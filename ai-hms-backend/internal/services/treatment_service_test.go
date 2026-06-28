package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestTreatmentService(t *testing.T) *TreatmentService {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db failed: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	statements := []string{
		`CREATE TABLE ["Treatment_Treatment"] (
			"Id" INTEGER PRIMARY KEY,
			"TenantId" INTEGER,
			"PatientId" INTEGER,
			"ScheduleId" INTEGER,
			"ReceptionDrId" INTEGER,
			"SignInTime" DATETIME,
			"QueueNo" TEXT,
			"ReceptionTime" DATETIME,
			"DayProgrammeId" INTEGER,
			"WardId" INTEGER,
			"WardName" TEXT,
			"BedId" INTEGER,
			"ShiftId" INTEGER,
			"ShiftName" TEXT,
			"StartTime" DATETIME,
			"EndTime" DATETIME,
			"RealDuration" REAL,
			"RealUFQuantity" REAL,
			"RealSubstituateVolume" REAL,
			"NurseSummary" TEXT,
			"TreatmentSummary" TEXT,
			"Status" TEXT,
			"CaseStatus" TEXT,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME,
			"LastModifyTime" DATETIME,
			"TmrPath" TEXT,
			"TmrTime" DATETIME,
			"TmrPages" INTEGER
		)`,
		`CREATE TABLE ["Treatment_BeforeSigns"] (
			"Id" INTEGER PRIMARY KEY,
			"TenantId" INTEGER,
			"TreatmentId" INTEGER,
			"SBP" REAL,
			"DBP" REAL,
			"Complication" TEXT,
			"Symptoms" TEXT,
			"Note" TEXT,
			"OperatorId" INTEGER,
			"OperateTime" DATETIME,
			"Weight" REAL,
			"ExtraWeight" REAL,
			"HeartRate" REAL,
			"Respiration" REAL,
			"BodyTemp" REAL,
			"PressurePoint" TEXT,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME,
			"LastModifyTime" DATETIME
		)`,
		`CREATE TABLE ["Treatment_AfterSigns"] (
			"Id" INTEGER PRIMARY KEY,
			"TenantId" INTEGER,
			"TreatmentId" INTEGER,
			"SBP" REAL,
			"DBP" REAL,
			"Complication" TEXT,
			"Symptoms" TEXT,
			"Note" TEXT,
			"Weight" REAL,
			"ExtraWeight" REAL,
			"LossWeight" REAL,
			"HeartRate" REAL,
			"Respiration" REAL,
			"BodyTemp" REAL,
			"RealIntake" REAL,
			"PressurePoint" TEXT,
			"CreatorId" INTEGER,
			"OperatorId" INTEGER,
			"OperateTime" DATETIME,
			"CreateTime" DATETIME,
			"LastModifyTime" DATETIME
		)`,
		`CREATE TABLE ["Treatment_BeforeCheck"] (
			"Id" INTEGER PRIMARY KEY,
			"TenantId" INTEGER,
			"TreatmentId" INTEGER,
			"BeforeSignsId" INTEGER,
			"BeforeSymptomId" INTEGER,
			"OperatorId" INTEGER,
			"OperateTime" DATETIME,
			"MaterialsResult" BOOLEAN,
			"MaterialsMistake" TEXT,
			"ParamResult" BOOLEAN,
			"ParamMistake" TEXT,
			"VascularAccessResult" BOOLEAN,
			"VascularAccessMistake" TEXT,
			"PipelineResult" BOOLEAN,
			"PipelineMistake" TEXT,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME,
			"LastModifyTime" DATETIME
		)`,
		`CREATE TABLE ["Treatment_BeforeSymptom"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "TreatmentId" INTEGER, "Code" TEXT, "Value" TEXT, "OperateTime" DATETIME)`,
		`CREATE TABLE ["Treatment_AfterSymptom"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "TreatmentId" INTEGER, "Code" TEXT, "Value" TEXT, "OperateTime" DATETIME)`,
		`CREATE TABLE ["Treatment_DuringParam"] (
			"Id" INTEGER PRIMARY KEY,
			"TenantId" INTEGER,
			"TreatmentId" INTEGER,
			"OperateTime" DATETIME,
			"VenousPressure" REAL,
			"ArterialPressure" REAL,
			"TMP" REAL,
			"Conductivity" REAL,
			"UFQuantity" REAL,
			"MachineTmp" REAL,
			"BF" REAL,
			"Note" TEXT,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME,
			"LastModifyTime" DATETIME
		)`,
		`CREATE TABLE ["Treatment_DuringSigns"] (
			"Id" INTEGER PRIMARY KEY,
			"TenantId" INTEGER,
			"TreatmentId" INTEGER,
			"OperateTime" DATETIME,
			"SBP" REAL,
			"DBP" REAL,
			"HeartRate" REAL,
			"BodyTemp" REAL,
			"Respiration" REAL,
			"SpO2" REAL,
			"OperatorId" INTEGER,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME,
			"LastModifyTime" DATETIME
		)`,
		`CREATE TABLE ["Treatment_Action"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "TreatmentId" INTEGER, "Name" TEXT, "OperatorId" INTEGER, "OperateTime" DATETIME, "CreatorId" INTEGER, "CreateTime" DATETIME, "LastModifyTime" DATETIME, "Code" TEXT)`,
		`CREATE TABLE ["Auxiliary_JsonData"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "TreatmentId" INTEGER, "Code" TEXT, "CreatorId" INTEGER, "CreateTime" DATETIME, "LastModifyTime" DATETIME, "Value" TEXT)`,
		`CREATE TABLE ["CodeDictionary_CodeDictionarys"] ("Id" INTEGER PRIMARY KEY, "Type" TEXT, "Code" TEXT, "Name" TEXT, "IsDisabled" BOOLEAN)`,
		`CREATE TABLE ["Plan_PatientPrescription"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "TreatmentId" INTEGER, "DialysisMethod" TEXT, "LastModifyTime" DATETIME)`,
		`CREATE TABLE Plan_PatientPlan ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "DialysisMethod" TEXT, "Dialysate" TEXT, "DialysateGroupId" INTEGER, "IsDisabled" BOOLEAN, "CreateTime" DATETIME, "LastModifyTime" DATETIME)`,
		`CREATE TABLE ["Organ_Employee"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "UserId" INTEGER, "Name" TEXT)`,
		`CREATE TABLE "Identity_Users" ("Id" INTEGER PRIMARY KEY, "UserName" TEXT)`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create test schema failed: %v", err)
		}
	}

	// 上机门禁（契约02 三）要求治疗方案完整：为默认测试患者 1001 播一张完整方案
	// （透析液配方非空），使 UpdateStatus(上机) 通过门禁。草稿态拦截的专项用例另行覆盖。
	planTime := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	if err := db.Table("Plan_PatientPlan").Create(map[string]any{
		"Id":               int64(7001),
		"TenantId":         LegacyTenantID,
		"PatientId":        int64(1001),
		"DialysisMethod":   "HD",
		"Dialysate":        "碳酸氢盐透析液",
		"DialysateGroupId": 0,
		"IsDisabled":       false,
		"CreateTime":       planTime,
		"LastModifyTime":   planTime,
	}).Error; err != nil {
		t.Fatalf("seed complete plan failed: %v", err)
	}

	return &TreatmentService{db: db}
}

func mustCreateLegacyTreatment(t *testing.T, db *gorm.DB, row map[string]any) int64 {
	t.Helper()
	baseTime := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	defaults := map[string]any{
		"Id":               int64(1),
		"TenantId":         LegacyTenantID,
		"PatientId":        int64(1001),
		"QueueNo":          "A01",
		"ShiftName":        "上午",
		"Status":           legacyStatusFromApp(models.TreatmentStatusPending),
		"CaseStatus":       "",
		"CreatorId":        int64(9),
		"CreateTime":       baseTime,
		"LastModifyTime":   baseTime,
		"NurseSummary":     "",
		"TreatmentSummary": "",
		"TmrPages":         0,
	}
	for key, value := range row {
		defaults[key] = value
	}
	if err := db.Table(`"Treatment_Treatment"`).Create(defaults).Error; err != nil {
		t.Fatalf("create treatment failed: %v", err)
	}
	return defaults["Id"].(int64)
}

func fetchTreatmentTimes(t *testing.T, db *gorm.DB, id int64) (*time.Time, *time.Time, string) {
	t.Helper()
	var row struct {
		StartTime *time.Time `gorm:"column:StartTime"`
		EndTime   *time.Time `gorm:"column:EndTime"`
		Status    string     `gorm:"column:Status"`
	}
	if err := db.Table(`"Treatment_Treatment"`).
		Select(`"StartTime", "EndTime", "Status"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		Take(&row).Error; err != nil {
		t.Fatalf("fetch treatment failed: %v", err)
	}
	return row.StartTime, row.EndTime, row.Status
}

func TestTreatmentServiceUpdateStatusSetsStartTimeOnlyWhenMissing(t *testing.T) {
	svc := newTestTreatmentService(t)

	t.Run("sets start time when nil", func(t *testing.T) {
		id := mustCreateLegacyTreatment(t, svc.db, map[string]any{"Id": int64(11)})

		if err := svc.UpdateStatus(id, models.TreatmentStatusInProgress); err != nil {
			t.Fatalf("UpdateStatus returned error: %v", err)
		}

		startTime, endTime, status := fetchTreatmentTimes(t, svc.db, id)
		if startTime == nil || startTime.IsZero() {
			t.Fatalf("expected start time to be set")
		}
		if endTime != nil {
			t.Fatalf("expected end time to remain nil, got %v", endTime)
		}
		if status != legacyStatusFromApp(models.TreatmentStatusInProgress) {
			t.Fatalf("expected status updated, got %s", status)
		}
	})

	t.Run("preserves existing start time", func(t *testing.T) {
		fixedStart := time.Date(2026, 4, 24, 9, 30, 0, 0, time.UTC)
		id := mustCreateLegacyTreatment(t, svc.db, map[string]any{
			"Id":        int64(12),
			"Status":    legacyStatusFromApp(models.TreatmentStatusInProgress),
			"StartTime": fixedStart,
		})

		if err := svc.UpdateStatus(id, models.TreatmentStatusInProgress); err != nil {
			t.Fatalf("UpdateStatus returned error: %v", err)
		}

		startTime, _, _ := fetchTreatmentTimes(t, svc.db, id)
		if startTime == nil || !startTime.Equal(fixedStart) {
			t.Fatalf("expected start time %v preserved, got %v", fixedStart, startTime)
		}
	})
}

// 方案完整性门禁（契约02 三）：草稿态方案（透析液配方为空）必须拦截上机，状态不得推进。
func TestTreatmentServiceUpdateStatusBlocksDraftPlan(t *testing.T) {
	svc := newTestTreatmentService(t)

	// 患者 2002 仅有草稿方案：透析液配方为空（模拟建档时生成、尚未补全的方案）。
	planTime := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	if err := svc.db.Table("Plan_PatientPlan").Create(map[string]any{
		"Id":               int64(7002),
		"TenantId":         LegacyTenantID,
		"PatientId":        int64(2002),
		"DialysisMethod":   "HD",
		"Dialysate":        "",
		"DialysateGroupId": 0,
		"IsDisabled":       false,
		"CreateTime":       planTime,
		"LastModifyTime":   planTime,
	}).Error; err != nil {
		t.Fatalf("seed draft plan failed: %v", err)
	}

	id := mustCreateLegacyTreatment(t, svc.db, map[string]any{
		"Id":        int64(21),
		"PatientId": int64(2002),
		"Status":    legacyStatusFromApp(models.TreatmentStatusPending),
	})

	err := svc.UpdateStatus(id, models.TreatmentStatusInProgress)
	if err == nil {
		t.Fatalf("expected 上机 to be blocked by draft plan, got nil error")
	}
	if !strings.Contains(err.Error(), "治疗方案尚未补全") {
		t.Fatalf("expected plan-incomplete error, got: %v", err)
	}

	// 门禁拦截后治疗状态不得推进到上机。
	_, _, status := fetchTreatmentTimes(t, svc.db, id)
	if status == legacyStatusFromApp(models.TreatmentStatusInProgress) {
		t.Fatalf("status must not advance to 上机 when blocked, got %s", status)
	}
}

func TestTreatmentServiceUpdateStatusUsesPrescriptionModeForPlanGate(t *testing.T) {
	svc := newTestTreatmentService(t)

	older := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	patientID := int64(3003)
	if err := svc.db.Table("Plan_PatientPlan").Create([]map[string]any{
		{
			"Id":               int64(7003),
			"TenantId":         LegacyTenantID,
			"PatientId":        patientID,
			"DialysisMethod":   "HDF",
			"Dialysate":        "碳酸氢盐透析液",
			"DialysateGroupId": 0,
			"IsDisabled":       false,
			"CreateTime":       older,
			"LastModifyTime":   older,
		},
		{
			"Id":               int64(7004),
			"TenantId":         LegacyTenantID,
			"PatientId":        patientID,
			"DialysisMethod":   "HD",
			"Dialysate":        "",
			"DialysateGroupId": 0,
			"IsDisabled":       false,
			"CreateTime":       newer,
			"LastModifyTime":   newer,
		},
	}).Error; err != nil {
		t.Fatalf("seed multi-mode plans failed: %v", err)
	}

	id := mustCreateLegacyTreatment(t, svc.db, map[string]any{
		"Id":        int64(24),
		"PatientId": patientID,
		"Status":    legacyStatusFromApp(models.TreatmentStatusPending),
	})
	if err := svc.db.Table(`"Plan_PatientPrescription"`).Create(map[string]any{
		"Id":             int64(8001),
		"TenantId":       LegacyTenantID,
		"TreatmentId":    id,
		"DialysisMethod": "HDF",
		"LastModifyTime": newer,
	}).Error; err != nil {
		t.Fatalf("seed prescription mode failed: %v", err)
	}

	if err := svc.UpdateStatus(id, models.TreatmentStatusInProgress); err != nil {
		t.Fatalf("expected HDF complete plan to pass gate, got: %v", err)
	}
	_, _, status := fetchTreatmentTimes(t, svc.db, id)
	if status != legacyStatusFromApp(models.TreatmentStatusInProgress) {
		t.Fatalf("expected treatment to advance to in-progress, got %s", status)
	}
}

func TestTreatmentServiceUpdateStatusSetsEndTimeOnComplete(t *testing.T) {
	svc := newTestTreatmentService(t)
	start := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)

	t.Run("sets end time when completing via status update", func(t *testing.T) {
		id := mustCreateLegacyTreatment(t, svc.db, map[string]any{
			"Id":        int64(21),
			"Status":    legacyStatusFromApp(models.TreatmentStatusInProgress),
			"StartTime": start,
		})

		if err := svc.UpdateStatus(id, models.TreatmentStatusCompleted); err != nil {
			t.Fatalf("UpdateStatus returned error: %v", err)
		}

		startTime, endTime, status := fetchTreatmentTimes(t, svc.db, id)
		if startTime == nil || !startTime.Equal(start) {
			t.Fatalf("expected start time preserved, got %v", startTime)
		}
		if endTime == nil {
			t.Fatalf("expected end time to be set when completing, got nil")
		}
		if status != legacyStatusFromApp(models.TreatmentStatusCompleted) {
			t.Fatalf("expected status updated, got %s", status)
		}
	})

	t.Run("preserves existing end time", func(t *testing.T) {
		fixedEnd := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
		id := mustCreateLegacyTreatment(t, svc.db, map[string]any{
			"Id":        int64(22),
			"Status":    legacyStatusFromApp(models.TreatmentStatusCompleted),
			"StartTime": start,
			"EndTime":   fixedEnd,
		})

		if err := svc.UpdateStatus(id, models.TreatmentStatusCompleted); err != nil {
			t.Fatalf("UpdateStatus returned error: %v", err)
		}

		_, endTime, _ := fetchTreatmentTimes(t, svc.db, id)
		if endTime == nil || !endTime.Equal(fixedEnd) {
			t.Fatalf("expected end time %v preserved, got %v", fixedEnd, endTime)
		}
	})
}

func TestTreatmentServiceSaveAfterSignsSetsExplicitEndTime(t *testing.T) {
	svc := newTestTreatmentService(t)
	start := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 24, 12, 30, 0, 0, time.UTC)
	id := mustCreateLegacyTreatment(t, svc.db, map[string]any{
		"Id":        int64(23),
		"Status":    legacyStatusFromApp(models.TreatmentStatusInProgress),
		"StartTime": start,
	})

	if _, err := svc.SaveAfterSigns(id, TreatmentAfterSignsRequest{StartTime: &start, EndTime: &end}, 99); err != nil {
		t.Fatalf("SaveAfterSigns returned error: %v", err)
	}

	startTime, endTime, _ := fetchTreatmentTimes(t, svc.db, id)
	if startTime == nil || !startTime.Equal(start) {
		t.Fatalf("expected start time %v preserved, got %v", start, startTime)
	}
	if endTime == nil || !endTime.Equal(end) {
		t.Fatalf("expected end time %v, got %v", end, endTime)
	}
}

func TestTreatmentServiceUpdatePrefersExplicitStartAndEndTime(t *testing.T) {
	svc := newTestTreatmentService(t)
	id := mustCreateLegacyTreatment(t, svc.db, map[string]any{"Id": int64(31)})
	explicitStart := time.Date(2026, 4, 24, 9, 5, 0, 0, time.UTC)
	explicitEnd := time.Date(2026, 4, 24, 13, 15, 0, 0, time.UTC)
	status := models.TreatmentStatusCompleted
	notes := "透后录入结束时间"

	updated, err := svc.Update(id, TreatmentUpdateRequest{
		Status:    &status,
		StartTime: &explicitStart,
		EndTime:   &explicitEnd,
		Notes:     &notes,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if updated.StartTime == nil || !updated.StartTime.Equal(explicitStart) {
		t.Fatalf("expected response start time %v, got %v", explicitStart, updated.StartTime)
	}
	if updated.EndTime == nil || !updated.EndTime.Equal(explicitEnd) {
		t.Fatalf("expected response end time %v, got %v", explicitEnd, updated.EndTime)
	}

	startTime, endTime, rawStatus := fetchTreatmentTimes(t, svc.db, id)
	if startTime == nil || !startTime.Equal(explicitStart) {
		t.Fatalf("expected stored start time %v, got %v", explicitStart, startTime)
	}
	if endTime == nil || !endTime.Equal(explicitEnd) {
		t.Fatalf("expected stored end time %v, got %v", explicitEnd, endTime)
	}
	if rawStatus != legacyStatusFromApp(models.TreatmentStatusCompleted) {
		t.Fatalf("expected completed raw status, got %s", rawStatus)
	}
}

func TestAppStatusFromLegacyHardCodeBeforeDict(t *testing.T) {
	startTime := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	dict := map[int]string{10: "签到", 30: "透中监测", 0: "待签到", 50: "取消治疗", 60: "已结束"}

	// raw "10" + StartTime 非空 + EndTime 空 → InProgress（不因字典名"签到"误判 Pending）
	if got := appStatusFromLegacy("10", &startTime, nil, dict); got != models.TreatmentStatusInProgress {
		t.Fatalf(`"10" + start!=nil + end==nil: expected InProgress, got %d`, got)
	}
	// raw "30" + StartTime 非空 + EndTime 空 → InProgress
	if got := appStatusFromLegacy("30", &startTime, nil, dict); got != models.TreatmentStatusInProgress {
		t.Fatalf(`"30" + start!=nil + end==nil: expected InProgress, got %d`, got)
	}
	// raw "0" + 无 start/end → Pending
	if got := appStatusFromLegacy("0", nil, nil, dict); got != models.TreatmentStatusPending {
		t.Fatalf(`"0" + start==nil + end==nil: expected Pending, got %d`, got)
	}
	// raw "50" → Cancelled
	if got := appStatusFromLegacy("50", nil, nil, dict); got != models.TreatmentStatusCancelled {
		t.Fatalf(`"50": expected Cancelled, got %d`, got)
	}
	// raw "60" → Completed
	if got := appStatusFromLegacy("60", nil, nil, dict); got != models.TreatmentStatusCompleted {
		t.Fatalf(`"60": expected Completed, got %d`, got)
	}
	// raw "10" + endTime 非空 + startTime 非空 → Completed（时间判据 fallback）
	endTime := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	if got := appStatusFromLegacy("10", &startTime, &endTime, dict); got != models.TreatmentStatusCompleted {
		t.Fatalf(`"10" + start!=nil + end!=nil: expected Completed, got %d`, got)
	}
	// raw "0" + start 非空 + end 空 → InProgress
	if got := appStatusFromLegacy("0", &startTime, nil, dict); got != models.TreatmentStatusInProgress {
		t.Fatalf(`"0" + start!=nil + end==nil: expected InProgress, got %d`, got)
	}
}

func TestTreatmentServiceSecondCheckIndependence(t *testing.T) {
	svc := newTestTreatmentService(t)
	id := mustCreateLegacyTreatment(t, svc.db, map[string]any{"Id": int64(31)})

	opA := int64(100)
	opB := int64(200)

	// 首次核对：操作人 A（落 Treatment_BeforeCheck.OperatorId=100）。
	if _, err := svc.SaveFirstCheck(id, TreatmentFirstCheckRequest{OperatorID: &opA}, opA); err != nil {
		t.Fatalf("SaveFirstCheck failed: %v", err)
	}

	// 二次核对同一人 A → 必须被服务端拒绝（双人核对独立性，不依赖前端过滤）。
	if _, err := svc.SaveSecondCheck(id, TreatmentSecondCheckRequest{OperatorID: &opA}, opA); err == nil {
		t.Fatalf("expected same-operator second check to be rejected")
	}

	// 二次核对换人 B → 通过。
	if _, err := svc.SaveSecondCheck(id, TreatmentSecondCheckRequest{OperatorID: &opB}, opB); err != nil {
		t.Fatalf("different-operator second check should succeed, got: %v", err)
	}
}
