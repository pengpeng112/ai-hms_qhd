// Package model 定义透析排班子程序的 GORM 数据模型。
//
// 表均在 PostgreSQL 的 "Schedule_" 命名空间下,沿用老系统(ai-hms)的多租户与审计约定。
// 设计依据:透析排班设计_数据模型与算法_v1.md(A 部分),规范 v1(决策 1-22)。
package model

import "time"

// BaseModel 通用主键 + 多租户 + 审计三件套,所有表内嵌。
type BaseModel struct {
	Id             int64     `gorm:"column:Id;primaryKey;autoIncrement" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;index;not null" json:"tenantId"`
	CreatorId      int64     `gorm:"column:CreatorId" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

// -------------------------------------------------------------------
// A.1 资源层
// -------------------------------------------------------------------

// Ward 分区(A/B/C + 子区树)。对应设计 A.1.1。
// 与老系统差异:用 ZoneType 三值枚举 + ParentWardId 子区树,删除 InfectionType。
type Ward struct {
	BaseModel
	Name         string `gorm:"column:Name;size:256;not null" json:"name"`
	ZoneType     string `gorm:"column:ZoneType;size:8;not null" json:"zoneType"` // A/B/C,见 ZoneType* 常量
	ParentWardId *int64 `gorm:"column:ParentWardId" json:"parentWardId"`         // 子区指向父区;顶级为 NULL
	IsSubZone    bool   `gorm:"column:IsSubZone;default:false" json:"isSubZone"`
	Sort         int    `gorm:"column:Sort" json:"sort"`
	IsDisabled   bool   `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
	Note         string `gorm:"column:Note;size:512" json:"note"`
}

func (Ward) TableName() string { return "Schedule_Ward" }

// Machine 机器(机型能力 + 物理排位 + 停用)。对应设计 A.1.2。
// 位 = 机器 × 班次 × 日期 动态表达(决策 20),不再有独立床表。
type Machine struct {
	BaseModel
	WardId        int64  `gorm:"column:WardId;not null;index" json:"wardId"`
	Code          string `gorm:"column:Code;size:64;not null" json:"code"`               // 院内唯一编号
	Name          string `gorm:"column:Name;size:256" json:"name"`
	MachineType   string `gorm:"column:MachineType;size:8;not null" json:"machineType"` // HD/HDF/CRRT
	PositionIndex int    `gorm:"column:PositionIndex;not null" json:"positionIndex"`    // 区内物理排位,用于连片/相邻判定
	IsDisabled    bool   `gorm:"column:IsDisabled;default:false" json:"isDisabled"`     // 永久报废/长期停用
	Sort          int    `gorm:"column:Sort" json:"sort"`
	LegacyBedId   *int64 `gorm:"column:LegacyBedId" json:"legacyBedId"` // 迁移影子列,对齐老 Schedule_Bed
	Note          string `gorm:"column:Note;size:512" json:"note"`
}

func (Machine) TableName() string { return "Schedule_Machine" }

// MachineOutage 机器停用时段。对应设计 A.1.3(决策 17)。
type MachineOutage struct {
	BaseModel
	MachineId int64      `gorm:"column:MachineId;not null;index" json:"machineId"`
	StartAt   time.Time  `gorm:"column:StartAt;not null" json:"startAt"`
	EndAt     *time.Time `gorm:"column:EndAt" json:"endAt"`                       // NULL=未定/长期
	OutageType int16     `gorm:"column:OutageType;not null" json:"outageType"`   // 10=临时(≤48h可归位) 20=长期/报废
	Reason    string     `gorm:"column:Reason;size:512" json:"reason"`
}

func (MachineOutage) TableName() string { return "Schedule_MachineOutage" }

// Shift 班次。对应设计 A.1.4。一个班次 × 一台机 = 1 个位(决策 20)。
type Shift struct {
	BaseModel
	Name       string `gorm:"column:Name;size:64;not null" json:"name"`
	ShiftCode  string `gorm:"column:ShiftCode;size:16;not null" json:"shiftCode"` // MORNING/AFTERNOON/NIGHT
	StartTime  string `gorm:"column:StartTime;size:8" json:"startTime"`
	EndTime    string `gorm:"column:EndTime;size:8" json:"endTime"`
	Sort       int    `gorm:"column:Sort" json:"sort"` // 上=1 下=2 晚=3
	IsDisabled bool   `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
}

func (Shift) TableName() string { return "Schedule_Shift" }

// Calendar 机构日历(透析日/非透析日/假日值班)。对应设计 A.1.5(决策 19)。
type Calendar struct {
	BaseModel
	CalDate       time.Time `gorm:"column:CalDate;type:date;not null" json:"calDate"`
	IsDialysisDay bool      `gorm:"column:IsDialysisDay;not null" json:"isDialysisDay"`
	HolidayMode   int16     `gorm:"column:HolidayMode;default:0" json:"holidayMode"`     // 0正常 10全院停 20假日值班
	OpenWardIds   string    `gorm:"column:OpenWardIds;type:text" json:"openWardIds"`     // 值班模式开放的区(逗号分隔;空=全开)
	OpenMachineIds string   `gorm:"column:OpenMachineIds;type:text" json:"openMachineIds"`
	Note          string    `gorm:"column:Note;size:256" json:"note"`
}

func (Calendar) TableName() string { return "Schedule_Calendar" }

// -------------------------------------------------------------------
// A.2 病人排班属性层
// -------------------------------------------------------------------

// PatientProfile 病人排班骨架(人工权限属性,算法只读)。对应设计 A.2.1。
type PatientProfile struct {
	BaseModel
	PatientId           int64  `gorm:"column:PatientId;not null;uniqueIndex" json:"patientId"`
	ZoneTag             string `gorm:"column:ZoneTag;size:8;not null" json:"zoneTag"`               // A/B/C 标签驱动
	HomeWardId          *int64 `gorm:"column:HomeWardId" json:"homeWardId"`                         // 归属区(含子区)
	FreqPattern         int16  `gorm:"column:FreqPattern;not null" json:"freqPattern"`              // 10/20/30/40/90
	ShiftId             *int64 `gorm:"column:ShiftId" json:"shiftId"`                               // 全周同一班
	DefaultMode         string `gorm:"column:DefaultMode;size:8;not null;default:HD" json:"defaultMode"`
	HdfEnabled          bool   `gorm:"column:HdfEnabled;default:false" json:"hdfEnabled"`           // 每两周一次 HDF 替换
	HdfWeekday          *int16 `gorm:"column:HdfWeekday" json:"hdfWeekday"`                         // 1=周一..6=周六
	HdfWeekParity       *int16 `gorm:"column:HdfWeekParity" json:"hdfWeekParity"`                   // 0=偶 1=奇(系统算)
	FixedHdMachineId    *int64 `gorm:"column:FixedHdMachineId" json:"fixedHdMachineId"`             // 双固定之一
	FixedHdfMachineId   *int64 `gorm:"column:FixedHdfMachineId" json:"fixedHdfMachineId"`           // 双固定之二(须 HDF 机)
	IsAdmissionRejected bool   `gorm:"column:IsAdmissionRejected;default:false" json:"isAdmissionRejected"`
	EffectiveFrom       *time.Time `gorm:"column:EffectiveFrom;type:date" json:"effectiveFrom"`
}

func (PatientProfile) TableName() string { return "Schedule_PatientProfile" }

// PlanChange 方案变更生效记录。对应设计 A.2.2(决策 14)。
type PlanChange struct {
	BaseModel
	PatientId     int64      `gorm:"column:PatientId;not null;index" json:"patientId"`
	ChangeType    string     `gorm:"column:ChangeType;size:16;not null" json:"changeType"` // FREQ/MODE/SHIFT/ZONE/HDF
	OldValue      string     `gorm:"column:OldValue;size:64" json:"oldValue"`
	NewValue      string     `gorm:"column:NewValue;size:64" json:"newValue"`
	EffectiveDate time.Time  `gorm:"column:EffectiveDate;type:date;not null" json:"effectiveDate"`
	AffectedCount int        `gorm:"column:AffectedCount" json:"affectedCount"`
	ProcessedAt   *time.Time `gorm:"column:ProcessedAt" json:"processedAt"`
}

func (PlanChange) TableName() string { return "Schedule_PlanChange" }

// -------------------------------------------------------------------
// A.3 排班记录层(核心)
// -------------------------------------------------------------------

// PatientShift 排班记录(核心,重构状态机)。对应设计 A.3.1。
type PatientShift struct {
	BaseModel
	PatientId    int64      `gorm:"column:PatientId;not null;index" json:"patientId"`
	ScheduleDate time.Time  `gorm:"column:ScheduleDate;type:date;not null;index" json:"scheduleDate"`
	ShiftId      *int64     `gorm:"column:ShiftId;index" json:"shiftId"` // CRRT 记录可空
	WardId       int64      `gorm:"column:WardId;not null;index" json:"wardId"`
	MachineId    *int64     `gorm:"column:MachineId;index" json:"machineId"` // 待排=NULL
	Status       int16      `gorm:"column:Status;not null;index" json:"status"`
	DialysisMode string     `gorm:"column:DialysisMode;size:8;not null" json:"dialysisMode"` // 按次 HD/HDF/CRRT
	SourceType   int16      `gorm:"column:SourceType;not null" json:"sourceType"`            // 10常规 20临时
	RecordForm   int16      `gorm:"column:RecordForm;not null;default:10" json:"recordForm"` // 10规律 20CRRT
	Confirm1At   *time.Time `gorm:"column:Confirm1At" json:"confirm1At"`
	Confirm2At   *time.Time `gorm:"column:Confirm2At" json:"confirm2At"`
	Confirm3At   *time.Time `gorm:"column:Confirm3At" json:"confirm3At"`
	Confirm1By   *int64     `gorm:"column:Confirm1By" json:"confirm1By"`
	Confirm2By   *int64     `gorm:"column:Confirm2By" json:"confirm2By"`
	Confirm3By   *int64     `gorm:"column:Confirm3By" json:"confirm3By"`
	IsBorrowedSlot       bool   `gorm:"column:IsBorrowedSlot;default:false" json:"isBorrowedSlot"`
	CancelReason         string `gorm:"column:CancelReason;size:256" json:"cancelReason"`
	MakeupOfShiftId      *int64 `gorm:"column:MakeupOfShiftId" json:"makeupOfShiftId"`
	SourceTemplateItemId *int64 `gorm:"column:SourceTemplateItemId" json:"sourceTemplateItemId"`
	IsLocked             bool   `gorm:"column:IsLocked;default:false" json:"isLocked"`
}

func (PatientShift) TableName() string { return "Schedule_PatientShift" }

// CrrtSession CRRT 占用(机 + 起止时间)。对应设计 A.3.2(决策 18)。
type CrrtSession struct {
	BaseModel
	PatientShiftId int64      `gorm:"column:PatientShiftId;not null;uniqueIndex" json:"patientShiftId"`
	MachineId      int64      `gorm:"column:MachineId;not null" json:"machineId"` // 必为 CRRT 机
	StartAt        time.Time  `gorm:"column:StartAt;not null" json:"startAt"`
	EndAt          *time.Time `gorm:"column:EndAt" json:"endAt"`
}

func (CrrtSession) TableName() string { return "Schedule_CrrtSession" }

// ScheduleTemplate 模板头(独立表,根除 Status=60)。对应设计 A.3.3(决策 16)。
type ScheduleTemplate struct {
	BaseModel
	Name     string `gorm:"column:Name;size:128;not null" json:"name"`
	Scope    string `gorm:"column:Scope;size:8" json:"scope"` // ALL/A/B/C
	IsActive bool   `gorm:"column:IsActive;default:true" json:"isActive"`
}

func (ScheduleTemplate) TableName() string { return "Schedule_ScheduleTemplate" }

// ScheduleTemplateItem 模板项(1 病人 1 项的稳定骨架)。对应设计 A.3.3。
type ScheduleTemplateItem struct {
	BaseModel
	TemplateId        int64  `gorm:"column:TemplateId;not null;index" json:"templateId"`
	PatientId         int64  `gorm:"column:PatientId;not null;index" json:"patientId"`
	ZoneTag           string `gorm:"column:ZoneTag;size:8;not null" json:"zoneTag"`
	WardId            *int64 `gorm:"column:WardId" json:"wardId"`
	ShiftId           *int64 `gorm:"column:ShiftId" json:"shiftId"`
	FreqPattern       int16  `gorm:"column:FreqPattern;not null" json:"freqPattern"`
	FixedHdMachineId  *int64 `gorm:"column:FixedHdMachineId" json:"fixedHdMachineId"`
	FixedHdfMachineId *int64 `gorm:"column:FixedHdfMachineId" json:"fixedHdfMachineId"`
	HdfEnabled        bool   `gorm:"column:HdfEnabled;default:false" json:"hdfEnabled"`
	HdfWeekday        *int16 `gorm:"column:HdfWeekday" json:"hdfWeekday"`
	HdfWeekParity     *int16 `gorm:"column:HdfWeekParity" json:"hdfWeekParity"` // 0=偶 1=奇
}

func (ScheduleTemplateItem) TableName() string { return "Schedule_ScheduleTemplateItem" }

// ConflictQueue 冲突/待处理队列(主→备→报警的统一落点)。对应设计 A.3.4。
type ConflictQueue struct {
	BaseModel
	PatientId        *int64     `gorm:"column:PatientId;index" json:"patientId"`
	ScheduleDate     *time.Time `gorm:"column:ScheduleDate;type:date" json:"scheduleDate"`
	ShiftId          *int64     `gorm:"column:ShiftId" json:"shiftId"`
	WardId           *int64     `gorm:"column:WardId" json:"wardId"`
	ConflictType     string     `gorm:"column:ConflictType;size:24;not null" json:"conflictType"`
	Severity         int16      `gorm:"column:Severity;default:10" json:"severity"` // 10提示 20报警
	Detail           string     `gorm:"column:Detail;type:text" json:"detail"`
	SuggestedShiftId *int64     `gorm:"column:SuggestedShiftId" json:"suggestedShiftId"`
	Status           int16      `gorm:"column:Status;not null;default:0" json:"status"` // 0待处理 10已处理 20已忽略
	ResolvedBy       *int64     `gorm:"column:ResolvedBy" json:"resolvedBy"`
	ResolvedAt       *time.Time `gorm:"column:ResolvedAt" json:"resolvedAt"`
}

func (ConflictQueue) TableName() string { return "Schedule_ConflictQueue" }

// Patient 病人主档(轻量本地档;真实系统对接老库 Register_PatientInfomation)。
// Id 即业务用的 PatientId(显式指定,不自增)。
type Patient struct {
	Id             int64     `gorm:"column:Id;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;index;not null" json:"tenantId"`
	Name           string    `gorm:"column:Name;size:64;not null" json:"name"`
	Gender         string    `gorm:"column:Gender;size:8" json:"gender"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (Patient) TableName() string { return "Schedule_Patient" }

// TenantSetting 租户级配置(承载算法依赖的可配置参数)。对应设计 §0.1。
type TenantSetting struct {
	BaseModel
	SettingKey   string `gorm:"column:SettingKey;size:64;not null" json:"settingKey"`
	SettingValue string `gorm:"column:SettingValue;size:256;not null" json:"settingValue"`
}

func (TenantSetting) TableName() string { return "Schedule_TenantSetting" }

// AllModels 返回全部模型,供 AutoMigrate 使用。
func AllModels() []interface{} {
	return []interface{}{
		&Ward{}, &Machine{}, &MachineOutage{}, &Shift{}, &Calendar{},
		&PatientProfile{}, &PlanChange{},
		&PatientShift{}, &CrrtSession{},
		&ScheduleTemplate{}, &ScheduleTemplateItem{}, &ConflictQueue{},
		&TenantSetting{}, &Patient{},
	}
}
