# API KNOWLEDGE BASE

**Scope:** `api/`
**Parent:** `../AGENTS.md`

## OVERVIEW

Gin API layer. Owns engine initialization, global middleware order, versioned route aggregation, response envelope, and API-facing DTO/error mapping.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Top-level Gin setup | `router.go` | `NewRouter(Options)`、显式依赖、trusted proxies、static、`/api` |
| Middleware order | `router.go` | Order is part of behavior, not style |
| Trace/logger injection | `middleware/trace-id.go` | Must run before access log and handlers |
| Access log summary | `middleware/access-log.go` | Logs final status in defer; must wrap recovery |
| Panic recovery | `middleware/recovery.go` | Writes unified internal response before access log defer records it |
| Body limit | `middleware/body-limit.go` | Rejects known lengths and converts streamed `MaxBytesError` before commit |
| JWT middleware | `middleware/jwt.go` | Stores decoded claims under `contextkey.JWTData` |
| Rate limiting | `middleware/rate-limit.go` | IP/token/custom key, bounded lazy TTL cleanup without goroutines |
| Pagination | `middleware/page.go` | Uses `config.DefaultPage*`; `-1` disables pagination |
| Unified response | `response/code.go`, `response/format.go` | Always JSON envelope with trace_id/timestamp |
| Version route tree | `app/router.go`, `app/v1/router.go` | `/api -> /v1 -> open/private` |
| Health module | `app/v1/open/health/` | Router, handler, DTO, API error mapping |

## ROUTING CONTRACT

Route registration chain is fixed; do not mount business routes directly in `api/router.go`:

```text
bootstrap -> api.NewRouter(DefaultOptions) -> app.RegisterRoutes -> v1.RegisterRoutes -> open/private -> module
```

- Each aggregation layer exposes `RegisterRoutes(*gin.RouterGroup)` and returns early on nil group.
- `NewRouter` requires the HTTP config snapshot, context/access/error logger providers, Trace ID factory, and route registrar; construction errors must reach bootstrap.
- `InitApi` is compatibility-only and must not replace the error-returning production path.
- Leaf modules expose module-specific registration such as `RegisterOpenRoutes`.
- `private` is a real route boundary even when empty; add private APIs under `api/app/v1/private/<module>`.

## MIDDLEWARE CONTRACT

Global order in `api/router.go`:

```text
TraceID -> AccessLog -> Recovery -> optional CORS -> optional IPRateLimit -> SecurityHeaders -> BodySizeLimit
```

- `TraceID` must stay first so downstream logs/responses can include trace_id.
- Trace IDs are canonical 36-character UUIDv7 values; only canonical incoming UUIDs are reused.
- `AccessLog` must wrap `Recovery`; its defer logs the final status and response size after recovery writes.
- `Recovery` must preserve committed responses; only pre-write application panics use `response.ReturnError(... INTERNAL ...)`.
- CORS preflight must abort with 204 before rate limiting and business handlers.
- Global rate limit and body size are driven by the injected `config.HTTPConfig` snapshot, not mutable package globals.
- Every limiter instance owns a bounded map and uses request-path TTL cleanup; do not add global caches or cleanup goroutines.

## RESPONSE CONTRACT

- API success and errors both return HTTP 200; semantic result lives in JSON `code/status/message/detail/total`.
- All response helpers inject `timestamp` and `trace_id` from context.
- Use `response.ReturnOk`, `ReturnOkWithTotal`, `ReturnSuccess`, `ReturnError`, or `ReturnErrorWithData`.
- Response logs contain only `code/status`; error responses return user-friendly messages and never log `detail`.

## ANTI-PATTERNS

- Do not bypass `app -> v1 -> open/private -> module` route layering.
- Do not reorder `TraceID`, `AccessLog`, `Recovery` without updating tests and documenting why.
- Do not return raw Gin JSON from normal API handlers; use `api/response` envelope.
- Do not expose domain or persistence structs in API responses; create DTOs.
- Do not log query, request bodies, credentials, cookies, panic values, or business error text in global HTTP logs.
- Do not put JWT context keys as string literals; use `utils/contextkey`.

## TEST NOTES

- Use `gin.SetMode(gin.TestMode)` + `httptest`; assert envelope body, not just HTTP status.
- Middleware order changes must update `api/router_test.go::TestInitApiMiddlewareOrder`.
