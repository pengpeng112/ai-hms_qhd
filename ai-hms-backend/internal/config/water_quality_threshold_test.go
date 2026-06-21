package config

import "testing"

func TestLoadWaterQualityThresholds(t *testing.T) {
	th, err := LoadWaterQualityThresholds()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	b, ok := th["bacteria"]
	if !ok || !b.Enabled || b.Max == nil || *b.Max != 100 {
		t.Fatalf("bacteria threshold wrong: %+v", b)
	}
	if fc, ok := th["free_chlorine"]; !ok || fc.Enabled {
		t.Fatalf("free_chlorine should be present but disabled: %+v", fc)
	}
	cond := th["conductivity"]
	if cond.LimitType != "range" || cond.Min == nil || *cond.Min != 13 || cond.Max == nil || *cond.Max != 15 {
		t.Fatalf("conductivity range wrong: %+v", cond)
	}
}
