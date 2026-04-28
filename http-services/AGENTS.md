# PROJECT KNOWLEDGE BASE

**Generated:** 2026-04-29 Asia/Shanghai
**Commit:** 8c8b648
**Branch:** main

## OVERVIEW

Gin + Viper + Zap 的 HTTP API 服务模板。仓库是单 Go module：`http-services`，可执行入口在根目录 `main.go`，没有 `cmd/`、`internal/`、`pkg/` 拆分。

## STRUCTURE

```text
http-services/
├── main.go          # CLI、配置、运行模式、日志、HTTP server、优雅关闭
├── Makefile         # build/run/dev/test/fmt/lint/verify
├── api/             # Gin 初始化、分层路由、中间件、统一响应
├── config/          # Viper 加载、默认值、环境变量、热重载、安全校验
├── domain/health/   # 当前唯一领域模块
├── utils/           # 日志、JWT、context key、pidfile、ID 等基础设施
└── vendor/          # 第三方代码；不要写入项目约束
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| 程序入口/启动顺序 | `main.go` | `LoadConfig -> runmodel.Detect -> log.GetLogger/StartMonitor -> WatchConfig -> CheckConfig -> api.InitApi` |
| 构建/测试/发布 | `Makefile` | 所有本地命令以 Make target 为准 |
| HTTP 路由/中间件 | `api/` | 见 `api/AGENTS.md` |
| 配置默认值与热重载 | `config/load.go` | Viper、`HTTP_SERVICES_` env、`parseSize`、`WatchConfig` |
| 配置安全门禁 | `config/check.go` | JWT key 非空、非示例值、长度至少 32 |
| 全局配置变量 | `config/config.go` | `ListenPort`、超时、限流、日志、分页默认值 |
| 日志生命周期 | `utils/log/log.go` | dev/release 分流、Gin 日志独立文件、轮转、热重建 |
| JWT 签发/解析 | `utils/authentication/jwt.go` | 直接读取 `config.JWTKey/JWTExpiration` |
| Context key | `utils/contextkey/keys.go` | trace_id、logger、jwtData、bound params 统一 key |
| 领域健康状态 | `domain/health/` | 领域错误与状态，不依赖 Gin |
| 根级集成测试 | `pidfile_integration_test.go` | 构建真实二进制，`testing.Short()` 跳过 |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `main` | func | `main.go` | 单一可执行入口、服务生命周期编排 |
| `CLI` | var | `main.go` | Kong flags：`-d/--dev`、`-v/--version` |
| `api.InitApi` | func | `api/router.go` | Gin engine、全局中间件、`/api` 挂载 |
| `config.LoadConfig` | func | `config/load.go` | 默认值、配置文件、环境变量、全局变量应用 |
| `config.WatchConfig` | func | `config/load.go` | 配置热重载，回调里刷新 logger |
| `config.CheckConfig` | func | `config/check.go` | JWT 配置安全校验 |
| `log.SetLogger` | func | `utils/log/log.go` | 按运行模式和日志级别重建业务/Gin logger |
| `middleware.CleanupAllLimiters` | func | `api/middleware/rate-limit.go` | 服务退出时清理限流器 goroutine |

## CONVENTIONS

- 启动顺序不可调换：运行模式必须在初始化 logger 前设置；logger 初始化后再启动配置热重载。
- 配置文件名固定为 `config.yaml`，查找顺序：程序目录、工作目录、`/etc/http-services/`。
- 环境变量前缀固定 `HTTP_SERVICES_`；层级用 `_` 代替点，如 `HTTP_SERVICES_SERVER_PORT`。
- `server.pid_file` 相对路径基于程序目录转换为绝对路径。
- `max_body_size` 支持 `B/KB/K/MB/M/GB/G`；非法单位会让配置加载失败。
- Dev 日志输出到终端；Release 日志输出到 `log/<程序名>.log` 与 `log/<程序名>.gin.log`。
- `log.gin_level` 为空时跟随 `log.level`；配置热重载后 logger 自动刷新。
- `config.yaml`、`bin/`、`dist/`、日志文件、`coverage.out` 不提交。
- Go 注释使用中文为主；只在非显而易见逻辑处注释。

## ANTI-PATTERNS (THIS PROJECT)

- 不要直接返回 DB/ORM/Service 内部结构体；API 对外数据必须经 DTO 显式挑选字段。
- 不要在单个 handler 混用 REST HTTP status 与本项目 `HTTP 200 + body.code/status` 策略。
- 不要把敏感数据放入 JWT；JWT 只放必要身份标识。
- 不要提交真实 `config.yaml` 或密钥；JWT key 必须至少 32 字符且不能使用示例值。
- 不要重新引入服务内 TLS/ACME 作为默认能力；当前服务只监听 HTTP，TLS 由反向代理/Ingress/负载均衡终止。
- 不要在 `vendor/` 下写项目规范或修改第三方代码。

## COMMANDS

```bash
make help
make build
make build CROSS=1
make build-cross
make run
make dev
make test
make fmt
make lint
make verify
make clean
make clean-dist
make version
```

Command internals:
- `make test` runs `go test -v -coverprofile=coverage.out -covermode=atomic ./...` and prints Chinese summary + total coverage.
- `make fmt` runs `gofmt -w $(find . -name "*.go" -not -path "./vendor/*")`.
- `make lint` runs `go vet ./...`.
- `make verify` runs `fmt -> lint -> test`.
- Cross build outputs archives to `dist/`; Windows uses zip when available, others use tar.gz.

## TESTS

- Package-local tests dominate; no central test helper package, no `TestMain`, no `testdata/`.
- API tests use `gin.SetMode(gin.TestMode)` + `httptest` and assert both HTTP 200 and JSON body contract.
- Table-driven tests appear in config, response, middleware, bcrypt, page parsing.
- Tests that mutate globals must restore via `t.Cleanup` or equivalent.
- Root `pidfile_integration_test.go` builds and runs a real binary; it must support `testing.Short()` skip.
- Only `utils/id/id_test.go` currently has benchmarks.

## NOTES

- Current checkout references `config.yaml.example` in docs/Makefile packaging, but the file is not present.
- `IMPROVEMENTS.md` mentions `make test-cover`; current `Makefile` does not define that target.
- `private` route tree is intentionally a placeholder; add modules under `api/app/v1/private/<module>` and register them in `api/app/v1/private/router.go`.
