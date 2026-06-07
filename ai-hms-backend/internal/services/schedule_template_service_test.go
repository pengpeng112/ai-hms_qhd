package services

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func int64Ptr(v int64) *int64 { return &v }

func newTestTemplateDB(t *testing.T) *gorm.DB {
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
	t.Cleanup(func() { _ = sqlDB.Close() })

	stmts := []string{
		`CREATE TABLE "Schedule_ScheduleTemplate" (
			"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"TenantId" INTEGER NOT NULL,
			"Name" TEXT NOT NULL,
			"Scope" TEXT,
			"WardId" INTEGER,
			"IsActive" INTEGER DEFAULT 1,
			"Version" INTEGER DEFAULT 1,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME DEFAULT CURRENT_TIMESTAMP,
			"LastModifyTime" DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE "Schedule_ScheduleTemplateItem" (
			"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"TenantId" INTEGER NOT NULL,
			"TemplateId" INTEGER NOT NULL,
			"PatientId" INTEGER NOT NULL,
			"ZoneTag" TEXT NOT NULL DEFAULT 'A',
			"WardId" INTEGER,
			"ShiftId" INTEGER,
			"FreqPattern" INTEGER NOT NULL DEFAULT 10,
			"FixedHdBedId" INTEGER,
			"FixedHdfBedId" INTEGER,
			"HdfEnabled" INTEGER DEFAULT 0,
			"HdfWeekday" INTEGER,
			"HdfWeekParity" INTEGER,
			"TemplateVersion" INTEGER NOT NULL DEFAULT 1,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME DEFAULT CURRENT_TIMESTAMP,
			"LastModifyTime" DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE "Schedule_PatientShift" (
			"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"TenantId" INTEGER NOT NULL,
			"PatientId" INTEGER NOT NULL,
			"TreatmentTime" DATETIME NOT NULL,
			"ShiftId" INTEGER NOT NULL,
			"BedId" INTEGER,
			"WardId" INTEGER,
			"PatientPlanId" INTEGER,
			"ShiftTiming" INTEGER,
			"Status" INTEGER,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME DEFAULT CURRENT_TIMESTAMP,
			"LastModifyTime" DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE "Schedule_PatientShiftExt" (
			"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"TenantId" INTEGER NOT NULL,
			"PatientShiftId" INTEGER NOT NULL,
			"DialysisMode" TEXT NOT NULL DEFAULT 'HD',
			"SourceType" INTEGER NOT NULL DEFAULT 10,
			"RecordForm" INTEGER NOT NULL DEFAULT 10,
			"Confirm1At" DATETIME,
			"Confirm2At" DATETIME,
			"Confirm3At" DATETIME,
			"Confirm1By" INTEGER,
			"Confirm2By" INTEGER,
			"Confirm3By" INTEGER,
			"IsBorrowedSlot" INTEGER DEFAULT 0,
			"BorrowedFromShiftId" INTEGER,
			"IsLocked" INTEGER DEFAULT 0,
			"CancelReason" TEXT,
			"SourceTemplateItemId" INTEGER,
			"SourceTemplateVersion" INTEGER,
			"RuleStatus" INTEGER NOT NULL DEFAULT 10,
			"ApprovedBy" INTEGER,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME DEFAULT CURRENT_TIMESTAMP,
			"LastModifyTime" DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE "Schedule_BedMachineExt" (
			"Id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"TenantId" INTEGER NOT NULL,
			"BedId" INTEGER NOT NULL,
			"MachineCode" TEXT,
			"MachineType" TEXT NOT NULL DEFAULT 'HD',
			"SupportedModes" TEXT NOT NULL DEFAULT 'HD',
			"PositionIndex" INTEGER NOT NULL DEFAULT 0,
			"IsDisabled" INTEGER DEFAULT 0,
			"LegacyBedName" TEXT,
			"Note" TEXT,
			"CreatorId" INTEGER,
			"CreateTime" DATETIME DEFAULT CURRENT_TIMESTAMP,
			"LastModifyTime" DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("create table failed: %v\nSQL: %s", err, s)
		}
	}

	// 插入测试机器配置
	db.Exec(`INSERT INTO "Schedule_BedMachineExt" ("TenantId","BedId","MachineType","SupportedModes") VALUES (3,10,'HD','HD'), (3,11,'HDF','HD,HDF,HF')`)

	return db
}

type svcFactory struct {
	newFn func(db *gorm.DB) *ScheduleTemplateService
}

// Test save then list
func TestSaveAndListTemplates(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	req := ScheduleTemplateSaveRequest{
		Name:  "默认模板",
		Scope: "A",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), FreqPattern: 10, HdfEnabled: true},
			{PatientId: 101, ZoneTag: "A", ShiftId: int64Ptr(1), FreqPattern: 20},
		},
	}

	result, err := svc.SaveTemplate(3, 1, req)
	if err != nil {
		t.Fatalf("SaveTemplate failed: %v", err)
	}

	if result.Template.Version != 1 {
		t.Errorf("新模板 Version 应为 1, 实际=%d", result.Template.Version)
	}
	if result.ItemCount != 2 {
		t.Errorf("ItemCount 应为 2, 实际=%d", result.ItemCount)
	}

	// List
	tmpls, err := svc.ListTemplates(3, nil)
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}
	if len(tmpls) != 1 {
		t.Fatalf("应有1个模板, 实际=%d", len(tmpls))
	}
	if len(tmpls[0].Items) != 2 {
		t.Errorf("模板应有2项, 实际=%d", len(tmpls[0].Items))
	}
}

// Test update version
func TestSaveTemplate_UpdateVersion(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	req1 := ScheduleTemplateSaveRequest{
		Name: "模板V1",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1)},
		},
	}
	r1, _ := svc.SaveTemplate(3, 1, req1)
	if r1.Template.Version != 1 {
		t.Fatalf("V1 version应为1, 实际=%d", r1.Template.Version)
	}

	req2 := ScheduleTemplateSaveRequest{
		Id:   r1.Template.Id,
		Name: "模板V2",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(2)},
		},
	}
	r2, err := svc.SaveTemplate(3, 1, req2)
	if err != nil {
		t.Fatalf("SaveTemplate V2 failed: %v", err)
	}
	if r2.Template.Version != 2 {
		t.Errorf("V2 version应为2, 实际=%d", r2.Template.Version)
	}
	if r2.Items[0].TemplateVersion != 2 {
		t.Errorf("templateItem.TemplateVersion 应为2, 实际=%d", r2.Items[0].TemplateVersion)
	}
}

// Test scope mismatch
func TestSaveTemplate_ScopeMismatch(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	req := ScheduleTemplateSaveRequest{
		Name:  "模板",
		Scope: "A",
		Items: []ScheduleTemplateItemRequest{{PatientId: 100, ZoneTag: "B", ShiftId: int64Ptr(1)}},
	}
	_, err := svc.SaveTemplate(3, 1, req)
	if err == nil {
		t.Fatal("Scope与ZoneTag不匹配应报错")
	}
}

// Test apply template
func TestApplyTemplate_CreatesDraft(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	// Save template
	req := ScheduleTemplateSaveRequest{
		Name:  "默认模板",
		Items: []ScheduleTemplateItemRequest{{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), FreqPattern: 10, FixedHdBedId: int64Ptr(10)}},
	}
	saveResult, err := svc.SaveTemplate(3, 1, req)
	if err != nil {
		t.Fatalf("SaveTemplate failed: %v", err)
	}

	// Apply
	applyReq := ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-01-10",
	}
	applyResult, err := svc.ApplyTemplate(3, 1, applyReq)
	if err != nil {
		t.Fatalf("ApplyTemplate failed: %v", err)
	}
	if applyResult.Count != 1 {
		t.Fatalf("应创建1条草稿, 实际=%d", applyResult.Count)
	}

	// Verify PatientShift Status=10
	var ps models.PatientShift
	if err := db.First(&ps, applyResult.CreatedShifts[0]).Error; err != nil {
		t.Fatalf("查 PatientShift 失败: %v", err)
	}
	if ps.Status != 10 {
		t.Errorf("Status 应为 10, 实际=%d", ps.Status)
	}
	if ps.Status == 20 {
		t.Error("不应生成 Status=20 已确认")
	}
	if ps.Status == 60 {
		t.Error("不应生成 Status=60")
	}

	// Verify PatientShiftExt
	var ext models.PatientShiftExt
	if err := db.Where(`"PatientShiftId" = ?`, ps.Id).First(&ext).Error; err != nil {
		t.Fatalf("查 PatientShiftExt 失败: %v", err)
	}
	if ext.RuleStatus != 10 {
		t.Errorf("RuleStatus 应为 10, 实际=%d", ext.RuleStatus)
	}
	if ext.SourceTemplateItemId == nil || *ext.SourceTemplateItemId != saveResult.Items[0].Id {
		t.Errorf("SourceTemplateItemId 不正确")
	}
}

// Test conflict detection
func TestApplyTemplate_PatientConflict(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	req := ScheduleTemplateSaveRequest{
		Name:  "默认模板",
		Items: []ScheduleTemplateItemRequest{{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), FixedHdBedId: int64Ptr(10)}},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	// Create pre-existing shift
	db.Create(&models.PatientShift{
		TenantId:     3,
		PatientId:    modeltypes.LegacyID(100),
		ScheduleDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		ShiftId:      1,
		Status:       10,
		CreatorId:    1,
	})

	_, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-01-10",
	})
	if err == nil {
		t.Fatal("同日同班冲突应报错")
	}
}

func TestApplyTemplate_BedConflict(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	bedID := int64(10)
	req := ScheduleTemplateSaveRequest{
		Name:  "默认模板",
		Items: []ScheduleTemplateItemRequest{{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), FixedHdBedId: &bedID}},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	// Pre-existing shift on same bed
	db.Create(&models.PatientShift{
		TenantId:     3,
		PatientId:    modeltypes.LegacyID(200),
		ScheduleDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		ShiftId:      1,
		BedId:        &bedID,
		Status:       20,
		CreatorId:    1,
	})

	_, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-01-10",
	})
	if err == nil {
		t.Fatal("同床冲突应报错")
	}
}

func TestApplyTemplate_InactiveLegacyStatusDoesNotConflict(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	bedID := int64(10)
	req := ScheduleTemplateSaveRequest{
		Name:  "默认模板",
		Items: []ScheduleTemplateItemRequest{{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), FixedHdBedId: &bedID}},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	// 老状态 40/50/60 分别表示用户取消、排班取消、转出人员，不应占用患者或床位。
	for i, status := range []int{40, 50, 60} {
		day := 10 + i
		db.Create(&models.PatientShift{
			TenantId:     3,
			PatientId:    modeltypes.LegacyID(100),
			ScheduleDate: time.Date(2026, 1, day, 0, 0, 0, 0, time.UTC),
			ShiftId:      1,
			BedId:        &bedID,
			Status:       status,
			CreatorId:    1,
		})

		result, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
			TemplateId: saveResult.Template.Id,
			TargetDate: fmt.Sprintf("2026-01-%02d", day),
		})
		if err != nil {
			t.Fatalf("Status=%d 不应触发冲突: %v", status, err)
		}
		if result.Count != 1 {
			t.Fatalf("Status=%d 应生成 1 条草稿, got %d", status, result.Count)
		}
	}
}

func TestApplyTemplate_ModeNotSupported(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	bedID := int64(10) // HD only bed
	req := ScheduleTemplateSaveRequest{
		Name: "默认模板",
		Items: []ScheduleTemplateItemRequest{{
			PatientId:    100,
			ZoneTag:      "A",
			ShiftId:      int64Ptr(1),
			HdfEnabled:   true,
			FixedHdBedId: &bedID,
		}},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	_, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-01-10",
	})
	if err == nil {
		t.Fatal("HD-only bed 不支持 HDF 应报错")
	}
}

func TestApplyTemplate_TransactionRollback(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	// 先查当前 PatientShift 总数
	var countBefore int64
	db.Model(&models.PatientShift{}).Count(&countBefore)

	req := ScheduleTemplateSaveRequest{
		Name: "默认模板",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), FixedHdBedId: int64Ptr(10)},
			// 第二项缺少 ShiftId，会使事务失败
			{PatientId: 101, ZoneTag: "A", ShiftId: nil},
		},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	_, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-01-10",
	})
	if err == nil {
		t.Fatal("缺 ShiftId 应报错")
	}

	// 确认没有半写入
	var countAfter int64
	db.Model(&models.PatientShift{}).Count(&countAfter)
	if countAfter != countBefore {
		t.Errorf("事务应回滚, 不应留下数据, before=%d, after=%d", countBefore, countAfter)
	}
}

// Test SaveTemplate does NOT write Status=60
func TestSaveTemplate_NoStatus60(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	req := ScheduleTemplateSaveRequest{
		Name:  "模板",
		Items: []ScheduleTemplateItemRequest{{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1)}},
	}
	_, err := svc.SaveTemplate(3, 1, req)
	if err != nil {
		t.Fatalf("SaveTemplate failed: %v", err)
	}

	var count60 int64
	db.Model(&models.PatientShift{}).Where(`"Status" = ?`, 60).Count(&count60)
	if count60 > 0 {
		t.Errorf("SaveTemplate 不应写入 Status=60, 实际=%d", count60)
	}
}

// Test ApplyTemplate respects WardId filter
func TestApplyTemplate_WardIdFilter(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	ward1 := int64(10)
	ward2 := int64(20)
	req := ScheduleTemplateSaveRequest{
		Name: "多病区模板",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), WardId: &ward1, FixedHdBedId: int64Ptr(10)},
			{PatientId: 200, ZoneTag: "A", ShiftId: int64Ptr(1), WardId: &ward2, FixedHdBedId: int64Ptr(11)},
		},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	applyReq := ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-02-01",
		WardId:     &ward1,
	}
	applyResult, err := svc.ApplyTemplate(3, 1, applyReq)
	if err != nil {
		t.Fatalf("ApplyTemplate failed: %v", err)
	}
	if applyResult.Count != 1 {
		t.Fatalf("只应生成 ward1 的1条草稿, 实际=%d", applyResult.Count)
	}
	var ps models.PatientShift
	if err := db.First(&ps, applyResult.CreatedShifts[0]).Error; err != nil {
		t.Fatalf("查 PatientShift 失败: %v", err)
	}
	if ps.WardId == nil || *ps.WardId != ward1 {
		t.Errorf("生成的排班 WardId 应为 %d, 实际=%v", ward1, ps.WardId)
	}
}

// Test ApplyTemplate excludes WardId IS NULL items when template Scope is not ALL
func TestApplyTemplate_WardIdFilter_NonAllScopeNullExcluded(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	ward1 := int64(10)
	req := ScheduleTemplateSaveRequest{
		Name:  "非全局模板",
		Scope: "A",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), WardId: &ward1, FixedHdBedId: int64Ptr(10)},
			{PatientId: 200, ZoneTag: "A", ShiftId: int64Ptr(1), WardId: nil, FixedHdBedId: int64Ptr(11)},
		},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	applyResult, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-03-01",
		WardId:     &ward1,
	})
	if err != nil {
		t.Fatalf("ApplyTemplate failed: %v", err)
	}
	if applyResult.Count != 1 {
		t.Fatalf("非 ALL Scope 的 WardId IS NULL 项应被排除, 只应生成 WardId=10 的1条, 实际=%d", applyResult.Count)
	}
	var ps models.PatientShift
	if err := db.First(&ps, applyResult.CreatedShifts[0]).Error; err != nil {
		t.Fatalf("查 PatientShift 失败: %v", err)
	}
	if ps.PatientId != modeltypes.LegacyID(100) {
		t.Errorf("应该是 PatientId=100 的排班, 实际=%d", ps.PatientId)
	}
}

// Test ApplyTemplate with Scope=ALL fills req.WardId into WardId IS NULL items
func TestApplyTemplate_GlobalTemplateWardIdFill(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	req := ScheduleTemplateSaveRequest{
		Name:  "全局模板",
		Scope: "ALL",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), FixedHdBedId: int64Ptr(10)},
		},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	ward10 := int64(10)
	applyResult, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-04-01",
		WardId:     &ward10,
	})
	if err != nil {
		t.Fatalf("ApplyTemplate failed: %v", err)
	}
	if applyResult.Count != 1 {
		t.Fatalf("应生成1条草稿, 实际=%d", applyResult.Count)
	}
	var ps models.PatientShift
	if err := db.First(&ps, applyResult.CreatedShifts[0]).Error; err != nil {
		t.Fatalf("查 PatientShift 失败: %v", err)
	}
	if ps.WardId == nil || *ps.WardId != ward10 {
		t.Errorf("全局模板项 WardId 应为 %d, 实际=%v", ward10, ps.WardId)
	}
}

// Test ApplyTemplate fails when BedMachineExt is missing
func TestApplyTemplate_NoBedMachineExt(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	bed999 := int64(999)
	req := ScheduleTemplateSaveRequest{
		Name: "测试模板",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1), FixedHdBedId: &bed999},
		},
	}
	saveResult, _ := svc.SaveTemplate(3, 1, req)

	var countBefore int64
	db.Model(&models.PatientShift{}).Count(&countBefore)

	_, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
		TemplateId: saveResult.Template.Id,
		TargetDate: "2026-03-01",
	})
	if err == nil {
		t.Fatal("缺少 BedMachineExt 应报错")
	}

	var countAfter int64
	db.Model(&models.PatientShift{}).Count(&countAfter)
	if countAfter != countBefore {
		t.Errorf("事务应回滚, before=%d, after=%d", countBefore, countAfter)
	}
}

// Test ApplyTemplate with invalid date returns TemplateBusinessError
func TestApplyTemplate_InvalidDate(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	_, err := svc.ApplyTemplate(3, 1, ScheduleTemplateApplyRequest{
		TemplateId: 1,
		TargetDate: "not-a-date",
	})
	if err == nil {
		t.Fatal("非法日期应报错")
	}
	if !IsTemplateBusinessError(err) {
		t.Errorf("非法日期应为 TemplateBusinessError, 实际=%T", err)
	}
}

// Test ListTemplates returns global templates when filtering by wardId
func TestListTemplates_GlobalTemplate(t *testing.T) {
	db := newTestTemplateDB(t)
	svc := &ScheduleTemplateService{db: db}

	// 全局模板 (Scope=ALL, WardId IS NULL)
	saveAll, _ := svc.SaveTemplate(3, 1, ScheduleTemplateSaveRequest{
		Name:  "全局模板",
		Scope: "ALL",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 100, ZoneTag: "A", ShiftId: int64Ptr(1)},
		},
	})

	// 病区模板 (WardId=10)
	ward10 := int64(10)
	saveWard, _ := svc.SaveTemplate(3, 1, ScheduleTemplateSaveRequest{
		Name:   "病区10模板",
		WardId: &ward10,
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 200, ZoneTag: "A", ShiftId: int64Ptr(1)},
		},
	})

	// 非全局 Zone 模板 (Scope="A", WardId IS NULL) — 应按病区查询时不返回
	saveZone, _ := svc.SaveTemplate(3, 1, ScheduleTemplateSaveRequest{
		Name:  "A区模板(无wardId)",
		Scope: "A",
		Items: []ScheduleTemplateItemRequest{
			{PatientId: 300, ZoneTag: "A", ShiftId: int64Ptr(1)},
		},
	})

	_ = saveAll
	_ = saveWard
	_ = saveZone

	tmpls, err := svc.ListTemplates(3, &ward10)
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	foundGlobal := false
	foundWard := false
	foundZone := false
	for _, t := range tmpls {
		if t.Template.Name == "全局模板" {
			foundGlobal = true
		}
		if t.Template.Name == "病区10模板" {
			foundWard = true
		}
		if t.Template.Name == "A区模板(无wardId)" {
			foundZone = true
		}
	}
	if !foundGlobal {
		t.Error("应返回全局模板(Scope=ALL)")
	}
	if !foundWard {
		t.Error("应返回病区10模板")
	}
	if foundZone {
		t.Error("不应返回非ALL Scope且WardId IS NULL的模板")
	}
}

func TestErrPatientShiftDuplicateIsSentinel(t *testing.T) {
	if ErrPatientShiftDuplicate == nil {
		t.Fatal("ErrPatientShiftDuplicate 未定义")
	}
	if ErrPatientShiftDuplicate.Error() != "患者或床位同日同班已有排班" {
		t.Errorf("sentinel 消息不匹配: %s", ErrPatientShiftDuplicate.Error())
	}
	if !errors.Is(ErrPatientShiftDuplicate, ErrPatientShiftDuplicate) {
		t.Error("errors.Is 应匹配自身")
	}
	if errors.Is(fmt.Errorf("wrapped: %w", ErrPatientShiftDuplicate), ErrPatientShiftDuplicate) {
		// wrapped error should be detectable
	} else {
		t.Error("errors.Is 应可检测 wrapping 后的 sentinel")
	}
}
