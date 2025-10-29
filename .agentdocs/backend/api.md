# 后端 API 路由架构

## 分层路由设计（每级目录均有 router.go）
采用“逐级嵌套路由”的组织方式：app 目录下每一个层级目录都必须包含一个 `router.go`，通过 `RegisterRoutes(...)` 向下注册子路由，形成清晰的分层编排。

- 顶层 `api/router.go`：仅负责 gin 初始化、全局中间件与挂载 `/api` 分组，然后调用 `app.RegisterRoutes(apiGroup)`。
- `api/app/router.go`：在 `/api` 下创建版本分组（当前为 `/v1`），调用 `v1.RegisterRoutes(v1Group)`。
- `api/app/v1/router.go`：在 `/api/v1` 下创建功能分组（如 `open`、`private`），分别调用各自的 `RegisterRoutes`。
- `api/app/v1/open/router.go` 与 `api/app/v1/private/router.go`：在各自分组内汇总并注册模块路由（如 `health`）。
- 具体模块（如 `health`）在自身目录内实现最贴近业务的注册函数（保留 `RegisterOpenRoutes`/`RegisterPrivateRoutes` 亦可）。

### 目录结构
```
http-services/
  api/
    router.go               # 顶层路由与全局中间件
    app/
      router.go             # /api 下编排版本（v1 等）
      v1/
        router.go           # /api/v1 下编排 open、private 等
        open/
          router.go         # /api/v1/open 下聚合公开模块
          health/
            router.go       # health 模块子路由注册
            health.go       # 健康检查 handler
        private/
          router.go         # /api/v1/private 下聚合私有模块
```

## 统一规范

- 顶层不直接写具体 handler 路由，所有业务路由在 app 分层逐级注册。
- 统一接口：各层目录使用 `RegisterRoutes(*gin.RouterGroup)` 作为对外注册入口。
- 模块可继续按需提供 `RegisterOpenRoutes` / `RegisterPrivateRoutes`，由上层在各自分组内调用。
- 安全与限流策略：
  - 全局策略（安全响应头、请求 ID、全局限流、跨域等）在顶层统一 `Use`。
  - 与模块强相关的策略在模块或分组级别的 `router.go` 中“就近维护”。

## 示例：health 模块

位于 `api/app/v1/open/health/router.go`，对外在 `open` 汇总层注册：

```go
// /api/v1/open
func (open *gin.RouterGroup) {
    health.RegisterOpenRoutes(open)
}
```

## 新增模块流程

1) 按层级创建目录并在每级新增 `router.go`（若不存在）：`app/` → `v1/` → `open|private/` → `<module>/`。
2) 在对应分组的 `router.go` 中调用模块的注册函数（或直接声明路由）。
3) 如需新增版本，仅在 `app/router.go` 中追加 `/v2`，并实现 `app/v2/router.go` 即可，不影响既有分层。

## 兼容性与路径

- 路径保持不变：健康检查仍为 `GET /api/v1/open/health`。
- 重构后顶层不再直接导入模块包，改由分层 `router.go` 统一编排，减少耦合。
