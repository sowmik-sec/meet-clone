package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.Engine, handler *AuthHandler) {
	users := router.Group("/users")
	{
		users.POST("/signup", handler.Signup)
		users.POST("/login", handler.Login)
		// users.POST("/refresh", handler.Refresh) // TODO: Implement refresh
	}
}
