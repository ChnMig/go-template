# 脚手架迁移：统一中间件与移除内置 TLS

## Goal

将 HTTP 服务脚手架迁移到更统一、可维护的默认形态：统一 panic 响应和 access log，补齐分页 helper，收敛 Gin context key 常量，并移除内置 TLS/ACME 能力，明确 TLS 由 Caddy、Nginx、Ingress 等反向代理负责。

## What I Already Know

- 用户已确认本次迁移范围，当前主会话确认 git status 干净。
- `api/router.go` 当前使用 `gin.Default()`，会自动启用 Gin 自带 logger/recovery，不符合统一响应要求。
- 全局中间件当前包含 rate limit、`SecurityHeaders`、`TraceID`、`BodySizeLimit`、`CorsDomainHandler`。
- CORS 继续使用 `api/middleware/cross-domain.go` 的 `CorsDomainHandler`，保持默认纯放开，不恢复配置化的 `CorssDomainHandler`。
- `api/response` 统一使用 HTTP 200 包装业务状态码，panic recovery 也必须返回 `INTERNAL` 且保持 HTTP 200。
- `utils/log` 已提供 Gin 独立 logger 和请求上下文 logger，`TraceID` 中间件当前会写入 `trace_id` 和 `logger`。
- `config.yaml.example`、`.agentdocs/index.md` 与 `.trellis/spec/backend/*` 中仍有 TLS/ACME 说明，需要同步删除或改写。

## Requirements

- 新增统一 panic recovery 中间件：
  - panic 时使用 `api/response` 返回 `INTERNAL`。
  - 保持 HTTP 200 响应策略。
  - 日志包含 `trace_id`、method、path、client_ip 等请求上下文字段。
  - 避免使用 `gin.Default()` 自带 recovery。
- 新增结构化 access log 中间件：
  - 记录 method、path、raw_query、status、latency、client_ip、user_agent、trace_id、error 等字段。
  - 成功请求不记录请求 body 或大量参数。
  - 使用 `utils/log` 的 Gin logger 或上下文 logger，符合现有日志规范。
- 增强分页 helper：
  - 保留 `GetPage`、`GetPageSize` 兼容现有调用。
  - 新增 `PageQuery`、`ParsePageQuery`，包含 `Page`、`PageSize`、`Offset`、`Limit`、`Disabled`/`IsDisabled`。
  - 支持 `page=-1` 或 `page_size=-1` 取消分页。
  - 非法值回落默认值。
  - 新增单元测试。
- 统一 Gin context key 常量：
  - `trace_id`、`logger`、`jwtData`、绑定参数等 key 不再散落硬编码。
  - 常量位置不能造成 import cycle。
  - 更新引用和测试。
- 移除内置 TLS/ACME：
  - 删除 `main.go` 对 `utils/acme` 和 `utils/tlsfile` 的导入、Setup 和关闭逻辑。
  - 移除 `config` 中 `server.enable_acme`、`server.acme_domain`、`server.acme_cache_dir`、`server.enable_tls`、`server.tls_cert_file`、`server.tls_key_file` 字段、默认值、`applyConfig` 映射、`CheckConfig` 冲突校验和相关测试。
  - 删除 `utils/acme`、`utils/tlsfile` 相关包和测试。
  - 更新 `config.yaml.example`、`README`、`.agentdocs/index.md`、`.trellis/spec/backend/*` 中的 TLS/ACME 说明。
- 复杂任务文档：
  - 初始化或更新 `.agentdocs/workflow/260428-bootstrap-guidelines.md`。
  - 在 `.agentdocs/index.md` 中登记。
  - 完成后同步 TODO 状态。

## Acceptance Criteria

- [x] `api/router.go` 不再使用 `gin.Default()`，而是显式挂载统一 recovery 和 access log。
- [x] panic 响应为 `response.INTERNAL` 的统一响应体，HTTP 状态码仍为 200。
- [x] access log 字段齐全，成功路径不记录 body。
- [x] 分页 helper 新旧 API 均可用，取消分页和非法值回落行为有单元测试覆盖。
- [x] Gin context key 常量集中定义，核心引用不再使用散落字符串。
- [x] TLS/ACME 配置、实现、测试和文档说明均已移除或改写。
- [x] `make fmt` 和 `GOFLAGS=-mod=readonly make verify` 通过。

## Definition of Done

- Go 代码通过 `gofmt`。
- 单元测试覆盖新增分页、recovery、access log 或 context key 相关关键行为。
- 文档只沉淀本任务决策和长期约束，不写无使用场景的工作汇报。
- 不回滚他人改动。

## Out of Scope

- 不恢复或配置化 CORS。
- 不引入新测试框架。
- 不实现反向代理、证书签发或部署脚本。
- 不改变统一响应的 HTTP 200 策略。

## Technical Notes

- 相关规范：`.trellis/spec/backend/index.md`、`directory-structure.md`、`quality-guidelines.md`、`logging-guidelines.md`、`error-handling.md`、`.trellis/spec/guides/code-reuse-thinking-guide.md`。
- 当前任务不需要外部技术调研，采用项目已有 Gin、Zap、Viper 与标准库测试模式。
