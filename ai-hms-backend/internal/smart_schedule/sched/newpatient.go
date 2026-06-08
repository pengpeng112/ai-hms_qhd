package sched

import (
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
)

// itemFromProfile 把病人骨架 Profile 转成临时模板项,以复用落位逻辑。
func itemFromProfile(p *model.PatientProfile) *model.ScheduleTemplateItem {
	return &model.ScheduleTemplateItem{
		PatientId:         p.PatientId,
		ZoneTag:           p.ZoneTag,
		WardId:            p.HomeWardId,
		ShiftId:           p.ShiftId,
		FreqPattern:       p.FreqPattern,
		DefaultMode:       p.DefaultMode,
		FixedHdMachineId:  p.FixedHdMachineId,
		FixedHdfMachineId: p.FixedHdfMachineId,
		HdfEnabled:        p.HdfEnabled,
		HdfWeekday:        p.HdfWeekday,
		HdfWeekParity:     p.HdfWeekParity,
	}
}

type dayMode struct {
	d    time.Time
	mode string
}

// PlaceNewPatient 新病人初始机位分配(决策 9/24,规范 §4.3):
// HD 类(HD/HFD/HP)尽量同一台 HD 机固定 → 逐次 → 报警;需 HDF 机的次(HDF/HF)走双固定/就近。
// 返回新固定的 HD 机 Id(若成功固定),供持久化层写回 Profile.FixedHdMachineId。
func (e *Engine) PlaceNewPatient(p *model.PatientProfile, dates []time.Time) *int64 {
	if p.IsAdmissionRejected || p.PatientStatus == PatientDischarged {
		return nil // 拒收或已出组:不生成排班(规范 §7.2,决策 27)
	}
	if p.HomeWardId == nil || p.ShiftId == nil {
		// 骨架不完整(无归属区/班次)→ 报警人工补全
		e.raiseConflict(ConflictNewUnplaced, p.PatientId, Cell{WardId: derefI64(p.HomeWardId)}, SeverityAlert, nil)
		return nil
	}
	ward, shift := *p.HomeWardId, *p.ShiftId
	item := itemFromProfile(p)

	// 按"是否需 HDF 机"分两组,各日带本次实际模式。
	var hdDays, hdfDays []dayMode
	for _, d := range dates {
		if !IsDue(p.FreqPattern, d) || !e.Board.IsDialysisDay(d) {
			continue
		}
		mode := DecideMode(e.Board.Anchor, p.DefaultMode, p.HdfEnabled, p.HdfWeekday, p.HdfWeekParity, d)
		if RequiresHdfMachine(mode) {
			hdfDays = append(hdfDays, dayMode{d, mode})
		} else {
			hdDays = append(hdDays, dayMode{d, mode})
		}
	}

	// 需 HDF 机的次:逐日走双固定/就近(决策 2)。
	for _, x := range hdfDays {
		cell := Cell{ward, shift, x.d}
		if m := e.placeHDFSession(cell, SessionItem{p.PatientId, x.mode, item}); m != nil {
			e.draft(p.PatientId, cell, m, x.mode, SourceRegular, nil, false)
		} else {
			e.raiseConflict(ConflictHdfNoMachine, p.PatientId, cell, SeverityAlert, nil)
		}
	}

	// HD 类 step1:找一台所有 HD 日都空闲的同一台 HD 机,一次性固定(连片优先)。
	if len(hdDays) > 0 {
		alldates := make([]time.Time, len(hdDays))
		for i, x := range hdDays {
			alldates[i] = x.d
		}
		if m := e.findMachineFreeOnAllDates(ward, shift, MachineHD, alldates); m != nil {
			for _, x := range hdDays {
				e.draft(p.PatientId, Cell{ward, shift, x.d}, m, x.mode, SourceRegular, nil, false)
			}
			return &m.Id // 写回固定 HD 机位
		}
	}

	// HD 类 step2:退而逐次分配(各日分别找空位,可能不同机)。
	for _, x := range hdDays {
		cell := Cell{ward, shift, x.d}
		if m := e.placeHDSession(cell, SessionItem{p.PatientId, x.mode, item}); m != nil {
			e.draft(p.PatientId, cell, m, x.mode, SourceRegular, nil, false)
		} else {
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
