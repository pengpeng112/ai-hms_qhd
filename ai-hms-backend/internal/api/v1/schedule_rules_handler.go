package v1

import (
	"log"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type ScheduleRulesHandler struct {
	svc *services.ScheduleBoardService
}

func NewScheduleRulesHandler() *ScheduleRulesHandler {
	return &ScheduleRulesHandler{svc: services.NewScheduleBoardService()}
}

// ===================== Board 快照 =====================

func (h *ScheduleRulesHandler) GetBoard(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	start, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.BadRequest(c, "date 格式应为 YYYY-MM-DD")
		return
	}
	end := start

	board, issues, slots, elapsed, err := h.svc.LoadBoardWithPrecheck(tenantID, start, end)
	if err != nil {
		log.Printf("[ERROR] board load failed: %v", err)
		response.InternalErrorSafe(c)
		return
	}

	response.Success(c, gin.H{
		"startDate": start.Format("2006-01-02"),
		"endDate":   end.Format("2006-01-02"),
		"wards":     board.Wards,
		"wardExts":  board.WardExts,
		"beds":      board.Beds,
		"bedExts":   board.BedMachineExts,
		"shifts":    board.Shifts,
		"profiles":  board.Profiles,
		"capacity":  slots,
		"issues":    issues,
		"elapsedMs": elapsed,
	})
}

// ===================== 预检 =====================

type PrecheckRequest struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

func (h *ScheduleRulesHandler) RunPrecheck(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	var req PrecheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	if req.StartDate == "" {
		response.BadRequest(c, "startDate 必填")
		return
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.BadRequest(c, "startDate 格式应为 YYYY-MM-DD")
		return
	}

	end := start
	if req.EndDate != "" {
		end, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			response.BadRequest(c, "endDate 格式应为 YYYY-MM-DD")
			return
		}
	}

	if end.Before(start) {
		response.BadRequest(c, "endDate 不能早于 startDate")
		return
	}

	maxDays := end.Sub(start).Hours() / 24
	if maxDays > 28 {
		response.BadRequest(c, "预检日期范围不能超过 28 天")
		return
	}

	board, issues, slots, elapsed, err := h.svc.LoadBoardWithPrecheck(tenantID, start, end)
	if err != nil {
		log.Printf("[ERROR] precheck failed: %v", err)
		response.InternalErrorSafe(c)
		return
	}

	response.Success(c, gin.H{
		"startDate": start.Format("2006-01-02"),
		"endDate":   end.Format("2006-01-02"),
		"issues":    issues,
		"capacity":  slots,
		"calendars": board.CalendarEntries,
		"outages":   board.Outages,
		"elapsedMs": elapsed,
	})
}

// ===================== 冲突列表 =====================

func (h *ScheduleRulesHandler) GetConflicts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	start, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.BadRequest(c, "date 格式应为 YYYY-MM-DD")
		return
	}
	end := start

	board, issues, _, elapsed, err := h.svc.LoadBoardWithPrecheck(tenantID, start, end)
	if err != nil {
		log.Printf("[ERROR] conflicts load failed: %v", err)
		response.InternalErrorSafe(c)
		return
	}

	criticals := []services.PrecheckIssue{}
	for _, iss := range issues {
		if iss.Severity == services.SeverityCritical {
			criticals = append(criticals, iss)
		}
	}

	response.Success(c, gin.H{
		"date":      dateStr,
		"conflicts": criticals,
		"allIssues": issues,
		"elapsedMs": elapsed,
		"board": gin.H{
			"wards":  board.Wards,
			"beds":   board.Beds,
			"shifts": board.Shifts,
		},
	})
}

// ===================== 路由注册 =====================

func RegisterScheduleRulesRoutes(r *gin.RouterGroup) {
	h := NewScheduleRulesHandler()

	rules := r.Group("/schedule/rules")
	{
		rules.GET("/board", h.GetBoard)
		rules.POST("/precheck", h.RunPrecheck)
	}

	r.GET("/schedule/conflicts", h.GetConflicts)
}
