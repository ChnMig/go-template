# Quality Guidelines

> 本项目后端质量标准。

---

## Overview

后端代码以简单、清晰、可测试为优先。变更应尽量复用现有模式：Gin 路由分层、统一响应、Zap 日志、Viper 配置、标准库测试。除非任务明确要求，不引入新的框架或测试体系。

当前 Go 版本由 `go.mod` 管理，验证命令由 `Makefile` 提供。

---

## Forbidden Patterns

- 不要跳过 `api/response` 直接返回自定义 JSON。
- 不要让 `domain/` 依赖 Gin、HTTP 响应结构或前端文案。
- 不要在 handler 中直接实现可复用中间件、日志、认证、限流、ID 生成等通用能力；先检查 `api/middleware/` 和 `utils/`。
- 不要在包初始化阶段创建生产日志文件或启动后台 goroutine；日志初始化由 `main.go` 控制。
- 不要记录密钥、token、证书私钥等敏感信息。
- 不要引入未使用的配置项、废弃代码或“以后可能用”的抽象。
- 不要在脚手架服务进程内重新引入 ACME 自动证书签发或本地证书文件 TLS 热更新；HTTPS/TLS 由反向代理或网关层处理。
- 单个代码文件不要超过 1000 行；接近上限时按职责拆分。

---

## Required Patterns

- Go 代码必须通过 `gofmt`，项目提供 `make fmt`。
- 新增配置项应同时更新默认值、全局配置变量、`applyConfig` 映射、配置校验或测试，并考虑 `config.yaml.example` 是否需要同步。
- 新增接口应遵循 `api -> app -> v1 -> open/private -> module` 路由注册链。
- 新增 handler 应使用 `response.Return*` 系列函数返回。
- 需要请求上下文日志时使用 `log.FromContext(c)`；只有错误排查路径使用 `log.WithRequest(c)`。
- 可取消或需清理的后台资源应提供停止逻辑，参考 `utils/log.StopMonitor()` 和 `api/middleware.CleanupAllLimiters()`。
- 修改常量、配置或工具函数前先搜索现有引用，确认没有遗漏联动点。

---

## Testing Requirements

测试使用 Go 标准库 `testing`，HTTP 测试使用 `net/http/httptest` 和 Gin test mode。测试文件与被测代码同目录。

常见模式：

```go
gin.SetMode(gin.TestMode)
w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)
```

配置、工具函数和中间件使用表驱动测试或直接断言。涉及全局配置或全局 logger 的测试必须保存旧值，并用 `t.Cleanup` 恢复。

```go
oldLogLevel := config.LogLevel
t.Cleanup(func() {
	config.LogLevel = oldLogLevel
	SetLogger()
})
```

集成测试可以构建测试二进制并使用临时目录；长耗时或外部进程类测试应支持 `testing.Short()` 跳过，例如 `pidfile_integration_test.go`。进程生命周期类测试需要给启动阶段预留足够时间，并在等待关键文件或端口就绪时同步检测子进程是否提前退出，避免全量并发测试负载下出现偶发超时且缺少诊断信息。

新增或变更功能必须补充单元测试；跨 HTTP 路由、配置、进程生命周期的变更应补充集成或端到端风格测试。

---

## Code Review Checklist

- 目录位置是否符合职责边界，是否有更合适的现有包可复用。
- 错误是否通过领域错误和 API 映射表达，响应格式是否统一。
- 日志字段是否足够排查问题，且没有泄漏敏感信息。
- 配置是否有默认值、环境变量覆盖、校验和测试。
- 后台 goroutine、文件 watcher、ticker、进程资源是否可停止或清理。
- 测试是否覆盖成功路径、错误路径和关键边界。
- 本地验证是否通过：

```bash
make fmt
make lint
make test
```

发布前或较大变更可直接运行：

```bash
make verify
```
