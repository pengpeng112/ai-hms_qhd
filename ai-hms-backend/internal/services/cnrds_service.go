package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

var periodRe = regexp.MustCompile(`^\d{4}-\d{2}$`)

var eventTypes = map[string]bool{
	models.CnrdsEventDeath:       true,
	models.CnrdsEventTransplant:  true,
	models.CnrdsEventTransferOut: true,
}

type labLookup interface {
	Value(patientID int64, conceptID string, start, end time.Time) *float64
}

type qcLabAdapter struct {
	svc *QCService
}

func (a qcLabAdapter) Value(patientID int64, conceptID string, start, end time.Time) *float64 {
	return a.svc.latestLabValue(patientID, conceptID, start, end)
}

type CnrdsService struct {
	db       *gorm.DB
	tenantID int64
	lab      labLookup
	exporter Exporter
}

func NewCnrdsService(tenantID int64) *CnrdsService {
	return &CnrdsService{
		db:       database.GetDB(),
		tenantID: tenantID,
		lab:      qcLabAdapter{svc: NewQCService()},
		exporter: CSVExporter{},
	}
}

func NewCnrdsServiceWith(db *gorm.DB, tenantID int64, lab labLookup, exporter Exporter) *CnrdsService {
	return &CnrdsService{db: db, tenantID: tenantID, lab: lab, exporter: exporter}
}

func (s *CnrdsService) GenerateMonthly(period string) (*models.CnrdsReport, error) {
	if !periodRe.MatchString(period) {
		return nil, fmt.Errorf("周期格式应为 YYYY-MM")
	}

	t, err := time.Parse("2006-01", period)
	if err != nil {
		return nil, fmt.Errorf("周期格式应为 YYYY-MM")
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	type patientInfo struct {
		ID     int64  `gorm:"column:Id"`
		Name   string `gorm:"column:Name"`
		Gender string `gorm:"column:Gender"`
	}
	var patients []patientInfo
	if err := s.db.Table(`"Register_PatientInfomation"`).
		Select(`"Id", "Name", "Gender"`).
		Where(`"TenantId" = ?`, s.tenantID).
		Find(&patients).Error; err != nil {
		return nil, fmt.Errorf("查询患者列表失败: %w", err)
	}

	rows := make([]models.CnrdsContentRow, 0, len(patients))
	for _, p := range patients {
		row := s.aggregateRow(p.ID, p.Name, p.Gender, start, end)
		rows = append(rows, row)
	}

	report := &models.CnrdsReport{
		ID:           utils.GenerateID(),
		TenantID:     s.tenantID,
		Period:       period,
		ReportType:   models.CnrdsTypeMonthly,
		PatientCount: len(rows),
		Status:       models.CnrdsStatusDraft,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	content := models.CnrdsContent{Period: period, Type: models.CnrdsTypeMonthly, Rows: rows}
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("序列化内容失败: %w", err)
	}
	report.Content = string(contentBytes)

	if err := s.db.Create(report).Error; err != nil {
		return nil, fmt.Errorf("保存月报失败: %w", err)
	}

	return report, nil
}

func (s *CnrdsService) GenerateEvent(patientID int64, eventType string) (*models.CnrdsReport, error) {
	if !eventTypes[eventType] {
		return nil, fmt.Errorf("未知事件类型: %s", eventType)
	}

	type patientInfo struct {
		Name   string `gorm:"column:Name"`
		Gender string `gorm:"column:Gender"`
	}
	var p patientInfo
	if err := s.db.Table(`"Register_PatientInfomation"`).
		Select(`"Name", "Gender"`).
		Where(`"Id" = ? AND "TenantId" = ?`, patientID, s.tenantID).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("患者不存在")
		}
		return nil, fmt.Errorf("查询患者失败: %w", err)
	}

	end := time.Now().UTC()
	start := end.AddDate(-1, 0, 0)

	type outcomeRow struct {
		Type        string    `gorm:"column:Type"`
		OutComeTime time.Time `gorm:"column:OutComeTime"`
		Reason      string    `gorm:"column:Reason"`
	}
	var oc outcomeRow
	q := s.db.Table(`"Register_OutCome"`).
		Select(`"Type", "OutComeTime", "Reason"`).
		Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, s.tenantID).
		Order(`"OutComeTime" DESC, "CreateTime" DESC`)

	switch eventType {
	case models.CnrdsEventDeath:
		q = q.Where(`"Type" = ? AND "Reason" = ?`, "20", "死亡")
	case models.CnrdsEventTransplant:
		q = q.Where(`"Type" = ? AND "Reason" = ?`, "20", "转肾移植")
	case models.CnrdsEventTransferOut:
		q = q.Where(`"Type" = ?`, "20")
	}

	if err := q.First(&oc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("未找到该患者的%s转归记录", eventType)
		}
		return nil, fmt.Errorf("查询转归失败: %w", err)
	}

	row := s.aggregateRow(patientID, p.Name, p.Gender, start, end)
	row.OutcomeType = eventType
	row.OutcomeDate = oc.OutComeTime.Format("2006-01-02 15:04")
	row.DeathReason = oc.Reason

	report := &models.CnrdsReport{
		ID:           utils.GenerateID(),
		TenantID:     s.tenantID,
		Period:       oc.OutComeTime.Format("2006-01"),
		ReportType:   models.CnrdsTypeEvent,
		EventType:    eventType,
		PatientID:    strconv.FormatInt(patientID, 10),
		PatientCount: 1,
		Status:       models.CnrdsStatusDraft,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	content := models.CnrdsContent{
		Period: report.Period,
		Type:   models.CnrdsTypeEvent,
		Rows:   []models.CnrdsContentRow{row},
	}
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("序列化内容失败: %w", err)
	}
	report.Content = string(contentBytes)

	if err := s.db.Create(report).Error; err != nil {
		return nil, fmt.Errorf("保存事件报失败: %w", err)
	}

	return report, nil
}

func (s *CnrdsService) aggregateRow(patientID int64, name, gender string, start, end time.Time) models.CnrdsContentRow {
	return models.CnrdsContentRow{
		PatientID: strconv.FormatInt(patientID, 10),
		Name:      name,
		Gender:    gender,

		Hb:  s.lab.Value(patientID, "HEMOGLOBIN", start, end),
		Ca:  s.lab.Value(patientID, "SERUM_CA", start, end),
		P:   s.lab.Value(patientID, "SERUM_P", start, end),
		PTH: s.lab.Value(patientID, "IPTH", start, end),
		Albumin: s.lab.Value(patientID, "ALBUMIN", start, end),
		KtV:     s.lab.Value(patientID, "KTV", start, end),
	}
}

type CnrdsListFilter struct {
	Period     string
	ReportType string
	Status     string
}

func (s *CnrdsService) List(f CnrdsListFilter) ([]models.CnrdsReport, error) {
	q := s.db.Where(`"tenant_id" = ?`, s.tenantID)
	if f.Period != "" {
		q = q.Where(`"period" = ?`, f.Period)
	}
	if f.ReportType != "" {
		q = q.Where(`"report_type" = ?`, f.ReportType)
	}
	if f.Status != "" {
		q = q.Where(`"status" = ?`, f.Status)
	}
	var reports []models.CnrdsReport
	if err := q.Order(`"created_at" DESC`).Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("查询报告列表失败: %w", err)
	}
	return reports, nil
}

func (s *CnrdsService) Get(id string) (*models.CnrdsReport, error) {
	var report models.CnrdsReport
	if err := s.db.Where(`"id" = ? AND "tenant_id" = ?`, id, s.tenantID).First(&report).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("查询报告失败: %w", err)
	}
	return &report, nil
}

func (s *CnrdsService) Export(id string) (filename string, data []byte, err error) {
	report, err := s.Get(id)
	if err != nil {
		return "", nil, err
	}
	if report.Status == models.CnrdsStatusSubmitted {
		return "", nil, fmt.Errorf("已提交的报告不能再导出")
	}
	filename, data, err = s.exporter.Export(report)
	if err != nil {
		return "", nil, err
	}
	report.ExportRef = filename
	report.Status = models.CnrdsStatusExported
	report.UpdatedAt = time.Now()
	if err := s.db.Model(report).Updates(map[string]interface{}{
		"status":     models.CnrdsStatusExported,
		"export_ref": filename,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return "", nil, fmt.Errorf("更新导出状态失败: %w", err)
	}
	return filename, data, nil
}

func (s *CnrdsService) Submit(id string, reviewedBy string) error {
	if reviewedBy == "" {
		return fmt.Errorf("提交须填写核对人")
	}
	report, err := s.Get(id)
	if err != nil {
		return err
	}
	if report.Status == models.CnrdsStatusSubmitted {
		return fmt.Errorf("报告已提交，不能重复提交")
	}
	if report.Status != models.CnrdsStatusExported {
		return fmt.Errorf("必须先导出再提交")
	}
	now := time.Now()
	report.ReviewedBy = reviewedBy
	report.Status = models.CnrdsStatusSubmitted
	report.SubmittedAt = &now
	report.UpdatedAt = now
	if err := s.db.Model(report).Updates(map[string]interface{}{
		"status":       models.CnrdsStatusSubmitted,
		"reviewed_by":  reviewedBy,
		"submitted_at": now,
		"updated_at":   now,
	}).Error; err != nil {
		return fmt.Errorf("更新提交状态失败: %w", err)
	}
	return nil
}

func unmarshalContent(raw string, target *models.CnrdsContent) error {
	return json.Unmarshal([]byte(raw), target)
}
