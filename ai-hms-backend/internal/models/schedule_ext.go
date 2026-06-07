package models

import "time"

// 透析排班新规则扩展模型
// 这些模型仅用于 GORM 查询/写入，不加入任何 AutoMigrate 流程。
// 新增表必须通过 docs/sql/schedule_extension_tables.sql 人工审核后执行。

// ---------- 4.1 病区扩展 ----------

type WardExt struct {
	Id             int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	WardId         int64     `gorm:"column:WardId;type:bigint;not null;uniqueIndex:idx_WardExt_Tenant_Ward" json:"wardId"`
	ZoneType       string    `gorm:"column:ZoneType;type:varchar(8);not null;default:A" json:"zoneType"`
	ParentWardId   *int64    `gorm:"column:ParentWardId;type:bigint" json:"parentWardId"`
	IsSubZone      bool      `gorm:"column:IsSubZone;default:false" json:"isSubZone"`
	Note           string    `gorm:"column:Note;type:varchar(512)" json:"note"`
	CreatorId      int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (WardExt) TableName() string { return "Schedule_WardExt" }

// ---------- 4.2 床位/机器扩展 ----------

type BedMachineExt struct {
	Id             int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	BedId          int64     `gorm:"column:BedId;type:bigint;not null;uniqueIndex:idx_BedMachineExt_Tenant_Bed" json:"bedId"`
	MachineCode    string    `gorm:"column:MachineCode;type:varchar(64)" json:"machineCode"`
	MachineType    string    `gorm:"column:MachineType;type:varchar(8);not null;default:HD" json:"machineType"`
	SupportedModes string    `gorm:"column:SupportedModes;type:varchar(64);not null;default:HD" json:"supportedModes"`
	PositionIndex  int       `gorm:"column:PositionIndex;not null;default:0" json:"positionIndex"`
	IsDisabled     bool      `gorm:"column:IsDisabled;default:false" json:"isDisabled"`
	LegacyBedName  string    `gorm:"column:LegacyBedName;type:varchar(256)" json:"legacyBedName"`
	Note           string    `gorm:"column:Note;type:varchar(512)" json:"note"`
	CreatorId      int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (BedMachineExt) TableName() string { return "Schedule_BedMachineExt" }

// ---------- 4.3 患者排班骨架 ----------

type PatientProfile struct {
	Id                  int64      `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId            int64      `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	PatientId           int64      `gorm:"column:PatientId;type:bigint;not null;uniqueIndex:idx_PatientProfile_Tenant_Patient" json:"patientId"`
	ZoneTag             string     `gorm:"column:ZoneTag;type:varchar(8);not null;default:A" json:"zoneTag"`
	HomeWardId          *int64     `gorm:"column:HomeWardId;type:bigint" json:"homeWardId"`
	FreqPattern         int16      `gorm:"column:FreqPattern;type:smallint;not null;default:10" json:"freqPattern"`
	ShiftId             *int64     `gorm:"column:ShiftId;type:bigint" json:"shiftId"`
	DefaultMode         string     `gorm:"column:DefaultMode;type:varchar(8);not null;default:HD" json:"defaultMode"`
	HdfEnabled          bool       `gorm:"column:HdfEnabled;default:false" json:"hdfEnabled"`
	HdfWeekday          *int16     `gorm:"column:HdfWeekday;type:smallint" json:"hdfWeekday"`
	HdfWeekParity       *int16     `gorm:"column:HdfWeekParity;type:smallint" json:"hdfWeekParity"`
	FixedHdBedId        *int64     `gorm:"column:FixedHdBedId;type:bigint" json:"fixedHdBedId"`
	FixedHdfBedId       *int64     `gorm:"column:FixedHdfBedId;type:bigint" json:"fixedHdfBedId"`
	IsAdmissionRejected bool       `gorm:"column:IsAdmissionRejected;default:false" json:"isAdmissionRejected"`
	EffectiveFrom       *time.Time `gorm:"column:EffectiveFrom;type:date" json:"effectiveFrom"`
	CreatorId           int64      `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime          time.Time  `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime      time.Time  `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (PatientProfile) TableName() string { return "Schedule_PatientProfile" }

// ---------- 4.4 排班记录扩展 ----------

type PatientShiftExt struct {
	Id                    int64      `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId              int64      `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	PatientShiftId        int64      `gorm:"column:PatientShiftId;type:bigint;not null;uniqueIndex:idx_PatientShiftExt_Tenant_Shift" json:"patientShiftId"`
	DialysisMode          string     `gorm:"column:DialysisMode;type:varchar(8);not null;default:HD" json:"dialysisMode"`
	SourceType            int16      `gorm:"column:SourceType;type:smallint;not null;default:10" json:"sourceType"`
	RecordForm            int16      `gorm:"column:RecordForm;type:smallint;not null;default:10" json:"recordForm"`
	Confirm1At            *time.Time `gorm:"column:Confirm1At" json:"confirm1At"`
	Confirm2At            *time.Time `gorm:"column:Confirm2At" json:"confirm2At"`
	Confirm3At            *time.Time `gorm:"column:Confirm3At" json:"confirm3At"`
	Confirm1By            *int64     `gorm:"column:Confirm1By;type:bigint" json:"confirm1By"`
	Confirm2By            *int64     `gorm:"column:Confirm2By;type:bigint" json:"confirm2By"`
	Confirm3By            *int64     `gorm:"column:Confirm3By;type:bigint" json:"confirm3By"`
	IsBorrowedSlot        bool       `gorm:"column:IsBorrowedSlot;default:false" json:"isBorrowedSlot"`
	BorrowedFromShiftId   *int64     `gorm:"column:BorrowedFromShiftId;type:bigint" json:"borrowedFromShiftId"`
	IsLocked              bool       `gorm:"column:IsLocked;default:false" json:"isLocked"`
	CancelReason          string     `gorm:"column:CancelReason;type:varchar(256)" json:"cancelReason"`
	SourceTemplateItemId  *int64     `gorm:"column:SourceTemplateItemId;type:bigint" json:"sourceTemplateItemId"`
	SourceTemplateVersion *int       `gorm:"column:SourceTemplateVersion;type:int" json:"sourceTemplateVersion"`
	RuleStatus            int16      `gorm:"column:RuleStatus;type:smallint;not null;default:10" json:"ruleStatus"`
	ApprovedBy            *int64     `gorm:"column:ApprovedBy;type:bigint" json:"approvedBy"`
	CreatorId             int64      `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime            time.Time  `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime        time.Time  `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (PatientShiftExt) TableName() string { return "Schedule_PatientShiftExt" }

// ---------- 4.5 模板头 ----------

type ScheduleTemplate struct {
	Id             int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	Name           string    `gorm:"column:Name;type:varchar(128);not null" json:"name"`
	Scope          string    `gorm:"column:Scope;type:varchar(8)" json:"scope"`
	WardId         *int64    `gorm:"column:WardId;type:bigint" json:"wardId"`
	IsActive       bool      `gorm:"column:IsActive;default:true" json:"isActive"`
	Version        int       `gorm:"column:Version;default:1" json:"version"`
	CreatorId      int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (ScheduleTemplate) TableName() string { return "Schedule_ScheduleTemplate" }

// ---------- 4.6 模板项 ----------

type ScheduleTemplateItem struct {
	Id              int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId        int64     `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	TemplateId      int64     `gorm:"column:TemplateId;type:bigint;not null" json:"templateId"`
	PatientId       int64     `gorm:"column:PatientId;type:bigint;not null" json:"patientId"`
	ZoneTag         string    `gorm:"column:ZoneTag;type:varchar(8);not null;default:A" json:"zoneTag"`
	WardId          *int64    `gorm:"column:WardId;type:bigint" json:"wardId"`
	ShiftId         *int64    `gorm:"column:ShiftId;type:bigint" json:"shiftId"`
	FreqPattern     int16     `gorm:"column:FreqPattern;type:smallint;not null;default:10" json:"freqPattern"`
	FixedHdBedId    *int64    `gorm:"column:FixedHdBedId;type:bigint" json:"fixedHdBedId"`
	FixedHdfBedId   *int64    `gorm:"column:FixedHdfBedId;type:bigint" json:"fixedHdfBedId"`
	HdfEnabled      bool      `gorm:"column:HdfEnabled;default:false" json:"hdfEnabled"`
	HdfWeekday      *int16    `gorm:"column:HdfWeekday;type:smallint" json:"hdfWeekday"`
	HdfWeekParity   *int16    `gorm:"column:HdfWeekParity;type:smallint" json:"hdfWeekParity"`
	TemplateVersion int       `gorm:"column:TemplateVersion;not null;default:1" json:"templateVersion"`
	CreatorId       int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime      time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime  time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (ScheduleTemplateItem) TableName() string { return "Schedule_ScheduleTemplateItem" }

// ---------- 4.7 冲突队列 ----------

type ConflictQueue struct {
	Id                      int64      `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId                int64      `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	PatientId               *int64     `gorm:"column:PatientId;type:bigint" json:"patientId"`
	ScheduleDate            *time.Time `gorm:"column:ScheduleDate;type:date" json:"scheduleDate"`
	ShiftId                 *int64     `gorm:"column:ShiftId;type:bigint" json:"shiftId"`
	WardId                  *int64     `gorm:"column:WardId;type:bigint" json:"wardId"`
	ConflictType            string     `gorm:"column:ConflictType;type:varchar(24);not null" json:"conflictType"`
	Severity                int16      `gorm:"column:Severity;type:smallint;not null;default:10" json:"severity"`
	Detail                  string     `gorm:"column:Detail;type:text" json:"detail"`
	SuggestedDate           *time.Time `gorm:"column:SuggestedDate;type:date" json:"suggestedDate"`
	SuggestedShiftId        *int64     `gorm:"column:SuggestedShiftId;type:bigint" json:"suggestedShiftId"`
	SuggestedBedId          *int64     `gorm:"column:SuggestedBedId;type:bigint" json:"suggestedBedId"`
	SuggestedPatientShiftId *int64     `gorm:"column:SuggestedPatientShiftId;type:bigint" json:"suggestedPatientShiftId"`
	Status                  int16      `gorm:"column:Status;type:smallint;not null;default:0" json:"status"`
	ResolvedBy              *int64     `gorm:"column:ResolvedBy;type:bigint" json:"resolvedBy"`
	ResolvedAt              *time.Time `gorm:"column:ResolvedAt" json:"resolvedAt"`
	CreatorId               int64      `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime              time.Time  `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime          time.Time  `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (ConflictQueue) TableName() string { return "Schedule_ConflictQueue" }

// ---------- 4.8 设备停机 ----------

type MachineOutage struct {
	Id             int64      `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64      `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	BedId          int64      `gorm:"column:BedId;type:bigint;not null" json:"bedId"`
	StartAt        time.Time  `gorm:"column:StartAt;not null" json:"startAt"`
	EndAt          *time.Time `gorm:"column:EndAt" json:"endAt"`
	ShiftId        *int64     `gorm:"column:ShiftId;type:bigint" json:"shiftId"`
	OutageType     int16      `gorm:"column:OutageType;type:smallint;not null;default:10" json:"outageType"`
	Reason         string     `gorm:"column:Reason;type:varchar(512)" json:"reason"`
	CreatorId      int64      `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time  `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time  `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (MachineOutage) TableName() string { return "Schedule_MachineOutage" }

// ---------- 4.9 机构日历 ----------

type Calendar struct {
	Id             int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	CalDate        time.Time `gorm:"column:CalDate;type:date;not null" json:"calDate"`
	IsDialysisDay  bool      `gorm:"column:IsDialysisDay;not null;default:true" json:"isDialysisDay"`
	HolidayMode    int16     `gorm:"column:HolidayMode;type:smallint;not null;default:0" json:"holidayMode"`
	Note           string    `gorm:"column:Note;type:varchar(256)" json:"note"`
	CreatorId      int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (Calendar) TableName() string { return "Schedule_Calendar" }

// ---------- 4.9a 日历开放病区 ----------

type CalendarOpenWard struct {
	Id         int64 `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId   int64 `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	CalendarId int64 `gorm:"column:CalendarId;type:bigint;not null" json:"calendarId"`
	WardId     int64 `gorm:"column:WardId;type:bigint;not null" json:"wardId"`
}

func (CalendarOpenWard) TableName() string { return "Schedule_CalendarOpenWard" }

// ---------- 4.9b 日历开放机器 ----------

type CalendarOpenBed struct {
	Id         int64 `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId   int64 `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	CalendarId int64 `gorm:"column:CalendarId;type:bigint;not null" json:"calendarId"`
	BedId      int64 `gorm:"column:BedId;type:bigint;not null" json:"bedId"`
}

func (CalendarOpenBed) TableName() string { return "Schedule_CalendarOpenBed" }

// ---------- 4.10 CRRT 占用 ----------

type CrrtSession struct {
	Id             int64      `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64      `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	PatientShiftId int64      `gorm:"column:PatientShiftId;type:bigint;not null;uniqueIndex:idx_CrrtSession_Tenant_Shift" json:"patientShiftId"`
	BedId          int64      `gorm:"column:BedId;type:bigint;not null" json:"bedId"`
	StartAt        time.Time  `gorm:"column:StartAt;not null" json:"startAt"`
	EndAt          *time.Time `gorm:"column:EndAt" json:"endAt"`
	CreatorId      int64      `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time  `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time  `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (CrrtSession) TableName() string { return "Schedule_CrrtSession" }

// ---------- 4.11 方案变更 ----------

type PlanChange struct {
	Id             int64      `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64      `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	PatientId      int64      `gorm:"column:PatientId;type:bigint;not null" json:"patientId"`
	ChangeType     string     `gorm:"column:ChangeType;type:varchar(16);not null" json:"changeType"`
	OldValue       string     `gorm:"column:OldValue;type:varchar(64)" json:"oldValue"`
	NewValue       string     `gorm:"column:NewValue;type:varchar(64)" json:"newValue"`
	EffectiveDate  time.Time  `gorm:"column:EffectiveDate;type:date;not null" json:"effectiveDate"`
	AffectedCount  int        `gorm:"column:AffectedCount;not null;default:0" json:"affectedCount"`
	ProcessedAt    *time.Time `gorm:"column:ProcessedAt" json:"processedAt"`
	CreatorId      int64      `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time  `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time  `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (PlanChange) TableName() string { return "Schedule_PlanChange" }

// ---------- 4.12 排班配置 ----------

type TenantSetting struct {
	Id             int64     `gorm:"column:Id;type:bigint;primaryKey" json:"id"`
	TenantId       int64     `gorm:"column:TenantId;type:bigint;not null" json:"tenantId"`
	SettingKey     string    `gorm:"column:SettingKey;type:varchar(64);not null" json:"settingKey"`
	SettingValue   string    `gorm:"column:SettingValue;type:varchar(256);not null" json:"settingValue"`
	SettingType    string    `gorm:"column:SettingType;type:varchar(16);not null;default:string" json:"settingType"`
	CreatorId      int64     `gorm:"column:CreatorId;type:bigint" json:"creatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime" json:"lastModifyTime"`
}

func (TenantSetting) TableName() string { return "Schedule_TenantSetting" }
