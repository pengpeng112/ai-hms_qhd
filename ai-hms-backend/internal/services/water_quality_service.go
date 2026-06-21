package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type WaterQualityService struct {
	db         *gorm.DB
	tenantID   int64
	thresholds map[string]config.WaterQualityThreshold
}

func NewWaterQualityService() *WaterQualityService {
	th, _ := config.LoadWaterQualityThresholds()
	if th == nil {
		th = map[string]config.WaterQualityThreshold{}
	}
	return &WaterQualityService{db: database.GetDB(), tenantID: LegacyTenantID, thresholds: th}
}

type RecordInput struct {
	TestDate    time.Time `json:"testDate"`
	TestType    string    `json:"testType"`
	SamplePoint string    `json:"samplePoint"`
	DeviceID    string    `json:"deviceId"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
}

// judgeWaterQuality judges pass/fail/pending against threshold + returns limit snapshot string.
func judgeWaterQuality(th config.WaterQualityThreshold, value float64) (result, limitSnap string) {
	switch th.LimitType {
	case "max":
		if th.Max == nil {
			return models.WQResultPending, ""
		}
		limitSnap = fmt.Sprintf("≤%g", *th.Max)
		if value > *th.Max {
			return models.WQResultFail, limitSnap
		}
		return models.WQResultPass, limitSnap
	case "range":
		if th.Min == nil || th.Max == nil {
			return models.WQResultPending, ""
		}
		limitSnap = fmt.Sprintf("%g–%g", *th.Min, *th.Max)
		if value < *th.Min || value > *th.Max {
			return models.WQResultFail, limitSnap
		}
		return models.WQResultPass, limitSnap
	}
	return models.WQResultPending, ""
}

func (s *WaterQualityService) Record(in RecordInput) (*models.WaterQuality, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if in.TestType == "conductivity" {
		// 电导为机器自动引用(ConductivityDaily)，read-only，绝不落 water_quality（规则§二"引用不重复采"）
		return nil, errors.New("电导为机器自动引用，无需手工录入")
	}
	th, ok := s.thresholds[in.TestType]
	if !ok {
		return nil, errors.New("未知检测项目")
	}
	if !th.Enabled {
		return nil, errors.New("该检测项目未启用（置灰待补阈值）")
	}
	result, limitSnap := judgeWaterQuality(th, in.Value)
	testDate := in.TestDate
	rec := &models.WaterQuality{
		ID:            utils.GenerateID(),
		TenantID:      s.tenantID,
		TestDate:      &testDate,
		TestType:      in.TestType,
		SamplePoint:   strings.TrimSpace(in.SamplePoint),
		DeviceID:      strings.TrimSpace(in.DeviceID),
		Value:         in.Value,
		Unit:          in.Unit,
		StandardLimit: limitSnap,
		Result:        result,
		Source:        "manual",
	}
	if th.FrequencyDays > 0 {
		due := testDate.AddDate(0, 0, th.FrequencyDays)
		rec.NextDueDate = &due
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

type HandleInput struct {
	Role       string `json:"role"` // engineer / head_nurse
	SignerID   string `json:"signerId"`
	SignerName string `json:"signerName"`
	Action     string `json:"action"`
}

func (s *WaterQualityService) Handle(id string, in HandleInput) (*models.WaterQuality, error) {
	if in.Role != "engineer" && in.Role != "head_nurse" {
		return nil, errors.New("非法处置角色")
	}
	var rec models.WaterQuality
	if err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&rec).Error; err != nil {
		return nil, errors.New("检测记录不存在")
	}
	if rec.Result != models.WQResultFail {
		return nil, errors.New("仅超标记录需双确认")
	}
	if rec.HandledAt != nil {
		return &rec, nil // 幂等
	}
	if in.Role == "engineer" {
		rec.HandledEngineerID = in.SignerID
	} else {
		rec.HandledHeadnurseID = in.SignerID
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		ss := &SignService{db: tx}
		if _, e := ss.Sign(s.tenantID, models.SignTargetWaterQualityHandling, rec.ID, in.SignerID, in.SignerName); e != nil {
			return e
		}
		updates := map[string]any{
			"handled_engineer_id":  rec.HandledEngineerID,
			"handled_headnurse_id": rec.HandledHeadnurseID,
			"updated_at":           time.Now(),
		}
		if strings.TrimSpace(in.Action) != "" {
			updates["action"] = strings.TrimSpace(in.Action)
		}
		if rec.HandledEngineerID != "" && rec.HandledHeadnurseID != "" {
			now := time.Now()
			updates["handled_at"] = &now
		}
		return tx.Model(&models.WaterQuality{}).Where("id = ?", rec.ID).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	var out models.WaterQuality
	s.db.First(&out, "id = ?", rec.ID)
	return &out, nil
}

type WaterQualityAlerts struct {
	Exceed []models.WaterQuality `json:"exceed"` // 超标未处置
	Due    []models.WaterQuality `json:"due"`    // 到期/将到期
}

func (s *WaterQualityService) Alerts() (*WaterQualityAlerts, error) {
	res := &WaterQualityAlerts{}
	if err := s.db.Where("tenant_id = ? AND result = ? AND handled_at IS NULL", s.tenantID, models.WQResultFail).
		Order("test_date DESC").Find(&res.Exceed).Error; err != nil {
		return nil, err
	}
	cutoff := time.Now().AddDate(0, 0, 14)
	if err := s.db.Where("tenant_id = ? AND next_due_date IS NOT NULL AND next_due_date < ?", s.tenantID, cutoff).
		Order("next_due_date ASC").Find(&res.Due).Error; err != nil {
		return nil, err
	}
	return res, nil
}

type WqListFilter struct {
	TestType    string
	SamplePoint string
}

func (s *WaterQualityService) List(f WqListFilter) ([]models.WaterQuality, error) {
	q := s.db.Where("tenant_id = ?", s.tenantID)
	if f.TestType != "" {
		q = q.Where("test_type = ?", f.TestType)
	}
	if f.SamplePoint != "" {
		q = q.Where("sample_point = ?", f.SamplePoint)
	}
	var rows []models.WaterQuality
	err := q.Order("test_date DESC, created_at DESC").Find(&rows).Error
	return rows, err
}

type ConductivityPoint struct {
	Day     string  `json:"day"`
	Value   float64 `json:"value"`
	InRange bool    `json:"inRange"`
}

// ConductivityDaily 引用实时监控：按日聚合透析机电导(Treatment_DuringParam.Conductivity)，
// 据 conductivity 配置 range 判定。read-only，不落 water_quality（规则§二"引用不重复采"）。
func (s *WaterQualityService) ConductivityDaily(days int) ([]ConductivityPoint, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	since := time.Now().AddDate(0, 0, -days)
	var rows []struct {
		Day string  `gorm:"column:day"`
		Avg float64 `gorm:"column:avg_cond"`
	}
	// 与全库一致：老库 StartTime 常 NULL，按 COALESCE(StartTime, CreateTime) 取治疗日（见 dashboard/monitoring 服务同款）
	err := s.db.Table(`"Treatment_DuringParam" AS dp`).
		Select(`date(COALESCE(t."StartTime", t."CreateTime")) AS day, AVG(dp."Conductivity") AS avg_cond`).
		Joins(`JOIN "Treatment_Treatment" t ON t."Id" = dp."TreatmentId"`).
		Where(`t."TenantId" = ? AND COALESCE(t."StartTime", t."CreateTime") >= ?`, s.tenantID, since).
		Group(`date(COALESCE(t."StartTime", t."CreateTime"))`).
		Order(`day`).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	th, ok := s.thresholds["conductivity"]
	pts := make([]ConductivityPoint, 0, len(rows))
	for _, r := range rows {
		inRange := false
		if ok && th.Min != nil && th.Max != nil {
			inRange = r.Avg >= *th.Min && r.Avg <= *th.Max
		}
		pts = append(pts, ConductivityPoint{Day: r.Day, Value: r.Avg, InRange: inRange})
	}
	return pts, nil
}
