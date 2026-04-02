package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// TreatmentPlan 透析治疗方案
type TreatmentPlan struct {
	ID                string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID         string         `gorm:"type:varchar(36);not null;index" json:"patientId"`
	WeeklyFrequency   int            `gorm:"default:3" json:"weeklyFrequency"`
	BiweeklyFrequency int            `gorm:"default:0" json:"biweeklyFrequency"`
	Duration          int            `gorm:"default:4" json:"duration"`           // 单位: 小时
	DryWeight         float64        `gorm:"type:decimal(5,2)" json:"dryWeight"`
	ExtraWeight       float64        `gorm:"type:decimal(5,2)" json:"extraWeight"`
	Status            string         `gorm:"type:varchar(20);default:启用" json:"status"` // 启用, 禁用
	DoctorID          *string        `gorm:"type:varchar(36)" json:"doctorId"`
	StartDate         *time.Time     `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	Notes       string    `gorm:"type:text" json:"notes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// 嵌入式 JSON 字段
	DialysisMode      DialysisMode     `gorm:"type:jsonb;serializer:json" json:"dialysisMode"`
	Anticoagulant     Anticoagulant    `gorm:"type:jsonb;serializer:json" json:"anticoagulant"`
	DialysisParameters DialysisParameters `gorm:"type:jsonb;serializer:json" json:"parameters"`
	Materials         MaterialList     `gorm:"type:jsonb;serializer:json" json:"materials"`

	// 关联
	Patient           *Patient        `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
}

// TableName 指定表名
func (TreatmentPlan) TableName() string {
	return "treatment_plans"
}

// DialysisMode 透析模式
type DialysisMode struct {
	Mode         string `json:"mode"`           // HD, HDF, HD+HP
	BloodFlow    int    `json:"bloodFlow"`      // 血流量
	SubstituteInputMode string  `json:"substituteInputMode"` // 置换液输入方式
	SubstituteFlow float64 `json:"substituteFlow"` // 置换液流速 (ml/min)
	SubstituteVolume float64 `json:"substituteVolume"` // 置换液总量 (L)
	BV           string `json:"bv"`             // 抗凝剂标识
	FrequencyDesc string `json:"frequencyDesc"` // 频率描述
	AutoConfirm  bool   `json:"autoConfirm"`    // 自动确认
	Status       string `json:"status"`         // 启用, 禁用
	Notes        string `json:"notes"`
}

// Anticoagulant 抗凝剂
type Anticoagulant struct {
	InitialDrug    string `json:"initialDrug"`    // 首剂量药物
	InitialDose    string `json:"initialDose"`    // 首剂量
	MaintenanceDrug string `json:"maintenanceDrug"` // 维持量药物
	InfusionRate   string `json:"infusionRate"`   // 输注速度
	InfusionTime   string `json:"infusionTime"`   // 输注时间
	MaintenanceDose string `json:"maintenanceDose"` // 维持量
	TotalDose      string `json:"totalDose"`      // 总量
}

// DialysisParameters 透析参数
type DialysisParameters struct {
	DialysateType string  `json:"dialysateType"` // 透析液类型
	DialysateGroup string `json:"dialysateGroup"` // 透析液组号
	FlowRate      int     `json:"flowRate"`      // 透析液流量
	Na            float64 `json:"na"`            // 钠
	Ca            float64 `json:"ca"`            // 钙
	K             float64 `json:"k"`             // 钾
	HCO3          float64 `json:"hco3"`          // 碳酸氢根
	Glucose       string  `json:"glucose"`       // 葡萄糖
	Conductivity  float64 `json:"conductivity"`  // 电导度
	Temp          float64 `json:"temp"`          // 温度
	Volume        float64 `json:"volume"`        // 透析液量
}

// Material 材料
type Material struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Count    int    `json:"count"`
	Code     string `json:"code"`
	Brand    string `json:"brand"`
	Spec     string `json:"spec"`
	Note     string `json:"note"`
}

// MaterialList 材料列表
type MaterialList []Material

// Scan 实现 sql.Scanner 接口
func (m *MaterialList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Value 实现 driver.Valuer 接口
func (m MaterialList) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "[]", nil
	}
	return json.Marshal(m)
}

// PrescriptionOrderItem 处方药品明细（快照）
type PrescriptionOrderItem struct {
	OrderID   string `json:"orderId"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Dose      string `json:"dose"`
	Unit      string `json:"unit"`
	Frequency string `json:"frequency"`
	Route     string `json:"route"`
	Spec      string `json:"spec"`
}

// PrescriptionOrderItemList 处方药品明细列表
type PrescriptionOrderItemList []PrescriptionOrderItem

// Scan 实现 sql.Scanner 接口
func (p *PrescriptionOrderItemList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, p)
}

// Value 实现 driver.Valuer 接口
func (p PrescriptionOrderItemList) Value() (driver.Value, error) {
	if len(p) == 0 {
		return "[]", nil
	}
	return json.Marshal(p)
}

// Prescription 每日处方
type Prescription struct {
	ID              string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID       string         `gorm:"type:varchar(36);not null;index" json:"patientId"`
	TreatmentPlanID string         `gorm:"type:varchar(36);not null" json:"treatmentPlanId"`
	PrescriptionDate time.Time     `gorm:"not null" json:"prescriptionDate"`
	DoctorID        string         `gorm:"type:varchar(36);not null" json:"doctorId"`
	DoctorName      string         `gorm:"type:varchar(50)" json:"doctorName"`
	Status          string         `gorm:"type:varchar(20);default:待执行" json:"status"` // 待执行, 执行中, 已执行, 已取消
	// 从治疗方案复制的字段
	Duration        int            `json:"duration"`
	DryWeight       float64        `gorm:"type:decimal(5,2)" json:"dryWeight"`
	ExtraWeight     float64        `gorm:"type:decimal(5,2)" json:"extraWeight"`
	DialysisMode    DialysisMode   `gorm:"type:jsonb;serializer:json" json:"dialysisMode"`
	Anticoagulant   Anticoagulant  `gorm:"type:jsonb;serializer:json" json:"anticoagulant"`
	Parameters      DialysisParameters `gorm:"type:jsonb;serializer:json" json:"parameters"`
	Materials       MaterialList   `gorm:"type:jsonb;serializer:json" json:"materials"`
	OrderItems      PrescriptionOrderItemList `gorm:"type:jsonb;serializer:json" json:"orderItems"`
	Notes      string    `gorm:"type:text" json:"notes"`
	ExecutedAt *time.Time `json:"executedAt"`
	ExecutedBy *string   `gorm:"type:varchar(36)" json:"executedBy"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (Prescription) TableName() string {
	return "prescriptions"
}

// PrescriptionStatus 常量
const (
	PrescriptionStatusPending   = "待执行"
	PrescriptionStatusExecuting = "执行中"
	PrescriptionStatusExecuted  = "已执行"
	PrescriptionStatusCancelled = "已取消"
)

// Order 医嘱
type Order struct {
	ID          string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID   string         `gorm:"type:varchar(36);not null;index" json:"patientId"`
	Type        string         `gorm:"type:varchar(20);not null" json:"type"`         // 长期, 临时
	Category    string         `gorm:"type:varchar(50)" json:"category"`               // 药品, 检查, 治疗, 护理
	Name        string         `gorm:"type:varchar(100)" json:"name"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Dose        string         `gorm:"type:varchar(50)" json:"dose"`
	Unit        string         `gorm:"type:varchar(20)" json:"unit"`
	Route       string         `gorm:"type:varchar(50)" json:"route"`
	Timing      string         `gorm:"type:varchar(50)" json:"timing"`
	ExecTiming  string         `gorm:"type:varchar(50)" json:"execTiming"`
	DrugID      *uint          `gorm:"index" json:"drugId,omitempty"`
	Spec        string         `gorm:"type:varchar(100)" json:"spec"`
	GroupID     *string        `gorm:"type:varchar(36)" json:"groupId,omitempty"`
	DoctorID    string         `gorm:"type:varchar(36);not null" json:"doctorId"`
	DoctorName  string         `gorm:"type:varchar(50)" json:"doctorName"`
	Status      string         `gorm:"type:varchar(20);default:待执行" json:"status"`  // 待执行, 执行中, 已执行, 已停止
	StartTime   time.Time      `json:"startTime"`
	EndTime     *time.Time     `json:"endTime"`
	Frequency   *string        `gorm:"type:varchar(50)" json:"frequency"`              // qd, bid, tid, qod, etc.
	Priority    string         `gorm:"type:varchar(20);default:普通" json:"priority"`  // 普通, 紧急, 临急
	Notes       string         `gorm:"type:text" json:"notes"`
	ExecutedAt  *time.Time     `json:"executedAt"`
	ExecutedBy  *string   `gorm:"type:varchar(36)" json:"executedBy"`
	StopReason *string   `gorm:"type:text" json:"stopReason"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// OrderType 常量
const (
	OrderTypeLongTerm  = "长期"
	OrderTypeTemporary = "临时"
)

// OrderCategory 常量
const (
	OrderCategoryMedicine = "药品"
	OrderCategoryExam     = "检查"
	OrderCategoryTreatment = "治疗"
	OrderCategoryNursing  = "护理"
	OrderCategoryDiet     = "饮食"
)

// OrderStatus 常量
const (
	OrderStatusPending   = "待执行"
	OrderStatusExecuting = "执行中"
	OrderStatusExecuted  = "已执行"
	OrderStatusStopped   = "已停止"
)

// OrderPriority 常量
const (
	OrderPriorityNormal  = "普通"
	OrderPriorityUrgent  = "紧急"
	OrderPriorityCritical = "临急"
)

// AdjustmentRecord 方案调整记录
type AdjustmentRecord struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	PatientID string    `gorm:"type:varchar(36);not null;index" json:"patientId"`
	Content   string    `gorm:"type:text;not null" json:"content"`       // 调整内容描述
	Operator  string    `gorm:"type:varchar(50)" json:"operator"`        // 调整人
	CreatedAt time.Time `json:"createdAt"`
}

// TableName 指定表名
func (AdjustmentRecord) TableName() string {
	return "adjustment_records"
}

// DialysisMode 透析模式常量
const (
	DialysisModeHD  = "HD"  // 血液透析
	DialysisModeHDF = "HDF" // 血液透析滤过
	DialysisModeHP  = "HP"  // 血液灌流
	DialysisModeHDHP = "HD+HP" // 血液透析+灌流
)

// TreatmentPlanStatus 常量
const (
	TreatmentPlanStatusActive   = "启用"
	TreatmentPlanStatusInactive = "禁用"
)

// ===== 透析治疗执行模块 =====

// Treatment 透析治疗主表
type Treatment struct {
	Id              int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId        int64       `gorm:"index:idx_treatment_tenant;not null" json:"tenantId"`
	PatientId       int64       `gorm:"index:idx_treatment_patient_date;not null" json:"patientId"`
	TreatmentDate   time.Time   `gorm:"index:idx_treatment_patient_date;not null" json:"treatmentDate"`
	ScheduleId      *int64      `gorm:"index" json:"scheduleId"` // 关联 Schedule_PatientShift
	ReceptionDrId   *int64      `json:"receptionDrId"` // 接诊医生
	SignInTime      *time.Time  `json:"signInTime"` // 签到时间
	QueueNo         string      `gorm:"size:32" json:"queueNo"` // 排队号
	ReceptionTime   *time.Time  `json:"receptionTime"` // 接诊时间
	DayProgrammeId  *int64      `json:"dayProgrammeId"` // 日间治疗方案ID
	WardId          *int64      `json:"wardId"`
	WardName        string      `gorm:"size:256" json:"wardName"`
	BedId           *int64      `gorm:"index" json:"bedId"`
	ShiftId         *int64      `gorm:"index" json:"shiftId"`
	ShiftTiming     int         `json:"shiftTiming"` // 班次时段
	Type            int         `gorm:"not null" json:"type"` // 治疗类型
	Status          int         `gorm:"not null;default:0" json:"status"` // 状态
	IsDisabled      bool      `gorm:"default:false" json:"isDisabled"`
	CreatorId       int64     `gorm:"not null" json:"creatorId"`
	CreateTime      time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime  time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Patient         *Patient         `gorm:"foreignKey:PatientId" json:"patient,omitempty"`
	Schedule        *PatientShift    `gorm:"foreignKey:ScheduleId" json:"schedule,omitempty"`
	Ward            *Ward            `gorm:"foreignKey:WardId" json:"ward,omitempty"`
	Bed             *Bed             `gorm:"foreignKey:BedId" json:"bed,omitempty"`
	Shift           *Shift           `gorm:"foreignKey:ShiftId" json:"shift,omitempty"`
	BeforeCheck     *TreatmentBeforeCheck `gorm:"foreignKey:TreatmentId" json:"beforeCheck,omitempty"`
	BeforeSigns     *TreatmentBeforeSigns `gorm:"foreignKey:TreatmentId" json:"beforeSigns,omitempty"`
	AfterSigns      *TreatmentAfterSigns  `gorm:"foreignKey:TreatmentId" json:"afterSigns,omitempty"`
	DuringParams    []TreatmentDuringParam `gorm:"foreignKey:TreatmentId" json:"duringParams,omitempty"`
	Alarms          []TreatmentAlarm       `gorm:"foreignKey:TreatmentId" json:"alarms,omitempty"`
}

// TableName 指定表名
func (Treatment) TableName() string {
	return "Treatment_Treatment"
}

// TreatmentType 治疗类型常量
const (
	TreatmentTypeHD    = 1 // HD (血液透析)
	TreatmentTypeHDF   = 2 // HDF (血液透析滤过)
	TreatmentTypeHP    = 3 // HP (血液灌流)
	TreatmentTypeHDHP  = 4 // HD+HP (血液透析+灌流)
)

// TreatmentStatus 治疗状态常量
const (
	TreatmentStatusPending    = 0 // 待开始
	TreatmentStatusInProgress = 1 // 进行中
	TreatmentStatusCompleted  = 2 // 已完成
	TreatmentStatusCancelled  = 3 // 已取消
)

// TreatmentBeforeCheck 透前检查
type TreatmentBeforeCheck struct {
	Id              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId        int64      `gorm:"index;not null" json:"tenantId"`
	TreatmentId     int64      `gorm:"index:idx_treatment;unique;not null" json:"treatmentId"`
	Weight          *float64   `json:"weight"` // 体重
	Temperature     *float64   `json:"temperature"` // 体温
	SBP             *int       `json:"sbp"` // 收缩压
	DBP             *int       `json:"dbp"` // 舒张压
	HeartRate       *int       `json:"heartRate"` // 心率
	Edema           *int       `json:"edema"` // 水肿
	Consciousness   *string    `gorm:"size:32" json:"consciousness"` // 意识状态
	Complication    *string    `gorm:"size:512" json:"complication"` // 并发症
	DryWeight       *float64   `json:"dryWeight"` // 干体重
	PreWeight       *float64   `json:"preWeight"` // 预估体重
	VascularAccess  *string    `gorm:"size:256" json:"vascularAccess"` // 血管通路
	CannulaType     *string    `gorm:"size:64" json:"cannulaType"` // 穿刺类型
	CannulaPosition *string    `gorm:"size:256" json:"cannulaPosition"` // 穿刺部位
	Catheter        *string    `gorm:"size:512" json:"catheter"` // 导管情况
	HeparingLock    *string    `gorm:"size:512" json:"heparingLock"` // 肝素封管
	MachineNo       *string    `gorm:"size:64" json:"machineNo"` // 机器号
	Dialyzer        *string    `gorm:"size:256" json:"dialyzer"` // 透析器
	Dialysate       *string    `gorm:"size:256" json:"dialysate"` // 透析液
	Calcium         *float64   `json:"calcium"` // 钙浓度
	Sodium          *float64   `json:"sodium"` // 钠浓度
	Bicarbonate     *float64   `json:"bicarbonate"` // 碳酸氢根
	Notes           *string    `gorm:"size:1024" json:"notes"` // 备注
	CreatorId      int64     `gorm:"not null" json:"creatorId"`
	CreateTime     time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Treatment       *Treatment `gorm:"foreignKey:TreatmentId" json:"treatment,omitempty"`
}

// TableName 指定表名
func (TreatmentBeforeCheck) TableName() string {
	return "Treatment_BeforeCheck"
}

// TreatmentBeforeSigns 透前体征
type TreatmentBeforeSigns struct {
	Id              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId        int64      `gorm:"index;not null" json:"tenantId"`
	TreatmentId     int64      `gorm:"index:idx_treatment;unique;not null" json:"treatmentId"`
	SBP             *int       `json:"sbp"` // 收缩压
	DBP             *int       `json:"dbp"` // 舒张压
	HeartRate       *int       `json:"heartRate"` // 心率
	SpO2            *int       `json:"spO2"` // 血氧饱和度
	Respiration     *int       `json:"respiration"` // 呼吸
	Temperature     *float64   `json:"temperature"` // 体温
	Weight          *float64   `json:"weight"` // 体重
	Symptoms        *string    `gorm:"size:1024" json:"symptoms"` // 症状描述
	Notes           *string    `gorm:"size:1024" json:"notes"` // 备注
	CreatorId      int64     `gorm:"not null" json:"creatorId"`
	CreateTime     time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Treatment       *Treatment `gorm:"foreignKey:TreatmentId" json:"treatment,omitempty"`
}

// TableName 指定表名
func (TreatmentBeforeSigns) TableName() string {
	return "Treatment_BeforeSigns"
}

// TreatmentDuringParam 透析中参数
type TreatmentDuringParam struct {
	Id              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId        int64      `gorm:"index:idx_during_param_tenant;not null" json:"tenantId"`
	TreatmentId     int64      `gorm:"index:idx_during_param_treatment_time;not null" json:"treatmentId"`
	RecordTime      time.Time  `gorm:"index:idx_during_param_treatment_time;not null" json:"recordTime"`
	Code            string     `gorm:"size:32;not null" json:"code"` // 参数代码
	BloodFlow       *float64   `json:"bloodFlow"` // 血流量
	DialysateFlow   *float64   `json:"dialysateFlow"` // 透析液流量
	UFVolume        *float64   `json:"ufVolume"` // 超滤量
	VenousPressure  *float64   `json:"venousPressure"` // 静脉压
	ArterialPressure *float64  `json:"arterialPressure"` // 动脉压
	TMP             *float64   `json:"tmp"` // 跨膜压
	Temperature     *float64   `json:"temperature"` // 温度
	Conductivity    *float64   `json:"conductivity"` // 电导度
	UFRate          *float64   `json:"ufRate"` // 超滤率
	Notes           *string    `gorm:"size:512" json:"notes"` // 备注
	CreatorId      int64     `gorm:"not null" json:"creatorId"`
	CreateTime     time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Treatment       *Treatment `gorm:"foreignKey:TreatmentId" json:"treatment,omitempty"`
}

// TableName 指定表名
func (TreatmentDuringParam) TableName() string {
	return "Treatment_DuringParam"
}

// TreatmentAfterSigns 透后体征
type TreatmentAfterSigns struct {
	Id              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId        int64      `gorm:"index;not null" json:"tenantId"`
	TreatmentId     int64      `gorm:"index:idx_treatment;unique;not null" json:"treatmentId"`
	SBP             *int       `json:"sbp"` // 收缩压
	DBP             *int       `json:"dbp"` // 舒张压
	HeartRate       *int       `json:"heartRate"` // 心率
	SpO2            *int       `json:"spO2"` // 血氧饱和度
	Weight          *float64   `json:"weight"` // 体重
	UFVolume        *float64   `json:"ufVolume"` // 实际超滤量
	DialysisTime    *int       `json:"dialysisTime"` // 透析时长(分钟)
	Complication    *string    `gorm:"size:1024" json:"complication"` // 并发症
	Symptoms        *string    `gorm:"size:1024" json:"symptoms"` // 症状描述
	Notes           *string    `gorm:"size:1024" json:"notes"` // 备注
	CreatorId      int64     `gorm:"not null" json:"creatorId"`
	CreateTime     time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Treatment       *Treatment `gorm:"foreignKey:TreatmentId" json:"treatment,omitempty"`
}

// TableName 指定表名
func (TreatmentAfterSigns) TableName() string {
	return "Treatment_AfterSigns"
}

// TreatmentAlarm 报警记录
type TreatmentAlarm struct {
	Id              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantId        int64      `gorm:"index:idx_alarm_tenant;not null" json:"tenantId"`
	TreatmentId     int64      `gorm:"index:idx_alarm_treatment;not null" json:"treatmentId"`
	AlarmTime       time.Time  `gorm:"index:idx_alarm_treatment;not null" json:"alarmTime"`
	AlarmCode       string     `gorm:"size:64;not null" json:"alarmCode"` // 报警代码
	AlarmLevel      int        `gorm:"not null" json:"alarmLevel"` // 报警级别
	AlarmMessage    string     `gorm:"size:512;not null" json:"alarmMessage"` // 报警信息
	IsHandled       bool       `gorm:"default:false" json:"isHandled"` // 是否已处理
	HandledBy       *int64     `json:"handledBy"` // 处理人ID
	HandledAt       *time.Time `json:"handledAt"` // 处理时间
	HandleNote      *string    `gorm:"size:512" json:"handleNote"` // 处理说明
	CreatorId      int64     `gorm:"not null" json:"creatorId"`
	CreateTime     time.Time `gorm:"autoCreateTime" json:"createTime"`
	LastModifyTime time.Time `gorm:"autoUpdateTime" json:"lastModifyTime"`

	// 关联
	Treatment       *Treatment `gorm:"foreignKey:TreatmentId" json:"treatment,omitempty"`
}

// TableName 指定表名
func (TreatmentAlarm) TableName() string {
	return "Treatment_Alarm"
}

// AlarmLevel 报警级别常量
const (
	AlarmLevelInfo     = 1 // 信息
	AlarmLevelWarning  = 2 // 警告
	AlarmLevelError    = 3 // 错误
	AlarmLevelCritical = 4 // 严重
)
