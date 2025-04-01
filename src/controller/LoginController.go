package controller

import (
	"UserManager/src/service"
	"UserManager/src/utils"
	"encoding/json"
	"html/template"
	"net/http"
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
	json.NewEncoder(w).Encode(utils.SuccessResult(map[string]string{"token": token}))
}
func (lc *LoginController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	//由中间件解析 JWT 后写入 Context
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](401, "无效的用户身份"))
		return
	}
	err := lc.Service.LogoutUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](500, "登出失败"))
		return
	}
	json.NewEncoder(w).Encode(utils.SuccessResult("登出成功"))
}
