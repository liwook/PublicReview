package router

import (
	"dianping/pkg/code"

	"github.com/gin-gonic/gin"
)

type User struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age" `
}

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		code.WriteResponse(c, code.ErrSuccess, "pong")
	})

	r.POST("/shouldbindjson", func(ctx *gin.Context) {
		var u User
		err := ctx.ShouldBindJSON(&u)
		if err != nil {
			//返回了500
			code.WriteResponse(ctx, code.ErrDatabase, err.Error())
			return
		}
		code.WriteResponse(ctx, code.ErrSuccess, gin.H{"name": u.Name, "age": u.Age})
	})
	r.POST("/bindjson", func(ctx *gin.Context) {
		var u User
		err := ctx.BindJSON(&u)
		if err != nil {
			//但却返回了400
			code.WriteResponse(ctx, code.ErrDatabase, "bindjson error") //ErrDatabase是返回500的
			return
		}
		code.WriteResponse(ctx, code.ErrSuccess, gin.H{"name": u.Name, "age": u.Age})
	})

	return r
}
