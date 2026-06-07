package sched

import (
	"time"

	"github.com/sdsph/dialysis-scheduling/internal/model"
)

// Cell 表示一个"区 × 班次 × 日期"的排班格子(决策 20:一格一机=一位)。
type Cell struct {
	WardId  int64
	ShiftId int64
	Date    time.Time
}

// occInfo 记录某机位被谁占用(用于挑机的组团/连片打分)。
type occInfo struct {
	PatientId int64
	ShiftId   int64
	Mode      string
}

// Board 是排班算法的内存快照:资源 + 当前占用 + 输出(草稿/冲突)。
// 把"读资源/算占用"集中到 Board,使分配逻辑近似纯函数,便于测试。
type Board struct {
	Anchor         time.Time
	wards          []*model.Ward
	wardByID       map[int64]*model.Ward
	machinesByWard map[int64][]*model.Machine // 已按 PositionIndex 升序
	machineByID    map[int64]*model.Machine
	shifts         []*model.Shift // 已按 Sort 升序
	occupied       map[string]*occInfo
	patientSlot    map[string]bool // 病人在某(日期+班次)是否已有任意记录(含取消/缺席),生成时据此跳过
	outages        []*model.MachineOutage
	calendarByDate map[string]*model.Calendar

	// 输出
	Drafts    []*model.PatientShift
	Conflicts []*model.ConflictQueue
}

// NewBoard 用快照数据构造 Board。machines 会按 ward 分组并按 PositionIndex 排序;
// existing 是库中已有的有效占用(规律排班记录),用于避免重复占位。
func NewBoard(
	anchor time.Time,
	wards []*model.Ward,
	machines []*model.Machine,
	shifts []*model.Shift,
	existing []*model.PatientShift,
	outages []*model.MachineOutage,
	calendar []*model.Calendar,
) *Board {
	b := &Board{
		Anchor:         anchor,
		wards:          wards,
		wardByID:       map[int64]*model.Ward{},
		machinesByWard: map[int64][]*model.Machine{},
		machineByID:    map[int64]*model.Machine{},
		shifts:         sortShifts(shifts),
		occupied:       map[string]*occInfo{},
		patientSlot:    map[string]bool{},
		outages:        outages,
		calendarByDate: map[string]*model.Calendar{},
	}
	for _, w := range wards {
		b.wardByID[w.Id] = w
	}
	for _, m := range machines {
		if m.IsDisabled {
			continue
		}
		b.machineByID[m.Id] = m
		b.machinesByWard[m.WardId] = append(b.machinesByWard[m.WardId], m)
	}
	for _, list := range b.machinesByWard {
		sortMachines(list)
	}
	for _, c := range calendar {
		b.calendarByDate[dkey(c.CalDate)] = c
	}
	for _, s := range existing {
		if s.ShiftId == nil {
			continue
		}
		// 病人占位集合:任意状态都登记(含取消/缺席),供生成时跳过——避免重复已排、复活已取消。
		b.patientSlot[pkey(s.PatientId, s.ScheduleDate, *s.ShiftId)] = true
		if s.MachineId == nil {
			continue
		}
		if s.Status == StatusCancelled || s.Status == StatusAbsent {
			continue // 取消/缺席不占机位(机位可借)
		}
		b.markOccupied(*s.MachineId, Cell{s.WardId, *s.ShiftId, s.ScheduleDate}, &occInfo{
			PatientId: s.PatientId, ShiftId: *s.ShiftId, Mode: s.DialysisMode,
		})
	}
	return b
}

func pkey(patientID int64, date time.Time, shiftID int64) string {
	return itoa(patientID) + "|" + dkey(date) + "|" + itoa(shiftID)
}

// PatientHasSlot 病人在该(日期+班次)是否已有任意记录(含取消/缺席)。
func (b *Board) PatientHasSlot(patientID int64, date time.Time, shiftID int64) bool {
	return b.patientSlot[pkey(patientID, date, shiftID)]
}

func dkey(d time.Time) string { return dateOnly(d).Format("2006-01-02") }

func okey(machineID int64, c Cell) string {
	return dkey(c.Date) + "|" + itoa(c.ShiftId) + "|" + itoa(machineID)
}

func (b *Board) markOccupied(machineID int64, c Cell, info *occInfo) {
	b.occupied[okey(machineID, c)] = info
}

// IsFree 机位在该格是否空闲(未被占 且 不在停机时段)。
func (b *Board) IsFree(machineID int64, c Cell) bool {
	if _, taken := b.occupied[okey(machineID, c)]; taken {
		return false
	}
	return b.NotOutage(machineID, c)
}

// NotOutage 该机器在该格日期是否未停机。停机以"覆盖当天"粗粒度判定(班次级可后续细化)。
func (b *Board) NotOutage(machineID int64, c Cell) bool {
	day := dateOnly(c.Date)
	dayEnd := day.AddDate(0, 0, 1)
	for _, o := range b.outages {
		if o.MachineId != machineID {
			continue
		}
		end := dayEnd
		if o.EndAt != nil {
			end = *o.EndAt
		}
		// 停机区间 [StartAt, end) 与当天 [day, dayEnd) 有交叠则视为冻结
		if o.StartAt.Before(dayEnd) && end.After(day) {
			return false
		}
	}
	return true
}

// machine 查机器。
func (b *Board) machine(id int64) *model.Machine { return b.machineByID[id] }

// freeMachines 返回某格内、指定机型、空闲可用的机器(按 PositionIndex 升序)。
func (b *Board) freeMachines(c Cell, machineType string) []*model.Machine {
	var out []*model.Machine
	for _, m := range b.machinesByWard[c.WardId] {
		if m.MachineType != machineType {
			continue
		}
		if b.IsFree(m.Id, c) {
			out = append(out, m)
		}
	}
	return out
}

// IsDialysisDay 是否透析日:查日历;无记录则周日=false、其余=true(规范 §1.3 兜底)。
func (b *Board) IsDialysisDay(d time.Time) bool {
	if c, ok := b.calendarByDate[dkey(d)]; ok {
		return c.IsDialysisDay
	}
	return d.Weekday() != time.Sunday
}

// ShiftList 返回已按 Sort 升序的班次(供扰动处理逐班次尝试)。
func (b *Board) ShiftList() []*model.Shift { return b.shifts }

// FindFreeForMode 在某(区×班×日)格内,按治疗模式找一台空闲且机型匹配的机器。
// HD 优先 HD 机、满则溢出 HDF 机;HDF/HF 用 HDF 机;CRRT 用 CRRT 机。无则返回 nil。
func (b *Board) FindFreeForMode(wardID, shiftID int64, date time.Time, mode string) *model.Machine {
	c := Cell{WardId: wardID, ShiftId: shiftID, Date: date}
	for _, mt := range ModeMachineTypes(mode) {
		if cands := b.freeMachines(c, mt); len(cands) > 0 {
			return cands[0] // 已按 PositionIndex 升序,取连片靠前者
		}
	}
	return nil
}

// MarkOccupied 标记某机位在该格被某病人占用(扰动处理连续分配时同步占用状态)。
func (b *Board) MarkOccupied(machineID int64, c Cell, patientID int64) {
	b.markOccupied(machineID, c, &occInfo{PatientId: patientID, ShiftId: c.ShiftId})
}

// neighborsOccupiedSameCell 统计某机位在该格内,左右相邻机位是否已被占(组团/连片打分用)。
func (b *Board) neighborsOccupiedSameCell(m *model.Machine, c Cell) int {
	list := b.machinesByWard[c.WardId]
	cnt := 0
	for i, x := range list {
		if x.Id != m.Id {
			continue
		}
		if i > 0 {
			if _, taken := b.occupied[okey(list[i-1].Id, c)]; taken {
				cnt++
			}
		}
		if i < len(list)-1 {
			if _, taken := b.occupied[okey(list[i+1].Id, c)]; taken {
				cnt++
			}
		}
		break
	}
	return cnt
}
