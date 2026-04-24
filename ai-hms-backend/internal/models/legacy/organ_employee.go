package legacy

// OrganEmployee 对应老库员工表 Organ_Employee。
type OrganEmployee struct {
	ID       int64  `gorm:"column:Id;primaryKey"`
	TenantID int64  `gorm:"column:TenantId"`
	UserID   int64  `gorm:"column:UserId"`
	Name     string `gorm:"column:Name"`
}

func (OrganEmployee) TableName() string {
	return "Organ_Employee"
}
