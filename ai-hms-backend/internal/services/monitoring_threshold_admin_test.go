package services

import "testing"

func f64(v float64) *float64 { return &v }

func TestValidatePayload_OK(t *testing.T) {
	p := ThresholdAdminPayload{
		NaFactor: 9.9,
		Fixed: []FixedThresholdDTO{
			{MetricKey: "map", DangerLow: f64(60), WarnLow: f64(70), WarnHigh: f64(110), DangerHigh: f64(120)},
			{MetricKey: "ufr", WarnHigh: f64(10), DangerHigh: f64(13)},
		},
		VPReference: []VPStratumDTO{
			{Access: "AVF", BFMin: 0, BFMax: 150, NormalLow: 29, WarnHigh: 79, DangerHigh: 99, Enabled: true},
		},
	}
	if err := ValidatePayload(p); err != nil {
		t.Fatalf("期望通过，实际 %v", err)
	}
}

func TestValidatePayload_NonMonotonicFixed(t *testing.T) {
	p := ThresholdAdminPayload{NaFactor: 9.9, Fixed: []FixedThresholdDTO{
		{MetricKey: "map", DangerLow: f64(80), WarnLow: f64(70)},
	}}
	if err := ValidatePayload(p); err == nil {
		t.Fatal("期望单调性校验失败")
	}
}

func TestValidatePayload_BadVPAccess(t *testing.T) {
	p := ThresholdAdminPayload{NaFactor: 9.9, VPReference: []VPStratumDTO{
		{Access: "XYZ", BFMin: 0, BFMax: 150, NormalLow: 1, WarnHigh: 2, DangerHigh: 3},
	}}
	if err := ValidatePayload(p); err == nil {
		t.Fatal("期望通路类型校验失败")
	}
}

func TestValidatePayload_BadVPBand(t *testing.T) {
	p := ThresholdAdminPayload{NaFactor: 9.9, VPReference: []VPStratumDTO{
		{Access: "AVF", BFMin: 200, BFMax: 150, NormalLow: 1, WarnHigh: 2, DangerHigh: 3},
	}}
	if err := ValidatePayload(p); err == nil {
		t.Fatal("期望 BF 区间校验失败")
	}
}

func TestValidatePayload_BadNaFactor(t *testing.T) {
	if err := ValidatePayload(ThresholdAdminPayload{NaFactor: 0}); err == nil {
		t.Fatal("期望 naFactor<=0 校验失败")
	}
}

func TestValidatePayload_BadVPMonotonic(t *testing.T) {
	p := ThresholdAdminPayload{NaFactor: 9.9, VPReference: []VPStratumDTO{
		{Access: "AVF", BFMin: 0, BFMax: 150, NormalLow: 50, WarnHigh: 30, DangerHigh: 99},
	}}
	if err := ValidatePayload(p); err == nil {
		t.Fatal("期望 VP 三档单调校验失败")
	}
}
