package utils

import (
	"UserManager/src/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// 密钥/签名
var jwtSecret = []byte("cxzzs")

type Claims struct {
	UserID    int    `json:"user_id"`
	Role      int    `json:"role"`
	AvatarUrl string `json:"avatar_url"`
	//预定义声明的结构体，里面的字段用于描述Token的基本信息
	jwt.RegisteredClaims
}

// 生成JWT
func GenerateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:    user.ID,
		Role:      user.Role,
		AvatarUrl: user.AvatarUrl,
		RegisteredClaims: jwt.RegisteredClaims{
			//过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			//签发时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("生成 Token 失败: %w", err)
	}
	return signedString, nil
}

// 解析JWI
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
