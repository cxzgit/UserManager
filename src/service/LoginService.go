package service

import (
	"UserManager/src/mapper"
	"UserManager/src/utils"
	"fmt"
	"golang.org/x/crypto/bcrypt"
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
	token, err := utils.GenerateToken(user)

	if err != nil {
		return "", fmt.Errorf("生成令牌失败")
	}

	return token, nil
}
