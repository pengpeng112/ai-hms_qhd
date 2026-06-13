// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DictType 字典类型
type DictType struct {
	ID          string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	Code        string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Description string     `gorm:"type:varchar(500)" json:"description"`
	Icon        string     `gorm:"type:varchar(50)" json:"icon"` // 图标（emoji）
	Source      string     `gorm:"type:varchar(20);default:local" json:"source"` // 来源：legacy 老库 / local 本地
	SortOrder   int        `gorm:"type:int;default:0" json:"sortOrder"`
	IsEnabled   bool       `gorm:"type:bool;default:true" json:"isEnabled"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Items       []DictItem `gorm:"foreignKey:TypeCode;references:Code" json:"items,omitempty"`
}

// BeforeCreate 创建前自动生成 UUID
func (d *DictType) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (DictType) TableName() string {
	return "dict_types"
}

// DictItem 字典项
type DictItem struct {
	ID          string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	TypeCode    string     `gorm:"type:varchar(50);index;not null" json:"typeCode"`
	Code        string     `gorm:"type:varchar(50);not null" json:"code"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Description string     `gorm:"type:varchar(500)" json:"description"`
	Source      string     `gorm:"type:varchar(20);default:local" json:"source"` // 来源：legacy 老库 / local 本地
	SortOrder   int        `gorm:"type:int;default:0" json:"sortOrder"`
	IsEnabled   bool       `gorm:"type:bool;default:true" json:"isEnabled"`
	Extra       string     `gorm:"type:varchar(500)" json:"extra,omitempty"`     // 扩展字段，如颜色标识
	ParentCode  string     `gorm:"type:varchar(50)" json:"parentCode,omitempty"` // 父级代码（用于树形结构）
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Children    []DictItem `gorm:"-" json:"children,omitempty"` // 子项（不存储在数据库）
}

// BeforeCreate 创建前自动生成 UUID
func (d *DictItem) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// DictItemTableIndex 定义 dict_items 表的唯一索引
// 注意：唯一索引已通过数据库迁移创建（idx_dict_items_unique）
// GORM 标签中的 uniqueIndex 在 AutoMigrate 时也会创建对应的约束

// TableName 指定表名
func (DictItem) TableName() string {
	return "dict_items"
}

// 字典类型代码常量
const (
	DictTypeDialysisMode   = "DIALYSIS_MODE"     // 透析方式
	DictTypeAnticoagulant  = "ANTICOAGULANT"     // 抗凝剂类型
	DictTypeDialysateType  = "DIALYSATE_TYPE"    // 透析液类型
	DictTypeDialysateGroup = "DIALYSATE_GROUP"   // 透析液组
	DictTypeDialysateFlow  = "DIALYSATE_FLOW"    // 透析液流量
	DictTypeGlucose        = "GLUCOSE"           // 葡萄糖类型
	DictTypeMaterialCat    = "MATERIAL_CATEGORY" // 材料分类
	DictTypeDrugCat        = "DRUG_CATEGORY"     // 药品分类
	DictTypeOrderType      = "ORDER_TYPE"        // 医嘱类型
	DictTypeOrderCategory  = "ORDER_CATEGORY"    // 医嘱分类
	DictTypePatientStatus  = "PATIENT_STATUS"    // 患者状态
	DictTypeVascularAccess = "VASCULAR_ACCESS"   // 血管通路类型
	DictTypeVascularSite   = "VASCULAR_SITE"     // 血管通路部位
	DictTypeVeinType       = "VEIN_TYPE"         // 静脉类型
	DictTypeArteryType     = "ARTERY_TYPE"       // 动脉类型
	DictTypeInsuranceType  = "INSURANCE_TYPE"    // 医保类型
	DictTypePatientType    = "PATIENT_TYPE"      // 患者类型
	DictTypeIDType         = "ID_TYPE"           // 证件类型
	DictTypeVisitCategory  = "VISIT_CATEGORY"    // 就诊类别
	DictTypeBloodTypeABO   = "BLOOD_TYPE_ABO"    // ABO血型
	DictTypeBloodTypeRH    = "BLOOD_TYPE_RH"     // Rh血型
	DictTypeEducationLevel = "EDUCATION_LEVEL"   // 文化程度
	DictTypeMaritalStatus  = "MARITAL_STATUS"    // 婚姻状况
	DictTypeDoctor         = "DOCTOR"            // 医生列表
	DictTypeHospital       = "HOSPITAL"          // 手术医院
	DictTypeSurgeryType    = "SURGERY_TYPE"      // 手术类型（血管通路干预）
	// 临床诊疗分类字典
	DictTypePrimaryDisease = "PRIMARY_DISEASE" // 原发病分类
	DictTypeComplication   = "COMPLICATION"    // 并发症类型
	DictTypePathology      = "PATHOLOGY"       // 病理诊断分类
	DictTypeTumor          = "TUMOR"           // 肿瘤分类
	DictTypeAllergen       = "ALLERGEN"        // 过敏原类型
	// 转归字典
	DictTypeOutcome = "OUTCOME" // 患者转归（一级：在科/转出，二级：具体原因）
	// 医嘱扩展字典
	DictTypeOrderRoute     = "ORDER_ROUTE"     // 用法
	DictTypeOrderFrequency = "ORDER_FREQUENCY" // 频次
	DictTypeOrderTiming    = "ORDER_TIMING"    // 使用时机
)
