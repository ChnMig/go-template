package contextkey

import "context"

const (
	// TraceIDHeader 是请求追踪 ID 的 HTTP header 名称。
	TraceIDHeader = "X-Trace-ID"
	// TraceID 是 Gin context 中存放请求追踪 ID 的 key。
	TraceID = "trace_id"
	// Logger 是 Gin context 中存放请求上下文 logger 的 key。
	Logger = "logger"
	// JWTData 是 Gin context 中存放 JWT 解密数据的 key。
	JWTData = "jwtData"
	// BoundParams 是 Gin context 中存放已绑定业务参数的 key。
	BoundParams = "__bound_params__"
	// RequestBody 是 Gin context 中存放已消费请求体副本的 key。
	RequestBody = "__request_body__"
)

// RequestBodyCapture 保存请求处理过程中实际读取的 body，供错误日志使用。
type RequestBodyCapture struct {
	Bytes []byte
}

type traceIDKey struct{}

// WithTraceID 将追踪 ID 写入标准 context，供非 Gin 的下游调用链读取。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// TraceIDFromContext 从标准 context 读取追踪 ID。
func TraceIDFromContext(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(traceIDKey{}).(string)
	return traceID, ok && traceID != ""
}
