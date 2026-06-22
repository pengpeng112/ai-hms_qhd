package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

const severeReportDeadline = 6 * time.Hour

type AdverseEventService struct {
	db       *gorm.DB
	tenantID int64
}

func NewAdverseEventService() *AdverseEventService {
	return &AdverseEventService{db: database.GetDB(), tenantID: LegacyTenantID}
}

var validAeSeverity = map[string]struct{}{
	models.AESeverityMild: {}, models.AESeverityModerate: {}, models.AESeveritySevere: {},
}

var aeNextStatus = map[string][]string{
	models.AEStatusRegistered:   {models.AEStatusReported},
	models.AEStatusReported:     {models.AEStatusAcknowledged},
	models.AEStatusAcknowledged: {models.AEStatusProcessing, models.AEStatusClosed},
	models.AEStatusProcessing:   {models.AEStatusClosed},
}

func validAeStatusTransition(from, to string) bool {
	nexts, ok := aeNextStatus[from]
	if !ok {
		return false
	}
	for _, n := range nexts {
		if n == to {
			return true
		}
	}
	return false
}

type AeRegisterInput struct {
	PatientID   int64     `json:"patientId"`
	TreatmentID *int64    `json:"treatmentId"`
	EventType   string    `json:"eventType"`
	Severity    string    `json:"severity"`
	OccurredAt  time.Time `json:"occurredAt"`
	Description string    `json:"description"`
	Handling    string    `json:"handling"`
	Outcome     string    `json:"outcome"`
	ReporterID  string    `json:"reporterId"`
}

func (s *AdverseEventService) Register(in AeRegisterInput) (*models.AdverseEvent, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if _, ok := validAeSeverity[in.Severity]; !ok {
		return nil, errors.New("非法分级：须为 mild/moderate/severe")
	}
	if strings.TrimSpace(in.EventType) == "" {
		return nil, errors.New("事件分类不能为空")
	}
	if in.OccurredAt.IsZero() {
		return nil, errors.New("发生时间不能为空")
	}
	oc := in.OccurredAt
	rec := &models.AdverseEvent{
		ID:          utils.GenerateID(),
		TenantID:    s.tenantID,
		PatientID:   in.PatientID,
		TreatmentID: in.TreatmentID,
		EventType:   strings.TrimSpace(in.EventType),
		Severity:    in.Severity,
		OccurredAt:  &oc,
		Description: in.Description,
		Handling:    in.Handling,
		Outcome:     in.Outcome,
		ReporterID:  in.ReporterID,
		Status:      models.AEStatusRegistered,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

type AeReportInput struct {
	ReportedTo []AeReportTarget `json:"reportedTo"`
}

type AeReportTarget struct {
	Role   string `json:"role"`
	UserID string `json:"userId"`
}

func (s *AdverseEventService) Report(id string, in AeReportInput) (*models.AdverseEvent, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var rec models.AdverseEvent
	if err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&rec).Error; err != nil {
		return nil, errors.New("记录不存在")
	}
	if rec.Status != models.AEStatusRegistered {
		return nil, errors.New("仅已登记且未上报的记录可上报")
	}
	if len(in.ReportedTo) == 0 {
		return nil, errors.New("上报对象不能为空")
	}
	b, _ := json.Marshal(in.ReportedTo)
	now := time.Now()
	var within6h bool
	if rec.Severity == models.AESeveritySevere && rec.OccurredAt != nil {
		within6h = now.Sub(*rec.OccurredAt) <= severeReportDeadline
	}
	updates := map[string]any{
		"reported_to": string(b),
		"reported_at": now,
		"within_6h":   within6h,
		"status":      models.AEStatusReported,
		"updated_at":  now,
	}
	if err := s.db.Model(&rec).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

type AeStatusInput struct {
	Status    string `json:"status"`
	CqiLinked *bool  `json:"cqiLinked"`
}

func (s *AdverseEventService) UpdateStatus(id string, in AeStatusInput) (*models.AdverseEvent, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var rec models.AdverseEvent
	if err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&rec).Error; err != nil {
		return nil, errors.New("记录不存在")
	}
	if !validAeStatusTransition(rec.Status, in.Status) {
		return nil, errors.New("状态流转不合法")
	}
	updates := map[string]any{
		"status":     in.Status,
		"updated_at": time.Now(),
	}
	if in.CqiLinked != nil {
		updates["cqi_linked"] = *in.CqiLinked
	}
	if err := s.db.Model(&rec).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *AdverseEventService) GetByID(id string) (*models.AdverseEvent, error) {
	var rec models.AdverseEvent
	if err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&rec).Error; err != nil {
		return nil, errors.New("记录不存在")
	}
	return &rec, nil
}

func (s *AdverseEventService) List(severity, status string, patientID *int64) ([]models.AdverseEvent, error) {
	q := s.db.Where("tenant_id = ?", s.tenantID)
	if severity != "" {
		q = q.Where("severity = ?", severity)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if patientID != nil {
		q = q.Where("patient_id = ?", *patientID)
	}
	var rows []models.AdverseEvent
	if err := q.Order("occurred_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type AeAlerts struct {
	SevereUnreported []models.AdverseEvent `json:"severeUnreported"`
	SevereOverdue    []models.AdverseEvent `json:"severeOverdue"`
	Pending          []models.AdverseEvent `json:"pending"`
}

func (s *AdverseEventService) Alerts() (*AeAlerts, error) {
	var all []models.AdverseEvent
	if err := s.db.Where("tenant_id = ?", s.tenantID).Order("occurred_at DESC").Find(&all).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	res := &AeAlerts{}
	for i := range all {
		ae := &all[i]
		if ae.Severity == models.AESeveritySevere && ae.Status == models.AEStatusRegistered {
			if ae.OccurredAt != nil && now.Sub(*ae.OccurredAt) > severeReportDeadline {
				res.SevereOverdue = append(res.SevereOverdue, *ae)
			} else {
				res.SevereUnreported = append(res.SevereUnreported, *ae)
			}
		}
		if ae.Status == models.AEStatusRegistered || ae.Status == models.AEStatusReported ||
			ae.Status == models.AEStatusAcknowledged || ae.Status == models.AEStatusProcessing {
			res.Pending = append(res.Pending, *ae)
		}
	}
	return res, nil
}
