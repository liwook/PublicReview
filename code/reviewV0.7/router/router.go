package router

import (
	"review/handler/order"
	"review/handler/shopservice"
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

	public := r.Group("/api/v1")
	public.Use(middleware.OptionalJWT())
	{
		public.POST("/send-code", user.SendCode)
		public.POST("/login", user.Login)

		// 添加公开访问的商店接口，这些都不需要登录的
		public.GET("/shop/:shopId", shopservice.QueryShopById)
		public.GET("/shop/type-list", shopservice.QueryShopTypeList)
		public.POST("/seckill/vouchers", order.SeckillVoucher) 
	}

	auth := r.Group("/api/v1")
	// 强制认证：token解析 + 认证检查
	auth.Use(middleware.OptionalJWT(), middleware.RequireAuth())
	{
		// 商户相关接口，需要登录验证
		auth.PUT("/shop", shopservice.UpdateShop)
		auth.DELETE("/shop/:shopId", shopservice.DelShop)
		auth.POST("/shop", shopservice.AddShop)

		// auth.POST("/seckill/vouchers", order.SeckillVoucher) //post /api/v1/seckill/vouchers
		auth.POST("/vouchers", order.AddVoucher)
	}

	return r
}
