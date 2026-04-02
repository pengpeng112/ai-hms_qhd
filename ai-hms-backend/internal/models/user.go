package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Username  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	RealName  string         `gorm:"type:varchar(50)" json:"realName"`
	Phone     string         `gorm:"type:varchar(20)" json:"phone"`
	Email     string         `gorm:"type:varchar(100)" json:"email"`
	Role      string         `gorm:"type:varchar(50);not null" json:"role"` // DOCTOR_CHIEF, NURSE_HEAD, etc.
	Status       string    `gorm:"type:varchar(20);default:active" json:"status"` // active, inactive
	DepartmentID *string   `gorm:"type:varchar(36)" json:"departmentId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Role 角色
const (
	RoleAdmin           = "ADMIN"
	RoleDoctorChief     = "DOCTOR_CHIEF"     // 主任医师
	RoleDoctorSupervisor = "DOCTOR_SUPERVISOR" // 主治医师
	RoleDoctorDuty      = "DOCTOR_DUTY"      // 值班医师
	RoleNurseHead       = "NURSE_HEAD"       // 护士长
	RoleNurseManager    = "NURSE_MANAGER"    // 护理组长
	RoleNurseResponsible = "NURSE_RESPONSIBLE" // 责任护士
	RoleEngineer        = "ENGINEER"         // 工程师
)

// UserStatus 用户状态
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)
