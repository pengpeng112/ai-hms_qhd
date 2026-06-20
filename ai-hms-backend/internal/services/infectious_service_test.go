package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newInfTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.PatientInfectious{}, &models.SignRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db.Exec(`CREATE TABLE ["Register_OutCome"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "Type" TEXT, "Reason" TEXT, "OutComeTime" DATETIME, "Note" TEXT, "CreateTime" DATETIME, "LastModifyTime" DATETIME)`)
	return db
}

func TestInf_Screen_JudgePositive(t *testing.T) {
	db := newInfTestDB(t)
	s := &InfectiousService{db: db, tenantID: 3}
	now := time.Now()
	rec, err := s.Screen(1001, ScreenInput{
		ScreenDate: now, Source: "manual",
		Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemPositive}, {Item: "抗-HCV", Result: models.InfItemNegative}},
	})
	if err != nil {
		t.Fatalf("Screen: %v", err)
	}
	if rec.ResultOverall != models.InfectiousPositive {
		t.Fatalf("want positive, got %s", rec.ResultOverall)
	}
	if rec.PositiveMarkers != "HBsAg" {
		t.Fatalf("want positive_markers HBsAg, got %q", rec.PositiveMarkers)
	}
	wantDue := now.AddDate(0, 6, 0).Format("2006-01-02")
	if rec.NextDueDate == nil || rec.NextDueDate.Format("2006-01-02") != wantDue {
		t.Fatalf("next_due_date want %s got %v", wantDue, rec.NextDueDate)
	}
}

func TestInf_Screen_JudgeNegativeAndPending(t *testing.T) {
	db := newInfTestDB(t)
	s := &InfectiousService{db: db, tenantID: 3}
	neg, _ := s.Screen(1002, ScreenInput{ScreenDate: time.Now(), Source: "manual",
		Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemNegative}}})
	if neg.ResultOverall != models.InfectiousNegative {
		t.Fatalf("want negative, got %s", neg.ResultOverall)
	}
	pend, _ := s.Screen(1003, ScreenInput{ScreenDate: time.Now(), Source: "manual",
		Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemIndeterminate}}})
	if pend.ResultOverall != models.InfectiousPending {
		t.Fatalf("want pending, got %s", pend.ResultOverall)
	}
	if _, err := s.Screen(1004, ScreenInput{ScreenDate: time.Now(), Source: "manual", Items: nil}); err == nil {
		t.Fatalf("empty items should error")
	}
}

func TestInf_Gate_FourStates(t *testing.T) {
	db := newInfTestDB(t)
	s := &InfectiousService{db: db, tenantID: 3}
	now := time.Now()

	if g := s.CanScheduleRoutine(2001); g.State != GateRequireCZone {
		t.Fatalf("未查 want REQUIRE_C_ZONE, got %s", g.State)
	}
	s.Screen(2002, ScreenInput{ScreenDate: now, Source: "manual", Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemNegative}}})
	if g := s.CanScheduleRoutine(2002); g.State != GateAllowNormal {
		t.Fatalf("阽性 want ALLOW_NORMAL, got %s", g.State)
	}
	s.Screen(2003, ScreenInput{ScreenDate: now.AddDate(0, -7, 0), Source: "manual", Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemNegative}}})
	if g := s.CanScheduleRoutine(2003); g.State != GateRequireCZone {
		t.Fatalf("过期 want REQUIRE_C_ZONE, got %s", g.State)
	}
	s.Screen(2004, ScreenInput{ScreenDate: now, Source: "manual", Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemPositive}}})
	if g := s.CanScheduleRoutine(2004); g.State != GateFrozen {
		t.Fatalf("阳性未处置 want FROZEN, got %s", g.State)
	}
}

func TestInf_Gate_FailOpenOnMissingTable(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared_gate_missing"), &gorm.Config{})
	s := &InfectiousService{db: db, tenantID: 3} // 未建表
	if g := s.CanScheduleRoutine(9999); g.State != GateAllowNormal {
		t.Fatalf("表缺应 fail-open ALLOW_NORMAL, got %s", g.State)
	}
}
