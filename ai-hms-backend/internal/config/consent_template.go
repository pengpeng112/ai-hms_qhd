// Package config 知情同意书模板清单（go:embed 内嵌 JSON，院方增删/调有效期重构建即生效）。
package config

import (
	_ "embed"
	"encoding/json"
)

//go:embed consent_templates.json
var consentTemplatesJSON []byte

// ConsentTemplate 一种同意书模板
type ConsentTemplate struct {
	ConsentType string `json:"consentType"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Timing      string `json:"timing"`
	ValidMonths int    `json:"validMonths"` // 默认有效期(月)，0=长期有效不设到期
	Enabled     bool   `json:"enabled"`
}

// LoadConsentTemplates 解析内嵌的同意书模板清单
func LoadConsentTemplates() ([]ConsentTemplate, error) {
	var out []ConsentTemplate
	if err := json.Unmarshal(consentTemplatesJSON, &out); err != nil {
		return nil, err
	}
	return out, nil
}
