# Error Handling

> 本项目错误处理约定。

---

## Overview

错误处理分为三层：

1. `domain/` 定义领域错误，只表达业务语义。
2. `api/app/...` 将领域错误映射为统一响应码、业务码和用户可见文案。
3. `api/response` 负责统一 JSON 响应格式，并始终通过 HTTP 200 返回业务结果。

返回客户端的响应结构包含 `code`、`status`、`description`、可选 `message`、`trace_id`、`timestamp`、`detail`、`total`。`trace_id` 由 `api/middleware/trace-id.go` 写入 Gin context，context key 统一来自 `utils/contextkey`，响应函数自动带出。

---

## Error Types

领域错误使用标准库 `errors.New` 定义，并通过 `errors.Is` 判断。领域错误不包含 HTTP 状态码、不包含前端文案。

真实示例：

```go
// domain/health/errors.go
var (
	ErrServiceNotReady  = errors.New("service not ready")
	ErrServiceUnhealthy = errors.New("service unhealthy")
)
```

模块级业务错误码定义在 API 模块内部，避免污染全局错误码空间。

```go
// api/app/v1/open/health/errors.go
const (
	CodeHealthServiceNotReady = 10001
	CodeHealthServiceUnhealthy = 10002
)
```

---

## Error Handling Patterns

handler 调用领域层后立即处理错误，先记录必要上下文，再调用模块内错误映射函数返回。

```go
status, err := domain.GetStatus()
if err != nil {
	log.WithRequest(c).Error("健康检查失败", zap.Error(err))
	ReturnDomainError(c, err)
	return
}
```

错误映射使用 `errors.Is`，已知错误映射到具体业务码，未知错误统一按 `response.INTERNAL`。

```go
switch {
case errors.Is(err, domain.ErrServiceNotReady):
	data := response.FAILED_PRECONDITION
	data.Code = CodeHealthServiceNotReady
	response.ReturnError(c, data, "服务尚未就绪，请稍后重试")
default:
	response.ReturnError(c, response.INTERNAL, "服务内部错误")
}
```

参数绑定错误通过 `middleware.CheckParam` 统一返回 `response.INVALID_ARGUMENT`，成功绑定后将参数挂到 context，供日志按需读取。

---

## API Error Responses

所有 API 响应当前都使用 HTTP 200，业务成功或失败由 JSON 中的 `code` 和 `status` 表达。不要在 handler 中手写 `c.JSON` 拼装错误响应，统一使用 `response.ReturnError`、`response.ReturnErrorWithData`、`response.ReturnOk`、`response.ReturnOkWithTotal`、`response.ReturnSuccess`。

真实响应结构定义：

```go
type responseData struct {
	Code        int         `json:"code"`
	Status      string      `json:"status"`
	Description string      `json:"description"`
	Message     string      `json:"message,omitempty"`
	TraceID     string      `json:"trace_id,omitempty"`
	Timestamp   int64       `json:"timestamp"`
	Detail      interface{} `json:"detail,omitempty"`
	Total       *int        `json:"total,omitempty"`
}
```

认证失败使用 `response.UNAUTHENTICATED`，限流失败使用 `response.RESOURCE_EXHAUSTED`，请求体过大与绑定失败当前使用 `response.INVALID_ARGUMENT`。

---

## Common Mistakes

- 不要让领域层返回面向前端的中文提示；中文提示在 API 错误映射层生成。
- 不要绕过 `api/response` 直接返回不一致的 JSON。
- 不要在成功响应中记录 error 级别日志；当前成功响应使用 debug，错误响应使用 error。
- 不要吞掉未知错误；未知错误至少映射为 `response.INTERNAL` 并记录日志。
