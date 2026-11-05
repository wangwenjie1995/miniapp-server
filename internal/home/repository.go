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
	// Entry是表的名称吗？
	// Entry 不是表名，而是一个 Go 结构体类型，通常与数据库中的表对应。GORM 使用结构体类型来映射数据库表的记录。
	// 在这里，gorm.G[Entry](r.db) 表示使用 GORM 的泛型方法来操作与 Entry 结构体对应的数据库表。
	// Find 方法会查询数据库并返回符合条件的记录，结果会被映射为 []Entry 类型的切片。
	// 怎么把Entry映射到数据库表的？
	// GORM 使用结构体的字段和标签来映射数据库表的列。每个结构体字段通常对应数据库表中的一列，字段名可以通过标签指定与列名的映射关系。
	// 在 Entry 结构体中，TableName 方法返回了对应的数据库表名 "home_entries"。GORM 会根据这个方法来确定查询的表。
	rows, err := gorm.G[Entry](r.db).Where("kind = ? AND enabled = ?", "featured", true).
		Order("sort_order ASC, id ASC").Find(ctx)
	
	if err != nil {
		return nil, err
	}
	result := make([]Featured, 0, len(rows))
	for _, row := range rows {
		// append 函数会在切片末尾添加元素，并返回新的切片。这里将每个查询结果映射为 Featured 结构体。
		// append 函数的第一个参数是目标切片，后续参数是要添加的元素。它会返回一个新的切片，如果原切片容量不足以容纳新元素，append 会分配一个新的底层数组。
		// 那么原来的result切片会被覆盖吗？
		// 不会被覆盖。append 函数返回一个新的切片，它可能指向原来的底层数组，也可能指向一个新的底层数组。如果原来的切片容量足够，append 会在原来的数组上添加元素；否则，它会分配一个新的数组并返回新的切片。
		// 因此，result 变量会被更新为 append 返回的新切片，但原来的 result 切片仍然存在，只是它的内容可能没有改变。
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
