package service

import (
	"UserManager/src/mapper"
	"UserManager/src/models"
	"UserManager/src/utils"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
)

type UserService struct {
	Mapper *mapper.UserMapper
}

func NewUserService(hm *mapper.UserMapper) *UserService {
	return &UserService{
		Mapper: hm,
	}
}

// GetUsers 支持分页 + 搜索 + 状态筛选
func (us *UserService) GetUsers(keyword, statusStr string, page, pageSize int) ([]*models.User, int, error) {
	// 计算 offset
	offset := (page - 1) * pageSize
	// 委托 Mapper 执行查询并返回总数
	return us.Mapper.QueryUsersWithPage(keyword, statusStr, offset, pageSize)
}

// 新增用户
func (us *UserService) CreateUser(email, passwordHash, nickname string, avatarFile io.Reader, avatarFileName string, role, status int) (*models.User, error) {
	// 上传头像到 OSS
	avatarURL, err := utils.UploadFileToOSS(avatarFile, avatarFileName)
	if err != nil {
		return nil, err
	}
	// 检查邮箱是否已注册
	if user, _ := us.Mapper.GetUserByEmail(email); user != nil {
		return nil, errors.New("邮箱已注册，请使用其他邮箱")
	}
	// 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(passwordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	user := &models.User{
		Email:        email,
		PasswordHash: string(hashed),
		Nickname:     nickname,
		AvatarUrl:    avatarURL,
		Role:         role,
		Status:       status,
	}
	if err := us.Mapper.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (us *UserService) UpdateUser(id int, email, passwordHash, nickname string, avatarFile io.Reader, avatarFileName string, role, status int) (*models.User, error) {
	// 1. 先取出数据库中的原用户
	user, err := us.Mapper.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	// 2. 如果邮箱被修改，检查是否重复
	if email != user.Email {
		if existing, _ := us.Mapper.GetUserByEmail(email); existing != nil {
			return nil, errors.New("邮箱已注册，请使用其他邮箱")
		}
	}

	// 3. 更新基本字段
	user.Email = email
	user.Nickname = nickname
	user.Role = role
	user.Status = status

	// 4. 如果前端传了非空密码，才加密并更新
	if passwordHash != "" {
		if !isValidPassword(passwordHash) {
			return nil, fmt.Errorf("密码必须是8-12位字母和数字组合")
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(passwordHash), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		user.PasswordHash = string(hashed)
	}

	// 5. 如果前端上传了新头像，才上传并更新 URL
	if avatarFile != nil {
		url, err := utils.UploadFileToOSS(avatarFile, avatarFileName)
		if err != nil {
			return nil, err
		}
		user.AvatarUrl = url
	}
	// 6. 写回数据库
	if err := us.Mapper.UpdateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

// 根据用户 ID 获取用户信息
func (us *UserService) GetUserByID(id int) (*models.User, error) {
	return us.Mapper.GetUserByID(id)
}

// 删除用户信息
func (us *UserService) DeleteUser(id int) error {
	return us.Mapper.DeleteUser(id)
}
