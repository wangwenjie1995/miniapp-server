package home

import "context"

// loadOverview 顺序查询各部分；任一查询失败就返回错误，不用演示数据掩盖数据库异常。
func loadOverview(ctx context.Context, repo Repository) (Overview, error) {
	featured, err := repo.Featured(ctx)
	if err != nil {
		return Overview{}, err
	}
	tools, err := repo.Tools(ctx)
	if err != nil {
		return Overview{}, err
	}
	banners, err := repo.Banners(ctx)
	if err != nil {
		return Overview{}, err
	}
	flash, err := repo.Flash(ctx)
	if err != nil {
		return Overview{}, err
	}
	return Overview{Featured: featured, Tools: tools, Banners: banners, FlashHeadlines: flash}, nil
}
