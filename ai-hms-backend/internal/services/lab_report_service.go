package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

// LabReportService 检验报告服务
type LabReportService struct {
	db *gorm.DB
}

type legacyLabExamination struct {
	ID             int64     `gorm:"column:Id"`
	TenantID       int64     `gorm:"column:TenantId"`
	PatientID      int64     `gorm:"column:PatientId"`
	Name           string    `gorm:"column:Name"`
	Type           string    `gorm:"column:Type"`
	ResultTime     *time.Time `gorm:"column:ResultTime"`
	SyncUserID     int64     `gorm:"column:SyncUserId"`
	SyncTime       *time.Time `gorm:"column:SyncTime"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime"`
	TestNO         string    `gorm:"column:TestNO"`
}

func (legacyLabExamination) TableName() string { return "LIS_Examination" }

type legacyLabExaminationApply struct {
	TestNO                string     `gorm:"column:TestNO"`
	ClinicalDiagnosisDesc string     `gorm:"column:ClinicalDiagnosisDesc"`
	Specimen              string     `gorm:"column:Specimen"`
	SpecimenDesc          string     `gorm:"column:SpecimenDesc"`
	SpecimenReceivedTime  *time.Time `gorm:"column:SpecimenReceivedTime"`
	SpecimenSampleTime    *time.Time `gorm:"column:SpecimenSampleTime"`
	ApplyTime             *time.Time `gorm:"column:ApplyTime"`
	ApplyUserName         string     `gorm:"column:ApplyUserName"`
	Priority              int        `gorm:"column:Priority"`
	ResultStatus          int        `gorm:"column:ResultStatus"`
	ResultRPTTime         *time.Time `gorm:"column:ResultRPTTime"`
}

func (legacyLabExaminationApply) TableName() string { return "LIS_ExaminationApply" }

type legacyLabExaminationItem struct {
	ID             int64     `gorm:"column:Id"`
	TenantID       int64     `gorm:"column:TenantId"`
	ExaminationID  int64     `gorm:"column:ExaminationId"`
	ItemName       string    `gorm:"column:ItemName"`
	ItemCode       string    `gorm:"column:ItemCode"`
	Result         string    `gorm:"column:Result"`
	Unit           string    `gorm:"column:Unit"`
	Reference      string    `gorm:"column:Reference"`
	ResultSign     string    `gorm:"column:ResultSign"`
	LastModifyTime time.Time `gorm:"column:LastModifyTime"`
}

func (legacyLabExaminationItem) TableName() string { return "LIS_ExaminationItem" }

// NewLabReportService 创建检验报告服务
func NewLabReportService() *LabReportService {
	return &LabReportService{
		db: database.GetDB(),
	}
}

// LabReportItemRequest 检验明细请求
type LabReportItemRequest struct {
	ItemCode       string `json:"itemCode" binding:"required"`
	ItemName       string `json:"itemName" binding:"required"`
	ResultValue    string `json:"resultValue" binding:"required"`
	Unit           string `json:"unit"`
	ReferenceRange string `json:"referenceRange"`
	AbnormalFlag   string `json:"abnormalFlag"` // H/L/N
	TestedAt       string `json:"testedAt"`     // RFC3339 / YYYY-MM-DD
}

// LabReportCreateRequest 创建检验报告请求
type LabReportCreateRequest struct {
	ReportNo          string `json:"reportNo"`
	ItemCode          string `json:"itemCode"`
	ItemName          string `json:"itemName" binding:"required"`
	ClinicalDiagnosis string `json:"clinicalDiagnosis"`
	SpecimenType      string `json:"specimenType"`
	Urgency           string `json:"urgency"`
	RequestDoctor     string `json:"requestDoctor"`
	RequestedAt       string `json:"requestedAt"`
	SampledAt         string `json:"sampledAt"`
	ReceivedAt        string `json:"receivedAt"`
	ReportedAt        string `json:"reportedAt"`
	Status            string `json:"status"`

	ExternalReportID *string `json:"externalReportId"`
	SourceSystem     string  `json:"sourceSystem"` // LOCAL/LIS/PACS
	SyncedAt         string  `json:"syncedAt"`

	Items []LabReportItemRequest `json:"items" binding:"required,min=1"`
}

// LabReportListRequest 检验报告列表请求
type LabReportListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	StartDate string `form:"startDate"` // 过滤 reportedAt >= startDate
	EndDate   string `form:"endDate"`   // 过滤 reportedAt <= endDate
}

// LabReportListResponse 检验报告列表响应
type LabReportListResponse struct {
	Items     []models.LabReport `json:"items"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"pageSize"`
	TotalPage int                `json:"totalPage"`
}

// LabReportUpdateRequest 更新检验报告请求
type LabReportUpdateRequest struct {
	ReportNo          string  `json:"reportNo"`
	ItemCode          string  `json:"itemCode"`
	ItemName          string  `json:"itemName"`
	ClinicalDiagnosis string  `json:"clinicalDiagnosis"`
	SpecimenType      string  `json:"specimenType"`
	Urgency           string  `json:"urgency"`
	RequestDoctor     string  `json:"requestDoctor"`
	RequestedAt       *string `json:"requestedAt"`
	SampledAt         *string `json:"sampledAt"`
	ReceivedAt        *string `json:"receivedAt"`
	ReportedAt        *string `json:"reportedAt"`
	Status            string  `json:"status"`

	ExternalReportID *string `json:"externalReportId"`
	SourceSystem     string  `json:"sourceSystem"`
	SyncedAt         *string `json:"syncedAt"`

	Items []LabReportItemRequest `json:"items"`
}

// ListByPatient 按患者查询检验报告
func (s *LabReportService) ListByPatient(patientID string, req LabReportListRequest) (*LabReportListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(patientID) == "" {
		return nil, errors.New("patient id is required")
	}

	page, pageSize := normalizePagination(req.Page, req.PageSize)
	return s.listLegacyByPatient(patientID, req, page, pageSize)
}

func (s *LabReportService) listLegacyByPatient(patientID string, req LabReportListRequest, page, pageSize int) (*LabReportListResponse, error) {
	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("patient id is required")
	}

	query := s.db.Model(&legacyLabExamination{}).
		Where(`"PatientId" = ? AND "TenantId" = ?`, legacyPatientID, legacyTenantID)

	if strings.TrimSpace(req.StartDate) != "" {
		startDate, err := parseOptionalTime(req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid startDate: %w", err)
		}
		query = query.Where(`COALESCE("ResultTime", "LastModifyTime") >= ?`, *startDate)
	}
	if strings.TrimSpace(req.EndDate) != "" {
		endDate, err := parseOptionalTime(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid endDate: %w", err)
		}
		query = query.Where(`COALESCE("ResultTime", "LastModifyTime") <= ?`, *endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var exams []legacyLabExamination
	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order(`COALESCE("ResultTime", "LastModifyTime") DESC`).
		Order(`"Id" DESC`).
		Find(&exams).Error; err != nil {
		return nil, err
	}

	examIDs := make([]int64, 0, len(exams))
	testNos := make([]string, 0, len(exams))
	for _, exam := range exams {
		examIDs = append(examIDs, exam.ID)
		if strings.TrimSpace(exam.TestNO) != "" {
			testNos = append(testNos, strings.TrimSpace(exam.TestNO))
		}
	}

	itemMap := make(map[int64][]models.LabReportItem)
	if len(examIDs) > 0 {
		var rows []legacyLabExaminationItem
		if err := s.db.Where(`"ExaminationId" IN ?`, examIDs).
			Order(`"ExaminationId" ASC`).
			Order(`"Id" ASC`).
			Find(&rows).Error; err != nil && !isIgnorableLegacyQueryError(err) {
			return nil, err
		} else if err == nil {
			for _, row := range rows {
				itemMap[row.ExaminationID] = append(itemMap[row.ExaminationID], models.LabReportItem{
					ID:             strconv.FormatInt(row.ID, 10),
					LabReportID:    strconv.FormatInt(row.ExaminationID, 10),
					ItemCode:       strings.TrimSpace(row.ItemCode),
					ItemName:       strings.TrimSpace(row.ItemName),
					ResultValue:    strings.TrimSpace(row.Result),
					Unit:           strings.TrimSpace(row.Unit),
					ReferenceRange: strings.TrimSpace(row.Reference),
					AbnormalFlag:   strings.TrimSpace(row.ResultSign),
					TestedAt:       timePtrOrNil(row.LastModifyTime),
					CreatedAt:      row.LastModifyTime,
					UpdatedAt:      row.LastModifyTime,
				})
			}
		}
	}

	applyMap := make(map[string]legacyLabExaminationApply)
	if len(testNos) > 0 {
		var applyRows []legacyLabExaminationApply
		if err := s.db.Where(`"TestNO" IN ?`, testNos).
			Order(`"LastModifyTime" DESC`).
			Find(&applyRows).Error; err == nil {
			for _, row := range applyRows {
				if _, ok := applyMap[strings.TrimSpace(row.TestNO)]; !ok {
					applyMap[strings.TrimSpace(row.TestNO)] = row
				}
			}
		}
	}

	reports := make([]models.LabReport, 0, len(exams))
	for _, exam := range exams {
		apply := applyMap[strings.TrimSpace(exam.TestNO)]
		reportedAt := firstTimePtr(exam.ResultTime, apply.ResultRPTTime, apply.ApplyTime, apply.SpecimenReceivedTime)
		report := models.LabReport{
			ID:                strconv.FormatInt(exam.ID, 10),
			PatientID:         patientID,
			ReportNo:          strings.TrimSpace(exam.TestNO),
			ItemCode:          firstLabItemCode(itemMap[exam.ID]),
			ItemName:          firstNonEmptyText(strings.TrimSpace(exam.Name), firstLabItemName(itemMap[exam.ID])),
			ClinicalDiagnosis: strings.TrimSpace(apply.ClinicalDiagnosisDesc),
			SpecimenType:      firstNonEmptyText(strings.TrimSpace(apply.SpecimenDesc), strings.TrimSpace(apply.Specimen)),
			Urgency:           legacyPriorityText(apply.Priority),
			RequestDoctor:     strings.TrimSpace(apply.ApplyUserName),
			RequestedAt:       apply.ApplyTime,
			SampledAt:         apply.SpecimenSampleTime,
			ReceivedAt:        apply.SpecimenReceivedTime,
			ReportedAt:        reportedAt,
			Status:            legacyLabStatusText(apply.ResultStatus),
			ExternalReportID:  nil,
			SourceSystem:      models.SourceSystemLIS,
			SyncedAt:          exam.SyncTime,
			CreatedAt:         exam.LastModifyTime,
			UpdatedAt:         exam.LastModifyTime,
			Items:             itemMap[exam.ID],
		}
		reports = append(reports, report)
	}

	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return &LabReportListResponse{
		Items:     reports,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}, nil
}

func firstTimePtr(values ...*time.Time) *time.Time {
	for _, value := range values {
		if value != nil && !value.IsZero() {
			return value
		}
	}
	return nil
}

func timePtrOrNil(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	v := value
	return &v
}

func firstLabItemCode(items []models.LabReportItem) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].ItemCode
}

func firstLabItemName(items []models.LabReportItem) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].ItemName
}

func legacyPriorityText(priority int) string {
	switch priority {
	case 1:
		return "急诊"
	default:
		return "常规"
	}
}

func legacyLabStatusText(status int) string {
	switch status {
	case 1:
		return "已出报告"
	case 2:
		return "已审核"
	default:
		return "已出报告"
	}
}

// Create 创建检验报告
func (s *LabReportService) Create(patientID string, req LabReportCreateRequest) (*models.LabReport, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(patientID) == "" {
		return nil, errors.New("patient id is required")
	}
	if strings.TrimSpace(req.ItemName) == "" {
		return nil, errors.New("item name is required")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("at least one lab report item is required")
	}

	// 验证患者是否存在
	var count int64
	if err := s.db.Model(&models.Patient{}).Where("id = ?", patientID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	sourceSystem, err := normalizeSourceSystem(req.SourceSystem)
	if err != nil {
		return nil, err
	}

	reportedAt, err := parseOptionalTime(req.ReportedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid reportedAt: %w", err)
	}
	requestedAt, err := parseOptionalTime(req.RequestedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid requestedAt: %w", err)
	}
	sampledAt, err := parseOptionalTime(req.SampledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid sampledAt: %w", err)
	}
	receivedAt, err := parseOptionalTime(req.ReceivedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid receivedAt: %w", err)
	}
	syncedAt, err := parseOptionalTime(req.SyncedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid syncedAt: %w", err)
	}

	report := models.LabReport{
		ID:                utils.GenerateID(),
		PatientID:         patientID,
		ReportNo:          strings.TrimSpace(req.ReportNo),
		ItemCode:          strings.TrimSpace(req.ItemCode),
		ItemName:          strings.TrimSpace(req.ItemName),
		ClinicalDiagnosis: strings.TrimSpace(req.ClinicalDiagnosis),
		SpecimenType:      strings.TrimSpace(req.SpecimenType),
		Urgency:           strings.TrimSpace(req.Urgency),
		RequestDoctor:     strings.TrimSpace(req.RequestDoctor),
		RequestedAt:       requestedAt,
		SampledAt:         sampledAt,
		ReceivedAt:        receivedAt,
		ReportedAt:        reportedAt,
		Status:            strings.TrimSpace(req.Status),
		ExternalReportID:  normalizeOptionalString(req.ExternalReportID),
		SourceSystem:      sourceSystem,
		SyncedAt:          syncedAt,
	}

	items, err := buildLabReportItems(report.ID, req.Items)
	if err != nil {
		return nil, err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&report).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var created models.LabReport
	if err := s.db.Preload("Items").First(&created, "id = ?", report.ID).Error; err != nil {
		return nil, err
	}

	return &created, nil
}

// Update 更新检验报告
func (s *LabReportService) Update(reportID string, req LabReportUpdateRequest) (*models.LabReport, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(reportID) == "" {
		return nil, errors.New("report id is required")
	}

	var report models.LabReport
	if err := s.db.First(&report, "id = ?", reportID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("lab report not found")
		}
		return nil, err
	}

	sourceSystem := report.SourceSystem
	if strings.TrimSpace(req.SourceSystem) != "" {
		normalizedSource, err := normalizeSourceSystem(req.SourceSystem)
		if err != nil {
			return nil, err
		}
		sourceSystem = normalizedSource
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		// 标量字段更新
		if req.ReportNo != "" {
			report.ReportNo = strings.TrimSpace(req.ReportNo)
		}
		if req.ItemCode != "" {
			report.ItemCode = strings.TrimSpace(req.ItemCode)
		}
		if req.ItemName != "" {
			report.ItemName = strings.TrimSpace(req.ItemName)
		}
		if req.ClinicalDiagnosis != "" {
			report.ClinicalDiagnosis = strings.TrimSpace(req.ClinicalDiagnosis)
		}
		if req.SpecimenType != "" {
			report.SpecimenType = strings.TrimSpace(req.SpecimenType)
		}
		if req.Urgency != "" {
			report.Urgency = strings.TrimSpace(req.Urgency)
		}
		if req.RequestDoctor != "" {
			report.RequestDoctor = strings.TrimSpace(req.RequestDoctor)
		}
		if req.Status != "" {
			report.Status = strings.TrimSpace(req.Status)
		}
		if err := applyTimeFieldUpdate(req.RequestedAt, &report.RequestedAt); err != nil {
			return fmt.Errorf("invalid requestedAt: %w", err)
		}
		if err := applyTimeFieldUpdate(req.SampledAt, &report.SampledAt); err != nil {
			return fmt.Errorf("invalid sampledAt: %w", err)
		}
		if err := applyTimeFieldUpdate(req.ReceivedAt, &report.ReceivedAt); err != nil {
			return fmt.Errorf("invalid receivedAt: %w", err)
		}
		if err := applyTimeFieldUpdate(req.ReportedAt, &report.ReportedAt); err != nil {
			return fmt.Errorf("invalid reportedAt: %w", err)
		}
		if err := applyTimeFieldUpdate(req.SyncedAt, &report.SyncedAt); err != nil {
			return fmt.Errorf("invalid syncedAt: %w", err)
		}
		report.SourceSystem = sourceSystem
		if req.ExternalReportID != nil {
			report.ExternalReportID = normalizeOptionalString(req.ExternalReportID)
		}

		if err := tx.Save(&report).Error; err != nil {
			return err
		}

		// 传入 items 时按新数据替换明细
		if req.Items != nil {
			if len(req.Items) == 0 {
				return errors.New("items cannot be empty")
			}
			items, err := buildLabReportItems(report.ID, req.Items)
			if err != nil {
				return err
			}
			if err := tx.Where("lab_report_id = ?", report.ID).Delete(&models.LabReportItem{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var updated models.LabReport
	if err := s.db.Preload("Items").First(&updated, "id = ?", report.ID).Error; err != nil {
		return nil, err
	}

	return &updated, nil
}

// Delete 删除检验报告
func (s *LabReportService) Delete(reportID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	if strings.TrimSpace(reportID) == "" {
		return errors.New("report id is required")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("lab_report_id = ?", reportID).Delete(&models.LabReportItem{}).Error; err != nil {
			return err
		}

		result := tx.Where("id = ?", reportID).Delete(&models.LabReport{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("lab report not found")
		}
		return nil
	})
}

func buildLabReportItems(reportID string, reqItems []LabReportItemRequest) ([]models.LabReportItem, error) {
	if len(reqItems) == 0 {
		return nil, errors.New("at least one lab report item is required")
	}

	items := make([]models.LabReportItem, 0, len(reqItems))
	for i, itemReq := range reqItems {
		if strings.TrimSpace(itemReq.ItemCode) == "" {
			return nil, fmt.Errorf("items[%d].itemCode is required", i)
		}
		if strings.TrimSpace(itemReq.ItemName) == "" {
			return nil, fmt.Errorf("items[%d].itemName is required", i)
		}
		if strings.TrimSpace(itemReq.ResultValue) == "" {
			return nil, fmt.Errorf("items[%d].resultValue is required", i)
		}

		flag, err := normalizeAbnormalFlag(itemReq.AbnormalFlag)
		if err != nil {
			return nil, fmt.Errorf("items[%d].abnormalFlag: %w", i, err)
		}

		testedAt, err := parseOptionalTime(itemReq.TestedAt)
		if err != nil {
			return nil, fmt.Errorf("items[%d].testedAt: %w", i, err)
		}

		items = append(items, models.LabReportItem{
			ID:             utils.GenerateID(),
			LabReportID:    reportID,
			ItemCode:       strings.TrimSpace(itemReq.ItemCode),
			ItemName:       strings.TrimSpace(itemReq.ItemName),
			ResultValue:    strings.TrimSpace(itemReq.ResultValue),
			Unit:           strings.TrimSpace(itemReq.Unit),
			ReferenceRange: strings.TrimSpace(itemReq.ReferenceRange),
			AbnormalFlag:   flag,
			TestedAt:       testedAt,
		})
	}

	return items, nil
}

func normalizeOptionalString(in *string) *string {
	if in == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*in)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func parseOptionalTime(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("unsupported time format: %s", value)
}

func applyTimeFieldUpdate(input *string, target **time.Time) error {
	if input == nil {
		return nil
	}
	parsed, err := parseOptionalTime(*input)
	if err != nil {
		return err
	}
	*target = parsed
	return nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

func normalizeAbnormalFlag(input string) (string, error) {
	flag := strings.ToUpper(strings.TrimSpace(input))
	if flag == "" {
		return "N", nil
	}

	switch flag {
	case "H", "L", "N":
		return flag, nil
	default:
		return "", errors.New("abnormalFlag must be one of H, L, N")
	}
}

func normalizeSourceSystem(input string) (string, error) {
	source := strings.ToUpper(strings.TrimSpace(input))
	if source == "" {
		return models.SourceSystemLocal, nil
	}

	switch source {
	case models.SourceSystemLocal, models.SourceSystemLIS, models.SourceSystemPACS:
		return source, nil
	default:
		return "", errors.New("sourceSystem must be one of LOCAL, LIS, PACS")
	}
}
