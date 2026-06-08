// Package schedule_engine 透析排班规则引擎
// 规范依据: 透析排班子程序规则规范_v1(1).md
// 核心铁律: 主方案→备选→报警入队, 算法只读骨架、只分机位
package schedule_engine

import "time"

// ── 标准化状态(规范 §7.1) ──
const (
	StatusPending    int = 0  // 待排
	StatusDraft      int = 10 // 草稿
	StatusConfirmed  int = 20 // 已确认
	StatusInDialysis int = 50 // 透析中
	StatusCompleted  int = 60 // 已完成
	StatusCancelled  int = 70 // 已取消
	StatusAbsent     int = 80 // 缺席
)

// ── 频率模式(规范 §2.1) ──
const (
	FreqMonWedFri int16 = 10 // 一三五
	FreqTueThuSat int16 = 20 // 二四六
	FreqTwoPerWk  int16 = 30 // 每周两次
	FreqOnePerWk  int16 = 40 // 每周一次
	FreqTemporary int16 = 90 // 临时(不参与规律生成)
)

// ── 机型(规范 §1.2) ──
const (
	MachineHD   = "HD"
	MachineHDF  = "HDF"
	MachineCRRT = "CRRT"
)

// ── 治疗模式 ──
const (
	ModeHD  = "HD"
	ModeHDF = "HDF"
	ModeHF  = "HF"
)

// ── 来源类型 ──
const (
	SourceRegular   int16 = 10 // 常规模板/算法
	SourceTemporary int16 = 20 // 临时透析
	SourceManual    int16 = 30 // 手工创建
)

// ── 记录形态 ──
const (
	RecordFormRegular int16 = 10 // 规律透析
	RecordFormCRRT    int16 = 20 // CRRT
)

// ── 冲突类型 ──
const (
	ConflictNoMachine    = "NO_MACHINE"
	ConflictHdfNoMachine = "HDF_NO_MACHINE"
	ConflictWardFull     = "WARD_FULL"
	ConflictSlotSpilled  = "SLOT_SPILLED"
)

// ── 严重度 ──
const (
	SeverityHint  int16 = 10
	SeverityAlert int16 = 20
)

// ── 全局配置 ──
const (
	DefaultAnchorMonday = "2025-01-06"
	DefaultSpillHorizon = 14
)

// DefaultAnchorMondayTime 奇偶周基准周一(2025-01-06, 永不漂移)
var DefaultAnchorMondayTime = time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
