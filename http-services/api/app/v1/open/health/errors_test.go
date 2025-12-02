package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"http-services/api/response"
	domain "http-services/domain/health"
)

// errorResponse 用于解析错误响应结构（示例）
type errorResponse struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
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
