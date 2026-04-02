package v1

import (
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ExamReportHandler 检查报告控制器
type ExamReportHandler struct {
	service *services.ExamReportService
}

// NewExamReportHandler 创建检查报告控制器
func NewExamReportHandler() *ExamReportHandler {
	return &ExamReportHandler{
		service: services.NewExamReportService(),
	}
}

// ListByPatient 获取患者检查报告列表
// GET /api/v1/patients/:id/exam-reports
func (h *ExamReportHandler) ListByPatient(c *gin.Context) {
	patientID := c.Param("id")
	if strings.TrimSpace(patientID) == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.ExamReportListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的查询参数")
		return
	}

	result, err := h.service.ListByPatient(patientID, req)
	if err != nil {
		if isExamReportBadRequestError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Paginated(c, result.Items, result.Page, result.PageSize, result.Total)
}

// RegisterExamReportRoutes 注册检查报告路由
func RegisterExamReportRoutes(r *gin.RouterGroup) {
	handler := NewExamReportHandler()

	patients := r.Group("/patients")
	{
		patients.GET("/:id/exam-reports", handler.ListByPatient)
	}
}

func isExamReportBadRequestError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "required") {
		return true
	}
	if strings.Contains(msg, "invalid") {
		return true
	}
	if strings.Contains(msg, "unsupported time format") {
		return true
	}
	return false
}
