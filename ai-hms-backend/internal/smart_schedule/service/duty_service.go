package service

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/config"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
)

// singleStaffRole 单名岗位（当班医生一室一名）；护士岗位（主班/当班）可多名（契约04 更正）。
func singleStaffRole(role string) bool { return role == model.DutyRoleDoctor }

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
	if !config.ValidShiftCode(in.Shift) {
		return nil, errors.New("非法班次（须为已启用班次码，如 early/late/long/short）")
	}
	if in.StaffId <= 0 || in.WardId <= 0 {
		return nil, errors.New("staffId 与 wardId 必填")
	}
	if in.DutyDate.IsZero() {
		return nil, errors.New("dutyDate 必填")
	}
	day := dutyDateOnly(in.DutyDate)

	// 唯一键：当班医生按(室,日,班,角色)单名替换；护士岗按(室,日,班,角色,人)幂等——同班可多名护士。
	q := g.Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" = ? AND "DutyRole" = ? AND "Shift" = ?`,
		tenant, in.WardId, day, in.DutyRole, in.Shift)
	if !singleStaffRole(in.DutyRole) {
		q = q.Where(`"StaffId" = ?`, in.StaffId)
	}
	var existing model.StaffDuty
	err := q.First(&existing).Error
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

// ResolveDuties 解析某(室,日,角色)当班全部人员（支持多名护士）。覆盖优先：当日有覆盖则返回覆盖名单，否则月基线名单。
func ResolveDuties(g *gorm.DB, tenant, wardId int64, date time.Time, dutyRole string) ([]ResolvedDuty, error) {
	if err := validateDutyRole(dutyRole); err != nil {
		return nil, err
	}
	day := dutyDateOnly(date)
	out := []ResolvedDuty{}

	var ovs []model.StaffDutyOverride
	if err := g.Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" = ? AND "DutyRole" = ?`, tenant, wardId, day, dutyRole).
		Order(`"Id" ASC`).Find(&ovs).Error; err != nil {
		return nil, err
	}
	if len(ovs) > 0 {
		for _, ov := range ovs {
			out = append(out, ResolvedDuty{StaffId: ov.ActualStaffId, StaffName: ov.ActualStaffName, DutyRole: ov.DutyRole, WardId: ov.WardId, DutyDate: ov.DutyDate, Source: "override"})
		}
		return out, nil
	}

	var rows []model.StaffDuty
	if err := g.Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" = ? AND "DutyRole" = ?`, tenant, wardId, day, dutyRole).
		Order(`"Shift" ASC`).Order(`"Id" ASC`).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		out = append(out, ResolvedDuty{StaffId: r.StaffId, StaffName: r.StaffName, DutyRole: r.DutyRole, WardId: r.WardId, DutyDate: r.DutyDate, Source: "baseline"})
	}
	return out, nil
}

// NurseRatioResult 护患比校验结果（1 护士 : ratio 台机）。
type NurseRatioResult struct {
	WardId         int64  `json:"wardId"`
	Shift          string `json:"shift"`
	Ratio          int    `json:"ratio"`          // 1:ratio
	MachineCount   int    `json:"machineCount"`   // 在用机台数
	NurseCount     int    `json:"nurseCount"`     // 当班护士数（主班+当班）
	RequiredNurses int    `json:"requiredNurses"` // 按机台数所需护士数 = ceil(M/ratio)
	Status         string `json:"status"`         // ok / understaffed(缺岗) / overstaffed(超配)
}

// CheckNurseRatio 护患比校验：当班护士数 vs 机台数（1:ratio）。缺岗=护士不足、超配=护士过剩。
func CheckNurseRatio(g *gorm.DB, tenant, wardId int64, date time.Time, shift string, ratio int) (*NurseRatioResult, error) {
	if wardId <= 0 {
		return nil, errors.New("wardId 必填")
	}
	if ratio <= 0 {
		ratio = 6
	}
	day := dutyDateOnly(date)

	var machines int64
	if err := g.Model(&model.Machine{}).Where(`"TenantId" = ? AND "WardId" = ? AND "IsDisabled" = ?`, tenant, wardId, false).Count(&machines).Error; err != nil {
		return nil, err
	}

	// 当班护士 = 主班护士 + 当班护士（按 StaffId 去重；在 Go 内去重以跨方言安全）
	nq := g.Model(&model.StaffDuty{}).
		Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" = ? AND "DutyRole" IN ?`,
			tenant, wardId, day, []string{model.DutyRoleChargeNurse, model.DutyRoleDutyNurse})
	if shift != "" {
		nq = nq.Where(`"Shift" = ?`, shift)
	}
	var nurseRows []model.StaffDuty
	if err := nq.Find(&nurseRows).Error; err != nil {
		return nil, err
	}
	uniq := map[int64]struct{}{}
	for _, r := range nurseRows {
		uniq[r.StaffId] = struct{}{}
	}
	nurses := int64(len(uniq))

	required := int((machines + int64(ratio) - 1) / int64(ratio)) // ceil(M/ratio)
	status := "ok"
	if int(nurses) < required {
		status = "understaffed"
	} else if int(nurses) > required && required > 0 {
		status = "overstaffed"
	}
	return &NurseRatioResult{
		WardId: wardId, Shift: shift, Ratio: ratio,
		MachineCount: int(machines), NurseCount: int(nurses), RequiredNurses: required, Status: status,
	}, nil
}
