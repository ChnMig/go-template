# 日志系统架构

## 概述

项目使用 Uber Zap 作为日志库，并实现了自动上下文注入机制。每个请求的日志自动包含追踪 ID 和请求信息。

## 核心功能

### 1. 自动上下文注入
RequestID 中间件自动为每个请求创建带上下文的 logger：
- `trace_id` - 请求追踪 ID（对应 X-Request-ID）
- `method` - HTTP 方法（GET、POST 等）
- `path` - 请求路径
- `client_ip` - 客户端 IP 地址

### 2. 双模式输出
- **开发模式** (`--dev`)：输出到终端，易读格式
- **生产模式**：输出到文件，JSON 格式

### 3. 日志轮转
使用 lumberjack 实现：
- 单文件最大大小：50MB（可配置）
- 保留备份数：3 个（可配置）
- 保留天数：30 天（可配置）

### 4. 日志监控
生产模式下自动监控日志文件，如被删除或重命名则自动重建。

## 使用方法

### 基本用法

```go
import "http-services/utils/log"

func Handler(c *gin.Context) {
    // 获取带上下文的 logger
    logger := log.FromContext(c)

    // 记录日志（自动包含 trace_id 等信息）
    logger.Info("处理用户请求")

    logger.Warn("可恢复的异常",
        zap.String("reason", "临时错误"),
    )

    logger.Error("操作失败",
        zap.Error(err),
        zap.String("operation", "create_user"),
    )
}
```

### 添加结构化字段

```go
logger.Info("用户登录成功",
    zap.String("user_id", user.ID),
    zap.String("username", user.Username),
    zap.String("ip", c.ClientIP()),
    zap.Duration("duration", time.Since(startTime)),
)

logger.Error("数据库查询失败",
    zap.Error(err),
    zap.String("query", sql),
    zap.Any("params", params),
)
```

## 日志级别

### 使用规范

| 级别 | 使用场景 | 示例 |
|------|----------|------|
| Debug | 详细的调试信息（仅开发模式） | 参数值、中间结果 |
| Info | 重要的业务流程节点 | 用户登录、订单创建 |
| Warn | 可恢复的异常情况 | 重试成功、降级服务 |
| Error | 需要关注的错误 | 数据库错误、外部服务失败 |
| Fatal | 致命错误（程序终止） | 配置错误、端口占用 |

### 示例

```go
// Debug - 仅开发模式可见
logger.Debug("解析请求参数",
    zap.Any("request", req),
)

// Info - 正常业务流程
logger.Info("创建订单",
    zap.String("order_id", orderID),
    zap.Float64("amount", amount),
)

// Warn - 可恢复的问题
logger.Warn("外部服务响应慢",
    zap.String("service", "payment-api"),
    zap.Duration("latency", latency),
)

// Error - 需要关注的错误
logger.Error("支付失败",
    zap.Error(err),
    zap.String("order_id", orderID),
)

// Fatal - 致命错误（慎用）
if config.JWTKey == "" {
    zap.L().Fatal("JWT key is required")
}
```

## 日志输出格式

### 开发模式
```
2024-10-28T10:30:00.123+0800  INFO  api/app/example/login.go:45  用户登录成功
    trace_id: abc123
    method: POST
    path: /api/v1/login
    client_ip: 192.168.1.1
    user_id: 12345
    username: john
```

### 生产模式（JSON）
```json
{
  "level": "info",
  "ts": "2024-10-28T10:30:00.123Z",
  "caller": "example/login.go:45",
  "msg": "用户登录成功",
  "trace_id": "abc123",
  "method": "POST",
  "path": "/api/v1/login",
  "client_ip": "192.168.1.1",
  "user_id": "12345",
  "username": "john"
}
```

## 配置管理

### 配置文件 (config.yaml)
```yaml
log:
  max_size: 50      # 单个日志文件最大大小（MB）
  max_backups: 3    # 保留的旧日志文件最大数量
  max_age: 30       # 保留旧日志文件的最大天数
```

### 环境变量覆盖
```bash
export HTTP_SERVICES_LOG_MAX_SIZE=100
export HTTP_SERVICES_LOG_MAX_BACKUPS=5
export HTTP_SERVICES_LOG_MAX_AGE=60
```

## 日志文件

### 文件位置
- 路径：`./log/http-services.log`
- 轮转文件：`http-services-2024-10-28T10-30-00.log`

### 日志清理
旧日志自动清理，保留策略：
1. 保留最近 3 个备份文件
2. 保留 30 天内的文件
3. 超出任一条件的文件将被删除

## 最佳实践

### ✅ 推荐做法

1. **使用 FromContext 获取 logger**
   ```go
   logger := log.FromContext(c)
   logger.Info("操作成功")
   ```

2. **添加有用的上下文字段**
   ```go
   logger.Info("处理订单",
       zap.String("order_id", orderID),
       zap.String("user_id", userID),
       zap.Float64("amount", amount),
   )
   ```

3. **记录关键业务节点**
   ```go
   logger.Info("订单状态变更",
       zap.String("order_id", orderID),
       zap.String("from_status", oldStatus),
       zap.String("to_status", newStatus),
   )
   ```

4. **错误日志包含足够信息**
   ```go
   logger.Error("数据库操作失败",
       zap.Error(err),
       zap.String("operation", "insert"),
       zap.String("table", "orders"),
       zap.Any("data", order),
   )
   ```

5. **错误响应前统一记录 Error 日志**
   ```go
   func ReturnDomainError(c *gin.Context, err error) {
       // 在统一错误映射处记录真实的领域错误和请求上下文，便于排查
       log.WithRequest(c).Error("健康检查领域错误", zap.Error(err))
       // ...根据 err 映射统一错误响应...
   }
   ```

### ❌ 避免做法

1. **不要在循环中记录过多日志**
   ```go
   // 错误示例
   for _, item := range items {
       logger.Info("处理项目", zap.Any("item", item))
   }

   // 正确做法
   logger.Info("开始批量处理", zap.Int("count", len(items)))
   // ... 处理逻辑 ...
   logger.Info("批量处理完成", zap.Int("success", successCount))
   ```

2. **不要记录敏感信息**
   ```go
   // 错误示例
   logger.Info("用户登录", zap.String("password", password))

   // 正确做法
   logger.Info("用户登录", zap.String("user_id", userID))
   ```

3. **不要使用全局 logger**
   ```go
   // 错误示例
   zap.L().Info("处理请求")

   // 正确做法
   logger := log.FromContext(c)
   logger.Info("处理请求")
   ```

4. **不要在日志中使用字符串拼接**
   ```go
   // 错误示例
   logger.Info("User " + userID + " login success")

   // 正确做法
   logger.Info("用户登录成功", zap.String("user_id", userID))
   ```

## 性能优化

### 1. 使用采样
生产模式已配置采样：每秒首4条日志全部记录，之后每秒记录1条。

### 2. 避免昂贵的字段计算
```go
// 推荐：延迟计算
logger.Debug("详细信息", zap.Any("data", func() interface{} {
    return expensiveOperation()
}))

// 或者只在必要时记录
if logger.Core().Enabled(zapcore.DebugLevel) {
    logger.Debug("详细信息", zap.Any("data", expensiveOperation()))
}
```

## 问题排查

### 使用 trace_id 追踪请求
1. 从错误响应或日志中获取 `trace_id`
2. 在日志文件中搜索该 `trace_id`
3. 查看该请求的完整处理流程

```bash
# 搜索特定请求的所有日志
grep "abc123" log/http-services.log
```

### 日志分析
```bash
# 统计错误数量
grep '"level":"error"' log/http-services.log | wc -l

# 查看特定用户的操作
grep '"user_id":"12345"' log/http-services.log

# 查看慢请求
grep '"latency"' log/http-services.log | grep -E '"latency":[0-9]{4,}'
```

## 中间件日志

### RequestID 中间件
自动记录：
- 请求开始：包含基本信息
- 请求完成：包含状态码

```json
{
  "level": "info",
  "msg": "Request started",
  "trace_id": "abc123",
  "method": "POST",
  "path": "/api/v1/users",
  "client_ip": "192.168.1.1"
}

{
  "level": "info",
  "msg": "Request completed",
  "trace_id": "abc123",
  "method": "POST",
  "path": "/api/v1/users",
  "client_ip": "192.168.1.1",
  "status_code": 200
}
```

## 生命周期管理

### 初始化
```go
// main.go
log.GetLogger()      // 初始化 logger
log.StartMonitor()   // 启动日志监控
```

### 优雅关闭
```go
// main.go
defer log.StopMonitor()  // 停止监控并刷新缓冲区
```

## 扩展建议

### 集成日志收集系统
生产环境建议将日志发送到集中式日志系统：
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Loki + Grafana
- 云服务日志（如阿里云SLS、腾讯云CLS）

### 日志分析和告警
基于日志建立监控告警：
- 错误率异常告警
- 慢请求告警
- 特定错误码告警

## 参考资料

- Zap 文档：https://pkg.go.dev/go.uber.org/zap
- Lumberjack：https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2
- 改进文档：`IMPROVEMENTS.md`
