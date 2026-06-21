package actrs

import "testing"

func TestMapXrayToPatientACTR(t *testing.T) {
	ctr := 0.48
	x := XrayOut{ID: 9, CTR: &ctr, QCPass: true, ModelVersion: "v8.5", AnalysisDate: "2026-06-20T10:00:00"}
	rec := MapXrayToPatientACTR(x, 3, "1001", "D-001", "manual")
	if rec.ActrsXrayID != 9 || rec.QCPass != 1 || rec.PatientID != "1001" || rec.DialysisNo != "D-001" {
		t.Fatalf("bad map: %+v", rec)
	}
	if rec.AnalysisDate == nil {
		t.Fatalf("analysis_date should parse")
	}

	x.QCPass = false
	if MapXrayToPatientACTR(x, 3, "1001", "D-001", "manual").QCPass != 0 {
		t.Fatalf("qc_pass should be 0 when QCPass false")
	}
}
