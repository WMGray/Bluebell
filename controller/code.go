package controller

// 定义一些程序中可能会用到的错误码

type ResCode int64

const (
	CodeSuccess ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy

	CodeNeedLogin
	CodeInvalidToken
)

var CodeMsg = map[ResCode]string{
	CodeSuccess:         "success",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户已存在",
	CodeUserNotExist:    "用户不存在",
	CodeInvalidPassword: "用户名或密码错误",
	CodeServerBusy:      "服务繁忙",

	CodeNeedLogin:    "需要登录",
	CodeInvalidToken: "无效的token",
}

func (c ResCode) Msg() string {
	msg, ok := CodeMsg[c]
	if !ok {
		msg = CodeMsg[CodeServerBusy]
	}
	return msg
}
