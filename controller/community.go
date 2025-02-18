package controller

import (
	"bluebell/logic"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CommunityHandler 跟社区相关的handler
// @Summary 查询社区列表
// @Description 查询社区列表的信息
// @Tags 信息查询
// @Param
// @Success 200 {object} _ResponseCommunityList
func CommunityHandler(ctx *gin.Context) {
	// 查询所有的社区（community_id, community_name）列表
	data, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("Logic.GetCommunityList() failed", zap.Error(err))
		ResponseError(ctx, CodeServerBusy) // 不能将服务器内部错误暴露给用户
		return
	}
	ResponseSuccess(ctx, data)
}

// CommunityDetailHandler 查询社区详情
func CommunityDetailHandler(ctx *gin.Context) {
	// 查询指定的社区详情
	// 1. 获取社区ID
	idStr := ctx.Param("id")
	// 2. 参数校验
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ResponseError(ctx, CodeInvalidParam)
		return
	}
	// 3. 业务处理
	data, err := logic.GetCommunityDetail(id)
	if err != nil {
		zap.L().Error("logic.GetCommunityDetail() failed", zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}
	ResponseSuccess(ctx, data)
}
