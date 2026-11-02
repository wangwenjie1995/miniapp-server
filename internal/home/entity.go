package home

import "time"

// 这些 Entity 结构体对应数据库表；model.go 中的结构体对应前端响应。
// 两者分开，数据库内部字段不会不小心暴露给前端。
// gorm 标签描述字段映射，json 标签描述接口 JSON，两者用途不同。
type Entry struct {
	ID        uint64 `gorm:"primaryKey"`
	Kind      string
	EntryKey  string
	Icon      string
	Edition   string
	SortOrder int
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Entry) TableName() string { return "home_entries" }

type BannerRecord struct {
	ID        uint64 `gorm:"primaryKey"`
	BannerKey string
	Tone      string
	SortOrder int
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (BannerRecord) TableName() string { return "home_banners" }

type FlashRecord struct {
	ID          uint64 `gorm:"primaryKey"`
	Content     string
	SortOrder   int
	Enabled     bool
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (FlashRecord) TableName() string { return "home_flash_items" }

// Article.ID 保留原接口的字符串 ID，例如 news-0，避免前端 ID 随迁移变化。
// 发布时间统一存 UTC，发布过滤避免提前展示未来发布的内容。
type Article struct {
	ID          string `gorm:"primaryKey"`
	Category    string
	Title       string
	ImageURL    string
	Published   bool
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Article) TableName() string { return "articles" }
