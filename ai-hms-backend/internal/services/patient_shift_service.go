package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// PatientShiftService 患者排班服务
type PatientShiftService struct {
	db *gorm.DB
}

// NewPatientShiftService 创建患者排班服务
func NewPatientShiftService() *PatientShiftService {
	return &PatientShiftService{
		db: database.GetDB(),
	}
}

// ListRequest 获取患者排班列表请求
type PatientShiftListRequest struct {
	Page      int        `form:"page"`
	PageSize  int        `form:"pageSize"`
	PatientId *int64     `form:"patientId"`
	ShiftId   *int64     `form:"shiftId"`
	WardId    *int64     `form:"wardId"`
	BedId     *int64     `form:"bedId"`
	StartDate *time.Time `form:"startDate"`
	EndDate   *time.Time `form:"endDate"`
	Status    *int       `form:"status"`
}

// ListResponse 获取患者排班列表响应
type PatientShiftListResponse struct {
	Items     []models.PatientShift `json:"items"`
	Total     int64                 `json:"total"`
	Page      int                   `json:"page"`
	PageSize  int                   `json:"pageSize"`
	TotalPage int                   `json:"totalPage"`
}

// List 获取患者排班列表
func (s *PatientShiftService) List(req PatientShiftListRequest) (*PatientShiftListResponse, error) {
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

	query := s.db.Model(&models.PatientShift{})

	// 筛选条件
	if req.PatientId != nil {
		query = query.Where("patient_id = ?", *req.PatientId)
	}
	if req.ShiftId != nil {
		query = query.Where("shift_id = ?", *req.ShiftId)
	}
	if req.WardId != nil {
		query = query.Where("ward_id = ?", *req.WardId)
	}
	if req.BedId != nil {
		query = query.Where("bed_id = ?", *req.BedId)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.StartDate != nil {
		query = query.Where("schedule_date >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("schedule_date <= ?", *req.EndDate)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var items []models.PatientShift
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Preload("Patient").
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		Offset(offset).
		Limit(req.PageSize).
		Order("schedule_date DESC, create_time DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &PatientShiftListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// Get 获取患者排班详情
func (s *PatientShiftService) Get(id int64) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	err := s.db.
		Preload("Patient").
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		First(&patientShift, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient shift not found")
		}
		return nil, err
	}

	return &patientShift, nil
}

// CreateRequest 创建患者排班请求
type PatientShiftCreateRequest struct {
	PatientId    int64     `json:"patientId" binding:"required"`
	ScheduleDate time.Time `json:"scheduleDate" binding:"required"`
	ShiftId      int64     `json:"shiftId" binding:"required"`
	BedId        *int64    `json:"bedId"`
	WardId       *int64    `json:"wardId"`
	Notes        string    `json:"notes"`
}

// Create 创建患者排班
func (s *PatientShiftService) Create(req PatientShiftCreateRequest, tenantId, creatorId int64) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	patientShift := models.PatientShift{
		TenantId:     tenantId,
		PatientId:    req.PatientId,
		ScheduleDate: req.ScheduleDate,
		ShiftId:      req.ShiftId,
		BedId:        req.BedId,
		WardId:       req.WardId,
		Status:       models.PatientShiftStatusPending,
		IsDisabled:   false,
		Notes:        req.Notes,
		CreatorId:    creatorId,
	}

	if err := s.db.Create(&patientShift).Error; err != nil {
		return nil, err
	}

	return &patientShift, nil
}

// UpdateRequest 更新患者排班请求
type PatientShiftUpdateRequest struct {
	ShiftId *int64  `json:"shiftId"`
	BedId   *int64  `json:"bedId"`
	WardId  *int64  `json:"wardId"`
	Status  *int    `json:"status"`
	Notes   *string `json:"notes"`
}

// Update 更新患者排班
func (s *PatientShiftService) Update(id int64, req PatientShiftUpdateRequest) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	if err := s.db.First(&patientShift, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient shift not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.ShiftId != nil {
		updates["shift_id"] = *req.ShiftId
	}
	if req.BedId != nil {
		updates["bed_id"] = *req.BedId
	}
	if req.WardId != nil {
		updates["ward_id"] = *req.WardId
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if err := s.db.Model(&patientShift).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.
		Preload("Patient").
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		First(&patientShift, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &patientShift, nil
}

// Delete 删除患者排班
func (s *PatientShiftService) Delete(id int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.PatientShift{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("patient shift not found")
	}

	return nil
}

// GetByPatientAndDate 根据患者ID和日期获取排班
func (s *PatientShiftService) GetByPatientAndDate(patientId int64, date time.Time) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	err := s.db.
		Where("patient_id = ? AND schedule_date = ?", patientId, date).
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		First(&patientShift).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有排班记录
		}
		return nil, err
	}

	return &patientShift, nil
}

// CheckConflict 检查排班冲突
func (s *PatientShiftService) CheckConflict(patientId int64, date time.Time, shiftId int64, excludeId *int64) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	query := s.db.Model(&models.PatientShift{}).
		Where("patient_id = ? AND schedule_date = ? AND shift_id = ?", patientId, date, shiftId)

	if excludeId != nil {
		query = query.Where("id != ?", *excludeId)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
