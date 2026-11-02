// home 包提供首页演示接口，运行时通过 GORM 泛型 API 读取 MySQL。
package home

// Featured 是首页顶部的特色入口。Key 和 Edition 对应前端的多语言文案键，
// Icon 对应前端已经支持的图标名，后端不拼接页面文案或绘制图标。
type Featured struct {
	Key     string `json:"key"`
	Icon    string `json:"icon"`
	Edition string `json:"edition"`
}

// Tool 是工具菜单入口。这里只返回展示配置，不代表对应业务功能已经实现。
type Tool struct {
	Key  string `json:"key"`
	Icon string `json:"icon"`
}

// Banner 与现有前端一致，用 Key 选择文案，用 Tone 选择轮播卡片样式。
// 当前轮播由前端绘制，不依赖远程图片。
type Banner struct {
	Key  string `json:"key"`
	Tone string `json:"tone"`
}

// Overview 是首页首屏聚合数据；资讯列表单独分页请求，避免首屏携带所有列表。
type Overview struct {
	Featured       []Featured `json:"featured"`
	Tools          []Tool     `json:"tools"`
	Banners        []Banner   `json:"banners"`
	FlashHeadlines []string   `json:"flashHeadlines"`
}

// NewsItem 对齐前端 NewsItem。omitempty 表示图片为空时不输出 imageUrl 字段。
type NewsItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// NewsQuery 从 GET 请求的 query 中读取参数，Gin 根据 form 和 binding 标签解析并校验。
// 缺省栏目为 discover，页码从 1 开始，每页默认 20 条、最多 100 条。
// 例如：/api/v1/home/news?tab=news&page=1&size=3。
type NewsQuery struct {
	Tab  string `form:"tab,default=discover" binding:"oneof=discover news flash watchlist journey"`
	Page int    `form:"page,default=1" binding:"min=1"`
	Size int    `form:"size,default=20" binding:"min=1,max=100"`
}

// NewsPage 对齐前端分页类型；外层响应还有一层 data，这里的 data 是当前页列表。
// 例如解包响应后，前端用 result.data 取列表，用 result.total 判断是否已加载完毕。
type NewsPage struct {
	Data  []NewsItem `json:"data"`
	Size  int        `json:"size"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
}
