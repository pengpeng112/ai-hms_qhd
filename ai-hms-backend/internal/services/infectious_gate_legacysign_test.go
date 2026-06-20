package services

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var lsGateSeq atomic.Int64

func newLegacySignGateDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:lsgate%d?mode=memory&cache=shared", lsGateSeq.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// infectious + sign_record tables
	if err := db.AutoMigrate(&models.PatientInfectious{}, &models.SignRecord{}); err != nil {
		t.Fatalf("migrate models: %v", err)
	}
	// legacy prescription table — use AutoMigrate on the package-private struct
	if err := db.AutoMigrate(&legacyPatientPrescription{}); err != nil {
		t.Fatalf("migrate prescription: %v", err)
	}
	return db
}

func TestLegacySign_InfectiousFrozenBlocks(t *testing.T) {
	db := newLegacySignGateDB(t)

	// Seed a prescription row: patient 5001, prescription 7001, Status=1 (active), HD mode
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","Status","DialysisMethod") VALUES (7001,3,5001,1,'HD')`)

	// Screen patient 5001 positive → FROZEN (unhandled positive)
	inf := &InfectiousService{db: db, tenantID: LegacyTenantID}
	if _, err := inf.Screen(5001, ScreenInput{
		ScreenDate: time.Now(), Source: "manual",
		Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemPositive}},
	}); err != nil {
		t.Fatalf("Screen: %v", err)
	}

	s := &PrescriptionService{db: db}
	_, err := s.LegacySign("5001", "7001", "9", "张医生", false)
	if err == nil || !strings.Contains(err.Error(), "冻结") {
		t.Fatalf("阳性未处置应拦签发(含'冻结'), got %v", err)
	}
}

func TestLegacySign_InfectiousCZoneAckRequired(t *testing.T) {
	db := newLegacySignGateDB(t)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","Status","DialysisMethod") VALUES (7002,3,5002,1,'HD')`)

	// No screening → gate returns REQUIRE_C_ZONE (no screening = unscreened)
	s := &PrescriptionService{db: db}

	// Without cZoneAck → should block
	_, err := s.LegacySign("5002", "7002", "9", "张医生", false)
	if err == nil || !strings.Contains(err.Error(), "cZoneAck") {
		t.Fatalf("未筛查应拦签发并提示 cZoneAck, got %v", err)
	}
}

func TestLegacySign_InfectiousCZoneCRRT_BlocksNonCRRT(t *testing.T) {
	db := newLegacySignGateDB(t)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","Status","DialysisMethod") VALUES (7003,3,5003,1,'HD')`)

	// Screen positive, then dispose as C_ZONE_CRRT (needs doctor + head_nurse dual sign)
	inf := &InfectiousService{db: db, tenantID: LegacyTenantID}
	rec, err := inf.Screen(5003, ScreenInput{
		ScreenDate: time.Now(), Source: "manual",
		Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemPositive}},
	})
	if err != nil {
		t.Fatalf("Screen: %v", err)
	}
	// Dual dispose to reach C_ZONE_CRRT state
	if _, err := inf.Dispose(5003, rec.ID, DispositionInput{
		Disposition: models.InfectiousDispCZoneCRRT, Role: "doctor", SignerID: "9", SignerName: "张医生",
	}); err != nil {
		t.Fatalf("doctor dispose: %v", err)
	}
	if _, err := inf.Dispose(5003, rec.ID, DispositionInput{
		Disposition: models.InfectiousDispCZoneCRRT, Role: "head_nurse", SignerID: "8", SignerName: "李护士长",
	}); err != nil {
		t.Fatalf("head_nurse dispose: %v", err)
	}

	// Verify gate is C_ZONE_CRRT
	g := inf.CanScheduleRoutine(5003)
	if g.State != GateCZoneCRRT {
		t.Fatalf("expected C_ZONE_CRRT, got %s", g.State)
	}

	// Prescription is HD (not CRRT) → should block
	s := &PrescriptionService{db: db}
	_, err = s.LegacySign("5003", "7003", "9", "张医生", false)
	if err == nil || !strings.Contains(err.Error(), "CRRT") {
		t.Fatalf("非CRRT处方应被拦截(含'CRRT'), got %v", err)
	}
}

func TestLegacySign_InfectiousCZoneCRRT_AllowsCRRT(t *testing.T) {
	db := newLegacySignGateDB(t)
	// Prescription with CRRT mode
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","Status","DialysisMethod") VALUES (7004,3,5004,1,'CRRT')`)

	inf := &InfectiousService{db: db, tenantID: LegacyTenantID}
	rec, err := inf.Screen(5004, ScreenInput{
		ScreenDate: time.Now(), Source: "manual",
		Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemPositive}},
	})
	if err != nil {
		t.Fatalf("Screen: %v", err)
	}
	if _, err := inf.Dispose(5004, rec.ID, DispositionInput{
		Disposition: models.InfectiousDispCZoneCRRT, Role: "doctor", SignerID: "9", SignerName: "张医生",
	}); err != nil {
		t.Fatalf("doctor dispose: %v", err)
	}
	if _, err := inf.Dispose(5004, rec.ID, DispositionInput{
		Disposition: models.InfectiousDispCZoneCRRT, Role: "head_nurse", SignerID: "8", SignerName: "李护士长",
	}); err != nil {
		t.Fatalf("head_nurse dispose: %v", err)
	}

	// CRRT prescription + C_ZONE_CRRT gate → should pass the infectious gate
	// (may fail at on-duty gate or other checks, but NOT at the infectious gate)
	s := &PrescriptionService{db: db}
	_, err = s.LegacySign("5004", "7004", "9", "张医生", false)
	// Should NOT contain CRRT-related block message
	if err != nil && strings.Contains(err.Error(), "CRRT") {
		t.Fatalf("CRRT处方不应被传染病门禁拦截, got %v", err)
	}
	if err != nil && strings.Contains(err.Error(), "冻结") {
		t.Fatalf("CRRT处方不应被传染病冻结拦截, got %v", err)
	}
	// It's OK if it fails at a later gate (e.g., on-duty doctor check), just not at the infectious gate
}

func TestLegacySign_InfectiousAllowNormal(t *testing.T) {
	db := newLegacySignGateDB(t)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","Status","DialysisMethod") VALUES (7005,3,5005,1,'HD')`)

	// Screen negative → ALLOW_NORMAL
	inf := &InfectiousService{db: db, tenantID: LegacyTenantID}
	if _, err := inf.Screen(5005, ScreenInput{
		ScreenDate: time.Now(), Source: "manual",
		Items: []ScreenItem{{Item: "HBsAg", Result: models.InfItemNegative}},
	}); err != nil {
		t.Fatalf("Screen: %v", err)
	}

	// Should pass infectious gate (may fail at later checks)
	s := &PrescriptionService{db: db}
	_, err := s.LegacySign("5005", "7005", "9", "张医生", false)
	// Should NOT fail with infectious-gate reasons
	if err != nil && (strings.Contains(err.Error(), "冻结") || strings.Contains(err.Error(), "cZoneAck") || strings.Contains(err.Error(), "CRRT")) {
		t.Fatalf("阴性筛查不应被传染病门禁拦截, got %v", err)
	}
}
