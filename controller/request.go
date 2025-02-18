package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

const CtxUserIDKey = "userID"

var ErrorUserNotLogin = errors.New("用户未登录")

// getCurrentUser 获取当前登录用户的ID
func getcurrentUser(ctx *gin.Context) (userID int64, err error) {
	uid, ok := ctx.Get(CtxUserIDKey)
	if !ok {
		ResponseError(ctx, CodeNeedLogin)
		return 0, ErrorUserNotLogin
	}
	userID, ok = uid.(int64)
	if !ok {
		ResponseError(ctx, CodeNeedLogin)
		return 0, ErrorUserNotLogin
	}
	return userID, nil
}

// getPageInfo 获取分页参数
func getPageInfo(ctx *gin.Context) (int64, int64) {
	// 获取分页参数
	pageStr := ctx.Query("page")
	sizeStr := ctx.Query("size")
	var (
		page int64
		size int64
		err  error
	)
	// 参数校验
	page, err = strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		page = 1
	}
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		size = 10
	}
	return page, size
}
