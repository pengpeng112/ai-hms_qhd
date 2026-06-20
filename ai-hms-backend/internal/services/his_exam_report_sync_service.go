package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/his_oracle"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type HisExamReportSyncService struct {
	db           *gorm.DB
	oracleCfg    his_oracle.Config
	mappingSvc   *ExternalPatientMappingService
	oracleClient *his_oracle.Client
	tenantID     int64
}

func NewHisExamReportSyncService(oracleCfg his_oracle.Config, tenantID int64) *HisExamReportSyncService {
	return &HisExamReportSyncService{
		db:         database.GetDB(),
		oracleCfg:  oracleCfg,
		mappingSvc: NewExternalPatientMappingService(),
		tenantID:   tenantID,
	}
}

func (s *HisExamReportSyncService) ensureOracleClient(ctx context.Context) (*his_oracle.Client, error) {
	if s.oracleClient != nil {
		return s.oracleClient, nil
	}
	client, err := his_oracle.NewClient(s.oracleCfg)
	if err != nil {
		return nil, fmt.Errorf("HIS Oracle connect failed: %w", err)
	}
	s.oracleClient = client
	return client, nil
}

func (s *HisExamReportSyncService) CloseOracleClient() {
	if s.oracleClient != nil {
		_ = s.oracleClient.Close()
		s.oracleClient = nil
	}
}

type HisExamSyncResult struct {
	Created        int
	Updated        int
	Skipped        int
	Failed         int
	Errors         []string
	MaxCursorTime  *time.Time
}

type examSyncItemRef struct {
	examNo   string
	reportID string
}

func (s *HisExamReportSyncService) SyncPatientExamReports(
	ctx context.Context,
	legacyPatientID int64,
) (*HisExamSyncResult, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	client, err := s.ensureOracleClient(ctx)
	if err != nil {
		return nil, err
	}

	mappings, err := s.mappingSvc.FindByLegacyPatient(legacyPatientID)
	if err != nil {
		return nil, fmt.Errorf("find external mappings failed: %w", err)
	}

	result := &HisExamSyncResult{}
	if len(mappings) == 0 {
		return result, nil
	}

	var itemRefs []examSyncItemRef

	for _, m := range mappings {
		if m.MatchStatus != models.MatchStatusConfirmed {
			continue
		}
		cursorTime := time.Time{}
		if m.LastSyncedAt != nil {
			cursorTime = *m.LastSyncedAt
		}

		rows, qErr := client.QueryExamReports(ctx, his_oracle.QueryExamReportsParams{
			CursorTime: cursorTime,
			BatchSize:  500,
			PatientID:  m.ExternalPatientID,
		})
		if qErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("query HIS failed (patient=%s): %v", m.ExternalPatientID, qErr))
			result.Failed++
			continue
		}

		for _, row := range rows {
			report := his_oracle.MapHisExamToExamReport(row, legacyPatientID)
			created, reportID, upErr := s.upsertExamReport(&report)
			if upErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("upsert failed (exam_no=%s): %v", row.ExamNo, upErr))
				result.Failed++
				continue
			}
			if created {
				result.Created++
			} else {
				result.Updated++
			}
			itemRefs = append(itemRefs, examSyncItemRef{examNo: row.ExamNo, reportID: reportID})
		}

		now := time.Now()
		if err := s.db.Model(&m).Update("last_synced_at", now).Error; err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("update mapping synced_at failed: %v", err))
		}
	}

	if len(itemRefs) > 0 {
		if itemErr := s.syncExamItems(ctx, client, itemRefs); itemErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("items sync warning: %v", itemErr))
		}
	}

	return result, nil
}

func (s *HisExamReportSyncService) SyncBatch(
	ctx context.Context,
	cursorTime time.Time,
	batchSize int,
) (*HisExamSyncResult, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	client, err := s.ensureOracleClient(ctx)
	if err != nil {
		return nil, err
	}

	rows, qErr := client.QueryExamReports(ctx, his_oracle.QueryExamReportsParams{
		CursorTime: cursorTime,
		BatchSize:  batchSize,
	})
	if qErr != nil {
		return nil, fmt.Errorf("HIS batch query failed: %w", qErr)
	}

	result := &HisExamSyncResult{}
	var itemRefs []examSyncItemRef

	for _, row := range rows {
		if row.CreateDate != nil && (result.MaxCursorTime == nil || row.CreateDate.After(*result.MaxCursorTime)) {
			result.MaxCursorTime = row.CreateDate
		}

		if strings.TrimSpace(row.PatientID) == "" {
			result.Skipped++
			continue
		}

		legacyPID, resolveErr := s.mappingSvc.ResolveLegacyPatientID(
			models.ExternalSystemHISOracle,
			row.PatientID,
			nil,
		)
		if resolveErr != nil {
			if row.IDNo != nil && strings.TrimSpace(*row.IDNo) != "" {
				m, autoErr := s.mappingSvc.AutoMatchByIDNo(
					models.ExternalSystemHISOracle,
					row.PatientID,
					row.IDNo,
					s.tenantID,
				)
				if autoErr == nil && m != nil {
					legacyPID = m.LegacyPatientID
				}
			}
		}
		if resolveErr != nil && legacyPID == 0 {
			result.Skipped++
			continue
		}

		report := his_oracle.MapHisExamToExamReport(row, legacyPID)
		created, reportID, upErr := s.upsertExamReport(&report)
		if upErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("upsert failed (exam_no=%s): %v", row.ExamNo, upErr))
			result.Failed++
			continue
		}
		if created {
			result.Created++
		} else {
			result.Updated++
		}
		itemRefs = append(itemRefs, examSyncItemRef{examNo: row.ExamNo, reportID: reportID})
	}

	if len(itemRefs) > 0 {
		if itemErr := s.syncExamItems(ctx, client, itemRefs); itemErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("items sync warning: %v", itemErr))
		}
	}

	return result, nil
}

func (s *HisExamReportSyncService) upsertExamReport(report *models.ExamReport) (created bool, reportID string, err error) {
	if report.ExternalReportID == nil || strings.TrimSpace(*report.ExternalReportID) == "" {
		return false, "", errors.New("external report id is required")
	}
	var existing models.ExamReport
	err = s.db.Where("patient_id = ? AND source_system = ? AND external_report_id = ?",
		report.PatientID, report.SourceSystem, *report.ExternalReportID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			report.ID = utils.GenerateID()
			rid := report.ID
			return true, rid, s.db.Create(report).Error
		}
		return false, "", err
	}
	updates := map[string]interface{}{
		"title":      report.Title,
		"conclusion": report.Conclusion,
		"department": report.Department,
		"exam_date":  report.ExamDate,
		"synced_at":  time.Now(),
		"updated_at": time.Now(),
	}
	return false, existing.ID, s.db.Model(&existing).Updates(updates).Error
}

func (s *HisExamReportSyncService) syncExamItems(ctx context.Context, client *his_oracle.Client, refs []examSyncItemRef) error {
	if len(refs) == 0 || s.db == nil {
		return nil
	}

	examNos := make([]string, len(refs))
	reportIDByExamNo := make(map[string]string, len(refs))
	seenReportIDs := make(map[string]bool, len(refs))
	var uniqueReportIDs []string

	for i, ref := range refs {
		examNos[i] = ref.examNo
		reportIDByExamNo[ref.examNo] = ref.reportID
		if !seenReportIDs[ref.reportID] {
			seenReportIDs[ref.reportID] = true
			uniqueReportIDs = append(uniqueReportIDs, ref.reportID)
		}
	}

	items, qErr := client.QueryExamItems(ctx, examNos)
	if qErr != nil {
		return qErr
	}

	dbItems := make([]models.ExamReportItem, 0, len(items))
	itemIdx := 0
	for _, item := range items {
		name := derefStr(item.ExamItem)
		if strings.TrimSpace(name) == "" {
			continue
		}
		reportID := reportIDByExamNo[item.ExamNo]
		if reportID == "" {
			continue
		}
		dbItem := models.ExamReportItem{
			ID:           utils.GenerateID(),
			ExamReportID: reportID,
			ItemName:     name,
			ItemCode:     item.ExamItemCode,
			SortOrder:    itemIdx,
		}
		dbItems = append(dbItems, dbItem)
		itemIdx++
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		const deleteBatch = 100
		for i := 0; i < len(uniqueReportIDs); i += deleteBatch {
			end := i + deleteBatch
			if end > len(uniqueReportIDs) {
				end = len(uniqueReportIDs)
			}
			if err := tx.Where("exam_report_id IN ?", uniqueReportIDs[i:end]).Delete(&models.ExamReportItem{}).Error; err != nil {
				return fmt.Errorf("delete old items failed: %w", err)
			}
		}

		const insertBatch = 200
		for i := 0; i < len(dbItems); i += insertBatch {
			end := i + insertBatch
			if end > len(dbItems) {
				end = len(dbItems)
			}
			if err := tx.CreateInBatches(dbItems[i:end], insertBatch).Error; err != nil {
				return fmt.Errorf("insert items failed: %w", err)
			}
		}
		return nil
	})
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *HisExamReportSyncService) PrematchAll(ctx context.Context) (int, error) {
	if s.db == nil {
		return 0, errors.New("database not available")
	}
	client, err := s.ensureOracleClient(ctx)
	if err != nil {
		return 0, err
	}

	var localIDs []struct {
		PatientID int64  `gorm:"column:PatientId"`
		IDNo      string `gorm:"column:IDNo"`
	}
	if err := s.db.Raw(`
		SELECT id."PatientId", TRIM(id."IDNo") AS "IDNo"
		FROM "Register_IDInfomation" id
		JOIN "Register_PatientInfomation" p ON p."Id" = id."PatientId"
		WHERE p."TenantId" = ?
		  AND COALESCE(id."IsDisabled", false) = false
		  AND NULLIF(TRIM(id."IDNo"), '') IS NOT NULL
	`, s.tenantID).Scan(&localIDs).Error; err != nil {
		return 0, fmt.Errorf("查询本地 ID_NO 失败: %w", err)
	}

	matched := 0
	for _, local := range localIDs {
		var count int64
		s.db.Model(&models.ExternalPatientMapping{}).
			Where("external_system = ? AND id_no = ? AND match_status = ?",
				models.ExternalSystemHISOracle, local.IDNo, models.MatchStatusConfirmed).
			Count(&count)
		if count > 0 {
			continue
		}

		hisPID, qErr := client.FindPatientIDByIDNo(ctx, local.IDNo)
		if qErr != nil || hisPID == "" {
			continue
		}

		m := &models.ExternalPatientMapping{
			ID:               fmt.Sprintf("epm_%s_%s", models.ExternalSystemHISOracle, hisPID),
			TenantID:         s.tenantID,
			LegacyPatientID:  local.PatientID,
			ExternalSystem:   models.ExternalSystemHISOracle,
			ExternalPatientID: hisPID,
			IDNo:             &local.IDNo,
			MatchStatus:      models.MatchStatusConfirmed,
		}
		if err := s.mappingSvc.CreateMapping(m); err != nil {
			continue
		}
		matched++
	}
	return matched, nil
}
