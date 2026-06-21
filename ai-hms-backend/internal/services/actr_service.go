package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/actrs"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

const ExternalSystemACTRS = "ACTRS"

type ActrService struct {
	db       *gorm.DB
	client   *actrs.Client
	enabled  bool
	tenantID int64
	mapping  *ExternalPatientMappingService
}

func NewActrService(cfg actrs.Config, enabled bool, tenantID int64) *ActrService {
	return NewActrServiceWith(database.GetDB(), actrs.NewClient(cfg), enabled, tenantID)
}

func NewActrServiceWith(db *gorm.DB, client *actrs.Client, enabled bool, tenantID int64) *ActrService {
	return &ActrService{db: db, client: client, enabled: enabled, tenantID: tenantID, mapping: &ExternalPatientMappingService{db: db}}
}

var ErrActrsDisabled = errors.New("ACTRS 未启用")
var ErrPrescriptionSigned = errors.New("处方已签发，无法从 ACTR 采纳到草稿")
var ErrNoDialysisNo = errors.New("该患者无透析号，无法关联影像")

func (s *ActrService) ensurePatientMapping(ctx context.Context, legacyPatientID int64) (int64, string, error) {
	var p struct {
		Name       string `gorm:"column:Name"`
		DialysisNo string `gorm:"column:DialysisNo"`
	}
	err := s.db.Table(`"Register_PatientInfomation"`).
		Select(`"Name", "DialysisNo"`).
		Where(`"Id" = ? AND "TenantId" = ?`, legacyPatientID, s.tenantID).
		First(&p).Error
	if err != nil {
		return 0, "", fmt.Errorf("患者不存在: %w", err)
	}
	dialysisNo := strings.TrimSpace(p.DialysisNo)
	if dialysisNo == "" {
		return 0, "", ErrNoDialysisNo
	}

	var existing models.ExternalPatientMapping
	e := s.db.Where("tenant_id = ? AND external_system = ? AND legacy_patient_id = ?",
		s.tenantID, ExternalSystemACTRS, legacyPatientID).First(&existing).Error
	if e == nil {
		actrsID, _ := strconv.ParseInt(existing.ExternalPatientID, 10, 64)
		return actrsID, dialysisNo, nil
	}
	if !errors.Is(e, gorm.ErrRecordNotFound) {
		return 0, "", e
	}

	if !s.enabled {
		return 0, "", ErrActrsDisabled
	}

	out, err := s.client.UpsertPatient(ctx, actrs.PatientCreate{DialysisID: dialysisNo, Name: p.Name})
	if err != nil {
		return 0, "", fmt.Errorf("ACTRS 建档失败: %w", err)
	}
	m := &models.ExternalPatientMapping{
		ID:               utils.GenerateID(),
		TenantID:         s.tenantID,
		LegacyPatientID:  legacyPatientID,
		ExternalSystem:   ExternalSystemACTRS,
		ExternalPatientID: strconv.FormatInt(out.ID, 10),
		DialysisNo:       &dialysisNo,
		MatchStatus:      models.MatchStatusConfirmed,
	}
	if err := s.mapping.CreateMapping(m); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") || errors.Is(err, gorm.ErrDuplicatedKey) {
			var fallback models.ExternalPatientMapping
			if e2 := s.db.Where("tenant_id = ? AND external_system = ? AND legacy_patient_id = ?",
				s.tenantID, ExternalSystemACTRS, legacyPatientID).First(&fallback).Error; e2 == nil {
				actrsID, _ := strconv.ParseInt(fallback.ExternalPatientID, 10, 64)
				return actrsID, dialysisNo, nil
			}
		}
		return 0, "", err
	}
	return out.ID, dialysisNo, nil
}

func (s *ActrService) Status() map[string]bool {
	return map[string]bool{
		"enabled":    s.enabled,
		"configured": s.client != nil,
	}
}

func (s *ActrService) History(legacyPatientID int64) ([]models.PatientACTR, error) {
	pidStr := strconv.FormatInt(legacyPatientID, 10)
	var rows []models.PatientACTR
	err := s.db.Where("tenant_id = ? AND patient_id = ?", s.tenantID, pidStr).
		Order("analysis_date DESC, created_at DESC").Find(&rows).Error
	return rows, err
}

func (s *ActrService) Analyze(ctx context.Context, legacyPatientID int64, filename string, file io.Reader) (*models.PatientACTR, error) {
	if !s.enabled {
		return nil, ErrActrsDisabled
	}
	actrsID, dialysisNo, err := s.ensurePatientMapping(ctx, legacyPatientID)
	if err != nil {
		return nil, err
	}
	out, err := s.client.AnalyzeXray(ctx, actrsID, filename, file)
	if err != nil {
		return nil, fmt.Errorf("ACTRS 分析失败: %w", err)
	}
	rec := actrs.MapXrayToPatientACTR(*out, s.tenantID, strconv.FormatInt(legacyPatientID, 10), dialysisNo, "manual")
	return s.upsertPatientACTR(&rec)
}

func (s *ActrService) upsertPatientACTR(rec *models.PatientACTR) (*models.PatientACTR, error) {
	var existing models.PatientACTR
	err := s.db.Where("tenant_id = ? AND patient_id = ? AND actrs_xray_id = ?",
		rec.TenantID, rec.PatientID, rec.ActrsXrayID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		rec.ID = utils.GenerateID()
		if err := s.db.Create(rec).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") || errors.Is(err, gorm.ErrDuplicatedKey) {
				_ = s.db.Where("tenant_id = ? AND patient_id = ? AND actrs_xray_id = ?",
					rec.TenantID, rec.PatientID, rec.ActrsXrayID).First(&existing)
				return &existing, nil
			}
			return nil, err
		}
		return rec, nil
	}
	if err != nil {
		return nil, err
	}
	now := time.Now()
	updates := map[string]any{
		"ctr": rec.CTR, "actr": rec.ACTR, "actr_norm": rec.ACTRNorm,
		"qc_pass": rec.QCPass, "qc_pa_ap": rec.QCPaAp, "qc_warnings": rec.QCWarnings,
		"model_version": rec.ModelVersion, "analysis_date": rec.AnalysisDate,
		"image_path": rec.ImagePath, "overlay_path": rec.OverlayPath, "mask_path": rec.MaskPath,
		"synced_at": &now, "updated_at": now,
	}
	if err := s.db.Model(&existing).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *ActrService) AdoptToPrescription(ctx context.Context, legacyPatientID, prescriptionID int64, actrRecordID string, dryWeight, ufQuantity *float64, adoptedBy string) error {
	if !s.enabled {
		return ErrActrsDisabled
	}
	var rec models.PatientACTR
	pidStr := strconv.FormatInt(legacyPatientID, 10)
	if err := s.db.Where("tenant_id = ? AND id = ? AND patient_id = ?", s.tenantID, actrRecordID, pidStr).First(&rec).Error; err != nil {
		return errors.New("ACTR 记录不存在")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table(`"Plan_PatientPrescription"`).
			Where(`"Id" = ? AND "TenantId" = ? AND "PatientId" = ? AND "ConfirmTime" IS NULL`,
				prescriptionID, s.tenantID, legacyPatientID).
			Updates(map[string]any{
				"DryWeight":     dryWeight,
				"UFQuantity":    ufQuantity,
				"LastModifyTime": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			var check struct{ ConfirmTime *time.Time `gorm:"column:ConfirmTime"` }
			if err := tx.Table(`"Plan_PatientPrescription"`).Where(`"Id" = ?`, prescriptionID).Select(`"ConfirmTime"`).Scan(&check).Error; err == nil && check.ConfirmTime != nil && !check.ConfirmTime.IsZero() {
				return ErrPrescriptionSigned
			}
			return errors.New("当日处方草稿不存在，无法采纳")
		}

		now := time.Now()
		if err := tx.Model(&rec).Updates(map[string]any{
			"adopted_by":              adoptedBy,
			"adopted_at":              &now,
			"adopted_prescription_id": strconv.FormatInt(prescriptionID, 10),
			"adopted_dry_weight":      dryWeight,
			"adopted_uf_quantity":     ufQuantity,
			"updated_at":              now,
		}).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func (s *ActrService) Correct(ctx context.Context, legacyPatientID int64, actrRecordID, correctedBy string, value float64) (*models.PatientACTR, error) {
	if !s.enabled {
		return nil, ErrActrsDisabled
	}
	var rec models.PatientACTR
	pidStr := strconv.FormatInt(legacyPatientID, 10)
	if err := s.db.Where("tenant_id = ? AND id = ? AND patient_id = ?", s.tenantID, actrRecordID, pidStr).First(&rec).Error; err != nil {
		return nil, errors.New("ACTR 记录不存在")
	}
	if _, err := s.client.ApplyCorrection(ctx, rec.ActrsXrayID, actrs.CorrectionRequest{DoctorCorrection: value}); err != nil {
		return nil, fmt.Errorf("ACTRS 矫正回流失败: %w", err)
	}
	now := time.Now()
	if err := s.db.Model(&rec).Updates(map[string]any{
		"doctor_correction": value,
		"corrected_by":      correctedBy,
		"corrected_at":      &now,
		"updated_at":        now,
	}).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}
