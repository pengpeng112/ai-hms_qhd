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
