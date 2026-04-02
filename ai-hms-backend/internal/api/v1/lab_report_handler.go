package v1

import (
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// LabReportHandler 检验报告控制器
type LabReportHandler struct {
	service *services.LabReportService
}

// NewLabReportHandler 创建检验报告控制器
func NewLabReportHandler() *LabReportHandler {
	return &LabReportHandler{
		service: services.NewLabReportService(),
	}
}

// ListByPatient 获取患者检验报告列表
func (h *LabReportHandler) ListByPatient(c *gin.Context) {
	patientID := c.Param("id")
	if strings.TrimSpace(patientID) == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.LabReportListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的查询参数")
		return
	}

	result, err := h.service.ListByPatient(patientID, req)
	if err != nil {
		if isLabReportBadRequestError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Paginated(c, result.Items, result.Page, result.PageSize, result.Total)
}

// Create 创建检验报告
func (h *LabReportHandler) Create(c *gin.Context) {
	patientID := c.Param("id")
	if strings.TrimSpace(patientID) == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.LabReportCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	report, err := h.service.Create(patientID, req)
	if err != nil {
		if err.Error() == "patient not found" {
			response.NotFound(c, "患者不存在")
			return
		}
		if isLabReportBadRequestError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, report)
}

// Update 更新检验报告
func (h *LabReportHandler) Update(c *gin.Context) {
	reportID := c.Param("id")
	if strings.TrimSpace(reportID) == "" {
		response.BadRequest(c, "报告ID不能为空")
		return
	}

	var req services.LabReportUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	report, err := h.service.Update(reportID, req)
	if err != nil {
		if err.Error() == "lab report not found" {
			response.NotFound(c, "检验报告不存在")
			return
		}
		if isLabReportBadRequestError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, report)
}

// Delete 删除检验报告
func (h *LabReportHandler) Delete(c *gin.Context) {
	reportID := c.Param("id")
	if strings.TrimSpace(reportID) == "" {
		response.BadRequest(c, "报告ID不能为空")
		return
	}

	if err := h.service.Delete(reportID); err != nil {
		if err.Error() == "lab report not found" {
			response.NotFound(c, "检验报告不存在")
			return
		}
		if isLabReportBadRequestError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// RegisterLabReportRoutes 注册检验报告路由
func RegisterLabReportRoutes(r *gin.RouterGroup) {
	handler := NewLabReportHandler()

	patients := r.Group("/patients")
	{
		patients.GET("/:id/lab-reports", handler.ListByPatient)
		patients.POST("/:id/lab-reports", handler.Create)
	}

	r.PUT("/lab-reports/:id", handler.Update)
	r.DELETE("/lab-reports/:id", handler.Delete)
}

func isLabReportBadRequestError(err error) bool {
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
	if strings.Contains(msg, "must be one of") {
		return true
	}
	if strings.Contains(msg, "cannot be empty") {
		return true
	}
	if strings.Contains(msg, "unsupported time format") {
		return true
	}
	return false
}
