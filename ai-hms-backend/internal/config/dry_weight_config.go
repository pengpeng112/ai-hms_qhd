package config

import _ "embed"

type DwConfigData struct {
	InductionMaxAdjustKg float64 `json:"induction_max_adjust_kg"`
	MaintenanceMaxAdjustKg float64 `json:"maintenance_max_adjust_kg"`
	InductionRNa float64 `json:"induction_rna"`
	MaintenanceRNa float64 `json:"maintenance_rna"`
	CTRThreshold float64 `json:"ctr_threshold"`
	SBPCriteria int `json:"sbp_criteria"`
	DBPCriteria int `json:"dbp_criteria"`
	HeartRateLow int `json:"heart_rate_low"`
	HeartRateHigh int `json:"heart_rate_high"`
}

var DwConfig DwConfigData

func init() {
	DwConfig = DwConfigData{
		InductionMaxAdjustKg:   1.0,
		MaintenanceMaxAdjustKg: 0.5,
		InductionRNa:           1.05,
		MaintenanceRNa:         1.025,
		CTRThreshold:           0.52,
		SBPCriteria:            110,
		DBPCriteria:            60,
		HeartRateLow:           60,
		HeartRateHigh:          100,
	}
}
