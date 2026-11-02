package home

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository 只描述首页需要的数据操作，Service 不必知道 MySQL 如何查询。
// 测试可以传入实现同样方法的对象，正式运行始终使用 MySQLRepository。
type Repository interface {
	Featured(context.Context) ([]Featured, error)
	Tools(context.Context) ([]Tool, error)
	Banners(context.Context) ([]Banner, error)
	Flash(context.Context) ([]string, error)
	News(context.Context, NewsQuery) (NewsPage, error)
}

type MySQLRepository struct{ db *gorm.DB }

// NewRepository 接收启动时创建的连接；这就是把数据库依赖注入首页模块。
func NewRepository(db *gorm.DB) *MySQLRepository { return &MySQLRepository{db: db} }

func (r *MySQLRepository) Featured(ctx context.Context) ([]Featured, error) {
	// [Entry] 指定结果类型；Find 直接返回 []Entry 和 error。
	rows, err := gorm.G[Entry](r.db).Where("kind = ? AND enabled = ?", "featured", true).
		Order("sort_order ASC, id ASC").Find(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Featured, 0, len(rows))
	for _, row := range rows {
		result = append(result, Featured{Key: row.EntryKey, Icon: row.Icon, Edition: row.Edition})
	}
	return result, nil
}

func (r *MySQLRepository) Tools(ctx context.Context) ([]Tool, error) {
	rows, err := gorm.G[Entry](r.db).Where("kind = ? AND enabled = ? AND entry_key <> ?", "tool", true, "ai").
		Order("sort_order ASC, id ASC").Find(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Tool, 0, len(rows))
	for _, row := range rows {
		result = append(result, Tool{Key: row.EntryKey, Icon: row.Icon})
	}
	return result, nil
}

func (r *MySQLRepository) Banners(ctx context.Context) ([]Banner, error) {
	rows, err := gorm.G[BannerRecord](r.db).Where("enabled = ?", true).
		Order("sort_order ASC, id ASC").Find(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Banner, 0, len(rows))
	for _, row := range rows {
		result = append(result, Banner{Key: row.BannerKey, Tone: row.Tone})
	}
	return result, nil
}

func (r *MySQLRepository) Flash(ctx context.Context) ([]string, error) {
	rows, err := gorm.G[FlashRecord](r.db).Where("enabled = ? AND published_at <= ?", true, time.Now().UTC()).
		Order("sort_order ASC, id ASC").Find(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Content)
	}
	return result, nil
}

// News 先查真实总数，再用 LIMIT/OFFSET 查询当前页，避免把所有资讯加载进内存。
func (r *MySQLRepository) News(ctx context.Context, query NewsQuery) (NewsPage, error) {
	now := time.Now().UTC()
	base := gorm.G[Article](r.db).Where("category = ? AND published = ? AND published_at <= ?", query.Tab, true, now)
	total, err := base.Count(ctx, "*")
	if err != nil {
		return NewsPage{}, err
	}
	result := NewsPage{Data: make([]NewsItem, 0), Size: query.Size, Total: total, Page: query.Page}
	// 先判断是否超出范围，避免极大 page 乘以 size 导致整数溢出。
	if total == 0 || int64(query.Page) > (total-1)/int64(query.Size)+1 {
		return result, nil
	}
	rows, err := base.Order("published_at DESC, id ASC").
		Offset((query.Page - 1) * query.Size).Limit(query.Size).Find(ctx)
	if err != nil {
		return NewsPage{}, err
	}
	for _, row := range rows {
		result.Data = append(result.Data, NewsItem{ID: row.ID, Title: row.Title, ImageURL: row.ImageURL})
	}
	return result, nil
}
