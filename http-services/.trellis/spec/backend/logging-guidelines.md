# Logging Guidelines

> 本项目日志约定。

---

## Overview

项目统一使用 `go.uber.org/zap`。`utils/log` 封装全局业务 logger、Gin 独立 logger、日志文件轮转、日志文件缺失监控和请求上下文 logger。业务代码优先通过 `log.FromContext(c)` 或 `log.WithRequest(c)` 获取 logger，不直接长期持有全局 logger 实例。

日志初始化顺序在 `main.go` 中固定：加载配置、设置运行模式、初始化日志、启动日志监控、启动配置热重载。`utils/log` 的 `init()` 不主动初始化 logger，避免测试或包初始化阶段创建日志文件。

---

## Log Levels

日志级别来自配置：

- `log.level`：业务日志级别，默认 `info`。
- `log.gin_level`：Gin access/error 日志级别；为空时沿用业务日志级别。

真实代码：

```go
businessLevel := parseLogLevel(config.LogLevel)
ginLevel := businessLevel
if strings.TrimSpace(config.GinLogLevel) != "" {
	ginLevel = parseLogLevel(config.GinLogLevel)
}
```

使用语义：

- `Debug`：请求开始/结束、成功响应、开发排查细节。
- `Info`：服务启动、配置加载、PID 文件写入、正常生命周期事件。
- `Warn`：非致命但需要关注的问题，例如日志文件缺失后重建、未识别日志级别回退。
- `Error`：业务处理失败、领域错误、服务异常退出、日志监控错误。
- `Fatal`：关键配置缺失或不安全，服务不能继续运行。

---

## Structured Logging

开发模式输出到终端，生产模式输出 JSON 文件并使用 ISO8601 时间。业务日志和 Gin 日志在生产模式下分文件：

- 业务日志：`log/<程序名>.log`
- Gin 日志：`log/<程序名>.gin.log`

`TraceID` 中间件会创建带上下文的 logger 并存入 Gin context，字段包括 `trace_id`、`method`、`path`、`client_ip`。

```go
contextLogger := zap.L().With(
	zap.String("trace_id", traceID),
	zap.String("method", c.Request.Method),
	zap.String("path", c.Request.URL.Path),
	zap.String("client_ip", c.ClientIP()),
)
c.Set(contextkey.Logger, contextLogger)
```

Gin context key 统一定义在 `utils/contextkey`，新增中间件或响应逻辑不要散落硬编码 `trace_id`、`logger`、`jwtData`、`__bound_params__` 等字符串。

业务 handler 默认使用：

```go
l := log.FromContext(c)
l.Debug("健康检查开始")
```

只有需要排查问题时使用 `log.WithRequest(c)`，它会额外带上 method、path、query、form、multipart 表单、路径参数和 `middleware.CheckParam` 绑定后的业务参数。

---

## What to Log

- 服务生命周期：启动、停止信号、异常退出、优雅关闭失败。
- 配置生命周期：配置文件加载、热重载后的关键配置字段。
- 安全关键配置错误：JWT 密钥缺失、默认密钥、长度不足。
- 请求链路：TraceID 中间件在 debug 级别记录请求开始和完成状态码，AccessLog 中间件记录结构化 method、path、status、latency、client_ip、user_agent、trace_id 和 error 字段。
- 错误路径：handler 出错时记录领域错误；统一响应函数记录错误响应。
- 日志基础设施：文件缺失重建、轮转失败、监控错误。

---

## What NOT to Log

- 不要记录 JWT 密钥、token 原文、证书私钥、反向代理敏感配置或完整认证头。
- 不要默认记录完整请求 body；`WithRequest` 明确避免主动读取 body。
- 不要在所有成功请求中记录大量参数；成功路径保持 debug 级别和有限字段。
- 不要在 `utils/log` 外重复实现第三方日志重定向。Gin 日志已通过 `NewZapWriterFunc(httplog.GetGinLogger, zapcore.InfoLevel)` 和 `GetGinErrorLogger` 接入独立日志文件。

---

## Scenario: Gin Recovery 与 AccessLog

### 1. Scope / Trigger

修改 Gin 全局中间件、panic 处理或请求日志时必须遵循本场景。目标是避免 Gin 默认 recovery 返回非统一响应，避免成功请求记录 body 或大体积参数。

### 2. Signatures

- `middleware.Recovery() gin.HandlerFunc`
- `middleware.AccessLog() gin.HandlerFunc`
- `middleware.TraceID() gin.HandlerFunc`
- `response.ReturnError(c, response.INTERNAL, "服务内部错误")`

### 3. Contracts

- `Recovery` 捕获 panic 后必须返回 `api/response` 统一响应，HTTP 状态码仍为 200。
- `AccessLog` 必须记录 `method`、`path`、`raw_query`、`status`、`latency`、`client_ip`、`user_agent`、`trace_id`、`error` 字段。
- `trace_id`、`logger`、`jwtData`、`__bound_params__` 等 Gin context key 必须来自 `utils/contextkey`。

### 4. Validation & Error Matrix

- handler panic -> Gin error logger 记录 panic、请求上下文和 stack -> 客户端收到 `INTERNAL` 统一响应。
- 成功请求 -> Gin access logger 记录结构化摘要字段 -> 不记录 body、表单或绑定参数。
- 缺少上游 trace header -> `TraceID` 生成新 ID 并写入响应 header。

### 5. Good/Base/Bad Cases

- Good: `gin.New()` 后显式挂载 `TraceID()`、`AccessLog()`、`Recovery()`，确保 panic 先被 recovery 写入统一响应，再由 access log 记录最终状态。
- Base: 业务 handler 用 `log.FromContext(c)`，错误排查路径才用 `log.WithRequest(c)`。
- Bad: 使用 `gin.Default()`，或在成功 access log 中记录完整 body。

### 6. Tests Required

- recovery 测试断言 panic 后 HTTP 200、响应体 code/status 为 `INTERNAL`，且带出 `trace_id`。
- access log 测试至少覆盖中间件不阻断正常请求，并验证 `TraceID` 响应头存在。
- context key 改动要更新 JWT、response、参数绑定和日志相关测试引用。

### 7. Wrong vs Correct

#### Wrong

```go
router := gin.Default()
```

#### Correct

```go
router := gin.New()
router.Use(middleware.TraceID())
router.Use(middleware.AccessLog())
router.Use(middleware.Recovery())
```
