package services

import (
	"time"
)

// PatientCoreResponse 患者核心信息聚合响应（/core 接口）
type PatientCoreResponse struct {
	Header          PatientCoreHeader      `json:"header"`
	Overview        PatientCoreOverview    `json:"overview"`
	ClinicalFocus   PatientCoreClinical   `json:"clinicalFocus"`
	Navigation      *PatientCoreNavigation `json:"navigation,omitempty"`
}

// ============ 页面头部信息 ============

type PatientCoreHeader struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Avatar      string     `json:"avatar,omitempty"`
	Gender      string     `json:"gender"`
	Age         int        `json:"age"`
	BedNumber   string     `json:"bedNumber"`
	PatientType string     `json:"patientType"`
	InsuranceType string    `json:"insuranceType"`
	DoctorName  string     `json:"doctorName"`
	RiskLevel   string     `json:"riskLevel"`   // 高危, 中危, 低危
	Status      string     `json:"status"`       // 治疗中, 待诊, 已结束
	DialysisAge string     `json:"dialysisAge,omitempty"` // 如 "3年2个月"
}

// ============ Overview Tab 数据 ============

type PatientCoreOverview struct {
	Infection     *PatientCoreInfection    `json:"infection,omitempty"` // 指针类型，无数据时为 nil
	CurrentPlan   *PatientCoreCurrentPlan  `json:"currentPlan,omitempty"` // 指针类型，无数据时为 nil
	ActiveOrders  []PatientCoreOrder       `json:"activeOrders"`
	LabTrends     []PatientCoreLabTrend     `json:"labTrends"`
}

// 传染病标志（5项）
type PatientCoreInfection struct {
	HbsAg     string    `json:"hbsag"`     // 乙肝表面抗原
	HcvAb     string    `json:"hcvab"`     // 丙肝抗体
	HivAb     string    `json:"hivab"`     // HIV抗体
	TpAb      string    `json:"tpab"`      // 梅毒抗体
	Tb        *string   `json:"tb"`         // 结核（可选）
	UpdateDate time.Time `json:"updateDate"`
}

// 当前核心诊疗方案
type PatientCoreCurrentPlan struct {
	DialysisMode   string  `json:"dialysisMode"`   // HD/HDF/CRRT
	Frequency      string  `json:"frequency"`      // "3次/周"
	Duration       int     `json:"duration"`       // 时长(小时)
	DryWeight      float64 `json:"dryWeight"`      // 干体重
	BloodFlow      int     `json:"bloodFlow"`      // 血流量
	Anticoagulant  string  `json:"anticoagulant"`  // 抗凝剂方案
	LastNote       string  `json:"lastTreatmentNote"` // 上次治疗动态
}

// 当前活跃医嘱
type PatientCoreOrder struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`      // 长期, 临时
	StartTime time.Time `json:"startTime"`
	Doctor    string    `json:"doctor"`
}

// 关键检验指标趋势
type PatientCoreLabTrend struct {
	Code         string                `json:"code"`         // 'HGB' | 'Ca' | 'P'
	Name         string                `json:"name"`
	Unit         string                `json:"unit"`
	NormalRange  string                `json:"normalRange"`
	Data         []PatientCoreLabData   `json:"data"`
}

type PatientCoreLabData struct {
	Date       string  `json:"date"`
	Value      float64 `json:"value"`
	IsAbnormal bool    `json:"isAbnormal"`
}

// ============ 临床焦点面板 ============

type PatientCoreClinical struct {
	CriticalAlerts  []PatientCoreAlert  `json:"criticalAlerts"`
	DocumentStatus  []PatientCoreDoc   `json:"documentStatus"`
	LastSyncAt      string             `json:"lastSyncAt"`
}

// 危急值提醒
type PatientCoreAlert struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`           // 'lab' | 'vital' | 'medication'
	Name           string    `json:"name"`
	Value          string    `json:"value"`
	Unit           string    `json:"unit"`
	Severity       string    `json:"severity"`       // 'critical' | 'warning' | 'info'
	ReferenceRange  string    `json:"referenceRange"`
	AISuggestion    *string   `json:"aiSuggestion,omitempty"`
	MeasuredAt     time.Time `json:"measuredAt"`
}

// 文书状态
type PatientCoreDoc struct {
	ID           string    `json:"id"`
	DocumentName string    `json:"documentName"`
	Status       string    `json:"status"`    // '待签署' | '已完成'
	DueDate      *string   `json:"dueDate,omitempty"`
	Priority     string    `json:"priority"`  // 'high' | 'medium' | 'low'
}

// ============ 患者导航 ============

type PatientCoreNavigation struct {
	Previous     *PatientCoreNavPatient `json:"previous,omitempty"`
	Next         *PatientCoreNavPatient `json:"next,omitempty"`
	Total        int                    `json:"total"`
	CurrentIndex int                    `json:"currentIndex"`
}

type PatientCoreNavPatient struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BedNumber string `json:"bedNumber"`
}

// ============ 危急值判断规则 ============

// LabCriticalThresholds 检验危急值阈值定义
type LabCriticalThresholds struct {
	Low    float64 `json:"low"`    // 低危阈值
	Normal float64 `json:"normal"` // 正常范围下限
	High   float64 `json:"high"`   // 正常范围上限
	CriticalLow  *float64 `json:"criticalLow,omitempty"`  // 危急低值
	CriticalHigh *float64 `json:"criticalHigh,omitempty"` // 危急高值
}

// GetLabThresholds 获取检验指标的危急值阈值
func GetLabThresholds(indicatorCode string) LabCriticalThresholds {
	thresholds := map[string]LabCriticalThresholds{
		"K": {
			Low:         3.5,
			Normal:      5.5,
			High:        5.5,
			CriticalLow:  &[]float64{2.8}[0],
			CriticalHigh:&[]float64{6.2}[0],
		},
		"Na": {
			Low:         137,
			Normal:      145,
			High:        145,
			CriticalLow:  &[]float64{120}[0],
			CriticalHigh:&[]float64{160}[0],
		},
		"Ca": {
			Low:         2.12,
			Normal:      2.5,
			High:        2.75,
			CriticalLow:  &[]float64{1.75}[0],
			CriticalHigh:&[]float64{3.5}[0],
		},
		"GLU": {
			Low:         3.9,
			Normal:      6.1,
			High:        6.1,
			CriticalLow:  &[]float64{2.2}[0],
			CriticalHigh:&[]float64{22.2}[0],
		},
		"HGB": {
			Low:         120,
			Normal:      160,
			High:        160,
			CriticalLow:  &[]float64{30}[0],
			CriticalHigh:&[]float64{200}[0],
		},
		"PLT": {
			Low:         100,
			Normal:      300,
			High:        300,
			CriticalLow:  &[]float64{20}[0],
			CriticalHigh:&[]float64{1000}[0],
		},
		"WBC": {
			Low:         4.0,
			Normal:      10.0,
			High:        10.0,
			CriticalLow:  &[]float64{0.5}[0],
			CriticalHigh: nil,
		},
	}

	if t, ok := thresholds[indicatorCode]; ok {
		return t
	}

	// 默认阈值
	return LabCriticalThresholds{
		Low:    0,
		Normal: 100,
		High:   100,
	}
}

// GetSeverityLevel 根据阈值判断严重程度
func GetSeverityLevel(value float64, thresholds LabCriticalThresholds) string {
	if thresholds.CriticalLow != nil && value < *thresholds.CriticalLow {
		return "critical"
	}
	if thresholds.CriticalHigh != nil && value > *thresholds.CriticalHigh {
		return "critical"
	}
	if value < thresholds.Low || value > thresholds.High {
		return "warning"
	}
	return "info"
}
