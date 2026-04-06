package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

// DashboardService 看板统计服务
type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService() *DashboardService {
	return &DashboardService{db: database.GetDB()}
}

// HourlyCount 按小时的计数
type HourlyCount struct {
	Name  string `json:"name"`  // "08:00"
	Value int    `json:"value"`
}

// DashboardStats 看板统计汇总
type DashboardStats struct {
	ActivePatients    int64         `json:"activePatients"`    // 在透患者数
	TodayTreatments   int64         `json:"todayTreatments"`   // 今日透析次数
	AlertItems        int64         `json:"alertItems"`        // 告警：库存不足 + 设备异常
	TreatmentsByHour  []HourlyCount `json:"treatmentsByHour"`  // 今日按时段分布（用于图表）
	QualityByHour     []HourlyCount `json:"qualityByHour"`     // 近7日每天完成次数（用于质量图表）
}

// GetStats 获取看板统计数据
func (s *DashboardService) GetStats() (*DashboardStats, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	today := time.Now().Format("2006-01-02")

	// 1. 今日透析次数
	var todayTreatments int64
	s.db.Table("treatments").
		Where("DATE(created_at) = ? AND status != 'cancelled'", today).
		Count(&todayTreatments)

	// 2. 在透患者数（status=active）
	var activePatients int64
	s.db.Table("patients").Where("status = 'active'").Count(&activePatients)

	// 3. 异常告警：库存告警 + 设备告警
	var inventoryAlerts int64
	s.db.Table("inventory_items").
		Where("is_disabled = false AND min_stock > 0 AND stock < min_stock").
		Count(&inventoryAlerts)
	var deviceAlerts int64
	s.db.Table("devices").
		Where("is_disabled = false AND status IN ('warning', 'alarm')").
		Count(&deviceAlerts)
	alertItems := inventoryAlerts + deviceAlerts

	// 4. 今日按小时分布（治疗创建时间统计）
	slots := []string{"08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00"}
	byHour := make([]HourlyCount, 0, len(slots))
	for _, slot := range slots {
		hour := slot[:2]
		var count int64
		s.db.Table("treatments").
			Where("DATE(created_at) = ? AND EXTRACT(hour FROM created_at) = ? AND status != 'cancelled'",
				today, hour).
			Count(&count)
		byHour = append(byHour, HourlyCount{Name: slot, Value: int(count)})
	}

	// 5. 近7日每日完成透析次数（用于质量统计图表）
	qualityByDay := make([]HourlyCount, 0, 7)
	for i := 6; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		label := fmt.Sprintf("%s", time.Now().AddDate(0, 0, -i).Format("01/02"))
		var count int64
		s.db.Table("treatments").
			Where("DATE(created_at) = ? AND status = 'completed'", day).
			Count(&count)
		qualityByDay = append(qualityByDay, HourlyCount{Name: label, Value: int(count)})
	}

	return &DashboardStats{
		ActivePatients:   activePatients,
		TodayTreatments:  todayTreatments,
		AlertItems:       alertItems,
		TreatmentsByHour: byHour,
		QualityByHour:    qualityByDay,
	}, nil
}
