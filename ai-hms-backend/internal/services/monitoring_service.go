package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/idh"
)

type MonitoringService struct {
	idhScorer idh.Scorer
}

func NewMonitoringService() *MonitoringService {
	return &MonitoringService{idhScorer: idh.StubScorer{}}
}

// SetIDHScorer 注入真 IDH 评分器（如 idh.NewHTTPScorer 接 Python 微服务）。插拔点。
func (s *MonitoringService) SetIDHScorer(scorer idh.Scorer) {
	s.idhScorer = scorer
}

func (s *MonitoringService) idhScorerOrStub() idh.Scorer {
	if s.idhScorer == nil {
		return idh.StubScorer{}
	}
	return s.idhScorer
}

type MonitoringLiveDevice struct {
	TreatmentID       int64     `json:"treatmentId"`
	PatientID         int64     `json:"patientId"`
	PatientName       string    `json:"patientName"`
	Age               int       `json:"age"`
	DialysisNo        string    `json:"dialysisNo"`
	BedID             int64     `json:"bedId"`
	BedName           string    `json:"bedName"`
	WardID            int64     `json:"wardId"`
	WardName          string    `json:"wardName"`
	Status            string    `json:"status"`
	StartTime         time.Time `json:"startTime"`
	EstimatedDuration float64   `json:"estimatedDuration"`
	DryWeight         float64   `json:"dryWeight"`
	DialysisMode      string    `json:"dialysisMode"`

	SBP              float64 `json:"sbp"`
	DBP              float64 `json:"dbp"`
	HeartRate        float64 `json:"heartRate"`
	Respiration      float64 `json:"respiration"`
	SpO2             float64 `json:"spO2"`
	BF               float64 `json:"bf"`
	TMP              float64 `json:"tmp"`
	UFVolume         float64 `json:"ufVolume"`
	UFGoal           float64 `json:"ufGoal"`
	Conductivity     float64 `json:"conductivity"`
	MachineTmp       float64 `json:"machineTmp"`
	ArterialPressure float64 `json:"arterialPressure"`
	VenousPressure   float64 `json:"venousPressure"`

	AccessType string            `json:"accessType"`
	AlarmLevel string            `json:"alarmLevel"`
	Alerts     []MonitoringAlert `json:"alerts"`

	IDHRisk idh.RiskResult `json:"idhRisk"`

	RNaCompletion RNaCompletion `json:"rnaCompletion"`

	VitalsSeries []VitalSample `json:"vitalsSeries"`
}

// VitalSample 整场体征序列的单点。
type VitalSample struct {
	T    time.Time `json:"t"`
	SBP  float64   `json:"sbp"`
	DBP  float64   `json:"dbp"`
	MAP  float64   `json:"map"`
	HR   float64   `json:"hr"`
	Kind string    `json:"kind"` // actual | predicted
}

// MonitoringAlert 单指标告警明细。
type MonitoringAlert struct {
	Metric string  `json:"metric"` // map | heartRate | vp | dialysateNa | ufr
	Level  string  `json:"level"`  // warning | danger
	Value  float64 `json:"value"`
}

// 透析液钠 ≈ 电导率 × 该系数（mmol/L）。来源经验，后续可移入配置。
const dialysateNaConductivityFactor = 9.9

// extrapolateVitals 简单外推占位（决②）：用最近窗口的线性斜率（限幅）把 MAP/SBP/DBP/心率
// 向「计划下机 plannedEnd」投影，生成 kind=predicted 的点（卡面渲染虚线）。
// 这是**占位**，真预测由 IDH/预测模型替换；故斜率与取值均保守夹紧，避免离谱投影。
// 首个 predicted 点 = 末个 actual 点（桥接，使虚线接上实线）。
func extrapolateVitals(actual []VitalSample, plannedEnd time.Time) []VitalSample {
	n := len(actual)
	if n < 2 {
		return nil
	}
	last := actual[n-1]
	if !plannedEnd.After(last.T) {
		return nil
	}
	win := 3
	if n < win {
		win = n
	}
	first := actual[n-win]
	dtMin := last.T.Sub(first.T).Minutes()
	if dtMin <= 0 {
		return nil
	}
	slope := func(a, b float64) float64 {
		s := (b - a) / dtMin
		if s > 0.5 {
			s = 0.5
		}
		if s < -0.5 {
			s = -0.5
		}
		return s
	}
	clamp := func(v, lo, hi float64) float64 {
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	}
	sSBP, sDBP, sHR := slope(first.SBP, last.SBP), slope(first.DBP, last.DBP), slope(first.HR, last.HR)

	out := []VitalSample{{T: last.T, SBP: last.SBP, DBP: last.DBP, MAP: last.MAP, HR: last.HR, Kind: "predicted"}}
	step := 30 * time.Minute
	for tcur := last.T.Add(step); !tcur.After(plannedEnd); tcur = tcur.Add(step) {
		m := tcur.Sub(last.T).Minutes()
		vs := VitalSample{T: tcur, Kind: "predicted"}
		if last.SBP > 0 {
			vs.SBP = clamp(last.SBP+sSBP*m, 50, 220)
		}
		if last.DBP > 0 {
			vs.DBP = clamp(last.DBP+sDBP*m, 30, 130)
		}
		if vs.SBP > 0 && vs.DBP > 0 {
			vs.MAP = (vs.SBP + 2*vs.DBP) / 3
		}
		if last.HR > 0 {
			vs.HR = clamp(last.HR+sHR*m, 30, 180)
		}
		out = append(out, vs)
	}
	if len(out) <= 1 {
		return nil
	}
	return out
}

// ageFromBirth 由出生日期粗算周岁；缺失返回 0。
func ageFromBirth(birth *time.Time) int {
	if birth == nil || birth.IsZero() {
		return 0
	}
	now := time.Now()
	age := now.Year() - birth.Year()
	if now.YearDay() < birth.YearDay() {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}

// dialysisDurationMinutes 把 "Plan_PatientPrescription"."DialysisDuration" 归一为分钟。
// 老库该列单位为小时（实测取值 1..8，均值约 4）；<=24 视为小时，>24 视为分钟，<=0 取默认 4h。
func dialysisDurationMinutes(v float64) float64 {
	if v <= 0 {
		return 240
	}
	if v <= 24 {
		return v * 60
	}
	return v
}

// mlToL 把 "Device_DMLog" 的 ml 量纲转为 L（<=0 返回 0）。
func mlToL(v float64) float64 {
	if v <= 0 {
		return 0
	}
	return v / 1000
}

func (s *MonitoringService) GetLiveData(tenantID int64) ([]MonitoringLiveDevice, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not available")
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	type treatmentRow struct {
		ID               int64      `gorm:"column:Id"`
		PatientID        int64      `gorm:"column:PatientId"`
		BedID            int64      `gorm:"column:BedId"`
		WardID           int64      `gorm:"column:WardId"`
		Status           string     `gorm:"column:Status"`
		StartTime        time.Time  `gorm:"column:StartTime"`
		PatientName      string     `gorm:"column:PatientName"`
		DialysisNo       string     `gorm:"column:DialysisNo"`
		BirthDate        *time.Time `gorm:"column:BirthDate"`
		Gender           string     `gorm:"column:Gender"`
		Height           *float64   `gorm:"column:Height"`
		BedName          string     `gorm:"column:BedName"`
		WardName         string     `gorm:"column:WardName"`
		DialysisDuration float64    `gorm:"column:DialysisDuration"`
		DryWeight        float64    `gorm:"column:DryWeight"`
		DialysisMethod   string     `gorm:"column:DialysisMethod"`
	}

	var rows []treatmentRow
	err := db.Table(`"Treatment_Treatment" AS t`).
		Select(`t."Id", t."PatientId", COALESCE(t."BedId", 0) AS "BedId", COALESCE(t."WardId", 0) AS "WardId", t."Status", COALESCE(t."StartTime", t."CreateTime") AS "StartTime",
			COALESCE(p."Name", '') AS "PatientName",
			COALESCE(p."DialysisNo", '') AS "DialysisNo",
			p."BirthDate" AS "BirthDate",
			COALESCE(p."Gender", '') AS "Gender",
			p."Height" AS "Height",
			COALESCE(b."Name", '') AS "BedName",
			COALESCE(w."Name", '') AS "WardName",
			COALESCE(
				(SELECT rx."DialysisDuration" FROM "Plan_PatientPrescription" rx
				 WHERE rx."TreatmentId" = t."Id" AND rx."TenantId" = t."TenantId" AND rx."DialysisDuration" IS NOT NULL
				 ORDER BY rx."CreateTime" DESC LIMIT 1),
				pl."DialysisDuration", 240) AS "DialysisDuration",
			COALESCE(pl."DryWeight", 0) AS "DryWeight",
			COALESCE(pl."DialysisMethod", '') AS "DialysisMethod"`).
		Joins(`LEFT JOIN "Register_PatientInfomation" AS p ON p."Id" = t."PatientId" AND p."TenantId" = t."TenantId"`).
		Joins(`LEFT JOIN "Schedule_Bed" AS b ON b."Id" = t."BedId" AND b."TenantId" = t."TenantId"`).
		Joins(`LEFT JOIN "Schedule_Ward" AS w ON w."Id" = t."WardId" AND w."TenantId" = t."TenantId"`).
		Joins(`LEFT JOIN "Plan_PatientPlan" AS pl ON pl."PatientId" = t."PatientId" AND pl."TenantId" = t."TenantId" AND COALESCE(pl."IsDisabled", false) = false`).
		Where(`t."TenantId" = ? AND t."StartTime" IS NOT NULL AND t."EndTime" IS NULL AND t."StartTime" >= ? AND t."StartTime" < ?`,
			tenantID, todayStart, todayStart.AddDate(0, 0, 1)).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []MonitoringLiveDevice{}, nil
	}

	treatmentIDs := make([]int64, len(rows))
	patientIDs := make([]int64, len(rows))
	for i, r := range rows {
		treatmentIDs[i] = r.ID
		patientIDs[i] = r.PatientID
	}

	// Latest signs
	type signRow struct {
		TreatmentID int64   `gorm:"column:TreatmentId"`
		SBP         float64 `gorm:"column:SBP"`
		DBP         float64 `gorm:"column:DBP"`
		HeartRate   float64 `gorm:"column:HeartRate"`
		Respiration float64 `gorm:"column:Respiration"`
		SpO2        float64 `gorm:"column:SpO2"`
	}
	var signs []signRow
	db.Table(`"Treatment_DuringSigns"`).
		Select(`"TreatmentId", COALESCE("SBP", 0) AS "SBP", COALESCE("DBP", 0) AS "DBP", COALESCE("HeartRate", 0) AS "HeartRate", COALESCE("Respiration", 0) AS "Respiration", COALESCE("SpO2", 0) AS "SpO2"`).
		Where(`"TreatmentId" IN ? AND "TenantId" = ?`, treatmentIDs, tenantID).
		Order(`"OperateTime" DESC`).
		Find(&signs)
	signMap := map[int64]signRow{}
	for _, r := range signs {
		if _, ok := signMap[r.TreatmentID]; !ok {
			signMap[r.TreatmentID] = r
		}
	}

	// Latest params from DuringParam (BF, TMP, Conductivity, MachineTmp, pressures)
	type paramRow struct {
		TreatmentID      int64   `gorm:"column:TreatmentId"`
		BF               float64 `gorm:"column:BF"`
		TMP              float64 `gorm:"column:TMP"`
		Conductivity     float64 `gorm:"column:Conductivity"`
		MachineTmp       float64 `gorm:"column:MachineTmp"`
		ArterialPressure float64 `gorm:"column:ArterialPressure"`
		VenousPressure   float64 `gorm:"column:VenousPressure"`
	}
	var params []paramRow
	db.Table(`"Treatment_DuringParam"`).
		Select(`"TreatmentId", COALESCE("BF", 0) AS "BF", COALESCE("TMP", 0) AS "TMP", COALESCE("Conductivity", 0) AS "Conductivity", COALESCE("MachineTmp", 0) AS "MachineTmp", COALESCE("ArterialPressure", 0) AS "ArterialPressure", COALESCE("VenousPressure", 0) AS "VenousPressure"`).
		Where(`"TreatmentId" IN ? AND "TenantId" = ?`, treatmentIDs, tenantID).
		Order(`"OperateTime" DESC`).
		Find(&params)
	paramMap := map[int64]paramRow{}
	for _, r := range params {
		if _, ok := paramMap[r.TreatmentID]; !ok {
			paramMap[r.TreatmentID] = r
		}
	}

	// Device_DMLog 实时数据
	type deviceRow struct {
		TreatmentID   int64   `gorm:"column:TreatmentId"`
		UFSetVolume   float64 `gorm:"column:UFSetVolume"`
		UFVolume      float64 `gorm:"column:UFVolume"`
		TreatmentTime float64 `gorm:"column:TreatmentTime"`
	}
	var devices []deviceRow
	db.Table(`"Device_DMLog"`).
		Select(`DISTINCT ON ("TreatmentId") "TreatmentId", COALESCE("UFSetVolume", 0) AS "UFSetVolume", COALESCE("UFVolume", 0) AS "UFVolume", COALESCE("TreatmentTime", 0) AS "TreatmentTime"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, tenantID, treatmentIDs).
		Order(`"TreatmentId", "LogTime" DESC`).
		Find(&devices)
	deviceMap := map[int64]deviceRow{}
	for _, r := range devices {
		if _, ok := deviceMap[r.TreatmentID]; !ok {
			deviceMap[r.TreatmentID] = r
		}
	}

	// 通路类型（VP 分层需要）
	type accessRow struct {
		PatientID  int64  `gorm:"column:PatientId"`
		AccessType string `gorm:"column:AccessType"`
	}
	var accesses []accessRow
	db.Table(`"Register_VascularAccess"`).
		Select(`"PatientId", COALESCE("AccessType", '') AS "AccessType"`).
		Where(`"PatientId" IN ? AND "TenantId" = ? AND COALESCE("IsDisabled", false) = false`, patientIDs, tenantID).
		Order(`"PatientId", "IsDefault" DESC, "OperationTime" DESC`).
		Find(&accesses)
	accessMap := map[int64]string{}
	for _, a := range accesses {
		if _, ok := accessMap[a.PatientID]; !ok {
			accessMap[a.PatientID] = a.AccessType
		}
	}

	// 报警阈值表
	thresholds, _ := config.LoadMonitoringThresholds()

	// 卡面整场曲线：取各治疗的 DuringSigns 全部行（升序）
	type signSeriesRow struct {
		TreatmentID int64     `gorm:"column:TreatmentId"`
		OperateTime time.Time `gorm:"column:OperateTime"`
		SBP         *float64  `gorm:"column:SBP"`
		DBP         *float64  `gorm:"column:DBP"`
		HeartRate   *float64  `gorm:"column:HeartRate"`
	}
	var signSeries []signSeriesRow
	db.Table(`"Treatment_DuringSigns"`).
		Select(`"TreatmentId", "OperateTime", "SBP", "DBP", "HeartRate"`).
		Where(`"TreatmentId" IN ? AND "TenantId" = ?`, treatmentIDs, tenantID).
		Order(`"TreatmentId", "OperateTime" ASC`).
		Find(&signSeries)
	vitalsSeriesMap := map[int64][]VitalSample{}
	for _, r := range signSeries {
		if r.SBP == nil && r.DBP == nil && r.HeartRate == nil {
			continue
		}
		vs := VitalSample{T: r.OperateTime, Kind: "actual"}
		if r.SBP != nil {
			vs.SBP = *r.SBP
		}
		if r.DBP != nil {
			vs.DBP = *r.DBP
		}
		if r.SBP != nil && r.DBP != nil {
			vs.MAP = (*r.SBP + 2*(*r.DBP)) / 3
		}
		if r.HeartRate != nil {
			vs.HR = *r.HeartRate
		}
		vitalsSeriesMap[r.TreatmentID] = append(vitalsSeriesMap[r.TreatmentID], vs)
	}

	// 床卡曲线点数上限保护：单床超出上限按等间距抽样，末点始终保留（外推需要最新值）。
	const vitalsSeriesMaxPerBed = 120
	for tid, series := range vitalsSeriesMap {
		if len(series) <= vitalsSeriesMaxPerBed {
			continue
		}
		stride := len(series) / vitalsSeriesMaxPerBed
		if stride < 1 {
			stride = 1
		}
		sampled := make([]VitalSample, 0, vitalsSeriesMaxPerBed+1)
		for i := 0; i < len(series); i += stride {
			sampled = append(sampled, series[i])
		}
		last := series[len(series)-1]
		if len(sampled) == 0 || !sampled[len(sampled)-1].T.Equal(last.T) {
			sampled = append(sampled, last)
		}
		vitalsSeriesMap[tid] = sampled
	}

	// ---- 实时 RNa 完成率输入 ----
	// C_pre：每患者最近一次血清钠（来源：老库 LIS_Examination + LIS_ExaminationItem）
	type cPreVal struct {
		V  float64
		At time.Time
	}
	cPreMap := map[int64]cPreVal{}
	{
		type cPreRow struct {
			PatientID int64     `gorm:"column:PatientId"`
			Value     string    `gorm:"column:Result"`
			TestedAt  time.Time `gorm:"column:ResultTime"`
		}
		var labs []cPreRow
		_ = db.Raw(`
			SELECT DISTINCT ON (e."PatientId") e."PatientId" AS "PatientId",
			       i."Result" AS "Result", e."ResultTime" AS "ResultTime"
			FROM "LIS_ExaminationItem" i
			JOIN "LIS_Examination" e ON e."Id" = i."ExaminationId" AND e."TenantId" = i."TenantId"
			WHERE e."TenantId" = ? AND e."PatientId" IN ?
			  AND (i."ItemCode" ILIKE 'NA' OR i."ItemName" LIKE '%钠%')
			ORDER BY e."PatientId", e."ResultTime" DESC`, tenantID, patientIDs).Scan(&labs)
		for _, r := range labs {
			cleaned := strings.TrimSpace(r.Value)
			for _, ch := range cleaned {
				if ch == '.' || (ch >= '0' && ch <= '9') {
					continue
				}
				cleaned = strings.ReplaceAll(cleaned, string(ch), " ")
			}
			fields := strings.Fields(cleaned)
			parsedVal := ""
			if len(fields) > 0 {
				parsedVal = fields[0]
			}
			v, verr := strconv.ParseFloat(parsedVal, 64)
			if verr == nil && v > 0 {
				cPreMap[r.PatientID] = cPreVal{V: v, At: r.TestedAt}
			}
		}
	}

	// 处方 RNa 目标：老库 dry_weight_assessment 不存在，Treatment_NA_Memo.R_Na 全为空。
	// 暂无可靠处方 RNa 落库来源，targetRNaMap 保持空 → RNa 完成率稳定降级为 Available=false。
	// TODO: 待确认 RNa 处方落库位置（新表或 Treatment_NA_Memo 写入逻辑）后接入。
	targetRNaMap := map[int64]float64{}

	// 透前体重
	preWeightMap := map[int64]float64{}
	{
		type pwRow struct {
			TreatmentID int64   `gorm:"column:TreatmentId"`
			Weight      float64 `gorm:"column:Weight"`
		}
		var pws []pwRow
		db.Table(`"Treatment_BeforeSigns"`).
			Select(`"TreatmentId", COALESCE("Weight", 0) AS "Weight"`).
			Where(`"TreatmentId" IN ? AND "TenantId" = ?`, treatmentIDs, tenantID).
			Order(`"TreatmentId", "CreateTime" DESC`).
			Find(&pws)
		for _, r := range pws {
			if _, ok := preWeightMap[r.TreatmentID]; !ok && r.Weight > 0 {
				preWeightMap[r.TreatmentID] = r.Weight
			}
		}
	}

	// 整场实测透析液钠均值 = 电导率均值 × 系数
	meanCdMap := map[int64]float64{}
	{
		type mcRow struct {
			TreatmentID int64   `gorm:"column:TreatmentId"`
			AvgCond     float64 `gorm:"column:avg_cond"`
		}
		var mcs []mcRow
		db.Table(`"Device_DMLog"`).
			Select(`"TreatmentId", AVG("Conductivity") AS avg_cond`).
			Where(`"TenantId" = ? AND "TreatmentId" IN ? AND "Conductivity" IS NOT NULL`, tenantID, treatmentIDs).
			Group(`"TreatmentId"`).
			Find(&mcs)
		for _, r := range mcs {
			if r.AvgCond > 0 {
				meanCdMap[r.TreatmentID] = r.AvgCond * dialysateNaConductivityFactor
			}
		}
	}

	result := make([]MonitoringLiveDevice, 0, len(rows))
	for _, r := range rows {
		d := MonitoringLiveDevice{
			TreatmentID:       r.ID,
			PatientID:         r.PatientID,
			PatientName:       r.PatientName,
			Age:               ageFromBirth(r.BirthDate),
			DialysisNo:        r.DialysisNo,
			BedID:             r.BedID,
			BedName:           r.BedName,
			WardID:            r.WardID,
			WardName:          r.WardName,
			Status:            r.Status,
			StartTime:         r.StartTime,
			EstimatedDuration: dialysisDurationMinutes(r.DialysisDuration),
			DryWeight:         r.DryWeight,
			DialysisMode:      r.DialysisMethod,
		}
		if s, ok := signMap[r.ID]; ok {
			d.SBP = s.SBP
			d.DBP = s.DBP
			d.HeartRate = s.HeartRate
			d.Respiration = s.Respiration
			d.SpO2 = s.SpO2
		}
		if p, ok := paramMap[r.ID]; ok {
			d.BF = p.BF
			d.TMP = p.TMP
			d.Conductivity = p.Conductivity
			d.MachineTmp = p.MachineTmp
			d.ArterialPressure = p.ArterialPressure
			d.VenousPressure = p.VenousPressure
		}
		if dev, ok := deviceMap[r.ID]; ok {
			d.UFGoal = mlToL(dev.UFSetVolume)
			d.UFVolume = mlToL(dev.UFVolume)
		}

		d.AccessType = accessMap[r.PatientID]
		d.VitalsSeries = vitalsSeriesMap[r.ID]
		if pred := extrapolateVitals(d.VitalsSeries, r.StartTime.Add(time.Duration(dialysisDurationMinutes(r.DialysisDuration))*time.Minute)); len(pred) > 0 {
			d.VitalsSeries = append(append([]VitalSample{}, d.VitalsSeries...), pred...)
		}
		s.evalAlarms(&d, thresholds)
		d.IDHRisk = s.idhScorerOrStub().Score(context.Background(), idh.RiskInput{TreatmentID: r.ID, AccessType: d.AccessType})

		if cp := cPreMap[r.PatientID]; cp.V > 0 {
			targetRNa := targetRNaMap[r.PatientID]
			meanCd := meanCdMap[r.ID]
			var elapsedH float64
			if dev, ok := deviceMap[r.ID]; ok {
				elapsedH = dev.TreatmentTime / 60
			}
			vuf := preWeightMap[r.ID] - r.DryWeight
			if vuf <= 0 {
				vuf = d.UFGoal
			}
			if targetRNa > 0 && r.Height != nil && *r.Height > 0 && r.DryWeight > 0 && meanCd > 0 && elapsedH > 0 && vuf > 0 {
				presc := CalculateRNaPrescription(RNaCalculateRequest{
					CPre: cp.V, DryWeight: r.DryWeight, HeightCm: *r.Height,
					AgeYears: float64(d.Age),
					IsMale:   r.Gender == "男" || strings.EqualFold(r.Gender, "M") || strings.EqualFold(r.Gender, "Male"),
					VUF:      &vuf, Driver: "rna", RNa: targetRNa,
				})
				rc := computeRNaCompletion(RNaCompletionInput{
					Presc: presc, CPre: cp.V, UFActual: d.UFVolume, MeanCd: meanCd, ElapsedH: elapsedH,
				})
				rc.CPreAt = cp.At.Format("2006-01-02")
				d.RNaCompletion = rc
			}
		}

		result = append(result, d)
	}
	return result, nil
}

// evalAlarms 按阈值表算每床各指标分级，写入 d.Alerts，并取最严重为 d.AlarmLevel。
// 缺失/无效读数（0 或缺前置条件）跳过，不报警。
func (s *MonitoringService) evalAlarms(d *MonitoringLiveDevice, th *config.MonitoringThresholds) {
	if th == nil {
		return
	}
	var levels []config.AlarmLevel
	add := func(metric string, level config.AlarmLevel, value float64) {
		levels = append(levels, level)
		if level != config.AlarmNormal {
			d.Alerts = append(d.Alerts, MonitoringAlert{Metric: metric, Level: string(level), Value: value})
		}
	}

	if d.SBP > 0 && d.DBP > 0 {
		mapVal := (d.SBP + 2*d.DBP) / 3
		add("map", th.EvalFixed("map", mapVal), mapVal)
	}
	if d.HeartRate > 0 {
		add("heartRate", th.EvalFixed("heartRate", d.HeartRate), d.HeartRate)
	}
	if d.VenousPressure > 0 {
		add("vp", th.EvalVP(d.AccessType, d.BF, d.VenousPressure), d.VenousPressure)
	}
	if d.Conductivity > 0 {
		na := d.Conductivity * dialysateNaConductivityFactor
		add("dialysateNa", th.EvalFixed("dialysateNa", na), na)
	}
	if d.UFGoal > 0 && d.DryWeight > 0 && d.EstimatedDuration > 0 {
		ufr := d.UFGoal * 1000 / d.DryWeight / (d.EstimatedDuration / 60)
		add("ufr", th.EvalFixed("ufr", ufr), ufr)
	}

	d.AlarmLevel = string(config.WorstLevel(levels...))
}

// ----- 整场趋势接口（决①） -----

type TrendPoint struct {
	T    time.Time `json:"t"`
	V    float64   `json:"v"`
	Kind string    `json:"kind"` // actual | predicted
}

type TreatmentTrend struct {
	TreatmentID int64                   `json:"treatmentId"`
	Start       time.Time               `json:"start"`
	Now         time.Time               `json:"now"`
	PlannedEnd  time.Time               `json:"plannedEnd"`
	Series      map[string][]TrendPoint `json:"series"`
}

const trendMaxDevicePoints = 360

func (s *MonitoringService) GetTreatmentTrend(tenantID, treatmentID int64) (*TreatmentTrend, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not available")
	}

	type basisRow struct {
		StartTime        time.Time `gorm:"column:StartTime"`
		DialysisDuration float64   `gorm:"column:DialysisDuration"`
	}
	var basis basisRow
	err := db.Table(`"Treatment_Treatment" AS t`).
		Select(`COALESCE(t."StartTime", t."CreateTime") AS "StartTime",
			COALESCE(
				(SELECT rx."DialysisDuration" FROM "Plan_PatientPrescription" rx
				 WHERE rx."TreatmentId" = t."Id" AND rx."TenantId" = t."TenantId" AND rx."DialysisDuration" IS NOT NULL
				 ORDER BY rx."CreateTime" DESC LIMIT 1),
				pl."DialysisDuration", 240) AS "DialysisDuration"`).
		Joins(`LEFT JOIN "Plan_PatientPlan" AS pl ON pl."PatientId" = t."PatientId" AND pl."TenantId" = t."TenantId" AND COALESCE(pl."IsDisabled", false) = false`).
		Where(`t."Id" = ? AND t."TenantId" = ?`, treatmentID, tenantID).
		Take(&basis).Error
	if err != nil {
		return nil, errors.New("治疗记录不存在")
	}

	out := &TreatmentTrend{
		TreatmentID: treatmentID,
		Start:       basis.StartTime,
		Now:         time.Now(),
		PlannedEnd:  basis.StartTime.Add(time.Duration(dialysisDurationMinutes(basis.DialysisDuration)) * time.Minute),
		Series:      map[string][]TrendPoint{},
	}

	// 生命体征（稀疏，护士观测）
	type signTrendRow struct {
		OperateTime time.Time `gorm:"column:OperateTime"`
		SBP         *float64  `gorm:"column:SBP"`
		DBP         *float64  `gorm:"column:DBP"`
		HeartRate   *float64  `gorm:"column:HeartRate"`
	}
	var signs []signTrendRow
	db.Table(`"Treatment_DuringSigns"`).
		Select(`"OperateTime", "SBP", "DBP", "HeartRate"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, tenantID).
		Order(`"OperateTime" ASC`).
		Find(&signs)
	for _, r := range signs {
		if r.SBP != nil {
			out.Series["sbp"] = append(out.Series["sbp"], TrendPoint{T: r.OperateTime, V: *r.SBP, Kind: "actual"})
		}
		if r.DBP != nil {
			out.Series["dbp"] = append(out.Series["dbp"], TrendPoint{T: r.OperateTime, V: *r.DBP, Kind: "actual"})
		}
		if r.HeartRate != nil {
			out.Series["heartRate"] = append(out.Series["heartRate"], TrendPoint{T: r.OperateTime, V: *r.HeartRate, Kind: "actual"})
		}
		if r.SBP != nil && r.DBP != nil {
			out.Series["map"] = append(out.Series["map"], TrendPoint{T: r.OperateTime, V: (*r.SBP + 2*(*r.DBP)) / 3, Kind: "actual"})
		}
	}

	// 设备参数（密集，DMLog 前置机流）
	type deviceTrendRow struct {
		LogTime          time.Time `gorm:"column:LogTime"`
		VenousPressure   *float64  `gorm:"column:VenousPressure"`
		ArterialPressure *float64  `gorm:"column:ArterialPressure"`
		TMP              *float64  `gorm:"column:TMP"`
		BF               *float64  `gorm:"column:BF"`
		Conductivity     *float64  `gorm:"column:Conductivity"`
		UFVolume         *float64  `gorm:"column:UFVolume"`
	}
	var devices []deviceTrendRow
	db.Table(`"Device_DMLog"`).
		Select(`"LogTime", "VenousPressure", "ArterialPressure", "TMP", "BF", "Conductivity", "UFVolume"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, tenantID).
		Order(`"LogTime" ASC`).
		Find(&devices)

	stride := 1
	if len(devices) > trendMaxDevicePoints {
		stride = (len(devices) + trendMaxDevicePoints - 1) / trendMaxDevicePoints
	}
	for i := 0; i < len(devices); i += stride {
		r := devices[i]
		if r.VenousPressure != nil {
			out.Series["vp"] = append(out.Series["vp"], TrendPoint{T: r.LogTime, V: *r.VenousPressure, Kind: "actual"})
		}
		if r.ArterialPressure != nil {
			out.Series["ap"] = append(out.Series["ap"], TrendPoint{T: r.LogTime, V: *r.ArterialPressure, Kind: "actual"})
		}
		if r.TMP != nil {
			out.Series["tmp"] = append(out.Series["tmp"], TrendPoint{T: r.LogTime, V: *r.TMP, Kind: "actual"})
		}
		if r.BF != nil {
			out.Series["bf"] = append(out.Series["bf"], TrendPoint{T: r.LogTime, V: *r.BF, Kind: "actual"})
		}
		if r.Conductivity != nil {
			out.Series["conductivity"] = append(out.Series["conductivity"], TrendPoint{T: r.LogTime, V: *r.Conductivity, Kind: "actual"})
		}
		if r.UFVolume != nil {
			out.Series["ufVolume"] = append(out.Series["ufVolume"], TrendPoint{T: r.LogTime, V: *r.UFVolume, Kind: "actual"})
		}
	}

	return out, nil
}
