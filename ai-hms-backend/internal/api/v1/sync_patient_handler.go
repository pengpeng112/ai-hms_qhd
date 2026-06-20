package v1

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/his_oracle"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SyncPatientHandler struct {
	mappingSvc *services.ExternalPatientMappingService
	oracleCfg  config.HisOracleConfig
	tenantID   int64
}

func NewSyncPatientHandler(oracleCfg config.HisOracleConfig, tenantID int64) *SyncPatientHandler {
	return &SyncPatientHandler{
		mappingSvc: services.NewExternalPatientMappingService(),
		oracleCfg:  oracleCfg,
		tenantID:   tenantID,
	}
}

type UnmatchedPatientItem struct {
	PatientID string `json:"patientId"`
	Name      string `json:"name"`
	ExamCnt   int    `json:"examCnt"`
}

type UnmatchedPatientResponse struct {
	Items    []UnmatchedPatientItem `json:"items"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

func (h *SyncPatientHandler) ListUnmatched(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := strings.TrimSpace(c.Query("keyword"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	hcfg := his_oracle.Config{
		Host:     h.oracleCfg.Host,
		Port:     h.oracleCfg.Port,
		Service:  h.oracleCfg.Service,
		Username: h.oracleCfg.Username,
		Password: h.oracleCfg.Password,
	}
	client, err := his_oracle.NewClient(hcfg)
	if err != nil {
		response.InternalError(c, "Oracle 连接失败: "+err.Error())
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, qErr := client.QueryUnmatchedPatients(ctx, his_oracle.UnmatchedPatientsParams{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
	})
	if qErr != nil {
		response.InternalError(c, qErr.Error())
		return
	}

	patientIDs := make([]string, 0, len(result.Items))
	for _, row := range result.Items {
		patientIDs = append(patientIDs, row.PatientID)
	}

	confirmedSet := make(map[string]bool)
	if len(patientIDs) > 0 {
		var confirmed []string
		database.GetDB().Model(&models.ExternalPatientMapping{}).
			Where("external_system = ? AND match_status = ? AND external_patient_id IN ?",
				models.ExternalSystemHISOracle, models.MatchStatusConfirmed, patientIDs).
			Pluck("external_patient_id", &confirmed)
		for _, pid := range confirmed {
			confirmedSet[pid] = true
		}
	}

	var items []UnmatchedPatientItem
	for _, row := range result.Items {
		if confirmedSet[row.PatientID] {
			continue
		}
		nameMasked := maskName(row.NameVal)
		items = append(items, UnmatchedPatientItem{
			PatientID: row.PatientID,
			Name:      nameMasked,
			ExamCnt:   row.ExamCnt,
		})
	}

	if items == nil {
		items = []UnmatchedPatientItem{}
	}

	response.Success(c, UnmatchedPatientResponse{
		Items:    items,
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	})
}

type BindMappingRequest struct {
	ExternalPatientID string `json:"externalPatientId" binding:"required"`
	LegacyPatientID   int64  `json:"legacyPatientId" binding:"required"`
}

func (h *SyncPatientHandler) BindMapping(c *gin.Context) {
	var req BindMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供 externalPatientId 和 legacyPatientId")
		return
	}

	existing, _ := h.mappingSvc.FindByExternal(
		models.ExternalSystemHISOracle,
		req.ExternalPatientID,
		nil,
	)
	if existing != nil && existing.MatchStatus == models.MatchStatusConfirmed {
		response.BadRequest(c, "该 HIS 患者已绑定本地患者，请勿重复绑定")
		return
	}

	m := &models.ExternalPatientMapping{
		ID:               "epm_" + models.ExternalSystemHISOracle + "_" + req.ExternalPatientID,
		TenantID:         h.tenantID,
		LegacyPatientID:  req.LegacyPatientID,
		ExternalSystem:   models.ExternalSystemHISOracle,
		ExternalPatientID: req.ExternalPatientID,
		MatchStatus:      models.MatchStatusConfirmed,
	}
	if err := h.mappingSvc.CreateMapping(m); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		hcfg := his_oracle.Config{
			Host:     h.oracleCfg.Host,
			Port:     h.oracleCfg.Port,
			Service:  h.oracleCfg.Service,
			Username: h.oracleCfg.Username,
			Password: h.oracleCfg.Password,
		}
		syncSvc := services.NewHisExamReportSyncService(hcfg, h.tenantID)
		defer syncSvc.CloseOracleClient()

		result, syncErr := syncSvc.SyncPatientExamReports(ctx, req.LegacyPatientID)
		if syncErr != nil {
			log.Printf("[sync-patient] auto-sync failed for legacyPatient=%d: %v", req.LegacyPatientID, syncErr)
			return
		}
		log.Printf("[sync-patient] auto-sync done for legacyPatient=%d: created=%d updated=%d skipped=%d failed=%d",
			req.LegacyPatientID, result.Created, result.Updated, result.Skipped, result.Failed)
	}()

	response.Success(c, "绑定成功")
}

type PatientSearchItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
	DialysisNo string `json:"dialysisNo"`
}

func (h *SyncPatientHandler) SearchPatients(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		response.Success(c, []PatientSearchItem{})
		return
	}

	var rows []PatientSearchItem
	err := database.GetDB().Raw(`
		SELECT p."Id" AS id, p."Name" AS name, p."Gender" AS gender,
		       EXTRACT(YEAR FROM AGE(CURRENT_DATE, p."BirthDate"))::INT AS age,
		       p."DialysisNo" AS dialysis_no
		FROM "Register_PatientInfomation" p
		WHERE p."TenantId" = ?
		  AND (p."Name" LIKE ? OR p."DialysisNo" LIKE ? OR CAST(p."Id" AS TEXT) LIKE ?)
		ORDER BY p."Id"
		LIMIT 20
	`, h.tenantID, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").Scan(&rows).Error
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if rows == nil {
		rows = []PatientSearchItem{}
	}
	response.Success(c, rows)
}

func maskName(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return "*"
	}
	if len(runes) <= 2 {
		return string(runes[0]) + "*"
	}
	return string(runes[0]) + string([]rune{'*'}) + string(runes[len(runes)-1])
}

func RegisterSyncPatientRoutes(r *gin.RouterGroup, oracleCfg config.HisOracleConfig, tenantID int64) {
	handler := NewSyncPatientHandler(oracleCfg, tenantID)
	sync := r.Group("/sync")
	{
		sync.GET("/unmatched-patients", handler.ListUnmatched)
		sync.POST("/external-mappings/bind", handler.BindMapping)
	}
	patients := r.Group("/patients")
	{
		patients.GET("/search", handler.SearchPatients)
	}
}
