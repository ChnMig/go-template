package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// statusResponse 用于解析健康检查接口的统一响应结构
type statusResponse struct {
	Code   int       `json:"code"`
	Status string    `json:"status"`
	Detail StatusDTO `json:"detail"`
}

// 合并后健康检查接口的测试
func TestStatus(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", Status)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status() status code = %d, want %d", w.Code, http.StatusOK)
	}

	var resp statusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Code != 200 {
		t.Errorf("Status() response code = %d, want 200", resp.Code)
	}

	if resp.Status != "OK" {
		t.Errorf("Status() wrapper status = %s, want 'OK'", resp.Status)
	}

	if resp.Detail.Status != "ok" {
		t.Errorf("Status() detail.status = %s, want 'ok'", resp.Detail.Status)
	}

	if !resp.Detail.Ready {
		t.Errorf("Status() detail.ready = %v, want true", resp.Detail.Ready)
	}

	if resp.Detail.Uptime == "" {
		t.Errorf("Status() missing uptime field")
	}

	if resp.Detail.Timestamp == 0 {
		t.Errorf("Status() missing timestamp field")
	}
}
