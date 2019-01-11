package entity

import (
	"github.com/jinzhu/gorm"
	"time"
)

type GovTask struct {
	gorm.Model
	// 网站主页
	URL string `gorm:"size:48;unique"`
	// 任务状态
	Status int `gorm:"index:task_status;default:0"`
}

// 爬取map时配置参数
type GovMapSetting struct {
	ID        uint `gorm:"primary_key;auto_increment:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	// 爬取深度
	Deep uint `gorm:"default:2"`
	// 目录如果不是以/结尾 是否添加
	AddSlash bool `gorm:"default:true"`
	// 爬取地图时用的client类型(有时网站全是需要js渲染时 选择chrome 0)
	ClientType uint `gorm:"default:1"`
	// 当client为http时选择是否使用代理
	HttpUseProxy bool `gorm:"default:false"`
	// goroutine 数
	Concurrent uint `gorm:"default:1"`
	// 额外匹配的host(有时需要采出本域名以外的域名)
	MustMatchHost string `gorm:"default:''"`
	// 一定有效的目录正则expr(爬取不精确时填写)
	MustRegExpr string `gorm:"default:''"`
	// 目录忽略关键字(map里面不需要内容!!!)
	IgnoreKeyword string `gorm:"default:''"`
	// base匹配的后缀(如:index.jsp匹配.jsp)
	MatcherSuffix string `gorm:"default:''"`
	// 当爬取的地图想要的目录没在里面 填写之
	MatcherRegBase string `gorm:"default:''"`
}

// 获取文章配置参数
type GovArtSetting struct {
	ID        uint `gorm:"primary_key;auto_increment:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	// 并发数
	Concurrent uint `gorm:"default:1"`
	// 翻页最大值
	MaxPageSize uint `gorm:"default:50"`
	// 下一页text值
	NextPage string `gorm:"default:'下一页'"`
	// ul ol table 之外的文章所在xpath
	ExtArtXpath string `gorm:"default:''"`
	// 翻页是否反向(page++变成page--)
	Reverse bool `gorm:"default:false"`
	// 获取第一页选择的client
	FirstPageClient uint `gorm:"default:1"`
	// 获取下一页行为(0 当有有效href分析下一页url; 1 无有效href'javascript:void(0)'时 选择点击下一页来获取下一页)
	NextPageActionType uint `gorm:"default:0"`
	IsSplitArt         bool `gorm:"default:false"`
	// 当client为http时选择是否使用代理
	HttpUseProxy bool `gorm:"default:false"`
}

// 获取文章字段配置参数
type GovArtUpdateSetting struct {
	ID        uint `gorm:"primary_key;auto_increment:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	// 更新内容选择使用代理?
	UseProxy bool `gorm:"default:false"`
	// 有些href也是通过js加载的
	ClientType uint `gorm:"default:1"`
	Concurrent uint `gorm:"default:10"`
}

// 网站地图
type GovMap struct {
	gorm.Model
	UID       uint   `gorm:"index:map_uid"`
	Crc32ID   uint32 `gorm:"index:map_crc32_id"`
	HDID      uint   `gorm:"column:hd_id"`
	HDTitle   string `gorm:"column:hd_title"` // 父标题
	Title     string
	URL       string
	SourceURL string
	Status    int `gorm:"index:map_status;default:0"`
}

// 文章
type GovArticle struct {
	gorm.Model
	UID       uint `gorm:"index:art_uid"`
	UUID      uint `gorm:"index:art_uuid"`
	URL       string
	SourceURL string
	Status    int    `gorm:"index:art_status;default:0"`
	Keyword1  string `gorm:"default:''"`
	Keyword2  string `gorm:"default:''"`
	Keyword3  string `gorm:"default:''"`
	Keyword4  string `gorm:"default:''"`
	Keyword5  string `gorm:"default:''"`
	Keyword6  string `gorm:"default:''"`
}

type GovArtRule struct {
	gorm.Model
	UUID uint `gorm:"index:art_rule_uuid"`
	// 文章类型(不要map!!!)
	Type int `gorm:"default:0"`
	// url 正则
	URLExpr string
	Example string
	// 该正则下的文章总数
	DataCount    int
	Keyword1Rule string `gorm:"default:''"`
	Keyword2Rule string `gorm:"default:''"`
	Keyword3Rule string `gorm:"default:''"`
	Keyword4Rule string `gorm:"default:''"`
	Keyword5Rule string `gorm:"default:''"`
	Keyword6Rule string `gorm:"default:''"`
}
