package log

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// testCore 是一个用于测试的 zapcore.Core 实现，用来收集日志字段。
type testCore struct {
	fields []zap.Field
}

func (c *testCore) Enabled(level zapcore.Level) bool {
	return true
}

func (c *testCore) With(fields []zap.Field) zapcore.Core {
	c.fields = append(c.fields, fields...)
	return c
}

func (c *testCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(ent.Level) {
		return ce
	}
	return ce.AddCore(ent, c)
}

func (c *testCore) Write(ent zapcore.Entry, fields []zap.Field) error {
	c.fields = append(c.fields, fields...)
	return nil
}

func (c *testCore) Sync() error {
	return nil
}

// TestWithRequest 确认 WithRequest 会把查询参数、表单参数和路径参数附加到日志字段中。
func TestWithRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core := &testCore{}
	baseLogger := zap.New(core)
	zap.ReplaceGlobals(baseLogger)
	defer zap.ReplaceGlobals(zap.NewNop())

	c, _ := gin.CreateTestContext(nil)
	req := &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Path:     "/api/v1/open/health/123",
			RawQuery: "q=hello&debug=1",
		},
		PostForm: url.Values{
			"name": {"tester"},
		},
	}
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "123"},
	}

	// 在 context 中设置基础 logger，模拟 RequestID 中间件行为
	c.Set("logger", baseLogger)

	logger := WithRequest(c)
	logger.Info("test message")

	assertFieldExists := func(key string) {
		for _, f := range core.fields {
			if f.Key == key {
				return
			}
		}
		t.Fatalf("expected field %q in log fields", key)
	}

	assertFieldExists("method")
	assertFieldExists("path")
	assertFieldExists("query")
	assertFieldExists("form")
	assertFieldExists("path_params")
}
