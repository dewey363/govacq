package http2

import (
	"github.com/gratno/tripod/hp"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36"

func newHeaders() http.Header {
	h := make(http.Header)
	h.Set("User-Agent", userAgent)
	h.Set("Connection", "close")
	return h
}

func NewGetRequest(s string) *http.Request {
	req, _ := http.NewRequest("GET", s, nil)
	req.Header = newHeaders()
	req.Header.Set("Referer", s)
	return req
}

type RedisConfig struct {
	ProxyName string
	Addr      string
	Password  string
	InitDB    int
	MaxActive int
}

type HttpClient struct {
	cfg     *RedisConfig
	Client  *http.Client
	IsProxy bool
}

func New(client *http.Client, isProxy bool, cfg *RedisConfig) *HttpClient {
	if isProxy && hp.RedisPool == nil {
		hp.InitRedis(cfg.Addr, cfg.Password, cfg.InitDB, cfg.MaxActive)
	}
	hc := &HttpClient{Client: client, IsProxy: isProxy, cfg: cfg}
	return hc
}

func (hc *HttpClient) Do(req *http.Request) (resp *http.Response, err error) {
	if hc.IsProxy {
		conn := hp.ProxyConn{Conn: hp.RedisPool.Get(), Name: hc.cfg.ProxyName}
		defer conn.Close()
		proxy, err := conn.RPopProxy()
		if err != nil {
			logrus.Fatalln(err)
		}
		tp := hp.CloneDefaultTransport()
		tp.Proxy = http.ProxyURL(&url.URL{Host: proxy.String()})
		hc.Client.Transport = tp
		hc.Client.Timeout = 10 * time.Second
		resp, err = hc.Client.Do(req)
		if err != nil {
			return resp, err
		}
		if resp.StatusCode >= 500 {
			return resp, err
		}
		conn.RPushProxy(proxy)
		return resp, err
	}

	return hc.Client.Do(req)
}
