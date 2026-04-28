# Quality check middleware and TLS cleanup

## Goal

Review the current uncommitted changes in `/Users/chenming/work/go-template/http-services`, fix concrete quality issues directly, and verify the result without reverting other participants' work.

## Requirements

* Verify middleware order for `Recovery`, `AccessLog`, and `TraceID` so a handler panic is recovered into the unified response before the access log records the request. Fix the order if access logging currently runs before recovery completes.
* Ensure panic responses use the `api/response` unified response format, the project's HTTP 200 response strategy, and include `trace_id`.
* Ensure access logs are structured and include `method`, `path`, `raw_query`, `status`, `latency`, `client_ip`, `user_agent`, `trace_id`, and `error`; do not log request or response bodies or large parameter payloads.
* Verify `PageQuery`, `ParsePageQuery`, `GetPage`, and `GetPageSize` behavior: `page=-1` or `page_size=-1` disables pagination, invalid values fall back to defaults, and offset/limit are correct.
* Ensure context keys are unified through `utils/contextkey`, no import cycle is introduced, and hard-coded context key usage is not left behind. Logging field name strings are allowed.
* Ensure TLS/ACME runtime behavior, configuration, examples, README, `.agentdocs`, and `.trellis/spec` no longer expose usable built-in TLS/ACME configuration such as `server.enable_acme` or `server.enable_tls`; only retain the constraint that HTTPS/TLS is terminated by a reverse proxy.
* Do not restore configurable CORS behavior or `CorssDomainHandler`.

## Acceptance Criteria

* Relevant tests cover middleware ordering and panic recovery/access logging behavior.
* `GOFLAGS=-mod=readonly make verify` is run, or any failure is reported with concrete cause.
* `git diff --check` passes.
* Final report lists fixed findings, unfixed findings, verification results, and modified files.
