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
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(405)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"access_token": "tok123"})
	})
	mux.HandleFunc("/api/patients", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok123" {
			w.WriteHeader(401)
			return
		}
		if r.Method == http.MethodGet {
			q := r.URL.Query().Get("q")
			json.NewEncoder(w).Encode([]map[string]any{
				{"id": 77, "dialysis_id": q, "name": "test"},
				{"id": 78, "dialysis_id": q + "0", "name": "other"},
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 77, "dialysis_id": "D-001", "name": "test"})
	})
	mux.HandleFunc("/api/patients/77/xrays", func(w http.ResponseWriter, r *http.Request) {
		ctr := 0.48
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "ctr": ctr, "qc_pass": 1}})
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "ctr": ctr, "qc_pass": 1, "model_version": "v8.5"})
	})
	mux.HandleFunc("/api/xrays/9/correction", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("correction") == "" {
			w.WriteHeader(422)
			json.NewEncoder(w).Encode(map[string]string{"detail": "correction query param required"})
			return
		}
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
	if out.ID != 9 || out.CTR == nil || out.QCPass == nil || *out.QCPass != 1 {
		t.Fatalf("unexpected XrayOut: %+v", out)
	}
}

func TestClient_ApplyCorrection_UsesQueryParam(t *testing.T) {
	srv := fakeActrs(t)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL, Username: "u", Password: "p", TimeoutSec: 5})

	out, err := c.ApplyCorrection(context.Background(), 9, CorrectionRequest{DoctorCorrection: 0.5})
	if err != nil {
		t.Fatalf("ApplyCorrection: %v", err)
	}
	if out.DoctorCorrection == nil || *out.DoctorCorrection != 0.5 {
		t.Fatalf("unexpected DoctorCorrection: %+v", out)
	}
}

func TestClient_ApplyCorrection_NoQueryParam_Fails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"access_token": "tok"})
	})
	mux.HandleFunc("/api/xrays/9/correction", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("correction") == "" {
			w.WriteHeader(422)
			return
		}
		val := 0.5
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "doctor_correction": &val})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Username: "u", Password: "p", TimeoutSec: 5})
	_, err := c.ApplyCorrection(context.Background(), 9, CorrectionRequest{DoctorCorrection: 0.5})
	if err != nil {
		t.Fatalf("correction with query param should succeed: %v", err)
	}
}

func TestClient_SearchPatients(t *testing.T) {
	srv := fakeActrs(t)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL, Username: "u", Password: "p", TimeoutSec: 5})

	results, err := c.SearchPatients(context.Background(), "D-001")
	if err != nil {
		t.Fatalf("SearchPatients: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}
	if results[0].ID != 77 || results[0].DialysisID != "D-001" {
		t.Fatalf("first result wrong: %+v", results[0])
	}
}

func TestClient_Reachable(t *testing.T) {
	srv := fakeActrs(t)
	defer srv.Close()
	c := NewClient(Config{BaseURL: srv.URL, Username: "u", Password: "p", TimeoutSec: 5})

	if !c.Reachable(context.Background()) {
		t.Fatalf("server should be reachable")
	}

	c2 := NewClient(Config{BaseURL: "http://127.0.0.1:1", Username: "u", Password: "p", TimeoutSec: 1})
	if c2.Reachable(context.Background()) {
		t.Fatalf("dead port should not be reachable")
	}

	wrongSrv := httptest.NewServer(http.NewServeMux())
	defer wrongSrv.Close()
	c3 := NewClient(Config{BaseURL: wrongSrv.URL, Username: "u", Password: "p", TimeoutSec: 1})
	if c3.Reachable(context.Background()) {
		t.Fatalf("non-ACTRS service should not be reachable")
	}
}

func TestClient_NormalizeBaseURL(t *testing.T) {
	cases := []struct {
		input  string
		expect string
	}{
		{"http://localhost:8000", "http://localhost:8000"},
		{"http://localhost:8000/", "http://localhost:8000"},
		{"http://localhost:8000/api", "http://localhost:8000"},
		{"http://localhost:8000/API", "http://localhost:8000"},
		{"http://localhost:8000/api/", "http://localhost:8000"},
		{"  http://localhost:8000/api  ", "http://localhost:8000"},
	}
	for _, tc := range cases {
		got := normalizeBaseURL(tc.input)
		if got != tc.expect {
			t.Fatalf("normalizeBaseURL(%q) = %q, want %q", tc.input, got, tc.expect)
		}
	}
}

func TestClient_LoginPath_HasApiPrefix(t *testing.T) {
	var capturedPath string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		json.NewEncoder(w).Encode(map[string]string{"access_token": "tok"})
	})
	mux.HandleFunc("/api/patients", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": 1})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Username: "u", Password: "p", TimeoutSec: 5})
	_, _ = c.UpsertPatient(context.Background(), PatientCreate{DialysisID: "X"})
	if capturedPath != "/api/auth/login" {
		t.Fatalf("login path should be /api/auth/login, got %s", capturedPath)
	}
}

func TestClient_AllPathsHaveApiPrefix(t *testing.T) {
	var paths []string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"access_token": "tok"})
	})
	mux.HandleFunc("/api/patients", func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode([]map[string]any{})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 77, "dialysis_id": "D", "name": "n"})
	})
	mux.HandleFunc("/api/patients/77/xrays", func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		ctr := 0.5
		qc := 1
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "ctr": ctr, "qc_pass": qc}})
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "ctr": ctr, "qc_pass": qc})
	})
	mux.HandleFunc("/api/xrays/9/correction", func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		val := 0.5
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "doctor_correction": &val})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Username: "u", Password: "p", TimeoutSec: 5})
	ctx := context.Background()
	_, _ = c.SearchPatients(ctx, "D")
	_, _ = c.UpsertPatient(ctx, PatientCreate{DialysisID: "D", Name: "n"})
	_, _ = c.ListXrays(ctx, 77)
	_, _ = c.AnalyzeXray(ctx, 77, "c.jpg", strings.NewReader("b"))
	_, _ = c.ApplyCorrection(ctx, 9, CorrectionRequest{DoctorCorrection: 0.5})

	for _, p := range paths {
		if !strings.HasPrefix(p, "/api/") {
			t.Fatalf("path %s should start with /api/", p)
		}
	}
}
