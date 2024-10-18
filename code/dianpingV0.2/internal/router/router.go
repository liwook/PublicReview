package router

import (
	"dianping/internal/user"
	"dianping/pkg/code"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	// r.Use(middleware.JWT()) //使用jwt中间件

	r.GET("/ping", func(c *gin.Context) {
		code.WriteResponse(c, code.ErrSuccess, "pong")
	})

	r.GET("/user/verificationcode/:phone", user.GetVerificationCode)
	r.POST("/user/login", user.Login)
	return r
}
