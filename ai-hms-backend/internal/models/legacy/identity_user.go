package legacy

// IdentityUser 对应老库登录用户表 Identity_Users（标准 ASP.NET Core Identity，无 TenantId 列）。
type IdentityUser struct {
	ID           int64  `gorm:"column:Id;primaryKey"`
	UserName     string `gorm:"column:UserName"`
	PasswordHash string `gorm:"column:PasswordHash"`
}

func (IdentityUser) TableName() string {
	return "Identity_Users"
}
