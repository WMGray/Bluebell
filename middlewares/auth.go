package middlewares

import (
	"bluebell/controller"
	"bluebell/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URL
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头
		// Authorization: Bearer xxx.xxx.xxx
		// 这里的具体实现方式要依据你的实际业务情况而定
		authHeader := ctx.Request.Header.Get("Authorization")
		if authHeader == "" {
			controller.ResponseError(ctx, controller.CodeNeedLogin)
			ctx.Abort()
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controller.ResponseError(ctx, controller.CodeInvalidToken)
			ctx.Abort()
			return
		}
		// part[1]是获取到的tokenString，需要进行解析
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controller.ResponseError(ctx, controller.CodeInvalidToken)
			ctx.Abort()
			return
		}
		// 将当前请求的userID信息保存到请求的上下文ctx上
		ctx.Set(controller.CtxUserIDKey, mc.UserID)
		ctx.Next() // 后续的处理函数可以用 ctx.Get(CtxUserIDKey) 来获取当前请求的用户信息
	}
}
