package services

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestPatientWardToday(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE ["Treatment_Treatment"] (
		"Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "WardId" INTEGER,
		"StartTime" DATETIME, "SignInTime" DATETIME, "ReceptionTime" DATETIME, "CreateTime" DATETIME)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	s := &PrescriptionService{db: db}
	now := time.Now()

	mk := func(id, pid, ward int64, when time.Time) {
		if err := db.Table(`"Treatment_Treatment"`).Create(map[string]any{
			"Id": id, "TenantId": LegacyTenantID, "PatientId": pid, "WardId": ward, "StartTime": when,
		}).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	mk(1, 1001, 5, now)
	mk(2, 3003, 7, now.AddDate(0, 0, -1))

	if w := s.patientWardToday(1001, now); w != 5 {
		t.Errorf("今日治疗病人应得病区 5, 得 %d", w)
	}
	if w := s.patientWardToday(2002, now); w != 0 {
		t.Errorf("无治疗病人应得 0, 得 %d", w)
	}
	if w := s.patientWardToday(3003, now); w != 0 {
		t.Errorf("仅昨日治疗应得 0（不算今日）, 得 %d", w)
	}
}
