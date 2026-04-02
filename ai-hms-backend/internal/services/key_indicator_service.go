package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/hdis"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

// KeyIndicatorService 患者关键指标服务
type KeyIndicatorService struct {
	db           *gorm.DB
	tokenManager *HDISTokenManager
	timeout      time.Duration
}

// KeyIndicatorListRequest 关键指标列表查询参数
type KeyIndicatorListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
}

// KeyIndicatorListResponse 关键指标列表响应
type KeyIndicatorListResponse struct {
	Items     []models.PatientKeyIndicator `json:"items"`
	Total     int64                        `json:"total"`
	Page      int                          `json:"page"`
	PageSize  int                          `json:"pageSize"`
	TotalPage int                          `json:"totalPage"`
}

// NewKeyIndicatorService 创建关键指标服务
func NewKeyIndicatorService(cfg config.HdisConfig) *KeyIndicatorService {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	return &KeyIndicatorService{
		db:           database.GetDB(),
		tokenManager: NewHDISTokenManager(cfg),
		timeout:      timeout,
	}
}

// ListByPatient 查询患者关键指标
func (s *KeyIndicatorService) ListByPatient(patientID string, req KeyIndicatorListRequest) (*KeyIndicatorListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(patientID) == "" {
		return nil, errors.New("patient id is required")
	}

	page, pageSize := normalizePagination(req.Page, req.PageSize)
	query := s.db.Model(&models.PatientKeyIndicator{}).Where("patient_id = ?", patientID)

	if strings.TrimSpace(req.StartDate) != "" {
		startDate, err := parseOptionalTime(req.StartDate)
		if err != nil {
			return nil, err
		}
		query = query.Where("test_time >= ?", *startDate)
	}
	if strings.TrimSpace(req.EndDate) != "" {
		endDate, err := parseOptionalTime(req.EndDate)
		if err != nil {
			return nil, err
		}
		query = query.Where("test_time <= ?", *endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var indicators []models.PatientKeyIndicator
	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("test_time DESC NULLS LAST, created_at DESC").
		Find(&indicators).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return &KeyIndicatorListResponse{
		Items:     indicators,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}, nil
}

// SyncPatientKeyIndicators 同步患者关键指标
func (s *KeyIndicatorService) SyncPatientKeyIndicators(patientID string) (*LabReportSyncResult, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(patientID) == "" {
		return nil, ErrSyncPatientIDRequired
	}
	runtimeCfg, err := s.tokenManager.GetRuntimeConfig(context.Background())
	if err != nil {
		return nil, ErrSyncNotConfigured
	}
	if runtimeCfg == nil || runtimeCfg.GraphqlURL == "" || runtimeCfg.Token == "" {
		return nil, ErrSyncNotConfigured
	}
	graphqlClient := hdis.NewGraphQLClient(runtimeCfg.GraphqlURL, runtimeCfg.Token, runtimeCfg.Timeout)

	hdisPatientID, err := s.getHDISPatientID(patientID)
	if err != nil {
		return nil, err
	}

	result := &LabReportSyncResult{}
	page := 1
	pageSize := 50
	for {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		records, queryErr := graphqlClient.GetRecords(ctx, hdisPatientID, page, pageSize)
		cancel()
		if queryErr != nil {
			return nil, queryErr
		}
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			payload := hdis.MapGraphQLRecordToKeyIndicatorPayload(record)
			if payload.ExternalRecordID == "" || payload.IndexName == "" {
				result.Errors++
				continue
			}

			existing, exists, findErr := s.findExistingRecord(patientID, payload.ExternalRecordID)
			if findErr != nil {
				result.Errors++
				continue
			}

			if upsertErr := s.upsertRecord(patientID, existing, exists, payload); upsertErr != nil {
				result.Errors++
				continue
			}

			if exists {
				result.Updated++
			} else {
				result.Created++
			}
		}

		if len(records) < pageSize {
			break
		}
		page++
		if page > 20 {
			break
		}
	}

	return result, nil
}

func (s *KeyIndicatorService) getHDISPatientID(patientID string) (int, error) {
	var basicInfo models.PatientBasicInfo
	if err := s.db.Where("patient_id = ?", patientID).First(&basicInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrSyncPatientBasicNotFound
		}
		return 0, err
	}

	if basicInfo.HdisPatientID == nil || *basicInfo.HdisPatientID <= 0 {
		return 0, ErrSyncPatientHDISIDMissing
	}
	return *basicInfo.HdisPatientID, nil
}

func (s *KeyIndicatorService) findExistingRecord(patientID, externalRecordID string) (*models.PatientKeyIndicator, bool, error) {
	var row models.PatientKeyIndicator
	err := s.db.Where(
		"patient_id = ? AND external_record_id = ? AND source_system = ?",
		patientID,
		externalRecordID,
		models.SourceSystemHDISRecord,
	).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &row, true, nil
}

func (s *KeyIndicatorService) upsertRecord(
	patientID string,
	existing *models.PatientKeyIndicator,
	exists bool,
	payload hdis.KeyIndicatorPayload,
) error {
	var row models.PatientKeyIndicator
	if exists && existing != nil {
		row = *existing
	} else {
		row = models.PatientKeyIndicator{
			ID: utils.GenerateID(),
		}
	}

	row.PatientID = patientID
	row.ExternalRecordID = payload.ExternalRecordID
	row.SourceSystem = models.SourceSystemHDISRecord
	row.IndexName = payload.IndexName
	row.IndexCode = payload.IndexCode
	row.Result = payload.Result
	row.Unit = payload.Unit
	row.Reference = payload.Reference
	row.ResultSign = payload.ResultSign
	row.TestTime = payload.TestedAt
	row.EvaluationResult = payload.EvaluationResult
	row.SyncedAt = payload.SyncedAt

	if exists {
		return s.db.Save(&row).Error
	}
	return s.db.Create(&row).Error
}
