package router

import (
	"review/handler/blog"
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

		public.GET("/shop/distance-list", shopservice.QueryShopDistance)
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

		auth.POST("/blog/images", blog.UploadImages)
		auth.POST("/blogs", blog.SaveBlog)
		auth.GET("/blogs/:blogId", blog.GetBlogById)
		auth.POST("/blogs/:blogId/like", blog.LikeBlog)
		auth.GET("/blogs/:blogId/likes", blog.GetBlogLikes)
		auth.GET("/users/:userId/blogs", blog.GetBlogsByUserId)
		auth.GET("/users/:userId/following-blogs", blog.QueryBlogOfFollow)

		auth.GET("/users/:userId/follow/:followId", user.FollowUser)
		auth.POST("/users/:userId/follow/:followId", user.FollowUser)
		auth.DELETE("/users/:userId/follow/:followId", user.UnfollowUser)
		auth.GET("/users/follow/commons", user.FollowCommons)

		auth.POST("/unique-visitor", user.AddUniqueVisitor)
		auth.GET("/blogs/:blogId/unique-visitor", user.ContinuousSigninStatistics)

		auth.GET("/user/:userId", user.QueryUserById)

		auth.POST("/user/:userId/signIn", user.SignIn)
		auth.GET("/user/:userId/signin-statistics", user.ContinuousSigninStatistics)
	}

	return r
}
