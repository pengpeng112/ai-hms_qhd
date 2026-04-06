package services

import (
	"errors"
	"fmt"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

// InventoryService 库存服务
type InventoryService struct {
	db *gorm.DB
}

func NewInventoryService() *InventoryService {
	return &InventoryService{db: database.GetDB()}
}

// ─────────────── InventoryItem ───────────────

type InventoryListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Category string `form:"category"`
	Alert    *bool  `form:"alert"` // 仅返回告警项
	Keyword  string `form:"keyword"`
}

type InventoryListResponse struct {
	Items     []InventoryItemView `json:"items"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"pageSize"`
	TotalPage int                 `json:"totalPage"`
}

// InventoryItemView 在 InventoryItem 基础上增加计算字段
type InventoryItemView struct {
	models.InventoryItem
	Alert       bool   `json:"alert"`       // stock < minStock
	LastUpdated string `json:"lastUpdated"` // UpdatedAt 格式化
}

func (s *InventoryService) ListItems(req InventoryListRequest) (*InventoryListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	query := s.db.Model(&models.InventoryItem{}).Where("is_disabled = ?", false)
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR spec LIKE ? OR supplier LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.InventoryItem
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("category, name").Find(&items).Error; err != nil {
		return nil, err
	}

	views := make([]InventoryItemView, 0, len(items))
	for _, item := range items {
		alert := item.MinStock > 0 && item.Stock < item.MinStock
		if req.Alert != nil && *req.Alert != alert {
			continue
		}
		views = append(views, InventoryItemView{
			InventoryItem: item,
			Alert:         alert,
			LastUpdated:   item.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &InventoryListResponse{
		Items: views, Total: total,
		Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage,
	}, nil
}

type InventoryCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Spec     string `json:"spec"`
	Category string `json:"category"`
	Stock    int    `json:"stock"`
	Unit     string `json:"unit"`
	MinStock int    `json:"minStock"`
	MaxStock int    `json:"maxStock"`
	Location string `json:"location"`
	Supplier string `json:"supplier"`
}

func (s *InventoryService) CreateItem(req InventoryCreateRequest, tenantId, creatorId int64) (*InventoryItemView, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var count int64
	s.db.Model(&models.InventoryItem{}).Count(&count)

	item := models.InventoryItem{
		ID:        fmt.Sprintf("INV-%04d", count+1),
		TenantId:  tenantId,
		Name:      req.Name,
		Spec:      req.Spec,
		Category:  req.Category,
		Stock:     req.Stock,
		Unit:      req.Unit,
		MinStock:  req.MinStock,
		MaxStock:  req.MaxStock,
		Location:  req.Location,
		Supplier:  req.Supplier,
		CreatorId: creatorId,
	}

	if err := s.db.Create(&item).Error; err != nil {
		return nil, err
	}

	view := &InventoryItemView{
		InventoryItem: item,
		Alert:         item.MinStock > 0 && item.Stock < item.MinStock,
		LastUpdated:   item.UpdatedAt.Format("2006-01-02 15:04"),
	}
	return view, nil
}

type InventoryUpdateRequest struct {
	Name     *string `json:"name"`
	Spec     *string `json:"spec"`
	Category *string `json:"category"`
	Unit     *string `json:"unit"`
	MinStock *int    `json:"minStock"`
	MaxStock *int    `json:"maxStock"`
	Location *string `json:"location"`
	Supplier *string `json:"supplier"`
}

func (s *InventoryService) UpdateItem(id string, req InventoryUpdateRequest) (*InventoryItemView, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var item models.InventoryItem
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("item not found")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Spec != nil {
		updates["spec"] = *req.Spec
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Unit != nil {
		updates["unit"] = *req.Unit
	}
	if req.MinStock != nil {
		updates["min_stock"] = *req.MinStock
	}
	if req.MaxStock != nil {
		updates["max_stock"] = *req.MaxStock
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Supplier != nil {
		updates["supplier"] = *req.Supplier
	}

	if err := s.db.Model(&item).Updates(updates).Error; err != nil {
		return nil, err
	}

	view := &InventoryItemView{
		InventoryItem: item,
		Alert:         item.MinStock > 0 && item.Stock < item.MinStock,
		LastUpdated:   item.UpdatedAt.Format("2006-01-02 15:04"),
	}
	return view, nil
}

func (s *InventoryService) DeleteItem(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Model(&models.InventoryItem{}).Where("id = ?", id).Update("is_disabled", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}

// ─────────────── StockLog ───────────────

type StockLogListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	ItemId   string `form:"itemId"`
	Type     string `form:"type"` // in | out
}

type StockLogListResponse struct {
	Items     []models.StockLog `json:"items"`
	Total     int64             `json:"total"`
	Page      int               `json:"page"`
	PageSize  int               `json:"pageSize"`
	TotalPage int               `json:"totalPage"`
}

func (s *InventoryService) ListLogs(req StockLogListRequest) (*StockLogListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	query := s.db.Model(&models.StockLog{})
	if req.ItemId != "" {
		query = query.Where("item_id = ?", req.ItemId)
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.StockLog
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &StockLogListResponse{
		Items: items, Total: total,
		Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage,
	}, nil
}

// AdjustStockRequest 出入库操作请求
type AdjustStockRequest struct {
	ItemId   string `json:"itemId" binding:"required"`
	Type     string `json:"type" binding:"required"` // in | out
	Quantity int    `json:"quantity" binding:"required,min=1"`
	Operator string `json:"operator"`
	Note     string `json:"note"`
}

// AdjustStock 出入库（事务：更新库存 + 创建记录）
func (s *InventoryService) AdjustStock(req AdjustStockRequest, tenantId int64) (*models.StockLog, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Type != "in" && req.Type != "out" {
		return nil, errors.New("type must be 'in' or 'out'")
	}

	var log models.StockLog

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var item models.InventoryItem
		if err := tx.First(&item, "id = ? AND is_disabled = ?", req.ItemId, false).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("item not found")
			}
			return err
		}

		newStock := item.Stock
		if req.Type == "in" {
			newStock += req.Quantity
		} else {
			if item.Stock < req.Quantity {
				return fmt.Errorf("库存不足，当前库存 %d %s", item.Stock, item.Unit)
			}
			newStock -= req.Quantity
		}

		if err := tx.Model(&item).Update("stock", newStock).Error; err != nil {
			return err
		}

		var logCount int64
		tx.Model(&models.StockLog{}).Count(&logCount)

		log = models.StockLog{
			ID:       fmt.Sprintf("LOG-%06d", logCount+1),
			TenantId: tenantId,
			ItemId:   item.ID,
			ItemName: item.Name,
			Type:     req.Type,
			Quantity: req.Quantity,
			Unit:     item.Unit,
			Operator: req.Operator,
			Note:     req.Note,
		}
		return tx.Create(&log).Error
	})

	if err != nil {
		return nil, err
	}
	return &log, nil
}

// ─────────────── LabelTask ───────────────

type LabelTaskListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Status   string `form:"status"`
}

type LabelTaskListResponse struct {
	Items     []models.LabelTask `json:"items"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"pageSize"`
	TotalPage int                `json:"totalPage"`
}

func (s *InventoryService) ListLabelTasks(req LabelTaskListRequest) (*LabelTaskListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	query := s.db.Model(&models.LabelTask{})
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.LabelTask
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &LabelTaskListResponse{
		Items: items, Total: total,
		Page: req.Page, PageSize: req.PageSize, TotalPage: totalPage,
	}, nil
}

type LabelTaskCreateRequest struct {
	ItemId   string `json:"itemId" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

func (s *InventoryService) CreateLabelTask(req LabelTaskCreateRequest, tenantId, creatorId int64) (*models.LabelTask, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	var item models.InventoryItem
	if err := s.db.First(&item, "id = ? AND is_disabled = ?", req.ItemId, false).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("item not found")
		}
		return nil, err
	}

	var count int64
	s.db.Model(&models.LabelTask{}).Count(&count)

	task := models.LabelTask{
		ID:        fmt.Sprintf("LBL-%04d", count+1),
		TenantId:  tenantId,
		ItemId:    item.ID,
		ItemName:  item.Name,
		Spec:      item.Spec,
		Quantity:  req.Quantity,
		Status:    "pending",
		CreatorId: creatorId,
	}

	if err := s.db.Create(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *InventoryService) UpdateLabelTaskStatus(id, status string) error {
	if s.db == nil {
		return errors.New("database not available")
	}

	validStatuses := map[string]bool{"pending": true, "printing": true, "completed": true}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	result := s.db.Model(&models.LabelTask{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("label task not found")
	}
	return nil
}
