package idh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	r := h.Score(context.Background(), RiskInput{TreatmentID: 1, Window: []Sample{{BF: 220}}})
	if !r.Available || r.Probability != 0.73 || r.Level != "high" {
		t.Errorf("unexpected result: %+v", r)
	}
}
