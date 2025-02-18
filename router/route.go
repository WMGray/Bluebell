package router

import (
	"bluebell/controller"
	"bluebell/logger"
	"bluebell/middlewares"
	"net/http"

	"github.com/gin-contrib/pprof"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "bluebell/docs" // 千万不要忘了导入把你上面生成的docs

	"github.com/gin-gonic/gin"
)

// gin-swagger middleware
// swagger embed files

// SetupRouter 路由
func SetupRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // gin设置成发布模式
	}
	r := gin.New()

	//r.Use(logger.GinLogger(), logger.GinRecovery(true), middlewares.RateLimitMiddleware(2*time.Second, 1))
	// 令牌桶中间件
	r.Use(logger.GinLogger(), logger.GinRecovery(true), middlewares.RateLimitMiddleware(20, 10000))

	// 加载静态文件
	r.LoadHTMLFiles("./templates/index.html")
	r.Static("/static", "./static")
	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	v1 := r.Group("/api/v1")
	// 注册业务路由 --> controller.SignupHandler
	v1.POST("./signup", controller.SignupHandler)

	// 登录业务路由 --> controller.LoginHandler
	v1.POST("./login", controller.LoginHandler)
	//v1.GET("/posts", controller.GetPostListHandler)

	// 使用中间件
	// JWT 认证中间件
	v1.Use(middlewares.JWTAuthMiddleware())
	{
		// 发帖业务路由 --> controller.CreatePostHandler
		v1.GET("/community", controller.CommunityHandler)
		v1.GET("/community/:id", controller.CommunityDetailHandler)

		v1.POST("/post", controller.CreatePostHandler)
		v1.GET("/post/:id", controller.GetPostDetailHandler)
		//v1.GET("/posts", controller.GetPostListHandler)
		// 根据帖子时间或者分数进行排序，然后返回
		v1.GET("/posts2", controller.GetPostListHandler2)
		v1.POST("/vote", controller.PostVoteHandler)

		// 文档
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "ping --> pong")
	})
	pprof.Register(r) // 注册 pprof 相关路由
	r.NoRoute(func(ctx *gin.Context) {
		controller.ResponseErrorWithMsg(ctx, controller.CodeInvalidParam, "404")
	})
	return r
}
