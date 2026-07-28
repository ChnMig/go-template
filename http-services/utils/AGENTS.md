# UTILS KNOWLEDGE

## OVERVIEW

与业务无关的基础设施工具层；提供日志、认证、加密、ID、路径、pidfile、运行模式、context key 与并发任务组。

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| 日志 | `log/` | Zap 生命周期、Gin 日志、轮转、request context logger |
| JWT | `authentication/` | JWT 签发、解析与 claims |
| 密码 | `encryption/` | bcrypt hash、verify 与格式识别 |
| ID | `id/` | Sonyflake、UUIDv7 与其 MD5 表示 |
| Context key | `contextkey/` | trace_id、logger、JWT 数据和绑定参数的统一 key |
| 路径 | `pathtool/` | 工作目录、路径存在性、目录和空文件创建 |
| pidfile | `pidfile/` | 独占写入、PID 所有权校验与清理 |
| 随机值 | `random/` | crypto/rand 驱动的十六进制随机值 |
| 运行模式 | `runmodel/` | CLI 与环境变量的 dev/release 判定 |
| 并发任务组 | `taskgroup/` | panic 恢复、按策略取消、等待全部退出、按输入顺序返回错误 |

## CONVENTIONS

- `utils/` 只放业务无关能力；业务常量、Redis key 和领域 DTO 不放这里。
- 底层 helper 返回带上下文的 error，由边界层决定是否记录日志，避免重复日志。
- Gin context key 统一使用 `contextkey`，不要散落裸字符串。
- 并发任务通过 `taskgroup.CancelOnError` / `ContinueOnError` 声明策略；业务错误优先级由调用方决定。
- TLS 在反向代理、Ingress 或负载均衡终止，服务进程不内置 ACME/TLS 文件监听。

## ANTI-PATTERNS

- 不要在工具层引入业务模型或业务配置。
- 不要在 helper 内吞掉 error 或重复记录同一个 error。
- 不要启动没有 context 退出路径和等待机制的 goroutine。
- 不要重新引入服务内 TLS/ACME 作为模板默认能力。
