package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

var (
	ErrNotConfirmed         = errors.New("该排班未确认,不能上机")
	ErrNotToday             = errors.New("只能在透析当日上机")
	ErrNotInDialysis        = errors.New("该排班不在透析中,无法下机")
	ErrInfectionUnconfirmed = errors.New("院感指标未出,需护士长确认(豁免)后方可上机")
	ErrInfectionPositive    = errors.New("确诊传染病,仅C区可收治,不可在本区上机")
)

func infectionGate(g *gorm.DB, tenant, patientID int64) error {
	var pt model.Patient
	if err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, patientID).First(&pt).Error; err != nil {
		return ErrNotFound
	}
	switch pt.InfectionStatus {
	case sched.InfectionNegative:
		return nil
	case sched.InfectionPositive:
		var prof model.PatientProfile
		if err := g.Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).First(&prof).Error; err == nil {
			if prof.ZoneTag == sched.ZoneC {
				return nil
			}
		}
		return ErrInfectionPositive
	default:
		if pt.InfectionWaivedAt != nil {
			return nil
		}
		return ErrInfectionUnconfirmed
	}
}

func StartTreatment(g *gorm.DB, tenant, by, id int64) error {
	s, err := loadShift(g, tenant, id)
	if err != nil {
		return err
	}
	if s.Status != sched.StatusConfirmed {
		return ErrNotConfirmed
	}
	if s.ScheduleDate.Format("2006-01-02") != time.Now().Format("2006-01-02") {
		return ErrNotToday
	}
	if err := infectionGate(g, tenant, s.PatientId); err != nil {
		return err
	}
	now := time.Now()
	upd := map[string]interface{}{"Status": sched.StatusInDialysis}
	if s.Confirm1At == nil {
		upd["Confirm1At"] = now
		upd["Confirm1By"] = by
	}
	return g.Model(s).Updates(upd).Error
}

func CompleteTreatment(g *gorm.DB, tenant, by, id int64) error {
	s, err := loadShift(g, tenant, id)
	if err != nil {
		return err
	}
	if s.Status != sched.StatusInDialysis {
		return ErrNotInDialysis
	}
	now := time.Now()
	return g.Model(s).Updates(map[string]interface{}{
		"Status":      sched.StatusCompleted,
		"Confirm3At":  now,
		"Confirm3By":  by,
	}).Error
}
