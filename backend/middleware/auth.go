package middleware

import (
	"net/http"
	"strings"

	"aiops/controllers"

	"github.com/gin-gonic/gin"
)

// AuthRequired 需要认证的中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// SSE 等 GET 场景无法携带自定义请求头，允许通过 query 参数 token  fallback 认证
		if authHeader == "" && c.Request.Method == http.MethodGet {
			authHeader = c.Query("token")
		}

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"message": "未登录或登录已过期",
			})
			c.Abort()
			return
		}

		// 支持 "Bearer <token>" 格式
		token := authHeader
		if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		value, ok := controllers.GlobalTokenStore.Load(token)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"message": "登录已过期，请重新登录",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		info := value.(controllers.TokenInfo)
		c.Set("user_id", info.UserID)
		c.Set("username", info.Username)
		c.Set("user_name", info.Name)

		c.Next()
	}
}