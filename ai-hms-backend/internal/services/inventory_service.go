package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

type InventoryService struct {
	db *gorm.DB
}

func NewInventoryService() *InventoryService {
	return &InventoryService{db: database.GetDB()}
}

// ─────────────── DTOs ───────────────

type InventoryListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Category string `form:"category"`
	Keyword  string `form:"keyword"`
}

type InventoryListResponse struct {
	Items     []InventoryItemView `json:"items"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"pageSize"`
	TotalPage int                 `json:"totalPage"`
}

type InventoryItemView struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Spec        string  `json:"spec"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit"`
	Supplier    string  `json:"supplier"`
	Stock       float64 `json:"stock"`
	MinStock    float64 `json:"minStock"`
	Price       float64 `json:"price"`
	Position    string  `json:"position"`
	Location    string  `json:"location"`
	IsDisabled  bool    `json:"isDisabled"`
	Alert       bool    `json:"alert"`
	LastUpdated string  `json:"lastUpdated"`
}

type StockLogRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Type     string `form:"type"`
}

type StockLogResponse struct {
	Items     []StockLogView `json:"items"`
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	PageSize  int            `json:"pageSize"`
	TotalPage int            `json:"totalPage"`
}

type StockLogView struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	ItemName    string  `json:"itemName"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	Operator    string  `json:"operator"`
	Note        string  `json:"note"`
	CreatedAt   string  `json:"createdAt"`
	OperateTime string  `json:"operateTime"`
}

// ─────────────── List Items ───────────────

type legacyStockItemRow struct {
	ID        int64     `gorm:"column:Id"`
	Name      string    `gorm:"column:Name"`
	Batch     string    `gorm:"column:Batch"`
	Num       float64   `gorm:"column:Num"`
	Price     float64   `gorm:"column:Price"`
	Unit      string    `gorm:"column:Unit"`
	Location  string    `gorm:"column:Location"`
	UpdatedAt time.Time `gorm:"column:LastModifyTime"`
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

	baseQuery := s.db.Table(`"Stock_Stock" AS s`).
		Select(`s."Id", COALESCE(c."Name", '') AS "Name", s."Batch", COALESCE(s."Num", 0) AS "Num",
			COALESCE(s."Price", c."Price", 0) AS "Price", COALESCE(c."Unit", '') AS "Unit",
			COALESCE(st."Name", st."Position", '') AS "Location", s."LastModifyTime"`).
		Joins(`LEFT JOIN "Stock_ChargeItem" AS c ON c."Id" = s."ChargeItemId" AND c."TenantId" = s."TenantId"`).
		Joins(`LEFT JOIN "Stock_Storage" AS st ON st."Id" = s."StorageId" AND st."TenantId" = s."TenantId"`).
		Where(`s."TenantId" = ?`, LegacyTenantID)

	countQuery := s.db.Table(`"Stock_Stock" AS s`).
		Joins(`LEFT JOIN "Stock_ChargeItem" AS c ON c."Id" = s."ChargeItemId" AND c."TenantId" = s."TenantId"`).
		Where(`s."TenantId" = ?`, LegacyTenantID)

	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		baseQuery = baseQuery.Where(`c."Name" LIKE ?`, like)
		countQuery = countQuery.Where(`c."Name" LIKE ?`, like)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []legacyStockItemRow
	offset := (req.Page - 1) * req.PageSize
	if err := baseQuery.Offset(offset).Limit(req.PageSize).Order(`c."Name"`).Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]InventoryItemView, 0, len(rows))
	for _, r := range rows {
		items = append(items, InventoryItemView{
			ID:          r.ID,
			Name:        r.Name,
			Spec:        r.Batch,
			Stock:       r.Num,
			Price:       r.Price,
			Unit:        r.Unit,
			Position:    r.Location,
			Location:    r.Location,
			LastUpdated: r.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &InventoryListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// ─────────────── Stock Logs ───────────────

type legacyInOutRow struct {
	ID          int64     `gorm:"column:Id"`
	BillNo      string    `gorm:"column:BillNo"`
	BillType    int64     `gorm:"column:BillType"`
	ItemName    string    `gorm:"column:ItemName"`
	Num         float64   `gorm:"column:Num"`
	Unit        string    `gorm:"column:Unit"`
	Note        string    `gorm:"column:Note"`
	HandlerName string    `gorm:"column:HandlerName"`
	CreateTime  time.Time `gorm:"column:CreateTime"`
}

func (s *InventoryService) ListLogs(req StockLogRequest) (*StockLogResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	query := s.db.Table(`"Stock_InOutStorage" AS io`).
		Select(`d."Id", io."BillNo", io."BillType", COALESCE(c."Name", io."BillNo", '') AS "ItemName",
			COALESCE(d."Num", 0) AS "Num", COALESCE(c."Unit", '') AS "Unit",
			COALESCE(io."Note", '') AS "Note", COALESCE(e."Name", '') AS "HandlerName", io."CreateTime"`).
		Joins(`JOIN "Stock_InOutStorageDetail" AS d ON d."InOutStorageId" = io."Id" AND d."TenantId" = io."TenantId"`).
		Joins(`LEFT JOIN "Stock_ChargeItem" AS c ON c."Id" = d."ChargeItemId" AND c."TenantId" = d."TenantId"`).
		Joins(`LEFT JOIN "Organ_Employee" AS e ON e."Id" = io."HandlerId"`).
		Where(`io."TenantId" = ?`, LegacyTenantID)

	if req.Type != "" {
		switch req.Type {
		case "in", "入库":
			query = query.Where(`io."BillType" = ?`, 10)
		case "out", "出库":
			query = query.Where(`io."BillType" = ?`, 20)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []legacyInOutRow
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order(`io."CreateTime" DESC`).Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]StockLogView, 0, len(rows))
	for _, r := range rows {
		logType := "in"
		if r.BillType == 20 {
			logType = "out"
		}
		createdAt := r.CreateTime.Format("2006-01-02 15:04")
		items = append(items, StockLogView{
			ID:          r.ID,
			Type:        logType,
			ItemName:    r.ItemName,
			Quantity:    r.Num,
			Unit:        r.Unit,
			Operator:    r.HandlerName,
			Note:        r.Note,
			CreatedAt:   createdAt,
			OperateTime: createdAt,
		})
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &StockLogResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

// ─────────────── CRUD (stub) ───────────────

type CreateInventoryItemRequest struct {
	Name string  `json:"name"`
	Spec string  `json:"spec"`
	Unit string  `json:"unit"`
	Num  float64 `json:"num"`
	Note string  `json:"note"`
}

func (s *InventoryService) CreateItem(req CreateInventoryItemRequest, tenantID, creatorID int64) (*InventoryItemView, error) {
	return nil, fmt.Errorf("库存新增需通过 HIS 同步或采购流程，暂不支持直接新增")
}

func (s *InventoryService) DeleteItem(id, tenantID int64) error {
	return fmt.Errorf("库存删除暂不支持")
}
