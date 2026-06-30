package services

import (
	"testing"
	"time"
)

func fcPts(start time.Time, stepMin int, vals ...float64) []forecastPoint {
	out := make([]forecastPoint, 0, len(vals))
	for i, v := range vals {
		out = append(out, forecastPoint{T: start.Add(time.Duration(i*stepMin) * time.Minute), V: v})
	}
	return out
}

func TestForecastSeries_RisingClampedByEnvelope(t *testing.T) {
	start := time.Date(2026, 6, 28, 8, 0, 0, 0, time.UTC)
	actual := fcPts(start, 30, 100, 110, 120, 130)
	plannedEnd := start.Add(5 * time.Hour)
	pred := forecastSeries(actual, plannedEnd, forecastOpts{})
	if len(pred) < 2 {
		t.Fatalf("期望多个预测点，得 %d", len(pred))
	}
	if pred[0].V != 130 || !pred[0].T.Equal(actual[len(actual)-1].T) {
		t.Errorf("首点应桥接末actual(130@90min)，得 %+v", pred[0])
	}
	if pred[len(pred)-1].V <= pred[0].V {
		t.Errorf("升序应继续上升，末预测 %.1f 首 %.1f", pred[len(pred)-1].V, pred[0].V)
	}
	for _, p := range pred {
		if p.V > 145.0001 {
			t.Errorf("超出 envHi=145: %.2f", p.V)
		}
	}
}

func TestForecastSeries_FlatStaysFlat(t *testing.T) {
	start := time.Date(2026, 6, 28, 8, 0, 0, 0, time.UTC)
	actual := fcPts(start, 30, 140, 141, 139, 140)
	pred := forecastSeries(actual, start.Add(3*time.Hour), forecastOpts{})
	for _, p := range pred {
		if p.V < 135 || p.V > 145 {
			t.Errorf("平稳序列预测应近平(135-145)，得 %.2f", p.V)
		}
	}
}

func TestForecastSeries_ClinicalClampHi(t *testing.T) {
	start := time.Date(2026, 6, 28, 8, 0, 0, 0, time.UTC)
	actual := fcPts(start, 30, 100, 115, 130, 145)
	hi := 120.0
	pred := forecastSeries(actual, start.Add(4*time.Hour), forecastOpts{ClampHi: &hi})
	for _, p := range pred[1:] {
		if p.V > 120.0001 {
			t.Errorf("临床上限120应生效(投影点)，得 %.2f", p.V)
		}
	}
}

func TestForecastSeries_TooFewPoints(t *testing.T) {
	start := time.Date(2026, 6, 28, 8, 0, 0, 0, time.UTC)
	if got := forecastSeries(fcPts(start, 30, 100), start.Add(2*time.Hour), forecastOpts{}); got != nil {
		t.Errorf("单点应返回 nil，得 %+v", got)
	}
}

func TestForecastSeries_PlannedEndBeforeLast(t *testing.T) {
	start := time.Date(2026, 6, 28, 8, 0, 0, 0, time.UTC)
	actual := fcPts(start, 30, 100, 110, 120)
	if got := forecastSeries(actual, start.Add(10*time.Minute), forecastOpts{}); got != nil {
		t.Errorf("plannedEnd 早于末点应返回 nil，得 %+v", got)
	}
}
