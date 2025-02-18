package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"fmt"

	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

type VoteData struct {
	// UserID 可以从请求中获取
	PostID    int64 `json:"post_id,string" binding:"required"`                // 帖子ID
	Direction int8  `json:"direction,string" binding:"required,oneof=1 0 -1"` // 赞成票(1)还是反对票(-1)还是取消投票(0)
}

// PostVoteHandler 投票
func PostVoteHandler(ctx *gin.Context) {
	// 1. 获取参数及参数校验: 用户ID、帖子ID、投票类型
	p := new(models.ParamVoteData)
	if err := ctx.ShouldBindJSON(p); err != nil {
		errs, ok := err.(validator.ValidationErrors) // 类型断言
		fmt.Println(err)
		if !ok {
			ResponseError(ctx, CodeInvalidParam)
			return
		}
		errData := removeTagStruct(errs.Translate(trans)) // 翻译并去掉错误提示中的结构体标签
		zap.L().Error("controller.PostVoteHandler with invalid param", zap.Error(err))
		ResponseErrorWithMsg(ctx, CodeInvalidParam, errData)
		return
	}
	// 获取用户ID
	uid, err := getcurrentUser(ctx)
	if err != nil {
		ResponseError(ctx, CodeNeedLogin)
		return
	}
	// 2. 投票的逻辑处理
	if err := logic.VoteForPost(uid, p); err != nil {
		zap.L().Error("logic.VoteForPost failed", zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}
	// 3. 返回响应
	ResponseSuccess(ctx, nil)
}
