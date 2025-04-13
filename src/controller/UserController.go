package controller

import (
	"UserManager/src/service"
	"UserManager/src/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type UserController struct {
	Service *service.UserService
	Tmpl    *template.Template
}

func NewUserController(s *service.UserService) *UserController {
	//加载模板文件
	tmpl := template.Must(template.ParseFiles("views/userList.html"))
	return &UserController{
		Service: s,
		Tmpl:    tmpl,
	}
}

// 渲染登录页面
func (lc *UserController) UserPage(w http.ResponseWriter, r *http.Request) {
	if err := lc.Tmpl.ExecuteTemplate(w, "userList.html", nil); err != nil {
		http.Error(w, "页面渲染错误", http.StatusInternalServerError)
	}
}

func (uc *UserController) GetUsers(w http.ResponseWriter, r *http.Request) {
	// 1. 解析分页参数
	page := 1
	pageSize := 5
	if p := r.URL.Query().Get("page"); p != "" {
		if pi, err := strconv.Atoi(p); err == nil && pi > 0 {
			page = pi
		}
	}
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if psi, err := strconv.Atoi(ps); err == nil && psi > 0 {
			pageSize = psi
		}
	}

	// 2. 解析搜索 + 状态
	keyword := r.URL.Query().Get("keyword")  // 模糊查询
	statusStr := r.URL.Query().Get("status") // "0" 启用, "1" 禁用, "" 全部

	// 3. 调用 Service 层
	users, totalCount, err := uc.Service.GetUsers(keyword, statusStr, page, pageSize)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](500, "获取用户列表失败："+err.Error()))
		return
	}

	// 4. 取当前登录用户角色
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusUnauthorized, "未登录"))
		return
	}
	token, err := utils.ParseToken(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusUnauthorized, "解析 Token 失败"))
		return
	}

	// 5. 计算总页数
	pageCount := (totalCount + pageSize - 1) / pageSize

	// 6. 返回结果
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.SuccessResult(map[string]interface{}{
		"data":      users,
		"page":      page,
		"pageSize":  pageSize,
		"pageCount": pageCount,
		"CurRole":   token.Role,
	}))
}

// 新增用户
func (uc *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	// 只允许 POST 方法
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusMethodNotAllowed, "Method Not Allowed"))
		return
	}

	// 获取 JWT token
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusUnauthorized, fmt.Sprintf("获取 JWT 失败：%v", err)))
		return
	}

	token, err := utils.ParseToken(cookie.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusUnauthorized, fmt.Sprintf("解析 Token 失败：%v", err)))
		return
	}
	CurRole := token.Role

	// 解析 multipart 表单数据（最大 10MB）
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, fmt.Sprintf("解析表单失败：%v", err)))
		return
	}

	// 获取表单字段，简单校验必填字段
	email := r.FormValue("email")
	password := r.FormValue("password")
	nickname := r.FormValue("nickname")
	if email == "" || password == "" || nickname == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, "请完整填写所有必填项"))
		return
	}
	if !isValidPassword(password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, "密码必须是8-12位字母和数字组合"))
		return
	}

	// 角色和状态：这里默认普通用户、启用状态
	role := 0   // 普通用户
	status := 1 // 启用

	// 获取上传的头像文件（字段名 "avatar"）
	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, fmt.Sprintf("读取头像文件失败：%v", err)))
		return
	}
	defer file.Close()

	// 调用 Service 层创建用户
	newUser, err := uc.Service.CreateUser(email, password, nickname, file, fileHeader.Filename, role, status)
	if err != nil {
		// 如果错误信息中包含“邮箱已注册”，返回 400 提示用户更换邮箱
		if strings.Contains(err.Error(), "邮箱已注册") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, err.Error()))
			return
		}
		// 其它错误返回 500
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusInternalServerError, fmt.Sprintf("创建用户失败：%v", err)))
		return
	}

	// 正常返回成功的 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.SuccessResult(map[string]interface{}{
		"Newuser": newUser,
		"CurRole": CurRole,
	}))
}

// 修改用户
func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// 只允许 PUT 方法
	if r.Method != http.MethodPut {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusMethodNotAllowed, "Method Not Allowed"))
		return
	}

	// 获取 JWT token
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusUnauthorized, fmt.Sprintf("获取 JWT 失败：%v", err)))
		return
	}

	token, err := utils.ParseToken(cookie.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusUnauthorized, fmt.Sprintf("解析 Token 失败：%v", err)))
		return
	}
	CurRole := token.Role

	// 解析 multipart 表单数据（最大 10MB）
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, fmt.Sprintf("解析表单失败：%v", err)))
		return
	}

	// 获取表单字段，简单校验必填字段
	idStr := r.FormValue("userId")
	email := r.FormValue("email")
	nickname := r.FormValue("nickname")
	roleStr := r.FormValue("role")
	statusStr := r.FormValue("status")
	if email == "" || statusStr == "" || nickname == "" || roleStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, "请完整填写所有必填项"))
		return
	}

	// 可选字段：password、avatar
	password := r.FormValue("password")

	// avatar 可选
	var avatarFile io.Reader
	var avatarFilename string
	file, fileHeader, err := r.FormFile("avatar")
	if err == nil {
		avatarFile = file
		avatarFilename = fileHeader.Filename
		defer file.Close()
	} else if err != http.ErrMissingFile {
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, "读取头像失败："+err.Error()))
		return
	}
	id, _ := strconv.Atoi(idStr)
	role, _ := strconv.Atoi(roleStr)
	status, _ := strconv.Atoi(statusStr)
	fmt.Println("id:", id)
	// 调用 Service 层创建用户
	UpdateUser, err := uc.Service.UpdateUser(id, email, password, nickname, avatarFile, avatarFilename, role, status)
	if err != nil {
		// 如果错误信息中包含“邮箱已注册”，返回 400 提示用户更换邮箱
		if strings.Contains(err.Error(), "邮箱已注册") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, err.Error()))
			return
		}
		// 其它错误返回 500
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusInternalServerError, fmt.Sprintf("修改用户失败：%v", err)))
		return
	}

	// 正常返回成功的 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.SuccessResult(map[string]interface{}{
		"UpdateUser": UpdateUser,
		"CurRole":    CurRole,
	}))
}

func (uc *UserController) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// 只允许 GET 方法
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusMethodNotAllowed, "Method Not Allowed"))
		return
	}

	// 从 URL 中获取 id 参数
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, "缺少 id 参数"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, fmt.Sprintf("无效的 id 参数：%v", err)))
		return
	}

	// 调用 Service 层获取用户
	user, err := uc.Service.GetUserByID(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		// 如果找不到用户，可以返回 404
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusNotFound, fmt.Sprintf("用户不存在：%v", err)))
		return
	}

	// 正常返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.SuccessResult(map[string]interface{}{
		"user": user,
	}))
}

// DeleteUser 删除用户，只允许管理员执行
func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// 只允许 DELETE 方法
	if r.Method != http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusMethodNotAllowed, "Method Not Allowed"))
		return
	}

	// 从 query 里获取 userId
	idStr := r.URL.Query().Get("id")

	// 获取 JWT token
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusUnauthorized, fmt.Sprintf("获取 JWT 失败：%v", err)))
		return
	}

	token, err := utils.ParseToken(cookie.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusUnauthorized, fmt.Sprintf("解析 Token 失败：%v", err)))
		return
	}

	if idStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, "缺少用户 ID"))
		return
	}
	id, err := strconv.Atoi(idStr)

	if id == token.UserID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, "不能删除当前用户"))
		return
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusBadRequest, "用户 ID 无效"))
		return
	}

	// 调用 Service 层
	if err := uc.Service.DeleteUser(id); err != nil {
		// 如果没有找到要删除的用户
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusNotFound, fmt.Sprintf("未找到 ID 为 %d 的用户", id)))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(utils.ErrorResult[any](http.StatusInternalServerError, "删除用户失败："+err.Error()))
		}
		return
	}
	// 删除成功
	json.NewEncoder(w).Encode(utils.SuccessResult[any](nil))
}

// 退出登录
func (uc *UserController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 清除 Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // 兼容性处理（设置为过去的时间）
		HttpOnly: true,
	})
	json.NewEncoder(w).Encode(utils.SuccessResult("登出成功"))
}

// 头像
func (uc *UserController) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("2345678")
	// 从 Cookie 中获取 token
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}
	tokenString := cookie.Value

	// 解析 token
	token, err := utils.ParseToken(tokenString)
	if err != nil {
		http.Error(w, "token解析错误", http.StatusUnauthorized)
	}
	// 返回用户头像、用户名等数据，这里只返回头像 URL 作为示例
	resp := map[string]string{
		"avatar": token.AvatarUrl,
		// 还可以加入其他信息，例如用户名、角色等
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.SuccessResult(resp))
}

// isValidPassword 校验密码是否满足8-12位字母和数字组合
func isValidPassword(password string) bool {
	// 1. 检查长度是否为8-12位
	if len(password) < 8 || len(password) > 12 {
		return false
	}
	// 2. 检查是否只包含字母和数字
	validChars, err := regexp.MatchString(`^[0-9A-Za-z]+$`, password)
	if err != nil || !validChars {
		return false
	}
	// 3. 检查是否至少包含一个字母
	hasLetter, err := regexp.MatchString(`[A-Za-z]`, password)
	if err != nil || !hasLetter {
		return false
	}
	// 4. 检查是否至少包含一个数字
	hasDigit, err := regexp.MatchString(`\d`, password)
	if err != nil || !hasDigit {
		return false
	}

	return true
}
