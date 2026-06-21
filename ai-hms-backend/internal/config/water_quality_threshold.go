package config

import (
	_ "embed"
	"encoding/json"
)

//go:embed water_quality_thresholds.json
var waterQualityThresholdsJSON []byte

// WaterQualityThreshold 单个检测项阈值配置
type WaterQualityThreshold struct {
	Label         string   `json:"label"`
	Enabled       bool     `json:"enabled"`
	LimitType     string   `json:"limitType"` // max / range
	Max           *float64 `json:"max"`
	Min           *float64 `json:"min"`
	Intervention  *float64 `json:"intervention"`
	Unit          string   `json:"unit"`
	FrequencyDays int      `json:"frequencyDays"`
	SamplePoint   string   `json:"samplePoint"`
}

// LoadWaterQualityThresholds 解析内嵌阈值表（key=test_type）
func LoadWaterQualityThresholds() (map[string]WaterQualityThreshold, error) {
	out := map[string]WaterQualityThreshold{}
	if err := json.Unmarshal(waterQualityThresholdsJSON, &out); err != nil {
		return nil, err
	}
	return out, nil
}
