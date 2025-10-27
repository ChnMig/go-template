# 中间件架构文档

## 概述
项目提供了一套完整的 Gin 中间件，用于处理认证、跨域、分页等常见需求。

## 中间件列表

### 1. 跨域中间件 (CORS)
**文件**: `api/middleware/cross-domain.go`

**功能**: 处理跨域请求，允许所有域名访问

**使用方式**:
```go
router.Use(middleware.CorssDomainHandler())
```

### 2. JWT 认证中间件
**文件**: `api/middleware/jwt.go`

**功能**: 验证JWT token

**使用方式**:
```go
router.Use(middleware.TokenVerify)
```

**上下文数据**:
- `jwtData`: JWT中的数据字段

### 3. 参数验证中间件
**文件**: `api/middleware/params.go`

**功能**: 验证和绑定请求参数

**使用方式**:
```go
type LoginParams struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
    var params LoginParams
    if !middleware.CheckParam(&params, c) {
        return
    }
    // 处理登录逻辑
}
```

### 4. 分页中间件
**文件**: `api/middleware/page.go`

**功能**: 解析分页参数

**使用方式**:
```go
func ListUsers(c *gin.Context) {
    page := middleware.GetPage(c)           // 默认1
    pageSize := middleware.GetPageSize(c)   // 默认20
    // 查询数据
}
```

**查询参数**:
- `page`: 页码
- `pageSize`: 每页大小

## 认证架构

### JWT认证流程
1. 用户登录 -> 调用 `authentication.JWTIssue(data)` 生成token
2. 请求API -> 使用 `middleware.TokenVerify` 验证
3. 业务处理 -> 从 `c.Get("jwtData")` 获取用户信息

## 工具函数

### JWT工具 (util/authentication)
- `PrepareRegisteredClaims(rc)`: 填充JWT默认字段
- `SignHS256(claims)`: HS256签名
- `ParseHS256(token, claims)`: HS256解析
- `JWTIssue(data)`: 签发JWT token
- `JWTDecrypt(token)`: 解析JWT token

### 加密工具 (util/encryption)
- `HashPasswordWithBcrypt(password)`: 密码加密
- `VerifyBcryptPassword(password, hash)`: 密码验证
- `IsBcryptHash(hash)`: 检查bcrypt格式

## 中间件使用示例

### 完整路由配置示例
```go
router := gin.Default()

// 全局中间件
router.Use(middleware.CorssDomainHandler())

// 公开接口
public := router.Group("/api/public")
{
    public.POST("/login", handler.Login)
}

// 需要认证的接口
auth := router.Group("/api")
auth.Use(middleware.TokenVerify)
{
    auth.GET("/user/info", handler.GetUserInfo)
    auth.POST("/users", handler.CreateUser)
}
```
