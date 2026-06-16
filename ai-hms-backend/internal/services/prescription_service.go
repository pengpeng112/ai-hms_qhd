package services

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	legacymodels "github.com/elliotxin/ai-hms-backend/internal/models/legacy"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/elliotxin/ai-hms-backend/internal/utils/idgen"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type PrescriptionService struct {
	db *gorm.DB
}

type legacyPrescriptionMaterial struct {
	ID                    int64     `gorm:"column:Id"`
	TenantID              int64     `gorm:"column:TenantId"`
	PatientPrescriptionID int64     `gorm:"column:PatientPrescriptionId"`
	MaterialID            int64     `gorm:"column:MaterialId"`
	MaterialGroup         int64     `gorm:"column:MaterialGroup"`
	Num                   float64   `gorm:"column:Num"`
	Note                  string    `gorm:"column:Note"`
	LastModifyTime        time.Time `gorm:"column:LastModifyTime"`
	ChargeItemID          int64     `gorm:"column:ChargeItemId"`
}

func (legacyPrescriptionMaterial) TableName() string { return "Plan_PatientPrescriptionMaterial" }

type legacyPrescriptionNote struct {
	Notes            string                           `json:"notes,omitempty"`
	ExtraWeight      float64                          `json:"extraWeight,omitempty"`
	OrderItems       models.PrescriptionOrderItemList `json:"orderItems,omitempty"`
	PrescriptionDate string                           `json:"prescriptionDate,omitempty"`
	DoctorName       string                           `json:"doctorName,omitempty"`
}

func NewPrescriptionService() *PrescriptionService {
	return &PrescriptionService{
		db: database.GetDB(),
	}
}

type PrescriptionCreateRequest struct {
	PrescriptionDate string                           `json:"prescriptionDate" binding:"required"`
	Duration         int                              `json:"duration"`
	DryWeight        float64                          `json:"dryWeight"`
	ExtraWeight      float64                          `json:"extraWeight"`
	DialysisMode     models.DialysisMode              `json:"dialysisMode"`
	Anticoagulant    models.Anticoagulant             `json:"anticoagulant"`
	Parameters       models.DialysisParameters        `json:"parameters"`
	Materials        models.MaterialList              `json:"materials"`
	OrderItems       models.PrescriptionOrderItemList `json:"orderItems"`
	Notes            string                           `json:"notes"`
}

type PrescriptionUpdateRequest struct {
	Duration      *int                              `json:"duration"`
	DryWeight     *float64                          `json:"dryWeight"`
	ExtraWeight   *float64                          `json:"extraWeight"`
	DialysisMode  *models.DialysisMode              `json:"dialysisMode"`
	Anticoagulant *models.Anticoagulant             `json:"anticoagulant"`
	Parameters    *models.DialysisParameters        `json:"parameters"`
	Materials     *models.MaterialList              `json:"materials"`
	OrderItems    *models.PrescriptionOrderItemList `json:"orderItems"`
	Notes         *string                           `json:"notes"`
}

type PrescriptionExtractRequest struct {
	Date string `json:"date" binding:"required"`
}

func mapLegacyPrescriptionStatus(status int) string {
	switch status {
	case 2:
		return models.PrescriptionStatusExecuted
	case 3:
		return models.PrescriptionStatusCancelled
	case 1, 0:
		return models.PrescriptionStatusPending
	default:
		return models.PrescriptionStatusPending
	}
}

func mapNewPrescriptionStatus(status string) int {
	switch strings.TrimSpace(status) {
	case models.PrescriptionStatusPending:
		return 1 // 草稿
	case models.PrescriptionStatusExecuting:
		return 2 // 确认（老库无“执行中”，折叠为确认）
	case models.PrescriptionStatusExecuted:
		return 2 // 确认
	case models.PrescriptionStatusCancelled:
		return 3 // 作废
	default:
		return 0
	}
}

func (s *PrescriptionService) loadLegacyPrescriptionTreatmentDate(treatmentID int64) *time.Time {
	if treatmentID <= 0 {
		return nil
	}

	var row struct {
		StartTime     *time.Time `gorm:"column:StartTime"`
		ReceptionTime *time.Time `gorm:"column:ReceptionTime"`
		CreateTime    time.Time  `gorm:"column:CreateTime"`
	}
	err := s.db.Table(`"Treatment_Treatment"`).
		Select(`"StartTime", "ReceptionTime", "CreateTime"`).
		Where(`"Id" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Limit(1).
		First(&row).Error
	if err != nil {
		return nil
	}
	if row.StartTime != nil && !row.StartTime.IsZero() {
		return row.StartTime
	}
	if row.ReceptionTime != nil && !row.ReceptionTime.IsZero() {
		return row.ReceptionTime
	}
	if !row.CreateTime.IsZero() {
		t := row.CreateTime
		return &t
	}
	return nil
}

func parseLegacyPrescriptionNote(raw string) (*legacyPrescriptionNote, bool) {
	if strings.TrimSpace(raw) == "" {
		return &legacyPrescriptionNote{}, true
	}
	var note legacyPrescriptionNote
	if err := json.Unmarshal([]byte(raw), &note); err != nil {
		return nil, false
	}
	return &note, true
}

func buildLegacyPrescriptionNote(notes string, extraWeight float64, date time.Time, doctorName string, orderItems models.PrescriptionOrderItemList) string {
	payload := legacyPrescriptionNote{
		Notes:       strings.TrimSpace(notes),
		ExtraWeight: extraWeight,
		DoctorName:  strings.TrimSpace(doctorName),
		OrderItems:  orderItems,
	}
	if !date.IsZero() {
		payload.PrescriptionDate = date.Format("2006-01-02")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return strings.TrimSpace(notes)
	}
	return string(data)
}

func isMissingLegacyColumnError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42703" {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "undefined_column") || strings.Contains(lower, "column") && strings.Contains(lower, "does not exist")
}

func (s *PrescriptionService) resolveLegacyUserID(userIDRaw, fallbackName string) (int64, string, error) {
	if userID := parseLegacyNumericID(userIDRaw); userID > 0 {
		name, err := s.lookupLegacyUserDisplayName(userID)
		if err != nil {
			return 0, "", err
		}
		if name == "" {
			name = strings.TrimSpace(fallbackName)
		}
		return userID, name, nil
	}

	fallbackName = strings.TrimSpace(fallbackName)
	if fallbackName == "" {
		return 0, "", nil
	}

	var employee struct {
		ID     int64  `gorm:"column:Id"`
		UserID int64  `gorm:"column:UserId"`
		Name   string `gorm:"column:Name"`
	}
	queries := []struct {
		selectSQL string
		whereSQL  string
		args      []any
	}{
		{`"Id", "UserId", "Name"`, `"Name" = ? AND "TenantId" = ?`, []any{fallbackName, LegacyTenantID}},
		{`"Id", "UserId", "Name"`, `"Name" = ?`, []any{fallbackName}},
		{`"Id", "Name"`, `"Name" = ? AND "TenantId" = ?`, []any{fallbackName, LegacyTenantID}},
		{`"Id", "Name"`, `"Name" = ?`, []any{fallbackName}},
	}
	var lastErr error
	for _, query := range queries {
		err := s.db.Table(`"Organ_Employee"`).
			Select(query.selectSQL).
			Where(query.whereSQL, query.args...).
			Order(`"Id" DESC`).
			First(&employee).Error
		if err == nil {
			resolved := employee.UserID
			if resolved <= 0 {
				resolved = employee.ID
			}
			return resolved, strings.TrimSpace(employee.Name), nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fallbackName, nil
		}
		if isMissingLegacyColumnError(err) {
			lastErr = err
			continue
		}
		return 0, "", err
	}
	if lastErr != nil {
		return 0, fallbackName, nil
	}
	return 0, fallbackName, nil
}

func (s *PrescriptionService) lookupLegacyUserDisplayName(userID int64) (string, error) {
	if userID <= 0 {
		return "", nil
	}

	var employee struct {
		Name string `gorm:"column:Name"`
	}
	queries := []struct {
		whereSQL string
		args     []any
	}{
		{`"UserId" = ? AND "TenantId" = ?`, []any{userID, LegacyTenantID}},
		{`"UserId" = ?`, []any{userID}},
		{`"Id" = ? AND "TenantId" = ?`, []any{userID, LegacyTenantID}},
		{`"Id" = ?`, []any{userID}},
	}
	var err error
	for _, query := range queries {
		err = s.db.Table(`"Organ_Employee"`).
			Select(`"Name"`).
			Where(query.whereSQL, query.args...).
			Order(`"Id" ASC`).
			First(&employee).Error
		if err == nil && strings.TrimSpace(employee.Name) != "" {
			return strings.TrimSpace(employee.Name), nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break
		}
		if isMissingLegacyColumnError(err) {
			continue
		}
		if err != nil {
			return "", err
		}
	}

	var identityUser legacymodels.IdentityUser
	err = s.db.Model(&legacymodels.IdentityUser{}).
		Select(`"UserName"`).
		Where(`"Id" = ?`, userID).
		First(&identityUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return strconv.FormatInt(userID, 10), nil
		}
		return "", err
	}
	if strings.TrimSpace(identityUser.UserName) != "" {
		return strings.TrimSpace(identityUser.UserName), nil
	}
	return strconv.FormatInt(userID, 10), nil
}

func (s *PrescriptionService) getLegacyPlanForPrescription(patientID, mode string) (*legacyPatientPlan, error) {
	planService := &PatientService{db: s.db}
	if strings.TrimSpace(mode) != "" {
		plan, err := planService.legacyPlanByMode(parseLegacyIDOrZero(patientID), mode)
		if err == nil {
			return plan, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return planService.legacyPlanByMode(parseLegacyIDOrZero(patientID), "")
}

func parseLegacyIDOrZero(raw string) modeltypes.LegacyID {
	id, err := parseLegacyID(raw)
	if err != nil {
		return 0
	}
	return id
}

func (s *PrescriptionService) loadLegacyPrescriptionMaterials(prescriptionID int64) (models.MaterialList, error) {
	var rows []legacyPrescriptionMaterial
	if err := s.db.Where(`"PatientPrescriptionId" = ?`, prescriptionID).
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

func (s *PrescriptionService) syncLegacyPrescriptionMaterials(tx *gorm.DB, prescriptionID int64, materials models.MaterialList) error {
	if err := tx.Where(`"PatientPrescriptionId" = ?`, prescriptionID).Delete(&legacyPrescriptionMaterial{}).Error; err != nil {
		return err
	}

	planService := &PatientService{db: tx}
	now := time.Now()
	for idx, material := range materials {
		materialID, err := planService.findLegacyMaterialID(material)
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
			"Id":                    id,
			"TenantId":              LegacyTenantID,
			"PatientPrescriptionId": prescriptionID,
			"MaterialId":            materialID,
			"MaterialGroup":         idx + 1,
			"Num":                   material.Count,
			"Note":                  strings.TrimSpace(material.Note),
			"LastModifyTime":        now,
			"ChargeItemId":          0,
		}
		if err := tx.Table(`"Plan_PatientPrescriptionMaterial"`).Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *PrescriptionService) toPrescriptionDTO(item legacyPatientPrescription) (*models.Prescription, error) {
	planService := &PatientService{db: s.db}
	drugNames, err := planService.loadLegacyDrugNames(item.FirstAnticoagulant, item.MaintainAnticoagulant)
	if err != nil {
		return nil, err
	}
	materials, err := s.loadLegacyPrescriptionMaterials(item.ID)
	if err != nil {
		return nil, err
	}

	notePayload, noteIsJSON := parseLegacyPrescriptionNote(item.Note)
	orderItems := models.PrescriptionOrderItemList{}
	notes := strings.TrimSpace(item.Note)
	extraWeight := item.AdjustQuantity
	prescriptionDate := item.CreateTime
	if treatmentDate := s.loadLegacyPrescriptionTreatmentDate(item.TreatmentID); treatmentDate != nil {
		prescriptionDate = *treatmentDate
	}
	doctorName, err := s.lookupLegacyUserDisplayName(item.CreatorID)
	if err != nil {
		return nil, err
	}
	if noteIsJSON && notePayload != nil {
		orderItems = notePayload.OrderItems
		if strings.TrimSpace(notePayload.Notes) != "" {
			notes = strings.TrimSpace(notePayload.Notes)
		} else {
			notes = ""
		}
		if notePayload.ExtraWeight != 0 {
			extraWeight = notePayload.ExtraWeight
		}
		if notePayload.PrescriptionDate != "" && item.TreatmentID <= 0 {
			if parsed, parseErr := time.Parse("2006-01-02", notePayload.PrescriptionDate); parseErr == nil {
				prescriptionDate = parsed
			}
		}
		if strings.TrimSpace(notePayload.DoctorName) != "" {
			doctorName = strings.TrimSpace(notePayload.DoctorName)
		}
	}

	status := mapLegacyPrescriptionStatus(item.Status)
	var executedAt *time.Time
	if item.ConfirmTime != nil && !item.ConfirmTime.IsZero() {
		executedAt = item.ConfirmTime
	}

	var executedBy *string
	if item.ConfirmUserID > 0 {
		value := strconv.FormatInt(item.ConfirmUserID, 10)
		executedBy = &value
	}

	dto := &models.Prescription{
		ID:               strconv.FormatInt(item.ID, 10),
		PatientID:        item.PatientID,
		TreatmentPlanID:  strconv.FormatInt(item.PatientPlanID, 10),
		TreatmentID:      item.TreatmentID,
		PrescriptionDate: prescriptionDate,
		DoctorID:         strconv.FormatInt(item.CreatorID, 10),
		DoctorName:       doctorName,
		Status:           status,
		Duration:         int(item.DialysisDuration),
		DryWeight:        item.DryWeight,
		ExtraWeight:      extraWeight,
		DialysisMode: models.DialysisMode{
			Mode:                normalizeLegacyDialysisMode(item.DialysisMethod),
			BloodFlow:           int(item.BF),
			SubstituteInputMode: strings.TrimSpace(item.DilutionMnt),
			SubstituteFlow:      item.SubstituateFlow,
			SubstituteVolume:    item.SubstituateVolume,
			BV:                  formatLegacyNumber(item.BV),
			Status:              status,
		},
		Anticoagulant: models.Anticoagulant{
			InitialDrug:     legacyDrugNameByID(drugNames, item.FirstAnticoagulant),
			InitialDose:     formatLegacyNumber(item.FirstDosage),
			MaintenanceDrug: legacyDrugNameByID(drugNames, item.MaintainAnticoagulant),
			InfusionRate:    formatLegacyNumber(item.InjectionRate),
			InfusionTime:    formatLegacyNumber(item.InjectionDuration),
			MaintenanceDose: formatLegacyNumber(item.DilutionProportion),
			TotalDose:       formatLegacyNumber(item.InjectionVolume),
		},
		Parameters: models.DialysisParameters{
			DialysateType:  strings.TrimSpace(item.Dialysate),
			DialysateGroup: strconv.FormatInt(item.DialysateGroupID, 10),
			FlowRate:       int(item.DialysateFlow),
			Na:             item.NaIonCon,
			Ca:             item.CaIonCon,
			K:              item.KIonCon,
			HCO3:           item.HCO3IonCon,
			Glucose:        formatLegacyNumber(item.GlucoseCon),
			Conductivity:   item.Conductivity,
			Temp:           item.DialysateTmp,
			Volume:         item.DialysateVolume,
		},
		Materials:  materials,
		OrderItems: orderItems,
		Notes:      notes,
		ExecutedAt: executedAt,
		ExecutedBy: executedBy,
		CreatedAt:  item.CreateTime,
		UpdatedAt:  item.LastModifyTime,
	}

	return dto, nil
}

func (s *PrescriptionService) loadLegacyPrescription(patientID, prescriptionID string) (*legacyPatientPrescription, error) {
	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}
	id := parseLegacyNumericID(prescriptionID)
	if id <= 0 {
		return nil, errors.New("prescription not found")
	}

	var item legacyPatientPrescription
	err = s.db.Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, id, legacyPatientID, LegacyTenantID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("prescription not found")
		}
		return nil, err
	}
	return &item, nil
}

func (s *PrescriptionService) List(patientID string) ([]models.Prescription, error) {
	return s.LegacyList(patientID)
}

func (s *PrescriptionService) Get(patientID, prescriptionID string) (*models.Prescription, error) {
	return s.LegacyGet(patientID, prescriptionID)
}

func (s *PrescriptionService) Create(patientID, doctorID, doctorName string, req PrescriptionCreateRequest) (*models.Prescription, error) {
	return s.LegacyCreate(patientID, doctorID, doctorName, req)
}

func (s *PrescriptionService) Update(patientID, prescriptionID string, req PrescriptionUpdateRequest) (*models.Prescription, error) {
	return s.LegacyUpdate(patientID, prescriptionID, req)
}

func (s *PrescriptionService) Execute(patientID, prescriptionID, executedBy string) (*models.Prescription, error) {
	return s.LegacyExecute(patientID, prescriptionID, executedBy)
}

func (s *PrescriptionService) Cancel(patientID, prescriptionID string) (*models.Prescription, error) {
	return s.LegacyCancel(patientID, prescriptionID)
}

func (s *PrescriptionService) ExtractFromLongTermOrders(patientID, doctorID, doctorName, dateStr string) (*models.Prescription, error) {
	return s.LegacyExtractFromLongTermOrders(patientID, doctorID, doctorName, dateStr)
}

func (s *PrescriptionService) LegacyList(patientID string) ([]models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}

	var rows []legacyPatientPrescription
	if err := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, legacyPatientID, LegacyTenantID).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]models.Prescription, 0, len(rows))
	for _, row := range rows {
		dto, convErr := s.toPrescriptionDTO(row)
		if convErr != nil {
			return nil, convErr
		}
		result = append(result, *dto)
	}
	return result, nil
}

// DayPrescriptionStatus 当日某患者的处方开单/签发状态（驾驶舱医生墙用）。
type DayPrescriptionStatus struct {
	PatientID       string `json:"patientId"`
	HasPrescription bool   `json:"hasPrescription"`
	Signed          bool   `json:"signed"` // 已签 = 老库 ConfirmTime 非空（契约02：已确认=已签）
	PrescriptionID  string `json:"prescriptionId,omitempty"`
}

// DayStatus 批量返回指定日期「有处方的患者」的开方/签发状态，供驾驶舱医生墙。
// 有关联治疗记录时按治疗业务日期匹配；无关联治疗记录时回退到处方 CreateTime，避免漏掉手工开方。
func (s *PrescriptionService) DayStatus(date time.Time) ([]DayPrescriptionStatus, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	var rows []legacyPatientPrescription
	if err := s.db.Table(`"Plan_PatientPrescription" AS p`).
		Select(`p.*`).
		Joins(`LEFT JOIN "Treatment_Treatment" AS t ON t."Id" = p."TreatmentId" AND t."TenantId" = p."TenantId"`).
		Where(`p."TenantId" = ? AND (
			(t."Id" IS NOT NULL AND DATE(COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime")) = DATE(?))
			OR (t."Id" IS NULL AND p."CreateTime" >= ? AND p."CreateTime" < ?)
		)`, LegacyTenantID, dayStart, dayStart, dayEnd).
		Order(`COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime", p."CreateTime") DESC`).
		Order(`p."Id" DESC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(rows))
	result := make([]DayPrescriptionStatus, 0, len(rows))
	for _, row := range rows {
		pid := strconv.FormatInt(int64(row.PatientID), 10)
		if _, ok := seen[pid]; ok {
			continue // 同患者只取当日最新一条
		}
		seen[pid] = struct{}{}
		result = append(result, DayPrescriptionStatus{
			PatientID:       pid,
			HasPrescription: true,
			Signed:          row.ConfirmTime != nil && !row.ConfirmTime.IsZero(),
			PrescriptionID:  strconv.FormatInt(row.ID, 10),
		})
	}
	return result, nil
}

func (s *PrescriptionService) LegacyGet(patientID, prescriptionID string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	item, err := s.loadLegacyPrescription(patientID, prescriptionID)
	if err != nil {
		return nil, err
	}
	return s.toPrescriptionDTO(*item)
}

func (s *PrescriptionService) LegacyCreate(patientID, doctorID, doctorName string, req PrescriptionCreateRequest) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}
	if strings.TrimSpace(req.PrescriptionDate) == "" {
		return nil, errors.New("prescriptionDate is required")
	}
	prescriptionDate, err := time.Parse("2006-01-02", req.PrescriptionDate)
	if err != nil {
		return nil, errors.New("日期格式错误，应为 yyyy-MM-dd")
	}

	resolvedUserID, resolvedDoctorName, err := s.resolveLegacyUserID(doctorID, doctorName)
	if err != nil {
		return nil, err
	}

	plan, err := s.getLegacyPlanForPrescription(patientID, req.DialysisMode.Mode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("请先创建启用的治疗方案")
		}
		return nil, err
	}
	// 方案完整性门禁（契约02 三）：草稿态方案（透析液配方为空）不得开当日处方，
	// 提示医生先到治疗方案页补全。建档时生成的草稿方案在此被拦截。
	if !isLegacyPlanComplete(plan) {
		return nil, errors.New("治疗方案尚未补全（透析液配方为空），请先到治疗方案页完善方案后再开当日处方")
	}

	firstDrugID, err := (&PatientService{db: s.db}).findLegacyDrugIDByName(req.Anticoagulant.InitialDrug)
	if err != nil {
		return nil, err
	}
	maintainDrugID, err := (&PatientService{db: s.db}).findLegacyDrugIDByName(req.Anticoagulant.MaintenanceDrug)
	if err != nil {
		return nil, err
	}

	prescriptionIDValue, err := idgen.NextID()
	if err != nil {
		return nil, err
	}

	duration := req.Duration
	if duration == 0 {
		duration = int(plan.DialysisDuration)
	}
	dryWeight := req.DryWeight
	if dryWeight == 0 {
		dryWeight = plan.DryWeight
	}
	mode := req.DialysisMode
	if strings.TrimSpace(mode.Mode) == "" {
		mode.Mode = normalizeLegacyDialysisMode(plan.DialysisMethod)
		mode.BloodFlow = int(plan.BF)
		mode.SubstituteInputMode = plan.DilutionMnt
		mode.SubstituteFlow = plan.SubstituateFlow
		mode.SubstituteVolume = plan.SubstituateVolume
		mode.BV = formatLegacyNumber(plan.BV)
	}
	anticoagulant := req.Anticoagulant
	if anticoagulant.InitialDrug == "" {
		drugNames, loadErr := (&PatientService{db: s.db}).loadLegacyDrugNames(plan.FirstAnticoagulant, plan.MaintainAnticoagulant)
		if loadErr != nil {
			return nil, loadErr
		}
		anticoagulant.InitialDrug = legacyDrugNameByID(drugNames, plan.FirstAnticoagulant)
		anticoagulant.InitialDose = formatLegacyNumber(plan.FirstDosage)
		anticoagulant.MaintenanceDrug = legacyDrugNameByID(drugNames, plan.MaintainAnticoagulant)
		anticoagulant.InfusionRate = formatLegacyNumber(plan.InjectionRate)
		anticoagulant.InfusionTime = formatLegacyNumber(plan.InjectionDuration)
		anticoagulant.MaintenanceDose = formatLegacyNumber(plan.DilutionProportion)
		anticoagulant.TotalDose = formatLegacyNumber(plan.InjectionVolume)
	}
	parameters := req.Parameters
	if strings.TrimSpace(parameters.DialysateType) == "" {
		parameters = models.DialysisParameters{
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
		}
	}
	materials := req.Materials
	if len(materials) == 0 {
		if copied, loadErr := (&PatientService{db: s.db}).loadLegacyPlanMaterials(plan.ID); loadErr == nil {
			materials = copied
		}
	}

	row := map[string]any{
		"Id":                           prescriptionIDValue,
		"TenantId":                     LegacyTenantID,
		"PatientId":                    legacyPatientID,
		"TreatmentId":                  0,
		"PatientPlanId":                plan.ID,
		"CreatorId":                    resolvedUserID,
		"CreateTime":                   prescriptionDate,
		"ConfirmUserId":                0,
		"Status":                       mapNewPrescriptionStatus(models.PrescriptionStatusPending),
		"CaseStatus":                   "",
		"DialysisMethod":               normalizeLegacyDialysisMode(mode.Mode),
		"DialysisDuration":             duration,
		"DryWeight":                    dryWeight,
		"AdjustQuantity":               req.ExtraWeight,
		"BF":                           mode.BloodFlow,
		"BV":                           parseStringFloat(mode.BV),
		"FirstAnticoagulant":           firstDrugID,
		"FirstDosage":                  parseStringFloat(anticoagulant.InitialDose),
		"MaintainAnticoagulant":        maintainDrugID,
		"DilutionProportion":           parseStringFloat(anticoagulant.MaintenanceDose),
		"InjectionRate":                parseStringFloat(anticoagulant.InfusionRate),
		"InjectionDuration":            parseStringFloat(anticoagulant.InfusionTime),
		"InjectionVolume":              parseStringFloat(anticoagulant.TotalDose),
		"VascularAccessId":             plan.VascularAccessID,
		"Dialysate":                    strings.TrimSpace(parameters.DialysateType),
		"DialysateFlow":                parameters.FlowRate,
		"DialysateVolume":              parameters.Volume,
		"NaIonCon":                     parameters.Na,
		"CaIonCon":                     parameters.Ca,
		"KIonCon":                      parameters.K,
		"HCO3IonCon":                   parameters.HCO3,
		"Conductivity":                 parameters.Conductivity,
		"DialysateTmp":                 parameters.Temp,
		"SubstituateVolume":            mode.SubstituteVolume,
		"DilutionMnt":                  strings.TrimSpace(mode.SubstituteInputMode),
		"LastModifyTime":               time.Now(),
		"SalineQuantity":               plan.SalineQuantity,
		"SealQuantity":                 plan.SealQuantity,
		"ArterialQuantity":             plan.ArterialQuantity,
		"VenousQuantity":               plan.VenousQuantity,
		"UFQuantity":                   plan.ExtraWeight,
		"SealType":                     strings.TrimSpace(plan.SealType),
		"GlucoseCon":                   parseStringFloat(parameters.Glucose),
		"DialysateGroupId":             parseLegacyNumericID(parameters.DialysateGroup),
		"Note":                         buildLegacyPrescriptionNote(req.Notes, req.ExtraWeight, prescriptionDate, resolvedDoctorName, req.OrderItems),
		"SubstituateFlow":              mode.SubstituteFlow,
		"IsInduceDialysisPrescription": false,
		"HeparinType":                  0,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(`"Plan_PatientPrescription"`).Create(row).Error; err != nil {
			return err
		}
		return s.syncLegacyPrescriptionMaterials(tx, prescriptionIDValue, materials)
	}); err != nil {
		return nil, err
	}

	return s.LegacyGet(patientID, strconv.FormatInt(prescriptionIDValue, 10))
}

func (s *PrescriptionService) LegacyUpdate(patientID, prescriptionID string, req PrescriptionUpdateRequest) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	item, err := s.loadLegacyPrescription(patientID, prescriptionID)
	if err != nil {
		return nil, err
	}
	if mapLegacyPrescriptionStatus(item.Status) != models.PrescriptionStatusPending {
		return nil, errors.New("仅待执行状态的处方可编辑")
	}

	notePayload, noteIsJSON := parseLegacyPrescriptionNote(item.Note)
	if !noteIsJSON || notePayload == nil {
		notePayload = &legacyPrescriptionNote{Notes: strings.TrimSpace(item.Note), ExtraWeight: item.AdjustQuantity}
	}

	updates := map[string]any{
		"LastModifyTime": time.Now(),
	}
	if req.Duration != nil {
		updates["DialysisDuration"] = *req.Duration
	}
	if req.DryWeight != nil {
		updates["DryWeight"] = *req.DryWeight
	}
	if req.ExtraWeight != nil {
		updates["AdjustQuantity"] = *req.ExtraWeight
		notePayload.ExtraWeight = *req.ExtraWeight
	}
	if req.DialysisMode != nil {
		updates["DialysisMethod"] = normalizeLegacyDialysisMode(req.DialysisMode.Mode)
		updates["BF"] = req.DialysisMode.BloodFlow
		updates["BV"] = parseStringFloat(req.DialysisMode.BV)
		updates["DilutionMnt"] = strings.TrimSpace(req.DialysisMode.SubstituteInputMode)
		updates["SubstituateFlow"] = req.DialysisMode.SubstituteFlow
		updates["SubstituateVolume"] = req.DialysisMode.SubstituteVolume
		if plan, planErr := s.getLegacyPlanForPrescription(patientID, req.DialysisMode.Mode); planErr == nil && plan != nil {
			updates["PatientPlanId"] = plan.ID
		}
	}
	if req.Anticoagulant != nil {
		firstDrugID, findErr := (&PatientService{db: s.db}).findLegacyDrugIDByName(req.Anticoagulant.InitialDrug)
		if findErr != nil {
			return nil, findErr
		}
		maintainDrugID, findErr := (&PatientService{db: s.db}).findLegacyDrugIDByName(req.Anticoagulant.MaintenanceDrug)
		if findErr != nil {
			return nil, findErr
		}
		updates["FirstAnticoagulant"] = firstDrugID
		updates["FirstDosage"] = parseStringFloat(req.Anticoagulant.InitialDose)
		updates["MaintainAnticoagulant"] = maintainDrugID
		updates["DilutionProportion"] = parseStringFloat(req.Anticoagulant.MaintenanceDose)
		updates["InjectionRate"] = parseStringFloat(req.Anticoagulant.InfusionRate)
		updates["InjectionDuration"] = parseStringFloat(req.Anticoagulant.InfusionTime)
		updates["InjectionVolume"] = parseStringFloat(req.Anticoagulant.TotalDose)
	}
	if req.Parameters != nil {
		updates["Dialysate"] = strings.TrimSpace(req.Parameters.DialysateType)
		updates["DialysateFlow"] = req.Parameters.FlowRate
		updates["DialysateVolume"] = req.Parameters.Volume
		updates["NaIonCon"] = req.Parameters.Na
		updates["CaIonCon"] = req.Parameters.Ca
		updates["KIonCon"] = req.Parameters.K
		updates["HCO3IonCon"] = req.Parameters.HCO3
		updates["Conductivity"] = req.Parameters.Conductivity
		updates["DialysateTmp"] = req.Parameters.Temp
		updates["GlucoseCon"] = parseStringFloat(req.Parameters.Glucose)
		if groupID := parseLegacyNumericID(req.Parameters.DialysateGroup); groupID > 0 {
			updates["DialysateGroupId"] = groupID
		}
	}
	if req.Notes != nil {
		notePayload.Notes = strings.TrimSpace(*req.Notes)
	}
	if req.OrderItems != nil {
		notePayload.OrderItems = *req.OrderItems
	}
	prescriptionDate := item.CreateTime
	if notePayload.PrescriptionDate != "" {
		if parsed, parseErr := time.Parse("2006-01-02", notePayload.PrescriptionDate); parseErr == nil {
			prescriptionDate = parsed
		}
	}
	doctorName, lookupErr := s.lookupLegacyUserDisplayName(item.CreatorID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if notePayload.DoctorName != "" {
		doctorName = notePayload.DoctorName
	}
	updates["Note"] = buildLegacyPrescriptionNote(notePayload.Notes, notePayload.ExtraWeight, prescriptionDate, doctorName, notePayload.OrderItems)

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(`"Plan_PatientPrescription"`).Where(`"Id" = ? AND "TenantId" = ?`, item.ID, LegacyTenantID).Updates(updates).Error; err != nil {
			return err
		}
		if req.Materials != nil {
			return s.syncLegacyPrescriptionMaterials(tx, item.ID, *req.Materials)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.LegacyGet(patientID, prescriptionID)
}

// LegacySign 签发当日处方（待签 → 已签，契约02 待签线）：
// 落 ConfirmTime/ConfirmUserId 作"已签"信号（**不改执行态**，护士仍可后续执行），并写 sign_record 统一留痕。
// 已签幂等（重复签不重复落 ConfirmTime，但仍补一条留痕便于审计）。
func (s *PrescriptionService) LegacySign(patientID, prescriptionID, signerID, signerName string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	item, err := s.loadLegacyPrescription(patientID, prescriptionID)
	if err != nil {
		return nil, err
	}
	if mapLegacyPrescriptionStatus(item.Status) == models.PrescriptionStatusCancelled {
		return nil, errors.New("已取消的处方不能签发")
	}

	resolvedUserID, resolvedName, err := s.resolveLegacyUserID(signerID, signerName)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if item.ConfirmTime == nil || item.ConfirmTime.IsZero() {
			if err := tx.Table(`"Plan_PatientPrescription"`).
				Where(`"Id" = ? AND "TenantId" = ?`, item.ID, LegacyTenantID).
				Updates(map[string]any{
					"ConfirmUserId":  resolvedUserID,
					"ConfirmTime":    now,
					"LastModifyTime": now,
				}).Error; err != nil {
				return err
			}
		}
		signService := &SignService{db: tx}
		_, err := signService.Sign(LegacyTenantID, models.SignTargetPrescription,
			strconv.FormatInt(item.ID, 10), strconv.FormatInt(resolvedUserID, 10), resolvedName)
		return err
	}); err != nil {
		return nil, err
	}

	return s.LegacyGet(patientID, prescriptionID)
}

func (s *PrescriptionService) LegacyExecute(patientID, prescriptionID, executedBy string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	item, err := s.loadLegacyPrescription(patientID, prescriptionID)
	if err != nil {
		return nil, err
	}

	status := mapLegacyPrescriptionStatus(item.Status)
	if status == models.PrescriptionStatusExecuted {
		return s.toPrescriptionDTO(*item)
	}
	if status == models.PrescriptionStatusCancelled {
		return nil, errors.New("已取消的处方不能执行")
	}

	resolvedUserID, _, err := s.resolveLegacyUserID(executedBy, "")
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if err := s.db.Table(`"Plan_PatientPrescription"`).
		Where(`"Id" = ? AND "TenantId" = ?`, item.ID, LegacyTenantID).
		Updates(map[string]any{
			"Status":         mapNewPrescriptionStatus(models.PrescriptionStatusExecuted),
			"ConfirmUserId":  resolvedUserID,
			"ConfirmTime":    now,
			"LastModifyTime": now,
		}).Error; err != nil {
		return nil, err
	}

	return s.LegacyGet(patientID, prescriptionID)
}

func (s *PrescriptionService) LegacyCancel(patientID, prescriptionID string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	item, err := s.loadLegacyPrescription(patientID, prescriptionID)
	if err != nil {
		return nil, err
	}

	status := mapLegacyPrescriptionStatus(item.Status)
	if status == models.PrescriptionStatusCancelled {
		return s.toPrescriptionDTO(*item)
	}
	if status == models.PrescriptionStatusExecuted {
		return nil, errors.New("已执行的处方不能取消")
	}

	if err := s.db.Table(`"Plan_PatientPrescription"`).
		Where(`"Id" = ? AND "TenantId" = ?`, item.ID, LegacyTenantID).
		Updates(map[string]any{
			"Status":         mapNewPrescriptionStatus(models.PrescriptionStatusCancelled),
			"LastModifyTime": time.Now(),
		}).Error; err != nil {
		return nil, err
	}
	return s.LegacyGet(patientID, prescriptionID)
}

func (s *PrescriptionService) LegacyExtractFromLongTermOrders(patientID, doctorID, doctorName, dateStr string) (*models.Prescription, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("日期格式错误，应为 yyyy-MM-dd")
	}

	plan, err := s.getLegacyPlanForPrescription(patientID, "")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("请先创建启用的治疗方案")
		}
		return nil, err
	}

	orders, err := (&OrderService{db: s.db}).listLegacyOrders(OrderListRequest{
		PatientID: patientID,
		TenantID:  LegacyTenantID,
		Type:      models.OrderTypeLongTerm,
		Statuses:  strings.Join([]string{models.OrderStatusPending, models.OrderStatusExecuting}, ","),
	})
	if err != nil {
		return nil, err
	}

	orderItems := make(models.PrescriptionOrderItemList, 0, len(orders))
	for _, o := range orders {
		orderItems = append(orderItems, models.PrescriptionOrderItem{
			OrderID:   o.ID,
			Name:      o.Name,
			Category:  o.Category,
			Dose:      o.Dose,
			Unit:      o.Unit,
			Frequency: firstNonEmptyText(valueOrEmpty(o.Frequency)),
			Route:     o.Route,
			Spec:      o.Spec,
		})
	}

	req := PrescriptionCreateRequest{
		PrescriptionDate: date.Format("2006-01-02"),
		Duration:         int(plan.DialysisDuration),
		DryWeight:        plan.DryWeight,
		ExtraWeight:      plan.ExtraWeight,
		DialysisMode: models.DialysisMode{
			Mode:                normalizeLegacyDialysisMode(plan.DialysisMethod),
			BloodFlow:           int(plan.BF),
			SubstituteInputMode: plan.DilutionMnt,
			SubstituteFlow:      plan.SubstituateFlow,
			SubstituteVolume:    plan.SubstituateVolume,
			BV:                  formatLegacyNumber(plan.BV),
		},
		Anticoagulant: models.Anticoagulant{
			InitialDose:     formatLegacyNumber(plan.FirstDosage),
			MaintenanceDose: formatLegacyNumber(plan.DilutionProportion),
			InfusionRate:    formatLegacyNumber(plan.InjectionRate),
			InfusionTime:    formatLegacyNumber(plan.InjectionDuration),
			TotalDose:       formatLegacyNumber(plan.InjectionVolume),
		},
		Parameters: models.DialysisParameters{
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
		OrderItems: orderItems,
	}

	drugNames, err := (&PatientService{db: s.db}).loadLegacyDrugNames(plan.FirstAnticoagulant, plan.MaintainAnticoagulant)
	if err != nil {
		return nil, err
	}
	req.Anticoagulant.InitialDrug = legacyDrugNameByID(drugNames, plan.FirstAnticoagulant)
	req.Anticoagulant.MaintenanceDrug = legacyDrugNameByID(drugNames, plan.MaintainAnticoagulant)

	if copied, loadErr := (&PatientService{db: s.db}).loadLegacyPlanMaterials(plan.ID); loadErr == nil {
		req.Materials = copied
	}

	return s.LegacyCreate(patientID, doctorID, doctorName, req)
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}
