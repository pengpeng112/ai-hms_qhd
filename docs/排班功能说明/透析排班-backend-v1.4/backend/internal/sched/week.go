package sched

import (
	"math"
	"time"
)

// 本文件是排班算法的确定性核心:周序号 / 奇偶周(决策 21 / D-6)、频率推导、
// HDF 替换判定(决策 1/3)。全为纯函数,便于单元测试。

// dateOnly 归一到当日零点(同一时区),消除"日内时间"对日期运算的干扰。
func dateOnly(d time.Time) time.Time {
	y, m, day := d.Date()
	return time.Date(y, m, day, 0, 0, 0, 0, d.Location())
}

// isoWeekday 返回 ISO 星期:周一=1 … 周日=7(与 PatientProfile.HdfWeekday 取值一致)。
func isoWeekday(d time.Time) int16 {
	wd := d.Weekday() // time: 周日=0 … 周六=6
	if wd == time.Sunday {
		return 7
	}
	return int16(wd)
}

// MondayOf 返回该日期所在自然周的周一(零点)。
func MondayOf(d time.Time) time.Time {
	d = dateOnly(d)
	return d.AddDate(0, 0, -int(isoWeekday(d)-1))
}

// daysBetween 两个零点日期间的整天数(可为负)。因入参均为零点,差值是 24h 的整数倍。
func daysBetween(a, b time.Time) int {
	return int(math.Round(b.Sub(a).Hours() / 24))
}

// WeekIndex 全局连续周序号(决策 21):以基准周一为第 0 周,逐周 +1,可跨年、可为负。
// 因 MondayOf(anchor) 与 MondayOf(d) 均为周一,差值必为 7 的整数倍,整除精确无截断问题。
func WeekIndex(anchor, d time.Time) int {
	return daysBetween(MondayOf(anchor), MondayOf(d)) / 7
}

// WeekParity 奇偶周:0=偶数周 / 1=奇数周。对负序号做归一,保证基准前的周也稳定。
func WeekParity(anchor, d time.Time) int16 {
	p := WeekIndex(anchor, d) % 2
	if p < 0 {
		p += 2
	}
	return int16(p)
}

// IsDue 按频率模式判断某日是否为该病人的透析日(规范 §2.1)。临时模式无固定日,恒 false。
func IsDue(freq int16, d time.Time) bool {
	for _, wd := range freqWeekdays[freq] {
		if wd == d.Weekday() {
			return true
		}
	}
	return false
}

// DecideMode 判定某次透析的治疗模式(决策 1/25)。
// 常规日返回病人**基础模式** baseMode(HD/HFD/HF 等;空则按 HD);
// 仅当 (启用HDF) 且 (本日==HDF日) 且 (本周奇偶==该病人HDF所属周奇偶) 时,本次替换为 HDF。
// hdfWeekday 用 ISO(周一=1..周六=6);hdfParity 取值 0/1,与 WeekParity 对齐。
func DecideMode(anchor time.Time, baseMode string, hdfEnabled bool, hdfWeekday, hdfParity *int16, d time.Time) string {
	base := baseMode
	if base == "" {
		base = ModeHD
	}
	if !hdfEnabled || hdfWeekday == nil || hdfParity == nil {
		return base
	}
	if isoWeekday(d) != *hdfWeekday {
		return base
	}
	if WeekParity(anchor, d) != *hdfParity {
		return base
	}
	return ModeHDF // 本次基础模式被替换为 HDF
}
