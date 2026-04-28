package contextkey

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
)
