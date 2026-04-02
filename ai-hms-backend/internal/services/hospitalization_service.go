package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// HospitalizationService 住院信息服务
type HospitalizationService struct {
	db *gorm.DB
}

// NewHospitalizationService 创建住院信息服务
func NewHospitalizationService() *HospitalizationService {
	return &HospitalizationService{
		db: database.GetDB(),
	}
}

// ListRequest 获取住院信息列表请求
type HospitalizationListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	PatientId *int64 `form:"patientId"`
	Status    *int   `form:"status"`
	HospWard  string `form:"hospWard"`
}

// ListResponse 获取住院信息列表响应
type HospitalizationListResponse struct {
	Items     []models.Hospitalization `json:"items"`
	Total     int64                    `json:"total"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"pageSize"`
	TotalPage int                      `json:"totalPage"`
}

// List 获取住院信息列表
func (s *HospitalizationService) List(req HospitalizationListRequest) (*HospitalizationListResponse, error) {
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

	query := s.db.Model(&models.Hospitalization{})

	// 筛选条件
	if req.PatientId != nil {
		query = query.Where("patient_id = ?", *req.PatientId)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.HospWard != "" {
		query = query.Where("hosp_ward LIKE ?", "%"+req.HospWard+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.Hospitalization
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Preload("Patient").
		Offset(offset).
		Limit(req.PageSize).
		Order("create_time DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &HospitalizationListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取住院信息详情
func (s *HospitalizationService) Get(id int64) (*models.Hospitalization, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var hospitalization models.Hospitalization
	err := s.db.
		Preload("Patient").
		First(&hospitalization, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("hospitalization not found")
		}
		return nil, err
	}

	return &hospitalization, nil
}

// CreateRequest 创建住院信息请求
type HospitalizationCreateRequest struct {
	PatientId       int64      `json:"patientId" binding:"required"`
	CaseNo          string     `json:"caseNo"`
	HospNo          string     `json:"hospNo"`
	BarCode         string     `json:"barCode"`
	HospPatientType string     `json:"hospPatientType"`
	HospReceiveDept string     `json:"hospReceiveDept"`
	HospWard        string     `json:"hospWard"`
	HospBed         string     `json:"hospBed"`
	AttendDr        string     `json:"attendDr"`
	ReceptionDr     string     `json:"receptionDr"`
	AdmissionDate   *time.Time `json:"admissionDate"`
	Notes           string     `json:"notes"`
}

// Create 创建住院信息
func (s *HospitalizationService) Create(req HospitalizationCreateRequest) (*models.Hospitalization, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	hospitalization := models.Hospitalization{
		TenantId:        1, // TODO: 从上下文获取
		PatientId:       req.PatientId,
		CaseNo:          req.CaseNo,
		HospNo:          req.HospNo,
		BarCode:         req.BarCode,
		HospPatientType: req.HospPatientType,
		HospReceiveDept: req.HospReceiveDept,
		HospWard:        req.HospWard,
		HospBed:         req.HospBed,
		AttendDr:        req.AttendDr,
		ReceptionDr:     req.ReceptionDr,
		Status:          models.HospitalizationStatusInPatient,
		AdmissionDate:   req.AdmissionDate,
		Notes:           req.Notes,
		CreatorId:       1, // TODO: 从上下文获取
	}

	if err := s.db.Create(&hospitalization).Error; err != nil {
		return nil, err
	}

	return &hospitalization, nil
}

// UpdateRequest 更新住院信息请求
type HospitalizationUpdateRequest struct {
	HospWard        *string    `json:"hospWard"`
	HospBed         *string    `json:"hospBed"`
	AttendDr        *string    `json:"attendDr"`
	Status          *int       `json:"status"`
	DischargeDate   *time.Time `json:"dischargeDate"`
	Notes           *string    `json:"notes"`
}

// Update 更新住院信息
func (s *HospitalizationService) Update(id int64, req HospitalizationUpdateRequest) (*models.Hospitalization, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var hospitalization models.Hospitalization
	if err := s.db.First(&hospitalization, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("hospitalization not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.HospWard != nil {
		updates["hosp_ward"] = *req.HospWard
	}
	if req.HospBed != nil {
		updates["hosp_bed"] = *req.HospBed
	}
	if req.AttendDr != nil {
		updates["attend_dr"] = *req.AttendDr
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.DischargeDate != nil {
		updates["discharge_date"] = *req.DischargeDate
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if err := s.db.Model(&hospitalization).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.
		Preload("Patient").
		First(&hospitalization, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &hospitalization, nil
}

// Delete 删除住院信息
func (s *HospitalizationService) Delete(id int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.Hospitalization{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("hospitalization not found")
	}

	return nil
}

// GetByPatientId 获取患者的当前住院信息
func (s *HospitalizationService) GetByPatientId(patientId int64) (*models.Hospitalization, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var hospitalization models.Hospitalization
	err := s.db.
		Where("patient_id = ? AND status = ?", patientId, models.HospitalizationStatusInPatient).
		Preload("Patient").
		First(&hospitalization).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有在院记录
		}
		return nil, err
	}

	return &hospitalization, nil
}
