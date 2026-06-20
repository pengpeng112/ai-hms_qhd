package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/sijms/go-ora/v2"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/his_oracle"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/services"
)

func main() {
	jobCode := flag.String("job", "", "sync job code (his_exam_report / his_patient_archive)")
	once := flag.Bool("once", false, "run once and exit")
	flag.Parse()

	if *jobCode == "" {
		log.Fatal("--job is required")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	services.LegacyTenantID = cfg.LegacyTenantID

	if err := database.Initialize(&cfg.Database, &cfg.Logging, cfg.ParameterizedQueries); err != nil {
		log.Fatalf("database connect failed: %v", err)
	}
	defer database.Close()

	log.Printf("[his-sync] job=%s", *jobCode)

	jobSvc := services.NewSyncJobService()
	job, err := jobSvc.GetJob(*jobCode)
	if err != nil {
		log.Fatalf("sync job not found: %v", err)
	}
	if !job.Enabled && !*once {
		log.Fatalf("job %s is disabled; enable it first or use --once", *jobCode)
	}

	oracleCfg := his_oracle.Config{
		Host:     cfg.HisOracle.Host,
		Port:     cfg.HisOracle.Port,
		Service:  cfg.HisOracle.Service,
		Username: cfg.HisOracle.Username,
		Password: cfg.HisOracle.Password,
	}

	switch *jobCode {
	case models.SyncJobCodeExamReport:
		runExamReportSync(jobSvc, cfg.LegacyTenantID, oracleCfg, job)
	default:
		log.Fatalf("unknown job code: %s", *jobCode)
	}

	if *once {
		os.Exit(0)
	}
}

func runExamReportSync(
	jobSvc *services.SyncJobService,
	tenantID int64,
	oracleCfg his_oracle.Config,
	job *models.SyncJobConfig,
) {
	ctx := context.Background()

	cursorTime := time.Time{}
	if job.CursorValue != nil && *job.CursorValue != "" {
		if t, err := time.Parse(time.RFC3339, *job.CursorValue); err == nil {
			cursorTime = t
		}
	}

	cursorBefore := cursorTime.Format(time.RFC3339)
	run := &models.SyncJobRun{
		ID:           fmt.Sprintf("run_%s_%d", job.JobCode, time.Now().UnixMilli()),
		JobCode:      job.JobCode,
		SourceSystem: job.SourceSystem,
		SyncType:     job.SyncType,
		Status:       models.SyncJobStatusRunning,
		StartedAt:    time.Now(),
		CursorBefore: &cursorBefore,
	}
	if err := jobSvc.CreateRun(run); err != nil {
		log.Printf("[his-sync] create run record failed: %v", err)
		return
	}

	syncSvc := services.NewHisExamReportSyncService(oracleCfg, tenantID)
	defer syncSvc.CloseOracleClient()

	// 预匹配：先尝试用 ID_NO 建立所有可匹配的患者映射
	prematchCount, prematchErr := syncSvc.PrematchAll(ctx)
	if prematchErr != nil {
		log.Printf("[his-sync] prematch warning: %v", prematchErr)
	} else {
		log.Printf("[his-sync] prematch: %d mappings created", prematchCount)
	}

	result, syncErr := syncSvc.SyncBatch(ctx, cursorTime, job.BatchSize)
	if syncErr != nil {
		log.Printf("[his-sync] sync batch error: %v", syncErr)
		if finishErr := jobSvc.FinishRun(run.ID, models.SyncJobStatusFailed, nil, "", cursorBefore, syncErr.Error()); finishErr != nil {
			log.Printf("[his-sync] finish run record failed: %v", finishErr)
		}
		return
	}

	status := models.SyncJobStatusSuccess
	counts := map[string]int{
		"created": result.Created, "updated": result.Updated,
		"skipped": result.Skipped, "failed": result.Failed,
		"fetched": result.Created + result.Updated + result.Skipped + result.Failed,
	}
	errMsg := ""
	if result.Failed > 0 {
		status = models.SyncJobStatusPartial
		errMsg = fmt.Sprintf("%d records failed", result.Failed)
	}

	log.Printf("[his-sync] result: fetched=%d created=%d updated=%d skipped=%d failed=%d status=%s",
		counts["fetched"], result.Created, result.Updated, result.Skipped, result.Failed, status)
	if len(result.Errors) > 0 {
		for i, e := range result.Errors {
			if i < 10 {
				log.Printf("[his-sync] error[%d]: %s", i, e)
			}
		}
		if len(result.Errors) > 10 {
			log.Printf("[his-sync] ... and %d more errors", len(result.Errors)-10)
		}
	}

	cursorAfter := time.Now().Format(time.RFC3339)
	if result.MaxCursorTime != nil {
		cursorAfter = result.MaxCursorTime.Format(time.RFC3339)
	}
	if err := jobSvc.FinishRun(run.ID, status, counts, cursorAfter, cursorBefore, errMsg); err != nil {
		log.Printf("[his-sync] finish run record failed: %v", err)
	}
}
