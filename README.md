# go-template

Golang 项目模板集合

## 项目简介

go-template 的目标是**提高开发效率，而不是极致简化**。😉  
因此模板中会默认集成一些“够好用”的第三方库😊，并在实践中持续打磨，让它们在工程化和可维护性上也“足够好”😋。

本仓库目前主要包含一个 HTTP API 服务模板 `http-services`，适合作为中小型后端服务的起步工程。

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
   - `jwt.key`: **必须修改为至少32字符的强密钥** (服务启动时会进行安全检查)
   - `jwt.expiration`: 访问令牌有效期（例如 `"12h"`、`"24h"`、`"30m"`）

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
```

**重要说明**：`config.yaml` 已被加入 `.gitignore`，避免敏感配置被提交到仓库。请始终以 `config.yaml.example` 为模版创建本地配置。

### Configuration Reload & Restart

当前模板虽然使用 Viper 支持监控 `config.yaml` 变更，但大部分配置（例如 `server.port`、TLS/ACME 开关、全局限流、超时配置等）只在进程启动时读取并应用，运行中的 HTTP 服务器和路由不会自动根据新配置重新构建。

因此在生产环境中，**修改配置后建议始终重启服务进程**，以确保所有配置项都按预期生效；不要依赖“热更新配置”来切换是否启用 TLS、修改端口或调整全局限流策略。

## 功能特性（Features）

### 核心能力（Core Components）

- **JWT 鉴权**：基于 `JWT` 的标准认证流程，包含基础安全校验
- **CORS**：跨域请求中间件
- **密码加密**：使用 `BCrypt` 进行安全密码哈希
- **分页能力**：内置分页工具，支持可配置默认值
- **优雅下线（Graceful Shutdown）**：HTTP 服务器支持 10s 超时的优雅关闭
- **健康检查**：提供 `/health` 和 `/ready` 等监控探针接口

### 中间件（Middleware）

- `RequestID`：为每个请求生成并传播请求 ID，支持链路追踪
- `SecurityHeaders`：设置常用安全响应头（如 `X-Content-Type-Options`、`X-Frame-Options` 等）
- `BodySizeLimit`：请求体大小限制（默认 10MB），防止大包攻击
- `TokenVerify`：JWT 鉴权中间件
- `CorssDomainHandler`：跨域（CORS）处理中间件
- `IPRateLimit`：基于 IP 的令牌桶限流
- `TokenRateLimit`：基于登录用户 Token 的限流

### 工具类（Utilities）

- **Authentication**（`util/authentication`）：
  - JWT 生成与解析
  - HS256 签名与校验
  - 标准 claims 支持

- **Encryption**（`util/encryption`）：
  - 使用 `BCrypt` 的密码哈希
  - 密码校验工具

- **ID Generation**（`util/id`）：
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
- `github.com/goccy/go-yaml`：YAML 解析
- `github.com/alecthomas/kong`：命令行解析
- `golang.org/x/crypto/bcrypt`：密码加密
- `go.uber.org/zap`：结构化日志

### 日志说明：接口内 zap 使用约定（Logging）

本模板统一使用 `go.uber.org/zap` 作为日志组件，相关封装位于 `http-services/utils/log` 包：

- 运行模式（开发 / 生产）由 `config.RunModel` 控制：开发模式输出到终端，生产模式输出到按日期滚动的日志文件。
- 在 `api.InitApi` 中通过 `httplog.NewZapWriter` 将 Gin 的默认日志输出（访问日志、panic 等）重定向到 zap，框架日志与业务日志走同一管道。
- `RequestID` 中间件会为每个请求生成 `trace_id`，并在 `gin.Context` 中注入带 `trace_id`、`method`、`path`、`client_ip` 等字段的 logger。

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
    // 默认不会记录请求体（body），只记录 query / form / 路径参数 / 预绑定业务参数
    debugLogger := httplog.WithRequest(c)
    debugLogger.Debug("请求参数详情日志")
}
```

一般情况下：

- 正常业务日志：优先使用 `httplog.FromContext(c)`，保证所有日志都带有 `trace_id`，便于链路追踪。
- 深度排查问题时：在局部（例如特定 handler）使用 `httplog.WithRequest(c)` 打印请求参数，避免对所有请求都记录大体积参数导致日志膨胀。
- 错误处理：在向客户端返回错误响应之前，必须至少记录一条 `Error` 级别日志，且日志消息与字段应能够反映真实错误原因，避免使用“操作失败”这类模糊描述；推荐在统一错误映射函数（如 `ReturnDomainError`）中调用 `httplog.WithRequest(c).Error("健康检查领域错误", zap.Error(err))` 统一记录错误日志。

## 项目结构（Project Structure）

```text
http-services/
├── api/
│   ├── app/              # 业务接口入口
│   │   ├── example/      # 示例接口（如有）
│   │   └── health/       # 健康检查接口
│   ├── middleware/       # 中间件组件
│   └── response/         # 统一响应封装
├── config/               # 配置加载与校验
├── utils/                # 通用工具包
│   ├── authentication/   # JWT 工具（含测试）
│   ├── encryption/       # 密码加密工具（含测试）
│   ├── id/               # ID 生成工具（含测试）
│   ├── log/              # 日志封装
│   ├── path-tool/       # 路径工具
│   └── run-model/       # 运行模式工具
├── db/                   # 数据库层（预留）
├── services/             # 业务服务层（预留）
├── common/               # 公共通用层（预留）
└── main.go              # 应用入口
```

## 测试（Testing）

运行全部测试：

```bash
make test
```

运行带覆盖率的测试：

```bash
go test -cover ./...
```

当前已覆盖的主要测试包括：
- JWT 鉴权与 Token 处理
- BCrypt 密码哈希与校验
- ID 生成（Sonyflake + MD5）

## 安全特性（Security Features）

- **JWT 密钥校验**：启动时会检测 JWT 密钥强度，弱密钥或默认值会拒绝启动
- **请求体大小限制**：通过 `BodySizeLimit` 防止大体积请求导致的 DoS 风险
- **安全响应头**：自动为响应添加通用安全 Header
- **限流能力**：支持按 IP 或登录用户 Token 进行限流
- **请求 ID 追踪**：通过 `RequestID` 中间件支持链路追踪与问题排查

## 开发说明（Development Notes）

本模板基于 [art-design-pro-edge-go-server](https://github.com/ChnMig/art-design-pro-edge-go-server) 演进而来，会持续同步其核心能力更新，并针对通用场景做适配与精简。
