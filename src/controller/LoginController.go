package controller

import (
	"UserManager/src/service"
	"UserManager/src/utils"
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

type LoginController struct {
	Service *service.LoginService
	Tmpl    *template.Template
}

func NewLoginController(s *service.LoginService) *LoginController {
	//加载模板文件，templates/login1.html
	tmpl := template.Must(template.ParseFiles("views/login.html"))
	return &LoginController{
		Service: s,
		Tmpl:    tmpl,
	}
}

// 渲染登录页面
func (lc *LoginController) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := lc.Tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, "页面渲染错误", http.StatusInternalServerError)
	}
}
func (lc *LoginController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "请求方法不支持", http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](405, "请求方法不支持"))
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](400, "请求解析失败"))
		return
	}

	token, err := lc.Service.LoginUser(req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](401, err.Error()))
		return
	}
	// 设置 Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,                          // 生产环境应设为 true
		Expires:  time.Now().Add(24 * time.Hour), // 24 小时后失效，不自动续期
	})
	json.NewEncoder(w).Encode(utils.SuccessResult(map[string]string{"token": token}))
}
