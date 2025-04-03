package controller

import (
	"UserManager/src/service"
	"UserManager/src/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type HomeController struct {
	Service *service.HomeService
	Tmpl    *template.Template
}

func NewHomeController(s *service.HomeService) *HomeController {
	//加载模板文件，templates/login1.html
	tmpl := template.Must(template.ParseFiles("views/index.html"))
	return &HomeController{
		Service: s,
		Tmpl:    tmpl,
	}
}

// 渲染登录页面
func (hc *HomeController) HomePage(w http.ResponseWriter, r *http.Request) {
	if err := hc.Tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "页面渲染错误", http.StatusInternalServerError)
	}
}

// DashboardStats 处理仪表盘数据请求
func (hc *HomeController) DashboardStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if ok {
		// 记录访问
		if err := hc.Service.AddVisitCounts(userID); err != nil {
			// 如果记录失败，可以记录日志，不必阻止登录流程
			log.Printf("记录访问失败: %v", err)
		}
	}
	stats, err := hc.Service.GetDashboardStats()
	if err != nil {
		http.Error(w, "获取数据失败", http.StatusInternalServerError)
		return
	}
	fmt.Println("数据", stats)

	json.NewEncoder(w).Encode(utils.SuccessResult(stats))
}

func (hc *HomeController) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	// 清除 Cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt_token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0), // 兼容性处理（设置为过去的时间）
	})
	json.NewEncoder(w).Encode(utils.SuccessResult("登出成功"))
}

// 处理 GET /accessTrend 请求
func (hc *HomeController) GetAccessTrend(w http.ResponseWriter, r *http.Request) {
	// 默认统计 7 天数据；获取查询参数 days
	daysStr := r.URL.Query().Get("days")
	if daysStr == "" {
		daysStr = "7"
	}
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		http.Error(w, "无效的天数参数", http.StatusBadRequest)
		return
	}

	trends, err := hc.Service.GetAccessTrends(days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头为 JSON 格式
	w.Header().Set("Content-Type", "application/json")
	// 返回 JSON 数据
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"data": trends}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
