## 产品文档

目前暂无专门的产品文档，HTTP 服务主要作为通用 HTTP API 模板使用。

## 前端文档

当前仓库不包含前端代码，无前端文档。

## 后端文档

`backend/architecture.md` - 预留：后端整体架构与技术约束（待根据实际需要补充）。

## 当前任务文档

暂无正在跟踪的任务文档。

## 已完成任务文档

- `workflow/done/260323-sync-gin-log-level.md` - 已将独立 Gin 日志级别配置与热重载 logger 刷新能力同步到脚手架仓库。
- `workflow/done/260428-bootstrap-guidelines.md` - 脚手架迁移：统一中间件、分页 helper、context key，并移除内置 TLS/ACME。

## 全局重要记忆

- 配置与运行：
  - Go 版本由 `go.mod` 管理，后端通过 `make build` / `make run` / `make dev` 进行构建与运行。
  - 配置使用 Viper 管理，配置文件为 `config.yaml`，支持环境变量覆盖与热重载。
  - 配置文件已加入 `.gitignore`，使用 `config.yaml.example` 作为模板。
  - 运行模式判断统一走 `utils/runmodel`，命令行 `--dev` 优先于环境变量 `model`。
  - 路径工具统一放在 `utils/pathtool`，避免继续使用短横线目录命名。
  - 服务进程只监听 HTTP；HTTPS/TLS 终止由 Caddy、Nginx、Ingress、Traefik 或云负载均衡等反向代理负责。
  - 不再新增 `server.enable_acme`、`server.enable_tls` 等服务内 TLS 配置项；如业务项目确需内置 TLS，应作为项目级定制重新设计。
- API 与中间件：
  - 路由规范：`api/app` 下每一层级目录必须包含 `router.go`，通过 `RegisterRoutes(*gin.RouterGroup)` 逐级嵌套注册（app -> v1 -> open/private -> module），顶层仅挂载 `/api`。
  - 全局中间件顺序要求：`TraceID -> AccessLog -> Recovery`，确保 panic 先被 recovery 写成统一响应，再由 access log 记录最终状态。
  - CORS 中间件使用 `CorsDomainHandler`，脚手架默认纯放开跨域；生产项目如需收紧，应在项目内替换中间件策略。
  - 所有 API 响应自动包含 `trace_id` 字段，方便问题追踪。
  - Gin context key 统一放在 `utils/contextkey`，避免散落硬编码。
  - 分页参数优先使用 `middleware.ParsePageQuery(c)`，旧调用 `GetPage` / `GetPageSize` 保留兼容。
- 日志与错误：
  - 使用 `log.FromContext(c)` 获取带上下文的 logger，自动包含 trace_id、method、path 等信息。
  - 错误处理使用 `response.ReturnError()` 返回统一格式，日志使用 `log.FromContext(c)` 或 `log.WithRequest(c)` 记录必要上下文。
  - 业务日志与 Gin 日志分离，Gin 日志级别由 `log.gin_level` 控制；为空时跟随 `log.level`。
- 测试与验证（Go 后端）：
  - 代码变更后应通过 `make fmt` 进行格式化。
  - 使用 `make test` 运行单元测试并检查覆盖率。
  - 如需静态检查，可执行 `make lint`。
  - 发布前或较大变更优先执行 `GOFLAGS=-mod=readonly make verify`。
  - 进程生命周期类集成测试应支持 `testing.Short()` 跳过；等待 pid 文件、端口或外部进程就绪时要预留足够时间，并同步检测子进程提前退出以输出诊断信息。
- 构建与发布：
  - 构建与发版统一使用 Makefile：`make verify` 完整校验，`make build` 本地构建，`make build CROSS=1` 或 `make build-cross` 跨平台打包到 `dist/`。
  - 默认跨平台矩阵：`linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64`，可通过 `PLATFORMS` 覆盖；随包自动包含 `README.md` 与 `config.yaml.example`。
