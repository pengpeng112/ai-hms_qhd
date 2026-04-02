package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/hdis"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

// ExamReportSyncService 检查报告同步服务
type ExamReportSyncService struct {
	db           *gorm.DB
	tokenManager *HDISTokenManager
	timeout      time.Duration
}

// NewExamReportSyncService 创建检查报告同步服务
func NewExamReportSyncService(cfg config.HdisConfig) *ExamReportSyncService {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	return &ExamReportSyncService{
		db:           database.GetDB(),
		tokenManager: NewHDISTokenManager(cfg),
		timeout:      timeout,
	}
}

// SyncPatientExamReports 同步患者检查报告
func (s *ExamReportSyncService) SyncPatientExamReports(patientID string) (*LabReportSyncResult, error) {
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
	if runtimeCfg == nil || runtimeCfg.Token == "" || (runtimeCfg.WebcmdURL == "" && runtimeCfg.GraphqlURL == "") {
		return nil, ErrSyncNotConfigured
	}
	var webcmdClient *hdis.WebcmdClient
	var graphqlClient *hdis.GraphQLClient
	if runtimeCfg.WebcmdURL != "" {
		webcmdClient = hdis.NewWebcmdClient(runtimeCfg.WebcmdURL, runtimeCfg.Token, runtimeCfg.Timeout)
	}
	if runtimeCfg.GraphqlURL != "" {
		graphqlClient = hdis.NewGraphQLClient(runtimeCfg.GraphqlURL, runtimeCfg.Token, runtimeCfg.Timeout)
	}

	hdisPatientID, err := s.getHDISPatientID(patientID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	result := &LabReportSyncResult{}
	// 优先使用 GraphQL Examination 链路（真检查链路）。
	// 仅在 GraphQL 未配置时，才回退到 webcmd 分流。
	if graphqlClient != nil {
		exams, err := graphqlClient.GetExaminations(ctx, hdisPatientID)
		if err != nil {
			return nil, err
		}
		s.syncGraphQLExams(patientID, exams, result)
		return result, nil
	}

	headers, err := webcmdClient.GetExaminationList(ctx, []int{hdisPatientID}, nil, nil)
	if err != nil {
		return nil, err
	}

	for _, header := range headers {
		if header.ID <= 0 {
			result.Errors++
			continue
		}
		if hdis.ClassifyExamOrLab(header) == hdis.DataCategoryExam {
			examDate, _ := examParseOptionalTime(header.ResultTime)
			syncedAt, _ := examParseOptionalTime(header.LastModifyTime)
			if syncedAt == nil {
				now := time.Now()
				syncedAt = &now
			}
			payload := hdis.ExamReportPayload{
				ExternalReportID: strconv.Itoa(header.ID),
				Title:            strings.TrimSpace(header.Name),
				Department:       "检查报告",
				Conclusion:       "",
				ExamDate:         examDate,
				SyncedAt:         syncedAt,
			}
			s.syncOneExamReport(patientID, payload, result)
			continue
		}
		result.Skipped++
	}

	return result, nil
}

func (s *ExamReportSyncService) getHDISPatientID(patientID string) (int, error) {
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

func (s *ExamReportSyncService) syncGraphQLExams(patientID string, exams []hdis.GraphQLExamination, result *LabReportSyncResult) {
	for _, exam := range exams {
		payload := hdis.MapGraphQLExaminationToExamReportPayload(exam)
		if strings.TrimSpace(payload.ExternalReportID) == "" {
			result.Errors++
			continue
		}
		s.syncOneExamReport(patientID, payload, result)
	}
}

func (s *ExamReportSyncService) syncOneExamReport(patientID string, payload hdis.ExamReportPayload, result *LabReportSyncResult) {
	existing, exists, err := s.findExistingByExternalID(patientID, payload.ExternalReportID)
	if err != nil {
		result.Errors++
		return
	}
	if exists && examShouldSkipSync(existing.SyncedAt, payload.SyncedAt) {
		result.Skipped++
		return
	}
	if err := s.upsertExamReport(patientID, existing, exists, payload); err != nil {
		result.Errors++
		return
	}

	if exists {
		result.Updated++
	} else {
		result.Created++
	}
}

func (s *ExamReportSyncService) findExistingByExternalID(patientID, externalReportID string) (*models.ExamReport, bool, error) {
	var report models.ExamReport
	err := s.db.Where(
		"patient_id = ? AND external_report_id = ? AND source_system = ?",
		patientID,
		externalReportID,
		models.SourceSystemHDISExam,
	).First(&report).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &report, true, nil
}

func (s *ExamReportSyncService) upsertExamReport(
	patientID string,
	existing *models.ExamReport,
	exists bool,
	payload hdis.ExamReportPayload,
) error {
	var report models.ExamReport
	if exists && existing != nil {
		report = *existing
	} else {
		report = models.ExamReport{
			ID:        utils.GenerateID(),
			PatientID: patientID,
		}
	}

	report.Title = payload.Title
	report.Department = payload.Department
	report.Conclusion = payload.Conclusion
	report.ExamDate = payload.ExamDate
	report.SourceSystem = models.SourceSystemHDISExam
	report.ExternalReportID = &payload.ExternalReportID
	report.SyncedAt = payload.SyncedAt

	if exists {
		return s.db.Save(&report).Error
	}
	return s.db.Create(&report).Error
}

func examShouldSkipSync(existingSyncedAt, sourceLastModified *time.Time) bool {
	if existingSyncedAt == nil || sourceLastModified == nil {
		return false
	}
	return !sourceLastModified.After(*existingSyncedAt)
}
