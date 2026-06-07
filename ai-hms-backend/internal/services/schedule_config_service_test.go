package services

import (
	"strings"
	"testing"
)

func TestNormalizeModes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"single lowercase", "hd", "HD"},
		{"single with spaces", " hd ", "HD"},
		{"comma list", "HD,HDF,HF", "HD,HDF,HF"},
		{"comma with spaces", " HD , HDF , HF ", "HD,HDF,HF"},
		{"lowercase mixed", "hd,hdf,hf", "HD,HDF,HF"},
		{"duplicates removed", "HD,HD,HDF", "HD,HDF"},
		{"empty entries skipped", "HD,,HDF", "HD,HDF"},
		{"single CRRT", "CRRT", "CRRT"},
		{"empty input", "", ""},
		{"whitespace only", "  ,  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeModes(tt.raw)
			joined := strings.Join(got, ",")
			if joined != tt.want {
				t.Errorf("normalizeModes(%q) = %q, want %q", tt.raw, joined, tt.want)
			}
		})
	}
}

func TestValidateMachineTypeAndModes(t *testing.T) {
	tests := []struct {
		name          string
		machineType   string
		supportedMode string
		wantModes     string
		wantErr       bool
		errContains   string
	}{
		// HD machine
		{"HD defaults to HD", "HD", "", "HD", false, ""},
		{"HD only HD allowed", "HD", "HD", "HD", false, ""},
		{"HD rejects HDF", "HD", "HD,HDF", "", true, "只支持 HD"},
		{"HD rejects HF", "HD", "HD,HF", "", true, "只支持 HD"},
		{"HD rejects CRRT", "HD", "CRRT", "", true, "只支持 HD"},
		{"HD rejects non-HD mode", "HD", "HDF", "", true, "只支持 HD"},

		// HDF machine
		{"HDF defaults to HD,HDF,HF", "HDF", "", "HD,HDF,HF", false, ""},
		{"HDF with HD,HDF,HF ok", "HDF", "HD,HDF,HF", "HD,HDF,HF", false, ""},
		{"HDF with HD,HDF ok", "HDF", "HD,HDF", "HD,HDF", false, ""},
		{"HDF must include HD and HDF", "HDF", "HD", "", true, "必须包含 HD 和 HDF"},
		{"HDF must include HD and HDF - HF only", "HDF", "HF", "", true, "必须包含 HD 和 HDF"},
		{"HDF rejects CRRT", "HDF", "HD,HDF,CRRT", "", true, "不支持 CRRT"},

		// CRRT machine
		{"CRRT defaults to CRRT", "CRRT", "", "CRRT", false, ""},
		{"CRRT only CRRT allowed", "CRRT", "CRRT", "CRRT", false, ""},
		{"CRRT rejects HD", "CRRT", "CRRT,HD", "", true, "只支持 CRRT"},
		{"CRRT rejects HDF", "CRRT", "HDF", "", true, "只支持 CRRT"},
		{"CRRT rejects non-CRRT mode", "CRRT", "HD", "", true, "只支持 CRRT"},

		// Invalid machine type
		{"invalid machine type", "X", "HD", "", true, "必须为 HD、HDF 或 CRRT"},

		// Lowercase / space normalization
		{"lowercase hd with spaces", "hd", " hd ", "HD", false, ""},
		{"lowercase hdf with spaces", "hdf", " hd , hdf ", "HD,HDF", false, ""},
		{"mixed case", "HdF", "hd,HDF,HF", "HD,HDF,HF", false, ""},

		// HDF should not be misdetected as HD-containing
		{"HDF machine with valid lowercase", "HDF", "hd,hdf", "HD,HDF", false, ""},

		// Empty modes after normalization
		{"empty modes not allowed", "HD", "  ,  ", "", true, "不能为空"},
		// Invalid mode token
		{"invalid mode token", "HD", "HD,XYZ", "", true, "只能包含 HD、HDF、HF、CRRT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateMachineTypeAndModes(tt.machineType, tt.supportedMode)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if got != tt.wantModes {
					t.Errorf("validateMachineTypeAndModes(%q, %q) = %q, want %q",
						tt.machineType, tt.supportedMode, got, tt.wantModes)
				}
			}
		})
	}
}
