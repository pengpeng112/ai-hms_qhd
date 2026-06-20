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
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type InfectiousService struct {
	db       *gorm.DB
	tenantID int64
}

func NewInfectiousService() *InfectiousService {
	return &InfectiousService{db: database.GetDB(), tenantID: LegacyTenantID}
}

type ScreenItem struct {
	Item   string `json:"item"`
	Result string `json:"result"` // negative/positive/indeterminate
}

type ScreenInput struct {
	ScreenDate time.Time    `json:"screenDate"`
	Source     string       `json:"source"`
	Items      []ScreenItem `json:"items"`
	Note       string       `json:"note"`
}

func (s *InfectiousService) Screen(patientID int64, in ScreenInput) (*models.PatientInfectious, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if len(in.Items) == 0 {
		return nil, errors.New("筛查项目不能为空")
	}
	overall := models.InfectiousNegative
	var positives []string
	hasIndeterminate := false
	for _, it := range in.Items {
		switch it.Result {
		case models.InfItemPositive:
			positives = append(positives, it.Item)
		case models.InfItemIndeterminate:
			hasIndeterminate = true
		}
	}
	if len(positives) > 0 {
		overall = models.InfectiousPositive
	} else if hasIndeterminate {
		overall = models.InfectiousPending
	}
	itemsJSON, _ := json.Marshal(in.Items)
	screenDate := in.ScreenDate
	due := screenDate.AddDate(0, 6, 0)
	source := in.Source
	if source == "" {
		source = "manual"
	}
	rec := &models.PatientInfectious{
		ID:              utils.GenerateID(),
		TenantID:        s.tenantID,
		PatientID:       strconv.FormatInt(patientID, 10),
		ScreenDate:      &screenDate,
		Items:           string(itemsJSON),
		Source:          source,
		ResultOverall:   overall,
		PositiveMarkers: strings.Join(positives, ","),
		NextDueDate:     &due,
		ZoneTag:         "normal",
		Note:            in.Note,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

// latest 取该患者最新一条筛查记录（无则 nil,nil）。
// 内置 recover：若 DB 驱动 panic（如表缺失），转为 error 返回，由调用方 fail-open。
func (s *InfectiousService) latest(patientID int64) (rec *models.PatientInfectious, err error) {
	defer func() {
		if r := recover(); r != nil {
			rec = nil
			err = fmt.Errorf("infectious latest panic: %v", r)
		}
	}()
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var row models.PatientInfectious
	e := s.db.Where("patient_id = ?", strconv.FormatInt(patientID, 10)).
		Order("screen_date DESC, created_at DESC").First(&row).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	return &row, nil
}

type GateState string

const (
	GateAllowNormal  GateState = "ALLOW_NORMAL"
	GateRequireCZone GateState = "REQUIRE_C_ZONE"
	GateFrozen       GateState = "FROZEN"
	GateCZoneCRRT    GateState = "C_ZONE_CRRT"
)

type GateResult struct {
	State  GateState `json:"state"`
	Reason string    `json:"reason"`
}

// CanScheduleRoutine 四态门禁。基础设施错误(表缺/查询失败)→fail-open ALLOW_NORMAL+日志；业务判定→fail-closed。
func (s *InfectiousService) CanScheduleRoutine(patientID int64) GateResult {
	rec, err := s.latest(patientID)
	if err != nil {
		log.Printf("[infectious-gate] fail-open: latest(%d) error: %v", patientID, err)
		return GateResult{State: GateAllowNormal}
	}
	today := time.Now()
	if rec == nil {
		return GateResult{State: GateRequireCZone, Reason: "入院前未完成传染病筛查，不得安排常规透析（可 C 区全警戒）"}
	}
	if rec.ResultOverall == models.InfectiousPositive {
		if rec.HandledAt == nil {
			return GateResult{State: GateFrozen, Reason: "阳性未完成双人处置，排班已冻结，须医生+护士长双处理"}
		}
		if rec.Disposition == models.InfectiousDispCZoneCRRT {
			return GateResult{State: GateCZoneCRRT, Reason: "阳性患者仅可 C 区全警戒 + CRRT 机器"}
		}
		if rec.Disposition == models.InfectiousDispTransferOut {
			return GateResult{State: GateFrozen, Reason: "阳性患者已转出退册，不再排班"}
		}
		return GateResult{State: GateFrozen, Reason: "阳性患者处置状态异常，已冻结"}
	}
	if rec.NextDueDate != nil && rec.NextDueDate.Before(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())) {
		return GateResult{State: GateRequireCZone, Reason: "传染病筛查已过期(>6月)，请复查后再排常规（可 C 区全警戒）"}
	}
	if rec.ResultOverall == models.InfectiousNegative {
		return GateResult{State: GateAllowNormal}
	}
	// pending/未出结果 等 → 结果未明，fail-closed 仅可 C 区
	return GateResult{State: GateRequireCZone, Reason: "传染病筛查结果未明（待定/待复核），暂不得上常规，可 C 区全警戒"}
}
