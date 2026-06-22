package services

import (
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

// ConsentService 知情同意（C2）：开具 → 患者/家属签署（复用 sign_record）→ 到期复签。
type ConsentService struct {
	db        *gorm.DB
	tenantID  int64
	templates []config.ConsentTemplate
	signSvc   *SignService
}

func NewConsentService() *ConsentService {
	tpls, _ := config.LoadConsentTemplates()
	return &ConsentService{db: database.GetDB(), tenantID: LegacyTenantID, templates: tpls, signSvc: NewSignService()}
}

// EnabledTemplates 返回启用的同意书模板（供前端开具时选择）
func (s *ConsentService) EnabledTemplates() []config.ConsentTemplate {
	out := make([]config.ConsentTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		if t.Enabled {
			out = append(out, t)
		}
	}
	return out
}

func (s *ConsentService) findTemplate(consentType string) *config.ConsentTemplate {
	for i := range s.templates {
		if s.templates[i].ConsentType == consentType {
			return &s.templates[i]
		}
	}
	return nil
}

// ── 开具 ──────────────────────────────────────────────────

type ConsentIssueInput struct {
	PatientID   string `json:"patientId"`
	ConsentType string `json:"consentType"`
	IssuedBy    string `json:"issuedBy"`
	Note        string `json:"note"`
}

func (s *ConsentService) Issue(in ConsentIssueInput) (*models.ConsentRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	tpl := s.findTemplate(in.ConsentType)
	if tpl == nil || !tpl.Enabled {
		return nil, errors.New("未知或未启用的同意书类型")
	}
	if strings.TrimSpace(in.PatientID) == "" {
		return nil, errors.New("患者必填")
	}
	rec := &models.ConsentRecord{
		ID: utils.GenerateID(), TenantID: s.tenantID, PatientID: in.PatientID,
		ConsentType: in.ConsentType, TemplateVersion: tpl.Version, IssuedBy: in.IssuedBy,
		Status: models.ConsentStatusPending, Note: in.Note,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

// ── 签署（复用 sign_record）────────────────────────────────

type ConsentSignInput struct {
	SignedBy string `json:"signedBy"` // 患者本人 / 家属(关系-姓名)
	DocRef   string `json:"docRef"`   // 签署文档/影像
}

func (s *ConsentService) Sign(id string, in ConsentSignInput) (*models.ConsentRecord, error) {
	rec, err := s.get(id)
	if err != nil {
		return nil, err
	}
	if rec.Status == models.ConsentStatusRevoked {
		return nil, errors.New("已撤销的同意书不可签署")
	}
	signedBy := strings.TrimSpace(in.SignedBy)
	if signedBy == "" {
		return nil, errors.New("签署人必填")
	}
	// 复用统一电子签留痕：签署人=患者域（signerID 用患者ID关联，signerName=签署人描述）
	sign, err := s.signSvc.Sign(s.tenantID, models.SignTargetConsent, rec.ID, rec.PatientID, signedBy)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	upd := map[string]any{
		"signed_by": signedBy, "sign_record_id": sign.ID, "signed_at": &now,
		"status": models.ConsentStatusSigned, "updated_at": now,
	}
	if strings.TrimSpace(in.DocRef) != "" {
		upd["doc_ref"] = in.DocRef
	}
	if tpl := s.findTemplate(rec.ConsentType); tpl != nil && tpl.ValidMonths > 0 {
		exp := now.AddDate(0, tpl.ValidMonths, 0)
		upd["expires_at"] = &exp
	}
	if err := s.db.Model(&models.ConsentRecord{}).Where("id = ?", rec.ID).Updates(upd).Error; err != nil {
		return nil, err
	}
	return s.get(id)
}

func (s *ConsentService) Revoke(id string) (*models.ConsentRecord, error) {
	rec, err := s.get(id)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.ConsentRecord{}).Where("id = ?", rec.ID).
		Updates(map[string]any{"status": models.ConsentStatusRevoked, "updated_at": time.Now()}).Error; err != nil {
		return nil, err
	}
	return s.get(id)
}

func (s *ConsentService) get(id string) (*models.ConsentRecord, error) {
	var rec models.ConsentRecord
	if err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&rec).Error; err != nil {
		return nil, errors.New("同意书不存在")
	}
	return &rec, nil
}

// ── 查询 ──────────────────────────────────────────────────

type ConsentListFilter struct {
	PatientID   string
	ConsentType string
	Status      string
}

func (s *ConsentService) List(f ConsentListFilter) ([]models.ConsentRecord, error) {
	q := s.db.Where("tenant_id = ?", s.tenantID)
	if f.PatientID != "" {
		q = q.Where("patient_id = ?", f.PatientID)
	}
	if f.ConsentType != "" {
		q = q.Where("consent_type = ?", f.ConsentType)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var rows []models.ConsentRecord
	if err := q.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	applyConsentExpiry(rows)
	return rows, nil
}

// applyConsentExpiry 读取时把"已签但已过期"的视图状态翻成 expired（趋势提示用，不强制持久化）
func applyConsentExpiry(rows []models.ConsentRecord) {
	now := time.Now()
	for i := range rows {
		if rows[i].Status == models.ConsentStatusSigned && rows[i].ExpiresAt != nil && now.After(*rows[i].ExpiresAt) {
			rows[i].Status = models.ConsentStatusExpired
		}
	}
}

// Alerts 待办：待签 + 已过期（需复签）。供②带患者域/待办提示。
func (s *ConsentService) Alerts() (map[string][]models.ConsentRecord, error) {
	var rows []models.ConsentRecord
	if err := s.db.Where("tenant_id = ? AND status IN ?", s.tenantID,
		[]string{models.ConsentStatusPending, models.ConsentStatusSigned}).
		Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	applyConsentExpiry(rows)
	out := map[string][]models.ConsentRecord{"pending": {}, "expired": {}}
	for _, r := range rows {
		switch r.Status {
		case models.ConsentStatusPending:
			out["pending"] = append(out["pending"], r)
		case models.ConsentStatusExpired:
			out["expired"] = append(out["expired"], r)
		}
	}
	return out, nil
}
