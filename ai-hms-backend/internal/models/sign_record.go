package models

import "time"

// 统一电子签名留痕（契约05 #5 / 契约02 待签线）。
// 处方 / 方案 / 小结 三类待签共用此表，避免改老库加字段。
// v1 = 轻量留痕（谁 / 何时 / 签了什么）；法律级签名（CA/图像）留 SignatureBlob 接口后做。
//
// ⚠️ 本项目 AutoMigrate 永久禁用：此独立新表应由部署阶段按 docs/sql/deploy_new_tables.sql 建表后方可运行。
type SignRecord struct {
	ID            string    `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID      int64     `gorm:"column:tenant_id;index;not null" json:"tenantId"`
	TargetType    string    `gorm:"column:target_type;type:varchar(16);index;not null" json:"targetType"` // prescription / plan / summary
	TargetID      string    `gorm:"column:target_id;type:varchar(64);index;not null" json:"targetId"`
	SignerID      string    `gorm:"column:signer_id;type:varchar(64);not null" json:"signerId"`
	SignerName    string    `gorm:"column:signer_name;type:varchar(64)" json:"signerName"`
	SignTime      time.Time `gorm:"column:sign_time;not null" json:"signTime"`
	SignatureBlob string    `gorm:"column:signature_blob;type:text" json:"signatureBlob,omitempty"` // 法律级签名，v1 留空
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName 新表，snake_case（契约03/05）。
func (SignRecord) TableName() string { return "sign_record" }

// 待签对象类型（契约02 待签线，三类共用）。
const (
	SignTargetPrescription          = "prescription"
	SignTargetPlan                  = "plan"
	SignTargetSummary               = "summary"
	SignTargetInfectiousDisposition = "infectious_disposition"
)
