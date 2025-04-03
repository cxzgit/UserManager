package mapper

import (
	"UserManager/src/models"
	"database/sql"
	"fmt"
	"time"
)

type RegisterMapper struct {
	DB *sql.DB
}

func NewRegisterMapper(db *sql.DB) *RegisterMapper {
	return &RegisterMapper{DB: db}
}

// GetUserByEmail 根据邮箱查询用户（用于判断邮箱是否已注册）
func (rm *RegisterMapper) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var createdAt string
	err := rm.DB.QueryRow("SELECT id, email, password_hash,role,nickname,avatar_url,created_at FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Nickname, &user.AvatarUrl, &createdAt)
	if err != nil {
		return nil, err
	}
	user.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		fmt.Errorf("时间解析错误")
		return nil, err
	}
	return user, nil
}

// InsertUser 插入新用户数据
func (rm *RegisterMapper) InsertUser(user *models.User) error {
	result, err := rm.DB.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", user.Email, user.PasswordHash)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)
	return nil
}
