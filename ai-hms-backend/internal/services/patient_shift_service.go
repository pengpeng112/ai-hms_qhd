package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
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
	StartDate *time.Time `form:"startDate" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"endDate" time_format:"2006-01-02"`
	Status    *int       `form:"status"`
	TenantId  int64      `form:"-"`
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
	if req.TenantId > 0 {
		query = query.Where("\"TenantId\" = ?", req.TenantId)
	}

	// 筛选条件
	if req.PatientId != nil {
		query = query.Where("\"PatientId\" = ?", *req.PatientId)
	}
	if req.ShiftId != nil {
		query = query.Where("\"ShiftId\" = ?", *req.ShiftId)
	}
	if req.WardId != nil {
		query = query.Where("\"WardId\" = ?", *req.WardId)
	}
	if req.BedId != nil {
		query = query.Where("\"BedId\" = ?", *req.BedId)
	}
	if req.Status != nil {
		query = query.Where("\"Status\" = ?", MapPatientShiftStatusNewToLegacy(*req.Status))
	}
	if req.StartDate != nil {
		query = query.Where("DATE(\"TreatmentTime\") >= DATE(?)", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("DATE(\"TreatmentTime\") <= DATE(?)", *req.EndDate)
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
		Order("\"TreatmentTime\" DESC, \"CreateTime\" DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	for i := range items {
		items[i].Status = MapPatientShiftStatusLegacyToNew(items[i].Status)
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
func (s *PatientShiftService) Get(id, tenantId int64) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	err := s.db.
		Preload("Patient").
		Preload("Shift").
		Preload("Bed").
		Preload("Ward").
		Where("\"Id\" = ?", id).
		Where("\"TenantId\" = ?", tenantId).
		First(&patientShift).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient shift not found")
		}
		return nil, err
	}

	patientShift.Status = MapPatientShiftStatusLegacyToNew(patientShift.Status)
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
		PatientId:    modeltypes.LegacyID(req.PatientId),
		ScheduleDate: req.ScheduleDate,
		ShiftId:      req.ShiftId,
		BedId:        req.BedId,
		WardId:       req.WardId,
		Status:       MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusPending),
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
func (s *PatientShiftService) Update(id, tenantId int64, req PatientShiftUpdateRequest) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	if err := s.db.Where("\"Id\" = ?", id).Where("\"TenantId\" = ?", tenantId).First(&patientShift).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient shift not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.ShiftId != nil {
		updates["ShiftId"] = *req.ShiftId
	}
	if req.BedId != nil {
		updates["BedId"] = *req.BedId
	}
	if req.WardId != nil {
		updates["WardId"] = *req.WardId
	}
	if req.Status != nil {
		updates["Status"] = MapPatientShiftStatusNewToLegacy(*req.Status)
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
		Where("\"Id\" = ?", id).
		Where("\"TenantId\" = ?", tenantId).
		First(&patientShift).Error; err != nil {
		return nil, err
	}

	patientShift.Status = MapPatientShiftStatusLegacyToNew(patientShift.Status)
	if req.Notes != nil {
		patientShift.Notes = *req.Notes
	}

	return &patientShift, nil
}

// Delete 删除患者排班
func (s *PatientShiftService) Delete(id, tenantId int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Model(&models.PatientShift{}).
		Where("\"Id\" = ?", id).
		Where("\"TenantId\" = ?", tenantId).
		Update("Status", MapPatientShiftStatusNewToLegacy(models.PatientShiftStatusCancelled))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("patient shift not found")
	}

	return nil
}

// GetByPatientAndDate 根据患者ID和日期获取排班
func (s *PatientShiftService) GetByPatientAndDate(patientId, tenantId int64, date time.Time) (*models.PatientShift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var patientShift models.PatientShift
	err := s.db.
		Where("\"TenantId\" = ?", tenantId).
		Where("\"PatientId\" = ? AND DATE(\"TreatmentTime\") = DATE(?)", patientId, date).
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

	patientShift.Status = MapPatientShiftStatusLegacyToNew(patientShift.Status)
	return &patientShift, nil
}

// CheckConflict 检查排班冲突
func (s *PatientShiftService) CheckConflict(patientId, tenantId int64, date time.Time, shiftId int64, excludeId *int64) (bool, error) {
	if s.db == nil {
		return false, errors.New("database not available")
	}

	query := s.db.Model(&models.PatientShift{}).
		Where("\"TenantId\" = ?", tenantId).
		Where("\"PatientId\" = ? AND DATE(\"TreatmentTime\") = DATE(?) AND \"ShiftId\" = ?", patientId, date, shiftId)

	if excludeId != nil {
		query = query.Where("\"Id\" != ?", *excludeId)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
