package controller

import (
	"bluebell/dao/mysql"
	"bluebell/logic"
	"bluebell/models"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func SignupHandler(ctx *gin.Context) {
	// 1. 获取参数和参数校验
	p := new(models.ParamSignUp)

	if err := ctx.ShouldBindJSON(p); err != nil {
		// 请求参数有误，直接返回响应
		zap.L().Error("Signup with invalid param", zap.Error(err))
		// 判断 err 是否为 validator.ValidationErrors 类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(ctx, CodeInvalidParam)
			return
		}
		// 是 validator.ValidationErrors 类型，进行翻译
		ResponseErrorWithMsg(ctx, CodeInvalidParam, removeTagStruct(errs.Translate(trans)))
		return
	}

	//// 手动对请求参数进行详细的业务规则校验
	//if len(p.Username) == 0 || len(p.Password) == 0 || len(p.RePassword) == 0 || p.Password != p.RePassword {
	//	ctx.JSON(http.StatusOK, gin.H{
	//		"msg": "请求参数错误",
	//	})
	//	return
	//}
	fmt.Println(p)
	// 2. 业务处理
	if err := logic.SignUp(p); err != nil {
		zap.L().Error("logic.Signup failed", zap.Error(err))
		if errors.Is(err, mysql.ErrorUserExist) {
			ResponseError(ctx, CodeUserExist)
			return
		}
		ResponseError(ctx, CodeServerBusy)
		return
	}
	// 3. 返回响应
	ResponseSuccess(ctx, nil)
}

func LoginHandler(ctx *gin.Context) {
	// 1. 获取参数和参数校验
	p := new(models.ParamLogin)
	if err := ctx.ShouldBindJSON(p); err != nil {
		zap.L().Error("Login with invalid param", zap.Error(err))
		// 判断 err 是否为 validator.ValidationErrors 类型
		err, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(ctx, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(ctx, CodeInvalidParam, removeTagStruct(err.Translate(trans)))
	}
	// 2. 业务处理 --> 调用 logic 函数
	user, err := logic.Login(p)
	if err != nil {
		zap.L().Error("Logic.Login failed", zap.String("username: ", p.Username),
			zap.Error(err))
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(ctx, CodeUserNotExist)
			return
		}
		ResponseError(ctx, CodeInvalidPassword)
		return
	}
	// 3. 返回响应
	ResponseSuccess(ctx, gin.H{
		"user_id":       fmt.Sprintf("%d", user.UserID), //js识别的最大值：id值大于1<<53-1  int64: i<<63-1
		"user_name":     user.Username,
		"access_token":  user.AccessToken,
		"refresh_token": user.RefreshToken,
	})
}
