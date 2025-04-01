package service

import (
	"UserManager/src/db"
	"UserManager/src/mapper"
	"UserManager/src/utils"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type LoginService struct {
	Mapper *mapper.LoginMapper
}

func NewLoginService(lm *mapper.LoginMapper) *LoginService {
	return &LoginService{
		Mapper: lm,
	}
}

// 用户登录
func (ls *LoginService) LoginUser(email string, password string) (string, error) {
	user, err := ls.Mapper.GetUserByEmail(email)

	if err != nil || user == nil {
		return "", fmt.Errorf("用户不存在")
	}

	//验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("密码错误")
	}
	//生成JWT
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("生成令牌失败")
	}

	//保存session到Redis
	sessionKey := fmt.Sprintf("session:%d", user.ID)
	err = db.RedisClient.Set(db.Ctx, sessionKey, token, 24*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("保存 session 失败")
	}
	return token, nil
}

// 用于登出
func (ls *LoginService) LogoutUser(userID int) error {
	sessionKey := fmt.Sprintf("session:%d", userID)
	err := db.RedisClient.Del(db.Ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("登出失败")
	}
	return nil
}
