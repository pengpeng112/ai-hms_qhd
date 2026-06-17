package v1

import (
	"strconv"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type QCHandler struct {
	service *services.QCService
}

func NewQCHandler() *QCHandler {
	return &QCHandler{service: services.NewQCService()}
}

func qcMonth(c *gin.Context) (int, int) {
	if t, err := time.Parse("2006-01", c.Query("month")); err == nil {
		return t.Year(), int(t.Month())
	}
	now := time.Now()
	return now.Year(), int(now.Month())
}

func (h *QCHandler) Doctors(c *gin.Context) {
	year, month := qcMonth(c)
	res, err := h.service.ScoreDoctors(year, month)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, res)
}

func (h *QCHandler) DoctorDetail(c *gin.Context) {
	doctorID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的医生ID")
		return
	}
	year, month := qcMonth(c)
	doctor, patients, err := h.service.ScoreDoctorDetail(doctorID, year, month)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"doctor": doctor, "patients": patients})
}
