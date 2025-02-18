package jwt

import (
	"errors"
	"time"

	"github.com/spf13/viper"

	"github.com/golang-jwt/jwt/v5"
)

var mySecret = []byte("浅吟轻唱一曲离歌")

func keyFunc(_ *jwt.Token) (i interface{}, err error) {
	return mySecret, nil
}

type MyClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.MapClaims
}

// GenToken 生成JWT
func GenToken(userID int64, username string) (atoken, rtoken string, err error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		userID,   // 自定义字段
		username, // 自定义字段
		jwt.MapClaims{
			"exp": time.Now().Add(
				time.Hour * time.Duration(viper.GetInt("auth.jwt_expire"))).Unix(), // 过期时间
			"iat": time.Now().Unix(), // 发布时间
			"iss": "bluebell",        // 签发人
		},
	}
	// 加密并获取完整的编码后的字符串token
	atoken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)

	// Refresh Token 不需要任何自定义字段
	rtoken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * time.Duration(viper.GetInt("auth.jwt_expire")) * 30).Unix(), // 过期时间
		"iat": time.Now().Unix(),                                                                      // 发布时间
		"iss": "bluebell",                                                                             // 签发人
	}).SignedString(mySecret)
	return
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (claims *MyClaims, err error) {
	// 解析token
	var token *jwt.Token
	claims = new(MyClaims)
	token, err = jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return
	}
	if !token.Valid {
		err = errors.New("invalid token")
	}
	return
}

// RefreshToken 刷新Token
func RefreshToken(atoken, rtoken string) (NewAToken, NewRToken string, err error) {
	// refresh token 无效直接返回
	if _, err = jwt.Parse(rtoken, keyFunc); err != nil {
		return
	}
	// 从旧Access Token中解析出claims数据
	var claims MyClaims
	_, err = jwt.ParseWithClaims(atoken, &claims, keyFunc)
	if err != nil {
		// 检查是否为过期错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			// 生成新的Access Token
			NewAToken, _, err = GenToken(claims.UserID, claims.Username)
			return
		}
		return
	}
	return
}
