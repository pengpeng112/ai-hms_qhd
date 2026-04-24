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

const (
	legacyPatientTable      = `"Register_PatientInfomation"`
	legacyPatientShiftTable = `"Schedule_PatientShift"`
	legacyShiftTable        = `"Schedule_Shift"`
	legacyTreatmentTable    = `"Treatment_Treatment"`
	legacyEquipmentTable    = `"Auxiliary_EquipmentInfomation"`
	inventoryTable          = "inventory_items"
)

func NewDashboardService() *DashboardService {
	return &DashboardService{db: database.GetDB()}
}

func countRows(query *gorm.DB) (int64, error) {
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// HourlyCount 按小时的计数
type HourlyCount struct {
	Name  string `json:"name"` // "08:00"
	Value int    `json:"value"`
}

// DashboardStats 看板统计汇总
type DashboardStats struct {
	ActivePatients   int64         `json:"activePatients"`   // 在透患者数
	ShiftCount       int64         `json:"shiftCount"`       // 启用班次数量
	EquipmentCount   int64         `json:"equipmentCount"`   // 启用设备数量
	TodaySchedules   int64         `json:"todaySchedules"`   // 今日排班数量
	TodayTreatments  int64         `json:"todayTreatments"`  // 今日透析次数
	AlertItems       int64         `json:"alertItems"`       // 告警：库存不足 + 设备异常
	TreatmentsByHour []HourlyCount `json:"treatmentsByHour"` // 今日按时段分布（用于图表）
	QualityByHour    []HourlyCount `json:"qualityByHour"`    // 近7日每天完成次数（用于质量图表）
}

// GetStats 获取看板统计数据
func (s *DashboardService) GetStats() (*DashboardStats, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	// 1. 今日透析次数（真实表 Treatment_Treatment）
	todayTreatmentsQuery := s.db.Table(legacyTreatmentTable).
		Where(`"TenantId" = ?`, legacyTenantID).
		Where(`DATE("StartTime") = ?`, today)
	todayTreatments, err := countRows(todayTreatmentsQuery)
	if err != nil {
		return nil, err
	}

	// 2. 在科患者数（真实表 Register_PatientInfomation）
	activePatientsQuery := s.db.Table(legacyPatientTable).
		Where(`"TenantId" = ?`, legacyTenantID)
	activePatients, err := countRows(activePatientsQuery)
	if err != nil {
		return nil, err
	}

	// 3. 启用班次数量（真实表 Schedule_Shift）
	shiftCountQuery := s.db.Table(legacyShiftTable).
		Where(`"TenantId" = ?`, legacyTenantID).
		Where(`COALESCE("IsDisabled", false) = false`)
	shiftCount, err := countRows(shiftCountQuery)
	if err != nil {
		return nil, err
	}

	// 4. 启用设备数量（真实表 Auxiliary_EquipmentInfomation）
	equipmentCountQuery := s.db.Table(legacyEquipmentTable).
		Where(`"TenantId" = ?`, legacyTenantID).
		Where(`COALESCE("IsDisabled", false) = false`)
	equipmentCount, err := countRows(equipmentCountQuery)
	if err != nil {
		return nil, err
	}

	// 5. 今日排班数量（真实表 Schedule_PatientShift）
	todaySchedulesQuery := s.db.Table(legacyPatientShiftTable).
		Where(`"TenantId" = ?`, legacyTenantID).
		Where(`DATE("TreatmentTime") = ?`, today)
	todaySchedules, err := countRows(todaySchedulesQuery)
	if err != nil {
		return nil, err
	}

	// 6. 异常告警：当前仅保留本地库存告警
	var inventoryAlerts int64
	if s.db.Migrator().HasTable(inventoryTable) {
		inventoryAlertsQuery := s.db.Table(inventoryTable).
			Where("is_disabled = false AND min_stock > 0 AND stock < min_stock")
		inventoryAlerts, err = countRows(inventoryAlertsQuery)
		if err != nil {
			return nil, err
		}
	}
	alertItems := inventoryAlerts

	// 7. 今日按小时分布（按真实上机时间 StartTime 统计）
	slots := []string{"08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00"}
	byHour := make([]HourlyCount, 0, len(slots))
	for _, slot := range slots {
		parsedTime, _ := time.Parse("15:04", slot)
		hour := parsedTime.Hour()
		byHourQuery := s.db.Table(legacyTreatmentTable).
			Where(`"TenantId" = ?`, legacyTenantID).
			Where(`DATE("StartTime") = ? AND EXTRACT(hour FROM "StartTime") = ?`, today, hour)
		count, err := countRows(byHourQuery)
		if err != nil {
			return nil, err
		}
		byHour = append(byHour, HourlyCount{Name: slot, Value: int(count)})
	}

	// 8. 近7日每日已结束透析次数（状态 60 = 已结束）
	qualityByDay := make([]HourlyCount, 0, 7)
	for i := 6; i >= 0; i-- {
		dayTime := now.AddDate(0, 0, -i)
		day := dayTime.Format("2006-01-02")
		label := fmt.Sprintf("%s", dayTime.Format("01/02"))
		qualityByDayQuery := s.db.Table(legacyTreatmentTable).
			Where(`"TenantId" = ?`, legacyTenantID).
			Where(`DATE("StartTime") = ? AND "Status" = ?`, day, 60)
		count, err := countRows(qualityByDayQuery)
		if err != nil {
			return nil, err
		}
		qualityByDay = append(qualityByDay, HourlyCount{Name: label, Value: int(count)})
	}

	return &DashboardStats{
		ActivePatients:   activePatients,
		ShiftCount:       shiftCount,
		EquipmentCount:   equipmentCount,
		TodaySchedules:   todaySchedules,
		TodayTreatments:  todayTreatments,
		AlertItems:       alertItems,
		TreatmentsByHour: byHour,
		QualityByHour:    qualityByDay,
	}, nil
}
