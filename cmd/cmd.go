package main

import (
	"context"
	"github.com/gratno/govacq/pipelines"
	"github.com/gratno/govacq/task"
	"github.com/gratno/tripod"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"os"
	"sync"
)

var spider task.Spider

func init() {
	const settingsFile = "settings.toml"
	config := task.ParseConfig(settingsFile)
	spider = task.New(config)
}

func CreateTask(urls ...string) {
	pipelines.CreateGovTask(urls...)
}

func SiteMap(ids ...uint) {
	ctx := context.Background()
	pool := semaphore.NewWeighted(10)
	wg := sync.WaitGroup{}
	for _, id := range ids {
		pool.Acquire(ctx, 1)
		wg.Add(1)
		go func(id uint) {
			defer pool.Release(1)
			defer wg.Done()
			govTask := pipelines.FindGovTask(id)
			spider.SiteMap(govTask)
		}(id)
	}
	wg.Wait()
	logrus.Infoln("sitemap任务全部完成! ")
}

func GetArticles(ids ...uint) {
	ctx := context.Background()
	pool := semaphore.NewWeighted(10)
	wg := sync.WaitGroup{}
	for _, id := range ids {
		pool.Acquire(ctx, 1)
		wg.Add(1)
		go func(id uint) {
			defer pool.Release(1)
			defer wg.Done()
			govTask := pipelines.FindGovTask(id)
			maps := pipelines.FindGovMaps(id)
			if len(maps) == 0 {
				return
			}
			// 获取文章
			spider.GetArticles(govTask, maps)
		}(id)
	}
	wg.Wait()
	logrus.Infoln("get articles任务全部完成! ")
}

func SplitArticles(id uint) {
	govTask := pipelines.FindGovTask(id)
	govArticles := pipelines.FindGovArticles(id, 0)
	var urls = make([]string, len(govArticles))
	for i, v := range govArticles {
		urls[i] = v.URL
	}
	spider.SpiltArticles(govTask, urls)
}

func UpdateArticles(id uint) {
	// 确认rules
	govArticles := pipelines.FindGovArticles(id, 0)
	// 更新内容
	govTask := pipelines.FindGovTask(id)
	spider.UpdateArtFields(govTask, govArticles)

	govArticles = pipelines.FindGovArticles(id, -1)
	spider.UpdateArtFields(govTask, govArticles)
	logrus.Infoln("更新成功数:", pipelines.CountGovArticles(id, 1),
		"失败数:", pipelines.CountGovArticles(id, -1), "未更新:", pipelines.CountGovArticles(id, 0))
}

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: tripod.Birth, DisableLevelTruncation: true})
	//CreateTask("http://www.chongyang.gov.cn/")
	//SiteMap(1)
	//GetArticles(1)
	//SplitArticles(1)
	UpdateArticles(1)
}
