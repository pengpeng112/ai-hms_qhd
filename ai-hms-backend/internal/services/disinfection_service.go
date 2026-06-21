package services

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type DisinfectionService struct {
	db       *gorm.DB
	tenantID int64
}

func NewDisinfectionService() *DisinfectionService {
	return &DisinfectionService{db: database.GetDB(), tenantID: LegacyTenantID}
}

var validDisinfectTypes = map[string]struct{}{
	models.DisinfectTypeHeat: {}, models.DisinfectTypeTerminal: {},
	models.DisinfectTypeDecalc: {}, models.DisinfectTypeEnhanced: {},
}

type DisinfectRecordInput struct {
	DeviceID      int64     `json:"deviceId"`
	DisinfectType string    `json:"disinfectType"`
	Disinfectant  string    `json:"disinfectant"`
	Concentration string    `json:"concentration"`
	OperatorID    int64     `json:"operatorId"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	TreatmentID   int64     `json:"treatmentId"`
	ResidualCheck string    `json:"residualCheck"`
	Result        string    `json:"result"`
	DocRef        string    `json:"docRef"`
	Source        string    `json:"source"`
}

type RecordResult struct {
	DisinfectionID int64  `json:"disinfectionId"`
	ComplianceID   string `json:"complianceId"`
}

func (s *DisinfectionService) Record(in DisinfectRecordInput) (*RecordResult, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if _, ok := validDisinfectTypes[in.DisinfectType]; !ok {
		return nil, errors.New("未知消毒类型")
	}
	if in.DeviceID == 0 {
		return nil, errors.New("缺少设备")
	}
	source := in.Source
	if source == "" {
		source = "manual"
	}
	var out RecordResult
	err := s.db.Transaction(func(tx *gorm.DB) error {
		baseID, idErr := nextLegacyID()
		if idErr != nil {
			return idErr
		}
		now := time.Now()
		base := map[string]any{
			"Id": int64(baseID), "TenantId": s.tenantID, "EquipmentId": in.DeviceID,
			"DisinfectUserId": in.OperatorID, "Disinfectant": in.Disinfectant, "Type": in.DisinfectType,
			"StartTime": in.StartTime, "EndTime": in.EndTime, "TreatmentId": in.TreatmentID,
			"Status": 1, "CreatorId": in.OperatorID, "CreateTime": now, "LastModifyTime": now,
		}
		if e := tx.Table(`"Auxiliary_EquipmentDisinfection"`).Create(base).Error; e != nil {
			return e
		}
		comp := &models.DisinfectionCompliance{
			ID: utils.GenerateID(), TenantID: s.tenantID, DisinfectionID: int64(baseID), DeviceID: in.DeviceID,
			Concentration: strings.TrimSpace(in.Concentration), ResidualCheck: in.ResidualCheck,
			Result: in.Result, Source: source, DocRef: strings.TrimSpace(in.DocRef),
		}
		if e := tx.Create(comp).Error; e != nil {
			return e
		}
		out = RecordResult{DisinfectionID: int64(baseID), ComplianceID: comp.ID}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type ComplianceInput struct {
	Concentration string `json:"concentration"`
	ResidualCheck string `json:"residualCheck"`
	Result        string `json:"result"`
	DocRef        string `json:"docRef"`
}

// ---------- Machine status (safety-critical) ----------

const (
	DisinfMachineOK      = "OK"
	DisinfMachineWarn    = "WARN"
	DisinfMachineBlocked = "BLOCKED_RESIDUAL"
)

type MachineDisinfStatus struct {
	DeviceID      int64      `json:"deviceId"`
	State         string     `json:"state"`
	LastTerminal  *time.Time `json:"lastTerminal"`
	LastDecalc    *time.Time `json:"lastDecalc"`
	LastHeat      *time.Time `json:"lastHeat"`
	TerminalToday bool       `json:"terminalToday"`
	DecalcOverdue bool       `json:"decalcOverdue"`
	HeatLag       bool       `json:"heatLag"`
	ResidualFail  bool       `json:"residualFail"`
	Reasons       []string   `json:"reasons"`
}

func (s *DisinfectionService) lastDisinfect(deviceID int64, dtype string) *time.Time {
	// Scan into *string then parse: SQLite MAX() returns text which cannot be
	// scanned directly into *time.Time by the pure-Go driver.  This approach
	// works portably across SQLite (tests) and PostgreSQL (production).
	var raw struct {
		T *string `gorm:"column:t"`
	}
	s.db.Table(`"Auxiliary_EquipmentDisinfection"`).
		Select(`MAX("StartTime") AS t`).
		Where(`"TenantId" = ? AND "EquipmentId" = ? AND "Type" = ?`, s.tenantID, deviceID, dtype).
		Scan(&raw)
	if raw.T == nil || *raw.T == "" {
		return nil
	}
	for _, layout := range []string{
		time.RFC3339Nano, time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, *raw.T); err == nil {
			return &t
		}
	}
	return nil
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func (s *DisinfectionService) MachineStatus(deviceID int64) MachineDisinfStatus {
	now := time.Now()
	st := MachineDisinfStatus{DeviceID: deviceID, State: DisinfMachineOK}
	st.LastTerminal = s.lastDisinfect(deviceID, models.DisinfectTypeTerminal)
	st.LastDecalc = s.lastDisinfect(deviceID, models.DisinfectTypeDecalc)
	st.LastHeat = s.lastDisinfect(deviceID, models.DisinfectTypeHeat)

	// Residual check: most recent disinfection for this device
	var resid struct {
		Residual string `gorm:"column:residual_check"`
	}
	// 取该机"最近一条"消毒的残留结果。disinfection_id 来自 nextLegacyID()，时间单调递增，
	// 故最大 disinfection_id ＝ 最近一条 —— 直接按它倒序，避免 JOIN 老库表（老库带引号标识符
	// 在 GORM/SQLite 与 PG 间的引用方式不一致，JOIN 易引入 PG 语法错/测试不一致）。
	if e := s.db.Table("disinfection_compliance").
		Select("residual_check").
		Where("tenant_id = ? AND device_id = ?", s.tenantID, deviceID).
		Order("disinfection_id DESC").Limit(1).Scan(&resid).Error; e != nil {
		log.Printf("[disinfection] 残留检测查询失败 device=%d: %v", deviceID, e)
	}
	if resid.Residual == models.DisinfectResultFail {
		st.ResidualFail = true
		st.State = DisinfMachineBlocked
		st.Reasons = append(st.Reasons, "残留检测不合格，该机停用")
		return st
	}

	st.TerminalToday = st.LastTerminal != nil && sameDay(*st.LastTerminal, now)
	st.DecalcOverdue = st.LastDecalc == nil || st.LastDecalc.Before(now.AddDate(0, 0, -7))
	st.HeatLag = s.treatmentsToday(deviceID) > s.heatToday(deviceID)

	if !st.TerminalToday {
		st.Reasons = append(st.Reasons, "今日终末消毒未做")
	}
	if st.DecalcOverdue {
		st.Reasons = append(st.Reasons, "除钙到期(>7天)")
	}
	if st.HeatLag {
		st.Reasons = append(st.Reasons, "热消毒滞后")
	}
	if len(st.Reasons) > 0 {
		st.State = DisinfMachineWarn
	}
	return st
}

func (s *DisinfectionService) heatToday(deviceID int64) int64 {
	var n int64
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	s.db.Table(`"Auxiliary_EquipmentDisinfection"`).
		Where(`"TenantId" = ? AND "EquipmentId" = ? AND "Type" = ? AND "StartTime" >= ?`, s.tenantID, deviceID, models.DisinfectTypeHeat, start).
		Count(&n)
	return n
}

// treatmentsToday 今日该机治疗数（治疗→床→机器绑定 Schedule_BedEquipmentRel）。
// 绑定/治疗表缺时返回 0（不 false-warn，符合 spec"缺则计 0 不崩"）。
func (s *DisinfectionService) treatmentsToday(deviceID int64) (result int64) {
	defer func() {
		if r := recover(); r != nil {
			result = 0
		}
	}()
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var n int64
	err := s.db.Table(`"Treatment_Treatment" AS t`).
		Joins(`JOIN "Schedule_BedEquipmentRel" r ON r."BedId" = t."BedId"`).
		Where(`t."TenantId" = ? AND r."EquipmentId" = ? AND COALESCE(t."StartTime", t."CreateTime") >= ?`, s.tenantID, deviceID, start).
		Count(&n).Error
	if err != nil {
		return 0
	}
	return n
}

// CanUseMachine is a convenience alias for MachineStatus.
func (s *DisinfectionService) CanUseMachine(deviceID int64) MachineDisinfStatus {
	return s.MachineStatus(deviceID)
}

type DisinfAlerts struct {
	Blocked []MachineDisinfStatus `json:"blocked"`
	Warn    []MachineDisinfStatus `json:"warn"`
}

// Alerts 对给定机器集合算三态，分桶 Blocked / Warn。deviceIDs 由调用方(设备服务)提供在用机列表。
func (s *DisinfectionService) Alerts(deviceIDs []int64) (*DisinfAlerts, error) {
	res := &DisinfAlerts{}
	for _, id := range deviceIDs {
		st := s.MachineStatus(id)
		switch st.State {
		case DisinfMachineBlocked:
			res.Blocked = append(res.Blocked, st)
		case DisinfMachineWarn:
			res.Warn = append(res.Warn, st)
		}
	}
	return res, nil
}

type DisinfStats struct {
	HeatToday      int64 `json:"heatToday"`
	TreatmentToday int64 `json:"treatmentToday"`
}

// Stats 执行率原料：今日热消毒数 / 今日治疗数（按机器集合汇总）。
func (s *DisinfectionService) Stats(deviceIDs []int64) DisinfStats {
	var st DisinfStats
	for _, id := range deviceIDs {
		st.HeatToday += s.heatToday(id)
		st.TreatmentToday += s.treatmentsToday(id)
	}
	return st
}

func (s *DisinfectionService) SaveCompliance(disinfectionID int64, in ComplianceInput) (*models.DisinfectionCompliance, error) {
	var comp models.DisinfectionCompliance
	err := s.db.Where("disinfection_id = ? AND tenant_id = ?", disinfectionID, s.tenantID).First(&comp).Error
	if err != nil {
		return nil, errors.New("消毒记录不存在")
	}
	updates := map[string]any{"updated_at": time.Now()}
	if strings.TrimSpace(in.Concentration) != "" {
		updates["concentration"] = strings.TrimSpace(in.Concentration)
	}
	if in.ResidualCheck != "" {
		updates["residual_check"] = in.ResidualCheck
	}
	if in.Result != "" {
		updates["result"] = in.Result
	}
	if strings.TrimSpace(in.DocRef) != "" {
		updates["doc_ref"] = strings.TrimSpace(in.DocRef)
	}
	if err := s.db.Model(&models.DisinfectionCompliance{}).Where("disinfection_id = ?", disinfectionID).Updates(updates).Error; err != nil {
		return nil, err
	}
	var out models.DisinfectionCompliance
	s.db.First(&out, "disinfection_id = ?", disinfectionID)
	return &out, nil
}
