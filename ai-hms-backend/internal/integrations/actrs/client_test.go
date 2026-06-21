package actrs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fakeActrs(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"access_token": "tok123"})
	})
	mux.HandleFunc("/patients", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok123" {
			w.WriteHeader(401)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 77, "dialysis_id": "D-001", "name": "test"})
	})
	mux.HandleFunc("/patients/77/xrays", func(w http.ResponseWriter, r *http.Request) {
		ctr := 0.48
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "ctr": ctr, "qc_pass": true}})
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "ctr": ctr, "qc_pass": true, "model_version": "v8.5"})
	})
	mux.HandleFunc("/xrays/9/correction", func(w http.ResponseWriter, _ *http.Request) {
		val := 0.50
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "ctr": &val, "doctor_correction": &val})
	})
	return httptest.NewServer(mux)
}

func TestClient_UpsertPatient(t *testing.T) {
	srv := fakeActrs(t)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL, Username: "u", Password: "p", TimeoutSec: 5})
	p, err := c.UpsertPatient(context.Background(), PatientCreate{DialysisID: "D-001", Name: "test"})
	if err != nil {
		t.Fatalf("UpsertPatient: %v", err)
	}
	if p.ID != 77 {
		t.Fatalf("want id 77, got %d", p.ID)
	}
}

func TestClient_AnalyzeXray(t *testing.T) {
	srv := fakeActrs(t)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL, Username: "u", Password: "p", TimeoutSec: 5})
	out, err := c.AnalyzeXray(context.Background(), 77, "chest.jpg", strings.NewReader("fakejpegbytes"))
	if err != nil {
		t.Fatalf("AnalyzeXray: %v", err)
	}
	if out.ID != 9 || out.CTR == nil || !out.QCPass {
		t.Fatalf("unexpected XrayOut: %+v", out)
	}
}
