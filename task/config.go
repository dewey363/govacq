package task

import (
	"github.com/BurntSushi/toml"
	"github.com/gratno/govacq/driver"
	"github.com/gratno/govacq/http2"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DriverConfig driver.Config     `toml:"DriverConfig"`
	DBConfig     DBConfig          `toml:"DBConfig"`
	RedisConfig  http2.RedisConfig `toml:"RedisConfig"`
}

type DBConfig struct {
	Dialect  string
	User     string
	Password string
	DBName   string
}

func ParseConfig(settingFile string) Config {
	var c Config
	if _, err := toml.DecodeFile(settingFile, &c); err != nil {
		logrus.Fatalln(err)
	}
	return c
}
