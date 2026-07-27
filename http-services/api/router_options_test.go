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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewRouterUsesInjectedDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var output bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&output),
		zap.DebugLevel,
	))
	provider := middleware.LoggerProvider(func() *zap.Logger { return logger })
	traceID := uuid.MustParse("018f47a5-7b8c-7c11-8000-123456789abc")
	router, err := NewRouter(Options{
		Server: config.HTTPConfig{MaxBodySize: 10 << 20},
		Loggers: LoggerProviders{
			Context: provider,
			Access:  provider,
			Error:   provider,
		},
		IDFactory: func() (uuid.UUID, error) { return traceID, nil },
		RegisterRoutes: func(group *gin.RouterGroup) {
			group.GET("/injected", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/injected", nil))
	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("response = %d %q", w.Code, w.Body.String())
	}
	if got := w.Header().Get(middleware.TraceIDHeaderKey); got != traceID.String() {
		t.Fatalf("trace ID = %q, want %q", got, traceID.String())
	}
	if !strings.Contains(output.String(), "http.request") || !strings.Contains(output.String(), traceID.String()) {
		t.Fatalf("injected logger did not receive access log: %s", output.String())
	}
}

func TestNewRouterRejectsInvalidOptionsAndTrustedProxy(t *testing.T) {
	provider := middleware.LoggerProvider(func() *zap.Logger { return zap.NewNop() })
	valid := Options{
		Server:         config.HTTPConfig{MaxBodySize: 10 << 20},
		Loggers:        LoggerProviders{Context: provider, Access: provider, Error: provider},
		IDFactory:      func() (uuid.UUID, error) { return uuid.NewV7() },
		RegisterRoutes: func(*gin.RouterGroup) {},
	}

	tests := []struct {
		name   string
		mutate func(*Options)
	}{
		{name: "missing context logger", mutate: func(options *Options) { options.Loggers.Context = nil }},
		{name: "nil context logger", mutate: func(options *Options) {
			options.Loggers.Context = func() *zap.Logger { return nil }
		}},
		{name: "missing access logger", mutate: func(options *Options) { options.Loggers.Access = nil }},
		{name: "missing error logger", mutate: func(options *Options) { options.Loggers.Error = nil }},
		{name: "missing trace factory", mutate: func(options *Options) { options.IDFactory = nil }},
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
