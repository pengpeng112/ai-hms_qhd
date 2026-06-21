package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elliotxin/ai-hms-backend/internal/integrations/actrs"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newActrTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.PatientACTR{}, &models.ExternalPatientMapping{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS ["Register_PatientInfomation"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "Name" TEXT, "DialysisNo" TEXT)`)
	db.Exec(`INSERT INTO ["Register_PatientInfomation"] ("Id","TenantId","Name","DialysisNo") VALUES (1001,3,'张三','D-001')`)
	db.Exec(`INSERT INTO ["Register_PatientInfomation"] ("Id","TenantId","Name","DialysisNo") VALUES (1002,3,'李四','')`)
	return db
}

func fakeActrsServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"access_token": "tok"})
	})
	mux.HandleFunc("/patients", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": 77, "dialysis_id": "D-001", "name": "张三"})
	})
	mux.HandleFunc("/patients/77/xrays", func(w http.ResponseWriter, r *http.Request) {
		ctr := 0.48
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "ctr": ctr, "qc_pass": true}})
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "ctr": ctr, "qc_pass": true, "model_version": "v8.5"})
	})
	mux.HandleFunc("/xrays/9/correction", func(w http.ResponseWriter, _ *http.Request) {
		val := 0.50
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "ctr": &val, "doctor_correction": &val})
	})
	return httptest.NewServer(mux)
}

func newActrSvc(t *testing.T, db *gorm.DB, baseURL string) *ActrService {
	t.Helper()
	return NewActrServiceWith(db, actrs.NewClient(actrs.Config{BaseURL: baseURL, Username: "u", Password: "p", TimeoutSec: 5}), true, 3)
}

func TestActr_EnsurePatientMapping_LazyCreate(t *testing.T) {
	db := newActrTestDB(t)
	srv := fakeActrsServer(t)
	defer srv.Close()
	s := newActrSvc(t, db, srv.URL)

	actrsID, dialysisNo, err := s.ensurePatientMapping(context.Background(), 1001)
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if actrsID != 77 || dialysisNo != "D-001" {
		t.Fatalf("want 77/D-001, got %d/%s", actrsID, dialysisNo)
	}
	var cnt int64
	db.Model(&models.ExternalPatientMapping{}).Where("external_system = ?", ExternalSystemACTRS).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("mapping should be created, got %d", cnt)
	}
	if _, _, err := s.ensurePatientMapping(context.Background(), 1001); err != nil {
		t.Fatalf("ensure#2: %v", err)
	}
	db.Model(&models.ExternalPatientMapping{}).Where("external_system = ?", ExternalSystemACTRS).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("mapping should stay 1, got %d", cnt)
	}
}

func TestActr_EnsurePatientMapping_NoDialysisNo(t *testing.T) {
	db := newActrTestDB(t)
	srv := fakeActrsServer(t)
	defer srv.Close()
	s := newActrSvc(t, db, srv.URL)

	_, _, err := s.ensurePatientMapping(context.Background(), 1002)
	if err == nil || err.Error() == "" {
		t.Fatalf("should reject patient without dialysis no")
	}
}

func TestActr_Analyze_PersistsAndIdempotent(t *testing.T) {
	db := newActrTestDB(t)
	srv := fakeActrsServer(t)
	defer srv.Close()
	s := newActrSvc(t, db, srv.URL)

	rec, err := s.Analyze(context.Background(), 1001, "chest.jpg", strings.NewReader("bytes"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if rec.ActrsXrayID != 9 || rec.CTR == nil || rec.QCPass != 1 || rec.Source != "manual" {
		t.Fatalf("bad rec: %+v", rec)
	}
	if _, err := s.Analyze(context.Background(), 1001, "chest.jpg", strings.NewReader("bytes")); err != nil {
		t.Fatalf("Analyze#2: %v", err)
	}
	var cnt int64
	db.Model(&models.PatientACTR{}).Where("tenant_id = ? AND patient_id = ? AND actrs_xray_id = ?", 3, "1001", 9).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("should be idempotent, got %d rows", cnt)
	}
}

func TestActr_AdoptWritesDraftNoSign(t *testing.T) {
	db := newActrTestDB(t)
	srv := fakeActrsServer(t)
	defer srv.Close()
	s := newActrSvc(t, db, srv.URL)

	db.Exec(`CREATE TABLE IF NOT EXISTS ["Plan_PatientPrescription"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "DryWeight" REAL, "UFQuantity" REAL, "ConfirmTime" DATETIME, "Note" TEXT, "LastModifyTime" DATETIME)`)
	db.Exec(`INSERT INTO ["Plan_PatientPrescription"] ("Id","TenantId","PatientId") VALUES (7001,3,1001)`)
	db.Exec(`INSERT INTO ["Plan_PatientPrescription"] ("Id","TenantId","PatientId","ConfirmTime") VALUES (7002,3,1001,datetime('now'))`)

	rec, _ := s.Analyze(context.Background(), 1001, "c.jpg", strings.NewReader("b"))

	dw, uf := 62.5, 2.8
	if err := s.AdoptToPrescription(context.Background(), 1001, 7001, rec.ID, &dw, &uf, "医生A"); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var row struct {
		DryWeight   *float64 `gorm:"column:DryWeight"`
		UFQuantity  *float64 `gorm:"column:UFQuantity"`
		ConfirmTime *string  `gorm:"column:ConfirmTime"`
	}
	db.Table(`"Plan_PatientPrescription"`).Where(`"Id" = ?`, 7001).Scan(&row)
	if row.DryWeight == nil || *row.DryWeight != 62.5 || row.UFQuantity == nil || *row.UFQuantity != 2.8 {
		t.Fatalf("draft not written: %+v", row)
	}
	if row.ConfirmTime != nil {
		t.Fatalf("adopt must NOT sign (ConfirmTime should stay null)")
	}

	var actr models.PatientACTR
	db.Where("id = ?", rec.ID).First(&actr)
	if actr.AdoptedBy != "医生A" || actr.AdoptedPrescriptionID != "7001" {
		t.Fatalf("audit fields not set: %+v", actr)
	}
}

func TestActr_AdoptRejectSigned(t *testing.T) {
	db := newActrTestDB(t)
	srv := fakeActrsServer(t)
	defer srv.Close()
	s := newActrSvc(t, db, srv.URL)

	db.Exec(`CREATE TABLE IF NOT EXISTS ["Plan_PatientPrescription"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "DryWeight" REAL, "UFQuantity" REAL, "ConfirmTime" DATETIME, "Note" TEXT, "LastModifyTime" DATETIME)`)
	db.Exec(`INSERT INTO ["Plan_PatientPrescription"] ("Id","TenantId","PatientId","ConfirmTime") VALUES (7002,3,1001,datetime('now'))`)

	rec, _ := s.Analyze(context.Background(), 1001, "c.jpg", strings.NewReader("b"))
	dw, uf := 62.5, 2.8
	err := s.AdoptToPrescription(context.Background(), 1001, 7002, rec.ID, &dw, &uf, "医生A")
	if err != ErrPrescriptionSigned {
		t.Fatalf("should reject signed prescription, got: %v", err)
	}
}

func TestActr_DegradeReadOnly(t *testing.T) {
	db := newActrTestDB(t)
	s := NewActrServiceWith(db, nil, false, 3)

	if _, err := s.Analyze(context.Background(), 1001, "c.jpg", strings.NewReader("b")); err != ErrActrsDisabled {
		t.Fatalf("disabled Analyze should return ErrActrsDisabled, got %v", err)
	}
	dw, uf := 60.0, 2.0
	if err := s.AdoptToPrescription(context.Background(), 1001, 7001, "x", &dw, &uf, "医生A"); err != ErrActrsDisabled {
		t.Fatalf("disabled Adopt should return ErrActrsDisabled, got %v", err)
	}
	if _, err := s.Correct(context.Background(), 1001, "x", "医生A", 0.5); err != ErrActrsDisabled {
		t.Fatalf("disabled Correct should return ErrActrsDisabled, got %v", err)
	}
	rows, err := s.History(1001)
	if err != nil {
		t.Fatalf("History should work offline: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected empty history, got %d", len(rows))
	}
	st := s.Status()
	if st["enabled"] != false {
		t.Fatalf("status should show disabled")
	}
}

func TestActr_History(t *testing.T) {
	db := newActrTestDB(t)
	srv := fakeActrsServer(t)
	defer srv.Close()
	s := newActrSvc(t, db, srv.URL)

	_, err := s.Analyze(context.Background(), 1001, "c.jpg", strings.NewReader("b"))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	rows, err := s.History(1001)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestActr_Correct(t *testing.T) {
	db := newActrTestDB(t)
	srv := fakeActrsServer(t)
	defer srv.Close()
	s := newActrSvc(t, db, srv.URL)

	rec, _ := s.Analyze(context.Background(), 1001, "c.jpg", strings.NewReader("b"))
	updated, err := s.Correct(context.Background(), 1001, rec.ID, "医生B", 0.55)
	if err != nil {
		t.Fatalf("Correct: %v", err)
	}
	if updated.CorrectedBy != "医生B" || updated.DoctorCorrection == nil || *updated.DoctorCorrection != 0.55 {
		t.Fatalf("correction not saved: %+v", updated)
	}
}
