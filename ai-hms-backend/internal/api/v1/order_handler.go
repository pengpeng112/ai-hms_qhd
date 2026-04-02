package v1

import (
	"errors"
	"net/http"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// OrderHandler 医嘱控制器
type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{service: services.NewOrderService()}
}

func (h *OrderHandler) List(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}
	req.PatientID = patientID

	orders, err := h.service.List(req)
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, orders)
}

func (h *OrderHandler) Create(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	order, err := h.service.Create(patientID, middleware.GetUserID(c), middleware.GetUsername(c), req)
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, order)
}

func (h *OrderHandler) Update(c *gin.Context) {
	patientID, orderID := c.Param("id"), c.Param("oid")
	if patientID == "" || orderID == "" {
		response.BadRequest(c, "患者ID和医嘱ID不能为空")
		return
	}

	var req services.OrderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	order, err := h.service.Update(patientID, orderID, req)
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, order)
}

func (h *OrderHandler) Revise(c *gin.Context) {
	patientID, orderID := c.Param("id"), c.Param("oid")
	if patientID == "" || orderID == "" {
		response.BadRequest(c, "患者ID和医嘱ID不能为空")
		return
	}

	var req services.OrderReviseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	order, err := h.service.Revise(patientID, orderID, middleware.GetUserID(c), middleware.GetUsername(c), req)
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, order)
}

func (h *OrderHandler) Copy(c *gin.Context) {
	patientID, orderID := c.Param("id"), c.Param("oid")
	if patientID == "" || orderID == "" {
		response.BadRequest(c, "患者ID和医嘱ID不能为空")
		return
	}

	order, err := h.service.Copy(patientID, orderID, middleware.GetUserID(c), middleware.GetUsername(c))
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, order)
}

func (h *OrderHandler) Stop(c *gin.Context) {
	patientID, orderID := c.Param("id"), c.Param("oid")
	if patientID == "" || orderID == "" {
		response.BadRequest(c, "患者ID和医嘱ID不能为空")
		return
	}

	var req services.OrderStopRequest
	_ = c.ShouldBindJSON(&req)

	orders, err := h.service.Stop(patientID, orderID, req.StopReason, req.StopDate)
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, orders)
}

func (h *OrderHandler) CreateFromTemplate(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.CreateFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	orders, err := h.service.CreateFromTemplate(patientID, middleware.GetUserID(c), middleware.GetUsername(c), req)
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, orders)
}

func (h *OrderHandler) Group(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.OrderGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	orders, err := h.service.Group(patientID, req.OrderIDs)
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, orders)
}

func (h *OrderHandler) Ungroup(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.BadRequest(c, "患者ID不能为空")
		return
	}

	var req services.OrderUngroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	orders, err := h.service.Ungroup(patientID, req.OrderIDs)
	if err != nil {
		h.respondOrderError(c, err)
		return
	}
	response.Success(c, orders)
}

func (h *OrderHandler) respondOrderError(c *gin.Context, err error) {
	var svcErr *services.OrderServiceError
	if !errors.As(err, &svcErr) {
		response.InternalError(c, err.Error())
		return
	}

	switch svcErr.Status {
	case http.StatusBadRequest:
		response.BadRequest(c, svcErr.Message)
	case http.StatusNotFound:
		response.NotFound(c, svcErr.Message)
	case http.StatusConflict:
		response.Error(c, http.StatusConflict, "CONFLICT", svcErr.Message)
	default:
		response.InternalError(c, svcErr.Message)
	}
}
