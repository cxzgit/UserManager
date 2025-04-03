package service

import (
	"UserManager/src/mapper"
	"fmt"
	"time"
)

type HomeService struct {
	Mapper *mapper.HomeMapper
}

// TrendData 用于返回给前端的访问趋势数据
type TrendData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// 仪表数据
type DashboardData struct {
	RegisteredUsers        int     `json:"registered_users"`
	Visits                 int     `json:"visits"`
	DeactivatedUsers       int     `json:"deactivated_users"`
	RegisteredUsersGrowth  float64 `json:"registered_users_growth"`
	VisitsGrowth           float64 `json:"visits_growth"`
	DeactivatedUsersGrowth float64 `json:"deactivated_users_growth"`
}

func NewHomeService(hm *mapper.HomeMapper) *HomeService {
	return &HomeService{
		Mapper: hm,
	}
}

// 获取仪表盘统计数据
func (hs *HomeService) GetDashboardStats() (DashboardData, error) {
	// 获取当前时间
	now := time.Now()
	currentUsers, err := hs.Mapper.CountRegisteredUsers(now)
	fmt.Println("currentUsers: ", currentUsers)
	if err != nil {
		return DashboardData{}, err
	}
	currentVisits, err := hs.Mapper.CountVisits(time.Now())

	fmt.Println("currentVisits: ", currentVisits)
	if err != nil {
		return DashboardData{}, err
	}
	currentDeactivated, err := hs.Mapper.CountDeactivatedUsers(time.Now())
	fmt.Println("currentDeactivated: ", currentDeactivated)
	if err != nil {
		return DashboardData{}, err
	}

	//获取上个月的数据

	// 获取当前年份和月份
	year, month, _ := now.Date()

	// 计算上个月的年份和月份
	if month == time.January {
		year--
		month = time.December
	} else {
		month--
	}

	// 获取上个月的第一天
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())

	// 上个月的最后一天是本月第一天的前一天
	lastMonth := firstOfMonth.AddDate(0, 1, -1)
	previousUsers, err := hs.Mapper.CountRegisteredUsers(lastMonth)
	if err != nil {
		return DashboardData{}, err
	}
	previousVisits, err := hs.Mapper.CountVisits(lastMonth)
	if err != nil {
		return DashboardData{}, err
	}
	previousDeactivated, err := hs.Mapper.CountDeactivatedUsers(lastMonth)
	if err != nil {
		return DashboardData{}, err
	}

	// 计算增长率
	usersGrowth := calculateGrowth(previousUsers, currentUsers)
	visitsGrowth := calculateGrowth(previousVisits, currentVisits)
	deactivatedGrowth := calculateGrowth(previousDeactivated, currentDeactivated)

	return DashboardData{
		RegisteredUsers:        currentUsers,
		Visits:                 currentVisits,
		DeactivatedUsers:       currentDeactivated,
		RegisteredUsersGrowth:  usersGrowth,
		VisitsGrowth:           visitsGrowth,
		DeactivatedUsersGrowth: deactivatedGrowth,
	}, nil
}

// calculateGrowth 计算增长率
func calculateGrowth(previous, current int) float64 {
	if previous == 0 {
		if current > 0 {
			return 100.0 // 从0增长到正数，视为100%增长
		}
		return 0.0 // 无变化
	}
	return (float64(current-previous) / float64(previous)) * 100
}
func (hs *HomeService) AddVisitCounts(userID int) error {
	err := hs.Mapper.AddVisitCounts(userID)
	if err != nil {
		return err
	}
	return nil
}

// 根据传入天数获取访问趋势数据
func (hs *HomeService) GetAccessTrends(days int) ([]TrendData, error) {
	// 计算起始时间（包含当天）
	startTime := time.Now().AddDate(0, 0, -days+1)

	visits, err := hs.Mapper.GetVisitsFrom(startTime)
	if err != nil {
		return nil, err
	}

	// 初始化每天的访问计数（key 为日期字符串）
	trendMap := make(map[string]int)
	for i := 0; i < days; i++ {
		dateStr := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		trendMap[dateStr] = 0
	}

	// 遍历访问记录进行计数
	for _, visit := range visits {
		dateStr := visit.VisitTime.Format("2006-01-02")
		trendMap[dateStr]++
	}

	// 按时间顺序组织返回数据（例如从最早到最新）
	var trends []TrendData
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		trends = append(trends, TrendData{
			Date:  date,
			Count: trendMap[date],
		})
	}

	return trends, nil
}
