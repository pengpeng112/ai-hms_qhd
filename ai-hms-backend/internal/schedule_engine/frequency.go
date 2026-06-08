package schedule_engine

import "time"

// FrequencyWeekdays 返回频率模式对应的透析日集合(规范 §2.1)
func FrequencyWeekdays(freq int16) []time.Weekday {
	switch freq {
	case FreqMonWedFri:
		return []time.Weekday{time.Monday, time.Wednesday, time.Friday}
	case FreqTueThuSat:
		return []time.Weekday{time.Tuesday, time.Thursday, time.Saturday}
	case FreqTwoPerWk:
		return []time.Weekday{time.Tuesday, time.Thursday}
	case FreqOnePerWk:
		return []time.Weekday{time.Thursday}
	case FreqTemporary:
		return nil
	default:
		return nil
	}
}

// IsDue 判断给定日期是否属于该频率的透析日
func IsDue(freqPattern int16, date time.Time) bool {
	wd := date.Weekday()
	for _, d := range FrequencyWeekdays(freqPattern) {
		if wd == d {
			return true
		}
	}
	return false
}

// FrequencyWeeklyCount 返回频率模式对应的每周透析次数
func FrequencyWeeklyCount(freq int16) int {
	switch freq {
	case FreqMonWedFri, FreqTueThuSat:
		return 3
	case FreqTwoPerWk:
		return 2
	case FreqOnePerWk:
		return 1
	default:
		return 0
	}
}

// ExpandDialysisDates 从 start 起展开 weeks 周内的透析日(跳过非透析日)
func ExpandDialysisDates(start time.Time, weeks int, isDialysisDay func(time.Time) bool) []time.Time {
	var dates []time.Time
	for i := 0; i < weeks*7; i++ {
		d := dateOnly(start).AddDate(0, 0, i)
		if isDialysisDay == nil || isDialysisDay(d) {
			dates = append(dates, d)
		}
	}
	return dates
}

// dateOnly 截断到日期精度
func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// WeekNumber 计算从 anchor 起第几周
func WeekNumber(anchor, date time.Time) int {
	days := int(dateOnly(date).Sub(dateOnly(anchor)).Hours() / 24)
	if days < 0 {
		return days / 7 // 负数结果向下取整
	}
	return days / 7
}

// IsOddWeek 判断是否为奇数周(基于 anchor)
func IsOddWeek(anchor, date time.Time) bool {
	return WeekNumber(anchor, date)%2 == 0
}
