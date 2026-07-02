package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"project20260702/internal/auth"
	"project20260702/internal/response"
)

const userIDContextKey = "userID"

// Auth 校验登录 token。
//
// 小程序请求受保护接口时，需要带：
// Authorization: Bearer <token>
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			response.Error(c, 401, 40101, "请先登录")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		if tokenString == header {
			response.Error(c, 401, 40101, "Authorization 格式不正确")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(tokenString, jwtSecret)
		if err != nil {
			response.Error(c, 401, 40101, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set(userIDContextKey, claims.UserID)
		c.Next()
	}
}

// CurrentUserID 从 Gin 上下文里取出当前登录用户 ID。
//
// 这个值由 Auth 中间件解析 token 后写入。
func CurrentUserID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(userIDContextKey)
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint64)
	return userID, ok
}
