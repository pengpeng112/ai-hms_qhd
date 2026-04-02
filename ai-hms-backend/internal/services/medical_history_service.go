package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MedicalHistoryService 临床病史服务
type MedicalHistoryService struct {
	db *gorm.DB
}

// NewMedicalHistoryService 创建临床病史服务
func NewMedicalHistoryService() *MedicalHistoryService {
	return &MedicalHistoryService{
		db: database.GetDB(),
	}
}

// ========== 临床病史 ==========

// MedicalHistoryResponse 临床病史响应
type MedicalHistoryResponse struct {
	// 基础临床病史
	Current    HistoryContent `json:"current"`
	Past       HistoryContent `json:"past"`
	Transfusion HistoryContent `json:"transfusion"`
	Marital    HistoryContent `json:"marital"`
	Family     HistoryContent `json:"family"`
	Diagnosis  HistoryContent `json:"diagnosis"`

	// 专科记录
	Primary      HistoryNamedContent `json:"primary"`
	Pathology    HistoryNamedContent `json:"pathology"`
	Allergen     HistoryNamedContent `json:"allergen"`
	Tumor        HistoryNamedContent `json:"tumor"`
	Complication HistoryNamedContent `json:"complication"`
}

// HistoryContent 简单内容
type HistoryContent struct {
	Content string `json:"content"`
}

// HistoryNamedContent 带名称的内容
type HistoryNamedContent struct {
	Name        string `json:"name"`
	Content     string `json:"content"`
	Type        string `json:"type,omitempty"`
	CheckTime   string `json:"checkTime,omitempty"`
	CheckDoctor string `json:"checkDoctor,omitempty"`
}

// MedicalHistoryRequest 更新临床病史请求
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

// GetMedicalHistory 获取临床病史
func (s *MedicalHistoryService) GetMedicalHistory(patientID string) (*MedicalHistoryResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 验证患者存在
	var count int64
	s.db.Model(&models.Patient{}).Where("id = ?", patientID).Count(&count)
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	var history models.MedicalHistory
	err := s.db.Where("patient_id = ?", patientID).First(&history).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 返回空数据
		return &MedicalHistoryResponse{}, nil
	}
	if err != nil {
		return nil, err
	}

	return s.buildResponse(history), nil
}

// SaveMedicalHistory 保存/更新临床病史
func (s *MedicalHistoryService) SaveMedicalHistory(patientID string, req *MedicalHistoryRequest) (*MedicalHistoryResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 验证患者存在
	var count int64
	s.db.Model(&models.Patient{}).Where("id = ?", patientID).Count(&count)
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	var history models.MedicalHistory
	err := s.db.Where("patient_id = ?", patientID).First(&history).Error
	isNew := errors.Is(err, gorm.ErrRecordNotFound)

	if isNew {
		history = models.MedicalHistory{
			ID:        uuid.New().String(),
			PatientID: patientID,
		}
	} else if err != nil {
		return nil, err
	}

	// 更新字段
	s.applyRequest(&history, req)

	if isNew {
		err = s.db.Create(&history).Error
	} else {
		err = s.db.Save(&history).Error
	}
	if err != nil {
		return nil, err
	}

	return s.buildResponse(history), nil
}

func (s *MedicalHistoryService) applyRequest(h *models.MedicalHistory, req *MedicalHistoryRequest) {
	if req.Current != nil {
		h.CurrentIllness = req.Current.Content
	}
	if req.Past != nil {
		h.PastHistory = req.Past.Content
	}
	if req.Transfusion != nil {
		h.TransfusionHistory = req.Transfusion.Content
	}
	if req.Marital != nil {
		h.MaritalHistory = req.Marital.Content
	}
	if req.Family != nil {
		h.FamilyHistory = req.Family.Content
	}
	if req.Diagnosis != nil {
		h.DiseaseDiagnosis = req.Diagnosis.Content
	}
	if req.Primary != nil {
		h.PrimaryDiseaseName = req.Primary.Name
		h.PrimaryDiseaseContent = req.Primary.Content
		h.PrimaryDiseaseType = req.Primary.Type
		h.PrimaryDiseaseCheckTime = req.Primary.CheckTime
		h.PrimaryDiseaseCheckDoc = req.Primary.CheckDoctor
	}
	if req.Pathology != nil {
		h.PathologyName = req.Pathology.Name
		h.PathologyContent = req.Pathology.Content
		h.PathologyType = req.Pathology.Type
		h.PathologyCheckTime = req.Pathology.CheckTime
		h.PathologyCheckDoc = req.Pathology.CheckDoctor
	}
	if req.Allergen != nil {
		h.AllergenName = req.Allergen.Name
		h.AllergenContent = req.Allergen.Content
		h.AllergenType = req.Allergen.Type
		h.AllergenCheckTime = req.Allergen.CheckTime
		h.AllergenCheckDoc = req.Allergen.CheckDoctor
	}
	if req.Tumor != nil {
		h.TumorHistoryName = req.Tumor.Name
		h.TumorHistoryContent = req.Tumor.Content
		h.TumorHistoryType = req.Tumor.Type
		h.TumorHistoryCheckTime = req.Tumor.CheckTime
		h.TumorHistoryCheckDoc = req.Tumor.CheckDoctor
	}
	if req.Complication != nil {
		h.ComplicationName = req.Complication.Name
		h.ComplicationContent = req.Complication.Content
		h.ComplicationType = req.Complication.Type
		h.ComplicationCheckTime = req.Complication.CheckTime
		h.ComplicationCheckDoc = req.Complication.CheckDoctor
	}
}

func (s *MedicalHistoryService) buildResponse(h models.MedicalHistory) *MedicalHistoryResponse {
	return &MedicalHistoryResponse{
		Current:     HistoryContent{Content: h.CurrentIllness},
		Past:        HistoryContent{Content: h.PastHistory},
		Transfusion: HistoryContent{Content: h.TransfusionHistory},
		Marital:     HistoryContent{Content: h.MaritalHistory},
		Family:      HistoryContent{Content: h.FamilyHistory},
		Diagnosis:   HistoryContent{Content: h.DiseaseDiagnosis},
		Primary:     HistoryNamedContent{Name: h.PrimaryDiseaseName, Content: h.PrimaryDiseaseContent, Type: h.PrimaryDiseaseType, CheckTime: h.PrimaryDiseaseCheckTime, CheckDoctor: h.PrimaryDiseaseCheckDoc},
		Pathology:   HistoryNamedContent{Name: h.PathologyName, Content: h.PathologyContent, Type: h.PathologyType, CheckTime: h.PathologyCheckTime, CheckDoctor: h.PathologyCheckDoc},
		Allergen:    HistoryNamedContent{Name: h.AllergenName, Content: h.AllergenContent, Type: h.AllergenType, CheckTime: h.AllergenCheckTime, CheckDoctor: h.AllergenCheckDoc},
		Tumor:       HistoryNamedContent{Name: h.TumorHistoryName, Content: h.TumorHistoryContent, Type: h.TumorHistoryType, CheckTime: h.TumorHistoryCheckTime, CheckDoctor: h.TumorHistoryCheckDoc},
		Complication: HistoryNamedContent{Name: h.ComplicationName, Content: h.ComplicationContent, Type: h.ComplicationType, CheckTime: h.ComplicationCheckTime, CheckDoctor: h.ComplicationCheckDoc},
	}
}

// ========== 治疗转归记录 ==========

// OutcomeRecordResponse 转归记录响应
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

// OutcomeRecordRequest 创建/更新转归记录请求
type OutcomeRecordRequest struct {
	Type             string `json:"type" binding:"required"`
	Reason           string `json:"reason"`
	Time             string `json:"time" binding:"required"`
	Remarks          string `json:"remarks"`
	Registrar        string `json:"registrar"`
	RegistrationTime string `json:"registrationTime"`
	IsDoorRule       bool   `json:"isDoorRule"`
}

// ListOutcomeRecords 获取转归记录列表
func (s *MedicalHistoryService) ListOutcomeRecords(patientID string) ([]OutcomeRecordResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var records []models.OutcomeRecord
	err := s.db.Where("patient_id = ?", patientID).
		Order("time DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	result := make([]OutcomeRecordResponse, len(records))
	for i, r := range records {
		result[i] = s.buildOutcomeResponse(r)
	}
	return result, nil
}

// CreateOutcomeRecord 创建转归记录
func (s *MedicalHistoryService) CreateOutcomeRecord(patientID string, req *OutcomeRecordRequest) (*OutcomeRecordResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 验证患者存在
	var count int64
	s.db.Model(&models.Patient{}).Where("id = ?", patientID).Count(&count)
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	t, err := time.Parse("2006-01-02 15:04", req.Time)
	if err != nil {
		return nil, errors.New("invalid time format, expected YYYY-MM-DD HH:mm")
	}

	regTime := time.Now()
	if req.RegistrationTime != "" {
		if parsed, e := time.Parse("2006-01-02 15:04", req.RegistrationTime); e == nil {
			regTime = parsed
		}
	}

	record := models.OutcomeRecord{
		ID:               uuid.New().String(),
		PatientID:        patientID,
		Type:             req.Type,
		Reason:           req.Reason,
		Time:             t,
		Remarks:          req.Remarks,
		Registrar:        req.Registrar,
		RegistrationTime: regTime,
		IsDoorRule:       req.IsDoorRule,
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}

	resp := s.buildOutcomeResponse(record)
	return &resp, nil
}

// UpdateOutcomeRecord 更新转归记录
func (s *MedicalHistoryService) UpdateOutcomeRecord(patientID, recordID string, req *OutcomeRecordRequest) (*OutcomeRecordResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var record models.OutcomeRecord
	err := s.db.Where("id = ? AND patient_id = ?", recordID, patientID).First(&record).Error
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

	record.Type = req.Type
	record.Reason = req.Reason
	record.Time = t
	record.Remarks = req.Remarks
	record.Registrar = req.Registrar
	record.IsDoorRule = req.IsDoorRule
	if req.RegistrationTime != "" {
		if parsed, e := time.Parse("2006-01-02 15:04", req.RegistrationTime); e == nil {
			record.RegistrationTime = parsed
		}
	}

	if err := s.db.Save(&record).Error; err != nil {
		return nil, err
	}

	resp := s.buildOutcomeResponse(record)
	return &resp, nil
}

// DeleteOutcomeRecord 删除转归记录
func (s *MedicalHistoryService) DeleteOutcomeRecord(patientID, recordID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Where("id = ? AND patient_id = ?", recordID, patientID).Delete(&models.OutcomeRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("outcome record not found")
	}
	return nil
}

func (s *MedicalHistoryService) buildOutcomeResponse(r models.OutcomeRecord) OutcomeRecordResponse {
	return OutcomeRecordResponse{
		ID:               r.ID,
		Type:             r.Type,
		Reason:           r.Reason,
		Time:             r.Time.Format("2006-01-02 15:04"),
		Remarks:          r.Remarks,
		Registrar:        r.Registrar,
		RegistrationTime: r.RegistrationTime.Format("2006-01-02 15:04"),
		IsDoorRule:       r.IsDoorRule,
	}
}
