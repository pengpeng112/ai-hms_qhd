package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// RNaHandler RNa 处方计算接口
// 专利 P2/P4/P6 的服务端计算入口（前端可直接用 JS 做实时计算，此接口用于留痕/审计）
type RNaHandler struct{}

func NewRNaHandler() *RNaHandler { return &RNaHandler{} }

// RegisterRNaRoutes 注册路由（在 main.go 用 protected 组调用）
// v1.RegisterRNaRoutes(protected)
func RegisterRNaRoutes(r *gin.RouterGroup) {
	h := NewRNaHandler()
	patients := r.Group("/patients")
	patients.POST("/:id/prescriptions/calculate-na", h.CalculateNa)
	patients.GET("/:id/prescriptions/na-history", h.GetNaHistory)
}

// CalculateNaRequest 计算请求（钠清除比模型 v2）
type CalculateNaRequest struct {
	CPre      float64  `json:"cPre" binding:"required,min=100,max=175"`
	PreWeight float64  `json:"preWeight" binding:"required,min=20,max=200"`
	DryWeight float64  `json:"dryWeight" binding:"required,min=20,max=200"`
	HeightCm  float64  `json:"heightCm" binding:"required,min=100,max=230"`
	AgeYears  float64  `json:"ageYears" binding:"required,min=1,max=120"`
	IsMale    bool     `json:"isMale"`
	VUF       *float64 `json:"vuf" binding:"omitempty,min=0,max=8"` // 超滤量 (L)；缺省用 preWeight−dryWeight

	// 钠目标：driver=rna 用 rNa；driver=cpost 用 cPost
	Driver string  `json:"driver" binding:"omitempty,oneof=rna cpost"`
	RNa    float64 `json:"rNa" binding:"omitempty,min=0.5,max=1.6"`
	CPost  float64 `json:"cPost" binding:"omitempty,min=110,max=160"`

	// 高级参数（可选）
	Alpha float64 `json:"alpha" binding:"omitempty,min=0.5,max=1"`
	D     float64 `json:"d" binding:"omitempty,min=1,max=40"`
	T     float64 `json:"t" binding:"omitempty,min=1,max=10"`
}

// CalculateNa 执行 RNa 钠清除比处方计算（POST /api/v1/patients/:id/prescriptions/calculate-na）
func (h *RNaHandler) CalculateNa(c *gin.Context) {
	var req CalculateNaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	driver := req.Driver
	if driver == "" {
		driver = "rna"
	}

	result := services.CalculateRNaPrescription(services.RNaCalculateRequest{
		CPre:      req.CPre,
		PreWeight: req.PreWeight,
		DryWeight: req.DryWeight,
		HeightCm:  req.HeightCm,
		AgeYears:  req.AgeYears,
		IsMale:    req.IsMale,
		VUF:       req.VUF,
		Driver:    driver,
		RNa:       req.RNa,
		CPost:     req.CPost,
		Alpha:     req.Alpha,
		D:         req.D,
		T:         req.T,
	})

	response.Success(c, result)
}

// NaHistoryPoint 血清钠历史数据点
type NaHistoryPoint struct {
	Date      string   `json:"date"`
	CPre      float64  `json:"cPre"`
	CDUsed    *float64 `json:"cdUsed,omitempty"`
	DeltaUsed *float64 `json:"deltaUsed,omitempty"`
}

// GetNaHistory 获取患者近期透前血清钠历史（GET /api/v1/patients/:id/prescriptions/na-history）
// 用于前端棘轮进度图。数据来源：lab_report_items（SERUM_NA）+ prescriptions 历史
func (h *RNaHandler) GetNaHistory(c *gin.Context) {
	patientID := c.Param("id")
	if patientID == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_PATIENT_ID", "患者ID不能为空")
		return
	}

	limitStr := c.DefaultQuery("limit", "8")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 30 {
		limit = 8
	}

	// TODO: 从 lab_report_items（ItemName 含"血清钠/Na+"）+ prescriptions 联查历史
	// 当前返回 stub，待 lab_report_items 数据积累后接真实查询
	// 真实实现：
	//   SELECT lri.ResultValue, lri.TestedAt, p.Parameters->>'na' AS cd_used
	//   FROM lab_report_items lri
	//   LEFT JOIN prescriptions p ON p.PatientID = lri.PatientID
	//     AND DATE(p.PrescriptionDate) = DATE(lri.TestedAt)
	//   WHERE lri.PatientID = patientID
	//     AND (lri.ItemName ILIKE '%血清钠%' OR lri.ItemName ILIKE '%Na+%')
	//   ORDER BY lri.TestedAt DESC
	//   LIMIT limit
	_ = limit

	stub := []NaHistoryPoint{
		{Date: time.Now().AddDate(0, 0, -21).Format("01-02"), CPre: 141.0},
		{Date: time.Now().AddDate(0, 0, -18).Format("01-02"), CPre: 140.5, DeltaUsed: floatPtr(2.0)},
		{Date: time.Now().AddDate(0, 0, -15).Format("01-02"), CPre: 140.0, DeltaUsed: floatPtr(2.0)},
		{Date: time.Now().AddDate(0, 0, -12).Format("01-02"), CPre: 139.5, DeltaUsed: floatPtr(2.0)},
		{Date: time.Now().AddDate(0, 0, -9).Format("01-02"), CPre: 139.0, DeltaUsed: floatPtr(2.0)},
		{Date: time.Now().AddDate(0, 0, -6).Format("01-02"), CPre: 138.5, DeltaUsed: floatPtr(1.5)},
		{Date: time.Now().AddDate(0, 0, -3).Format("01-02"), CPre: 138.0, DeltaUsed: floatPtr(2.0)},
	}

	response.Success(c, stub)
}

func floatPtr(f float64) *float64 { return &f }
