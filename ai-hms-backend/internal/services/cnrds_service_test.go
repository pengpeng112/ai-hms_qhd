package services

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var cnrdsTestDBSeq atomic.Int64

func newCnrdsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:cnrdstest%d?mode=memory&cache=shared", cnrdsTestDBSeq.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.CnrdsReport{}); err != nil {
		t.Fatalf("migrate cnrds_report: %v", err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS ["Register_PatientInfomation"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "Name" TEXT, "Gender" TEXT)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS ["Register_OutCome"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "Type" TEXT, "Reason" TEXT, "OutComeTime" DATETIME, "CreateTime" DATETIME)`)

	db.Exec(`INSERT INTO ["Register_PatientInfomation"] ("Id","TenantId","Name","Gender") VALUES (1001,3,'张三','男')`)
	db.Exec(`INSERT INTO ["Register_PatientInfomation"] ("Id","TenantId","Name","Gender") VALUES (1002,3,'李四','女')`)
	db.Exec(`INSERT INTO ["Register_OutCome"] ("Id","TenantId","PatientId","Type","Reason","OutComeTime","CreateTime") VALUES (1,3,1001,'20','死亡','2026-06-15 10:00:00','2026-06-15 10:00:00')`)
	db.Exec(`INSERT INTO ["Register_OutCome"] ("Id","TenantId","PatientId","Type","Reason","OutComeTime","CreateTime") VALUES (2,3,1002,'20','转肾移植','2026-06-10 08:00:00','2026-06-10 08:00:00')`)
	return db
}

type fakeLab struct{}

func (fakeLab) Value(patientID int64, conceptID string, start, end time.Time) *float64 {
	vals := map[string]float64{
		"HEMOGLOBIN": 110,
		"SERUM_CA":   2.2,
		"SERUM_P":    1.5,
		"IPTH":       300,
		"ALBUMIN":    38,
		"KTV":        1.4,
	}
	if v, ok := vals[conceptID]; ok {
		return &v
	}
	return nil
}

func TestCnrds_GenerateMonthly(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, err := svc.GenerateMonthly("2026-06")
	if err != nil {
		t.Fatalf("GenerateMonthly: %v", err)
	}
	if rep.ReportType != models.CnrdsTypeMonthly {
		t.Fatalf("want reportType monthly, got %s", rep.ReportType)
	}
	if rep.Status != models.CnrdsStatusDraft {
		t.Fatalf("want status draft, got %s", rep.Status)
	}
	if rep.PatientCount != 2 {
		t.Fatalf("want patientCount 2, got %d", rep.PatientCount)
	}

	var content models.CnrdsContent
	if err := json.Unmarshal([]byte(rep.Content), &content); err != nil {
		t.Fatalf("unmarshal content: %v", err)
	}
	if len(content.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(content.Rows))
	}

	r1 := content.Rows[0]
	if r1.PatientID != "1001" || r1.Name != "张三" || r1.Gender != "男" {
		t.Fatalf("row0 mismatched: %+v", r1)
	}
	if r1.Hb == nil || *r1.Hb != 110 {
		t.Fatalf("want Hb=110, got %v", r1.Hb)
	}
	if r1.KtV == nil || *r1.KtV != 1.4 {
		t.Fatalf("want KtV=1.4, got %v", r1.KtV)
	}
	if r1.PrimaryDiagnosis != "" || r1.DialysisMode != "" || r1.InfMarkers != "" {
		t.Fatalf("best-effort fields should be empty")
	}
	if r1.OutcomeType != models.CnrdsEventDeath || r1.DeathReason != "死亡" {
		t.Fatalf("want latest death outcome, got outcome=%q reason=%q", r1.OutcomeType, r1.DeathReason)
	}
}

func TestCnrds_GenerateMonthly_EmptyPeriodRejected(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	if _, err := svc.GenerateMonthly("bad"); err == nil {
		t.Fatal("expected error for bad period")
	}
}

func TestCnrds_GenerateMonthly_NoPatients(t *testing.T) {
	db := newCnrdsTestDB(t)
	db.Exec(`DELETE FROM ["Register_PatientInfomation"]`)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, err := svc.GenerateMonthly("2026-06")
	if err != nil {
		t.Fatalf("GenerateMonthly: %v", err)
	}
	if rep.PatientCount != 0 {
		t.Fatalf("want 0 patients, got %d", rep.PatientCount)
	}
	var content models.CnrdsContent
	json.Unmarshal([]byte(rep.Content), &content)
	if len(content.Rows) != 0 {
		t.Fatalf("want 0 rows, got %d", len(content.Rows))
	}
}

func TestCnrds_GenerateEvent_Death(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, err := svc.GenerateEvent(1001, models.CnrdsEventDeath)
	if err != nil {
		t.Fatalf("GenerateEvent death: %v", err)
	}
	if rep.ReportType != models.CnrdsTypeEvent {
		t.Fatalf("want reportType event, got %s", rep.ReportType)
	}
	if rep.EventType != models.CnrdsEventDeath {
		t.Fatalf("want eventType death, got %s", rep.EventType)
	}
	if rep.PatientID != "1001" {
		t.Fatalf("want patientID 1001, got %s", rep.PatientID)
	}

	var content models.CnrdsContent
	if err := json.Unmarshal([]byte(rep.Content), &content); err != nil {
		t.Fatalf("unmarshal content: %v", err)
	}
	if len(content.Rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(content.Rows))
	}
	row := content.Rows[0]
	if row.DeathReason != "死亡" {
		t.Fatalf("want deathReason '死亡', got %q", row.DeathReason)
	}
	if row.OutcomeType != models.CnrdsEventDeath {
		t.Fatalf("want outcomeType death, got %q", row.OutcomeType)
	}
}

func TestCnrds_GenerateEvent_Transplant(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, err := svc.GenerateEvent(1002, models.CnrdsEventTransplant)
	if err != nil {
		t.Fatalf("GenerateEvent transplant: %v", err)
	}
	if rep.EventType != models.CnrdsEventTransplant {
		t.Fatalf("want eventType transplant, got %s", rep.EventType)
	}
}

func TestCnrds_GenerateEvent_InvalidType(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	if _, err := svc.GenerateEvent(1001, "xxx"); err == nil {
		t.Fatal("expected error for invalid eventType")
	}
}

func TestCnrds_GenerateEvent_PatientNotFound(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	if _, err := svc.GenerateEvent(9999, models.CnrdsEventDeath); err == nil {
		t.Fatal("expected error for non-existent patient")
	}
}

func TestCnrds_ListAndGet(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep1, _ := svc.GenerateMonthly("2026-06")
	rep2, _ := svc.GenerateEvent(1001, models.CnrdsEventDeath)

	list, err := svc.List(CnrdsListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 reports, got %d", len(list))
	}

	listByType, _ := svc.List(CnrdsListFilter{ReportType: models.CnrdsTypeEvent})
	if len(listByType) != 1 || listByType[0].ID != rep2.ID {
		t.Fatalf("list by type mismatch")
	}

	fetched, err := svc.Get(rep1.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if fetched.ID != rep1.ID {
		t.Fatalf("Get returned wrong report")
	}
}

func TestCnrds_ExportAndSubmit(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, _ := svc.GenerateMonthly("2026-06")

	filename, data, err := svc.Export(rep.ID)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if filename == "" {
		t.Fatal("export filename empty")
	}
	if len(data) == 0 {
		t.Fatal("export data empty")
	}

	fetched, _ := svc.Get(rep.ID)
	if fetched.Status != models.CnrdsStatusExported {
		t.Fatalf("want status exported, got %s", fetched.Status)
	}
	if fetched.ExportRef == "" {
		t.Fatal("exportRef empty")
	}

	err = svc.Submit(rep.ID, "王医生")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	fetched, _ = svc.Get(rep.ID)
	if fetched.Status != models.CnrdsStatusSubmitted {
		t.Fatalf("want status submitted, got %s", fetched.Status)
	}
	if fetched.ReviewedBy != "王医生" {
		t.Fatalf("want reviewedBy '王医生', got %s", fetched.ReviewedBy)
	}
	if fetched.SubmittedAt == nil {
		t.Fatal("submittedAt nil")
	}
}

func TestCnrds_ExportSubmittedFails(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, _ := svc.GenerateMonthly("2026-06")
	svc.Export(rep.ID)
	svc.Submit(rep.ID, "医生")

	if _, _, err := svc.Export(rep.ID); err == nil {
		t.Fatal("expected error exporting submitted report")
	}
}

func TestCnrds_SubmitDraftFails(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, _ := svc.GenerateMonthly("2026-06")
	if err := svc.Submit(rep.ID, "医生"); err == nil {
		t.Fatal("expected error submitting draft without export")
	}
}

func TestCnrds_SubmitEmptyReviewerFails(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, _ := svc.GenerateMonthly("2026-06")
	svc.Export(rep.ID)
	if err := svc.Submit(rep.ID, ""); err == nil {
		t.Fatal("expected error for empty reviewer")
	}
}

func TestCnrds_SubmitTwiceFails(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, _ := svc.GenerateMonthly("2026-06")
	svc.Export(rep.ID)
	svc.Submit(rep.ID, "医生")
	if err := svc.Submit(rep.ID, "医生"); err == nil {
		t.Fatal("expected error for duplicate submit")
	}
}

func TestCnrds_CSVExportContents(t *testing.T) {
	db := newCnrdsTestDB(t)
	svc := NewCnrdsServiceWith(db, 3, fakeLab{}, CSVExporter{})

	rep, _ := svc.GenerateMonthly("2026-06")
	_, data, err := svc.Export(rep.ID)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if len(data) < 4 {
		t.Fatal("csv too short")
	}
	if data[0] != 0xEF || data[1] != 0xBB || data[2] != 0xBF {
		t.Fatal("missing UTF-8 BOM")
	}

	fetched, _ := svc.Get(rep.ID)
	if fetched.Status != models.CnrdsStatusExported {
		t.Fatalf("status not exported after export")
	}
}
