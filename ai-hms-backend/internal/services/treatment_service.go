package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"gorm.io/gorm"
)

type TreatmentService struct {
	db *gorm.DB
}

func NewTreatmentService() *TreatmentService {
	return &TreatmentService{db: database.GetDB()}
}

type TreatmentListRequest struct {
	Page               int                  `form:"page"`
	PageSize           int                  `form:"pageSize"`
	PatientId          *modeltypes.LegacyID `form:"patientId"`
	Status             *int                 `form:"status"`
	Type               *int                 `form:"type"`
	TreatmentDate      *time.Time           `form:"treatmentDate" time_format:"2006-01-02"`
	TreatmentDateStart *time.Time           `form:"treatmentDateStart" time_format:"2006-01-02"`
	TreatmentDateEnd   *time.Time           `form:"treatmentDateEnd" time_format:"2006-01-02"`
}

type TreatmentListResponse struct {
	Items     []TreatmentRealtimeResponse `json:"items"`
	Total     int64                       `json:"total"`
	Page      int                         `json:"page"`
	PageSize  int                         `json:"pageSize"`
	TotalPage int                         `json:"totalPage"`
}

type TreatmentRealtimeResponse struct {
	ID                 int64                         `json:"id"`
	TenantID           int64                         `json:"tenantId"`
	PatientID          string                        `json:"patientId"`
	WardID             *int64                        `json:"wardId,omitempty"`
	TreatmentDate      string                        `json:"treatmentDate"`
	ShiftID            *int64                        `json:"shiftId,omitempty"`
	TreatmentType      string                        `json:"treatmentType"`
	Status             int                           `json:"status"`
	LegacyStatus       string                        `json:"legacyStatus,omitempty"` // 老库原始状态码(10签到/20透前/30透中/40透后/50中断/60结束),供驾驶舱细分卡态;附加只读、不影响 Status
	StartTime          *time.Time                    `json:"startTime,omitempty"`
	EndTime            *time.Time                    `json:"endTime,omitempty"`
	Notes              string                        `json:"notes,omitempty"`
	CreatorID          int64                         `json:"creatorId"`
	CreateTime         time.Time                     `json:"createTime"`
	LastModifyTime     time.Time                     `json:"lastModifyTime"`
	DoctorSummary      string                        `json:"doctorSummary,omitempty"`
	TreatmentSummary   string                        `json:"treatmentSummary,omitempty"`
	TimeRange          string                        `json:"timeRange,omitempty"`
	DurationMinutes    int                           `json:"durationMinutes,omitempty"`
	WeightLossKG       float64                       `json:"weightLossKg,omitempty"`
	ShiftName          string                        `json:"shiftName,omitempty"`
	QueueNo            string                        `json:"queueNo,omitempty"`
	CaseStatus         string                        `json:"caseStatus,omitempty"`
	TmrPath            string                        `json:"tmrPath,omitempty"`
	TmrTime            *time.Time                    `json:"tmrTime,omitempty"`
	TmrPages           int                           `json:"tmrPages,omitempty"`
	DoctorName         string                        `json:"doctorName,omitempty"`
	StartBP            string                        `json:"startBp,omitempty"`
	EndBP              string                        `json:"endBp,omitempty"`
	Complications      string                        `json:"complications,omitempty"`
	BeforeSigns        *TreatmentBeforeSnapshot      `json:"beforeSigns,omitempty"`
	FirstCheck         *TreatmentFirstCheckSnapshot  `json:"firstCheck,omitempty"`
	SecondCheck        *TreatmentSecondCheckSnapshot `json:"secondCheck,omitempty"`
	BeforeSymptomItems []TreatmentSymptomItem        `json:"beforeSymptomItems,omitempty"`
	AfterSymptomItems  []TreatmentSymptomItem        `json:"afterSymptomItems,omitempty"`
	Actions            []ActionDTO                   `json:"actions,omitempty"`
	DuringParams       []TreatmentDuringParamDTO     `json:"duringParams,omitempty"`
}

type TreatmentBeforeSnapshot struct {
	Weight        *float64   `json:"weight,omitempty"`
	ExtraWeight   *float64   `json:"extraWeight,omitempty"`
	SBP           *float64   `json:"sbp,omitempty"`
	DBP           *float64   `json:"dbp,omitempty"`
	HeartRate     *float64   `json:"heartRate,omitempty"`
	Respiration   *float64   `json:"respiration,omitempty"`
	Temperature   *float64   `json:"temperature,omitempty"`
	PressurePoint string     `json:"pressurePoint,omitempty"`
	Symptoms      string     `json:"symptoms,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	OperateTime   *time.Time `json:"operateTime,omitempty"`
}

type TreatmentFirstCheckSnapshot struct {
	ID                    int64      `json:"id"`
	TreatmentID           int64      `json:"treatmentId"`
	BeforeSignsID         *int64     `json:"beforeSignsId,omitempty"`
	BeforeSymptomID       *int64     `json:"beforeSymptomId,omitempty"`
	OperatorID            *int64     `json:"operatorId,omitempty"`
	OperateTime           *time.Time `json:"operateTime,omitempty"`
	MaterialsResult       *bool      `json:"materialsResult,omitempty"`
	MaterialsMistake      string     `json:"materialsMistake,omitempty"`
	ParamResult           *bool      `json:"paramResult,omitempty"`
	ParamMistake          string     `json:"paramMistake,omitempty"`
	VascularAccessResult  *bool      `json:"vascularAccessResult,omitempty"`
	VascularAccessMistake string     `json:"vascularAccessMistake,omitempty"`
	PipelineResult        *bool      `json:"pipelineResult,omitempty"`
	PipelineMistake       string     `json:"pipelineMistake,omitempty"`
	CreatorID             int64      `json:"creatorId"`
	CreateTime            *time.Time `json:"createTime,omitempty"`
	LastModifyTime        *time.Time `json:"lastModifyTime,omitempty"`
}

type TreatmentSecondCheckSnapshot struct {
	ActionID              int64      `json:"actionId"`
	TreatmentID           int64      `json:"treatmentId"`
	OperatorID            *int64     `json:"operatorId,omitempty"`
	RecheckNurseID        *int64     `json:"recheckNurseId,omitempty"`
	QCNurseID             *int64     `json:"qcNurseId,omitempty"`
	OperateTime           *time.Time `json:"operateTime,omitempty"`
	ParamResult           *bool      `json:"paramResult,omitempty"`
	ParamMistake          string     `json:"paramMistake,omitempty"`
	VascularAccessResult  *bool      `json:"vascularAccessResult,omitempty"`
	VascularAccessMistake string     `json:"vascularAccessMistake,omitempty"`
	PipelineResult        *bool      `json:"pipelineResult,omitempty"`
	PipelineMistake       string     `json:"pipelineMistake,omitempty"`
	DialysisModeResult    *bool      `json:"dialysisModeResult,omitempty"`
	DialysisModeMistake   string     `json:"dialysisModeMistake,omitempty"`
	PrescriptionResult    *bool      `json:"prescriptionResult,omitempty"`
	PrescriptionMistake   string     `json:"prescriptionMistake,omitempty"`
	AnticoagulantResult   *bool      `json:"anticoagulantResult,omitempty"`
	AnticoagulantMistake  string     `json:"anticoagulantMistake,omitempty"`
	LineConnectionResult  *bool      `json:"lineConnectionResult,omitempty"`
	LineConnectionMistake string     `json:"lineConnectionMistake,omitempty"`
	CreateTime            *time.Time `json:"createTime,omitempty"`
	LastModifyTime        *time.Time `json:"lastModifyTime,omitempty"`
}

type ActionDTO struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	OperatorID  int64     `json:"operatorId"`
	OperateTime time.Time `json:"operateTime"`
	Code        string    `json:"code,omitempty"`
	Operator    string    `json:"operator,omitempty"`
}

type TreatmentDuringParamDTO struct {
	ID               int64      `json:"id"`
	TenantID         int64      `json:"tenantId"`
	TreatmentID      int64      `json:"treatmentId"`
	RecordTime       time.Time  `json:"recordTime"`
	Code             string     `json:"code"`
	BloodFlow        *float64   `json:"bloodFlow,omitempty"`
	DialysateFlow    *float64   `json:"dialysateFlow,omitempty"`
	UFVolume         *float64   `json:"ufVolume,omitempty"`
	VenousPressure   *float64   `json:"venousPressure,omitempty"`
	ArterialPressure *float64   `json:"arterialPressure,omitempty"`
	TMP              *float64   `json:"tmp,omitempty"`
	Temperature      *float64   `json:"temperature,omitempty"`
	Conductivity     *float64   `json:"conductivity,omitempty"`
	UFRate           *float64   `json:"ufRate,omitempty"`
	SBP              *float64   `json:"sbp,omitempty"`
	DBP              *float64   `json:"dbp,omitempty"`
	HeartRate        *float64   `json:"heartRate,omitempty"`
	Respiration      *float64   `json:"respiration,omitempty"`
	SpO2             *float64   `json:"spO2,omitempty"`
	Notes            string     `json:"notes,omitempty"`
	CreatorID        int64      `json:"creatorId"`
	CreateTime       *time.Time `json:"createTime,omitempty"`
	LastModifyTime   *time.Time `json:"lastModifyTime,omitempty"`
}

type legacyTreatmentHistoryRow struct {
	ID                   int64      `gorm:"column:Id"`
	TenantID             int64      `gorm:"column:TenantId"`
	PatientID            int64      `gorm:"column:PatientId"`
	ReceptionDrID        *int64     `gorm:"column:ReceptionDrId"`
	ShiftID              *int64     `gorm:"column:ShiftId"`
	ShiftName            string     `gorm:"column:ShiftName"`
	QueueNo              string     `gorm:"column:QueueNo"`
	Status               string     `gorm:"column:Status"`
	CaseStatus           string     `gorm:"column:CaseStatus"`
	StartTime            *time.Time `gorm:"column:StartTime"`
	EndTime              *time.Time `gorm:"column:EndTime"`
	SignInTime           *time.Time `gorm:"column:SignInTime"`
	ReceptionTime        *time.Time `gorm:"column:ReceptionTime"`
	RealDuration         *float64   `gorm:"column:RealDuration"`
	RealUFQuantity       *float64   `gorm:"column:RealUFQuantity"`
	RealSubstituteVolume *float64   `gorm:"column:RealSubstituateVolume"`
	NurseSummary         string     `gorm:"column:NurseSummary"`
	TreatmentSummary     string     `gorm:"column:TreatmentSummary"`
	CreatorID            int64      `gorm:"column:CreatorId"`
	CreateTime           time.Time  `gorm:"column:CreateTime"`
	LastModifyTime       time.Time  `gorm:"column:LastModifyTime"`
	TmrPath              string     `gorm:"column:TmrPath"`
	TmrTime              *time.Time `gorm:"column:TmrTime"`
	TmrPages             int        `gorm:"column:TmrPages"`
	WardID               *int64     `gorm:"column:WardId"`
}

type legacyTreatmentSigns struct {
	SBP          *float64 `gorm:"column:SBP"`
	DBP          *float64 `gorm:"column:DBP"`
	Complication string   `gorm:"column:Complication"`
	Symptoms     string   `gorm:"column:Symptoms"`
	Notes        string   `gorm:"column:Note"`
}

type legacySymptomRow struct {
	TreatmentID int64  `gorm:"column:TreatmentId"`
	Code        string `gorm:"column:Code"`
	Value       string `gorm:"column:Value"`
}

type legacyDuringParamRow struct {
	ID               int64      `gorm:"column:Id"`
	TreatmentID      int64      `gorm:"column:TreatmentId"`
	OperateTime      time.Time  `gorm:"column:OperateTime"`
	VenousPressure   *float64   `gorm:"column:VenousPressure"`
	ArterialPressure *float64   `gorm:"column:ArterialPressure"`
	TMP              *float64   `gorm:"column:TMP"`
	Conductivity     *float64   `gorm:"column:Conductivity"`
	UFQuantity       *float64   `gorm:"column:UFQuantity"`
	MachineTmp       *float64   `gorm:"column:MachineTmp"`
	BF               *float64   `gorm:"column:BF"`
	Note             string     `gorm:"column:Note"`
	CreatorID        int64      `gorm:"column:CreatorId"`
	CreateTime       *time.Time `gorm:"column:CreateTime"`
	LastModifyTime   *time.Time `gorm:"column:LastModifyTime"`
}

type legacyDuringSignsRow struct {
	ID             int64      `gorm:"column:Id"`
	TreatmentID    int64      `gorm:"column:TreatmentId"`
	OperateTime    time.Time  `gorm:"column:OperateTime"`
	SBP            *float64   `gorm:"column:SBP"`
	DBP            *float64   `gorm:"column:DBP"`
	HeartRate      *float64   `gorm:"column:HeartRate"`
	BodyTemp       *float64   `gorm:"column:BodyTemp"`
	Respiration    *float64   `gorm:"column:Respiration"`
	SpO2           *float64   `gorm:"column:SpO2"`
	OperatorID     int64      `gorm:"column:OperatorId"`
	CreatorID      int64      `gorm:"column:CreatorId"`
	CreateTime     *time.Time `gorm:"column:CreateTime"`
	LastModifyTime *time.Time `gorm:"column:LastModifyTime"`
}

type legacyTreatmentJSONDataRow struct {
	ID             int64           `gorm:"column:Id"`
	PatientID      int64           `gorm:"column:PatientId"`
	TreatmentID    int64           `gorm:"column:TreatmentId"`
	Code           string          `gorm:"column:Code"`
	CreatorID      int64           `gorm:"column:CreatorId"`
	CreateTime     time.Time       `gorm:"column:CreateTime"`
	LastModifyTime time.Time       `gorm:"column:LastModifyTime"`
	Value          json.RawMessage `gorm:"column:Value"`
}

type legacyBeforeCheckRow struct {
	ID                    int64      `gorm:"column:Id"`
	TreatmentID           int64      `gorm:"column:TreatmentId"`
	BeforeSignsID         *int64     `gorm:"column:BeforeSignsId"`
	BeforeSymptomID       *int64     `gorm:"column:BeforeSymptomId"`
	OperatorID            *int64     `gorm:"column:OperatorId"`
	OperateTime           time.Time  `gorm:"column:OperateTime"`
	MaterialsResult       *bool      `gorm:"column:MaterialsResult"`
	MaterialsMistake      string     `gorm:"column:MaterialsMistake"`
	ParamResult           *bool      `gorm:"column:ParamResult"`
	ParamMistake          string     `gorm:"column:ParamMistake"`
	VascularAccessResult  *bool      `gorm:"column:VascularAccessResult"`
	VascularAccessMistake string     `gorm:"column:VascularAccessMistake"`
	PipelineResult        *bool      `gorm:"column:PipelineResult"`
	PipelineMistake       string     `gorm:"column:PipelineMistake"`
	CreatorID             int64      `gorm:"column:CreatorId"`
	CreateTime            *time.Time `gorm:"column:CreateTime"`
	LastModifyTime        *time.Time `gorm:"column:LastModifyTime"`
}

const (
	legacyJSONCodeBeforeSymptom        = "hp_before_symptom"
	legacyJSONCodeAfterSymptom         = "hp_after_symptom"
	legacyJSONCodeDuringOther          = "hp_during_other"
	legacyJSONCodeTreatmentDetail      = "hp_treatment_details"
	legacyJSONCodeTreatmentFeelContent = "hp_treatment_feelcontent"
	legacyJSONCodeAgainCheck           = "hp_again_check"
	legacyActionCodeFirstCheck         = "70"
	legacyActionCodeAgainCheck         = "150"
)

func (s *TreatmentService) List(req TreatmentListRequest) (*TreatmentListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 200 {
		req.PageSize = 200
	}

	query := s.db.Table(`"Treatment_Treatment"`).Where(`"TenantId" = ?`, LegacyTenantID)
	if req.PatientId != nil {
		query = query.Where(`"PatientId" = ?`, *req.PatientId)
	}
	if req.Status != nil {
		status := legacyStatusFromApp(*req.Status)
		// 兼容旧数据：上机状态新写"30"，但旧记录可能为"10"，筛选"进行中"时同时匹配两者。
		if *req.Status == models.TreatmentStatusInProgress {
			query = query.Where(`"Status" IN ?`, []string{"10", status})
		} else {
			query = query.Where(`"Status" = ?`, status)
		}
	}
	if req.TreatmentDate != nil {
		query = query.Where(`DATE(COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime")) = DATE(?)`, *req.TreatmentDate)
	}
	if req.TreatmentDateStart != nil {
		query = query.Where(`DATE(COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime")) >= DATE(?)`, *req.TreatmentDateStart)
	}
	if req.TreatmentDateEnd != nil {
		query = query.Where(`DATE(COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime")) <= DATE(?)`, *req.TreatmentDateEnd)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []legacyTreatmentHistoryRow
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).
		Limit(req.PageSize).
		Order(`COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime") DESC`).
		Order(`"Id" DESC`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items, err := s.enrichTreatmentRows(rows, false)
	if err != nil {
		return nil, err
	}

	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPage++
	}

	return &TreatmentListResponse{
		Items:     items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPage: totalPage,
	}, nil
}

func (s *TreatmentService) Get(id int64) (*TreatmentRealtimeResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var row legacyTreatmentHistoryRow
	if err := s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("treatment not found")
		}
		return nil, err
	}
	items, err := s.enrichTreatmentRows([]legacyTreatmentHistoryRow{row}, true)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("treatment not found")
	}
	return &items[0], nil
}

func (s *TreatmentService) enrichTreatmentRows(rows []legacyTreatmentHistoryRow, includeDuringParams bool) ([]TreatmentRealtimeResponse, error) {
	if len(rows) == 0 {
		return []TreatmentRealtimeResponse{}, nil
	}

	actionMap, err := s.loadActionMap(rows)
	if err != nil {
		return nil, err
	}
	beforeMap, err := s.loadSignsMap(rows, `"Treatment_BeforeSigns"`)
	if err != nil {
		return nil, err
	}
	beforeDetailMap, err := s.loadBeforeSignsDetailMap(rows)
	if err != nil {
		return nil, err
	}
	firstCheckMap, err := s.loadBeforeCheckMap(rows)
	if err != nil {
		return nil, err
	}
	secondCheckMap, err := s.loadSecondCheckMap(rows, actionMap)
	if err != nil {
		return nil, err
	}
	afterMap, err := s.loadSignsMap(rows, `"Treatment_AfterSigns"`)
	if err != nil {
		return nil, err
	}
	beforeSymptomMap, err := s.loadSymptomItemMap(rows, `"Treatment_BeforeSymptom"`)
	if err != nil {
		return nil, err
	}
	afterSymptomMap, err := s.loadSymptomItemMap(rows, `"Treatment_AfterSymptom"`)
	if err != nil {
		return nil, err
	}
	afterJSONMap, err := s.loadTreatmentJSONSignsMap(rows, legacyJSONCodeAfterSymptom)
	if err != nil {
		return nil, err
	}
	detailJSONMap, err := s.loadTreatmentJSONSignsMap(rows, legacyJSONCodeTreatmentDetail)
	if err != nil {
		return nil, err
	}
	feelJSONMap, err := s.loadTreatmentJSONSignsMap(rows, legacyJSONCodeTreatmentFeelContent)
	if err != nil {
		return nil, err
	}
	statusDict, err := s.loadLegacyTreatmentStatusDict()
	if err != nil {
		return nil, err
	}
	beforeJSONItemsMap, afterJSONItemsMap, err := s.loadTreatmentJSONSymptomItemsMaps(rows)
	if err != nil {
		return nil, err
	}
	modeMap, err := s.loadTreatmentModeMap(rows)
	if err != nil {
		return nil, err
	}
	duringMap := map[int64][]TreatmentDuringParamDTO{}
	if includeDuringParams {
		duringMap, err = s.loadDuringParamMap(rows)
		if err != nil {
			return nil, err
		}
	}

	prescriptionService := &PrescriptionService{db: s.db}
	result := make([]TreatmentRealtimeResponse, 0, len(rows))
	for _, row := range rows {
		doctorName := ""
		if row.ReceptionDrID != nil && *row.ReceptionDrID > 0 {
			if name, lookupErr := prescriptionService.lookupLegacyUserDisplayName(*row.ReceptionDrID); lookupErr == nil {
				doctorName = name
			}
		}
		if doctorName == "" && row.CreatorID > 0 {
			if name, lookupErr := prescriptionService.lookupLegacyUserDisplayName(row.CreatorID); lookupErr == nil {
				doctorName = name
			}
		}

		actions := actionMap[row.ID]
		for i := range actions {
			if actions[i].OperatorID > 0 {
				if name, lookupErr := prescriptionService.lookupLegacyUserDisplayName(actions[i].OperatorID); lookupErr == nil {
					actions[i].Operator = name
				}
			}
		}

		afterSigns := afterMap[row.ID]
		if fallback, ok := afterJSONMap[row.ID]; ok {
			if afterSigns.Complication == "" {
				afterSigns.Complication = fallback.Complication
			}
			if afterSigns.Symptoms == "" {
				afterSigns.Symptoms = fallback.Symptoms
			}
			if afterSigns.Notes == "" {
				afterSigns.Notes = fallback.Notes
			}
		}
		if fallback, ok := detailJSONMap[row.ID]; ok {
			if afterSigns.Complication == "" {
				afterSigns.Complication = fallback.Complication
			}
			if afterSigns.Symptoms == "" {
				afterSigns.Symptoms = fallback.Symptoms
			}
			if afterSigns.Notes == "" {
				afterSigns.Notes = fallback.Notes
			}
		}
		if fallback, ok := feelJSONMap[row.ID]; ok {
			if afterSigns.Notes == "" {
				afterSigns.Notes = fallback.Notes
			}
		}
		beforeSymptomItems := beforeSymptomMap[row.ID]
		if len(beforeSymptomItems) == 0 {
			beforeSymptomItems = beforeJSONItemsMap[row.ID]
		}
		afterSymptomItems := afterSymptomMap[row.ID]
		if len(afterSymptomItems) == 0 {
			afterSymptomItems = afterJSONItemsMap[row.ID]
		}

		result = append(result, buildTreatmentRealtimeResponse(
			row,
			modeMap[row.ID],
			beforeMap[row.ID],
			afterSigns,
			beforeDetailMap[row.ID],
			firstCheckMap[row.ID],
			secondCheckMap[row.ID],
			doctorName,
			statusDict,
			beforeSymptomItems,
			afterSymptomItems,
			actions,
			duringMap[row.ID],
		))
	}
	return result, nil
}

func (s *TreatmentService) loadTreatmentJSONSignsMap(rows []legacyTreatmentHistoryRow, code string) (map[int64]legacyTreatmentSigns, error) {
	result := make(map[int64]legacyTreatmentSigns, len(rows))
	ids := collectTreatmentIDs(rows)
	if len(ids) == 0 || strings.TrimSpace(code) == "" {
		return result, nil
	}
	var items []legacyTreatmentJSONDataRow
	if err := s.db.Table(`"Auxiliary_JsonData"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ? AND "Code" = ?`, LegacyTenantID, ids, code).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		if _, exists := result[item.TreatmentID]; exists {
			continue
		}
		result[item.TreatmentID] = parseTreatmentSignsFromJSON(item.Value)
	}
	return result, nil
}

func (s *TreatmentService) loadLegacyTreatmentStatusDict() (map[int]string, error) {
	result := make(map[int]string)
	var rows []struct {
		Code string `gorm:"column:Code"`
		Name string `gorm:"column:Name"`
	}
	if err := s.db.Table(`"CodeDictionary_CodeDictionarys"`).
		Select(`"Code", "Name"`).
		Where(`"Type" = ? AND COALESCE("IsDisabled", false) = false`, "Treatment_TreatmentStatus").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		codeInt, err := strconv.Atoi(strings.TrimSpace(row.Code))
		if err != nil {
			continue
		}
		result[codeInt] = strings.TrimSpace(row.Name)
	}
	return result, nil
}

func parseTreatmentSignsFromJSON(raw json.RawMessage) legacyTreatmentSigns {
	var payload struct {
		Complication string                 `json:"complication"`
		Symptoms     string                 `json:"symptoms"`
		Notes        string                 `json:"notes"`
		SymptomItems []TreatmentSymptomItem `json:"symptomItems"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return legacyTreatmentSigns{}
	}
	out := legacyTreatmentSigns{
		Complication: strings.TrimSpace(payload.Complication),
		Symptoms:     strings.TrimSpace(payload.Symptoms),
		Notes:        strings.TrimSpace(payload.Notes),
	}
	for _, item := range payload.SymptomItems {
		code := strings.TrimSpace(item.Code)
		value := strings.TrimSpace(item.Value)
		if value == "" {
			continue
		}
		if code == "complication" && out.Complication == "" {
			out.Complication = value
		}
		if code == "symptoms" && out.Symptoms == "" {
			out.Symptoms = value
		}
		if code == "notes" && out.Notes == "" {
			out.Notes = value
		}
	}
	return out
}

func parseTreatmentSymptomItemsFromJSON(raw json.RawMessage) []TreatmentSymptomItem {
	var payload struct {
		SymptomItems []TreatmentSymptomItem `json:"symptomItems"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return []TreatmentSymptomItem{}
	}
	result := make([]TreatmentSymptomItem, 0, len(payload.SymptomItems))
	for _, item := range payload.SymptomItems {
		code := strings.TrimSpace(item.Code)
		value := strings.TrimSpace(item.Value)
		if code == "" || value == "" {
			continue
		}
		result = append(result, TreatmentSymptomItem{Code: code, Value: value})
	}
	return result
}

func (s *TreatmentService) loadTreatmentJSONSymptomItemsMaps(rows []legacyTreatmentHistoryRow) (map[int64][]TreatmentSymptomItem, map[int64][]TreatmentSymptomItem, error) {
	before := make(map[int64][]TreatmentSymptomItem, len(rows))
	after := make(map[int64][]TreatmentSymptomItem, len(rows))
	ids := collectTreatmentIDs(rows)
	if len(ids) == 0 {
		return before, after, nil
	}
	var items []legacyTreatmentJSONDataRow
	if err := s.db.Table(`"Auxiliary_JsonData"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ? AND "Code" IN ?`, LegacyTenantID, ids, []string{legacyJSONCodeBeforeSymptom, legacyJSONCodeAfterSymptom}).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		Find(&items).Error; err != nil {
		return nil, nil, err
	}
	seenBefore := make(map[int64]struct{}, len(rows))
	seenAfter := make(map[int64]struct{}, len(rows))
	for _, item := range items {
		switch strings.TrimSpace(item.Code) {
		case legacyJSONCodeBeforeSymptom:
			if _, ok := seenBefore[item.TreatmentID]; ok {
				continue
			}
			before[item.TreatmentID] = parseTreatmentSymptomItemsFromJSON(item.Value)
			seenBefore[item.TreatmentID] = struct{}{}
		case legacyJSONCodeAfterSymptom:
			if _, ok := seenAfter[item.TreatmentID]; ok {
				continue
			}
			after[item.TreatmentID] = parseTreatmentSymptomItemsFromJSON(item.Value)
			seenAfter[item.TreatmentID] = struct{}{}
		}
	}
	return before, after, nil
}

func (s *TreatmentService) loadTreatmentModeMap(rows []legacyTreatmentHistoryRow) (map[int64]string, error) {
	result := make(map[int64]string, len(rows))
	ids := collectTreatmentIDs(rows)
	if len(ids) == 0 {
		return result, nil
	}

	var items []struct {
		TreatmentID    int64  `gorm:"column:TreatmentId"`
		DialysisMethod string `gorm:"column:DialysisMethod"`
	}
	if err := s.db.Table(`"Plan_PatientPrescription"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, LegacyTenantID, ids).
		Order(`"LastModifyTime" DESC`).
		Order(`"Id" DESC`).
		Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		if _, ok := result[item.TreatmentID]; ok {
			continue
		}
		result[item.TreatmentID] = normalizeLegacyDialysisMode(item.DialysisMethod)
	}
	return result, nil
}

func (s *TreatmentService) loadActionMap(rows []legacyTreatmentHistoryRow) (map[int64][]ActionDTO, error) {
	result := make(map[int64][]ActionDTO, len(rows))
	ids := collectTreatmentIDs(rows)

	var actions []struct {
		ID          int64     `gorm:"column:Id"`
		TreatmentID int64     `gorm:"column:TreatmentId"`
		Name        string    `gorm:"column:Name"`
		OperatorID  int64     `gorm:"column:OperatorId"`
		OperateTime time.Time `gorm:"column:OperateTime"`
		Code        string    `gorm:"column:Code"`
	}
	if err := s.db.Table(`"Treatment_Action"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, LegacyTenantID, ids).
		Order(`"OperateTime" ASC`).
		Order(`"Id" ASC`).
		Find(&actions).Error; err != nil {
		return nil, err
	}
	for _, action := range actions {
		result[action.TreatmentID] = append(result[action.TreatmentID], ActionDTO{
			ID:          action.ID,
			Name:        action.Name,
			OperatorID:  action.OperatorID,
			OperateTime: action.OperateTime,
			Code:        action.Code,
		})
	}
	return result, nil
}

func (s *TreatmentService) loadSignsMap(rows []legacyTreatmentHistoryRow, table string) (map[int64]legacyTreatmentSigns, error) {
	result := make(map[int64]legacyTreatmentSigns, len(rows))
	ids := collectTreatmentIDs(rows)

	var signs []struct {
		TreatmentID int64 `gorm:"column:TreatmentId"`
		legacyTreatmentSigns
	}
	if err := s.db.Table(table).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, LegacyTenantID, ids).
		Find(&signs).Error; err != nil {
		return nil, err
	}
	for _, sign := range signs {
		result[sign.TreatmentID] = sign.legacyTreatmentSigns
	}
	return result, nil
}

func (s *TreatmentService) loadBeforeSignsDetailMap(rows []legacyTreatmentHistoryRow) (map[int64]legacyBeforeSignsRow, error) {
	result := make(map[int64]legacyBeforeSignsRow, len(rows))
	ids := collectTreatmentIDs(rows)
	if len(ids) == 0 {
		return result, nil
	}

	var signs []legacyBeforeSignsRow
	if err := s.db.Table(`"Treatment_BeforeSigns"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, LegacyTenantID, ids).
		Order(`"OperateTime" DESC`).
		Order(`"Id" DESC`).
		Find(&signs).Error; err != nil {
		return nil, err
	}
	for _, sign := range signs {
		if _, exists := result[sign.TreatmentID]; exists {
			continue
		}
		result[sign.TreatmentID] = sign
	}
	return result, nil
}

func (s *TreatmentService) loadBeforeCheckMap(rows []legacyTreatmentHistoryRow) (map[int64]*TreatmentFirstCheckSnapshot, error) {
	result := make(map[int64]*TreatmentFirstCheckSnapshot, len(rows))
	ids := collectTreatmentIDs(rows)
	if len(ids) == 0 {
		return result, nil
	}

	var items []legacyBeforeCheckRow
	if err := s.db.Table(`"Treatment_BeforeCheck"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, LegacyTenantID, ids).
		Order(`"OperateTime" DESC`).
		Order(`"Id" DESC`).
		Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		if _, exists := result[item.TreatmentID]; exists {
			continue
		}
		result[item.TreatmentID] = mapFirstCheckSnapshot(item)
	}
	return result, nil
}

func mapFirstCheckSnapshot(row legacyBeforeCheckRow) *TreatmentFirstCheckSnapshot {
	var operateTime *time.Time
	if !row.OperateTime.IsZero() {
		t := row.OperateTime
		operateTime = &t
	}
	return &TreatmentFirstCheckSnapshot{
		ID:                    row.ID,
		TreatmentID:           row.TreatmentID,
		BeforeSignsID:         row.BeforeSignsID,
		BeforeSymptomID:       row.BeforeSymptomID,
		OperatorID:            row.OperatorID,
		OperateTime:           operateTime,
		MaterialsResult:       row.MaterialsResult,
		MaterialsMistake:      strings.TrimSpace(row.MaterialsMistake),
		ParamResult:           row.ParamResult,
		ParamMistake:          strings.TrimSpace(row.ParamMistake),
		VascularAccessResult:  row.VascularAccessResult,
		VascularAccessMistake: strings.TrimSpace(row.VascularAccessMistake),
		PipelineResult:        row.PipelineResult,
		PipelineMistake:       strings.TrimSpace(row.PipelineMistake),
		CreatorID:             row.CreatorID,
		CreateTime:            row.CreateTime,
		LastModifyTime:        row.LastModifyTime,
	}
}

type legacySecondCheckJSONRow struct {
	TreatmentID    int64           `gorm:"column:TreatmentId"`
	CreatorID      int64           `gorm:"column:CreatorId"`
	CreateTime     *time.Time      `gorm:"column:CreateTime"`
	LastModifyTime *time.Time      `gorm:"column:LastModifyTime"`
	Value          json.RawMessage `gorm:"column:Value"`
}

type treatmentSecondCheckPayload struct {
	OperatorID            *int64     `json:"operatorId,omitempty"`
	RecheckNurseID        *int64     `json:"recheckNurseId,omitempty"`
	QCNurseID             *int64     `json:"qcNurseId,omitempty"`
	OperateTime           *time.Time `json:"operateTime,omitempty"`
	ParamResult           *bool      `json:"paramResult,omitempty"`
	ParamMistake          string     `json:"paramMistake,omitempty"`
	VascularAccessResult  *bool      `json:"vascularAccessResult,omitempty"`
	VascularAccessMistake string     `json:"vascularAccessMistake,omitempty"`
	PipelineResult        *bool      `json:"pipelineResult,omitempty"`
	PipelineMistake       string     `json:"pipelineMistake,omitempty"`
	DialysisModeResult    *bool      `json:"dialysisModeResult,omitempty"`
	DialysisModeMistake   string     `json:"dialysisModeMistake,omitempty"`
	PrescriptionResult    *bool      `json:"prescriptionResult,omitempty"`
	PrescriptionMistake   string     `json:"prescriptionMistake,omitempty"`
	AnticoagulantResult   *bool      `json:"anticoagulantResult,omitempty"`
	AnticoagulantMistake  string     `json:"anticoagulantMistake,omitempty"`
	LineConnectionResult  *bool      `json:"lineConnectionResult,omitempty"`
	LineConnectionMistake string     `json:"lineConnectionMistake,omitempty"`
}

func (s *TreatmentService) loadSecondCheckMap(rows []legacyTreatmentHistoryRow, actionMap map[int64][]ActionDTO) (map[int64]*TreatmentSecondCheckSnapshot, error) {
	result := make(map[int64]*TreatmentSecondCheckSnapshot, len(rows))
	ids := collectTreatmentIDs(rows)
	if len(ids) == 0 {
		return result, nil
	}

	var jsonRows []legacySecondCheckJSONRow
	if err := s.db.Table(`"Auxiliary_JsonData"`).
		Select(`"TreatmentId", "CreatorId", "CreateTime", "LastModifyTime", "Value"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ? AND "Code" = ?`, LegacyTenantID, ids, legacyJSONCodeAgainCheck).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		Find(&jsonRows).Error; err != nil {
		return nil, err
	}
	for _, row := range jsonRows {
		if _, exists := result[row.TreatmentID]; exists {
			continue
		}
		payload := treatmentSecondCheckPayload{}
		if len(row.Value) > 0 {
			if err := json.Unmarshal(row.Value, &payload); err != nil {
				continue
			}
		}
		result[row.TreatmentID] = &TreatmentSecondCheckSnapshot{
			ActionID:              0,
			TreatmentID:           row.TreatmentID,
			OperatorID:            payload.OperatorID,
			RecheckNurseID:        payload.RecheckNurseID,
			QCNurseID:             payload.QCNurseID,
			OperateTime:           payload.OperateTime,
			ParamResult:           payload.ParamResult,
			ParamMistake:          strings.TrimSpace(payload.ParamMistake),
			VascularAccessResult:  payload.VascularAccessResult,
			VascularAccessMistake: strings.TrimSpace(payload.VascularAccessMistake),
			PipelineResult:        payload.PipelineResult,
			PipelineMistake:       strings.TrimSpace(payload.PipelineMistake),
			DialysisModeResult:    payload.DialysisModeResult,
			DialysisModeMistake:   strings.TrimSpace(payload.DialysisModeMistake),
			PrescriptionResult:    payload.PrescriptionResult,
			PrescriptionMistake:   strings.TrimSpace(payload.PrescriptionMistake),
			AnticoagulantResult:   payload.AnticoagulantResult,
			AnticoagulantMistake:  strings.TrimSpace(payload.AnticoagulantMistake),
			LineConnectionResult:  payload.LineConnectionResult,
			LineConnectionMistake: strings.TrimSpace(payload.LineConnectionMistake),
			CreateTime:            row.CreateTime,
			LastModifyTime:        row.LastModifyTime,
		}
	}
	for _, row := range rows {
		snapshot := result[row.ID]
		action := pickLatestSecondCheckAction(actionMap[row.ID])
		if action == nil {
			continue
		}
		if snapshot == nil {
			snapshot = &TreatmentSecondCheckSnapshot{TreatmentID: row.ID}
			result[row.ID] = snapshot
		}
		snapshot.ActionID = action.ID
		if snapshot.OperatorID == nil && action.OperatorID > 0 {
			value := action.OperatorID
			snapshot.OperatorID = &value
		}
		if snapshot.OperateTime == nil {
			t := action.OperateTime
			snapshot.OperateTime = &t
		}
	}
	return result, nil
}

func pickLatestSecondCheckAction(actions []ActionDTO) *ActionDTO {
	for i := len(actions) - 1; i >= 0; i-- {
		code := strings.TrimSpace(actions[i].Code)
		name := strings.TrimSpace(actions[i].Name)
		if code == legacyActionCodeAgainCheck || strings.Contains(name, "二次核对") {
			return &actions[i]
		}
	}
	return nil
}

func (s *TreatmentService) loadSymptomItemMap(rows []legacyTreatmentHistoryRow, table string) (map[int64][]TreatmentSymptomItem, error) {
	result := make(map[int64][]TreatmentSymptomItem, len(rows))
	ids := collectTreatmentIDs(rows)
	if len(ids) == 0 {
		return result, nil
	}
	var items []legacySymptomRow
	if err := s.db.Table(table).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, LegacyTenantID, ids).
		Order(`"OperateTime" ASC`).
		Order(`"Id" ASC`).
		Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		value := strings.TrimSpace(item.Value)
		if code == "" || value == "" {
			continue
		}
		result[item.TreatmentID] = append(result[item.TreatmentID], TreatmentSymptomItem{
			Code:  code,
			Value: value,
		})
	}
	return result, nil
}

func (s *TreatmentService) loadDuringParamMap(rows []legacyTreatmentHistoryRow) (map[int64][]TreatmentDuringParamDTO, error) {
	result := make(map[int64][]TreatmentDuringParamDTO, len(rows))
	ids := collectTreatmentIDs(rows)
	if len(ids) == 0 {
		return result, nil
	}

	var params []legacyDuringParamRow
	if err := s.db.Table(`"Treatment_DuringParam"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, LegacyTenantID, ids).
		Order(`"OperateTime" ASC`).
		Order(`"Id" ASC`).
		Find(&params).Error; err != nil {
		return nil, err
	}

	duringIndex := make(map[int64]map[int64]*TreatmentDuringParamDTO, len(ids))
	for _, item := range params {
		dto := TreatmentDuringParamDTO{
			ID:               item.ID,
			TenantID:         LegacyTenantID,
			TreatmentID:      item.TreatmentID,
			RecordTime:       item.OperateTime,
			Code:             "legacy",
			BloodFlow:        item.BF,
			UFVolume:         item.UFQuantity,
			VenousPressure:   item.VenousPressure,
			ArterialPressure: item.ArterialPressure,
			TMP:              item.TMP,
			Temperature:      item.MachineTmp,
			Conductivity:     item.Conductivity,
			Notes:            strings.TrimSpace(item.Note),
			CreatorID:        item.CreatorID,
			CreateTime:       item.CreateTime,
			LastModifyTime:   item.LastModifyTime,
		}
		result[item.TreatmentID] = append(result[item.TreatmentID], dto)
		if _, ok := duringIndex[item.TreatmentID]; !ok {
			duringIndex[item.TreatmentID] = make(map[int64]*TreatmentDuringParamDTO)
		}
		idx := len(result[item.TreatmentID]) - 1
		duringIndex[item.TreatmentID][item.OperateTime.Unix()] = &result[item.TreatmentID][idx]
	}

	var signs []legacyDuringSignsRow
	if err := s.db.Table(`"Treatment_DuringSigns"`).
		Where(`"TenantId" = ? AND "TreatmentId" IN ?`, LegacyTenantID, ids).
		Order(`"OperateTime" ASC`).
		Order(`"Id" ASC`).
		Find(&signs).Error; err != nil {
		return nil, err
	}
	for _, sign := range signs {
		byTime, ok := duringIndex[sign.TreatmentID]
		if !ok {
			continue
		}
		target, exists := byTime[sign.OperateTime.Unix()]
		if !exists || target == nil {
			continue
		}
		target.SBP = sign.SBP
		target.DBP = sign.DBP
		target.HeartRate = sign.HeartRate
		target.Respiration = sign.Respiration
		target.SpO2 = sign.SpO2
		if target.Temperature == nil {
			target.Temperature = sign.BodyTemp
		}
	}
	return result, nil
}

func collectTreatmentIDs(rows []legacyTreatmentHistoryRow) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

func buildTreatmentRealtimeResponse(row legacyTreatmentHistoryRow, treatmentMode string, beforeSigns, afterSigns legacyTreatmentSigns, beforeDetail legacyBeforeSignsRow, firstCheck *TreatmentFirstCheckSnapshot, secondCheck *TreatmentSecondCheckSnapshot, doctorName string, statusDict map[int]string, beforeSymptomItems, afterSymptomItems []TreatmentSymptomItem, actions []ActionDTO, duringParams []TreatmentDuringParamDTO) TreatmentRealtimeResponse {
	treatmentDate := firstNonNilTime(row.StartTime, row.SignInTime, row.ReceptionTime, &row.CreateTime)
	doctorSummary := strings.TrimSpace(row.TreatmentSummary)
	treatmentSummary := strings.TrimSpace(row.NurseSummary)
	notes := doctorSummary
	if notes == "" {
		notes = treatmentSummary
	}

	duration := 0
	if row.RealDuration != nil {
		duration = int(*row.RealDuration)
	} else if row.StartTime != nil && row.EndTime != nil {
		duration = int(row.EndTime.Sub(*row.StartTime).Minutes())
	}

	weightLoss := 0.0
	if row.RealUFQuantity != nil {
		weightLoss = *row.RealUFQuantity / 1000.0
	}

	complications := strings.TrimSpace(afterSigns.Complication)
	if complications == "" {
		complications = strings.TrimSpace(afterSigns.Symptoms)
	}
	if complications == "" {
		complications = strings.TrimSpace(afterSigns.Notes)
	}

	startSBP := beforeSigns.SBP
	startDBP := beforeSigns.DBP
	if beforeDetail.SBP != nil {
		startSBP = beforeDetail.SBP
	}
	if beforeDetail.DBP != nil {
		startDBP = beforeDetail.DBP
	}

	return TreatmentRealtimeResponse{
		ID:                 row.ID,
		TenantID:           row.TenantID,
		PatientID:          strconv.FormatInt(row.PatientID, 10),
		WardID:             row.WardID,
		TreatmentDate:      treatmentDate.Format(time.RFC3339),
		ShiftID:            row.ShiftID,
		TreatmentType:      inferTreatmentType(row, treatmentMode),
		Status:             appStatusFromLegacy(row.Status, row.StartTime, row.EndTime, statusDict),
		LegacyStatus:       strings.TrimSpace(row.Status),
		StartTime:          row.StartTime,
		EndTime:            row.EndTime,
		Notes:              notes,
		CreatorID:          row.CreatorID,
		CreateTime:         row.CreateTime,
		LastModifyTime:     row.LastModifyTime,
		DoctorSummary:      doctorSummary,
		TreatmentSummary:   treatmentSummary,
		TimeRange:          formatTimeRange(row.StartTime, row.EndTime),
		DurationMinutes:    duration,
		WeightLossKG:       weightLoss,
		ShiftName:          row.ShiftName,
		QueueNo:            row.QueueNo,
		CaseStatus:         row.CaseStatus,
		TmrPath:            row.TmrPath,
		TmrTime:            row.TmrTime,
		TmrPages:           row.TmrPages,
		DoctorName:         doctorName,
		StartBP:            formatBP(startSBP, startDBP),
		EndBP:              formatBP(afterSigns.SBP, afterSigns.DBP),
		Complications:      complications,
		BeforeSigns:        buildBeforeSnapshot(beforeDetail, beforeSymptomItems),
		FirstCheck:         firstCheck,
		SecondCheck:        secondCheck,
		BeforeSymptomItems: beforeSymptomItems,
		AfterSymptomItems:  afterSymptomItems,
		Actions:            actions,
		DuringParams:       duringParams,
	}
}

func buildBeforeSnapshot(detail legacyBeforeSignsRow, symptomItems []TreatmentSymptomItem) *TreatmentBeforeSnapshot {
	pressurePoint := strings.TrimSpace(detail.PressurePoint)
	symptoms := ""
	for _, item := range symptomItems {
		code := strings.TrimSpace(item.Code)
		value := strings.TrimSpace(item.Value)
		if code == "bp_site" && pressurePoint == "" {
			pressurePoint = value
		}
		if code == "symptoms" && symptoms == "" {
			symptoms = value
		}
	}

	isEmpty := detail.Weight == nil &&
		detail.ExtraWeight == nil &&
		detail.SBP == nil &&
		detail.DBP == nil &&
		detail.HeartRate == nil &&
		detail.Respiration == nil &&
		detail.BodyTemp == nil &&
		strings.TrimSpace(detail.Note) == "" &&
		pressurePoint == "" &&
		symptoms == "" &&
		detail.OperateTime.IsZero()
	if isEmpty {
		return nil
	}

	var operateTime *time.Time
	if !detail.OperateTime.IsZero() {
		t := detail.OperateTime
		operateTime = &t
	}

	return &TreatmentBeforeSnapshot{
		Weight:        detail.Weight,
		ExtraWeight:   detail.ExtraWeight,
		SBP:           detail.SBP,
		DBP:           detail.DBP,
		HeartRate:     detail.HeartRate,
		Respiration:   detail.Respiration,
		Temperature:   detail.BodyTemp,
		PressurePoint: pressurePoint,
		Symptoms:      symptoms,
		Notes:         strings.TrimSpace(detail.Note),
		OperateTime:   operateTime,
	}
}

func firstNonNilTime(values ...*time.Time) time.Time {
	for _, value := range values {
		if value != nil && !value.IsZero() {
			return *value
		}
	}
	return time.Time{}
}

func formatTimeRange(startTime, endTime *time.Time) string {
	if startTime == nil && endTime == nil {
		return ""
	}
	if startTime != nil && endTime != nil {
		return startTime.Format("15:04") + "-" + endTime.Format("15:04")
	}
	if startTime != nil {
		return startTime.Format("15:04") + "-"
	}
	return "-" + endTime.Format("15:04")
}

func formatBP(sbp, dbp *float64) string {
	if sbp == nil && dbp == nil {
		return ""
	}
	left := "-"
	right := "-"
	if sbp != nil {
		left = strconv.Itoa(int(*sbp))
	}
	if dbp != nil {
		right = strconv.Itoa(int(*dbp))
	}
	return left + "/" + right
}

func inferTreatmentType(row legacyTreatmentHistoryRow, treatmentMode string) string {
	if strings.TrimSpace(treatmentMode) != "" {
		return normalizeLegacyDialysisMode(treatmentMode)
	}
	text := strings.ToUpper(strings.Join([]string{
		row.NurseSummary,
		row.TreatmentSummary,
		row.CaseStatus,
	}, " "))
	switch {
	case strings.Contains(text, "HDF") || strings.Contains(text, "血液透析滤过"):
		return "HDF"
	case strings.Contains(text, "HD+HP") || strings.Contains(text, "血液透析+灌流"):
		return "HD+HP"
	case strings.Contains(text, "HP") || strings.Contains(text, "血液灌流"):
		return "HP"
	case row.RealSubstituteVolume != nil && *row.RealSubstituteVolume > 0:
		return "HDF"
	default:
		return "HD"
	}
}

func appStatusFromLegacy(raw string, startTime, endTime *time.Time, statusDict map[int]string) int {
	// 明确状态码优先于字典名判断：老库字典中 "10"="签到"，但实际业务中
	// StartTime 非空、EndTime 空的记录应识别为"进行中"。将 hard-code 分支
	// 提前，避免字典名误判。
	switch strings.TrimSpace(raw) {
	case "60", "100":
		return models.TreatmentStatusCompleted
	case "90", "50":
		return models.TreatmentStatusCancelled
	case "0", "10", "30":
		if startTime != nil && endTime == nil {
			return models.TreatmentStatusInProgress
		}
		if endTime != nil {
			return models.TreatmentStatusCompleted
		}
		if strings.TrimSpace(raw) == "0" {
			return models.TreatmentStatusPending
		}
	}
	// 字典名 fallback：非明确治疗状态码或缺失 start/end 时按字典语义补漏。
	if code, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		if name := strings.TrimSpace(statusDict[code]); name != "" {
			switch {
			case strings.Contains(name, "取消"), strings.Contains(name, "作废"), strings.Contains(name, "终止"):
				return models.TreatmentStatusCancelled
			case strings.Contains(name, "完成"), strings.Contains(name, "结束"), strings.Contains(name, "下机"):
				return models.TreatmentStatusCompleted
			case strings.Contains(name, "进行"), strings.Contains(name, "治疗中"), strings.Contains(name, "开始"):
				return models.TreatmentStatusInProgress
			case strings.Contains(name, "待"), strings.Contains(name, "排队"), strings.Contains(name, "签到"), strings.Contains(name, "接诊"):
				return models.TreatmentStatusPending
			}
		}
	}
	if endTime != nil {
		return models.TreatmentStatusCompleted
	}
	if startTime != nil {
		return models.TreatmentStatusInProgress
	}
	return models.TreatmentStatusPending
}

func legacyStatusFromApp(status int) string {
	switch status {
	case models.TreatmentStatusInProgress:
		return "30" // 透中监测（治疗中）
	case models.TreatmentStatusCompleted:
		return "60" // 已结束
	case models.TreatmentStatusCancelled:
		return "50" // 取消治疗
	default:
		return "0" // 待签到
	}
}

type TreatmentCreateRequest struct {
	PatientId      modeltypes.LegacyID `json:"patientId" binding:"required"`
	TreatmentDate  time.Time           `json:"treatmentDate" binding:"required"`
	ScheduleId     *int64              `json:"scheduleId"`
	ReceptionDrId  *int64              `json:"receptionDrId"`
	SignInTime     *time.Time          `json:"signInTime"`
	QueueNo        string              `json:"queueNo"`
	ReceptionTime  *time.Time          `json:"receptionTime"`
	DayProgrammeId *int64              `json:"dayProgrammeId"`
	WardId         *int64              `json:"wardId"`
	WardName       string              `json:"wardName"`
	BedId          *int64              `json:"bedId"`
	ShiftId        *int64              `json:"shiftId"`
	ShiftTiming    int                 `json:"shiftTiming"`
	Type           int                 `json:"type" binding:"required"`
	Status         int                 `json:"status"`
	Notes          string              `json:"notes"`
}

func (s *TreatmentService) Create(req TreatmentCreateRequest, tenantId, creatorId int64) (*TreatmentRealtimeResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 方案完整性门禁（契约02 三）：若直接以上机态建治疗，同样校验方案已补全。
	if req.Status == models.TreatmentStatusInProgress {
		if err := s.ensurePlanCompleteForPatient(req.PatientId); err != nil {
			return nil, err
		}
	}

	id, err := nextLegacyID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	status := legacyStatusFromApp(req.Status)
	startTime := (*time.Time)(nil)
	if req.Status == models.TreatmentStatusInProgress {
		startTime = &now
	}
	signInTime := req.SignInTime
	if signInTime == nil {
		signInTime = &now
	}

	values := map[string]any{
		"Id":               id,
		"TenantId":         LegacyTenantID,
		"PatientId":        req.PatientId,
		"ScheduleId":       req.ScheduleId,
		"ReceptionDrId":    req.ReceptionDrId,
		"SignInTime":       signInTime,
		"QueueNo":          req.QueueNo,
		"ReceptionTime":    req.ReceptionTime,
		"DayProgrammeId":   req.DayProgrammeId,
		"WardId":           req.WardId,
		"WardName":         req.WardName,
		"BedId":            req.BedId,
		"ShiftId":          req.ShiftId,
		"ShiftName":        "",
		"StartTime":        startTime,
		"NurseSummary":     req.Notes,
		"TreatmentSummary": req.Notes,
		"Status":           status,
		"CaseStatus":       "",
		"CreatorId":        creatorId,
		"CreateTime":       now,
		"LastModifyTime":   now,
	}
	if err := s.db.Table(`"Treatment_Treatment"`).Create(values).Error; err != nil {
		return nil, err
	}
	return s.Get(id.Int64())
}

type TreatmentUpdateRequest struct {
	SignInTime    *time.Time `json:"signInTime"`
	QueueNo       *string    `json:"queueNo"`
	ReceptionTime *time.Time `json:"receptionTime"`
	ReceptionDrId *int64     `json:"receptionDrId"`
	WardId        *int64     `json:"wardId"`
	WardName      *string    `json:"wardName"`
	BedId         *int64     `json:"bedId"`
	ShiftId       *int64     `json:"shiftId"`
	ShiftTiming   *int       `json:"shiftTiming"`
	Status        *int       `json:"status"`
	StartTime     *time.Time `json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
	IsDisabled    *bool      `json:"isDisabled"`
	Notes         *string    `json:"notes"`
}

func (s *TreatmentService) Update(id int64, req TreatmentUpdateRequest) (*TreatmentRealtimeResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	// 方案完整性门禁（契约02 三）：经 Update 推进到上机（20→30）同样需校验方案已补全，
	// 与 UpdateStatus 两条上机写入路径保持一致。
	if req.Status != nil && *req.Status == models.TreatmentStatusInProgress {
		if err := s.ensurePlanCompleteForTreatment(id); err != nil {
			return nil, err
		}
	}

	updates := map[string]any{
		"LastModifyTime": time.Now(),
	}
	if req.SignInTime != nil {
		updates["SignInTime"] = *req.SignInTime
	}
	if req.QueueNo != nil {
		updates["QueueNo"] = *req.QueueNo
	}
	if req.ReceptionTime != nil {
		updates["ReceptionTime"] = *req.ReceptionTime
	}
	if req.ReceptionDrId != nil {
		updates["ReceptionDrId"] = *req.ReceptionDrId
	}
	if req.WardId != nil {
		updates["WardId"] = *req.WardId
	}
	if req.WardName != nil {
		updates["WardName"] = *req.WardName
	}
	if req.BedId != nil {
		updates["BedId"] = *req.BedId
	}
	if req.ShiftId != nil {
		updates["ShiftId"] = *req.ShiftId
	}
	if req.Status != nil {
		updates["Status"] = legacyStatusFromApp(*req.Status)
		if *req.Status == models.TreatmentStatusInProgress {
			updates["StartTime"] = gorm.Expr(`COALESCE("StartTime", ?)`, time.Now())
		}
		if *req.Status == models.TreatmentStatusCompleted {
			updates["EndTime"] = gorm.Expr(`COALESCE("EndTime", ?)`, time.Now())
		}
	}
	if req.StartTime != nil {
		updates["StartTime"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["EndTime"] = *req.EndTime
	}
	if req.Notes != nil {
		updates["NurseSummary"] = *req.Notes
		updates["TreatmentSummary"] = *req.Notes
	}

	result := s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("treatment not found")
	}
	return s.Get(id)
}

func (s *TreatmentService) Delete(id int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	// Treatment_Treatment 没有 IsDisabled 列，使用 CaseStatus='20' 表示封存。
	updates := map[string]any{
		"CaseStatus":     "20", // 封存
		"EndTime":        gorm.Expr(`COALESCE("EndTime", ?)`, time.Now()),
		"LastModifyTime": time.Now(),
	}
	result := s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ? AND COALESCE("CaseStatus", '10') <> '20'`, id, LegacyTenantID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("treatment not found")
	}
	return nil
}

func (s *TreatmentService) UpdateStatus(id int64, status int) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	// 方案完整性门禁（契约02 三）：上机（20→30）前校验治疗方案已补全。
	// 草稿态方案（透析液配方为空，含建档时生成的草稿）不得上机，提示先到方案页完善。
	if status == models.TreatmentStatusInProgress {
		if err := s.ensurePlanCompleteForTreatment(id); err != nil {
			return err
		}
	}
	updates := map[string]any{
		"Status":         legacyStatusFromApp(status),
		"LastModifyTime": time.Now(),
	}
	if status == models.TreatmentStatusInProgress {
		updates["StartTime"] = gorm.Expr(`COALESCE("StartTime", ?)`, time.Now())
	}
	// 修复：原 UpdateStatus 只在「上机」补 StartTime，「下机/完成」却不写 EndTime，
	//       导致治疗时长无法计算、且记录的 EndTime 长期为空。
	if status == models.TreatmentStatusCompleted {
		updates["EndTime"] = gorm.Expr(`COALESCE("EndTime", ?)`, time.Now())
	}
	result := s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, id, LegacyTenantID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("treatment not found")
	}
	if status == models.TreatmentStatusInProgress {
		s.syncScheduleStatus(id, 50)
	}
	// 下机(完成)时把计划排班一并流转到"已完成(60)"，保持计划视图与执行一致。
	if status == models.TreatmentStatusCompleted {
		s.syncScheduleStatus(id, 60)
	}
	return nil
}

func (s *TreatmentService) GetByPatientAndDate(patientId modeltypes.LegacyID, date time.Time) (*TreatmentRealtimeResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var row legacyTreatmentHistoryRow
	err := s.db.Table(`"Treatment_Treatment"`).
		Where(`"PatientId" = ? AND "TenantId" = ? AND DATE(COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime")) = DATE(?)`, patientId, LegacyTenantID, date).
		Order(`COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime") DESC`).
		Order(`"Id" DESC`).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	items, err := s.enrichTreatmentRows([]legacyTreatmentHistoryRow{row}, true)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

type TreatmentDuringParamRequest struct {
	RecordTime       *time.Time `json:"recordTime"`
	Code             string     `json:"code"`
	BloodFlow        *float64   `json:"bloodFlow"`
	DialysateFlow    *float64   `json:"dialysateFlow"`
	UFVolume         *float64   `json:"ufVolume"`
	SBP              *float64   `json:"sbp"`
	DBP              *float64   `json:"dbp"`
	HeartRate        *float64   `json:"heartRate"`
	Respiration      *float64   `json:"respiration"`
	SpO2             *float64   `json:"spO2"`
	VenousPressure   *float64   `json:"venousPressure"`
	ArterialPressure *float64   `json:"arterialPressure"`
	TMP              *float64   `json:"tmp"`
	Temperature      *float64   `json:"temperature"`
	Conductivity     *float64   `json:"conductivity"`
	UFRate           *float64   `json:"ufRate"`
	Notes            string     `json:"notes"`
}

type TreatmentBeforeSignsRequest struct {
	Weight       *float64               `json:"weight"`
	ExtraWeight  *float64               `json:"extraWeight"`
	SBP          *float64               `json:"sbp"`
	DBP          *float64               `json:"dbp"`
	HeartRate    *float64               `json:"heartRate"`
	Respiration  *float64               `json:"respiration"`
	Temperature  *float64               `json:"temperature"`
	PressurePos  string                 `json:"pressurePoint"`
	Notes        string                 `json:"notes"`
	SymptomItems []TreatmentSymptomItem `json:"symptomItems"`
}

type TreatmentAfterSignsRequest struct {
	StartTime     *time.Time             `json:"startTime"`
	EndTime       *time.Time             `json:"endTime"`
	RealUFVolume  *float64               `json:"realUfVolume"`
	RealSubVolume *float64               `json:"realSubstituteVolume"`
	Weight        *float64               `json:"weight"`
	ExtraWeight   *float64               `json:"extraWeight"`
	LossWeight    *float64               `json:"lossWeight"`
	SBP           *float64               `json:"sbp"`
	DBP           *float64               `json:"dbp"`
	HeartRate     *float64               `json:"heartRate"`
	Respiration   *float64               `json:"respiration"`
	Temperature   *float64               `json:"temperature"`
	RealIntake    *float64               `json:"realIntake"`
	PressurePos   string                 `json:"pressurePoint"`
	Complication  string                 `json:"complication"`
	Symptoms      string                 `json:"symptoms"`
	Notes         string                 `json:"notes"`
	SymptomItems  []TreatmentSymptomItem `json:"symptomItems"`
}

type TreatmentFirstCheckRequest struct {
	BeforeSignsID         *int64     `json:"beforeSignsId"`
	BeforeSymptomID       *int64     `json:"beforeSymptomId"`
	OperatorID            *int64     `json:"operatorId"`
	OperateTime           *time.Time `json:"operateTime"`
	MaterialsResult       *bool      `json:"materialsResult"`
	MaterialsMistake      *string    `json:"materialsMistake"`
	ParamResult           *bool      `json:"paramResult"`
	ParamMistake          *string    `json:"paramMistake"`
	VascularAccessResult  *bool      `json:"vascularAccessResult"`
	VascularAccessMistake *string    `json:"vascularAccessMistake"`
	PipelineResult        *bool      `json:"pipelineResult"`
	PipelineMistake       *string    `json:"pipelineMistake"`
}

type TreatmentSecondCheckRequest struct {
	OperatorID            *int64     `json:"operatorId"`
	RecheckNurseID        *int64     `json:"recheckNurseId"`
	QCNurseID             *int64     `json:"qcNurseId"`
	OperateTime           *time.Time `json:"operateTime"`
	ParamResult           *bool      `json:"paramResult"`
	ParamMistake          *string    `json:"paramMistake"`
	VascularAccessResult  *bool      `json:"vascularAccessResult"`
	VascularAccessMistake *string    `json:"vascularAccessMistake"`
	PipelineResult        *bool      `json:"pipelineResult"`
	PipelineMistake       *string    `json:"pipelineMistake"`
	DialysisModeResult    *bool      `json:"dialysisModeResult"`
	DialysisModeMistake   *string    `json:"dialysisModeMistake"`
	PrescriptionResult    *bool      `json:"prescriptionResult"`
	PrescriptionMistake   *string    `json:"prescriptionMistake"`
	AnticoagulantResult   *bool      `json:"anticoagulantResult"`
	AnticoagulantMistake  *string    `json:"anticoagulantMistake"`
	LineConnectionResult  *bool      `json:"lineConnectionResult"`
	LineConnectionMistake *string    `json:"lineConnectionMistake"`
}

type TreatmentSymptomItem struct {
	Code  string `json:"code"`
	Value string `json:"value"`
}

type TreatmentSignsResponse struct {
	ID             int64      `json:"id"`
	TenantID       int64      `json:"tenantId"`
	TreatmentID    int64      `json:"treatmentId"`
	Weight         *float64   `json:"weight,omitempty"`
	ExtraWeight    *float64   `json:"extraWeight,omitempty"`
	LossWeight     *float64   `json:"lossWeight,omitempty"`
	SBP            *float64   `json:"sbp,omitempty"`
	DBP            *float64   `json:"dbp,omitempty"`
	HeartRate      *float64   `json:"heartRate,omitempty"`
	Respiration    *float64   `json:"respiration,omitempty"`
	Temperature    *float64   `json:"temperature,omitempty"`
	RealIntake     *float64   `json:"realIntake,omitempty"`
	PressurePoint  string     `json:"pressurePoint,omitempty"`
	Complication   string     `json:"complication,omitempty"`
	Symptoms       string     `json:"symptoms,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	CreatorID      int64      `json:"creatorId"`
	CreateTime     *time.Time `json:"createTime,omitempty"`
	LastModifyTime *time.Time `json:"lastModifyTime,omitempty"`
}

type legacyBeforeSignsRow struct {
	ID             int64      `gorm:"column:Id"`
	TreatmentID    int64      `gorm:"column:TreatmentId"`
	OperatorID     int64      `gorm:"column:OperatorId"`
	OperateTime    time.Time  `gorm:"column:OperateTime"`
	Weight         *float64   `gorm:"column:Weight"`
	ExtraWeight    *float64   `gorm:"column:ExtraWeight"`
	SBP            *float64   `gorm:"column:SBP"`
	DBP            *float64   `gorm:"column:DBP"`
	HeartRate      *float64   `gorm:"column:HeartRate"`
	Respiration    *float64   `gorm:"column:Respiration"`
	BodyTemp       *float64   `gorm:"column:BodyTemp"`
	PressurePoint  string     `gorm:"column:PressurePoint"`
	Note           string     `gorm:"column:Note"`
	CreatorID      int64      `gorm:"column:CreatorId"`
	CreateTime     *time.Time `gorm:"column:CreateTime"`
	LastModifyTime *time.Time `gorm:"column:LastModifyTime"`
}

type legacyAfterSignsRow struct {
	ID             int64      `gorm:"column:Id"`
	TreatmentID    int64      `gorm:"column:TreatmentId"`
	Weight         *float64   `gorm:"column:Weight"`
	ExtraWeight    *float64   `gorm:"column:ExtraWeight"`
	LossWeight     *float64   `gorm:"column:LossWeight"`
	SBP            *float64   `gorm:"column:SBP"`
	DBP            *float64   `gorm:"column:DBP"`
	HeartRate      *float64   `gorm:"column:HeartRate"`
	Respiration    *float64   `gorm:"column:Respiration"`
	BodyTemp       *float64   `gorm:"column:BodyTemp"`
	RealIntake     *float64   `gorm:"column:RealIntake"`
	PressurePoint  string     `gorm:"column:PressurePoint"`
	Complication   string     `gorm:"column:Complication"`
	Symptoms       string     `gorm:"column:Symptoms"`
	Note           string     `gorm:"column:Note"`
	CreatorID      int64      `gorm:"column:CreatorId"`
	CreateTime     *time.Time `gorm:"column:CreateTime"`
	LastModifyTime *time.Time `gorm:"column:LastModifyTime"`
}

type legacyDuringParamIdentityRow struct {
	ID          int64     `gorm:"column:Id"`
	TreatmentID int64     `gorm:"column:TreatmentId"`
	OperateTime time.Time `gorm:"column:OperateTime"`
	CreatorID   int64     `gorm:"column:CreatorId"`
}

func (s *TreatmentService) upsertDuringSignsByTime(treatmentID int64, oldTime *time.Time, targetTime time.Time, req TreatmentDuringParamRequest, creatorID int64) error {
	if req.SBP == nil && req.DBP == nil && req.HeartRate == nil && req.Respiration == nil && req.SpO2 == nil && req.Temperature == nil {
		return nil
	}
	now := time.Now()
	query := s.db.Table(`"Treatment_DuringSigns"`).
		Where(`"TenantId" = ? AND "TreatmentId" = ?`, LegacyTenantID, treatmentID)
	if oldTime != nil && !oldTime.IsZero() {
		query = query.Where(`"OperateTime" = ?`, *oldTime)
	} else {
		query = query.Where(`"OperateTime" = ?`, targetTime)
	}
	var existing struct {
		ID int64 `gorm:"column:Id"`
	}
	err := query.Select(`"Id"`).Take(&existing).Error
	values := map[string]any{
		"OperateTime":    targetTime,
		"LastModifyTime": now,
	}
	if req.SBP != nil {
		values["SBP"] = req.SBP
	}
	if req.DBP != nil {
		values["DBP"] = req.DBP
	}
	if req.HeartRate != nil {
		values["HeartRate"] = req.HeartRate
	}
	if req.Respiration != nil {
		values["Respiration"] = req.Respiration
	}
	if req.SpO2 != nil {
		values["SpO2"] = req.SpO2
	}
	if req.Temperature != nil {
		values["BodyTemp"] = req.Temperature
	}
	if err == nil {
		return s.db.Table(`"Treatment_DuringSigns"`).
			Where(`"Id" = ? AND "TenantId" = ?`, existing.ID, LegacyTenantID).
			Updates(values).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	newID, idErr := nextLegacyID()
	if idErr != nil {
		return idErr
	}
	values["Id"] = newID
	values["TenantId"] = LegacyTenantID
	values["TreatmentId"] = treatmentID
	values["OperatorId"] = creatorID
	values["CreatorId"] = creatorID
	values["CreateTime"] = now
	return s.db.Table(`"Treatment_DuringSigns"`).Create(values).Error
}

func (s *TreatmentService) CreateDuringParam(treatmentID int64, req TreatmentDuringParamRequest, creatorID int64) (*TreatmentDuringParamDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var count int64
	if err := s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("treatment not found")
	}

	paramID, err := nextLegacyID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	recordTime := now
	if req.RecordTime != nil && !req.RecordTime.IsZero() {
		recordTime = *req.RecordTime
	}

	values := map[string]any{
		"Id":               paramID,
		"TenantId":         LegacyTenantID,
		"TreatmentId":      treatmentID,
		"OperateTime":      recordTime,
		"VenousPressure":   req.VenousPressure,
		"ArterialPressure": req.ArterialPressure,
		"TMP":              req.TMP,
		"Conductivity":     req.Conductivity,
		"UFQuantity":       req.UFVolume,
		"MachineTmp":       req.Temperature,
		"BF":               req.BloodFlow,
		"CreatorId":        creatorID,
		"CreateTime":       now,
		"LastModifyTime":   now,
	}
	if err := s.db.Table(`"Treatment_DuringParam"`).Create(values).Error; err != nil {
		return nil, err
	}
	if err := s.upsertDuringSignsByTime(treatmentID, nil, recordTime, req, creatorID); err != nil {
		return nil, err
	}
	if err := s.upsertTreatmentSignsJSONSnapshot(treatmentID, creatorID, legacyJSONCodeDuringOther, map[string]any{
		"recordTime":    recordTime,
		"code":          strings.TrimSpace(req.Code),
		"notes":         strings.TrimSpace(req.Notes),
		"bloodFlow":     req.BloodFlow,
		"dialysateFlow": req.DialysateFlow,
		"ufVolume":      req.UFVolume,
		"sbp":           req.SBP,
		"dbp":           req.DBP,
		"heartRate":     req.HeartRate,
		"respiration":   req.Respiration,
		"spO2":          req.SpO2,
	}); err != nil {
		return nil, err
	}

	return &TreatmentDuringParamDTO{
		ID:               paramID.Int64(),
		TenantID:         LegacyTenantID,
		TreatmentID:      treatmentID,
		RecordTime:       recordTime,
		Code:             strings.TrimSpace(req.Code),
		BloodFlow:        req.BloodFlow,
		DialysateFlow:    req.DialysateFlow,
		UFVolume:         req.UFVolume,
		VenousPressure:   req.VenousPressure,
		ArterialPressure: req.ArterialPressure,
		TMP:              req.TMP,
		Temperature:      req.Temperature,
		Conductivity:     req.Conductivity,
		UFRate:           req.UFRate,
		SBP:              req.SBP,
		DBP:              req.DBP,
		HeartRate:        req.HeartRate,
		Respiration:      req.Respiration,
		SpO2:             req.SpO2,
		Notes:            strings.TrimSpace(req.Notes),
		CreatorID:        creatorID,
		CreateTime:       &now,
		LastModifyTime:   &now,
	}, nil
}

func (s *TreatmentService) UpdateDuringParam(treatmentID, paramID int64, req TreatmentDuringParamRequest) (*TreatmentDuringParamDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var existing legacyDuringParamIdentityRow
	if err := s.db.Table(`"Treatment_DuringParam"`).
		Select(`"Id", "TreatmentId", "OperateTime", "CreatorId"`).
		Where(`"Id" = ? AND "TreatmentId" = ? AND "TenantId" = ?`, paramID, treatmentID, LegacyTenantID).
		Take(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("during param not found")
		}
		return nil, err
	}

	updates := map[string]any{
		"LastModifyTime": time.Now(),
	}
	targetRecordTime := existing.OperateTime
	if req.RecordTime != nil && !req.RecordTime.IsZero() {
		updates["OperateTime"] = *req.RecordTime
		targetRecordTime = *req.RecordTime
	}
	if req.VenousPressure != nil {
		updates["VenousPressure"] = req.VenousPressure
	}
	if req.ArterialPressure != nil {
		updates["ArterialPressure"] = req.ArterialPressure
	}
	if req.TMP != nil {
		updates["TMP"] = req.TMP
	}
	if req.Conductivity != nil {
		updates["Conductivity"] = req.Conductivity
	}
	if req.UFVolume != nil {
		updates["UFQuantity"] = req.UFVolume
	}
	if req.Temperature != nil {
		updates["MachineTmp"] = req.Temperature
	}
	if req.BloodFlow != nil {
		updates["BF"] = req.BloodFlow
	}

	result := s.db.Table(`"Treatment_DuringParam"`).
		Where(`"Id" = ? AND "TreatmentId" = ? AND "TenantId" = ?`, paramID, treatmentID, LegacyTenantID).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("during param not found")
	}
	if err := s.upsertDuringSignsByTime(treatmentID, &existing.OperateTime, targetRecordTime, req, existing.CreatorID); err != nil {
		return nil, err
	}
	if err := s.upsertTreatmentSignsJSONSnapshot(treatmentID, existing.CreatorID, legacyJSONCodeDuringOther, map[string]any{
		"recordTime":    targetRecordTime,
		"code":          strings.TrimSpace(req.Code),
		"notes":         strings.TrimSpace(req.Notes),
		"bloodFlow":     req.BloodFlow,
		"dialysateFlow": req.DialysateFlow,
		"ufVolume":      req.UFVolume,
		"sbp":           req.SBP,
		"dbp":           req.DBP,
		"heartRate":     req.HeartRate,
		"respiration":   req.Respiration,
		"spO2":          req.SpO2,
	}); err != nil {
		return nil, err
	}

	params, err := s.loadDuringParamMap([]legacyTreatmentHistoryRow{{ID: treatmentID}})
	if err != nil {
		return nil, err
	}
	for _, item := range params[treatmentID] {
		if item.ID == paramID {
			item.Code = strings.TrimSpace(req.Code)
			if item.Code == "" {
				item.Code = "legacy"
			}
			if req.DialysateFlow != nil {
				item.DialysateFlow = req.DialysateFlow
			}
			if req.UFRate != nil {
				item.UFRate = req.UFRate
			}
			if req.SBP != nil {
				item.SBP = req.SBP
			}
			if req.DBP != nil {
				item.DBP = req.DBP
			}
			if req.HeartRate != nil {
				item.HeartRate = req.HeartRate
			}
			if req.Respiration != nil {
				item.Respiration = req.Respiration
			}
			if req.SpO2 != nil {
				item.SpO2 = req.SpO2
			}
			return &item, nil
		}
	}
	return nil, errors.New("during param not found")
}

func (s *TreatmentService) DeleteDuringParam(treatmentID, paramID int64) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Table(`"Treatment_DuringParam"`).
		Where(`"Id" = ? AND "TreatmentId" = ? AND "TenantId" = ?`, paramID, treatmentID, LegacyTenantID).
		Delete(map[string]any{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("during param not found")
	}
	return nil
}

func (s *TreatmentService) SaveBeforeSigns(treatmentID int64, req TreatmentBeforeSignsRequest, creatorID int64) (*TreatmentSignsResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	values := map[string]any{
		"Weight":        req.Weight,
		"ExtraWeight":   req.ExtraWeight,
		"SBP":           req.SBP,
		"DBP":           req.DBP,
		"HeartRate":     req.HeartRate,
		"Respiration":   req.Respiration,
		"BodyTemp":      req.Temperature,
		"PressurePoint": strings.TrimSpace(req.PressurePos),
		"Note":          strings.TrimSpace(req.Notes),
	}
	row, err := s.upsertTreatmentSigns(`"Treatment_BeforeSigns"`, treatmentID, creatorID, values)
	if err != nil {
		return nil, err
	}
	if err := s.replaceTreatmentSymptomItems(`"Treatment_BeforeSymptom"`, treatmentID, creatorID, req.SymptomItems); err != nil {
		return nil, err
	}
	if err := s.upsertTreatmentSignsJSONSnapshot(treatmentID, creatorID, legacyJSONCodeBeforeSymptom, req); err != nil {
		return nil, err
	}
	return mapBeforeSignsResponse(row), nil
}

func (s *TreatmentService) SaveAfterSigns(treatmentID int64, req TreatmentAfterSignsRequest, creatorID int64) (*TreatmentSignsResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	values := map[string]any{
		"Weight":        req.Weight,
		"ExtraWeight":   req.ExtraWeight,
		"LossWeight":    req.LossWeight,
		"SBP":           req.SBP,
		"DBP":           req.DBP,
		"HeartRate":     req.HeartRate,
		"Respiration":   req.Respiration,
		"BodyTemp":      req.Temperature,
		"RealIntake":    req.RealIntake,
		"PressurePoint": strings.TrimSpace(req.PressurePos),
		"Note":          strings.TrimSpace(req.Notes),
	}
	row, err := s.upsertTreatmentSigns(`"Treatment_AfterSigns"`, treatmentID, creatorID, values)
	if err != nil {
		return nil, err
	}
	if err := s.updateTreatmentSummaryFields(treatmentID, req); err != nil {
		return nil, err
	}
	if err := s.replaceTreatmentSymptomItems(`"Treatment_AfterSymptom"`, treatmentID, creatorID, req.SymptomItems); err != nil {
		return nil, err
	}
	if err := s.upsertTreatmentSignsJSONSnapshot(treatmentID, creatorID, legacyJSONCodeAfterSymptom, req); err != nil {
		return nil, err
	}
	if err := s.upsertTreatmentSignsJSONSnapshot(treatmentID, creatorID, legacyJSONCodeTreatmentDetail, map[string]any{
		"complication": req.Complication,
		"symptoms":     req.Symptoms,
		"symptomItems": req.SymptomItems,
		"notes":        req.Notes,
	}); err != nil {
		return nil, err
	}
	if err := s.upsertTreatmentSignsJSONSnapshot(treatmentID, creatorID, legacyJSONCodeTreatmentFeelContent, map[string]any{
		"notes": req.Notes,
	}); err != nil {
		return nil, err
	}
	return mapAfterSignsResponse(row), nil
}

func (s *TreatmentService) SaveFirstCheck(treatmentID int64, req TreatmentFirstCheckRequest, creatorID int64) (*TreatmentFirstCheckSnapshot, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var count int64
	if err := s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("treatment not found")
	}

	now := time.Now()
	operatorID := creatorID
	if req.OperatorID != nil && *req.OperatorID > 0 {
		operatorID = *req.OperatorID
	}
	operateTime := now
	if req.OperateTime != nil && !req.OperateTime.IsZero() {
		operateTime = *req.OperateTime
	}
	_, _, actionErr := s.upsertLegacyAction(treatmentID, "首次核对", legacyActionCodeFirstCheck, operatorID, operateTime, creatorID)
	if actionErr != nil {
		return nil, actionErr
	}

	values := map[string]any{
		"OperatorId":     operatorID,
		"OperateTime":    operateTime,
		"LastModifyTime": now,
	}
	if req.BeforeSignsID != nil {
		values["BeforeSignsId"] = req.BeforeSignsID
	}
	if req.BeforeSymptomID != nil {
		values["BeforeSymptomId"] = req.BeforeSymptomID
	}
	if req.MaterialsResult != nil {
		values["MaterialsResult"] = req.MaterialsResult
	}
	if req.MaterialsMistake != nil {
		values["MaterialsMistake"] = strings.TrimSpace(*req.MaterialsMistake)
	}
	if req.ParamResult != nil {
		values["ParamResult"] = req.ParamResult
	}
	if req.ParamMistake != nil {
		values["ParamMistake"] = strings.TrimSpace(*req.ParamMistake)
	}
	if req.VascularAccessResult != nil {
		values["VascularAccessResult"] = req.VascularAccessResult
	}
	if req.VascularAccessMistake != nil {
		values["VascularAccessMistake"] = strings.TrimSpace(*req.VascularAccessMistake)
	}
	if req.PipelineResult != nil {
		values["PipelineResult"] = req.PipelineResult
	}
	if req.PipelineMistake != nil {
		values["PipelineMistake"] = strings.TrimSpace(*req.PipelineMistake)
	}

	var existing struct {
		ID int64 `gorm:"column:Id"`
	}
	err := s.db.Table(`"Treatment_BeforeCheck"`).
		Select(`"Id"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Take(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newID, idErr := nextLegacyID()
		if idErr != nil {
			return nil, idErr
		}
		values["Id"] = newID
		values["TenantId"] = LegacyTenantID
		values["TreatmentId"] = treatmentID
		values["CreatorId"] = creatorID
		values["CreateTime"] = now
		if err := s.db.Table(`"Treatment_BeforeCheck"`).Create(values).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Table(`"Treatment_BeforeCheck"`).
			Where(`"Id" = ? AND "TreatmentId" = ? AND "TenantId" = ?`, existing.ID, treatmentID, LegacyTenantID).
			Updates(values).Error; err != nil {
			return nil, err
		}
	}

	var saved legacyBeforeCheckRow
	if err := s.db.Table(`"Treatment_BeforeCheck"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		First(&saved).Error; err != nil {
		return nil, err
	}
	return mapFirstCheckSnapshot(saved), nil
}

func (s *TreatmentService) SaveSecondCheck(treatmentID int64, req TreatmentSecondCheckRequest, creatorID int64) (*TreatmentSecondCheckSnapshot, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var count int64
	if err := s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("treatment not found")
	}

	now := time.Now()
	operatorID := creatorID
	if req.OperatorID != nil && *req.OperatorID > 0 {
		operatorID = *req.OperatorID
	}
	operateTime := now
	if req.OperateTime != nil && !req.OperateTime.IsZero() {
		operateTime = *req.OperateTime
	}

	// 双人核对独立性（防错核心）：二次核对操作人不得与首次核对同一人。
	// 服务端强制，不仅依赖前端下拉过滤（否则直调接口可绕过）。首核未做则不在此拦。
	var firstCheck struct {
		OperatorID int64 `gorm:"column:OperatorId"`
	}
	ferr := s.db.Table(`"Treatment_BeforeCheck"`).
		Select(`"OperatorId"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Take(&firstCheck).Error
	if ferr != nil && !errors.Is(ferr, gorm.ErrRecordNotFound) {
		return nil, ferr
	}
	if ferr == nil && firstCheck.OperatorID > 0 && firstCheck.OperatorID == operatorID {
		return nil, errors.New("二次核对需独立复核，不可与首次核对为同一人")
	}

	actionID, actionCreateTime, actionErr := s.upsertLegacyAction(treatmentID, "二次核对", legacyActionCodeAgainCheck, operatorID, operateTime, creatorID)
	if actionErr != nil {
		return nil, actionErr
	}

	mergeResult := func(primary, a, b, c *bool) *bool {
		if primary != nil {
			return primary
		}
		hasTrue := false
		hasFalse := false
		candidates := []*bool{a, b, c}
		for _, item := range candidates {
			if item == nil {
				continue
			}
			if *item {
				hasTrue = true
			} else {
				hasFalse = true
			}
		}
		if hasFalse {
			v := false
			return &v
		}
		if hasTrue {
			v := true
			return &v
		}
		return nil
	}
	mergeMistake := func(primary *string, values ...*string) string {
		if primary != nil {
			return strings.TrimSpace(*primary)
		}
		parts := make([]string, 0, len(values))
		for _, item := range values {
			if item == nil {
				continue
			}
			text := strings.TrimSpace(*item)
			if text == "" {
				continue
			}
			parts = append(parts, text)
		}
		return strings.Join(parts, "；")
	}

	payload := treatmentSecondCheckPayload{
		OperatorID:            &operatorID,
		RecheckNurseID:        req.RecheckNurseID,
		QCNurseID:             req.QCNurseID,
		OperateTime:           &operateTime,
		ParamResult:           mergeResult(req.ParamResult, req.DialysisModeResult, req.PrescriptionResult, req.AnticoagulantResult),
		ParamMistake:          mergeMistake(req.ParamMistake, req.DialysisModeMistake, req.PrescriptionMistake, req.AnticoagulantMistake),
		VascularAccessResult:  req.VascularAccessResult,
		VascularAccessMistake: strings.TrimSpace(derefSecondCheckString(req.VascularAccessMistake)),
		PipelineResult:        mergeResult(req.PipelineResult, req.LineConnectionResult, nil, nil),
		PipelineMistake:       mergeMistake(req.PipelineMistake, req.LineConnectionMistake),
		DialysisModeResult:    req.DialysisModeResult,
		DialysisModeMistake:   strings.TrimSpace(derefSecondCheckString(req.DialysisModeMistake)),
		PrescriptionResult:    req.PrescriptionResult,
		PrescriptionMistake:   strings.TrimSpace(derefSecondCheckString(req.PrescriptionMistake)),
		AnticoagulantResult:   req.AnticoagulantResult,
		AnticoagulantMistake:  strings.TrimSpace(derefSecondCheckString(req.AnticoagulantMistake)),
		LineConnectionResult:  req.LineConnectionResult,
		LineConnectionMistake: strings.TrimSpace(derefSecondCheckString(req.LineConnectionMistake)),
	}
	if err := s.upsertTreatmentSignsJSONSnapshot(treatmentID, creatorID, legacyJSONCodeAgainCheck, payload); err != nil {
		return nil, err
	}

	return &TreatmentSecondCheckSnapshot{
		ActionID:              actionID,
		TreatmentID:           treatmentID,
		OperatorID:            &operatorID,
		RecheckNurseID:        req.RecheckNurseID,
		QCNurseID:             req.QCNurseID,
		OperateTime:           &operateTime,
		ParamResult:           payload.ParamResult,
		ParamMistake:          payload.ParamMistake,
		VascularAccessResult:  payload.VascularAccessResult,
		VascularAccessMistake: payload.VascularAccessMistake,
		PipelineResult:        payload.PipelineResult,
		PipelineMistake:       payload.PipelineMistake,
		DialysisModeResult:    payload.DialysisModeResult,
		DialysisModeMistake:   payload.DialysisModeMistake,
		PrescriptionResult:    payload.PrescriptionResult,
		PrescriptionMistake:   payload.PrescriptionMistake,
		AnticoagulantResult:   payload.AnticoagulantResult,
		AnticoagulantMistake:  payload.AnticoagulantMistake,
		LineConnectionResult:  payload.LineConnectionResult,
		LineConnectionMistake: payload.LineConnectionMistake,
		CreateTime:            actionCreateTime,
		LastModifyTime:        &now,
	}, nil
}

func (s *TreatmentService) upsertLegacyAction(treatmentID int64, name, code string, operatorID int64, operateTime time.Time, creatorID int64) (int64, *time.Time, error) {
	now := time.Now()
	var existingAction struct {
		ID         int64      `gorm:"column:Id"`
		CreateTime *time.Time `gorm:"column:CreateTime"`
	}
	actionErr := s.db.Table(`"Treatment_Action"`).
		Select(`"Id", "CreateTime"`).
		Where(`"TenantId" = ? AND "TreatmentId" = ? AND ("Code" = ? OR "Name" = ?)`,
			LegacyTenantID, treatmentID, strings.TrimSpace(code), strings.TrimSpace(name)).
		Order(`"OperateTime" DESC`).
		Order(`"Id" DESC`).
		Take(&existingAction).Error
	if actionErr != nil && !errors.Is(actionErr, gorm.ErrRecordNotFound) {
		return 0, nil, actionErr
	}
	if errors.Is(actionErr, gorm.ErrRecordNotFound) {
		newActionID, idErr := nextLegacyID()
		if idErr != nil {
			return 0, nil, idErr
		}
		row := map[string]any{
			"Id":             newActionID,
			"TenantId":       LegacyTenantID,
			"TreatmentId":    treatmentID,
			"Name":           strings.TrimSpace(name),
			"OperatorId":     operatorID,
			"OperateTime":    operateTime,
			"CreatorId":      creatorID,
			"CreateTime":     now,
			"LastModifyTime": now,
			"Code":           strings.TrimSpace(code),
		}
		if err := s.db.Table(`"Treatment_Action"`).Create(row).Error; err != nil {
			return 0, nil, err
		}
		return newActionID.Int64(), &now, nil
	}

	if err := s.db.Table(`"Treatment_Action"`).
		Where(`"Id" = ? AND "TenantId" = ?`, existingAction.ID, LegacyTenantID).
		Updates(map[string]any{
			"Name":           strings.TrimSpace(name),
			"OperatorId":     operatorID,
			"OperateTime":    operateTime,
			"LastModifyTime": now,
			"Code":           strings.TrimSpace(code),
		}).Error; err != nil {
		return 0, nil, err
	}
	return existingAction.ID, existingAction.CreateTime, nil
}

func derefSecondCheckString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *TreatmentService) resolveTreatmentPatientID(treatmentID int64) (int64, error) {
	var row struct {
		PatientID int64 `gorm:"column:PatientId"`
	}
	err := s.db.Table(`"Treatment_Treatment"`).
		Select(`"PatientId"`).
		Where(`"Id" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Limit(1).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("treatment not found")
		}
		return 0, err
	}
	return row.PatientID, nil
}

func (s *TreatmentService) resolveTreatmentPatientAndMode(treatmentID int64) (modeltypes.LegacyID, string, error) {
	patientID, err := s.resolveTreatmentPatientID(treatmentID)
	if err != nil {
		return 0, "", err
	}

	var row struct {
		DialysisMethod string `gorm:"column:DialysisMethod"`
	}
	err = s.db.Table(`"Plan_PatientPrescription"`).
		Select(`"DialysisMethod"`).
		Where(`"TenantId" = ? AND "TreatmentId" = ?`, LegacyTenantID, treatmentID).
		Order(`"LastModifyTime" DESC`).
		Order(`"Id" DESC`).
		Limit(1).
		Take(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, "", err
	}
	return modeltypes.LegacyID(patientID), strings.TrimSpace(row.DialysisMethod), nil
}

// ensurePlanCompleteForTreatment 上机门禁：先按治疗解析患者和本次处方模式，再校验方案完整（契约02 三）。
func (s *TreatmentService) ensurePlanCompleteForTreatment(treatmentID int64) error {
	patientID, mode, err := s.resolveTreatmentPatientAndMode(treatmentID)
	if err != nil {
		return err
	}
	return s.ensurePlanCompleteForPatientAndMode(patientID, mode)
}

// ensurePlanCompleteForPatient 校验患者治疗方案是否完整（契约02 三，上机门禁核心）。
// 草稿态（透析液配方为空）或无方案时返回拦截错误，引导医生先到治疗方案页补全。
// 三条上机写入路径（Create / Update / UpdateStatus 进入 InProgress）共用此校验。
func (s *TreatmentService) ensurePlanCompleteForPatient(patientID modeltypes.LegacyID) error {
	return s.ensurePlanCompleteForPatientAndMode(patientID, "")
}

func (s *TreatmentService) ensurePlanCompleteForPatientAndMode(patientID modeltypes.LegacyID, mode string) error {
	planService := &PatientService{db: s.db}
	plan, err := planService.legacyPlanByMode(patientID, mode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("该患者尚无治疗方案，请先到治疗方案页创建并补全方案后再上机")
		}
		return err
	}
	if !isLegacyPlanComplete(plan) {
		return errors.New("治疗方案尚未补全（透析液配方为空），请先到治疗方案页完善方案后再上机")
	}
	return nil
}

func (s *TreatmentService) upsertTreatmentSignsJSONSnapshot(treatmentID, creatorID int64, code string, payload any) error {
	if strings.TrimSpace(code) == "" {
		return nil
	}
	patientID, err := s.resolveTreatmentPatientID(treatmentID)
	if err != nil {
		return err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	now := time.Now()
	var existing struct {
		ID int64 `gorm:"column:Id"`
	}
	findErr := s.db.Table(`"Auxiliary_JsonData"`).
		Select(`"Id"`).
		Where(`"TenantId" = ? AND "TreatmentId" = ? AND "Code" = ?`, LegacyTenantID, treatmentID, code).
		Order(`"LastModifyTime" DESC`).
		Order(`"CreateTime" DESC`).
		Order(`"Id" DESC`).
		First(&existing).Error
	if findErr == nil {
		return s.db.Table(`"Auxiliary_JsonData"`).
			Where(`"Id" = ? AND "TenantId" = ?`, existing.ID, LegacyTenantID).
			Updates(map[string]any{
				"PatientId":      patientID,
				"Value":          json.RawMessage(data),
				"LastModifyTime": now,
			}).Error
	}
	if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return findErr
	}
	newID, idErr := nextLegacyID()
	if idErr != nil {
		return idErr
	}
	row := map[string]any{
		"Id":             newID,
		"TenantId":       LegacyTenantID,
		"PatientId":      patientID,
		"TreatmentId":    treatmentID,
		"Code":           code,
		"CreatorId":      creatorID,
		"CreateTime":     now,
		"LastModifyTime": now,
		"Value":          json.RawMessage(data),
	}
	return s.db.Table(`"Auxiliary_JsonData"`).Create(row).Error
}

func (s *TreatmentService) updateTreatmentSummaryFields(treatmentID int64, req TreatmentAfterSignsRequest) error {
	updates := map[string]any{
		"LastModifyTime": time.Now(),
	}
	if req.StartTime != nil && !req.StartTime.IsZero() {
		updates["StartTime"] = *req.StartTime
	}
	if req.EndTime != nil && !req.EndTime.IsZero() {
		updates["EndTime"] = *req.EndTime
	}
	if req.RealUFVolume != nil {
		updates["RealUFQuantity"] = *req.RealUFVolume
	}
	if req.RealSubVolume != nil {
		updates["RealSubstituateVolume"] = *req.RealSubVolume
	}
	if req.StartTime != nil && req.EndTime != nil && !req.StartTime.IsZero() && !req.EndTime.IsZero() && req.EndTime.After(*req.StartTime) {
		updates["RealDuration"] = req.EndTime.Sub(*req.StartTime).Minutes()
	}
	return s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Updates(updates).Error
}

func (s *TreatmentService) replaceTreatmentSymptomItems(table string, treatmentID, creatorID int64, items []TreatmentSymptomItem) error {
	if err := s.db.Table(table).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Delete(map[string]any{}).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		value := strings.TrimSpace(item.Value)
		if code == "" || value == "" {
			continue
		}
		id, err := nextLegacyID()
		if err != nil {
			return err
		}
		values := map[string]any{
			"Id":             id,
			"TenantId":       LegacyTenantID,
			"TreatmentId":    treatmentID,
			"OperatorId":     creatorID,
			"OperateTime":    now,
			"Code":           code,
			"Value":          value,
			"CreatorId":      creatorID,
			"CreateTime":     now,
			"LastModifyTime": now,
		}
		if err := s.db.Table(table).Create(values).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *TreatmentService) upsertTreatmentSigns(table string, treatmentID, creatorID int64, values map[string]any) (map[string]any, error) {
	var count int64
	if err := s.db.Table(`"Treatment_Treatment"`).
		Where(`"Id" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("treatment not found")
	}

	now := time.Now()
	var existing struct {
		ID int64 `gorm:"column:Id"`
	}
	err := s.db.Table(table).
		Select(`"Id"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Take(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newID, idErr := nextLegacyID()
		if idErr != nil {
			return nil, idErr
		}
		values["Id"] = newID
		values["TenantId"] = LegacyTenantID
		values["TreatmentId"] = treatmentID
		values["CreatorId"] = creatorID
		values["OperatorId"] = creatorID
		values["OperateTime"] = now
		values["CreateTime"] = now
		values["LastModifyTime"] = now
		if err := s.db.Table(table).Create(values).Error; err != nil {
			return nil, err
		}
		values["Id"] = newID.Int64()
		values["CreatorId"] = creatorID
		values["CreateTime"] = now
		values["LastModifyTime"] = now
		return values, nil
	}

	values["LastModifyTime"] = now
	if err := s.db.Table(table).
		Where(`"Id" = ? AND "TreatmentId" = ? AND "TenantId" = ?`, existing.ID, treatmentID, LegacyTenantID).
		Updates(values).Error; err != nil {
		return nil, err
	}
	values["Id"] = existing.ID
	values["TenantId"] = LegacyTenantID
	values["TreatmentId"] = treatmentID
	values["LastModifyTime"] = now
	return values, nil
}

func mapBeforeSignsResponse(values map[string]any) *TreatmentSignsResponse {
	return &TreatmentSignsResponse{
		ID:             valueInt64(values["Id"]),
		TenantID:       LegacyTenantID,
		TreatmentID:    valueInt64(values["TreatmentId"]),
		Weight:         valueFloat64Ptr(values["Weight"]),
		ExtraWeight:    valueFloat64Ptr(values["ExtraWeight"]),
		SBP:            valueFloat64Ptr(values["SBP"]),
		DBP:            valueFloat64Ptr(values["DBP"]),
		HeartRate:      valueFloat64Ptr(values["HeartRate"]),
		Respiration:    valueFloat64Ptr(values["Respiration"]),
		Temperature:    valueFloat64Ptr(values["BodyTemp"]),
		PressurePoint:  valueString(values["PressurePoint"]),
		Notes:          valueString(values["Note"]),
		CreatorID:      valueInt64(values["CreatorId"]),
		CreateTime:     valueTimePtr(values["CreateTime"]),
		LastModifyTime: valueTimePtr(values["LastModifyTime"]),
	}
}

func mapAfterSignsResponse(values map[string]any) *TreatmentSignsResponse {
	return &TreatmentSignsResponse{
		ID:             valueInt64(values["Id"]),
		TenantID:       LegacyTenantID,
		TreatmentID:    valueInt64(values["TreatmentId"]),
		Weight:         valueFloat64Ptr(values["Weight"]),
		ExtraWeight:    valueFloat64Ptr(values["ExtraWeight"]),
		LossWeight:     valueFloat64Ptr(values["LossWeight"]),
		SBP:            valueFloat64Ptr(values["SBP"]),
		DBP:            valueFloat64Ptr(values["DBP"]),
		HeartRate:      valueFloat64Ptr(values["HeartRate"]),
		Respiration:    valueFloat64Ptr(values["Respiration"]),
		Temperature:    valueFloat64Ptr(values["BodyTemp"]),
		RealIntake:     valueFloat64Ptr(values["RealIntake"]),
		PressurePoint:  valueString(values["PressurePoint"]),
		Complication:   valueString(values["Complication"]),
		Symptoms:       valueString(values["Symptoms"]),
		Notes:          valueString(values["Note"]),
		CreatorID:      valueInt64(values["CreatorId"]),
		CreateTime:     valueTimePtr(values["CreateTime"]),
		LastModifyTime: valueTimePtr(values["LastModifyTime"]),
	}
}

func valueFloat64Ptr(v any) *float64 {
	switch n := v.(type) {
	case *float64:
		return n
	case float64:
		return &n
	case *int:
		if n == nil {
			return nil
		}
		f := float64(*n)
		return &f
	case int:
		f := float64(n)
		return &f
	default:
		return nil
	}
}

func valueInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	default:
		return 0
	}
}

func valueString(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func valueTimePtr(v any) *time.Time {
	switch t := v.(type) {
	case time.Time:
		return &t
	case *time.Time:
		return t
	default:
		return nil
	}
}
func (s *TreatmentService) SubmitPostAssessment(treatmentID int64, req TreatmentAfterSignsRequest, creatorID int64) (*TreatmentRealtimeResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		txService := &TreatmentService{db: tx}
		if _, err := txService.SaveAfterSigns(treatmentID, req, creatorID); err != nil {
			return err
		}
		if err := txService.UpdateStatus(treatmentID, models.TreatmentStatusCompleted); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.Get(treatmentID)
}

// TreatmentDisinfectionRequest 消毒登记请求
type TreatmentDisinfectionRequest struct {
	EquipmentID     *int64  `json:"equipmentId"`
	DisinfectUserID *int64  `json:"disinfectUserId"`
	DisinfectWay    *string `json:"disinfectWay"`
	Type            *string `json:"type"`
	Disinfectant    *string `json:"disinfectant"`
	StartTime       *string `json:"startTime"`
	EndTime         *string `json:"endTime"`
	Description     *string `json:"description"`
	Note            *string `json:"note"`
}

// TreatmentSummaryRequest 治疗小结保存请求
type TreatmentSummaryRequest struct {
	TreatmentSummary *string `json:"treatmentSummary"`
	NurseSummary     *string `json:"nurseSummary"`
	DoctorSummary    *string `json:"doctorSummary"`
}

func (s *TreatmentService) SaveDisinfection(treatmentID int64, req TreatmentDisinfectionRequest, creatorID int64) (*TreatmentRealtimeResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	now := time.Now()

	equipmentID := int64(0)
	if req.EquipmentID != nil {
		equipmentID = *req.EquipmentID
	}
	disinfectUserID := creatorID
	if req.DisinfectUserID != nil {
		disinfectUserID = *req.DisinfectUserID
	}

	var startTime *time.Time
	if req.StartTime != nil && *req.StartTime != "" {
		if t, err := parseTimeString(*req.StartTime); err == nil {
			startTime = &t
		}
	}
	if startTime == nil {
		startTime = &now
	}

	var endTime *time.Time
	if req.EndTime != nil && *req.EndTime != "" {
		if t, err := parseTimeString(*req.EndTime); err == nil {
			endTime = &t
		}
	}

	columns := map[string]interface{}{
		`"TenantId"`:        LegacyTenantID,
		`"TreatmentId"`:     treatmentID,
		`"EquipmentId"`:     equipmentID,
		`"DisinfectUserId"`: disinfectUserID,
		`"StartTime"`:       startTime,
		`"CreatorId"`:       creatorID,
		`"LastModifyTime"`:  now,
	}
	if req.DisinfectWay != nil {
		columns[`"DisinfectWay"`] = *req.DisinfectWay
	}
	if req.Type != nil {
		columns[`"Type"`] = *req.Type
	}
	if req.Disinfectant != nil {
		columns[`"Disinfectant"`] = *req.Disinfectant
	}
	if req.Description != nil {
		columns[`"Description"`] = *req.Description
	}
	if req.Note != nil {
		columns[`"Note"`] = *req.Note
	}
	if endTime != nil {
		columns[`"EndTime"`] = endTime
	}

	var existing struct {
		ID int64 `gorm:"column:Id"`
	}
	err := s.db.Table(`"Auxiliary_EquipmentDisinfection"`).
		Select(`"Id"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).
		Order(`"CreateTime" DESC`).
		Limit(1).
		First(&existing).Error
	if err == nil {
		res := s.db.Table(`"Auxiliary_EquipmentDisinfection"`).
			Where(`"Id" = ? AND "TenantId" = ?`, existing.ID, LegacyTenantID).
			Updates(columns)
		if res.Error != nil {
			return nil, fmt.Errorf("更新消毒登记失败: %w", res.Error)
		}
		return s.Get(treatmentID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询消毒登记失败: %w", err)
	}

	newID, err := nextLegacyID()
	if err != nil {
		return nil, err
	}
	columns[`"Id"`] = newID
	columns[`"CreateTime"`] = now

	result := s.db.Table(`"Auxiliary_EquipmentDisinfection"`).Create(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("保存消毒登记失败: %w", result.Error)
	}

	return s.Get(treatmentID)
}

func (s *TreatmentService) SaveSummary(treatmentID int64, req TreatmentSummaryRequest, creatorID int64) (*TreatmentRealtimeResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	updates := map[string]interface{}{}
	if req.TreatmentSummary != nil {
		updates[`"TreatmentSummary"`] = *req.TreatmentSummary
	}
	if req.NurseSummary != nil {
		updates[`"NurseSummary"`] = *req.NurseSummary
	}
	if req.DoctorSummary != nil {
		updates[`"TreatmentSummary"`] = *req.DoctorSummary
	}

	if len(updates) == 0 {
		return s.Get(treatmentID)
	}

	result := s.db.Table(`"Treatment_Treatment"`).Where(`"Id" = ? AND "TenantId" = ?`, treatmentID, LegacyTenantID).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("保存治疗小结失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("treatment not found")
	}

	return s.Get(treatmentID)
}

func (s *TreatmentService) syncScheduleStatus(treatmentID int64, targetStatus int16) {
	var ref struct {
		ScheduleID    int64      `gorm:"column:schedule_id"`
		PatientID     int64      `gorm:"column:patient_id"`
		StartTime     *time.Time `gorm:"column:StartTime"`
		SignInTime    *time.Time `gorm:"column:SignInTime"`
		ReceptionTime *time.Time `gorm:"column:ReceptionTime"`
		CreateTime    *time.Time `gorm:"column:CreateTime"`
	}
	if err := s.db.Table(`"Treatment_Treatment"`).
		Select(`COALESCE("ScheduleId", 0) AS schedule_id, "PatientId" AS patient_id, "StartTime", "SignInTime", "ReceptionTime", "CreateTime"`).
		Where(`"Id" = ?`, treatmentID).
		Scan(&ref).Error; err != nil {
		return
	}
	refTime := ref.StartTime
	if refTime == nil {
		refTime = ref.SignInTime
	}
	if refTime == nil {
		refTime = ref.ReceptionTime
	}
	if refTime == nil {
		refTime = ref.CreateTime
	}

	// 路径一：治疗直接挂了 ScheduleId（如从排班格上机创建），按主键精确流转。
	if ref.ScheduleID != 0 {
		res := s.db.Table(`"Schedule_PatientShift"`).
			Where(`"Id" = ? AND "TenantId" = ?`, ref.ScheduleID, LegacyTenantID).
			Update("Status", targetStatus)
		if res.Error != nil {
			log.Printf("[treatment] syncScheduleStatus failed: scheduleId=%d target=%d err=%v", ref.ScheduleID, targetStatus, res.Error)
		}
		return
	}

	// 路径二：工作流上机创建的治疗不带 ScheduleId，按 患者+当日 回退关联到计划排班。
	// 生产库存在同患者同日多条 Status IN (20,50) 的历史脏排班且无法靠 ShiftId/WardId/MachineId 区分，
	// 故只更新唯一候选；不止一条则跳过并记日志，由人工或后续修正，绝不批量误改。
	if ref.PatientID == 0 || refTime == nil {
		return
	}
	type candidateRow struct {
		ID int64 `gorm:"column:Id"`
	}
	var candidates []candidateRow
	_ = s.db.Table(`"Schedule_PatientShift"`).
		Select(`"Id"`).
		Where(`"TenantId" = ? AND "PatientId" = ? AND DATE("TreatmentTime") = DATE(?) AND "Status" IN ?`,
			LegacyTenantID, ref.PatientID, *refTime, []int16{20, 50}).
		Scan(&candidates)
	if len(candidates) != 1 {
		log.Printf("[treatment] syncScheduleStatus skipped: patientId=%d date=%v candidates=%d (need exactly 1)",
			ref.PatientID, *refTime, len(candidates))
		return
	}
	shiftID := candidates[0].ID
	res := s.db.Table(`"Schedule_PatientShift"`).
		Where(`"Id" = ? AND "TenantId" = ?`, shiftID, LegacyTenantID).
		Update("Status", targetStatus)
	if res.Error != nil {
		log.Printf("[treatment] syncScheduleStatus failed: shiftId=%d target=%d err=%v", shiftID, targetStatus, res.Error)
	}
}

func parseTimeString(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339,
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Now(), fmt.Errorf("无法解析时间字符串: %s", s)
}
