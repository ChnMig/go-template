# HTTP 请求日志

统一响应通过 `log.WithRequest(c)` 记录请求信息。`TraceID` 在限流之前安装请求上下文和请求体采集，普通 Handler 使用 `log.FromContext(c)`，需要参数时使用 `log.WithRequest(c)`。参数按原值记录，不做脱敏。

## 字段

| 字段 | 含义 |
| --- | --- |
| `trace_id/method/path/client_ip` | 同一请求的通用上下文，由 logger helper 统一附加一次 |
| `query` | 原始 Query 字符串 |
| `form/multipart_form/path_params` | 已解析的表单、multipart 文本字段、路由参数 |
| `params` | 最近一次成功绑定的参数；再次绑定失败时不会残留上一次的值 |
| `request_body` | 没有成功绑定参数时，回退到已采集的原始请求体；合法 JSON 保持对象、数组和数字类型，语法错误保留原始文本 |
| `request_body_bytes` | 业务处理已读取的请求体字节数 |
| `content_type/content_length` | 请求体的 Content-Type 和声明长度，未知长度可为 -1 |
| `request_body_incomplete` | 请求流尚未读完，当前快照只有已读取的部分 |
| `request_body_read_error` | 底层请求流出现读取错误 |
| `*_truncated` | 相应字段超过 64 KiB，日志只保留前缀，实际请求未被日志逻辑截断 |
| `request_body_status` | `unread/empty/truncated/invalid_json/invalid_form/unsupported_content_type` 等状态 |

例如，整数参数 `id` 收到数组时，绑定失败和统一错误响应日志都会保留原始输入：

```json
{
  "msg": "Returning error response",
  "trace_id": "018f47a5-7b8c-7c11-8000-123456789abc",
  "method": "PUT",
  "path": "/example",
  "request_body": {"id": [1, 2], "token": "original-value"}
}
```

## 采集边界

采集器只观察业务代码对 Body 的读取，不提前消费请求流。S2S 验签后的 Body 恢复、直接 Gin 绑定、JSON 校验失败、普通绑定 helper 都能使用同一份快照；原始读取错误、Close 和 HTTP BodySizeLimit 语义保持不变。请求在读取 Body 前被鉴权或限流拒绝时，记录 `unread`，不会为写日志额外读取请求。

文本快照上限为 64 KiB，超长字段保留前缀并标记截断。上传文件和二进制 Body 不输出内容，multipart 文本参数仍按原值记录。日志字段不会加入对外 API 响应。

## 同类问题检查

- 参数绑定失败：用原始请求体回退，不能仅把部分解码后的结构体提前放入 `params`。
- JWT/S2S 鉴权：必须使用请求 logger；JWT 解析工具只返回错误，由调用方带上下文记录，避免重复的全局错误日志。
- 限流：`TraceID` 必须先执行，否则提前返回的响应和日志没有 trace。
- panic：使用请求上下文记录 panic 和堆栈，保持 HTTP 500 行为；断开的连接不再写响应。
- 业务 Handler 和中间件必须显式使用 `log.FromContext(c)` 或 `log.WithRequest(c)`；`zap.L()` 不会自动继承当前请求上下文。后台任务使用自身上下文，不持有 Gin Context。
