# 小程序 Go 后端

基于 Go 1.25 和 Gin 的独立 HTTP 服务骨架，后续扩展登录、用户和首页能力。当前首页接口通过 GORM 泛型 API 读取本地 MySQL；登录、用户资料和 AI 功能尚未实现。

## 环境与启动

需要 Go 1.25.0 或更高版本。首次运行前先完成下方“本地 MySQL 初始化”，再启动服务。

```sh
go mod download
go run ./cmd/server
```

## 本地 MySQL 初始化

使用你已经安装的 MySQL，无需 Docker。MySQL 服务需要启动，并提前创建业务数据库，例如在 MySQL 客户端执行：

```sql
CREATE DATABASE miniapp CHARACTER SET utf8mb4;
```

如使用已有数据库，直接填写它的名称，不要重复创建。复制 `.env.example` 为未被 Git 跟踪的 `.env.local`，填写本地账号密码：

```dotenv
HTTP_ADDR=:8080
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=miniapp
DB_USER=root
DB_PASSWORD=你的本地密码
```

密码包含 `#`、空格等字符时用引号包裹。不要提交或在聊天里发送真实密码。运行命令时应位于项目根目录。

```sh
# 显式执行首页建表迁移，并导入原有演示数据。
go run ./cmd/dbinit -seed

# 然后启动服务，或使用下方 Air 热重载。
go run ./cmd/server
```

`go run ./cmd/dbinit` 只建表，不导入内容。`-seed` 的重复导入会跳过已经存在的记录，不覆盖已修改的内容。原始演示数据集中在 `internal/home/data.go`，只有初始化命令使用，HTTP 接口不会读取它作为兜底。

只创建 `home_entries`、`home_banners`、`home_flash_items`、`articles` 四张首页业务表，以及一张 `home_schema_migrations` 迁移版本表。迁移按文件名执行并记录校验和；已执行的文件不能修改，后续变更新增 SQL 文件。当前迁移执行器面向普通 DDL，不支持存储过程等包含内部分号的语法。MySQL DDL 不整体回滚，部分失败后修复原因再重跑。

正常启动只连接数据库，不自动建表、不执行 AutoMigrate、不自动补种子数据。连接失败则退出；接口查询失败返回 HTTP 500 和通用错误说明，详细错误只记日志。数据库密码不会被打印。

连接池最多 10 个连接、最多保留 5 个空闲连接，连接最长使用 30 分钟，空闲 5 分钟后回收；连接和网络读写设置 5 秒超时，单次首页请求也最多查询 5 秒。服务退出时，在 HTTP 请求收尾后关闭连接池。生产运行建议使用单独的数据库账号，初始化账号需有建表权限，运行账号只需首页读权限。

数据库参数默认值为 `127.0.0.1:3306`、数据库 `miniapp`、用户 `root`，密码默认空。请按本机实际情况填写。

## 开发热重载

项目通过 Go tool 锁定 Air v1.64.0（兼容 Go 1.25），无需全局安装。开发时运行：

```sh
go tool air -c .air.toml
```

保存业务 `.go` 文件或修改 `go.mod`、`go.sum` 后，会自动重新编译并重启服务；修改测试文件不触发重启。首次运行需要下载工具依赖。热重载是重新启动进程，不会保留内存中的状态，也不会自动刷新浏览器。构建失败会停止旧服务并在终端输出错误，修好保存后重新构建。

按 Ctrl+C 退出。Air 会先向业务进程发送中断信号，允许最多 10 秒优雅关闭；超过 Air 配置的 12 秒等待后再强制终止。编译产物和错误日志存放在被忽略的 `tmp/` 目录。

同一端口只能启动一个服务，不要同时运行 `go run` 和 Air。如果提示 `address already in use`，先在原来的服务终端按 Ctrl+C；需要同时启动第二个实例时，可以换端口：

```sh
HTTP_ADDR=:8081 go tool air -c .air.toml
```

普通 `go run ./cmd/server` 仍然可用，但不会监听文件变化。以上开发热重载不用于生产部署。

默认监听 `:8080`。程序可读取项目根目录的 `.env.local`，已设置的进程环境变量优先；不自动加载 `.env` 或 `.env.example`。

```sh
HTTP_ADDR=127.0.0.1:8081 go run ./cmd/server
curl -i http://127.0.0.1:8081/healthz
```

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | host:port 格式，端口范围 1–65535；显式空值视为配置错误 |

健康检查为 `GET /healthz`，返回 HTTP 200、JSON `{"status":"ok"}`。监听所有网卡时应根据运行环境控制网络访问。

## 项目组织

- `cmd/server`：启动入口、信号监听和退出码。
- `internal/app`：Gin 路由、HTTP 服务和优雅关闭。
- `internal/config`：环境变量、本地私有配置读取及校验。
- `internal/database`：MySQL 连接池、GORM 初始化和启动连接检查。
- `migrations`、`cmd/dbinit`：首页版本化 SQL 迁移及显式初始化命令。
- `internal/middleware`：使用 `slog` 的 JSON 请求日志与异常恢复。
- `internal/response`：JSON 响应工具。
- `internal/home`：首页展示配置、快讯和分页资讯接口，通过 repository 查询 MySQL。
- `internal/user`：用户注册与登录凭据校验；`internal/auth`：认证模块职责预留。

Gin 使用 `gin.New()` 创建，由标准库 `http.Server` 托管。请求日志包含方法、路径、状态码和耗时，不记录请求头、请求体及查询参数。异常恢复记录错误及堆栈；响应尚未写出时返回 HTTP 500 和纯文本 `服务内部错误`。如果响应已经写出，则终止当前处理，无法再改写状态码。

HTTP 超时：请求头 5 秒、读取 10 秒、写入 15 秒、空闲连接 60 秒。收到 Ctrl+C 或 SIGTERM 后最多等待 10 秒完成在途请求；超时则关闭连接并以非零退出码退出。配置错误、监听失败或服务错误同样以非零退出码退出。

## 验证

```sh
go test ./...
go vet ./...
go build ./...
```

普通测试覆盖健康检查、异常恢复、结构化日志、环境配置、非法分页参数和数据库错误响应，无需真实数据库。

真实 MySQL 集成测试显式启用：

```sh
MYSQL_INTEGRATION=1 go test ./internal/app -v
```

集成测试读取 `.env.local`，使用同一账号创建名称为 `miniapp_test_时间戳` 的独立临时库并在完成后删除。账号需要创建、删除测试库权限。测试不会清空业务库，覆盖六个首页接口、五栏目分页、重复迁移与导入、数据修改保留、发布过滤和取消查询。未启用时相关集成用例会显示 SKIP，不代表已完成真实 MySQL 验证。

构建独立可执行文件：

```sh
go build -o bin/server ./cmd/server
HTTP_ADDR=127.0.0.1:8081 ./bin/server
```

在另一终端请求健康检查，再对服务进程发送 SIGTERM（或在前台按 Ctrl+C），日志应包含“HTTP 服务已关闭”，退出码为 0。直接验证信号时使用构建后的可执行文件，避免将信号发给 `go run` 的父进程。

```sh
HTTP_ADDR=invalid ./bin/server
```

上述无效地址应输出配置错误并以退出码 1 退出；地址被占用时应输出监听错误并以退出码 1 退出。

## 首页接口

首页接口位于 `/api/v1/home`，支持首屏聚合、特色入口、工具、轮播、快讯和分页资讯。全部为公开只读接口，运行时读取 MySQL；相邻 `uni-preset-vue-vite-ts` 前端已接入聚合和资讯分页接口。

```sh
curl http://localhost:8080/api/v1/home
curl 'http://localhost:8080/api/v1/home/news?tab=news&page=1&size=3'
```

参数、返回字段和示例见 [首页接口说明](internal/home/README.md)。首页成功响应使用 `{code:0,message:"success",data:...}`；健康检查返回独立的状态信息。

## 后续扩展

用户模块已支持注册与登录凭据校验，登录成功返回用户 ID 和用户名，尚未签发 token 或建立会话。后续再实现登录凭证及用户资料功能。

## 业务响应约定

业务成功与业务失败均返回 HTTP 200，前端必须检查 JSON 的 `code`：`0` 成功、`4000` 参数错误、`5000` 用户名或密码错误、`5001` 用户名已存在。JSON 包含 `code/message/data`，失败时 `data` 为 `null`。

数据库异常和未处理的 panic 返回 HTTP 500，响应体为纯文本通用提示，不包含 JSON 或内部错误详情。前端先检查 HTTP 状态，HTTP 200 再解析业务 JSON。健康检查 `/healthz` 维持独立的状态响应。
