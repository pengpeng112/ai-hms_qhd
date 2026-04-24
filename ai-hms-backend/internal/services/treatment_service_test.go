package services

import (
	"fmt"
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
		`CREATE TABLE ["Treatment_Action"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "TreatmentId" INTEGER, "Name" TEXT, "OperatorId" INTEGER, "OperateTime" DATETIME, "Code" TEXT)`,
		`CREATE TABLE ["Auxiliary_JsonData"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "PatientId" INTEGER, "TreatmentId" INTEGER, "Code" TEXT, "CreatorId" INTEGER, "CreateTime" DATETIME, "LastModifyTime" DATETIME, "Value" TEXT)`,
		`CREATE TABLE ["CodeDictionary_CodeDictionarys"] ("Id" INTEGER PRIMARY KEY, "Type" TEXT, "Code" TEXT, "Name" TEXT, "IsDisabled" BOOLEAN)`,
		`CREATE TABLE ["Plan_PatientPrescription"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "TreatmentId" INTEGER, "DialysisMethod" TEXT, "LastModifyTime" DATETIME)`,
		`CREATE TABLE ["Organ_Employee"] ("Id" INTEGER PRIMARY KEY, "TenantId" INTEGER, "UserId" INTEGER, "Name" TEXT)`,
		`CREATE TABLE "Identity_Users" ("Id" INTEGER PRIMARY KEY, "UserName" TEXT)`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create test schema failed: %v", err)
		}
	}

	return &TreatmentService{db: db}
}

func mustCreateLegacyTreatment(t *testing.T, db *gorm.DB, row map[string]any) int64 {
	t.Helper()
	baseTime := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	defaults := map[string]any{
		"Id":               int64(1),
		"TenantId":         legacyTenantID,
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
		Where(`"Id" = ? AND "TenantId" = ?`, id, legacyTenantID).
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

func TestTreatmentServiceUpdateStatusDoesNotSetEndTime(t *testing.T) {
	svc := newTestTreatmentService(t)
	start := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)

	t.Run("does not set end time when completing via status update", func(t *testing.T) {
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
		if endTime != nil {
			t.Fatalf("expected end time to remain nil, got %v", endTime)
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
