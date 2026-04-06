package models

import "time"

// InventoryItem 库存耗材/药品
type InventoryItem struct {
	ID         string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantId   int64     `gorm:"type:bigint;index" json:"tenantId"`
	Name       string    `gorm:"type:varchar(100);not null" json:"name"`
	Spec       string    `gorm:"type:varchar(100)" json:"spec"`         // 规格
	Category   string    `gorm:"type:varchar(50);index" json:"category"` // 类别
	Stock      int       `gorm:"default:0" json:"stock"`                // 当前库存
	Unit       string    `gorm:"type:varchar(20)" json:"unit"`           // 单位
	MinStock   int       `gorm:"default:0" json:"minStock"`             // 最低库存告警阈值
	MaxStock   int       `gorm:"default:0" json:"maxStock"`             // 最高库存
	Location   string    `gorm:"type:varchar(100)" json:"location"`     // 存放位置
	Supplier   string    `gorm:"type:varchar(100)" json:"supplier"`     // 供应商
	IsDisabled bool      `gorm:"default:false" json:"isDisabled"`
	CreatorId  int64     `gorm:"type:bigint" json:"creatorId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (InventoryItem) TableName() string { return "inventory_items" }

// StockLog 出入库记录
type StockLog struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantId  int64     `gorm:"type:bigint;index" json:"tenantId"`
	ItemId    string    `gorm:"type:varchar(36);index" json:"itemId"`
	ItemName  string    `gorm:"type:varchar(100)" json:"itemName"` // 冗余，便于展示
	Type      string    `gorm:"type:varchar(10)" json:"type"`      // in | out
	Quantity  int       `gorm:"not null" json:"quantity"`
	Unit      string    `gorm:"type:varchar(20)" json:"unit"`
	Operator  string    `gorm:"type:varchar(50)" json:"operator"`
	Note      string    `gorm:"type:varchar(255)" json:"note"`
	CreatedAt time.Time `json:"createdAt"`
}

func (StockLog) TableName() string { return "stock_logs" }

// LabelTask 标签打印任务
type LabelTask struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantId  int64     `gorm:"type:bigint;index" json:"tenantId"`
	ItemId    string    `gorm:"type:varchar(36);index" json:"itemId"`
	ItemName  string    `gorm:"type:varchar(100)" json:"itemName"`
	Spec      string    `gorm:"type:varchar(100)" json:"spec"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Status    string    `gorm:"type:varchar(20);default:pending" json:"status"` // pending | printing | completed
	CreatorId int64     `gorm:"type:bigint" json:"creatorId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (LabelTask) TableName() string { return "label_tasks" }
