# 中间件架构文档

## 概述
项目提供了一套完整的 Gin 中间件，用于处理认证、授权、限流、跨域、分页等常见需求。

## 中间件列表

### 1. 跨域中间件 (CORS)
**文件**: `api/middleware/cross-domain.go`

**功能**: 处理跨域请求，允许所有域名访问

**使用方式**:
```go
router.Use(middleware.CorssDomainHandler())
```

### 2. JWT 认证中间件

#### 2.1 基础JWT认证
**文件**: `api/middleware/jwt.go`

**功能**: 验证单体应用的JWT token

**使用方式**:
```go
router.Use(middleware.TokenVerify)
```

**上下文数据**:
- `jwtData`: JWT中的数据字段

#### 2.2 多租户JWT认证
**文件**: `api/middleware/multi_tenant_jwt.go`

**功能**: 验证多租户场景的JWT token，解析用户ID、租户ID、账号信息

**使用方式**:
```go
router.Use(middleware.MultiTenantTokenVerify)
```

**上下文数据**:
- `user_id`: 用户ID (uint)
- `tenant_id`: 租户ID (uint)
- `account`: 账号名称 (string)

**辅助函数**:
- `GetCurrentUserID(c)`: 获取当前用户ID
- `GetTenantID(c)`: 获取租户ID
- `GetCurrentAccount(c)`: 获取当前账号

### 3. 权限验证中间件
**文件**: `api/middleware/super_admin.go`

#### 3.1 超级管理员验证
**功能**: 验证是否为超级管理员（用户ID为1）

**使用方式**:
```go
router.Use(middleware.SuperAdminVerify)
```

#### 3.2 租户管理员验证
**功能**: 验证是否为租户管理员或超级管理员

**使用方式**:
```go
router.Use(middleware.TenantAdminVerify)
```

**辅助函数**:
- `IsSuperAdmin(c)`: 判断是否为超级管理员

**注意**: 简化版本，完整实现需要结合数据库进行用户状态和角色验证

### 4. 参数验证中间件
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

### 5. 分页中间件
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

## 认证授权架构

### 单体应用认证流程
1. 用户登录 -> 调用 `authentication.JWTIssue(data)` 生成token
2. 请求API -> 使用 `middleware.TokenVerify` 验证
3. 业务处理 -> 从 `c.Get("jwtData")` 获取用户信息

### 多租户认证流程
1. 用户登录 -> 调用 `auth.JWTIssue(userID, tenantID, account)` 生成token
2. 请求API -> 使用 `middleware.MultiTenantTokenVerify` 验证
3. 业务处理 -> 使用辅助函数获取用户信息:
   - `middleware.GetCurrentUserID(c)`
   - `middleware.GetTenantID(c)`
   - `middleware.GetCurrentAccount(c)`

## 工具函数

### JWT工具 (util/authentication)
- `PrepareRegisteredClaims(rc)`: 填充JWT默认字段
- `SignHS256(claims)`: HS256签名
- `ParseHS256(token, claims)`: HS256解析
- `JWTIssue(data)`: 签发普通JWT
- `JWTDecrypt(token)`: 解析普通JWT

### 多租户JWT工具 (api/auth)
- `JWTIssue(userID, tenantID, account)`: 签发多租户JWT
- `JWTDecrypt(token)`: 解析多租户JWT
- `EncodeUserInfo/DecodeUserInfo`: 用户信息编解码（兼容性）

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
auth.Use(middleware.MultiTenantTokenVerify)
{
    auth.GET("/user/info", handler.GetUserInfo)
}

// 需要管理员权限的接口
admin := router.Group("/api/admin")
admin.Use(middleware.MultiTenantTokenVerify)
admin.Use(middleware.TenantAdminVerify)
{
    admin.POST("/users", handler.CreateUser)
}

// 需要超级管理员权限的接口
super := router.Group("/api/super")
super.Use(middleware.MultiTenantTokenVerify)
super.Use(middleware.SuperAdminVerify)
{
    super.POST("/tenants", handler.CreateTenant)
}
```

## 兼容性说明
- 保留了原有的 `TokenVerify` 中间件，确保向后兼容
- 新增 `MultiTenantTokenVerify` 用于多租户场景
- 两种认证方式可以共存，根据业务需求选择使用
