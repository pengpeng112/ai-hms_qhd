package actrs

import "testing"

func TestMapXrayToPatientACTR(t *testing.T) {
	ctr := 0.48
	one := 1
	x := XrayOut{ID: 9, CTR: &ctr, QCPass: &one, ModelVersion: "v8.5", AnalysisDate: "2026-06-20T10:00:00"}
	rec := MapXrayToPatientACTR(x, 3, "1001", "D-001", "manual")
	if rec.ActrsXrayID != 9 || rec.QCPass != 1 || rec.PatientID != "1001" || rec.DialysisNo != "D-001" {
		t.Fatalf("bad map: %+v", rec)
	}
	if rec.AnalysisDate == nil {
		t.Fatalf("analysis_date should parse")
	}

	zero := 0
	if MapXrayToPatientACTR(XrayOut{QCPass: &zero}, 3, "1001", "D-001", "manual").QCPass != 0 {
		t.Fatalf("qc_pass should be 0 when *QCPass=0")
	}

	if MapXrayToPatientACTR(XrayOut{QCPass: nil}, 3, "1001", "D-001", "manual").QCPass != 0 {
		t.Fatalf("qc_pass should be 0 when QCPass nil")
	}
}

func TestMapXrayToPatientACTR_V85Fields(t *testing.T) {
	ctr := 0.48
	actr1 := 0.45
	actr2 := 0.50
	hw := 120
	lw := 250
	ta := 1.5
	qc := 1
	x := XrayOut{
		ID: 9, CTR: &ctr, QCPass: &qc,
		ACTR1: &actr1, ACTR2: &actr2,
		HeartWidth: &hw, LungWidth: &lw, TiltAngle: &ta,
		ModelVersion: "v8.5", MaskPath: "/m/mask.png",
	}
	rec := MapXrayToPatientACTR(x, 3, "1001", "D-001", "manual")
	if rec.ACTR1 == nil || *rec.ACTR1 != 0.45 {
		t.Fatalf("actr1 not mapped: %+v", rec.ACTR1)
	}
	if rec.ACTR2 == nil || *rec.ACTR2 != 0.50 {
		t.Fatalf("actr2 not mapped: %+v", rec.ACTR2)
	}
	if rec.HeartWidth == nil || *rec.HeartWidth != 120 {
		t.Fatalf("heart_width not mapped: %+v", rec.HeartWidth)
	}
	if rec.LungWidth == nil || *rec.LungWidth != 250 {
		t.Fatalf("lung_width not mapped: %+v", rec.LungWidth)
	}
	if rec.TiltAngle == nil || *rec.TiltAngle != 1.5 {
		t.Fatalf("tilt_angle not mapped: %+v", rec.TiltAngle)
	}
	if rec.MaskPath != "/m/mask.png" {
		t.Fatalf("mask_path not mapped: %s", rec.MaskPath)
	}
}
