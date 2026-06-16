// Package config 提供临床指标映射表的加载（go:embed 内嵌 JSON，无运行时文件依赖）。
package config

import (
	_ "embed"
	"encoding/json"
)

//go:embed clinical_indicator_mapping.json
var clinicalIndicatorMappingJSON []byte

// IndicatorConcept 单个临床指标概念（对应 clinical_indicator_mapping.json 一条）
type IndicatorConcept struct {
	ConceptID      string   `json:"conceptId"`
	ConceptNameZh  string   `json:"conceptNameZh"`
	ConceptNameEn  string   `json:"conceptNameEn"`
	Category       string   `json:"category"`
	NameKeywords   []string `json:"nameKeywords"`
	ItemCodeHints  []string `json:"itemCodeHints"`
	IndexCodeHints []string `json:"indexCodeHints"`
	Loinc          []string `json:"loinc"`
	SourceTables   []string `json:"sourceTables"`
	Unit           string   `json:"unit"`
	Priority       int      `json:"priority"`
}

// LoadIndicatorConcepts 解析内嵌的指标映射表
func LoadIndicatorConcepts() ([]IndicatorConcept, error) {
	var out []IndicatorConcept
	if err := json.Unmarshal(clinicalIndicatorMappingJSON, &out); err != nil {
		return nil, err
	}
	return out, nil
}
