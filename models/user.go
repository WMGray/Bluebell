package models

type User struct {
	UserID       int64  `json:"user_id,string" db:"user_id"` // 指定json序列化/反序列化时使用小写user_id
	Username     string `json:"username" db:"username"`
	Password     string `json:"password" db:"password"`
	Email        string `json:"email" db:"email"`   // 邮箱
	Gender       int    `json:"gender" db:"gender"` // 性别
	AccessToken  string
	RefreshToken string
}
