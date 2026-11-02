package home

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Seed 仅由显式的 dbinit -seed 命令执行，正常启动不会覆盖后台改过的内容。
// 使用稳定主键/唯一键和 DoNothing，重复初始化不会重复插入或更新原记录。
// 所有数据插入处于同一事务，一部分失败时全部回滚。
func Seed(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		entries := make([]Entry, 0)
		for i, item := range seedFeatured() {
			entries = append(entries, Entry{Kind: "featured", EntryKey: item.Key, Icon: item.Icon, Edition: item.Edition, SortOrder: i, Enabled: true})
		}
		for i, item := range seedTools() {
			entries = append(entries, Entry{Kind: "tool", EntryKey: item.Key, Icon: item.Icon, SortOrder: i, Enabled: true})
		}
		if err := gorm.G[Entry](tx, clause.OnConflict{DoNothing: true}).CreateInBatches(ctx, &entries, 100); err != nil {
			return err
		}
		banners := make([]BannerRecord, 0)
		for i, item := range seedBanners() {
			banners = append(banners, BannerRecord{BannerKey: item.Key, Tone: item.Tone, SortOrder: i, Enabled: true})
		}
		if err := gorm.G[BannerRecord](tx, clause.OnConflict{DoNothing: true}).CreateInBatches(ctx, &banners, 100); err != nil {
			return err
		}
		// 固定发布时间，避免每次导入后排序发生变化。
		publishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		flashes := make([]FlashRecord, 0)
		for i, content := range seedFlash() {
			flashes = append(flashes, FlashRecord{ID: uint64(i + 1), Content: content, SortOrder: i, Enabled: true, PublishedAt: publishedAt})
		}
		if err := gorm.G[FlashRecord](tx, clause.OnConflict{DoNothing: true}).CreateInBatches(ctx, &flashes, 100); err != nil {
			return err
		}
		articles := make([]Article, 0)
		for _, tab := range []string{"discover", "news", "flash", "watchlist", "journey"} {
			for i, title := range seedNewsTitles(tab) {
				articles = append(articles, Article{ID: tab + "-" + strconv.Itoa(i), Category: tab, Title: title, Published: true, PublishedAt: publishedAt})
			}
		}
		return gorm.G[Article](tx, clause.OnConflict{DoNothing: true}).CreateInBatches(ctx, &articles, 100)
	})
}
