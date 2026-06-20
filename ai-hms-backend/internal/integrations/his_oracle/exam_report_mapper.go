package his_oracle

import (
	"fmt"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
)

const validMinDate = "1900-01-01"

func MapHisExamToExamReport(row HisExamRow, legacyPatientID int64) models.ExamReport {
	examDate := resolveExamDate(row)
	title := resolveTitle(row)
	department := resolveDepartment(row)
	conclusion := buildConclusion(row)
	externalID := row.ExamNo

	return models.ExamReport{
		PatientID:        fmt.Sprintf("%d", legacyPatientID),
		ExamDate:         examDate,
		Title:            title,
		Conclusion:       conclusion,
		Department:       department,
		ExternalReportID: &externalID,
		SourceSystem:     models.SourceSystemHISOracleExam,
	}
}

func resolveExamDate(row HisExamRow) *time.Time {
	var minValid time.Time
	parsed, _ := time.Parse("2006-01-02", validMinDate)
	minValid = parsed

	candidates := []*time.Time{
		row.ReportTime,
		row.ReportDateTime,
		row.ExamDateTime,
		row.ReqDateTime,
	}
	for _, c := range candidates {
		if c != nil && !c.IsZero() && c.After(minValid) {
			return c
		}
	}
	return nil
}

func resolveTitle(row HisExamRow) string {
	if s := strings.TrimSpace(strPtr(row.ReportExamItems)); s != "" {
		return truncate(s, 200)
	}
	if s := strings.TrimSpace(strPtr(row.ItemNames)); s != "" {
		return truncate(s, 200)
	}
	if s := strings.TrimSpace(strPtr(row.ExamSubClass)); s != "" {
		return s
	}
	if s := strings.TrimSpace(strPtr(row.ExamClass)); s != "" {
		return s
	}
	return "检查报告"
}

func resolveDepartment(row HisExamRow) string {
	if s := strings.TrimSpace(strPtr(row.PerformedBy)); s != "" {
		return s
	}
	if s := strings.TrimSpace(strPtr(row.ReqDept)); s != "" {
		return s
	}
	return ""
}

func buildConclusion(row HisExamRow) string {
	var parts []string
	add := func(label, value string) {
		v := strings.TrimSpace(value)
		if v == "" {
			return
		}
		parts = append(parts, fmt.Sprintf("【%s】%s", label, v))
	}
	add("检查所见", strPtr(row.Description))
	add("印象", strPtr(row.Impression))
	add("诊断", strPtr(row.ExamDiag))
	add("建议", strPtr(row.Recommendation))
	add("备注", strPtr(row.Memo))
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n")
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
