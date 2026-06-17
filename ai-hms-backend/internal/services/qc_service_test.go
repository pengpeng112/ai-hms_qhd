package services

import (
	"testing"

	"github.com/elliotxin/ai-hms-backend/internal/config"
)

func TestQCConceptMatches(t *testing.T) {
	concepts, err := config.LoadIndicatorConcepts()
	if err != nil {
		t.Fatalf("load concepts: %v", err)
	}
	byID := map[string]*config.IndicatorConcept{}
	for i := range concepts {
		byID[concepts[i].ConceptID] = &concepts[i]
	}

	cases := []struct {
		conceptID string
		code      string
		name      string
		want      bool
	}{
		{"HEMOGLOBIN", "HGB", "血红蛋白", true},
		{"HEMOGLOBIN", "", "血红蛋白测定", true},
		{"HEMOGLOBIN", "CA", "血钙", false},
		{"SERUM_CA", "CA", "血清钙", true},
		{"SERUM_P", "P", "血磷", true},
		{"IPTH", "PTH", "全段甲状旁腺激素", true},
		{"ALBUMIN", "ALB", "白蛋白", true},
		{"URR", "URR", "尿素清除率", true},
		{"KTV", "SPKTV", "Kt/V", true},
		{"SERUM_CA", "P", "血磷", false},
	}
	for _, tc := range cases {
		c := byID[tc.conceptID]
		if c == nil {
			t.Fatalf("缺概念 %s", tc.conceptID)
		}
		if got := conceptMatches(c, tc.code, tc.name); got != tc.want {
			t.Errorf("%s vs (code=%q name=%q): got %v want %v", tc.conceptID, tc.code, tc.name, got, tc.want)
		}
	}
}

func TestQCParseFloatPtr(t *testing.T) {
	if v := parseFloatPtr(" 12.5 "); v == nil || *v != 12.5 {
		t.Fatalf("parse 12.5 fail: %v", v)
	}
	if v := parseFloatPtr("阴性"); v != nil {
		t.Fatalf("non-numeric should be nil, got %v", *v)
	}
}
