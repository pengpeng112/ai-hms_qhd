// Package config 实时监控报警阈值表（go:embed 内嵌 JSON，院方改阈值重构建即生效；
// 后续可升级为可编辑 DB 表 + admin UI，本文件为种子默认）。
// 弃用设备上报的 device.status，改由后端按本表逐指标算 warning/danger。
// 静脉压 VP 为特例：按「通路类型 × 血流量 BF」分层查表（源本院 VP 大样本研究/CMPB）。
package config

import (
	_ "embed"
	"encoding/json"
	"math"
	"strings"
)

//go:embed monitoring_thresholds.json
var monitoringThresholdsJSON []byte

// AlarmLevel 报警分级。
type AlarmLevel string

const (
	AlarmNormal  AlarmLevel = "normal"
	AlarmWarning AlarmLevel = "warning"
	AlarmDanger  AlarmLevel = "danger"
)

// FixedThreshold 五档固定阈值（危险低 / 警戒低 / 正常 / 警戒高 / 危险高）。
// 任一侧为 nil 表示该侧无界（如超滤率只有上界）。
type FixedThreshold struct {
	Label      string   `json:"label"`
	Enabled    bool     `json:"enabled"`
	Unit       string   `json:"unit"`
	DangerLow  *float64 `json:"dangerLow"`
	WarnLow    *float64 `json:"warnLow"`
	WarnHigh   *float64 `json:"warnHigh"`
	DangerHigh *float64 `json:"dangerHigh"`
}

// Eval 按五档给单值分级。未启用返回 normal。
func (t FixedThreshold) Eval(v float64) AlarmLevel {
	if !t.Enabled {
		return AlarmNormal
	}
	if (t.DangerLow != nil && v < *t.DangerLow) || (t.DangerHigh != nil && v > *t.DangerHigh) {
		return AlarmDanger
	}
	if (t.WarnLow != nil && v < *t.WarnLow) || (t.WarnHigh != nil && v > *t.WarnHigh) {
		return AlarmWarning
	}
	return AlarmNormal
}

// VPStratum 静脉压分层（通路类型 × BF 区间 [BFMin, BFMax)）阈值。
// NormalLow=P10（低于→警戒低）、WarnHigh=P90（高于→警戒高）、DangerHigh=P95（高于→危险高）。
// 研究聚焦高侧，故不设危险低。
type VPStratum struct {
	Access     string  `json:"access"`
	BFMin      float64 `json:"bfMin"`
	BFMax      float64 `json:"bfMax"`
	NormalLow  float64 `json:"normalLow"`
	WarnHigh   float64 `json:"warnHigh"`
	DangerHigh float64 `json:"dangerHigh"`
}

// MonitoringThresholds 阈值表整体。
type MonitoringThresholds struct {
	Fixed       map[string]FixedThreshold `json:"fixed"`
	VPReference []VPStratum               `json:"vpReference"`
}

// LoadMonitoringThresholds 解析内嵌阈值表。
func LoadMonitoringThresholds() (*MonitoringThresholds, error) {
	var out MonitoringThresholds
	if err := json.Unmarshal(monitoringThresholdsJSON, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EvalFixed 按 key（map/heartRate/ufr/dialysateNa）查固定阈值并分级；无该 key 返回 normal。
func (m *MonitoringThresholds) EvalFixed(key string, v float64) AlarmLevel {
	t, ok := m.Fixed[key]
	if !ok {
		return AlarmNormal
	}
	return t.Eval(v)
}

// ClassifyAccess4 把原始通路类型串细分为 AVF/AVG/TCC/NCC（VP 分层需区分 AVF 与 AVG）。
// 无法识别返回 ""。注意 AVG（移植物动静脉内瘘）字串含"内瘘"，故必须先判 AVG 再判 AVF。
func ClassifyAccess4(raw string) string {
	u := strings.ToUpper(raw)
	switch {
	case strings.Contains(u, "NCC") || strings.Contains(raw, "临时") || strings.Contains(raw, "无隧道"):
		return "NCC"
	case strings.Contains(u, "TCC") || strings.Contains(raw, "长期") || strings.Contains(raw, "带隧道") || strings.Contains(raw, "涤纶"):
		return "TCC"
	case strings.Contains(u, "AVG") || strings.Contains(raw, "移植") || strings.Contains(raw, "人工血管"):
		return "AVG"
	case strings.Contains(u, "AVF") || strings.Contains(raw, "内瘘") || strings.Contains(raw, "自体"):
		return "AVF"
	default:
		return ""
	}
}

// EvalVP 按通路类型 + 实时 BF 查分层表给静脉压分级。
// 通路无法识别 / 该通路无任何分层 → 返回 normal（不误报）；
// BF 不落任何区间 → 退到该通路最近 BF 档（兜底，不报错）。
func (m *MonitoringThresholds) EvalVP(accessRaw string, bf, vp float64) AlarmLevel {
	access := ClassifyAccess4(accessRaw)
	if access == "" {
		return AlarmNormal
	}
	st := m.matchVPStratum(access, bf)
	if st == nil {
		return AlarmNormal
	}
	if vp > st.DangerHigh {
		return AlarmDanger
	}
	if vp > st.WarnHigh || vp < st.NormalLow {
		return AlarmWarning
	}
	return AlarmNormal
}

// matchVPStratum 命中 [BFMin, BFMax) 区间；不命中则退到 BF 区间中点最近的档。
func (m *MonitoringThresholds) matchVPStratum(access string, bf float64) *VPStratum {
	var candidates []int
	for i := range m.VPReference {
		if m.VPReference[i].Access == access {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	for _, i := range candidates {
		if bf >= m.VPReference[i].BFMin && bf < m.VPReference[i].BFMax {
			return &m.VPReference[i]
		}
	}
	best := candidates[0]
	bestDist := math.MaxFloat64
	for _, i := range candidates {
		mid := (m.VPReference[i].BFMin + m.VPReference[i].BFMax) / 2
		if d := math.Abs(bf - mid); d < bestDist {
			bestDist = d
			best = i
		}
	}
	return &m.VPReference[best]
}

// WorstLevel 取多个分级中最严重者（danger > warning > normal）。
func WorstLevel(levels ...AlarmLevel) AlarmLevel {
	worst := AlarmNormal
	for _, l := range levels {
		if l == AlarmDanger {
			return AlarmDanger
		}
		if l == AlarmWarning {
			worst = AlarmWarning
		}
	}
	return worst
}
