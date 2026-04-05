package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PatientBasicService 患者基本信息服务
type PatientBasicService struct {
	db *gorm.DB
}

// NewPatientBasicService 创建患者基本信息服务
func NewPatientBasicService() *PatientBasicService {
	return &PatientBasicService{
		db: database.GetDB(),
	}
}

// GetBasicInfo 获取患者基本信息档案
func (s *PatientBasicService) GetBasicInfo(patientID string) (*PatientBasicInfoResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patient models.Patient
	err := s.db.Where("id = ?", patientID).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	// 查询扩展信息（可选记录）
	var basicInfo models.PatientBasicInfo
	s.db.Where("patient_id = ?", patientID).First(&basicInfo)

	// 查询关联信息（可选记录 - 使用 Find 避免 ErrRecordNotFound）
	var vascularAccess models.VascularAccess
	s.db.Where("patient_id = ?", patientID).First(&vascularAccess)

	var medicalHistory models.MedicalHistory
	s.db.Where("patient_id = ?", patientID).First(&medicalHistory)

	// 构建响应
	response := &PatientBasicInfoResponse{
		PersonalInfo:    s.buildPersonalInfo(patient, basicInfo),
		MedicalInfo:     s.buildMedicalInfo(patient, basicInfo),
		VitalSocialInfo: s.buildVitalSocialInfo(patient, basicInfo),
		ContactInfo:     s.buildContactInfo(basicInfo),
	}

	return response, nil
}

// UpdateBasicInfo 更新患者基本信息档案
func (s *PatientBasicService) UpdateBasicInfo(patientID string, req *PatientBasicInfoRequest) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	var patient models.Patient
	err := s.db.Where("id = ?", patientID).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("patient not found")
		}
		return err
	}

	// 更新患者基本信息（主表字段）
	patientUpdates := make(map[string]interface{})

	if req.PersonalInfo.Name != nil {
		patientUpdates["name"] = *req.PersonalInfo.Name
	}
	if req.PersonalInfo.Gender != nil {
		patientUpdates["gender"] = *req.PersonalInfo.Gender
	}
	if req.PersonalInfo.PatientType != nil {
		patientUpdates["patient_type"] = *req.PersonalInfo.PatientType
	}
	if req.MedicalInfo.InsuranceType != nil {
		patientUpdates["insurance_type"] = *req.MedicalInfo.InsuranceType
	}
	if req.MedicalInfo.DoctorName != nil {
		patientUpdates["doctor_name"] = *req.MedicalInfo.DoctorName
	}
	if req.VitalSocialInfo.DryWeight != nil {
		patientUpdates["dry_weight"] = *req.VitalSocialInfo.DryWeight
	}

	if len(patientUpdates) > 0 {
		err = s.db.Model(&patient).Updates(patientUpdates).Error
		if err != nil {
			return fmt.Errorf("failed to update patient: %w", err)
		}
	}

	// 更新或创建扩展信息
	var basicInfo models.PatientBasicInfo
	err = s.db.Where("patient_id = ?", patientID).First(&basicInfo).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新记录
		basicInfo = models.PatientBasicInfo{
			ID:        uuid.New().String(),
			PatientID: patientID,
		}
	}

	// 更新扩展字段
	basicInfo.Pinyin = req.PersonalInfo.Pinyin
	basicInfo.Ethnicity = req.PersonalInfo.Ethnicity
	basicInfo.IDType = s.stringValue(req.PersonalInfo.IDType, models.IDTypeIDCard)
	basicInfo.IDNumber = req.PersonalInfo.IDNumber

	basicInfo.VisitCategory = req.MedicalInfo.VisitCategory
	basicInfo.AdmissionNo = req.MedicalInfo.AdmissionNo
	basicInfo.VisitNo = req.MedicalInfo.VisitNo
	basicInfo.MedicalRecordNo = req.MedicalInfo.MedicalRecordNo
	basicInfo.InsuranceNo = req.MedicalInfo.InsuranceNo
	basicInfo.HdisPatientID = req.MedicalInfo.HdisPatientId
	basicInfo.DialysisNo = req.MedicalInfo.DialysisNo
	basicInfo.NurseName = req.MedicalInfo.NurseName
	basicInfo.FirstDialysisDate = parseDatePtr(req.MedicalInfo.FirstDialysisDate)
	basicInfo.FirstHospitalDate = parseDatePtr(req.MedicalInfo.FirstHospitalDate)
	basicInfo.FirstDialysisHospital = req.MedicalInfo.FirstDialysisHospital

	basicInfo.Height = req.VitalSocialInfo.Height
	basicInfo.ABOBloodType = req.VitalSocialInfo.ABOBloodType
	basicInfo.RhBloodType = req.VitalSocialInfo.RhBloodType
	basicInfo.EducationLevel = req.VitalSocialInfo.EducationLevel
	basicInfo.Occupation = req.VitalSocialInfo.Occupation
	basicInfo.MaritalStatus = req.VitalSocialInfo.MaritalStatus
	basicInfo.Workplace = req.VitalSocialInfo.Workplace

	basicInfo.Phone = req.ContactInfo.Phone
	basicInfo.Wechat = req.ContactInfo.Wechat
	basicInfo.Landline = req.ContactInfo.Landline
	basicInfo.Address = req.ContactInfo.Address
	basicInfo.District = req.ContactInfo.District
	basicInfo.ContactName = req.ContactInfo.ContactName
	basicInfo.ContactPhone = req.ContactInfo.ContactPhone

	// 保存扩展信息
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 新建：使用显式列插入，避免 GORM 在 *time.Time 字段上的反射写入 panic
		err = insertPatientBasicInfo(s.db, basicInfo)
	} else {
		// 更新
		err = s.db.Save(&basicInfo).Error
	}

	if err != nil {
		return fmt.Errorf("failed to save basic info: %w", err)
	}

	return nil
}

// buildPersonalInfo 构建身份核心信息
func (s *PatientBasicService) buildPersonalInfo(patient models.Patient, basicInfo models.PatientBasicInfo) PatientBasicPersonal {
	return PatientBasicPersonal{
		Name:        patient.Name,
		Pinyin:      basicInfo.Pinyin,
		Birthday:    formatDatePtr(basicInfo.Birthday),
		Age:         patient.Age,
		Gender:      patient.Gender,
		Ethnicity:   basicInfo.Ethnicity,
		IDType:      basicInfo.IDType,
		IDNumber:    s.stringPtrOrEmpty(basicInfo.IDNumber),
		PatientType: s.stringOrDefault(patient.PatientType, "门诊"),
	}
}

// buildMedicalInfo 构建医疗登记信息
func (s *PatientBasicService) buildMedicalInfo(patient models.Patient, basicInfo models.PatientBasicInfo) PatientBasicMedical {
	// 计算透析龄
	var dialysisAge *string
	if basicInfo.FirstDialysisDate != nil {
		age := CalculateDialysisAge(*basicInfo.FirstDialysisDate)
		dialysisAge = &age
	}

	return PatientBasicMedical{
		VisitCategory:         basicInfo.VisitCategory,
		AdmissionNo:           basicInfo.AdmissionNo,
		VisitNo:               basicInfo.VisitNo,
		MedicalRecordNo:       basicInfo.MedicalRecordNo,
		InsuranceNo:           basicInfo.InsuranceNo,
		HdisPatientId:         basicInfo.HdisPatientID,
		InsuranceType:         s.stringOrDefault(patient.InsuranceType, "自费"),
		DialysisNo:            basicInfo.DialysisNo,
		DoctorName:            patient.DoctorName,
		NurseName:             basicInfo.NurseName,
		FirstDialysisDate:     formatDatePtr(basicInfo.FirstDialysisDate),
		FirstHospitalDate:     formatDatePtr(basicInfo.FirstHospitalDate),
		FirstDialysisHospital: basicInfo.FirstDialysisHospital,
		CurrentDialysisAge:    dialysisAge,
	}
}

// buildVitalSocialInfo 构建生命体征与社会信息
func (s *PatientBasicService) buildVitalSocialInfo(patient models.Patient, basicInfo models.PatientBasicInfo) PatientBasicVitalSocial {
	return PatientBasicVitalSocial{
		Height:         basicInfo.Height,
		DryWeight:      patient.DryWeight,
		ABOBloodType:   basicInfo.ABOBloodType,
		RhBloodType:    basicInfo.RhBloodType,
		EducationLevel: basicInfo.EducationLevel,
		Occupation:     basicInfo.Occupation,
		MaritalStatus:  basicInfo.MaritalStatus,
		Workplace:      basicInfo.Workplace,
	}
}

// buildContactInfo 构建联系信息
func (s *PatientBasicService) buildContactInfo(basicInfo models.PatientBasicInfo) PatientBasicContact {
	return PatientBasicContact{
		Phone:        basicInfo.Phone,
		Wechat:       basicInfo.Wechat,
		Landline:     basicInfo.Landline,
		Address:      basicInfo.Address,
		District:     basicInfo.District,
		ContactName:  basicInfo.ContactName,
		ContactPhone: basicInfo.ContactPhone,
	}
}

// stringOrDefault 安全获取字符串值
func (s *PatientBasicService) stringOrDefault(str string, defaultVal string) string {
	if str == "" {
		return defaultVal
	}
	return str
}

// stringValue 安全获取字符串指针值
func (s *PatientBasicService) stringValue(str *string, defaultVal string) string {
	if str == nil || *str == "" {
		return defaultVal
	}
	return *str
}

// stringPtrOrEmpty 安全获取字符串指针的值，空返回空字符串
func (s *PatientBasicService) stringPtrOrEmpty(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

// formatDatePtr 格式化日期指针为字符串
func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format("2006-01-02")
	return &formatted
}

// parseDatePtr 解析字符串日期为时间指针
func parseDatePtr(str *string) *time.Time {
	if str == nil || *str == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *str)
	if err != nil {
		return nil
	}
	return &t
}
