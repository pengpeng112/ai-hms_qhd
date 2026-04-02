package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VascularAccessService 血管通路服务
type VascularAccessService struct {
	db *gorm.DB
}

// NewVascularAccessService 创建血管通路服务
func NewVascularAccessService() *VascularAccessService {
	return &VascularAccessService{
		db: database.GetDB(),
	}
}

// VascularAccessResponse 血管通路响应
type VascularAccessResponse struct {
	ID                string   `json:"id"`
	AccessType        string   `json:"accessType"`
	Site              string   `json:"site"`
	Artery            []string `json:"artery"`
	Vein              []string `json:"vein"`
	Side              string   `json:"side"`
	Hospital          string   `json:"hospital"`
	Surgeon           string   `json:"surgeon"`
	SurgeryDate       string   `json:"surgeryDate"`
	FirstUseDate      string   `json:"firstUseDate"`
	AccessNumber      int      `json:"accessNumber"`
	InterventionCount int      `json:"interventionCount"`
	InterventionDate  string   `json:"interventionDate"`
	CatheterMethod    *string  `json:"catheterMethod"`
	CatheterDepth     *string  `json:"catheterDepth"`
	VPuncturePosition []string `json:"vPuncturePosition"`
	APuncturePosition []string `json:"aPuncturePosition"`
	Notes             string   `json:"notes"`
	Images            []string `json:"images"`
	IsDefault         bool     `json:"isDefault"`
	IsDisabled        bool     `json:"isDisabled"`
	CreatedAt         string   `json:"createdAt"`
}

// VascularAccessRequest 创建/更新血管通路请求
type VascularAccessRequest struct {
	AccessType        string   `json:"accessType" binding:"required"`
	Site              string   `json:"site"`
	Artery            []string `json:"artery"`
	Vein              []string `json:"vein"`
	Side              string   `json:"side"`
	Hospital          string   `json:"hospital"`
	Surgeon           string   `json:"surgeon"`
	SurgeryDate       string   `json:"surgeryDate"`
	FirstUseDate      string   `json:"firstUseDate"`
	AccessNumber      int      `json:"accessNumber"`
	InterventionCount int      `json:"interventionCount"`
	InterventionDate  string   `json:"interventionDate"`
	CatheterMethod    *string  `json:"catheterMethod"`
	CatheterDepth     *string  `json:"catheterDepth"`
	VPuncturePosition []string `json:"vPuncturePosition"`
	APuncturePosition []string `json:"aPuncturePosition"`
	Notes             string   `json:"notes"`
	Images            []string `json:"images"`
	IsDefault         bool     `json:"isDefault"`
	IsDisabled        bool     `json:"isDisabled"`
}

// List 获取患者的血管通路列表
func (s *VascularAccessService) List(patientID string) ([]VascularAccessResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var accesses []models.VascularAccess
	err := s.db.Where("patient_id = ?", patientID).
		Order("is_default DESC, created_at DESC").
		Find(&accesses).Error
	if err != nil {
		return nil, err
	}

	result := make([]VascularAccessResponse, len(accesses))
	for i, a := range accesses {
		result[i] = s.buildResponse(a)
	}
	return result, nil
}

// Create 创建血管通路
func (s *VascularAccessService) Create(patientID string, req *VascularAccessRequest) (*VascularAccessResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 验证患者存在
	var count int64
	s.db.Model(&models.Patient{}).Where("id = ?", patientID).Count(&count)
	if count == 0 {
		return nil, errors.New("patient not found")
	}

	access := models.VascularAccess{
		ID:                uuid.New().String(),
		PatientID:         patientID,
		AccessType:        req.AccessType,
		Site:              req.Site,
		Artery:            req.Artery,
		Vein:              req.Vein,
		Side:              req.Side,
		Hospital:          req.Hospital,
		Surgeon:           req.Surgeon,
		AccessNumber:      req.AccessNumber,
		InterventionCount: req.InterventionCount,
		CatheterMethod:    req.CatheterMethod,
		CatheterDepth:     req.CatheterDepth,
		VPuncturePosition: req.VPuncturePosition,
		APuncturePosition: req.APuncturePosition,
		Notes:             req.Notes,
		Images:            req.Images,
		IsDefault:         req.IsDefault,
		IsDisabled:        req.IsDisabled,
	}

	// 解析日期
	if req.SurgeryDate != "" {
		if t, err := time.Parse("2006-01-02", req.SurgeryDate); err == nil {
			access.SurgeryDate = &t
		}
	}
	if req.FirstUseDate != "" {
		if t, err := time.Parse("2006-01-02", req.FirstUseDate); err == nil {
			access.FirstUseDate = &t
		}
	}
	if req.InterventionDate != "" {
		if t, err := time.Parse("2006-01-02", req.InterventionDate); err == nil {
			access.InterventionDate = &t
		}
	}

	// 如果设置为默认，需要将其他记录的默认状态取消
	if req.IsDefault {
		s.db.Model(&models.VascularAccess{}).
			Where("patient_id = ? AND is_default = ?", patientID, true).
			Update("is_default", false)
	}

	if err := s.db.Create(&access).Error; err != nil {
		return nil, err
	}

	resp := s.buildResponse(access)
	return &resp, nil
}

// Update 更新血管通路
func (s *VascularAccessService) Update(patientID, accessID string, req *VascularAccessRequest) (*VascularAccessResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var access models.VascularAccess
	err := s.db.Where("id = ? AND patient_id = ?", accessID, patientID).First(&access).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vascular access not found")
	}
	if err != nil {
		return nil, err
	}

	// 更新字段
	access.AccessType = req.AccessType
	access.Site = req.Site
	access.Artery = req.Artery
	access.Vein = req.Vein
	access.Side = req.Side
	access.Hospital = req.Hospital
	access.Surgeon = req.Surgeon
	access.AccessNumber = req.AccessNumber
	access.InterventionCount = req.InterventionCount
	access.CatheterMethod = req.CatheterMethod
	access.CatheterDepth = req.CatheterDepth
	access.VPuncturePosition = req.VPuncturePosition
	access.APuncturePosition = req.APuncturePosition
	access.Notes = req.Notes
	access.Images = req.Images
	access.IsDefault = req.IsDefault
	access.IsDisabled = req.IsDisabled

	// 解析日期
	if req.SurgeryDate != "" {
		if t, err := time.Parse("2006-01-02", req.SurgeryDate); err == nil {
			access.SurgeryDate = &t
		}
	} else {
		access.SurgeryDate = nil
	}
	if req.FirstUseDate != "" {
		if t, err := time.Parse("2006-01-02", req.FirstUseDate); err == nil {
			access.FirstUseDate = &t
		}
	} else {
		access.FirstUseDate = nil
	}
	if req.InterventionDate != "" {
		if t, err := time.Parse("2006-01-02", req.InterventionDate); err == nil {
			access.InterventionDate = &t
		}
	} else {
		access.InterventionDate = nil
	}

	// 如果设置为默认，需要将其他记录的默认状态取消
	if req.IsDefault {
		s.db.Model(&models.VascularAccess{}).
			Where("patient_id = ? AND id != ? AND is_default = ?", patientID, accessID, true).
			Update("is_default", false)
	}

	if err := s.db.Save(&access).Error; err != nil {
		return nil, err
	}

	resp := s.buildResponse(access)
	return &resp, nil
}

// Delete 删除血管通路
func (s *VascularAccessService) Delete(patientID, accessID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Where("id = ? AND patient_id = ?", accessID, patientID).Delete(&models.VascularAccess{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("vascular access not found")
	}
	return nil
}

func (s *VascularAccessService) buildResponse(a models.VascularAccess) VascularAccessResponse {
	resp := VascularAccessResponse{
		ID:                a.ID,
		AccessType:        a.AccessType,
		Site:              a.Site,
		Artery:            a.Artery,
		Vein:              a.Vein,
		Side:              a.Side,
		Hospital:          a.Hospital,
		Surgeon:           a.Surgeon,
		AccessNumber:      a.AccessNumber,
		InterventionCount: a.InterventionCount,
		CatheterMethod:    a.CatheterMethod,
		CatheterDepth:     a.CatheterDepth,
		VPuncturePosition: a.VPuncturePosition,
		APuncturePosition: a.APuncturePosition,
		Notes:             a.Notes,
		Images:            a.Images,
		IsDefault:         a.IsDefault,
		IsDisabled:        a.IsDisabled,
		CreatedAt:         a.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 确保数组不为 nil
	if resp.Artery == nil {
		resp.Artery = []string{}
	}
	if resp.Vein == nil {
		resp.Vein = []string{}
	}
	if resp.VPuncturePosition == nil {
		resp.VPuncturePosition = []string{}
	}
	if resp.APuncturePosition == nil {
		resp.APuncturePosition = []string{}
	}
	if resp.Images == nil {
		resp.Images = []string{}
	}

	// 格式化日期
	if a.SurgeryDate != nil {
		resp.SurgeryDate = a.SurgeryDate.Format("2006-01-02")
	}
	if a.FirstUseDate != nil {
		resp.FirstUseDate = a.FirstUseDate.Format("2006-01-02")
	}
	if a.InterventionDate != nil {
		resp.InterventionDate = a.InterventionDate.Format("2006-01-02")
	}

	return resp
}

// ===== 血管通路干预记录相关 =====

// VascularAccessInterventionResponse 干预记录响应
type VascularAccessInterventionResponse struct {
	ID                 string `json:"id"`
	VascularAccessID   string `json:"vascularAccessId"`
	PatientID          string `json:"patientId"`
	AccessType         string `json:"accessType"`
	AvgBloodFlow       int    `json:"avgBloodFlow"`
	UsageDays          int    `json:"usageDays"`
	SurgeryType        string `json:"surgeryType"`
	InterventionReason string `json:"interventionReason"`
	Doctor             string `json:"doctor"`
	InterventionDate   string `json:"interventionDate"`
	Description        string `json:"description"`
	CreatedAt          string `json:"createdAt"`
}

// VascularAccessInterventionRequest 创建干预记录请求
type VascularAccessInterventionRequest struct {
	VascularAccessID   string `json:"vascularAccessId" binding:"required"`
	AccessType         string `json:"accessType"`
	AvgBloodFlow       int    `json:"avgBloodFlow"`
	UsageDays          int    `json:"usageDays"`
	SurgeryType        string `json:"surgeryType" binding:"required"`
	InterventionReason string `json:"interventionReason" binding:"required"`
	Doctor             string `json:"doctor"`
	InterventionDate   string `json:"interventionDate" binding:"required"`
	Description        string `json:"description"`
}

// ListInterventions 获取血管通路的干预记录列表
func (s *VascularAccessService) ListInterventions(patientID, vascularAccessID string) ([]VascularAccessInterventionResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var interventions []models.VascularAccessIntervention
	query := s.db.Where("patient_id = ?", patientID)

	// 如果指定了血管通路ID，则过滤
	if vascularAccessID != "" {
		query = query.Where("vascular_access_id = ?", vascularAccessID)
	}

	err := query.Order("intervention_date DESC, created_at DESC").Find(&interventions).Error
	if err != nil {
		return nil, err
	}

	result := make([]VascularAccessInterventionResponse, len(interventions))
	for i, iv := range interventions {
		result[i] = s.buildInterventionResponse(iv)
	}
	return result, nil
}

// CreateIntervention 创建干预记录
func (s *VascularAccessService) CreateIntervention(patientID string, req *VascularAccessInterventionRequest) (*VascularAccessInterventionResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 验证血管通路存在且属于该患者
	var access models.VascularAccess
	err := s.db.Where("id = ? AND patient_id = ?", req.VascularAccessID, patientID).First(&access).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vascular access not found")
	}
	if err != nil {
		return nil, err
	}

	// 解析干预日期
	interventionDate, err := time.Parse("2006-01-02", req.InterventionDate)
	if err != nil {
		return nil, errors.New("invalid intervention date format")
	}

	intervention := models.VascularAccessIntervention{
		ID:                 uuid.New().String(),
		VascularAccessID:   req.VascularAccessID,
		PatientID:          patientID,
		AccessType:         req.AccessType,
		AvgBloodFlow:       req.AvgBloodFlow,
		UsageDays:          req.UsageDays,
		SurgeryType:        req.SurgeryType,
		InterventionReason: req.InterventionReason,
		Doctor:             req.Doctor,
		InterventionDate:   interventionDate,
		Description:        req.Description,
	}

	// 如果未指定通路类型，使用血管通路的类型
	if intervention.AccessType == "" {
		intervention.AccessType = access.AccessType
	}

	if err := s.db.Create(&intervention).Error; err != nil {
		return nil, err
	}

	// 更新血管通路的干预次数和干预日期
	access.InterventionCount++
	access.InterventionDate = &interventionDate
	s.db.Save(&access)

	resp := s.buildInterventionResponse(intervention)
	return &resp, nil
}

// DeleteIntervention 删除干预记录
func (s *VascularAccessService) DeleteIntervention(patientID, interventionID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Where("id = ? AND patient_id = ?", interventionID, patientID).Delete(&models.VascularAccessIntervention{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("intervention not found")
	}
	return nil
}

func (s *VascularAccessService) buildInterventionResponse(iv models.VascularAccessIntervention) VascularAccessInterventionResponse {
	return VascularAccessInterventionResponse{
		ID:                 iv.ID,
		VascularAccessID:   iv.VascularAccessID,
		PatientID:          iv.PatientID,
		AccessType:         iv.AccessType,
		AvgBloodFlow:       iv.AvgBloodFlow,
		UsageDays:          iv.UsageDays,
		SurgeryType:        iv.SurgeryType,
		InterventionReason: iv.InterventionReason,
		Doctor:             iv.Doctor,
		InterventionDate:   iv.InterventionDate.Format("2006-01-02"),
		Description:        iv.Description,
		CreatedAt:          iv.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
