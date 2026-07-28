# PROJECT KNOWLEDGE BASE

## OVERVIEW

Gin + Viper + Zap 的 HTTP API 服务模板。仓库是单 Go module：`http-services`，可执行入口在根目录 `main.go`，没有 `cmd/`、`internal/`、`pkg/` 拆分。

## STRUCTURE

```text
http-services/
├── main.go          # CLI、配置、运行模式、日志、迁移、HTTP server、优雅关闭
├── Makefile         # build/run/dev/migrate/test/fmt/lint/verify
├── api/             # Gin 初始化、分层路由、中间件、统一响应
├── config/          # Viper 加载、默认值、环境变量、热重载、安全校验
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
| 程序入口/启动顺序 | `main.go` | `LoadConfig -> runmodel.Detect -> log.GetLogger/StartMonitor -> WatchConfig -> CheckConfig -> optional MigrateAll -> api.InitApi` |
| 构建/测试/发布 | `Makefile` | 所有本地命令以 Make target 为准 |
| HTTP 路由/中间件 | `api/` | 见 `api/AGENTS.md` |
| 配置默认值与热重载 | `config/load.go` | Viper、`HTTP_SERVICES_` env、`parseSize`、`WatchConfig` |
| 配置安全门禁 | `config/check.go` | JWT key 非空、非示例值、长度至少 32 |
| 全局配置变量 | `config/config.go` | `ListenPort`、超时、限流、日志、分页默认值 |
| 日志生命周期 | `utils/log/log.go` | dev/release 分流、Gin 日志独立文件、轮转、热重建 |
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
| `main` | func | `main.go` | 单一可执行入口、服务生命周期编排 |
| `CLI` | var | `main.go` | Kong flags：`-d/--dev`、`-v/--version`、`-m/--migrate` |
| `api.InitApi` | func | `api/router.go` | Gin engine、全局中间件、`/api` 挂载 |
| `config.LoadConfig` | func | `config/load.go` | 默认值、配置文件、环境变量、全局变量应用 |
| `config.WatchConfig` | func | `config/load.go` | 配置热重载，回调里刷新 logger |
| `config.CheckConfig` | func | `config/check.go` | JWT 配置安全校验 |
| `log.SetLogger` | func | `utils/log/log.go` | 按运行模式和日志级别重建业务/Gin logger |
| `taskgroup.Run` | func | `utils/taskgroup/taskgroup.go` | 执行命名并发任务，等待全部退出并按输入顺序返回错误 |
| `middleware.CleanupAllLimiters` | func | `api/middleware/rate-limit.go` | 服务退出时清理限流器 goroutine |
| `db.MigrateAll` | func | `db/migrate.go` | 顶层数据库迁移聚合入口 |
| `msqldb.GetClient` | func | `db/msqldb/client.go` | 懒初始化 GORM MySQL client |
| `rdb.GetClient` | func | `db/rdb/client.go` | 懒初始化 Redis client |

## CONVENTIONS

- 启动顺序不可调换：运行模式必须在初始化 logger 前设置；logger 初始化后再启动配置热重载。
- 配置文件名固定为 `config.yaml`，查找顺序：程序目录、工作目录、`/etc/http-services/`。
- 环境变量前缀固定 `HTTP_SERVICES_`；层级用 `_` 代替点，如 `HTTP_SERVICES_SERVER_PORT`。
- `server.pid_file` 相对路径基于程序目录转换为绝对路径。
- `max_body_size` 支持 `B/KB/K/MB/M/GB/G`；非法单位会让配置加载失败。
- Dev 日志输出到终端；Release 日志输出到 `log/<程序名>.log` 与 `log/<程序名>.gin.log`。
- `log.gin_level` 为空时跟随 `log.level`；配置热重载后 logger 自动刷新。
- `api/` 只处理 HTTP 入参、DTO、响应和错误翻译；核心业务规则放 `domain/`。
- `domain/` 放可被 handler、worker、cron 复用的业务规则；`services/` 只放长驻任务的调度、消费循环和生命周期管理。
- `db/` 只表达存储结构、数据库常量、查询 helper、迁移和外部存储适配；跨表业务流程、状态机和领域错误不要下沉到 `db/`。
- 新增业务表域时，在 `db/msqldb/<module>/` 放 `model.go/query.go/constants.go/migrate.go`，并在 `db/migrate.go` 按依赖顺序注册迁移。
- `database.mysql_dsn` 与 Redis 配置通过 Viper 加载；普通 HTTP 启动不强制连接 MySQL/Redis，迁移或业务首次调用 client 时才初始化。
- `common/` 放跨模块共享业务语义；`utils/` 只放与业务无关的基础设施工具。
- 数据库常量优先跟随对应数据库模块放在 `db/` 下；只有被多个业务模块作为业务语义共同使用时，才抽到 `common/`。
- 新增业务模块默认按 `db/msqldb/<module> -> domain/<module> -> api/app/v1/{open|private}/<module>` 链路落地，不新增顶层 `repository/dao/biz/model` 目录。
- API 层负责 DTO、Gin binding、response envelope 和领域错误到用户文案的映射；domain 层不 import Gin，不返回 HTTP 状态，不写用户提示文案。
- Redis key 拼接和缓存/session 访问放在使用它的 domain 包附近，通过 `db/rdb.Client()` 获取 client 并处理错误；业务 Redis key 不进 `utils/`。
- MySQL 查询 helper 优先使用 `msqldb.Client()` 返回的 `(*gorm.DB, error)`；`GetClient()` 只作为兼容便捷入口，不建议在新业务里忽略初始化错误。
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
- `make test` runs `go test -v -coverprofile=coverage.out -covermode=atomic ./...` and prints Chinese summary + total coverage.
- `make test-race` runs `go test -race -shuffle=on -count=1 ./...`.
- `make fmt` runs `gofmt -w $(find . -name "*.go" -not -path "./vendor/*")`.
- `make lint` runs `go vet ./...`.
- `make verify` runs `fmt -> lint -> test -> test-race`.
- Cross build outputs archives to `dist/`; Windows uses zip when available, others use tar.gz.

## TESTS

- Package-local tests dominate; no central test helper package, no `TestMain`, no `testdata/`.
- API tests use `gin.SetMode(gin.TestMode)` + `httptest` and assert both HTTP 200 and JSON body contract.
- Table-driven tests appear in config, response, middleware, bcrypt, page parsing.
- Tests that mutate globals must restore via `t.Cleanup` or equivalent.
- Root `pidfile_integration_test.go` builds and runs a real binary; it must support `testing.Short()` skip.
- Only `utils/id/id_test.go` currently has benchmarks.

## NOTES

- `config.yaml.example` 是配置模板来源，会随默认 `PACKAGE_FILES` 一起打包；真实 `config.yaml` 不提交。
- `IMPROVEMENTS.md` 是历史改进记录；当前可用命令以本文件 COMMANDS、README 和 `Makefile` 为准。
- `private` route tree is intentionally a placeholder; add modules under `api/app/v1/private/<module>` and register them in `api/app/v1/private/router.go`.
