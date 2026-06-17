// Package model 定义透析排班子程序的 GORM 数据模型。
//
// 所有表直接复用老库 Schedule_* 系列表,不再新建 Schedule_v2_* 表。
// 新增列通过手动 DDL 脚本添加,老列保留以实现零破坏融合。
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

// Ward 分区(A/B/C + 子区树)。复用老表 Schedule_Ward,新增 ZoneType/ParentWardId/IsSubZone。
type Ward struct {
	BaseModel
	Name             string `gorm:"column:Name;size:256;not null" json:"name"`
	PatientType      string `gorm:"column:PatientType;size:64" json:"patientType"`           // 保留老列
	InfectionType    string `gorm:"column:InfectionType;size:64" json:"infectionType"`       // 保留老列
	ResponsibleUsers string `gorm:"column:ResponsibleUsers;size:512" json:"responsibleUsers"` // 保留老列
	ZoneType         string `gorm:"column:ZoneType;size:8;not null;default:A" json:"zoneType"` // 新增列
	ParentWardId     *int64 `gorm:"column:ParentWardId" json:"parentWardId"`                  // 新增列
	IsSubZone        bool   `gorm:"column:IsSubZone;default:false" json:"isSubZone"`          // 新增列
	Sort             int    `gorm:"column:Sort" json:"sort"`
	IsDisabled       bool   `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
	Note             string `gorm:"column:Note;size:512" json:"note"`
}

func (Ward) TableName() string { return "Schedule_Ward" }

// Machine 机器(实际复用老表 Schedule_Bed,新增 MachineType 等列)。
// 位 = 机器 × 班次 × 日期 动态表达(决策 20)。
type Machine struct {
	BaseModel
	WardId        int64  `gorm:"column:WardId;not null;index;default:0" json:"wardId"`             // 老库可空, GORM读NULL转0
	Code          string `gorm:"column:Code;size:64" json:"code"`               // 新增列
	Name          string `gorm:"column:Name;size:256;not null" json:"name"`
	MachineType   string `gorm:"column:MachineType;size:8;not null;default:HD" json:"machineType"`   // 新增列
	SupportedModes string `gorm:"column:SupportedModes;size:64;not null;default:HD" json:"supportedModes"` // 新增列
	PositionIndex int    `gorm:"column:PositionIndex;not null;default:0" json:"positionIndex"`       // 新增列
	LegacyBedName string `gorm:"column:LegacyBedName;size:256" json:"legacyBedName"`                 // 新增列
	IsDisabled    bool   `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
	Sort          int    `gorm:"column:Sort" json:"sort"`
	Note          string `gorm:"column:Note;size:512" json:"note"`
}

func (Machine) TableName() string { return "Schedule_Bed" }

// MachineOutage 机器停用时段。对应设计 A.1.3(决策 17)。复用老表 Schedule_MachineOutage。
type MachineOutage struct {
	BaseModel
	BedId      int64      `gorm:"column:BedId;not null;default:0" json:"bedId"`        // 保留老列
	MachineId  int64      `gorm:"column:MachineId;not null;index;default:0" json:"machineId"` // 新增列
	ShiftId    *int64     `gorm:"column:ShiftId" json:"shiftId"`                        // 保留老列
	StartAt    time.Time  `gorm:"column:StartAt;not null" json:"startAt"`
	EndAt      *time.Time `gorm:"column:EndAt" json:"endAt"`
	OutageType int16      `gorm:"column:OutageType;not null" json:"outageType"` // 10=临时 20=长期/报废
	Reason     string     `gorm:"column:Reason;size:512" json:"reason"`
}

func (MachineOutage) TableName() string { return "Schedule_MachineOutage" }

// Shift 班次。对应设计 A.1.4。复用老表 Schedule_Shift,新增 ShiftCode。
type Shift struct {
	BaseModel
	Name       string `gorm:"column:Name;size:256" json:"name"`
	ShiftCode  string `gorm:"column:ShiftCode;size:16;not null;default:MORNING" json:"shiftCode"` // 新增列
	StartTime  string `gorm:"column:StartTime;size:32" json:"startTime"`   // 老库 timestamptz
	EndTime    string `gorm:"column:EndTime;size:32" json:"endTime"`       // 老库 timestamptz
	Type       int    `gorm:"column:Type;default:1" json:"type"`           // 保留老列 integer
	Sort       int    `gorm:"column:Sort" json:"sort"`
	IsDisabled bool   `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
	Note       string `gorm:"column:Note;size:512" json:"note"`            // 保留老列
}

func (Shift) TableName() string { return "Schedule_Shift" }

// Calendar 机构日历(透析日/非透析日/假日值班)。对应设计 A.1.5。复用老表 Schedule_Calendar,新增 OpenWardIds/OpenMachineIds。
type Calendar struct {
	BaseModel
	CalDate        time.Time `gorm:"column:CalDate;type:date;not null" json:"calDate"`
	IsDialysisDay  bool      `gorm:"column:IsDialysisDay;not null" json:"isDialysisDay"`
	HolidayMode    int16     `gorm:"column:HolidayMode;default:0" json:"holidayMode"`
	OpenWardIds    string    `gorm:"column:OpenWardIds;type:text" json:"openWardIds"`       // 新增列
	OpenMachineIds string    `gorm:"column:OpenMachineIds;type:text" json:"openMachineIds"` // 新增列
	Note           string    `gorm:"column:Note;size:256" json:"note"`
}

func (Calendar) TableName() string { return "Schedule_Calendar" }

// -------------------------------------------------------------------
// A.2 病人排班属性层
// -------------------------------------------------------------------

// PatientProfile 病人排班骨架。复用老表 Schedule_PatientProfile,新增 WeeklyCount/PatientStatus 等列。
type PatientProfile struct {
	BaseModel
	PatientId           int64  `gorm:"column:PatientId;not null;uniqueIndex" json:"patientId"`
	ZoneTag             string `gorm:"column:ZoneTag;size:8;not null" json:"zoneTag"`
	HomeWardId          *int64 `gorm:"column:HomeWardId" json:"homeWardId"`
	WeeklyCount         int16  `gorm:"column:WeeklyCount" json:"weeklyCount"`   // 新增列
	FreqPattern         int16  `gorm:"column:FreqPattern;not null" json:"freqPattern"`
	ShiftId             *int64 `gorm:"column:ShiftId" json:"shiftId"`
	DefaultMode         string `gorm:"column:DefaultMode;size:8;not null;default:HD" json:"defaultMode"`
	HdfEnabled          bool   `gorm:"column:HdfEnabled;default:false" json:"hdfEnabled"`
	HdfWeekday          *int16 `gorm:"column:HdfWeekday" json:"hdfWeekday"`
	HdfWeekParity       *int16 `gorm:"column:HdfWeekParity" json:"hdfWeekParity"`
	FixedHdBedId        *int64 `gorm:"column:FixedHdBedId" json:"fixedHdBedId"`           // 保留老列
	FixedHdfBedId       *int64 `gorm:"column:FixedHdfBedId" json:"fixedHdfBedId"`         // 保留老列
	FixedHdMachineId    *int64 `gorm:"column:FixedHdMachineId" json:"fixedHdMachineId"`   // 新增列
	FixedHdfMachineId   *int64 `gorm:"column:FixedHdfMachineId" json:"fixedHdfMachineId"` // 新增列
	IsAdmissionRejected bool   `gorm:"column:IsAdmissionRejected;default:false" json:"isAdmissionRejected"`
	EffectiveFrom       *time.Time `gorm:"column:EffectiveFrom;type:date" json:"effectiveFrom"`
	PatientStatus       int16      `gorm:"column:PatientStatus;not null;default:10" json:"patientStatus"`   // 新增列
	DischargeReason     string     `gorm:"column:DischargeReason;size:64" json:"dischargeReason"`           // 新增列
	DischargedAt        *time.Time `gorm:"column:DischargedAt" json:"dischargedAt"`                        // 新增列
	DischargedBy        *int64     `gorm:"column:DischargedBy" json:"dischargedBy"`                        // 新增列
}

func (PatientProfile) TableName() string { return "Schedule_PatientProfile" }

// PlanChange 方案变更生效记录。复用老表 Schedule_PlanChange,字段完全一致。
type PlanChange struct {
	BaseModel
	PatientId     int64      `gorm:"column:PatientId;not null;index" json:"patientId"`
	ChangeType    string     `gorm:"column:ChangeType;size:16;not null" json:"changeType"`
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

// PatientShift 排班记录(核心)。复用老表 Schedule_PatientShift,合并 PatientShiftExt 字段。
// 老库列名 TreatmentTime 映射到 GORM 字段 ScheduleDate。
type PatientShift struct {
	BaseModel
	PatientId    int64     `gorm:"column:PatientId;not null;index" json:"patientId"`
	ScheduleDate time.Time `gorm:"column:TreatmentTime;not null;index" json:"scheduleDate"` // 老库列名 TreatmentTime
	ShiftId      int64     `gorm:"column:ShiftId;not null;index;default:0" json:"shiftId"`   // 老库 NOT NULL,CRRT用0表示无班次
WardId       int64      `gorm:"column:WardId;not null;index" json:"wardId"`                         // 老库 NOT NULL
	BedId        int64      `gorm:"column:BedId;not null;default:0" json:"bedId"`              // 保留老列(与MachineId同值)
	MachineId    int64     `gorm:"column:MachineId;not null;default:0;index" json:"machineId"` // 与BedId同值
	PatientPlanId *int64   `gorm:"column:PatientPlanId;default:0" json:"patientPlanId"`       // 保留老列
	ShiftTiming   *int     `gorm:"column:ShiftTiming;default:0" json:"shiftTiming"`           // 保留老列
Status       int16      `gorm:"column:Status;not null;index" json:"status"`                 // V2状态机 int16
	// 以下为 PatientShiftExt 合并进来的列(新增)
	DialysisMode string `gorm:"column:DialysisMode;size:8;not null;default:HD" json:"dialysisMode"`
	SourceType   int16  `gorm:"column:SourceType;not null;default:10" json:"sourceType"`
	RecordForm   int16  `gorm:"column:RecordForm;not null;default:10" json:"recordForm"`
	Confirm1At   *time.Time `gorm:"column:Confirm1At" json:"confirm1At"`
	Confirm2At   *time.Time `gorm:"column:Confirm2At" json:"confirm2At"`
	Confirm3At   *time.Time `gorm:"column:Confirm3At" json:"confirm3At"`
	Confirm1By   *int64     `gorm:"column:Confirm1By" json:"confirm1By"`
	Confirm2By   *int64     `gorm:"column:Confirm2By" json:"confirm2By"`
	Confirm3By   *int64     `gorm:"column:Confirm3By" json:"confirm3By"`
	IsBorrowedSlot       bool   `gorm:"column:IsBorrowedSlot;default:false" json:"isBorrowedSlot"`
	CancelReason         string `gorm:"column:CancelReason;size:256" json:"cancelReason"`
	MakeupOfShiftId      *int64 `gorm:"column:MakeupOfShiftId" json:"makeupOfShiftId"`             // V2 新增
	SourceTemplateItemId *int64 `gorm:"column:SourceTemplateItemId" json:"sourceTemplateItemId"`
	IsLocked             bool   `gorm:"column:IsLocked;default:false" json:"isLocked"`
}

func (PatientShift) TableName() string { return "Schedule_PatientShift" }

// CrrtSession CRRT 占用。复用老表 Schedule_CrrtSession,新增 MachineId。
type CrrtSession struct {
	BaseModel
	PatientShiftId int64      `gorm:"column:PatientShiftId;not null;uniqueIndex" json:"patientShiftId"`
	BedId          int64      `gorm:"column:BedId;not null;default:0" json:"bedId"`     // 保留老列
	MachineId      int64      `gorm:"column:MachineId;not null;default:0" json:"machineId"` // 新增列
	StartAt        time.Time  `gorm:"column:StartAt;not null" json:"startAt"`
	EndAt          *time.Time `gorm:"column:EndAt" json:"endAt"`
}

func (CrrtSession) TableName() string { return "Schedule_CrrtSession" }

// ScheduleTemplate 模板头。复用老表 Schedule_ScheduleTemplate。
type ScheduleTemplate struct {
	BaseModel
	Name     string `gorm:"column:Name;size:128;not null" json:"name"`
	Scope    string `gorm:"column:Scope;size:8" json:"scope"`
	WardId   *int64 `gorm:"column:WardId" json:"wardId"`        // 保留老列
	Version  int    `gorm:"column:Version;default:1" json:"version"` // 保留老列
	IsActive bool   `gorm:"column:IsActive;default:true" json:"isActive"`
}

func (ScheduleTemplate) TableName() string { return "Schedule_ScheduleTemplate" }

// ScheduleTemplateItem 模板项。复用老表 Schedule_ScheduleTemplateItem,新增 DefaultMode/MachineId。
type ScheduleTemplateItem struct {
	BaseModel
	TemplateId        int64  `gorm:"column:TemplateId;not null;index" json:"templateId"`
	PatientId         int64  `gorm:"column:PatientId;not null;index" json:"patientId"`
	ZoneTag           string `gorm:"column:ZoneTag;size:8;not null" json:"zoneTag"`
	WardId            *int64 `gorm:"column:WardId" json:"wardId"`
	ShiftId           *int64 `gorm:"column:ShiftId" json:"shiftId"`
	FreqPattern       int16  `gorm:"column:FreqPattern;not null" json:"freqPattern"`
	FixedHdBedId      *int64 `gorm:"column:FixedHdBedId" json:"fixedHdBedId"`           // 保留老列
	FixedHdfBedId     *int64 `gorm:"column:FixedHdfBedId" json:"fixedHdfBedId"`         // 保留老列
	DefaultMode       string `gorm:"column:DefaultMode;size:8;not null;default:HD" json:"defaultMode"`       // 新增列
	FixedHdMachineId  *int64 `gorm:"column:FixedHdMachineId" json:"fixedHdMachineId"`   // 新增列
	FixedHdfMachineId *int64 `gorm:"column:FixedHdfMachineId" json:"fixedHdfMachineId"` // 新增列
	HdfEnabled        bool   `gorm:"column:HdfEnabled;default:false" json:"hdfEnabled"`
	HdfWeekday        *int16 `gorm:"column:HdfWeekday" json:"hdfWeekday"`
	HdfWeekParity     *int16 `gorm:"column:HdfWeekParity" json:"hdfWeekParity"`
	TemplateVersion   int    `gorm:"column:TemplateVersion;default:1" json:"templateVersion"` // 保留老列
}

func (ScheduleTemplateItem) TableName() string { return "Schedule_ScheduleTemplateItem" }

// ConflictQueue 冲突/待处理队列。复用老表 Schedule_ConflictQueue。
type ConflictQueue struct {
	BaseModel
	PatientId               *int64     `gorm:"column:PatientId;index" json:"patientId"`
	ScheduleDate            *time.Time `gorm:"column:ScheduleDate;type:date" json:"scheduleDate"`
	ShiftId                 *int64     `gorm:"column:ShiftId" json:"shiftId"`
	WardId                  *int64     `gorm:"column:WardId" json:"wardId"`
	ConflictType            string     `gorm:"column:ConflictType;size:24;not null" json:"conflictType"`
	Severity                int16      `gorm:"column:Severity;default:10" json:"severity"`
	Detail                  string     `gorm:"column:Detail;type:text" json:"detail"`
	SuggestedDate           *time.Time `gorm:"column:SuggestedDate;type:date" json:"suggestedDate"`           // 保留老列
	SuggestedShiftId        *int64     `gorm:"column:SuggestedShiftId" json:"suggestedShiftId"`
	SuggestedBedId          *int64     `gorm:"column:SuggestedBedId" json:"suggestedBedId"`                   // 保留老列
	SuggestedPatientShiftId *int64     `gorm:"column:SuggestedPatientShiftId" json:"suggestedPatientShiftId"` // 保留老列
	Status                  int16      `gorm:"column:Status;not null;default:0" json:"status"`
	ResolvedBy              *int64     `gorm:"column:ResolvedBy" json:"resolvedBy"`
	ResolvedAt              *time.Time `gorm:"column:ResolvedAt" json:"resolvedAt"`
}

func (ConflictQueue) TableName() string { return "Schedule_ConflictQueue" }

// Patient 病人主档(轻量本地档)。新表 Schedule_Patient,需手动创建。
type Patient struct {
	Id               int64      `gorm:"column:Id;primaryKey" json:"id"`
	TenantId         int64      `gorm:"column:TenantId;index;not null" json:"tenantId"`
	Name             string     `gorm:"column:Name;size:64;not null" json:"name"`
	Gender           string     `gorm:"column:Gender;size:8" json:"gender"`
	InfectionStatus  string     `gorm:"column:InfectionStatus;size:16;not null;default:unknown" json:"infectionStatus"`
	InfectionWaivedBy *int64    `gorm:"column:InfectionWaivedBy" json:"infectionWaivedBy"`
	InfectionWaivedAt *time.Time `gorm:"column:InfectionWaivedAt" json:"infectionWaivedAt"`
	CreateTime       time.Time  `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime   time.Time  `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (Patient) TableName() string { return "Schedule_Patient" }

// TenantSetting 租户级配置。复用老表 Schedule_TenantSetting。
type TenantSetting struct {
	BaseModel
	SettingKey   string `gorm:"column:SettingKey;size:64;not null" json:"settingKey"`
	SettingValue string `gorm:"column:SettingValue;size:256;not null" json:"settingValue"`
	SettingType  string `gorm:"column:SettingType;size:16;not null;default:string" json:"settingType"` // 保留老列
}

func (TenantSetting) TableName() string { return "Schedule_TenantSetting" }

// StaffDuty 医护人力排班·月基线（④ v1，契约04/05）
type StaffDuty struct {
	BaseModel
	StaffId   int64     `gorm:"column:StaffId;index;not null" json:"staffId"`
	StaffName string    `gorm:"column:StaffName;size:64" json:"staffName"`
	DutyRole  string    `gorm:"column:DutyRole;size:32;not null" json:"dutyRole"`
	WardId    int64     `gorm:"column:WardId;index;not null" json:"wardId"`
	DutyDate  time.Time `gorm:"column:DutyDate;type:date;index;not null" json:"dutyDate"`
	Shift     string    `gorm:"column:Shift;size:16" json:"shift"`
}

func (StaffDuty) TableName() string { return "Schedule_StaffDuty" }

const (
	DutyRoleDoctor      = "当班医生"
	DutyRoleChargeNurse = "主班护士"
	DutyRoleDutyNurse   = "当班护士"
)

// StaffDutyOverride 当日覆盖（顶班/换班/请假）（④ v2）
type StaffDutyOverride struct {
	BaseModel
	DutyDate        time.Time `gorm:"column:DutyDate;type:date;index;not null" json:"dutyDate"`
	WardId          int64     `gorm:"column:WardId;index;not null" json:"wardId"`
	DutyRole        string    `gorm:"column:DutyRole;size:32;not null" json:"dutyRole"`
	OriginalStaffId int64     `gorm:"column:OriginalStaffId" json:"originalStaffId"`
	ActualStaffId   int64     `gorm:"column:ActualStaffId;not null" json:"actualStaffId"`
	ActualStaffName string    `gorm:"column:ActualStaffName;size:64" json:"actualStaffName"`
	Reason          string    `gorm:"column:Reason;size:128" json:"reason"`
	ChangedBy       int64     `gorm:"column:ChangedBy" json:"changedBy"`
}

func (StaffDutyOverride) TableName() string { return "Schedule_StaffDutyOverride" }

// CheckIn 接班记录（复用老库 Schedule_CheckIn）
type CheckIn struct {
	Id             int64     `gorm:"column:Id;primaryKey;autoIncrement" json:"id"`
	TenantId       int64     `gorm:"column:TenantId" json:"tenantId"`
	ShiftId        int64     `gorm:"column:ShiftId" json:"shiftId"`
	WardId         int64     `gorm:"column:WardId" json:"wardId"`
	ClockInTime    time.Time `gorm:"column:ClockInTime" json:"clockInTime"`
	OperatorType   int64     `gorm:"column:OperatorType" json:"operatorType"`
	Type           int64     `gorm:"column:Type" json:"type"`
	Note           string    `gorm:"column:Note" json:"note"`
	OperatorId     int64     `gorm:"column:OperatorId" json:"operatorId"`
	CreatorId      int64     `gorm:"column:CreatorId" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime" json:"lastModifyTime"`
}

func (CheckIn) TableName() string { return "Schedule_CheckIn" }

// AllModels 返回全部模型,供 AutoMigrate 使用(当前系统禁止 AutoMigrate,此函数仅作文档参考)。
func AllModels() []interface{} {
	return []interface{}{
		&Ward{}, &Machine{}, &MachineOutage{}, &Shift{}, &Calendar{},
		&PatientProfile{}, &PlanChange{},
		&PatientShift{}, &CrrtSession{},
		&ScheduleTemplate{}, &ScheduleTemplateItem{}, &ConflictQueue{},
		&TenantSetting{}, &Patient{}, &StaffDuty{}, &StaffDutyOverride{},
	}
}
