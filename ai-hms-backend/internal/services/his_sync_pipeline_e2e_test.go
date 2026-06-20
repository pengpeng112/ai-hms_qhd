package services

// 端到端冒烟（服务层 + 真实 DB 引擎 SQLite）：HIS 同步中心「落库管线」。
// 覆盖 b31161f 新增、当前零测试的 his_oracle / 同步服务的本地 DB 一侧：
//   1) upsertExamReport —— 幂等 upsert（核心去重逻辑）
//   2) ExternalPatientMappingService —— 解析/身份证号自动匹配（患者对人）
//   3) SyncJobService —— 跑批运行记录生命周期 + 游标推进
//
// 离线不可测部分：HIS Oracle 实连查询（client.QueryExamReports 等需真实 Oracle），
// 这部分在测试报告 §2/§7 已标注，需在接通生产 Oracle 后用 his-sync --once 灰度验证。

import (
	"context"
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newHisSyncTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.ExamReport{},
		&models.ExamReportItem{},
		&models.ExternalPatientMapping{},
		&models.SyncJobConfig{},
		&models.SyncJobRun{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func strptr(s string) *string { return &s }

// ---- 1) upsertExamReport 幂等 ----

func TestHisSync_UpsertExamReport_Idempotent(t *testing.T) {
	db := newHisSyncTestDB(t)
	s := &HisExamReportSyncService{db: db, tenantID: LegacyTenantID}

	examDate := time.Now().Add(-24 * time.Hour)
	r1 := models.ExamReport{
		PatientID:        "1001",
		Title:            "胸部正位片",
		Conclusion:       "心影增大，CTR 0.55",
		Department:       "放射科",
		ExamDate:         &examDate,
		ExternalReportID: strptr("HIS-R-0001"),
		SourceSystem:     models.SourceSystemHISOracleExam,
	}

	created, id1, err := s.upsertExamReport(&r1)
	if err != nil {
		t.Fatalf("首次 upsert 失败: %v", err)
	}
	if !created {
		t.Errorf("首次应为新建 created=true")
	}
	if id1 == "" {
		t.Errorf("应返回报告 ID")
	}

	// 同一 external_report_id 再次同步：内容有更新，但不应新增行
	r2 := r1
	r2.ID = ""
	r2.Title = "胸部正位片（修订）"
	r2.Conclusion = "心影增大，CTR 0.58"
	created2, id2, err := s.upsertExamReport(&r2)
	if err != nil {
		t.Fatalf("二次 upsert 失败: %v", err)
	}
	if created2 {
		t.Errorf("重复 external_report_id 应为更新 created=false")
	}
	if id2 != id1 {
		t.Errorf("重复同步应命中同一行: id1=%s id2=%s", id1, id2)
	}

	var n int64
	db.Model(&models.ExamReport{}).Count(&n)
	if n != 1 {
		t.Errorf("幂等：应只有 1 行报告，得 %d", n)
	}
	var got models.ExamReport
	db.First(&got, "id = ?", id1)
	if got.Title != "胸部正位片（修订）" || got.Conclusion != "心影增大，CTR 0.58" {
		t.Errorf("更新应覆盖标题/结论，得 title=%q conclusion=%q", got.Title, got.Conclusion)
	}
	if got.SyncedAt == nil {
		t.Errorf("更新应回填 synced_at")
	}
}

func TestHisSync_UpsertExamReport_RequiresExternalID(t *testing.T) {
	db := newHisSyncTestDB(t)
	s := &HisExamReportSyncService{db: db, tenantID: LegacyTenantID}

	r := models.ExamReport{PatientID: "1001", Title: "无外部ID"}
	if _, _, err := s.upsertExamReport(&r); err == nil {
		t.Fatalf("缺少 external_report_id 应报错")
	}
}

// ---- 2) ExternalPatientMappingService 解析与自动匹配 ----

func TestHisSync_ResolveLegacyPatientID(t *testing.T) {
	db := newHisSyncTestDB(t)
	svc := &ExternalPatientMappingService{db: db}

	// 已确认映射
	if err := svc.CreateMapping(&models.ExternalPatientMapping{
		ID: "epm_HIS_ORACLE_P9", TenantID: LegacyTenantID, LegacyPatientID: 1001,
		ExternalSystem: models.ExternalSystemHISOracle, ExternalPatientID: "P9",
		MatchStatus: models.MatchStatusConfirmed,
	}); err != nil {
		t.Fatalf("create mapping: %v", err)
	}
	// 候选（未确认）映射，不应被解析
	if err := svc.CreateMapping(&models.ExternalPatientMapping{
		ID: "epm_HIS_ORACLE_P8", TenantID: LegacyTenantID, LegacyPatientID: 1002,
		ExternalSystem: models.ExternalSystemHISOracle, ExternalPatientID: "P8",
		MatchStatus: models.MatchStatusCandidate,
	}); err != nil {
		t.Fatalf("create candidate: %v", err)
	}

	if pid, err := svc.ResolveLegacyPatientID(models.ExternalSystemHISOracle, "P9", nil); err != nil || pid != 1001 {
		t.Errorf("已确认映射应解析为 1001，得 pid=%d err=%v", pid, err)
	}
	if _, err := svc.ResolveLegacyPatientID(models.ExternalSystemHISOracle, "P8", nil); err == nil {
		t.Errorf("候选映射不应被解析为确认")
	}
	if _, err := svc.ResolveLegacyPatientID(models.ExternalSystemHISOracle, "NOPE", nil); err == nil {
		t.Errorf("不存在映射应报错")
	}
}

func TestHisSync_AutoMatchByIDNo(t *testing.T) {
	db := newHisSyncTestDB(t)
	// 旧库患者表（标准名建表即可，AutoMatchByIDNo 用普通表名 + 原始 SQL 引号）
	if err := db.Exec(`CREATE TABLE Register_IDInfomation (
		"Id" INTEGER PRIMARY KEY, "PatientId" INTEGER, "IDNo" TEXT, "IsDisabled" BOOLEAN)`).Error; err != nil {
		t.Fatalf("create Register_IDInfomation: %v", err)
	}
	if err := db.Exec(`CREATE TABLE Register_PatientInfomation (
		"Id" INTEGER PRIMARY KEY, "Name" TEXT, "TenantId" INTEGER)`).Error; err != nil {
		t.Fatalf("create Register_PatientInfomation: %v", err)
	}
	db.Exec(`INSERT INTO Register_PatientInfomation ("Id","Name","TenantId") VALUES (2002,'王患者',?)`, LegacyTenantID)
	db.Exec(`INSERT INTO Register_IDInfomation ("Id","PatientId","IDNo","IsDisabled") VALUES (1,2002,'X123456',0)`)

	svc := &ExternalPatientMappingService{db: db}
	m, err := svc.AutoMatchByIDNo(models.ExternalSystemHISOracle, "PX", strptr("X123456"), LegacyTenantID)
	if err != nil {
		t.Fatalf("AutoMatchByIDNo 失败: %v", err)
	}
	if m == nil || m.LegacyPatientID != 2002 || m.MatchStatus != models.MatchStatusConfirmed {
		t.Fatalf("应按身份证号匹配到本地患者 2002 并确认，得 %+v", m)
	}
	// 持久化后可被解析
	if pid, err := svc.ResolveLegacyPatientID(models.ExternalSystemHISOracle, "PX", nil); err != nil || pid != 2002 {
		t.Errorf("自动匹配应已落库并可解析为 2002，得 pid=%d err=%v", pid, err)
	}
	// 查不到身份证号 → 返回 nil，不报错
	if m2, err := svc.AutoMatchByIDNo(models.ExternalSystemHISOracle, "PY", strptr("NOTEXIST"), LegacyTenantID); err != nil || m2 != nil {
		t.Errorf("身份证号无匹配应返回 (nil,nil)，得 m=%+v err=%v", m2, err)
	}
}

// ---- 3) SyncJobService 跑批生命周期 + 游标推进 ----

func TestHisSync_JobRunLifecycle(t *testing.T) {
	db := newHisSyncTestDB(t)
	svc := &SyncJobService{db: db}

	oldCursor := "2026-06-19T00:00:00Z"
	if err := db.Create(&models.SyncJobConfig{
		ID: "job1", JobCode: models.SyncJobCodeExamReport,
		SourceSystem: "HIS_ORACLE", SyncType: models.SyncTypeExamReport,
		Enabled: true, BatchSize: 500, CursorType: models.CursorTypeTime,
		CursorValue: &oldCursor, OverwritePolicy: models.OverwritePolicyFillEmpty,
	}).Error; err != nil {
		t.Fatalf("seed job config: %v", err)
	}

	if job, err := svc.GetJob(models.SyncJobCodeExamReport); err != nil || job == nil {
		t.Fatalf("GetJob 失败: %v", err)
	}

	before := "2026-06-19T00:00:00Z"
	run := &models.SyncJobRun{
		ID: "run1", JobCode: models.SyncJobCodeExamReport, SourceSystem: "HIS_ORACLE",
		SyncType: models.SyncTypeExamReport, Status: models.SyncJobStatusRunning,
		StartedAt: time.Now().Add(-2 * time.Second), CursorBefore: &before,
	}
	if err := svc.CreateRun(run); err != nil {
		t.Fatalf("CreateRun 失败: %v", err)
	}

	newCursor := "2026-06-20T08:00:00Z"
	counts := map[string]int{"created": 3, "updated": 1, "skipped": 2, "failed": 0, "fetched": 6}
	if err := svc.FinishRun("run1", models.SyncJobStatusSuccess, counts, newCursor, before, ""); err != nil {
		t.Fatalf("FinishRun 失败: %v", err)
	}

	var got models.SyncJobRun
	db.First(&got, "id = ?", "run1")
	if got.Status != models.SyncJobStatusSuccess {
		t.Errorf("状态应为 success，得 %s", got.Status)
	}
	if got.CreatedCount != 3 || got.UpdatedCount != 1 || got.SkippedCount != 2 || got.FetchedCount != 6 {
		t.Errorf("计数落库错误: %+v", got)
	}
	if got.FinishedAt == nil || got.DurationMs == nil || *got.DurationMs <= 0 {
		t.Errorf("应落 finished_at 与正向 duration_ms，得 finished=%v dur=%v", got.FinishedAt, got.DurationMs)
	}

	// 成功跑批应推进 config 游标 + 落 last_run_at
	var cfg models.SyncJobConfig
	db.First(&cfg, "job_code = ?", models.SyncJobCodeExamReport)
	if cfg.CursorValue == nil || *cfg.CursorValue != newCursor {
		t.Errorf("成功跑批应推进游标到 %s，得 %v", newCursor, cfg.CursorValue)
	}
	if cfg.LastRunAt == nil {
		t.Errorf("成功跑批应落 last_run_at")
	}

	_ = context.Background()
}
