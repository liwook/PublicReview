package router

import (
	"review/handler/user"
	"review/middleware"
	"review/pkg/response"

	"github.com/gin-gonic/gin"
)

func HandleNotFound(c *gin.Context) {
	response.Error(c, response.ErrNotFound, "route not found")
}

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.NoRoute(HandleNotFound)
	r.GET("/ping", func(c *gin.Context) {
		response.Success(c, "pong")
	})

	public := r.Group("/api/v1")
	public.Use(middleware.OptionalJWT())
	{
		public.POST("/send-code", user.SendCode)
		public.POST("/login", user.Login)
	}

	auth := r.Group("/api/v1")
	// 强制认证：token解析 + 认证检查
	auth.Use(middleware.OptionalJWT(), middleware.RequireAuth())
	{
	}

	return r
}
