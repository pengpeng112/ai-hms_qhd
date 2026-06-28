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
		_ = json.NewEncoder(w).Encode(scoreResponse{Probability: 0.73})
	}))
	defer srv.Close()

	h := NewHTTPScorer(Config{BaseURL: srv.URL})
	bf := 220.0
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: &bf}}})
	if !r.Available || r.Probability != 0.73 || r.Level != "high" {
		t.Errorf("unexpected result: %+v", r)
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
