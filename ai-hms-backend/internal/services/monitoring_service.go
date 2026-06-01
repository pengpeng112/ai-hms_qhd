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

	type treatmentRow struct {
		ID                int64     `gorm:"column:Id"`
		PatientID         int64     `gorm:"column:PatientId"`
		BedID             int64     `gorm:"column:BedId"`
		WardID            int64     `gorm:"column:WardId"`
		Status            string    `gorm:"column:Status"`
		StartTime         time.Time `gorm:"column:StartTime"`
		PatientName       string    `gorm:"column:PatientName"`
		BedName           string    `gorm:"column:BedName"`
		WardName          string    `gorm:"column:WardName"`
		DialysisDuration  float64   `gorm:"column:DialysisDuration"`
		DryWeight         float64   `gorm:"column:DryWeight"`
		DialysisMethod    string    `gorm:"column:DialysisMethod"`
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
		Joins(`LEFT JOIN "Schedule_Ward" AS w ON w."Id" = t."WardId"`).
		Joins(`LEFT JOIN "Plan_PatientPlan" AS pl ON pl."PatientId" = t."PatientId" AND pl."TenantId" = t."TenantId" AND COALESCE(pl."IsDisabled", false) = false`).
		Where(`t."TenantId" = ? AND t."Status" = ?`, tenantID, "30").
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

	// Latest params
	type paramRow struct {
		TreatmentID     int64   `gorm:"column:TreatmentId"`
		BF              float64 `gorm:"column:BF"`
		TMP             float64 `gorm:"column:TMP"`
		UFQuantity      float64 `gorm:"column:UFQuantity"`
		Conductivity    float64 `gorm:"column:Conductivity"`
		MachineTmp      float64 `gorm:"column:MachineTmp"`
		ArterialPressure float64 `gorm:"column:ArterialPressure"`
		VenousPressure  float64 `gorm:"column:VenousPressure"`
	}
	var params []paramRow
	db.Table(`"Treatment_DuringParam"`).
		Select(`"TreatmentId", COALESCE("BF", 0) AS "BF", COALESCE("TMP", 0) AS "TMP", COALESCE("UFQuantity", 0) AS "UFQuantity", COALESCE("Conductivity", 0) AS "Conductivity", COALESCE("MachineTmp", 0) AS "MachineTmp", COALESCE("ArterialPressure", 0) AS "ArterialPressure", COALESCE("VenousPressure", 0) AS "VenousPressure"`).
		Where(`"TreatmentId" IN ? AND "TenantId" = ?`, treatmentIDs, tenantID).
		Order(`"OperateTime" DESC`).
		Find(&params)
	paramMap := map[int64]paramRow{}
	for _, r := range params {
		if _, ok := paramMap[r.TreatmentID]; !ok {
			paramMap[r.TreatmentID] = r
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
			d.UFVolume = p.UFQuantity
			d.Conductivity = p.Conductivity
			d.MachineTmp = p.MachineTmp
			d.ArterialPressure = p.ArterialPressure
			d.VenousPressure = p.VenousPressure
		}
		d.UFGoal = r.DryWeight // UF Goal from Plan
		result = append(result, d)
	}
	return result, nil
}
