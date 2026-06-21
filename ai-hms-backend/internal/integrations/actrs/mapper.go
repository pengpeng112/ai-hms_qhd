package actrs

import (
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
)

func MapXrayToPatientACTR(x XrayOut, tenantID int64, patientID, dialysisNo, source string) models.PatientACTR {
	qc := 0
	if x.QCPass {
		qc = 1
	}
	var analysisAt *time.Time
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, x.AnalysisDate); err == nil {
			analysisAt = &t
			break
		}
	}
	now := time.Now()
	return models.PatientACTR{
		TenantID:     tenantID,
		PatientID:    patientID,
		DialysisNo:   dialysisNo,
		ActrsXrayID:  x.ID,
		AnalysisDate: analysisAt,
		CTR:          x.CTR,
		ACTR:         x.ACTR,
		ACTR1:        x.ACTR1,
		ACTR2:        x.ACTR2,
		ACTRNorm:     x.ACTRNorm,
		HeartWidth:   x.HeartWidth,
		LungWidth:    x.LungWidth,
		TiltAngle:    x.TiltAngle,
		QCPass:       qc,
		QCPaAp:       x.QCPaAp,
		QCWarnings:   x.QCWarnings,
		ModelVersion: x.ModelVersion,
		Source:       source,
		ImagePath:    x.ImagePath,
		OverlayPath:  x.OverlayPath,
		MaskPath:     x.MaskPath,
		DoctorCorrection: x.DoctorCorrection,
		Notes:            x.Notes,
		SyncedAt:         &now,
	}
}
