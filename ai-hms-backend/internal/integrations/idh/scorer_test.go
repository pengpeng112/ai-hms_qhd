package idh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStubScorer(t *testing.T) {
	r := StubScorer{}.Score(context.Background(), RiskInput{TreatmentID: 1})
	if r.Available {
		t.Errorf("stub should be unavailable, got %+v", r)
	}
}

func TestLevelFromProbability(t *testing.T) {
	cases := []struct {
		p    float64
		want string
	}{{0.0, "low"}, {0.1, "low"}, {0.3, "medium"}, {0.5, "high"}, {0.9, "high"}}
	for _, c := range cases {
		if got := LevelFromProbability(c.p); got != c.want {
			t.Errorf("p=%.2f: got %s want %s", c.p, got, c.want)
		}
	}
}

func TestLevelFromProbabilityWithCuts(t *testing.T) {
	cases := []struct {
		p, high, medium float64
		want            string
	}{
		{0.0, 0.5, 0.2, "low"},
		{0.3, 0.8, 0.4, "low"},
		{0.5, 0.8, 0.4, "medium"},
		{0.85, 0.8, 0.4, "high"},
	}
	for _, c := range cases {
		if got := LevelFromProbabilityWithCuts(c.p, c.high, c.medium); got != c.want {
			t.Errorf("p=%.2f high=%.2f medium=%.2f: got %s want %s", c.p, c.high, c.medium, got, c.want)
		}
	}
}

func TestHTTPScorer_EmptyWindowUnavailable(t *testing.T) {
	h := NewHTTPScorer(Config{BaseURL: "http://example.invalid"})
	if r := h.Score(context.Background(), RiskInput{TreatmentID: 1}); r.Available {
		t.Errorf("empty window should be unavailable, got %+v", r)
	}
}

func TestHTTPScorer_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/idh/score" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: true, Probability: 0.73})
	}))
	defer srv.Close()

	h := NewHTTPScorer(Config{BaseURL: srv.URL})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if !r.Available || r.Probability != 0.73 || r.Level != "high" {
		t.Errorf("unexpected result: %+v", r)
	}
}

func TestHTTPScorer_CustomCutoffs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/idh/score" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: true, Probability: 0.45})
	}))
	defer srv.Close()

	// 默认切点: 0.45<0.5 → medium(>=0.2)
	h := NewHTTPScorer(Config{BaseURL: srv.URL})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r.Level != "medium" {
		t.Errorf("default cuts: want medium, got %s", r.Level)
	}

	// 自定义切点: high=0.4 medium=0.3 → 0.45>=0.4 high
	h2 := NewHTTPScorer(Config{BaseURL: srv.URL, LevelHigh: 0.4, LevelMedium: 0.3})
	r2 := h2.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r2.Level != "high" {
		t.Errorf("custom cuts high=0.4: want high, got %s", r2.Level)
	}
}

func TestNewHTTPScorer_DefaultCuts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: true, Probability: 0.55})
	}))
	defer srv.Close()

	// 零值→默认 0.5/0.2
	h := NewHTTPScorer(Config{BaseURL: srv.URL})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r.Level != "high" {
		t.Errorf("default cuts: 0.55 should be high, got %s", r.Level)
	}
}

func TestNewHTTPScorer_ValidCustomCuts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: true, Probability: 0.45})
	}))
	defer srv.Close()

	h := NewHTTPScorer(Config{BaseURL: srv.URL, LevelHigh: 0.4, LevelMedium: 0.3})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r.Level != "high" {
		t.Errorf("custom cuts 0.4/0.3: 0.45 should be high, got %s", r.Level)
	}
}

func TestNewHTTPScorer_HighOutOfRange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: true, Probability: 0.55})
	}))
	defer srv.Close()

	h := NewHTTPScorer(Config{BaseURL: srv.URL, LevelHigh: 1.5, LevelMedium: 0.2})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r.Level != "high" {
		t.Errorf("high out-of-range → default 0.5: 0.55 should be high, got %s", r.Level)
	}
}

func TestNewHTTPScorer_MediumOutOfRange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: true, Probability: 0.25})
	}))
	defer srv.Close()

	h := NewHTTPScorer(Config{BaseURL: srv.URL, LevelHigh: 0.5, LevelMedium: 1.5})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r.Level != "medium" {
		t.Errorf("medium out-of-range → default 0.2: 0.25 should be medium, got %s", r.Level)
	}
}

func TestNewHTTPScorer_MediumAboveHigh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: true, Probability: 0.3})
	}))
	defer srv.Close()

	h := NewHTTPScorer(Config{BaseURL: srv.URL, LevelHigh: 0.3, LevelMedium: 0.6})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r.Level != "medium" {
		t.Errorf("medium>high → default 0.5/0.2: 0.3 should be medium, got %s", r.Level)
	}
}

func TestNewHTTPScorer_LevelAtOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: true, Probability: 0.95})
	}))
	defer srv.Close()

	// LevelHigh=1.0 应合法（边界）
	h := NewHTTPScorer(Config{BaseURL: srv.URL, LevelHigh: 1.0, LevelMedium: 0.5})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r.Level != "medium" {
		t.Errorf("LevelHigh=1.0/Medium=0.5: 0.95 should be medium, got %s", r.Level)
	}
}

func TestHTTPScorer_ServiceUnavailableResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/idh/score" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(scoreResponse{Available: false, Probability: 0})
	}))
	defer srv.Close()

	h := NewHTTPScorer(Config{BaseURL: srv.URL})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if r.Available {
		t.Errorf("service available=false should degrade, got %+v", r)
	}
}

func TestRiskInputJSONContract(t *testing.T) {
	bf := 220.0
	g := 1
	age := 65.0
	pw := 62.5
	in := RiskInput{
		TreatmentID: 7, AccessType: "AVF",
		Window: []Sample{{BF: &bf}},
		Basic:  BasicInfo{Gender: &g, Age: &age, PreWeight: &pw},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, k := range []string{`"BF":220`, `"window"`, `"basic"`, `"treatmentId":7`} {
		if !strings.Contains(s, k) {
			t.Errorf("缺 %s; got %s", k, s)
		}
	}
	for _, k := range []string{`"Gender":1`, `"Age":65`, `"pre-Weight":62.5`} {
		if !strings.Contains(s, k) {
			t.Errorf("缺 %s; got %s", k, s)
		}
	}
}
