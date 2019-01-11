package driver

// http response 日志
type Performance struct {
	Message struct {
		Method string `json:"method"`
		Params struct {
			Response struct {
				Status     int    `json:"status"`
				StatusText string `json:"statusText"`
				URL        string `json:"url"`
			} `json:"response"`
		} `json:"params"`
	} `json:"message"`
}
