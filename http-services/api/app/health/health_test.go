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

func TestHealth(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查状态码
	if w.Code != http.StatusOK {
		t.Errorf("Health() status code = %d, want %d", w.Code, http.StatusOK)
	}

	// 解析响应
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 检查响应内容
	if status, ok := response["status"].(string); !ok || status != "ok" {
		t.Errorf("Health() status = %v, want 'ok'", response["status"])
	}
}

func TestReady(t *testing.T) {
	router := setupTestRouter()
	router.GET("/ready", Ready)

	req, _ := http.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查状态码
	if w.Code != http.StatusOK {
		t.Errorf("Ready() status code = %d, want %d", w.Code, http.StatusOK)
	}

	// 解析响应
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 检查响应内容
	if status, ok := response["status"].(string); !ok || status != "ready" {
		t.Errorf("Ready() status = %v, want 'ready'", response["status"])
	}

	// 检查 uptime 字段存在
	if _, ok := response["uptime"].(string); !ok {
		t.Errorf("Ready() missing uptime field")
	}
}

func TestReadyUptime(t *testing.T) {
	router := setupTestRouter()
	router.GET("/ready", Ready)

	// 第一次请求
	req1, _ := http.NewRequest("GET", "/ready", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	uptime1 := response1["uptime"].(string)

	// 等待一小段时间
	// time.Sleep(10 * time.Millisecond)

	// 第二次请求
	req2, _ := http.NewRequest("GET", "/ready", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	uptime2 := response2["uptime"].(string)

	// uptime 应该是字符串格式
	if uptime1 == "" || uptime2 == "" {
		t.Error("Ready() uptime should not be empty")
	}

	t.Logf("First uptime: %s", uptime1)
	t.Logf("Second uptime: %s", uptime2)
}
