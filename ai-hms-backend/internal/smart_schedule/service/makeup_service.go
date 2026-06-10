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

// MakeupResult 补透结果摘要。
type MakeupResult struct {
	Placed    int `json:"placed"`    // 已补排的次数(草稿)
	Conflicts int `json:"conflicts"` // 无空机入冲突队列的次数
}

// MakeupPatient 为某病人在 [weekStart,+weeks] 内补排所有"应排未排"的透析次(决策 13:人工触发)。
// 仅填补缺口(该日该班无有效排班),生成草稿待确认;无空机则报警入冲突队列。
func MakeupPatient(g *gorm.DB, tenant, patientID int64, weekStart time.Time, weeks int) (*MakeupResult, error) {
	var prof model.PatientProfile
	if err := g.Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, patientID).First(&prof).Error; err != nil {
		return nil, errors.New("病人排班骨架不存在")
	}
	if prof.IsAdmissionRejected {
		return nil, errors.New("该病人为拒收状态")
	}
	if prof.HomeWardId == nil || prof.ShiftId == nil {
		return nil, errors.New("病人缺归属区或班次,无法补排")
	}
	if prof.FreqPattern == sched.FreqTemporary {
		return nil, errors.New("临时病人无固定应排次数,不适用补排")
	}

	anchor := config.AnchorMonday(g, tenant)
	start := dayStart(weekStart)
	end := start.AddDate(0, 0, weeks*7)
	board, err := repo.LoadBoard(g, tenant, anchor, start, end)
	if err != nil {
		return nil, err
	}

	ward, shift := *prof.HomeWardId, *prof.ShiftId
	res := &MakeupResult{}
	for i := 0; i < weeks*7; i++ {
		d := start.AddDate(0, 0, i)
		if !sched.IsDue(prof.FreqPattern, d) || !board.IsDialysisDay(d) {
			continue
		}
		if patientBookedAt(g, tenant, patientID, d, shift) {
			continue // 该日已有有效排班,不重复
		}
		mode := sched.DecideMode(anchor, prof.DefaultMode, prof.HdfEnabled, prof.HdfWeekday, prof.HdfWeekParity, d)
		m := board.FindFreeForMode(ward, shift, d, mode)
		if m == nil {
			pid := patientID
			dd := d
			wid := ward
			sid := shift
			_ = raiseConflictDB(g, tenant, &pid, &dd, &sid, &wid, sched.ConflictMakeupSuggest, sched.SeverityAlert, "补透:本区本班无空机")
			res.Conflicts++
			continue
		}
		shiftID, machineID := shift, m.Id
		rec := &model.PatientShift{
			BaseModel:    model.BaseModel{TenantId: tenant},
			PatientId:    patientID,
			ScheduleDate: d,
			ShiftId:      shiftID,
			WardId:       m.WardId,
			MachineId:    machineID,
			Status:       sched.StatusDraft,
			DialysisMode: mode,
			SourceType:   sched.SourceRegular,
			RecordForm:   sched.RecordFormRegular,
		}
		if err := g.Create(rec).Error; err != nil {
			if isUniqueViolation(err) {
				continue // 并发:该位/该次已被占,跳过
			}
			return nil, err
		}
		board.MarkOccupied(m.Id, sched.Cell{WardId: m.WardId, ShiftId: shift, Date: d}, patientID)
		res.Placed++
	}
	return res, nil
}
