package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type DryWeightService struct {
	db       *gorm.DB
	tenantID int64
}

func NewDryWeightService() *DryWeightService {
	return &DryWeightService{db: database.GetDB(), tenantID: LegacyTenantID}
}

type DwAssessInput struct {
	AssessType   string   `json:"assessType"`
	Phase        string   `json:"phase"`
	SBP          *int     `json:"sbp"`
	DBP          *int     `json:"dbp"`
	HeartRate    *int     `json:"heartRate"`
	Edema        bool     `json:"edema"`
	Palpitation  bool     `json:"palpitation"`
	HeartFailure bool     `json:"heartFailure"`
	Cramp        bool     `json:"cramp"`
	CTR          *float64 `json:"ctr"`
	ACTR         *float64 `json:"actr"`
	BIAOH        *float64 `json:"biaOh"`
	BIATBW       *float64 `json:"biaTbw"`
	BIAECW       *float64 `json:"biaEcw"`
	PostWeight   *float64 `json:"postWeight"`
	TargetWeight *float64 `json:"targetWeight"`
	Decision     string   `json:"decision"`
	AdjustKg     *float64 `json:"adjustKg"`
	RNaSetting   *float64 `json:"rnaSetting"`
	AssessorID   string   `json:"assessorId"`
	AssessorName string   `json:"assessorName"`
}

func (s *DryWeightService) Assess(patientID int64, in DwAssessInput) (*models.DryWeightAssessment, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if in.Phase != models.DWPhaseInduction && in.Phase != models.DWPhaseMaintenance {
		return nil, errors.New("阶段必须为 induction 或 maintenance")
	}
	if in.AssessType != models.DWAssessDaily && in.AssessType != models.DWAssessCycle {
		return nil, errors.New("评估类型必须为 daily 或 cycle")
	}

	// 主判据判定
	mainMet := true
	failed := make([]string, 0)

	// 血压 ≥110/60
	sbpOk := in.SBP != nil && *in.SBP >= config.DwConfig.SBPCriteria
	dbpOk := in.DBP != nil && *in.DBP >= config.DwConfig.DBPCriteria
	if !sbpOk || !dbpOk {
		mainMet = false
		failed = append(failed, "血压低于标准")
	}

	// 心率 60-100
	if in.HeartRate != nil {
		if *in.HeartRate < config.DwConfig.HeartRateLow || *in.HeartRate > config.DwConfig.HeartRateHigh {
			mainMet = false
			failed = append(failed, "心率异常")
		}
	} else {
		mainMet = false
		failed = append(failed, "缺心率")
	}

	// 症状
	if in.Edema {
		mainMet = false
		failed = append(failed, "显性水肿")
	}
	if in.Palpitation {
		mainMet = false
		failed = append(failed, "心慌气短")
	}
	if in.HeartFailure {
		mainMet = false
		failed = append(failed, "心衰")
	}
	if in.Cramp {
		mainMet = false
		failed = append(failed, "肌肉痉挛")
	}

	// 阶段限幅校验
	maxAdj := config.DwConfig.MaintenanceMaxAdjustKg
	if in.Phase == models.DWPhaseInduction {
		maxAdj = config.DwConfig.InductionMaxAdjustKg
	}
	if in.AdjustKg != nil && math.Abs(*in.AdjustKg) > maxAdj {
		return nil, fmt.Errorf("调整幅度 %.2f kg 超过 %s 上限 %.1f kg", *in.AdjustKg, in.Phase, maxAdj)
	}

	failedJSON, _ := json.Marshal(failed)
	dwa := &models.DryWeightAssessment{
		ID: utils.GenerateID(), TenantID: s.tenantID, PatientID: patientID,
		AssessType: in.AssessType, Phase: in.Phase,
		SBP: in.SBP, DBP: in.DBP, HeartRate: in.HeartRate,
		Edema: in.Edema, Palpitation: in.Palpitation, HeartFailure: in.HeartFailure, Cramp: in.Cramp,
		CTR: in.CTR, ACTR: in.ACTR, BIAOH: in.BIAOH, BIATBW: in.BIATBW, BIAECW: in.BIAECW,
		PostWeight: in.PostWeight, TargetWeight: in.TargetWeight,
		Decision: in.Decision, AdjustKg: in.AdjustKg, RNaSetting: in.RNaSetting,
		MainMet: mainMet, FailedReasons: string(failedJSON),
		AssessorID: in.AssessorID, AssessorName: in.AssessorName,
	}
	if err := s.db.Create(dwa).Error; err != nil {
		return nil, err
	}
	return dwa, nil
}

func (s *DryWeightService) ListAssessments(patientID int64) ([]models.DryWeightAssessment, error) {
	var rows []models.DryWeightAssessment
	if err := s.db.Where("tenant_id = ? AND patient_id = ?", s.tenantID, patientID).
		Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type DwConfirmInput struct {
	DryWeight    float64  `json:"dryWeight"`
	Phase        string   `json:"phase"`
	ACTR         *float64 `json:"actr"`
	CTR          *float64 `json:"ctr"`
	ConfirmedBy  string   `json:"confirmedBy"`
	ConfirmedName string  `json:"confirmedName"`
}

type DwConfirmResult struct {
	*models.PatientDryWeight
	LegacyPlanUpdated bool `json:"legacyPlanUpdated"`
}

func (s *DryWeightService) Confirm(patientID int64, in DwConfirmInput) (*DwConfirmResult, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if strings.TrimSpace(in.ConfirmedBy) == "" {
		return nil, errors.New("确定人不能为空")
	}
	if in.DryWeight <= 0 {
		return nil, errors.New("干体重必须大于0")
	}
	if in.Phase != models.DWPhaseInduction && in.Phase != models.DWPhaseMaintenance {
		return nil, errors.New("阶段非法")
	}

	now := time.Now()
	var existing models.PatientDryWeight
	err := s.db.Where("tenant_id = ? AND patient_id = ?", s.tenantID, patientID).First(&existing).Error
	pdw := &models.PatientDryWeight{
		TenantID: s.tenantID, PatientID: patientID,
		DryWeight: in.DryWeight, StandardACTR: in.ACTR, StandardCTR: in.CTR,
		Phase: in.Phase, ConfirmedBy: in.ConfirmedBy, ConfirmedName: in.ConfirmedName, ConfirmedAt: now,
	}
	var result *models.PatientDryWeight
	if errors.Is(err, gorm.ErrRecordNotFound) {
		pdw.ID = utils.GenerateID()
		if e := s.db.Create(pdw).Error; e != nil {
			return nil, e
		}
		result = pdw
	} else if err != nil {
		return nil, err
	} else {
		updates := map[string]any{
			"dry_weight":    in.DryWeight,
			"standard_actr": in.ACTR,
			"standard_ctr":  in.CTR,
			"phase":         in.Phase,
			"confirmed_by":  in.ConfirmedBy,
			"confirmed_name": in.ConfirmedName,
			"confirmed_at":  now,
			"updated_at":    now,
		}
		if e := s.db.Model(&existing).Updates(updates).Error; e != nil {
			return nil, e
		}
		existing.DryWeight = in.DryWeight
		existing.StandardACTR = in.ACTR
		existing.StandardCTR = in.CTR
		existing.Phase = in.Phase
		result = &existing
	}

	// 写回老库 Plan_PatientPlan.DryWeight
	legacyUpdated := false
	result2 := s.db.Table(`"Plan_PatientPlan"`).
		Where(`"PatientId" = ? AND "TenantId" = ? AND "IsDisabled" = false`, patientID, s.tenantID).
		Update("DryWeight", in.DryWeight)
	if result2.Error == nil && result2.RowsAffected > 0 {
		legacyUpdated = true
	}

	return &DwConfirmResult{PatientDryWeight: result, LegacyPlanUpdated: legacyUpdated}, nil
}

type DwCurrentData struct {
	DryWeight    *float64 `json:"dryWeight"`
	StandardACTR *float64 `json:"standardActr"`
	StandardCTR  *float64 `json:"standardCtr"`
	Phase        string   `json:"phase"`
	SuggestedRNa float64  `json:"suggestedRNa"`
	ConfirmedAt  *string  `json:"confirmedAt"`
}

func (s *DryWeightService) Current(patientID int64) (*DwCurrentData, error) {
	var pdw models.PatientDryWeight
	err := s.db.Where("tenant_id = ? AND patient_id = ?", s.tenantID, patientID).First(&pdw).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &DwCurrentData{
			Phase:        models.DWPhaseInduction,
			SuggestedRNa: config.DwConfig.InductionRNa,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	rna := config.DwConfig.MaintenanceRNa
	if pdw.Phase == models.DWPhaseInduction {
		rna = config.DwConfig.InductionRNa
	}
	at := pdw.ConfirmedAt.Format(time.RFC3339)
	dw := pdw.DryWeight
	return &DwCurrentData{
		DryWeight: &dw, StandardACTR: pdw.StandardACTR, StandardCTR: pdw.StandardCTR,
		Phase: pdw.Phase, SuggestedRNa: rna, ConfirmedAt: &at,
	}, nil
}

func (s *DryWeightService) fetchLatestACTR(patientID int64) *float64 {
	var actr models.PatientACTR
	err := s.db.Where("tenant_id = ? AND patient_id = ?", s.tenantID, strconv.FormatInt(patientID, 10)).
		Order("created_at DESC").First(&actr).Error
	if err != nil {
		return nil
	}
	if actr.DoctorCorrection != nil {
		return actr.DoctorCorrection
	}
	return actr.ACTR
}
