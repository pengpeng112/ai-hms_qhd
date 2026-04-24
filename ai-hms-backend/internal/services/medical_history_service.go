package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"gorm.io/gorm"
)

type MedicalHistoryService struct {
	db *gorm.DB
}

type legacyMedicalHistory struct {
	ID                         modeltypes.LegacyID `gorm:"column:Id"`
	TenantID                   int64               `gorm:"column:TenantId"`
	PatientID                  modeltypes.LegacyID `gorm:"column:PatientId"`
	Complaints                 string              `gorm:"column:Complaints"`
	PresentIllnessHistory      string              `gorm:"column:PresentIllnessHistory"`
	PastIllnessHistory         string              `gorm:"column:PastIllnessHistory"`
	PersonalHistory            string              `gorm:"column:PersonalHistory"`
	MaritalReproductiveHistory string              `gorm:"column:MaritalReproductiveHistory"`
	FamilyHistory              string              `gorm:"column:FamilyHistory"`
	DiagnosisDesc              string              `gorm:"column:DiagnosisDesc"`
	PhysicalExamination        string              `gorm:"column:PhysicalExamination"`
	SpecialistExamination      string              `gorm:"column:SpecialistExamination"`
	AncillaryExamination       string              `gorm:"column:AncillaryExamination"`
	Note                       string              `gorm:"column:Note"`
	CreatorID                  int64               `gorm:"column:CreatorId"`
	CreateTime                 time.Time           `gorm:"column:CreateTime"`
	LastModifyTime             time.Time           `gorm:"column:LastModifyTime"`
	Narrator                   string              `gorm:"column:Narrator"`
}

func (legacyMedicalHistory) TableName() string { return "Register_MedicalHistory" }

type legacyHistoryNamedRow struct {
	ID             modeltypes.LegacyID `gorm:"column:Id"`
	TenantID       int64               `gorm:"column:TenantId"`
	PatientID      modeltypes.LegacyID `gorm:"column:PatientId"`
	Type           string              `gorm:"column:Type"`
	Name           string              `gorm:"column:Name"`
	Description    string              `gorm:"column:Description"`
	TreatmentDesc  string              `gorm:"column:TreatmentDesc"`
	ExamineTime    *time.Time          `gorm:"column:ExamineTime"`
	ExamineDr      string              `gorm:"column:ExamineDr"`
	Note           string              `gorm:"column:Note"`
	CreatorID      int64               `gorm:"column:CreatorId"`
	CreateTime     time.Time           `gorm:"column:CreateTime"`
	LastModifyTime time.Time           `gorm:"column:LastModifyTime"`
}

type legacyOutcomeRecord struct {
	ID             modeltypes.LegacyID `gorm:"column:Id"`
	TenantID       int64               `gorm:"column:TenantId"`
	PatientID      modeltypes.LegacyID `gorm:"column:PatientId"`
	Type           string              `gorm:"column:Type"`
	Reason         string              `gorm:"column:Reason"`
	OutComeTime    time.Time           `gorm:"column:OutComeTime"`
	Note           string              `gorm:"column:Note"`
	CreatorID      int64               `gorm:"column:CreatorId"`
	CreateTime     time.Time           `gorm:"column:CreateTime"`
	LastModifyTime time.Time           `gorm:"column:LastModifyTime"`
}

func (legacyOutcomeRecord) TableName() string { return "Register_OutCome" }

type legacyJsonDataRow struct {
	ID             int64               `gorm:"column:Id"`
	TenantID       int64               `gorm:"column:TenantId"`
	PatientID      modeltypes.LegacyID `gorm:"column:PatientId"`
	TreatmentID    int64               `gorm:"column:TreatmentId"`
	Code           string              `gorm:"column:Code"`
	CreatorID      int64               `gorm:"column:CreatorId"`
	CreateTime     time.Time           `gorm:"column:CreateTime"`
	LastModifyTime time.Time           `gorm:"column:LastModifyTime"`
	Value          json.RawMessage     `gorm:"column:Value"`
}

func (legacyJsonDataRow) TableName() string { return "Auxiliary_JsonData" }

const legacyJSONCodeBloodTransfusion = "hp_bloodtransfusion_history"

func NewMedicalHistoryService() *MedicalHistoryService {
	return &MedicalHistoryService{
		db: database.GetDB(),
	}
}

type MedicalHistoryResponse struct {
	Current      HistoryContent      `json:"current"`
	Past         HistoryContent      `json:"past"`
	Transfusion  HistoryContent      `json:"transfusion"`
	Marital      HistoryContent      `json:"marital"`
	Family       HistoryContent      `json:"family"`
	Diagnosis    HistoryContent      `json:"diagnosis"`
	Primary      HistoryNamedContent `json:"primary"`
	Pathology    HistoryNamedContent `json:"pathology"`
	Allergen     HistoryNamedContent `json:"allergen"`
	Tumor        HistoryNamedContent `json:"tumor"`
	Complication HistoryNamedContent `json:"complication"`
}

type HistoryContent struct {
	Content string `json:"content"`
}

type HistoryNamedContent struct {
	Name        string `json:"name"`
	Content     string `json:"content"`
	Type        string `json:"type,omitempty"`
	CheckTime   string `json:"checkTime,omitempty"`
	CheckDoctor string `json:"checkDoctor,omitempty"`
}

type MedicalHistoryRequest struct {
	Current      *HistoryContent      `json:"current"`
	Past         *HistoryContent      `json:"past"`
	Transfusion  *HistoryContent      `json:"transfusion"`
	Marital      *HistoryContent      `json:"marital"`
	Family       *HistoryContent      `json:"family"`
	Diagnosis    *HistoryContent      `json:"diagnosis"`
	Primary      *HistoryNamedContent `json:"primary"`
	Pathology    *HistoryNamedContent `json:"pathology"`
	Allergen     *HistoryNamedContent `json:"allergen"`
	Tumor        *HistoryNamedContent `json:"tumor"`
	Complication *HistoryNamedContent `json:"complication"`
}

type OutcomeRecordResponse struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Reason           string `json:"reason"`
	Time             string `json:"time"`
	Remarks          string `json:"remarks"`
	Registrar        string `json:"registrar"`
	RegistrationTime string `json:"registrationTime"`
	IsDoorRule       bool   `json:"isDoorRule"`
}

type OutcomeRecordRequest struct {
	Type             string `json:"type" binding:"required"`
	Reason           string `json:"reason"`
	Time             string `json:"time" binding:"required"`
	Remarks          string `json:"remarks"`
	Registrar        string `json:"registrar"`
	RegistrationTime string `json:"registrationTime"`
	IsDoorRule       bool   `json:"isDoorRule"`
}

func (s *MedicalHistoryService) ensurePatientExists(patientID modeltypes.LegacyID) error {
	var count int64
	if err := s.db.Model(&models.Patient{}).
		Where(`"Id" = ? AND "TenantId" = ?`, patientID, legacyTenantID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("patient not found")
	}
	return nil
}

func (s *MedicalHistoryService) latestNamedHistory(table string, patientID modeltypes.LegacyID) (*legacyHistoryNamedRow, error) {
	var row legacyHistoryNamedRow
	err := s.db.Table(table).
		Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, legacyTenantID).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func formatHistoryCheckTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func buildNamedContent(row *legacyHistoryNamedRow, content string) HistoryNamedContent {
	if row == nil {
		return HistoryNamedContent{}
	}
	return HistoryNamedContent{
		Name:        strings.TrimSpace(row.Name),
		Content:     strings.TrimSpace(content),
		Type:        strings.TrimSpace(row.Type),
		CheckTime:   formatHistoryCheckTime(row.ExamineTime),
		CheckDoctor: strings.TrimSpace(row.ExamineDr),
	}
}

func (s *MedicalHistoryService) GetMedicalHistory(patientID string) (*MedicalHistoryResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}
	if err := s.ensurePatientExists(legacyPatientID); err != nil {
		return nil, err
	}

	resp := &MedicalHistoryResponse{}

	var main legacyMedicalHistory
	err = s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, legacyPatientID, legacyTenantID).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		First(&main).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		resp.Current = HistoryContent{Content: strings.TrimSpace(main.PresentIllnessHistory)}
		resp.Past = HistoryContent{Content: strings.TrimSpace(main.PastIllnessHistory)}
		resp.Transfusion = HistoryContent{Content: strings.TrimSpace(main.PersonalHistory)}
		resp.Marital = HistoryContent{Content: strings.TrimSpace(main.MaritalReproductiveHistory)}
		resp.Family = HistoryContent{Content: strings.TrimSpace(main.FamilyHistory)}
		resp.Diagnosis = HistoryContent{Content: strings.TrimSpace(main.DiagnosisDesc)}
	}
	if transfusion, transfusionErr := s.getBloodTransfusionHistoryFromJSON(legacyPatientID); transfusionErr != nil {
		return nil, transfusionErr
	} else if strings.TrimSpace(transfusion) != "" {
		resp.Transfusion = HistoryContent{Content: transfusion}
	}

	primary, err := s.latestNamedHistory(`"Register_Protopathy"`, legacyPatientID)
	if err != nil {
		return nil, err
	}
	resp.Primary = buildNamedContent(primary, firstNonEmptyText(primaryValue(primary, "Note"), primaryValue(primary, "Description")))

	pathology, err := s.latestNamedHistory(`"Register_Pathology"`, legacyPatientID)
	if err != nil {
		return nil, err
	}
	resp.Pathology = buildNamedContent(pathology, firstNonEmptyText(primaryValue(pathology, "Description"), primaryValue(pathology, "Note")))

	allergen, err := s.latestNamedHistory(`"Register_Allergen"`, legacyPatientID)
	if err != nil {
		return nil, err
	}
	resp.Allergen = buildNamedContent(allergen, strings.TrimSpace(primaryValue(allergen, "Note")))

	tumor, err := s.latestNamedHistory(`"Register_Tumor"`, legacyPatientID)
	if err != nil {
		return nil, err
	}
	resp.Tumor = buildNamedContent(tumor, firstNonEmptyText(primaryValue(tumor, "TreatmentDesc"), primaryValue(tumor, "Note")))

	complication, err := s.latestNamedHistory(`"Register_Complication"`, legacyPatientID)
	if err != nil {
		return nil, err
	}
	resp.Complication = buildNamedContent(complication, firstNonEmptyText(primaryValue(complication, "Description"), primaryValue(complication, "Note")))

	return resp, nil
}

func primaryValue(row *legacyHistoryNamedRow, field string) string {
	if row == nil {
		return ""
	}
	switch field {
	case "Description":
		return row.Description
	case "TreatmentDesc":
		return row.TreatmentDesc
	case "Note":
		return row.Note
	default:
		return ""
	}
}

func (s *MedicalHistoryService) SaveMedicalHistory(patientID string, req *MedicalHistoryRequest) (*MedicalHistoryResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}
	if err := s.ensurePatientExists(legacyPatientID); err != nil {
		return nil, err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var main legacyMedicalHistory
		findErr := tx.Where(`"PatientId" = ? AND "TenantId" = ?`, legacyPatientID, legacyTenantID).
			Order(`"LastModifyTime" DESC`).
			Order(`"CreateTime" DESC`).
			Order(`"Id" DESC`).
			First(&main).Error
		isNew := errors.Is(findErr, gorm.ErrRecordNotFound)
		if findErr != nil && !isNew {
			return findErr
		}
		now := time.Now()
		if isNew {
			id, idErr := nextLegacyID()
			if idErr != nil {
				return idErr
			}
			main = legacyMedicalHistory{
				ID:         id,
				TenantID:   legacyTenantID,
				PatientID:  legacyPatientID,
				CreateTime: now,
			}
		}
		if req.Current != nil {
			main.PresentIllnessHistory = req.Current.Content
		}
		if req.Past != nil {
			main.PastIllnessHistory = req.Past.Content
		}
		if req.Transfusion != nil {
			main.PersonalHistory = req.Transfusion.Content
			if err := s.upsertBloodTransfusionHistoryJSON(tx, legacyPatientID, 0, req.Transfusion.Content); err != nil {
				return err
			}
		}
		if req.Marital != nil {
			main.MaritalReproductiveHistory = req.Marital.Content
		}
		if req.Family != nil {
			main.FamilyHistory = req.Family.Content
		}
		if req.Diagnosis != nil {
			main.DiagnosisDesc = req.Diagnosis.Content
		}
		main.LastModifyTime = now

		if isNew {
			if err := tx.Table(`"Register_MedicalHistory"`).Create(&main).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Table(`"Register_MedicalHistory"`).
				Where(`"Id" = ?`, main.ID).
				Updates(map[string]any{
					"PresentIllnessHistory":      main.PresentIllnessHistory,
					"PastIllnessHistory":         main.PastIllnessHistory,
					"PersonalHistory":            main.PersonalHistory,
					"MaritalReproductiveHistory": main.MaritalReproductiveHistory,
					"FamilyHistory":              main.FamilyHistory,
					"DiagnosisDesc":              main.DiagnosisDesc,
					"LastModifyTime":             main.LastModifyTime,
				}).Error; err != nil {
				return err
			}
		}

		if req.Primary != nil {
			if err := s.upsertNamedHistory(tx, `"Register_Protopathy"`, legacyPatientID, req.Primary, "Note"); err != nil {
				return err
			}
		}
		if req.Pathology != nil {
			if err := s.upsertNamedHistory(tx, `"Register_Pathology"`, legacyPatientID, req.Pathology, "Description"); err != nil {
				return err
			}
		}
		if req.Allergen != nil {
			if err := s.upsertNamedHistory(tx, `"Register_Allergen"`, legacyPatientID, req.Allergen, "Note"); err != nil {
				return err
			}
		}
		if req.Tumor != nil {
			if err := s.upsertNamedHistory(tx, `"Register_Tumor"`, legacyPatientID, req.Tumor, "TreatmentDesc"); err != nil {
				return err
			}
		}
		if req.Complication != nil {
			if err := s.upsertNamedHistory(tx, `"Register_Complication"`, legacyPatientID, req.Complication, "Description"); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetMedicalHistory(patientID)
}

func (s *MedicalHistoryService) getBloodTransfusionHistoryFromJSON(patientID modeltypes.LegacyID) (string, error) {
	var row legacyJsonDataRow
	err := s.db.Table(`"Auxiliary_JsonData"`).
		Where(`"TenantId" = ? AND "PatientId" = ? AND "Code" = ?`, legacyTenantID, patientID, legacyJSONCodeBloodTransfusion).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return parseBloodTransfusionHistoryValue(row.Value), nil
}

func parseBloodTransfusionHistoryValue(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return ""
	}
	var asObject struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &asObject); err == nil && strings.TrimSpace(asObject.Content) != "" {
		return strings.TrimSpace(asObject.Content)
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return strings.TrimSpace(asString)
	}
	return strings.Trim(strings.TrimSpace(trimmed), `"`)
}

func (s *MedicalHistoryService) upsertBloodTransfusionHistoryJSON(tx *gorm.DB, patientID modeltypes.LegacyID, creatorID int64, content string) error {
	now := time.Now()
	payload, err := json.Marshal(map[string]string{
		"content": strings.TrimSpace(content),
	})
	if err != nil {
		return err
	}
	var existing legacyJsonDataRow
	findErr := tx.Table(`"Auxiliary_JsonData"`).
		Where(`"TenantId" = ? AND "PatientId" = ? AND "Code" = ?`, legacyTenantID, patientID, legacyJSONCodeBloodTransfusion).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		First(&existing).Error
	if findErr == nil {
		return tx.Table(`"Auxiliary_JsonData"`).
			Where(`"Id" = ?`, existing.ID).
			Updates(map[string]any{
				"Value":          json.RawMessage(payload),
				"LastModifyTime": now,
			}).Error
	}
	if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return findErr
	}
	newID, idErr := nextLegacyID()
	if idErr != nil {
		return idErr
	}
	row := map[string]any{
		"Id":             newID,
		"TenantId":       legacyTenantID,
		"PatientId":      patientID,
		"TreatmentId":    0,
		"Code":           legacyJSONCodeBloodTransfusion,
		"CreatorId":      creatorID,
		"CreateTime":     now,
		"LastModifyTime": now,
		"Value":          json.RawMessage(payload),
	}
	return tx.Table(`"Auxiliary_JsonData"`).Create(row).Error
}

func (s *MedicalHistoryService) upsertNamedHistory(tx *gorm.DB, table string, patientID modeltypes.LegacyID, req *HistoryNamedContent, contentColumn string) error {
	var existing legacyHistoryNamedRow
	findErr := tx.Table(table).
		Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, legacyTenantID).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		First(&existing).Error
	isNew := errors.Is(findErr, gorm.ErrRecordNotFound)
	if findErr != nil && !isNew {
		return findErr
	}

	now := time.Now()
	values := map[string]any{
		"Type":           strings.TrimSpace(req.Type),
		"Name":           strings.TrimSpace(req.Name),
		contentColumn:    strings.TrimSpace(req.Content),
		"ExamineDr":      strings.TrimSpace(req.CheckDoctor),
		"LastModifyTime": now,
	}
	if strings.TrimSpace(req.CheckTime) != "" {
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(req.CheckTime)); err == nil {
			values["ExamineTime"] = t
		}
	}

	if isNew {
		id, err := nextLegacyID()
		if err != nil {
			return err
		}
		values["Id"] = id
		values["TenantId"] = legacyTenantID
		values["PatientId"] = patientID
		values["CreateTime"] = now
		return tx.Table(table).Create(values).Error
	}

	return tx.Table(table).Where(`"Id" = ?`, existing.ID).Updates(values).Error
}

func (s *MedicalHistoryService) ListOutcomeRecords(patientID string) ([]OutcomeRecordResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}

	var records []legacyOutcomeRecord
	if err := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, legacyPatientID, legacyTenantID).
		Order(`"OutComeTime" DESC`).
		Order(`"CreateTime" DESC`).
		Find(&records).Error; err != nil {
		return nil, err
	}

	result := make([]OutcomeRecordResponse, 0, len(records))
	for _, record := range records {
		result = append(result, s.buildOutcomeResponse(record))
	}
	return result, nil
}

func (s *MedicalHistoryService) CreateOutcomeRecord(patientID string, req *OutcomeRecordRequest) (*OutcomeRecordResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}
	if err := s.ensurePatientExists(legacyPatientID); err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-01-02 15:04", req.Time)
	if err != nil {
		return nil, errors.New("invalid time format, expected YYYY-MM-DD HH:mm")
	}

	recordID, err := nextLegacyID()
	if err != nil {
		return nil, err
	}

	creatorID, registrarName, err := (&PrescriptionService{db: s.db}).resolveLegacyUserID(req.Registrar, req.Registrar)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	row := legacyOutcomeRecord{
		ID:             recordID,
		TenantID:       legacyTenantID,
		PatientID:      legacyPatientID,
		Type:           strings.TrimSpace(req.Type),
		Reason:         strings.TrimSpace(req.Reason),
		OutComeTime:    t,
		Note:           strings.TrimSpace(req.Remarks),
		CreatorID:      creatorID,
		CreateTime:     now,
		LastModifyTime: now,
	}
	if err := s.db.Table(`"Register_OutCome"`).Create(&row).Error; err != nil {
		return nil, err
	}

	resp := s.buildOutcomeResponse(row)
	if registrarName != "" {
		resp.Registrar = registrarName
	}
	return &resp, nil
}

func (s *MedicalHistoryService) UpdateOutcomeRecord(patientID, recordID string, req *OutcomeRecordRequest) (*OutcomeRecordResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}
	legacyRecordID, err := parseLegacyID(recordID)
	if err != nil {
		return nil, errors.New("invalid outcome record id")
	}

	var record legacyOutcomeRecord
	err = s.db.Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, legacyRecordID, legacyPatientID, legacyTenantID).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("outcome record not found")
	}
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-01-02 15:04", req.Time)
	if err != nil {
		return nil, errors.New("invalid time format, expected YYYY-MM-DD HH:mm")
	}

	creatorID := record.CreatorID
	if strings.TrimSpace(req.Registrar) != "" {
		resolvedCreatorID, _, resolveErr := (&PrescriptionService{db: s.db}).resolveLegacyUserID(req.Registrar, req.Registrar)
		if resolveErr != nil {
			return nil, resolveErr
		}
		if resolvedCreatorID > 0 {
			creatorID = resolvedCreatorID
		}
	}

	record.Type = strings.TrimSpace(req.Type)
	record.Reason = strings.TrimSpace(req.Reason)
	record.OutComeTime = t
	record.Note = strings.TrimSpace(req.Remarks)
	record.CreatorID = creatorID
	record.LastModifyTime = time.Now()

	if err := s.db.Table(`"Register_OutCome"`).
		Where(`"Id" = ?`, record.ID).
		Updates(map[string]any{
			"Type":           record.Type,
			"Reason":         record.Reason,
			"OutComeTime":    record.OutComeTime,
			"Note":           record.Note,
			"CreatorId":      record.CreatorID,
			"LastModifyTime": record.LastModifyTime,
		}).Error; err != nil {
		return nil, err
	}

	resp := s.buildOutcomeResponse(record)
	return &resp, nil
}

func (s *MedicalHistoryService) DeleteOutcomeRecord(patientID, recordID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return errors.New("invalid patient id")
	}
	legacyRecordID, err := parseLegacyID(recordID)
	if err != nil {
		return errors.New("invalid outcome record id")
	}

	result := s.db.Table(`"Register_OutCome"`).
		Where(`"Id" = ? AND "PatientId" = ? AND "TenantId" = ?`, legacyRecordID, legacyPatientID, legacyTenantID).
		Delete(&legacyOutcomeRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("outcome record not found")
	}
	return nil
}

func (s *MedicalHistoryService) buildOutcomeResponse(r legacyOutcomeRecord) OutcomeRecordResponse {
	registrar, _ := (&PrescriptionService{db: s.db}).lookupLegacyUserDisplayName(r.CreatorID)
	return OutcomeRecordResponse{
		ID:               legacyIDString(r.ID),
		Type:             strings.TrimSpace(r.Type),
		Reason:           strings.TrimSpace(r.Reason),
		Time:             r.OutComeTime.Format("2006-01-02 15:04"),
		Remarks:          strings.TrimSpace(r.Note),
		Registrar:        registrar,
		RegistrationTime: r.CreateTime.Format("2006-01-02 15:04"),
		IsDoorRule:       false,
	}
}
