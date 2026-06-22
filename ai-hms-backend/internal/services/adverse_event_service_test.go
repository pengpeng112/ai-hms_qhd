package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newAeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.AdverseEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestAe_Register(t *testing.T) {
	db := newAeTestDB(t)
	s := &AdverseEventService{db: db, tenantID: 3}
	now := time.Now()

	rec, err := s.Register(AeRegisterInput{
		PatientID: 1001, EventType: "低血压", Severity: models.AESeveritySevere,
		OccurredAt: now, Description: "透析2h血压70/40", ReporterID: "9",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if rec.Status != models.AEStatusRegistered || rec.Severity != models.AESeveritySevere || rec.ReportedAt != nil {
		t.Fatalf("bad event: %+v", rec)
	}

	if _, err := s.Register(AeRegisterInput{PatientID: 1001, EventType: "x", Severity: "bogus", OccurredAt: now}); err == nil {
		t.Fatal("非法分级应拒")
	}
	if _, err := s.Register(AeRegisterInput{PatientID: 1001, EventType: "", Severity: models.AESeverityMild, OccurredAt: now}); err == nil {
		t.Fatal("空分类应拒")
	}
	if _, err := s.Register(AeRegisterInput{PatientID: 1001, EventType: "x", Severity: models.AESeverityMild}); err == nil {
		t.Fatal("空发生时间应拒")
	}
}

func TestAe_Report_6h(t *testing.T) {
	db := newAeTestDB(t)
	s := &AdverseEventService{db: db, tenantID: 3}

	within, _ := s.Register(AeRegisterInput{
		PatientID: 1002, EventType: "空气栓塞", Severity: models.AESeveritySevere,
		OccurredAt: time.Now(), Description: "test", ReporterID: "9",
	})

	sevenHoursAgo := time.Now().Add(-7 * time.Hour)
	overdue, _ := s.Register(AeRegisterInput{
		PatientID: 1003, EventType: "过敏反应", Severity: models.AESeveritySevere,
		OccurredAt: sevenHoursAgo, Description: "test", ReporterID: "9",
	})

	r1, err := s.Report(within.ID, AeReportInput{ReportedTo: []AeReportTarget{{Role: "护士长", UserID: "10"}}})
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if r1.Within6h == nil || !*r1.Within6h {
		t.Fatal("6h内上报应within_6h=true")
	}

	r2, err := s.Report(overdue.ID, AeReportInput{ReportedTo: []AeReportTarget{{Role: "科主任", UserID: "20"}}})
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if r2.Within6h == nil || *r2.Within6h {
		t.Fatal("超过6h上报应within_6h=false")
	}

	if _, err := s.Report(r1.ID, AeReportInput{ReportedTo: []AeReportTarget{{Role: "x", UserID: "1"}}}); err == nil {
		t.Fatal("重复上报应拒")
	}
	if _, err := s.Report(within.ID, AeReportInput{}); err == nil {
		t.Fatal("空上报对象应拒")
	}
}

func TestAe_StatusFlow(t *testing.T) {
	db := newAeTestDB(t)
	s := &AdverseEventService{db: db, tenantID: 3}
	now := time.Now()

	rec, _ := s.Register(AeRegisterInput{
		PatientID: 1004, EventType: "发热", Severity: models.AESeverityModerate,
		OccurredAt: now, Description: "test", ReporterID: "9",
	})
	if rec.Status != models.AEStatusRegistered {
		t.Fatal("初始状态应为registered")
	}

	if _, err := s.UpdateStatus(rec.ID, AeStatusInput{Status: models.AEStatusClosed}); err == nil {
		t.Fatal("registered不能直跳closed")
	}

	s.Report(rec.ID, AeReportInput{ReportedTo: []AeReportTarget{{Role: "护士长", UserID: "10"}}})

	r2, _ := s.UpdateStatus(rec.ID, AeStatusInput{Status: models.AEStatusAcknowledged})
	if r2.Status != models.AEStatusAcknowledged {
		t.Fatal("reported→acknowledged应成功")
	}

	r3, _ := s.UpdateStatus(rec.ID, AeStatusInput{Status: models.AEStatusProcessing})
	if r3.Status != models.AEStatusProcessing {
		t.Fatal("acknowledged→processing应成功")
	}

	r4, _ := s.UpdateStatus(rec.ID, AeStatusInput{Status: models.AEStatusClosed})
	if r4.Status != models.AEStatusClosed {
		t.Fatal("processing→closed应成功")
	}
}

func TestAe_Alerts(t *testing.T) {
	db := newAeTestDB(t)
	s := &AdverseEventService{db: db, tenantID: 3}
	now := time.Now()

	s.Register(AeRegisterInput{PatientID: 2001, EventType: "低血压", Severity: models.AESeveritySevere, OccurredAt: now, ReporterID: "9"})
	s.Register(AeRegisterInput{PatientID: 2002, EventType: "空气栓塞", Severity: models.AESeveritySevere, OccurredAt: time.Now().Add(-8 * time.Hour), ReporterID: "9"})
	s.Register(AeRegisterInput{PatientID: 2003, EventType: "发热", Severity: models.AESeverityMild, OccurredAt: now, ReporterID: "9"})

	a, err := s.Alerts()
	if err != nil {
		t.Fatalf("alerts: %v", err)
	}
	if len(a.SevereUnreported) != 1 {
		t.Fatalf("severeUnreported != 1: %d", len(a.SevereUnreported))
	}
	if len(a.SevereOverdue) != 1 {
		t.Fatalf("severeOverdue != 1: %d", len(a.SevereOverdue))
	}
	if len(a.Pending) != 3 {
		t.Fatalf("pending != 3: %d", len(a.Pending))
	}
}

func TestAe_List(t *testing.T) {
	db := newAeTestDB(t)
	s := &AdverseEventService{db: db, tenantID: 3}
	now := time.Now()

	s.Register(AeRegisterInput{PatientID: 3001, EventType: "低血压", Severity: models.AESeveritySevere, OccurredAt: now, ReporterID: "9"})
	s.Register(AeRegisterInput{PatientID: 3001, EventType: "发热", Severity: models.AESeverityMild, OccurredAt: now.Add(-1 * time.Hour), ReporterID: "9"})

	all, err := s.List("", "", nil)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2: %d", len(all))
	}

	pid := int64(3001)
	filtered, err := s.List("", "", &pid)
	if err != nil {
		t.Fatalf("list filter: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 for pid: %d", len(filtered))
	}

	severe, err := s.List(models.AESeveritySevere, "", nil)
	if err != nil {
		t.Fatalf("list severe: %v", err)
	}
	if len(severe) != 1 {
		t.Fatalf("expected 1 severe: %d", len(severe))
	}
}
