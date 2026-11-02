package home

import "context"

// Service 负责业务组装，不直接读取环境变量，也不在函数中创建数据库连接。
// repo 是外部传入的数据操作对象；正式运行传入 MySQLRepository。
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Overview 顺序查询各部分；任一查询失败就返回错误，不用演示数据掩盖数据库异常。
func (s *Service) Overview(ctx context.Context) (Overview, error) {
	featured, err := s.repo.Featured(ctx)
	if err != nil {
		return Overview{}, err
	}
	tools, err := s.repo.Tools(ctx)
	if err != nil {
		return Overview{}, err
	}
	banners, err := s.repo.Banners(ctx)
	if err != nil {
		return Overview{}, err
	}
	flash, err := s.repo.Flash(ctx)
	if err != nil {
		return Overview{}, err
	}
	return Overview{Featured: featured, Tools: tools, Banners: banners, FlashHeadlines: flash}, nil
}
func (s *Service) Featured(ctx context.Context) ([]Featured, error) { return s.repo.Featured(ctx) }
func (s *Service) Tools(ctx context.Context) ([]Tool, error)        { return s.repo.Tools(ctx) }
func (s *Service) Banners(ctx context.Context) ([]Banner, error)    { return s.repo.Banners(ctx) }
func (s *Service) Flash(ctx context.Context) ([]string, error)      { return s.repo.Flash(ctx) }
func (s *Service) News(ctx context.Context, query NewsQuery) (NewsPage, error) {
	return s.repo.News(ctx, query)
}
