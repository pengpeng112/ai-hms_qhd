package sched

import (
	"time"

	"github.com/sdsph/dialysis-scheduling/internal/model"
)

// itemFromProfile 把病人骨架 Profile 转成临时模板项,以复用落位逻辑。
func itemFromProfile(p *model.PatientProfile) *model.ScheduleTemplateItem {
	return &model.ScheduleTemplateItem{
		PatientId:         p.PatientId,
		ZoneTag:           p.ZoneTag,
		WardId:            p.HomeWardId,
		ShiftId:           p.ShiftId,
		FreqPattern:       p.FreqPattern,
		FixedHdMachineId:  p.FixedHdMachineId,
		FixedHdfMachineId: p.FixedHdfMachineId,
		HdfEnabled:        p.HdfEnabled,
		HdfWeekday:        p.HdfWeekday,
		HdfWeekParity:     p.HdfWeekParity,
	}
}

// PlaceNewPatient 新病人初始机位分配(决策 9,规范 §4.3):
// 三天同一台 HD 机 → 逐次分配 → 报警。HDF 日单独走 HDF 双固定/就近逻辑。
// 返回新固定的 HD 机 Id(若成功固定),供持久化层写回 Profile.FixedHdMachineId。
func (e *Engine) PlaceNewPatient(p *model.PatientProfile, dates []time.Time) *int64 {
	if p.IsAdmissionRejected {
		return nil // 拒收:不生成排班(规范 §7.2)
	}
	if p.HomeWardId == nil || p.ShiftId == nil {
		// 骨架不完整(无归属区/班次)→ 报警人工补全
		e.raiseConflict(ConflictNewUnplaced, p.PatientId, Cell{WardId: derefI64(p.HomeWardId)}, SeverityAlert, nil)
		return nil
	}
	ward, shift := *p.HomeWardId, *p.ShiftId
	item := itemFromProfile(p)

	// 区分 HD 日与 HDF 日(HDF 日不能落在 HD 固定机上)。
	var hdDates, hdfDates []time.Time
	for _, d := range dates {
		if !IsDue(p.FreqPattern, d) || !e.Board.IsDialysisDay(d) {
			continue
		}
		if DecideMode(e.Board.Anchor, p.HdfEnabled, p.HdfWeekday, p.HdfWeekParity, d) == ModeHDF {
			hdfDates = append(hdfDates, d)
		} else {
			hdDates = append(hdDates, d)
		}
	}

	// HDF 日:逐日走 HDF 双固定/就近(决策 2)。
	for _, d := range hdfDates {
		cell := Cell{ward, shift, d}
		if m := e.placeHDFSession(cell, SessionItem{p.PatientId, ModeHDF, item}); m != nil {
			e.draft(p.PatientId, cell, m, ModeHDF, SourceRegular, nil, false)
		} else {
			e.raiseConflict(ConflictHdfNoMachine, p.PatientId, cell, SeverityAlert, nil)
		}
	}

	// HD 日 step1:找一台所有 HD 日都空闲的同一台 HD 机,一次性固定(连片优先)。
	if len(hdDates) > 0 {
		if m := e.findMachineFreeOnAllDates(ward, shift, MachineHD, hdDates); m != nil {
			for _, d := range hdDates {
				e.draft(p.PatientId, Cell{ward, shift, d}, m, ModeHD, SourceRegular, nil, false)
			}
			return &m.Id // 写回固定 HD 机位
		}
	}

	// HD 日 step2:退而逐次分配(各日分别找空位,可能不同机)。
	for _, d := range hdDates {
		cell := Cell{ward, shift, d}
		if m := e.placeHDSession(cell, SessionItem{p.PatientId, ModeHD, item}); m != nil {
			e.draft(p.PatientId, cell, m, ModeHD, SourceRegular, nil, false)
		} else {
			// step3:排不开 → 报警人工
			e.raiseConflict(ConflictNewUnplaced, p.PatientId, cell, SeverityAlert, nil)
		}
	}
	return nil
}

// findMachineFreeOnAllDates 找一台在所有给定日期(同区同班)都空闲、且机型匹配的机器。
// 候选已按 PositionIndex 升序,首个全空者即返回(倾向连片靠前)。
func (e *Engine) findMachineFreeOnAllDates(ward, shift int64, machineType string, dates []time.Time) *model.Machine {
	for _, m := range e.Board.machinesByWard[ward] {
		if m.MachineType != machineType {
			continue
		}
		allFree := true
		for _, d := range dates {
			if !e.Board.IsFree(m.Id, Cell{ward, shift, d}) {
				allFree = false
				break
			}
		}
		if allFree {
			return m
		}
	}
	return nil
}
