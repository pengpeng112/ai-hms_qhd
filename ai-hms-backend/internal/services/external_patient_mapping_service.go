package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

type ExternalPatientMappingService struct {
	db *gorm.DB
}

func NewExternalPatientMappingService() *ExternalPatientMappingService {
	return &ExternalPatientMappingService{db: database.GetDB()}
}

func (s *ExternalPatientMappingService) FindByExternal(
	externalSystem, externalPatientID string,
	externalVisitID *string,
) (*models.ExternalPatientMapping, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var m models.ExternalPatientMapping
	q := s.db.Where("external_system = ? AND external_patient_id = ?", externalSystem, externalPatientID)
	if externalVisitID != nil && *externalVisitID != "" {
		q = q.Where("external_visit_id = ?", *externalVisitID)
	} else {
		q = q.Where("external_visit_id IS NULL OR external_visit_id = ''")
	}
	if err := q.First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *ExternalPatientMappingService) FindByLegacyPatient(legacyPatientID int64) ([]models.ExternalPatientMapping, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var mappings []models.ExternalPatientMapping
	err := s.db.Where("legacy_patient_id = ?", legacyPatientID).Find(&mappings).Error
	return mappings, err
}

func (s *ExternalPatientMappingService) CreateMapping(mapping *models.ExternalPatientMapping) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	return s.db.Create(mapping).Error
}

func (s *ExternalPatientMappingService) ResolveLegacyPatientID(
	externalSystem, externalPatientID string,
	externalVisitID *string,
) (int64, error) {
	m, err := s.FindByExternal(externalSystem, externalPatientID, externalVisitID)
	if err != nil {
		return 0, err
	}
	if m != nil && m.MatchStatus == models.MatchStatusConfirmed {
		return m.LegacyPatientID, nil
	}
	return 0, fmt.Errorf("no confirmed mapping for %s/%s", externalSystem, externalPatientID)
}

func (s *ExternalPatientMappingService) AutoMatchByIDNo(
	externalSystem, externalPatientID string,
	idNo *string,
	tenantID int64,
) (*models.ExternalPatientMapping, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if idNo == nil || strings.TrimSpace(*idNo) == "" {
		return nil, nil
	}
	var row struct {
		PatientID int64  `gorm:"column:PatientId"`
		Name      string `gorm:"column:name"`
	}
	err := s.db.Table("Register_IDInfomation").
		Joins(`JOIN "Register_PatientInfomation" p ON p."Id" = "Register_IDInfomation"."PatientId"`).
		Where(`"Register_IDInfomation"."IDNo" = ? AND p."TenantId" = ? AND COALESCE("Register_IDInfomation"."IsDisabled", false) = false`,
			strings.TrimSpace(*idNo), tenantID).
		Select(`"Register_IDInfomation"."PatientId", p."Name" AS name`).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	m := &models.ExternalPatientMapping{
		ID:               fmt.Sprintf("epm_%s_%s", externalSystem, externalPatientID),
		TenantID:         tenantID,
		LegacyPatientID:  row.PatientID,
		ExternalSystem:   externalSystem,
		ExternalPatientID: externalPatientID,
		IDNo:             idNo,
		MatchStatus:      models.MatchStatusConfirmed,
	}
	return m, s.CreateMapping(m)
}
