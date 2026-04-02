package hdis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	labDeptKeywords = []string{
		"检验", "输血",
	}
	examDeptKeywords = []string{
		"放射", "超声", "功能", "病理", "内镜",
	}
	labNameKeywords = []string{
		"血", "尿", "生化", "凝血", "病毒", "hiv", "丙肝", "梅毒",
	}
	examNameKeywords = []string{
		"x-ray", "ct", "mri", "超声", "彩超", "心电", "ecg", "影像",
	}
)

// LabReportPayload 报告头映射结果
type LabReportPayload struct {
	ExternalReportID  string
	ReportNo          string
	ItemCode          string
	ItemName          string
	ClinicalDiagnosis string
	SpecimenType      string
	Urgency           string
	RequestDoctor     string
	RequestedAt       *time.Time
	SampledAt         *time.Time
	ReceivedAt        *time.Time
	ReportedAt        *time.Time
	Status            string
	SyncedAt          *time.Time
}

// LabReportItemPayload 报告明细映射结果
type LabReportItemPayload struct {
	ItemCode       string
	ItemName       string
	ResultValue    string
	Unit           string
	ReferenceRange string
	AbnormalFlag   string
}

// KeyIndicatorPayload 关键指标映射结果
type KeyIndicatorPayload struct {
	ExternalRecordID string
	IndexName        string
	IndexCode        string
	Result           string
	Unit             string
	Reference        string
	ResultSign       string
	TestedAt         *time.Time
	EvaluationResult string
	SyncedAt         *time.Time
}

// ExamReportPayload 检查报告头映射结果
type ExamReportPayload struct {
	ExternalReportID string
	Title            string
	Department       string
	Conclusion       string
	ExamDate         *time.Time
	SyncedAt         *time.Time
}

// ParseWebcmdExaminationList 解析 webcmd 的 Data 字段
func ParseWebcmdExaminationList(raw json.RawMessage) ([]WebcmdExamination, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		return []WebcmdExamination{}, nil
	}

	var groups []WebcmdGroup
	if err := json.Unmarshal(raw, &groups); err == nil {
		exams := make([]WebcmdExamination, 0)
		for _, group := range groups {
			exams = append(exams, group.GroupValue...)
		}
		return exams, nil
	}

	var exams []WebcmdExamination
	if err := json.Unmarshal(raw, &exams); err == nil {
		return exams, nil
	}

	var singleGroup WebcmdGroup
	if err := json.Unmarshal(raw, &singleGroup); err == nil {
		return singleGroup.GroupValue, nil
	}

	return nil, fmt.Errorf("unsupported webcmd data format: %s", trimmed)
}

// MapWebcmdToLabReportPayload 报告头字段映射
func MapWebcmdToLabReportPayload(exam WebcmdExamination) LabReportPayload {
	reportedAt := parseOptionalTime(exam.ResultRPTTime)
	if reportedAt == nil {
		reportedAt = parseOptionalTime(exam.ResultTime)
	}

	syncedAt := parseOptionalTime(exam.LastModifyTime)
	if syncedAt == nil {
		syncedAt = parseOptionalTime(exam.SyncTime)
	}
	if syncedAt == nil {
		now := time.Now()
		syncedAt = &now
	}

	return LabReportPayload{
		ExternalReportID:  strconv.Itoa(exam.ID),
		ReportNo:          strings.TrimSpace(exam.TestNO),
		ItemCode:          strings.TrimSpace(exam.Type),
		ItemName:          strings.TrimSpace(exam.Name),
		ClinicalDiagnosis: strings.TrimSpace(normalizeString(exam.ClinicalDiagnosisDesc)),
		SpecimenType:      strings.TrimSpace(exam.Specimen),
		Urgency:           mapUrgency(exam.Priority),
		RequestDoctor:     strings.TrimSpace(exam.ApplyUserName),
		RequestedAt:       parseOptionalTime(exam.ApplyTime),
		SampledAt:         parseOptionalTime(exam.SpecimenSampleTime),
		ReceivedAt:        parseOptionalTime(exam.SpecimenReceivedTime),
		ReportedAt:        reportedAt,
		Status:            mapResultStatus(exam.ResultStatus),
		SyncedAt:          syncedAt,
	}
}

// MapGraphQLItemsToLabReportItems 明细字段映射
func MapGraphQLItemsToLabReportItems(items []GraphQLExaminationItem) []LabReportItemPayload {
	mapped := make([]LabReportItemPayload, 0, len(items))

	for _, item := range items {
		referenceRange := firstNonEmpty(
			normalizeString(item.Reference),
			normalizeString(item.RefRange),
		)
		abnormal := firstNonEmpty(
			normalizeString(item.ResultSign),
			normalizeString(item.AbnormalFlag),
		)

		mapped = append(mapped, LabReportItemPayload{
			ItemCode:       strings.TrimSpace(normalizeString(item.ItemCode)),
			ItemName:       strings.TrimSpace(normalizeString(item.ItemName)),
			ResultValue:    strings.TrimSpace(normalizeString(item.Result)),
			Unit:           strings.TrimSpace(normalizeString(item.Unit)),
			ReferenceRange: strings.TrimSpace(referenceRange),
			AbnormalFlag:   normalizeAbnormalFlag(abnormal),
		})
	}

	return mapped
}

// MapGraphQLRecordToKeyIndicatorPayload Record 字段映射
func MapGraphQLRecordToKeyIndicatorPayload(record GraphQLRecord) KeyIndicatorPayload {
	now := time.Now()
	return KeyIndicatorPayload{
		ExternalRecordID: strings.TrimSpace(normalizeString(record.ID)),
		IndexName:        strings.TrimSpace(normalizeString(record.IndexName)),
		IndexCode:        strings.TrimSpace(normalizeString(record.IndexCode)),
		Result:           strings.TrimSpace(normalizeString(record.Result)),
		Unit:             strings.TrimSpace(normalizeString(record.Unit)),
		Reference:        strings.TrimSpace(normalizeString(record.Reference)),
		ResultSign:       strings.TrimSpace(normalizeString(record.ResultSign)),
		TestedAt:         parseOptionalTime(normalizeString(record.TestTime)),
		EvaluationResult: strings.TrimSpace(normalizeString(record.EvaluationResult)),
		SyncedAt:         &now,
	}
}

// MapGraphQLExaminationToExamReportPayload Examination 报告头字段映射
func MapGraphQLExaminationToExamReportPayload(exam GraphQLExamination) ExamReportPayload {
	syncedAt := parseOptionalTime(normalizeString(exam.LastModifyTime))
	if syncedAt == nil {
		now := time.Now()
		syncedAt = &now
	}

	examDate := parseOptionalTime(normalizeString(exam.ResultTime))
	if examDate == nil {
		examDate = syncedAt
	}

	title := strings.TrimSpace(normalizeString(exam.Name))
	if title == "" {
		title = "-"
	}

	return ExamReportPayload{
		ExternalReportID: strings.TrimSpace(normalizeString(exam.ID)),
		Title:            title,
		Department:       "检查报告",
		Conclusion:       "",
		ExamDate:         examDate,
		SyncedAt:         syncedAt,
	}
}

// ClassifyExamOrLab 检查/检验分流
func ClassifyExamOrLab(exam WebcmdExamination) DataCategory {
	dealDept := strings.TrimSpace(strings.ToLower(exam.DealDept))
	if containsAny(dealDept, labDeptKeywords) {
		return DataCategoryLab
	}
	if containsAny(dealDept, examDeptKeywords) {
		return DataCategoryExam
	}

	typeCode := strings.TrimSpace(strings.ToUpper(exam.Type))
	if strings.HasPrefix(typeCode, "JY") {
		return DataCategoryLab
	}

	name := strings.TrimSpace(strings.ToLower(exam.Name))
	if containsAny(name, examNameKeywords) {
		return DataCategoryExam
	}
	if containsAny(name, labNameKeywords) {
		return DataCategoryLab
	}

	return DataCategoryLab
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func containsAny(source string, keywords []string) bool {
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		if strings.Contains(source, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func normalizeString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func normalizeAbnormalFlag(flag string) string {
	normalized := strings.ToUpper(strings.TrimSpace(flag))
	switch normalized {
	case "H", "L":
		return normalized
	default:
		return "N"
	}
}

func mapUrgency(priority *int) string {
	if priority != nil && *priority == 1 {
		return "急诊"
	}
	return "常规"
}

func mapResultStatus(status *int) string {
	if status == nil {
		return "已出报告"
	}

	switch *status {
	case 1:
		return "待检"
	case 2:
		return "检验中"
	case 3:
		return "已完成"
	case 4:
		return "已出报告"
	default:
		return strconv.Itoa(*status)
	}
}

func parseOptionalTime(value string) *time.Time {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return &parsed
		}
	}

	return nil
}
