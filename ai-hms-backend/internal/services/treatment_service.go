package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// TreatmentService 透析治疗服务
type TreatmentService struct {
	db *gorm.DB
}

// NewTreatmentService 创建透析治疗服务
func NewTreatmentService() *TreatmentService {
	return &TreatmentService{
		db: database.GetDB(),
	}
}

// ListRequest 获取治疗记录列表请求
type TreatmentListRequest struct {
	Page               int        `form:"page"`
	PageSize           int        `form:"pageSize"`
	PatientId          *int64     `form:"patientId"`
	Status             *int       `form:"status"`
	Type               *int       `form:"type"`
	TreatmentDate      *time.Time `form:"treatmentDate" time_format:"2006-01-02"`
	TreatmentDateStart *time.Time `form:"treatmentDateStart" time_format:"2006-01-02"`
	TreatmentDateEnd   *time.Time `form:"treatmentDateEnd" time_format:"2006-01-02"`
}

// ListResponse 获取治疗记录列表响应
type TreatmentListResponse struct {
	Items     []models.Treatment `json:"items"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"pageSize"`
	TotalPage int                `json:"totalPage"`
}

// List 获取治疗记录列表
func (s *TreatmentService) List(req TreatmentListRequest) (*TreatmentListResponse, error) {
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

	query := s.db.Model(&models.Treatment{})

	// 筛选条件
	if req.PatientId != nil {
		query = query.Where("patient_id = ?", *req.PatientId)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.TreatmentDate != nil {
		query = query.Where("DATE(treatment_date) = DATE(?)", *req.TreatmentDate)
	}
	if req.TreatmentDateStart != nil {
		query = query.Where("DATE(treatment_date) >= DATE(?)", *req.TreatmentDateStart)
	}
	if req.TreatmentDateEnd != nil {
		query = query.Where("DATE(treatment_date) <= DATE(?)", *req.TreatmentDateEnd)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.Treatment
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Preload("Patient").
		Preload("Ward").
		Preload("Bed").
		Preload("Shift").
		Preload("BeforeCheck").
		Preload("BeforeSigns").
		Preload("AfterSigns").
		Offset(offset).
		Limit(req.PageSize).
		Order("treatment_date DESC, create_time DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &TreatmentListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取治疗记录详情
func (s *TreatmentService) Get(id int64) (*models.Treatment, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var treatment models.Treatment
	err := s.db.
		Preload("Patient").
		Preload("Schedule").
		Preload("Ward").
		Preload("Bed").
		Preload("Shift").
		Preload("BeforeCheck").
		Preload("BeforeSigns").
		Preload("AfterSigns").
		Preload("DuringParams", func(db *gorm.DB) *gorm.DB {
			return db.Order("record_time ASC")
		}).
		Preload("Alarms", func(db *gorm.DB) *gorm.DB {
			return db.Order("alarm_time ASC")
		}).
		First(&treatment, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("treatment not found")
		}
		return nil, err
	}

	return &treatment, nil
}

// CreateRequest 创建治疗记录请求
type TreatmentCreateRequest struct {
	PatientId      int64      `json:"patientId" binding:"required"`
	TreatmentDate  time.Time  `json:"treatmentDate" binding:"required"`
	ScheduleId     *int64     `json:"scheduleId"`
	ReceptionDrId  *int64     `json:"receptionDrId"`
	SignInTime     *time.Time `json:"signInTime"`
	QueueNo        string     `json:"queueNo"`
	ReceptionTime  *time.Time `json:"receptionTime"`
	DayProgrammeId *int64     `json:"dayProgrammeId"`
	WardId         *int64     `json:"wardId"`
	WardName       string     `json:"wardName"`
	BedId          *int64     `json:"bedId"`
	ShiftId        *int64     `json:"shiftId"`
	ShiftTiming    int        `json:"shiftTiming"`
	Type           int        `json:"type" binding:"required"`
	Status         int        `json:"status"`
}

// Create 创建治疗记录
func (s *TreatmentService) Create(req TreatmentCreateRequest, tenantId, creatorId int64) (*models.Treatment, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 如果没有指定状态，默认为待开始
	status := req.Status
	if status == 0 {
		status = models.TreatmentStatusPending
	}

	treatment := models.Treatment{
		TenantId:       tenantId,
		PatientId:      req.PatientId,
		TreatmentDate:  req.TreatmentDate,
		ScheduleId:     req.ScheduleId,
		ReceptionDrId:  req.ReceptionDrId,
		SignInTime:     req.SignInTime,
		QueueNo:        req.QueueNo,
		ReceptionTime:  req.ReceptionTime,
		DayProgrammeId: req.DayProgrammeId,
		WardId:         req.WardId,
		WardName:       req.WardName,
		BedId:          req.BedId,
		ShiftId:        req.ShiftId,
		ShiftTiming:    req.ShiftTiming,
		Type:           req.Type,
		Status:         status,
		IsDisabled:     false,
		CreatorId:      creatorId,
	}

	if err := s.db.Create(&treatment).Error; err != nil {
		return nil, err
	}

	return &treatment, nil
}

// UpdateRequest 更新治疗记录请求
type TreatmentUpdateRequest struct {
	SignInTime    *time.Time `json:"signInTime"`
	QueueNo       *string    `json:"queueNo"`
	ReceptionTime *time.Time `json:"receptionTime"`
	ReceptionDrId *int64     `json:"receptionDrId"`
	WardId        *int64     `json:"wardId"`
	WardName      *string    `json:"wardName"`
	BedId         *int64     `json:"bedId"`
	ShiftId       *int64     `json:"shiftId"`
	ShiftTiming   *int       `json:"shiftTiming"`
	Status        *int       `json:"status"`
	IsDisabled    *bool      `json:"isDisabled"`
}

// Update 更新治疗记录
func (s *TreatmentService) Update(id int64, req TreatmentUpdateRequest) (*models.Treatment, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var treatment models.Treatment
	if err := s.db.First(&treatment, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("treatment not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.SignInTime != nil {
		updates["sign_in_time"] = *req.SignInTime
	}
	if req.QueueNo != nil {
		updates["queue_no"] = *req.QueueNo
	}
	if req.ReceptionTime != nil {
		updates["reception_time"] = *req.ReceptionTime
	}
	if req.ReceptionDrId != nil {
		updates["reception_dr_id"] = *req.ReceptionDrId
	}
	if req.WardId != nil {
		updates["ward_id"] = *req.WardId
	}
	if req.WardName != nil {
		updates["ward_name"] = *req.WardName
	}
	if req.BedId != nil {
		updates["bed_id"] = *req.BedId
	}
	if req.ShiftId != nil {
		updates["shift_id"] = *req.ShiftId
	}
	if req.ShiftTiming != nil {
		updates["shift_timing"] = *req.ShiftTiming
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.IsDisabled != nil {
		updates["is_disabled"] = *req.IsDisabled
	}

	if err := s.db.Model(&treatment).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	return s.Get(id)
}

// Delete 删除治疗记录
func (s *TreatmentService) Delete(id int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.Treatment{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("treatment not found")
	}

	return nil
}

// UpdateStatus 更新治疗状态
func (s *TreatmentService) UpdateStatus(id int64, status int) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Model(&models.Treatment{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("treatment not found")
	}

	return nil
}

// GetByPatientAndDate 获取患者在指定日期的治疗记录
func (s *TreatmentService) GetByPatientAndDate(patientId int64, date time.Time) (*models.Treatment, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var treatment models.Treatment
	err := s.db.
		Where("patient_id = ? AND DATE(treatment_date) = DATE(?)", patientId, date).
		Preload("BeforeCheck").
		Preload("BeforeSigns").
		Preload("AfterSigns").
		Preload("DuringParams").
		First(&treatment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &treatment, nil
}
