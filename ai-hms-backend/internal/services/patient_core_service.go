package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

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

	var patient models.Patient
	err := s.db.First(&patient, "id = ?", patientID).Error
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
	// navigation, _ := s.buildNavigation(patientID)

	return &PatientCoreResponse{
		Header:        header,
		Overview:      *overview,
		ClinicalFocus: clinicalFocus,
		Navigation:    nil, // 暂不实现，需要从患者列表缓存获取
	}, nil
}

// buildHeader 构建页面头部信息
func (s *PatientCoreService) buildHeader(patient models.Patient) PatientCoreHeader {
	// 计算透析龄
	dialysisAge := s.calculateDialysisAge(patient.CreatedAt)

	// 转换状态：active -> 治疗中
	status := "待诊"
	if patient.Status == "active" {
		// TODO: 判断是否有当前治疗会话
		status = "透析中"
	}

	return PatientCoreHeader{
		ID:          patient.ID,
		Name:        patient.Name,
		Gender:      patient.Gender,
		Age:         patient.Age,
		BedNumber:   patient.BedNumber,
		PatientType:  s.stringOrDefault(patient.PatientType, "门诊"),
		InsuranceType: s.stringOrDefault(patient.InsuranceType, "自费"),
		DoctorName:   s.stringOrDefault(patient.DoctorName, ""),
		RiskLevel:    s.stringOrDefault(patient.RiskLevel, "低危"),
		Status:       status,
		DialysisAge:  dialysisAge,
	}
}

// buildOverview 构建 Overview Tab 数据
func (s *PatientCoreService) buildOverview(patient models.Patient) (*PatientCoreOverview, error) {
	infection, _ := s.buildInfection(patient.ID)
	currentPlan := s.buildCurrentPlan(patient)
	activeOrders, _ := s.buildActiveOrders(patient.ID)
	labTrends, _ := s.buildLabTrends(patient.ID)

	return &PatientCoreOverview{
		Infection:    infection,  // 可能是 nil
		CurrentPlan:  currentPlan, // 可能是 nil
		ActiveOrders: activeOrders,
		LabTrends:   labTrends,
	}, nil
}

// buildInfection 构建感染信息
func (s *PatientCoreService) buildInfection(patientID string) (*PatientCoreInfection, error) {
	var infection models.InfectionInfo
	err := s.db.Where("patient_id = ?", patientID).First(&infection).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有感染记录时，返回 nil
			return nil, nil
		}
		return nil, err
	}

	return &PatientCoreInfection{
		HbsAg:      s.stringOrDefault(infection.HbsAg, "阴性"),
		HcvAb:      s.stringOrDefault(infection.HcvAb, "阴性"),
		HivAb:      s.stringOrDefault(infection.HivAb, "阴性"),
		TpAb:       s.stringOrDefault(infection.TpaB, "阴性"),
		UpdateDate: infection.UpdateDate,
	}, nil
}

// buildCurrentPlan 构建当前治疗方案，如果没有治疗方案返回 nil
func (s *PatientCoreService) buildCurrentPlan(patient models.Patient) *PatientCoreCurrentPlan {
	// 获取治疗方案
	var treatmentPlan models.TreatmentPlan
	err := s.db.Where("patient_id = ?", patient.ID).
		Order("created_at DESC").
		First(&treatmentPlan).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有治疗方案记录，返回 nil
			return nil
		}
		return nil
	}

	// 判断方案是否启用
	status := s.stringOrDefault(treatmentPlan.Status, "启用")
	if status != "启用" {
		// 方案未启用，返回 nil
		return nil
	}

	// 从 JSON 字段提取数据
	dialysisMode := "HD"
	bloodFlow := 200

	// 如果有 dialysis_mode JSON，从中提取
	if treatmentPlan.DialysisMode.Mode != "" {
		dialysisMode = treatmentPlan.DialysisMode.Mode
		bloodFlow = treatmentPlan.DialysisMode.BloodFlow
	}

	return &PatientCoreCurrentPlan{
		DialysisMode:  dialysisMode,
		Frequency:     s.buildFrequency(int64(treatmentPlan.WeeklyFrequency)),
		Duration:      int(treatmentPlan.Duration),
		DryWeight:     treatmentPlan.DryWeight,
		BloodFlow:     bloodFlow,
		Anticoagulant: "肝素钠",
		LastNote:      s.stringOrEmpty(treatmentPlan.Notes),
	}
}

// buildActiveOrders 构建活跃医嘱列表
func (s *PatientCoreService) buildActiveOrders(_ string) ([]PatientCoreOrder, error) {
	// TODO: 从 Orders 表查询活跃医嘱
	// 当前先返回空数组
	return []PatientCoreOrder{}, nil
}

// buildLabTrends 构建检验指标趋势
func (s *PatientCoreService) buildLabTrends(_ string) ([]PatientCoreLabTrend, error) {
	// 定义关键指标：血红蛋白、钙、磷
	indicators := []struct {
		code   string
		name   string
		unit   string
		normal string
	}{
		{"HGB", "血红蛋白", "g/L", "120-160"},
		{"Ca", "钙", "mmol/L", "2.12-2.75"},
		{"P", "磷", "mmol/L", "0.87-1.45"},
	}

	var trends []PatientCoreLabTrend

	for _, indicator := range indicators {
		// TODO: 从 Examination 表查询最近 6 个月的数据
		// 当前先返回空趋势
		trends = append(trends, PatientCoreLabTrend{
			Code:        indicator.code,
			Name:        indicator.name,
			Unit:        indicator.unit,
			NormalRange: indicator.normal,
			Data:        []PatientCoreLabData{},
		})
	}

	return trends, nil
}

// buildClinicalFocus 构建临床焦点面板
func (s *PatientCoreService) buildClinicalFocus(_ models.Patient) PatientCoreClinical {
	// TODO: 实现危急值判断和文书状态查询
	// 当前先返回空数据
	return PatientCoreClinical{
		CriticalAlerts: []PatientCoreAlert{},
		DocumentStatus: []PatientCoreDoc{},
		LastSyncAt:    time.Now().Format(time.RFC3339),
	}
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
