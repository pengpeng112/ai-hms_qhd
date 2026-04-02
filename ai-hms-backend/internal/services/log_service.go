package services

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	defaultLogLines = 200
	minLogLines     = 50
	maxLogLines     = 1000
)

type LogSource string

const (
	LogSourceApp   LogSource = "app"
	LogSourceError LogSource = "error"
	LogSourceAll   LogSource = "all"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

type LogQuery struct {
	Source  LogSource
	Lines   int
	Keyword string
	Level   LogLevel
}

type LogEntry struct {
	Raw       string `json:"raw"`
	Timestamp string `json:"timestamp,omitempty"`
	Source    string `json:"source"`
	Level     string `json:"level,omitempty"`
}

type LogMeta struct {
	Source              string `json:"source"`
	Lines               int    `json:"lines"`
	Keyword             string `json:"keyword,omitempty"`
	Level               string `json:"level,omitempty"`
	Redacted            bool   `json:"redacted"`
	LevelApplied        *bool  `json:"levelApplied,omitempty"`
	LevelAppliedOnApp   *bool  `json:"levelAppliedOnApp,omitempty"`
	LevelAppliedOnError *bool  `json:"levelAppliedOnError,omitempty"`
	FetchedAt           string `json:"fetchedAt"`
}

type LogsResponse struct {
	Entries *[]LogEntry `json:"entries,omitempty"`
	Merged  *[]LogEntry `json:"merged,omitempty"`
	Meta    LogMeta     `json:"meta"`
}

type LogService struct {
	logDir string
}

func NewLogService() *LogService {
	return &LogService{
		logDir: "logs",
	}
}

func NewLogServiceWithDir(logDir string) *LogService {
	clean := strings.TrimSpace(logDir)
	if clean == "" {
		clean = "logs"
	}
	return &LogService{
		logDir: clean,
	}
}

func (s *LogService) GetLogs(query LogQuery) (*LogsResponse, error) {
	normalized, err := normalizeLogQuery(query)
	if err != nil {
		return nil, err
	}

	keyword := strings.ToLower(strings.TrimSpace(normalized.Keyword))
	fetchedAt := time.Now()

	switch normalized.Source {
	case LogSourceApp:
		appLines, readErr := tailLines(filepath.Join(s.logDir, "app.log"), normalized.Lines)
		if readErr != nil {
			return nil, readErr
		}
		parsed := buildLogEntries(appLines, LogSourceApp, keyword, "", false)
		entries := toLogEntries(parsed)
		levelApplied := false
		return &LogsResponse{
			Entries: &entries,
			Meta: LogMeta{
				Source:       string(LogSourceApp),
				Lines:        normalized.Lines,
				Keyword:      normalized.Keyword,
				Level:        string(normalized.Level),
				Redacted:     true,
				LevelApplied: &levelApplied,
				FetchedAt:    fetchedAt.Format(time.RFC3339),
			},
		}, nil
	case LogSourceError:
		errorLines, readErr := tailLines(filepath.Join(s.logDir, "error.log"), normalized.Lines)
		if readErr != nil {
			return nil, readErr
		}
		parsed := buildLogEntries(errorLines, LogSourceError, keyword, normalized.Level, normalized.Level != "")
		entries := toLogEntries(parsed)
		levelApplied := normalized.Level != ""
		return &LogsResponse{
			Entries: &entries,
			Meta: LogMeta{
				Source:       string(LogSourceError),
				Lines:        normalized.Lines,
				Keyword:      normalized.Keyword,
				Level:        string(normalized.Level),
				Redacted:     true,
				LevelApplied: &levelApplied,
				FetchedAt:    fetchedAt.Format(time.RFC3339),
			},
		}, nil
	case LogSourceAll:
		appLines, readErr := tailLines(filepath.Join(s.logDir, "app.log"), normalized.Lines)
		if readErr != nil {
			return nil, readErr
		}
		errorLines, readErr := tailLines(filepath.Join(s.logDir, "error.log"), normalized.Lines)
		if readErr != nil {
			return nil, readErr
		}

		appParsed := buildLogEntries(appLines, LogSourceApp, keyword, "", false)
		errorParsed := buildLogEntries(errorLines, LogSourceError, keyword, normalized.Level, normalized.Level != "")
		merged := mergeAndSort(appParsed, errorParsed)
		result := toLogEntries(merged)

		levelAppliedOnApp := false
		levelAppliedOnError := normalized.Level != ""
		return &LogsResponse{
			Merged: &result,
			Meta: LogMeta{
				Source:              string(LogSourceAll),
				Lines:               normalized.Lines,
				Keyword:             normalized.Keyword,
				Level:               string(normalized.Level),
				Redacted:            true,
				LevelAppliedOnApp:   &levelAppliedOnApp,
				LevelAppliedOnError: &levelAppliedOnError,
				FetchedAt:           fetchedAt.Format(time.RFC3339),
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported log source: %s", normalized.Source)
	}
}

type parsedLogEntry struct {
	entry       LogEntry
	effectiveTS time.Time
	hasTS       bool
	seq         int64
}

func mergeAndSort(appEntries, errorEntries []parsedLogEntry) []parsedLogEntry {
	merged := make([]parsedLogEntry, 0, len(appEntries)+len(errorEntries))
	merged = append(merged, appEntries...)
	merged = append(merged, errorEntries...)

	sort.SliceStable(merged, func(i, j int) bool {
		left := merged[i]
		right := merged[j]

		if left.hasTS && right.hasTS {
			if !left.effectiveTS.Equal(right.effectiveTS) {
				return left.effectiveTS.Before(right.effectiveTS)
			}
			return left.seq < right.seq
		}
		if left.hasTS != right.hasTS {
			return left.hasTS
		}
		return left.seq < right.seq
	})

	return merged
}

func toLogEntries(entries []parsedLogEntry) []LogEntry {
	result := make([]LogEntry, 0, len(entries))
	for _, item := range entries {
		result = append(result, item.entry)
	}
	return result
}

func buildLogEntries(lines []string, source LogSource, keyword string, levelFilter LogLevel, applyLevel bool) []parsedLogEntry {
	result := make([]parsedLogEntry, 0, len(lines))
	var lastTS time.Time
	hasLastTS := false
	lastLevel := LogLevel("")

	for idx, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		redacted := redactLogLine(line)

		ts, hasTS := parseTimestamp(source, line)
		if hasTS {
			lastTS = ts
			hasLastTS = true
		}

		level := LogLevel("")
		if source == LogSourceError {
			level = parseErrorLevel(line)
			if level != "" {
				lastLevel = level
			}
			if level == "" && lastLevel != "" {
				level = lastLevel
			}
		}

		if keyword != "" && !strings.Contains(strings.ToLower(redacted), keyword) {
			continue
		}
		if applyLevel && source == LogSourceError && levelFilter != "" && level != levelFilter {
			continue
		}

		entry := LogEntry{
			Raw:    redacted,
			Source: string(source),
		}
		if hasLastTS {
			entry.Timestamp = lastTS.Format(time.RFC3339)
		}
		if level != "" {
			entry.Level = string(level)
		}

		result = append(result, parsedLogEntry{
			entry:       entry,
			effectiveTS: lastTS,
			hasTS:       hasLastTS,
			seq:         int64(idx),
		})
	}

	return result
}

func normalizeLogQuery(query LogQuery) (LogQuery, error) {
	normalized := query

	if normalized.Source == "" {
		normalized.Source = LogSourceAll
	}
	if normalized.Source != LogSourceApp && normalized.Source != LogSourceError && normalized.Source != LogSourceAll {
		return LogQuery{}, errors.New("invalid source, expected app|error|all")
	}

	if normalized.Lines == 0 {
		normalized.Lines = defaultLogLines
	}
	if normalized.Lines < minLogLines || normalized.Lines > maxLogLines {
		return LogQuery{}, fmt.Errorf("invalid lines, expected %d-%d", minLogLines, maxLogLines)
	}

	normalized.Keyword = strings.TrimSpace(normalized.Keyword)
	normalized.Level = LogLevel(strings.ToUpper(strings.TrimSpace(string(normalized.Level))))
	if normalized.Level != "" && normalized.Level != LogLevelInfo && normalized.Level != LogLevelWarn && normalized.Level != LogLevelError {
		return LogQuery{}, errors.New("invalid level, expected INFO|WARN|ERROR")
	}

	return normalized, nil
}

var (
	errorTimestampPattern = regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})`)
	ginTimestampPattern   = regexp.MustCompile(`^\[GIN\]\s+(\d{4}/\d{2}/\d{2})\s+-\s+(\d{2}:\d{2}:\d{2})`)
	errorLevelPattern     = regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\s+(INFO|WARN|ERROR)\b`)
	errorKVLevelPattern   = regexp.MustCompile(`(?i)\blevel=(info|warn|error)\b`)

	authBearerPattern = regexp.MustCompile(`(?i)(Authorization:\s*Bearer\s+)[^\s"]+`)
	jsonSecretPattern = regexp.MustCompile(`(?i)"(access_token|refresh_token|password|secret|client_secret|service_password)"\s*:\s*"[^"]*"`)
	jsonSecretKey     = regexp.MustCompile(`(?i)^"([^"]+)"`)
	kvSecretPattern   = regexp.MustCompile(`(?i)\b(access_token|refresh_token|password|secret|client_secret|service_password)\b(\s*[:=]\s*)([^,\s&]+)`)
)

func parseTimestamp(source LogSource, line string) (time.Time, bool) {
	switch source {
	case LogSourceError:
		matched := errorTimestampPattern.FindStringSubmatch(line)
		if len(matched) < 2 {
			return time.Time{}, false
		}
		ts, err := time.ParseInLocation("2006/01/02 15:04:05", matched[1], time.Local)
		if err != nil {
			return time.Time{}, false
		}
		return ts, true
	case LogSourceApp:
		matched := ginTimestampPattern.FindStringSubmatch(line)
		if len(matched) < 3 {
			return time.Time{}, false
		}
		ts, err := time.ParseInLocation("2006/01/02 15:04:05", matched[1]+" "+matched[2], time.Local)
		if err != nil {
			return time.Time{}, false
		}
		return ts, true
	default:
		return time.Time{}, false
	}
}

func parseErrorLevel(line string) LogLevel {
	if matched := errorLevelPattern.FindStringSubmatch(line); len(matched) >= 2 {
		return normalizeLevelToken(matched[1])
	}
	if matched := errorKVLevelPattern.FindStringSubmatch(line); len(matched) >= 2 {
		return normalizeLevelToken(matched[1])
	}
	return ""
}

func normalizeLevelToken(token string) LogLevel {
	switch strings.ToUpper(strings.TrimSpace(token)) {
	case string(LogLevelInfo):
		return LogLevelInfo
	case string(LogLevelWarn):
		return LogLevelWarn
	case string(LogLevelError):
		return LogLevelError
	default:
		return ""
	}
}

func redactLogLine(line string) string {
	redacted := authBearerPattern.ReplaceAllString(line, "${1}***")
	redacted = jsonSecretPattern.ReplaceAllStringFunc(redacted, func(segment string) string {
		matched := jsonSecretKey.FindStringSubmatch(segment)
		if len(matched) < 2 {
			return segment
		}
		return `"` + matched[1] + `":"***"`
	})
	redacted = kvSecretPattern.ReplaceAllString(redacted, "${1}${2}***")
	return redacted
}

func tailLines(path string, maxLines int) ([]string, error) {
	if maxLines <= 0 {
		return []string{}, nil
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	if fileInfo.Size() == 0 {
		return []string{}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	const chunkSize int64 = 4096
	size := fileInfo.Size()
	offset := size
	buffer := make([]byte, 0, minInt64(size, chunkSize*4))
	need := maxLines + 1

	for offset > 0 {
		readSize := chunkSize
		if offset < readSize {
			readSize = offset
		}
		offset -= readSize

		chunk := make([]byte, readSize)
		if _, readErr := file.ReadAt(chunk, offset); readErr != nil {
			return nil, readErr
		}
		buffer = append(chunk, buffer...)

		if bytes.Count(buffer, []byte{'\n'}) >= need {
			break
		}
	}

	content := strings.ReplaceAll(string(buffer), "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if offset > 0 && len(lines) > 0 {
		lines = lines[1:]
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return lines, nil
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
