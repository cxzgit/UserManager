package controller

import (
	"UserManager/src/service"
	"UserManager/src/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type RegisterController struct {
	Service *service.RegisterService
	Tmpl    *template.Template
}

func NewRegisterController(s *service.RegisterService) *RegisterController {
	// 加载模板文件：templates/register.html
	tmpl := template.Must(template.ParseFiles("views/login.html"))
	return &RegisterController{
		Service: s,
		Tmpl:    tmpl,
	}
}

// SendCodePage 渲染请求验证码页面
func (rc *RegisterController) SendCodePage(w http.ResponseWriter, r *http.Request) {
	if err := rc.Tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, "页面渲染错误", http.StatusInternalServerError)
	}
}

// SendCodeHandler 处理发送验证码请求
func (rc *RegisterController) SendCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](405, "请求方法不支持"))
		return
	}
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](400, "解析表单错误"))
		return
	}
	email := r.FormValue("email")
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](400, "邮箱不能为空"))
		return
	}
	err := rc.Service.SendVerificationCode(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](500, "验证码发送失败"))
		return
	}
	json.NewEncoder(w).Encode(utils.SuccessResult(fmt.Sprintf("验证码已发送至 %s，请查收邮件", email)))
}

// RegisterPage 渲染注册页面
func (rc *RegisterController) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if err := rc.Tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, "页面渲染错误", http.StatusInternalServerError)
	}
}

// RegisterHandler 处理注册表单提交
func (rc *RegisterController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](405, "请求方法不支持"))
		return
	}
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](400, "解析表单错误"))
		return
	}
	// 解析 JSON 请求体
	var req struct {
		Email           string `json:"email"`
		Code            string `json:"code"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求解析失败", http.StatusBadRequest)
		return
	}
	err := rc.Service.RegisterUser(req.Email, req.Code, req.Password, req.ConfirmPassword)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[string](400, fmt.Sprintf("注册失败：%v", err)))
		return
	}
	json.NewEncoder(w).Encode(utils.SuccessResult("注册成功，欢迎使用"))
}
