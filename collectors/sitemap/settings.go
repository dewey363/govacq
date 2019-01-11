package sitemap

import (
	"github.com/gratno/govacq/driver"
	"github.com/gratno/govacq/http2"
	"github.com/gratno/tripod/hp"
	"github.com/sclevine/agouti"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

var ValidSuffix = []string{
	"/", ".htm", ".html", ".shtm", ".shtml", ".jhtml", ".jhtm", ".jsp", ".aspx", ".asp", ".php",
}

var IgnoreFiled = []string{
	"content", "article",
}

var ValidBase = []string{
	"(?i)^index", "(?i)^default.",
}

const (
	mustRegExprLevel levelVerifier = iota
	mustMatchHostLevel
	ignoreKeywordLevel
	matcherSuffixLevel
	matcherRegBaseLevel
)

type Verifier interface {
	Validate(URL *url.URL) bool
}

type levelVerifier int

func (l *levelVerifier) Validate(URL *url.URL, vs []string) bool {
	switch *l {
	case mustRegExprLevel:
		for _, v := range vs {
			if regexp.MustCompile(v).MatchString(URL.String()) {
				return true
			}
		}
	case mustMatchHostLevel:
		for _, v := range vs {
			if strings.Contains(URL.Host, v) {
				return true
			}
		}
	case ignoreKeywordLevel:
		for _, v := range vs {
			if strings.Contains(URL.RequestURI(), v) {
				return false
			}
		}
		return true
	case matcherSuffixLevel:
		for _, v := range vs {
			if strings.HasSuffix(URL.Path, v) {
				return true
			}
		}
	case matcherRegBaseLevel:
		if strings.HasSuffix(URL.Path, "/") {
			return true
		}
		for _, v := range vs {
			if regexp.MustCompile(v).MatchString(path.Base(URL.Path)) {
				return true
			}
		}
	}
	return false
}

type GetClient interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	ChromeClient ClientType = iota
	HttpClient
)

type ClientType int

func GenClients(clientType ClientType, concurrent int,
	config *driver.Config, allowedRedirect, isProxy bool, redisCfg *http2.RedisConfig) (getClients []GetClient, chrome *agouti.WebDriver, pages []*agouti.Page) {
	switch clientType {
	case ChromeClient:
		chrome = driver.New(config)
		if err := chrome.Start(); err != nil {
			logrus.Fatalln("chrome start failed! ", err)
		}
		for i := 0; i < concurrent; i++ {
			page, _ := chrome.NewPage()
			client := driver.ChromeClient{Page: page}
			client.SetTimeout(30 * time.Second)
			pages = append(pages, page)
			getClients = append(getClients, &client)
		}

	case HttpClient:
		for i := 0; i < concurrent; i++ {
			client := hp.CloneDefaultClient()
			if allowedRedirect {
				client.CheckRedirect = nil
			}
			httpClient := http2.New(client, isProxy, redisCfg)
			getClients = append(getClients, httpClient)
		}
	}
	return
}

type Config struct {
	DriverCfg   *driver.Config
	Deep        int
	rules       [5][]string
	AddSlash    bool
	ClientType  ClientType
	Concurrent  int
	HttpIsProxy bool
	RedisCfg    *http2.RedisConfig
}

const defaultSpiderDeep = 2

func DefaultConfig() Config {
	c := Config{Deep: defaultSpiderDeep, AddSlash: true, Concurrent: 1}
	c.AddMatcherSuffix(ValidSuffix...)
	c.AddMatcherRegBase(ValidBase...)
	c.AddIgnoreKeyword(IgnoreFiled...)
	return c
}

func (c *Config) AddMustMatchHost(hosts ...string) {
	c.rules[mustMatchHostLevel] = append(c.rules[mustMatchHostLevel], hosts...)
}

func (c *Config) AddMustRegExpr(exprs ...string) bool {
	for _, expr := range exprs {
		_, err := regexp.Compile(expr)
		if err != nil {
			return false
		}
		c.rules[mustRegExprLevel] = append(c.rules[mustRegExprLevel], expr)
	}
	return true
}

func (c *Config) AddIgnoreKeyword(keywords ...string) {
	c.rules[ignoreKeywordLevel] = append(c.rules[ignoreKeywordLevel], keywords...)
}

func (c *Config) AddMatcherSuffix(suffixes ...string) {
	c.rules[matcherSuffixLevel] = append(c.rules[matcherSuffixLevel], suffixes...)
}

func (c *Config) AddMatcherRegBase(bases ...string) bool {
	for _, expr := range bases {
		_, err := regexp.Compile(expr)
		if err != nil {
			return false
		}
		c.rules[matcherRegBaseLevel] = append(c.rules[matcherRegBaseLevel], expr)
	}
	return true
}

func (c *Config) validate(level levelVerifier, URL *url.URL) bool {
	return level.Validate(URL, c.rules[level])
}

func (c *Config) Validate(URL *url.URL) bool {
	for i := range c.rules {
		level := levelVerifier(i)
		if level == mustRegExprLevel {
			if c.validate(level, URL) {
				return true
			}
			continue
		}
		if !c.validate(level, URL) {
			return false
		}
	}
	return true
}
