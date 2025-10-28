package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReturnOk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]string{"message": "test"}
	ReturnOk(c, testData)

	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp responseData
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != 200 {
		t.Errorf("Expected code 200, got %d", resp.Code)
	}

	if resp.Status != "OK" {
		t.Errorf("Expected status 'OK', got '%s'", resp.Status)
	}

	if resp.Timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}

	if resp.Detail == nil {
		t.Error("Expected detail to be set")
	}
}

func TestReturnOkWithTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := []string{"item1", "item2"}
	total := 100
	ReturnOkWithTotal(c, total, testData)

	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp responseData
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Total == nil {
		t.Fatal("Expected total to be set")
	}

	if *resp.Total != total {
		t.Errorf("Expected total %d, got %d", total, *resp.Total)
	}
}

func TestReturnError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errorMsg := "Invalid parameter"
	ReturnError(c, INVALID_ARGUMENT, errorMsg)

	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp responseData
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != 400 {
		t.Errorf("Expected code 400, got %d", resp.Code)
	}

	if resp.Status != "INVALID_ARGUMENT" {
		t.Errorf("Expected status 'INVALID_ARGUMENT', got '%s'", resp.Status)
	}

	if resp.Message != errorMsg {
		t.Errorf("Expected message '%s', got '%s'", errorMsg, resp.Message)
	}

	if resp.Timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}
}

func TestReturnSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ReturnSuccess(c)

	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp responseData
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != 200 {
		t.Errorf("Expected code 200, got %d", resp.Code)
	}

	if resp.Status != "OK" {
		t.Errorf("Expected status 'OK', got '%s'", resp.Status)
	}

	if resp.Detail != nil {
		t.Error("Expected detail to be nil")
	}
}

func TestReturnErrorWithData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testDetail := map[string]string{"field": "username", "error": "required"}
	ReturnErrorWithData(c, INVALID_ARGUMENT, testDetail)

	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp responseData
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != 400 {
		t.Errorf("Expected code 400, got %d", resp.Code)
	}

	if resp.Detail == nil {
		t.Error("Expected detail to be set")
	}
}

// 测试所有预定义的错误码
func TestAllErrorCodes(t *testing.T) {
	testCases := []struct {
		name         string
		errorData    responseData
		expectedCode int
	}{
		{"OK", OK, 200},
		{"INVALID_ARGUMENT", INVALID_ARGUMENT, 400},
		{"FAILED_PRECONDITION", FAILED_PRECONDITION, 400},
		{"OUT_OF_RANGE", OUT_OF_RANGE, 400},
		{"UNAUTHENTICATED", UNAUTHENTICATED, 401},
		{"PERMISSION_DENIED", PERMISSION_DENIED, 403},
		{"NOT_FOUND", NOT_FOUND, 404},
		{"ABORTED", ABORTED, 409},
		{"ALREADY_EXISTS", ALREADY_EXISTS, 409},
		{"RESOURCE_EXHAUSTED", RESOURCE_EXHAUSTED, 429},
		{"CANCELLED", CANCELLED, 499},
		{"DATA_LOSS", DATA_LOSS, 500},
		{"UNKNOWN", UNKNOWN, 500},
		{"INTERNAL", INTERNAL, 500},
		{"NOT_IMPLEMENTED", NOT_IMPLEMENTED, 501},
		{"UNAVAILABLE", UNAVAILABLE, 503},
		{"DEADLINE_EXCEEDED", DEADLINE_EXCEEDED, 504},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errorData.Code != tc.expectedCode {
				t.Errorf("%s: expected code %d, got %d", tc.name, tc.expectedCode, tc.errorData.Code)
			}
			if tc.errorData.Status == "" {
				t.Errorf("%s: status should not be empty", tc.name)
			}
			if tc.errorData.Description == "" {
				t.Errorf("%s: description should not be empty", tc.name)
			}
		})
	}
}
