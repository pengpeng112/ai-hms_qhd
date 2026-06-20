package v1

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/his_oracle"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SyncJobHandler struct {
	service   *services.SyncJobService
	oracleCfg his_oracle.Config
	tenantID  int64
}

func NewSyncJobHandler(oracleCfg config.HisOracleConfig, tenantID int64) *SyncJobHandler {
	return &SyncJobHandler{
		service: services.NewSyncJobService(),
		oracleCfg: his_oracle.Config{
			Host:     oracleCfg.Host,
			Port:     oracleCfg.Port,
			Service:  oracleCfg.Service,
			Username: oracleCfg.Username,
			Password: oracleCfg.Password,
		},
		tenantID: tenantID,
	}
}

func (h *SyncJobHandler) ListJobs(c *gin.Context) {
	jobs, err := h.service.ListJobs()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, jobs)
}

func (h *SyncJobHandler) GetJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		response.BadRequest(c, "任务代码不能为空")
		return
	}
	job, err := h.service.GetJob(code)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, job)
}

func (h *SyncJobHandler) UpdateJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		response.BadRequest(c, "任务代码不能为空")
		return
	}
	var req services.UpdateSyncJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}
	job, err := h.service.UpdateJob(code, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, job)
}

func (h *SyncJobHandler) GetJobRuns(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		response.BadRequest(c, "任务代码不能为空")
		return
	}
	runs, err := h.service.GetJobRuns(code, 50)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, runs)
}

func (h *SyncJobHandler) RunJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		response.BadRequest(c, "任务代码不能为空")
		return
	}
	job, err := h.service.GetJob(code)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	if !job.Enabled {
		response.BadRequest(c, "任务未启用，请先启用后再运行")
		return
	}

	run := &models.SyncJobRun{
		ID:           fmt.Sprintf("run_%s_%d", code, time.Now().UnixNano()),
		JobCode:      job.JobCode,
		SourceSystem: job.SourceSystem,
		SyncType:     job.SyncType,
		Status:       models.SyncJobStatusRunning,
		StartedAt:    time.Now(),
	}
	if err := h.service.CreateRun(run); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	go func() {
		ctx := context.Background()
		syncSvc := services.NewHisExamReportSyncService(h.oracleCfg, h.tenantID)
		defer syncSvc.CloseOracleClient()

		var cursorTime time.Time
		if job.CursorValue != nil && *job.CursorValue != "" {
			if t, err := time.Parse(time.RFC3339, *job.CursorValue); err == nil {
				cursorTime = t
			}
		}
		cursorBefore := cursorTime.Format(time.RFC3339)

		result, syncErr := syncSvc.SyncBatch(ctx, cursorTime, job.BatchSize)
		errMsg := ""
		var status string
		var cursorAfter string
		var counts map[string]int

		if syncErr != nil {
			status = models.SyncJobStatusFailed
			errMsg = syncErr.Error()
			cursorAfter = time.Now().Format(time.RFC3339)
			counts = map[string]int{"created": 0, "updated": 0, "skipped": 0, "failed": 0, "fetched": 0}
		} else {
			counts = map[string]int{
				"created": result.Created, "updated": result.Updated,
				"skipped": result.Skipped, "failed": result.Failed,
				"fetched": result.Created + result.Updated + result.Skipped + result.Failed,
			}
			status = models.SyncJobStatusSuccess
			if result.Failed > 0 {
				status = models.SyncJobStatusPartial
			}
			cursorAfter = time.Now().Format(time.RFC3339)
			if result.MaxCursorTime != nil {
				cursorAfter = result.MaxCursorTime.Format(time.RFC3339)
			}
		}

		if err := h.service.FinishRun(run.ID, status, counts, cursorAfter, cursorBefore, errMsg); err != nil {
			log.Printf("[sync-job] finish run failed: %v", err)
		}
		log.Printf("[sync-job] %s done: fetched=%d created=%d status=%s", code, counts["fetched"], counts["created"], status)
	}()

	response.Success(c, gin.H{"runId": run.ID, "status": "started"})
}

func RegisterSyncJobRoutes(r *gin.RouterGroup, oracleCfg config.HisOracleConfig, tenantID int64) {
	handler := NewSyncJobHandler(oracleCfg, tenantID)
	sync := r.Group("/sync")
	{
		sync.GET("/jobs", handler.ListJobs)
		sync.GET("/jobs/:code", handler.GetJob)
		sync.PUT("/jobs/:code", handler.UpdateJob)
		sync.GET("/jobs/:code/runs", handler.GetJobRuns)
		sync.POST("/jobs/:code/run", handler.RunJob)
	}
}
