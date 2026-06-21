package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newVascTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.VascularAccessEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS "Register_VascularAccess" (
		"Id" BIGINT PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "AccessType" TEXT,
		"OperationTime" DATETIME, "IsDefault" BOOLEAN DEFAULT 0, "IsDisabled" BOOLEAN DEFAULT 0)`)
	return db
}

func TestVasc_RecordEvent(t *testing.T) {
	db := newVascTestDB(t)
	db.Exec(`INSERT INTO "Register_VascularAccess" ("Id","TenantId","PatientId","AccessType","OperationTime","IsDefault") VALUES (501,3,1001,'AVF',?,1)`, time.Now().AddDate(0,0,-40))
	s := &VascularAccessMonitor{db: db, tenantID: 3}
	rec, err := s.RecordEvent(1001, 501, VascEventInput{EventType: models.VAEMaturation, EventDate: time.Now(), Detail: `{"bloodFlow":600,"usable":true}`, OperatorID: "9"})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if rec.EventType != "maturation" || rec.AccessID != 501 {
		t.Fatalf("bad event: %+v", rec)
	}
	if _, err := s.RecordEvent(1001, 501, VascEventInput{EventType: "bogus", EventDate: time.Now()}); err == nil {
		t.Fatalf("非法类型应拒")
	}
	if _, err := s.RecordEvent(9999, 501, VascEventInput{EventType: models.VAEMaturation, EventDate: time.Now()}); err == nil {
		t.Fatalf("越权 accessId 应拒")
	}
	if _, err := s.RecordEvent(1001, 501, VascEventInput{EventType: models.VAEMaturation, EventDate: time.Now(), Detail: "{bad"}); err == nil {
		t.Fatalf("非法 detail 应拒")
	}
}

func TestVasc_Timeline(t *testing.T) {
	db := newVascTestDB(t)
	surgery := time.Now().AddDate(0, 0, -40)
	db.Exec(`INSERT INTO "Register_VascularAccess" ("Id","TenantId","PatientId","AccessType","OperationTime","IsDefault") VALUES (601,3,2001,'AVF',?,1)`, surgery)
	s := &VascularAccessMonitor{db: db, tenantID: 3}
	s.RecordEvent(2001, 601, VascEventInput{EventType: models.VAEMaturation, EventDate: time.Now().AddDate(0, 0, -10), Detail: `{"usable":true}`})
	tl, err := s.Timeline(2001)
	if err != nil {
		t.Fatalf("timeline: %v", err)
	}
	var hasEstablish, hasMaturation bool
	for _, e := range tl {
		if e.EventType == models.VAEEstablish {
			hasEstablish = true
		}
		if e.EventType == models.VAEMaturation {
			hasMaturation = true
		}
	}
	if !hasEstablish || !hasMaturation {
		t.Fatalf("时间线缺 establish/maturation: %+v", tl)
	}
	if len(tl) < 2 || tl[0].EventType != "maturation" {
		t.Fatalf("时间线首条应为最近的 maturation(10天前), got %+v", tl)
	}
}

func TestVasc_Reminders(t *testing.T) {
	db := newVascTestDB(t)
	s := &VascularAccessMonitor{db: db, tenantID: 3}
	now := time.Now()
	db.Exec(`INSERT INTO "Register_VascularAccess" ("Id","TenantId","PatientId","AccessType","OperationTime","IsDefault") VALUES (701,3,3001,'AVF',?,1)`, now.AddDate(0,0,-35))
	if rs := s.PatientReminders(3001); !hasKind(rs, "maturation_due") {
		t.Fatalf("AVF 5周无评估应 maturation_due, got %+v", rs)
	}
	db.Exec(`INSERT INTO "Register_VascularAccess" ("Id","TenantId","PatientId","AccessType","OperationTime","IsDefault") VALUES (702,3,3002,'CVC-NCC',?,1)`, now.AddDate(0,0,-35))
	if rs := s.PatientReminders(3002); !hasKind(rs, "cvc_over_limit") {
		t.Fatalf("NCC 5周应 cvc_over_limit, got %+v", rs)
	}
	db.Exec(`INSERT INTO "Register_VascularAccess" ("Id","TenantId","PatientId","AccessType","OperationTime","IsDefault") VALUES (703,3,3003,'AVF',?,1)`, now.AddDate(0,0,-200))
	s.RecordEvent(3003, 703, VascEventInput{EventType: models.VAEPhysicalCheck, EventDate: now, Detail: `{"abnormal":true}`})
	if rs := s.PatientReminders(3003); !hasKind(rs, "physical_abnormal") {
		t.Fatalf("物理异常应 physical_abnormal, got %+v", rs)
	}
}

func TestVasc_Alerts(t *testing.T) {
	db := newVascTestDB(t)
	s := &VascularAccessMonitor{db: db, tenantID: 3}
	now := time.Now()
	db.Exec(`INSERT INTO "Register_VascularAccess" ("Id","TenantId","PatientId","AccessType","OperationTime","IsDefault") VALUES (801,3,4001,'CVC-NCC',?,1)`, now.AddDate(0,0,-35))
	a, err := s.Alerts()
	if err != nil {
		t.Fatalf("alerts: %v", err)
	}
	if len(a) == 0 {
		t.Fatalf("应有 NCC 超时限提醒")
	}
	found := false
	for _, r := range a {
		if r.PatientID == 4001 && r.Kind == "cvc_over_limit" {
			found = true
		}
	}
	if !found {
		t.Fatalf("应含 4001 cvc_over_limit, got %+v", a)
	}
}

func TestVasc_PeriodicDueNoFalsePositive(t *testing.T) {
	db := newVascTestDB(t)
	s := &VascularAccessMonitor{db: db, tenantID: 3}
	now := time.Now()
	db.Exec(`INSERT INTO "Register_VascularAccess" ("Id","TenantId","PatientId","AccessType","OperationTime","IsDefault") VALUES (901,3,5001,'AVF',?,1)`, now.AddDate(0,0,-200))
	s.RecordEvent(5001, 901, VascEventInput{EventType: models.VAEMaturation, EventDate: now.AddDate(0,0,-100), Detail: `{"usable":true}`})
	s.RecordEvent(5001, 901, VascEventInput{EventType: models.VAEPhysicalCheck, EventDate: now.AddDate(0,0,-1), Detail: `{"abnormal":false}`})
	rs := s.PatientReminders(5001)
	if hasKind(rs, "periodic_due") {
		t.Fatalf("maturation 100天前但physical_check昨天，不应 periodic_due, got %+v", rs)
	}
}

func TestVasc_PeriodicDueTrigger(t *testing.T) {
	db := newVascTestDB(t)
	s := &VascularAccessMonitor{db: db, tenantID: 3}
	now := time.Now()
	db.Exec(`INSERT INTO "Register_VascularAccess" ("Id","TenantId","PatientId","AccessType","OperationTime","IsDefault") VALUES (902,3,5002,'AVF',?,1)`, now.AddDate(0,0,-200))
	s.RecordEvent(5002, 902, VascEventInput{EventType: models.VAEMaturation, EventDate: now.AddDate(0,0,-100), Detail: `{"usable":true}`})
	rs := s.PatientReminders(5002)
	if !hasKind(rs, "periodic_due") {
		t.Fatalf("maturation 100天前且无physical_check，应 periodic_due, got %+v", rs)
	}
}

func TestVasc_ClassifyAccess(t *testing.T) {
	tests := []struct{ t, want string }{
		{"CVC-NCC", "ncc"},
		{"临时无隧道导管", "ncc"},
		{"CVC-TCC", "tcc"},
		{"带隧道和涤纶套的透析导管TCC", "tcc"},
		{"自体动静脉内瘘AVF", "avf_avg"},
		{"AVG", "avf_avg"},
		{"内瘘", "avf_avg"},
		{"未知类型XYZ", "other"},
	}
	for _, tt := range tests {
		if got := classifyAccess(tt.t); got != tt.want {
			t.Errorf("classifyAccess(%q) = %q, want %q", tt.t, got, tt.want)
		}
	}
}

func hasKind(rs []VascReminder, kind string) bool {
	for _, r := range rs {
		if r.Kind == kind {
			return true
		}
	}
	return false
}
