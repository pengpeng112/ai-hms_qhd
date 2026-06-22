package v1

import (
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type MedicationHandler struct{ svc *services.MedicationService }

func RegisterMedicationRoutes(rg *gin.RouterGroup) {
	h := &MedicationHandler{svc: services.NewMedicationService()}
	rg.POST("/medication-admins", h.RecordAdmin)
	rg.GET("/medication-admins", h.List)
	rg.POST("/medication-admins/:id/second-check", h.SecondCheck)
	rg.GET("/patients/:id/medication-suggestions", h.Suggestions)
	rg.GET("/medication-default-doses", h.DefaultDoses)
}

func (h *MedicationHandler) RecordAdmin(c *gin.Context) {
	var raw struct {
		PatientID        int64  `json:"patientId"`
		OrderID          int64  `json:"orderId"`
		TreatmentID      int64  `json:"treatmentId"`
		DrugName         string `json:"drugName"`
		Category         string `json:"category"`
		Dose             string `json:"dose"`
		Route            string `json:"route"`
		Timing           string `json:"timing"`
		Note             string `json:"note"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	userID := middleware.GetUserID(c)
	userName := middleware.GetUsername(c)
	rec, err := h.svc.RecordAdmin(raw.PatientID, services.MaRecordInput{
		OrderID: raw.OrderID, TreatmentID: raw.TreatmentID, DrugName: raw.DrugName,
		Category: raw.Category, Dose: raw.Dose, Route: raw.Route, Timing: raw.Timing,
		AdministeredBy: userID, AdministeredName: userName, Note: raw.Note,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

func (h *MedicationHandler) List(c *gin.Context) {
	var treatmentID, patientID, orderID *int64
	if v := c.Query("treatmentId"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效treatmentId")
			return
		}
		treatmentID = &n
	}
	if v := c.Query("patientId"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效patientId")
			return
		}
		patientID = &n
	}
	if v := c.Query("orderId"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效orderId")
			return
		}
		orderID = &n
	}
	rows, err := h.svc.List(treatmentID, patientID, orderID)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rows)
}

func (h *MedicationHandler) SecondCheck(c *gin.Context) {
	var raw struct {
		CheckerID   string `json:"checkerId"`
		CheckerName string `json:"checkerName"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	if raw.CheckerID == "" {
		raw.CheckerID = middleware.GetUserID(c)
		raw.CheckerName = middleware.GetUsername(c)
	}
	rec, err := h.svc.SecondCheck(c.Param("id"), raw.CheckerID, raw.CheckerName)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *MedicationHandler) Suggestions(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	sugs, err := h.svc.Suggestions(pid)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, sugs)
}

func (h *MedicationHandler) DefaultDoses(c *gin.Context) {
	doses, err := h.svc.DefaultDoses()
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, doses)
}
