package services

import (
	"context"
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/integrations/idh"
)

type fakeScorer struct {
	calls  int
	result idh.RiskResult
}

func (f *fakeScorer) Score(_ context.Context, _ idh.RiskInput) idh.RiskResult {
	f.calls++
	return f.result
}

func fptr(v float64) *float64 { return &v }

func TestBuildIDHInput_OrderAndMapping(t *testing.T) {
	rows := []dmlogWindowRow{
		{LogTime: time.Unix(100, 0), BF: fptr(210), TMP: fptr(120)},
		{LogTime: time.Unix(160, 0), BF: fptr(220), TMP: fptr(130)},
	}
	g := 1
	basic := idhBasic{Gender: &g, PreWeight: fptr(62.5)}
	in := buildIDHInput(7, "AVF", rows, basic)
	if in.TreatmentID != 7 || in.AccessType != "AVF" {
		t.Fatalf("头部映射错: %+v", in)
	}
	if len(in.Window) != 2 || in.Window[0].BF == nil || *in.Window[0].BF != 210 {
		t.Fatalf("窗口顺序/映射错: %+v", in.Window)
	}
	if in.Basic.Gender == nil || *in.Basic.Gender != 1 || in.Basic.PreWeight == nil || *in.Basic.PreWeight != 62.5 {
		t.Fatalf("基本信息映射错: %+v", in.Basic)
	}
}

func TestRefreshIDHNow_PopulatesCache(t *testing.T) {
	resetIDHStateForTest()
	t.Cleanup(resetIDHStateForTest)
	SetIDHScorer(&fakeScorer{result: idh.RiskResult{Available: true, Probability: 0.8, Level: "high"}})
	refreshIDHNow(3, 42, "AVF", idhBasic{})
	got := lookupIDHCached(3, 42, "AVF", idhBasic{})
	if !got.Available || got.Probability != 0.8 {
		t.Fatalf("期望缓存命中 0.8，got %+v", got)
	}
}

func TestLookupIDHCached_MissReturnsUnavailable(t *testing.T) {
	resetIDHStateForTest()
	t.Cleanup(resetIDHStateForTest)
	got := lookupIDHCached(3, 999, "AVF", idhBasic{})
	if got.Available {
		t.Fatalf("无缓存应不可用，got %+v", got)
	}
}

func TestTriggerIDHRefresh_StubSkips(t *testing.T) {
	resetIDHStateForTest()
	t.Cleanup(resetIDHStateForTest)
	triggerIDHRefresh(3, 7, "AVF", idhBasic{})
	idhCache.mu.RLock()
	_, ok := idhCache.m[idhCacheKey{tenantID: 3, treatmentID: 7}]
	idhCache.mu.RUnlock()
	if ok {
		t.Fatal("Stub 不应触发刷新写缓存")
	}
}
