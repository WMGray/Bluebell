package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreatePostHandler 发帖
func CreatePostHandler(ctx *gin.Context) {
	// 1. 获取参数及参数校验
	p := new(models.Post)
	// ctx.ShouldBindJSON() // validator --> binding
	if err := ctx.ShouldBindJSON(p); err != nil {
		zap.L().Error("controller.CreatePostHandler: ctx.ShouldBindJSON() failed", zap.Error(err))
		ResponseError(ctx, CodeInvalidParam)
		return
	}
	// 从请求中获取当前用户的ID
	userID, err := getcurrentUser(ctx)
	if err != nil {
		ResponseError(ctx, CodeNeedLogin)
		return
	}
	fmt.Println("userID:", userID)
	p.AuthorID = userID
	// 2. 创建帖子
	if err := logic.CreatePost(p); err != nil {
		zap.L().Error("controller.CreatePostHandler: logic.CreatePost() failed", zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}
	// 3. 返回响应
	ResponseSuccess(ctx, nil)
}

// GetPostDetailHandler 获取帖子详情
//
//	@Summary		获取帖子详情接口
//	@Description	可按社区按时间或分数排序查询帖子列表接口
//	@Tags			帖子相关接口(api分组展示使用的)
//	@Accept			application/json
//	@Produce		application/json
//	@Param			Authorization	header	string	true	"Bearer JWT"
//	@Param			id				path	int		true	"帖子ID"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	_ResponsePostList
//	@Router			/post/:id [get]
func GetPostDetailHandler(ctx *gin.Context) {
	// 1. 获取帖子ID
	pidStr := ctx.Param("id")
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		zap.L().Error("controller.GetPostDetailHandler: invalid param", zap.Error(err))
		ResponseError(ctx, CodeInvalidParam)
		return
	}
	// 2. 根据ID取出帖子数据
	data, err := logic.GetPostByID(pid)
	if err != nil {
		zap.L().Error("controller.GetPostDetailHandler: logic.GetPostByID() failed", zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}
	ResponseSuccess(ctx, data)
}

// GetPostListHandler 获取帖子列表
func GetPostListHandler(ctx *gin.Context) {
	page, size := getPageInfo(ctx)
	// 获取数据
	data, err := logic.GetPostList(page, size)
	if err != nil {
		zap.L().Error("controller.GetPostListHandler: logic.GetPostList() failed", zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(ctx, data)
}

// GetPostListHandler2 升级版获取帖子接口
// 根据前端传递的参数来获取帖子列表：创建时间/分数
//
//	@Summary		升级版帖子列表接口
//	@Description	可按社区按时间或分数排序查询帖子列表接口
//	@Tags			帖子相关接口(api分组展示使用的)
//	@Accept			application/json
//	@Produce		application/json
//	@Param			Authorization	header	string					true	"Bearer JWT"
//	@Param			object			query	models.ParamPostList	false	"查询参数"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	_ResponsePostList
//	@Router			/posts2 [get]
func GetPostListHandler2(ctx *gin.Context) {
	// 1. 获取参数： 时间 or 分数
	// Get 请求参数(query string): /api/v1/posts2?page=1&size=10&order=time
	p := &models.ParamPostList{
		Page:        1,
		Size:        10,
		CommunityID: 0,
		Order:       models.OrderTime,
	}
	// ctx.ShouldBind() 根据请求的数据类型选择相应的方法去获取数据
	// ctx.ShouldBindJSON() 如果请求中携带的是josn格式的数据，才能用这个方法获取到数据
	if err := ctx.ShouldBindQuery(p); err != nil {
		zap.L().Error("controller.GetPostListHandler2 ctx.ShouldBindQuery failed", zap.Error(err))
		ResponseError(ctx, CodeInvalidParam)
		return
	}

	data, err := logic.GetPostListNew(p)
	if err != nil {
		zap.L().Error("logic.GetPostList2 failed", zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}
	// 4. 返回帖子信息列表
	ResponseSuccess(ctx, data)
}

// GetCommunityPostListHander 根据社区查询帖子列表
//func GetCommunityPostListHander(ctx *gin.Context) {
//	// 1. 获取参数： 时间 or 分数
//	// Get 请求参数(query string): /api/v1/posts2?page=1&size=10&order=time
//	p := &models.ParamCommunityPostList{
//		CommunityID: 1,
//		ParamPostList: &models.ParamPostList{
//			Page:  1,
//			Size:  10,
//			Order: "score",
//		},
//	}
//	// ctx.ShouldBind() 根据请求的数据类型选择相应的方法去获取数据
//	// ctx.ShouldBindJSON() 如果请求中携带的是josn格式的数据，才能用这个方法获取到数据
//	if err := ctx.ShouldBindQuery(p); err != nil {
//		zap.L().Error("controller.GetPostListHandler2 ctx.ShouldBindQuery failed", zap.Error(err))
//		ResponseError(ctx, CodeInvalidParam)
//		return
//	}
//
//	data, err := logic.GetCommunityPostList(p)
//	if err != nil {
//		zap.L().Error("logic.GetPostList2 failed", zap.Error(err))
//		ResponseError(ctx, CodeServerBusy)
//		return
//	}
//	// 4. 返回帖子信息列表
//	ResponseSuccess(ctx, data)
//}
