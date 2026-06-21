package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/elliotxin/ai-hms-backend/internal/models"
)

type Exporter interface {
	Format() string
	Export(rep *models.CnrdsReport) (filename string, data []byte, err error)
}

var csvHeader = []string{
	"patientId", "name", "gender", "birthDate",
	"primaryDiagnosis", "comorbidity", "firstDialysisDate",
	"dialysisMode", "frequency", "vascularAccess",
	"hb", "ca", "p", "pth", "albumin", "ktv",
	"infMarkers",
	"outcomeType", "outcomeDate", "deathReason",
}

type CSVExporter struct{}

func (CSVExporter) Format() string { return "csv" }

func (CSVExporter) Export(rep *models.CnrdsReport) (string, []byte, error) {
	var content models.CnrdsContent
	if err := unmarshalContent(rep.Content, &content); err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)
	if err := w.Write(csvHeader); err != nil {
		return "", nil, err
	}

	for _, row := range content.Rows {
		record := []string{
			row.PatientID,
			row.Name,
			row.Gender,
			row.BirthDate,
			row.PrimaryDiagnosis,
			row.Comorbidity,
			row.FirstDialysisDate,
			row.DialysisMode,
			row.Frequency,
			row.VascularAccess,
			fstr(row.Hb),
			fstr(row.Ca),
			fstr(row.P),
			fstr(row.PTH),
			fstr(row.Albumin),
			fstr(row.KtV),
			row.InfMarkers,
			row.OutcomeType,
			row.OutcomeDate,
			row.DeathReason,
		}
		if err := w.Write(record); err != nil {
			return "", nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", nil, err
	}

	filename := fmt.Sprintf("CNRDS_%s_%s_%s.csv", rep.ReportType, rep.Period, rep.ID)
	return filename, buf.Bytes(), nil
}

func fstr(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}
