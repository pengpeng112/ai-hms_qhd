package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetLogsSingleSourceShape(t *testing.T) {
	logDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(logDir, "app.log"), []byte("[GIN] 2026/03/04 - 12:00:00 | 200 | 1ms | ::1 | GET \"/health\"\n"), 0o644); err != nil {
		t.Fatalf("write app.log failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "error.log"), []byte("2026/03/04 12:00:01 INFO server started\n"), 0o644); err != nil {
		t.Fatalf("write error.log failed: %v", err)
	}

	svc := NewLogServiceWithDir(logDir)
	res, err := svc.GetLogs(LogQuery{
		Source: LogSourceApp,
		Lines:  200,
	})
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}
	if res.Entries == nil {
		t.Fatalf("expected entries to be present")
	}
	if res.Merged != nil {
		t.Fatalf("expected merged to be nil for app source")
	}
	if len(*res.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(*res.Entries))
	}
}

func TestGetLogsAllKeepsStackTraceNearErrorLine(t *testing.T) {
	logDir := t.TempDir()
	appLog := "" +
		"[GIN] 2026/03/04 - 12:00:00 | 200 | 1ms | ::1 | GET \"/health\"\n" +
		"[GIN] 2026/03/04 - 12:00:02 | 500 | 2m0s | ::1 | POST \"/api/v1/settings/integrations/hdis/refresh-token\"\n"
	errorLog := "" +
		"2026/03/04 12:00:01 ERROR refresh token failed\n" +
		"runtime error: access_token not found\n" +
		"goroutine 100 [running]:\n"

	if err := os.WriteFile(filepath.Join(logDir, "app.log"), []byte(appLog), 0o644); err != nil {
		t.Fatalf("write app.log failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "error.log"), []byte(errorLog), 0o644); err != nil {
		t.Fatalf("write error.log failed: %v", err)
	}

	svc := NewLogServiceWithDir(logDir)
	res, err := svc.GetLogs(LogQuery{
		Source: LogSourceAll,
		Lines:  200,
	})
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}
	if res.Merged == nil {
		t.Fatalf("expected merged to be present")
	}
	if res.Entries != nil {
		t.Fatalf("expected entries to be nil for all source")
	}

	got := *res.Merged
	if len(got) != 5 {
		t.Fatalf("expected 5 merged entries, got %d", len(got))
	}
	if got[0].Source != string(LogSourceApp) || got[0].Raw == "" {
		t.Fatalf("unexpected first entry: %+v", got[0])
	}
	if got[1].Source != string(LogSourceError) || got[1].Level != string(LogLevelError) {
		t.Fatalf("expected error headline at index 1, got %+v", got[1])
	}
	if got[2].Source != string(LogSourceError) || got[3].Source != string(LogSourceError) {
		t.Fatalf("expected stack lines immediately after error headline, got %+v %+v", got[2], got[3])
	}
	if got[4].Source != string(LogSourceApp) {
		t.Fatalf("expected trailing app line at index 4, got %+v", got[4])
	}
}

func TestGetLogsLevelFilterOnErrorKeepsContinuationLines(t *testing.T) {
	logDir := t.TempDir()
	errorLog := "" +
		"2026/03/04 12:00:01 INFO background job started\n" +
		"2026/03/04 12:00:02 ERROR refresh token failed\n" +
		"stack line A\n" +
		"stack line B\n" +
		"2026/03/04 12:00:03 WARN fallback used\n"
	if err := os.WriteFile(filepath.Join(logDir, "error.log"), []byte(errorLog), 0o644); err != nil {
		t.Fatalf("write error.log failed: %v", err)
	}

	svc := NewLogServiceWithDir(logDir)
	res, err := svc.GetLogs(LogQuery{
		Source: LogSourceError,
		Lines:  200,
		Level:  LogLevelError,
	})
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}
	if res.Entries == nil {
		t.Fatalf("expected entries to be present")
	}

	got := *res.Entries
	if len(got) != 3 {
		t.Fatalf("expected 3 lines (error + 2 stack), got %d", len(got))
	}
	if got[0].Level != string(LogLevelError) {
		t.Fatalf("expected first line level=ERROR, got %+v", got[0])
	}
	if got[1].Level != string(LogLevelError) || got[2].Level != string(LogLevelError) {
		t.Fatalf("expected continuation lines inherit ERROR level, got %+v %+v", got[1], got[2])
	}
}

func TestRedactLogLine(t *testing.T) {
	input := "Authorization: Bearer abcdef password=foo access_token=bar client_secret=baz"
	got := redactLogLine(input)
	if got == input {
		t.Fatalf("expected redaction to modify line")
	}
	if containsAny(got, "abcdef", "foo", "bar", "baz") {
		t.Fatalf("expected sensitive values to be masked, got: %s", got)
	}
}

func TestTailLines(t *testing.T) {
	logDir := t.TempDir()
	path := filepath.Join(logDir, "app.log")
	content := "l1\nl2\nl3\nl4\nl5\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write app.log failed: %v", err)
	}

	got, err := tailLines(path, 2)
	if err != nil {
		t.Fatalf("tailLines failed: %v", err)
	}
	if len(got) != 2 || got[0] != "l4" || got[1] != "l5" {
		t.Fatalf("unexpected tail result: %#v", got)
	}
}

func TestParseErrorLevelSupportsKeyValueFormat(t *testing.T) {
	line := `2026/03/04 12:00:01 level=ERROR msg="refresh failed"`
	if got := parseErrorLevel(line); got != LogLevelError {
		t.Fatalf("expected ERROR from key-value line, got %q", got)
	}

	line2 := `2026/03/04 12:00:02 LEVEL=warn msg="fallback"`
	if got := parseErrorLevel(line2); got != LogLevelWarn {
		t.Fatalf("expected WARN from key-value line (case-insensitive), got %q", got)
	}
}

func containsAny(s string, values ...string) bool {
	for _, v := range values {
		if v != "" && strings.Contains(s, v) {
			return true
		}
	}
	return false
}
