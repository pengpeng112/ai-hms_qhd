package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
)

type MonitoringService struct{}

func NewMonitoringService() *MonitoringService {
	return &MonitoringService{}
}

type MonitoringLiveDevice struct {
	TreatmentID        int64     `json:"treatmentId"`
	PatientID          int64     `json:"patientId"`
	PatientName        string    `json:"patientName"`
	BedID              int64     `json:"bedId"`
	BedName            string    `json:"bedName"`
	WardID             int64     `json:"wardId"`
	WardName           string    `json:"wardName"`
	Status             string    `json:"status"`
	StartTime          time.Time `json:"startTime"`
	EstimatedDuration  float64   `json:"estimatedDuration"`
	DryWeight          float64   `json:"dryWeight"`
	DialysisMode       string    `json:"dialysisMode"`

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
}

func (s *MonitoringService) GetLiveData(tenantID int64) ([]MonitoringLiveDevice, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not available")
	}

	// 今日 [00:00, 次日 00:00)，用于限定实时监控只看当天在机的治疗。
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	type treatmentRow struct {
		ID               int64     `gorm:"column:Id"`
		PatientID        int64     `gorm:"column:PatientId"`
		BedID            int64     `gorm:"column:BedId"`
		WardID           int64     `gorm:"column:WardId"`
		Status           string    `gorm:"column:Status"`
		StartTime        time.Time `gorm:"column:StartTime"`
		PatientName      string    `gorm:"column:PatientName"`
		BedName          string    `gorm:"column:BedName"`
		WardName         string    `gorm:"column:WardName"`
		DialysisDuration float64   `gorm:"column:DialysisDuration"`
		DryWeight        float64   `gorm:"column:DryWeight"`
		DialysisMethod   string    `gorm:"column:DialysisMethod"`
	}

	var rows []treatmentRow
	err := db.Table(`"Treatment_Treatment" AS t`).
		Select(`t."Id", t."PatientId", COALESCE(t."BedId", 0) AS "BedId", COALESCE(t."WardId", 0) AS "WardId", t."Status", COALESCE(t."StartTime", t."CreateTime") AS "StartTime",
			COALESCE(p."Name", '') AS "PatientName",
			COALESCE(b."Name", '') AS "BedName",
			COALESCE(w."Name", '') AS "WardName",
			COALESCE(pl."DialysisDuration", 240) AS "DialysisDuration",
			COALESCE(pl."DryWeight", 0) AS "DryWeight",
			COALESCE(pl."DialysisMethod", '') AS "DialysisMethod"`).
		Joins(`LEFT JOIN "Register_PatientInfomation" AS p ON p."Id" = t."PatientId" AND p."TenantId" = t."TenantId"`).
		Joins(`LEFT JOIN "Schedule_Bed" AS b ON b."Id" = t."BedId" AND b."TenantId" = t."TenantId"`).
		Joins(`LEFT JOIN "Schedule_Ward" AS w ON w."Id" = t."WardId" AND w."TenantId" = t."TenantId"`).
		Joins(`LEFT JOIN "Plan_PatientPlan" AS pl ON pl."PatientId" = t."PatientId" AND pl."TenantId" = t."TenantId" AND COALESCE(pl."IsDisabled", false) = false`).
		// 不用 Status='30' 过滤：用时间判据「今日已上机未下机」，绕开状态码字典之争。
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
	for i, r := range rows {
		treatmentIDs[i] = r.ID
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

	// Device_DMLog 实时数据：取每个治疗最近一条设备数据，提取 UF 相关字段。
	// Device_DMLog 由透析机实时上传，是 UFSetVolume(目标) 和 UFVolume(实际) 的真实数据源。
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

	result := make([]MonitoringLiveDevice, 0, len(rows))
	for _, r := range rows {
		d := MonitoringLiveDevice{
			TreatmentID:       r.ID,
			PatientID:         r.PatientID,
			PatientName:       r.PatientName,
			BedID:             r.BedID,
			BedName:           r.BedName,
			WardID:            r.WardID,
			WardName:          r.WardName,
			Status:            r.Status,
			StartTime:         r.StartTime,
			EstimatedDuration: r.DialysisDuration,
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
		// UF数据来源修正：从 Device_DMLog 获取真实目标超滤量和实际超滤量。
		// 原 UFGoal = DryWeight (干体重) 是错误数据源。
		if dev, ok := deviceMap[r.ID]; ok {
			d.UFGoal = dev.UFSetVolume   // 设定超滤总量
			d.UFVolume = dev.UFVolume    // 实际累计超滤量
		}
		// Fallback: 无设备数据时 UFGoal/UFVolume 保持 0（前端触发"暂无超滤数据"降级）。
		result = append(result, d)
	}
	return result, nil
}
