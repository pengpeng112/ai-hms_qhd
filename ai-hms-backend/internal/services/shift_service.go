package services

import (
	"errors"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// ShiftService 班次服务
type ShiftService struct {
	db *gorm.DB
}

// NewShiftService 创建班次服务
func NewShiftService() *ShiftService {
	return &ShiftService{
		db: database.GetDB(),
	}
}

// List 获取班次列表
func (s *ShiftService) List() ([]models.Shift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var shifts []models.Shift
	err := s.db.
		Where("is_disabled = ?", false).
		Order("sort ASC, create_time DESC").
		Find(&shifts).Error

	if err != nil {
		return nil, err
	}

	return shifts, nil
}

// Get 获取班次详情
func (s *ShiftService) Get(id int64) (*models.Shift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var shift models.Shift
	err := s.db.First(&shift, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("shift not found")
		}
		return nil, err
	}

	return &shift, nil
}

// CreateRequest 创建班次请求
type ShiftCreateRequest struct {
	Name       string `json:"name" binding:"required"`
	StartTime string `json:"startTime" binding:"required"` // HH:MM
	EndTime   string `json:"endTime" binding:"required"`   // HH:MM
	Type       string `json:"type"`
	Sort       *int   `json:"sort"`
	Notes      string `json:"notes"`
}

// Create 创建班次
func (s *ShiftService) Create(req ShiftCreateRequest) (*models.Shift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	shift := models.Shift{
		TenantId:  1, // TODO: 从上下文获取
		Name:      req.Name,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Type:      req.Type,
		IsDisabled: false,
		Sort:      0,
		Notes:     req.Notes,
		CreatorId: 1, // TODO: 从上下文获取
	}

	if req.Sort != nil {
		shift.Sort = *req.Sort
	}

	if err := s.db.Create(&shift).Error; err != nil {
		return nil, err
	}

	return &shift, nil
}

// UpdateRequest 更新班次请求
type ShiftUpdateRequest struct {
	Name       *string `json:"name"`
	StartTime *string `json:"startTime"`
	EndTime   *string `json:"endTime"`
	Type       *string `json:"type"`
	IsDisabled *bool   `json:"isDisabled"`
	Sort       *int    `json:"sort"`
	Notes      *string `json:"notes"`
}

// Update 更新班次
func (s *ShiftService) Update(id int64, req ShiftUpdateRequest) (*models.Shift, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var shift models.Shift
	if err := s.db.First(&shift, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("shift not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.IsDisabled != nil {
		updates["is_disabled"] = *req.IsDisabled
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if err := s.db.Model(&shift).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的数据
	if err := s.db.First(&shift, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &shift, nil
}

// Delete 删除班次
func (s *ShiftService) Delete(id int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Delete(&models.Shift{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("shift not found")
	}

	return nil
}
