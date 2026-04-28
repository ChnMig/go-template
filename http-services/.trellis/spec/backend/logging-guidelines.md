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
c.Set("logger", contextLogger)
```

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
- 安全关键配置错误：JWT 密钥缺失、默认密钥、长度不足，ACME/TLS 配置冲突。
- 请求链路：TraceID 中间件在 debug 级别记录请求开始和完成状态码。
- 错误路径：handler 出错时记录领域错误；统一响应函数记录错误响应。
- 日志基础设施：文件缺失重建、轮转失败、监控错误。

---

## What NOT to Log

- 不要记录 JWT 密钥、token 原文、TLS 私钥、证书敏感内容或完整认证头。
- 不要默认记录完整请求 body；`WithRequest` 明确避免主动读取 body。
- 不要在所有成功请求中记录大量参数；成功路径保持 debug 级别和有限字段。
- 不要在 `utils/log` 外重复实现第三方日志重定向。Gin 日志已通过 `NewZapWriterFunc(httplog.GetGinLogger, zapcore.InfoLevel)` 和 `GetGinErrorLogger` 接入独立日志文件。
