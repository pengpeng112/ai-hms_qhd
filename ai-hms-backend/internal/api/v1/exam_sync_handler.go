package v1

import (
	"errors"
	"strings"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ExamReportSyncHandler 检查报告同步控制器
type ExamReportSyncHandler struct {
	service *services.ExamReportSyncService
}

// NewExamReportSyncHandler 创建检查报告同步控制器
func NewExamReportSyncHandler(cfg config.HdisConfig) *ExamReportSyncHandler {
	return &ExamReportSyncHandler{
		service: services.NewExamReportSyncService(cfg),
	}
}

// SyncPatientExamReports 触发患者检查报告同步
// POST /api/v1/patients/:id/exam-reports/sync
func (h *ExamReportSyncHandler) SyncPatientExamReports(c *gin.Context) {
	patientID := strings.TrimSpace(c.Param("id"))
	result, err := h.service.SyncPatientExamReports(patientID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSyncPatientIDRequired):
			response.BadRequest(c, err.Error())
			return
		case errors.Is(err, services.ErrSyncNotConfigured):
			response.BadRequest(c, "HDIS 对接未配置，请先在系统设置 > Integration 中完成配置")
			return
		case errors.Is(err, services.ErrSyncPatientBasicNotFound):
			response.NotFound(c, "患者基础档案不存在，请先完善患者基本信息")
			return
		case errors.Is(err, services.ErrSyncPatientHDISIDMissing):
			response.BadRequest(c, "患者缺少 HDIS 对应 ID，请先在患者基本信息中填写 hdisPatientId")
			return
		default:
			response.InternalError(c, err.Error())
			return
		}
	}

	response.Success(c, result)
}

// RegisterExamSyncRoutes 注册检查报告同步路由
func RegisterExamSyncRoutes(r *gin.RouterGroup, cfg config.HdisConfig) {
	handler := NewExamReportSyncHandler(cfg)

	patients := r.Group("/patients")
	{
		patients.POST("/:id/exam-reports/sync", handler.SyncPatientExamReports)
	}
}
