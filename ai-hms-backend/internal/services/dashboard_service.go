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
	legacyOutcomeTable      = `"Register_OutCome"`
	legacyPatientShiftTable = `"Schedule_PatientShift"`
	legacyShiftTable        = `"Schedule_Shift"`
	legacyTreatmentTable    = `"Treatment_Treatment"`
	legacyEquipmentTable    = `"Auxiliary_EquipmentInfomation"`
	legacyDMLogTable        = `"Device_DMLog"`
	inventoryTable          = "inventory_items"
	treatmentBusinessDate   = `DATE(COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime"))`
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
	ActivePatients      int64         `json:"activePatients"`
	ShiftCount          int64         `json:"shiftCount"`
	EquipmentCount      int64         `json:"equipmentCount"`
	TodaySchedules      int64         `json:"todaySchedules"`
	TodayTreatments     int64         `json:"todayTreatments"`
	RunningTreatments   int64         `json:"runningTreatments"`
	CompletedTreatments int64         `json:"completedTreatments"`
	AlertItems          int64         `json:"alertItems"`
	TreatmentsByHour    []HourlyCount `json:"treatmentsByHour"`
	QualityByHour       []HourlyCount `json:"qualityByHour"`
	AvgDialysisHours    float64       `json:"avgDialysisHours"`
}

// GetStats 获取看板统计数据
func (s *DashboardService) GetStats() (*DashboardStats, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	// 1. 今日透析次数（真实表 Treatment_Treatment）
	// 老库未上机记录常只有 SignInTime/CreateTime，不能仅按 StartTime 统计。
	todayTreatmentsQuery := s.db.Table(legacyTreatmentTable).
		Where(`"TenantId" = ?`, LegacyTenantID).
		Where(treatmentBusinessDate+` = ?`, today)
	todayTreatments, err := countRows(todayTreatmentsQuery)
	if err != nil {
		return nil, err
	}

	// 设备日志治疗状态：取每台设备当天最新 UF 记录。
	latestDMLogQuery := s.db.Table(legacyDMLogTable).
		Select(`DISTINCT ON ("FEPId") "FEPId", "UFVolume", "UFSetVolume", "LogTime"`).
		Where(`DATE("LogTime") = ?`, today).
		Where(`"UFSetVolume" IS NOT NULL AND "UFSetVolume" > 0 AND "UFVolume" IS NOT NULL`).
		Order(`"FEPId", "LogTime" DESC`)
	runningTreatments, err := countRows(s.db.Table(`(?) AS dm`, latestDMLogQuery).
		Where(`dm."UFVolume" > 0 AND dm."UFVolume" < dm."UFSetVolume"`))
	if err != nil {
		return nil, err
	}
	completedTreatments, err := countRows(s.db.Table(`(?) AS dm`, latestDMLogQuery).
		Where(`dm."UFVolume" >= dm."UFSetVolume"`))
	if err != nil {
		return nil, err
	}

	// 2. 在科患者数（与患者列表 onlyActive 保持一致：最新转归 Type=10）
	activeSubquery := s.db.Table(legacyOutcomeTable).
		Select(`DISTINCT ON ("PatientId") "PatientId", "Type"`).
		Where(`"TenantId" = ?`, LegacyTenantID).
		Order(`"PatientId", "OutComeTime" DESC, "CreateTime" DESC`)
	activePatientsQuery := s.db.Table(legacyPatientTable+` AS p`).
		Joins(`INNER JOIN (?) AS oc ON oc."PatientId" = p."Id" AND oc."Type" = '10'`, activeSubquery).
		Where(`p."TenantId" = ?`, LegacyTenantID)
	activePatients, err := countRows(activePatientsQuery)
	if err != nil {
		return nil, err
	}

	// 3. 启用班次数量（真实表 Schedule_Shift）
	shiftCountQuery := s.db.Table(legacyShiftTable).
		Where(`"TenantId" = ?`, LegacyTenantID).
		Where(`COALESCE("IsDisabled", false) = false`)
	shiftCount, err := countRows(shiftCountQuery)
	if err != nil {
		return nil, err
	}

	// 4. 启用设备数量（真实表 Auxiliary_EquipmentInfomation）
	equipmentCountQuery := s.db.Table(legacyEquipmentTable).
		Where(`"TenantId" = ?`, LegacyTenantID).
		Where(`COALESCE("IsDisabled", false) = false`)
	equipmentCount, err := countRows(equipmentCountQuery)
	if err != nil {
		return nil, err
	}

	// 5. 今日排班数量（真实表 Schedule_PatientShift）
	todaySchedulesQuery := s.db.Table(legacyPatientShiftTable).
		Where(`"TenantId" = ?`, LegacyTenantID).
		Where(`"TreatmentTime" >= ? AND "TreatmentTime" < ?`,
			today+" 00:00:00", today+" 23:59:59")
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

	// 7. 今日按小时分布（按治疗业务时间统计）
	slots := []string{"08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00"}
	byHour := make([]HourlyCount, 0, len(slots))
	for _, slot := range slots {
		parsedTime, _ := time.Parse("15:04", slot)
		hour := parsedTime.Hour()
		byHourQuery := s.db.Table(legacyTreatmentTable).
			Where(`"TenantId" = ?`, LegacyTenantID).
			Where(treatmentBusinessDate+` = ?`, today).
			Where(`EXTRACT(hour FROM COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime")) = ?`, hour)
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
			Where(`"TenantId" = ?`, LegacyTenantID).
			Where(treatmentBusinessDate+` = ? AND "Status" = ?`, day, "60")
		count, err := countRows(qualityByDayQuery)
		if err != nil {
			return nil, err
		}
		qualityByDay = append(qualityByDay, HourlyCount{Name: label, Value: int(count)})
	}

	// 今日平均透析时长
	var avgHours float64
	var avgResult struct{ AvgMinutes float64 }
	s.db.Table(legacyTreatmentTable).
		Select(`COALESCE(AVG(COALESCE("RealDuration", 0)), 0) AS "AvgMinutes"`).
		Where(`"TenantId" = ?`, LegacyTenantID).
		Where(treatmentBusinessDate+` = ?`, today).
		Scan(&avgResult)
	avgHours = avgResult.AvgMinutes / 60.0

	return &DashboardStats{
		ActivePatients:      activePatients,
		ShiftCount:          shiftCount,
		EquipmentCount:      equipmentCount,
		TodaySchedules:      todaySchedules,
		TodayTreatments:     todayTreatments,
		RunningTreatments:   runningTreatments,
		CompletedTreatments: completedTreatments,
		AlertItems:          alertItems,
		TreatmentsByHour:    byHour,
		QualityByHour:       qualityByDay,
		AvgDialysisHours:    avgHours,
	}, nil
}
