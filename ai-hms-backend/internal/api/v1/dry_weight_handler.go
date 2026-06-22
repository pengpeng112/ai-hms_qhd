package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type DryWeightHandler struct{ svc *services.DryWeightService }

func RegisterDryWeightRoutes(rg *gin.RouterGroup) {
	h := &DryWeightHandler{svc: services.NewDryWeightService()}
	rg.POST("/patients/:id/dry-weight-assessments", h.Assess)
	rg.GET("/patients/:id/dry-weight-assessments", h.ListAssessments)
	rg.POST("/patients/:id/dry-weight/confirm", h.Confirm)
	rg.GET("/patients/:id/dry-weight", h.Current)
}

func (h *DryWeightHandler) Assess(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	var raw struct {
		AssessType   string   `json:"assessType"`
		Phase        string   `json:"phase"`
		SBP          *int     `json:"sbp"`
		DBP          *int     `json:"dbp"`
		HeartRate    *int     `json:"heartRate"`
		Edema        bool     `json:"edema"`
		Palpitation  bool     `json:"palpitation"`
		HeartFailure bool     `json:"heartFailure"`
		Cramp        bool     `json:"cramp"`
		CTR          *float64 `json:"ctr"`
		ACTR         *float64 `json:"actr"`
		BIAOH        *float64 `json:"biaOh"`
		BIATBW       *float64 `json:"biaTbw"`
		BIAECW       *float64 `json:"biaEcw"`
		PostWeight   *float64 `json:"postWeight"`
		TargetWeight *float64 `json:"targetWeight"`
		Decision     string   `json:"decision"`
		AdjustKg     *float64 `json:"adjustKg"`
		RNaSetting   *float64 `json:"rnaSetting"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	userID := middleware.GetUserID(c)
	userName := middleware.GetUsername(c)
	dwa, err := h.svc.Assess(pid, services.DwAssessInput{
		AssessType: raw.AssessType, Phase: raw.Phase,
		SBP: raw.SBP, DBP: raw.DBP, HeartRate: raw.HeartRate,
		Edema: raw.Edema, Palpitation: raw.Palpitation, HeartFailure: raw.HeartFailure, Cramp: raw.Cramp,
		CTR: raw.CTR, ACTR: raw.ACTR, BIAOH: raw.BIAOH, BIATBW: raw.BIATBW, BIAECW: raw.BIAECW,
		PostWeight: raw.PostWeight, TargetWeight: raw.TargetWeight,
		Decision: raw.Decision, AdjustKg: raw.AdjustKg, RNaSetting: raw.RNaSetting,
		AssessorID: userID, AssessorName: userName,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, dwa)
}

func (h *DryWeightHandler) ListAssessments(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	rows, err := h.svc.ListAssessments(pid)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rows)
}

func (h *DryWeightHandler) Confirm(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	var raw struct {
		DryWeight     float64  `json:"dryWeight"`
		Phase         string   `json:"phase"`
		ACTR          *float64 `json:"actr"`
		CTR           *float64 `json:"ctr"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	userID := middleware.GetUserID(c)
	userName := middleware.GetUsername(c)
	result, err := h.svc.Confirm(pid, services.DwConfirmInput{
		DryWeight: raw.DryWeight, Phase: raw.Phase,
		ACTR: raw.ACTR, CTR: raw.CTR,
		ConfirmedBy: userID, ConfirmedName: userName,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DryWeightHandler) Current(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	data, err := h.svc.Current(pid)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, data)
}
