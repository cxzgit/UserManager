package mapper

import (
	"UserManager/src/models"
	"database/sql"
	"fmt"
	"time"
)

type LoginMapper struct {
	DB *sql.DB
}

func NewLoginMapper(db *sql.DB) *LoginMapper {
	return &LoginMapper{DB: db}
}

// GetUserByEmail 根据邮箱查询用户（用于判断邮箱是否已注册）
func (lm *LoginMapper) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var createdAt string
	err := lm.DB.QueryRow("SELECT id, email, password_hash,role,nickname,avatar_url, created_at FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Nickname, &user.AvatarUrl, &createdAt)
	if err != nil {
		return nil, err
	}
	user.CreatedAt, err = time.Parse("2006-01-02T15:04:05Z07:00", createdAt)
	if err != nil {
		fmt.Errorf("时间解析错误")
		return nil, err
	}
	return user, nil
}
