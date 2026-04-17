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

// PatientShiftStatus: 新（0/1/2/3/4）→ 老（10/20/30/40/50）
var patientShiftStatusNewToLegacy = map[int]int{
	0: 10, // 待执行 -> 待确认
	1: 20, // 已确认
	2: 30, // 进行中
	3: 40, // 已完成
	4: 50, // 已取消
}

var patientShiftStatusLegacyToNew = map[int]int{
	10: 0,
	20: 1,
	30: 2,
	40: 3,
	50: 4,
}

func MapPatientShiftStatusNewToLegacy(v int) int {
	if mapped, ok := patientShiftStatusNewToLegacy[v]; ok {
		return mapped
	}
	return v
}

func MapPatientShiftStatusLegacyToNew(v int) int {
	if mapped, ok := patientShiftStatusLegacyToNew[v]; ok {
		return mapped
	}
	return v
}

// TODO: 后续 Phase 补充 DialysisMode / OrderStatus / OrderType / ...
