package config

import "testing"

func mustLoadThresholds(t *testing.T) *MonitoringThresholds {
	t.Helper()
	m, err := LoadMonitoringThresholds()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return m
}

func TestEvalFixed_MAP(t *testing.T) {
	m := mustLoadThresholds(t)
	cases := []struct {
		v    float64
		want AlarmLevel
	}{
		{55, AlarmDanger},   // < 60
		{65, AlarmWarning},  // 60–70
		{90, AlarmNormal},   // 70–110
		{115, AlarmWarning}, // 110–120
		{130, AlarmDanger},  // > 120
	}
	for _, c := range cases {
		if got := m.EvalFixed("map", c.v); got != c.want {
			t.Errorf("MAP %.0f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_HeartRate(t *testing.T) {
	m := mustLoadThresholds(t)
	cases := []struct {
		v    float64
		want AlarmLevel
	}{
		{45, AlarmDanger}, {55, AlarmWarning}, {80, AlarmNormal}, {105, AlarmWarning}, {120, AlarmDanger},
	}
	for _, c := range cases {
		if got := m.EvalFixed("heartRate", c.v); got != c.want {
			t.Errorf("HR %.0f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_UFR_HighOnly(t *testing.T) {
	m := mustLoadThresholds(t)
	// 超滤率只有上界：低值也是 normal。
	for _, c := range []struct {
		v    float64
		want AlarmLevel
	}{{0, AlarmNormal}, {8, AlarmNormal}, {11, AlarmWarning}, {14, AlarmDanger}} {
		if got := m.EvalFixed("ufr", c.v); got != c.want {
			t.Errorf("UFR %.0f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_DialysateNa(t *testing.T) {
	m := mustLoadThresholds(t)
	for _, c := range []struct {
		v    float64
		want AlarmLevel
	}{{128, AlarmDanger}, {133, AlarmWarning}, {140, AlarmNormal}, {147, AlarmWarning}, {150, AlarmDanger}} {
		if got := m.EvalFixed("dialysateNa", c.v); got != c.want {
			t.Errorf("Na %.0f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_SpO2_LowOnly(t *testing.T) {
	m := mustLoadThresholds(t)
	// SpO2 只有低侧：<90 危险、90–94 警戒、>=95 正常；高侧无界(100 仍正常)。
	for _, c := range []struct {
		v    float64
		want AlarmLevel
	}{{89, AlarmDanger}, {92, AlarmWarning}, {97, AlarmNormal}, {100, AlarmNormal}} {
		if got := m.EvalFixed("spo2", c.v); got != c.want {
			t.Errorf("SpO2 %.0f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_Respiration(t *testing.T) {
	m := mustLoadThresholds(t)
	for _, c := range []struct {
		v    float64
		want AlarmLevel
	}{{6, AlarmDanger}, {10, AlarmWarning}, {16, AlarmNormal}, {23, AlarmWarning}, {28, AlarmDanger}} {
		if got := m.EvalFixed("respiration", c.v); got != c.want {
			t.Errorf("RR %.0f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_TMP_HighOnly(t *testing.T) {
	m := mustLoadThresholds(t)
	for _, c := range []struct {
		v    float64
		want AlarmLevel
	}{{0, AlarmNormal}, {200, AlarmNormal}, {300, AlarmWarning}, {400, AlarmDanger}} {
		if got := m.EvalFixed("tmp", c.v); got != c.want {
			t.Errorf("TMP %.0f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_AP_HighOnly(t *testing.T) {
	m := mustLoadThresholds(t)
	// AP 阈值表存绝对值高侧两档(引擎喂 |AP|);此处直接喂正数即绝对值。
	for _, c := range []struct {
		v    float64
		want AlarmLevel
	}{{150, AlarmNormal}, {220, AlarmWarning}, {260, AlarmDanger}} {
		if got := m.EvalFixed("ap", c.v); got != c.want {
			t.Errorf("AP %.0f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_Conductivity(t *testing.T) {
	m := mustLoadThresholds(t)
	for _, c := range []struct {
		v    float64
		want AlarmLevel
	}{{12.8, AlarmDanger}, {13.2, AlarmWarning}, {14.0, AlarmNormal}, {14.8, AlarmWarning}, {15.3, AlarmDanger}} {
		if got := m.EvalFixed("conductivity", c.v); got != c.want {
			t.Errorf("Cond %.1f: got %s want %s", c.v, got, c.want)
		}
	}
}

func TestEvalFixed_UnknownKey(t *testing.T) {
	m := mustLoadThresholds(t)
	if got := m.EvalFixed("nope", 999); got != AlarmNormal {
		t.Errorf("unknown key should be normal, got %s", got)
	}
}

func TestClassifyAccess4(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"AVF", "AVF"},
		{"自体动静脉内瘘", "AVF"},
		{"AVG", "AVG"},
		{"移植物动静脉内瘘", "AVG"}, // 含"内瘘"但必须判为 AVG
		{"人工血管", "AVG"},
		{"TCC", "TCC"},
		{"带隧道带涤纶套导管", "TCC"},
		{"长期导管", "TCC"},
		{"NCC", "NCC"},
		{"无隧道临时导管", "NCC"},
		{"未知", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := ClassifyAccess4(c.raw); got != c.want {
			t.Errorf("ClassifyAccess4(%q): got %q want %q", c.raw, got, c.want)
		}
	}
}

func TestEvalVP_Stratified(t *testing.T) {
	m := mustLoadThresholds(t)
	// AVF, BF 200–250：normalLow 64 / warnHigh 145 / dangerHigh 157。
	if got := m.EvalVP("AVF", 220, 100); got != AlarmNormal {
		t.Errorf("AVF/220/VP100: got %s want normal", got)
	}
	if got := m.EvalVP("AVF", 220, 50); got != AlarmWarning {
		t.Errorf("AVF/220/VP50(<P10): got %s want warning", got)
	}
	if got := m.EvalVP("AVF", 220, 150); got != AlarmWarning {
		t.Errorf("AVF/220/VP150(>P90): got %s want warning", got)
	}
	if got := m.EvalVP("AVF", 220, 170); got != AlarmDanger {
		t.Errorf("AVF/220/VP170(>P95): got %s want danger", got)
	}
	// AVG 与 AVF 阈值不同：同 BF 200–250，AVG normalLow 108。VP100 在 AVG 应为警戒低。
	if got := m.EvalVP("AVG", 220, 100); got != AlarmWarning {
		t.Errorf("AVG/220/VP100(<P10=108): got %s want warning", got)
	}
}

func TestEvalVP_FallbackAndUnknown(t *testing.T) {
	m := mustLoadThresholds(t)
	// AVG 无 >250 档：BF 300 退到最近档（200–250），其 dangerHigh=154。
	if got := m.EvalVP("AVG", 300, 200); got != AlarmDanger {
		t.Errorf("AVG/300(fallback)/VP200: got %s want danger", got)
	}
	// 通路无法识别 → normal，不误报。
	if got := m.EvalVP("未知通路", 220, 999); got != AlarmNormal {
		t.Errorf("unknown access: got %s want normal", got)
	}
}

func TestWorstLevel(t *testing.T) {
	if WorstLevel(AlarmNormal, AlarmWarning, AlarmNormal) != AlarmWarning {
		t.Error("expected warning")
	}
	if WorstLevel(AlarmNormal, AlarmWarning, AlarmDanger) != AlarmDanger {
		t.Error("expected danger")
	}
	if WorstLevel(AlarmNormal, AlarmNormal) != AlarmNormal {
		t.Error("expected normal")
	}
}
