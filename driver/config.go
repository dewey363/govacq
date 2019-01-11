package driver

type Config struct {
	Binary     string
	ProxyAddr  string
	Headless   bool
	ShowImages bool
}

func DefaultConfig() *Config {
	return &Config{Headless: true, ShowImages: false}
}
