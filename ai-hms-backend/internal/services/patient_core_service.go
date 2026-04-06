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
	navigation, _ := s.buildNavigation(patientID)

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
		ID:            patient.ID,
		Name:          patient.Name,
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

// buildOverview 构建 Overview Tab 数据
func (s *PatientCoreService) buildOverview(patient models.Patient) (*PatientCoreOverview, error) {
	infection, _ := s.buildInfection(patient.ID)
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
func (s *PatientCoreService) buildActiveOrders(patientID string) ([]PatientCoreOrder, error) {
	var orders []models.Order
	err := s.db.Where("patient_id = ? AND status IN ?", patientID, []string{
		models.OrderStatusPending,
		models.OrderStatusExecuting,
	}).
		Order("start_time DESC").
		Limit(10).
		Find(&orders).Error
	if err != nil {
		return []PatientCoreOrder{}, nil
	}

	result := make([]PatientCoreOrder, 0, len(orders))
	for _, o := range orders {
		result = append(result, PatientCoreOrder{
			ID:        o.ID,
			Content:   o.Content,
			Type:      o.Type,
			StartTime: o.StartTime,
			Doctor:    o.DoctorName,
		})
	}
	return result, nil
}

// buildLabTrends 构建检验指标趋势
func (s *PatientCoreService) buildLabTrends(patientID string) ([]PatientCoreLabTrend, error) {
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

	// 查询近6个月的检验报告明细
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	var trends []PatientCoreLabTrend

	for _, indicator := range indicators {
		var items []models.LabReportItem

		// 联表查询：通过 lab_reports 关联 lab_report_items
		err := s.db.Table("lab_report_items").
			Joins("JOIN lab_reports ON lab_reports.id = lab_report_items.lab_report_id").
			Where("lab_reports.patient_id = ? AND lab_report_items.item_code = ? AND lab_reports.reported_at >= ?",
				patientID, indicator.code, sixMonthsAgo).
			Order("lab_report_items.tested_at ASC").
			Limit(12). // 最多12个数据点（约6个月，每月2次）
			Find(&items).Error

		var data []PatientCoreLabData
		if err == nil {
			for _, item := range items {
				var val float64
				if _, err2 := fmt.Sscanf(item.ResultValue, "%f", &val); err2 == nil {
					dateStr := ""
					if item.TestedAt != nil {
						dateStr = item.TestedAt.Format("2006-01-02")
					}
					data = append(data, PatientCoreLabData{
						Date:       dateStr,
						Value:      val,
						IsAbnormal: item.AbnormalFlag == "H" || item.AbnormalFlag == "L",
					})
				}
			}
		}

		if data == nil {
			data = []PatientCoreLabData{}
		}

		trends = append(trends, PatientCoreLabTrend{
			Code:        indicator.code,
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

	var abnormalItems []models.LabReportItem
	err := s.db.Table("lab_report_items").
		Joins("JOIN lab_reports ON lab_reports.id = lab_report_items.lab_report_id").
		Where("lab_reports.patient_id = ? AND lab_report_items.abnormal_flag IN ? AND lab_report_items.tested_at >= ?",
			patient.ID, []string{"H", "L"}, thirtyDaysAgo).
		Order("lab_report_items.tested_at DESC").
		Limit(5).
		Find(&abnormalItems).Error

	if err == nil {
		for _, item := range abnormalItems {
			var val float64
			fmt.Sscanf(item.ResultValue, "%f", &val)

			severity := "warning"
			thresholds := GetLabThresholds(item.ItemCode)
			if thresholds.CriticalLow != nil && val < *thresholds.CriticalLow {
				severity = "critical"
			} else if thresholds.CriticalHigh != nil && val > *thresholds.CriticalHigh {
				severity = "critical"
			}

			measuredAt := item.CreatedAt
			if item.TestedAt != nil {
				measuredAt = *item.TestedAt
			}

			alertID := fmt.Sprintf("lab_%s_%s", item.ItemCode, item.ID)
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
func (s *PatientCoreService) buildNavigation(patientID string) (*PatientCoreNavigation, error) {
	// 简单实现：获取所有活跃患者，找到当前患者的位置
	var patients []models.Patient
	err := s.db.Where("status = ?", "active").Order("bed_number ASC").Find(&patients).Error
	if err != nil {
		return nil, err
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
			ID:        p.ID,
			Name:      p.Name,
			BedNumber: p.BedNumber,
		}
	}
	if currentIndex < total-1 {
		p := patients[currentIndex+1]
		next = &PatientCoreNavPatient{
			ID:        p.ID,
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
