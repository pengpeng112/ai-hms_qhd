package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

// CRRT 落位(决策 18 / 规范 §10):C 区特殊排班,不走三班/模板/固定机位,
// 只记机器 + 起止时间(可跨班跨天)。主记录走状态机(RecordForm=CRRT),扩展记录存时间区间。

var farFuture = time.Date(2999, 1, 1, 0, 0, 0, 0, time.UTC)

func overlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

// crrtMachineFree CRRT 机在 [start,end] 是否无重叠占用(排除取消/缺席)。
func crrtMachineFree(g *gorm.DB, tenant, machineID int64, start, end time.Time) bool {
	var sessions []model.CrrtSession
	g.Joins(`JOIN "Schedule_v2_PatientShift" ps ON ps."Id" = "Schedule_v2_CrrtSession"."PatientShiftId"`).
		Where(`"Schedule_v2_CrrtSession"."TenantId" = ? AND "Schedule_v2_CrrtSession"."MachineId" = ? AND ps."Status" NOT IN ?`,
			tenant, machineID, []int16{sched.StatusCancelled, sched.StatusAbsent}).
		Find(&sessions)
	for _, s := range sessions {
		e := farFuture
		if s.EndAt != nil {
			e = *s.EndAt
		}
		if overlap(start, end, s.StartAt, e) {
			return false
		}
	}
	return true
}

// InsertCrrt 安排一次 CRRT 治疗。machineID=0 时在该(C)区自动选一台空闲 CRRT 机。
func InsertCrrt(g *gorm.DB, tenant, patientID, wardID, machineID int64, startAt time.Time, endAt *time.Time) (*model.PatientShift, error) {
	var ward model.Ward
	if err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, wardID).First(&ward).Error; err != nil {
		return nil, errors.New("病区不存在")
	}
	if ward.ZoneType != sched.ZoneC {
		return nil, errors.New("CRRT 须安排在 C 区(全警戒区)")
	}
	end := farFuture
	if endAt != nil {
		end = *endAt
	}

	// 选机:指定则校验,否则在本区找空闲 CRRT 机。
	var chosen *model.Machine
	if machineID > 0 {
		var m model.Machine
		if err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, machineID).First(&m).Error; err != nil {
			return nil, errors.New("机器不存在")
		}
		if m.MachineType != sched.MachineCRRT {
			return nil, errors.New("所选机器非 CRRT 机")
		}
		if !crrtMachineFree(g, tenant, m.Id, startAt, end) {
			return nil, ErrOccupied
		}
		chosen = &m
	} else {
		var machines []model.Machine
		g.Where(`"TenantId" = ? AND "WardId" = ? AND "MachineType" = ? AND "IsDisabled" = false`,
			tenant, wardID, sched.MachineCRRT).Order(`"PositionIndex"`).Find(&machines)
		for i := range machines {
			if crrtMachineFree(g, tenant, machines[i].Id, startAt, end) {
				chosen = &machines[i]
				break
			}
		}
		if chosen == nil {
			pid := patientID
			d := dayStart(startAt)
			wid := wardID
			raiseConflictDB(g, tenant, &pid, &d, nil, &wid, sched.ConflictNoMachine, sched.SeverityAlert, "CRRT:本区无空闲 CRRT 机")
			return nil, ErrNoSlot
		}
	}

	var rec *model.PatientShift
	err := g.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		rec = &model.PatientShift{
			BaseModel:    model.BaseModel{TenantId: tenant},
			PatientId:    patientID,
			ScheduleDate: dayStart(startAt),
			ShiftId:      nil, // CRRT 不走三班
			WardId:       wardID,
			MachineId:    &chosen.Id,
			Status:       sched.StatusConfirmed,
			DialysisMode: sched.ModeCRRT,
			SourceType:   sched.SourceTemporary,
			RecordForm:   sched.RecordFormCRRT,
			Confirm1At:   &now,
		}
		if err := tx.Create(rec).Error; err != nil {
			return err
		}
		cs := &model.CrrtSession{
			BaseModel: model.BaseModel{TenantId: tenant}, PatientShiftId: rec.Id,
			MachineId: chosen.Id, StartAt: startAt, EndAt: endAt,
		}
		return tx.Create(cs).Error
	})
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// CrrtItem CRRT 占用展示项。
type CrrtItem struct {
	Id          int64      `json:"id"`
	PatientId   int64      `json:"patientId"`
	PatientName string     `json:"patientName"`
	MachineId   int64      `json:"machineId"`
	MachineCode string     `json:"machineCode"`
	WardId      int64      `json:"wardId"`
	StartAt     time.Time  `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	Status      int16      `json:"status"`
}

// ListCrrt 列出与某日有交叠的 CRRT 占用(进行中或当日)。
func ListCrrt(g *gorm.DB, tenant int64, date time.Time) ([]CrrtItem, error) {
	dayS := dayStart(date)
	dayE := dayS.AddDate(0, 0, 1)

	var sessions []model.CrrtSession
	if err := g.Where(`"TenantId" = ?`, tenant).Find(&sessions).Error; err != nil {
		return nil, err
	}
	names := patientNames(g, tenant)

	// 机器码
	var machines []model.Machine
	g.Where(`"TenantId" = ?`, tenant).Find(&machines)
	mcode := map[int64]string{}
	for _, m := range machines {
		mcode[m.Id] = m.Code
	}

	var out []CrrtItem
	for _, cs := range sessions {
		e := farFuture
		if cs.EndAt != nil {
			e = *cs.EndAt
		}
		if !overlap(cs.StartAt, e, dayS, dayE) {
			continue
		}
		var ps model.PatientShift
		if err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, cs.PatientShiftId).First(&ps).Error; err != nil {
			continue
		}
		if ps.Status == sched.StatusCancelled {
			continue
		}
		out = append(out, CrrtItem{
			Id: cs.Id, PatientId: ps.PatientId, PatientName: names[ps.PatientId],
			MachineId: cs.MachineId, MachineCode: mcode[cs.MachineId], WardId: ps.WardId,
			StartAt: cs.StartAt, EndAt: cs.EndAt, Status: ps.Status,
		})
	}
	return out, nil
}
