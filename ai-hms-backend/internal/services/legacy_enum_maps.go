package services

// legacy_enum_maps.go
// 集中收纳“新血透枚举 → 老血透字段值”的双向映射，避免散落在多 service。
// 每个枚举：NewToLegacyXxx / LegacyToNewXxx 一对；未命中返回原值。

// PatientType: 新（门诊/住院）→ 老（10/20）
var patientTypeNewToLegacy = map[string]string{
	"门诊": "10",
	"住院": "20",
}

var patientTypeLegacyToNew = map[string]string{
	"10": "门诊",
	"20": "住院",
}

func MapPatientTypeNewToLegacy(v string) string {
	if mapped, ok := patientTypeNewToLegacy[v]; ok {
		return mapped
	}
	return v
}

func MapPatientTypeLegacyToNew(v string) string {
	if mapped, ok := patientTypeLegacyToNew[v]; ok {
		return mapped
	}
	return v
}

// TODO: 后续 Phase 补充 DialysisMode / OrderStatus / OrderType / ...
