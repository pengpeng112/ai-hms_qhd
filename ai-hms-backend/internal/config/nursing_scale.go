// Package config 护理量表评分规则（go:embed 内嵌 JSON，院方改配置重构建即生效，无运行时文件依赖）。
package config

import (
	_ "embed"
	"encoding/json"
)

//go:embed nursing_scale_rules.json
var nursingScaleRulesJSON []byte

// NursingScaleOption 量表条目的一个可选项
type NursingScaleOption struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

// NursingScaleItem 量表的一个条目
type NursingScaleItem struct {
	Key     string               `json:"key"`
	Label   string               `json:"label"`
	Options []NursingScaleOption `json:"options"`
}

// NursingScaleBand 分数→风险分档（[min,max] 闭区间，方向无关）
type NursingScaleBand struct {
	Min   int    `json:"min"`
	Max   int    `json:"max"`
	Level string `json:"level"` // high / moderate / low / none
	Label string `json:"label"`
}

// NursingScale 一张量表的完整定义
type NursingScale struct {
	ScaleType string             `json:"scaleType"`
	Name      string             `json:"name"`
	Enabled   bool               `json:"enabled"`
	Direction string             `json:"direction"` // higher_worse / lower_worse（仅说明，分档已直接给区间）
	Items     []NursingScaleItem `json:"items"`
	Bands     []NursingScaleBand `json:"bands"`
}

// LoadNursingScales 解析内嵌的护理量表规则
func LoadNursingScales() ([]NursingScale, error) {
	var out []NursingScale
	if err := json.Unmarshal(nursingScaleRulesJSON, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ScoreBand 在量表分档里定位分数所属档；找不到返回 nil
func (s NursingScale) ScoreBand(score int) *NursingScaleBand {
	for i := range s.Bands {
		b := s.Bands[i]
		if score >= b.Min && score <= b.Max {
			return &b
		}
	}
	return nil
}
