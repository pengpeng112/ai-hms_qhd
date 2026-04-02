package services

import (
	"testing"
	"time"
)

func TestNormalizeAbnormalFlag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty defaults to N", input: "", want: "N", wantErr: false},
		{name: "n lowercase", input: "n", want: "N", wantErr: false},
		{name: "high", input: "H", want: "H", wantErr: false},
		{name: "low", input: "l", want: "L", wantErr: false},
		{name: "invalid", input: "X", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAbnormalFlag(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestNormalizeSourceSystem(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty defaults to LOCAL", input: "", want: "LOCAL", wantErr: false},
		{name: "local lowercase", input: "local", want: "LOCAL", wantErr: false},
		{name: "lis", input: "LIS", want: "LIS", wantErr: false},
		{name: "pacs", input: "pacs", want: "PACS", wantErr: false},
		{name: "invalid", input: "ERP", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSourceSystem(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestNormalizePagination(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{name: "default pagination", page: 0, pageSize: 0, wantPage: 1, wantPageSize: 20},
		{name: "negative pagination", page: -1, pageSize: -1, wantPage: 1, wantPageSize: 20},
		{name: "normal pagination", page: 2, pageSize: 50, wantPage: 2, wantPageSize: 50},
		{name: "limit page size", page: 1, pageSize: 1000, wantPage: 1, wantPageSize: 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := normalizePagination(tt.page, tt.pageSize)
			if gotPage != tt.wantPage {
				t.Fatalf("expected page=%d, got=%d", tt.wantPage, gotPage)
			}
			if gotPageSize != tt.wantPageSize {
				t.Fatalf("expected pageSize=%d, got=%d", tt.wantPageSize, gotPageSize)
			}
		})
	}
}

func TestApplyTimeFieldUpdate(t *testing.T) {
	base := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	t.Run("nil input should not update", func(t *testing.T) {
		current := &base
		if err := applyTimeFieldUpdate(nil, &current); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current == nil || !current.Equal(base) {
			t.Fatalf("expected unchanged time")
		}
	})

	t.Run("empty string should clear", func(t *testing.T) {
		current := &base
		empty := ""
		if err := applyTimeFieldUpdate(&empty, &current); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current != nil {
			t.Fatalf("expected nil time, got %v", current)
		}
	})

	t.Run("valid date should update", func(t *testing.T) {
		current := &base
		in := "2026-03-03"
		if err := applyTimeFieldUpdate(&in, &current); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current == nil {
			t.Fatalf("expected non-nil time")
		}
		if got := current.Format("2006-01-02"); got != "2026-03-03" {
			t.Fatalf("expected 2026-03-03, got %s", got)
		}
	})

	t.Run("invalid date should error", func(t *testing.T) {
		current := &base
		in := "2026-13-99"
		if err := applyTimeFieldUpdate(&in, &current); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
