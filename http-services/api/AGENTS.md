# API KNOWLEDGE BASE

**Scope:** `api/`
**Parent:** `../AGENTS.md`

## OVERVIEW

Gin API layer. Owns engine initialization, global middleware order, versioned route aggregation, response envelope, and API-facing DTO/error mapping.

## 更新记录

| 更新时间 | 更新内容（概述） | 对应详细章节 |
| --- | --- | --- |
| 2026-11-15 | 增加严格 JSON、JWT 校验及接口错误和契约同步要求；核对 CORS 原有行为与测试。 | [MIDDLEWARE CONTRACT](#middleware-contract)、[RESPONSE CONTRACT](#response-contract) |

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Top-level Gin setup | `router.go` | `gin.New`、Gin log redirect、trusted proxies、static、`/api` |
| Middleware order | `router.go` | Order is part of behavior, not style |
| Trace/logger injection | `middleware/trace-id.go` | Must run before access log and handlers |
| JWT middleware | `middleware/jwt.go` | Stores decoded claims under `contextkey.JWTData` |
| Rate limiting | `middleware/rate-limit.go` | IP/token/custom key, global cache cleanup on shutdown |
| Pagination | `middleware/page.go` | Uses `config.DefaultPage*`; `-1` disables pagination |
| Unified response | `response/code.go`, `response/format.go` | Always JSON envelope with trace_id/timestamp |
| Version route tree | `app/router.go`, `app/v1/router.go` | `/api -> /v1 -> open/private` |
| Health module | `app/v1/open/health/` | Router, handler, DTO, API error mapping |

## ROUTING CONTRACT

Route registration chain is fixed; do not mount business routes directly in `api/router.go`:

```text
api.InitApi -> app.RegisterRoutes -> v1.RegisterRoutes -> open/private -> module
```

- Each aggregation layer exposes `RegisterRoutes(*gin.RouterGroup)` and returns early on nil group.
- Leaf modules expose module-specific registration such as `RegisterOpenRoutes`.
- `private` is a real route boundary even when empty; add private APIs under `api/app/v1/private/<module>`.

## MIDDLEWARE CONTRACT

Global order in `api/router.go`:

```text
Gin Logger / middleware.Recovery -> TraceID -> optional IPRateLimit -> SecurityHeaders -> BodySizeLimit -> CORS
```

- `gin.New` installs `gin.Logger()` and `middleware.Recovery()`; panic 日志包含请求上下文和堆栈，保持 HTTP 500 与断连处理。
- `TraceID` runs before rate limiting and business handlers, writes the ID to Gin and standard contexts, and uses `http.request.started/completed` with the `status` field.
- `TraceID` 安装被动 `CaptureRequestBody`；参数绑定失败通过 `log.WithRequest` 回退原始请求体，不提前读取、不脱敏，日志文本上限 64 KiB。
- Global rate limit is config-driven: `config.EnableRateLimit`, `GlobalRateLimit`, `GlobalRateBurst`.
- `BodySizeLimit` is config-driven via parsed `config.MaxBodySize`.
- Shutdown must call `middleware.CleanupAllLimiters()` from `main.go`.
- `CheckStrictJSONParam` / `BindStrictJSONParam` 按接口启用，拒绝未知字段和尾随 JSON，并执行 binding 校验；不改变普通 JSON helper 或 Gin 全局开关。失败诊断保留原始输入及具体错误，不能使用部分解码结构体替代。
- `TokenVerify` 通过统一解析入口限定 HS256 并要求有效 `exp`，空签名密钥拒绝；保持 Authorization 原请求方式、claims 的 data 结构和统一响应格式。
- 当前 CORS 预检沿用 HTTP 200 与 `Access-Control-Allow-Headers: *`；关闭 CORS 时，未注册的 OPTIONS 路由返回 404，测试按既有契约断言。

## RESPONSE CONTRACT

- API success and errors both return HTTP 200; semantic result lives in JSON `code/status/message/detail/total`.
- All response helpers inject `timestamp` and `trace_id` from context.
- Use `response.ReturnOk`, `ReturnOkWithTotal`, `ReturnSuccess`, `ReturnError`, or `ReturnErrorWithData`.
- Error responses should log internal context but return user-friendly messages.
- `message` 明确失败字段、条件或环节，内部异常不直接透传；已发布路径、鉴权、DTO、枚举、单位及同步／异步语义属于契约，修改时同步必要文档与测试，完整性要求见 `../docs/README.md`。
- 统一错误响应在请求 context 已取消或响应为 `CANCELLED` 时仅记 Debug；请求超时及活动请求继续按响应 code 记 Warn/Error，不改变 HTTP 200、JSON envelope 与 Abort 行为。
- Error paths use `log.WithRequest` so complete parsed request parameters and the response envelope remain available for troubleshooting; do not add redaction in the shared scaffold.

## ANTI-PATTERNS

- Do not bypass `app -> v1 -> open/private -> module` route layering.
- Do not return raw Gin JSON from normal API handlers; use `api/response` envelope.
- Do not expose domain or persistence structs in API responses; create DTOs.
- Do not put JWT context keys as string literals; use `utils/contextkey`.

## TEST NOTES

- Use `gin.SetMode(gin.TestMode)` + `httptest`; assert envelope body, not just HTTP status.
