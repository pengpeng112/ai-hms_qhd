package legacy

// OrganEmployee 对应老库员工表 Organ_Employee。
// 老库中 Organ_Employee.Id 与 Identity_Users.Id 共享同一 ID（无 UserId 列）。
type OrganEmployee struct {
	ID   int64  `gorm:"column:Id;primaryKey"`
	Name string `gorm:"column:Name"`
}

func (OrganEmployee) TableName() string {
	return "Organ_Employee"
}
