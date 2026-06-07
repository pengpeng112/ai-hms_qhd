package services

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ===== 方案模板服务 =====

// PlanTemplateService 方案模板服务
type PlanTemplateService struct {
	db *gorm.DB
}

type legacyPlanTemplate struct {
	ID                    int64     `gorm:"column:Id"`
	TenantID              int64     `gorm:"column:TenantId"`
	Name                  string    `gorm:"column:Name"`
	CreatorID             int64     `gorm:"column:CreatorId"`
	CreateTime            time.Time `gorm:"column:CreateTime"`
	IsDisabled            bool      `gorm:"column:IsDisabled"`
	DialysisMethod        string    `gorm:"column:DialysisMethod"`
	DialysisDuration      float64   `gorm:"column:DialysisDuration"`
	AdjustQuantity        float64   `gorm:"column:AdjustQuantity"`
	BF                    float64   `gorm:"column:BF"`
	BV                    float64   `gorm:"column:BV"`
	FirstAnticoagulant    int64     `gorm:"column:FirstAnticoagulant"`
	FirstDosage           float64   `gorm:"column:FirstDosage"`
	MaintainAnticoagulant int64     `gorm:"column:MaintainAnticoagulant"`
	DilutionProportion    float64   `gorm:"column:DilutionProportion"`
	InjectionRate         float64   `gorm:"column:InjectionRate"`
	InjectionDuration     float64   `gorm:"column:InjectionDuration"`
	InjectionVolume       float64   `gorm:"column:InjectionVolume"`
	VascularAccessID      int64     `gorm:"column:VascularAccessId"`
	Dialysate             string    `gorm:"column:Dialysate"`
	DialysateFlow         float64   `gorm:"column:DialysateFlow"`
	DialysateVolume       float64   `gorm:"column:DialysateVolume"`
	NaIonCon              float64   `gorm:"column:NaIonCon"`
	CaIonCon              float64   `gorm:"column:CaIonCon"`
	KIonCon               float64   `gorm:"column:KIonCon"`
	Conductivity          float64   `gorm:"column:Conductivity"`
	DialysateTmp          float64   `gorm:"column:DialysateTmp"`
	SubstituateVolume     float64   `gorm:"column:SubstituateVolume"`
	DilutionMnt           string    `gorm:"column:DilutionMnt"`
	LastModifyTime        time.Time `gorm:"column:LastModifyTime"`
	HCO3IonCon            float64   `gorm:"column:HCO3IonCon"`
	GlucoseCon            float64   `gorm:"column:GlucoseCon"`
	DialysateGroupID      int64     `gorm:"column:DialysateGroupId"`
	Note                  string    `gorm:"column:Note"`
	SubstituateFlow       float64   `gorm:"column:SubstituateFlow"`
}

func (legacyPlanTemplate) TableName() string { return "Plan_PlanTPL" }

type legacyPlanTemplateMaterial struct {
	ID             int64     `gorm:"column:Id"`
	TenantID       int64     `gorm:"column:TenantId"`
	PlanTemplateID int64     `gorm:"column:PlanTPLId"`
	MaterialID     int64     `gorm:"column:MaterialId"`
	MaterialGroup  int64     `gorm:"column:MaterialGroup"`
	Num            float64   `gorm:"column:Num"`
	Note           string    `gorm:"column:Note"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime"`
}

func (legacyPlanTemplateMaterial) TableName() string { return "Plan_PlanTPLMaterial" }

type legacyMaterialCatalogRow struct {
	ID             int64     `gorm:"column:Id"`
	TenantID       int64     `gorm:"column:TenantId"`
	Name           string    `gorm:"column:Name"`
	Spell          string    `gorm:"column:Spell"`
	Classification string    `gorm:"column:Classification"`
	Code           string    `gorm:"column:Code"`
	Brand          string    `gorm:"column:Brand"`
	Specification  string    `gorm:"column:Specification"`
	Package        string    `gorm:"column:Package"`
	Manufacturer   string    `gorm:"column:Manufacturer"`
	Note           string    `gorm:"column:Note"`
	Type           string    `gorm:"column:Type"`
	IsDisabled     bool      `gorm:"column:IsDisabled"`
	CreatorID      int64     `gorm:"column:CreatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime"`
	StdCat         string    `gorm:"column:StdCat"`
	Unit           string    `gorm:"column:Unit"`
	ShortName      string    `gorm:"column:ShortName"`
	Sort           float64   `gorm:"column:Sort"`
}

func (legacyMaterialCatalogRow) TableName() string { return "Auxiliary_MaterialInfomation" }

type legacyOrderTemplateRow struct {
	ID             int64     `gorm:"column:Id"`
	TenantID       int64     `gorm:"column:TenantId"`
	Name           string    `gorm:"column:Name"`
	OrderGroup     int64     `gorm:"column:OrderGroup"`
	IsDisabled     bool      `gorm:"column:IsDisabled"`
	Classification string    `gorm:"column:Classification"`
	DrugID         int64     `gorm:"column:DrugId"`
	Content        string    `gorm:"column:Content"`
	Dosage         string    `gorm:"column:Dosage"`
	UseOpportunity string    `gorm:"column:UseOpportunity"`
	UseMethod      string    `gorm:"column:UseMethod"`
	UseWay         string    `gorm:"column:UseWay"`
	Note           string    `gorm:"column:Note"`
	CreatorID      int64     `gorm:"column:CreatorId"`
	CreateTime     time.Time `gorm:"column:CreateTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime"`
	UseNum         float64   `gorm:"column:UseNum"`
	AllDosage      float64   `gorm:"column:AllDosage"`
}

func (legacyOrderTemplateRow) TableName() string { return "Order_OrderTPL" }

// NewPlanTemplateService 创建方案模板服务
func NewPlanTemplateService() *PlanTemplateService {
	return &PlanTemplateService{
		db: database.GetDB(),
	}
}

func toMaterialCatalogDTO(item legacyMaterialCatalogRow) models.MaterialCatalog {
	var shortName *string
	if trimmed := strings.TrimSpace(item.ShortName); trimmed != "" {
		shortName = &trimmed
	}
	var mnemonic *string
	if trimmed := strings.TrimSpace(item.Spell); trimmed != "" {
		mnemonic = &trimmed
	}
	var standardType *string
	if trimmed := strings.TrimSpace(item.StdCat); trimmed != "" {
		standardType = &trimmed
	}
	var packaging *string
	if trimmed := strings.TrimSpace(item.Package); trimmed != "" {
		packaging = &trimmed
	}
	var manufacturer *string
	if trimmed := strings.TrimSpace(item.Manufacturer); trimmed != "" {
		manufacturer = &trimmed
	}
	return models.MaterialCatalog{
		ID:           uint(item.ID),
		Code:         strings.TrimSpace(item.Code),
		Name:         strings.TrimSpace(item.Name),
		ShortName:    shortName,
		Mnemonic:     mnemonic,
		Category:     strings.TrimSpace(item.Classification),
		Spec:         strings.TrimSpace(item.Specification),
		StandardType: standardType,
		Brand:        strings.TrimSpace(item.Brand),
		Unit:         strings.TrimSpace(item.Unit),
		Packaging:    packaging,
		Manufacturer: manufacturer,
		SortOrder:    int(item.Sort),
		IsEnabled:    !item.IsDisabled,
		Notes:        strings.TrimSpace(item.Note),
		CreatedAt:    item.CreateTime,
		UpdatedAt:    item.LastModifyTime,
	}
}

func normalizeLegacyTemplateMode(raw string) string {
	v := strings.ToUpper(strings.TrimSpace(raw))
	switch v {
	case "HD", "HDF", "HP", "HF", "HFD", "HD+HP", "HP+HD", "PE":
		return v
	default:
		return strings.TrimSpace(raw)
	}
}

func (s *PlanTemplateService) loadLegacyPlanTemplateMaterials(templateIDs []int64) (map[int64][]models.PlanTemplateMaterial, error) {
	result := make(map[int64][]models.PlanTemplateMaterial, len(templateIDs))
	if len(templateIDs) == 0 {
		return result, nil
	}

	var rows []legacyPlanTemplateMaterial
	if err := s.db.Where(`"TenantId" = ? AND "PlanTPLId" IN ?`, LegacyTenantID, templateIDs).
		Order(`"MaterialGroup" ASC`).
		Order(`"Id" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return result, nil
	}

	materialIDs := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.MaterialID]; ok || row.MaterialID <= 0 {
			continue
		}
		seen[row.MaterialID] = struct{}{}
		materialIDs = append(materialIDs, row.MaterialID)
	}

	var materials []legacyMaterialCatalogRow
	if len(materialIDs) > 0 {
		if err := s.db.Where(`"TenantId" = ? AND "Id" IN ?`, LegacyTenantID, materialIDs).Find(&materials).Error; err != nil {
			return nil, err
		}
	}
	materialMap := make(map[int64]legacyMaterialCatalogRow, len(materials))
	for _, material := range materials {
		materialMap[material.ID] = material
	}

	for _, row := range rows {
		material := materialMap[row.MaterialID]
		result[row.PlanTemplateID] = append(result[row.PlanTemplateID], models.PlanTemplateMaterial{
			ID:       strconv.FormatInt(row.MaterialID, 10),
			Name:     strings.TrimSpace(material.Name),
			Category: strings.TrimSpace(material.Classification),
			Count:    int(row.Num),
			Code:     strings.TrimSpace(material.Code),
			Brand:    strings.TrimSpace(material.Brand),
			Spec:     strings.TrimSpace(material.Specification),
			Note:     strings.TrimSpace(row.Note),
		})
	}

	return result, nil
}

func (s *PlanTemplateService) loadLegacyDrugNames(ids ...int64) (map[int64]string, error) {
	result := make(map[int64]string)
	unique := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return result, nil
	}

	var rows []legacyDrugCatalog
	if err := s.db.Where(`"TenantId" = ? AND "Id" IN ?`, LegacyTenantID, unique).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ID] = strings.TrimSpace(row.Name)
	}
	return result, nil
}

func toPlanTemplateDTO(item legacyPlanTemplate, materials []models.PlanTemplateMaterial, drugNames map[int64]string) models.PlanTemplate {
	dto := models.PlanTemplate{
		ID:          strconv.FormatInt(item.ID, 10),
		Name:        strings.TrimSpace(item.Name),
		Description: "",
		Mode:        normalizeLegacyTemplateMode(item.DialysisMethod),
		Category:    "",
		IsDefault:   false,
		IsEnabled:   !item.IsDisabled,
		TenantID:    &item.TenantID,
		CreatedAt:   item.CreateTime,
		UpdatedAt:   item.LastModifyTime,
	}
	dto.TemplateContent.Duration = int(item.DialysisDuration)
	dto.TemplateContent.DryWeight = 0
	dto.TemplateContent.DialysisMode.Mode = normalizeLegacyTemplateMode(item.DialysisMethod)
	dto.TemplateContent.DialysisMode.BloodFlow = int(item.BF)
	dto.TemplateContent.DialysisMode.SubstituteInputMode = strings.TrimSpace(item.DilutionMnt)
	dto.TemplateContent.DialysisMode.SubstituteFlow = item.SubstituateFlow
	dto.TemplateContent.DialysisMode.SubstituteVolume = item.SubstituateVolume
	dto.TemplateContent.DialysisMode.BV = formatLegacyNumber(item.BV)
	dto.TemplateContent.DialysisMode.Notes = strings.TrimSpace(item.Note)
	dto.TemplateContent.Anticoagulant.InitialDrug = strings.TrimSpace(drugNames[item.FirstAnticoagulant])
	dto.TemplateContent.Anticoagulant.InitialDose = formatLegacyNumber(item.FirstDosage)
	dto.TemplateContent.Anticoagulant.TotalDose = formatLegacyNumber(item.InjectionVolume)
	dto.TemplateContent.Anticoagulant.MaintenanceDrug = strings.TrimSpace(drugNames[item.MaintainAnticoagulant])
	dto.TemplateContent.Anticoagulant.InfusionRate = formatLegacyNumber(item.InjectionRate)
	dto.TemplateContent.Anticoagulant.InfusionTime = formatLegacyNumber(item.InjectionDuration)
	dto.TemplateContent.Anticoagulant.MaintenanceDose = formatLegacyNumber(item.DilutionProportion)
	dto.TemplateContent.Parameters.DialysateType = strings.TrimSpace(item.Dialysate)
	dto.TemplateContent.Parameters.DialysateGroup = strconv.FormatInt(item.DialysateGroupID, 10)
	dto.TemplateContent.Parameters.FlowRate = int(item.DialysateFlow)
	dto.TemplateContent.Parameters.Na = item.NaIonCon
	dto.TemplateContent.Parameters.Ca = item.CaIonCon
	dto.TemplateContent.Parameters.K = item.KIonCon
	dto.TemplateContent.Parameters.HCO3 = item.HCO3IonCon
	dto.TemplateContent.Parameters.Glucose = formatLegacyNumber(item.GlucoseCon)
	dto.TemplateContent.Parameters.Conductivity = item.Conductivity
	dto.TemplateContent.Parameters.Temp = item.DialysateTmp
	dto.TemplateContent.Parameters.Volume = item.DialysateVolume
	dto.TemplateContent.Materials = materials
	return dto
}

func (s *PlanTemplateService) syncLegacyPlanTemplateMaterials(tx *gorm.DB, templateID int64, materials []models.PlanTemplateMaterial) error {
	if err := tx.Where(`"PlanTPLId" = ? AND "TenantId" = ?`, templateID, LegacyTenantID).
		Delete(&legacyPlanTemplateMaterial{}).Error; err != nil {
		return err
	}

	patientService := &PatientService{db: tx}
	now := time.Now()
	for idx, material := range materials {
		materialID, err := patientService.findLegacyMaterialID(models.Material{
			ID:    material.ID,
			Name:  material.Name,
			Code:  material.Code,
			Brand: material.Brand,
			Spec:  material.Spec,
			Note:  material.Note,
		})
		if err != nil {
			return err
		}
		if materialID == 0 {
			continue
		}
		id, err := nextLegacyNumericID(tx, `"Plan_PlanTPLMaterial"`)
		if err != nil {
			return err
		}
		row := map[string]any{
			"Id":             id,
			"TenantId":       LegacyTenantID,
			"PlanTPLId":      templateID,
			"MaterialId":     materialID,
			"MaterialGroup":  idx,
			"Num":            material.Count,
			"Note":           strings.TrimSpace(material.Note),
			"LastModifyTime": now,
		}
		if err := tx.Table(`"Plan_PlanTPLMaterial"`).Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *PlanTemplateService) buildLegacyPlanTemplateValues(req PlanTemplateCreateRequest) (map[string]any, error) {
	patientService := &PatientService{db: s.db}
	firstDrugID, err := patientService.findLegacyDrugIDByName(req.TemplateContent.Anticoagulant.InitialDrug)
	if err != nil {
		return nil, err
	}
	maintainDrugID, err := patientService.findLegacyDrugIDByName(req.TemplateContent.Anticoagulant.MaintenanceDrug)
	if err != nil {
		return nil, err
	}

	mode := req.TemplateContent.DialysisMode
	params := req.TemplateContent.Parameters
	anticoagulant := req.TemplateContent.Anticoagulant
	now := time.Now()
	values := map[string]any{
		"TenantId":              LegacyTenantID,
		"Name":                  strings.TrimSpace(req.Name),
		"CreatorId":             int64(0),
		"CreateTime":            now,
		"IsDisabled":            !req.IsEnabled,
		"DialysisMethod":        normalizeLegacyTemplateMode(firstNonEmpty(mode.Mode, req.Mode)),
		"DialysisDuration":      req.TemplateContent.Duration,
		"AdjustQuantity":        0,
		"BF":                    mode.BloodFlow,
		"BV":                    parseStringFloat(mode.BV),
		"FirstAnticoagulant":    firstDrugID,
		"FirstDosage":           parseStringFloat(anticoagulant.InitialDose),
		"MaintainAnticoagulant": maintainDrugID,
		"DilutionProportion":    parseStringFloat(anticoagulant.MaintenanceDose),
		"InjectionRate":         parseStringFloat(anticoagulant.InfusionRate),
		"InjectionDuration":     parseStringFloat(anticoagulant.InfusionTime),
		"InjectionVolume":       parseStringFloat(anticoagulant.TotalDose),
		"VascularAccessId":      0,
		"Dialysate":             strings.TrimSpace(params.DialysateType),
		"DialysateFlow":         params.FlowRate,
		"DialysateVolume":       params.Volume,
		"NaIonCon":              params.Na,
		"CaIonCon":              params.Ca,
		"KIonCon":               params.K,
		"Conductivity":          params.Conductivity,
		"DialysateTmp":          params.Temp,
		"SubstituateVolume":     mode.SubstituteVolume,
		"DilutionMnt":           strings.TrimSpace(mode.SubstituteInputMode),
		"LastModifyTime":        now,
		"HCO3IonCon":            params.HCO3,
		"GlucoseCon":            parseStringFloat(params.Glucose),
		"DialysateGroupId":      parseLegacyNumericID(params.DialysateGroup),
		"Note":                  strings.TrimSpace(mode.Notes),
		"SubstituateFlow":       mode.SubstituteFlow,
	}
	return values, nil
}

// PlanTemplateListRequest 获取方案模板列表请求
type PlanTemplateListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Search    string `form:"search"`
	Mode      string `form:"mode"`      // HD, HDF, HP, HF, HFD
	Category  string `form:"category"`  // 分类
	IsEnabled *bool  `form:"isEnabled"` // 启用状态
}

// PlanTemplateListResponse 获取方案模板列表响应
type PlanTemplateListResponse struct {
	Items     []models.PlanTemplate `json:"items"`
	Total     int64                `json:"total"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"pageSize"`
	TotalPage int                  `json:"totalPage"`
}

// List 获取方案模板列表
func (s *PlanTemplateService) List(req PlanTemplateListRequest) (*PlanTemplateListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.PlanTemplate{})

	// 筛选条件
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}
	if req.Mode != "" {
		query = query.Where("mode = ?", req.Mode)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.PlanTemplate
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("is_default DESC, created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &PlanTemplateListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

func (s *PlanTemplateService) LegacyList(req PlanTemplateListRequest) (*PlanTemplateListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&legacyPlanTemplate{}).Where(`"TenantId" = ?`, LegacyTenantID)
	if req.Search != "" {
		query = query.Where(`"Name" LIKE ?`, "%"+strings.TrimSpace(req.Search)+"%")
	}
	if req.Mode != "" {
		query = query.Where(`"DialysisMethod" = ?`, strings.TrimSpace(req.Mode))
	}
	if req.IsEnabled != nil {
		query = query.Where(`COALESCE("IsDisabled", false) = ?`, !*req.IsEnabled)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []legacyPlanTemplate
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).
		Order(`COALESCE("IsDisabled", false) ASC`).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	templateIDs := make([]int64, 0, len(rows))
	drugIDs := make([]int64, 0, len(rows)*2)
	for _, row := range rows {
		templateIDs = append(templateIDs, row.ID)
		drugIDs = append(drugIDs, row.FirstAnticoagulant, row.MaintainAnticoagulant)
	}
	materialsMap, err := s.loadLegacyPlanTemplateMaterials(templateIDs)
	if err != nil {
		return nil, err
	}
	drugNames, err := s.loadLegacyDrugNames(drugIDs...)
	if err != nil {
		return nil, err
	}

	items := make([]models.PlanTemplate, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPlanTemplateDTO(row, materialsMap[row.ID], drugNames))
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}
	return &PlanTemplateListResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage}, nil
}

// Get 获取方案模板详情
func (s *PlanTemplateService) Get(id string) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var template models.PlanTemplate
	err := s.db.First(&template, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan template not found")
		}
		return nil, err
	}

	return &template, nil
}

func (s *PlanTemplateService) LegacyGet(id string) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var row legacyPlanTemplate
	if err := s.db.Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan template not found")
		}
		return nil, err
	}
	materialsMap, err := s.loadLegacyPlanTemplateMaterials([]int64{row.ID})
	if err != nil {
		return nil, err
	}
	drugNames, err := s.loadLegacyDrugNames(row.FirstAnticoagulant, row.MaintainAnticoagulant)
	if err != nil {
		return nil, err
	}
	dto := toPlanTemplateDTO(row, materialsMap[row.ID], drugNames)
	return &dto, nil
}

// PlanTemplateCreateRequest 创建方案模板请求
type PlanTemplateCreateRequest struct {
	Name        string                         `json:"name" binding:"required"`
	Description string                         `json:"description"`
	Mode        string                         `json:"mode" binding:"required,oneof=HD HFD HP HF HDF"`
	Category    string                         `json:"category"`
	IsDefault   bool                           `json:"isDefault"`
	IsEnabled   bool                           `json:"isEnabled"`
	TemplateContent models.PlanTemplateContent `json:"templateContent"`
}

// Create 创建方案模板
func (s *PlanTemplateService) Create(req PlanTemplateCreateRequest) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 如果设置为默认模板，先取消其他默认模板
	if req.IsDefault {
		s.db.Model(&models.PlanTemplate{}).
			Where("mode = ? AND is_default = ?", req.Mode, true).
			Update("is_default", false)
	}

	template := models.PlanTemplate{
		ID:              utils.GenerateID(),
		Name:            req.Name,
		Description:     req.Description,
		Mode:            req.Mode,
		Category:        req.Category,
		IsDefault:       req.IsDefault,
		IsEnabled:       req.IsEnabled,
		TemplateContent: req.TemplateContent,
	}

	if err := s.db.Create(&template).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

func (s *PlanTemplateService) LegacyCreate(req PlanTemplateCreateRequest) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	templateID, err := nextLegacyNumericID(s.db, `"Plan_PlanTPL"`)
	if err != nil {
		return nil, err
	}
	values, err := s.buildLegacyPlanTemplateValues(req)
	if err != nil {
		return nil, err
	}
	values["Id"] = templateID

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(`"Plan_PlanTPL"`).Create(values).Error; err != nil {
			return err
		}
		return s.syncLegacyPlanTemplateMaterials(tx, templateID, req.TemplateContent.Materials)
	}); err != nil {
		return nil, err
	}
	return s.LegacyGet(strconv.FormatInt(templateID, 10))
}

// PlanTemplateUpdateRequest 更新方案模板请求
type PlanTemplateUpdateRequest struct {
	Name           *string                         `json:"name"`
	Description    *string                         `json:"description"`
	Mode           *string                         `json:"mode"`
	Category       *string                         `json:"category"`
	IsDefault      *bool                           `json:"isDefault"`
	IsEnabled      *bool                           `json:"isEnabled"`
	TemplateContent *models.PlanTemplateContent    `json:"templateContent"`
}

// Update 更新方案模板
func (s *PlanTemplateService) Update(id string, req PlanTemplateUpdateRequest) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var template models.PlanTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan template not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Mode != nil {
		updates["mode"] = *req.Mode
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.IsDefault != nil {
		// 如果设置为默认模板，先取消同模式的其他默认模板
		if *req.IsDefault {
			s.db.Model(&models.PlanTemplate{}).
				Where("mode = ? AND is_default = ? AND id != ?", template.Mode, true, id).
				Update("is_default", false)
		}
		updates["is_default"] = *req.IsDefault
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.TemplateContent != nil {
		updates["template_content"] = *req.TemplateContent
	}

	if err := s.db.Model(&template).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

func (s *PlanTemplateService) LegacyUpdate(id string, req PlanTemplateUpdateRequest) (*models.PlanTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var current legacyPlanTemplate
	if err := s.db.Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan template not found")
		}
		return nil, err
	}

	merged := PlanTemplateCreateRequest{
		Name:      strings.TrimSpace(current.Name),
		Mode:      normalizeLegacyTemplateMode(current.DialysisMethod),
		IsEnabled: !current.IsDisabled,
		TemplateContent: toPlanTemplateDTO(current, nil, map[int64]string{}).TemplateContent,
	}
	if req.Name != nil {
		merged.Name = strings.TrimSpace(*req.Name)
	}
	if req.Mode != nil {
		merged.Mode = strings.TrimSpace(*req.Mode)
		merged.TemplateContent.DialysisMode.Mode = strings.TrimSpace(*req.Mode)
	}
	if req.IsEnabled != nil {
		merged.IsEnabled = *req.IsEnabled
	}
	if req.TemplateContent != nil {
		merged.TemplateContent = *req.TemplateContent
		if strings.TrimSpace(merged.Mode) == "" {
			merged.Mode = merged.TemplateContent.DialysisMode.Mode
		}
	}

	values, err := s.buildLegacyPlanTemplateValues(merged)
	if err != nil {
		return nil, err
	}
	delete(values, "Id")
	delete(values, "TenantId")
	delete(values, "CreateTime")
	values["LastModifyTime"] = time.Now()

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(`"Plan_PlanTPL"`).
			Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
			Updates(values).Error; err != nil {
			return err
		}
		if req.TemplateContent != nil {
			return s.syncLegacyPlanTemplateMaterials(tx, current.ID, req.TemplateContent.Materials)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.LegacyGet(id)
}

// Delete 删除方案模板
func (s *PlanTemplateService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.PlanTemplate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("plan template not found")
	}

	return nil
}

func (s *PlanTemplateService) LegacyDelete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Table(`"Plan_PlanTPL"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		Updates(map[string]any{
			"IsDisabled":     true,
			"LastModifyTime": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("plan template not found")
	}
	return nil
}

// ToggleEnabled 切换启用状态
func (s *PlanTemplateService) ToggleEnabled(id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var template models.PlanTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("plan template not found")
		}
		return false, err
	}

	newStatus := !template.IsEnabled
	if err := s.db.Model(&template).Update("is_enabled", newStatus).Error; err != nil {
		return false, err
	}

	return newStatus, nil
}

func (s *PlanTemplateService) LegacyToggleEnabled(id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}
	var row legacyPlanTemplate
	if err := s.db.Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("plan template not found")
		}
		return false, err
	}
	newIsDisabled := !row.IsDisabled
	if err := s.db.Table(`"Plan_PlanTPL"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		Updates(map[string]any{
			"IsDisabled":     newIsDisabled,
			"LastModifyTime": time.Now(),
		}).Error; err != nil {
		return false, err
	}
	return !newIsDisabled, nil
}

// SetDefault 设置默认模板
func (s *PlanTemplateService) SetDefault(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	var template models.PlanTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("plan template not found")
		}
		return err
	}

	// 取消同模式的其他默认模板
	s.db.Model(&models.PlanTemplate{}).
		Where("mode = ? AND is_default = ?", template.Mode, true).
		Update("is_default", false)

	// 设置当前模板为默认
	if err := s.db.Model(&template).Update("is_default", true).Error; err != nil {
		return err
	}

	return nil
}

// ===== 材料目录服务 =====

// MaterialCatalogService 材料目录服务
type MaterialCatalogService struct {
	db *gorm.DB
}

// NewMaterialCatalogService 创建材料目录服务
func NewMaterialCatalogService() *MaterialCatalogService {
	return &MaterialCatalogService{
		db: database.GetDB(),
	}
}

// MaterialCatalogListRequest 获取材料目录列表请求
type MaterialCatalogListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Search    string `form:"search"`
	Category  string `form:"category"`
	IsEnabled *bool  `form:"isEnabled"`
}

// MaterialCatalogListResponse 获取材料目录列表响应
type MaterialCatalogListResponse struct {
	Items     []models.MaterialCatalog `json:"items"`
	Total     int64                    `json:"total"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"pageSize"`
	TotalPage int                      `json:"totalPage"`
}

// List 获取材料目录列表
func (s *MaterialCatalogService) List(req MaterialCatalogListRequest) (*MaterialCatalogListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.MaterialCatalog{})

	// 筛选条件
	if req.Search != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.MaterialCatalog
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("sort_order ASC, category, code, created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &MaterialCatalogListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

func (s *MaterialCatalogService) LegacyList(req MaterialCatalogListRequest) (*MaterialCatalogListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&legacyMaterialCatalogRow{}).Where(`"TenantId" = ?`, LegacyTenantID)
	if req.Search != "" {
		like := "%" + strings.TrimSpace(req.Search) + "%"
		query = query.Where(`("Name" LIKE ? OR "Code" LIKE ? OR "Spell" LIKE ?)`, like, like, like)
	}
	if req.Category != "" {
		query = query.Where(`"Classification" = ?`, strings.TrimSpace(req.Category))
	}
	if req.IsEnabled != nil {
		query = query.Where(`COALESCE("IsDisabled", false) = ?`, !*req.IsEnabled)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []legacyMaterialCatalogRow
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).
		Order(`"Classification" ASC`).
		Order(`"Sort" ASC`).
		Order(`"Code" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]models.MaterialCatalog, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMaterialCatalogDTO(row))
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}
	return &MaterialCatalogListResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage}, nil
}

// Get 获取材料目录详情
func (s *MaterialCatalogService) Get(id string) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var catalog models.MaterialCatalog
	err := s.db.First(&catalog, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material catalog not found")
		}
		return nil, err
	}

	return &catalog, nil
}

func (s *MaterialCatalogService) LegacyGet(id string) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var row legacyMaterialCatalogRow
	if err := s.db.Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material catalog not found")
		}
		return nil, err
	}
	dto := toMaterialCatalogDTO(row)
	return &dto, nil
}

// MaterialCatalogCreateRequest 创建材料目录请求
type MaterialCatalogCreateRequest struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	ShortName    string `json:"shortName"`
	Mnemonic     string `json:"mnemonic"`
	Category     string `json:"category"`
	Spec         string `json:"spec"`
	StandardType string `json:"standardType"`
	Brand        string `json:"brand"`
	Unit         string `json:"unit"`
	Packaging    string `json:"packaging"`
	Manufacturer string `json:"manufacturer"`
	SortOrder    int    `json:"sortOrder"`
	IsEnabled    bool   `json:"isEnabled"`
	Notes        string `json:"notes"`
}

// ValidateMaterialCatalogCreateRequest 校验创建材料目录请求
func ValidateMaterialCatalogCreateRequest(req MaterialCatalogCreateRequest) error {
	missing := make([]string, 0, 4)
	if strings.TrimSpace(req.Category) == "" {
		missing = append(missing, "材料分类")
	}
	if strings.TrimSpace(req.Name) == "" {
		missing = append(missing, "材料名称")
	}
	if strings.TrimSpace(req.ShortName) == "" {
		missing = append(missing, "简称")
	}
	if strings.TrimSpace(req.Unit) == "" {
		missing = append(missing, "单位")
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少必填字段：%s", strings.Join(missing, "、"))
	}
	return nil
}

func ensureMaterialCatalogCode(code string) string {
	trimmed := strings.TrimSpace(code)
	if trimmed != "" {
		return trimmed
	}
	return "AUTO-" + uuid.NewString()
}

func optionalStringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *MaterialCatalogService) resolveMaterialCategory(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || s.db == nil {
		return trimmed
	}

	var dictItem models.DictItem
	err := s.db.
		Where("type_code = ? AND is_enabled = ? AND code = ?", models.DictTypeMaterialCat, true, trimmed).
		Order("sort_order ASC").
		First(&dictItem).Error
	if err == nil {
		return strings.TrimSpace(dictItem.Name)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return trimmed
	}

	err = s.db.
		Where("type_code = ? AND is_enabled = ? AND name = ?", models.DictTypeMaterialCat, true, trimmed).
		Order("sort_order ASC").
		First(&dictItem).Error
	if err == nil {
		return strings.TrimSpace(dictItem.Name)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return trimmed
	}

	// 兼容历史停用字典项
	err = s.db.
		Where("type_code = ? AND code = ?", models.DictTypeMaterialCat, trimmed).
		Order("sort_order ASC").
		First(&dictItem).Error
	if err == nil {
		return strings.TrimSpace(dictItem.Name)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return trimmed
	}

	err = s.db.
		Where("type_code = ? AND name = ?", models.DictTypeMaterialCat, trimmed).
		Order("sort_order ASC").
		First(&dictItem).Error
	if err == nil {
		return strings.TrimSpace(dictItem.Name)
	}

	return trimmed
}

func mapMaterialCatalogWriteError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "material_catalogs_code") {
			return errors.New("material code already exists")
		}
		return errors.New("material catalog already exists")
	}

	return err
}

// Create 创建材料目录
func (s *MaterialCatalogService) Create(req MaterialCatalogCreateRequest) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	code := ensureMaterialCatalogCode(req.Code)

	// 检查代码是否已存在
	var count int64
	if err := s.db.Model(&models.MaterialCatalog{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("material code already exists")
	}

	catalog := models.MaterialCatalog{
		Code:         code,
		Name:         strings.TrimSpace(req.Name),
		ShortName:    optionalStringPtr(req.ShortName),
		Mnemonic:     optionalStringPtr(req.Mnemonic),
		Category:     s.resolveMaterialCategory(req.Category),
		Spec:         strings.TrimSpace(req.Spec),
		StandardType: optionalStringPtr(req.StandardType),
		Brand:        strings.TrimSpace(req.Brand),
		Unit:         strings.TrimSpace(req.Unit),
		Packaging:    optionalStringPtr(req.Packaging),
		Manufacturer: optionalStringPtr(req.Manufacturer),
		SortOrder:    req.SortOrder,
		IsEnabled:    req.IsEnabled,
		Notes:        strings.TrimSpace(req.Notes),
	}

	if err := s.db.Create(&catalog).Error; err != nil {
		return nil, mapMaterialCatalogWriteError(err)
	}

	return &catalog, nil
}

func (s *MaterialCatalogService) LegacyCreate(req MaterialCatalogCreateRequest) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	id, err := nextLegacyNumericID(s.db, `"Auxiliary_MaterialInfomation"`)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	row := map[string]any{
		"Id":             id,
		"TenantId":       LegacyTenantID,
		"Name":           strings.TrimSpace(req.Name),
		"Spell":          strings.TrimSpace(req.Mnemonic),
		"Classification": strings.TrimSpace(req.Category),
		"Code":           ensureMaterialCatalogCode(req.Code),
		"Brand":          strings.TrimSpace(req.Brand),
		"Specification":  strings.TrimSpace(req.Spec),
		"Package":        strings.TrimSpace(req.Packaging),
		"Manufacturer":   strings.TrimSpace(req.Manufacturer),
		"Note":           strings.TrimSpace(req.Notes),
		"Type":           strings.TrimSpace(req.StandardType),
		"IsDisabled":     !req.IsEnabled,
		"CreatorId":      int64(0),
		"CreateTime":     now,
		"LastModifyTime": now,
		"StdCat":         strings.TrimSpace(req.StandardType),
		"Unit":           strings.TrimSpace(req.Unit),
		"ShortName":      strings.TrimSpace(req.ShortName),
		"Sort":           req.SortOrder,
	}
	if err := s.db.Table(`"Auxiliary_MaterialInfomation"`).Create(row).Error; err != nil {
		return nil, err
	}
	return s.LegacyGet(strconv.FormatInt(id, 10))
}

// MaterialCatalogUpdateRequest 更新材料目录请求
type MaterialCatalogUpdateRequest struct {
	Code         *string `json:"code"`
	Name         *string `json:"name"`
	ShortName    *string `json:"shortName"`
	Mnemonic     *string `json:"mnemonic"`
	Category     *string `json:"category"`
	Spec         *string `json:"spec"`
	StandardType *string `json:"standardType"`
	Brand        *string `json:"brand"`
	Unit         *string `json:"unit"`
	Packaging    *string `json:"packaging"`
	Manufacturer *string `json:"manufacturer"`
	SortOrder    *int    `json:"sortOrder"`
	IsEnabled    *bool   `json:"isEnabled"`
	Notes        *string `json:"notes"`
}

// Update 更新材料目录
func (s *MaterialCatalogService) Update(id uint, req MaterialCatalogUpdateRequest) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var catalog models.MaterialCatalog
	if err := s.db.First(&catalog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material catalog not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Code != nil {
		code := strings.TrimSpace(*req.Code)
		if code == "" {
			return nil, errors.New("material code cannot be empty")
		}

		var duplicate int64
		if err := s.db.Model(&models.MaterialCatalog{}).
			Where("code = ? AND id <> ?", code, id).
			Count(&duplicate).Error; err != nil {
			return nil, err
		}
		if duplicate > 0 {
			return nil, errors.New("material code already exists")
		}
		updates["code"] = code
	}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.ShortName != nil {
		updates["short_name"] = strings.TrimSpace(*req.ShortName)
	}
	if req.Mnemonic != nil {
		updates["mnemonic"] = strings.TrimSpace(*req.Mnemonic)
	}
	if req.Category != nil {
		updates["category"] = s.resolveMaterialCategory(*req.Category)
	}
	if req.Spec != nil {
		updates["spec"] = strings.TrimSpace(*req.Spec)
	}
	if req.StandardType != nil {
		updates["standard_type"] = strings.TrimSpace(*req.StandardType)
	}
	if req.Brand != nil {
		updates["brand"] = strings.TrimSpace(*req.Brand)
	}
	if req.Unit != nil {
		updates["unit"] = strings.TrimSpace(*req.Unit)
	}
	if req.Packaging != nil {
		updates["packaging"] = strings.TrimSpace(*req.Packaging)
	}
	if req.Manufacturer != nil {
		updates["manufacturer"] = strings.TrimSpace(*req.Manufacturer)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.Notes != nil {
		updates["notes"] = strings.TrimSpace(*req.Notes)
	}

	if err := s.db.Model(&catalog).Updates(updates).Error; err != nil {
		return nil, mapMaterialCatalogWriteError(err)
	}

	// 重新获取更新后的数据
	if err := s.db.First(&catalog, id).Error; err != nil {
		return nil, err
	}

	return &catalog, nil
}

func (s *MaterialCatalogService) LegacyUpdate(id uint, req MaterialCatalogUpdateRequest) (*models.MaterialCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var row legacyMaterialCatalogRow
	if err := s.db.Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material catalog not found")
		}
		return nil, err
	}
	updates := map[string]any{"LastModifyTime": time.Now()}
	if req.Code != nil {
		updates["Code"] = strings.TrimSpace(*req.Code)
	}
	if req.Name != nil {
		updates["Name"] = strings.TrimSpace(*req.Name)
	}
	if req.ShortName != nil {
		updates["ShortName"] = strings.TrimSpace(*req.ShortName)
	}
	if req.Mnemonic != nil {
		updates["Spell"] = strings.TrimSpace(*req.Mnemonic)
	}
	if req.Category != nil {
		updates["Classification"] = strings.TrimSpace(*req.Category)
	}
	if req.Spec != nil {
		updates["Specification"] = strings.TrimSpace(*req.Spec)
	}
	if req.StandardType != nil {
		updates["StdCat"] = strings.TrimSpace(*req.StandardType)
		updates["Type"] = strings.TrimSpace(*req.StandardType)
	}
	if req.Brand != nil {
		updates["Brand"] = strings.TrimSpace(*req.Brand)
	}
	if req.Unit != nil {
		updates["Unit"] = strings.TrimSpace(*req.Unit)
	}
	if req.Packaging != nil {
		updates["Package"] = strings.TrimSpace(*req.Packaging)
	}
	if req.Manufacturer != nil {
		updates["Manufacturer"] = strings.TrimSpace(*req.Manufacturer)
	}
	if req.SortOrder != nil {
		updates["Sort"] = *req.SortOrder
	}
	if req.IsEnabled != nil {
		updates["IsDisabled"] = !*req.IsEnabled
	}
	if req.Notes != nil {
		updates["Note"] = strings.TrimSpace(*req.Notes)
	}
	if err := s.db.Table(`"Auxiliary_MaterialInfomation"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.LegacyGet(strconv.FormatUint(uint64(id), 10))
}

// Delete 删除材料目录
func (s *MaterialCatalogService) Delete(id uint) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.MaterialCatalog{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("material catalog not found")
	}

	return nil
}

func (s *MaterialCatalogService) LegacyDelete(id uint) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Table(`"Auxiliary_MaterialInfomation"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		Updates(map[string]any{"IsDisabled": true, "LastModifyTime": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("material catalog not found")
	}
	return nil
}

// ToggleEnabled 切换启用状态
func (s *MaterialCatalogService) ToggleEnabled(id uint) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var catalog models.MaterialCatalog
	if err := s.db.First(&catalog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("material catalog not found")
		}
		return false, err
	}

	newStatus := !catalog.IsEnabled
	if err := s.db.Model(&catalog).Update("is_enabled", newStatus).Error; err != nil {
		return false, err
	}

	return newStatus, nil
}

func (s *MaterialCatalogService) LegacyToggleEnabled(id uint) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}
	var row legacyMaterialCatalogRow
	if err := s.db.Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("material catalog not found")
		}
		return false, err
	}
	newIsDisabled := !row.IsDisabled
	if err := s.db.Table(`"Auxiliary_MaterialInfomation"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		Updates(map[string]any{"IsDisabled": newIsDisabled, "LastModifyTime": time.Now()}).Error; err != nil {
		return false, err
	}
	return !newIsDisabled, nil
}

// GetCategories 获取材料分类列表
func (s *MaterialCatalogService) GetCategories() ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var categories []string
	if err := s.db.Model(&models.MaterialCatalog{}).
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *MaterialCatalogService) LegacyGetCategories() ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var categories []string
	if err := s.db.Model(&legacyMaterialCatalogRow{}).
		Where(`"TenantId" = ? AND COALESCE("IsDisabled", false) = false`, LegacyTenantID).
		Distinct(`"Classification"`).
		Pluck(`"Classification"`, &categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func toOrderTemplateType(item legacyOrderTemplateRow) string {
	lower := strings.ToLower(strings.TrimSpace(item.Content + " " + item.Note))
	if strings.Contains(lower, "临时") {
		return "临时"
	}
	return "长期"
}

func toOrderTemplateDTO(rows []legacyOrderTemplateRow, drugMap map[int64]legacyDrugCatalog) models.OrderTemplate {
	first := rows[0]
	templateID := first.ID
	if first.OrderGroup > 0 {
		templateID = first.OrderGroup
	}
	template := models.OrderTemplate{
		ID:        strconv.FormatInt(templateID, 10),
		Name:      strings.TrimSpace(first.Name),
		Type:      toOrderTemplateType(first),
		Category:  strings.TrimSpace(first.Classification),
		Content:   strings.TrimSpace(first.Content),
		Priority:  models.OrderPriorityNormal,
		IsDefault: false,
		IsEnabled: !first.IsDisabled,
		CreatedAt: first.CreateTime,
		UpdatedAt: first.LastModifyTime,
	}
	items := make([]models.OrderTemplateItem, 0, len(rows))
	for idx, row := range rows {
		if row.LastModifyTime.After(template.UpdatedAt) {
			template.UpdatedAt = row.LastModifyTime
		}
		if row.CreateTime.Before(template.CreatedAt) {
			template.CreatedAt = row.CreateTime
		}
		if strings.TrimSpace(template.Content) == "" && strings.TrimSpace(row.Content) != "" {
			template.Content = strings.TrimSpace(row.Content)
		}
		drug := drugMap[row.DrugID]
		var drugID *uint
		if row.DrugID > 0 {
			v := uint(row.DrugID)
			drugID = &v
		}
		var spec *string
		if trimmed := strings.TrimSpace(drug.Specification); trimmed != "" {
			spec = &trimmed
		}
		var dosage *string
		if trimmed := strings.TrimSpace(row.Dosage); trimmed != "" {
			dosage = &trimmed
		}
		var unit *string
		if trimmed := strings.TrimSpace(drug.BasicUnit); trimmed != "" {
			unit = &trimmed
		}
		var route *string
		if trimmed := strings.TrimSpace(row.UseWay); trimmed != "" {
			route = &trimmed
		}
		var frequency *string
		if trimmed := strings.TrimSpace(row.UseMethod); trimmed != "" {
			frequency = &trimmed
		}
		var timing *string
		if trimmed := strings.TrimSpace(row.UseOpportunity); trimmed != "" {
			timing = &trimmed
			if template.Frequency == nil {
				template.Frequency = &trimmed
			}
		}
		var minUnitDose *string
		if trimmed := strings.TrimSpace(row.Dosage); trimmed != "" {
			minUnitDose = &trimmed
		} else if drug.MinUnitDosage > 0 {
			v := strconv.FormatInt(drug.MinUnitDosage, 10)
			minUnitDose = &v
		}
		groupID := ""
		if row.OrderGroup > 0 {
			groupID = strconv.FormatInt(row.OrderGroup, 10)
		}
		items = append(items, models.OrderTemplateItem{
			ID:          strconv.FormatInt(row.ID, 10),
			TemplateID:  template.ID,
			DrugID:      drugID,
			DrugName:    strings.TrimSpace(firstNonEmpty(strings.TrimSpace(drug.Name), strings.TrimSpace(row.Content), strings.TrimSpace(row.Name))),
			Spec:        spec,
			MinUnitDose: minUnitDose,
			Dosage:      dosage,
			Unit:        unit,
			Route:       route,
			Frequency:   frequency,
			Timing:      timing,
			GroupID:     optionalStringPtr(groupID),
			SortOrder:   idx,
		})
	}
	template.Items = items
	return template
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// ===== 药品目录服务 =====

// DrugCatalogService 药品目录服务
type DrugCatalogService struct {
	db *gorm.DB
}

type legacyDrugCatalog struct {
	ID                int64     `gorm:"column:Id"`
	TenantID          int64     `gorm:"column:TenantId"`
	Name              string    `gorm:"column:Name"`
	Classification    string    `gorm:"column:Classification"`
	Code              string    `gorm:"column:Code"`
	Brand             string    `gorm:"column:Brand"`
	Specification     string    `gorm:"column:Specification"`
	Package           string    `gorm:"column:Package"`
	Manufacturer      string    `gorm:"column:Manufacturer"`
	Note              string    `gorm:"column:Note"`
	IsDisabled        bool      `gorm:"column:IsDisabled"`
	CreateTime        time.Time `gorm:"column:CreateTime"`
	Spell             string    `gorm:"column:Spell"`
	BasicUnit         string    `gorm:"column:BasicUnit"`
	SpecificationUnit string    `gorm:"column:SpecificationUnit"`
	Sort              int       `gorm:"column:Sort"`
	StdCat            string    `gorm:"column:StdCat"`
	LastModifyTime    time.Time `gorm:"column:LastModifyTime"`
	ShortName         string    `gorm:"column:ShortName"`
	UseTips           string    `gorm:"column:UseTips"`
	MinUnitDosage     int64     `gorm:"column:MinUnitDosage"`
	UseOpportunity    string    `gorm:"column:UseOpportunity"`
}

func (legacyDrugCatalog) TableName() string { return "Auxiliary_DrugInfomation" }

func toDrugCatalogDTO(item legacyDrugCatalog) models.DrugCatalog {
	var shortName *string
	if trimmed := strings.TrimSpace(item.ShortName); trimmed != "" {
		shortName = &trimmed
	}
	var mnemonic *string
	if trimmed := strings.TrimSpace(item.Spell); trimmed != "" {
		mnemonic = &trimmed
	}
	var packaging *string
	if trimmed := strings.TrimSpace(item.Package); trimmed != "" {
		packaging = &trimmed
	}
	var specUnit *string
	if trimmed := strings.TrimSpace(item.SpecificationUnit); trimmed != "" {
		specUnit = &trimmed
	}
	var timing *string
	if trimmed := strings.TrimSpace(item.UseOpportunity); trimmed != "" {
		timing = &trimmed
	}
	var tips *string
	if trimmed := strings.TrimSpace(item.UseTips); trimmed != "" {
		tips = &trimmed
	}
	var minUnitDose *string
	if item.MinUnitDosage > 0 {
		v := strconv.FormatInt(item.MinUnitDosage, 10)
		minUnitDose = &v
	}

	return models.DrugCatalog{
		ID:           uint(item.ID),
		Code:         strings.TrimSpace(item.Code),
		Name:         strings.TrimSpace(item.Name),
		ShortName:    shortName,
		Mnemonic:     mnemonic,
		Category:     strings.TrimSpace(item.Classification),
		Spec:         strings.TrimSpace(item.Specification),
		SpecUnit:     specUnit,
		MinUnitDose:  minUnitDose,
		BaseUnit:     strings.TrimSpace(item.BasicUnit),
		Brand:        optionalStringPtr(strings.TrimSpace(item.Brand)),
		Packaging:    packaging,
		Manufacturer: strings.TrimSpace(item.Manufacturer),
		Timing:       timing,
		Tips:         tips,
		SortOrder:    item.Sort,
		IsEnabled:    !item.IsDisabled,
		TenantID:     &item.TenantID,
		Note:         strings.TrimSpace(item.Note),
		CreatedAt:    item.CreateTime,
		UpdatedAt:    item.LastModifyTime,
	}
}

// NewDrugCatalogService 创建药品目录服务
func NewDrugCatalogService() *DrugCatalogService {
	return &DrugCatalogService{
		db: database.GetDB(),
	}
}

// DrugCatalogListRequest 获取药品目录列表请求
type DrugCatalogListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Search    string `form:"search"`
	Category  string `form:"category"`
	IsEnabled *bool  `form:"isEnabled"`
}

// DrugCatalogListResponse 获取药品目录列表响应
type DrugCatalogListResponse struct {
	Items     []models.DrugCatalog `json:"items"`
	Total     int64                `json:"total"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"pageSize"`
	TotalPage int                  `json:"totalPage"`
}

// List 获取药品目录列表
func (s *DrugCatalogService) List(req DrugCatalogListRequest) (*DrugCatalogListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.DrugCatalog{})

	// 筛选条件
	if req.Search != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR generic_name LIKE ?",
			"%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.DrugCatalog
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("category, code, created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &DrugCatalogListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取药品目录详情
func (s *DrugCatalogService) Get(id string) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var catalog models.DrugCatalog
	err := s.db.First(&catalog, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drug catalog not found")
		}
		return nil, err
	}

	return &catalog, nil
}

// DrugCatalogCreateRequest 创建药品目录请求
type DrugCatalogCreateRequest struct {
	Code         string `json:"code"`
	Name         string `json:"name" binding:"required"`
	ShortName    string `json:"shortName"`
	Mnemonic     string `json:"mnemonic"`
	GenericName  string `json:"genericName"`
	Category     string `json:"category" binding:"required"`
	Spec         string `json:"spec"`
	Concentration string `json:"concentration"`
	SpecUnit     string `json:"specUnit"`
	MinUnitDose  string `json:"minUnitDose"`
	BaseUnit     string `json:"baseUnit"`
	Brand        string `json:"brand"`
	Packaging    string `json:"packaging"`
	Manufacturer string `json:"manufacturer"`
	StandardType string `json:"standardType"`
	Timing       string `json:"timing"`
	Tips         string `json:"tips"`
	SortOrder    int    `json:"sortOrder"`
	IsEnabled    bool   `json:"isEnabled"`
	Note         string `json:"note"`
}

func ensureDrugCatalogCode(code string) string {
	trimmed := strings.TrimSpace(code)
	if trimmed != "" {
		return trimmed
	}
	return "AUTO-" + uuid.NewString()
}

// Create 创建药品目录
func (s *DrugCatalogService) Create(req DrugCatalogCreateRequest) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	code := ensureDrugCatalogCode(req.Code)

	// 检查代码是否已存在
	var count int64
	if err := s.db.Model(&models.DrugCatalog{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("drug code already exists")
	}

	catalog := models.DrugCatalog{
		Code:         code,
		Name:         strings.TrimSpace(req.Name),
		ShortName:    optionalStringPtr(req.ShortName),
		Mnemonic:     optionalStringPtr(req.Mnemonic),
		GenericName:  strings.TrimSpace(req.GenericName),
		Category:     strings.TrimSpace(req.Category),
		Spec:         strings.TrimSpace(req.Spec),
		Concentration: optionalStringPtr(req.Concentration),
		SpecUnit:     optionalStringPtr(req.SpecUnit),
		MinUnitDose:  optionalStringPtr(req.MinUnitDose),
		BaseUnit:     strings.TrimSpace(req.BaseUnit),
		Brand:        optionalStringPtr(req.Brand),
		Packaging:    optionalStringPtr(req.Packaging),
		Manufacturer: strings.TrimSpace(req.Manufacturer),
		StandardType: optionalStringPtr(req.StandardType),
		Timing:       optionalStringPtr(req.Timing),
		Tips:         optionalStringPtr(req.Tips),
		SortOrder:    req.SortOrder,
		IsEnabled:    req.IsEnabled,
		Note:         strings.TrimSpace(req.Note),
	}

	if err := s.db.Create(&catalog).Error; err != nil {
		return nil, err
	}

	return &catalog, nil
}

// DrugCatalogUpdateRequest 更新药品目录请求
type DrugCatalogUpdateRequest struct {
	Name         *string `json:"name"`
	ShortName    *string `json:"shortName"`
	Mnemonic     *string `json:"mnemonic"`
	GenericName  *string `json:"genericName"`
	Category     *string `json:"category"`
	Spec         *string `json:"spec"`
	Concentration *string `json:"concentration"`
	SpecUnit     *string `json:"specUnit"`
	MinUnitDose  *string `json:"minUnitDose"`
	BaseUnit     *string `json:"baseUnit"`
	Brand        *string `json:"brand"`
	Packaging    *string `json:"packaging"`
	Manufacturer *string `json:"manufacturer"`
	StandardType *string `json:"standardType"`
	Timing       *string `json:"timing"`
	Tips         *string `json:"tips"`
	SortOrder    *int    `json:"sortOrder"`
	IsEnabled    *bool   `json:"isEnabled"`
	Note         *string `json:"note"`
}

// Update 更新药品目录
func (s *DrugCatalogService) Update(id uint, req DrugCatalogUpdateRequest) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var catalog models.DrugCatalog
	if err := s.db.First(&catalog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drug catalog not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.ShortName != nil {
		updates["short_name"] = strings.TrimSpace(*req.ShortName)
	}
	if req.Mnemonic != nil {
		updates["mnemonic"] = strings.TrimSpace(*req.Mnemonic)
	}
	if req.GenericName != nil {
		updates["generic_name"] = strings.TrimSpace(*req.GenericName)
	}
	if req.Category != nil {
		updates["category"] = strings.TrimSpace(*req.Category)
	}
	if req.Spec != nil {
		updates["spec"] = strings.TrimSpace(*req.Spec)
	}
	if req.Concentration != nil {
		updates["concentration"] = strings.TrimSpace(*req.Concentration)
	}
	if req.SpecUnit != nil {
		updates["spec_unit"] = strings.TrimSpace(*req.SpecUnit)
	}
	if req.MinUnitDose != nil {
		updates["min_unit_dose"] = strings.TrimSpace(*req.MinUnitDose)
	}
	if req.BaseUnit != nil {
		updates["unit"] = strings.TrimSpace(*req.BaseUnit)
	}
	if req.Brand != nil {
		updates["brand"] = strings.TrimSpace(*req.Brand)
	}
	if req.Packaging != nil {
		updates["packaging"] = strings.TrimSpace(*req.Packaging)
	}
	if req.Manufacturer != nil {
		updates["manufacturer"] = strings.TrimSpace(*req.Manufacturer)
	}
	if req.StandardType != nil {
		updates["standard_type"] = strings.TrimSpace(*req.StandardType)
	}
	if req.Timing != nil {
		updates["timing"] = strings.TrimSpace(*req.Timing)
	}
	if req.Tips != nil {
		updates["tips"] = strings.TrimSpace(*req.Tips)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.Note != nil {
		updates["notes"] = strings.TrimSpace(*req.Note)
	}

	if err := s.db.Model(&catalog).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.First(&catalog, id).Error; err != nil {
		return nil, err
	}

	return &catalog, nil
}

// Delete 删除药品目录
func (s *DrugCatalogService) Delete(id uint) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.DrugCatalog{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("drug catalog not found")
	}

	return nil
}

// ToggleEnabled 切换启用状态
func (s *DrugCatalogService) ToggleEnabled(id uint) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var catalog models.DrugCatalog
	if err := s.db.First(&catalog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("drug catalog not found")
		}
		return false, err
	}

	newStatus := !catalog.IsEnabled
	if err := s.db.Model(&catalog).Update("is_enabled", newStatus).Error; err != nil {
		return false, err
	}

	return newStatus, nil
}

// GetCategories 获取药品分类列表
func (s *DrugCatalogService) GetCategories() ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var categories []string
	if err := s.db.Model(&models.DrugCatalog{}).
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *DrugCatalogService) LegacyList(req DrugCatalogListRequest) (*DrugCatalogListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&legacyDrugCatalog{}).Where(`"TenantId" = ?`, LegacyTenantID)
	if req.Search != "" {
		like := "%" + req.Search + "%"
		query = query.Where(`("Name" LIKE ? OR "Code" LIKE ? OR "Spell" LIKE ?)`, like, like, like)
	}
	if req.Category != "" {
		query = query.Where(`"Classification" = ?`, req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where(`COALESCE("IsDisabled", false) = ?`, !*req.IsEnabled)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []legacyDrugCatalog
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order(`"Classification" ASC`).
		Order(`"Sort" ASC`).
		Order(`"Code" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]models.DrugCatalog, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDrugCatalogDTO(row))
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &DrugCatalogListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

func (s *DrugCatalogService) LegacyGet(id string) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var row legacyDrugCatalog
	if err := s.db.Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drug catalog not found")
		}
		return nil, err
	}

	dto := toDrugCatalogDTO(row)
	return &dto, nil
}

func (s *DrugCatalogService) LegacyGetCategories() ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var categories []string
	if err := s.db.Model(&legacyDrugCatalog{}).
		Where(`"TenantId" = ? AND COALESCE("IsDisabled", false) = false`, LegacyTenantID).
		Distinct(`"Classification"`).
		Pluck(`"Classification"`, &categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// ===== 医嘱模板服务 =====

// OrderTemplateService 医嘱模板服务
func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func legacyMinUnitDosage(value string) int64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}
	if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return i
	}
	return int64(math.Round(parseStringFloat(trimmed)))
}

func (s *DrugCatalogService) LegacyCreate(req DrugCatalogCreateRequest) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	code := ensureDrugCatalogCode(req.Code)
	var count int64
	if err := s.db.Model(&legacyDrugCatalog{}).
		Where(`"TenantId" = ? AND "Code" = ?`, LegacyTenantID, code).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("drug code already exists")
	}

	var created legacyDrugCatalog
	err := s.db.Transaction(func(tx *gorm.DB) error {
		nextID, err := nextLegacyNumericID(tx, `public."Auxiliary_DrugInfomation"`)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		created = legacyDrugCatalog{
			ID:                nextID,
			TenantID:          LegacyTenantID,
			Name:              strings.TrimSpace(req.Name),
			Classification:    strings.TrimSpace(req.Category),
			Code:              code,
			Brand:             strings.TrimSpace(req.Brand),
			Specification:     strings.TrimSpace(req.Spec),
			Package:           strings.TrimSpace(req.Packaging),
			Manufacturer:      strings.TrimSpace(req.Manufacturer),
			Note:              strings.TrimSpace(req.Note),
			IsDisabled:        !req.IsEnabled,
			CreateTime:        now,
			Spell:             strings.TrimSpace(req.Mnemonic),
			BasicUnit:         strings.TrimSpace(req.BaseUnit),
			SpecificationUnit: strings.TrimSpace(req.SpecUnit),
			Sort:              req.SortOrder,
			StdCat:            strings.TrimSpace(req.StandardType),
			LastModifyTime:    now,
			ShortName:         strings.TrimSpace(req.ShortName),
			UseTips:           strings.TrimSpace(req.Tips),
			MinUnitDosage:     legacyMinUnitDosage(req.MinUnitDose),
			UseOpportunity:    strings.TrimSpace(req.Timing),
		}
		return tx.Table(created.TableName()).Create(&created).Error
	})
	if err != nil {
		return nil, err
	}

	dto := toDrugCatalogDTO(created)
	return &dto, nil
}

func (s *DrugCatalogService) LegacyUpdate(id uint, req DrugCatalogUpdateRequest) (*models.DrugCatalog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var current legacyDrugCatalog
	if err := s.db.Where(`"TenantId" = ? AND "Id" = ?`, LegacyTenantID, id).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drug catalog not found")
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"LastModifyTime": time.Now().UTC(),
	}
	if req.Name != nil {
		updates["Name"] = strings.TrimSpace(*req.Name)
	}
	if req.ShortName != nil {
		updates["ShortName"] = strings.TrimSpace(*req.ShortName)
	}
	if req.Mnemonic != nil {
		updates["Spell"] = strings.TrimSpace(*req.Mnemonic)
	}
	if req.Category != nil {
		updates["Classification"] = strings.TrimSpace(*req.Category)
	}
	if req.Spec != nil {
		updates["Specification"] = strings.TrimSpace(*req.Spec)
	}
	if req.SpecUnit != nil {
		updates["SpecificationUnit"] = strings.TrimSpace(*req.SpecUnit)
	}
	if req.MinUnitDose != nil {
		updates["MinUnitDosage"] = legacyMinUnitDosage(*req.MinUnitDose)
	}
	if req.BaseUnit != nil {
		updates["BasicUnit"] = strings.TrimSpace(*req.BaseUnit)
	}
	if req.Brand != nil {
		updates["Brand"] = strings.TrimSpace(*req.Brand)
	}
	if req.Packaging != nil {
		updates["Package"] = strings.TrimSpace(*req.Packaging)
	}
	if req.Manufacturer != nil {
		updates["Manufacturer"] = strings.TrimSpace(*req.Manufacturer)
	}
	if req.StandardType != nil {
		updates["StdCat"] = strings.TrimSpace(*req.StandardType)
	}
	if req.Timing != nil {
		updates["UseOpportunity"] = strings.TrimSpace(*req.Timing)
	}
	if req.Tips != nil {
		updates["UseTips"] = strings.TrimSpace(*req.Tips)
	}
	if req.SortOrder != nil {
		updates["Sort"] = *req.SortOrder
	}
	if req.IsEnabled != nil {
		updates["IsDisabled"] = !*req.IsEnabled
	}
	if req.Note != nil {
		updates["Note"] = strings.TrimSpace(*req.Note)
	}

	if err := s.db.Model(&legacyDrugCatalog{}).
		Where(`"TenantId" = ? AND "Id" = ?`, LegacyTenantID, id).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.LegacyGet(strconv.FormatUint(uint64(id), 10))
}

func (s *DrugCatalogService) LegacyDelete(id uint) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Model(&legacyDrugCatalog{}).
		Where(`"TenantId" = ? AND "Id" = ?`, LegacyTenantID, id).
		Updates(map[string]interface{}{
			"IsDisabled":     true,
			"LastModifyTime": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("drug catalog not found")
	}
	return nil
}

func (s *DrugCatalogService) LegacyToggleEnabled(id uint) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var current legacyDrugCatalog
	if err := s.db.Where(`"TenantId" = ? AND "Id" = ?`, LegacyTenantID, id).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("drug catalog not found")
		}
		return false, err
	}

	newEnabled := current.IsDisabled
	if err := s.db.Model(&legacyDrugCatalog{}).
		Where(`"TenantId" = ? AND "Id" = ?`, LegacyTenantID, id).
		Updates(map[string]interface{}{
			"IsDisabled":     !newEnabled,
			"LastModifyTime": time.Now().UTC(),
		}).Error; err != nil {
		return false, err
	}

	return newEnabled, nil
}

type OrderTemplateService struct {
	db *gorm.DB
}

// NewOrderTemplateService 创建医嘱模板服务
func NewOrderTemplateService() *OrderTemplateService {
	return &OrderTemplateService{
		db: database.GetDB(),
	}
}

// OrderTemplateListRequest 获取医嘱模板列表请求
type OrderTemplateListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Search    string `form:"search"`
	Type      string `form:"type"`
	Category  string `form:"category"`
	IsEnabled *bool  `form:"isEnabled"`
}

// OrderTemplateListResponse 获取医嘱模板列表响应
type OrderTemplateListResponse struct {
	Items     []models.OrderTemplate `json:"items"`
	Total     int64                  `json:"total"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"pageSize"`
	TotalPage int                    `json:"totalPage"`
}

// List 获取医嘱模板列表
func (s *OrderTemplateService) List(req OrderTemplateListRequest) (*OrderTemplateListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.OrderTemplate{})

	// 筛选条件
	if req.Search != "" {
		query = query.Where("name LIKE ? OR content LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.OrderTemplate
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("is_default DESC, created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &OrderTemplateListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

func (s *OrderTemplateService) LegacyList(req OrderTemplateListRequest) (*OrderTemplateListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&legacyOrderTemplateRow{}).Where(`"TenantId" = ?`, LegacyTenantID)
	if req.Search != "" {
		like := "%" + strings.TrimSpace(req.Search) + "%"
		query = query.Where(`("Name" LIKE ? OR "Content" LIKE ?)`, like, like)
	}
	if req.Category != "" {
		query = query.Where(`"Classification" = ?`, strings.TrimSpace(req.Category))
	}
	if req.IsEnabled != nil {
		query = query.Where(`COALESCE("IsDisabled", false) = ?`, !*req.IsEnabled)
	}

	var rows []legacyOrderTemplateRow
	if err := query.Order(`"OrderGroup" ASC`).
		Order(`"Name" ASC`).
		Order(`"Id" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	grouped := make(map[int64][]legacyOrderTemplateRow)
	order := make([]int64, 0)
	for _, row := range rows {
		key := row.OrderGroup
		if key <= 0 {
			key = row.ID
		}
		if _, ok := grouped[key]; !ok {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], row)
	}

	drugIDs := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row.DrugID <= 0 {
			continue
		}
		if _, ok := seen[row.DrugID]; ok {
			continue
		}
		seen[row.DrugID] = struct{}{}
		drugIDs = append(drugIDs, row.DrugID)
	}
	drugMap := make(map[int64]legacyDrugCatalog)
	if len(drugIDs) > 0 {
		var drugs []legacyDrugCatalog
		if err := s.db.Where(`"TenantId" = ? AND "Id" IN ?`, LegacyTenantID, drugIDs).Find(&drugs).Error; err != nil {
			return nil, err
		}
		for _, drug := range drugs {
			drugMap[drug.ID] = drug
		}
	}

	templates := make([]models.OrderTemplate, 0, len(grouped))
	for _, key := range order {
		template := toOrderTemplateDTO(grouped[key], drugMap)
		if req.Type != "" && template.Type != strings.TrimSpace(req.Type) {
			continue
		}
		templates = append(templates, template)
	}

	total := int64(len(templates))
	start := (req.Page - 1) * req.PageSize
	if start > len(templates) {
		start = len(templates)
	}
	end := start + req.PageSize
	if end > len(templates) {
		end = len(templates)
	}
	pageItems := templates[start:end]
	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}
	return &OrderTemplateListResponse{Items: pageItems, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage}, nil
}

// Get 获取医嘱模板详情（含 items）
func (s *OrderTemplateService) Get(id string) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var template models.OrderTemplate
	err := s.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&template, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order template not found")
		}
		return nil, err
	}

	return &template, nil
}

func (s *OrderTemplateService) LegacyGet(id string) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	key, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil {
		return nil, errors.New("order template not found")
	}

	var rows []legacyOrderTemplateRow
	if err := s.db.Where(`"TenantId" = ? AND ("OrderGroup" = ? OR ("OrderGroup" IS NULL AND "Id" = ?) OR ("OrderGroup" = 0 AND "Id" = ?))`, LegacyTenantID, key, key, key).
		Order(`"Id" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("order template not found")
	}

	drugIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.DrugID > 0 {
			drugIDs = append(drugIDs, row.DrugID)
		}
	}
	drugMap := make(map[int64]legacyDrugCatalog)
	if len(drugIDs) > 0 {
		var drugs []legacyDrugCatalog
		if err := s.db.Where(`"TenantId" = ? AND "Id" IN ?`, LegacyTenantID, drugIDs).Find(&drugs).Error; err != nil {
			return nil, err
		}
		for _, drug := range drugs {
			drugMap[drug.ID] = drug
		}
	}

	dto := toOrderTemplateDTO(rows, drugMap)
	return &dto, nil
}

// OrderTemplateItemRequest 医嘱模板条目请求
type OrderTemplateItemRequest struct {
	DrugID      *uint   `json:"drugId"`
	DrugName    string  `json:"drugName" binding:"required"`
	Spec        *string `json:"spec"`
	MinUnitDose *string `json:"minUnitDose"`
	Dosage      *string `json:"dosage"`
	Unit        *string `json:"unit"`
	Route       *string `json:"route"`
	Frequency   *string `json:"frequency"`
	Timing      *string `json:"timing"`
	GroupID     *string `json:"groupId"`
	SortOrder   int     `json:"sortOrder"`
}

// OrderTemplateCreateRequest 创建医嘱模板请求
type OrderTemplateCreateRequest struct {
	Name      string                     `json:"name" binding:"required"`
	Type      string                     `json:"type" binding:"required,oneof=长期 临时"`
	Category  string                     `json:"category" binding:"required"`
	Content   string                     `json:"content" binding:"required"`
	Frequency *string                    `json:"frequency"`
	Priority  string                     `json:"priority"`
	IsDefault bool                       `json:"isDefault"`
	IsEnabled bool                       `json:"isEnabled"`
	Items     []OrderTemplateItemRequest `json:"items"`
}

// Create 创建医嘱模板
func (s *OrderTemplateService) Create(req OrderTemplateCreateRequest) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 如果设置为默认模板，先取消其他默认模板
	if req.IsDefault {
		s.db.Model(&models.OrderTemplate{}).
			Where("category = ? AND is_default = ?", req.Category, true).
			Update("is_default", false)
	}

	// 设置默认优先级
	if req.Priority == "" {
		req.Priority = models.OrderPriorityNormal
	}

	template := models.OrderTemplate{
		ID:        utils.GenerateID(),
		Name:      req.Name,
		Type:      req.Type,
		Category:  req.Category,
		Content:   req.Content,
		Frequency: req.Frequency,
		Priority:  req.Priority,
		IsDefault: req.IsDefault,
		IsEnabled: req.IsEnabled,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&template).Error; err != nil {
			return err
		}
		// 批量创建 items
		for _, itemReq := range req.Items {
			item := models.OrderTemplateItem{
				TemplateID:  template.ID,
				DrugID:      itemReq.DrugID,
				DrugName:    itemReq.DrugName,
				Spec:        itemReq.Spec,
				MinUnitDose: itemReq.MinUnitDose,
				Dosage:      itemReq.Dosage,
				Unit:        itemReq.Unit,
				Route:       itemReq.Route,
				Frequency:   itemReq.Frequency,
				Timing:      itemReq.Timing,
				GroupID:     itemReq.GroupID,
				SortOrder:   itemReq.SortOrder,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 重新加载含 items
	return s.Get(template.ID)
}

// OrderTemplateUpdateRequest 更新医嘱模板请求
type OrderTemplateUpdateRequest struct {
	Name      *string                     `json:"name"`
	Type      *string                     `json:"type"`
	Category  *string                     `json:"category"`
	Content   *string                     `json:"content"`
	Frequency *string                     `json:"frequency"`
	Priority  *string                     `json:"priority"`
	IsDefault *bool                       `json:"isDefault"`
	IsEnabled *bool                       `json:"isEnabled"`
	Items     *[]OrderTemplateItemRequest `json:"items"`
}

// Update 更新医嘱模板
func (s *OrderTemplateService) Update(id string, req OrderTemplateUpdateRequest) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var template models.OrderTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order template not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Frequency != nil {
		updates["frequency"] = *req.Frequency
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.IsDefault != nil {
		// 如果设置为默认模板，先取消同分类的其他默认模板
		if *req.IsDefault {
			s.db.Model(&models.OrderTemplate{}).
				Where("category = ? AND is_default = ? AND id != ?", template.Category, true, id).
				Update("is_default", false)
		}
		updates["is_default"] = *req.IsDefault
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 更新模板字段
		if len(updates) > 0 {
			if err := tx.Model(&template).Updates(updates).Error; err != nil {
				return err
			}
		}
		// 如果传了 items，全量替换
		if req.Items != nil {
			// 删除旧 items
			if err := tx.Where("template_id = ?", id).Delete(&models.OrderTemplateItem{}).Error; err != nil {
				return err
			}
			// 创建新 items
			for _, itemReq := range *req.Items {
				item := models.OrderTemplateItem{
					TemplateID:  id,
					DrugID:      itemReq.DrugID,
					DrugName:    itemReq.DrugName,
					Spec:        itemReq.Spec,
					MinUnitDose: itemReq.MinUnitDose,
					Dosage:      itemReq.Dosage,
					Unit:        itemReq.Unit,
					Route:       itemReq.Route,
					Frequency:   itemReq.Frequency,
					Timing:      itemReq.Timing,
					GroupID:     itemReq.GroupID,
					SortOrder:   itemReq.SortOrder,
				}
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 重新加载含 items
	return s.Get(id)
}

// Delete 删除医嘱模板（级联删除 items）
func (s *OrderTemplateService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除关联的 items
		if err := tx.Where("template_id = ?", id).Delete(&models.OrderTemplateItem{}).Error; err != nil {
			return err
		}
		// 再删除模板
		result := tx.Delete(&models.OrderTemplate{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("order template not found")
		}
		return nil
	})
	return err
}

// ToggleEnabled 切换启用状态
func (s *OrderTemplateService) ToggleEnabled(id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	var template models.OrderTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("order template not found")
		}
		return false, err
	}

	newStatus := !template.IsEnabled
	if err := s.db.Model(&template).Update("is_enabled", newStatus).Error; err != nil {
		return false, err
	}

	return newStatus, nil
}

// SetDefault 设置默认模板
func (s *OrderTemplateService) SetDefault(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	var template models.OrderTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order template not found")
		}
		return err
	}

	// 取消同分类的其他默认模板
	s.db.Model(&models.OrderTemplate{}).
		Where("category = ? AND is_default = ?", template.Category, true).
		Update("is_default", false)

	// 设置当前模板为默认
	if err := s.db.Model(&template).Update("is_default", true).Error; err != nil {
		return err
	}

	return nil
}

func buildLegacyOrderTemplateNote(orderType, content string) string {
	parts := make([]string, 0, 2)
	if trimmed := strings.TrimSpace(orderType); trimmed != "" {
		parts = append(parts, "type="+trimmed)
	}
	if trimmed := strings.TrimSpace(content); trimmed != "" {
		parts = append(parts, "content="+trimmed)
	}
	return strings.Join(parts, "; ")
}

func resolveLegacyOrderTemplateGroupKey(id int64, row legacyOrderTemplateRow) int64 {
	if row.OrderGroup > 0 {
		return row.OrderGroup
	}
	if id > 0 {
		return id
	}
	return row.ID
}

func (s *OrderTemplateService) replaceLegacyOrderTemplateRows(tx *gorm.DB, groupKey int64, req OrderTemplateCreateRequest) error {
	deleteQuery := tx.Where(`"TenantId" = ? AND ("OrderGroup" = ? OR ("OrderGroup" IS NULL AND "Id" = ?) OR ("OrderGroup" = 0 AND "Id" = ?))`, LegacyTenantID, groupKey, groupKey, groupKey)
	if err := deleteQuery.Delete(&legacyOrderTemplateRow{}).Error; err != nil {
		return err
	}

	rows := req.Items
	if len(rows) == 0 {
		rows = []OrderTemplateItemRequest{{
			DrugName:  strings.TrimSpace(req.Content),
			SortOrder: 1,
		}}
	}

	now := time.Now().UTC()
	for idx, item := range rows {
		nextID, err := nextLegacyNumericID(tx, `public."Order_OrderTPL"`)
		if err != nil {
			return err
		}

		var drugID int64
		if item.DrugID != nil && *item.DrugID > 0 {
			drugID = int64(*item.DrugID)
		} else if name := strings.TrimSpace(item.DrugName); name != "" {
			if lookedUpID, err := (&PatientService{db: tx}).findLegacyDrugIDByName(name); err == nil {
				drugID = lookedUpID
			}
		}

		content := firstNonEmpty(strings.TrimSpace(item.DrugName), strings.TrimSpace(req.Content), strings.TrimSpace(req.Name))
		name := firstNonEmpty(strings.TrimSpace(req.Name), content)
		sortOrder := item.SortOrder
		if sortOrder <= 0 {
			sortOrder = idx + 1
		}

		row := legacyOrderTemplateRow{
			ID:             nextID,
			TenantID:       LegacyTenantID,
			Name:           name,
			Classification: strings.TrimSpace(req.Category),
			OrderGroup:     groupKey,
			DrugID:         drugID,
			Content:        content,
			Dosage:         firstNonEmpty(derefString(item.Dosage), derefString(item.MinUnitDose)),
			UseOpportunity: derefString(item.Timing),
			UseMethod:      derefString(item.Frequency),
			UseWay:         derefString(item.Route),
			UseNum:         float64(sortOrder),
			IsDisabled:     !req.IsEnabled,
			CreatorID:      0,
			CreateTime:     now,
			LastModifyTime: now,
			Note:           buildLegacyOrderTemplateNote(req.Type, req.Content),
		}
		if err := tx.Table(row.TableName()).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *OrderTemplateService) LegacyCreate(req OrderTemplateCreateRequest) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	groupKey, err := nextLegacyNumericID(s.db, `public."Order_OrderTPL"`)
	if err != nil {
		return nil, err
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		return s.replaceLegacyOrderTemplateRows(tx, groupKey, req)
	}); err != nil {
		return nil, err
	}
	return s.LegacyGet(strconv.FormatInt(groupKey, 10))
}

func (s *OrderTemplateService) LegacyUpdate(id string, req OrderTemplateUpdateRequest) (*models.OrderTemplate, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	current, err := s.LegacyGet(id)
	if err != nil {
		return nil, err
	}

	key, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil {
		return nil, errors.New("order template not found")
	}

	createReq := OrderTemplateCreateRequest{
		Name:      current.Name,
		Type:      current.Type,
		Category:  current.Category,
		Content:   current.Content,
		Frequency: current.Frequency,
		Priority:  current.Priority,
		IsDefault: current.IsDefault,
		IsEnabled: current.IsEnabled,
		Items:     make([]OrderTemplateItemRequest, 0, len(current.Items)),
	}
	if req.Name != nil {
		createReq.Name = strings.TrimSpace(*req.Name)
	}
	if req.Type != nil {
		createReq.Type = strings.TrimSpace(*req.Type)
	}
	if req.Category != nil {
		createReq.Category = strings.TrimSpace(*req.Category)
	}
	if req.Content != nil {
		createReq.Content = strings.TrimSpace(*req.Content)
	}
	if req.Frequency != nil {
		createReq.Frequency = req.Frequency
	}
	if req.Priority != nil {
		createReq.Priority = strings.TrimSpace(*req.Priority)
	}
	if req.IsDefault != nil {
		createReq.IsDefault = *req.IsDefault
	}
	if req.IsEnabled != nil {
		createReq.IsEnabled = *req.IsEnabled
	}

	if req.Items != nil {
		createReq.Items = append(createReq.Items, (*req.Items)...)
	} else {
		for _, item := range current.Items {
			createReq.Items = append(createReq.Items, OrderTemplateItemRequest{
				DrugID:      item.DrugID,
				DrugName:    item.DrugName,
				Spec:        item.Spec,
				MinUnitDose: item.MinUnitDose,
				Dosage:      item.Dosage,
				Unit:        item.Unit,
				Route:       item.Route,
				Frequency:   item.Frequency,
				Timing:      item.Timing,
				GroupID:     item.GroupID,
				SortOrder:   item.SortOrder,
			})
		}
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		return s.replaceLegacyOrderTemplateRows(tx, key, createReq)
	}); err != nil {
		return nil, err
	}
	return s.LegacyGet(id)
}

func (s *OrderTemplateService) LegacyDelete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	key, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil {
		return errors.New("order template not found")
	}

	result := s.db.Model(&legacyOrderTemplateRow{}).
		Where(`"TenantId" = ? AND ("OrderGroup" = ? OR ("OrderGroup" IS NULL AND "Id" = ?) OR ("OrderGroup" = 0 AND "Id" = ?))`, LegacyTenantID, key, key, key).
		Updates(map[string]interface{}{
			"IsDisabled":     true,
			"LastModifyTime": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("order template not found")
	}
	return nil
}

func (s *OrderTemplateService) LegacyToggleEnabled(id string) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	template, err := s.LegacyGet(id)
	if err != nil {
		return false, err
	}
	key, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil {
		return false, errors.New("order template not found")
	}

	newEnabled := !template.IsEnabled
	if err := s.db.Model(&legacyOrderTemplateRow{}).
		Where(`"TenantId" = ? AND ("OrderGroup" = ? OR ("OrderGroup" IS NULL AND "Id" = ?) OR ("OrderGroup" = 0 AND "Id" = ?))`, LegacyTenantID, key, key, key).
		Updates(map[string]interface{}{
			"IsDisabled":     !newEnabled,
			"LastModifyTime": time.Now().UTC(),
		}).Error; err != nil {
		return false, err
	}
	return newEnabled, nil
}

func (s *OrderTemplateService) LegacySetDefault(id string) error {
	if _, err := s.LegacyGet(id); err != nil {
		return err
	}
	return errors.New("legacy order template does not support default flag")
}

func (s *PlanTemplateService) LegacySetDefault(id string) error {
	if _, err := s.LegacyGet(id); err != nil {
		return err
	}
	return errors.New("legacy plan template does not support default flag")
}
