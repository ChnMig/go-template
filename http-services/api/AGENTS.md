# API KNOWLEDGE BASE

**Scope:** `api/`
**Parent:** `../AGENTS.md`

## OVERVIEW

Gin API layer. Owns engine initialization, global middleware order, versioned route aggregation, response envelope, and API-facing DTO/error mapping.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Top-level Gin setup | `router.go` | `gin.New`、Gin log redirect、trusted proxies、static、`/api` |
| Middleware order | `router.go` | Order is part of behavior, not style |
| Trace/logger injection | `middleware/trace-id.go` | Must run before access log and handlers |
| Access log summary | `middleware/access-log.go` | Logs final status in defer; must wrap recovery |
| Panic recovery | `middleware/recovery.go` | Writes unified internal response before access log defer records it |
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
TraceID -> AccessLog -> Recovery -> optional IPRateLimit -> SecurityHeaders -> BodySizeLimit -> CORS
```

- `TraceID` must stay first so downstream logs/responses can include trace_id.
- `AccessLog` must wrap `Recovery`; its defer logs the final status and response size after recovery writes.
- `Recovery` must use `response.ReturnError(... INTERNAL ...)` so panic responses keep the project envelope.
- Global rate limit is config-driven: `config.EnableRateLimit`, `GlobalRateLimit`, `GlobalRateBurst`.
- `BodySizeLimit` is config-driven via parsed `config.MaxBodySize`.
- Shutdown must call `middleware.CleanupAllLimiters()` from `main.go`.

## RESPONSE CONTRACT

- API success and errors both return HTTP 200; semantic result lives in JSON `code/status/message/detail/total`.
- All response helpers inject `timestamp` and `trace_id` from context.
- Use `response.ReturnOk`, `ReturnOkWithTotal`, `ReturnSuccess`, `ReturnError`, or `ReturnErrorWithData`.
- Error responses should log internal context but return user-friendly messages.

## ANTI-PATTERNS

- Do not bypass `app -> v1 -> open/private -> module` route layering.
- Do not reorder `TraceID`, `AccessLog`, `Recovery` without updating tests and documenting why.
- Do not return raw Gin JSON from normal API handlers; use `api/response` envelope.
- Do not expose domain or persistence structs in API responses; create DTOs.
- Do not log request bodies in access logs; existing access-log test guards this.
- Do not put JWT context keys as string literals; use `utils/contextkey`.

## TEST NOTES

- Use `gin.SetMode(gin.TestMode)` + `httptest`; assert envelope body, not just HTTP status.
- Middleware order changes must update `api/router_test.go::TestInitApiMiddlewareOrder`.
