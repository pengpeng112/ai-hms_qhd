package v1

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/his_oracle"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type HisExamSyncHandler struct {
	service *services.HisExamReportSyncService
}

func NewHisExamSyncHandler(oracleCfg config.HisOracleConfig, tenantID int64) *HisExamSyncHandler {
	hcfg := his_oracle.Config{
		Host:     oracleCfg.Host,
		Port:     oracleCfg.Port,
		Service:  oracleCfg.Service,
		Username: oracleCfg.Username,
		Password: oracleCfg.Password,
	}
	return &HisExamSyncHandler{
		service: services.NewHisExamReportSyncService(hcfg, tenantID),
	}
}

func RegisterHisExamSyncRoutes(r *gin.RouterGroup, oracleCfg config.HisOracleConfig, tenantID int64) {
	handler := NewHisExamSyncHandler(oracleCfg, tenantID)
	patients := r.Group("/patients")
	{
		patients.POST("/:id/exam-reports/sync", handler.SyncPatientExamReports)
	}
}

func (h *HisExamSyncHandler) SyncPatientExamReports(c *gin.Context) {
	patientID := strings.TrimSpace(c.Param("id"))
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}
	legacyID, err := parseLegacyPatientID(patientID)
	if err != nil {
		response.BadRequest(c, "无效的患者ID格式")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	result, syncErr := h.service.SyncPatientExamReports(ctx, legacyID)
	if syncErr != nil {
		response.InternalError(c, syncErr.Error())
		return
	}
	response.Success(c, gin.H{
		"created": result.Created,
		"updated": result.Updated,
		"skipped": result.Skipped,
		"failed":  result.Failed,
	})
}

func parseLegacyPatientID(patientID string) (int64, error) {
	patientID = strings.TrimSpace(patientID)
	if patientID == "" {
		return 0, errors.New("empty patient id")
	}
	var id int64
	for _, ch := range patientID {
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid patient id format: must be numeric")
		}
		id = id*10 + int64(ch-'0')
	}
	if id <= 0 {
		return 0, errors.New("invalid patient id")
	}
	return id, nil
}
