package main

import (
	"UserManager/src/controller"
	"UserManager/src/db"
	"UserManager/src/mapper"
	"UserManager/src/middleware"
	"UserManager/src/service"
	"UserManager/src/utils"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// 初始化mysql数据库连接
	db.InitDB()
	defer db.DB.Close()
	//初始化redis
	db.InitRedis("192.168.24.101:6379", "", 0)
	defer db.RedisClient.Close()
	// **创建 EmailService 实例**
	emailService := utils.NewEmailService("3287670282@qq.com", "xsnwkuvczgvbdaeg")

	// **创建 VerificationService**
	verificationService := utils.NewVerificationService(db.RedisClient)

	// 初始化各层

	//注册
	registerMapper := mapper.NewRegisterMapper(db.DB)
	registerService := service.NewRegisterService(registerMapper, emailService, verificationService)
	registerController := controller.NewRegisterController(registerService)

	//登录
	loginMapper := mapper.NewLoginMapper(db.DB)
	loginService := service.NewLoginService(loginMapper)
	loginController := controller.NewLoginController(loginService)

	// 使用 Gorilla mux 路由
	router := mux.NewRouter()
	//注册时发送邮箱验证码
	router.HandleFunc("/send_code", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			registerController.SendCodePage(w, r)
		} else if r.Method == http.MethodPost {
			registerController.SendCodeHandler(w, r)
		}
	}).Methods(http.MethodGet, http.MethodPost)
	//注册
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			registerController.RegisterPage(w, r)
		} else if r.Method == http.MethodPost {
			registerController.RegisterHandler(w, r)
		}
	}).Methods(http.MethodGet, http.MethodPost)
	//登录
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			loginController.LoginPage(w, r)
		} else if r.Method == http.MethodPost {
			loginController.LoginHandler(w, r)
		}
	}).Methods(http.MethodGet, http.MethodPost)

	//受保护路由（需要登录才能访问）
	protected := router.PathPrefix("/").Subrouter()
	// JWT 拦截器，中间件会解析 JWT，并把用户信息写入 Context
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/logout", loginController.LogoutHandler).Methods(http.MethodPost)
	fmt.Println("服务器启动，监听端口 :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
