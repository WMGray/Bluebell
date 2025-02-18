package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*

{
	"code": 10001, // 程序中的错误码
	"msg": "xx", // 提示信息
	"data": {} // 数据
}

*/

type ResponseData struct {
	Code ResCode     `json:"code"`           // 程序中的错误码
	Msg  interface{} `json:"msg"`            // 提示信息
	Data interface{} `json:"data,omitempty"` // 数据
}

func ResponseError(ctx *gin.Context, code ResCode) {
	ctx.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  code.Msg(),
		Data: nil,
	})
}

func ResponseErrorWithMsg(ctx *gin.Context, code ResCode, msg interface{}) {
	ctx.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func ResponseSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, &ResponseData{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}
