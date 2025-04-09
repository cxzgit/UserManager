package models

import "time"

// User 定义用户数据模型
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Role         int       `json:"role"`
	Nickname     string    `json:"nickname"`
	AvatarUrl    string    `json:"avatar_url"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
