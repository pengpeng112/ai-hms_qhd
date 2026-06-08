package schedule_engine

import "time"

// ParityAssignment HDF奇偶周分配结果(需写回持久层)
type ParityAssignment struct {
	PatientID int64
	Parity    int16 // 0=偶, 1=奇
}

// AssignHdfWeekParity 为HDF患者分配奇偶周(规范 §5)
// 已分配者保持固定, 新入组者分到负载较轻一侧
func AssignHdfWeekParity(profiles []PatientProfile) []ParityAssignment {
	type key struct {
		weekday int16
		parity  int16
	}
	load := map[key]int{}

	// 统计已有分配负载
	for _, p := range profiles {
		if p.HdfEnabled && p.HdfWeekday != nil && p.HdfWeekParity != nil {
			load[key{*p.HdfWeekday, *p.HdfWeekParity}]++
		}
	}

	var assignments []ParityAssignment
	for _, p := range profiles {
		if !p.HdfEnabled || p.HdfWeekday == nil {
			continue
		}
		if p.HdfWeekParity != nil {
			continue // 已分配, 保持不变
		}
		evenLoad := load[key{*p.HdfWeekday, 0}]
		oddLoad := load[key{*p.HdfWeekday, 1}]
		var chosen int16 = 0
		if evenLoad > oddLoad {
			chosen = 1
		}
		load[key{*p.HdfWeekday, chosen}]++
		assignments = append(assignments, ParityAssignment{PatientID: p.PatientID, Parity: chosen})
	}
	return assignments
}

// DecideMode 决策本次透析的治疗模式(规范 §5 HDF替换语义)
// HDF不是新增次数, 而是替换某一次HD
func DecideMode(anchor time.Time, hdfEnabled bool, hdfWeekday, hdfWeekParity *int16, date time.Time) string {
	if !hdfEnabled || hdfWeekday == nil {
		return ModeHD
	}
	wd := int16(date.Weekday())
	if *hdfWeekday != wd {
		return ModeHD // 不是HDF日
	}
	if hdfWeekParity == nil {
		return ModeHD // 未分配奇偶周
	}
	odd := IsOddWeek(anchor, date)
	// hdfWeekParity: 0=偶周做HDF, 1=奇周做HDF
	if (odd && *hdfWeekParity == 1) || (!odd && *hdfWeekParity == 0) {
		return ModeHDF
	}
	return ModeHD
}

// AnchorFromString 从配置字符串获取基准周一
func AnchorFromString(s string) time.Time {
	if s == "" {
		return DefaultAnchorMondayTime
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil || t.IsZero() {
		return DefaultAnchorMondayTime
	}
	return t
}
