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

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if status, ok := response["status"].(string); !ok || status != "ok" {
		t.Errorf("Status() status = %v, want 'ok'", response["status"])
	}

	if ready, ok := response["ready"].(bool); !ok || !ready {
		t.Errorf("Status() ready = %v, want true", response["ready"])
	}

	if _, ok := response["uptime"].(string); !ok {
		t.Errorf("Status() missing uptime field")
	}
}
