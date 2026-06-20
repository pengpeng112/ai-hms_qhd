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
	// 纯数字
	num := []struct {
		in   string
		want float64
	}{
		{" 12.5 ", 12.5},
		{"12.5↑", 12.5},
		{"12.5 g/L", 12.5},
		{"<0.01", 0.01},
		{">120", 120},
		{"6.5%", 6.5},
		{"-3.2", -3.2},
	}
	for _, c := range num {
		v := parseFloatPtr(c.in)
		if v == nil || *v != c.want {
			t.Errorf("parse %q = %v, want %v", c.in, v, c.want)
		}
	}
	// 非数值 → nil
	for _, s := range []string{"阴性", "未见异常", "", "  "} {
		if v := parseFloatPtr(s); v != nil {
			t.Errorf("parse %q 应为 nil, got %v", s, *v)
		}
	}
}
