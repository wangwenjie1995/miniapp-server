# 用户模块

按 handler、service、repository 划分实现，通过 MySQL 存储用户。

- `POST /api/v1/user/register`：注册用户。
- `POST /api/v1/user/login`：校验用户名和密码，成功返回用户 ID 和用户名；尚未签发 token 或建立会话。

请求 JSON 包含 `username`（4 至 24 个字符）和 `password`（8 至 72 个字符），均为必填。

业务结果统一返回 HTTP 200 和 `code/message/data`：`0` 成功、`4000` 参数错误、`5000` 用户名或密码错误、`5001` 用户名已存在。失败时 `data` 为 `null`。数据库异常返回 HTTP 500 和纯文本提示，具体原因只记录日志。

密码当前沿用 MD5 存储，正式上线前需要迁移到适合密码存储的哈希方案。
