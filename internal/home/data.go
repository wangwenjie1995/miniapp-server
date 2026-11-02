package home

// 本文件集中存放演示内容，来源于现有前端的静态配置和模拟标题。
// 这些标题不是实时行情或真实资讯流；仅供显式的初始化命令导入 MySQL，运行时接口不会读取本文件。
// 每次调用返回新切片，不在服务端维护请求计数、用户状态或可变的全局列表。

// seedFeatured 返回特色入口，保持与前端多语言键和图标名称一致。
func seedFeatured() []Featured {
	return []Featured{
		{Key: "quant", Icon: "chart-bar", Edition: "dragon"},
		{Key: "tech", Icon: "chart", Edition: "dragon"},
		{Key: "profit", Icon: "star-filled", Edition: "dragon"},
		{Key: "value", Icon: "chart-pie", Edition: "mobile"},
		{Key: "trend", Icon: "star", Edition: "mobile"},
	}
}

// seedTools 返回工具入口；按当前需求暂时移除 ai 项。
// 前端展示时仍可按每 10 个入口分成一页。
func seedTools() []Tool {
	return []Tool{
		{Key: "trends", Icon: "chart"},
		{Key: "market", Icon: "chart-bar"},
		{Key: "topics", Icon: "flag"},
		{Key: "protection", Icon: "secured"},
		{Key: "watchlist", Icon: "add-rectangle"},
		{Key: "smart", Icon: "chart-pie"},
		{Key: "diagnosis", Icon: "detection"},
		{Key: "indicators", Icon: "search"},
		{Key: "school", Icon: "books"},
		{Key: "quotes", Icon: "chart"},
		{Key: "sectors", Icon: "view-module"},
		{Key: "capital", Icon: "money-circle"},
		{Key: "calendar", Icon: "calendar"},
		{Key: "reports", Icon: "file"},
		{Key: "review", Icon: "history"},
		{Key: "alerts", Icon: "notification"},
		{Key: "portfolio", Icon: "wallet"},
		{Key: "learning", Icon: "books"},
		{Key: "support", Icon: "service"},
	}
}

// seedBanners 返回轮播配置，具体标题和副标题由前端翻译文件决定。
func seedBanners() []Banner {
	return []Banner{
		{Key: "security", Tone: "blue"},
		{Key: "insight", Tone: "purple"},
	}
}

// seedFlash 返回首页滚动快讯的固定演示文案。
func seedFlash() []string {
	return []string{
		"市场快讯：主要指数窄幅震荡，成交量较前一交易日有所变化",
		"板块快讯：科技方向活跃，资金关注产业链后续表现",
		"公司快讯：多家公司披露最新公告，投资者可留意业绩信息",
		"财经快讯：本周重要经济数据将陆续公布，关注市场反应",
	}
}

// seedNewsTitles 返回指定栏目的演示标题。栏目合法性已在 HTTP 参数绑定时校验。
// watchlist 当前也是公共演示列表，不代表某个用户的真实自选记录。
func seedNewsTitles(tab string) []string {
	titles := map[string][]string{
		"discover": {
			"先进封装与高端装备方向受到市场关注",
			"算力需求增长，AI 基础设施持续升温",
			"商业航天产业链迎来新的观察窗口",
			"新能源板块出现结构性轮动",
			"今日热门主题与资金关注方向一览",
			"高股息策略仍受长期资金关注",
			"从成交量变化观察市场交易情绪",
			"本周值得留意的公司公告",
		},
		"news": {
			"主要指数窄幅震荡，市场交投保持活跃",
			"多家公司披露最新季度经营数据",
			"产业政策动态引发市场关注",
			"海外市场收盘表现及主要资产走势",
			"财报季临近，上市公司预告陆续公布",
			"机构发布最新行业研究观点",
			"重要经济数据公布，市场关注后续影响",
			"今日板块涨跌情况回顾",
		},
		"flash": {
			"开盘：主要指数走势分化",
			"快讯：科技板块交投活跃",
			"快讯：多家公司披露公告",
			"午间：市场成交情况更新",
			"快讯：热门板块出现轮动",
			"快讯：资金流向数据更新",
			"快讯：主要指数波动收窄",
			"收盘：今日市场表现汇总",
		},
		"watchlist": {
			"自选观察：关注公司最新公告",
			"自选观察：重点标的今日走势",
			"自选观察：盘中成交量变化",
			"自选观察：相关行业市场动态",
			"自选观察：近期财报披露日程",
			"自选观察：机构观点更新",
			"自选观察：价格区间与波动情况",
			"自选观察：本周重要事件提醒",
		},
		"journey": {
			"投资入门：如何建立自己的观察清单",
			"市场基础：读懂指数与板块",
			"风险管理：仓位与波动的关系",
			"阅读财报：先关注哪些指标",
			"长期投资：认识复利与时间",
			"行业研究：从产业链开始",
			"交易记录：复盘的重要性",
			"市场新知：常见术语解读",
		},
	}
	return titles[tab]
}
