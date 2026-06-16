package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// SignHandler 统一电子签名留痕（处方/方案/小结共用）。
type SignHandler struct {
	service *services.SignService
}

func NewSignHandler() *SignHandler {
	return &SignHandler{service: services.NewSignService()}
}

// Create 通用签发：为方案/小结等对象写一条签名留痕（处方签发走 prescriptions/:pid/sign，另含 ConfirmTime）。
// POST /api/v1/sign-records  body: {targetType, targetId}
func (h *SignHandler) Create(c *gin.Context) {
	var req struct {
		TargetType string `json:"targetType"`
		TargetID   string `json:"targetId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	rec, err := h.service.Sign(middleware.GetTenantID(c), req.TargetType, req.TargetID, middleware.GetUserID(c), "")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rec)
}

// List 查询某对象的签名留痕（审计/展示）
// GET /api/v1/sign-records?targetType=prescription&targetId=123
func (h *SignHandler) List(c *gin.Context) {
	targetType := c.Query("targetType")
	targetID := c.Query("targetId")
	if targetType == "" || targetID == "" {
		response.BadRequest(c, "targetType 和 targetId 不能为空")
		return
	}
	rows, err := h.service.ListSigns(middleware.GetTenantID(c), targetType, targetID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rows)
}
