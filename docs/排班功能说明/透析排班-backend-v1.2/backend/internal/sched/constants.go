// Package sched 实现透析排班子程序的核心算法:周序号/奇偶周、频率推导、
// 两轮分配、HDF 双固定、排满顺延、新病人初始分配、冲突入队。
//
// 设计依据:透析排班设计_数据模型与算法_v1.md(B 部分),规范 v1(决策 1-22)。
// 贯穿铁律:主方案 → 备选 → 报警入队(绝不硬排);算法只读骨架、只分机位。
package sched

import "time"

// ── 排班记录状态机(规范 §7.1,决策 16)──
const (
	StatusPending    int16 = 0  // 待排:应排未排的缺口
	StatusDraft      int16 = 10 // 草稿:模板复制/算法生成,未确认
	StatusConfirmed  int16 = 20 // 已确认:计划生效(确认进度由 Confirm1/2/3At 表达)
	StatusInDialysis int16 = 50 // 透析中:已上机
	StatusCompleted  int16 = 60 // 已完成
	StatusCancelled  int16 = 70 // 已取消:提前请假/计划取消(留痕)
	StatusAbsent     int16 = 80 // 缺席:当日爽约(留痕,机位可借)
)

// ── 频率模式(规范 §2.1 五种,决策 0)──
const (
	FreqMonWedFri int16 = 10 // 一三五:每周 3 次
	FreqTueThuSat int16 = 20 // 二四六:每周 3 次
	FreqTwoPerWk  int16 = 30 // 每周两次:周二、周四
	FreqOnePerWk  int16 = 40 // 每周一次:周四
	FreqTemporary int16 = 90 // 临时:无固定透析日
)

// ── 机型(规范 §1.2)──
const (
	MachineHD   = "HD"
	MachineHDF  = "HDF"
	MachineCRRT = "CRRT"
)

// ── 治疗模式(按次,决策 1)──
const (
	ModeHD   = "HD"
	ModeHDF  = "HDF"
	ModeHF   = "HF"
	ModeCRRT = "CRRT"
)

// ── 分区类型(规范 §1.1,决策 5)──
const (
	ZoneA = "A" // 门诊
	ZoneB = "B" // 住院
	ZoneC = "C" // 全警戒
)

// ── 排班来源(正交,规范 §7.2)──
const (
	SourceRegular   int16 = 10 // 常规(模板复制)
	SourceTemporary int16 = 20 // 临时(急诊插入)
)

// ── 数据形态 ──
const (
	RecordFormRegular int16 = 10 // 规律透析(区+机+班+日)
	RecordFormCRRT    int16 = 20 // CRRT(机+起止时间)
)

// ── 停机类型(决策 17)──
const (
	OutageTemp int16 = 10 // 临时(默认 ≤48h,可归位)
	OutageLong int16 = 20 // 长期/报废(>48h,人工永久迁移)
)

// ── 冲突队列严重度与类型(设计 A.3.4)──
const (
	SeverityHint  int16 = 10 // 提示级
	SeverityAlert int16 = 20 // 报警级
)

const (
	ConflictNoMachine     = "NO_MACHINE"            // 完全无可用时段/机器
	ConflictHdfNoMachine  = "HDF_NO_MACHINE"        // 无空闲 HDF 机
	ConflictWardFull      = "WARD_FULL"             // 区满
	ConflictNewUnplaced   = "NEW_PATIENT_UNPLACED"  // 新病人排不进
	ConflictMachineOutage = "MACHINE_OUTAGE"        // 停机受影响
	ConflictHolidayReplan = "HOLIDAY_REPLAN"        // 假日需挪班
	ConflictPlanChange    = "PLAN_CHANGE"           // 方案变更需审核
	ConflictMakeupSuggest = "MAKEUP_SUGGEST"        // 补透建议
	ConflictSlotSpilled   = "SLOT_SPILLED"          // 排满顺延到后续时段(提示级,决策 22)
)

// ── 配置键(设计 §0.1)──
const (
	CfgAnchorMonday      = "OddEvenWeekAnchorMonday" // 奇偶周基准周一
	CfgLowSlotWarn       = "LowSlotWarnThreshold"    // 余位预警阈值
	CfgDraftWeeks        = "DraftWeeks"              // 一次生成草稿周数
	CfgSpillHorizonDays  = "SpillHorizonDays"        // 排满顺延窗口天数
)

// DefaultAnchorMonday 奇偶周默认基准周一(决策 21 / D-6):2025-01-06 是周一。
// 不变量:必须是周一;上线后不应再改(改它会翻转全体 HDF 病人奇偶周)。
var DefaultAnchorMonday = time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)

// machineCapability 机型能力包含关系(规范 §1.2),代码常量,不入库。
var machineCapability = map[string]map[string]bool{
	MachineHD:   {ModeHD: true},
	MachineHDF:  {ModeHD: true, ModeHDF: true, ModeHF: true},
	MachineCRRT: {ModeCRRT: true},
}

// MachineSupports 判断某机型能否执行某治疗模式。
func MachineSupports(machineType, mode string) bool {
	caps, ok := machineCapability[machineType]
	if !ok {
		return false
	}
	return caps[mode]
}

// freqWeekdays 频率模式 → 透析日集合(规范 §2.1),代码常量,不入库。
var freqWeekdays = map[int16][]time.Weekday{
	FreqMonWedFri: {time.Monday, time.Wednesday, time.Friday},
	FreqTueThuSat: {time.Tuesday, time.Thursday, time.Saturday},
	FreqTwoPerWk:  {time.Tuesday, time.Thursday},
	FreqOnePerWk:  {time.Thursday},
	FreqTemporary: {},
}

// FreqWeekdays 返回某频率模式的透析日集合(只读副本语义)。
func FreqWeekdays(freq int16) []time.Weekday {
	return freqWeekdays[freq]
}

// ModeMachineTypes 返回某治疗模式可用的机型(按优先级:HD 优先 HD 机、满则溢出 HDF 机)。
func ModeMachineTypes(mode string) []string {
	switch mode {
	case ModeHD:
		return []string{MachineHD, MachineHDF}
	case ModeHDF, ModeHF:
		return []string{MachineHDF}
	case ModeCRRT:
		return []string{MachineCRRT}
	}
	return nil
}
