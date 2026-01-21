## 产品文档

目前暂无专门的产品文档，HTTP 服务主要作为通用 HTTP API 模板使用。

## 前端文档

当前仓库不包含前端代码，无前端文档。

## 后端文档

`backend/architecture.md` - 预留：后端整体架构与技术约束（待根据实际需要补充）。

## 当前任务文档

暂无正在跟踪的任务文档。本次改动为在 HTTP 服务中增加可选的 ACME 自动 TLS 能力，改动范围较集中，暂不单独建立任务文档。

## 全局重要记忆

- 配置与运行：
  - Go 版本由 `go.mod` 管理，后端通过 `make build` / `make run` / `make dev` 进行构建与运行。
  - 配置使用 Viper 管理，配置文件为 `config.yaml`，支持环境变量覆盖与热重载。
- 测试与验证（Go 后端）：
  - 代码变更后应通过 `make fmt` 进行格式化。
  - 使用 `make test` 运行单元测试并检查覆盖率。
  - 如需静态检查，可执行 `make lint`。
- TLS / ACME 相关约定：
  - HTTP 服务内置可选的 ACME 自动 TLS 能力，通过 `server.enable_acme` 开关与 `server.acme_domain` 配置启用。
  - 启用 ACME 后，服务会在 `server.port`（推荐 443）上以 HTTPS 形式提供服务，并在 80 端口开启仅供 ACME 验证与 HTTP→HTTPS 跳转的辅助 HTTP 服务。
  - ACME 证书缓存目录由 `server.acme_cache_dir` 配置，允许使用相对路径，相对路径基于程序所在目录。
  - HTTP 服务同时支持基于本地证书文件的 TLS 模式，通过 `server.enable_tls`、`server.tls_cert_file`、`server.tls_key_file` 配置启用，证书与私钥文件发生变更时会自动热更新。
  - ACME 与本地证书文件 TLS 模式互斥：`server.enable_acme` 与 `server.enable_tls` 不能同时为 `true`，否则启动时会直接报错。
  - 若在前置网关（如 Nginx / Ingress / Caddy / Traefik）层统一管理证书，建议保持 `server.enable_acme` 与 `server.enable_tls` 均为关闭，由网关处理 TLS。
