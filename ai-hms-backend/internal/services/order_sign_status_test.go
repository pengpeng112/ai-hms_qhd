package services

import (
	"testing"

	"github.com/elliotxin/ai-hms-backend/internal/models"
)

func TestMapLegacyDaySignStatusByDictName(t *testing.T) {
	dict := map[int]string{
		10: "草稿",
		20: "已提交",
		30: "已确认",
		40: "已核对",
		50: "已执行",
		60: "已作废",
		70: "停止",
	}
	cases := map[int]string{
		10: models.OrderSignDraft,
		20: models.OrderSignSubmitted,
		30: models.OrderSignConfirmed,
		40: models.OrderSignChecked,
		50: models.OrderSignExecuted,
		60: models.OrderSignVoided,
		70: models.OrderSignVoided,
	}
	for raw, want := range cases {
		if got := mapLegacyDaySignStatus(raw, dict); got != want {
			t.Errorf("raw=%d dict=%q: got %q, want %q", raw, dict[raw], got, want)
		}
	}
}

func TestSignStatusPreservesGranularityVsCoarse(t *testing.T) {
	dict := map[int]string{11: "草稿", 12: "已提交", 13: "已确认", 14: "已核对"}
	for raw := 11; raw <= 14; raw++ {
		if got := mapLegacyDayOrderStatus(raw, dict); got != models.OrderStatusPending {
			t.Fatalf("coarse raw=%d: 预期压平为待执行, got %q", raw, got)
		}
	}
	got := []string{
		mapLegacyDaySignStatus(11, dict),
		mapLegacyDaySignStatus(12, dict),
		mapLegacyDaySignStatus(13, dict),
		mapLegacyDaySignStatus(14, dict),
	}
	want := []string{models.OrderSignDraft, models.OrderSignSubmitted, models.OrderSignConfirmed, models.OrderSignChecked}
	seen := map[string]bool{}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("粒度 idx=%d: got %q want %q", i, got[i], want[i])
		}
		seen[got[i]] = true
	}
	if len(seen) != 4 {
		t.Errorf("粒度应区分出 4 个不同状态, 实得 %d 个: %v", len(seen), got)
	}
}

func TestMapLegacyDaySignStatusByCodeRange(t *testing.T) {
	empty := map[int]string{}
	cases := map[int]string{
		0:  models.OrderSignDraft,
		10: models.OrderSignSubmitted,
		20: models.OrderSignConfirmed,
		30: models.OrderSignChecked,
		40: models.OrderSignExecuted,
		50: models.OrderSignVoided,
	}
	for raw, want := range cases {
		if got := mapLegacyDaySignStatus(raw, empty); got != want {
			t.Errorf("raw=%d (no dict): got %q, want %q", raw, got, want)
		}
	}
}
