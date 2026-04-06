package services

import (
	"errors"
	"fmt"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// DeviceService 设备服务
type DeviceService struct {
	db *gorm.DB
}

// NewDeviceService 创建设备服务
func NewDeviceService() *DeviceService {
	return &DeviceService{db: database.GetDB()}
}

// DeviceListRequest 获取设备列表请求
type DeviceListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	Status    string `form:"status"`
	BedNumber string `form:"bedNumber"`
	WardId    *int64 `form:"wardId"`
	Keyword   string `form:"keyword"` // 名称/序列号模糊搜索
}

// DeviceListResponse 设备列表响应
type DeviceListResponse struct {
	Items     []models.Device `json:"items"`
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"pageSize"`
	TotalPage int             `json:"totalPage"`
}

// List 获取设备列表
func (s *DeviceService) List(req DeviceListRequest) (*DeviceListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	query := s.db.Model(&models.Device{}).Where("is_disabled = ?", false)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.BedNumber != "" {
		query = query.Where("bed_number LIKE ?", "%"+req.BedNumber+"%")
	}
	if req.WardId != nil {
		query = query.Where("ward_id = ?", *req.WardId)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR serial_no LIKE ? OR model LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.Device
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("bed_number").Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &DeviceListResponse{
		Items: items, Total: total,
		Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage,
	}, nil
}

// DeviceCreateRequest 创建设备请求
type DeviceCreateRequest struct {
	Name         string `json:"name" binding:"required"`
	SerialNo     string `json:"serialNo"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	BedNumber    string `json:"bedNumber"`
	WardId       *int64 `json:"wardId"`
	Notes        string `json:"notes"`
}

// Create 创建设备
func (s *DeviceService) Create(req DeviceCreateRequest, tenantId, creatorId int64) (*models.Device, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	device := models.Device{
		ID:           fmt.Sprintf("DEV-%d", creatorId*1000+int64(s.count())+1),
		TenantId:     tenantId,
		Name:         req.Name,
		SerialNo:     req.SerialNo,
		Model:        req.Model,
		Manufacturer: req.Manufacturer,
		BedNumber:    req.BedNumber,
		WardId:       req.WardId,
		Status:       models.DeviceStatusOffline,
		Notes:        req.Notes,
		IsDisabled:   false,
		CreatorId:    creatorId,
	}

	if err := s.db.Create(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

// count 获取设备总数（用于生成 ID）
func (s *DeviceService) count() int {
	var count int64
	s.db.Model(&models.Device{}).Count(&count)
	return int(count)
}

// DeviceUpdateRequest 更新设备请求
type DeviceUpdateRequest struct {
	Name         *string `json:"name"`
	SerialNo     *string `json:"serialNo"`
	Model        *string `json:"model"`
	Manufacturer *string `json:"manufacturer"`
	BedNumber    *string `json:"bedNumber"`
	WardId       *int64  `json:"wardId"`
	Status       *string `json:"status"`
	Notes        *string `json:"notes"`
}

// Update 更新设备
func (s *DeviceService) Update(id string, req DeviceUpdateRequest) (*models.Device, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var device models.Device
	if err := s.db.First(&device, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("device not found")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.SerialNo != nil {
		updates["serial_no"] = *req.SerialNo
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.Manufacturer != nil {
		updates["manufacturer"] = *req.Manufacturer
	}
	if req.BedNumber != nil {
		updates["bed_number"] = *req.BedNumber
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

	if err := s.db.Model(&device).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

// Delete 删除设备（软删除）
func (s *DeviceService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	result := s.db.Model(&models.Device{}).Where("id = ?", id).Update("is_disabled", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("device not found")
	}
	return nil
}

// UpdateStatus 批量更新设备状态（供 HDIS 回调使用）
func (s *DeviceService) UpdateStatus(id, status string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	validStatuses := map[string]bool{
		models.DeviceStatusNormal:      true,
		models.DeviceStatusWarning:     true,
		models.DeviceStatusAlarm:       true,
		models.DeviceStatusOffline:     true,
		models.DeviceStatusMaintenance: true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	return s.db.Model(&models.Device{}).Where("id = ?", id).Update("status", status).Error
}
