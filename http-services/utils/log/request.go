package log

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"http-services/utils/contextkey"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	requestBodyKey     = contextkey.RequestLogBody
	requestLogMaxBytes = 64 << 10
)

// requestBodyCapture 只观察实际读取，不提前消费请求流，也不改变读取错误或 Close。
// 即使验签后恢复 Request.Body，Context 中仍保留原始快照。
type requestBodyCapture struct {
	io.ReadCloser
	data      []byte
	readBytes int64
	eof       bool
	readError bool
}

func (b *requestBodyCapture) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if remaining := requestLogMaxBytes - len(b.data); remaining > 0 {
		b.data = append(b.data, p[:min(n, remaining)]...)
	}
	b.readBytes += int64(n)
	b.eof = b.eof || errors.Is(err, io.EOF)
	b.readError = b.readError || (err != nil && !errors.Is(err, io.EOF))
	return n, err
}

// CaptureRequestBody 在绑定或验签前安装有大小上限的被动采集器。
// 不主动读取 Body，也不改变 BodySizeLimit。
func CaptureRequestBody(c *gin.Context) {
	if c == nil || c.Request == nil || c.Request.Body == nil || c.Request.Body == http.NoBody {
		return
	}
	if _, exists := c.Get(requestBodyKey); exists {
		return
	}
	body := &requestBodyCapture{ReadCloser: c.Request.Body}
	c.Set(requestBodyKey, body)
	c.Request.Body = body
}

// WithRequest 按原值记录请求参数，不脱敏；绑定失败时回退到保留原始类型的请求体。
// 文本快照有大小上限，未读完或截断时显式标记。
func WithRequest(c *gin.Context) *zap.Logger {
	base := FromContext(c)
	if c == nil || c.Request == nil {
		return base
	}
	fields := make([]zap.Field, 0, 8)
	if c.Request.URL != nil && c.Request.URL.RawQuery != "" {
		query := c.Request.URL.RawQuery
		if len(query) > requestLogMaxBytes {
			query = query[:requestLogMaxBytes]
			fields = append(fields, zap.Bool("query_truncated", true))
		}
		fields = append(fields, zap.String("query", query))
	}
	if len(c.Request.PostForm) > 0 {
		fields = append(fields, requestValueFields("form", c.Request.PostForm)...)
	}
	if form := c.Request.MultipartForm; form != nil && len(form.Value) > 0 {
		fields = append(fields, requestValueFields("multipart_form", form.Value)...)
	}
	if len(c.Params) > 0 {
		params := make(map[string]string, len(c.Params))
		for _, param := range c.Params {
			params[param.Key] = param.Value
		}
		fields = append(fields, requestValueFields("path_params", params)...)
	}
	if bound, exists := c.Get(BoundParamsKey); exists && bound != nil {
		fields = append(fields, requestValueFields("params", bound)...)
	} else if value, exists := c.Get(requestBodyKey); exists {
		if body, ok := value.(*requestBodyCapture); ok {
			fields = append(fields, body.logFields(c.Request)...)
		}
	}
	return base.With(fields...)
}

func (b *requestBodyCapture) logFields(req *http.Request) []zap.Field {
	fields := []zap.Field{
		zap.String("content_type", req.Header.Get("Content-Type")),
		zap.Int64("content_length", req.ContentLength),
		zap.Int64("request_body_bytes", b.readBytes),
	}
	if b.readError {
		fields = append(fields, zap.Bool("request_body_read_error", true))
	}
	if !b.eof && (req.ContentLength <= 0 || b.readBytes < req.ContentLength) {
		fields = append(fields, zap.Bool("request_body_incomplete", true))
	}
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(req.Header.Get("Content-Type"), ";")[0]))
	textBody := contentType == "" || contentType == "application/json" || strings.HasSuffix(contentType, "+json") || contentType == "application/x-www-form-urlencoded" || strings.HasPrefix(contentType, "text/")
	status := ""
	switch {
	case len(b.data) == 0:
		status = "unread"
		if b.eof {
			status = "empty"
		}
	case !textBody:
		status = "unsupported_content_type"
	case b.readBytes > requestLogMaxBytes:
		status = "truncated"
		fields = append(fields, zap.ByteString("request_body", b.data), zap.Bool("request_body_truncated", true))
	case json.Valid(b.data):
		fields = append(fields, zap.Reflect("request_body", json.RawMessage(b.data)))
	case contentType == "application/x-www-form-urlencoded":
		if values, err := url.ParseQuery(string(b.data)); err == nil {
			fields = append(fields, requestValueFields("request_body", values)...)
		} else {
			status = "invalid_form"
			fields = append(fields, zap.ByteString("request_body", b.data))
		}
	default:
		status = "invalid_json"
		fields = append(fields, zap.ByteString("request_body", b.data))
	}
	if status != "" {
		fields = append(fields, zap.String("request_body_status", status))
	}
	return fields
}

func requestValueFields(name string, value any) []zap.Field {
	data, err := json.Marshal(value)
	if err != nil {
		return []zap.Field{zap.String(name+"_status", "unserializable")}
	}
	if len(data) > requestLogMaxBytes {
		return []zap.Field{zap.ByteString(name, data[:requestLogMaxBytes]), zap.Bool(name+"_truncated", true)}
	}
	return []zap.Field{zap.Reflect(name, json.RawMessage(data))}
}
