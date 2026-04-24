// DEPRECATED: legacy new-db model, will be rewritten to map legacy hemodialysis DB in Phase 1~5.
package models

import "time"

// IntegrationHDISSetting HDIS 对接配置（单租户）
type IntegrationHDISSetting struct {
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

	WebcmdURL  string `gorm:"type:varchar(255);not null" json:"webcmdUrl"`
	GraphqlURL string `gorm:"type:varchar(255);not null" json:"graphqlUrl"`
	AuthURL    string `gorm:"type:varchar(255);not null" json:"authUrl"`
	ClientID   string `gorm:"type:varchar(64);not null" json:"clientId"`

	ServiceUsername          string `gorm:"type:varchar(64);not null" json:"serviceUsername"`
	ServicePasswordEncrypted string `gorm:"type:text;not null" json:"-"`
	AccessTokenEncrypted     string `gorm:"type:text" json:"-"`
	TokenExpiresAt           *time.Time

	AutoRefreshEnabled bool   `gorm:"not null;default:true" json:"autoRefreshEnabled"`
	RefreshLeadSeconds int    `gorm:"not null;default:1800" json:"refreshLeadSeconds"`
	LastError          string `gorm:"type:text" json:"lastError"`
	UpdatedBy          string `gorm:"type:varchar(36)" json:"updatedBy"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (IntegrationHDISSetting) TableName() string {
	return "integration_hdis_settings"
}
