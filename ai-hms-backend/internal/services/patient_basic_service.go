package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	legacymodels "github.com/elliotxin/ai-hms-backend/internal/models/legacy"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"gorm.io/gorm"
)

type PatientBasicService struct {
	db *gorm.DB
}

type legacyPatientBasicRow struct {
	Spell                    *string    `gorm:"column:Spell"`
	BirthDate                *time.Time `gorm:"column:BirthDate"`
	Nation                   *string    `gorm:"column:Nation"`
	ABOType                  *string    `gorm:"column:ABOType"`
	RHType                   *string    `gorm:"column:RHType"`
	Height                   *string    `gorm:"column:Height"`
	Occupation               *string    `gorm:"column:Occupation"`
	MaritalStatus            *string    `gorm:"column:MaritalStatus"`
	EducationLevel           *string    `gorm:"column:EducationLevel"`
	Province                 *string    `gorm:"column:Province"`
	City                     *string    `gorm:"column:City"`
	County                   *string    `gorm:"column:County"`
	Address                  *string    `gorm:"column:Address"`
	PhoneNo                  *string    `gorm:"column:PhoneNo"`
	HomePhoneNo              *string    `gorm:"column:HomePhoneNo"`
	ExpenseType              *string    `gorm:"column:ExpenseType"`
	SSN                      *string    `gorm:"column:SSN"`
	DialysisNo               *string    `gorm:"column:DialysisNo"`
	ResponsibilityDrID       *int64     `gorm:"column:ResponsibilityDrId"`
	ResponsibilityNurseID    *int64     `gorm:"column:ResponsibilityNurseId"`
	FirstDialysisDate        *time.Time `gorm:"column:FirstDialysisDate"`
	FirstDialysisHospital    *string    `gorm:"column:FirstDialysisHospital"`
	OurHospitalFirstDialysis *time.Time `gorm:"column:OurHospitalFirstDialysisDate"`
	WeChatNo                 *string    `gorm:"column:WeChatNo"`
	IDName                   *string    `gorm:"column:IDName"`
	PatientType              *string    `gorm:"column:PatientType"`
	Workunit                 *string    `gorm:"column:Workunit"`
}

type legacyHospitalizationRow struct {
	ID              int64   `gorm:"column:Id"`
	CaseNo          *string `gorm:"column:CaseNo"`
	HospNo          *string `gorm:"column:HospNo"`
	BarCode         *string `gorm:"column:BarCode"`
	HospPatientType *string `gorm:"column:HospPatientType"`
	AttendDr        *string `gorm:"column:AttendDr"`
	ReceptionDr     *string `gorm:"column:ReceptionDr"`
	MedicalRecordNo *string `gorm:"column:MedicalRecordNo"`
}

type legacyIDInfoRow struct {
	ID     int64   `gorm:"column:Id"`
	IDType *string `gorm:"column:IDType"`
	IDNo   *string `gorm:"column:IDNo"`
}

type legacyFamilyMemberRow struct {
	ID      int64   `gorm:"column:Id"`
	Name    *string `gorm:"column:Name"`
	PhoneNo *string `gorm:"column:PhoneNo"`
}

func NewPatientBasicService() *PatientBasicService {
	return &PatientBasicService{db: database.GetDB()}
}

func (s *PatientBasicService) GetBasicInfo(patientID string) (*PatientBasicInfoResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}

	var patient models.Patient
	if err := s.db.Where(`"Id" = ?`, legacyPatientID).First(&patient).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	var basicInfo models.PatientBasicInfo

	legacyBasic, err := s.getLegacyPatientBasic(int64(legacyPatientID))
	if err != nil {
		return nil, err
	}
	hospitalization, err := s.getLegacyHospitalization(int64(legacyPatientID))
	if err != nil {
		return nil, err
	}
	idInfo, err := s.getLegacyIDInfo(int64(legacyPatientID))
	if err != nil {
		return nil, err
	}
	familyMember, err := s.getLegacyFamilyMember(int64(legacyPatientID))
	if err != nil {
		return nil, err
	}

	if err := s.applyLegacyFallbacks(&patient, &basicInfo, legacyBasic, hospitalization, idInfo, familyMember); err != nil {
		return nil, err
	}

	return &PatientBasicInfoResponse{
		PersonalInfo:    s.buildPersonalInfo(patient, basicInfo),
		MedicalInfo:     s.buildMedicalInfo(patient, basicInfo),
		VitalSocialInfo: s.buildVitalSocialInfo(patient, basicInfo),
		ContactInfo:     s.buildContactInfo(basicInfo),
	}, nil
}

func (s *PatientBasicService) UpdateBasicInfo(patientID string, req *PatientBasicInfoRequest) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return errors.New("invalid patient id")
	}

	var patient models.Patient
	if err := s.db.Where(`"Id" = ?`, legacyPatientID).First(&patient).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("patient not found")
		}
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		patientUpdates := make(map[string]interface{})
		if req.PersonalInfo.Name != nil {
			patientUpdates["Name"] = strings.TrimSpace(*req.PersonalInfo.Name)
		}
		if req.PersonalInfo.Gender != nil {
			patientUpdates["Gender"] = strings.TrimSpace(*req.PersonalInfo.Gender)
		}
		if req.PersonalInfo.PatientType != nil {
			patientUpdates["PatientType"] = strings.TrimSpace(*req.PersonalInfo.PatientType)
		}
		if req.PersonalInfo.Pinyin != nil {
			patientUpdates["Spell"] = strings.TrimSpace(*req.PersonalInfo.Pinyin)
		}
		if req.PersonalInfo.Birthday != nil {
			patientUpdates["BirthDate"] = parseDatePtr(req.PersonalInfo.Birthday)
		}
		if req.PersonalInfo.Ethnicity != nil {
			patientUpdates["Nation"] = strings.TrimSpace(*req.PersonalInfo.Ethnicity)
		}
		if req.MedicalInfo.InsuranceType != nil {
			patientUpdates["ExpenseType"] = strings.TrimSpace(*req.MedicalInfo.InsuranceType)
		}
		if req.MedicalInfo.InsuranceNo != nil {
			patientUpdates["SSN"] = strings.TrimSpace(*req.MedicalInfo.InsuranceNo)
		}
		if req.MedicalInfo.DialysisNo != nil {
			patientUpdates["DialysisNo"] = strings.TrimSpace(*req.MedicalInfo.DialysisNo)
		}
		if req.MedicalInfo.FirstDialysisDate != nil {
			patientUpdates["FirstDialysisDate"] = parseDatePtr(req.MedicalInfo.FirstDialysisDate)
		}
		if req.MedicalInfo.FirstHospitalDate != nil {
			patientUpdates["OurHospitalFirstDialysisDate"] = parseDatePtr(req.MedicalInfo.FirstHospitalDate)
		}
		if req.MedicalInfo.FirstDialysisHospital != nil {
			patientUpdates["FirstDialysisHospital"] = strings.TrimSpace(*req.MedicalInfo.FirstDialysisHospital)
		}
		if req.VitalSocialInfo.Height != nil {
			patientUpdates["Height"] = strings.TrimSpace(*req.VitalSocialInfo.Height)
		}
		if req.VitalSocialInfo.DryWeight != nil {
			patientUpdates["Weight"] = *req.VitalSocialInfo.DryWeight
		}
		if req.VitalSocialInfo.ABOBloodType != nil {
			patientUpdates["ABOType"] = strings.TrimSpace(*req.VitalSocialInfo.ABOBloodType)
		}
		if req.VitalSocialInfo.RhBloodType != nil {
			patientUpdates["RHType"] = strings.TrimSpace(*req.VitalSocialInfo.RhBloodType)
		}
		if req.VitalSocialInfo.EducationLevel != nil {
			patientUpdates["EducationLevel"] = strings.TrimSpace(*req.VitalSocialInfo.EducationLevel)
		}
		if req.VitalSocialInfo.Occupation != nil {
			patientUpdates["Occupation"] = strings.TrimSpace(*req.VitalSocialInfo.Occupation)
		}
		if req.VitalSocialInfo.MaritalStatus != nil {
			patientUpdates["MaritalStatus"] = strings.TrimSpace(*req.VitalSocialInfo.MaritalStatus)
		}
		if req.VitalSocialInfo.Workplace != nil {
			patientUpdates["Workunit"] = strings.TrimSpace(*req.VitalSocialInfo.Workplace)
		}
		if req.ContactInfo.Phone != nil {
			patientUpdates["PhoneNo"] = strings.TrimSpace(*req.ContactInfo.Phone)
		}
		if req.ContactInfo.Wechat != nil {
			patientUpdates["WeChatNo"] = strings.TrimSpace(*req.ContactInfo.Wechat)
		}
		if req.ContactInfo.Landline != nil {
			patientUpdates["HomePhoneNo"] = strings.TrimSpace(*req.ContactInfo.Landline)
		}
		if req.ContactInfo.Address != nil {
			patientUpdates["Address"] = strings.TrimSpace(*req.ContactInfo.Address)
		}
		if req.ContactInfo.District != nil {
			province, city, county := splitLegacyDistrict(*req.ContactInfo.District)
			patientUpdates["Province"] = province
			patientUpdates["City"] = city
			patientUpdates["County"] = county
		}

		if req.MedicalInfo.DoctorName != nil {
			doctorName := strings.TrimSpace(*req.MedicalInfo.DoctorName)
			if doctorID, err := s.resolveEmployeeIDByName(tx, doctorName); err != nil {
				return err
			} else if doctorID != nil {
				patientUpdates["ResponsibilityDrId"] = *doctorID
			}
		}
		if req.MedicalInfo.NurseName != nil {
			nurseName := strings.TrimSpace(*req.MedicalInfo.NurseName)
			if nurseID, err := s.resolveEmployeeIDByName(tx, nurseName); err != nil {
				return err
			} else if nurseID != nil {
				patientUpdates["ResponsibilityNurseId"] = *nurseID
			}
		}

		if len(patientUpdates) > 0 {
			patientUpdates["LastModifyTime"] = now
			if err := tx.Table("Register_PatientInfomation").Where(`"Id" = ?`, legacyPatientID).Updates(patientUpdates).Error; err != nil {
				return fmt.Errorf("failed to update legacy patient: %w", err)
			}
		}

		hospUpdates := make(map[string]interface{})
		if req.MedicalInfo.VisitCategory != nil {
			hospUpdates["HospPatientType"] = strings.TrimSpace(*req.MedicalInfo.VisitCategory)
		}
		if req.MedicalInfo.AdmissionNo != nil {
			hospUpdates["HospNo"] = strings.TrimSpace(*req.MedicalInfo.AdmissionNo)
		}
		if req.MedicalInfo.VisitNo != nil {
			hospUpdates["CaseNo"] = strings.TrimSpace(*req.MedicalInfo.VisitNo)
		}
		if req.MedicalInfo.MedicalRecordNo != nil {
			hospUpdates["MedicalRecordNo"] = strings.TrimSpace(*req.MedicalInfo.MedicalRecordNo)
		}
		if req.MedicalInfo.DoctorName != nil {
			hospUpdates["AttendDr"] = strings.TrimSpace(*req.MedicalInfo.DoctorName)
		}
		if len(hospUpdates) > 0 {
			if err := s.upsertLegacyHospitalization(tx, patient, int64(legacyPatientID), hospUpdates, now); err != nil {
				return err
			}
		}

		idInfoUpdates := make(map[string]interface{})
		if req.PersonalInfo.IDType != nil {
			idInfoUpdates["IDType"] = strings.TrimSpace(*req.PersonalInfo.IDType)
		}
		if req.PersonalInfo.IDNumber != nil {
			idInfoUpdates["IDNo"] = strings.TrimSpace(*req.PersonalInfo.IDNumber)
		}
		if len(idInfoUpdates) > 0 {
			if err := s.upsertLegacyIDInfo(tx, patient, int64(legacyPatientID), idInfoUpdates, now); err != nil {
				return err
			}
		}

		familyUpdates := make(map[string]interface{})
		if req.ContactInfo.ContactName != nil {
			familyUpdates["Name"] = strings.TrimSpace(*req.ContactInfo.ContactName)
		}
		if req.ContactInfo.ContactPhone != nil {
			familyUpdates["PhoneNo"] = strings.TrimSpace(*req.ContactInfo.ContactPhone)
		}
		if len(familyUpdates) > 0 {
			if err := s.upsertLegacyFamilyMember(tx, patient, int64(legacyPatientID), familyUpdates, now); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *PatientBasicService) buildPersonalInfo(patient models.Patient, basicInfo models.PatientBasicInfo) PatientBasicPersonal {
	age := patient.Age
	if age == 0 {
		if basicInfo.Birthday != nil {
			age = calculateAge(*basicInfo.Birthday)
		} else if patient.BirthDate != nil {
			age = calculateAge(*patient.BirthDate)
		}
	}

	return PatientBasicPersonal{
		Name:        patient.Name,
		Pinyin:      basicInfo.Pinyin,
		Birthday:    formatDatePtr(basicInfo.Birthday),
		Age:         age,
		Gender:      patient.Gender,
		Ethnicity:   basicInfo.Ethnicity,
		IDType:      basicInfo.IDType,
		IDNumber:    stringOrEmpty(basicInfo.IDNumber),
		PatientType: strings.TrimSpace(patient.PatientType),
	}
}

func (s *PatientBasicService) buildMedicalInfo(patient models.Patient, basicInfo models.PatientBasicInfo) PatientBasicMedical {
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
		HdisPatientId:         nil,
		InsuranceType:         strings.TrimSpace(patient.InsuranceType),
		DialysisNo:            basicInfo.DialysisNo,
		DoctorName:            strings.TrimSpace(patient.DoctorName),
		NurseName:             basicInfo.NurseName,
		FirstDialysisDate:     formatDatePtr(basicInfo.FirstDialysisDate),
		FirstHospitalDate:     formatDatePtr(basicInfo.FirstHospitalDate),
		FirstDialysisHospital: basicInfo.FirstDialysisHospital,
		CurrentDialysisAge:    dialysisAge,
	}
}

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

func (s *PatientBasicService) getLegacyPatientBasic(patientID int64) (*legacyPatientBasicRow, error) {
	var row legacyPatientBasicRow
	err := s.db.Table("Register_PatientInfomation").
		Select(`
"Spell",
"BirthDate",
"Nation",
"ABOType",
"RHType",
CAST("Height" AS text) AS "Height",
"Occupation",
"MaritalStatus",
"EducationLevel",
"Province",
"City",
"County",
"Address",
"PhoneNo",
"HomePhoneNo",
"ExpenseType",
"SSN",
"DialysisNo",
"ResponsibilityDrId",
"ResponsibilityNurseId",
"FirstDialysisDate",
"FirstDialysisHospital",
"OurHospitalFirstDialysisDate",
"WeChatNo",
"IDName",
"PatientType",
"Workunit"`).
		Where(`"Id" = ?`, patientID).
		Take(&row).Error
	if isIgnorableLegacyQueryError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *PatientBasicService) getLegacyHospitalization(patientID int64) (*legacyHospitalizationRow, error) {
	var row legacyHospitalizationRow
	err := s.db.Table("Register_Hospitalization").
		Select(`"Id", "CaseNo", "HospNo", "BarCode", "HospPatientType", "AttendDr", "ReceptionDr", "MedicalRecordNo"`).
		Where(`"PatientId" = ?`, patientID).
		Order(`"Id" DESC`).
		Take(&row).Error
	if isIgnorableLegacyQueryError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *PatientBasicService) getLegacyIDInfo(patientID int64) (*legacyIDInfoRow, error) {
	var row legacyIDInfoRow
	err := s.db.Table("Register_IDInfomation").
		Select(`"Id", "IDType", "IDNo"`).
		Where(`"PatientId" = ? AND COALESCE("IsDisabled", false) = false`, patientID).
		Order(`"Id" DESC`).
		Take(&row).Error
	if isIgnorableLegacyQueryError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *PatientBasicService) getLegacyFamilyMember(patientID int64) (*legacyFamilyMemberRow, error) {
	var row legacyFamilyMemberRow
	err := s.db.Table("Register_FamilyMember").
		Select(`"Id", "Name", "PhoneNo"`).
		Where(`"PatientId" = ? AND COALESCE("IsDisabled", false) = false`, patientID).
		Order(`"Id" DESC`).
		Take(&row).Error
	if isIgnorableLegacyQueryError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *PatientBasicService) applyLegacyFallbacks(
	patient *models.Patient,
	basicInfo *models.PatientBasicInfo,
	legacyBasic *legacyPatientBasicRow,
	hospitalization *legacyHospitalizationRow,
	idInfo *legacyIDInfoRow,
	familyMember *legacyFamilyMemberRow,
) error {
	if patient == nil || basicInfo == nil {
		return nil
	}

	if legacyBasic != nil {
		if patient.BirthDate == nil && legacyBasic.BirthDate != nil {
			patient.BirthDate = legacyBasic.BirthDate
		}
		if strings.TrimSpace(patient.PatientType) == "" {
			patient.PatientType = stringValue(legacyBasic.PatientType)
		}
		if strings.TrimSpace(patient.InsuranceType) == "" {
			patient.InsuranceType = stringValue(legacyBasic.ExpenseType)
		}
		if strings.TrimSpace(patient.DialysisNo) == "" {
			patient.DialysisNo = stringValue(legacyBasic.DialysisNo)
		}
		if strings.TrimSpace(patient.PhoneNo) == "" {
			patient.PhoneNo = stringValue(legacyBasic.PhoneNo)
		}
	}
	if strings.TrimSpace(patient.PatientType) == "" && hospitalization != nil {
		patient.PatientType = stringValue(hospitalization.HospPatientType)
	}
	if strings.TrimSpace(patient.DoctorName) == "" {
		doctorName, err := s.resolveEmployeeName(legacyBasicResponsibilityID(legacyBasic, true))
		if err != nil {
			return err
		}
		patient.DoctorName = doctorName
		if strings.TrimSpace(patient.DoctorName) == "" && hospitalization != nil {
			patient.DoctorName = firstNonEmptyStringValue(
				stringValue(hospitalization.AttendDr),
				stringValue(hospitalization.ReceptionDr),
			)
		}
	}

	basicInfo.Pinyin = firstNonEmptyPtr(basicInfo.Pinyin, normalizeStringPtr(fieldStringPtr(legacyBasic, "spell")))
	if basicInfo.Birthday == nil {
		if patient.BirthDate != nil {
			basicInfo.Birthday = patient.BirthDate
		} else if legacyBasic != nil {
			basicInfo.Birthday = legacyBasic.BirthDate
		}
	}
	basicInfo.Ethnicity = firstNonEmptyPtr(basicInfo.Ethnicity, normalizeStringPtr(fieldStringPtr(legacyBasic, "nation")))
	if strings.TrimSpace(basicInfo.IDType) == "" {
		basicInfo.IDType = firstNonEmptyStringValue(idInfoType(idInfo), stringValue(fieldStringPtr(legacyBasic, "idName")))
	}
	basicInfo.IDNumber = firstNonEmptyPtr(basicInfo.IDNumber, idInfoNumberPtr(idInfo), normalizeStringPtr(fieldStringPtr(legacyBasic, "ssn")))
	basicInfo.VisitCategory = firstNonEmptyPtr(basicInfo.VisitCategory, normalizeStringPtr(fieldStringPtr(hospitalization, "visitCategory")))
	basicInfo.AdmissionNo = firstNonEmptyPtr(basicInfo.AdmissionNo, normalizeStringPtr(fieldStringPtr(hospitalization, "admissionNo")))
	basicInfo.VisitNo = firstNonEmptyPtr(
		basicInfo.VisitNo,
		normalizeStringPtr(fieldStringPtr(hospitalization, "visitNo")),
		normalizeStringPtr(fieldStringPtr(hospitalization, "barCode")),
	)
	basicInfo.MedicalRecordNo = firstNonEmptyPtr(basicInfo.MedicalRecordNo, normalizeStringPtr(fieldStringPtr(hospitalization, "medicalRecordNo")))
	basicInfo.InsuranceNo = firstNonEmptyPtr(basicInfo.InsuranceNo, normalizeStringPtr(fieldStringPtr(legacyBasic, "ssn")))
	basicInfo.DialysisNo = firstNonEmptyPtr(basicInfo.DialysisNo, ptrFromString(patient.DialysisNo), normalizeStringPtr(fieldStringPtr(legacyBasic, "dialysisNo")))
	if basicInfo.NurseName == nil {
		nurseName, err := s.resolveEmployeeName(legacyBasicResponsibilityID(legacyBasic, false))
		if err != nil {
			return err
		}
		basicInfo.NurseName = ptrFromString(nurseName)
	}
	if basicInfo.FirstDialysisDate == nil && legacyBasic != nil {
		basicInfo.FirstDialysisDate = legacyBasic.FirstDialysisDate
	}
	if basicInfo.FirstHospitalDate == nil && legacyBasic != nil {
		basicInfo.FirstHospitalDate = legacyBasic.OurHospitalFirstDialysis
	}
	basicInfo.FirstDialysisHospital = firstNonEmptyPtr(basicInfo.FirstDialysisHospital, normalizeStringPtr(fieldStringPtr(legacyBasic, "firstDialysisHospital")))
	basicInfo.Height = firstNonEmptyPtr(basicInfo.Height, normalizeStringPtr(fieldStringPtr(legacyBasic, "height")))
	basicInfo.ABOBloodType = firstNonEmptyPtr(basicInfo.ABOBloodType, normalizeStringPtr(fieldStringPtr(legacyBasic, "aboType")))
	basicInfo.RhBloodType = firstNonEmptyPtr(basicInfo.RhBloodType, normalizeStringPtr(fieldStringPtr(legacyBasic, "rhType")))
	basicInfo.EducationLevel = firstNonEmptyPtr(basicInfo.EducationLevel, normalizeStringPtr(fieldStringPtr(legacyBasic, "educationLevel")))
	basicInfo.Occupation = firstNonEmptyPtr(basicInfo.Occupation, normalizeStringPtr(fieldStringPtr(legacyBasic, "occupation")))
	basicInfo.MaritalStatus = firstNonEmptyPtr(basicInfo.MaritalStatus, normalizeStringPtr(fieldStringPtr(legacyBasic, "maritalStatus")))
	basicInfo.Workplace = firstNonEmptyPtr(basicInfo.Workplace, normalizeStringPtr(fieldStringPtr(legacyBasic, "workunit")))
	basicInfo.Phone = firstNonEmptyPtr(basicInfo.Phone, ptrFromString(patient.PhoneNo), normalizeStringPtr(fieldStringPtr(legacyBasic, "phoneNo")))
	basicInfo.Wechat = firstNonEmptyPtr(basicInfo.Wechat, normalizeStringPtr(fieldStringPtr(legacyBasic, "weChatNo")))
	basicInfo.Landline = firstNonEmptyPtr(basicInfo.Landline, normalizeStringPtr(fieldStringPtr(legacyBasic, "homePhoneNo")))
	basicInfo.Address = firstNonEmptyPtr(basicInfo.Address, normalizeStringPtr(fieldStringPtr(legacyBasic, "address")))
	if basicInfo.District == nil && legacyBasic != nil {
		basicInfo.District = ptrFromString(strings.TrimSpace(
			stringValue(legacyBasic.Province) +
				stringValue(legacyBasic.City) +
				stringValue(legacyBasic.County),
		))
	}
	if familyMember != nil {
		basicInfo.ContactName = firstNonEmptyPtr(basicInfo.ContactName, normalizeStringPtr(familyMember.Name))
		basicInfo.ContactPhone = firstNonEmptyPtr(basicInfo.ContactPhone, normalizeStringPtr(familyMember.PhoneNo))
	}

	return nil
}

func (s *PatientBasicService) upsertLegacyHospitalization(
	tx *gorm.DB,
	patient models.Patient,
	legacyPatientID int64,
	updates map[string]interface{},
	now time.Time,
) error {
	row, err := s.getLegacyHospitalization(legacyPatientID)
	if err != nil {
		return fmt.Errorf("failed to query legacy hospitalization: %w", err)
	}

	updates["LastModifyTime"] = now
	if row != nil {
		if err := tx.Table("Register_Hospitalization").Where(`"Id" = ?`, row.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update legacy hospitalization: %w", err)
		}
		return nil
	}

	id, err := nextLegacyID()
	if err != nil {
		return fmt.Errorf("failed to generate hospitalization id: %w", err)
	}

	updates["Id"] = id
	updates["TenantId"] = patient.TenantID
	updates["PatientId"] = patient.ID
	updates["CreatorId"] = 0
	updates["CreateTime"] = now
	if err := tx.Table("Register_Hospitalization").Create(updates).Error; err != nil {
		return fmt.Errorf("failed to create legacy hospitalization: %w", err)
	}
	return nil
}

func (s *PatientBasicService) upsertLegacyIDInfo(
	tx *gorm.DB,
	patient models.Patient,
	legacyPatientID int64,
	updates map[string]interface{},
	now time.Time,
) error {
	row, err := s.getLegacyIDInfo(legacyPatientID)
	if err != nil {
		return fmt.Errorf("failed to query legacy id info: %w", err)
	}

	updates["LastModifyTime"] = now
	updates["IsDisabled"] = false
	if row != nil {
		if err := tx.Table("Register_IDInfomation").Where(`"Id" = ?`, row.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update legacy id info: %w", err)
		}
		return nil
	}

	id, err := nextLegacyID()
	if err != nil {
		return fmt.Errorf("failed to generate id info id: %w", err)
	}

	updates["Id"] = id
	updates["TenantId"] = patient.TenantID
	updates["PatientId"] = patient.ID
	updates["CreatorId"] = 0
	updates["CreateTime"] = now
	if err := tx.Table("Register_IDInfomation").Create(updates).Error; err != nil {
		return fmt.Errorf("failed to create legacy id info: %w", err)
	}
	return nil
}

func (s *PatientBasicService) upsertLegacyFamilyMember(
	tx *gorm.DB,
	patient models.Patient,
	legacyPatientID int64,
	updates map[string]interface{},
	now time.Time,
) error {
	row, err := s.getLegacyFamilyMember(legacyPatientID)
	if err != nil {
		return fmt.Errorf("failed to query legacy family member: %w", err)
	}

	updates["LastModifyTime"] = now
	updates["IsDisabled"] = false
	if row != nil {
		if err := tx.Table("Register_FamilyMember").Where(`"Id" = ?`, row.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update legacy family member: %w", err)
		}
		return nil
	}

	id, err := nextLegacyID()
	if err != nil {
		return fmt.Errorf("failed to generate family member id: %w", err)
	}

	updates["Id"] = id
	updates["TenantId"] = patient.TenantID
	updates["PatientId"] = patient.ID
	updates["CreatorId"] = 0
	updates["CreateTime"] = now
	updates["Type"] = ""
	if err := tx.Table("Register_FamilyMember").Create(updates).Error; err != nil {
		return fmt.Errorf("failed to create legacy family member: %w", err)
	}
	return nil
}

func (s *PatientBasicService) resolveEmployeeName(employeeID *int64) (string, error) {
	if employeeID == nil || *employeeID <= 0 {
		return "", nil
	}

	var employee legacymodels.OrganEmployee
	err := s.db.Where(`"Id" = ?`, modeltypes.LegacyID(*employeeID)).Take(&employee).Error
	if isIgnorableLegacyQueryError(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(employee.Name), nil
}

func (s *PatientBasicService) resolveEmployeeIDByName(tx *gorm.DB, name string) (*int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}

	var employee legacymodels.OrganEmployee
	err := tx.Where(`"Name" = ?`, name).Order(`"Id" DESC`).Take(&employee).Error
	if isIgnorableLegacyQueryError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to resolve employee by name: %w", err)
	}

	return &employee.ID, nil
}

func isIgnorableLegacyQueryError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "undefined_table") ||
		strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "undefined_column")
}

func normalizeStringPtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func ptrFromString(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func stringOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func firstNonEmptyPtr(values ...*string) *string {
	for _, value := range values {
		if normalized := normalizeStringPtr(value); normalized != nil {
			return normalized
		}
	}
	return nil
}

func firstNonEmptyStringValue(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func formatDatePtr(v *time.Time) *string {
	if v == nil {
		return nil
	}
	formatted := v.Format("2006-01-02")
	return &formatted
}

func parseDatePtr(v *string) *time.Time {
	if v == nil {
		return nil
	}
	value := strings.TrimSpace(*v)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &parsed
}

func calculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}

func splitLegacyDistrict(raw string) (string, string, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", "", ""
	}

	replacer := strings.NewReplacer("，", ",", "、", ",", "/", ",", "\\", ",", "|", ",", ">", ",", "-", ",")
	normalized := replacer.Replace(trimmed)
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == ',' || r == ' '
	})
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(strings.Join(parts[2:], ""))
	}
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), ""
	}

	province, rest := cutBySuffix(trimmed, []string{"特别行政区", "自治区", "省", "市"})
	city, county := cutBySuffix(rest, []string{"自治州", "地区", "盟", "州", "市", "区", "县"})
	if province == "" && city == "" {
		return "", "", trimmed
	}
	return province, city, county
}

func cutBySuffix(value string, suffixes []string) (string, string) {
	for _, suffix := range suffixes {
		if idx := strings.Index(value, suffix); idx >= 0 {
			end := idx + len(suffix)
			return strings.TrimSpace(value[:end]), strings.TrimSpace(value[end:])
		}
	}
	return "", strings.TrimSpace(value)
}

func legacyBasicResponsibilityID(row *legacyPatientBasicRow, doctor bool) *int64 {
	if row == nil {
		return nil
	}
	if doctor {
		return row.ResponsibilityDrID
	}
	return row.ResponsibilityNurseID
}

func idInfoType(row *legacyIDInfoRow) string {
	if row == nil || row.IDType == nil {
		return ""
	}
	return strings.TrimSpace(*row.IDType)
}

func idInfoNumberPtr(row *legacyIDInfoRow) *string {
	if row == nil {
		return nil
	}
	return normalizeStringPtr(row.IDNo)
}

func fieldStringPtr(row interface{}, field string) *string {
	switch r := row.(type) {
	case *legacyPatientBasicRow:
		if r == nil {
			return nil
		}
		switch field {
		case "spell":
			return r.Spell
		case "nation":
			return r.Nation
		case "aboType":
			return r.ABOType
		case "rhType":
			return r.RHType
		case "height":
			return r.Height
		case "occupation":
			return r.Occupation
		case "maritalStatus":
			return r.MaritalStatus
		case "educationLevel":
			return r.EducationLevel
		case "address":
			return r.Address
		case "phoneNo":
			return r.PhoneNo
		case "homePhoneNo":
			return r.HomePhoneNo
		case "ssn":
			return r.SSN
		case "dialysisNo":
			return r.DialysisNo
		case "firstDialysisHospital":
			return r.FirstDialysisHospital
		case "weChatNo":
			return r.WeChatNo
		case "idName":
			return r.IDName
		case "workunit":
			return r.Workunit
		}
	case *legacyHospitalizationRow:
		if r == nil {
			return nil
		}
		switch field {
		case "visitCategory":
			return r.HospPatientType
		case "admissionNo":
			return r.HospNo
		case "visitNo":
			return r.CaseNo
		case "barCode":
			return r.BarCode
		case "medicalRecordNo":
			return r.MedicalRecordNo
		}
	}
	return nil
}
