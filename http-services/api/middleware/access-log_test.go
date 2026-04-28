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
	oldLogLevel := config.LogLevel
	oldGinLogLevel := config.GinLogLevel

	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	restored := false
	restoreGlobals := func() {
		os.Stdout = oldStdout
		config.RunModel = oldRunModel
		config.LogLevel = oldLogLevel
		config.GinLogLevel = oldGinLogLevel
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
	config.LogLevel = "info"
	config.GinLogLevel = "info"
	httplog.SetLogger()

	router := gin.New()
	router.Use(TraceID(), AccessLog())
	router.POST("/ok", func(c *gin.Context) {
		response.ReturnSuccess(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/ok?foo=bar", bytes.NewBufferString("secret-body"))
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
		`"raw_query": "foo=bar"`,
		`"status": 200`,
		`"latency":`,
		`"client_ip":`,
		`"user_agent": "access-log-test"`,
		`"trace_id":`,
		`"error": ""`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("access log output missing %s: %s", want, output)
		}
	}
	if strings.Contains(output, "secret-body") {
		t.Fatalf("access log output should not contain request body: %s", output)
	}
}
