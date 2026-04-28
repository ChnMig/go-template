# 脚手架迁移：统一中间件与移除内置 TLS

## 任务背景

本任务迁移 HTTP 服务脚手架默认能力：统一 Gin recovery 与 access log、增强分页 helper、统一 Gin context key，并移除内置 TLS/ACME。TLS 证书签发和 HTTPS 终止由 Caddy、Nginx、Ingress、Traefik 等反向代理负责，服务本身保持 HTTP API 模板职责。

## 决策

- 全局中间件顺序固定为 `TraceID -> AccessLog -> Recovery`，确保 panic 时 recovery 先写统一响应，access log 再记录最终状态。
- panic recovery 使用 `api/response` 的 `INTERNAL` 响应，延续项目 HTTP 200 包装业务状态码的策略。
- access log 只记录请求摘要字段，不记录 body 或大量业务参数。
- 分页统一使用 `middleware.ParsePageQuery(c)`，旧的 `GetPage` / `GetPageSize` 保留。
- Gin context key 统一放在 `utils/contextkey`。
- 脚手架不再内置 ACME 自动证书签发或本地证书文件 TLS 热更新。

## TODO

- [x] 新增统一 recovery 中间件。
- [x] 新增结构化 access log 中间件。
- [x] 增强分页 helper 并补测试。
- [x] 统一 Gin context key 常量。
- [x] 移除 TLS/ACME 配置、实现、测试与文档说明。
- [x] 更新 README、`.agentdocs` 与 Trellis 规范。
- [x] 执行 `make fmt`、`GOFLAGS=-mod=readonly make verify` 与 `git diff --check`。

## 验证

- `make fmt`：通过。
- `GOFLAGS=-mod=readonly make verify`：通过。
- `git diff --check`：通过。

## 后续约束

- 不恢复 `CorssDomainHandler` 旧拼写兼容入口。
- 不恢复 CORS 配置化；脚手架默认纯放开，业务项目需要收紧时自行替换。
- 不在脚手架服务进程内重新引入 TLS/ACME 能力；HTTPS/TLS 应放在反向代理或网关层。
