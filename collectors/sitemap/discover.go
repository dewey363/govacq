package sitemap

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gratno/govacq/http2"
	"github.com/gratno/govacq/tools"
	"github.com/gratno/tripod/hp"
	"github.com/pkg/errors"
	"github.com/samclarke/robotstxt"
	"github.com/sclevine/agouti"
	"github.com/sirupsen/logrus"
	"hash/crc32"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type SiteDir struct {
	CrcID     uint32 `json:"crc_id"`
	URL       string `json:"url"`
	Alias     string `json:"alias"`
	From      string `json:"from"`
	FromAlias string `json:"from_alias"`
	curDeep   int
}

// web site map collector
type SiteMap struct {
	Config
	root      *url.URL
	dirs      sync.Map
	taskQueue chan SiteDir
	robots    *robotstxt.RobotsTxt
}

func New(URL string) *SiteMap {
	var err error
	sm := new(SiteMap)
	sm.root, err = url.Parse(URL)
	if err != nil {
		panic(err)
	}
	sm.Config = DefaultConfig()
	sm.taskQueue = make(chan SiteDir, 1000)
	sm.AddMustMatchHost(sm.root.Host)
	crcID := crc32.ChecksumIEEE([]byte(URL))
	rootDir := SiteDir{CrcID: crcID, URL: URL, Alias: "首页"}
	sm.dirs.Store(crcID, rootDir)
	sm.taskQueue <- rootDir
	return sm
}

func (sm *SiteMap) Run() error {
	if sm.Concurrent < 1 {
		return errors.New("invalid concurrent value for sitemap! ")
	}
	sm.setRobots()
	rootDir := <-sm.taskQueue
	var (
		getClients []GetClient
		chrome     *agouti.WebDriver
	)

	getClients, chrome, _ = GenClients(sm.ClientType, sm.Concurrent, sm.DriverCfg, false, sm.HttpIsProxy, sm.RedisCfg)
	defer func() {
		if chrome != nil {
			chrome.Stop()
		}
	}()
	logrus.Infoln("Processing "+rootDir.URL, "From:", "/")
	err := sm.discover(rootDir, getClients[0])
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for i := 0; i < sm.Concurrent; i++ {
		wg.Add(1)
		client := getClients[i]
		go func(client GetClient) {
			sm.worker(client)
			wg.Done()
		}(client)
	}
	wg.Wait()
	close(sm.taskQueue)

	return nil
}

func (sm *SiteMap) Result() (dirs []SiteDir) {
	sm.dirs.Range(func(key, value interface{}) bool {
		dirs = append(dirs, value.(SiteDir))
		return true
	})

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].URL < dirs[j].URL
	})
	return
}

func (sm *SiteMap) CleanDuplicateURLType(dirs []SiteDir) {
	urls := make([]string, len(dirs))
	for i, v := range dirs {
		fmt.Println("Url:", v.URL, "Alias:", v.Alias, "From:", v.From)
		urls[i] = v.URL
	}
	fmt.Println("===================")
	matchers := tools.SplitURL1(urls)
	fmt.Println(matchers)
	panic("implemented me")
	return
}

func (sm *SiteMap) discover(oriDir SiteDir, client GetClient) error {
	resp, err := client.Do(http2.NewGetRequest(oriDir.URL))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode >= 500:
		return errors.Wrap(err, "服务器出错! ")
	case resp.StatusCode > 400 && resp.StatusCode < 500:
		logrus.Warnln("GET", oriDir.URL, tools.ErrStatusCode(resp.StatusCode))
		return nil
	case resp.StatusCode > 300 && resp.StatusCode < 400:
		loc, _ := resp.Location()
		location := loc.String()
		crcID := crc32.ChecksumIEEE([]byte(oriDir.URL))
		v, _ := sm.dirs.Load(crcID)
		if v != nil {
			sm.dirs.Delete(crcID)
			crcID = crc32.ChecksumIEEE([]byte(location))
			vdir := v.(SiteDir)
			vdir.URL = location
			sm.dirs.Store(crcID, vdir)
			oriDir.URL = location
		}
	case resp.StatusCode != 200:
		logrus.Infoln("Ignore", oriDir.URL)
		return nil
	}
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	hrefAttrs := []string{"href", "src"}
	for _, v := range hrefAttrs {
		document.Find("*[" + v + "]").Each(func(_ int, sel *goquery.Selection) {
			href, _ := sel.Attr(v)
			ref := tools.CleanHref(oriDir.URL, href, sm.AddSlash)
			if ref == nil {
				return
			}
			if !sm.Validate(ref) {
				return
			}

			for query := range ref.Query() {
				if strings.Contains(strings.ToLower(query), "page") {
					return
				}
			}
			refUrl := ref.String()
			// check pass
			crcID := crc32.ChecksumIEEE([]byte(refUrl))
			if _, ok := sm.dirs.Load(crcID); ok {
				return
			}
			alias := sel.AttrOr("alt", sel.AttrOr("title", sel.Text()))
			if len(alias) == 0 || strings.HasPrefix(alias, "更多") || strings.HasPrefix(strings.ToLower(alias), "more") {
				alias = ref.Path
			}
			dir := SiteDir{CrcID: crcID, URL: refUrl, Alias: strings.TrimSpace(alias), From: oriDir.URL, FromAlias: strings.TrimSpace(oriDir.Alias), curDeep: oriDir.curDeep + 1}
			if _, ok := sm.dirs.LoadOrStore(crcID, dir); !ok {
				logrus.Infoln("Add "+dir.URL, "From "+dir.From)
				sm.taskQueue <- dir
			}
		})
	}
	return nil
}

func (sm *SiteMap) worker(client GetClient) {
	for {
		select {
		case v := <-sm.taskQueue:
			if v.curDeep > sm.Deep {
				logrus.Infoln("drop", v.URL, "已达到最大深度!")
				continue
			}
			if sm.robots != nil {
				if allowed, _ := sm.robots.IsAllowed("*", v.URL); !allowed {
					logrus.Warnln(v.URL, "is not allowed in robots! ")
					continue
				}
			}
			logrus.Infoln("Processing "+v.URL, "From:", v.From)
			for i := 0; i < 2; i++ {
				err := sm.discover(v, client)
				if err == nil {
					break
				}
				logrus.Errorln(&url.Error{Op: "GET", URL: v.URL, Err: err}, "From:", v.From)
				logrus.Infoln("After 5s retry...")
				time.Sleep(5 * time.Second)
			}
		case <-time.After(30 * time.Second):
			logrus.Infoln(sm.root, "获取列表完成! ")
			return
		}
	}
}

func (sm *SiteMap) setRobots() {
	logrus.Infoln("正在设置", sm.root, "robots.txt")
	robotsURL := &url.URL{Scheme: sm.root.Scheme, Host: sm.root.Host, Path: "/robots.txt"}
	resp, err := hp.DefaultClient.Get(robotsURL.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}
	b, _ := ioutil.ReadAll(resp.Body)
	logrus.Infoln(sm.root, "robots.txt\n", len(b))
	sm.robots, _ = robotstxt.Parse(string(b), sm.root.String())
}
