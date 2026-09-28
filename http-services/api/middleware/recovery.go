package middleware

import (
	"errors"
	"net/http"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"http-services/utils/log"
)

// Recovery 保留 Gin 的 HTTP 500 和连接断开处理方式，
// 同时记录与普通错误响应一致的 trace 和请求参数。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if err, ok := recovered.(error); ok && (errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, http.ErrAbortHandler)) {
					log.WithRequest(c).Debug("request connection aborted", zap.Error(err))
					_ = c.Error(err)
					c.Abort()
					return
				}
				log.WithRequest(c).Error("request panic recovered", zap.Any("panic", recovered), zap.Stack("stack"))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
