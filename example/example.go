package main

import (
	"encoding/json"
	"fmt"
	"github.com/gratno/govacq/collectors/articles"
	"github.com/gratno/govacq/collectors/sitemap"
	"github.com/gratno/govacq/driver"
	"github.com/gratno/govacq/pipelines"
	"github.com/gratno/govacq/task"
	"github.com/gratno/govacq/tools"
	"github.com/gratno/tripod"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

var spider task.Spider

func init() {
	const settingsFile = "settings.toml"
	config := task.ParseConfig(settingsFile)
	spider = task.New(config)
}

func ExampleDir() {
	sm := sitemap.New("http://www.xianfeng.gov.cn/xfgov/index.html")
	cfg := driver.DefaultConfig()
	cfg.Binary = `E:\Program Files (x86)\Google\Chrome\Application\chrome.exe`
	sm.Config.DriverCfg = cfg
	sm.Config.AddSlash = true
	sm.ClientType = sitemap.HttpClient
	sm.Concurrent = 3
	if err := sm.Run(); err != nil {
		logrus.Fatalln(err)
	}
	dirs := sm.Result()
	b, _ := json.Marshal(dirs)
	err := ioutil.WriteFile("example_dir.txt", b, 0644)
	if err != nil {
		logrus.Errorln(err)
	}
}

func ExampleArticle() {
	b, _ := ioutil.ReadFile("example_dir.txt")
	var dirs []sitemap.SiteDir
	_ = json.Unmarshal(b, &dirs)
	finder := articles.NewFinder(dirs, "", "下一页")
	finder.Concurrent = 3
	finder.MaxPageSize = 10
	pages := finder.Find()
	logrus.Infoln("page len:", len(pages))
	b, _ = json.Marshal(pages)
	_ = ioutil.WriteFile("example_article.txt", b, 0644)
}

func ExampleTask() {
	var id uint = 91
	govTask := pipelines.FindGovTask(id)
	// 地图
	spider.SiteMap(govTask)
	return
	maps := pipelines.FindGovMaps(id)
	if len(maps) == 0 {
		return
	}
	// 获取文章
	spider.GetArticles(govTask, maps)
	return
	// 确认rules
	govArticles := pipelines.FindGovArticles(id, 0)
	// 更新内容
	spider.UpdateArtFields(govTask, govArticles)

	govArticles = pipelines.FindGovArticles(id, -1)
	spider.UpdateArtFields(govTask, govArticles)
	logrus.Infoln("更新成功数:", pipelines.CountGovArticles(id, 1),
		"失败数:", pipelines.CountGovArticles(id, -1), "未更新:", pipelines.CountGovArticles(id, 0))
}

func ExampleTaskExt(id uint) {
	arts := pipelines.FindKeyword2(id)
	for _, v := range arts {
		if _, err := time.Parse(tools.StandardBirth, v.Keyword2); err != nil {
			old := v.Keyword2
			v.Keyword2 = tools.FormatTime(old)
			fmt.Println(v.URL, "old:", old, "new:", v.Keyword2)
			pipelines.UpdateArticle(v)
		}
	}
}

func ExampleTaskExt2(id uint) {
	arts := pipelines.FindKeyword2(id)
	birth := "20060102"
	reg := regexp.MustCompile(`^t(\d+)_\d+\.shtml`)
	for _, v := range arts {
		if _, err := time.Parse(tools.StandardBirth, v.Keyword2); err != nil {
			old := v.Keyword2
			base := v.URL[strings.LastIndex(v.URL, "/")+1:]
			if reg.MatchString(base) {
				str := reg.FindStringSubmatch(base)[1]
				t, _ := time.Parse(birth, str)
				v.Keyword2 = t.Format(tools.StandardBirth)
			}
			fmt.Println(v.URL, "old:", old, "new:", v.Keyword2)
			pipelines.UpdateArticle(v)
		}
	}
}

func ExampleSpiltUrls() {
	const settingsFile = "settings.toml"
	config := task.ParseConfig(settingsFile)
	spider := task.New(config)
	var id uint = 91
	govTask := pipelines.FindGovTask(id)
	govArticles := pipelines.FindGovArticles(id, 0)
	var urls = make([]string, len(govArticles))
	for i, v := range govArticles {
		urls[i] = v.URL
	}
	spider.SpiltArticles(govTask, urls)
}

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: tripod.Birth, DisableLevelTruncation: true})
	//ExampleTask()
	//ExampleTaskExt(1)
	ExampleTaskExt2(1)

}
