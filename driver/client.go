package driver

import (
	"github.com/sclevine/agouti"
	"net/http"
	"time"
)

type ChromeClient struct {
	Page *agouti.Page
}

func (c *ChromeClient) SetTimeout(timeout time.Duration) {
	_ = c.Page.SetImplicitWait(int(timeout / time.Millisecond))
}

func (c *ChromeClient) Do(req *http.Request) (*http.Response, error) {
	return Get(c.Page, req)
}
