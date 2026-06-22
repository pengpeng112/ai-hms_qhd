package models

import "time"

// ConsentRecord 知情同意记录（C2）。同意书开具→患者/家属签署（复用 sign_record 留痕）→归档→到期复签。
type ConsentRecord struct {
	ID              string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID        int64      `gorm:"column:tenant_id;index:idx_cr_tenant_patient;not null" json:"tenantId"`
	PatientID       string     `gorm:"column:patient_id;type:varchar(64);index:idx_cr_tenant_patient" json:"patientId"`
	ConsentType     string     `gorm:"column:consent_type;type:varchar(16);index:idx_cr_type" json:"consentType"` // dialysis/cvc/avf/transfusion/drug/self_pay
	TemplateVersion string     `gorm:"column:template_version;type:varchar(32)" json:"templateVersion"`
	SignedBy        string     `gorm:"column:signed_by;type:varchar(64)" json:"signedBy"`          // 患者/家属（含关系）
	SignRecordID    string     `gorm:"column:sign_record_id;type:varchar(64)" json:"signRecordId"` // 关联 sign_record
	IssuedBy        string     `gorm:"column:issued_by;type:varchar(64)" json:"issuedBy"`          // 开具医生
	SignedAt        *time.Time `gorm:"column:signed_at" json:"signedAt"`
	ExpiresAt       *time.Time `gorm:"column:expires_at" json:"expiresAt"`
	Status          string     `gorm:"column:status;type:varchar(12);index:idx_cr_status" json:"status"` // pending/signed/expired/revoked
	DocRef          string     `gorm:"column:doc_ref;type:varchar(256)" json:"docRef"`
	Note            string     `gorm:"column:note;type:varchar(256)" json:"note"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (ConsentRecord) TableName() string { return "consent_record" }

// 同意书状态
const (
	ConsentStatusPending = "pending" // 待签
	ConsentStatusSigned  = "signed"  // 已签
	ConsentStatusExpired = "expired" // 已过期
	ConsentStatusRevoked = "revoked" // 已撤销
)
