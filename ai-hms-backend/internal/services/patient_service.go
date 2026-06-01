package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/elliotxin/ai-hms-backend/internal/utils/idgen"
	"gorm.io/gorm"
)

// PatientService 患者服务
type PatientService struct {
	db *gorm.DB
}

type legacyPatientPlan struct {
	ID                         int64               `gorm:"column:Id"`
	TenantID                   int64               `gorm:"column:TenantId"`
	PatientID                  modeltypes.LegacyID `gorm:"column:PatientId"`
	Name                       string              `gorm:"column:Name"`
	CreatorID                  int64               `gorm:"column:CreatorId"`
	CreateTime                 time.Time           `gorm:"column:CreateTime"`
	PlanTemplateID             int64               `gorm:"column:PlanTPLId"`
	OddWeekFrequency           int                 `gorm:"column:OddWeekFrequency"`
	EvenWeekFrequency          int                 `gorm:"column:EvenWeekFrequency"`
	DialysisMethod             string              `gorm:"column:DialysisMethod"`
	DialysisDuration           float64             `gorm:"column:DialysisDuration"`
	DryWeight                  float64             `gorm:"column:DryWeight"`
	ExtraWeight                float64             `gorm:"column:ExtraWeight"`
	AdjustQuantity             float64             `gorm:"column:AdjustQuantity"`
	BF                         float64             `gorm:"column:BF"`
	BV                         float64             `gorm:"column:BV"`
	FirstAnticoagulant         int64               `gorm:"column:FirstAnticoagulant"`
	FirstDosage                float64             `gorm:"column:FirstDosage"`
	MaintainAnticoagulant      int64               `gorm:"column:MaintainAnticoagulant"`
	DilutionProportion         float64             `gorm:"column:DilutionProportion"`
	InjectionRate              float64             `gorm:"column:InjectionRate"`
	InjectionDuration          float64             `gorm:"column:InjectionDuration"`
	InjectionVolume            float64             `gorm:"column:InjectionVolume"`
	VascularAccessID           int64               `gorm:"column:VascularAccessId"`
	Dialysate                  string              `gorm:"column:Dialysate"`
	DialysateFlow              float64             `gorm:"column:DialysateFlow"`
	DialysateVolume            float64             `gorm:"column:DialysateVolume"`
	NaIonCon                   float64             `gorm:"column:NaIonCon"`
	CaIonCon                   float64             `gorm:"column:CaIonCon"`
	KIonCon                    float64             `gorm:"column:KIonCon"`
	HCO3IonCon                 float64             `gorm:"column:HCO3IonCon"`
	Conductivity               float64             `gorm:"column:Conductivity"`
	DialysateTmp               float64             `gorm:"column:DialysateTmp"`
	SubstituateVolume          float64             `gorm:"column:SubstituateVolume"`
	DilutionMnt                string              `gorm:"column:DilutionMnt"`
	IsDisabled                 bool                `gorm:"column:IsDisabled"`
	LastModifyTime             time.Time           `gorm:"column:LastModifyTime"`
	SalineQuantity             float64             `gorm:"column:SalineQuantity"`
	SealQuantity               float64             `gorm:"column:SealQuantity"`
	ArterialQuantity           float64             `gorm:"column:ArterialQuantity"`
	VenousQuantity             float64             `gorm:"column:VenousQuantity"`
	SealType                   string              `gorm:"column:SealType"`
	Frequency                  string              `gorm:"column:Frequency"`
	GlucoseCon                 float64             `gorm:"column:GlucoseCon"`
	DialysateGroupID           int64               `gorm:"column:DialysateGroupId"`
	AutoConfirmPrescriptionRaw string              `gorm:"column:AutoConfirmPrescription"`
	Note                       string              `gorm:"column:Note"`
	SubstituateFlow            float64             `gorm:"column:SubstituateFlow"`
}

func (legacyPatientPlan) TableName() string { return "Plan_PatientPlan" }

type legacyPatientPlanMaterial struct {
	ID             int64     `gorm:"column:Id"`
	TenantID       int64     `gorm:"column:TenantId"`
	PatientPlanID  int64     `gorm:"column:PatientPlanId"`
	MaterialID     int64     `gorm:"column:MaterialId"`
	MaterialGroup  int64     `gorm:"column:MaterialGroup"`
	Num            float64   `gorm:"column:Num"`
	Note           string    `gorm:"column:Note"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime"`
}

func (legacyPatientPlanMaterial) TableName() string { return "Plan_PatientPlanMaterial" }

type legacyMaterialCatalog struct {
	ID             int64     `gorm:"column:Id"`
	Name           string    `gorm:"column:Name"`
	Classification string    `gorm:"column:Classification"`
	Code           string    `gorm:"column:Code"`
	Brand          string    `gorm:"column:Brand"`
	Specification  string    `gorm:"column:Specification"`
	Note           string    `gorm:"column:Note"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime"`
}

func (legacyMaterialCatalog) TableName() string { return "Auxiliary_MaterialInfomation" }

type legacyPatientPrescription struct {
	ID                           int64               `gorm:"column:Id"`
	TenantID                     int64               `gorm:"column:TenantId"`
	PatientID                    modeltypes.LegacyID `gorm:"column:PatientId"`
	TreatmentID                  int64               `gorm:"column:TreatmentId"`
	PatientPlanID                int64               `gorm:"column:PatientPlanId"`
	CreatorID                    int64               `gorm:"column:CreatorId"`
	CreateTime                   time.Time           `gorm:"column:CreateTime"`
	ConfirmUserID                int64               `gorm:"column:ConfirmUserId"`
	ConfirmTime                  *time.Time          `gorm:"column:ConfirmTime"`
	Status                       int                 `gorm:"column:Status"`
	CaseStatus                   string              `gorm:"column:CaseStatus"`
	DialysisMethod               string              `gorm:"column:DialysisMethod"`
	DialysisDuration             float64             `gorm:"column:DialysisDuration"`
	DryWeight                    float64             `gorm:"column:DryWeight"`
	AdjustQuantity               float64             `gorm:"column:AdjustQuantity"`
	BF                           float64             `gorm:"column:BF"`
	BV                           float64             `gorm:"column:BV"`
	FirstAnticoagulant           int64               `gorm:"column:FirstAnticoagulant"`
	FirstDosage                  float64             `gorm:"column:FirstDosage"`
	MaintainAnticoagulant        int64               `gorm:"column:MaintainAnticoagulant"`
	DilutionProportion           float64             `gorm:"column:DilutionProportion"`
	InjectionRate                float64             `gorm:"column:InjectionRate"`
	InjectionDuration            float64             `gorm:"column:InjectionDuration"`
	InjectionVolume              float64             `gorm:"column:InjectionVolume"`
	VascularAccessID             int64               `gorm:"column:VascularAccessId"`
	Dialysate                    string              `gorm:"column:Dialysate"`
	DialysateFlow                float64             `gorm:"column:DialysateFlow"`
	DialysateVolume              float64             `gorm:"column:DialysateVolume"`
	NaIonCon                     float64             `gorm:"column:NaIonCon"`
	CaIonCon                     float64             `gorm:"column:CaIonCon"`
	KIonCon                      float64             `gorm:"column:KIonCon"`
	HCO3IonCon                   float64             `gorm:"column:HCO3IonCon"`
	Conductivity                 float64             `gorm:"column:Conductivity"`
	DialysateTmp                 float64             `gorm:"column:DialysateTmp"`
	SubstituateVolume            float64             `gorm:"column:SubstituateVolume"`
	DilutionMnt                  string              `gorm:"column:DilutionMnt"`
	LastModifyTime               time.Time           `gorm:"column:LastModifyTime"`
	SalineQuantity               float64             `gorm:"column:SalineQuantity"`
	SealQuantity                 float64             `gorm:"column:SealQuantity"`
	ArterialQuantity             float64             `gorm:"column:ArterialQuantity"`
	VenousQuantity               float64             `gorm:"column:VenousQuantity"`
	UFQuantity                   float64             `gorm:"column:UFQuantity"`
	SealType                     string              `gorm:"column:SealType"`
	GlucoseCon                   float64             `gorm:"column:GlucoseCon"`
	DialysateGroupID             int64               `gorm:"column:DialysateGroupId"`
	Note                         string              `gorm:"column:Note"`
	SubstituateFlow              float64             `gorm:"column:SubstituateFlow"`
	IsInduceDialysisPrescription bool                `gorm:"column:IsInduceDialysisPrescription"`
	HeparinType                  int                 `gorm:"column:HeparinType"`
}

func (legacyPatientPrescription) TableName() string { return "Plan_PatientPrescription" }

type legacyAdjustmentRecord struct {
	ID                        int64     `gorm:"column:Id"`
	TenantID                  int64     `gorm:"column:TenantId"`
	PatientPlanPrescriptionID int64     `gorm:"column:PatientPlanPrescriptionId"`
	AdjustUserID              int64     `gorm:"column:AdjustUserId"`
	AdjustTime                time.Time `gorm:"column:AdjustTime"`
	AdjustReason              string    `gorm:"column:AdjustReason"`
	CreateTime                time.Time `gorm:"column:CreateTime"`
}

func (legacyAdjustmentRecord) TableName() string { return "Plan_PatientPlanPrescriptionAdjustment" }

// NewPatientService 创建患者服务
func NewPatientService() *PatientService {
	return &PatientService{
		db: database.GetDB(),
	}
}

func normalizeLegacyDialysisMode(raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return models.DialysisModeHD
	}
	upper := strings.ToUpper(v)
	switch upper {
	case "HD", "HDF", "HP", "HF", "HFD", "HD+HP":
		return upper
	}
	return v
}

func boolFromLegacyFlag(raw string) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "1", "true", "yes", "y", "10":
		return true
	default:
		return false
	}
}

func formatLegacyNumber(raw float64) string {
	if raw == 0 {
		return ""
	}
	return strconv.FormatFloat(raw, 'f', -1, 64)
}

func parseLegacyPlanNoteJSON(note string) map[string]any {
	if strings.TrimSpace(note) == "" {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(note), &payload); err != nil {
		return nil
	}
	return payload
}

func legacyNoteString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if payload == nil {
			return ""
		}
		if value, ok := payload[key]; ok {
			if s, ok := value.(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func legacyNoteBool(payload map[string]any, keys ...string) bool {
	for _, key := range keys {
		if payload == nil {
			return false
		}
		if value, ok := payload[key]; ok {
			switch v := value.(type) {
			case bool:
				return v
			case string:
				return boolFromLegacyFlag(v)
			case float64:
				return v != 0
			}
		}
	}
	return false
}

func legacyDrugNameByID(items map[int64]string, id int64) string {
	if id <= 0 {
		return ""
	}
	if name, ok := items[id]; ok {
		return name
	}
	return strconv.FormatInt(id, 10)
}

func (s *PatientService) loadLegacyDrugNames(ids ...int64) (map[int64]string, error) {
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
		return map[int64]string{}, nil
	}

	var rows []struct {
		ID   int64  `gorm:"column:Id"`
		Name string `gorm:"column:Name"`
	}
	if err := s.db.Table(`"Auxiliary_DrugInfomation"`).
		Select(`"Id", "Name"`).
		Where(`"Id" IN ?`, unique).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[int64]string, len(rows))
	for _, row := range rows {
		result[row.ID] = strings.TrimSpace(row.Name)
	}
	return result, nil
}

func (s *PatientService) loadLegacyPlanMaterials(planID int64) (models.MaterialList, error) {
	var rows []legacyPatientPlanMaterial
	if err := s.db.Where(`"PatientPlanId" = ?`, planID).
		Order(`"MaterialGroup" ASC`).
		Order(`"Id" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return models.MaterialList{}, nil
	}

	materialIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.MaterialID > 0 {
			materialIDs = append(materialIDs, row.MaterialID)
		}
	}

	var catalogs []legacyMaterialCatalog
	if len(materialIDs) > 0 {
		if err := s.db.Where(`"Id" IN ?`, materialIDs).Find(&catalogs).Error; err != nil {
			return nil, err
		}
	}
	catalogMap := make(map[int64]legacyMaterialCatalog, len(catalogs))
	for _, catalog := range catalogs {
		catalogMap[catalog.ID] = catalog
	}

	materials := make(models.MaterialList, 0, len(rows))
	for _, row := range rows {
		catalog := catalogMap[row.MaterialID]
		materials = append(materials, models.Material{
			ID:       strconv.FormatInt(row.MaterialID, 10),
			Name:     strings.TrimSpace(catalog.Name),
			Category: strings.TrimSpace(catalog.Classification),
			Count:    int(row.Num),
			Code:     strings.TrimSpace(catalog.Code),
			Brand:    strings.TrimSpace(catalog.Brand),
			Spec:     strings.TrimSpace(catalog.Specification),
			Note:     firstNonEmptyText(row.Note, catalog.Note),
		})
	}

	return materials, nil
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (s *PatientService) toTreatmentPlanDTO(plan legacyPatientPlan, drugNames map[int64]string) (*models.TreatmentPlan, error) {
	materials, err := s.loadLegacyPlanMaterials(plan.ID)
	if err != nil {
		return nil, err
	}
	notePayload := parseLegacyPlanNoteJSON(plan.Note)
	status := models.TreatmentPlanStatusActive
	if plan.IsDisabled {
		status = models.TreatmentPlanStatusInactive
	}

	dto := &models.TreatmentPlan{
		ID:                strconv.FormatInt(plan.ID, 10),
		PatientID:         plan.PatientID,
		WeeklyFrequency:   plan.OddWeekFrequency,
		BiweeklyFrequency: plan.EvenWeekFrequency,
		Duration:          int(plan.DialysisDuration),
		DryWeight:         plan.DryWeight,
		ExtraWeight:       plan.ExtraWeight,
		Status:            status,
		Notes:             strings.TrimSpace(plan.Note),
		CreatedAt:         plan.CreateTime,
		UpdatedAt:         plan.LastModifyTime,
		DialysisMode: models.DialysisMode{
			Mode:                normalizeLegacyDialysisMode(plan.DialysisMethod),
			BloodFlow:           int(plan.BF),
			SubstituteInputMode: firstNonEmptyText(plan.DilutionMnt, legacyNoteString(notePayload, "substituteInputMode", "substituteMode")),
			SubstituteFlow:      plan.SubstituateFlow,
			SubstituteVolume:    plan.SubstituateVolume,
			BV:                  firstNonEmptyText(formatLegacyNumber(plan.BV), legacyNoteString(notePayload, "bv")),
			FrequencyDesc:       firstNonEmptyText(plan.Frequency),
			AutoConfirm:         boolFromLegacyFlag(plan.AutoConfirmPrescriptionRaw) || legacyNoteBool(notePayload, "autoConfirm"),
			Status:              status,
			Notes:               legacyNoteString(notePayload, "dialysisModeNotes"),
		},
		Anticoagulant: models.Anticoagulant{
			InitialDrug:     legacyDrugNameByID(drugNames, plan.FirstAnticoagulant),
			InitialDose:     formatLegacyNumber(plan.FirstDosage),
			MaintenanceDrug: legacyDrugNameByID(drugNames, plan.MaintainAnticoagulant),
			InfusionRate:    formatLegacyNumber(plan.InjectionRate),
			InfusionTime:    formatLegacyNumber(plan.InjectionDuration),
			MaintenanceDose: formatLegacyNumber(plan.DilutionProportion),
			TotalDose:       formatLegacyNumber(plan.InjectionVolume),
		},
		DialysisParameters: models.DialysisParameters{
			DialysateType:  strings.TrimSpace(plan.Dialysate),
			DialysateGroup: strconv.FormatInt(plan.DialysateGroupID, 10),
			FlowRate:       int(plan.DialysateFlow),
			Na:             plan.NaIonCon,
			Ca:             plan.CaIonCon,
			K:              plan.KIonCon,
			HCO3:           plan.HCO3IonCon,
			Glucose:        formatLegacyNumber(plan.GlucoseCon),
			Conductivity:   plan.Conductivity,
			Temp:           plan.DialysateTmp,
			Volume:         plan.DialysateVolume,
		},
		Materials: materials,
	}

	return dto, nil
}

// ListRequest 获取患者列表请求
type ListRequest struct {
	Page             int    `form:"page"`
	PageSize         int    `form:"pageSize"`
	Status           string `form:"status"`
	BedNumber        string `form:"bedNumber"`
	Name             string `form:"name"`
	RiskLevel        string `form:"riskLevel"`
	OnlyActive       bool   `form:"onlyActive"`
	OnlyTransferred  bool   `form:"onlyTransferred"`
}

// ListResponse 获取患者列表响应
type ListResponse struct {
	Items     []models.Patient `json:"items"`
	Total     int64            `json:"total"`
	Page      int              `json:"page"`
	PageSize  int              `json:"pageSize"`
	TotalPage int              `json:"totalPage"`
}

// List 获取患者列表
func (s *PatientService) List(req ListRequest) (*ListResponse, error) {
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

	// TenantId=3 过滤（老血透库多租户）
	query := s.db.Model(&models.Patient{}).Where(`"TenantId" = ?`, legacyTenantID)

	// 筛选条件
	if req.Name != "" {
		query = query.Where(`"Name" LIKE ?`, "%"+req.Name+"%")
	}

	// 通过 Register_OutCome 过滤活跃/转出患者
	if req.OnlyActive {
		activeSubquery := s.db.Table(`"Register_OutCome"`).
			Select(`DISTINCT ON ("PatientId") "PatientId", "Type"`).
			Where(`"TenantId" = ?`, legacyTenantID).
			Order(`"PatientId", "OutComeTime" DESC, "CreateTime" DESC`)
		query = query.Joins(`INNER JOIN (?) AS oc ON oc."PatientId" = "Register_PatientInfomation"."Id" AND oc."Type" = '10'`, activeSubquery)
	}
	if req.OnlyTransferred {
		activeSubquery := s.db.Table(`"Register_OutCome"`).
			Select(`DISTINCT ON ("PatientId") "PatientId", "Type"`).
			Where(`"TenantId" = ?`, legacyTenantID).
			Order(`"PatientId", "OutComeTime" DESC, "CreateTime" DESC`)
		query = query.Joins(`INNER JOIN (?) AS oc ON oc."PatientId" = "Register_PatientInfomation"."Id" AND oc."Type" = '20'`, activeSubquery)
	}
	// 旧 status 筛选保留兼容，但老库 TreatmentStatus 通常为空
	if req.Status != "" && !req.OnlyActive && !req.OnlyTransferred {
		query = query.Where(`"TreatmentStatus" = ?`, req.Status)
	}
	// BedNumber 和 RiskLevel 在老库无对应列，忽略过滤

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.Patient
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order(`"CreateTime" DESC`).
		Find(&items).Error; err != nil {
		return nil, err
	}

	// 计算 Age（老库存 BirthDate，无 Age 列）
	now := time.Now()
	for i := range items {
		if items[i].BirthDate != nil {
			items[i].Age = int(now.Sub(*items[i].BirthDate).Hours() / 8766)
		}
	}

	// 后置填充 DryWeight（从 Plan_PatientPlan）和 Diagnosis（从 Register_Diagnosis）
	s.fillDryWeightAndDiagnosis(items)

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &ListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// fillDryWeightAndDiagnosis 后置填充列表中的干体重和诊断字段
func (s *PatientService) fillDryWeightAndDiagnosis(items []models.Patient) {
	if len(items) == 0 {
		return
	}
	ids := make([]int64, len(items))
	for i, p := range items {
		ids[i] = int64(p.ID)
	}

	// 批量取每个患者最新的 Plan_PatientPlan.DryWeight
	type planRow struct {
		PatientID int64   `gorm:"column:PatientId"`
		DryWeight float64 `gorm:"column:DryWeight"`
	}
	var plans []planRow
	s.db.Table(`"Plan_PatientPlan"`).
		Select(`"PatientId", "DryWeight"`).
		Where(`"PatientId" IN ? AND "TenantId" = ? AND COALESCE("IsDisabled", false) = false`, ids, legacyTenantID).
		Order(`"CreateTime" DESC`).
		Find(&plans)
	planMap := map[int64]float64{}
	for _, r := range plans {
		if _, ok := planMap[r.PatientID]; !ok {
			planMap[r.PatientID] = r.DryWeight
		}
	}

	// 批量取每个患者最新的 Register_Diagnosis.DiagnosisDesc
	type diagRow struct {
		PatientID     int64  `gorm:"column:PatientId"`
		DiagnosisDesc string `gorm:"column:DiagnosisDesc"`
	}
	var diags []diagRow
	s.db.Table(`"Register_Diagnosis"`).
		Select(`"PatientId", COALESCE("DiagnosisDesc", '') AS "DiagnosisDesc"`).
		Where(`"PatientId" IN ? AND "TenantId" = ?`, ids, legacyTenantID).
		Order(`"CreateTime" DESC`).
		Find(&diags)
	diagMap := map[int64]string{}
	for _, r := range diags {
		if _, ok := diagMap[r.PatientID]; !ok {
			diagMap[r.PatientID] = r.DiagnosisDesc
		}
	}

	for i := range items {
		pid := int64(items[i].ID)
		if v, ok := planMap[pid]; ok {
			items[i].DryWeight = v
		}
		if v, ok := diagMap[pid]; ok {
			items[i].Diagnosis = v
		}
	}
}

// PatientStatsResponse 患者统计响应
type PatientStatsResponse struct {
	TotalCount      int64 `json:"totalCount"`
	ActiveCount     int64 `json:"activeCount"`
	OutpatientCount int64 `json:"outpatientCount"`
	InpatientCount  int64 `json:"inpatientCount"`
}

// GetStats 获取患者统计数据
func (s *PatientService) GetStats() (*PatientStatsResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 总人数
	var totalCount int64
	if err := s.db.Model(&models.Patient{}).Where(`"TenantId" = ?`, legacyTenantID).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// 在科活跃：最新 Register_OutCome.Type = '10'（非转出）的患者
	activeSubquery := s.db.Table(`"Register_OutCome"`).
		Select(`DISTINCT ON ("PatientId") "PatientId", "Type"`).
		Where(`"TenantId" = ?`, legacyTenantID).
		Order(`"PatientId", "OutComeTime" DESC, "CreateTime" DESC`)

	var activeCount int64
	if err := s.db.Table(`"Register_PatientInfomation" AS p`).
		Joins(`INNER JOIN (?) AS oc ON oc."PatientId" = p."Id" AND oc."Type" = '10'`, activeSubquery).
		Where(`p."TenantId" = ?`, legacyTenantID).
		Count(&activeCount).Error; err != nil {
		return nil, err
	}

	// 门诊人数
	var outpatientCount int64
	if err := s.db.Model(&models.Patient{}).Where(`"TenantId" = ? AND "PatientType" = ?`, legacyTenantID, "门诊").Count(&outpatientCount).Error; err != nil {
		return nil, err
	}

	// 住院人数
	var inpatientCount int64
	if err := s.db.Model(&models.Patient{}).Where(`"TenantId" = ? AND "PatientType" = ?`, legacyTenantID, "住院").Count(&inpatientCount).Error; err != nil {
		return nil, err
	}

	return &PatientStatsResponse{
		TotalCount:      totalCount,
		ActiveCount:     activeCount,
		OutpatientCount: outpatientCount,
		InpatientCount:  inpatientCount,
	}, nil
}

// Get 获取患者详情
func (s *PatientService) Get(id modeltypes.LegacyID) (*models.Patient, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patient models.Patient
	err := s.db.
		Preload("VascularAccesses").
		Preload("MedicalHistory").
		Preload("TreatmentPlan").
		Where(`"TenantId" = ?`, legacyTenantID).
		First(&patient, `"Id" = ?`, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	// 后置填充 DryWeight 和 Diagnosis
	s.fillDryWeightAndDiagnosis([]models.Patient{patient})

	return &patient, nil
}

// CreateRequest 创建患者请求
type CreateRequest struct {
	Name          string  `json:"name" binding:"required"`
	Age           int     `json:"age" binding:"required,min=0,max=150"`
	Gender        string  `json:"gender" binding:"required,oneof=M F"`
	BedNumber     string  `json:"bedNumber"`
	Diagnosis     string  `json:"diagnosis"`
	RiskLevel     string  `json:"riskLevel"`
	Status        string  `json:"status"`
	PatientType   string  `json:"patientType"`
	InsuranceType string  `json:"insuranceType"`
	DryWeight     float64 `json:"dryWeight"`
	DefaultMode   string  `json:"defaultMode"`
	DoctorID      *string `json:"doctorId"`
	DoctorName    string  `json:"doctorName"`
	// 基本信息档案（可选）
	Pinyin                string  `json:"pinyin"`
	Birthday              *string `json:"birthday"`
	Ethnicity             string  `json:"ethnicity"`
	IdType                string  `json:"idType"`
	IdNumber              string  `json:"idNumber"`
	VisitCategory         string  `json:"visitCategory"`
	AdmissionNo           string  `json:"admissionNo"`
	VisitNo               string  `json:"visitNo"`
	MedicalRecordNo       string  `json:"medicalRecordNo"`
	InsuranceNo           string  `json:"insuranceNo"`
	DialysisNo            string  `json:"dialysisNo"`
	NurseName             string  `json:"nurseName"`
	FirstDialysisDate     *string `json:"firstDialysisDate"`
	FirstHospitalDate     *string `json:"firstHospitalDate"`
	FirstDialysisHospital string  `json:"firstDialysisHospital"`
	Height                string  `json:"height"`
	AboBloodType          string  `json:"aboBloodType"`
	RhBloodType           string  `json:"rhBloodType"`
	EducationLevel        string  `json:"educationLevel"`
	Occupation            string  `json:"occupation"`
	MaritalStatus         string  `json:"maritalStatus"`
	Workplace             string  `json:"workplace"`
	Phone                 string  `json:"phone"`
	Wechat                string  `json:"wechat"`
	Landline              string  `json:"landline"`
	Address               string  `json:"address"`
	District              string  `json:"district"`
	ContactName           string  `json:"contactName"`
	ContactPhone          string  `json:"contactPhone"`
}

// Create 创建患者
func (s *PatientService) Create(req CreateRequest, tenantID int64, creatorID string) (*models.Patient, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(creatorID) == "" {
		return nil, errors.New("creator id is required")
	}
	if tenantID <= 0 {
		return nil, errors.New("tenant id is required")
	}

	createdPatientID, err := idgen.NextID()
	if err != nil {
		return nil, err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		tx = tx.Set("tenant_id", tenantID).Set("creator_id", creatorID)

		patient := models.Patient{
			ID:            modeltypes.LegacyID(createdPatientID),
			Name:          req.Name,
			Age:           req.Age,
			Gender:        req.Gender,
			BedNumber:     req.BedNumber,
			RiskLevel:     req.RiskLevel,
			Status:        req.Status,
			PatientType:   req.PatientType,
			InsuranceType: req.InsuranceType,
			DefaultMode:   req.DefaultMode,
			DoctorID:      req.DoctorID,
			DoctorName:    req.DoctorName,
		}

		if patient.RiskLevel == "" {
			patient.RiskLevel = models.RiskLevelLow
		}
		if patient.Status == "" {
			patient.Status = models.PatientStatusActive
		}

		if err := tx.Create(&patient).Error; err != nil {
			return err
		}

		if err := s.createBasicInfo(tx, patient.ID, req); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	var patient models.Patient
	if err := s.db.First(&patient, `"Id" = ?`, modeltypes.LegacyID(createdPatientID)).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

// createBasicInfo 创建患者基本信息档案
func (s *PatientService) createBasicInfo(tx *gorm.DB, patientID modeltypes.LegacyID, req CreateRequest) error {
	// 创建基本信息档案（可选字段）
	basicInfo := models.PatientBasicInfo{
		ID:                    utils.GenerateID(),
		PatientID:             patientID,
		Pinyin:                stringPtr(req.Pinyin),
		Birthday:              parseTimePointer(req.Birthday),
		Ethnicity:             stringPtr(req.Ethnicity),
		IDType:                req.IdType,
		IDNumber:              stringPtr(req.IdNumber),
		VisitCategory:         stringPtr(req.VisitCategory),
		AdmissionNo:           stringPtr(req.AdmissionNo),
		VisitNo:               stringPtr(req.VisitNo),
		MedicalRecordNo:       stringPtr(req.MedicalRecordNo),
		InsuranceNo:           stringPtr(req.InsuranceNo),
		DialysisNo:            stringPtr(req.DialysisNo),
		NurseName:             stringPtr(req.NurseName),
		FirstDialysisDate:     parseTimePointer(req.FirstDialysisDate),
		FirstHospitalDate:     parseTimePointer(req.FirstHospitalDate),
		FirstDialysisHospital: stringPtr(req.FirstDialysisHospital),
		Height:                stringPtr(req.Height),
		ABOBloodType:          stringPtr(req.AboBloodType),
		RhBloodType:           stringPtr(req.RhBloodType),
		EducationLevel:        stringPtr(req.EducationLevel),
		Occupation:            stringPtr(req.Occupation),
		MaritalStatus:         stringPtr(req.MaritalStatus),
		Workplace:             stringPtr(req.Workplace),
		Phone:                 stringPtr(req.Phone),
		Wechat:                stringPtr(req.Wechat),
		Landline:              stringPtr(req.Landline),
		Address:               stringPtr(req.Address),
		District:              stringPtr(req.District),
		ContactName:           stringPtr(req.ContactName),
		ContactPhone:          stringPtr(req.ContactPhone),
	}
	if err := insertPatientBasicInfo(tx, basicInfo); err != nil {
		return err
	}

	return nil
}

// isDuplicateKeyError 检查是否是主键冲突错误
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL 错误代码 23505 表示 unique_violation
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "23505")
}

// UpdateRequest 更新患者请求
type UpdateRequest struct {
	BedNumber     *string  `json:"bedNumber"`
	Diagnosis     *string  `json:"diagnosis"`
	RiskLevel     *string  `json:"riskLevel"`
	Status        *string  `json:"status"`
	PatientType   *string  `json:"patientType"`
	InsuranceType *string  `json:"insuranceType"`
	DryWeight     *float64 `json:"dryWeight"`
	DefaultMode   *string  `json:"defaultMode"`
	DoctorID      *string  `json:"doctorId"`
	DoctorName    *string  `json:"doctorName"`
}

// Update 更新患者
func (s *PatientService) Update(id modeltypes.LegacyID, req UpdateRequest) (*models.Patient, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patient models.Patient
	if err := s.db.First(&patient, `"Id" = ?`, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.BedNumber != nil {
		updates["bed_number"] = *req.BedNumber
	}
	if req.RiskLevel != nil {
		updates["risk_level"] = *req.RiskLevel
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.PatientType != nil {
		updates["patient_type"] = *req.PatientType
	}
	if req.InsuranceType != nil {
		updates["insurance_type"] = *req.InsuranceType
	}
	if req.DefaultMode != nil {
		updates["default_mode"] = *req.DefaultMode
	}
	if req.DoctorID != nil {
		updates["doctor_id"] = *req.DoctorID
	}
	if req.DoctorName != nil {
		updates["doctor_name"] = *req.DoctorName
	}

	if err := s.db.Model(&patient).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.
		Preload("VascularAccesses").
		Preload("MedicalHistory").
		Preload("TreatmentPlan").
		First(&patient, `"Id" = ?`, id).Error; err != nil {
		return nil, err
	}

	return &patient, nil
}

// Delete 删除患者（硬删除 - 从数据库中真正删除）
func (s *PatientService) Delete(id modeltypes.LegacyID) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 先删除关联的基本信息档案
	s.db.Where("patient_id = ?", id).Delete(&models.PatientBasicInfo{})

	// 硬删除患者记录
	result := s.db.Delete(&models.Patient{}, `"Id" = ?`, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("patient not found")
	}

	return nil
}

// parseTimePointer 解析时间字符串指针
func parseTimePointer(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

// insertPatientBasicInfo 使用显式列插入，避免 GORM 在 *time.Time 字段上的反射写入 panic。
func insertPatientBasicInfo(tx *gorm.DB, basicInfo models.PatientBasicInfo) error {
	now := time.Now()
	idType := basicInfo.IDType
	if strings.TrimSpace(idType) == "" {
		idType = models.IDTypeIDCard
	}

	return tx.Table("patient_basic_infos").Create(map[string]interface{}{
		"id":                      basicInfo.ID,
		"patient_id":              basicInfo.PatientID,
		"pinyin":                  basicInfo.Pinyin,
		"birthday":                basicInfo.Birthday,
		"ethnicity":               basicInfo.Ethnicity,
		"id_type":                 idType,
		"id_number":               basicInfo.IDNumber,
		"visit_category":          basicInfo.VisitCategory,
		"admission_no":            basicInfo.AdmissionNo,
		"visit_no":                basicInfo.VisitNo,
		"medical_record_no":       basicInfo.MedicalRecordNo,
		"insurance_no":            basicInfo.InsuranceNo,
		"hdis_patient_id":         basicInfo.HdisPatientID,
		"dialysis_no":             basicInfo.DialysisNo,
		"nurse_name":              basicInfo.NurseName,
		"first_dialysis_date":     basicInfo.FirstDialysisDate,
		"first_hospital_date":     basicInfo.FirstHospitalDate,
		"first_dialysis_hospital": basicInfo.FirstDialysisHospital,
		"height":                  basicInfo.Height,
		"abo_blood_type":          basicInfo.ABOBloodType,
		"rh_blood_type":           basicInfo.RhBloodType,
		"education_level":         basicInfo.EducationLevel,
		"occupation":              basicInfo.Occupation,
		"marital_status":          basicInfo.MaritalStatus,
		"workplace":               basicInfo.Workplace,
		"phone":                   basicInfo.Phone,
		"wechat":                  basicInfo.Wechat,
		"landline":                basicInfo.Landline,
		"address":                 basicInfo.Address,
		"district":                basicInfo.District,
		"contact_name":            basicInfo.ContactName,
		"contact_phone":           basicInfo.ContactPhone,
		"created_at":              now,
		"updated_at":              now,
	}).Error
}

// stringPtr 将字符串转换为指针，空字符串返回 nil
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ===== 治疗方案相关方法 =====

// GetTreatmentPlans 获取患者的所有治疗方案
func (s *PatientService) GetTreatmentPlans(patientID modeltypes.LegacyID) ([]*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var legacyPlans []legacyPatientPlan
	err := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, legacyTenantID).
		Order(`"IsDisabled" ASC`).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Find(&legacyPlans).Error
	if err != nil {
		return nil, err
	}

	if len(legacyPlans) == 0 {
		return []*models.TreatmentPlan{}, nil
	}

	drugIDs := make([]int64, 0, len(legacyPlans)*2)
	for _, plan := range legacyPlans {
		drugIDs = append(drugIDs, plan.FirstAnticoagulant, plan.MaintainAnticoagulant)
	}
	drugNames, err := s.loadLegacyDrugNames(drugIDs...)
	if err != nil {
		return nil, err
	}

	plans := make([]*models.TreatmentPlan, 0, len(legacyPlans))
	for _, plan := range legacyPlans {
		dto, convErr := s.toTreatmentPlanDTO(plan, drugNames)
		if convErr != nil {
			return nil, convErr
		}
		plans = append(plans, dto)
	}

	return plans, nil
}

// GetTreatmentPlan 获取患者治疗方案（可指定透析模式）
func (s *PatientService) GetTreatmentPlan(patientID modeltypes.LegacyID, mode ...string) (*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var plan models.TreatmentPlan
	var err error

	// 如果提供了模式参数，则查询特定模式的治疗方案
	if len(mode) > 0 && mode[0] != "" {
		err = s.db.Where("patient_id = ? AND dialysis_mode->>'mode' = ?", patientID, mode[0]).First(&plan).Error
	} else {
		// 否则返回患者的第一个治疗方案
		err = s.db.Where("patient_id = ?", patientID).First(&plan).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有治疗方案，返回 nil 而不是错误
		}
		return nil, err
	}

	return &plan, nil
}

// CreateTreatmentPlanRequest 创建治疗方案请求
type CreateTreatmentPlanRequest struct {
	WeeklyFrequency    int                       `json:"weeklyFrequency"`
	BiweeklyFrequency  int                       `json:"biweeklyFrequency"`
	Duration           int                       `json:"duration"`
	DryWeight          float64                   `json:"dryWeight"`
	ExtraWeight        float64                   `json:"extraWeight"`
	Status             string                    `json:"status"`
	Notes              string                    `json:"notes"`
	DialysisMode       models.DialysisMode       `json:"dialysisMode"`
	Anticoagulant      models.Anticoagulant      `json:"anticoagulant"`
	DialysisParameters models.DialysisParameters `json:"parameters"`
	Materials          models.MaterialList       `json:"materials"`
}

// CreateTreatmentPlan 创建患者治疗方案
func (s *PatientService) CreateTreatmentPlan(patientID modeltypes.LegacyID, req CreateTreatmentPlanRequest) (*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 检查患者是否存在
	var patient models.Patient
	if err := s.db.First(&patient, `"Id" = ?`, patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	// 检查该患者是否已存在相同透析模式的治疗方案
	var existingPlan models.TreatmentPlan
	err := s.db.Where("patient_id = ? AND dialysis_mode->>'mode' = ?", patientID, req.DialysisMode.Mode).First(&existingPlan).Error
	if err == nil {
		return nil, fmt.Errorf("该患者已存在 %s 模式的治疗方案", req.DialysisMode.Mode)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	plan := models.TreatmentPlan{
		ID:                 utils.GenerateID(),
		PatientID:          patientID,
		WeeklyFrequency:    req.WeeklyFrequency,
		BiweeklyFrequency:  req.BiweeklyFrequency,
		Duration:           req.Duration,
		DryWeight:          req.DryWeight,
		ExtraWeight:        req.ExtraWeight,
		Status:             req.Status,
		Notes:              req.Notes,
		DialysisMode:       req.DialysisMode,
		Anticoagulant:      req.Anticoagulant,
		DialysisParameters: req.DialysisParameters,
		Materials:          req.Materials,
	}

	if err := s.db.Create(&plan).Error; err != nil {
		return nil, err
	}

	return &plan, nil
}

// UpdateTreatmentPlanRequest 更新治疗方案请求
type UpdateTreatmentPlanRequest struct {
	WeeklyFrequency    *int                       `json:"weeklyFrequency"`
	BiweeklyFrequency  *int                       `json:"biweeklyFrequency"`
	Duration           *int                       `json:"duration"`
	DryWeight          *float64                   `json:"dryWeight"`
	ExtraWeight        *float64                   `json:"extraWeight"`
	Status             *string                    `json:"status"`
	Notes              *string                    `json:"notes"`
	DialysisMode       *models.DialysisMode       `json:"dialysisMode"`
	Anticoagulant      *models.Anticoagulant      `json:"anticoagulant"`
	DialysisParameters *models.DialysisParameters `json:"parameters"`
	Materials          *models.MaterialList       `json:"materials"`
}

// UpdateTreatmentPlan 更新患者治疗方案
func (s *PatientService) UpdateTreatmentPlan(patientID modeltypes.LegacyID, req UpdateTreatmentPlanRequest) (*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var plan models.TreatmentPlan
	// 必须指定透析模式才能精确匹配方案
	if req.DialysisMode == nil || req.DialysisMode.Mode == "" {
		return nil, errors.New("dialysisMode.mode is required for update")
	}
	query := s.db.Where("patient_id = ? AND dialysis_mode->>'mode' = ?", patientID, req.DialysisMode.Mode)
	if err := query.First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("treatment plan not found")
		}
		return nil, err
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.WeeklyFrequency != nil {
		updates["weekly_frequency"] = *req.WeeklyFrequency
	}
	if req.BiweeklyFrequency != nil {
		updates["biweekly_frequency"] = *req.BiweeklyFrequency
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.DryWeight != nil {
		updates["dry_weight"] = *req.DryWeight
	}
	if req.ExtraWeight != nil {
		updates["extra_weight"] = *req.ExtraWeight
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.DialysisMode != nil {
		updates["dialysis_mode"] = *req.DialysisMode
	}
	if req.Anticoagulant != nil {
		updates["anticoagulant"] = *req.Anticoagulant
	}
	if req.DialysisParameters != nil {
		updates["dialysis_parameters"] = *req.DialysisParameters
	}
	if req.Materials != nil {
		updates["materials"] = *req.Materials
	}

	if err := s.db.Model(&plan).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.Where("patient_id = ?", patientID).Where("id = ?", plan.ID).First(&plan).Error; err != nil {
		return nil, err
	}

	return &plan, nil
}

// DeleteTreatmentPlan 删除患者治疗方案
func (s *PatientService) DeleteTreatmentPlan(patientID modeltypes.LegacyID) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Where("patient_id = ?", patientID).Delete(&models.TreatmentPlan{})
	if result.Error != nil {
		return result.Error
	}
	// 不检查 RowsAffected，允许删除不存在的方案
	return nil
}

// ===== 方案调整记录 =====

// CreateAdjustmentRecordRequest 创建调整记录请求
type CreateAdjustmentRecordRequest struct {
	Content                   string `json:"content" binding:"required"`
	Operator                  string `json:"operator"`
	PatientPlanPrescriptionID *int64 `json:"patientPlanPrescriptionId,omitempty"`
}

// GetAdjustmentRecords 获取患者方案调整记录列表
func (s *PatientService) GetAdjustmentRecords(patientID modeltypes.LegacyID) ([]models.AdjustmentRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var records []models.AdjustmentRecord
	if err := s.db.Where("patient_id = ?", patientID).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// CreateAdjustmentRecord 创建方案调整记录
func (s *PatientService) CreateAdjustmentRecord(patientID modeltypes.LegacyID, req CreateAdjustmentRecordRequest) (*models.AdjustmentRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 校验患者是否存在
	var count int64
	if err := s.db.Model(&models.Patient{}).Where(`"Id" = ?`, patientID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	record := models.AdjustmentRecord{
		ID:        utils.GenerateID(),
		PatientID: patientID,
		Content:   req.Content,
		Operator:  req.Operator,
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *PatientService) GetLegacyTreatmentPlans(patientID modeltypes.LegacyID) ([]*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var legacyPlans []legacyPatientPlan
	if err := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, legacyTenantID).
		Order(`"IsDisabled" ASC`).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Find(&legacyPlans).Error; err != nil {
		return nil, err
	}

	if len(legacyPlans) == 0 {
		return []*models.TreatmentPlan{}, nil
	}

	drugIDs := make([]int64, 0, len(legacyPlans)*2)
	for _, plan := range legacyPlans {
		drugIDs = append(drugIDs, plan.FirstAnticoagulant, plan.MaintainAnticoagulant)
	}

	drugNames, err := s.loadLegacyDrugNames(drugIDs...)
	if err != nil {
		return nil, err
	}

	plans := make([]*models.TreatmentPlan, 0, len(legacyPlans))
	for _, plan := range legacyPlans {
		dto, convErr := s.toTreatmentPlanDTO(plan, drugNames)
		if convErr != nil {
			return nil, convErr
		}
		plans = append(plans, dto)
	}

	return plans, nil
}

func (s *PatientService) GetLegacyTreatmentPlan(patientID modeltypes.LegacyID, mode ...string) (*models.TreatmentPlan, error) {
	plans, err := s.GetLegacyTreatmentPlans(patientID)
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, nil
	}

	if len(mode) > 0 && strings.TrimSpace(mode[0]) != "" {
		targetMode := normalizeLegacyDialysisMode(mode[0])
		for _, plan := range plans {
			if normalizeLegacyDialysisMode(plan.DialysisMode.Mode) == targetMode {
				return plan, nil
			}
		}
		return nil, nil
	}

	return plans[0], nil
}

func (s *PatientService) GetLegacyAdjustmentRecords(patientID modeltypes.LegacyID) ([]models.AdjustmentRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var prescriptions []legacyPatientPrescription
	if err := s.db.Where(`"PatientId" = ?`, patientID).Find(&prescriptions).Error; err != nil {
		return nil, err
	}
	if len(prescriptions) == 0 {
		return []models.AdjustmentRecord{}, nil
	}

	prescriptionIDs := make([]int64, 0, len(prescriptions))
	for _, prescription := range prescriptions {
		prescriptionIDs = append(prescriptionIDs, prescription.ID)
	}

	var legacyRecords []legacyAdjustmentRecord
	if err := s.db.Where(`"PatientPlanPrescriptionId" IN ?`, prescriptionIDs).
		Order(`"AdjustTime" DESC`).
		Find(&legacyRecords).Error; err != nil {
		return nil, err
	}

	records := make([]models.AdjustmentRecord, 0, len(legacyRecords))
	prescriptionService := &PrescriptionService{db: s.db}
	for _, record := range legacyRecords {
		operatorName, err := prescriptionService.lookupLegacyUserDisplayName(record.AdjustUserID)
		if err != nil || strings.TrimSpace(operatorName) == "" {
			operatorName = "--"
		}
		createdAt := record.CreateTime
		if createdAt.IsZero() {
			createdAt = record.AdjustTime
		}
		records = append(records, models.AdjustmentRecord{
			ID:        strconv.FormatInt(record.ID, 10),
			PatientID: patientID,
			Content:   strings.TrimSpace(record.AdjustReason),
			Operator:  operatorName,
			CreatedAt: createdAt,
		})
	}

	return records, nil
}

func parseLegacyNumericID(raw string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseStringFloat(raw string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}
	return v
}

func normalizePlanStatus(status string, modeStatus string) string {
	if strings.TrimSpace(status) != "" {
		return strings.TrimSpace(status)
	}
	if strings.TrimSpace(modeStatus) != "" {
		return strings.TrimSpace(modeStatus)
	}
	return models.TreatmentPlanStatusActive
}

func legacyPlanDisabled(status string) bool {
	return strings.TrimSpace(status) == models.TreatmentPlanStatusInactive
}

func boolToLegacyFlag(v bool) string {
	if v {
		return "10"
	}
	return "0"
}

func (s *PatientService) findLegacyDrugIDByName(name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, nil
	}

	var row struct {
		ID int64 `gorm:"column:Id"`
	}
	if err := s.db.Table(`"Auxiliary_DrugInfomation"`).
		Select(`"Id"`).
		Where(`"TenantId" = ? AND "Name" = ?`, legacyTenantID, name).
		Limit(1).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return row.ID, nil
}

func (s *PatientService) findLegacyMaterialID(material models.Material) (int64, error) {
	if id := parseLegacyNumericID(material.ID); id > 0 {
		return id, nil
	}
	name := strings.TrimSpace(material.Name)
	if name == "" {
		return 0, nil
	}

	var row struct {
		ID int64 `gorm:"column:Id"`
	}
	if err := s.db.Table(`"Auxiliary_MaterialInfomation"`).
		Select(`"Id"`).
		Where(`"TenantId" = ? AND "Name" = ?`, legacyTenantID, name).
		Limit(1).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return row.ID, nil
}

func (s *PatientService) syncLegacyPlanMaterials(tx *gorm.DB, planID int64, materials models.MaterialList) error {
	if err := tx.Where(`"PatientPlanId" = ?`, planID).Delete(&legacyPatientPlanMaterial{}).Error; err != nil {
		return err
	}

	now := time.Now()
	for idx, material := range materials {
		materialID, err := s.findLegacyMaterialID(material)
		if err != nil {
			return err
		}
		if materialID == 0 {
			continue
		}
		id, err := idgen.NextID()
		if err != nil {
			return err
		}
		row := map[string]any{
			"Id":             id,
			"TenantId":       legacyTenantID,
			"PatientPlanId":  planID,
			"MaterialId":     materialID,
			"MaterialGroup":  idx + 1,
			"Num":            material.Count,
			"Note":           strings.TrimSpace(material.Note),
			"LastModifyTime": now,
		}
		if err := tx.Table(`"Plan_PatientPlanMaterial"`).Create(row).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *PatientService) legacyPlanByMode(patientID modeltypes.LegacyID, mode string) (*legacyPatientPlan, error) {
	var plan legacyPatientPlan
	query := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, legacyTenantID)
	if strings.TrimSpace(mode) != "" {
		query = query.Where(`"DialysisMethod" = ?`, normalizeLegacyDialysisMode(mode))
	}
	err := query.
		Order(`"IsDisabled" ASC`).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *PatientService) LegacyCreateTreatmentPlan(patientID modeltypes.LegacyID, req CreateTreatmentPlanRequest) (*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patient models.Patient
	if err := s.db.First(&patient, `"Id" = ?`, patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	if existing, err := s.legacyPlanByMode(patientID, req.DialysisMode.Mode); err == nil && existing != nil {
		return nil, fmt.Errorf("该患者已存在 %s 模式的治疗方案", req.DialysisMode.Mode)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	firstDrugID, err := s.findLegacyDrugIDByName(req.Anticoagulant.InitialDrug)
	if err != nil {
		return nil, err
	}
	maintainDrugID, err := s.findLegacyDrugIDByName(req.Anticoagulant.MaintenanceDrug)
	if err != nil {
		return nil, err
	}
	planID, err := idgen.NextID()
	if err != nil {
		return nil, err
	}

	status := normalizePlanStatus(req.Status, req.DialysisMode.Status)
	now := time.Now()
	createMap := map[string]any{
		"Id":                      planID,
		"TenantId":                legacyTenantID,
		"PatientId":               patientID,
		"Name":                    patient.Name,
		"CreatorId":               0,
		"CreateTime":              now,
		"OddWeekFrequency":        req.WeeklyFrequency,
		"EvenWeekFrequency":       req.BiweeklyFrequency,
		"DialysisMethod":          normalizeLegacyDialysisMode(req.DialysisMode.Mode),
		"DialysisDuration":        req.Duration,
		"DryWeight":               req.DryWeight,
		"ExtraWeight":             req.ExtraWeight,
		"BF":                      req.DialysisMode.BloodFlow,
		"BV":                      parseStringFloat(req.DialysisMode.BV),
		"FirstAnticoagulant":      firstDrugID,
		"FirstDosage":             parseStringFloat(req.Anticoagulant.InitialDose),
		"MaintainAnticoagulant":   maintainDrugID,
		"DilutionProportion":      parseStringFloat(req.Anticoagulant.MaintenanceDose),
		"InjectionRate":           parseStringFloat(req.Anticoagulant.InfusionRate),
		"InjectionDuration":       parseStringFloat(req.Anticoagulant.InfusionTime),
		"InjectionVolume":         parseStringFloat(req.Anticoagulant.TotalDose),
		"Dialysate":               strings.TrimSpace(req.DialysisParameters.DialysateType),
		"DialysateFlow":           req.DialysisParameters.FlowRate,
		"DialysateVolume":         req.DialysisParameters.Volume,
		"NaIonCon":                req.DialysisParameters.Na,
		"CaIonCon":                req.DialysisParameters.Ca,
		"KIonCon":                 req.DialysisParameters.K,
		"HCO3IonCon":              req.DialysisParameters.HCO3,
		"Conductivity":            req.DialysisParameters.Conductivity,
		"DialysateTmp":            req.DialysisParameters.Temp,
		"SubstituateVolume":       req.DialysisMode.SubstituteVolume,
		"DilutionMnt":             strings.TrimSpace(req.DialysisMode.SubstituteInputMode),
		"IsDisabled":              legacyPlanDisabled(status),
		"LastModifyTime":          now,
		"Frequency":               strings.TrimSpace(req.DialysisMode.FrequencyDesc),
		"GlucoseCon":              parseStringFloat(req.DialysisParameters.Glucose),
		"DialysateGroupId":        parseLegacyNumericID(req.DialysisParameters.DialysateGroup),
		"AutoConfirmPrescription": boolToLegacyFlag(req.DialysisMode.AutoConfirm),
		"Note":                    strings.TrimSpace(req.Notes),
		"SubstituateFlow":         req.DialysisMode.SubstituteFlow,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(`"Plan_PatientPlan"`).Create(createMap).Error; err != nil {
			return err
		}
		return s.syncLegacyPlanMaterials(tx, planID, req.Materials)
	}); err != nil {
		return nil, err
	}

	return s.GetLegacyTreatmentPlan(patientID, req.DialysisMode.Mode)
}

func (s *PatientService) LegacyUpdateTreatmentPlan(patientID modeltypes.LegacyID, req UpdateTreatmentPlanRequest) (*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.DialysisMode == nil || strings.TrimSpace(req.DialysisMode.Mode) == "" {
		return nil, errors.New("dialysisMode.mode is required for update")
	}

	plan, err := s.legacyPlanByMode(patientID, req.DialysisMode.Mode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("treatment plan not found")
		}
		return nil, err
	}

	updates := map[string]any{
		"LastModifyTime": time.Now(),
	}
	if req.WeeklyFrequency != nil {
		updates["OddWeekFrequency"] = *req.WeeklyFrequency
	}
	if req.BiweeklyFrequency != nil {
		updates["EvenWeekFrequency"] = *req.BiweeklyFrequency
	}
	if req.Duration != nil {
		updates["DialysisDuration"] = *req.Duration
	}
	if req.DryWeight != nil {
		updates["DryWeight"] = *req.DryWeight
	}
	if req.ExtraWeight != nil {
		updates["ExtraWeight"] = *req.ExtraWeight
	}
	if req.Status != nil {
		updates["IsDisabled"] = legacyPlanDisabled(*req.Status)
	}
	if req.Notes != nil {
		updates["Note"] = strings.TrimSpace(*req.Notes)
	}
	if req.DialysisMode != nil {
		updates["DialysisMethod"] = normalizeLegacyDialysisMode(req.DialysisMode.Mode)
		updates["BF"] = req.DialysisMode.BloodFlow
		updates["BV"] = parseStringFloat(req.DialysisMode.BV)
		updates["SubstituateVolume"] = req.DialysisMode.SubstituteVolume
		updates["DilutionMnt"] = strings.TrimSpace(req.DialysisMode.SubstituteInputMode)
		updates["Frequency"] = strings.TrimSpace(req.DialysisMode.FrequencyDesc)
		updates["AutoConfirmPrescription"] = boolToLegacyFlag(req.DialysisMode.AutoConfirm)
		if req.Status == nil {
			updates["IsDisabled"] = legacyPlanDisabled(normalizePlanStatus("", req.DialysisMode.Status))
		}
		updates["SubstituateFlow"] = req.DialysisMode.SubstituteFlow
	}
	if req.Anticoagulant != nil {
		firstDrugID, err := s.findLegacyDrugIDByName(req.Anticoagulant.InitialDrug)
		if err != nil {
			return nil, err
		}
		maintainDrugID, err := s.findLegacyDrugIDByName(req.Anticoagulant.MaintenanceDrug)
		if err != nil {
			return nil, err
		}
		updates["FirstAnticoagulant"] = firstDrugID
		updates["FirstDosage"] = parseStringFloat(req.Anticoagulant.InitialDose)
		updates["MaintainAnticoagulant"] = maintainDrugID
		updates["DilutionProportion"] = parseStringFloat(req.Anticoagulant.MaintenanceDose)
		updates["InjectionRate"] = parseStringFloat(req.Anticoagulant.InfusionRate)
		updates["InjectionDuration"] = parseStringFloat(req.Anticoagulant.InfusionTime)
		updates["InjectionVolume"] = parseStringFloat(req.Anticoagulant.TotalDose)
	}
	if req.DialysisParameters != nil {
		updates["Dialysate"] = strings.TrimSpace(req.DialysisParameters.DialysateType)
		updates["DialysateFlow"] = req.DialysisParameters.FlowRate
		updates["DialysateVolume"] = req.DialysisParameters.Volume
		updates["NaIonCon"] = req.DialysisParameters.Na
		updates["CaIonCon"] = req.DialysisParameters.Ca
		updates["KIonCon"] = req.DialysisParameters.K
		updates["HCO3IonCon"] = req.DialysisParameters.HCO3
		updates["Conductivity"] = req.DialysisParameters.Conductivity
		updates["DialysateTmp"] = req.DialysisParameters.Temp
		updates["GlucoseCon"] = parseStringFloat(req.DialysisParameters.Glucose)
		if groupID := parseLegacyNumericID(req.DialysisParameters.DialysateGroup); groupID > 0 {
			updates["DialysateGroupId"] = groupID
		}
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(`"Plan_PatientPlan"`).Where(`"Id" = ?`, plan.ID).Updates(updates).Error; err != nil {
			return err
		}
		if req.Materials != nil {
			return s.syncLegacyPlanMaterials(tx, plan.ID, *req.Materials)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.GetLegacyTreatmentPlan(patientID, req.DialysisMode.Mode)
}

func (s *PatientService) LegacyDeleteTreatmentPlan(patientID modeltypes.LegacyID) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	return s.db.Table(`"Plan_PatientPlan"`).
		Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, legacyTenantID).
		Updates(map[string]any{
			"IsDisabled":     true,
			"LastModifyTime": time.Now(),
		}).Error
}

func (s *PatientService) LegacyCreateAdjustmentRecord(patientID modeltypes.LegacyID, req CreateAdjustmentRecordRequest) (*models.AdjustmentRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	resolvedPrescriptionID := int64(0)
	if req.PatientPlanPrescriptionID != nil && *req.PatientPlanPrescriptionID > 0 {
		var matchCount int64
		if err := s.db.Table(`"Plan_PatientPrescription"`).
			Where(`"TenantId" = ? AND "Id" = ? AND "PatientId" = ?`, legacyTenantID, *req.PatientPlanPrescriptionID, patientID).
			Count(&matchCount).Error; err != nil {
			return nil, err
		}
		if matchCount > 0 {
			resolvedPrescriptionID = *req.PatientPlanPrescriptionID
		}
	}
	if resolvedPrescriptionID <= 0 {
		var prescription legacyPatientPrescription
		if err := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, legacyTenantID).
			Order(`"LastModifyTime" DESC`).
			Order(`"Id" DESC`).
			First(&prescription).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("no legacy prescription found for adjustment record")
			}
			return nil, err
		}
		resolvedPrescriptionID = prescription.ID
	}

	recordID, err := idgen.NextID()
	if err != nil {
		return nil, err
	}

	prescriptionService := &PrescriptionService{db: s.db}
	operatorUserID, operatorName, err := prescriptionService.resolveLegacyUserID(req.Operator, req.Operator)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	row := map[string]any{
		"Id":                        recordID,
		"TenantId":                  legacyTenantID,
		"Type":                      0,
		"PatientPlanPrescriptionId": resolvedPrescriptionID,
		"AdjustUserId":              operatorUserID,
		"AdjustTime":                now,
		"AdjustReason":              strings.TrimSpace(req.Content),
		"CreateTime":                now,
	}
	if err := s.db.Table(`"Plan_PatientPlanPrescriptionAdjustment"`).Create(row).Error; err != nil {
		return nil, err
	}

	return &models.AdjustmentRecord{
		ID:        strconv.FormatInt(recordID, 10),
		PatientID: patientID,
		Content:   strings.TrimSpace(req.Content),
		Operator:  operatorName,
		CreatedAt: now,
	}, nil
}
