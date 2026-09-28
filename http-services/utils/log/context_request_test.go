package log

import (
	"net/http"
	"testing"

	"http-services/utils/contextkey"
)

func TestFromContextRetainsStandardTraceWithBaseLogger(t *testing.T) {
	c, output := requestLogContext(t, http.MethodGet, "/trace", "")
	c.Set(contextkey.TraceID, "")
	c.Request = c.Request.WithContext(WithTraceID(c.Request.Context(), "standard-trace"))
	FromContext(c).Info("standard context fallback")
	entry := requestLogEntry(t, output)
	if entry["trace_id"] != "standard-trace" || entry["path"] != "/trace" || entry["method"] != "GET" {
		t.Fatalf("standard context or request metadata lost: %s", output.String())
	}
}
