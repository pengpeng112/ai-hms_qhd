// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ===== 方案模板 =====

// PlanTemplate 治疗方案模板
type PlanTemplate struct {
	ID              string              `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name            string              `gorm:"type:varchar(100);not null" json:"name"`
	Description     string              `gorm:"type:text" json:"description"`
	Mode            string              `gorm:"type:varchar(20);not null" json:"mode"` // HD, HDF, HP, HF, HFD
	IsDefault       bool                `gorm:"default:false" json:"isDefault"`
	IsEnabled       bool                `json:"isEnabled"`
	Category        string              `gorm:"type:varchar(50)" json:"category"` // 分类: 急性, 慢性, 导管, 等
	TenantID        *int64              `gorm:"index" json:"tenantId,omitempty"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
	TemplateContent PlanTemplateContent `gorm:"type:jsonb;serializer:json" json:"templateContent"`
}

// PlanTemplateContent 模板内容
type PlanTemplateContent struct {
	// 基础配置
	WeeklyFrequency   int     `json:"weeklyFrequency"`
	BiweeklyFrequency int     `json:"biweeklyFrequency"`
	Duration          int     `json:"duration"`
	DryWeight         float64 `json:"dryWeight"`

	// 透析模式
	DialysisMode struct {
		Mode                string  `json:"mode"`
		BloodFlow           int     `json:"bloodFlow"`
		SubstituteInputMode string  `json:"substituteInputMode"`
		SubstituteFlow      float64 `json:"substituteFlow"`
		SubstituteVolume    float64 `json:"substituteVolume"`
		BV                  string  `json:"bv"`
		FrequencyDesc       string  `json:"frequencyDesc"`
		AutoConfirm         bool    `json:"autoConfirm"`
		Status              string  `json:"status"`
		Notes               string  `json:"notes"`
	} `json:"dialysisMode"`

	// 抗凝剂
	Anticoagulant struct {
		InitialDrug     string `json:"initialDrug"`
		InitialDose     string `json:"initialDose"`
		TotalDose       string `json:"totalDose"`
		MaintenanceDrug string `json:"maintenanceDrug"`
		InfusionRate    string `json:"infusionRate"`
		InfusionTime    string `json:"infusionTime"`
		MaintenanceDose string `json:"maintenanceDose"`
	} `json:"anticoagulant"`

	// 透析参数
	Parameters struct {
		DialysateType  string  `json:"dialysateType"`
		DialysateGroup string  `json:"dialysateGroup"`
		FlowRate       int     `json:"flowRate"`
		Na             float64 `json:"na"`
		Ca             float64 `json:"ca"`
		K              float64 `json:"k"`
		HCO3           float64 `json:"hco3"`
		Glucose        string  `json:"glucose"`
		Conductivity   float64 `json:"conductivity"`
		Temp           float64 `json:"temp"`
		Volume         float64 `json:"volume"`
	} `json:"parameters"`

	// 材料
	Materials []PlanTemplateMaterial `json:"materials"`
}

// PlanTemplateMaterial 模板材料项
type PlanTemplateMaterial struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Count    int    `json:"count"`
	Code     string `json:"code"`
	Brand    string `json:"brand"`
	Spec     string `json:"spec"`
	Note     string `json:"note"`
}

// TableName 指定表名
func (PlanTemplate) TableName() string {
	return "plan_templates"
}

// PlanTemplateMode 模式常量
const (
	PlanTemplateModeHD  = "HD"
	PlanTemplateModeHFD = "HFD"
	PlanTemplateModeHP  = "HP"
	PlanTemplateModeHF  = "HF"
	PlanTemplateModeHDF = "HDF"
)

// ===== 材料目录 =====

// MaterialCatalog 材料目录（主数据）
type MaterialCatalog struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code         string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	ShortName    *string   `gorm:"type:varchar(50)" json:"shortName,omitempty"` // 简称
	Mnemonic     *string   `gorm:"type:varchar(50)" json:"mnemonic,omitempty"`  // 助记符
	Category     string    `gorm:"type:varchar(50);not null;index" json:"category"`
	Spec         string    `gorm:"type:varchar(100)" json:"spec"`
	StandardType *string   `gorm:"type:varchar(50)" json:"standardType,omitempty"` // 标准类型
	Brand        string    `gorm:"type:varchar(100)" json:"brand"`
	Unit         string    `gorm:"type:varchar(20)" json:"unit"`
	Packaging    *string   `gorm:"type:varchar(50)" json:"packaging,omitempty"`     // 包装
	Manufacturer *string   `gorm:"type:varchar(100)" json:"manufacturer,omitempty"` // 生产厂家
	SortOrder    int       `gorm:"default:0" json:"sortOrder"`                      // 排序
	IsEnabled    bool      `json:"isEnabled"`
	TenantID     *int64    `gorm:"index" json:"tenantId,omitempty"`
	Notes        string    `gorm:"type:text" json:"notes"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (MaterialCatalog) TableName() string {
	return "material_catalogs"
}

// MaterialCategory 材料分类常量
const (
	MaterialCategoryDialyzer     = "透析器"
	MaterialCategoryBloodLine    = "血路管"
	MaterialCategoryCatheter     = "导管"
	MaterialCategoryNeedle       = "穿刺针"
	MaterialCategorySyringe      = "注射器"
	MaterialCategoryInfusionSet  = "输液器"
	MaterialCategoryDisinfectant = "消毒剂"
	MaterialCategoryDressing     = "敷料"
	MaterialCategoryGlove        = "手套"
	MaterialCategoryOther        = "其他"
)

// ===== 药品目录 =====

// DrugCatalog 药品目录（主数据）
type DrugCatalog struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code          string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name          string    `gorm:"type:varchar(100);not null" json:"name"`
	ShortName     *string   `gorm:"type:varchar(50)" json:"shortName,omitempty"`
	Mnemonic      *string   `gorm:"type:varchar(50)" json:"mnemonic,omitempty"`
	GenericName   string    `gorm:"type:varchar(100)" json:"genericName"`
	Category      string    `gorm:"type:varchar(50);not null;index" json:"category"`
	Spec          string    `gorm:"type:varchar(100)" json:"spec"`
	Concentration *string   `gorm:"type:varchar(50)" json:"concentration,omitempty"`
	SpecUnit      *string   `gorm:"type:varchar(20)" json:"specUnit,omitempty"`
	MinUnitDose   *string   `gorm:"type:varchar(20)" json:"minUnitDose,omitempty"`
	BaseUnit      string    `gorm:"column:unit;type:varchar(20)" json:"baseUnit"`
	Brand         *string   `gorm:"type:varchar(100)" json:"brand,omitempty"`
	Packaging     *string   `gorm:"type:varchar(50)" json:"packaging,omitempty"`
	Manufacturer  string    `gorm:"type:varchar(100)" json:"manufacturer"`
	StandardType  *string   `gorm:"type:varchar(50)" json:"standardType,omitempty"`
	Timing        *string   `gorm:"type:varchar(50)" json:"timing,omitempty"`
	Tips          *string   `gorm:"type:varchar(200)" json:"tips,omitempty"`
	SortOrder     int       `gorm:"default:0" json:"sortOrder"`
	IsEnabled     bool      `json:"isEnabled"`
	TenantID      *int64    `gorm:"index" json:"tenantId,omitempty"`
	Note          string    `gorm:"column:notes;type:text" json:"note"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (DrugCatalog) TableName() string {
	return "drug_catalogs"
}

// DrugCategory 药品分类常量
const (
	DrugCategoryAnticoagulant        = "抗凝剂"
	DrugCategoryAnticoagulantLow     = "低分子肝素"
	DrugCategoryAnticoagulantCitrate = "柠檬酸钠"
	DrugCategoryErythropoietin       = "促红素"
	DrugCategoryIron                 = "铁剂"
	DrugCategoryCalcium              = "钙剂"
	DrugCategoryVitaminD             = "维生素D"
	DrugCategoryAntihypertensive     = "降压药"
	DrugCategoryDiuretic             = "利尿剂"
	DrugCategoryOther                = "其他"
	DrugCategoryMethod               = "METHOD"
)

// ===== 医嘱模板 =====

// OrderTemplate 医嘱模板
type OrderTemplate struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Type      string    `gorm:"type:varchar(20);not null" json:"type"` // 长期, 临时
	Category  string    `gorm:"type:varchar(50);not null" json:"category"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Frequency *string   `gorm:"type:varchar(50)" json:"frequency"`
	Priority  string    `gorm:"type:varchar(20);default:普通" json:"priority"`
	IsDefault bool      `gorm:"default:false" json:"isDefault"`
	IsEnabled bool      `json:"isEnabled"`
	TenantID  *int64    `gorm:"index" json:"tenantId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 关联
	Items []OrderTemplateItem `gorm:"foreignKey:TemplateID" json:"items,omitempty"`
}

// TableName 指定表名
func (OrderTemplate) TableName() string {
	return "order_templates"
}

// OrderTemplateItem 医嘱模板条目
type OrderTemplateItem struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TemplateID  string    `gorm:"type:varchar(36);not null;index" json:"templateId"`
	DrugID      *uint     `gorm:"index" json:"drugId,omitempty"`
	DrugName    string    `gorm:"type:varchar(100);not null" json:"drugName"`
	Spec        *string   `gorm:"type:varchar(100)" json:"spec,omitempty"`
	MinUnitDose *string   `gorm:"type:varchar(20)" json:"minUnitDose,omitempty"`
	Dosage      *string   `gorm:"type:varchar(50)" json:"dosage,omitempty"`
	Unit        *string   `gorm:"type:varchar(20)" json:"unit,omitempty"`
	Route       *string   `gorm:"type:varchar(50)" json:"route,omitempty"`
	Frequency   *string   `gorm:"type:varchar(50)" json:"frequency,omitempty"`
	Timing      *string   `gorm:"type:varchar(50)" json:"timing,omitempty"`
	GroupID     *string   `gorm:"type:varchar(36)" json:"groupId,omitempty"`
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (OrderTemplateItem) TableName() string {
	return "order_template_items"
}

// BeforeCreate 自动生成 UUID
func (o *OrderTemplateItem) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// OrderTemplateCategory 医嘱模板分类常量
const (
	OrderTemplateCategoryMedicine  = "药品"
	OrderTemplateCategoryExam      = "检查"
	OrderTemplateCategoryTreatment = "治疗"
	OrderTemplateCategoryNursing   = "护理"
	OrderTemplateCategoryDiet      = "饮食"
)
