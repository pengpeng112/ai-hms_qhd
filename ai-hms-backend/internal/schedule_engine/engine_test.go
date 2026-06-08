package schedule_engine

import (
	"testing"
	"time"
)

func TestFrequencyWeekdays(t *testing.T) {
	tests := []struct {
		freq     int16
		expected int
		name     string
	}{
		{FreqMonWedFri, 3, "一三五"},
		{FreqTueThuSat, 3, "二四六"},
		{FreqTwoPerWk, 2, "每周两次"},
		{FreqOnePerWk, 1, "每周一次"},
		{FreqTemporary, 0, "临时"},
	}
	for _, tt := range tests {
		days := FrequencyWeekdays(tt.freq)
		if len(days) != tt.expected {
			t.Errorf("%s: expected %d days, got %d", tt.name, tt.expected, len(days))
		}
	}
}

func TestIsDue(t *testing.T) {
	mon := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC) // Monday
	wed := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC) // Wednesday
	tue := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC) // Tuesday
	thu := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC) // Thursday
	sun := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC) // Sunday

	if !IsDue(FreqMonWedFri, mon) {
		t.Error("周一应该在周一三五频率中")
	}
	if !IsDue(FreqMonWedFri, wed) {
		t.Error("周三应该在周一三五频率中")
	}
	if IsDue(FreqMonWedFri, tue) {
		t.Error("周二不应该在周一三五频率中")
	}
	if !IsDue(FreqTwoPerWk, tue) {
		t.Error("周二应该在每周两次频率中")
	}
	if !IsDue(FreqTwoPerWk, thu) {
		t.Error("周四应该在每周两次频率中")
	}
	if IsDue(FreqTemporary, mon) {
		t.Error("临时频率不应该在任何日期匹配")
	}
	if IsDue(FreqOnePerWk, thu) == false {
		t.Error("每周一次应该匹配周四")
	}
	if IsDue(FreqOnePerWk, sun) {
		t.Error("每周一次不应该匹配周日")
	}
}

func TestFrequencyWeeklyCount(t *testing.T) {
	if n := FrequencyWeeklyCount(FreqMonWedFri); n != 3 {
		t.Errorf("一三五每周%d次", n)
	}
	if n := FrequencyWeeklyCount(FreqTwoPerWk); n != 2 {
		t.Errorf("每周两次=%d", n)
	}
	if n := FrequencyWeeklyCount(FreqTemporary); n != 0 {
		t.Errorf("临时=%d", n)
	}
}

func TestIsOddWeek(t *testing.T) {
	anchor := DefaultAnchorMondayTime // 2025-01-06 Monday
	d1 := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)  // Week 0
	d2 := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC) // Week 1
	d3 := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC) // Week 2

	if !IsOddWeek(anchor, d1) {
		t.Error("Week 0 should be odd")
	}
	if IsOddWeek(anchor, d2) {
		t.Error("Week 1 should be even")
	}
	if !IsOddWeek(anchor, d3) {
		t.Error("Week 2 should be odd")
	}
}

func TestDecideMode(t *testing.T) {
	anchor := DefaultAnchorMondayTime
	// 2026-06-01 is Monday, Week 73 (73%2=1 -> not odd, even week)
	mon := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	// 2025-01-06 is Monday, Week 0 (odd)
	oddMon := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)

	// HDF patient with HDF on Monday, odd parity, on odd week → HDF
	hdfWD := int16(1)
	oddParity := int16(1)
	mode := DecideMode(anchor, true, &hdfWD, &oddParity, oddMon)
	if mode != ModeHDF {
		t.Errorf("Expected HDF on HDF day+odd week, got %s", mode)
	}

	// Non-HDF patient
	mode = DecideMode(anchor, false, nil, nil, mon)
	if mode != ModeHD {
		t.Errorf("Expected HD for non-HDF patient, got %s", mode)
	}

	// HDF patient on wrong weekday
	hdfWD = int16(2) // Tuesday
	mode = DecideMode(anchor, true, &hdfWD, &oddParity, oddMon)
	if mode != ModeHD {
		t.Errorf("Expected HD on non-HDF weekday, got %s", mode)
	}

	// HDF patient, right weekday, even parity on even week → HDF
	// June 2 2026 is Tuesday, Week 73 (even week)
	tueEven := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	hdfWD = int16(2) // Tuesday
	evenParity := int16(0) // even week = HDF
	mode = DecideMode(anchor, true, &hdfWD, &evenParity, tueEven)
	if mode != ModeHDF {
		t.Errorf("Expected HDF with evenParity on even week, got %s", mode)
	}
}

func TestMachineSupports(t *testing.T) {
	if !MachineSupports("HD", "HD") {
		t.Error("HD machine should support HD")
	}
	if MachineSupports("HD", "HDF") {
		t.Error("HD machine should NOT support HDF")
	}
	if !MachineSupports("HDF", "HDF") {
		t.Error("HDF machine should support HDF")
	}
	if !MachineSupports("HDF", "HD") {
		t.Error("HDF machine should support HD")
	}
	if MachineSupports("CRRT", "HD") {
		t.Error("CRRT machine should NOT support HD")
	}
}

func TestExpandDialysisDates(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC) // Monday
	dates := ExpandDialysisDates(start, 2, func(t time.Time) bool {
		return t.Weekday() != time.Sunday
	})
	if len(dates) != 12 {
		t.Errorf("2 weeks excluding Sundays: expected 12 days, got %d", len(dates))
	}
}

func TestCellKey(t *testing.T) {
	cell := Cell{1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 1}
	key := cell.Key()
	if key != "2026-06-01-1-1" {
		t.Errorf("Cell key mismatch: %s", key)
	}
}

func TestIsOccupied(t *testing.T) {
	occupied := map[int64][]Occupancy{
		1: {{BedID: 1, Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), ShiftID: 1}},
	}
	cell := Cell{1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 1}
	if !IsOccupied(occupied, 1, cell) {
		t.Error("Bed 1 should be occupied")
	}
	cell2 := Cell{1, time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC), 1}
	if IsOccupied(occupied, 1, cell2) {
		t.Error("Bed 1 should NOT be occupied on different date")
	}
}

func TestFindFreeBeds(t *testing.T) {
	beds := []BedInfo{
		{ID: 1, WardID: 1, MachineType: "HD", PositionIndex: 0},
		{ID: 2, WardID: 1, MachineType: "HDF", PositionIndex: 1},
		{ID: 3, WardID: 2, MachineType: "HD", PositionIndex: 2},
	}
	occupied := map[int64][]Occupancy{
		1: {{BedID: 1, Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), ShiftID: 1}},
	}
	cell := Cell{1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 1}
	free := FindFreeBeds(beds, occupied, 1, cell, "HD")
	if len(free) != 1 {
		t.Errorf("Expected 1 free bed (HDF supports HD), got %d", len(free))
	}
}

func TestPlaceHdSession(t *testing.T) {
	beds := []BedInfo{
		{ID: 1, WardID: 1, MachineType: "HD", PositionIndex: 0},
		{ID: 2, WardID: 1, MachineType: "HD", PositionIndex: 1},
		{ID: 3, WardID: 1, MachineType: "HDF", PositionIndex: 2},
	}
	occupied := map[int64][]Occupancy{}
	cell := Cell{1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 1}

	fixedHD := int64(1)
	session := SessionItem{PatientID: 100, FixedHdBedID: &fixedHD, Mode: ModeHD}
	bed := PlaceHdSession(beds, occupied, 1, cell, session)
	if bed == nil || bed.ID != 1 {
		t.Error("Should pick fixed HD bed")
	}

	// Occupy bed 1
	occupied[1] = []Occupancy{{BedID: 1, Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), ShiftID: 1}}
	bed = PlaceHdSession(beds, occupied, 1, cell, session)
	if bed == nil || bed.ID != 2 {
		t.Error("Should pick next HD bed when fixed is occupied")
	}
}
