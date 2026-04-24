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

// KeyIndicatorService 患者关键指标服务
type KeyIndicatorService struct {
	db           *gorm.DB
	tokenManager *HDISTokenManager
	timeout      time.Duration
}

type legacyKeyIndicatorConfigRow struct {
	ItemCode           string `gorm:"column:ItemCode"`
	ItemName           string `gorm:"column:ItemName"`
	RetItemName        string `gorm:"column:RetItemName"`
	RetExaminationName string `gorm:"column:RetExaminationName"`
	ExaminationName    string `gorm:"column:ExaminationName"`
}

type legacyKeyIndicatorConfigV2Row struct {
	RetItemName        string `gorm:"column:RetItemName"`
	RetExaminationName string `gorm:"column:RetExaminationName"`
	Unit               string `gorm:"column:Unit"`
	Reference          string `gorm:"column:Reference"`
}

type legacyKeyIndicatorConfigRowFallback struct {
	ItemCode           string `gorm:"column:ItemCode"`
	ItemName           string `gorm:"column:ItemName"`
	RetItemName        string `gorm:"column:RetItemName"`
	RetExaminationName string `gorm:"column:RetExaminationName"`
	ExaminationName    string `gorm:"column:ExaminationName"`
}

type legacyKeyIndicatorRow struct {
	ItemID         int64      `gorm:"column:item_id"`
	ExamID         int64      `gorm:"column:exam_id"`
	TestNO         string     `gorm:"column:test_no"`
	ItemCode       string     `gorm:"column:item_code"`
	ItemName       string     `gorm:"column:item_name"`
	ResultValue    string     `gorm:"column:result_value"`
	Unit           string     `gorm:"column:unit"`
	ReferenceRange string     `gorm:"column:reference_range"`
	ResultSign     string     `gorm:"column:result_sign"`
	TestedAt       *time.Time `gorm:"column:tested_at"`
	CreatedAt      *time.Time `gorm:"column:created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
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

	return s.listLegacyByPatient(patientID, req)
}

func (s *KeyIndicatorService) listLegacyByPatient(patientID string, req KeyIndicatorListRequest) (*KeyIndicatorListResponse, error) {
	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("patient id is required")
	}

	page, pageSize := normalizePagination(req.Page, req.PageSize)

	cfgRows := make([]legacyKeyIndicatorConfigRow, 0)
	var cfgRowsV2 []legacyKeyIndicatorConfigV2Row
	err = s.db.Table(`"LIS_ExaminationItem_Config"`).
		Select(`"RetItemName", "RetExaminationName", "Unit", "Reference"`).
		Where(`COALESCE("TenantId", ?) = ? AND "RetExaminationName" = ?`, legacyTenantID, legacyTenantID, "重要指标").
		Find(&cfgRowsV2).Error
	if err != nil && !isIgnorableLegacyQueryError(err) {
		return nil, err
	}
	for _, row := range cfgRowsV2 {
		cfgRows = append(cfgRows, legacyKeyIndicatorConfigRow{
			ItemCode:           "",
			ItemName:           strings.TrimSpace(row.RetItemName),
			RetItemName:        strings.TrimSpace(row.RetItemName),
			RetExaminationName: strings.TrimSpace(row.RetExaminationName),
			ExaminationName:    strings.TrimSpace(row.RetExaminationName),
		})
	}
	if len(cfgRows) == 0 {
		var fallbackRows []legacyKeyIndicatorConfigRowFallback
		if err := s.db.Table(`"LIS_ExaminationItem_Ret"`).
			Select(`"ItemCode", "ItemName", "RetItemName", "RetExaminationName", "ExaminationName"`).
			Where(`COALESCE("TenantId", ?) = ?`, legacyTenantID, legacyTenantID).
			Find(&fallbackRows).Error; err != nil {
			if isIgnorableLegacyQueryError(err) {
				return &KeyIndicatorListResponse{
					Items:     []models.PatientKeyIndicator{},
					Total:     0,
					Page:      page,
					PageSize:  pageSize,
					TotalPage: 0,
				}, nil
			}
			return nil, err
		}
		for _, row := range fallbackRows {
			cfgRows = append(cfgRows, legacyKeyIndicatorConfigRow{
				ItemCode:           row.ItemCode,
				ItemName:           row.ItemName,
				RetItemName:        row.RetItemName,
				RetExaminationName: row.RetExaminationName,
				ExaminationName:    row.ExaminationName,
			})
		}
	}

	codeSet := make(map[string]struct{}, len(cfgRows))
	nameSet := make(map[string]struct{}, len(cfgRows)*2)
	displayByCode := make(map[string]string, len(cfgRows))
	displayByName := make(map[string]string, len(cfgRows)*2)

	for _, cfg := range cfgRows {
		displayName := firstNonEmptyText(strings.TrimSpace(cfg.RetExaminationName), strings.TrimSpace(cfg.ExaminationName), strings.TrimSpace(cfg.RetItemName), strings.TrimSpace(cfg.ItemName))
		if code := strings.TrimSpace(cfg.ItemCode); code != "" {
			codeSet[code] = struct{}{}
			if displayName != "" {
				displayByCode[code] = displayName
			}
		}
		for _, rawName := range []string{cfg.ItemName, cfg.RetItemName} {
			name := strings.TrimSpace(rawName)
			if name == "" {
				continue
			}
			nameSet[name] = struct{}{}
			if displayName != "" {
				displayByName[name] = displayName
			}
		}
	}

	if len(codeSet) == 0 && len(nameSet) == 0 {
		return &KeyIndicatorListResponse{
			Items:     []models.PatientKeyIndicator{},
			Total:     0,
			Page:      page,
			PageSize:  pageSize,
			TotalPage: 0,
		}, nil
	}

	codes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		codes = append(codes, code)
	}
	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}

	query := s.db.Table(`"LIS_ExaminationItem" AS i`).
		Joins(`JOIN "LIS_Examination" AS e ON e."Id" = i."ExaminationId"`).
		Where(`e."PatientId" = ? AND e."TenantId" = ?`, legacyPatientID, legacyTenantID)

	if len(codes) > 0 && len(names) > 0 {
		query = query.Where(`(i."ItemCode" IN ? OR i."ItemName" IN ?)`, codes, names)
	} else if len(codes) > 0 {
		query = query.Where(`i."ItemCode" IN ?`, codes)
	} else {
		query = query.Where(`i."ItemName" IN ?`, names)
	}

	if strings.TrimSpace(req.StartDate) != "" {
		startDate, err := parseOptionalTime(req.StartDate)
		if err != nil {
			return nil, err
		}
		query = query.Where(`COALESCE(e."ResultTime", i."LastModifyTime", e."LastModifyTime") >= ?`, *startDate)
	}
	if strings.TrimSpace(req.EndDate) != "" {
		endDate, err := parseOptionalTime(req.EndDate)
		if err != nil {
			return nil, err
		}
		query = query.Where(`COALESCE(e."ResultTime", i."LastModifyTime", e."LastModifyTime") <= ?`, *endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		if isIgnorableLegacyQueryError(err) {
			return &KeyIndicatorListResponse{
				Items:     []models.PatientKeyIndicator{},
				Total:     0,
				Page:      page,
				PageSize:  pageSize,
				TotalPage: 0,
			}, nil
		}
		return nil, err
	}

	var rows []legacyKeyIndicatorRow
	offset := (page - 1) * pageSize
	if err := query.
		Select(`i."Id" AS item_id, e."Id" AS exam_id, e."TestNO" AS test_no, i."ItemCode" AS item_code, i."ItemName" AS item_name, i."Result" AS result_value, i."Unit" AS unit, i."Reference" AS reference_range, i."ResultSign" AS result_sign, COALESCE(e."ResultTime", i."LastModifyTime", e."LastModifyTime") AS tested_at, i."LastModifyTime" AS updated_at, e."CreateTime" AS created_at`).
		Order(`COALESCE(e."ResultTime", i."LastModifyTime", e."LastModifyTime") DESC`).
		Order(`i."Id" DESC`).
		Offset(offset).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		if isIgnorableLegacyQueryError(err) {
			return &KeyIndicatorListResponse{
				Items:     []models.PatientKeyIndicator{},
				Total:     0,
				Page:      page,
				PageSize:  pageSize,
				TotalPage: 0,
			}, nil
		}
		return nil, err
	}

	items := make([]models.PatientKeyIndicator, 0, len(rows))
	for _, row := range rows {
		itemCode := strings.TrimSpace(row.ItemCode)
		itemName := strings.TrimSpace(row.ItemName)
		indexName := firstNonEmptyText(displayByCode[itemCode], displayByName[itemName], itemName, itemCode)
		externalID := strings.TrimSpace(row.TestNO)
		if externalID == "" {
			externalID = strconv.FormatInt(row.ExamID, 10)
		}
		externalID = externalID + "-" + strconv.FormatInt(row.ItemID, 10)

		createdAt := time.Now()
		if row.CreatedAt != nil && !row.CreatedAt.IsZero() {
			createdAt = *row.CreatedAt
		} else if row.TestedAt != nil && !row.TestedAt.IsZero() {
			createdAt = *row.TestedAt
		}
		updatedAt := createdAt
		if row.UpdatedAt != nil && !row.UpdatedAt.IsZero() {
			updatedAt = *row.UpdatedAt
		}

		items = append(items, models.PatientKeyIndicator{
			ID:               strconv.FormatInt(row.ItemID, 10),
			PatientID:        patientID,
			ExternalRecordID: externalID,
			SourceSystem:     models.SourceSystemLIS,
			IndexName:        indexName,
			IndexCode:        itemCode,
			Result:           strings.TrimSpace(row.ResultValue),
			Unit:             strings.TrimSpace(row.Unit),
			Reference:        strings.TrimSpace(row.ReferenceRange),
			ResultSign:       strings.TrimSpace(row.ResultSign),
			TestTime:         row.TestedAt,
			EvaluationResult: legacyIndicatorEvaluation(strings.TrimSpace(row.ResultSign)),
			SyncedAt:         nil,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		})
	}

	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return &KeyIndicatorListResponse{
		Items:     items,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}, nil
}

func legacyIndicatorEvaluation(sign string) string {
	v := strings.ToUpper(strings.TrimSpace(sign))
	switch {
	case strings.Contains(v, "H") || strings.Contains(sign, "↑"):
		return "偏高"
	case strings.Contains(v, "L") || strings.Contains(sign, "↓"):
		return "偏低"
	case v == "":
		return ""
	default:
		return "正常"
	}
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
