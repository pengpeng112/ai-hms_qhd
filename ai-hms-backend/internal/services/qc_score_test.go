package services

import "testing"

func fp(v float64) *float64 { return &v }

func TestQCIndicatorThresholds(t *testing.T) {
	if sc, ok := scoreBloodPressure(fp(150), fp(90)); sc != 10 || !ok {
		t.Errorf("BP 150/90 -> %d,%v want 10,true", sc, ok)
	}
	if sc, _ := scoreBloodPressure(fp(160), fp(95)); sc != 5 {
		t.Errorf("BP 160/95 -> %d want 5", sc)
	}
	if sc, _ := scoreBloodPressure(fp(181), fp(80)); sc != -5 {
		t.Errorf("BP 181/80 -> %d want -5", sc)
	}
	if sc, ok := scoreHeartRate(fp(72)); sc != 5 || !ok {
		t.Errorf("HR72 -> %d,%v want 5,true", sc, ok)
	}
	if sc, _ := scoreHeartRate(fp(110)); sc != 0 {
		t.Errorf("HR110 -> %d want 0", sc)
	}
	if sc, ok := scoreAdequacy(fp(72), nil); sc != 10 || !ok {
		t.Errorf("URR72 -> %d,%v want 10,true", sc, ok)
	}
	if sc, _ := scoreAdequacy(fp(55), fp(1.25)); sc != 5 {
		t.Errorf("URR55/KtV1.25 -> %d want 5", sc)
	}
	if sc, _ := scoreAdequacy(fp(40), fp(1.0)); sc != -5 {
		t.Errorf("URR40/KtV1.0 -> %d want -5", sc)
	}
	if sc, ok := scoreFluid(fp(3.5)); sc != 5 || !ok {
		t.Errorf("TUF3.5 -> %d,%v want 5,true", sc, ok)
	}
	if sc, _ := scoreFluid(fp(6.5)); sc != -3 {
		t.Errorf("TUF6.5 -> %d want -3", sc)
	}
	if sc, ok := scoreAnemia(fp(120)); sc != 10 || !ok {
		t.Errorf("Hb120 -> %d,%v want 10,true", sc, ok)
	}
	if sc, ok := scoreNutrition(fp(45)); sc != 10 || !ok {
		t.Errorf("Alb45 -> %d", sc)
	}
	if sc, ok := scoreCalcium(fp(2.3)); sc != 5 || !ok {
		t.Errorf("Ca2.3 -> %d", sc)
	}
	if sc, ok := scorePhosphorus(fp(1.5)); sc != 5 || !ok {
		t.Errorf("P1.5 -> %d", sc)
	}
	if sc, ok := scorePTH(fp(200)); sc != 5 || !ok {
		t.Errorf("PTH200 -> %d", sc)
	}
}

func TestQCMissingData(t *testing.T) {
	p := ScorePatient(QCInput{})
	if p.Quality != 0 {
		t.Fatalf("全缺数据质量分应 0, got %d", p.Quality)
	}
	if p.Total != 20 {
		t.Fatalf("全缺数据总分应=基础20, got %d", p.Total)
	}
	for k, ok := range p.OnTarget {
		if ok {
			t.Fatalf("缺数据 %s 不应达标", k)
		}
	}
}

func TestQCFullScorePatient(t *testing.T) {
	p := ScorePatient(QCInput{
		AvgSP: fp(130), AvgDP: fp(80), AvgHR: fp(72),
		CTRPre: fp(50), URR: fp(75), TUFPercent: fp(3),
		Hb: fp(120), Alb: fp(45), Ca: fp(2.3), P: fp(1.5), PTH: fp(200),
	})
	if p.Quality != 75 {
		t.Fatalf("满分质量分应 75, got %d (items=%v)", p.Quality, p.Items)
	}
	if p.Total != 95 {
		t.Fatalf("满分总分应 95, got %d", p.Total)
	}
	for k, ok := range p.OnTarget {
		if !ok {
			t.Fatalf("满分病人 %s 应达标", k)
		}
	}
}

func TestQCAggregateDoctor(t *testing.T) {
	full := ScorePatient(QCInput{AvgSP: fp(130), AvgDP: fp(80), AvgHR: fp(72), CTRPre: fp(50), URR: fp(75), TUFPercent: fp(3), Hb: fp(120), Alb: fp(45), Ca: fp(2.3), P: fp(1.5), PTH: fp(200)})
	mixed := ScorePatient(QCInput{AvgSP: fp(170), AvgDP: fp(98), AvgHR: fp(72), CTRPre: fp(50), URR: fp(75), TUFPercent: fp(3), Hb: fp(120), Alb: fp(45), Ca: fp(2.3), P: fp(1.5), PTH: fp(200)})

	d := AggregateDoctor("9001", []QCPatientScore{full, mixed})
	if d.PatientCount != 2 || d.QuantityScore != 40 {
		t.Fatalf("数量分应 40(2x20), got %d/%d", d.PatientCount, d.QuantityScore)
	}
	if d.QualityScore != 140 || d.TotalScore != 180 {
		t.Fatalf("质量分应140 总分180, got q=%d t=%d", d.QualityScore, d.TotalScore)
	}
	if d.OnTargetRate["bloodPressure"] != 0.5 {
		t.Fatalf("血压达标率应 0.5, got %v", d.OnTargetRate["bloodPressure"])
	}
	if d.OnTargetRate["heartRate"] != 1.0 {
		t.Fatalf("心率达标率应 1.0, got %v", d.OnTargetRate["heartRate"])
	}
}
