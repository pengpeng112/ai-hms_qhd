package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
)

func newDutyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	g, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := g.AutoMigrate(&model.StaffDuty{}, &model.StaffDutyOverride{}, &model.CheckIn{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return g
}

func TestUpsertStaffDutyAndResolve(t *testing.T) {
	g := newDutyTestDB(t)
	const tenant int64 = 1
	day := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	in := StaffDutyInput{StaffId: 9001, StaffName: "王医生", DutyRole: model.DutyRoleDoctor, WardId: 10, DutyDate: day, Shift: "early"}
	if _, err := UpsertStaffDuty(g, tenant, in, 1); err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	r, err := ResolveDuty(g, tenant, 10, day, model.DutyRoleDoctor)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if r == nil || r.StaffId != 9001 {
		t.Fatalf("expected 王医生(9001) as 当班医生, got %+v", r)
	}

	r2, err := ResolveDuty(g, tenant, 10, day, model.DutyRoleDutyNurse)
	if err != nil {
		t.Fatalf("resolve duty nurse: %v", err)
	}
	if r2 != nil {
		t.Fatalf("expected nil for unassigned duty nurse, got %+v", r2)
	}
}

func TestUpsertStaffDutyIdempotentOverwrite(t *testing.T) {
	g := newDutyTestDB(t)
	const tenant int64 = 1
	day := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	// 首次排
	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 1, StaffName: "A", DutyRole: model.DutyRoleDoctor, WardId: 1, DutyDate: day, Shift: "early"}, 1)
	// 覆盖
	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 2, StaffName: "B", DutyRole: model.DutyRoleDoctor, WardId: 1, DutyDate: day, Shift: "early"}, 1)

	r, _ := ResolveDuty(g, tenant, 1, day, model.DutyRoleDoctor)
	if r.StaffId != 2 {
		t.Fatalf("expected overwritten to B(2), got %d", r.StaffId)
	}
	count := int64(0)
	g.Model(&model.StaffDuty{}).Where(`"TenantId" = ? AND "WardId" = ? AND "DutyDate" = ?`, tenant, 1, day).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 row after idempotent overwrite, got %d", count)
	}
}

func TestResolveDutyOverrideWins(t *testing.T) {
	g := newDutyTestDB(t)
	const tenant int64 = 1
	day := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 1, StaffName: "A", DutyRole: model.DutyRoleDoctor, WardId: 1, DutyDate: day, Shift: "early"}, 1)
	CreateOverride(g, tenant, OverrideInput{DutyDate: day, WardId: 1, DutyRole: model.DutyRoleDoctor, ActualStaffId: 99, ActualStaffName: "顶班"}, 1)

	r, _ := ResolveDuty(g, tenant, 1, day, model.DutyRoleDoctor)
	if r.StaffId != 99 || r.Source != "override" {
		t.Fatalf("override should win, got %+v", r)
	}
}

func TestResolveMyDutiesAndCheckIn(t *testing.T) {
	g := newDutyTestDB(t)
	const tenant int64 = 1
	day := dutyDateOnly(time.Now())

	// 排 9001 为 A室医生
	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 9001, StaffName: "张医生", DutyRole: model.DutyRoleDoctor, WardId: 10, DutyDate: day, Shift: "early"}, 1)
	// 9002 不在排班里
	mine, _ := ResolveMyDuties(g, tenant, 9001, day)
	if len(mine) != 1 || mine[0].WardId != 10 {
		t.Fatalf("9001 should have 1 ward duty, got %+v", mine)
	}
	if m2, _ := ResolveMyDuties(g, tenant, 9002, day); len(m2) != 0 {
		t.Fatalf("9002 should have no duties")
	}
	// 接班
	if _, err := CheckIn(g, tenant, 9001, 10, 0, 10, 10, ""); err != nil {
		t.Fatalf("check-in failed: %v", err)
	}
	ok, _ := IsCheckedIn(g, tenant, 9001, day)
	if !ok {
		t.Fatalf("expected checked-in true")
	}
	ok2, _ := IsCheckedIn(g, tenant, 9002, day)
	if ok2 {
		t.Fatalf("9002 should not be checked in")
	}
	// 覆盖 9003 → A室医生, 9001 不再出现在我的职责
	CreateOverride(g, tenant, OverrideInput{DutyDate: day, WardId: 10, DutyRole: model.DutyRoleDoctor, OriginalStaffId: 9001, ActualStaffId: 9003, ActualStaffName: "替班"}, 1)
	if m3, _ := ResolveMyDuties(g, tenant, 9003, day); len(m3) != 1 || m3[0].Source != "override" {
		t.Fatalf("9003 should have 1 override duty, got %+v", m3)
	}
}

func TestUpsertStaffDutyValidation(t *testing.T) {
	g := newDutyTestDB(t)
	day := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	if _, err := UpsertStaffDuty(g, 1, StaffDutyInput{DutyRole: "未知", StaffId: 1, WardId: 1, DutyDate: day, Shift: "early"}, 1); err == nil {
		t.Fatal("expected validation error for unknown duty role")
	}
	if _, err := UpsertStaffDuty(g, 1, StaffDutyInput{DutyRole: model.DutyRoleDoctor, WardId: 1, DutyDate: day, Shift: "early"}, 1); err == nil {
		t.Fatal("expected validation error for zero staffId")
	}
	if _, err := UpsertStaffDuty(g, 1, StaffDutyInput{DutyRole: model.DutyRoleDoctor, StaffId: 1, WardId: 1, DutyDate: day, Shift: "无此班"}, 1); err == nil {
		t.Fatal("expected validation error for invalid shift code")
	}
}

// 护士岗可多名（契约04 更正）：同(室,日,班,角色)多名护士各占一行；医生仍单名替换。
func TestMultiNursePerShift(t *testing.T) {
	g := newDutyTestDB(t)
	const tenant int64 = 1
	day := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	for _, n := range []struct {
		id   int64
		role string
	}{{201, model.DutyRoleChargeNurse}, {202, model.DutyRoleChargeNurse}, {301, model.DutyRoleDutyNurse}} {
		if _, err := UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: n.id, StaffName: "N", DutyRole: n.role, WardId: 5, DutyDate: day, Shift: "early"}, 1); err != nil {
			t.Fatalf("upsert nurse failed: %v", err)
		}
	}
	charge, _ := ResolveDuties(g, tenant, 5, day, model.DutyRoleChargeNurse)
	if len(charge) != 2 {
		t.Fatalf("主班护士应 2 名, got %d", len(charge))
	}

	// 同一护士重复排=幂等
	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 201, StaffName: "N-改名", DutyRole: model.DutyRoleChargeNurse, WardId: 5, DutyDate: day, Shift: "early"}, 1)
	if charge2, _ := ResolveDuties(g, tenant, 5, day, model.DutyRoleChargeNurse); len(charge2) != 2 {
		t.Fatalf("重复排同一护士应幂等仍 2 名, got %d", len(charge2))
	}

	// 医生单名替换
	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 1, StaffName: "甲", DutyRole: model.DutyRoleDoctor, WardId: 5, DutyDate: day, Shift: "early"}, 1)
	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 2, StaffName: "乙", DutyRole: model.DutyRoleDoctor, WardId: 5, DutyDate: day, Shift: "early"}, 1)
	docs, _ := ResolveDuties(g, tenant, 5, day, model.DutyRoleDoctor)
	if len(docs) != 1 || docs[0].StaffId != 2 {
		t.Fatalf("当班医生应单名且替换为乙(2), got %+v", docs)
	}
}

func TestCheckNurseRatio(t *testing.T) {
	g := newDutyTestDB(t)
	if err := g.AutoMigrate(&model.Machine{}); err != nil {
		t.Fatalf("migrate machine: %v", err)
	}
	const tenant int64 = 1
	day := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 13; i++ {
		g.Create(&model.Machine{BaseModel: model.BaseModel{TenantId: tenant}, WardId: 5, Name: "M"})
	}
	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 301, StaffName: "N", DutyRole: model.DutyRoleDutyNurse, WardId: 5, DutyDate: day, Shift: "early"}, 1)
	UpsertStaffDuty(g, tenant, StaffDutyInput{StaffId: 302, StaffName: "N", DutyRole: model.DutyRoleDutyNurse, WardId: 5, DutyDate: day, Shift: "early"}, 1)

	res, err := CheckNurseRatio(g, tenant, 5, day, "early", 6)
	if err != nil {
		t.Fatalf("ratio check failed: %v", err)
	}
	if res.MachineCount != 13 || res.RequiredNurses != 3 || res.NurseCount != 2 || res.Status != "understaffed" {
		t.Fatalf("期望 13机/需3/有2/缺岗, got %+v", res)
	}
}
