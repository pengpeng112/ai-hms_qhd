package models

import "time"

type Permission struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string    `gorm:"size:100;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	Module      string    `gorm:"size:50;index" json:"module"`
	Action      string    `gorm:"size:50" json:"action"`
	Status      string    `gorm:"size:20;default:active" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Permission) TableName() string {
	return "permissions"
}

type RolePermission struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Role           string    `gorm:"size:50;index:idx_role_perm,unique;not null" json:"role"`
	PermissionCode string    `gorm:"size:100;index:idx_role_perm,unique;not null" json:"permissionCode"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
