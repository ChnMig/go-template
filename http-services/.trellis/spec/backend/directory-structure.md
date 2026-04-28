# Directory Structure

> 本项目后端代码组织方式。

---

## Overview

这是一个单仓库 Go HTTP 服务模板。入口在 `main.go`，HTTP 路由在 `api/`，领域层逻辑在 `domain/`，配置在 `config/`，可复用工具在 `utils/`。新增模块时优先延续现有的“API 适配层 + domain 领域层 + utils 通用能力”分层，不把业务语义直接塞进全局中间件或工具包。

---

## Directory Layout

```
.
├── main.go                         # CLI 参数、配置加载、日志初始化、HTTP server 生命周期
├── config/                         # Viper 配置加载、默认值、关键配置校验
├── api/
│   ├── router.go                   # Gin 初始化、全局中间件、/api 分组挂载
│   ├── middleware/                 # JWT、TraceID、限流、安全头、参数绑定等通用 HTTP 中间件
│   ├── response/                   # 统一响应结构和错误码
│   └── app/
│       └── v1/
│           ├── open/               # 公开接口分组
│           └── private/            # 私有接口分组
├── domain/                         # 领域模型、领域服务、领域错误
├── utils/                          # 日志、JWT、ID、随机数、Gin context key、pidfile 等通用工具
└── *_test.go                       # 测试与被测代码同包或同目录放置
```

---

## Module Organization

### 路由注册

路由采用逐层注册模式：

- `api/router.go` 只负责 Gin 初始化、全局中间件、静态文件和 `/api` 根分组。
- `api/app/router.go` 负责挂载版本分组，例如 `/api/v1`。
- `api/app/v1/router.go` 负责挂载 `open` / `private` 子分组。
- 业务模块在自己的目录里暴露 `RegisterOpenRoutes` / `RegisterPrivateRoutes`。

真实示例：

```go
// api/app/v1/open/health/router.go
func RegisterOpenRoutes(open *gin.RouterGroup) {
	if open == nil {
		return
	}
	open.GET("/health", Status)
}
```

### API 层与领域层

API handler 负责读取 Gin 上下文、调用领域层、组装 DTO、映射错误并返回统一响应。领域层不依赖 Gin，也不关心前端文案。

真实示例：

```go
// api/app/v1/open/health/health.go
status, err := domain.GetStatus()
if err != nil {
	log.WithRequest(c).Error("健康检查失败", zap.Error(err))
	ReturnDomainError(c, err)
	return
}
```

```go
// domain/health/status.go
func GetStatus() (Status, error) {
	return Status{
		Status:    "ok",
		Ready:     true,
		Uptime:    time.Since(startTime),
		Timestamp: time.Now().Unix(),
	}, nil
}
```

### 配置与工具

- 配置默认值、文件读取、环境变量覆盖在 `config/load.go`。
- 安全校验在 `config/check.go`。
- 可跨模块复用的能力放入 `utils/<topic>/`，例如 `utils/log`、`utils/authentication`、`utils/id`。
- 工具包不能反向依赖业务模块。

---

## Naming Conventions

- 包名使用小写短名，必要时用目录表达层级，例如 `api/app/v1/open/health`。
- 路由注册函数使用 `RegisterRoutes`、`RegisterOpenRoutes`、`RegisterPrivateRoutes`。
- DTO 与领域对象分开命名，例如 `api/app/v1/open/health/dto.go` 中的 `StatusDTO` 不直接复用 `domain/health/status.go` 的 `Status`。
- 测试文件使用 `*_test.go`，测试函数使用 `TestXxx`。
- `utils/` 下目录使用 Go 常见的小写短名，不使用短横线；例如 `utils/pathtool`、`utils/runmodel`。

---

## Examples

- `api/router.go`：顶层路由与全局中间件顺序。
- `api/app/v1/open/health/`：完整的公开模块示例，包含 router、handler、DTO、错误映射和测试。
- `domain/health/`：领域实体与领域错误示例。
- `api/response/`：统一响应结构。
- `utils/log/`：跨模块基础设施工具示例。

## Common Mistakes

- 不要在 `api/router.go` 中直接写具体业务接口；它只做顶层编排。
- 不要让 `domain/` 依赖 `gin.Context`、`api/response` 或 HTTP 文案。
- 不要为单个模块重复实现通用能力；先检查 `utils/`、`api/middleware/` 和 `api/response/` 是否已有可复用实现。
