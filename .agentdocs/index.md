# 项目文档索引

## 后端文档
`backend/configuration.md` - 后端配置管理架构与使用说明，修改配置相关代码时必读
`backend/middleware.md` - 中间件架构与使用文档，开发API时必读
`backend/cli.md` - 命令行参数文档，启动和部署时必读
`backend/logging.md` - 日志系统架构与使用规范，记录日志时必读
`backend/api.md` - 后端 API 路由分层架构与使用规范，开发/调整路由时必读
`backend/build.md` - 构建与打包规范，发布与交付时必读

## 全局重要记忆
- 项目使用 YAML 配置文件管理配置项，配置文件位于 `http-services/config.yaml`
- 配置文件已加入 `.gitignore`，使用 `config.yaml.example` 作为模板
- 所有配置支持通过环境变量覆盖，格式为 `HTTP_SERVICES_<SECTION>_<KEY>`
- 所有配置加载在程序启动时完成，并在 logger 初始化后进行校验
- 项目基于 art-design-pro-edge-go-server 框架，定期同步基础组件更新
- 使用标准JWT认证，简洁高效
- 所有 API 响应自动包含 `trace_id` 字段，方便问题追踪
- 使用 `log.FromContext(c)` 获取带上下文的 logger，自动包含 trace_id、method、path 等信息
- 错误处理使用 `response.ReturnError()` 返回统一格式，日志使用 `log.FromContext(c)` 记录详细信息
- 路由规范：`api/app` 下每一层级目录必须包含 `router.go`，通过 `RegisterRoutes(*gin.RouterGroup)` 逐级嵌套注册（app → v1 → open/private → module），顶层仅挂载 `/api`
- 构建与发版统一使用 Makefile：`make verify` 完整校验，`make build` 本地构建，`make build CROSS=1` 或 `make build-cross` 跨平台打包到 `dist/`
- 默认跨平台矩阵：`linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64`，可通过 `PLATFORMS` 覆盖；随包自动包含 `README.md` 与 `config.yaml.example`
