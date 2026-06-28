// Package idh 透中低血压（IDH）风险预警的可插拔评分接口。
//
// 本期只搭骨架：默认 StubScorer（恒不可用占位），真模型随后以 HTTPScorer 接入
// Python「IDH 预警」FastAPI 微服务（与 ACTRS 平级）。真模型来自 dixueya 研究
// （CatBoost+SMOTE，AUC≈0.91），详见 桌面\HMS开发\实时监控_IDH低血压预测模型_评估_20260625.md。
package idh

// Config IDH 预警微服务配置。
type Config struct {
	BaseURL    string
	TimeoutSec int
	Enabled    bool
}

// Sample 单个 Device_DMLog 设备时点（模型输入窗口一行）。
// json key 必须与 Python features.py::FLATTEN_COLS 逐字一致（PascalCase）。
// 全部指针：缺列→null→Python 端 NaN。
type Sample struct {
	LogTime             *float64 `json:"LogTime"`
	TMP                 *float64 `json:"TMP"`
	UFVolume            *float64 `json:"UFVolume"`
	VenousPressure      *float64 `json:"VenousPressure"`
	ArterialPressure    *float64 `json:"ArterialPressure"`
	BF                  *float64 `json:"BF"`
	Conductivity        *float64 `json:"Conductivity"`
	APumpSpeedDeviation *float64 `json:"APumpSpeedDeviation"`
	BPumpSpeedDeviation *float64 `json:"BPumpSpeedDeviation"`
	HeparinPumpFlow     *float64 `json:"HeparinPumpFlow"`
	AConductivity       *float64 `json:"AConductivity"`
	DialysateTemp       *float64 `json:"DialysateTemp"`
	TreatmentTime       *float64 `json:"TreatmentTime"`
	UFSetVolume         *float64 `json:"UFSetVolume"`
	UFQuantity          *float64 `json:"UFQuantity"`
	BConductivity       *float64 `json:"BConductivity"`
	DeviceId            *float64 `json:"DeviceId"`
	SubstituateVolume   *float64 `json:"SubstituateVolume"`
	HeparinVolume       *float64 `json:"HeparinVolume"`
	SubstituateSpeed    *float64 `json:"SubstituateSpeed"`
}

// BasicInfo 病人基本信息（json key/alias 必须与 Python BasicInfo 逐字一致）。
type BasicInfo struct {
	Gender         *int     `json:"Gender"` // 男=1/女=0
	Age            *float64 `json:"Age"`
	DialysisMethod *string  `json:"DialysisMethod"`
	DryWeight      *float64 `json:"DryWeight"`
	UFQuantityY    *float64 `json:"UFQuantity_y"`
	PreWeight      *float64 `json:"pre-Weight"`
	PreSBP         *float64 `json:"pre-SBP"`
	PreDBP         *float64 `json:"pre-DBP"`
	SBP            *float64 `json:"SBP"`
	DBP            *float64 `json:"DBP"`
}

// RiskInput 一次 IDH 风险评分输入：某治疗近 30 个 DMLog 时点窗口（升序）+ 基本信息。
type RiskInput struct {
	TreatmentID int64     `json:"treatmentId"`
	AccessType  string    `json:"accessType"`
	Window      []Sample  `json:"window"`
	Basic       BasicInfo `json:"basic"`
}

// RiskResult IDH 风险评分输出。Available=false 时墙上不显风险（链路完整、不报错）。
type RiskResult struct {
	Available   bool    `json:"available"`
	Probability float64 `json:"probability"` // 下次测量低血压概率 [0,1]
	Level       string  `json:"level"`       // high | medium | low（按概率映射）
}
