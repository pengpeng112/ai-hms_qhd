package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

// PatientService 患者服务
type PatientService struct {
	db *gorm.DB
}

// NewPatientService 创建患者服务
func NewPatientService() *PatientService {
	return &PatientService{
		db: database.GetDB(),
	}
}

// ListRequest 获取患者列表请求
type ListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Status    string `form:"status"`
	BedNumber string `form:"bedNumber"`
	Name      string `form:"name"`
	RiskLevel string `form:"riskLevel"`
}

// ListResponse 获取患者列表响应
type ListResponse struct {
	Items     []models.Patient `json:"items"`
	Total     int64            `json:"total"`
	Page      int              `json:"page"`
	PageSize  int              `json:"pageSize"`
	TotalPage int              `json:"totalPage"`
}

// List 获取患者列表
func (s *PatientService) List(req ListRequest) (*ListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.Patient{})

	// 筛选条件
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.BedNumber != "" {
		query = query.Where("bed_number LIKE ?", "%"+req.BedNumber+"%")
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.RiskLevel != "" {
		query = query.Where("risk_level = ?", req.RiskLevel)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询 - 列表页面只返回必要字段
	var items []models.Patient
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Select("id", "name", "age", "gender", "bed_number", "status",
			"patient_type", "insurance_type", "dry_weight", "default_mode", "doctor_name").
		Offset(offset).
		Limit(req.PageSize).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &ListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取患者详情
func (s *PatientService) Get(id string) (*models.Patient, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patient models.Patient
	err := s.db.
		Preload("VascularAccesses").
		Preload("MedicalHistory").
		Preload("TreatmentPlan").
		First(&patient, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	return &patient, nil
}

// CreateRequest 创建患者请求
type CreateRequest struct {
	Name          string  `json:"name" binding:"required"`
	Age           int     `json:"age" binding:"required,min=0,max=150"`
	Gender        string  `json:"gender" binding:"required,oneof=M F"`
	BedNumber     string  `json:"bedNumber"`
	Diagnosis     string  `json:"diagnosis"`
	RiskLevel     string  `json:"riskLevel"`
	Status        string  `json:"status"`
	PatientType   string  `json:"patientType"`
	InsuranceType string  `json:"insuranceType"`
	DryWeight     float64 `json:"dryWeight"`
	DefaultMode   string  `json:"defaultMode"`
	DoctorID      *string `json:"doctorId"`
	DoctorName    string  `json:"doctorName"`
	// 基本信息档案（可选）
	Pinyin                string  `json:"pinyin"`
	Birthday              *string `json:"birthday"`
	Ethnicity             string  `json:"ethnicity"`
	IdType                string  `json:"idType"`
	IdNumber              string  `json:"idNumber"`
	VisitCategory         string  `json:"visitCategory"`
	AdmissionNo           string  `json:"admissionNo"`
	VisitNo               string  `json:"visitNo"`
	MedicalRecordNo       string  `json:"medicalRecordNo"`
	InsuranceNo           string  `json:"insuranceNo"`
	DialysisNo            string  `json:"dialysisNo"`
	NurseName             string  `json:"nurseName"`
	FirstDialysisDate     *string `json:"firstDialysisDate"`
	FirstHospitalDate     *string `json:"firstHospitalDate"`
	FirstDialysisHospital string  `json:"firstDialysisHospital"`
	Height                string  `json:"height"`
	AboBloodType          string  `json:"aboBloodType"`
	RhBloodType           string  `json:"rhBloodType"`
	EducationLevel        string  `json:"educationLevel"`
	Occupation            string  `json:"occupation"`
	MaritalStatus         string  `json:"maritalStatus"`
	Workplace             string  `json:"workplace"`
	Phone                 string  `json:"phone"`
	Wechat                string  `json:"wechat"`
	Landline              string  `json:"landline"`
	Address               string  `json:"address"`
	District              string  `json:"district"`
	ContactName           string  `json:"contactName"`
	ContactPhone          string  `json:"contactPhone"`
}

// Create 创建患者
func (s *PatientService) Create(req CreateRequest, tenantID string, creatorID string) (*models.Patient, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(creatorID) == "" {
		return nil, errors.New("creator id is required")
	}
	if strings.TrimSpace(tenantID) == "" {
		tenantID = creatorID
	}

	// 生成业务ID，不作为主键，允许失败后重试
	today := time.Now().Format("20060102")

	// 重试机制：处理并发时的 ID 冲突
	maxRetries := 10
	var lastErr error

	var createdPatientID string
	for retry := 0; retry < maxRetries; retry++ {
		err := s.db.Transaction(func(tx *gorm.DB) error {
			tx = tx.Set("tenant_id", tenantID).Set("creator_id", creatorID)

			// 查询今天已有的患者数量
			var todayCount int64
			if err := tx.Model(&models.Patient{}).
				Where("id LIKE ?", "P"+today+"%").
				Count(&todayCount).Error; err != nil {
				return err
			}

			// 生成患者编号: P + 年月日 + 4位序号
			// 序号从 0001 开始，使用较大的偏移量避免并发冲突
			seq := int(todayCount) + retry + 1
			patientID := fmt.Sprintf("P%s%04d", today, seq)

			patient := models.Patient{
				ID:            patientID,
				Name:          req.Name,
				Age:           req.Age,
				Gender:        req.Gender,
				BedNumber:     req.BedNumber,
				Diagnosis:     req.Diagnosis,
				RiskLevel:     req.RiskLevel,
				Status:        req.Status,
				PatientType:   req.PatientType,
				InsuranceType: req.InsuranceType,
				DryWeight:     req.DryWeight,
				DefaultMode:   req.DefaultMode,
				DoctorID:      req.DoctorID,
				DoctorName:    req.DoctorName,
			}

			// 设置默认值
			if patient.RiskLevel == "" {
				patient.RiskLevel = models.RiskLevelLow
			}
			if patient.Status == "" {
				patient.Status = models.PatientStatusActive
			}

			if err := tx.Create(&patient).Error; err != nil {
				return err
			}

			if err := s.createBasicInfo(tx, patient.ID, req); err != nil {
				return err
			}

			createdPatientID = patient.ID
			return nil
		})
		if err == nil {
			var patient models.Patient
			if err := s.db.First(&patient, "id = ?", createdPatientID).Error; err != nil {
				return nil, err
			}
			return &patient, nil
		}

		// 检查是否是主键冲突错误
		if isDuplicateKeyError(err) {
			lastErr = err
			// 短暂延迟后重试，避免紧邻重试
			time.Sleep(time.Millisecond * time.Duration(10*(retry+1)))
			continue
		}

		// 其他错误直接返回
		return nil, err
	}

	return nil, fmt.Errorf("failed to create patient after %d retries: %w", maxRetries, lastErr)
}

// createBasicInfo 创建患者基本信息档案
func (s *PatientService) createBasicInfo(tx *gorm.DB, patientID string, req CreateRequest) error {
	// 创建基本信息档案（可选字段）
	basicInfo := models.PatientBasicInfo{
		ID:                    utils.GenerateID(),
		PatientID:             patientID,
		Pinyin:                stringPtr(req.Pinyin),
		Birthday:              parseTimePointer(req.Birthday),
		Ethnicity:             stringPtr(req.Ethnicity),
		IDType:                req.IdType,
		IDNumber:              stringPtr(req.IdNumber),
		VisitCategory:         stringPtr(req.VisitCategory),
		AdmissionNo:           stringPtr(req.AdmissionNo),
		VisitNo:               stringPtr(req.VisitNo),
		MedicalRecordNo:       stringPtr(req.MedicalRecordNo),
		InsuranceNo:           stringPtr(req.InsuranceNo),
		DialysisNo:            stringPtr(req.DialysisNo),
		NurseName:             stringPtr(req.NurseName),
		FirstDialysisDate:     parseTimePointer(req.FirstDialysisDate),
		FirstHospitalDate:     parseTimePointer(req.FirstHospitalDate),
		FirstDialysisHospital: stringPtr(req.FirstDialysisHospital),
		Height:                stringPtr(req.Height),
		ABOBloodType:          stringPtr(req.AboBloodType),
		RhBloodType:           stringPtr(req.RhBloodType),
		EducationLevel:        stringPtr(req.EducationLevel),
		Occupation:            stringPtr(req.Occupation),
		MaritalStatus:         stringPtr(req.MaritalStatus),
		Workplace:             stringPtr(req.Workplace),
		Phone:                 stringPtr(req.Phone),
		Wechat:                stringPtr(req.Wechat),
		Landline:              stringPtr(req.Landline),
		Address:               stringPtr(req.Address),
		District:              stringPtr(req.District),
		ContactName:           stringPtr(req.ContactName),
		ContactPhone:          stringPtr(req.ContactPhone),
	}
	if err := insertPatientBasicInfo(tx, basicInfo); err != nil {
		return err
	}

	return nil
}

// isDuplicateKeyError 检查是否是主键冲突错误
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL 错误代码 23505 表示 unique_violation
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "23505")
}

// UpdateRequest 更新患者请求
type UpdateRequest struct {
	BedNumber     *string  `json:"bedNumber"`
	Diagnosis     *string  `json:"diagnosis"`
	RiskLevel     *string  `json:"riskLevel"`
	Status        *string  `json:"status"`
	PatientType   *string  `json:"patientType"`
	InsuranceType *string  `json:"insuranceType"`
	DryWeight     *float64 `json:"dryWeight"`
	DefaultMode   *string  `json:"defaultMode"`
	DoctorID      *string  `json:"doctorId"`
	DoctorName    *string  `json:"doctorName"`
}

// Update 更新患者
func (s *PatientService) Update(id string, req UpdateRequest) (*models.Patient, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patient models.Patient
	if err := s.db.First(&patient, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.BedNumber != nil {
		updates["bed_number"] = *req.BedNumber
	}
	if req.Diagnosis != nil {
		updates["diagnosis"] = *req.Diagnosis
	}
	if req.RiskLevel != nil {
		updates["risk_level"] = *req.RiskLevel
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.PatientType != nil {
		updates["patient_type"] = *req.PatientType
	}
	if req.InsuranceType != nil {
		updates["insurance_type"] = *req.InsuranceType
	}
	if req.DryWeight != nil {
		updates["dry_weight"] = *req.DryWeight
	}
	if req.DefaultMode != nil {
		updates["default_mode"] = *req.DefaultMode
	}
	if req.DoctorID != nil {
		updates["doctor_id"] = *req.DoctorID
	}
	if req.DoctorName != nil {
		updates["doctor_name"] = *req.DoctorName
	}

	if err := s.db.Model(&patient).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.
		Preload("VascularAccesses").
		Preload("MedicalHistory").
		Preload("TreatmentPlan").
		First(&patient, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &patient, nil
}

// Delete 删除患者（硬删除 - 从数据库中真正删除）
func (s *PatientService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	// 先删除关联的基本信息档案
	s.db.Where("patient_id = ?", id).Delete(&models.PatientBasicInfo{})

	// 硬删除患者记录
	result := s.db.Delete(&models.Patient{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("patient not found")
	}

	return nil
}

// parseTimePointer 解析时间字符串指针
func parseTimePointer(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

// insertPatientBasicInfo 使用显式列插入，避免 GORM 在 *time.Time 字段上的反射写入 panic。
func insertPatientBasicInfo(tx *gorm.DB, basicInfo models.PatientBasicInfo) error {
	now := time.Now()
	idType := basicInfo.IDType
	if strings.TrimSpace(idType) == "" {
		idType = models.IDTypeIDCard
	}

	return tx.Table("patient_basic_infos").Create(map[string]interface{}{
		"id":                      basicInfo.ID,
		"patient_id":              basicInfo.PatientID,
		"pinyin":                  basicInfo.Pinyin,
		"birthday":                basicInfo.Birthday,
		"ethnicity":               basicInfo.Ethnicity,
		"id_type":                 idType,
		"id_number":               basicInfo.IDNumber,
		"visit_category":          basicInfo.VisitCategory,
		"admission_no":            basicInfo.AdmissionNo,
		"visit_no":                basicInfo.VisitNo,
		"medical_record_no":       basicInfo.MedicalRecordNo,
		"insurance_no":            basicInfo.InsuranceNo,
		"hdis_patient_id":         basicInfo.HdisPatientID,
		"dialysis_no":             basicInfo.DialysisNo,
		"nurse_name":              basicInfo.NurseName,
		"first_dialysis_date":     basicInfo.FirstDialysisDate,
		"first_hospital_date":     basicInfo.FirstHospitalDate,
		"first_dialysis_hospital": basicInfo.FirstDialysisHospital,
		"height":                  basicInfo.Height,
		"abo_blood_type":          basicInfo.ABOBloodType,
		"rh_blood_type":           basicInfo.RhBloodType,
		"education_level":         basicInfo.EducationLevel,
		"occupation":              basicInfo.Occupation,
		"marital_status":          basicInfo.MaritalStatus,
		"workplace":               basicInfo.Workplace,
		"phone":                   basicInfo.Phone,
		"wechat":                  basicInfo.Wechat,
		"landline":                basicInfo.Landline,
		"address":                 basicInfo.Address,
		"district":                basicInfo.District,
		"contact_name":            basicInfo.ContactName,
		"contact_phone":           basicInfo.ContactPhone,
		"created_at":              now,
		"updated_at":              now,
	}).Error
}

// stringPtr 将字符串转换为指针，空字符串返回 nil
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ===== 治疗方案相关方法 =====

// GetTreatmentPlans 获取患者的所有治疗方案
func (s *PatientService) GetTreatmentPlans(patientID string) ([]*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var plans []*models.TreatmentPlan
	err := s.db.Where("patient_id = ?", patientID).Find(&plans).Error
	if err != nil {
		return nil, err
	}

	return plans, nil
}

// GetTreatmentPlan 获取患者治疗方案（可指定透析模式）
func (s *PatientService) GetTreatmentPlan(patientID string, mode ...string) (*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var plan models.TreatmentPlan
	var err error

	// 如果提供了模式参数，则查询特定模式的治疗方案
	if len(mode) > 0 && mode[0] != "" {
		err = s.db.Where("patient_id = ? AND dialysis_mode->>'mode' = ?", patientID, mode[0]).First(&plan).Error
	} else {
		// 否则返回患者的第一个治疗方案
		err = s.db.Where("patient_id = ?", patientID).First(&plan).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有治疗方案，返回 nil 而不是错误
		}
		return nil, err
	}

	return &plan, nil
}

// CreateTreatmentPlanRequest 创建治疗方案请求
type CreateTreatmentPlanRequest struct {
	WeeklyFrequency    int                       `json:"weeklyFrequency"`
	BiweeklyFrequency  int                       `json:"biweeklyFrequency"`
	Duration           int                       `json:"duration"`
	DryWeight          float64                   `json:"dryWeight"`
	ExtraWeight        float64                   `json:"extraWeight"`
	Status             string                    `json:"status"`
	Notes              string                    `json:"notes"`
	DialysisMode       models.DialysisMode       `json:"dialysisMode"`
	Anticoagulant      models.Anticoagulant      `json:"anticoagulant"`
	DialysisParameters models.DialysisParameters `json:"parameters"`
	Materials          models.MaterialList       `json:"materials"`
}

// CreateTreatmentPlan 创建患者治疗方案
func (s *PatientService) CreateTreatmentPlan(patientID string, req CreateTreatmentPlanRequest) (*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 检查患者是否存在
	var patient models.Patient
	if err := s.db.First(&patient, "id = ?", patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	// 检查该患者是否已存在相同透析模式的治疗方案
	var existingPlan models.TreatmentPlan
	err := s.db.Where("patient_id = ? AND dialysis_mode->>'mode' = ?", patientID, req.DialysisMode.Mode).First(&existingPlan).Error
	if err == nil {
		return nil, fmt.Errorf("该患者已存在 %s 模式的治疗方案", req.DialysisMode.Mode)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	plan := models.TreatmentPlan{
		ID:                 utils.GenerateID(),
		PatientID:          patientID,
		WeeklyFrequency:    req.WeeklyFrequency,
		BiweeklyFrequency:  req.BiweeklyFrequency,
		Duration:           req.Duration,
		DryWeight:          req.DryWeight,
		ExtraWeight:        req.ExtraWeight,
		Status:             req.Status,
		Notes:              req.Notes,
		DialysisMode:       req.DialysisMode,
		Anticoagulant:      req.Anticoagulant,
		DialysisParameters: req.DialysisParameters,
		Materials:          req.Materials,
	}

	if err := s.db.Create(&plan).Error; err != nil {
		return nil, err
	}

	return &plan, nil
}

// UpdateTreatmentPlanRequest 更新治疗方案请求
type UpdateTreatmentPlanRequest struct {
	WeeklyFrequency    *int                       `json:"weeklyFrequency"`
	BiweeklyFrequency  *int                       `json:"biweeklyFrequency"`
	Duration           *int                       `json:"duration"`
	DryWeight          *float64                   `json:"dryWeight"`
	ExtraWeight        *float64                   `json:"extraWeight"`
	Status             *string                    `json:"status"`
	Notes              *string                    `json:"notes"`
	DialysisMode       *models.DialysisMode       `json:"dialysisMode"`
	Anticoagulant      *models.Anticoagulant      `json:"anticoagulant"`
	DialysisParameters *models.DialysisParameters `json:"parameters"`
	Materials          *models.MaterialList       `json:"materials"`
}

// UpdateTreatmentPlan 更新患者治疗方案
func (s *PatientService) UpdateTreatmentPlan(patientID string, req UpdateTreatmentPlanRequest) (*models.TreatmentPlan, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var plan models.TreatmentPlan
	// 必须指定透析模式才能精确匹配方案
	if req.DialysisMode == nil || req.DialysisMode.Mode == "" {
		return nil, errors.New("dialysisMode.mode is required for update")
	}
	query := s.db.Where("patient_id = ? AND dialysis_mode->>'mode' = ?", patientID, req.DialysisMode.Mode)
	if err := query.First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("treatment plan not found")
		}
		return nil, err
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.WeeklyFrequency != nil {
		updates["weekly_frequency"] = *req.WeeklyFrequency
	}
	if req.BiweeklyFrequency != nil {
		updates["biweekly_frequency"] = *req.BiweeklyFrequency
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.DryWeight != nil {
		updates["dry_weight"] = *req.DryWeight
	}
	if req.ExtraWeight != nil {
		updates["extra_weight"] = *req.ExtraWeight
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.DialysisMode != nil {
		updates["dialysis_mode"] = *req.DialysisMode
	}
	if req.Anticoagulant != nil {
		updates["anticoagulant"] = *req.Anticoagulant
	}
	if req.DialysisParameters != nil {
		updates["dialysis_parameters"] = *req.DialysisParameters
	}
	if req.Materials != nil {
		updates["materials"] = *req.Materials
	}

	if err := s.db.Model(&plan).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.Where("patient_id = ?", patientID).Where("id = ?", plan.ID).First(&plan).Error; err != nil {
		return nil, err
	}

	return &plan, nil
}

// DeleteTreatmentPlan 删除患者治疗方案
func (s *PatientService) DeleteTreatmentPlan(patientID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Where("patient_id = ?", patientID).Delete(&models.TreatmentPlan{})
	if result.Error != nil {
		return result.Error
	}
	// 不检查 RowsAffected，允许删除不存在的方案
	return nil
}

// ===== 方案调整记录 =====

// CreateAdjustmentRecordRequest 创建调整记录请求
type CreateAdjustmentRecordRequest struct {
	Content  string `json:"content" binding:"required"`
	Operator string `json:"operator"`
}

// GetAdjustmentRecords 获取患者方案调整记录列表
func (s *PatientService) GetAdjustmentRecords(patientID string) ([]models.AdjustmentRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var records []models.AdjustmentRecord
	if err := s.db.Where("patient_id = ?", patientID).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// CreateAdjustmentRecord 创建方案调整记录
func (s *PatientService) CreateAdjustmentRecord(patientID string, req CreateAdjustmentRecordRequest) (*models.AdjustmentRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 校验患者是否存在
	var count int64
	if err := s.db.Model(&models.Patient{}).Where("id = ?", patientID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	record := models.AdjustmentRecord{
		ID:        utils.GenerateID(),
		PatientID: patientID,
		Content:   req.Content,
		Operator:  req.Operator,
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}
