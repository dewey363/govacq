package articles

import (
	"github.com/gratno/govacq/collectors/sitemap"
	"github.com/gratno/govacq/driver"
	"github.com/gratno/govacq/http2"
)

// 主要通过翻页去获取文章
type Config struct {
	Concurrent         int
	DriverCfg          *driver.Config
	MaxPageSize        int
	NextPageList       []string // 下一页文本标识
	Reverse            bool
	FirstPageClient    sitemap.ClientType
	NextPageActionType ActionType
	HttpIsProxy        bool
	RedisCfg           *http2.RedisConfig
}

const (
	AnalysisNextPage ActionType = iota
	ClickNextPage
)

type ActionType int
