package router

import (
	"dianping/internal/shopservice"
	"dianping/internal/user"
	"dianping/pkg/code"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	// r := gin.Default()
	// r.Use(middleware.JWT()) //使用jwt中间件
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		code.WriteResponse(c, code.ErrSuccess, "pong")
	})

	r.GET("/user/verificationcode/:phone", user.GetVerificationCode)
	r.POST("/user/login", user.Login)

	r.GET("/shop/:id", shopservice.QueryShopById)
	r.GET("/shop/type-list", shopservice.QueryShopTypeList)
	r.POST("/shop/update", shopservice.UpdateShop)
	r.POST("/shop/add", shopservice.AddShop)
	r.DELETE("/shop/delete/:id", shopservice.DelShop)

	r.POST("/voucher/add", shopservice.AddVoucher)
	return r
}
