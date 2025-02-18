package mysql

import (
	"bluebell/models"
	"crypto/md5"
	"database/sql"
	"encoding/hex"

	"go.uber.org/zap"
)

// 把每一步数据库操作封装成函数，等待logic层调用

const secret = "WMGray"

// CheckUserExist 检查指定用户名的用户是否存在
func CheckUserExist(username string) (err error) {
	sqlStr := "select count(user_id) from user where username = ?"
	var count int64
	if err := db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return ErrorUserExist
	}
	return
}

// InsertUser 向数据库中插入一条新的用户记录
func InsertUser(user *models.User) (err error) {
	// 密码加密
	user.Password = encryptPassword(user.Password)

	// 执行SQL语句
	sqlStr := "insert into user(user_id, username, password) values(?, ?, ?)"
	_, err = db.Exec(sqlStr, user.UserID, user.Username, user.Password)
	return
}

func encryptPassword(password string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(password)))
}

func Login(user *models.User) (err error) {
	opassword := user.Password
	sqlStr := "select user_id, username, password from user where username = ?"
	err = db.Get(user, sqlStr, user.Username)
	if err == sql.ErrNoRows {
		return ErrorUserNotExist
	}
	if err != nil {
		// 数据库查询失败
		return err
	}
	// 判断密码是否正确
	//fmt.Println("user.Password", user.Password)
	if encryptPassword(opassword) != user.Password {
		//fmt.Println(user.Password, opassword)
		return ErrorInvalidPassword
	}
	return
}

// GetUserByID 根据用户ID查询用户信息
func GetUserByID(id int64) (user *models.User, err error) {
	user = new(models.User)
	sqlStr := "select user_id, username from user where user_id = ?"
	if err = db.Get(user, sqlStr, id); err != nil {
		zap.L().Error("mysql.GetUserID() failed. ", zap.Error(err))
		return
	}
	return
}
