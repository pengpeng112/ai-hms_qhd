package sched

import (
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func i16(v int16) *int16 { return &v }

// anchor 用默认基准周一 2025-01-06。
var anchor = DefaultAnchorMonday

func TestAnchorIsMonday(t *testing.T) {
	if DefaultAnchorMonday.Weekday() != time.Monday {
		t.Fatalf("默认基准必须是周一,实际 %v", DefaultAnchorMonday.Weekday())
	}
}

func TestWeekIndexAndParity(t *testing.T) {
	cases := []struct {
		name      string
		d         time.Time
		wantIndex int
		wantPar   int16
	}{
		{"基准周一本身", day(2025, 1, 6), 0, 0},
		{"同周周日归一到周一", day(2025, 1, 12), 0, 0},
		{"下一周周一", day(2025, 1, 13), 1, 1},
		{"第2周", day(2025, 1, 20), 2, 0},
		{"基准前一周(负序号归一)", day(2024, 12, 30), -1, 1},
		{"跨年连续无ISO第53周漂移", day(2025, 12, 29), 51, 1},
	}
	for _, c := range cases {
		if got := WeekIndex(anchor, c.d); got != c.wantIndex {
			t.Errorf("%s: WeekIndex=%d, want %d", c.name, got, c.wantIndex)
		}
		if got := WeekParity(anchor, c.d); got != c.wantPar {
			t.Errorf("%s: WeekParity=%d, want %d", c.name, got, c.wantPar)
		}
	}
}

func TestIsDue(t *testing.T) {
	cases := []struct {
		freq int16
		d    time.Time
		want bool
	}{
		{FreqMonWedFri, day(2025, 1, 6), true},   // 周一
		{FreqMonWedFri, day(2025, 1, 7), false},  // 周二
		{FreqMonWedFri, day(2025, 1, 8), true},   // 周三
		{FreqTueThuSat, day(2025, 1, 7), true},   // 周二
		{FreqTwoPerWk, day(2025, 1, 9), true},    // 周四
		{FreqTwoPerWk, day(2025, 1, 10), false},  // 周五
		{FreqOnePerWk, day(2025, 1, 9), true},    // 仅周四
		{FreqOnePerWk, day(2025, 1, 7), false},   // 周二
		{FreqTemporary, day(2025, 1, 6), false},  // 临时无固定日
	}
	for _, c := range cases {
		if got := IsDue(c.freq, c.d); got != c.want {
			t.Errorf("IsDue(freq=%d, %s)=%v, want %v", c.freq, c.d.Format("2006-01-02"), got, c.want)
		}
	}
}

func TestDecideMode(t *testing.T) {
	// 病人:一三五,启用 HDF,HDF 日=周三(ISO 3),HDF 所属周=偶(0)。
	wd, par := i16(3), i16(0)

	// 周三且偶数周 → HDF(本次替换 HD)
	if got := DecideMode(anchor, true, wd, par, day(2025, 1, 8)); got != ModeHDF {
		t.Errorf("偶周周三应为 HDF,得 %s", got)
	}
	// 周三但奇数周 → HD(奇偶错峰)
	if got := DecideMode(anchor, true, wd, par, day(2025, 1, 15)); got != ModeHD {
		t.Errorf("奇周周三应为 HD,得 %s", got)
	}
	// 偶数周但周一(非 HDF 日)→ HD
	if got := DecideMode(anchor, true, wd, par, day(2025, 1, 6)); got != ModeHD {
		t.Errorf("非 HDF 日应为 HD,得 %s", got)
	}
	// 未启用 HDF → 恒 HD
	if got := DecideMode(anchor, false, wd, par, day(2025, 1, 8)); got != ModeHD {
		t.Errorf("未启用 HDF 应为 HD,得 %s", got)
	}
}

func TestMachineSupports(t *testing.T) {
	cases := []struct {
		mt, mode string
		want     bool
	}{
		{MachineHD, ModeHD, true},
		{MachineHD, ModeHDF, false},
		{MachineHDF, ModeHD, true},   // HDF 机向下兼容 HD
		{MachineHDF, ModeHDF, true},
		{MachineHDF, ModeHF, true},
		{MachineCRRT, ModeCRRT, true},
		{MachineCRRT, ModeHD, false},
	}
	for _, c := range cases {
		if got := MachineSupports(c.mt, c.mode); got != c.want {
			t.Errorf("MachineSupports(%s,%s)=%v, want %v", c.mt, c.mode, got, c.want)
		}
	}
}
