# API KNOWLEDGE BASE

**Scope:** `api/`
**Parent:** `../AGENTS.md`

## OVERVIEW

Gin API layer. Owns engine initialization, global middleware order, versioned route aggregation, response envelope, and API-facing DTO/error mapping.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Top-level Gin setup | `router.go` | `gin.Default`、Gin log redirect、trusted proxies、static、`/api` |
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
Gin Logger/Recovery -> TraceID -> optional IPRateLimit -> SecurityHeaders -> BodySizeLimit -> CORS
```

- `gin.Default` provides the framework access logger and panic recovery.
- `TraceID` runs before rate limiting and business handlers, writes the ID to Gin and standard contexts, and uses `http.request.started/completed` with the `status` field.
- Global rate limit is config-driven: `config.EnableRateLimit`, `GlobalRateLimit`, `GlobalRateBurst`.
- `BodySizeLimit` is config-driven via parsed `config.MaxBodySize`.
- Shutdown must call `middleware.CleanupAllLimiters()` from `main.go`.

## RESPONSE CONTRACT

- API success and errors both return HTTP 200; semantic result lives in JSON `code/status/message/detail/total`.
- All response helpers inject `timestamp` and `trace_id` from context.
- Use `response.ReturnOk`, `ReturnOkWithTotal`, `ReturnSuccess`, `ReturnError`, or `ReturnErrorWithData`.
- Error responses should log internal context but return user-friendly messages.
- Error paths use `log.WithRequest` so complete parsed request parameters and the response envelope remain available for troubleshooting; do not add redaction in the shared scaffold.

## ANTI-PATTERNS

- Do not bypass `app -> v1 -> open/private -> module` route layering.
- Do not return raw Gin JSON from normal API handlers; use `api/response` envelope.
- Do not expose domain or persistence structs in API responses; create DTOs.
- Do not put JWT context keys as string literals; use `utils/contextkey`.

## TEST NOTES

- Use `gin.SetMode(gin.TestMode)` + `httptest`; assert envelope body, not just HTTP status.
