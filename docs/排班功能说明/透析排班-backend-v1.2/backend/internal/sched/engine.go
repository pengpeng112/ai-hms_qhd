package sched

import (
	"time"

	"github.com/sdsph/dialysis-scheduling/internal/model"
)

// Engine 在 Board 快照上执行排班算法。所有"排不开"分支统一写冲突队列,绝不硬排。
type Engine struct {
	Board           *Board
	SpillHorizonDays int // 排满顺延窗口(决策 22),默认 14
}

// NewEngine 构造引擎。
func NewEngine(b *Board) *Engine {
	return &Engine{Board: b, SpillHorizonDays: 14}
}

// SessionItem 一次透析的待排单元(病人 + 本次模式 + 来源模板项)。
type SessionItem struct {
	PatientId int64
	Mode      string
	Item      *model.ScheduleTemplateItem
}

// ParityAssignment 待写回的 HDF 奇偶周分配(供持久化层落 Profile/TemplateItem)。
type ParityAssignment struct {
	PatientId int64
	Parity    int16
}

// ExpandDialysisDates 从 start 起展开 weeks 周内的透析日(跳过非透析日)。
func (e *Engine) ExpandDialysisDates(start time.Time, weeks int) []time.Time {
	var out []time.Time
	for i := 0; i < weeks*7; i++ {
		d := dateOnly(start).AddDate(0, 0, i)
		if e.Board.IsDialysisDay(d) {
			out = append(out, d)
		}
	}
	return out
}

// AssignHdfWeekParity 决策 21/3:已排定者保持不变并计入负载,新入组者分到较轻一侧。
// 直接修改 items 的 HdfWeekParity,并返回需写回持久层的分配。
func (e *Engine) AssignHdfWeekParity(items []*model.ScheduleTemplateItem) []ParityAssignment {
	type key struct {
		weekday int16
		parity  int16
	}
	load := map[key]int{}
	for _, it := range items {
		if it.HdfEnabled && it.HdfWeekday != nil && it.HdfWeekParity != nil {
			load[key{*it.HdfWeekday, *it.HdfWeekParity}]++
		}
	}
	var out []ParityAssignment
	for _, it := range items {
		if !it.HdfEnabled || it.HdfWeekday == nil || it.HdfWeekParity != nil {
			continue // 未启用 / 无 HDF 日 / 已排定(尽量固定)
		}
		even := load[key{*it.HdfWeekday, 0}]
		odd := load[key{*it.HdfWeekday, 1}]
		var chosen int16 = 0
		if even > odd {
			chosen = 1
		}
		it.HdfWeekParity = &chosen
		load[key{*it.HdfWeekday, chosen}]++
		out = append(out, ParityAssignment{PatientId: it.PatientId, Parity: chosen})
	}
	return out
}

// Generate 主流程:模板复制生成 dates 范围内的草稿(规范 §6,决策 11)。
// items 为生效模板的稳定病人骨架(WardId/ShiftId 非空者参与规律生成)。
func (e *Engine) Generate(items []*model.ScheduleTemplateItem, dates []time.Time) {
	e.AssignHdfWeekParity(items)

	// 按 (区,班) 分组,加速逐格查找;仅取骨架完整、非临时频率的项。
	type ws struct{ ward, shift int64 }
	group := map[ws][]*model.ScheduleTemplateItem{}
	for _, it := range items {
		if it.WardId == nil || it.ShiftId == nil || it.FreqPattern == FreqTemporary {
			continue
		}
		k := ws{*it.WardId, *it.ShiftId}
		group[k] = append(group[k], it)
	}

	for _, d := range dates {
		for _, ward := range e.Board.wards {
			if ward.IsDisabled {
				continue
			}
			if !e.Board.WardOpenOn(d, ward.Id) {
				continue // 假日值班模式:该区当天不开放,跳过
			}
			for _, shift := range e.Board.shifts {
				if shift.IsDisabled {
					continue
				}
				cell := Cell{WardId: ward.Id, ShiftId: shift.Id, Date: d}
				var sessions []SessionItem
				for _, it := range group[ws{ward.Id, shift.Id}] {
					if !IsDue(it.FreqPattern, d) {
						continue
					}
					// 已有记录(含已取消/缺席)的病人本次跳过:不重复已排,也不复活已取消(尊重请假)。
					if e.Board.PatientHasSlot(it.PatientId, d, shift.Id) {
						continue
					}
					mode := DecideMode(e.Board.Anchor, it.HdfEnabled, it.HdfWeekday, it.HdfWeekParity, d)
					sessions = append(sessions, SessionItem{PatientId: it.PatientId, Mode: mode, Item: it})
				}
				if len(sessions) > 0 {
					e.placeCell(cell, sessions)
				}
			}
		}
	}
}

// placeCell 两轮分配(规范 §4.1,决策 4):先 HDF 次进 HDF 机,再 HD 次。
func (e *Engine) placeCell(cell Cell, sessions []SessionItem) {
	for _, s := range sessions {
		if s.Mode != ModeHDF {
			continue
		}
		m := e.placeHDFSession(cell, s) // 双固定 → 就近 → 报警
		if m == nil {
			e.raiseConflict(ConflictHdfNoMachine, s.PatientId, cell, SeverityAlert, nil)
		} else {
			e.draft(s.PatientId, cell, m, ModeHDF, SourceRegular, &s.Item.Id, false)
		}
	}
	for _, s := range sessions {
		if s.Mode == ModeHDF {
			continue
		}
		m := e.placeHDSession(cell, s) // 固定 → 就近 → 溢出 HDF 机
		if m != nil {
			e.draft(s.PatientId, cell, m, ModeHD, SourceRegular, &s.Item.Id, false)
			continue
		}
		// 本格排满:不加床(决策 22),沿时间维度顺延找空位。
		if spill := e.spillToLaterSlot(s, cell); spill == nil {
			e.raiseConflict(ConflictNoMachine, s.PatientId, cell, SeverityAlert, nil)
		}
	}
}

// placeHDSession HD 次落位:固定 HD 机 → 同区就近 HD 机 → 溢出空闲 HDF 机(规范 §4.2/§4.1)。
func (e *Engine) placeHDSession(cell Cell, s SessionItem) *model.Machine {
	// 1) 固定机位:回到上次同台 HD 机
	if fid := s.Item.FixedHdMachineId; fid != nil {
		if m := e.Board.machine(*fid); m != nil && m.MachineType == MachineHD && e.Board.IsFree(m.Id, cell) {
			return m
		}
	}
	// 2) 同区空闲 HD 机,按"集中连片+组团"挑
	if hd := e.Board.freeMachines(cell, MachineHD); len(hd) > 0 {
		return e.pickBest(hd, s, cell)
	}
	// 3) HD 机全满 → 溢出到空闲 HDF 机(决策 4)
	if hdf := e.Board.freeMachines(cell, MachineHDF); len(hdf) > 0 {
		return e.pickBest(hdf, s, cell)
	}
	return nil
}

// placeHDFSession HDF 次落位:双固定 → 就近 → 报警(规范 §5,决策 2)。
func (e *Engine) placeHDFSession(cell Cell, s SessionItem) *model.Machine {
	// 1) 双固定:回到记忆的固定 HDF 机
	if fid := s.Item.FixedHdfMachineId; fid != nil {
		if m := e.Board.machine(*fid); m != nil && m.MachineType == MachineHDF && e.Board.IsFree(m.Id, cell) {
			return m
		}
	}
	// 2) 就近:同区空闲 HDF 机,取离参考机位最近者
	hdf := e.Board.freeMachines(cell, MachineHDF)
	if len(hdf) == 0 {
		return nil // 3) 报警
	}
	ref := s.Item.FixedHdfMachineId
	if ref == nil {
		ref = s.Item.FixedHdMachineId
	}
	return e.nearest(hdf, ref)
}

// pickBest 挑机偏好(决策 10,规范 §4.4):集中连片 + 组团,不做负载均衡。
func (e *Engine) pickBest(cands []*model.Machine, s SessionItem, cell Cell) *model.Machine {
	var best *model.Machine
	bestScore := int(-1 << 31)
	refPos, hasRef := e.refPosition(s)
	for _, m := range cands {
		score := 0
		// 组团 + 连片:相邻机位已被占 → 加分(填连片块,保护整台全空机)
		score += e.Board.neighborsOccupiedSameCell(m, cell) * 10
		// 固定位记忆:靠近病人历史机位
		if hasRef {
			score -= absInt(m.PositionIndex - refPos)
		}
		if score > bestScore {
			bestScore, best = score, m
		}
	}
	return best
}

// nearest 取离参考机位(PositionIndex)最近的候选;无参考则取首个(已按位序)。
func (e *Engine) nearest(cands []*model.Machine, refMachineID *int64) *model.Machine {
	if refMachineID == nil {
		return cands[0]
	}
	ref := e.Board.machine(*refMachineID)
	if ref == nil {
		return cands[0]
	}
	best := cands[0]
	bestD := absInt(best.PositionIndex - ref.PositionIndex)
	for _, m := range cands[1:] {
		if d := absInt(m.PositionIndex - ref.PositionIndex); d < bestD {
			bestD, best = d, m
		}
	}
	return best
}

func (e *Engine) refPosition(s SessionItem) (int, bool) {
	if fid := s.Item.FixedHdMachineId; fid != nil {
		if m := e.Board.machine(*fid); m != nil {
			return m.PositionIndex, true
		}
	}
	return 0, false
}

// spillToLaterSlot 排满顺延(决策 22,规范 §9.4):先同日后续班次,再后续透析日(优先原班次)。
// 命中则生成"建议草稿 + 提示队列",由护士长人工确认;窗口内仍无 → 返回 nil 交上层报警。
func (e *Engine) spillToLaterSlot(s SessionItem, cell Cell) *Cell {
	for _, cand := range e.laterSlots(cell) {
		if !e.patientAvailable(s.PatientId, cand) {
			continue
		}
		if m := e.placeHDSession(cand, s); m != nil {
			sug := e.draft(s.PatientId, cand, m, ModeHD, SourceRegular, &s.Item.Id, false)
			e.raiseConflict(ConflictSlotSpilled, s.PatientId, cand, SeverityHint, sug)
			c := cand
			return &c
		}
	}
	return nil
}

// laterSlots 生成顺延候选时段:先同日后续班次,再后续透析日(每日优先原班次,再退其它班次)。
func (e *Engine) laterSlots(cell Cell) []Cell {
	var out []Cell
	// 1) 同一天,排在当前班次之后的班次
	started := false
	for _, sh := range e.Board.shifts {
		if sh.Id == cell.ShiftId {
			started = true
			continue
		}
		if started && !sh.IsDisabled {
			out = append(out, Cell{cell.WardId, sh.Id, cell.Date})
		}
	}
	// 2) 后续透析日(窗口内),优先病人原班次,再退其它班次
	added := 0
	for i := 1; i <= e.SpillHorizonDays && added < e.SpillHorizonDays; i++ {
		d := dateOnly(cell.Date).AddDate(0, 0, i)
		if !e.Board.IsDialysisDay(d) {
			continue
		}
		out = append(out, Cell{cell.WardId, cell.ShiftId, d}) // 先试原班次
		for _, sh := range e.Board.shifts {
			if sh.Id == cell.ShiftId || sh.IsDisabled {
				continue
			}
			out = append(out, Cell{cell.WardId, sh.Id, d})
		}
		added++
	}
	return out
}

// patientAvailable 病人在该格所在(日期)是否已有有效排班(避免同日重复)。
// 冲突检测口径见规范 §8:同病人同"日期+班次"不可重复有效排班。
func (e *Engine) patientAvailable(patientID int64, cell Cell) bool {
	for _, d := range e.Board.Drafts {
		if d.PatientId != patientID {
			continue
		}
		if dkey(d.ScheduleDate) == dkey(cell.Date) && d.ShiftId != nil && *d.ShiftId == cell.ShiftId {
			return false
		}
	}
	return true
}

// draft 生成一条草稿排班记录,占用机位并登记到输出。
func (e *Engine) draft(patientID int64, cell Cell, m *model.Machine, mode string, source int16, templateItemID *int64, borrowed bool) *model.PatientShift {
	shiftID := cell.ShiftId
	machineID := m.Id
	s := &model.PatientShift{
		PatientId:            patientID,
		ScheduleDate:         cell.Date,
		ShiftId:              &shiftID,
		WardId:               cell.WardId,
		MachineId:            &machineID,
		Status:               StatusDraft,
		DialysisMode:         mode,
		SourceType:           source,
		RecordForm:           RecordFormRegular,
		IsBorrowedSlot:       borrowed,
		SourceTemplateItemId: templateItemID,
	}
	e.Board.markOccupied(machineID, cell, &occInfo{PatientId: patientID, ShiftId: shiftID, Mode: mode})
	e.Board.Drafts = append(e.Board.Drafts, s)
	return s
}

// raiseConflict 写一条冲突/待处理队列记录(主→备→报警的统一落点)。
func (e *Engine) raiseConflict(ctype string, patientID int64, cell Cell, severity int16, suggested *model.PatientShift) {
	c := &model.ConflictQueue{
		ConflictType: ctype,
		Severity:     severity,
		Status:       0,
	}
	if patientID != 0 {
		pid := patientID
		c.PatientId = &pid
	}
	d := dateOnly(cell.Date)
	c.ScheduleDate = &d
	sid := cell.ShiftId
	c.ShiftId = &sid
	wid := cell.WardId
	c.WardId = &wid
	if suggested != nil {
		c.SuggestedShiftId = &suggested.Id
	}
	e.Board.Conflicts = append(e.Board.Conflicts, c)
}
