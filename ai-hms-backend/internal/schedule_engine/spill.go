package schedule_engine

import "time"

// SpillToLaterSlot 排满顺延: 后续班次→后续透析日(规范 §9.4, 决策22)
// 返回建议草稿, nil表示完全无位
func SpillToLaterSlot(
	beds []BedInfo,
	occupied map[int64][]Occupancy,
	session SessionItem,
	cell Cell,
	shifts []ShiftInfo,
	horizonDays int,
	isDialysisDay func(time.Time) bool,
) *DraftResult {
	for _, cand := range laterSlots(cell, shifts, horizonDays, isDialysisDay) {
		if PatientHasShift(occupied, session.PatientID, cand.Date, cand.ShiftID) {
			continue
		}
		bed := PlaceHdSession(beds, occupied, cand.WardID, cand, session)
		if bed == nil {
			continue
		}
		status := StatusDraft
		planID := int64(0)
		if session.PatientPlanID != nil {
			planID = *session.PatientPlanID
		}
		return &DraftResult{
			PatientID:      session.PatientID,
			Date:           cand.Date,
			ShiftID:        cand.ShiftID,
			WardID:         cand.WardID,
			BedID:          bed.ID,
			DialysisMode:   session.Mode,
			Status:         status,
			SourceType:     SourceRegular,
			RecordForm:     RecordFormRegular,
			TemplateItemID: session.TemplateItemID,
			PatientPlanID:  &planID,
			ShiftTiming:    20,
			IsSpilled:      true,
		}
	}
	return nil
}

// laterSlots 生成顺延候选时段
func laterSlots(cell Cell, shifts []ShiftInfo, horizonDays int, isDialysisDay func(time.Time) bool) []Cell {
	var out []Cell

	// 1) 同一天, 后续班次
	for _, sh := range shifts {
		if sh.Sort <= shiftSort(cell.ShiftID, shifts) || sh.IsDisabled {
			continue
		}
		out = append(out, Cell{cell.WardID, cell.Date, sh.ID})
	}

	// 2) 后续透析日(窗口内)
	for i := 1; i <= horizonDays; i++ {
		d := dateOnly(cell.Date).AddDate(0, 0, i)
		if isDialysisDay != nil && !isDialysisDay(d) {
			continue
		}
		// 优先原班次
		out = append(out, Cell{cell.WardID, d, cell.ShiftID})
		// 再其他班次
		for _, sh := range shifts {
			if sh.ID == cell.ShiftID || sh.IsDisabled {
				continue
			}
			out = append(out, Cell{cell.WardID, d, sh.ID})
		}
	}
	return out
}

func shiftSort(shiftID int64, shifts []ShiftInfo) int {
	for _, sh := range shifts {
		if sh.ID == shiftID {
			return sh.Sort
		}
	}
	return 0
}
