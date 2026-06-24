package services

import (
	"errors"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

var errDB = errors.New("database not available")

type HisPriceService struct {
	db *gorm.DB
}

func NewHisPriceService() *HisPriceService {
	return &HisPriceService{db: database.GetDB()}
}

type HisPriceSearchRequest struct {
	Keyword    string  `form:"keyword"`
	ItemClass  *string `form:"itemClass"`
	ActiveOnly *bool   `form:"activeOnly"`
	Page       int     `form:"page"`
	PageSize   int     `form:"pageSize"`
}

type HisPriceSearchResponse struct {
	Items     []models.HisPriceItem `json:"items"`
	Total     int64                 `json:"total"`
	Page      int                   `json:"page"`
	PageSize  int                   `json:"pageSize"`
	TotalPage int                   `json:"totalPage"`
}

func (s *HisPriceService) Search(req HisPriceSearchRequest) (*HisPriceSearchResponse, error) {
	if s.db == nil {
		return nil, errDB
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}
	if req.PageSize > 200 {
		req.PageSize = 200
	}

	query := s.db.Model(&models.HisPriceItem{}).Where("source_system = ?", "HIS_ORACLE")

	activeOnly := true
	if req.ActiveOnly != nil {
		activeOnly = *req.ActiveOnly
	}
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if req.ItemClass != nil && *req.ItemClass != "" {
		query = query.Where("item_class = ?", *req.ItemClass)
	}

	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where(
			"item_name LIKE ? OR input_code LIKE ? OR input_code_wb LIKE ? OR item_code LIKE ?",
			kw, kw, kw, kw,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.HisPriceItem
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("item_code").Offset(offset).Limit(req.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &HisPriceSearchResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

func (s *HisPriceService) FindByItemCode(itemCode string) (*models.HisPriceItem, error) {
	if s.db == nil {
		return nil, errDB
	}
	var item models.HisPriceItem
	err := s.db.Where("source_system = ? AND item_code = ?", "HIS_ORACLE", itemCode).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *HisPriceService) MatchByName(name string, itemClass *string, requireActive bool) ([]models.HisPriceItem, error) {
	if s.db == nil {
		return nil, errDB
	}
	if name == "" {
		return nil, nil
	}

	query := s.db.Model(&models.HisPriceItem{}).
		Where("source_system = ? AND UPPER(item_name) = UPPER(?)", "HIS_ORACLE", name)

	if itemClass != nil && *itemClass != "" {
		query = query.Where("item_class = ?", *itemClass)
	}
	if requireActive {
		query = query.Where("is_active = ?", true)
	}

	var items []models.HisPriceItem
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	if len(items) == 0 {
		kw := "%" + name + "%"
		query2 := s.db.Model(&models.HisPriceItem{}).
			Where("source_system = ? AND item_name LIKE ?", "HIS_ORACLE", kw)
		if itemClass != nil && *itemClass != "" {
			query2 = query2.Where("item_class = ?", *itemClass)
		}
		if requireActive {
			query2 = query2.Where("is_active = ?", true)
		}
		if err := query2.Limit(20).Find(&items).Error; err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (s *HisPriceService) LastSyncTime() (string, error) {
	if s.db == nil {
		return "", errDB
	}
	var item models.HisPriceItem
	err := s.db.Where("source_system = ?", "HIS_ORACLE").Order("synced_at DESC").First(&item).Error
	if err != nil {
		return "", err
	}
	return item.SyncedAt.Format("2006-01-02 15:04:05"), nil
}
