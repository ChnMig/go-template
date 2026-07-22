package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"

	"http-services/api/response"
	"http-services/config"
	"http-services/utils/contextkey"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
)

func TestRecoveryReturnsUnifiedInternalResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceID(), AccessLog(), Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	const traceID = "0123456789abcdef0123456789abcdef"
	req.Header.Set(contextkey.TraceIDHeader, traceID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", w.Code, http.StatusOK)
	}

	var body struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
		TraceID string `json:"trace_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if body.Code != response.INTERNAL.Code || body.Status != response.INTERNAL.Status {
		t.Fatalf("响应状态 = %d/%s, want %d/%s", body.Code, body.Status, response.INTERNAL.Code, response.INTERNAL.Status)
	}
	if body.TraceID != traceID {
		t.Fatalf("trace_id = %q, want %q", body.TraceID, traceID)
	}
	if body.Message != "服务内部错误" {
		t.Fatalf("message = %q, want 服务内部错误", body.Message)
	}
}

func TestRecoveryDoesNotWriteForAbortedConnections(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, panicErr := range []error{
		http.ErrAbortHandler,
		syscall.EPIPE,
		fmt.Errorf("wrapped reset: %w", syscall.ECONNRESET),
	} {
		router := gin.New()
		router.Use(Recovery())
		router.GET("/panic", func(*gin.Context) { panic(panicErr) })

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/panic?secret=query", nil))
		if w.Body.Len() != 0 {
			t.Fatalf("panic %v wrote response body %q", panicErr, w.Body.String())
		}
	}
}

func TestRecoveryPreservesCommittedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		c.String(http.StatusAccepted, "partial")
		panic(errors.New("sensitive panic value"))
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if w.Code != http.StatusAccepted || w.Body.String() != "partial" {
		t.Fatalf("committed response = %d %q, want %d %q", w.Code, w.Body.String(), http.StatusAccepted, "partial")
	}
}

func TestRecoveryLogDoesNotLeakPanicValueOrQuery(t *testing.T) {
	output := captureRecoveryLogs(t, func() {
		router := gin.New()
		router.Use(TraceID(), Recovery())
		router.GET("/panic", func(*gin.Context) { panic(errors.New("secret-panic-value")) })

		req := httptest.NewRequest(http.MethodGet, "/panic?secret-query=query-value", nil)
		req.Header.Set("Authorization", "Bearer secret-token")
		req.Header.Set("Cookie", "session=secret-cookie")
		router.ServeHTTP(httptest.NewRecorder(), req)
	})

	for _, forbidden := range []string{
		"secret-panic-value", "secret-query", "query-value", "secret-token", "secret-cookie", "raw_query",
	} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("recovery log leaked %q: %s", forbidden, output)
		}
	}
	for _, required := range []string{"HTTP panic recovered", "panic_type", "stack"} {
		if !strings.Contains(output, required) {
			t.Fatalf("recovery log missing %q: %s", required, output)
		}
	}
}

func captureRecoveryLogs(t *testing.T, run func()) string {
	t.Helper()
	oldStdout := os.Stdout
	oldRunModel := config.RunModel
	oldLogConfig := config.CurrentLogConfig()
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	t.Cleanup(func() {
		_ = readPipe.Close()
		_ = writePipe.Close()
	})

	os.Stdout = writePipe
	config.RunModel = config.RunModelDevValue
	config.UpdateLogConfig(config.LogConfig{MaxSize: 50, MaxAge: 30, Level: "info", GinLevel: "info"})
	httplog.SetLogger()
	run()
	_ = httplog.GetGinErrorLogger().Sync()
	_ = httplog.GetLogger().Sync()
	os.Stdout = oldStdout
	config.RunModel = oldRunModel
	config.UpdateLogConfig(oldLogConfig)
	httplog.SetLogger()
	if err := writePipe.Close(); err != nil {
		t.Fatalf("close log pipe: %v", err)
	}
	output, err := io.ReadAll(readPipe)
	if err != nil {
		t.Fatalf("read recovery log: %v", err)
	}
	return string(output)
}

func TestRecoveryWritesResponseBeforeOuterAccessLogDefer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	responseSizeInOuterDefer := -1
	router := gin.New()
	router.Use(TraceID())
	router.Use(func(c *gin.Context) {
		defer func() {
			responseSizeInOuterDefer = c.Writer.Size()
		}()
		c.Next()
	})
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", w.Code, http.StatusOK)
	}
	if responseSizeInOuterDefer <= 0 {
		t.Fatalf("outer access-log-style defer saw response size %d, want a written recovery response", responseSizeInOuterDefer)
	}
}

func TestAccessLogDoesNotBlockRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceID(), AccessLog())
	router.GET("/ok", func(c *gin.Context) {
		response.ReturnSuccess(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok?foo=bar", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get(contextkey.TraceIDHeader) == "" {
		t.Fatalf("未写入 %s 响应头", contextkey.TraceIDHeader)
	}
}
