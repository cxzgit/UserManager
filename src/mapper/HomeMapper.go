package mapper

import (
	"UserManager/src/models"
	"database/sql"
	"time"
)

type HomeMapper struct {
	DB *sql.DB
}

func NewHomeMapper(db *sql.DB) *HomeMapper {
	return &HomeMapper{DB: db}
}

// 查询注册的数量
func (hm *HomeMapper) CountRegisteredUsers(t time.Time) (int, error) {
	var count int
	err := hm.DB.QueryRow("SELECT COUNT(*) FROM users where created_at< ?", t).Scan(&count)
	return count, err
}

// 查询访问数量
func (hm *HomeMapper) CountVisits(t time.Time) (int, error) {
	var count int
	err := hm.DB.QueryRow("SELECT COUNT(*) FROM visits where visit_time < ?", t).Scan(&count)
	return count, err
}

// 查询注销用户数量
func (hm *HomeMapper) CountDeactivatedUsers(t time.Time) (int, error) {
	var count int
	err := hm.DB.QueryRow("SELECT COUNT(*) FROM users WHERE status = ? and created_at< ? ", 0, t).Scan(&count)
	return count, err
}

// 每次登录成功后，访问数量加一
func (hm *HomeMapper) AddVisitCounts(userID int) error {
	// 插入一条记录到 site_visits 表中
	_, err := hm.DB.Exec("INSERT INTO visits (user_id) VALUES (?)", userID)
	return err
}

// 返回指定时间之后的所有访问记录
func (hm *HomeMapper) GetVisitsFrom(startTime time.Time) ([]models.Visit, error) {
	var visits []models.Visit
	rows, err := hm.DB.Query("SELECT id, user_id, visit_time FROM visits WHERE visit_time >= ?", startTime)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var v models.Visit
		if err = rows.Scan(&v.ID, &v.UserID, &v.VisitTime); err != nil {
			return nil, err
		}
		visits = append(visits, v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return visits, err
}
