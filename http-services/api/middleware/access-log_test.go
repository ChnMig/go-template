package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"http-services/api/response"
	"http-services/config"
	httplog "http-services/utils/log"

	"github.com/gin-gonic/gin"
)

func TestAccessLogWritesStructuredSummaryFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldStdout := os.Stdout
	oldRunModel := config.RunModel
	oldLogConfig := config.CurrentLogConfig()

	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	restored := false
	restoreGlobals := func() {
		os.Stdout = oldStdout
		config.RunModel = oldRunModel
		config.UpdateLogConfig(oldLogConfig)
		httplog.SetLogger()
	}
	restore := func() {
		if restored {
			return
		}
		restored = true
		restoreGlobals()
		_ = readPipe.Close()
		_ = writePipe.Close()
	}
	t.Cleanup(restore)

	os.Stdout = writePipe
	config.RunModel = config.RunModelDevValue
	config.UpdateLogConfig(config.LogConfig{MaxSize: 50, MaxAge: 30, Level: "info", GinLevel: "info"})
	httplog.SetLogger()

	router := gin.New()
	router.Use(TraceID(), AccessLog())
	router.POST("/ok", func(c *gin.Context) {
		response.ReturnSuccess(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/ok?secret_query=query-value", bytes.NewBufferString("secret-body"))
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Cookie", "session=secret-cookie")
	req.Header.Set("User-Agent", "access-log-test")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	_ = httplog.GetGinLogger().Sync()
	restoreGlobals()
	_ = writePipe.Close()
	outBytes, readErr := io.ReadAll(readPipe)
	if readErr != nil {
		t.Fatalf("read access log output: %v", readErr)
	}
	restored = true
	_ = readPipe.Close()

	output := string(outBytes)
	for _, want := range []string{
		`"method": "POST"`,
		`"path": "/ok"`,
		`"status": 200`,
		`"response_bytes":`,
		`"latency":`,
		`"client_ip":`,
		`"trace_id":`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("access log output missing %s: %s", want, output)
		}
	}
	for _, forbidden := range []string{
		"secret_query", "query-value", "secret-body", "secret-token", "secret-cookie", "access-log-test", "raw_query",
	} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("access log output leaked %q: %s", forbidden, output)
		}
	}
}
