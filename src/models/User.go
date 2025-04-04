package models

import "time"

// User 定义用户数据模型
type User struct {
	ID           int
	Email        string
	PasswordHash string
	Role         int
	Nickname     string
	AvatarUrl    string
	CreatedAt    time.Time
}
