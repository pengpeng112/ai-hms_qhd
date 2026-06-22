package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

// NursingDocService 护理文书（C1）：量表评估算分 + 护理记录 + 护理计划。
type NursingDocService struct {
	db       *gorm.DB
	tenantID int64
	scales   []config.NursingScale
}

func NewNursingDocService() *NursingDocService {
	scales, _ := config.LoadNursingScales()
	return &NursingDocService{db: database.GetDB(), tenantID: LegacyTenantID, scales: scales}
}

// EnabledScales 返回启用的量表定义（供前端渲染录入表单）
func (s *NursingDocService) EnabledScales() []config.NursingScale {
	out := make([]config.NursingScale, 0, len(s.scales))
	for _, sc := range s.scales {
		if sc.Enabled {
			out = append(out, sc)
		}
	}
	return out
}

func (s *NursingDocService) findScale(scaleType string) *config.NursingScale {
	for i := range s.scales {
		if s.scales[i].ScaleType == scaleType {
			return &s.scales[i]
		}
	}
	return nil
}

// ── 量表评估 ──────────────────────────────────────────────

type ScaleRecordInput struct {
	PatientID   string         `json:"patientId"`
	TreatmentID string         `json:"treatmentId"`
	ScaleType   string         `json:"scaleType"`
	Items       map[string]int `json:"items"` // 条目key → 选中分值
	NurseID     string         `json:"nurseId"`
	NurseName   string         `json:"nurseName"`
}

func (s *NursingDocService) RecordScale(in ScaleRecordInput) (*models.NursingDoc, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	scale := s.findScale(in.ScaleType)
	if scale == nil || !scale.Enabled {
		return nil, errors.New("未知或未启用的量表")
	}
	if strings.TrimSpace(in.PatientID) == "" {
		return nil, errors.New("患者必填")
	}
	if len(in.Items) == 0 {
		return nil, errors.New("量表条目必填")
	}
	// 服务端按配置重算总分（不信任前端总分）
	total := 0
	for _, item := range scale.Items {
		v, ok := in.Items[item.Key]
		if !ok {
			return nil, errors.New("量表条目不完整：" + item.Label)
		}
		if !validOptionValue(item, v) {
			return nil, errors.New("非法条目取值：" + item.Label)
		}
		total += v
	}
	risk := models.NursingRiskNone
	if band := scale.ScoreBand(total); band != nil {
		risk = band.Level
	}
	contentJSON, _ := json.Marshal(in.Items)
	now := time.Now()
	rec := &models.NursingDoc{
		ID: utils.GenerateID(), TenantID: s.tenantID, PatientID: in.PatientID, TreatmentID: in.TreatmentID,
		DocType: models.NursingDocScale, ScaleType: in.ScaleType, Score: &total, RiskLevel: risk,
		Content: string(contentJSON), NurseID: in.NurseID, NurseName: in.NurseName, RecordedAt: &now,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

func validOptionValue(item config.NursingScaleItem, v int) bool {
	for _, o := range item.Options {
		if o.Value == v {
			return true
		}
	}
	return false
}

// ── 护理记录 / 护理计划 ────────────────────────────────────

type DocRecordInput struct {
	PatientID   string `json:"patientId"`
	TreatmentID string `json:"treatmentId"`
	DocType     string `json:"docType"` // record / plan
	Content     string `json:"content"` // 前端拼好的 JSON（护理观察/操作/宣教/交班 或 问题-措施-评价）
	NurseID     string `json:"nurseId"`
	NurseName   string `json:"nurseName"`
}

func (s *NursingDocService) RecordDoc(in DocRecordInput) (*models.NursingDoc, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if in.DocType != models.NursingDocRecord && in.DocType != models.NursingDocPlan {
		return nil, errors.New("非法文书类型")
	}
	if strings.TrimSpace(in.PatientID) == "" {
		return nil, errors.New("患者必填")
	}
	if strings.TrimSpace(in.Content) == "" {
		return nil, errors.New("内容必填")
	}
	now := time.Now()
	rec := &models.NursingDoc{
		ID: utils.GenerateID(), TenantID: s.tenantID, PatientID: in.PatientID, TreatmentID: in.TreatmentID,
		DocType: in.DocType, Content: in.Content, NurseID: in.NurseID, NurseName: in.NurseName, RecordedAt: &now,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

// ── 查询 ──────────────────────────────────────────────────

type NursingListFilter struct {
	PatientID   string
	TreatmentID string
	DocType     string
	ScaleType   string
}

func (s *NursingDocService) List(f NursingListFilter) ([]models.NursingDoc, error) {
	q := s.db.Where("tenant_id = ?", s.tenantID)
	if f.PatientID != "" {
		q = q.Where("patient_id = ?", f.PatientID)
	}
	if f.TreatmentID != "" {
		q = q.Where("treatment_id = ?", f.TreatmentID)
	}
	if f.DocType != "" {
		q = q.Where("doc_type = ?", f.DocType)
	}
	if f.ScaleType != "" {
		q = q.Where("scale_type = ?", f.ScaleType)
	}
	var rows []models.NursingDoc
	err := q.Order("recorded_at DESC, created_at DESC").Find(&rows).Error
	return rows, err
}

// HighRiskAlerts 高风险量表（每患者每量表取最新一条，风险=high 的）→ 驾驶舱护士墙
func (s *NursingDocService) HighRiskAlerts() ([]models.NursingDoc, error) {
	var rows []models.NursingDoc
	// 取所有量表评估按时间倒序，内存里每 (patient,scale) 保留最新，再筛 high
	err := s.db.Where("tenant_id = ? AND doc_type = ?", s.tenantID, models.NursingDocScale).
		Order("recorded_at DESC, created_at DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]models.NursingDoc, 0)
	for _, r := range rows {
		key := r.PatientID + "|" + r.ScaleType
		if seen[key] {
			continue
		}
		seen[key] = true
		if r.RiskLevel == models.NursingRiskHigh {
			out = append(out, r)
		}
	}
	return out, nil
}
