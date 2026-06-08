package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/schedule_engine"
	"gorm.io/gorm"
)

// ScheduleConfirmService 排班确认服务
type ScheduleConfirmService struct {
	db *gorm.DB
}

func NewScheduleConfirmService() *ScheduleConfirmService {
	return &ScheduleConfirmService{db: database.GetDB()}
}

// ConfirmPlan 第一次确认(护士长): 周期内草稿→已确认
func (s *ScheduleConfirmService) ConfirmPlan(tenantID, confirmBy int64, weekStart time.Time, weeks int) (int64, error) {
	if s.db == nil {
		return 0, errors.New("database not available")
	}
	weekStart = dateOnly(weekStart)
	end := weekStart.AddDate(0, 0, weeks*7)
	now := time.Now()

	res := s.db.Model(&models.PatientShift{}).
		Where(`"TenantId" = ? AND "ScheduleDate" >= ? AND "ScheduleDate" < ? AND "Status" = ?`,
			tenantID, weekStart, end, schedule_engine.StatusDraft).
		Updates(map[string]interface{}{
			"Status": schedule_engine.StatusConfirmed,
		})

	// 同步更新确认信息到 PatientShiftExt
	var shifts []models.PatientShift
	s.db.Where(`"TenantId" = ? AND "ScheduleDate" >= ? AND "ScheduleDate" < ?`,
		tenantID, weekStart, end).Find(&shifts)
	for _, ps := range shifts {
		s.db.Model(&models.PatientShiftExt{}).
			Where(`"TenantId" = ? AND "PatientShiftId" = ?`, tenantID, ps.Id).
			Updates(map[string]interface{}{
				"Confirm1At": now,
				"Confirm1By": confirmBy,
			})
	}

	return res.RowsAffected, res.Error
}

// ConfirmDay 第二/三次确认: 对某日已确认(20)的排班更新 Confirm2At 或 Confirm3At
func (s *ScheduleConfirmService) ConfirmDay(tenantID, confirmBy int64, date time.Time, level int) (int64, error) {
	if s.db == nil {
		return 0, errors.New("database not available")
	}
	if level != 2 && level != 3 {
		return 0, errors.New("level must be 2 or 3")
	}

	col := "Confirm2At"
	byCol := "Confirm2By"
	if level == 3 {
		col = "Confirm3At"
		byCol = "Confirm3By"
	}

	now := time.Now()
	d := dateOnly(date)

	// 更新 PatientShift status
	s.db.Model(&models.PatientShift{}).
		Where(`"TenantId" = ? AND "ScheduleDate" = ? AND "Status" = ?`, tenantID, d, schedule_engine.StatusConfirmed).
		Update("Status", schedule_engine.StatusConfirmed) // 保持已确认, 但写入最终锁定标记

	// 更新 PatientShiftExt 确认时间
	res := s.db.Model(&models.PatientShiftExt{}).
		Joins(`JOIN "Schedule_PatientShift" ps ON ps."Id" = "Schedule_PatientShiftExt"."PatientShiftId"`).
		Where(`"Schedule_PatientShiftExt"."TenantId" = ? AND ps."ScheduleDate" = ? AND ps."Status" = ?`,
			tenantID, d, schedule_engine.StatusConfirmed).
		Updates(map[string]interface{}{
			col:    now,
			byCol:  confirmBy,
		})

	return res.RowsAffected, res.Error
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
