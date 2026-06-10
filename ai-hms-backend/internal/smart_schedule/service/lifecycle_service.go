package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/config"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/repo"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

// 病人生命周期与院感(决策 26/27)。

// DischargePatient 出组:标记已出组 + 取消其未来未执行排班(决策 27)。
func DischargePatient(g *gorm.DB, tenant, by, patientID int64, reason string) error {
	res := g.Model(&model.PatientProfile{}).Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).
		Updates(map[string]interface{}{
			"PatientStatus": sched.PatientDischarged, "DischargeReason": reason,
			"DischargedAt": time.Now(), "DischargedBy": by,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	today := dayStart(time.Now())
	return g.Model(&model.PatientShift{}).Where(
		`"TenantId" = ? AND "PatientId" = ? AND "TreatmentTime" >= ? AND "Status" IN ?`,
		tenant, patientID, today, []int16{sched.StatusPending, sched.StatusDraft, sched.StatusConfirmed},
	).Updates(map[string]interface{}{"Status": sched.StatusCancelled, "CancelReason": "出组:" + reason}).Error
}

// PlaceNewPatientService 中途入组"录完即排"(决策 27):为单个病人排入 [start,+weeks]。
func PlaceNewPatientService(g *gorm.DB, tenant, patientID int64, start time.Time, weeks int) (placed, conflicts int, err error) {
	var prof model.PatientProfile
	if e := g.Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).First(&prof).Error; e != nil {
		return 0, 0, ErrNotFound
	}
	if prof.IsAdmissionRejected || prof.PatientStatus == sched.PatientDischarged {
		return 0, 0, errors.New("拒收或已出组病人,不予排班")
	}
	anchor := config.AnchorMonday(g, tenant)

	// 启用 HDF 但未定奇偶周 → 简单置 0 并持久化(单人入组,均衡由后续整批生成维护)。
	hdfNeedsInit := prof.HdfEnabled && prof.HdfWeekParity == nil
	if hdfNeedsInit {
		var z int16 = 0
		prof.HdfWeekParity = &z
	}

	end := start.AddDate(0, 0, weeks*7)
	board, e := repo.LoadBoard(g, tenant, anchor, start, end)
	if e != nil {
		return 0, 0, e
	}
	eng := sched.NewEngine(board)
	eng.SpillHorizonDays = config.SpillHorizonDays(g, tenant)
	dates := eng.ExpandDialysisDates(start, weeks)
	fixedHd := eng.PlaceNewPatient(&prof, dates)

	e = g.Transaction(func(tx *gorm.DB) error {
		if hdfNeedsInit {
			if err := tx.Model(&model.PatientProfile{}).Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).Update("HdfWeekParity", 0).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.ScheduleTemplateItem{}).Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).Update("HdfWeekParity", 0).Error; err != nil {
				return err
			}
		}
		if fixedHd != nil {
			if err := tx.Model(&model.PatientProfile{}).Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).
				Update("FixedHdMachineId", *fixedHd).Error; err != nil {
				return err
			}
		}
		if err := repo.SaveDrafts(tx, tenant, board.Drafts); err != nil {
			return err
		}
		return repo.SaveConflicts(tx, tenant, board.Conflicts)
	})
	return len(board.Drafts), len(board.Conflicts), e
}

// SetInfectionStatus 设置院感状态(LIS 同步或手工,决策 26)。
func SetInfectionStatus(g *gorm.DB, tenant, patientID int64, status string) error {
	switch status {
	case sched.InfectionNegative, sched.InfectionPositive, sched.InfectionUnknown:
	default:
		return errors.New("院感状态须为 negative/positive/unknown")
	}
	res := g.Model(&model.Patient{}).Where(`"TenantId" = ? AND "Id" = ?`, tenant, patientID).
		Update("InfectionStatus", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// WaiveInfection 院感无指标时,护士长确认放行上机(决策 26 闸门)。
func WaiveInfection(g *gorm.DB, tenant, by, patientID int64) error {
	now := time.Now()
	res := g.Model(&model.Patient{}).Where(`"TenantId" = ? AND "Id" = ?`, tenant, patientID).
		Updates(map[string]interface{}{"InfectionWaivedBy": by, "InfectionWaivedAt": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// IncompleteItem 资料待补项。
type IncompleteItem struct {
	PatientId   int64    `json:"patientId"`
	PatientName string   `json:"patientName"`
	Missing     []string `json:"missing"`
}

// ListIncompleteProfiles 列出"资料待补"病人(决策 29:缺项提示而非阻断)。
func ListIncompleteProfiles(g *gorm.DB, tenant int64) ([]IncompleteItem, error) {
	var profs []model.PatientProfile
	if err := g.Where(`"TenantId" = ? AND "PatientStatus" = ? AND "IsAdmissionRejected" = false`,
		tenant, sched.PatientActive).Find(&profs).Error; err != nil {
		return nil, err
	}
	names := patientNames(g, tenant)
	var out []IncompleteItem
	for _, p := range profs {
		if p.FreqPattern == sched.FreqTemporary {
			continue
		}
		var miss []string
		if p.HomeWardId == nil {
			miss = append(miss, "归属区")
		}
		if p.ShiftId == nil {
			miss = append(miss, "班次")
		}
		if p.WeeklyCount == 0 {
			miss = append(miss, "每周次数")
		}
		if len(miss) > 0 {
			out = append(out, IncompleteItem{PatientId: p.PatientId, PatientName: names[p.PatientId], Missing: miss})
		}
	}
	return out, nil
}
