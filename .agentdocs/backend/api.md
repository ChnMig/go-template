# 后端 API 路由架构

## 分层路由设计
本项目采用“多层级 Router”模式来组织 Gin 路由：

- 顶层在 `api/router.go` 中定义全局中间件、版本分组（如 `/api/v1`）以及一级分组（如 `open`、`private`）。
- 各业务模块（app）在自身目录下定义子路由注册函数，仅关注本模块的路由与中间件组合，避免在全局路由中堆叠具体业务逻辑。

### 目录结构
```
http-services/
  api/
    router.go            # 顶层路由与全局中间件
    app/
      v1/
        open/
          health/
            router.go    # health 模块子路由注册
            health.go    # 健康检查 handler
```

## 统一规范

- 顶层仅做“分组与编排”，不直接写具体 handler 路由。
- app 内提供以下约定函数（按需实现）：
- `RegisterOpenRoutes(open *gin.RouterGroup)`：开放接口子路由，前缀为 `/api/v1/open/<module>`。
- `RegisterPrivateRoutes(private *gin.RouterGroup)`：需认证接口子路由，前缀为 `/api/v1/private/<module>`。
- 安全与限流策略：
  - 全局通用策略（如安全响应头、请求ID、监控、全局限流、跨域）在顶层统一 `Use`。
  - 与业务绑定的策略（如某模块特定的限流或认证）在 app 的 `router.go` 中自行叠加，做到“就近维护”。

## 示例：example 模块

位于 `api/app/v1/open/health/router.go`，将健康检查路由迁移到模块内部：

```go
// 开放接口：/api/v1/open/health
open.GET("/health", Status)

// 私有接口：/api/v1/private/example
r := private.Group("/example", middleware.TokenVerify, middleware.TokenRateLimit(100, 200))
r.GET("/pong", Pong)
```

## 为新模块添加路由

1) 在 `api/app/<your-app>/` 下创建 `router.go`，实现上述 `RegisterOpenRoutes`/`RegisterPrivateRoutes`（可按需选择）。

2) 在 `api/router.go` 中相应的 open/private 编排处引入你的模块注册函数（health 作为示例已采用模块化注册）：
```go
// 省略其他模块...
health.RegisterOpenRoutes(open)
yourapp.RegisterOpenRoutes(open)
yourapp.RegisterPrivateRoutes(private)
```

3) 中间件放置建议：
- 与模块强相关的限流、认证放在模块子路由（更灵活，易于调优）。
- 需要所有模块统一生效的策略（如统一鉴权）可放在顶层某个分组上（例如对 `private` 统一 `Use(middleware.TokenVerify)`）。

## 兼容性说明与约束

此次重构属于“结构化收敛”，路径与中间件语义保持不变，除以下调整：
- 合并健康检查接口为单一路径：`GET /api/v1/open/health`（包含 ready 与 uptime 信息）。
- 移除 Prometheus 相关：不再注册 `router.Use(middleware.Metrics())`，且不暴露 `/metrics` 端点。
- 删除示例模块 example 及其路由示例，保持模板最简。

如需统一权限或全局私有组中间件，可在 `api/router.go` 的 `private` 组上追加 `Use(...)`，但属于行为变更，请评估后再调整。
