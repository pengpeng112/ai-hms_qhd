package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// newTestPatientService 构造内存库，建一张与老库一致列集的 Plan_PatientPlan 表，
// 用于验证建档时同建草稿方案的逻辑（createDraftTreatmentPlan）。
func newTestPatientService(t *testing.T) *PatientService {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	// 生产代码 createDraftTreatmentPlan 用 .Table(`"Plan_PatientPlan"`) 写入；
	// 在 sqlite 下该引用解析为带引号的表名，故 DDL 与读取均沿用同一形式（Postgres 下两者等价）。
	stmt := `CREATE TABLE ["Plan_PatientPlan"] (
		"Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "Name" TEXT,
		"CreatorId" INTEGER, "CreateTime" DATETIME, "OddWeekFrequency" INTEGER, "EvenWeekFrequency" INTEGER,
		"DialysisMethod" TEXT, "DialysisDuration" REAL, "DryWeight" REAL, "ExtraWeight" REAL,
		"BF" REAL, "BV" REAL, "FirstAnticoagulant" INTEGER, "FirstDosage" REAL, "MaintainAnticoagulant" INTEGER,
		"DilutionProportion" REAL, "InjectionRate" REAL, "InjectionDuration" REAL, "InjectionVolume" REAL,
		"Dialysate" TEXT, "DialysateFlow" REAL, "DialysateVolume" REAL, "NaIonCon" REAL, "CaIonCon" REAL,
		"KIonCon" REAL, "HCO3IonCon" REAL, "Conductivity" REAL, "DialysateTmp" REAL, "SubstituateVolume" REAL,
		"DilutionMnt" TEXT, "IsDisabled" BOOLEAN, "LastModifyTime" DATETIME, "Frequency" TEXT, "GlucoseCon" REAL,
		"DialysateGroupId" INTEGER, "AutoConfirmPrescription" TEXT, "Note" TEXT, "SubstituateFlow" REAL,
		"VascularAccessId" INTEGER )`
	if err := db.Exec(stmt).Error; err != nil {
		t.Fatalf("create Plan_PatientPlan table failed: %v", err)
	}

	return &PatientService{db: db}
}

// 建档侧：同建一条草稿治疗方案，干体重落库、种子与临床默认正确、草稿标记成立。
// 一举验证两个历史病根的修复——干体重不再丢弃、裸患者不再无方案。
func TestCreateDraftTreatmentPlanSeedsAndMarksDraft(t *testing.T) {
	s := newTestPatientService(t)

	req := CreateRequest{
		Name:        "张三",
		DryWeight:   62.5,
		DefaultMode: "HDF",
	}
	pid := modeltypes.LegacyID(100001)
	now := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)

	if err := s.createDraftTreatmentPlan(s.db, pid, req.Name, req, now); err != nil {
		t.Fatalf("createDraftTreatmentPlan failed: %v", err)
	}

	var plan legacyPatientPlan
	if err := s.db.Table(`"Plan_PatientPlan"`).
		Where(`"PatientId" = ? AND "TenantId" = ?`, pid, LegacyTenantID).
		First(&plan).Error; err != nil {
		t.Fatalf("draft plan not persisted: %v", err)
	}

	// 种子：干体重落库正确（修复历史丢弃）。
	if plan.DryWeight != 62.5 {
		t.Fatalf("DryWeight = %v, want 62.5", plan.DryWeight)
	}
	// 种子：默认透析模式。
	if plan.DialysisMethod != "HDF" {
		t.Fatalf("DialysisMethod = %q, want HDF", plan.DialysisMethod)
	}
	// 临床默认：标准钾 2.0 / 钠基线 140。
	if plan.KIonCon != draftPlanDefaultK {
		t.Fatalf("KIonCon = %v, want %v", plan.KIonCon, draftPlanDefaultK)
	}
	if plan.NaIonCon != draftPlanDefaultNa {
		t.Fatalf("NaIonCon = %v, want %v", plan.NaIonCon, draftPlanDefaultNa)
	}
	// 草稿标记（契约02 三）：通路 0 占位 + 配方为空。
	if plan.VascularAccessID != 0 {
		t.Fatalf("VascularAccessID = %v, want 0 (draft placeholder)", plan.VascularAccessID)
	}
	if strings.TrimSpace(plan.Dialysate) != "" {
		t.Fatalf("Dialysate = %q, want empty (draft marker)", plan.Dialysate)
	}
	// 频率留 0：未补全前不可进入排班。
	if plan.OddWeekFrequency != 0 {
		t.Fatalf("OddWeekFrequency = %v, want 0", plan.OddWeekFrequency)
	}
	// 租户与启用状态。
	if plan.TenantID != LegacyTenantID {
		t.Fatalf("TenantID = %v, want %v", plan.TenantID, LegacyTenantID)
	}
	if plan.IsDisabled {
		t.Fatalf("draft plan should be enabled (IsDisabled=false)")
	}

	// 完整性判定：草稿（配方空）= 未完成，必须被门禁拦截。
	if isLegacyPlanComplete(&plan) {
		t.Fatalf("draft plan (empty dialysate) must be judged incomplete")
	}

	// 方案页补全透析液配方后 → 完整，可开方/上机。
	plan.Dialysate = "碳酸氢盐透析液"
	if !isLegacyPlanComplete(&plan) {
		t.Fatalf("plan with dialysate filled must be judged complete")
	}
}
