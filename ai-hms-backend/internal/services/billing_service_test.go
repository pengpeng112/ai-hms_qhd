package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupBillingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS charge_record (
			id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL, patient_id INTEGER,
			treatment_id INTEGER NOT NULL, prescription_id INTEGER,
			charge_date DATETIME, shift TEXT, dialysis_mode TEXT, access_type TEXT,
			crrt_hours REAL, total_amount REAL, status TEXT NOT NULL DEFAULT 'draft',
			recorded_by TEXT, recorded_name TEXT, checked_by TEXT, checked_name TEXT,
			checked_at DATETIME, exported_at DATETIME, pushed_at DATETIME, note TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS charge_line (
			id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL,
			charge_record_id TEXT NOT NULL, category TEXT NOT NULL,
			item_code TEXT, item_name TEXT NOT NULL, spec TEXT, unit TEXT,
			quantity REAL, unit_price REAL, amount REAL,
			billable INTEGER NOT NULL DEFAULT 1, source TEXT NOT NULL DEFAULT 'auto',
			charge_item_id INTEGER, his_price_item_id TEXT, his_item_code TEXT,
			his_item_class TEXT, his_item_name TEXT, price_source TEXT,
			matched_status TEXT, note TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS his_price_item (
			id TEXT PRIMARY KEY, source_system TEXT NOT NULL DEFAULT 'HIS_ORACLE',
			item_class TEXT, item_code TEXT NOT NULL, item_name TEXT, item_spec TEXT,
			units TEXT, price REAL, prefer_price REAL, foreigner_price REAL,
			performed_by TEXT, fee_type_mask INTEGER, class_on_inp_rcpt TEXT,
			class_on_outp_rcpt TEXT, class_on_reckoning TEXT, subj_code TEXT,
			class_on_mr TEXT, memo TEXT, start_date DATETIME, stop_date DATETIME,
			operator_code TEXT, enter_date DATETIME, high_price REAL,
			material_code TEXT, score_1 REAL, score_2 REAL, price_name_code TEXT,
			control_flag TEXT, input_code TEXT, input_code_wb TEXT, std_code_1 TEXT,
			changed_memo TEXT, class_on_insur_mr TEXT, package_spec TEXT,
			firm_id TEXT, charge_according TEXT, license_id TEXT, update_flag REAL,
			dept_name TEXT, update_flag_syb REAL, mr_bill_class TEXT,
			class_on_mr_add TEXT, cwtj_code TEXT, high_value REAL, drg_code TEXT,
			insur_update INTEGER, stop_operator TEXT, limit_quantity REAL,
			is_active INTEGER NOT NULL DEFAULT 1, synced_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			sync_run_id TEXT, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS medication_admin (
			id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL, patient_id INTEGER,
			order_id INTEGER NOT NULL, treatment_id INTEGER, drug_name TEXT NOT NULL,
			category TEXT, dose TEXT, route TEXT, timing TEXT,
			administered_by TEXT NOT NULL, administered_name TEXT,
			administered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			second_check_by TEXT, second_check_name TEXT, second_check_at DATETIME,
			status TEXT NOT NULL DEFAULT 'recorded', note TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
`CREATE TABLE IF NOT EXISTS "Treatment_MaterialTrace" (
			"Id" INTEGER PRIMARY KEY, "TenantId" INTEGER NOT NULL,
			"TreatmentId" INTEGER, "ChargeItemId" INTEGER, "Num" REAL
		)`,
		`CREATE TABLE IF NOT EXISTS "Treatment_Treatment" (
			"Id" INTEGER PRIMARY KEY, "TenantId" INTEGER NOT NULL,
			"PatientId" INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS "Stock_ChargeItem" (
			"Id" INTEGER PRIMARY KEY, "TenantId" INTEGER NOT NULL,
			"Name" TEXT, "Unit" TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_cr_tenant_treatment_active
		 ON charge_record (tenant_id, treatment_id) WHERE status <> 'cancelled'`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("ddl failed: %v", err)
		}
	}
	return db
}

func TestBilling_BuildDraft(t *testing.T) {
	svc := &BillingService{
		db:          setupBillingTestDB(t),
		hisPriceSvc: &HisPriceService{},
		tenantID:    3,
	}

	mode := "hd"
	access := "avf"
	req := BuildDraftRequest{
		TreatmentID:  100,
		DialysisMode: &mode,
		AccessType:   &access,
	}

	rec, err := svc.BuildDraft(req, "user1", "测试护士")
	if err != nil {
		t.Fatalf("BuildDraft failed: %v", err)
	}
	if rec.ID == "" {
		t.Error("expected id")
	}
	if rec.Status != models.ChargeStatusDraft {
		t.Errorf("expected status draft, got %s", rec.Status)
	}
	if rec.TreatmentID != 100 {
		t.Errorf("expected treatment_id 100, got %d", rec.TreatmentID)
	}

	found := false
	for _, l := range rec.Lines {
		if l.Category == models.ChargeCatTreatment {
			found = true
			if l.ItemName == "" {
				t.Error("treatment line should have item name")
			}
		}
	}
	if !found {
		t.Error("expected at least one treatment line")
	}
}

func TestBilling_BuildDraft_Idempotent(t *testing.T) {
	svc := &BillingService{
		db:          setupBillingTestDB(t),
		hisPriceSvc: &HisPriceService{},
		tenantID:    3,
	}

	mode := "hd"
	req := BuildDraftRequest{TreatmentID: 200, DialysisMode: &mode}

	rec1, err := svc.BuildDraft(req, "user1", "护士A")
	if err != nil {
		t.Fatalf("first build failed: %v", err)
	}

	rec2, err := svc.BuildDraft(req, "user2", "护士B")
	if err != nil {
		t.Fatalf("second build failed: %v", err)
	}

	if rec1.ID != rec2.ID {
		t.Errorf("expected same record id, got %s vs %s", rec1.ID, rec2.ID)
	}

	var count int64
	svc.db.Model(&models.ChargeRecord{}).Where("treatment_id = ? AND status <> ?", 200, "cancelled").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 non-cancelled record, got %d", count)
	}
}

func TestBilling_StatusFlow(t *testing.T) {
	svc := &BillingService{
		db:          setupBillingTestDB(t),
		hisPriceSvc: &HisPriceService{},
		tenantID:    3,
	}

	mode := "hd"
	req := BuildDraftRequest{TreatmentID: 300, DialysisMode: &mode}
	rec, err := svc.BuildDraft(req, "user1", "护士")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if _, err := svc.Confirm(rec.ID, "user1", "护士"); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	rec, _ = svc.GetCharge(rec.ID)
	if rec.Status != models.ChargeStatusConfirmed {
		t.Errorf("expected confirmed, got %s", rec.Status)
	}

	if _, err := svc.Check(rec.ID, "user2", "护士长"); err != nil {
		t.Fatalf("check failed: %v", err)
	}
	rec, _ = svc.GetCharge(rec.ID)
	if rec.Status != models.ChargeStatusChecked {
		t.Errorf("expected checked, got %s", rec.Status)
	}

	if _, err := svc.Cancel(rec.ID, "测试取消"); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	rec, _ = svc.GetCharge(rec.ID)
	if rec.Status != models.ChargeStatusCancelled {
		t.Errorf("expected cancelled, got %s", rec.Status)
	}

	if _, err := svc.Confirm(rec.ID, "user3", "护士"); err == nil {
		t.Error("confirm should fail on cancelled record")
	}
}

func TestBilling_CheckedLock(t *testing.T) {
	svc := &BillingService{
		db:          setupBillingTestDB(t),
		hisPriceSvc: &HisPriceService{},
		tenantID:    3,
	}

	mode := "hd"
	req := BuildDraftRequest{TreatmentID: 400, DialysisMode: &mode}
	rec, err := svc.BuildDraft(req, "user1", "护士")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	line := &models.ChargeLine{
		ItemName: "测试手工项目",
		Category: models.ChargeCatMaterial,
		Billable: true,
	}
	if _, err := svc.AddLine(rec.ID, line); err != nil {
		t.Fatalf("add line failed: %v", err)
	}

	svc.Confirm(rec.ID, "user1", "护士")
	svc.Check(rec.ID, "user2", "护士长")

	_, err = svc.AddLine(rec.ID, line)
	if err == nil {
		t.Error("add line should fail on checked record")
	}

	err = svc.DeleteLine(rec.Lines[0].ID)
	if err == nil {
		t.Error("delete line should fail on checked record")
	}
}

func TestBilling_MarkExported(t *testing.T) {
	svc := &BillingService{
		db:          setupBillingTestDB(t),
		hisPriceSvc: &HisPriceService{},
		tenantID:    3,
	}

	mode := "hd"
	req := BuildDraftRequest{TreatmentID: 500, DialysisMode: &mode}
	rec, err := svc.BuildDraft(req, "user1", "护士")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	svc.Confirm(rec.ID, "user1", "护士")
	svc.Check(rec.ID, "user2", "护士长")

	exported, err := svc.MarkExported(rec.ID)
	if err != nil {
		t.Fatalf("mark exported failed: %v", err)
	}
	if exported.ExportedAt == nil {
		t.Error("expected exported_at")
	}
	if exported.Status != models.ChargeStatusChecked {
		t.Errorf("status should still be checked after export, got %s", exported.Status)
	}
}

func TestBilling_ConcurrentBuild(t *testing.T) {
	svc := &BillingService{
		db:          setupBillingTestDB(t),
		hisPriceSvc: &HisPriceService{},
		tenantID:    3,
	}

	mode := "hd"
	req := BuildDraftRequest{TreatmentID: 600, DialysisMode: &mode}

	errs := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			time.Sleep(time.Duration(idx*10) * time.Millisecond)
			_, err := svc.BuildDraft(req, fmt.Sprintf("user%d", idx), fmt.Sprintf("护士%d", idx))
			errs <- err
		}(i)
	}

	for i := 0; i < 5; i++ {
		if err := <-errs; err != nil {
			t.Logf("goroutine %d error: %v", i, err)
		}
	}

	var count int64
	svc.db.Model(&models.ChargeRecord{}).Where("treatment_id = ? AND status <> ?", 600, "cancelled").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 record after concurrent builds, got %d", count)
	}
}

func TestBilling_NoopPusher(t *testing.T) {
	p := NoopPusher{}
	if p.Channel() != "noop" {
		t.Error("unexpected channel")
	}
	rec := &models.ChargeRecord{ID: "x", Status: models.ChargeStatusChecked}
	res, err := p.Push(rec)
	if err != nil {
		t.Error("noop push should not error on checked record")
	}
	if res.Accepted {
		t.Error("noop should not accept")
	}
}
