package middlewares

import (
	"net/http"
	"time"

	"github.com/juju/ratelimit"

	"github.com/gin-gonic/gin"
)

// 限流中间件
func RateLimitMiddleware(fillInterval time.Duration, cap int64) func(ctx *gin.Context) {
	// 令牌桶算法实现限流
	bucket := ratelimit.NewBucket(fillInterval, cap)
	return func(ctx *gin.Context) {
		// 如果取不到令牌，就返回限流提示
		if bucket.TakeAvailable(1) == 0 {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"msg": "too many requests",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
