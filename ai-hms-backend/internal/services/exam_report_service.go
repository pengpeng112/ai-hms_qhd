package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// ExamReportService 检查报告服务
type ExamReportService struct {
	db *gorm.DB
}

// ExamReportListRequest 检查报告列表请求
type ExamReportListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	StartDate string `form:"startDate"` // examDate >= startDate
	EndDate   string `form:"endDate"`   // examDate <= endDate
}

// ExamReportListResponse 检查报告列表响应
type ExamReportListResponse struct {
	Items     []models.ExamReport `json:"items"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"pageSize"`
	TotalPage int                 `json:"totalPage"`
}

// NewExamReportService 创建检查报告服务
func NewExamReportService() *ExamReportService {
	return &ExamReportService{
		db: database.GetDB(),
	}
}

// ListByPatient 按患者查询检查报告
func (s *ExamReportService) ListByPatient(patientID string, req ExamReportListRequest) (*ExamReportListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(patientID) == "" {
		return nil, errors.New("patient id is required")
	}

	page, pageSize := examNormalizePagination(req.Page, req.PageSize)
	query := s.db.Model(&models.ExamReport{}).Where("patient_id = ?", patientID)

	if strings.TrimSpace(req.StartDate) != "" {
		startDate, err := examParseOptionalTime(req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid startDate: %w", err)
		}
		query = query.Where("exam_date >= ?", *startDate)
	}
	if strings.TrimSpace(req.EndDate) != "" {
		endDate, err := examParseOptionalTime(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid endDate: %w", err)
		}
		query = query.Where("exam_date <= ?", *endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var reports []models.ExamReport
	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("exam_date DESC NULLS LAST, created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return &ExamReportListResponse{
		Items:     reports,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}, nil
}

func examNormalizePagination(page, pageSize int) (int, int) {
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

func examParseOptionalTime(value string) (*time.Time, error) {
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
