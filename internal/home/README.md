# 首页接口

当前接口全部为公开 GET 请求，通过 GORM 泛型 API 查询 MySQL，不要求登录。相邻 `uni-preset-vue-vite-ts` 前端已通过统一请求层接入聚合和资讯分页接口。

## 接口列表

| 方法 | 路径 | 内容 |
| --- | --- | --- |
| GET | `/api/v1/home` | 首屏聚合数据：featured、tools、banners、flashHeadlines |
| GET | `/api/v1/home/featured` | 顶部特色入口数组 |
| GET | `/api/v1/home/tools` | 工具入口数组，暂不包含 ai |
| GET | `/api/v1/home/banners` | 轮播配置数组 |
| GET | `/api/v1/home/flash` | 滚动快讯字符串数组 |
| GET | `/api/v1/home/news` | 按栏目分页的资讯列表 |

首次可通过 `go run ./cmd/dbinit -seed` 将原有演示内容导入数据库；之后以数据库内容为准，不会在启动时覆盖修改。初始化内容不是实时行情。`watchlist` 栏目也是公共演示内容，不是用户个人自选列表。当前快讯和资讯标题为中文，不按 X-Language 切换。

## 响应格式

成功响应与现有前端 `src/request` 的解包规则一致：

```json
{
  "code": 0,
  "message": "success",
  "data": []
}
```

上面的 `data` 会按接口变为对象或数组。首屏聚合对象包含：

- `featured`：`{key, icon, edition}[]`。
- `tools`：`{key, icon}[]`。
- `banners`：`{key, tone}[]`，由前端根据 key 读取翻译文案，根据 tone 选择样式，不是图片 URL。
- `flashHeadlines`：`string[]`。

特色入口和工具的 key、图标名称来自现有前端。资讯列表中的 `imageUrl` 为可选字段，当前不返回图片。

## 资讯分页

```sh
curl 'http://localhost:8080/api/v1/home/news?tab=news&page=1&size=3'
```

| 参数 | 默认值 | 规则 |
| --- | --- | --- |
| tab | discover | discover、news、flash、watchlist、journey |
| page | 1 | 正整数，从 1 开始 |
| size | 20 | 1 至 100 的整数 |

省略参数或传入空字符串时，Gin 使用默认值。其他非法参数返回 HTTP 400，响应为 `{"code":400,"message":"参数错误：…","data":null}`。

成功时外层 `data` 是分页对象：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {"id": "news-0", "title": "主要指数窄幅震荡，市场交投保持活跃"}
    ],
    "size": 1,
    "total": 8,
    "page": 1
  }
}
```

上述示例对应 `tab=news&page=1&size=1`。首次导入每个栏目有 8 条，total 通过数据库 COUNT 查询；新增、取消发布会改变总数。列表只返回已发布且发布时间不晚于当前时刻的内容，按发布时间倒序、ID 升序排列。超出最后一页返回空数组 `[]` 并保留 total、page、size；重复请求内容和 ID 稳定。

前端统一请求层解包外层 data 后，得到的结构正好是现有 `NewsPage`：`{data, size, total, page}`。前端接入时可通过统一 request 发送 GET，使用 data 传递 tab、page、size，并设置 auth:false；baseURL 需要包含 `/api/v1`，或者在接口路径中包含该前缀，两者选择一种。

## 调用和阅读顺序

```text
app/router.go 注册 /api/v1
    → home/handler.go 注册首页路由、校验参数、返回响应
    → home/service.go 组装首屏内容
    → home/repository.go 使用 gorm.G[类型] 查询 MySQL
    → 数据库返回数据，repository 转成接口结构
    → response.Success 输出 code/message/data
```

`model.go` 定义接口数据结构和参数规则；`entity.go` 对应四张业务表；`data.go`、`seed.go` 仅用于显式初始化。数据库连接在启动时创建并传入 repository，再传入 service 和 handler，这就是依赖注入。

```sh
curl http://localhost:8080/api/v1/home
curl http://localhost:8080/api/v1/home/featured
curl http://localhost:8080/api/v1/home/tools
curl http://localhost:8080/api/v1/home/banners
curl http://localhost:8080/api/v1/home/flash
```

查询错误返回 HTTP 500、`code:500`、通用说明和 `data:null`，不会伪装成空列表。空表正常查询返回空数组。聚合接口任何一个部分失败则整个请求失败；健康检查仍只表示 HTTP 可响应。

路由测试位于 `internal/app/home_test.go`。普通测试覆盖参数和异常响应；设置 `MYSQL_INTEGRATION=1` 后使用独立测试库验证实际查询、分页、修改和重复初始化，详见项目 README。
