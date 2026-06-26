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

// Sample 单个 Device_DMLog 设备时点，是模型输入窗口的一行。
// 字段按模型特征清单（前 30 个 DMLog 时点拉平 + 各自均值/标准差）补全，
// 此处先列模型最依赖的几项，接入时对齐训练特征列顺序。
type Sample struct {
	TMP              float64 `json:"tmp"`
	UFVolume         float64 `json:"ufVolume"`
	VenousPressure   float64 `json:"venousPressure"`
	ArterialPressure float64 `json:"arterialPressure"`
	BF               float64 `json:"bf"`
	Conductivity     float64 `json:"conductivity"`
}

// RiskInput 一次 IDH 风险评分输入：某治疗近 30 个 DMLog 时点窗口（升序）。
type RiskInput struct {
	TreatmentID int64    `json:"treatmentId"`
	AccessType  string   `json:"accessType"`
	Window      []Sample `json:"window"`
}

// RiskResult IDH 风险评分输出。Available=false 时墙上不显风险（链路完整、不报错）。
type RiskResult struct {
	Available   bool    `json:"available"`
	Probability float64 `json:"probability"` // 下次测量低血压概率 [0,1]
	Level       string  `json:"level"`       // high | medium | low（按概率映射）
}
