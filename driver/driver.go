package driver

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sclevine/agouti"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func New(config *Config) *agouti.WebDriver {
	options := []agouti.Option{
		agouti.Desired(agouti.Capabilities{
			"loggingPrefs": map[string]string{
				"browser":     "INFO",
				"performance": "ALL",
			},
			"browserName": "chrome",
		}),
	}
	// add chrome args
	args := []string{"no-first-run",
		"--disable-gpu",
		"--disable-notifications",
		"--allow-insecure-localhost",
		"no-sandbox",
	}
	if len(config.ProxyAddr) > 0 {
		args = append(args, "--proxy-server="+config.ProxyAddr)
	}
	if config.Headless {
		args = append(args, "--headless")
	}
	options = append(options, agouti.ChromeOptions("args", args))

	// add chrome prefs
	var prefs = newPrefs()
	prefs.AddProfile("default_content_setting_values", "notifications", 2)
	if !config.ShowImages {
		prefs.AddProfile("managed_default_content_settings", "images", 2)
	}
	options = append(options, agouti.ChromeOptions("prefs", prefs))

	// add chrome binary path must require Chrome 62 at least
	binary := `C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`
	if len(config.Binary) > 0 {
		binary = config.Binary
	}
	options = append(options, agouti.ChromeOptions("binary", binary))

	return agouti.ChromeDriver(options...)
}

type prefs map[string]map[string]map[string]int

func newPrefs() prefs {
	p := make(prefs)
	p["profile"] = make(map[string]map[string]int)
	return p
}

func (pf *prefs) AddProfile(k1, k2 string, v int) {
	if profile, ok := (*pf)["profile"]; ok {
		if profile[k1] == nil {
			profile[k1] = make(map[string]int)
		}
		profile[k1][k2] = v
	}
}

func Get(page *agouti.Page, r *http.Request) (*http.Response, error) {
	response := new(http.Response)
	response.Request = r
	response.Header = http.Header{}
	if r.Method != http.MethodGet {
		return response, errors.New("driver only support http get method!")
	}
	reqURL := r.URL.String()
	logrus.Debug("diver is getting " + reqURL)
	err := page.Navigate(reqURL)
	if err != nil {
		return response, err
	}
	if _, err := page.PopupText(); err == nil {
		page.ConfirmPopup()
	}
	response.StatusCode = 200
	if s, err := page.URL(); err == nil {
		curURL, _ := url.Parse(s)
		if curURL.RequestURI() != r.RequestURI {
			response.StatusCode = 302
			response.Header.Set("location", s)
		}
	}
	time.Sleep(200 * time.Millisecond)
	s, _ := page.HTML()
	if len(s) == 0 {
		return response, errors.New("response html is empty! ")
	}
	response.Body = ioutil.NopCloser(strings.NewReader(s))
	if err != nil {
		return response, err
	}

	logs, _ := page.ReadNewLogs("performance")
	for _, v := range logs {
		var performance Performance
		if err := json.Unmarshal([]byte(v.Message), &performance); err != nil {
			return response, err
		}
		if performance.Message.Method == "Network.responseReceived" && performance.Message.Params.Response.URL == reqURL {
			statusCode := performance.Message.Params.Response.Status
			if statusCode > 400 {
				response.StatusCode = statusCode
				return response, nil
			}
			break
		}
	}
	return response, nil

}
