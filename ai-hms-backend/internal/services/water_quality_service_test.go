package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newWqTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.WaterQuality{}, &models.SignRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newWqSvc(t *testing.T, db *gorm.DB) *WaterQualityService {
	th, err := config.LoadWaterQualityThresholds()
	if err != nil {
		t.Fatalf("load thresholds: %v", err)
	}
	return &WaterQualityService{db: db, tenantID: 3, thresholds: th}
}

func TestWq_Record_JudgeAndDue(t *testing.T) {
	db := newWqTestDB(t)
	s := newWqSvc(t, db)
	now := time.Now()
	rec, err := s.Record(RecordInput{TestDate: now, TestType: "endotoxin", SamplePoint: "ro_outlet", Value: 0.30, Unit: "EU/mL"})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if rec.Result != models.WQResultFail {
		t.Fatalf("want fail, got %s", rec.Result)
	}
	wantDue := now.AddDate(0, 0, 90).Format("2006-01-02")
	if rec.NextDueDate == nil || rec.NextDueDate.Format("2006-01-02") != wantDue {
		t.Fatalf("next_due want %s got %v", wantDue, rec.NextDueDate)
	}
	ok, _ := s.Record(RecordInput{TestDate: now, TestType: "bacteria", SamplePoint: "ro_outlet", Value: 80, Unit: "CFU/mL"})
	if ok.Result != models.WQResultPass {
		t.Fatalf("want pass, got %s", ok.Result)
	}
	if _, err := s.Record(RecordInput{TestDate: now, TestType: "free_chlorine", Value: 0.05}); err == nil {
		t.Fatalf("disabled item should be rejected")
	}
	if _, err := s.Record(RecordInput{TestDate: now, TestType: "bogus", Value: 1}); err == nil {
		t.Fatalf("unknown item should be rejected")
	}
}

func TestWq_ConductivityDaily(t *testing.T) {
	db := newWqTestDB(t)
	s := newWqSvc(t, db)
	if err := db.Exec(`CREATE TABLE "Treatment_Treatment" ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "StartTime" DATETIME, "CreateTime" DATETIME)`).Error; err != nil {
		t.Fatalf("create Treatment_Treatment: %v", err)
	}
	if err := db.Exec(`CREATE TABLE "Treatment_DuringParam" ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "TreatmentId" INTEGER, "Conductivity" REAL)`).Error; err != nil {
		t.Fatalf("create Treatment_DuringParam: %v", err)
	}
	day := time.Now().Format("2006-01-02")
	db.Exec(`INSERT INTO "Treatment_Treatment" ("Id","TenantId","StartTime") VALUES (1,3,?)`, day+" 08:00:00")
	db.Exec(`INSERT INTO "Treatment_DuringParam" ("Id","TenantId","TreatmentId","Conductivity") VALUES (1,3,1,16.2),(2,3,1,16.0)`)
	pts, err := s.ConductivityDaily(7)
	if err != nil {
		t.Fatalf("conductivity: %v", err)
	}
	if len(pts) != 1 {
		t.Fatalf("want 1 day point, got %d", len(pts))
	}
	if pts[0].InRange {
		t.Fatalf("16.1 应超出 13–15 范围，InRange 应为 false")
	}
}

func TestWq_Handle_DoubleConfirm(t *testing.T) {
	db := newWqTestDB(t)
	s := newWqSvc(t, db)
	now := time.Now()
	rec, _ := s.Record(RecordInput{TestDate: now, TestType: "endotoxin", SamplePoint: "ro_outlet", Value: 0.30, Unit: "EU/mL"})

	if _, err := s.Handle(rec.ID, HandleInput{Role: "engineer", SignerID: "e1", SignerName: "王工", Action: "更换内毒素滤器"}); err != nil {
		t.Fatalf("engineer: %v", err)
	}
	var mid models.WaterQuality
	db.First(&mid, "id = ?", rec.ID)
	if mid.HandledAt != nil {
		t.Fatalf("单签不应已处置")
	}
	if _, err := s.Handle(rec.ID, HandleInput{Role: "head_nurse", SignerID: "n1", SignerName: "李护士长", Action: "已复检合格"}); err != nil {
		t.Fatalf("headnurse: %v", err)
	}
	var done models.WaterQuality
	db.First(&done, "id = ?", rec.ID)
	if done.HandledAt == nil {
		t.Fatalf("双签齐应已处置")
	}
	ok, _ := s.Record(RecordInput{TestDate: now, TestType: "bacteria", SamplePoint: "ro_outlet", Value: 10, Unit: "CFU/mL"})
	if _, err := s.Handle(ok.ID, HandleInput{Role: "engineer", SignerID: "e1", SignerName: "王工", Action: "x"}); err == nil {
		t.Fatalf("非超标应拒处置")
	}
}

func TestWq_AlertsAndList(t *testing.T) {
	db := newWqTestDB(t)
	db.Exec("DELETE FROM water_quality")
	s := newWqSvc(t, db)
	now := time.Now()
	s.Record(RecordInput{TestDate: now, TestType: "endotoxin", SamplePoint: "ro_outlet", Value: 0.30, Unit: "EU/mL"})
	old := now.AddDate(0, 0, -100)
	s.Record(RecordInput{TestDate: old, TestType: "bacteria", SamplePoint: "ro_outlet", Value: 10, Unit: "CFU/mL"})
	a, err := s.Alerts()
	if err != nil {
		t.Fatalf("alerts: %v", err)
	}
	if len(a.Exceed) != 1 {
		t.Fatalf("超标卡应 1, got %d", len(a.Exceed))
	}
	if len(a.Due) != 1 {
		t.Fatalf("到期卡应 1, got %d", len(a.Due))
	}
	all, _ := s.List(WqListFilter{})
	if len(all) != 2 {
		t.Fatalf("列表应 2, got %d", len(all))
	}
}
