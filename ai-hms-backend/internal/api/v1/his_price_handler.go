package v1

import (
	"log"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

var HisPriceSyncRoles = []string{"ADMIN", "管理员"}

type HisPriceHandler struct {
	priceSvc  *services.HisPriceService
	syncSvc   *services.HisPriceSyncService
	tenantID  int64
}

func NewHisPriceHandler(oracleCfg config.HisOracleConfig, tenantID int64) *HisPriceHandler {
	syncSvc, err := services.NewHisPriceSyncService(oracleCfg)
	if err != nil {
		log.Printf("[his_price] init sync service warning: %v", err)
	}
	return &HisPriceHandler{
		priceSvc: services.NewHisPriceService(),
		syncSvc:  syncSvc,
		tenantID: tenantID,
	}
}

func RegisterHisPriceRoutes(rg *gin.RouterGroup, oracleCfg config.HisOracleConfig, tenantID int64) {
	h := NewHisPriceHandler(oracleCfg, tenantID)
	rg.GET("/his-price-items", h.Search)
	rg.GET("/his-price-items/:itemCode", h.GetByCode)

	syncWrite := rg.Group("")
	syncWrite.Use(middleware.RequireRoles(HisPriceSyncRoles...))
	syncWrite.POST("/his-price-items/sync", h.Sync)
}

func (h *HisPriceHandler) Search(c *gin.Context) {
	var req services.HisPriceSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.priceSvc.Search(req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *HisPriceHandler) Sync(c *gin.Context) {
	if h.syncSvc == nil {
		response.InternalError(c, "HIS Oracle 连接未初始化")
		return
	}

	if h.syncSvc.IsSyncRunning() {
		response.BadRequest(c, "HIS 价表同步任务正在进行中，请稍后再试")
		return
	}

	runID := utils.GenerateID()

	run := &models.SyncJobRun{
		ID:           runID,
		JobCode:      models.SyncJobCodeHisPriceList,
		SourceSystem: "HIS_ORACLE",
		SyncType:     models.SyncTypeHisPriceList,
		Status:       models.SyncJobStatusRunning,
		StartedAt:    time.Now(),
	}

	syncJobSvc := services.NewSyncJobService()
	if err := syncJobSvc.CreateRun(run); err != nil {
		response.InternalError(c, "创建同步运行记录失败: "+err.Error())
		return
	}

	go func() {
		fetched, created, updated, errMsg := h.syncSvc.SyncPriceList(runID)
		status := models.SyncJobStatusSuccess
		if errMsg != "" {
			status = models.SyncJobStatusFailed
			if fetched > 0 {
				status = models.SyncJobStatusPartial
			}
		}
		cursorAfter := time.Now().Format(time.RFC3339)
		counts := map[string]int{
			"fetched": fetched, "created": created, "updated": updated,
		}
		if finishErr := syncJobSvc.FinishRun(runID, status, counts, cursorAfter, "", errMsg); finishErr != nil {
			log.Printf("[his_price_sync] finish run failed: %v", finishErr)
		}
	}()

	response.Success(c, gin.H{
		"message": "HIS 价表同步任务已启动",
		"runId":   runID,
	})
}

func (h *HisPriceHandler) GetByCode(c *gin.Context) {
	itemCode := strings.TrimSpace(c.Param("itemCode"))
	if itemCode == "" {
		response.BadRequest(c, "项目编码不能为空")
		return
	}
	item, err := h.priceSvc.FindByItemCode(itemCode)
	if err != nil {
		response.NotFound(c, "未找到该项目: "+err.Error())
		return
	}
	response.Success(c, item)
}
