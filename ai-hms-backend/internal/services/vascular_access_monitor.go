package services

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type VascularAccessMonitor struct {
	db       *gorm.DB
	tenantID int64
}

func NewVascularAccessMonitor() *VascularAccessMonitor {
	return &VascularAccessMonitor{db: database.GetDB(), tenantID: LegacyTenantID}
}

const (
	vascMaturationDueDays = 28
	vascEscalateDays      = 56
	vascNccLimitDays      = 28
	vascTccLimitDays      = 547
	vascPeriodicAvfAvg    = 90
	vascPeriodicTcc       = 30
)

var validVascEvents = map[string]struct{}{
	models.VAEEstablish: {}, models.VAEMaturation: {}, models.VAEFirstUse: {}, models.VAEPhysicalCheck: {},
	models.VAEComplication: {}, models.VAEIntervention: {}, models.VAEFailure: {}, models.VAEReplacement: {},
}

type VascEventInput struct {
	EventType  string    `json:"eventType"`
	EventDate  time.Time `json:"eventDate"`
	Detail     string    `json:"detail"`
	OperatorID string    `json:"operatorId"`
	Note       string    `json:"note"`
}

func (s *VascularAccessMonitor) RecordEvent(patientID, accessID int64, in VascEventInput) (*models.VascularAccessEvent, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if _, ok := validVascEvents[in.EventType]; !ok {
		return nil, errors.New("未知事件类型")
	}
	if strings.TrimSpace(in.Detail) != "" {
		var probe any
		if err := json.Unmarshal([]byte(in.Detail), &probe); err != nil {
			return nil, errors.New("事件数据格式错误")
		}
	}
	var cnt int64
	s.db.Raw(`SELECT count(*) FROM "Register_VascularAccess" WHERE "Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, accessID, patientID, s.tenantID).Scan(&cnt)
	if cnt == 0 {
		return nil, errors.New("通路不存在或不属于该患者")
	}
	ed := in.EventDate
	rec := &models.VascularAccessEvent{
		ID: utils.GenerateID(), TenantID: s.tenantID, AccessID: accessID, PatientID: patientID,
		EventType: in.EventType, EventDate: &ed, Detail: strings.TrimSpace(in.Detail),
		OperatorID: in.OperatorID, Note: strings.TrimSpace(in.Note),
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

func classifyAccess(t string) string {
	u := strings.ToUpper(t)
	switch {
	case strings.Contains(u, "NCC") || strings.Contains(t, "临时") || strings.Contains(t, "无隧道"):
		return "ncc"
	case strings.Contains(u, "TCC") || strings.Contains(t, "长期") || strings.Contains(t, "带隧道") || strings.Contains(t, "涤纶"):
		return "tcc"
	case strings.Contains(u, "AVF") || strings.Contains(u, "AVG") || strings.Contains(t, "内瘘") || strings.Contains(t, "移植"):
		return "avf_avg"
	default:
		return "other"
	}
}

type VascTimelineEntry struct {
	AccessID  int64      `json:"accessId"`
	EventType string     `json:"eventType"`
	EventDate *time.Time `json:"eventDate"`
	Detail    string     `json:"detail"`
	Note      string     `json:"note"`
}

type vascAccessRow struct {
	ID         int64      `gorm:"column:Id"`
	AccessType string     `gorm:"column:AccessType"`
	Operation  *time.Time `gorm:"column:OperationTime"`
}

func (s *VascularAccessMonitor) patientAccesses(patientID int64) []vascAccessRow {
	var rows []vascAccessRow
	s.db.Raw(`SELECT "Id","AccessType","OperationTime" FROM "Register_VascularAccess" WHERE "PatientId" = ? AND "TenantId" = ? AND COALESCE("IsDisabled", false) = false ORDER BY "OperationTime" DESC`, patientID, s.tenantID).Scan(&rows)
	return rows
}

func (s *VascularAccessMonitor) Timeline(patientID int64) ([]VascTimelineEntry, error) {
	accesses := s.patientAccesses(patientID)
	out := make([]VascTimelineEntry, 0)
	for _, a := range accesses {
		out = append(out, VascTimelineEntry{AccessID: a.ID, EventType: models.VAEEstablish, EventDate: a.Operation, Note: a.AccessType})
	}
	var evs []models.VascularAccessEvent
	s.db.Where("patient_id = ? AND tenant_id = ?", patientID, s.tenantID).
		Order("event_date DESC, created_at DESC").Find(&evs)
	for _, e := range evs {
		out = append(out, VascTimelineEntry{AccessID: e.AccessID, EventType: e.EventType, EventDate: e.EventDate, Detail: e.Detail, Note: e.Note})
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].EventDate, out[j].EventDate
		if a == nil && b == nil {
			return false
		}
		if a == nil {
			return false
		}
		if b == nil {
			return true
		}
		return a.After(*b)
	})
	return out, nil
}

type VascReminder struct {
	AccessID  int64  `json:"accessId"`
	PatientID int64  `json:"patientId"`
	Kind      string `json:"kind"`
	Message   string `json:"message"`
}

func (s *VascularAccessMonitor) currentAccess(patientID int64) *vascAccessRow {
	var row vascAccessRow
	err := s.db.Raw(`SELECT "Id","AccessType","OperationTime" FROM "Register_VascularAccess" WHERE "PatientId" = ? AND "TenantId" = ? AND COALESCE("IsDisabled", false) = false ORDER BY "IsDefault" DESC, "OperationTime" DESC LIMIT 1`, patientID, s.tenantID).Scan(&row).Error
	if err != nil || row.ID == 0 {
		return nil
	}
	return &row
}

func (s *VascularAccessMonitor) lastEventDate(accessID int64, eventType string) *time.Time {
	var t time.Time
	err := s.db.Model(&models.VascularAccessEvent{}).
		Select("event_date").
		Where("access_id = ? AND event_type = ? AND tenant_id = ?", accessID, eventType, s.tenantID).
		Order("event_date DESC").Limit(1).
		Scan(&t).Error
	if err != nil || t.IsZero() {
		return nil
	}
	return &t
}

func (s *VascularAccessMonitor) latestEventDate(accessID int64, eventTypes ...string) *time.Time {
	var latest *time.Time
	for _, et := range eventTypes {
		d := s.lastEventDate(accessID, et)
		if d != nil && (latest == nil || d.After(*latest)) {
			latest = d
		}
	}
	return latest
}

func (s *VascularAccessMonitor) PatientReminders(patientID int64) []VascReminder {
	acc := s.currentAccess(patientID)
	if acc == nil || acc.Operation == nil {
		return nil
	}
	now := time.Now()
	days := int(now.Sub(*acc.Operation).Hours() / 24)
	class := classifyAccess(acc.AccessType)
	var out []VascReminder
	add := func(kind, msg string) { out = append(out, VascReminder{AccessID: acc.ID, PatientID: patientID, Kind: kind, Message: msg}) }

	switch class {
	case "avf_avg":
		if days >= vascMaturationDueDays && s.lastEventDate(acc.ID, models.VAEMaturation) == nil {
			if days >= vascEscalateDays {
				add("maturation_due", "内瘘成熟评估已严重逾期(>8周未评估)")
			} else {
				add("maturation_due", "内瘘成熟评估到期，请评估可用性")
			}
		}
		latestEval := s.latestEventDate(acc.ID, models.VAEMaturation, models.VAEPhysicalCheck)
		if latestEval != nil && int(now.Sub(*latestEval).Hours()/24) > vascPeriodicAvfAvg {
			add("periodic_due", "内瘘定期评估到期(>3月)")
		} else if latestEval == nil && days >= vascPeriodicAvfAvg {
			add("periodic_due", "内瘘定期评估到期(>3月，从未评估)")
		}
	case "ncc":
		if days > vascNccLimitDays {
			add("cvc_over_limit", "临时导管已超时限(>2-4周)，建议转长期/内瘘")
		}
	case "tcc":
		if days > vascTccLimitDays {
			add("cvc_over_limit", "长期导管已超时限(>18月)，建议更换")
		}
		last := s.lastEventDate(acc.ID, models.VAEPhysicalCheck)
		if last != nil && int(now.Sub(*last).Hours()/24) > vascPeriodicTcc {
			add("periodic_due", "长期导管定期评估到期(>1月)")
		} else if last == nil && days > vascPeriodicTcc {
			add("periodic_due", "长期导管定期评估到期(>1月，从未评估)")
		}
	case "other":
		add("type_unrecognized", "通路类型无法识别，请确认通路分类")
	}
	if s.latestPhysicalAbnormal(acc.ID) {
		add("physical_abnormal", "最近物理检查异常(震颤减弱/杂音异常)")
	}
	fail := s.lastEventDate(acc.ID, models.VAEFailure)
	repl := s.lastEventDate(acc.ID, models.VAEReplacement)
	if fail != nil && (repl == nil || repl.Before(*fail)) {
		add("failure_no_replace", "通路失功未完成更换")
	}
	return out
}

func (s *VascularAccessMonitor) latestPhysicalAbnormal(accessID int64) bool {
	var ev models.VascularAccessEvent
	err := s.db.Where("access_id = ? AND event_type = ? AND tenant_id = ?", accessID, models.VAEPhysicalCheck, s.tenantID).
		Order("event_date DESC, created_at DESC").Limit(1).First(&ev).Error
	if err != nil {
		return false
	}
	var d struct {
		Abnormal bool `json:"abnormal"`
	}
	_ = json.Unmarshal([]byte(ev.Detail), &d)
	return d.Abnormal
}

func (s *VascularAccessMonitor) Alerts() ([]VascReminder, error) {
	var pids []int64
	if err := s.db.Raw(`SELECT DISTINCT "PatientId" FROM "Register_VascularAccess" WHERE "TenantId" = ? AND COALESCE("IsDisabled", false) = false`, s.tenantID).Pluck(`"PatientId"`, &pids).Error; err != nil {
		return nil, err
	}
	var out []VascReminder
	for _, pid := range pids {
		out = append(out, s.PatientReminders(pid)...)
	}
	return out, nil
}
