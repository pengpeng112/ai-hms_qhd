package service

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
)

type StaffDutyInput struct {
	StaffId   int64
	StaffName string
	DutyRole  string
	WardId    int64
	DutyDate  time.Time
	Shift     string
}

type ResolvedDuty struct {
	StaffId   int64     `json:"staffId"`
	StaffName string    `json:"staffName"`
	DutyRole  string    `json:"dutyRole"`
	WardId    int64     `json:"wardId"`
	DutyDate  time.Time `json:"dutyDate"`
	Source    string    `json:"source"`
}

func validateDutyRole(role string) error {
	switch role {
	case model.DutyRoleDoctor, model.DutyRoleChargeNurse, model.DutyRoleDutyNurse:
		return nil
	}
	return errors.New("非法 dutyRole（须为 当班医生/主班护士/当班护士）")
}

func dutyDateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func UpsertStaffDuty(g *gorm.DB, tenant int64, in StaffDutyInput, createdBy int64) (*model.StaffDuty, error) {
	if err := validateDutyRole(in.DutyRole); err != nil {
		return nil, err
	}
	if in.StaffId <= 0 || in.WardId <= 0 {
		return nil, errors.New("staffId 与 wardId 必填")
	}
	if in.DutyDate.IsZero() {
		return nil, errors.New("dutyDate 必填")
	}
	day := dutyDateOnly(in.DutyDate)

	var existing model.StaffDuty
	err := g.Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" = ? AND "DutyRole" = ? AND "Shift" = ?`,
		tenant, in.WardId, day, in.DutyRole, in.Shift).First(&existing).Error
	switch {
	case err == nil:
		existing.StaffId = in.StaffId
		existing.StaffName = in.StaffName
		if e := g.Model(&existing).Updates(map[string]interface{}{
			"StaffId":   in.StaffId,
			"StaffName": in.StaffName,
		}).Error; e != nil {
			return nil, e
		}
		return &existing, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		row := model.StaffDuty{
			BaseModel: model.BaseModel{TenantId: tenant, CreatorId: createdBy},
			StaffId:   in.StaffId,
			StaffName: in.StaffName,
			DutyRole:  in.DutyRole,
			WardId:    in.WardId,
			DutyDate:  day,
			Shift:     in.Shift,
		}
		if e := g.Create(&row).Error; e != nil {
			return nil, e
		}
		return &row, nil
	default:
		return nil, err
	}
}

func ListStaffDuty(g *gorm.DB, tenant, wardId int64, monthStart, monthEnd time.Time) ([]model.StaffDuty, error) {
	var rows []model.StaffDuty
	if err := g.Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" >= ? AND "DutyDate" < ?`,
		tenant, wardId, dutyDateOnly(monthStart), dutyDateOnly(monthEnd)).
		Order(`"DutyDate" ASC`).Order(`"Shift" ASC`).Order(`"DutyRole" ASC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func DeleteStaffDuty(g *gorm.DB, tenant, id int64) error {
	res := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, id).Delete(&model.StaffDuty{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("排班记录不存在")
	}
	return nil
}

func ResolveDuty(g *gorm.DB, tenant, wardId int64, date time.Time, dutyRole string) (*ResolvedDuty, error) {
	if err := validateDutyRole(dutyRole); err != nil {
		return nil, err
	}
	day := dutyDateOnly(date)

	var ov model.StaffDutyOverride
	ovErr := g.Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" = ? AND "DutyRole" = ?`, tenant, wardId, day, dutyRole).
		Order(`"Id" DESC`).First(&ov).Error
	if ovErr == nil {
		return &ResolvedDuty{
			StaffId: ov.ActualStaffId, StaffName: ov.ActualStaffName,
			DutyRole: ov.DutyRole, WardId: ov.WardId, DutyDate: ov.DutyDate, Source: "override",
		}, nil
	}
	if ovErr != nil && !errors.Is(ovErr, gorm.ErrRecordNotFound) {
		return nil, ovErr
	}

	var row model.StaffDuty
	err := g.Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" = ? AND "DutyRole" = ?`, tenant, wardId, day, dutyRole).
		Order(`"Shift" ASC`).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ResolvedDuty{
		StaffId: row.StaffId, StaffName: row.StaffName,
		DutyRole: row.DutyRole, WardId: row.WardId, DutyDate: row.DutyDate, Source: "baseline",
	}, nil
}

type OverrideInput struct {
	DutyDate        time.Time
	WardId          int64
	DutyRole        string
	OriginalStaffId int64
	ActualStaffId   int64
	ActualStaffName string
	Reason          string
}

func CreateOverride(g *gorm.DB, tenant int64, in OverrideInput, changedBy int64) (*model.StaffDutyOverride, error) {
	if err := validateDutyRole(in.DutyRole); err != nil {
		return nil, err
	}
	if in.ActualStaffId <= 0 || in.WardId <= 0 || in.DutyDate.IsZero() {
		return nil, errors.New("actualStaffId/wardId/dutyDate 必填")
	}
	row := model.StaffDutyOverride{
		BaseModel:       model.BaseModel{TenantId: tenant, CreatorId: changedBy},
		DutyDate:        dutyDateOnly(in.DutyDate),
		WardId:          in.WardId,
		DutyRole:        in.DutyRole,
		OriginalStaffId: in.OriginalStaffId,
		ActualStaffId:   in.ActualStaffId,
		ActualStaffName: in.ActualStaffName,
		Reason:          in.Reason,
		ChangedBy:       changedBy,
	}
	if err := g.Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func ResolveMyDuties(g *gorm.DB, tenant, userId int64, date time.Time) ([]ResolvedDuty, error) {
	day := dutyDateOnly(date)
	out := []ResolvedDuty{}
	seen := map[string]bool{}

	var ovs []model.StaffDutyOverride
	if err := g.Where(`"TenantId" = ? AND "DutyDate" = ? AND "ActualStaffId" = ?`, tenant, day, userId).Find(&ovs).Error; err != nil {
		return nil, err
	}
	for _, o := range ovs {
		key := keyOf(o.WardId, o.DutyRole)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ResolvedDuty{StaffId: userId, StaffName: o.ActualStaffName, DutyRole: o.DutyRole, WardId: o.WardId, DutyDate: day, Source: "override"})
	}

	var bases []model.StaffDuty
	if err := g.Where(`"TenantId" = ? AND "DutyDate" = ? AND "StaffId" = ?`, tenant, day, userId).Find(&bases).Error; err != nil {
		return nil, err
	}
	for _, b := range bases {
		key := keyOf(b.WardId, b.DutyRole)
		if seen[key] {
			continue
		}
		var cnt int64
		g.Model(&model.StaffDutyOverride{}).Where(`"TenantId" = ? AND "DutyDate" = ? AND "WardId" = ? AND "DutyRole" = ?`, tenant, day, b.WardId, b.DutyRole).Count(&cnt)
		if cnt > 0 {
			continue
		}
		seen[key] = true
		out = append(out, ResolvedDuty{StaffId: userId, StaffName: b.StaffName, DutyRole: b.DutyRole, WardId: b.WardId, DutyDate: day, Source: "baseline"})
	}
	return out, nil
}

func keyOf(wardId int64, role string) string { return strconv.FormatInt(wardId, 10) + "|" + role }

func CheckIn(g *gorm.DB, tenant, userId, wardId, shiftId, operatorType, dutyType int64, note string) (*model.CheckIn, error) {
	if userId <= 0 || wardId <= 0 {
		return nil, errors.New("userId/wardId 必填")
	}
	now := time.Now()
	row := model.CheckIn{
		TenantId: tenant, ShiftId: shiftId, WardId: wardId,
		ClockInTime: now, OperatorType: operatorType, Type: dutyType,
		Note: note, OperatorId: userId, CreatorId: userId,
		CreateTime: now, LastModifyTime: now,
	}
	if err := g.Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func IsCheckedIn(g *gorm.DB, tenant, userId int64, date time.Time) (bool, error) {
	day := dutyDateOnly(date)
	next := day.AddDate(0, 0, 1)
	var cnt int64
	if err := g.Model(&model.CheckIn{}).Where(`"TenantId" = ? AND "OperatorId" = ? AND "ClockInTime" >= ? AND "ClockInTime" < ?`, tenant, userId, day, next).Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}
