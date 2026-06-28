package services

import (
	"testing"

	"github.com/elliotxin/ai-hms-backend/internal/config"
)

// nil db → 全量回退内嵌 JSON（固定阈值含 map/heartRate，VP 非空）。
func TestBuildThresholds_NilDBFallsBackToJSON(t *testing.T) {
	got := buildThresholds(nil, config.LoadMonitoringThresholds)
	if got == nil {
		t.Fatal("期望非 nil")
	}
	if _, ok := got.Fixed["map"]; !ok {
		t.Errorf("期望固定阈值含 map，实际 keys=%v", got.Fixed)
	}
	if len(got.VPReference) == 0 {
		t.Error("期望 VP 分层非空（回退 JSON）")
	}
	if got.NaFactor() != 9.9 {
		t.Errorf("期望默认 naFactor 9.9，实际 %v", got.NaFactor())
	}
}

// fallback 返回 nil 时不 panic，返回空 Fixed 的安全结构。
func TestBuildThresholds_NilFallbackSafe(t *testing.T) {
	got := buildThresholds(nil, func() (*config.MonitoringThresholds, error) { return nil, nil })
	if got == nil || got.Fixed == nil {
		t.Fatal("期望非 nil 且 Fixed 已初始化")
	}
}

// 缓存失效后 data 被清空。
func TestInvalidateThresholdCache(t *testing.T) {
	thresholdCache.data = &config.MonitoringThresholds{}
	InvalidateThresholdCache()
	if thresholdCache.data != nil {
		t.Error("失效后期望 data 为 nil")
	}
}
