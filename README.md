# go-template

Golang 项目模板集合

## 项目简介

go-template 的目标是**提高开发效率，而不是极致简化**。😉  
因此模板中会默认集成一些“够好用”的第三方库😊，并在实践中持续打磨，让它们在工程化和可维护性上也“足够好”😋。

本仓库目前主要包含一个 HTTP API 服务模板 `http-services`，适合作为中小型后端服务的起步工程。

模板默认提供健康检查与通用基础设施。业务模型、领域流程和系统间协议由使用模板的项目实现；通用能力反哺须保持业务无关，并附带独立测试。

## 更新记录

| 更新时间 | 更新内容（概述） | 对应详细章节 |
| --- | --- | --- |
| 2026-11-15 | 明确脚手架边界，补充严格校验、日志和文档规范入口，并校正配置与验证说明。 | [开发规范与文档](#开发规范与文档)、[日志说明：接口内 zap 使用约定（Logging）](#日志说明接口内-zap-使用约定logging)、[测试（Testing）](#测试testing) |

## 开发规范与文档

- [HTTP Services 使用与开发规范](http-services/README.md)：参数校验、JWT、错误、日志、接口契约和公共 ID 复用。
- [文档规范与导航](http-services/docs/README.md)：按实际版本区分页面、数据库、后端与接口说明，系统间协议独立维护；字段、错误、示例及更新记录要求。
- [HTTP 请求日志排查](http-services/docs/troubleshooting/request-logging.md)：原始数据、解析错误、采集上限与未读完标记。
- [代理开发约定](AGENT.md) 与 [模块约束](http-services/AGENTS.md)：变更范围、目录边界及验证要求。

## 使用 gonew 下载模板

推荐通过 [gonew](https://pkg.go.dev/golang.org/x/tools/cmd/gonew) 下载并初始化项目，这样可以直接指定新模块名，而不是简单拷贝。

## 目录结构

### http-services

适合作为 HTTP API 服务模板使用。

## 快速开始（Quick Start）

### 初始化配置

1. 拷贝示例配置文件：

   ```bash
   cd http-services
   cp config.yaml.example config.yaml
   ```

2. 编辑 `config.yaml` 并根据实际环境修改配置，尤其是：
   - `jwt.key`: 启用 JWT 时必须改为至少 32 字符的强密钥，并由项目在启动边界调用 `config.CheckConfig`；模板默认健康检查不强制启用 JWT
   - `jwt.expiration`: 访问令牌有效期（例如 `"12h"`、`"24h"`、`"30m"`）
   - `database.mysql_dsn`: 需要运行迁移或访问 MySQL 时填写
   - `redis.host` / `redis.password`: 需要 session、验证码、缓存等 Redis 能力时填写

3. 构建并运行：

   ```bash
   # 显示帮助
   make help

   # 构建
   make build

   # 运行（生产模式）
   make run

   # 运行（开发模式）
   make dev

   # 执行数据库迁移后退出（需要配置 database.mysql_dsn）
   make migrate
   ```

### 跨平台打包（Cross-Platform Packaging）

通过 `Makefile` 可以一次性为多个平台打包可执行文件，产物会输出到 `dist/` 目录下，文件名中包含版本号、操作系统和架构信息。

基础用法：

```bash
cd http-services

# 交叉编译并打包（Unix 使用 tar.gz，Windows 使用 zip）
make build CROSS=1
# 或者显式调用
make build-cross
```

通过环境变量 `PLATFORMS` 可以自定义目标平台（默认：`linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64`）：

```bash
make build CROSS=1 \
  PLATFORMS="linux/amd64 linux/arm64 darwin/arm64 windows/amd64"
```

说明：
- 产物二进制会内嵌版本信息：`Version`、`BuildTime`、`GitCommit`。
- 默认使用 `CGO_ENABLED=0`；如依赖 CGO，可自行覆盖。
- 若存在 `README.md`、`config.yaml.example` 等文件，会一并打包到发布目录。
- Windows 平台在系统存在 `zip` 命令时使用 `.zip`，其余平台使用 `.tar.gz`。

### 命令行参数（Command Line Options）

```bash
# 开发模式
./bin/http-services --dev

# 显示版本信息
./bin/http-services --version

# 执行数据库迁移后退出
./bin/http-services --migrate

# 显示帮助
./bin/http-services --help
```

## Configuration

该项目通过 YAML 配置文件管理服务参数。

### 配置文件示例结构（简化）

```yaml
server:
  port: 8080

jwt:
  key: "YOUR_SECRET_KEY_HERE"
  expiration: "12h"

database:
  mysql_dsn: ""

redis:
  host: "127.0.0.1:6379"
  password: ""
  key_prefix: "" # 非空时必须以 : 结尾，例如 service:env:
```

**重要说明**：`config.yaml` 已被加入 `.gitignore`，避免敏感配置被提交到仓库。请始终以 `config.yaml.example` 为模版创建本地配置。

`redis.key_prefix` 默认为空；每个非空值都必须以 `:` 结尾，加载器不会自动补齐或校验。启用或修改前缀后需要重启服务，且模板不会迁移旧前缀下的 Redis key。内置 hook 支持的命令族见 [HTTP Services Redis 公共 key 前缀说明](http-services/README.md#redis-公共-key-前缀)；其他自定义命令和 Redis module 命令需要显式增加 hook 支持。

### Configuration Reload & Restart

当前模板虽然使用 Viper 支持监控 `config.yaml` 变更，但大部分配置（例如 `server.port`、全局限流、超时配置等）只在进程启动时读取并应用，运行中的 HTTP 服务器和路由不会自动根据新配置重新构建。模板只监听 HTTP，TLS 由反向代理或负载均衡终止。

因此在生产环境中，**修改配置后建议始终重启服务进程**，以确保所有配置项都按预期生效；不要依赖“热更新配置”来修改端口或调整全局限流策略。

## 功能特性（Features）

### 核心能力（Core Components）

- **JWT 鉴权**：基于 `JWT` 的标准认证流程，包含基础安全校验
- **CORS**：跨域请求中间件
- **密码加密**：使用 `BCrypt` 进行安全密码哈希
- **分页能力**：内置分页工具，支持可配置默认值
- **优雅下线（Graceful Shutdown）**：HTTP 服务器支持 10s 超时的优雅关闭
- **健康检查**：提供 `/api/v1/open/health` 单一健康检查端点

### 中间件（Middleware）

- `TraceID`：为每个请求生成并传播请求 ID，支持链路追踪
- `SecurityHeaders`：设置常用安全响应头（如 `X-Content-Type-Options`、`X-Frame-Options` 等）
- `BodySizeLimit`：请求体大小限制（默认 10MB），防止大包攻击
- `TokenVerify`：JWT 鉴权中间件，只接受 HS256 和有效 `exp`
- `CheckStrictJSONParam`：按接口启用的严格 JSON 绑定，拒绝未知字段与尾随 JSON，复用参数校验及原始输入诊断
- `CorssDomainHandler`：跨域（CORS）处理中间件
- `IPRateLimit`：基于 IP 的令牌桶限流
- `TokenRateLimit`：基于登录用户 Token 的限流

### 工具类（Utilities）

- **Authentication**（`utils/authentication`）：
  - JWT 生成与解析
  - HS256 签名与校验；校验时拒绝空签名密钥、其他算法、无过期时间或已过期 Token
  - 标准 claims 支持

- **Encryption**（`utils/encryption`）：
  - 使用 `BCrypt` 的密码哈希
  - 密码校验工具

- **ID Generation**（`utils/id`）：
  - 基于 `Sonyflake` 的分布式唯一 ID
  - 基于 MD5 的唯一 ID 生成

- **Logging**（`utils/log`）：
  - 基于 `go.uber.org/zap` 的结构化日志（structured logging），支持开发/生产两种输出模式
  - 在 `api.InitApi` 中将 Gin 的默认访问日志和错误日志重定向到 zap，框架日志与业务日志统一输出
  - 提供从 `gin.Context` 获取带请求上下文信息的 logger，方便接口内按请求维度记录日志

### 依赖（Dependencies）

主要三方依赖包括：

- `github.com/gin-gonic/gin`：Web 框架
- `github.com/golang-jwt/jwt/v5`：JWT 实现
- `gorm.io/gorm` / `gorm.io/driver/mysql`：MySQL 持久化基础组件
- `github.com/redis/go-redis/v9`：Redis 客户端
- `github.com/goccy/go-yaml`：YAML 解析
- `github.com/alecthomas/kong`：命令行解析
- `golang.org/x/crypto/bcrypt`：密码加密
- `go.uber.org/zap`：结构化日志

### 日志说明：接口内 zap 使用约定（Logging）

本模板统一使用 `go.uber.org/zap` 作为日志组件，相关封装位于 `http-services/utils/log` 包：

- 运行模式（开发 / 生产）由 `config.RunModel` 控制：开发模式输出到终端，生产模式输出到按日期滚动的日志文件。
- 在 `api.InitApi` 中通过 `httplog.NewZapWriter` 将 Gin 的默认日志输出（访问日志、panic 等）重定向到 zap，框架日志与业务日志走同一管道。
- `TraceID` 中间件会为每个请求生成 `trace_id`，并在 `gin.Context` 中注入带 `trace_id`、`method`、`path`、`client_ip` 等字段的 logger。

在接口 handler 中，推荐按如下方式使用 zap：

```go
import (
    httplog "http-services/utils/log"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

func (h *Handler) GetUser(c *gin.Context) {
    // 推荐使用 FromContext 获取带 trace_id 等上下文信息的 logger
    logger := httplog.FromContext(c)
    logger.Info("获取用户信息",
        zap.String("user_id", c.Param("id")),
    )

    // 仅在排查请求参数相关问题时，使用 WithRequest 追加请求详情字段
    // 记录 query / form / 路径参数 / 预绑定业务参数，绑定失败时回退原始请求快照
    debugLogger := httplog.WithRequest(c)
    debugLogger.Debug("请求参数详情日志")
}
```

一般情况下：

- 正常业务日志：优先使用 `httplog.FromContext(c)`，保证所有日志都带有 `trace_id`，便于链路追踪。
- 深度排查问题时：在局部（例如特定 handler）使用 `httplog.WithRequest(c)` 打印请求参数，避免对所有请求都记录大体积参数导致日志膨胀。
- 错误处理：在边界层保留原始错误和请求上下文，写清操作、失败步骤与具体原因。业务拒绝通常记 Warn，内部异常记 Error，纯客户端取消沿用 Debug 降噪；统一响应已有诊断日志，不强制重复记录 Error。
- 解析失败：保留对应的原始入参、接口返回值或字段值，不能只输出部分结构体。通用请求采集只观察已读取内容，文本限 64 KiB，截断及未读完均明确标记；完整原文需求由具体集成在输入与耗时预算内处理，详见 [日志规范](http-services/README.md#日志记录要求)。

## 项目结构（Project Structure）

```text
http-services/
├── api/                  # Gin 传输层、版本路由、中间件、统一响应
│   ├── app/v1/open/health/  # 健康检查开放接口
│   ├── app/v1/private/      # 私有接口预留
│   ├── middleware/          # TraceID、访问日志、JWT、限流、CORS 等
│   └── response/            # 统一响应封装
├── common/              # 跨模块共享语义预留，如枚举、常量、跨模块 DTO
├── domain/              # 领域模型与领域服务，放核心业务规则
├── db/                  # 持久化适配层，详细约定见 http-services/README.md
│   ├── migrate.go       # 顶层数据库迁移聚合入口
│   ├── msqldb/          # MySQL/GORM client、BaseModel、model/query/constants 扩展位置
│   └── rdb/             # Redis client、session、cache、幂等访问封装扩展位置
├── services/            # 长驻服务和后台任务预留
│   ├── cron/            # 定时任务调度预留
│   └── health/          # health 相关后台任务占位
├── config/              # 配置加载、默认值、环境变量与安全校验
├── utils/               # 基础设施工具，不放具体业务规则
│   ├── authentication/  # JWT 工具
│   ├── encryption/      # 密码加密工具
│   ├── id/              # ID 生成工具
│   ├── log/             # 日志封装
│   ├── pathtool/        # 路径工具
│   └── runmodel/        # 运行模式工具
├── config.yaml.example  # 配置文件示例
├── Makefile             # 构建、运行、测试脚本
└── main.go              # 应用入口
```

根 README 只保留结构总览；`db/`、`domain/`、`services/` 等目录的细粒度放置规则见 `http-services/README.md`。

## 测试（Testing）

运行全部测试：

```bash
cd http-services
make test
```

运行带覆盖率的测试：

```bash
make test-coverage
```

`make test` 默认带 race、随机顺序并禁用测试结果缓存；`make verify` 执行静态检查、race 测试、临时构建与依赖校验，不修改源码。纯文档变更检查链接、锚点及示例。

当前已覆盖的主要测试包括：
- JWT 鉴权与 Token 处理
- BCrypt 密码哈希与校验
- ID 生成（Sonyflake + MD5）

## 安全特性（Security Features）

- **JWT 校验**：统一解析入口限定 HS256 和有效过期时间，空签名密钥拒绝；启用 JWT 的项目在启动边界调用 `config.CheckConfig` 校验密钥强度与有效期
- **请求体大小限制**：通过 `BodySizeLimit` 防止大体积请求导致的 DoS 风险
- **安全响应头**：自动为响应添加通用安全 Header
- **限流能力**：支持按 IP 或登录用户 Token 进行限流
- **请求 ID 追踪**：通过 `TraceID` 中间件支持链路追踪与问题排查

## 开发说明（Development Notes）

本模板基于 [art-design-pro-edge-go-server](https://github.com/ChnMig/art-design-pro-edge-go-server) 演进而来，会持续同步其核心能力更新，并针对通用场景做适配与精简。
