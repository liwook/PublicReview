package router

import (
	"fmt"
	"review/pkg/response"

	"github.com/gin-gonic/gin"
)

type User struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age" `
}

func HandleNotFound(c *gin.Context) {
	response.Error(c, response.ErrNotFound, "route not found")
}

func NewRouter() *gin.Engine {
	r := gin.Default()
	r.NoRoute(HandleNotFound)
	r.GET("/ping", func(c *gin.Context) {
		response.Success(c, "pong")
	})

	r.POST("/shouldbindjson", func(ctx *gin.Context) {
		var u User
		err := ctx.ShouldBindJSON(&u)
		if err != nil {
			fmt.Println(err)
			//response.ErrDatabase是应该返回500的，结果也是返回了500
			response.Error(ctx, response.ErrDatabase, "")
			return
		}
		response.Success(ctx, gin.H{"name": u.Name, "age": u.Age})
	})
	r.POST("/bindjson", func(ctx *gin.Context) {
		var u User
		err := ctx.BindJSON(&u)
		if err != nil {
			fmt.Println(err)
			//response.ErrDatabase是应该返回500的，使用BindJSON但却返回了400;
			//也出现警告：[GIN-debug] [WARNING] Headers were already written. Wanted to override status code 400 with 500
			//要是不使用response.Error，也返回了错误的。 BindJSON 提供了开箱即用的默认错误处理
			response.Error(ctx, response.ErrDatabase, "bindjson error")
			return
		}

		response.Success(ctx, gin.H{"name": u.Name, "age": u.Age})
	})

	return r
}
