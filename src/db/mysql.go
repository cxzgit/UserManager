package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() {
	var err error
	// DSN 示例：用户名:密码@tcp(127.0.0.1:3306)/数据库名
	DB, err = sql.Open("mysql", "root:cxz@tcp(127.0.0.1:3306)/goweb")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}
}
