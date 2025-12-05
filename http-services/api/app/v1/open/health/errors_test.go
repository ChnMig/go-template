package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"http-services/api/response"
	domain "http-services/domain/health"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// errorResponse 用于解析错误响应结构（示例）
type errorResponse struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// testCore 用于收集日志字段，验证 ReturnDomainError 是否正确记录错误日志
type testCore struct {
	entries []zapcore.Entry
	fields  [][]zap.Field
}

func (c *testCore) Enabled(level zapcore.Level) bool {
	return true
}

func (c *testCore) With(fields []zap.Field) zapcore.Core {
	c.fields = append(c.fields, fields)
	return c
}

func (c *testCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(ent.Level) {
		return ce
	}
	return ce.AddCore(ent, c)
}

func (c *testCore) Write(ent zapcore.Entry, fields []zap.Field) error {
	c.entries = append(c.entries, ent)
	c.fields = append(c.fields, fields)
	return nil
}

func (c *testCore) Sync() error {
	return nil
}

// TestReturnDomainError_ServiceNotReady 示例：验证 ErrServiceNotReady 映射到自定义业务码 + FAILED_PRECONDITION
func TestReturnDomainError_ServiceNotReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ReturnDomainError(c, domain.ErrServiceNotReady)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Code != CodeHealthServiceNotReady {
		t.Errorf("Code = %d, want %d", resp.Code, CodeHealthServiceNotReady)
	}
	if resp.Status != response.FAILED_PRECONDITION.Status {
		t.Errorf("Status = %s, want %s", resp.Status, response.FAILED_PRECONDITION.Status)
	}
	if resp.Message == "" {
		t.Errorf("Message should not be empty")
	}
}

// TestReturnDomainError_ServiceUnhealthy 示例：验证 ErrServiceUnhealthy 映射到自定义业务码 + UNAVAILABLE
func TestReturnDomainError_ServiceUnhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ReturnDomainError(c, domain.ErrServiceUnhealthy)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Code != CodeHealthServiceUnhealthy {
		t.Errorf("Code = %d, want %d", resp.Code, CodeHealthServiceUnhealthy)
	}
	if resp.Status != response.UNAVAILABLE.Status {
		t.Errorf("Status = %s, want %s", resp.Status, response.UNAVAILABLE.Status)
	}
	if resp.Message == "" {
		t.Errorf("Message should not be empty")
	}
}

// TestReturnDomainError_Logging 确保 ReturnDomainError 会使用 zap 记录错误日志
func TestReturnDomainError_Logging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core := &testCore{}
	logger := zap.New(core)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// 在 gin.Context 中注入 logger，模拟中间件行为
	c.Set("logger", logger)

	testErr := errors.New("domain error")
	ReturnDomainError(c, testErr)

	if len(core.entries) == 0 {
		t.Fatalf("expected at least one log entry, got none")
	}

	found := false
	for i, ent := range core.entries {
		if ent.Level != zap.ErrorLevel {
			continue
		}
		for _, f := range core.fields[i] {
			if f.Key == "error" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("expected error field in logged entries")
	}
}
