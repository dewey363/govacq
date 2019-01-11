package articles

import (
	"bytes"
	"github.com/gratno/govacq/collectors/sitemap"
	"github.com/gratno/govacq/entity"
	"github.com/gratno/govacq/http2"
	"github.com/gratno/govacq/tools"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/transform"
	"gopkg.in/xmlpath.v2"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var DeferParseHosts sync.Map

func ValidateRules(rules ...entity.GovArtRule) bool {
	for _, v := range rules {
		if len(v.Keyword1Rule) > 0 {
			_, err := xmlpath.Compile(v.Keyword1Rule)
			if err != nil {
				logrus.Warnln(v.URLExpr, "keyword1", "xpath规则校验未通过! ", err)
				return false
			}
		}
		if len(v.Keyword2Rule) > 0 {
			_, err := xmlpath.Compile(v.Keyword2Rule)
			if err != nil {
				logrus.Warnln(v.URLExpr, "keyword2", "xpath规则校验未通过! ", err)
				return false
			}
		}
		if len(v.Keyword3Rule) > 0 {
			_, err := xmlpath.Compile(v.Keyword3Rule)
			if err != nil {
				logrus.Warnln(v.URLExpr, "keyword3", "xpath规则校验未通过! ", err)
				return false
			}
		}
		if len(v.Keyword4Rule) > 0 {
			_, err := xmlpath.Compile(v.Keyword4Rule)
			if err != nil {
				logrus.Warnln(v.URLExpr, "keyword4", "xpath规则校验未通过! ", err)
				return false
			}
		}
		if len(v.Keyword5Rule) > 0 {
			_, err := xmlpath.Compile(v.Keyword5Rule)
			if err != nil {
				logrus.Warnln(v.URLExpr, "keyword5", "xpath规则校验未通过! ", err)
				return false
			}
		}
		if len(v.Keyword6Rule) > 0 {
			_, err := xmlpath.Compile(v.Keyword6Rule)
			if err != nil {
				logrus.Warnln(v.URLExpr, "keyword6", "xpath规则校验未通过! ", err)
				return false
			}
		}
	}

	return true
}

func ParseArticle(client sitemap.GetClient, govArt *entity.GovArticle, rule entity.GovArtRule) {
	// 匹配成功 获取文章内容
	u, _ := url.Parse(govArt.URL)
	var (
		count interface{}
		ok    bool
	)
	count, ok = DeferParseHosts.Load(u.Host)
	if ok {
		if count.(int) > 100 {
			logrus.Infoln(govArt.URL, "网站暂时无法访问!")
			govArt.Status = -1
			return
		}
	} else {
		DeferParseHosts.Store(u.Host, 0)
	}

	resp, err := client.Do(http2.NewGetRequest(govArt.URL))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 || resp.StatusCode == 403 {
		now := time.Now()
		govArt.DeletedAt = &now
		return
	}
	if resp.StatusCode != 200 {
		DeferParseHosts.Store(u.Host, count.(int)+1)
		govArt.Status = -1
		logrus.Warnln(govArt.URL, tools.ErrStatusCode(resp.StatusCode))
		return
	}
	DeferParseHosts.Store(u.Host, 1)

	var reader io.Reader
	bt, _ := ioutil.ReadAll(resp.Body)
	if !utf8.Valid(bt) {
		bt, _, _ = transform.Bytes(tools.GBKTrans.NewDecoder(), bt)
	}
	reader = bytes.NewReader(bt)
	doc, _ := xmlpath.ParseHTML(reader)
	// keyword1
	if len(rule.Keyword1Rule) > 0 {
		pattern := xmlpath.MustCompile(rule.Keyword1Rule)
		govArt.Keyword1, ok = pattern.String(doc)
		if !ok {
			logrus.Warnln(govArt.URL, "keyword1不匹配对应规则! ")
		}
		govArt.Keyword1 = strings.TrimSpace(govArt.Keyword1)
	}
	// keyword2
	if len(rule.Keyword2Rule) > 0 {
		pattern := xmlpath.MustCompile(rule.Keyword2Rule)
		govArt.Keyword2, ok = pattern.String(doc)
		if !ok {
			logrus.Warnln(govArt.URL, "keyword2不匹配对应规则! ")
		}
		govArt.Keyword2 = strings.TrimSpace(govArt.Keyword2)
	}

	// keyword3
	if len(rule.Keyword3Rule) > 0 {
		pattern := xmlpath.MustCompile(rule.Keyword3Rule)
		govArt.Keyword3, ok = pattern.String(doc)
		if !ok {
			logrus.Warnln(govArt.URL, "keyword3不匹配对应规则! ")
		}
		govArt.Keyword3 = strings.TrimSpace(govArt.Keyword3)
	}
	// keyword4
	if len(rule.Keyword4Rule) > 0 {
		pattern := xmlpath.MustCompile(rule.Keyword4Rule)
		govArt.Keyword4, ok = pattern.String(doc)
		if !ok {
			logrus.Warnln(govArt.URL, "keyword4不匹配对应规则! ")
		}
		govArt.Keyword4 = strings.TrimSpace(govArt.Keyword4)
	}
	// keyword5
	if len(rule.Keyword5Rule) > 0 {
		pattern := xmlpath.MustCompile(rule.Keyword5Rule)
		govArt.Keyword5, ok = pattern.String(doc)
		if !ok {
			logrus.Warnln(govArt.URL, "keyword5不匹配对应规则! ")
		}
		govArt.Keyword5 = strings.TrimSpace(govArt.Keyword5)
	}

	// keyword6
	if len(rule.Keyword6Rule) > 0 {
		pattern := xmlpath.MustCompile(rule.Keyword6Rule)
		govArt.Keyword6, ok = pattern.String(doc)
		if !ok {
			logrus.Warnln(govArt.URL, "keyword6不匹配对应规则! ")
		}
		govArt.Keyword6 = strings.TrimSpace(govArt.Keyword6)
	}
	govArt.Status = 1
}
