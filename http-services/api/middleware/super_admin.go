package middleware

import (
	"github.com/gin-gonic/gin"

	"go-services/api/response"
)

// GetTenantID 从上下文获取租户ID
func GetTenantID(c *gin.Context) uint {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return 0
	}
	return tenantID.(uint)
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// GetCurrentAccount 从上下文获取当前账号
func GetCurrentAccount(c *gin.Context) string {
	account, exists := c.Get("account")
	if !exists {
		return ""
	}
	return account.(string)
}

// SuperAdminVerify 超级管理员权限验证中间件（简化版）
// 只有用户ID为1的超级管理员才能执行特定操作
// 注意：完整版本应该结合数据库进行用户状态验证
func SuperAdminVerify(c *gin.Context) {
	// 获取当前用户ID
	userID := GetCurrentUserID(c)
	if userID == 0 {
		response.ReturnError(c, response.UNAUTHENTICATED, "用户认证失败")
		c.Abort()
		return
	}

	// 检查是否为超级管理员（用户ID为1）
	if userID != 1 {
		response.ReturnError(c, response.PERMISSION_DENIED, "权限不足，只有超级管理员可以执行此操作")
		c.Abort()
		return
	}

	c.Next()
}

// IsSuperAdmin 判断当前请求是否为平台超级管理员
func IsSuperAdmin(c *gin.Context) bool {
	return GetCurrentUserID(c) == 1
}

// TenantAdminVerify 租户管理员权限验证中间件（简化版）
// 允许超级管理员或租户管理员执行特定操作
// 注意：完整版本应该结合数据库进行用户状态和角色验证
func TenantAdminVerify(c *gin.Context) {
	// 获取当前用户ID和租户ID
	userID := GetCurrentUserID(c)
	tenantID := GetTenantID(c)

	if userID == 0 || tenantID == 0 {
		response.ReturnError(c, response.UNAUTHENTICATED, "用户认证失败")
		c.Abort()
		return
	}

	// 超级管理员（用户ID为1）可以访问任何租户数据
	if userID == 1 {
		c.Next()
		return
	}

	// TODO: 这里应该添加实际的数据库查询，验证用户角色和权限
	// 当前简化版本假设已通过JWT认证的用户都有租户管理员权限

	c.Next()
}
