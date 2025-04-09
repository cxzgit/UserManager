package mapper

import (
	"UserManager/src/models"
	"database/sql"
	"fmt"
	"time"
)

type UserMapper struct {
	DB *sql.DB
}

func NewUserMapper(db *sql.DB) *UserMapper {
	return &UserMapper{DB: db}
}

// 查询用户，分页
func (um *UserMapper) QueryUser(page, pageSize int) ([]models.User, error) {
	offset := (page - 1) * pageSize
	query := "SELECT id, email, password_hash, nickname, avatar_url, created_at, role, status FROM users LIMIT ? OFFSET ?"
	rows, err := um.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.AvatarUrl, &user.CreatedAt, &user.Role, &user.Status); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserByEmail 根据邮箱查询用户（用于判断邮箱是否已注册）
func (um *UserMapper) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var createdAt string
	err := um.DB.QueryRow("SELECT id, email, password_hash,role,nickname,avatar_url,created_at FROM users WHERE email = ?", email).
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

// 查询用户总数
// mapper/user.go

// QueryUsersWithPage 按关键字和状态分页查询用户，同时返回总记录数
func (um *UserMapper) QueryUsersWithPage(keyword, statusStr string, offset, limit int) ([]*models.User, int, error) {
	// 1. 基础查询
	baseSQL := `FROM users WHERE 1=1`
	var args []interface{}

	// 2. 模糊查询 nickname 或 email
	if keyword != "" {
		baseSQL += " AND (nickname LIKE ? OR email LIKE ?)"
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	// 3. 状态筛选
	if statusStr == "0" || statusStr == "1" {
		baseSQL += " AND status = ?"
		args = append(args, statusStr)
	}

	// 4. 查询总数
	countSQL := "SELECT COUNT(*) " + baseSQL
	var total int
	if err := um.DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 5. 查询数据页
	dataSQL := "SELECT id, email, nickname, avatar_url, role, status, created_at " +
		baseSQL + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := um.DB.Query(dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Nickname, &u.AvatarUrl, &u.Role, &u.Status, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}
	return users, total, nil
}

// 新增用户
func (um *UserMapper) CreateUser(user *models.User) error {
	query := "INSERT INTO users (email, password_hash, nickname, avatar_url, role, status) VALUES (?, ?, ?, ?, ?, ?) "
	// 使用当前时间作为创建时间
	result, err := um.DB.Exec(query, user.Email, user.PasswordHash, user.Nickname, user.AvatarUrl, user.Role, user.Status)
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

// 更新用户
func (um *UserMapper) UpdateUser(user *models.User) error {
	var query string

	var err error
	// 更新所有字段
	query = `
		UPDATE users
		SET email         = ?,
			password_hash = ?,
			nickname      = ?,
			avatar_url    = ?,
			role          = ?,
			status        = ?
		WHERE id = ?
	`
	_, err = um.DB.Exec(
		query,
		user.Email,
		user.PasswordHash,
		user.Nickname,
		user.AvatarUrl,
		user.Role,
		user.Status,
		user.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (um *UserMapper) GetUserByID(id int) (*models.User, error) {
	query := `
		SELECT id, email,password_hash,nickname, role, status, created_at, avatar_url 
		FROM users 
		WHERE id = ?
	`
	row := um.DB.QueryRow(query, id)
	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Role, &user.Status, &user.CreatedAt, &user.AvatarUrl)
	if err != nil {
		fmt.Println("陈炫祯")
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no user found with id %d", id)
		}
		return nil, err
	}
	return &user, nil
}

// 删除用户
func (um *UserMapper) DeleteUser(id int) error {
	result, err := um.DB.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// 没有任何行受影响，说明用户不存在
		return sql.ErrNoRows
	}
	return nil
}
