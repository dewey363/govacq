package task

import (
	"fmt"
	"github.com/gratno/govacq/collectors/articles"
	"github.com/gratno/govacq/collectors/sitemap"
	"github.com/gratno/govacq/driver"
	"github.com/gratno/govacq/entity"
	"github.com/gratno/govacq/http2"
	"github.com/gratno/govacq/pipelines"
	"github.com/gratno/govacq/tools"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/transform"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var MinMapNum = 10

type Spider struct {
	DBCfg     *DBConfig
	DriverCfg *driver.Config
	RedisCfg  *http2.RedisConfig
}

func New(config Config) Spider {
	if pipelines.DB == nil {
		dbConfig := config.DBConfig
		var dsn string
		switch strings.ToLower(dbConfig.Dialect) {
		case "mysql":
			dsn = fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true", dbConfig.User, dbConfig.Password, dbConfig.DBName)
		case "sqlite3":
			dsn = fmt.Sprintf("file:%[3]s?_auth&_auth_user=%[1]s&_auth_pass=%[2]s&_auth_crypt=sha1", dbConfig.User, dbConfig.Password, dbConfig.DBName)
		default:
			logrus.Fatalln("暂不支持该数据库类型!", dbConfig.Dialect)
		}
		pipelines.Init(dbConfig.Dialect, dsn)
	}
	return Spider{DriverCfg: &config.DriverConfig, RedisCfg: &config.RedisConfig, DBCfg: &config.DBConfig}
}

// 获取网站地图
func (sd *Spider) SiteMap(task entity.GovTask) {
	setting := pipelines.FindMapSetting(task.ID)
	sm := sitemap.New(task.URL)
	sm.Deep = int(setting.Deep)
	sm.AddSlash = setting.AddSlash
	sm.ClientType = sitemap.ClientType(setting.ClientType)
	sm.Concurrent = int(setting.Concurrent)
	if sd.DriverCfg != nil {
		sm.DriverCfg = sd.DriverCfg
	}
	sm.RedisCfg = sd.RedisCfg
	sm.HttpIsProxy = setting.HttpUseProxy
	if len(strings.TrimSpace(setting.MustMatchHost)) > 0 {
		sm.AddMustMatchHost(strings.Split(setting.MustMatchHost, tools.SplitStr)...)
	}
	if len(strings.TrimSpace(setting.MustRegExpr)) > 0 {
		sm.AddMustRegExpr(strings.Split(setting.MustRegExpr, tools.SplitStr)...)
	}
	if len(strings.TrimSpace(setting.IgnoreKeyword)) > 0 {
		sm.AddIgnoreKeyword(strings.Split(setting.IgnoreKeyword, tools.SplitStr)...)
	}
	if len(strings.TrimSpace(setting.MatcherSuffix)) > 0 {
		sm.AddMatcherSuffix(strings.Split(setting.MatcherSuffix, tools.SplitStr)...)
	}
	if len(strings.TrimSpace(setting.MatcherRegBase)) > 0 {
		sm.AddMatcherRegBase(strings.Split(setting.MatcherRegBase, tools.SplitStr)...)
	}
	if err := sm.Run(); err != nil {
		logrus.Fatalln(err)
	}
	dirs := sm.Result()

	if len(dirs) < MinMapNum {
		logrus.Warnln(task.URL, "采到的目录数太少,放弃本次任务! len:", len(dirs))
		return
	}
	govmaps := make([]entity.GovMap, len(dirs))
	for i, v := range dirs {
		alias := v.Alias
		if !utf8.Valid([]byte(alias)) {
			alias, _, _ = transform.String(tools.GBKTrans.NewDecoder(), alias)
		}
		govmaps[i] = entity.GovMap{UID: task.ID, Crc32ID: v.CrcID, Title: alias, URL: v.URL, SourceURL: v.From}
	}

	pipelines.SaveMaps(govmaps)
	pipelines.UpdateMapHDTitle(task.ID)
	pipelines.UpdateTaskStatus(task.ID, 1)
	logrus.Infoln(task.URL, "网站地图已采集完成! 有效目录数:", len(govmaps))
}

// 获取网站文章 如果要更新字段信息 需要在结束后填写gov_art_rules表中规则
func (sd *Spider) GetArticles(task entity.GovTask, govMaps []entity.GovMap) {
	if len(govMaps) == 0 {
		logrus.Warnln(task.URL, "maps is zero! ")
		return
	}
	setting := pipelines.FindArtSetting(task.ID)
	dirs := make([]sitemap.SiteDir, len(govMaps))
	for i, v := range govMaps {
		dirs[i] = sitemap.SiteDir{CrcID: uint32(v.ID), URL: v.URL, From: v.SourceURL}
	}
	finder := articles.NewFinder(dirs, setting.ExtArtXpath, setting.NextPage)
	if sd.DriverCfg != nil {
		finder.DriverCfg = sd.DriverCfg
	}
	finder.RedisCfg = sd.RedisCfg
	finder.HttpIsProxy = setting.HttpUseProxy
	finder.Concurrent = int(setting.Concurrent)
	finder.MaxPageSize = int(setting.MaxPageSize)
	finder.Reverse = setting.Reverse
	finder.FirstPageClient = sitemap.ClientType(setting.FirstPageClient)
	finder.NextPageActionType = articles.ActionType(setting.NextPageActionType)
	artPages := finder.Find()
	logrus.Infoln(task.URL, "发现文章数:", len(artPages), "正在存储文章...")
	arts := make([]entity.GovArticle, 0, len(artPages))
	urls := make([]string, 0, len(artPages))

	dirUrls := make(map[string]bool)
	for _, v := range dirs {
		dirUrls[v.URL] = true
	}
	for _, v := range artPages {
		if _, ok := dirUrls[v.URL]; !ok {
			arts = append(arts, entity.GovArticle{UID: uint(v.FromCrcID), UUID: task.ID, URL: v.URL, SourceURL: v.From})
			urls = append(urls, v.URL)
		}
	}
	pipelines.SaveArticles(arts)
	pipelines.UpdateMapsStatus(task.ID, 1)
	logrus.Infoln(task.URL, "存储文章完成! ")
	if !setting.IsSplitArt {
		return
	}

	sd.SpiltArticles(task, urls)

}

func (sd *Spider) SpiltArticles(task entity.GovTask, urls []string) {
	logrus.Infoln(task.URL, "正在对文章分类...")
	maps := pipelines.FindGovMaps(task.ID)
	urlCounterMatchers, err := tools.SplitURL0(urls, len(maps))
	if err != nil {
		logrus.Errorln(err)
		urlCounterMatchers = tools.SplitURL1(urls)
	}
	govArtRules := make([]entity.GovArtRule, len(urlCounterMatchers))
	for i, v := range urlCounterMatchers {
		govArtRules[i] = entity.GovArtRule{UUID: task.ID, URLExpr: v.Expr, Example: v.Example, DataCount: v.Count}
	}
	pipelines.SaveArtRules(govArtRules)
	logrus.Infoln(task.URL, "文章分类完成, 请确认将要采集字段匹配规则!")
}

// 根据gov_art_rules表中规则更新相关字段信息
func (sd *Spider) UpdateArtFields(task entity.GovTask, govArticles []entity.GovArticle) {

	setting := pipelines.FindArtUpdateSetting(task.ID)
	// type为1的表示文章
	artRules := pipelines.FindArtUpdateRules(task.ID, 1)
	if len(artRules) == 0 {
		logrus.Fatalln("还未给", task.ID, "文章配置规则!")
	}
	if !articles.ValidateRules(artRules...) {
		logrus.Fatalln("rules 校验未通过! ")
	}

	if sd.DBCfg.Dialect == "sqlite3" {
		// sqlite 不支持多并发写
		setting.Concurrent = 1
	}

	getCorrespondRule := func(article *entity.GovArticle) (entity.GovArtRule, bool) {
		for _, r := range artRules {
			reg := regexp.MustCompile(r.URLExpr)
			if reg.MatchString(article.URL) {
				return r, true
			}
		}
		return entity.GovArtRule{}, false
	}

	getClients, chrome, _ := sitemap.GenClients(sitemap.ClientType(setting.ClientType), int(setting.Concurrent), sd.DriverCfg, true, setting.UseProxy, sd.RedisCfg)
	defer func() {
		if chrome != nil {
			chrome.Stop()
		}
	}()

	taskQueue := make(chan entity.GovArticle, len(govArticles))
	defer close(taskQueue)
	for _, art := range govArticles {
		taskQueue <- art
	}

	wg := sync.WaitGroup{}
	for i := 0; i < int(setting.Concurrent); i++ {
		wg.Add(1)
		getClient := getClients[i]
		go func(getClient sitemap.GetClient) {
			defer wg.Done()
			for {
				select {
				case art := <-taskQueue:
					artRule, ok := getCorrespondRule(&art)
					if !ok {
						// 未在规则中会被删除!
						now := time.Now()
						art.DeletedAt = &now
					} else {
						articles.ParseArticle(getClient, &art, artRule)
					}
					logrus.Infoln("updating", art.URL)
					pipelines.UpdateArticle(art)
				case <-time.After(3 * time.Second):
					return
				}
			}
		}(getClient)
	}
	wg.Wait()
	pipelines.DelArtUnScoped()
	logrus.Infoln(task.ID, task.URL, "更新文章相关字段信息任务完成! ")
}
