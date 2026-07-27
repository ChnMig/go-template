# PROJECT KNOWLEDGE BASE

## OVERVIEW

Gin + Viper + Zap 的 HTTP API 服务模板。仓库是单 Go module：`http-services`，可执行入口在根目录 `main.go`，没有 `cmd/`、`internal/`、`pkg/` 拆分。

## STRUCTURE

```text
http-services/
├── main.go          # CLI、版本、信号 context 与进程退出边界
├── bootstrap/       # 配置初始化、迁移、HTTP 生命周期、PID 与反向清理
├── Makefile         # build/run/dev/migrate/test/fmt/lint/verify
├── api/             # Gin 初始化、分层路由、中间件、统一响应
├── config/          # Viper 加载、启动快照、日志专属热更新和安全校验
├── common/          # 跨模块共享业务语义预留，如枚举、常量、跨模块 DTO、事件定义
├── domain/health/   # 当前唯一领域模块
├── db/              # 持久化适配层；msqldb/rdb 放 client、模型、查询 helper、数据库常量、迁移与缓存访问
├── services/        # 长驻任务预留；cron/worker/consumer 的启动、调度、生命周期管理
├── utils/           # 日志、JWT、context key、pidfile、ID 等基础设施
└── vendor/          # 本地依赖镜像目录；默认忽略不提交，不写项目约束
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| 程序入口 | `main.go` | CLI、版本输出、signal context 和进程退出；不直接获取应用资源 |
| 启动顺序/资源清理 | `bootstrap/` | `Initialize -> optional Migrate` 或 `handler -> listener -> PID -> Serve`，退出时反向清理 |
| 构建/测试/发布 | `Makefile` | 所有本地命令以 Make target 为准 |
| HTTP 路由/中间件 | `api/` | 见 `api/AGENTS.md` |
| 配置加载与日志热更新 | `config/load.go`, `config/watch.go` | Viper、`HTTP_SERVICES_` env、类型化启动快照、实例级 `WatchLogConfig` |
| 配置安全门禁 | `config/check.go` | JWT key 非空、非示例值、长度至少 32 |
| 配置快照与兼容变量 | `config/config.go` | 生产路径使用 `Config`；旧工具函数使用的兼容变量不进入 bootstrap |
| 日志生命周期 | `utils/log/logger.go`, `utils/log/monitor.go` | dev/release 分流、Gin 日志独立文件、轮转、热重建 |
| JWT 签发/解析 | `utils/authentication/jwt.go` | 直接读取 `config.JWTKey/JWTExpiration` |
| Context key | `utils/contextkey/keys.go` | trace_id、logger、jwtData、bound params 统一 key |
| 并发任务组 | `utils/taskgroup/` | 按策略取消并发任务、恢复 panic，并按输入顺序返回错误 |
| 领域健康状态 | `domain/health/` | 领域错误与状态，不依赖 Gin |
| 跨模块业务语义 | `common/` | 枚举、业务常量、跨模块 DTO、事件定义；当前为扩展占位 |
| 持久化适配 | `db/migrate.go`, `db/msqldb/`, `db/rdb/` | MySQL/GORM client、BaseModel、迁移聚合、Redis client、模型/查询扩展位置 |
| 长驻任务 | `services/cron/` | cron、worker、consumer 的调度与生命周期；当前为扩展占位 |
| 根级集成测试 | `pidfile_integration_test.go` | 构建真实二进制，`testing.Short()` 跳过 |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `main` | func | `main.go` | 单一可执行入口与进程边界 |
| `CLI` | var | `main.go` | Kong flags：`-d/--dev`、`-v/--version`、`-m/--migrate` |
| `bootstrap.Run` | func | `bootstrap/app.go` | 初始化、迁移或 HTTP 运行、PID 与确定性清理 |
| `api.NewRouter` | func | `api/router.go` | 显式传入配置与 registrar，返回构建错误 |
| `api.InitApi` | func | `api/router.go` | 旧调用方兼容入口；生产 bootstrap 不使用 |
| `config.Load` | func | `config/load.go` | 返回验证后的类型化启动快照；`LoadConfig` 仅保留旧调用兼容 |
| `config.WatchLogConfig` | func | `config/watch.go` | 每个应用实例只监听并更新 `LogConfig`，返回可关闭、可等待的 watcher |
| `config.CheckConfig` | func | `config/check.go` | JWT 配置安全校验 |
| `log.SetLogger` | func | `utils/log/logger.go` | 按显式日志快照与 dev 标志重建业务/Gin logger |
| `taskgroup.Run` | func | `utils/taskgroup/taskgroup.go` | 执行命名并发任务，等待全部退出并按输入顺序返回错误 |
| `middleware.NewRateLimiter` | func | `api/middleware/rate-limit.go` | 创建有界、惰性 TTL 清理且无后台 goroutine 的限流器 |
| `db.MigrateAll` | func | `db/migrate.go` | 顶层数据库迁移聚合入口 |
| `msqldb.New` | func | `db/msqldb/client.go` | 创建由 bootstrap 生命周期持有的 GORM MySQL client |
| `rdb.New` | func | `db/rdb/client.go` | 创建由 bootstrap 生命周期持有的 Redis client |

## CONVENTIONS

- 所有 Go 命令使用 `GOTOOLCHAIN=go1.26.5`；优先通过 Makefile 的固定工具版本执行格式化、lint 和验证。
- `main.go` 只处理 CLI、信号与退出；资源获取、迁移、HTTP 生命周期和清理放在 `bootstrap/`。
- 启动顺序不可调换：加载配置后先检测运行模式，再初始化 logger；配置校验通过后才启动热重载、迁移或 HTTP listener。
- HTTP、生命周期、JWT、数据库和 Redis 配置加载后保持不变；每个应用实例持有自己的 watcher，且 watcher 只能更新 `config.LogConfig`。
- bootstrap 必须先获取配置中启用的 MySQL/Redis，再通过 `api.NewRouter(api.DefaultOptions(runtimeConfig.HTTP))` 构建 handler，并在 listener/PID 之前传播错误；退出时按 worker、listener/PID、Redis、MySQL、日志的反向顺序清理。
- 新增长驻 worker 时由 `bootstrap` 获取并拥有，HTTP 停止入口后 cancel/join worker，再关闭其依赖资源。
- 配置文件名固定为 `config.yaml`，查找顺序：程序目录、工作目录、`/etc/http-services/`。
- 环境变量前缀固定 `HTTP_SERVICES_`；层级用 `_` 代替点，如 `HTTP_SERVICES_SERVER_PORT`。
- `server.pid_file` 相对路径基于程序目录转换为绝对路径。
- `max_body_size` 支持 `B/KB/K/MB/M/GB/G`；非法单位会让配置加载失败。
- `server.trusted_proxies` 只接受明确的 IP/CIDR；`static_dir` 为空会关闭 static 托管，`enable_cors` 控制默认 CORS。
- Dev 日志输出到终端；Release 日志输出到 `log/<程序名>.log` 与 `log/<程序名>.gin.log`。
- `log.gin_level` 为空时跟随 `log.level`；配置热重载后 logger 自动刷新。
- `api/` 只处理 HTTP 入参、DTO、响应和错误翻译；核心业务规则放 `domain/`。
- `domain/` 放可被 handler、worker、cron 复用的业务规则；`services/` 只放长驻任务的调度、消费循环和生命周期管理。
- `db/` 只表达存储结构、数据库常量、查询 helper、迁移和外部存储适配；跨表业务流程、状态机和领域错误不要下沉到 `db/`。
- 新增业务表域时，在 `db/msqldb/<module>/` 放 `model.go/query.go/constants.go/migrate.go`，并在 `db/migrate.go` 按依赖顺序注册迁移。
- `database.mysql_dsn` 与 Redis 配置通过 Viper 加载；空值表示禁用，非空值由 bootstrap 在启动期连接并持有，禁止恢复 package-global client。
- `common/` 放跨模块共享业务语义；`utils/` 只放与业务无关的基础设施工具。
- 数据库常量优先跟随对应数据库模块放在 `db/` 下；只有被多个业务模块作为业务语义共同使用时，才抽到 `common/`。
- 新增业务模块默认按 `db/msqldb/<module> -> domain/<module> -> api/app/v1/{open|private}/<module>` 链路落地，不新增顶层 `repository/dao/biz/model` 目录。
- API 层负责 DTO、Gin binding、response envelope 和领域错误到用户文案的映射；domain 层不 import Gin，不返回 HTTP 状态，不写用户提示文案。
- Redis key 拼接和缓存/session 访问放在使用它的 domain 包附近，通过 bootstrap 注入的 `*rdb.Client` 访问；业务 Redis key 不进 `utils/`。
- MySQL 查询 helper 使用 bootstrap 注入的 `*msqldb.Client`，需要 GORM 时显式调用其 `Database()`；不要增加全局便捷入口。
- 事务边界由 domain 决定；db 层可提供 `InTx` helper，但不要在单表 query 函数里隐藏跨表事务。
- DB model 不定义对外 JSON 契约；接口字段和 JSON tag 放 API DTO。
- 新增配置项必须同步 `config/config.go`、`config/load.go`、`config.yaml.example` 和配置测试；环境变量继续遵循 `HTTP_SERVICES_<SECTION>_<KEY>`。
- `config.yaml`、`bin/`、`dist/`、日志文件、`coverage.out` 不提交。
- Go 注释使用中文为主；只在非显而易见逻辑处注释。

## ANTI-PATTERNS (THIS PROJECT)

- 不要直接返回 DB/ORM/Service 内部结构体；API 对外数据必须经 DTO 显式挑选字段。
- 不要在单个 handler 混用 REST HTTP status 与本项目 `HTTP 200 + body.code/status` 策略。
- 不要把敏感数据放入 JWT；JWT 只放必要身份标识。
- 不要把业务常量或数据库常量放进 `utils/`。
- 不要把 cron/consumer/worker 的运行循环塞进 `domain/`；`domain/` 应保持可复用业务规则。
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
make migrate
make test
make test-race
make fmt
make lint
make verify
make clean
make clean-dist
make version
```

Command internals:
- `make test` runs `go test -race -shuffle=on -count=1 ./...`; `make test-race` is its compatibility alias.
- `make test-coverage` writes `coverage.out` explicitly when coverage is wanted.
- `make fmt` uses the pinned gofumpt and golangci formatter; `make fmt-check` is non-mutating.
- `make lint` runs pinned golangci-lint, nilaway, and `go vet`.
- `make verify` is non-mutating and runs formatting checks, lint, race tests, pure-LOC, module verification, and a temporary build check.
- Cross build outputs archives to `dist/`; Windows uses zip when available, others use tar.gz.

## TESTS

- Package-local tests dominate; no central test helper package, no `TestMain`, no `testdata/`.
- API tests use `gin.SetMode(gin.TestMode)` + `httptest` and assert both HTTP 200 and JSON body contract.
- Table-driven tests appear in config, response, middleware, bcrypt, page parsing.
- Tests that mutate globals must restore via `t.Cleanup` or equivalent.
- Root `pidfile_integration_test.go` builds and runs a real binary; it must support `testing.Short()` skip.
- `bootstrap/` tests use injected initialization, listener, server, PID and cleanup seams; no real signals, ports or sleeps.
- Only `utils/id/id_test.go` currently has benchmarks.

## NOTES

- `config.yaml.example` 是配置模板来源，会随默认 `PACKAGE_FILES` 一起打包；真实 `config.yaml` 不提交。
- `IMPROVEMENTS.md` 是历史改进记录；当前可用命令以本文件 COMMANDS、README 和 `Makefile` 为准。
- `private` route tree is intentionally a placeholder; add modules under `api/app/v1/private/<module>` and register them in `api/app/v1/private/router.go`.
