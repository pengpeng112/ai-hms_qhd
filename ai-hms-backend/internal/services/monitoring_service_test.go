package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
)

func TestExtrapolateVitals(t *testing.T) {
	base := time.Date(2026, 6, 25, 8, 0, 0, 0, time.UTC)
	actual := []VitalSample{
		{T: base, SBP: 130, DBP: 80, MAP: 96.7, HR: 78, Kind: "actual"},
		{T: base.Add(30 * time.Minute), SBP: 128, DBP: 80, MAP: 96, HR: 80, Kind: "actual"},
		{T: base.Add(60 * time.Minute), SBP: 124, DBP: 78, MAP: 93.3, HR: 82, Kind: "actual"},
	}
	plannedEnd := base.Add(4 * time.Hour)
	pred := extrapolateVitals(actual, plannedEnd)
	if len(pred) < 2 {
		t.Fatalf("expected predicted points, got %d", len(pred))
	}
	// 首点 = 桥接（= 末个 actual 的时间/值），kind=predicted。
	if !pred[0].T.Equal(actual[2].T) || pred[0].Kind != "predicted" || pred[0].MAP != actual[2].MAP {
		t.Errorf("bridge point wrong: %+v", pred[0])
	}
	// 所有预测点 kind=predicted，且不超出计划下机、在生理夹紧范围内。
	for _, p := range pred {
		if p.Kind != "predicted" {
			t.Errorf("non-predicted kind: %+v", p)
		}
		if p.T.After(plannedEnd) {
			t.Errorf("point past plannedEnd: %v", p.T)
		}
		if p.SBP > 0 && (p.SBP < 50 || p.SBP > 220) {
			t.Errorf("SBP out of clamp: %v", p.SBP)
		}
	}
}

func TestExtrapolateVitals_TooShort(t *testing.T) {
	if extrapolateVitals(nil, time.Now().Add(time.Hour)) != nil {
		t.Error("nil input should yield nil")
	}
	one := []VitalSample{{T: time.Now(), MAP: 90, Kind: "actual"}}
	if extrapolateVitals(one, time.Now().Add(time.Hour)) != nil {
		t.Error("single point should yield nil")
	}
}

func TestDialysisDurationMinutes(t *testing.T) {
	cases := map[float64]float64{
		4:   240, // 小时→分钟
		3:   180,
		3.5: 210,
		8:   480,
		0:   240, // 缺失取默认 4h
		-1:  240,
		240: 240, // 已是分钟（>24）
		300: 300,
	}
	for in, want := range cases {
		if got := dialysisDurationMinutes(in); got != want {
			t.Errorf("dialysisDurationMinutes(%v)=%v, want %v", in, got, want)
		}
	}
}

func TestMlToL(t *testing.T) {
	cases := map[float64]float64{
		3000:  3,
		2000:  2,
		0:     0,
		-100:  0,
		500:   0.5,
		57631: 57.631,
	}
	for in, want := range cases {
		if got := mlToL(in); got != want {
			t.Errorf("mlToL(%v)=%v, want %v", in, got, want)
		}
	}
}

func TestEvalAlarms(t *testing.T) {
	th, err := config.LoadMonitoringThresholds()
	if err != nil {
		t.Fatalf("load thresholds: %v", err)
	}
	s := &MonitoringService{}

	t.Run("all normal", func(t *testing.T) {
		d := &MonitoringLiveDevice{
			SBP: 120, DBP: 80, HeartRate: 80, // MAP≈93
			AccessType: "AVF", BF: 220, VenousPressure: 100, // AVF 200-250 正常 64–145
			Conductivity: 14,                                         // Na≈138.6 正常
			UFGoal:       2.0, DryWeight: 60, EstimatedDuration: 240, // UFR≈8.3 正常
		}
		s.evalAlarms(d, th)
		if d.AlarmLevel != string(config.AlarmNormal) {
			t.Errorf("AlarmLevel=%s, alerts=%+v; want normal", d.AlarmLevel, d.Alerts)
		}
		if len(d.Alerts) != 0 {
			t.Errorf("want no alerts, got %+v", d.Alerts)
		}
	})

	t.Run("MAP danger dominates", func(t *testing.T) {
		d := &MonitoringLiveDevice{SBP: 80, DBP: 40, HeartRate: 80} // MAP≈53 <60 危险
		s.evalAlarms(d, th)
		if d.AlarmLevel != string(config.AlarmDanger) {
			t.Errorf("want danger, got %s", d.AlarmLevel)
		}
		if len(d.Alerts) != 1 || d.Alerts[0].Metric != "map" || d.Alerts[0].Level != "danger" {
			t.Errorf("want one map danger alert, got %+v", d.Alerts)
		}
	})

	t.Run("VP stratified danger", func(t *testing.T) {
		d := &MonitoringLiveDevice{AccessType: "AVF", BF: 220, VenousPressure: 170} // >P95(157) 危险
		s.evalAlarms(d, th)
		if d.AlarmLevel != string(config.AlarmDanger) {
			t.Errorf("want danger, got %s (alerts=%+v)", d.AlarmLevel, d.Alerts)
		}
	})

	t.Run("UFR warning", func(t *testing.T) {
		d := &MonitoringLiveDevice{UFGoal: 3.0, DryWeight: 60, EstimatedDuration: 240} // 3000/60/4=12.5 警戒
		s.evalAlarms(d, th)
		if d.AlarmLevel != string(config.AlarmWarning) {
			t.Errorf("want warning, got %s (alerts=%+v)", d.AlarmLevel, d.Alerts)
		}
	})

	t.Run("missing readings skipped", func(t *testing.T) {
		d := &MonitoringLiveDevice{} // 全 0 → 无可评估指标
		s.evalAlarms(d, th)
		if d.AlarmLevel != string(config.AlarmNormal) || len(d.Alerts) != 0 {
			t.Errorf("empty device should be normal/no-alerts, got %s %+v", d.AlarmLevel, d.Alerts)
		}
	})

	t.Run("nil thresholds no panic", func(t *testing.T) {
		d := &MonitoringLiveDevice{SBP: 80, DBP: 40}
		s.evalAlarms(d, nil)
		if d.AlarmLevel != "" || len(d.Alerts) != 0 {
			t.Errorf("nil thresholds should leave fields untouched, got %s %+v", d.AlarmLevel, d.Alerts)
		}
	})
}
