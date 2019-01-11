package articles

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gratno/govacq/collectors/sitemap"
	"github.com/gratno/govacq/driver"
	"github.com/gratno/govacq/http2"
	"github.com/gratno/govacq/tools"
	"github.com/pkg/errors"
	"github.com/sclevine/agouti"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/transform"
	"gopkg.in/xmlpath.v2"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

var nextPageXpaths = []string{
	`//*[text()="%s"]/@*`,
	`//*[text()="%s"]/ancestor-or-self::*/@*`,
	`//*[@title="%s"]/@*`,
	`//*[@title="%s"]/ancestor-or-self::*/@*`,
	`//*[@alt="%s"]/@*`,
	`//*[@alt="%s"]/ancestor-or-self::*/@*`,
}

type Article struct {
	URL       string `json:"url"`
	FromCrcID uint32 `json:"from_crc_id"`
	From      string `json:"from"`
}

type Finder struct {
	Config
	getClients        []sitemap.GetClient
	pages             []*agouti.Page
	articleXpaths     []*xmlpath.Path
	taskQueue         chan sitemap.SiteDir
	mustUnexpectedURL sync.Map
	articles          sync.Map
	delta             int
}

// each website should confirm article xpath and next page text about it
func NewFinder(siteDirs []sitemap.SiteDir, extraArticleXpath, nextPage string) Finder {
	cfg := Config{Concurrent: 1, DriverCfg: driver.DefaultConfig(), MaxPageSize: 30, NextPageList: []string{"下一页", "下页"}, FirstPageClient: sitemap.HttpClient, NextPageActionType: AnalysisNextPage}
	f := Finder{Config: cfg}
	f.taskQueue = make(chan sitemap.SiteDir, len(siteDirs))
	for _, v := range siteDirs {
		f.taskQueue <- v
	}
	for _, v := range []string{`//li/descendant-or-self::*//*[@href]/@href`, `//dd/descendant-or-self::*//*[@href]/@href`} {
		f.articleXpaths = append(f.articleXpaths, xmlpath.MustCompile(v))
	}
	if len(extraArticleXpath) > 0 {
		f.articleXpaths = append(f.articleXpaths, xmlpath.MustCompile(extraArticleXpath))
	}
	if len(nextPage) > 0 {
		f.NextPageList = strings.Split(nextPage, tools.SplitStr)
	}
	f.FirstPageClient = sitemap.HttpClient
	f.NextPageActionType = AnalysisNextPage
	return f
}

func New(siteDirs []sitemap.SiteDir) Finder {
	return NewFinder(siteDirs, "", "")
}

func (fd *Finder) addArticle(baseURL *url.URL, nodeStr string, siteDir sitemap.SiteDir) bool {
	if removeQuoteBorderReg.MatchString(nodeStr) {
		nodeStr = removeQuoteBorderReg.FindStringSubmatch(nodeStr)[1]
	}
	ref, err := baseURL.Parse(nodeStr)
	if err != nil {
		return false
	}
	ref = baseURL.ResolveReference(ref)
	ArticleURL, _ := url.QueryUnescape(ref.String())
	if !utf8.Valid([]byte(ArticleURL)) {
		ArticleURL, _, _ = transform.String(tools.GBKTrans.NewDecoder(), ArticleURL)
	}
	if strings.Contains(ref.Host, baseURL.Host) {
		// exclude URL
		if _, ok := fd.mustUnexpectedURL.Load(ArticleURL); !ok {
			if _, ok := fd.articles.Load(ArticleURL); ok {
				return false
			}
			if baseURL.RequestURI() == ref.RequestURI() {
				return false
			}
			if strings.Contains(strings.ToLower(ref.RawQuery), "page") {
				return false
			}
			var article Article
			article.From = siteDir.URL
			article.FromCrcID = siteDir.CrcID
			article.URL = ArticleURL
			logrus.Infoln(ArticleURL, "is add!")
			fd.articles.Store(ArticleURL, article)
			return true
		}
	}
	return false
}

func (fd *Finder) findNextPageByAnalysis(baseURL, currentURL *url.URL, doc *xmlpath.Node, delta int, isFirst bool) *url.URL {
	if isFirst {
		for _, v := range fd.NextPageList {
			for _, xpath := range nextPageXpaths {
				xpath = fmt.Sprintf(xpath, v)
				compile := xmlpath.MustCompile(xpath)
				href := fd.getHref(baseURL, compile, doc)
				if href != nil {
					return href
				}
			}
		}
		return nil
	}

	tmp := *currentURL
	if fd.foundNextPage(&tmp, delta) {
		return &tmp
	}
	return nil
}

func (fd *Finder) findNextPageByClick(page *agouti.Page, preBodyLen int, interval time.Duration) (string, bool) {
	err := page.FirstByXPath(fmt.Sprintf(`//*[text()="%s"]`, fd.NextPageList[0])).MouseToElement()
	if err == nil {
		err = page.Click(agouti.SingleClick, agouti.LeftButton)
		if err == nil {
			time.Sleep(interval)
			tmpBody, _ := page.HTML()
			if len(tmpBody) != preBodyLen {
				return tmpBody, true
			}
		}
	}
	return "", false
}

var removeQuoteBorderReg = regexp.MustCompile(`'(.*?)'`)

func (fd *Finder) getHref(oriURL *url.URL, compile *xmlpath.Path, doc *xmlpath.Node) *url.URL {
	oriDir := path.Dir(oriURL.Path)
	iter := compile.Iter(doc)
	attrs := make([]string, 0)
	for iter.Next() {
		attrs = append(attrs, iter.Node().String())
	}
	mayRefs := make([]*url.URL, 0, len(attrs))
	for i := len(attrs) - 1; i >= 0; i-- {
		attr := attrs[i]
		if removeQuoteBorderReg.MatchString(attr) {
			attr = removeQuoteBorderReg.FindStringSubmatch(attr)[1]
		}
		ref, _ := oriURL.Parse(attr)
		reference := oriURL.ResolveReference(ref)
		if reference.Host == oriURL.Host {
			if len(reference.RawQuery) == 0 {
				reference.Path, _ = url.PathUnescape(reference.Path)
				refDir := path.Dir(reference.Path)
				if oriDir == refDir {
					mayRefs = append(mayRefs, reference)
				}
			} else {
				for query := range reference.Query() {
					if strings.Contains(strings.ToLower(query), "page") {
						return reference
					}
				}
				mayRefs = append(mayRefs, reference)
			}
		}
	}

	if len(mayRefs) > 0 {
		return mayRefs[0]
	}
	return nil
}

func (fd *Finder) foundNextPage(pre *url.URL, delta int) bool {
	var found bool
	if len(pre.RawQuery) > 0 {
		query := pre.Query()
		for k, v := range query {
			if k == "p" || strings.Contains(k, "page") {
				if len(v) > 0 {
					page, err := strconv.Atoi(v[0])
					if err == nil {
						query.Set(k, strconv.Itoa(page+delta))
						query.Del("AspxAutoDetectCookieSupport")
						pre.RawQuery = query.Encode()
						return true
					}
				}
			}
		}
		logrus.Infoln(pre.String(), "请手动输入下一页url:")
		var text string
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text = scanner.Text()
		URL := strings.TrimSpace(text)
		if len(URL) == 0 {
			return false
		}
		logrus.Infoln(URL, "请手动输入下一页key:")
		scanner.Scan()
		text = scanner.Text()
		key := strings.TrimSpace(text)
		if len(key) == 0 {
			return false
		}
		value := pre.Query().Get(key)
		page, err := strconv.Atoi(value)
		if err == nil {
			pre.Query().Set(key, strconv.Itoa(page+delta))
			found = true
		}
		return found
	}
	dir, base := path.Split(pre.Path)
	i := strings.IndexFunc(base, func(r rune) bool {
		if r == '.' || r == '?' {
			return true
		}
		return false
	})
	var prefix, suffix = base[0:i], base[i:]
	page := ""
	var j int
	for j = len(prefix) - 1; j >= 0; j-- {
		if !unicode.IsDigit(rune(prefix[j])) {
			break
		}
		page = string(prefix[j]) + page
	}
	pageInt, _ := strconv.Atoi(page)
	base = prefix[0:j+1] + strconv.Itoa(pageInt+delta) + suffix
	newPath := path.Join(dir, base)
	if newPath != pre.Path {
		pre.Path = newPath
		found = true
	}
	return found
}

var baseHrefPattern = xmlpath.MustCompile(`//base[1]/@href`)

func (fd *Finder) run(siteDir sitemap.SiteDir, goroutineID int) error {
	currentURL, _ := url.Parse(siteDir.URL)
	var bodyBytes []byte
	resp, err := fd.getClients[goroutineID].Do(http2.NewGetRequest(siteDir.URL))
	if err != nil {
		return err
	}
	if resp.StatusCode > 400 && resp.StatusCode < 500 {
		resp.Body.Close()
		logrus.Warnln("Ignore GET", currentURL, resp.StatusCode)
		return nil
	}
	bodyBytes, _ = ioutil.ReadAll(resp.Body)
	if len(bodyBytes) == 0 {
		return errors.New("获取页面内容失败! code:" + strconv.Itoa(resp.StatusCode))
	}

	pageCount := 0
	bodyStr := string(bodyBytes)
	_ = bodyStr

Acq:
	if pageCount > fd.MaxPageSize {
		logrus.Infoln(currentURL, "采集达到最大页")
		return nil
	}

	if len(bodyBytes) == 0 {
		resp, err = fd.getClients[goroutineID].Do(http2.NewGetRequest(currentURL.String()))
		if err != nil {
			logrus.Errorln(err, currentURL)
			goto Acq
		}
		if resp.StatusCode > 400 && resp.StatusCode < 500 {
			logrus.Warnln("Ignore GET", currentURL, resp.StatusCode)
			resp.Body.Close()
			return nil
		}

		if resp.StatusCode >= 500 {
			logrus.Warnln("GET", currentURL, tools.ErrStatusCode(resp.StatusCode))
			resp.Body.Close()
			time.Sleep(3 * time.Second)
			goto Acq
		}

		bodyBytes, _ = ioutil.ReadAll(resp.Body)
	}

	if !utf8.Valid(bodyBytes) {
		bodyBytes, _, _ = transform.Bytes(tools.GBKTrans.NewDecoder(), bodyBytes)
	}

	fd.mustUnexpectedURL.Store(currentURL.String(), true)
	logrus.Infoln("Parsing", currentURL)
	doc, _ := xmlpath.ParseHTML(bytes.NewReader(bodyBytes))

	var baseURL *url.URL
	if s, ok := baseHrefPattern.String(doc); ok {
		baseURL, _ = url.Parse(s)
	} else {
		baseURL = currentURL
	}
	addCount := 0
	for _, xpath := range fd.articleXpaths {
		iter := xpath.Iter(doc)
		for iter.Next() {
			nodeStr := iter.Node().String()
			if fd.addArticle(baseURL, nodeStr, siteDir) {
				addCount++
			}
		}
	}

	pageCount++

	if strings.Contains(currentURL.String(), "http://www.zengdu.gov.cn/zhanqun/gongshang/wu/") {
		fmt.Println(currentURL)
	}

	if addCount > 0 {
		switch fd.NextPageActionType {
		case AnalysisNextPage:
			nextURL := fd.findNextPageByAnalysis(baseURL, currentURL, doc, fd.delta, pageCount == 1)
			if nextURL != nil {
				bodyBytes = []byte{} // need reget resp body
				currentURL = nextURL
				goto Acq
			}
		case ClickNextPage:
			if s, ok := fd.findNextPageByClick(fd.pages[goroutineID], len(bodyBytes), 500*time.Millisecond); ok {
				bodyBytes = []byte(s) // reset resp body
				goto Acq
			}
		}
	}
	logrus.Infoln("不是可翻页栏目", currentURL)
	return nil
}

func (fd *Finder) Find() (artPages []Article) {
	if fd.Reverse {
		fd.delta = -1
	} else {
		fd.delta = 1
	}

	var chrome *agouti.WebDriver
	fd.getClients, chrome, fd.pages = sitemap.GenClients(fd.FirstPageClient, fd.Concurrent, fd.DriverCfg, true, fd.HttpIsProxy, fd.RedisCfg)
	defer func() {
		if chrome != nil {
			chrome.Stop()
		}
	}()
	wg := sync.WaitGroup{}
	logrus.Infoln("任务目录数:", len(fd.taskQueue))
	for i := 0; i < fd.Concurrent; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				select {
				case siteDir := <-fd.taskQueue:
					if err := fd.run(siteDir, i); err != nil {
						logrus.Errorln(err, siteDir.URL)
						fd.taskQueue <- siteDir
					}
				case <-time.After(30 * time.Second):
					return
				}
			}
		}(i)
	}
	wg.Wait()
	logrus.Infoln("获取所有文章任务完成! ")

	fd.mustUnexpectedURL.Range(func(key, _ interface{}) bool {
		if _, ok := fd.articles.Load(key); ok {
			fd.articles.Delete(key)
		}
		return true
	})

	fd.articles.Range(func(_, value interface{}) bool {
		article := value.(Article)
		if len(strings.TrimSpace(article.URL)) > 0 {
			artPages = append(artPages, article)
		}
		return true
	})

	return artPages

}

var _ = []string{"下一页",
	"下页",
	">",
	">>",
	"next",
}
