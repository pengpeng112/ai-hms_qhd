package schedule_engine

import "time"

// Engine 排班规则引擎(规范 §6)
type Engine struct {
	Board            *Board
	SpillHorizonDays int
}

// NewEngine 构造引擎
func NewEngine(board *Board) *Engine {
	return &Engine{
		Board:            board,
		SpillHorizonDays: DefaultSpillHorizon,
	}
}

// Generate 主流程: 模板复制+两轮分配, 生成 dates 范围内的草稿(规范 §6, 决策11)
// items 为生效模板骨架(WardId/ShiftId 非空者参与规律生成)
func (e *Engine) Generate(items []SessionItem, dates []time.Time) GenerateResult {
	result := GenerateResult{
		StartDate:    dates[0].Format("2006-01-02"),
		Weeks:        len(dates) / 7,
		DialysisDays: len(dates),
	}

	// 按(Ward, Shift)分组
	type wsKey struct{ ward, shift int64 }
	groups := map[wsKey][]SessionItem{}
	for _, it := range items {
		if it.FreqPattern == FreqTemporary {
			continue
		}
		if it.WardID == nil || it.ShiftID == nil {
			continue
		}
		key := wsKey{*it.WardID, *it.ShiftID}
		groups[key] = append(groups[key], it)
	}

	beds := e.Board.Beds
	shifts := e.Board.Shifts
	occupied := e.Board.Occupied

	// 标记已占用
	markOccupied := func(bedID int64, cell Cell, pid int64) {
		occupied[bedID] = append(occupied[bedID], Occupancy{
			BedID: bedID, Date: cell.Date, ShiftID: cell.ShiftID, PatientID: pid,
		})
	}

	for _, d := range dates {
		if e.Board.NonDialysisDays[d.Format("2006-01-02")] {
			continue
		}
		for _, ward := range e.Board.Wards {
			wardID := ward.WardID
			for _, shift := range shifts {
				if shift.IsDisabled {
					continue
				}
				key := wsKey{wardID, shift.ID}
				items, ok := groups[key]
				if !ok {
					continue
				}

				cell := Cell{WardID: wardID, Date: d, ShiftID: shift.ID}
				var sessions []SessionItem

				for _, it := range items {
					if !IsDue(it.FreqPattern, d) {
						continue
					}
					if PatientHasShift(occupied, it.PatientID, d, shift.ID) {
						continue
					}
					mode := DecideMode(e.Board.Anchor, it.HdfEnabled, it.HdfWeekday, it.HdfWeekParity, d)
					sessions = append(sessions, SessionItem{
						PatientID:     it.PatientID,
						Mode:          mode,
						FreqPattern:   it.FreqPattern,
						FixedHdBedID:  it.FixedHdBedID,
						FixedHdfBedID: it.FixedHdfBedID,
						HdfEnabled:    it.HdfEnabled,
						HdfWeekday:    it.HdfWeekday,
						HdfWeekParity: it.HdfWeekParity,
						TemplateItemID: it.TemplateItemID,
						WardID:        it.WardID,
						ShiftID:       it.ShiftID,
						PatientPlanID: it.PatientPlanID,
					})
				}

				if len(sessions) == 0 {
					continue
				}
				e.placeCell(beds, occupied, cell, sessions, markOccupied, shifts, &result)
			}
		}
	}
	return result
}

// placeCell 两轮分配(规范 §4.1): 先HDF次进HDF机, 再HD次
func (e *Engine) placeCell(
	beds []BedInfo,
	occupied map[int64][]Occupancy,
	cell Cell,
	sessions []SessionItem,
	markOccupied func(int64, Cell, int64),
	shifts []ShiftInfo,
	result *GenerateResult,
) {
	// 第一轮: HDF次
	for _, s := range sessions {
		if s.Mode != ModeHDF {
			continue
		}
		bed := PlaceHdfSession(beds, occupied, cell.WardID, cell, s)
		if bed == nil {
			// HDF无空闲机→冲突
			result.Conflicts = append(result.Conflicts, Conflict{
				PatientID:    s.PatientID,
				Date:         cell.Date,
				ShiftID:      cell.ShiftID,
				WardID:       cell.WardID,
				ConflictType: ConflictHdfNoMachine,
				Severity:     SeverityAlert,
				Detail:       "无空闲HDF机",
			})
			continue
		}
		markOccupied(bed.ID, cell, s.PatientID)
		planID := int64(0)
		if s.PatientPlanID != nil {
			planID = *s.PatientPlanID
		}
		result.Drafts = append(result.Drafts, DraftResult{
			PatientID:      s.PatientID,
			Date:           cell.Date,
			ShiftID:        cell.ShiftID,
			WardID:         cell.WardID,
			BedID:          bed.ID,
			DialysisMode:   ModeHDF,
			Status:         StatusDraft,
			SourceType:     SourceRegular,
			RecordForm:     RecordFormRegular,
			TemplateItemID: s.TemplateItemID,
			PatientPlanID:  &planID,
			ShiftTiming:    20,
		})
	}

	// 第二轮: HD次
	for _, s := range sessions {
		if s.Mode == ModeHDF {
			continue
		}
		bed := PlaceHdSession(beds, occupied, cell.WardID, cell, s)
		if bed != nil {
			markOccupied(bed.ID, cell, s.PatientID)
			planID := int64(0)
			if s.PatientPlanID != nil {
				planID = *s.PatientPlanID
			}
			result.Drafts = append(result.Drafts, DraftResult{
				PatientID:      s.PatientID,
				Date:           cell.Date,
				ShiftID:        cell.ShiftID,
				WardID:         cell.WardID,
				BedID:          bed.ID,
				DialysisMode:   s.Mode,
				Status:         StatusDraft,
				SourceType:     SourceRegular,
				RecordForm:     RecordFormRegular,
				TemplateItemID: s.TemplateItemID,
				PatientPlanID:  &planID,
				ShiftTiming:    20,
			})
			continue
		}
		// 排满→顺延
		isDD := func(t time.Time) bool {
			return !e.Board.NonDialysisDays[t.Format("2006-01-02")]
		}
		spill := SpillToLaterSlot(beds, occupied, s, cell, shifts, e.SpillHorizonDays, isDD)
		if spill != nil {
			spill.IsSpilled = true
			result.Drafts = append(result.Drafts, *spill)
			result.Conflicts = append(result.Conflicts, Conflict{
				PatientID:    s.PatientID,
				Date:         cell.Date,
				ShiftID:      cell.ShiftID,
				WardID:       cell.WardID,
				ConflictType: ConflictSlotSpilled,
				Severity:     SeverityHint,
				Detail:       "排满已顺延到后续时段, 请确认",
				SuggestedDate:        &spill.Date,
				SuggestedShiftID:     &spill.ShiftID,
				SuggestedBedID:       &spill.BedID,
			})
			continue
		}
		// 完全无位
		result.Conflicts = append(result.Conflicts, Conflict{
			PatientID:    s.PatientID,
			Date:         cell.Date,
			ShiftID:      cell.ShiftID,
			WardID:       cell.WardID,
			ConflictType: ConflictNoMachine,
			Severity:     SeverityAlert,
			Detail:       "本区本班次无可用机位, 已顺延搜索仍无结果",
		})
	}
}
