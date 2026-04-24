package services

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"gorm.io/gorm"
)

type legacyCoreInfection struct {
	PatientID      modeltypes.LegacyID `gorm:"column:PatientId"`
	InfectionDesc  string              `gorm:"column:InfectionDesc"`
	OtherDesc      string              `gorm:"column:OtherDesc"`
	Note           string              `gorm:"column:Note"`
	LastModifyTime time.Time           `gorm:"column:LastModifyTime"`
}

func (legacyCoreInfection) TableName() string { return "Register_Infection" }

type legacyCorePlan struct {
	ID                    int64               `gorm:"column:Id"`
	TenantID              int64               `gorm:"column:TenantId"`
	PatientID             modeltypes.LegacyID `gorm:"column:PatientId"`
	Name                  string              `gorm:"column:Name"`
	CreateTime            time.Time           `gorm:"column:CreateTime"`
	LastModifyTime        time.Time           `gorm:"column:LastModifyTime"`
	OddWeekFrequency      int                 `gorm:"column:OddWeekFrequency"`
	EvenWeekFrequency     int                 `gorm:"column:EvenWeekFrequency"`
	DialysisMethod        string              `gorm:"column:DialysisMethod"`
	DialysisDuration      int                 `gorm:"column:DialysisDuration"`
	DryWeight             float64             `gorm:"column:DryWeight"`
	BF                    int                 `gorm:"column:BF"`
	FirstAnticoagulant    int64               `gorm:"column:FirstAnticoagulant"`
	MaintainAnticoagulant int64               `gorm:"column:MaintainAnticoagulant"`
	IsDisabled            bool                `gorm:"column:IsDisabled"`
	Note                  string              `gorm:"column:Note"`
}

func (legacyCorePlan) TableName() string { return "Plan_PatientPlan" }

type legacyCoreOrder struct {
	ID             int64               `gorm:"column:Id"`
	TenantID       int64               `gorm:"column:TenantId"`
	PatientID      modeltypes.LegacyID `gorm:"column:PatientId"`
	Type           string              `gorm:"column:Type"`
	Classification string              `gorm:"column:Classification"`
	Content        string              `gorm:"column:Content"`
	Dosage         string              `gorm:"column:Dosage"`
	UseMethod      string              `gorm:"column:UseMethod"`
	UseWay         string              `gorm:"column:UseWay"`
	Note           string              `gorm:"column:Note"`
	OperatorID     int64               `gorm:"column:OperatorId"`
	StartTime      time.Time           `gorm:"column:StartTime"`
	EndTime        *time.Time          `gorm:"column:EndTime"`
	IsDisabled     bool                `gorm:"column:IsDisabled"`
	CreateTime     time.Time           `gorm:"column:CreateTime"`
	LastModifyTime time.Time           `gorm:"column:LastModifyTime"`
}

func (legacyCoreOrder) TableName() string { return "Order_PatientOrder" }

type legacyCoreLabRow struct {
	ItemCode       string     `gorm:"column:item_code"`
	ItemName       string     `gorm:"column:item_name"`
	ResultValue    string     `gorm:"column:result_value"`
	Unit           string     `gorm:"column:unit"`
	ReferenceRange string     `gorm:"column:reference_range"`
	ResultSign     string     `gorm:"column:result_sign"`
	TestedAt       *time.Time `gorm:"column:tested_at"`
}

// PatientCoreService 患者核心信息服务
type PatientCoreService struct {
	db *gorm.DB
}

// NewPatientCoreService 创建患者核心信息服务
func NewPatientCoreService() *PatientCoreService {
	return &PatientCoreService{
		db: database.GetDB(),
	}
}

// GetCore 获取患者核心信息聚合数据
func (s *PatientCoreService) GetCore(patientID string) (*PatientCoreResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	legacyPatientID, err := parseLegacyID(patientID)
	if err != nil {
		return nil, errors.New("invalid patient id")
	}

	var patient models.Patient
	err = s.db.First(&patient, `"Id" = ?`, legacyPatientID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}

	// 并行查询各个部分
	header := s.buildHeader(patient)
	overview, err := s.buildOverview(patient)
	if err != nil {
		return nil, err
	}
	clinicalFocus := s.buildClinicalFocus(patient)
	navigation := &PatientCoreNavigation{Total: 1, CurrentIndex: 0}
	if nav, navErr := s.buildNavigation(legacyPatientID); navErr == nil && nav != nil {
		navigation = nav
	}

	return &PatientCoreResponse{
		Header:        header,
		Overview:      *overview,
		ClinicalFocus: clinicalFocus,
		Navigation:    navigation,
	}, nil
}

// buildHeader 构建页面头部信息
func (s *PatientCoreService) buildHeader(patient models.Patient) PatientCoreHeader {
	// 计算透析龄
	dialysisAge := s.calculateDialysisAge(patient.CreatedAt)

	// 转换状态：active -> 治疗中
	status := "待诊"
	if patient.Status == "active" {
		// 检查是否有当前进行中的治疗会话
		var currentTreatment models.Treatment
		err := s.db.Where("patient_id = ? AND status = ? AND treatment_date = ?",
			patient.ID, models.TreatmentStatusInProgress, time.Now().Format("2006-01-02")).
			First(&currentTreatment).Error
		if err == nil {
			status = "透析中"
		} else {
			status = "治疗中"
		}
	}

	return PatientCoreHeader{
		ID:            legacyIDString(patient.ID),
		Name:          patient.Name,
		Avatar:        normalizeLegacyPatientAvatar(patient.ImageBase64String),
		Gender:        patient.Gender,
		Age:           patient.Age,
		BedNumber:     patient.BedNumber,
		PatientType:   s.stringOrDefault(patient.PatientType, "门诊"),
		InsuranceType: s.stringOrDefault(patient.InsuranceType, "自费"),
		DoctorName:    s.stringOrDefault(patient.DoctorName, ""),
		RiskLevel:     s.stringOrDefault(patient.RiskLevel, "低危"),
		Status:        status,
		DialysisAge:   dialysisAge,
	}
}

func normalizeLegacyPatientAvatar(raw string) string {
	payload := strings.TrimSpace(raw)
	if payload == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(payload), "data:image/") {
		return payload
	}

	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(payload)
		if err != nil {
			return ""
		}
	}

	mime := detectImageMIME(data)
	if mime == "" {
		mime = "image/jpeg"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, payload)
}

func detectImageMIME(data []byte) string {
	switch {
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "image/png"
	case len(data) >= 3 && bytes.Equal(data[:3], []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg"
	case len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a"):
		return "image/gif"
	case len(data) >= 2 && bytes.Equal(data[:2], []byte{0x42, 0x4D}):
		return "image/bmp"
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "image/webp"
	default:
		return ""
	}
}

// buildOverview 构建 Overview Tab 数据
func (s *PatientCoreService) buildOverview(patient models.Patient) (*PatientCoreOverview, error) {
	infection, err := s.buildInfection(patient.ID)
	if err != nil {
		// 感染信息属于可选区块，legacy 环境下读失败时降级为 nil，避免影响核心页可用性
		infection = nil
	}
	currentPlan := s.buildCurrentPlan(patient)
	activeOrders, _ := s.buildActiveOrders(patient.ID)
	labTrends, _ := s.buildLabTrends(patient.ID)

	return &PatientCoreOverview{
		Infection:    infection,   // 可能是 nil
		CurrentPlan:  currentPlan, // 可能是 nil
		ActiveOrders: activeOrders,
		LabTrends:    labTrends,
	}, nil
}

// buildInfection 构建感染信息
func (s *PatientCoreService) buildInfection(patientID modeltypes.LegacyID) (*PatientCoreInfection, error) {
	var infection legacyCoreInfection
	err := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, patientID, legacyTenantID).First(&infection).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有感染记录时，返回 nil
			return nil, nil
		}
		return nil, err
	}

	combined := strings.Join(nonEmptyStrings(infection.InfectionDesc, infection.OtherDesc, infection.Note), "；")

	return &PatientCoreInfection{
		HbsAg:      detectLegacyInfectionValue(combined, []string{"hbsag", "hbv", "乙肝", "乙型"}),
		HcvAb:      detectLegacyInfectionValue(combined, []string{"hcv", "丙肝", "丙型"}),
		HivAb:      detectLegacyInfectionValue(combined, []string{"hiv", "艾滋"}),
		TpAb:       detectLegacyInfectionValue(combined, []string{"tp", "梅毒"}),
		UpdateDate: infection.LastModifyTime,
	}, nil
}

// buildCurrentPlan 构建当前治疗方案，如果没有治疗方案返回 nil
func (s *PatientCoreService) buildCurrentPlan(patient models.Patient) *PatientCoreCurrentPlan {
	var treatmentPlan legacyCorePlan
	err := s.db.Where(`"PatientId" = ? AND "TenantId" = ?`, patient.ID, legacyTenantID).
		Order("\"LastModifyTime\" DESC").
		Order("\"CreateTime\" DESC").
		First(&treatmentPlan).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有治疗方案记录，返回 nil
			return nil
		}
		return nil
	}

	// 判断方案是否启用
	if treatmentPlan.IsDisabled {
		return nil
	}

	weeklyFrequency := maxInt(treatmentPlan.OddWeekFrequency, treatmentPlan.EvenWeekFrequency)
	if weeklyFrequency <= 0 {
		weeklyFrequency = 3
	}
	dialysisMode := s.stringOrDefault(strings.TrimSpace(treatmentPlan.DialysisMethod), "HD")
	bloodFlow := treatmentPlan.BF
	if bloodFlow <= 0 {
		bloodFlow = 200
	}

	return &PatientCoreCurrentPlan{
		DialysisMode:  dialysisMode,
		Frequency:     s.buildFrequency(int64(weeklyFrequency)),
		Duration:      treatmentPlan.DialysisDuration,
		DryWeight:     treatmentPlan.DryWeight,
		BloodFlow:     bloodFlow,
		Anticoagulant: legacyAnticoagulantName(treatmentPlan),
		LastNote:      s.stringOrEmpty(treatmentPlan.Note),
	}
}

// buildActiveOrders 构建活跃医嘱列表
func (s *PatientCoreService) buildActiveOrders(patientID modeltypes.LegacyID) ([]PatientCoreOrder, error) {
	var orders []legacyCoreOrder
	err := s.db.Where("\"PatientId\" = ? AND \"TenantId\" = ? AND COALESCE(\"IsDisabled\", false) = false", patientID, legacyTenantID).
		Order("\"StartTime\" DESC").
		Limit(10).
		Find(&orders).Error
	if err != nil {
		return []PatientCoreOrder{}, nil
	}

	result := make([]PatientCoreOrder, 0, len(orders))
	for _, o := range orders {
		orderType := strings.TrimSpace(o.Classification)
		if orderType == "" {
			orderType = strings.TrimSpace(o.Type)
		}
		result = append(result, PatientCoreOrder{
			ID:        fmt.Sprintf("%d", o.ID),
			Content:   o.Content,
			Type:      orderType,
			StartTime: o.StartTime,
			Doctor:    legacyOperatorName(o.OperatorID),
		})
	}
	return result, nil
}

// buildLabTrends 构建检验指标趋势
func (s *PatientCoreService) buildLabTrends(patientID modeltypes.LegacyID) ([]PatientCoreLabTrend, error) {
	// 定义关键指标：血红蛋白、钙、磷
	indicators := []struct {
		codes  []string
		names  []string
		name   string
		unit   string
		normal string
	}{
		{[]string{"HGB", "HB", "血红蛋白"}, []string{"血红蛋白", "Hb", "HGB"}, "血红蛋白", "g/L", "120-160"},
		{[]string{"Ca", "CA", "钙"}, []string{"钙", "总钙", "血钙"}, "钙", "mmol/L", "2.12-2.75"},
		{[]string{"P", "PHOS", "磷"}, []string{"磷", "血磷"}, "磷", "mmol/L", "0.87-1.45"},
	}

	// 查询近6个月的检验报告明细
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	var trends []PatientCoreLabTrend

	for _, indicator := range indicators {
		var items []legacyCoreLabRow

		err := s.db.Table(`"LIS_ExaminationItem" AS i`).
			Select(`i."ItemCode" AS item_code, i."ItemName" AS item_name, i."Result" AS result_value, i."Unit" AS unit, i."Reference" AS reference_range, i."ResultSign" AS result_sign, COALESCE(e."ResultTime", i."LastModifyTime") AS tested_at`).
			Joins(`JOIN "LIS_Examination" AS e ON e."Id" = i."ExaminationId"`).
			Where(`e."PatientId" = ? AND e."TenantId" = ? AND e."ResultTime" >= ? AND (i."ItemCode" IN ? OR i."ItemName" IN ?)`,
				patientID, legacyTenantID, sixMonthsAgo, indicator.codes, indicator.names).
			Order(`COALESCE(e."ResultTime", i."LastModifyTime") ASC`).
			Limit(12).
			Find(&items).Error

		var data []PatientCoreLabData
		if err == nil {
			for _, item := range items {
				var val float64
				if _, err2 := fmt.Sscanf(strings.TrimSpace(item.ResultValue), "%f", &val); err2 == nil {
					dateStr := ""
					if item.TestedAt != nil {
						dateStr = item.TestedAt.Format("2006-01-02")
					}
					data = append(data, PatientCoreLabData{
						Date:       dateStr,
						Value:      val,
						IsAbnormal: isLegacyLabAbnormal(item.ResultSign),
					})
				}
			}
		}

		if data == nil {
			data = []PatientCoreLabData{}
		}

		trends = append(trends, PatientCoreLabTrend{
			Code:        indicator.codes[0],
			Name:        indicator.name,
			Unit:        indicator.unit,
			NormalRange: indicator.normal,
			Data:        data,
		})
	}

	return trends, nil
}

// buildClinicalFocus 构建临床焦点面板
func (s *PatientCoreService) buildClinicalFocus(patient models.Patient) PatientCoreClinical {
	var alerts []PatientCoreAlert

	// 查询近30天异常检验项目（AbnormalFlag = H 或 L）
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var abnormalItems []legacyCoreLabRow
	err := s.db.Table(`"LIS_ExaminationItem" AS i`).
		Select(`i."ItemCode" AS item_code, i."ItemName" AS item_name, i."Result" AS result_value, i."Unit" AS unit, i."Reference" AS reference_range, i."ResultSign" AS result_sign, COALESCE(e."ResultTime", i."LastModifyTime") AS tested_at`).
		Joins(`JOIN "LIS_Examination" AS e ON e."Id" = i."ExaminationId"`).
		Where(`e."PatientId" = ? AND e."TenantId" = ? AND COALESCE(e."ResultTime", i."LastModifyTime") >= ? AND COALESCE(i."ResultSign", '') <> ''`,
			patient.ID, legacyTenantID, thirtyDaysAgo).
		Order(`COALESCE(e."ResultTime", i."LastModifyTime") DESC`).
		Limit(5).
		Find(&abnormalItems).Error

	if err == nil {
		for _, item := range abnormalItems {
			var val float64
			fmt.Sscanf(strings.TrimSpace(item.ResultValue), "%f", &val)

			severity := "warning"
			thresholds := GetLabThresholds(item.ItemCode)
			if thresholds.CriticalLow != nil && val < *thresholds.CriticalLow {
				severity = "critical"
			} else if thresholds.CriticalHigh != nil && val > *thresholds.CriticalHigh {
				severity = "critical"
			}

			measuredAt := time.Now()
			if item.TestedAt != nil {
				measuredAt = *item.TestedAt
			}

			alertID := fmt.Sprintf("lab_%s_%s", item.ItemCode, measuredAt.Format(time.RFC3339))
			alerts = append(alerts, PatientCoreAlert{
				ID:             alertID,
				Type:           "lab",
				Name:           item.ItemName,
				Value:          item.ResultValue,
				Unit:           item.Unit,
				Severity:       severity,
				ReferenceRange: item.ReferenceRange,
				MeasuredAt:     measuredAt,
			})
		}
	}

	if alerts == nil {
		alerts = []PatientCoreAlert{}
	}

	return PatientCoreClinical{
		CriticalAlerts: alerts,
		DocumentStatus: []PatientCoreDoc{},
		LastSyncAt:     time.Now().Format(time.RFC3339),
	}
}

// buildNavigation 构建患者导航信息
func (s *PatientCoreService) buildNavigation(patientID modeltypes.LegacyID) (*PatientCoreNavigation, error) {
	// legacy 患者主表没有 bed_number / active 状态语义，按患者主键稳定排序即可
	var patients []models.Patient
	err := s.db.Where("\"TenantId\" = ?", legacyTenantID).Order("\"Id\" ASC").Find(&patients).Error
	if err != nil {
		// 某些 legacy 库可能缺少 TenantId 过滤语义，回退到全量患者排序
		if fallbackErr := s.db.Order("\"Id\" ASC").Find(&patients).Error; fallbackErr != nil {
			return &PatientCoreNavigation{Total: 1, CurrentIndex: 0}, nil
		}
	}

	total := len(patients)
	currentIndex := -1
	for i, p := range patients {
		if p.ID == patientID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return &PatientCoreNavigation{Total: total, CurrentIndex: 0}, nil
	}

	var prev, next *PatientCoreNavPatient
	if currentIndex > 0 {
		p := patients[currentIndex-1]
		prev = &PatientCoreNavPatient{
			ID:        legacyIDString(p.ID),
			Name:      p.Name,
			BedNumber: p.BedNumber,
		}
	}
	if currentIndex < total-1 {
		p := patients[currentIndex+1]
		next = &PatientCoreNavPatient{
			ID:        legacyIDString(p.ID),
			Name:      p.Name,
			BedNumber: p.BedNumber,
		}
	}

	return &PatientCoreNavigation{
		Previous:     prev,
		Next:         next,
		Total:        total,
		CurrentIndex: currentIndex,
	}, nil
}

func detectLegacyInfectionValue(source string, keywords []string) string {
	normalized := strings.ToLower(strings.TrimSpace(source))
	if normalized == "" {
		return "阴性"
	}
	for _, keyword := range keywords {
		if strings.Contains(normalized, strings.ToLower(keyword)) {
			return strings.TrimSpace(source)
		}
	}
	return "阴性"
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func legacyAnticoagulantName(plan legacyCorePlan) string {
	if plan.FirstAnticoagulant > 0 || plan.MaintainAnticoagulant > 0 {
		return "已配置"
	}
	return "未记录"
}

func legacyOperatorName(operatorID int64) string {
	if operatorID <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", operatorID)
}

func isLegacyLabAbnormal(sign string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(sign))
	if normalized == "" {
		return false
	}
	return strings.Contains(normalized, "H") || strings.Contains(normalized, "L") || strings.Contains(sign, "↑") || strings.Contains(sign, "↓") || strings.Contains(sign, "高") || strings.Contains(sign, "低")
}

// ============ 辅助方法 ============

// calculateDialysisAge 计算透析龄
func (s *PatientCoreService) calculateDialysisAge(startDate time.Time) string {
	if startDate.IsZero() {
		return ""
	}

	now := time.Now()
	years := now.Year() - startDate.Year()
	months := int(now.Month()) - int(startDate.Month())

	if months < 0 {
		years--
		months += 12
	}

	if years > 0 && months > 0 {
		return fmt.Sprintf("%d年%d个月", years, months)
	} else if years > 0 {
		return fmt.Sprintf("%d年", years)
	} else if months > 0 {
		return fmt.Sprintf("%d个月", months)
	}

	return ""
}

// buildFrequency 构建频次描述
func (s *PatientCoreService) buildFrequency(weeklyFreq int64) string {
	switch weeklyFreq {
	case 1:
		return "1次/周"
	case 2:
		return "2次/周"
	case 3:
		return "3次/周"
	case 4:
		return "4次/周"
	default:
		return "3次/周"
	}
}

// stringOrDefault 安全获取字符串值（非指针版本）
func (s *PatientCoreService) stringOrDefault(str string, defaultVal string) string {
	if str == "" {
		return defaultVal
	}
	return str
}

// stringOrEmpty 安全获取字符串值，空值返回空字符串
func (s *PatientCoreService) stringOrEmpty(str string) string {
	return str
}

// CalculateDialysisAge 计算透析龄（独立函数，供外部调用）
func CalculateDialysisAge(startDate time.Time) string {
	if startDate.IsZero() {
		return ""
	}

	now := time.Now()
	years := now.Year() - startDate.Year()
	months := int(now.Month()) - int(startDate.Month())

	if months < 0 {
		years--
		months += 12
	}

	if years > 0 && months > 0 {
		return fmt.Sprintf("%d年%d个月", years, months)
	} else if years > 0 {
		return fmt.Sprintf("%d年", years)
	} else if months > 0 {
		return fmt.Sprintf("%d个月", months)
	}

	return ""
}
