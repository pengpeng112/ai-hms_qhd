package v1

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/actrs"
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	maxImageUploadSize = 20 << 20
	maxDicomUploadSize = 100 << 20
)

var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".dcm": true, ".dicom": true, ".ima": true,
}

type ActrHandler struct {
	svc *services.ActrService
}

func RegisterActrRoutes(rg *gin.RouterGroup, cfg config.ActrsConfig, tenantID int64) {
	h := &ActrHandler{svc: services.NewActrService(actrs.Config{
		BaseURL:    cfg.BaseURL,
		Username:   cfg.Username,
		Password:   cfg.Password,
		TimeoutSec: cfg.TimeoutSec,
	}, cfg.Enabled, tenantID)}
	rg.GET("/actr/status", h.Status)
	rg.GET("/patients/:id/actr", h.History)
	rg.POST("/patients/:id/actr/analyze", h.Analyze)
	rg.POST("/patients/:id/actr/adopt", h.Adopt)
	rg.PATCH("/patients/:id/actr/:recordId/correction", h.Correct)
}

func (h *ActrHandler) Status(c *gin.Context) {
	response.Success(c, h.svc.Status(c.Request.Context()))
}

func (h *ActrHandler) History(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	rows, err := h.svc.History(pid)
	if err != nil {
		response.InternalErrorSafe(c)
		return
	}
	response.Success(c, rows)
}

func (h *ActrHandler) Analyze(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxDicomUploadSize))
	fh, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			response.Error(c, 413, "FILE_TOO_LARGE", fmt.Sprintf("文件过大，最大允许 %dMB", maxDicomUploadSize>>20))
			return
		}
		response.BadRequest(c, "缺少胸片文件")
		return
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedExtensions[ext] {
		response.BadRequest(c, "不支持的文件类型，仅支持 jpg/jpeg/png/dcm/dicom/ima")
		return
	}
	maxUploadSize := maxImageUploadSize
	if ext == ".dcm" || ext == ".dicom" || ext == ".ima" {
		maxUploadSize = maxDicomUploadSize
	}
	if fh.Size > int64(maxUploadSize) {
		response.Error(c, 413, "FILE_TOO_LARGE", fmt.Sprintf("文件过大，最大允许 %dMB", maxUploadSize>>20))
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.BadRequest(c, "文件无法读取")
		return
	}
	defer f.Close()
	rec, err := h.svc.Analyze(c.Request.Context(), pid, fh.Filename, f)
	if err != nil {
		response.Error(c, 502, "ACTRS_ANALYZE_FAILED", err.Error())
		return
	}
	response.SuccessCreated(c, rec)
}

type adoptReq struct {
	PrescriptionID string   `json:"prescriptionId"`
	ActrRecordID   string   `json:"actrRecordId"`
	DryWeight      *float64 `json:"dryWeight"`
	UFQuantity     *float64 `json:"ufQuantity"`
}

func (h *ActrHandler) Adopt(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	var req adoptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	rxID, err := strconv.ParseInt(req.PrescriptionID, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效处方ID")
		return
	}
	adoptedBy := middleware.GetUserID(c)
	if adoptedBy == "" {
		adoptedBy = middleware.GetUsername(c)
	}
	if err := h.svc.AdoptToPrescription(c.Request.Context(), pid, rxID, req.ActrRecordID, req.DryWeight, req.UFQuantity, adoptedBy); err != nil {
		if err == services.ErrPrescriptionSigned {
			response.BadRequest(c, err.Error())
		} else {
			response.Error(c, 502, "ACTRS_ADOPT_FAILED", err.Error())
		}
		return
	}
	response.Success(c, gin.H{"ok": true})
}

type correctReq struct {
	Value float64 `json:"value"`
	Notes string  `json:"notes,omitempty"`
}

func (h *ActrHandler) Correct(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效患者ID")
		return
	}
	var req correctReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求体无效")
		return
	}
	correctedBy := middleware.GetUserID(c)
	if correctedBy == "" {
		correctedBy = middleware.GetUsername(c)
	}
	rec, err := h.svc.Correct(c.Request.Context(), pid, c.Param("recordId"), correctedBy, req.Value)
	if err != nil {
		response.Error(c, 502, "ACTRS_CORRECT_FAILED", err.Error())
		return
	}
	response.Success(c, rec)
}
