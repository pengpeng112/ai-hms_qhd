package services

import "github.com/elliotxin/ai-hms-backend/internal/models"

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

// PatientShiftStatus: 新 -> 老
// 老库 Schedule_PatientShift.Status 语义（排班管理.md）：
//   10 草稿 / 20 已确认 / 30 用户确认 / 40 用户取消 / 50 排班取消 / 60 转出人员
var patientShiftStatusNewToLegacy = map[int]int{
	0: 10, // 待执行 -> 草稿
	1: 20, // 已确认
	2: 20, // 进行中 -> 已确认
	3: 30, // 已完成 -> 用户确认
	4: 50, // 系统取消 -> 排班取消
	5: 40, // 用户取消
	6: 60, // 转出人员
}

var patientShiftStatusLegacyToNew = map[int]int{
	10: 0, // 草稿
	20: 1, // 已确认
	30: 3, // 用户确认 -> 已完成
	40: 5, // 用户取消
	50: 4, // 排班取消 -> 已取消
	60: 6, // 转出人员
}

func MapPatientShiftStatusNewToLegacy(v int) int {
	if mapped, ok := patientShiftStatusNewToLegacy[v]; ok {
		return mapped
	}
	// 标准化状态 → 老库: 10=草稿, 20=已确认, 50=透析中, 60=已完成, 70=已取消, 80=缺席
	switch v {
	case models.StdStatusDraft:
		return 10
	case models.StdStatusConfirmed:
		return 20
	case models.StdStatusInDialysis:
		return 20
	case models.StdStatusCompleted:
		return 30
	case models.StdStatusCancelled, models.StdStatusAbsent:
		return 50
	}
	return v
}

func MapPatientShiftStatusLegacyToNew(v int) int {
	if mapped, ok := patientShiftStatusLegacyToNew[v]; ok {
		return mapped
	}
	return v
}

// MapPatientShiftStandardStatusToLegacy 标准化状态 → 老库状态
func MapPatientShiftStandardStatusToLegacy(v int) int {
	switch v {
	case models.StdStatusPending, models.StdStatusDraft:
		return 10
	case models.StdStatusConfirmed:
		return 20
	case models.StdStatusInDialysis:
		return 20
	case models.StdStatusCompleted:
		return 30
	case models.StdStatusCancelled:
		return 50
	case models.StdStatusAbsent:
		return 50 // 缺席映射到排班取消
	}
	return v
}

// MapPatientShiftLegacyToStandardStatus 老库状态 → 标准化状态
func MapPatientShiftLegacyToStandardStatus(v int) int {
	switch v {
	case 10:
		return models.StdStatusDraft
	case 20:
		return models.StdStatusConfirmed
	case 30:
		return models.StdStatusCompleted
	case 40, 50:
		return models.StdStatusCancelled
	case 60:
		return models.StdStatusCancelled
	}
	return v
}

// TODO: 后续 Phase 补充 DialysisMode / OrderStatus / OrderType / ...
