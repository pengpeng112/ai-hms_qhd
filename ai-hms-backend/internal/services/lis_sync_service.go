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

var (
	ErrSyncPatientIDRequired    = errors.New("patient id is required")
	ErrSyncNotConfigured        = errors.New("hdis integration is not configured")
	ErrSyncPatientBasicNotFound = errors.New("patient basic info not found")
	ErrSyncPatientHDISIDMissing = errors.New("patient has no hdis id mapping")
)

// LabReportSyncResult 同步结果
type LabReportSyncResult struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}

// LabReportSyncService 检验报告同步服务
type LabReportSyncService struct {
	db           *gorm.DB
	tokenManager *HDISTokenManager
	timeout      time.Duration
}

// NewLabReportSyncService 创建同步服务
func NewLabReportSyncService(cfg config.HdisConfig) *LabReportSyncService {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	return &LabReportSyncService{
		db:           database.GetDB(),
		tokenManager: NewHDISTokenManager(cfg),
		timeout:      timeout,
	}
}

// SyncPatientLabReports 按患者同步检验报告
func (s *LabReportSyncService) SyncPatientLabReports(patientID string) (*LabReportSyncResult, error) {
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
	if runtimeCfg == nil || runtimeCfg.WebcmdURL == "" || runtimeCfg.GraphqlURL == "" || runtimeCfg.Token == "" {
		return nil, ErrSyncNotConfigured
	}
	webcmdClient := hdis.NewWebcmdClient(runtimeCfg.WebcmdURL, runtimeCfg.Token, runtimeCfg.Timeout)
	graphqlClient := hdis.NewGraphQLClient(runtimeCfg.GraphqlURL, runtimeCfg.Token, runtimeCfg.Timeout)

	hdisPatientID, err := s.getHDISPatientID(patientID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	headers, err := webcmdClient.GetExaminationList(ctx, []int{hdisPatientID}, nil, nil)
	if err != nil {
		return nil, err
	}

	result := &LabReportSyncResult{}
	for _, header := range headers {
		if header.ID <= 0 {
			result.Errors++
			continue
		}
		if hdis.ClassifyExamOrLab(header) != hdis.DataCategoryLab {
			result.Skipped++
			continue
		}

		headerPayload := hdis.MapWebcmdToLabReportPayload(header)
		existing, exists, err := s.findExistingByExternalID(patientID, headerPayload.ExternalReportID)
		if err != nil {
			result.Errors++
			continue
		}

		hasItems := false
		if exists && existing != nil {
			hasItems, err = s.hasLabReportItems(existing.ID)
			if err != nil {
				result.Errors++
				continue
			}
		}

		// 仅当“已存在且已有明细且源数据未更新”时才跳过。
		// 已存在但明细为空（例如历史同步失败）的报告，必须允许重试回填。
		if exists && hasItems && shouldSkipSync(existing.SyncedAt, headerPayload.SyncedAt) {
			result.Skipped++
			continue
		}

		itemCtx, itemCancel := context.WithTimeout(context.Background(), s.timeout)
		graphItems, itemErr := graphqlClient.GetExaminationItems(itemCtx, header.ID)
		itemCancel()

		var itemPayloads []hdis.LabReportItemPayload
		replaceItems := true
		if itemErr != nil {
			// 明细查询失败时保留报告头，错误计数 +1。
			result.Errors++
			// 明细请求失败时，不覆盖已有明细，避免把历史明细误删。
			replaceItems = false
		} else {
			itemPayloads = hdis.MapGraphQLItemsToLabReportItems(graphItems)
		}
		if err := s.upsertLabReport(patientID, existing, exists, headerPayload, itemPayloads, replaceItems); err != nil {
			result.Errors++
			continue
		}

		if exists {
			result.Updated++
		} else {
			result.Created++
		}
	}

	return result, nil
}

func (s *LabReportSyncService) getHDISPatientID(patientID string) (int, error) {
	return 0, errors.New("HDIS患者ID获取暂不可用：老库无HdisPatientID对应列，patient_basic_infos表已弃用")
}

func (s *LabReportSyncService) hasLabReportItems(reportID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.LabReportItem{}).Where("lab_report_id = ?", reportID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *LabReportSyncService) findExistingByExternalID(patientID, externalReportID string) (*models.LabReport, bool, error) {
	var report models.LabReport
	err := s.db.Where(
		"patient_id = ? AND external_report_id = ? AND source_system = ?",
		patientID,
		externalReportID,
		models.SourceSystemLIS,
	).First(&report).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &report, true, nil
}

func (s *LabReportSyncService) upsertLabReport(
	patientID string,
	existing *models.LabReport,
	exists bool,
	header hdis.LabReportPayload,
	items []hdis.LabReportItemPayload,
	replaceItems bool,
) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var report models.LabReport
		if exists && existing != nil {
			report = *existing
		} else {
			report = models.LabReport{
				ID:           utils.GenerateID(),
				PatientID:    patientID,
				SourceSystem: models.SourceSystemLIS,
			}
		}

		report.ReportNo = header.ReportNo
		report.ItemCode = header.ItemCode
		report.ItemName = header.ItemName
		report.ClinicalDiagnosis = header.ClinicalDiagnosis
		report.SpecimenType = header.SpecimenType
		report.Urgency = header.Urgency
		report.RequestDoctor = header.RequestDoctor
		report.RequestedAt = header.RequestedAt
		report.SampledAt = header.SampledAt
		report.ReceivedAt = header.ReceivedAt
		report.ReportedAt = header.ReportedAt
		report.Status = header.Status
		report.SourceSystem = models.SourceSystemLIS

		externalReportID := header.ExternalReportID
		report.ExternalReportID = &externalReportID
		if header.SyncedAt != nil {
			report.SyncedAt = header.SyncedAt
		} else {
			now := time.Now()
			report.SyncedAt = &now
		}

		if exists {
			if err := tx.Save(&report).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&report).Error; err != nil {
				return err
			}
		}

		if replaceItems {
			if err := tx.Where("lab_report_id = ?", report.ID).Delete(&models.LabReportItem{}).Error; err != nil {
				return err
			}

			modelItems := make([]models.LabReportItem, 0, len(items))
			for _, item := range items {
				itemCode := strings.TrimSpace(item.ItemCode)
				if itemCode == "" {
					itemCode = report.ItemCode
				}

				itemName := strings.TrimSpace(item.ItemName)
				if itemName == "" {
					itemName = itemCode
				}

				resultValue := strings.TrimSpace(item.ResultValue)
				if itemCode == "" || itemName == "" || resultValue == "" {
					continue
				}

				modelItems = append(modelItems, models.LabReportItem{
					ID:             utils.GenerateID(),
					LabReportID:    report.ID,
					ItemCode:       itemCode,
					ItemName:       itemName,
					ResultValue:    resultValue,
					Unit:           strings.TrimSpace(item.Unit),
					ReferenceRange: strings.TrimSpace(item.ReferenceRange),
					AbnormalFlag:   strings.TrimSpace(item.AbnormalFlag),
				})
			}

			if len(modelItems) > 0 {
				if err := tx.Create(&modelItems).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func shouldSkipSync(existingSyncedAt, sourceLastModified *time.Time) bool {
	if existingSyncedAt == nil || sourceLastModified == nil {
		return false
	}

	return !sourceLastModified.After(*existingSyncedAt)
}
