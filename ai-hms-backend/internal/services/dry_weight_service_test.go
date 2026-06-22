package services

import (
	"testing"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newDwTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.DryWeightAssessment{}, &models.PatientDryWeight{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func intPtr(v int) *int       { return &v }
func floatPtr(v float64) *float64 { return &v }

func TestDw_Assess_MainMet(t *testing.T) {
	db := newDwTestDB(t)
	s := &DryWeightService{db: db, tenantID: 3}

	// 全满足 → mainMet=true
	dwa, err := s.Assess(1001, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseMaintenance,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
		Edema: false, Palpitation: false, HeartFailure: false, Cramp: false,
	})
	if err != nil {
		t.Fatalf("assess: %v", err)
	}
	if !dwa.MainMet {
		t.Fatal("全满足应mainMet=true")
	}

	// 低血压 → mainMet=false
	dwa2, _ := s.Assess(1001, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseMaintenance,
		SBP: intPtr(95), DBP: intPtr(55), HeartRate: intPtr(75),
	})
	if dwa2.MainMet {
		t.Fatal("低血压应mainMet=false")
	}
	if dwa2.FailedReasons == "" || dwa2.FailedReasons == "[]" {
		t.Fatal("应有failed_reasons")
	}

	// 症状 → mainMet=false
	dwa3, _ := s.Assess(1001, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseMaintenance,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
		Edema: true,
	})
	if dwa3.MainMet {
		t.Fatal("水肿应mainMet=false")
	}
}

func TestDw_Assess_PhaseLimit(t *testing.T) {
	db := newDwTestDB(t)
	s := &DryWeightService{db: db, tenantID: 3}

	// 诱导期 1.2 kg → 超限拒绝
	_, err := s.Assess(1002, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseInduction,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
		AdjustKg: floatPtr(1.2),
	})
	if err == nil {
		t.Fatal("诱导期1.2kg应超限被拒")
	}

	// 诱导期 0.8 kg → 通过
	_, err = s.Assess(1002, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseInduction,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
		AdjustKg: floatPtr(0.8),
	})
	if err != nil {
		t.Fatalf("诱导期0.8kg应通过: %v", err)
	}

	// 维持期 0.6 kg → 超限拒绝
	_, err = s.Assess(1002, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseMaintenance,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
		AdjustKg: floatPtr(0.6),
	})
	if err == nil {
		t.Fatal("维持期0.6kg应超限被拒")
	}

	// 维持期 -0.6 kg → 超限拒绝（负数下调也受限）
	_, err = s.Assess(1002, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseMaintenance,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
		AdjustKg: floatPtr(-0.6),
	})
	if err == nil {
		t.Fatal("维持期-0.6kg应超限被拒")
	}

	// 诱导期 -1.2 kg → 超限拒绝
	_, err = s.Assess(1002, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseInduction,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
		AdjustKg: floatPtr(-1.2),
	})
	if err == nil {
		t.Fatal("诱导期-1.2kg应超限被拒")
	}

	// 维持期 -0.4 kg → 通过（合法负数调整）
	_, err = s.Assess(1002, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseMaintenance,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
		AdjustKg: floatPtr(-0.4),
	})
	if err != nil {
		t.Fatalf("维持期-0.4kg应通过: %v", err)
	}
}

func TestDw_Confirm(t *testing.T) {
	db := newDwTestDB(t)
	s := &DryWeightService{db: db, tenantID: 3}

	// 首次确认
	result, err := s.Confirm(2001, DwConfirmInput{
		DryWeight: 65.0, Phase: models.DWPhaseMaintenance,
		ACTR: floatPtr(0.35), ConfirmedBy: "10",
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if result.DryWeight != 65.0 || result.Phase != models.DWPhaseMaintenance {
		t.Fatalf("bad confirm: %+v", result)
	}

	// upsert 再次确认
	result2, err := s.Confirm(2001, DwConfirmInput{
		DryWeight: 64.0, Phase: models.DWPhaseMaintenance,
		ACTR: floatPtr(0.34), ConfirmedBy: "10",
	})
	if err != nil {
		t.Fatalf("re-confirm: %v", err)
	}
	if result2.DryWeight != 64.0 {
		t.Fatalf("re-confirm dryWeight应为64.0: %f", result2.DryWeight)
	}

	// Current
	current, err := s.Current(2001)
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	if current.DryWeight == nil || *current.DryWeight != 64.0 {
		t.Fatalf("current dryWeight=%v", current.DryWeight)
	}
	if current.Phase != models.DWPhaseMaintenance {
		t.Fatalf("phase=%s", current.Phase)
	}
	if current.SuggestedRNa != 1.025 {
		t.Fatalf("维持期RNa=1.025: %f", current.SuggestedRNa)
	}

	// 无确定记录的患者
	current2, err := s.Current(9999)
	if err != nil || current2.Phase != models.DWPhaseInduction || current2.SuggestedRNa != 1.05 {
		t.Fatalf("无记录应诱导期RNa=1.05: %+v", current2)
	}
	if current2.DryWeight != nil {
		t.Fatalf("无确定记录应dryWeight=nil, got=%v", current2.DryWeight)
	}
}

func TestDw_ListAssessments(t *testing.T) {
	db := newDwTestDB(t)
	s := &DryWeightService{db: db, tenantID: 3}

	s.Assess(3001, DwAssessInput{
		AssessType: "daily", Phase: models.DWPhaseMaintenance,
		SBP: intPtr(120), DBP: intPtr(80), HeartRate: intPtr(75),
	})
	s.Assess(3001, DwAssessInput{
		AssessType: "cycle", Phase: models.DWPhaseInduction,
		SBP: intPtr(110), DBP: intPtr(70), HeartRate: intPtr(80),
	})

	rows, err := s.ListAssessments(3001)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("应有2条评估: %d", len(rows))
	}
}
