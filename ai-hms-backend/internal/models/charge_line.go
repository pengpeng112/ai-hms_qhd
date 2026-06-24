package models

import "time"

type ChargeLine struct {
	ID              string    `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID        int64     `gorm:"column:tenant_id;not null" json:"tenantId"`
	ChargeRecordID  string    `gorm:"column:charge_record_id;type:varchar(36);not null;index:idx_cl_record" json:"chargeRecordId"`
	Category        string    `gorm:"column:category;type:varchar(16);not null" json:"category"`
	ItemCode        *string   `gorm:"column:item_code;type:varchar(64)" json:"itemCode"`
	ItemName        string    `gorm:"column:item_name;type:varchar(128);not null" json:"itemName"`
	Spec            *string   `gorm:"column:spec;type:varchar(64)" json:"spec"`
	Unit            *string   `gorm:"column:unit;type:varchar(16)" json:"unit"`
	Quantity        *float64  `gorm:"column:quantity;type:decimal(10,2)" json:"quantity"`
	UnitPrice       *float64  `gorm:"column:unit_price;type:decimal(10,2)" json:"unitPrice"`
	Amount          *float64  `gorm:"column:amount;type:decimal(10,2)" json:"amount"`
	Billable        bool      `gorm:"column:billable;not null;default:true" json:"billable"`
	Source          string    `gorm:"column:source;type:varchar(8);not null;default:auto" json:"source"`
	ChargeItemID    *int64    `gorm:"column:charge_item_id" json:"chargeItemId"`
	HisPriceItemID  *string   `gorm:"column:his_price_item_id;type:varchar(36)" json:"hisPriceItemId"`
	HisItemCode     *string   `gorm:"column:his_item_code;type:varchar(20)" json:"hisItemCode"`
	HisItemClass    *string   `gorm:"column:his_item_class;type:varchar(1)" json:"hisItemClass"`
	HisItemName     *string   `gorm:"column:his_item_name;type:varchar(120)" json:"hisItemName"`
	PriceSource     *string   `gorm:"column:price_source;type:varchar(32)" json:"priceSource"`
	MatchedStatus   *string   `gorm:"column:matched_status;type:varchar(16)" json:"matchedStatus"`
	Note            *string   `gorm:"column:note;type:varchar(256)" json:"note"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
}

func (ChargeLine) TableName() string {
	return "charge_line"
}
