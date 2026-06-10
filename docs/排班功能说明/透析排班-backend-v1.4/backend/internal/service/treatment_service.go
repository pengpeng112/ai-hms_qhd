package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
)

// 上机(开始治疗)与下机(完成),含院感/当日闸门(决策 26)。
// 这是院感安全规则真正生效的执行点:无指标须护士长豁免、确诊传染病仅 C 区可上机。

var (
	ErrNotConfirmed         = errors.New("该排班未确认,不能上机")
	ErrNotToday             = errors.New("只能在透析当日上机")
	ErrNotInDialysis        = errors.New("该排班不在透析中,无法下机")
	ErrInfectionUnconfirmed = errors.New("院感指标未出,需护士长确认(豁免)后方可上机")
	ErrInfectionPositive    = errors.New("确诊传染病,仅 C 区可收治,不可在本区上机")
)

// infectionGate 院感闸门:返回 nil 表示放行。
func infectionGate(g *gorm.DB, tenant, patientID int64) error {
	var pt model.Patient
	if err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, patientID).First(&pt).Error; err != nil {
		return ErrNotFound
	}
	switch pt.InfectionStatus {
	case sched.InfectionNegative:
		return nil
	case sched.InfectionPositive:
		// 确诊阳性:仅 C 区可上机(人道急诊)
		var prof model.PatientProfile
		if err := g.Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).First(&prof).Error; err == nil {
			if prof.ZoneTag == sched.ZoneC {
				return nil
			}
		}
		return ErrInfectionPositive
	default: // unknown:须护士长豁免
		if pt.InfectionWaivedAt != nil {
			return nil
		}
		return ErrInfectionUnconfirmed
	}
}

// StartTreatment 上机(已确认 → 透析中)。闸门:当日 + 已确认 + 院感放行。
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
	return g.Model(s).Updates(map[string]interface{}{
		"Status": sched.StatusInDialysis, "TreatmentStartAt": now, "StartedBy": by,
	}).Error
}

// CompleteTreatment 下机(透析中 → 已完成)。
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
		"Status": sched.StatusCompleted, "TreatmentEndAt": now,
	}).Error
}
