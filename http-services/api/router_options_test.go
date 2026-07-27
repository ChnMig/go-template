package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"http-services/api/middleware"
	"http-services/config"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNewRouterUsesGlobalZapLoggers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var output bytes.Buffer
	require.NoError(t, httplog.SetLogger(config.LogConfig{
		MaxSize: config.LogFileSizeMB(1), MaxAge: config.LogRetentionDays(1),
		Level: config.LogLevelDebug, GinLevel: config.LogLevelDebug,
	}, true, &output))
	t.Cleanup(func() { require.NoError(t, httplog.Close()) })
	const traceID = "018f47a5-7b8c-7c11-8000-123456789abc"
	router, err := NewRouter(Options{
		Server: config.HTTPConfig{MaxBodySize: 10 << 20},
		RegisterRoutes: func(group *gin.RouterGroup) {
			group.GET("/global", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/global", nil)
	request.Header.Set(middleware.TraceIDHeader, traceID)
	router.ServeHTTP(w, request)
	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("response = %d %q", w.Code, w.Body.String())
	}
	if got := w.Header().Get(middleware.TraceIDHeaderKey); got != traceID {
		t.Fatalf("trace ID = %q, want %q", got, traceID)
	}
	if !strings.Contains(output.String(), "http.request") || !strings.Contains(output.String(), traceID) {
		t.Fatalf("global logger did not receive access log: %s", output.String())
	}
}

func TestNewRouterRejectsInvalidOptionsAndTrustedProxy(t *testing.T) {
	valid := Options{
		Server:         config.HTTPConfig{MaxBodySize: 10 << 20},
		RegisterRoutes: func(*gin.RouterGroup) {},
	}

	tests := []struct {
		name   string
		mutate func(*Options)
	}{
		{name: "missing registrar", mutate: func(options *Options) { options.RegisterRoutes = nil }},
		{name: "invalid trusted proxy", mutate: func(options *Options) {
			options.Server.TrustedProxies = []string{"not-a-proxy"}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := valid
			test.mutate(&options)
			if _, err := NewRouter(options); !errors.Is(err, ErrInvalidOptions) {
				t.Fatalf("NewRouter() error = %v, want ErrInvalidOptions", err)
			}
		})
	}
}
