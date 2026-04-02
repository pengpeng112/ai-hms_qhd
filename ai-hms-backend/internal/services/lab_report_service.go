package services

import (
	"errors"
	"fmt"
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

	query := s.db.Model(&models.LabReport{}).Where("patient_id = ?", patientID)

	if strings.TrimSpace(req.StartDate) != "" {
		startDate, err := parseOptionalTime(req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid startDate: %w", err)
		}
		query = query.Where("reported_at >= ?", *startDate)
	}
	if strings.TrimSpace(req.EndDate) != "" {
		endDate, err := parseOptionalTime(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid endDate: %w", err)
		}
		query = query.Where("reported_at <= ?", *endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var reports []models.LabReport
	offset := (page - 1) * pageSize
	if err := query.Preload("Items").
		Offset(offset).
		Limit(pageSize).
		Order("reported_at DESC NULLS LAST, created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, err
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
