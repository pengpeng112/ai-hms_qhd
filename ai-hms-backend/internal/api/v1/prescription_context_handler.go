package v1

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// PrescriptionContextHandler 处方开单参考数据聚合接口
type PrescriptionContextHandler struct{}

func NewPrescriptionContextHandler() *PrescriptionContextHandler {
	return &PrescriptionContextHandler{}
}

// RegisterPrescriptionContextRoutes 注册路由（在 main.go 用 protected 组调用）
func RegisterPrescriptionContextRoutes(r *gin.RouterGroup) {
	h := NewPrescriptionContextHandler()
	patients := r.Group("/patients")
	patients.GET("/:id/prescriptions/context", h.GetContext)
}

// ─── 响应结构体 ────────────────────────────────────────────────────────────

// LabIndicator 单项检验指标（含范围判断）
type LabIndicator struct {
	ConceptID   string   `json:"conceptId"`
	DisplayName string   `json:"displayName"`
	Value       *float64 `json:"value"`
	Unit        string   `json:"unit"`
	TargetLow   float64  `json:"targetLow"`
	TargetHigh  float64  `json:"targetHigh"`
	Status      string   `json:"status"` // normal / watch / high / low / critical_high / critical_low / missing
	StatusLabel string   `json:"statusLabel"`
	TestedAt    *string  `json:"testedAt"`
	DaysAgo     *int     `json:"daysAgo"`
	ActionHint  string   `json:"actionHint,omitempty"`
}

// WeightUFContext 体重与超滤
type WeightUFContext struct {
	PreWeight     *float64 `json:"preWeight"`
	DryWeight     *float64 `json:"dryWeight"`
	WeightGain    *float64 `json:"weightGain"`
	WeightGainPct *float64 `json:"weightGainPct"`
	UFTarget      *float64 `json:"ufTarget"`
	UFRatePerKg   *float64 `json:"ufRatePerKg"`  // 体重归一化超滤率 mL/kg/h（按 4h 估算）
	UFRateStatus  string   `json:"ufRateStatus"` // safe / watch / danger
	PreWeightAt   *string  `json:"preWeightAt"`
}

// LastTreatmentContext 上次治疗概况
type LastTreatmentContext struct {
	TreatmentDate  *string `json:"treatmentDate"`
	ActualUFml     *int    `json:"actualUFml"`
	PlannedUFml    *int    `json:"plannedUFml"`
	UFDiffMl       *int    `json:"ufDiffMl"`
	ActualMinutes  *int    `json:"actualMinutes"`
	PlannedMinutes *int    `json:"plannedMinutes"`
	AlarmCount     int     `json:"alarmCount"`
	AlarmSummary   string  `json:"alarmSummary"`
	Outcome        string  `json:"outcome"` // completed / early_stop / not_found
}

// PrescriptionHint 处方调整提示
type PrescriptionHint struct {
	Priority    int    `json:"priority"` // 1=紧急, 2=关注, 3=参考
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// TrendPoint 趋势数据点（最近 N 次治疗）
type TrendPoint struct {
	Date  string   `json:"date"`
	Value float64  `json:"value"`         // 主值（钠清除比 / 收缩压 / 心率）
	Aux   *float64 `json:"aux,omitempty"` // 辅值（舒张压，仅血压用）
}

// SodiumClearanceContext 钠清除比（专利核心指标：本次脱钠效果评估）
type SodiumClearanceContext struct {
	LastRatio   *float64     `json:"lastRatio"` // 上次钠清除比
	LastDate    *string      `json:"lastDate"`
	TargetLow   float64      `json:"targetLow"`  // 目标下限
	TargetHigh  float64      `json:"targetHigh"` // 目标上限
	Status      string       `json:"status"`     // good / low / high
	StatusLabel string       `json:"statusLabel"`
	Trend       []TrendPoint `json:"trend"` // 最近12次
}

// VitalSignsContext 血压心率（每点 = 该次透析全程监测的平均值）
type VitalSignsContext struct {
	LastSystolic  *float64     `json:"lastSystolic"`  // 上次透析平均收缩压
	LastDiastolic *float64     `json:"lastDiastolic"` // 上次透析平均舒张压
	LastHeartRate *float64     `json:"lastHeartRate"` // 上次透析平均心率
	LastDate      *string      `json:"lastDate"`
	BPStatus      string       `json:"bpStatus"` // normal / high / low
	BPStatusLabel string       `json:"bpStatusLabel"`
	HRStatus      string       `json:"hrStatus"` // normal / high / low
	HRStatusLabel string       `json:"hrStatusLabel"`
	BPTrend       []TrendPoint `json:"bpTrend"` // 各次透析平均收缩压(Value)/舒张压(Aux)，最近12次
	HRTrend       []TrendPoint `json:"hrTrend"` // 各次透析平均心率，最近12次

	// 正常范围界限（用于前端参考带 + 状态判定）
	SysLow  float64 `json:"sysLow"`  // 收缩压下限 90
	SysHigh float64 `json:"sysHigh"` // 收缩压上限 150
	DiaLow  float64 `json:"diaLow"`  // 舒张压下限 60
	DiaHigh float64 `json:"diaHigh"` // 舒张压上限 90
	HRLow   float64 `json:"hrLow"`   // 心率下限 60
	HRHigh  float64 `json:"hrHigh"`  // 心率上限 90
}

// PatientDemographics 患者人口学信息（供 RNa 取真实身高/年龄/性别）
// 来源：Register_PatientInfomation（Gender/BirthDate）+ patient_basic_infos（height）
type PatientDemographics struct {
	HeightCm   *float64 `json:"heightCm"`   // 身高 cm（patient_basic_infos.height）
	AgeYears   *int     `json:"ageYears"`   // 年龄（由 BirthDate 计算）
	IsMale     bool     `json:"isMale"`     // 性别（Gender == 男）
	GenderText string   `json:"genderText"` // 原始性别文本
}

// PrescriptionContextResponse 完整聚合响应
type PrescriptionContextResponse struct {
	PatientID       string                 `json:"patientId"`
	PatientName     string                 `json:"patientName"`
	BedCode         string                 `json:"bedCode"`
	Demographics    PatientDemographics    `json:"demographics"`
	Weight          WeightUFContext        `json:"weight"`
	Labs            []LabIndicator         `json:"labs"`
	LastTx          LastTreatmentContext   `json:"lastTreatment"`
	SodiumClearance SodiumClearanceContext `json:"sodiumClearance"`
	Vitals          VitalSignsContext      `json:"vitals"`
	Hints           []PrescriptionHint     `json:"hints"`
	DryWeight       *DwContextData         `json:"dryWeight,omitempty"`
	GeneratedAt     string                 `json:"generatedAt"`
}

// DwContextData 干体重评估上下文（供 RNa 面板阶段驱动）
type DwContextData struct {
	DryWeight    *float64 `json:"dryWeight"`
	Phase        string   `json:"phase"`
	SuggestedRNa float64  `json:"suggestedRNa"`
}

// ─── Handler ──────────────────────────────────────────────────────────────

// GetContext 聚合处方开单所需的全部参考数据
// GET /api/v1/patients/:id/prescriptions/context
func (h *PrescriptionContextHandler) GetContext(c *gin.Context) {
	patientID := strings.TrimSpace(c.Param("id"))
	if patientID == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_PATIENT_ID", "患者ID不能为空")
		return
	}

	db := database.GetDB()
	today := time.Now().Format("2006-01-02")

	// 老库主键为 bigint，非数字 ID 直接跳过 DB 查询（回退空值/模拟）
	pid, pidErr := strconv.ParseInt(strings.TrimSpace(patientID), 10, 64)

	// 1. 患者姓名（建档表 Register_PatientInfomation）+ 干体重（最新有效 Plan_PatientPlan）+ 床位（最近治疗）
	//    列名/表名查证自最新代码 patient_service.go（DryWeight 来自 Plan_PatientPlan，非 Prescription）。
	var patInfo struct {
		PatientName string
		BedCode     string
		DryWeight   float64
	}
	if pidErr == nil {
		var nameRow struct {
			Name string `gorm:"column:name"`
		}
		db.Raw(`SELECT "Name" AS name FROM "Register_PatientInfomation" WHERE "Id" = ? AND "TenantId" = ? LIMIT 1`, pid, services.LegacyTenantID).Scan(&nameRow)
		patInfo.PatientName = nameRow.Name

		var planRow struct {
			DryWeight float64 `gorm:"column:dry_weight"`
		}
		db.Raw(`SELECT "DryWeight" AS dry_weight FROM "Plan_PatientPlan"
		        WHERE "PatientId" = ? AND "TenantId" = ? AND COALESCE("IsDisabled", false) = false
		        ORDER BY "CreateTime" DESC LIMIT 1`, pid, services.LegacyTenantID).Scan(&planRow)
		patInfo.DryWeight = planRow.DryWeight

		var bedRow struct {
			BedName string `gorm:"column:bed_name"`
		}
		db.Raw(`SELECT "BedName" AS bed_name FROM "Treatment_Treatment"
		        WHERE "PatientId" = ? AND "TenantId" = ?
		        ORDER BY COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime") DESC LIMIT 1`, pid, services.LegacyTenantID).Scan(&bedRow)
		patInfo.BedCode = bedRow.BedName
	}

	// 2. 今日透前体重（Treatment_BeforeSigns.Weight，按 TreatmentId 关联当日治疗）
	var beforeCheck struct {
		Weight    *float64   `gorm:"column:weight"`
		CheckedAt *time.Time `gorm:"column:checked_at"`
	}
	if pidErr == nil {
		db.Raw(`
			SELECT s."Weight" AS weight, s."OperateTime" AS checked_at
			FROM "Treatment_BeforeSigns" s
			JOIN "Treatment_Treatment" t ON t."Id" = s."TreatmentId"
			WHERE t."PatientId" = ? AND t."TenantId" = ? AND s."TenantId" = ?
			  AND DATE(COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime")) = DATE(?)
			ORDER BY s."OperateTime" DESC
			LIMIT 1
		`, pid, services.LegacyTenantID, services.LegacyTenantID, today).Scan(&beforeCheck)
	}

	// 2b. 患者人口学信息（独立查询，失败不影响主流程）
	demographics := queryDemographics(patientID)

	// 3. 体重/超滤汇总
	weightCtx := buildWeightContext(beforeCheck.Weight, patInfo.DryWeight, beforeCheck.CheckedAt)

	// 4. 最近检验指标
	labs := queryRecentLabs(patientID)

	// 5. 上次治疗概况
	lastTx := queryLastTreatment(patientID, today)

	// 6. 钠清除比 + 血压心率趋势（近12次）
	//    血压心率：已接真实表 Treatment_DuringSigns（见 buildVitals），无数据时回退模拟。
	//    钠清除比：SRR 需"实际清除钠"，依赖透后血钠/在线离子清除，老库未常规存储，暂用模拟序列。
	naClearance := buildSodiumClearance(patientID)
	vitals := buildVitals(patientID)

	// 7. 生成处方提示（含钠清除比、血压）
	hints := buildPrescriptionHints(labs, weightCtx, lastTx, naClearance, vitals)

	response.Success(c, PrescriptionContextResponse{
		PatientID:       patientID,
		PatientName:     patInfo.PatientName,
		BedCode:         patInfo.BedCode,
		Demographics:    demographics,
		Weight:          weightCtx,
		Labs:            labs,
		LastTx:          lastTx,
		SodiumClearance: naClearance,
		Vitals:          vitals,
		Hints:           hints,
		DryWeight:       fetchDryWeightContext(c, pid),
		GeneratedAt:     time.Now().Format("2006-01-02 15:04"),
	})
}

// ─── 体重 / 超滤 ──────────────────────────────────────────────────────────

func buildWeightContext(preWeight *float64, dryWeight float64, checkedAt *time.Time) WeightUFContext {
	ctx := WeightUFContext{}
	if dryWeight > 0 {
		dw := dryWeight
		ctx.DryWeight = &dw
	}
	if preWeight == nil || dryWeight <= 0 {
		return ctx
	}
	ctx.PreWeight = preWeight
	gain := *preWeight - dryWeight
	ctx.WeightGain = &gain
	gainPct := gain / dryWeight * 100
	ctx.WeightGainPct = &gainPct
	if gain > 0 {
		ufTarget := gain
		ctx.UFTarget = &ufTarget
		// 体重归一化超滤率 = 增量(mL) / 4h / 干体重(kg) → mL/kg/h
		ufRate := gain * 1000 / 4.0 / dryWeight
		ctx.UFRatePerKg = &ufRate
		switch {
		case ufRate > 13:
			ctx.UFRateStatus = "danger"
		case ufRate > 11:
			ctx.UFRateStatus = "watch"
		default:
			ctx.UFRateStatus = "safe"
		}
	}
	if checkedAt != nil {
		s := checkedAt.Format("15:04")
		ctx.PreWeightAt = &s
	}
	return ctx
}

// ─── 患者人口学信息 ────────────────────────────────────────────────────────

// queryDemographics 取身高/性别/年龄（独立查询，列名以 DBA 确认为准）
//
//	性别/出生日期：Register_PatientInfomation（Gender / BirthDate）
//	身高：patient_basic_infos.height（varchar，cm）
func queryDemographics(patientID string) PatientDemographics {
	// 真实来源 = 建档表 Register_PatientInfomation（含 Gender/BirthDate/Height）。
	// 与 patient_basic_service.go 的 getLegacyPatientBasic 一致：带引号 CamelCase 标识符、
	// Height 为 numeric 需 CAST 成 text、Gender 存 'M'/'F'、主键 "Id" 为 bigint。
	id, perr := strconv.ParseInt(strings.TrimSpace(patientID), 10, 64)
	if perr != nil {
		return PatientDemographics{}
	}
	var row struct {
		Gender    string     `gorm:"column:gender"`
		BirthDate *time.Time `gorm:"column:birth_date"`
		Height    *string    `gorm:"column:height"`
	}
	database.GetDB().Raw(`
		SELECT "Gender" AS gender,
		       "BirthDate" AS birth_date,
		       CAST("Height" AS text) AS height
		FROM "Register_PatientInfomation"
		WHERE "Id" = ? AND "TenantId" = ?
		LIMIT 1
	`, id, services.LegacyTenantID).Scan(&row)

	g := strings.TrimSpace(row.Gender)
	d := PatientDemographics{
		GenderText: g,
		// 建档存 'M'/'F'；兼容历史 '男'/'male'
		IsMale: strings.EqualFold(g, "M") || strings.Contains(g, "男") || strings.EqualFold(g, "male"),
	}
	if row.Height != nil {
		if h, err := strconv.ParseFloat(strings.TrimSpace(*row.Height), 64); err == nil && h > 0 {
			d.HeightCm = &h
		}
	}
	if row.BirthDate != nil {
		age := int(time.Since(*row.BirthDate).Hours() / 24 / 365.25)
		if age >= 0 && age < 130 {
			d.AgeYears = &age
		}
	}
	return d
}

// ─── 检验指标 ──────────────────────────────────────────────────────────────

type labSpec struct {
	ConceptID   string
	DisplayName string
	Unit        string
	Keywords    []string
	TargetLow   float64
	TargetHigh  float64
	CritLow     float64
	CritHigh    float64
}

var labSpecs = []labSpec{
	{"SERUM_K", "钾 K⁺", "mmol/L", []string{"血清钾", "血钾", "K+", "钾"}, 3.5, 5.5, 2.5, 6.5},
	{"SERUM_CA", "钙 Ca²⁺", "mmol/L", []string{"血清钙", "血钙", "Ca", "钙"}, 2.1, 2.5, 1.75, 3.0},
	{"SERUM_P", "磷 P", "mmol/L", []string{"血清磷", "血磷", "磷"}, 1.1, 1.8, 0.5, 3.0},
	{"HEMOGLOBIN", "血红蛋白 Hb", "g/L", []string{"血红蛋白", "Hb", "HGB"}, 110, 130, 60, 180},
	{"FERRITIN", "铁蛋白 Ferritin", "ng/mL", []string{"铁蛋白", "Ferritin", "FERR"}, 200, 800, 50, 1200},
	{"KTV", "透析充分性 Kt/V", "", []string{"Kt/V", "KtV", "透析充分"}, 1.2, 9.0, 0.8, 9.0},
	{"PTH", "甲状旁腺素 iPTH", "pg/mL", []string{"甲状旁腺素", "PTH", "iPTH"}, 150, 600, 50, 1500},
	{"SERUM_NA", "血清钠 Na⁺", "mmol/L", []string{"血清钠", "血钠", "Na+", "钠"}, 135, 145, 120, 155},
}

func queryRecentLabs(patientID string) []LabIndicator {
	// lab_report_items / lab_reports 为新库 snake_case 表（HDIS/LIS 同步目标）。
	// 列名查证自 models/lab_report.go：item_name / result_value / tested_at（item 自带检验时间）。
	// 注：lab_reports.patient_id 为 varchar(36)，其取值口径由 LIS/HDIS 同步决定，
	//     与老库 bigint Id 是否一致需上线核对（见开发文档待办）。
	type rawRow struct {
		ItemName    string    `gorm:"column:item_name"`
		ResultValue string    `gorm:"column:result_value"`
		TestedAt    time.Time `gorm:"column:tested_at"`
	}
	var rows []rawRow
	database.GetDB().Raw(`
		SELECT lri.item_name AS item_name,
		       lri.result_value AS result_value,
		       lri.tested_at AS tested_at
		FROM lab_report_items lri
		JOIN lab_reports lr ON lr.id = lri.lab_report_id
		WHERE lr.patient_id = ?
		  AND lri.tested_at >= (CURRENT_DATE - INTERVAL '60 days')
		ORDER BY lri.tested_at DESC
		LIMIT 200
	`, patientID).Scan(&rows)

	now := time.Now()
	results := make([]LabIndicator, 0, len(labSpecs))

	for _, spec := range labSpecs {
		ind := LabIndicator{
			ConceptID:   spec.ConceptID,
			DisplayName: spec.DisplayName,
			Unit:        spec.Unit,
			TargetLow:   spec.TargetLow,
			TargetHigh:  spec.TargetHigh,
			Status:      "missing",
			StatusLabel: "无数据",
		}

		for _, row := range rows {
			if !matchesLabKeywords(row.ItemName, spec.Keywords) {
				continue
			}
			val, err := strconv.ParseFloat(strings.TrimSpace(row.ResultValue), 64)
			if err != nil {
				continue
			}
			v := val
			ind.Value = &v
			t := row.TestedAt.Format("01-02")
			ind.TestedAt = &t
			days := int(now.Sub(row.TestedAt).Hours() / 24)
			ind.DaysAgo = &days
			ind.Status, ind.StatusLabel = classifyLabValue(val, spec)
			break
		}

		results = append(results, ind)
	}
	return results
}

func matchesLabKeywords(name string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

func classifyLabValue(v float64, spec labSpec) (string, string) {
	switch {
	case v < spec.CritLow:
		return "critical_low", "极低"
	case v > spec.CritHigh:
		return "critical_high", "极高"
	case v < spec.TargetLow:
		if (spec.TargetLow-v)/spec.TargetLow > 0.15 {
			return "low", "↓ 偏低"
		}
		return "watch", "⚠ 临界"
	case v > spec.TargetHigh:
		if (v-spec.TargetHigh)/spec.TargetHigh > 0.15 {
			return "high", "↑ 偏高"
		}
		return "watch", "⚠ 临界"
	default:
		return "normal", "✓ 正常"
	}
}

// ─── 上次治疗 ──────────────────────────────────────────────────────────────

func queryLastTreatment(patientID, todayStr string) LastTreatmentContext {
	// 真实数据源 Treatment_Treatment（实际超滤/时长/状态）+ Treatment_Alarm（报警计数）。
	// 老库为带引号 CamelCase 标识符；列名均取自《老血透数据库表结构-合并版.md》已确认字段。
	// 计划超滤/时长来自当日处方（DayProgrammeId→处方表），join 关系尚未确认，暂留空（前端可缺省）。
	var row struct {
		StartTime  time.Time `gorm:"column:start_time"`
		RealUF     *float64  `gorm:"column:real_uf"`    // 实际超滤总量
		RealHours  *float64  `gorm:"column:real_hours"` // 实际治疗时长（小时）
		Status     string    `gorm:"column:status"`     // 已结束60 / 中断50 ...
		AlarmCount int       `gorm:"column:alarm_count"`
	}
	database.GetDB().Raw(`
		SELECT COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime") AS start_time,
		       t."RealUFQuantity" AS real_uf,
		       t."RealDuration"   AS real_hours,
		       t."Status"         AS status,
		       COALESCE(a.cnt, 0) AS alarm_count
		FROM "Treatment_Treatment" t
		LEFT JOIN (
		    SELECT "TreatmentId", COUNT(*) AS cnt
		    FROM "Treatment_Alarm"
		    WHERE "Levle" <> 1000
		    GROUP BY "TreatmentId"
		) a ON a."TreatmentId" = t."Id"
		WHERE t."PatientId" = ? AND t."TenantId" = ?
		  AND DATE(COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime")) < DATE(?)
		ORDER BY COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime") DESC
		LIMIT 1
	`, patientID, services.LegacyTenantID, todayStr).Scan(&row)

	if row.StartTime.IsZero() {
		return LastTreatmentContext{Outcome: "not_found"}
	}

	dateStr := row.StartTime.Format("2006-01-02")
	ctx := LastTreatmentContext{
		TreatmentDate: &dateStr,
		AlarmCount:    row.AlarmCount,
		Outcome:       "completed",
	}
	// 实际超滤：RealUFQuantity 单位不固定，<100 视为 L 转 mL，否则按 mL
	if row.RealUF != nil {
		ufml := int(math.Round(*row.RealUF))
		if *row.RealUF < 100 {
			ufml = int(math.Round(*row.RealUF * 1000))
		}
		ctx.ActualUFml = &ufml
	}
	// 实际时长：RealDuration 为小时 → 分钟
	if row.RealHours != nil {
		mins := int(math.Round(*row.RealHours * 60))
		ctx.ActualMinutes = &mins
	}
	if strings.Contains(row.Status, "50") || strings.Contains(row.Status, "中断") {
		ctx.Outcome = "early_stop"
	}
	return ctx
}

// ─── 钠清除比 + 血压心率（近12次趋势）────────────────────────────────────────
//
// 这两组数据是脱钠处方的关键依据：
//   · 钠清除比反映"上次实际脱了多少钠"，趋势平台化 = 钠池接近排空（专利棘轮自限）
//   · 血压心率反映患者对脱水/脱钠的耐受度，趋势下行 = 干体重可能设低或脱钠过猛
//
// 当前为确定性模拟序列（按 patientID 播种），待接 Treatment_DuringParam / Monitoring 真实数据。

// seedFromID 由患者ID生成稳定种子，保证同一患者每次返回一致的模拟曲线
func seedFromID(patientID string) int64 {
	var s int64
	for _, ch := range patientID {
		s = s*131 + int64(ch)
	}
	if s < 0 {
		s = -s
	}
	return s
}

// buildSodiumClearance 钠清除比近12次趋势（模拟：呈缓降平台，体现疗程棘轮自限）
func buildSodiumClearance(patientID string) SodiumClearanceContext {
	seed := seedFromID(patientID)
	const n = 12
	const targetLow, targetHigh = 0.90, 1.10

	trend := make([]TrendPoint, 0, n)
	base := 1.18 // 疗程初期清除比偏高（钠池满），随疗程缓降趋于平台
	for i := 0; i < n; i++ {
		// 缓降 + 轻微噪声
		decay := 0.22 * float64(i) / float64(n-1)
		noise := (float64((seed+int64(i*7))%11) - 5) / 100.0 // ±0.05
		v := base - decay + noise
		v = math.Round(v*100) / 100
		d := time.Now().AddDate(0, 0, -(n-1-i)*3) // 每3天一次（隔日透）
		trend = append(trend, TrendPoint{Date: d.Format("01-02"), Value: v})
	}

	last := trend[len(trend)-1].Value
	lastDate := trend[len(trend)-1].Date
	status, label := "good", "✓ 达标"
	switch {
	case last > targetHigh:
		status, label = "high", "↑ 偏高（钠池仍充盈）"
	case last < targetLow:
		status, label = "low", "↓ 偏低（脱钠不足/过度代偿）"
	}

	return SodiumClearanceContext{
		LastRatio:   &last,
		LastDate:    &lastDate,
		TargetLow:   targetLow,
		TargetHigh:  targetHigh,
		Status:      status,
		StatusLabel: label,
		Trend:       trend,
	}
}

// buildVitals 血压心率近12次趋势（每点 = 该次透析全程监测均值）
// 真实数据源：Treatment_DuringSigns（透中生命体征，单次透析多条 SBP/DBP/HeartRate）
//
//	按 TreatmentId 取均值，关联 Treatment_Treatment 取患者与时间，近 12 次。
//
// 查不到（DB 不可达 / 尚无监测数据）时回退确定性模拟序列，保证演示可用。
func buildVitals(patientID string) VitalSignsContext {
	if real, ok := queryVitalsReal(patientID); ok {
		return real
	}
	return buildVitalsMock(patientID)
}

// queryVitalsReal 从 Treatment_DuringSigns 取近 12 次透析的血压/心率均值
func queryVitalsReal(patientID string) (VitalSignsContext, bool) {
	type row struct {
		StartTime time.Time `gorm:"column:start_time"`
		Sys       *float64  `gorm:"column:sys"`
		Dia       *float64  `gorm:"column:dia"`
		HR        *float64  `gorm:"column:hr"`
	}
	var rows []row
	// 老库为带引号的 CamelCase 标识符（与 dashboard_service 一致）
	err := database.GetDB().Raw(`
		SELECT COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime") AS start_time,
		       AVG(s."SBP")       AS sys,
		       AVG(s."DBP")       AS dia,
		       AVG(s."HeartRate") AS hr
		FROM "Treatment_Treatment" t
		JOIN "Treatment_DuringSigns" s ON s."TreatmentId" = t."Id"
		WHERE t."PatientId" = ? AND t."TenantId" = ? AND s."TenantId" = ? AND s."SBP" IS NOT NULL
		GROUP BY t."Id", COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime")
		ORDER BY COALESCE(t."StartTime", t."SignInTime", t."ReceptionTime", t."CreateTime") DESC
		LIMIT 12
	`, patientID, services.LegacyTenantID, services.LegacyTenantID).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return VitalSignsContext{}, false
	}

	const sysLow, sysHigh = 90.0, 150.0
	const diaLow, diaHigh = 60.0, 90.0
	const hrLow, hrHigh = 60.0, 90.0

	// rows 为时间倒序，反转为正序（最早→最新）
	bpTrend := make([]TrendPoint, 0, len(rows))
	hrTrend := make([]TrendPoint, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		r := rows[i]
		d := r.StartTime.Format("01-02")
		sys := 0.0
		if r.Sys != nil {
			sys = math.Round(*r.Sys)
		}
		var diaPtr *float64
		if r.Dia != nil {
			dv := math.Round(*r.Dia)
			diaPtr = &dv
		}
		bpTrend = append(bpTrend, TrendPoint{Date: d, Value: sys, Aux: diaPtr})
		hr := 0.0
		if r.HR != nil {
			hr = math.Round(*r.HR)
		}
		hrTrend = append(hrTrend, TrendPoint{Date: d, Value: hr})
	}

	lastSys := bpTrend[len(bpTrend)-1].Value
	var lastDia float64
	if bpTrend[len(bpTrend)-1].Aux != nil {
		lastDia = *bpTrend[len(bpTrend)-1].Aux
	}
	lastHR := hrTrend[len(hrTrend)-1].Value
	lastDate := bpTrend[len(bpTrend)-1].Date

	bpStatus, bpLabel := "normal", "✓ 正常"
	switch {
	case lastSys > sysHigh || lastDia > diaHigh:
		bpStatus, bpLabel = "high", "↑ 偏高"
	case lastSys < sysLow || lastDia < diaLow:
		bpStatus, bpLabel = "low", "↓ 偏低"
	}
	hrStatus, hrLabel := "normal", "✓ 平稳"
	switch {
	case lastHR > hrHigh:
		hrStatus, hrLabel = "high", "↑ 偏快"
	case lastHR < hrLow:
		hrStatus, hrLabel = "low", "↓ 偏慢"
	}

	return VitalSignsContext{
		LastSystolic: &lastSys, LastDiastolic: &lastDia, LastHeartRate: &lastHR, LastDate: &lastDate,
		BPStatus: bpStatus, BPStatusLabel: bpLabel, HRStatus: hrStatus, HRStatusLabel: hrLabel,
		BPTrend: bpTrend, HRTrend: hrTrend,
		SysLow: sysLow, SysHigh: sysHigh, DiaLow: diaLow, DiaHigh: diaHigh, HRLow: hrLow, HRHigh: hrHigh,
	}, true
}

// buildVitalsMock 模拟序列（真实查询无数据时回退）
func buildVitalsMock(patientID string) VitalSignsContext {
	seed := seedFromID(patientID)
	const n = 12

	bpTrend := make([]TrendPoint, 0, n)
	hrTrend := make([]TrendPoint, 0, n)
	sysBase := 158.0 // 透前收缩压偏高，随容量管理改善缓降
	for i := 0; i < n; i++ {
		decay := 12.0 * float64(i) / float64(n-1)
		sysNoise := float64((seed+int64(i*13))%15) - 7 // ±7
		sys := math.Round(sysBase - decay + sysNoise)
		dia := math.Round(sys*0.58 + float64((seed+int64(i*5))%8) - 4)
		hr := math.Round(76 + float64((seed+int64(i*17))%17) - 8) // 68~84 波动
		d := time.Now().AddDate(0, 0, -(n-1-i)*3)
		diaCopy := dia
		bpTrend = append(bpTrend, TrendPoint{Date: d.Format("01-02"), Value: sys, Aux: &diaCopy})
		hrTrend = append(hrTrend, TrendPoint{Date: d.Format("01-02"), Value: hr})
	}

	// 正常范围界限
	const sysLow, sysHigh = 90.0, 150.0
	const diaLow, diaHigh = 60.0, 90.0
	const hrLow, hrHigh = 60.0, 90.0

	lastSys := bpTrend[len(bpTrend)-1].Value
	lastDia := *bpTrend[len(bpTrend)-1].Aux
	lastHR := hrTrend[len(hrTrend)-1].Value
	lastDate := bpTrend[len(bpTrend)-1].Date

	bpStatus, bpLabel := "normal", "✓ 正常"
	switch {
	case lastSys > sysHigh || lastDia > diaHigh:
		bpStatus, bpLabel = "high", "↑ 偏高"
	case lastSys < sysLow || lastDia < diaLow:
		bpStatus, bpLabel = "low", "↓ 偏低"
	}

	hrStatus, hrLabel := "normal", "✓ 平稳"
	switch {
	case lastHR > hrHigh:
		hrStatus, hrLabel = "high", "↑ 偏快"
	case lastHR < hrLow:
		hrStatus, hrLabel = "low", "↓ 偏慢"
	}

	return VitalSignsContext{
		LastSystolic:  &lastSys,
		LastDiastolic: &lastDia,
		LastHeartRate: &lastHR,
		LastDate:      &lastDate,
		BPStatus:      bpStatus,
		BPStatusLabel: bpLabel,
		HRStatus:      hrStatus,
		HRStatusLabel: hrLabel,
		BPTrend:       bpTrend,
		HRTrend:       hrTrend,
		SysLow:        sysLow,
		SysHigh:       sysHigh,
		DiaLow:        diaLow,
		DiaHigh:       diaHigh,
		HRLow:         hrLow,
		HRHigh:        hrHigh,
	}
}

// ─── 处方提示 ──────────────────────────────────────────────────────────────

func buildPrescriptionHints(labs []LabIndicator, weight WeightUFContext, lastTx LastTreatmentContext, naClr SodiumClearanceContext, vitals VitalSignsContext) []PrescriptionHint {
	hints := make([]PrescriptionHint, 0, 4)

	for _, lab := range labs {
		if lab.Value == nil {
			continue
		}
		v := *lab.Value
		switch lab.ConceptID {
		case "SERUM_K":
			if lab.Status == "high" || lab.Status == "critical_high" {
				hints = append(hints, PrescriptionHint{1, "⚡", "血钾偏高",
					fmt.Sprintf("K⁺ %.1f mmol/L，建议使用低钾透析液（K=2.0），确认处方钾浓度", v)})
			} else if lab.Status == "low" || lab.Status == "critical_low" {
				hints = append(hints, PrescriptionHint{1, "⚡", "血钾偏低",
					fmt.Sprintf("K⁺ %.1f mmol/L，注意透析液钾浓度不宜过低，预防心律失常", v)})
			}
		case "HEMOGLOBIN":
			if lab.Status == "low" || lab.Status == "critical_low" {
				hints = append(hints, PrescriptionHint{2, "📊", "贫血",
					fmt.Sprintf("Hb %.0f g/L 偏低，评估 EPO 剂量和铁储备状态", v)})
			}
		case "SERUM_P":
			if lab.Status == "high" || lab.Status == "critical_high" {
				hints = append(hints, PrescriptionHint{2, "🦴", "血磷偏高",
					fmt.Sprintf("P %.2f mmol/L，评估磷结合剂依从性和饮食管理", v)})
			}
		case "KTV":
			if lab.Status == "watch" || lab.Status == "low" {
				extra := ""
				if lastTx.ActualMinutes != nil && lastTx.PlannedMinutes != nil &&
					*lastTx.ActualMinutes < *lastTx.PlannedMinutes {
					extra = fmt.Sprintf("，上次短透 %d min", *lastTx.PlannedMinutes-*lastTx.ActualMinutes)
				}
				hints = append(hints, PrescriptionHint{2, "⏱", "透析充分性临界",
					fmt.Sprintf("Kt/V %.2f 接近下限%s，考虑延长透析时长或提高血流量", v, extra)})
			}
		}
	}

	if weight.UFRatePerKg != nil {
		switch weight.UFRateStatus {
		case "danger":
			hints = append(hints, PrescriptionHint{1, "💧", "超滤速率过高",
				fmt.Sprintf("超滤率约 %.1f mL/kg/h，超过报警阈值(13)，建议延长透析时间或下调超滤目标", *weight.UFRatePerKg)})
		case "watch":
			hints = append(hints, PrescriptionHint{3, "💧", "超滤速率偏高",
				fmt.Sprintf("超滤率约 %.1f mL/kg/h，进入警戒区(>11)，注意低血压风险", *weight.UFRatePerKg)})
		}
	}

	// 钠清除比提示（专利核心：评估脱钠空间）
	if naClr.LastRatio != nil {
		switch naClr.Status {
		case "high":
			hints = append(hints, PrescriptionHint{2, "🧂", "钠清除比偏高",
				fmt.Sprintf("上次钠清除比 %.2f（>%.2f），钠池仍充盈，本次可维持或加大脱钠力度（δ↑）", *naClr.LastRatio, naClr.TargetHigh)})
		case "low":
			hints = append(hints, PrescriptionHint{2, "🧂", "钠清除比偏低",
				fmt.Sprintf("上次钠清除比 %.2f（<%.2f），脱钠已趋平台，警惕过度脱钠，δ 宜收敛", *naClr.LastRatio, naClr.TargetLow)})
		}
	}

	// 血压提示（耐受度评估）
	if vitals.LastSystolic != nil {
		switch vitals.BPStatus {
		case "high":
			hints = append(hints, PrescriptionHint{2, "🫀", "透析平均血压偏高",
				fmt.Sprintf("上次透析平均 %.0f/%.0f mmHg，容量负荷偏重，可结合脱钠/超滤目标管理", *vitals.LastSystolic, *vitals.LastDiastolic)})
		case "low":
			hints = append(hints, PrescriptionHint{1, "🫀", "透析平均血压偏低",
				fmt.Sprintf("上次透析平均 %.0f/%.0f mmHg，耐受度差，慎用强力脱钠/高超滤，干体重可能设低", *vitals.LastSystolic, *vitals.LastDiastolic)})
		}
	}

	// 按优先级升序排列
	for i := 0; i < len(hints); i++ {
		for j := i + 1; j < len(hints); j++ {
			if hints[j].Priority < hints[i].Priority {
				hints[i], hints[j] = hints[j], hints[i]
			}
		}
	}
	return hints
}

func fetchDryWeightContext(_ *gin.Context, pid int64) *DwContextData {
	dws := services.NewDryWeightService()
	current, err := dws.Current(pid)
	if err != nil || current == nil {
		return &DwContextData{
			Phase:        "induction",
			SuggestedRNa: 1.05,
		}
	}
	if current.DryWeight == nil {
		return &DwContextData{
			Phase:        current.Phase,
			SuggestedRNa: current.SuggestedRNa,
		}
	}
	return &DwContextData{
		DryWeight:    current.DryWeight,
		Phase:        current.Phase,
		SuggestedRNa: current.SuggestedRNa,
	}
}
