package router

import (
	"dianping/pkg/code"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		code.WriteResponse(c, code.ErrSuccess, "pong")
	})

	r.GET("/test", func(c *gin.Context) {
		code.WriteResponse(c, code.ErrDatabase, "not find this data")
	})

	r.GET("/test2", func(c *gin.Context) {
		code.WriteResponse(c, code.ErrDatabase, nil)
	})
	return r
}
