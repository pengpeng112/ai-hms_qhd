package schedule_engine

import "time"

// Board 排班引擎工作快照，在事务外构建后传入引擎
type Board struct {
	TenantID  int64
	Anchor    time.Time    // 奇偶周基准周一
	StartDate time.Time    // 生成范围起始
	EndDate   time.Time    // 生成范围结束
	Wards     []WardInfo   // 病区列表
	Beds      []BedInfo    // 床位+机器
	Shifts    []ShiftInfo  // 班次列表
	Profiles  []PatientProfile // 患者骨架
	// 已有占用: key=BedId, value=占用列表
	Occupied map[int64][]Occupancy
	// 非透析日: key=日期字符串(YYYY-MM-DD)
	NonDialysisDays map[string]bool
	// 停机日: key=BedId, value=停机日期集合
	Outages map[int64]map[string]bool
}

// WardInfo 病区快照(算法只读)
type WardInfo struct {
	ID       int64
	Name     string
	ZoneType string // A/B/C
	WardID   int64  // 原病区ID
}

// BedInfo 床位+机器快照(算法只读)
// 当前系统用 Bed+BedMachineExt 表达机器位
type BedInfo struct {
	ID             int64
	Name           string
	WardID         int64
	MachineType    string   // HD/HDF/CRRT
	SupportedModes []string // 支持的透析模式
	PositionIndex  int      // 位置序号
	IsDisabled     bool
}

// ShiftInfo 班次快照
type ShiftInfo struct {
	ID        int64
	Name      string
	StartTime string
	EndTime   string
	Sort      int
	IsDisabled bool
}

// PatientProfile 患者排班骨架(算法只读)
type PatientProfile struct {
	PatientID      int64
	PatientName    string
	ZoneTag        string  // A/B/C
	HomeWardID     *int64
	ShiftID        *int64
	FreqPattern    int16
	DefaultMode    string
	HdfEnabled     bool
	HdfWeekday     *int16 // 0=周日..6=周六
	HdfWeekParity  *int16 // 0=偶, 1=奇
	FixedHdBedID   *int64
	FixedHdfBedID  *int64
}

// Occupancy 占用快照
type Occupancy struct {
	PatientShiftID int64
	PatientID      int64
	Date           time.Time
	ShiftID        int64
	BedID          int64
	WardID         int64
	Status         int
	DialysisMode   string
}

// Cell 排班格子: 病区×日期×班次
type Cell struct {
	WardID  int64
	Date    time.Time
	ShiftID int64
}

// Key 返回格子的唯一标识
func (c Cell) Key() string {
	return c.Date.Format("2006-01-02") + "-" + itoa(c.WardID) + "-" + itoa(c.ShiftID)
}

// SessionItem 一次透析的待排单元
type SessionItem struct {
	PatientID       int64
	Mode            string
	FreqPattern     int16
	FixedHdBedID    *int64
	FixedHdfBedID   *int64
	HdfEnabled      bool
	HdfWeekday      *int16
	HdfWeekParity   *int16
	TemplateItemID  *int64
	WardID          *int64
	ShiftID         *int64
	PatientPlanID   *int64
}

// DraftResult 一条生成结果
type DraftResult struct {
	PatientID       int64
	Date            time.Time
	ShiftID         int64
	WardID          int64
	BedID           int64
	DialysisMode    string
	Status          int
	SourceType      int16
	RecordForm      int16
	TemplateItemID  *int64
	PatientPlanID   *int64
	ShiftTiming     int
	IsBorrowedSlot  bool
	IsSpilled       bool
}

// Conflict 冲突记录
type Conflict struct {
	PatientID            int64
	Date                 time.Time
	ShiftID              int64
	WardID               int64
	BedID                int64
	ConflictType         string
	Severity             int16
	Detail               string
	SuggestedDate        *time.Time
	SuggestedShiftID     *int64
	SuggestedBedID       *int64
	SuggestedPatientShiftID *int64
}

// GenerateResult 生成结果摘要
type GenerateResult struct {
	StartDate      string          `json:"startDate"`
	Weeks          int             `json:"weeks"`
	DialysisDays   int             `json:"dialysisDays"`
	Drafts         []DraftResult   `json:"-"`
	Conflicts      []Conflict      `json:"-"`
	DraftCount     int             `json:"drafts"`
	ConflictCount  int             `json:"conflicts"`
	ParityAssigned int             `json:"parityAssigned"`
}

// 辅助
func itoa(i int64) string {
	if i == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		buf = append([]byte{byte('0' + i%10)}, buf...)
		i /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

// IsOccupied 检查指定床位在指定格是否已被占用
func IsOccupied(occupied map[int64][]Occupancy, bedID int64, cell Cell) bool {
	occs, ok := occupied[bedID]
	if !ok {
		return false
	}
	for _, o := range occs {
		if sameDate(o.Date, cell.Date) && o.ShiftID == cell.ShiftID {
			return true
		}
	}
	return false
}

// PatientHasShift 检查患者在该日期+班次是否已有有效排班
func PatientHasShift(occupied map[int64][]Occupancy, patientID int64, date time.Time, shiftID int64) bool {
	for _, occs := range occupied {
		for _, o := range occs {
			if o.PatientID == patientID && sameDate(o.Date, date) && o.ShiftID == shiftID {
				return true
			}
		}
	}
	return false
}

func sameDate(a, b time.Time) bool {
	ya, ma, da := a.Date()
	yb, mb, db := b.Date()
	return ya == yb && ma == mb && da == db
}

// MachineSupports 判断机型是否支持某治疗模式(规范 §1.2)
func MachineSupports(machineType, mode string) bool {
	switch machineType {
	case MachineHD:
		return mode == ModeHD
	case MachineHDF:
		return mode == ModeHD || mode == ModeHDF || mode == ModeHF
	case MachineCRRT:
		return mode == MachineCRRT
	}
	return false
}

// FindFreeBeds 在指定 Ward 的 Cell 中找空闲且支持指定模式的床位
func FindFreeBeds(beds []BedInfo, occupied map[int64][]Occupancy, wardID int64, cell Cell, mode string) []BedInfo {
	var candidates []BedInfo
	for _, b := range beds {
		if b.WardID != wardID || b.IsDisabled {
			continue
		}
		if !MachineSupports(b.MachineType, mode) {
			continue
		}
		if IsOccupied(occupied, b.ID, cell) {
			continue
		}
		candidates = append(candidates, b)
	}
	return candidates
}
