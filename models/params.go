package models

// 定义请求的参数结构体

// ParamSignUp 注册请求参数
type ParamSignUp struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}

// ParamLogin 登录请求参数
type ParamLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ParamVoteData 投票数据
type ParamVoteData struct {
	PostID    string `json:"post_id" binding:"required"`              // 帖子ID
	Direction int8   `json:"direction,string" binding:"oneof=1 0 -1"` // 赞成票(1)还是反对票(-1)还是取消投票(0)
}

// ParamPostList 获取帖子列表query string参数
const (
	OrderTime  = "time"
	OrderScore = "score"
)

// ParamPostList 获取帖子列表query string参数
type ParamPostList struct {
	Page        int64  `json:"page" form:"page"`
	Size        int64  `json:"size" form:"size"`
	CommunityID int64  `json:"community_id" form:"community_id"` // 社区ID 可以为空
	Order       string `json:"order" form:"order"`
}

// ParamCommunityList 按社区ID获取帖子列表的query string参数
//type ParamCommunityPostList struct {
//	*ParamPostList
//	CommunityID int64 `json:"community_id" form:"community_id"`
//}
