package models

import (
	"time"
)

// 内存对齐

// Post 帖子结构体
type Post struct {
	ID          int64     `json:"id,string" db:"post_id"`
	AuthorID    int64     `json:"author_id,string" db:"author_id"`
	CommunityID int64     `json:"community_id" db:"community_id" binding:"required"`
	Status      int32     `json:"status" db:"status"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Content     string    `json:"content" db:"content" binding:"required"`
	CreateTime  time.Time `json:"create_time" db:"create_time"`
}

// ApiPostDetail 帖子详情接口
type ApiPostDetail struct {
	AuthorName       string             `json:"author_name"`
	VoteNum          int64              `json:"vote_num"`
	*Post                               // 嵌入帖子结构体
	*CommunityDetail `json:"community"` // 嵌入社区结构体
}

// PostScore 帖子分数
type PostScore struct {
	ID    int64 `json:"id,string" db:"post_id"`
	Score int64 `json:"score" db:"score"`
}

// PostVoteData 帖子投票数据
type PostVoteData struct {
	PostID    int64 `json:"post_id,string" db:"post_id"`
	UserID    int64 `json:"user_id,string" db:"user_id"`
	Direction int32 `json:"direction" db:"direction"`
}

//
//// Value 给 PostScore 和 PostVoteData 实现 driver.Valuer 接口，使其可以被 sqlx.In() 函数使用
//// Value 实现 driver.Valuer 接口
//func (ps PostScore) Value() (driver.Value, error) {
//	return []interface{}{ps.ID, ps.Score}, nil
//}
//
//// Value 实现 driver.Valuer 接口
//func (pvd PostVoteData) Value() (driver.Value, error) {
//	return []interface{}{pvd.PostID, pvd.UserID, pvd.Direction}, nil
//}
