package services

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
)

func makeBoardForTest() *ScheduleBoard {
	start := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	return &ScheduleBoard{
		StartDate: start,
		EndDate:   start.AddDate(0, 0, 6),
		Wards: []models.Ward{
			{Id: 1, TenantId: 3, Name: "病区A", IsDisabled: false},
			{Id: 2, TenantId: 3, Name: "病区B", IsDisabled: false},
		},
		WardExts: map[int64]models.WardExt{
			1: {WardId: 1, ZoneType: "A"},
		},
		Beds: []models.Bed{
			{Id: 1, TenantId: 3, Name: "床1-HD", WardId: ptrI64(1), IsDisabled: false},
			{Id: 2, TenantId: 3, Name: "床2-HDF", WardId: ptrI64(1), IsDisabled: false},
			{Id: 3, TenantId: 3, Name: "床3-HD", WardId: ptrI64(2), IsDisabled: false},
			{Id: 4, TenantId: 3, Name: "床4-停机", WardId: ptrI64(1), IsDisabled: false},
			{Id: 5, TenantId: 3, Name: "床5-无机器", WardId: ptrI64(1), IsDisabled: false},
			{Id: 6, TenantId: 3, Name: "床6-扩展停用", WardId: ptrI64(1), IsDisabled: false},
		},
		BedMachineExts: map[int64]models.BedMachineExt{
			1: {BedId: 1, MachineType: "HD", SupportedModes: "HD"},
			2: {BedId: 2, MachineType: "HDF", SupportedModes: "HD,HDF,HF"},
			3: {BedId: 3, MachineType: "HD", SupportedModes: "HD"},
			4: {BedId: 4, MachineType: "HD", SupportedModes: "HD"},
			6: {BedId: 6, MachineType: "HD", SupportedModes: "HD", IsDisabled: true},
		},
		Shifts: []models.Shift{
			{Id: 1, TenantId: 3, Name: "上午"},
			{Id: 2, TenantId: 3, Name: "下午"},
		},
		Profiles: map[int64]models.PatientProfile{
			100: {PatientId: 100, ZoneTag: "A", HomeWardId: ptrI64(1), FreqPattern: 10, HdfEnabled: true},
			101: {PatientId: 101, ZoneTag: "B", HomeWardId: ptrI64(2), FreqPattern: 20},
			102: {PatientId: 102, ZoneTag: "X", HomeWardId: ptrI64(1), FreqPattern: 10},
			103: {PatientId: 103, ZoneTag: "A", HomeWardId: ptrI64(999), FreqPattern: 10},
			104: {PatientId: 104, ZoneTag: "A", FixedHdBedId: ptrI64(999), FreqPattern: 10},
			105: {PatientId: 105, ZoneTag: "A", FixedHdfBedId: ptrI64(999), FreqPattern: 10},
			106: {PatientId: 106, ZoneTag: "A", FixedHdBedId: ptrI64(5), FreqPattern: 10},
			107: {PatientId: 107, ZoneTag: "A", HomeWardId: ptrI64(2), FreqPattern: 10},
			108: {PatientId: 108, ZoneTag: "A", FixedHdfBedId: ptrI64(1), FreqPattern: 10},
		},
		Patients: map[int64]ScheduleBoardPatient{
			100: {Id: 100, Name: "患者100"},
			101: {Id: 101, Name: "患者101"},
			102: {Id: 102, Name: "患者102"},
			103: {Id: 103, Name: "患者103"},
			104: {Id: 104, Name: "患者104"},
			105: {Id: 105, Name: "患者105"},
			106: {Id: 106, Name: "患者106"},
			107: {Id: 107, Name: "患者107"},
			108: {Id: 108, Name: "患者108"},
		},
		Occupancies:     map[int64][]ScheduleBoardOccupancy{},
		CalendarEntries: map[string]ScheduleBoardCalendar{},
		Outages:         []models.MachineOutage{},
	}
}

func ptrI64(v int64) *int64 { return &v }

func TestRunPrecheck_BedNoMachineType(t *testing.T) {
	board := makeBoardForTest()
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "BED_NO_MACHINE_TYPE" && iss.BedId != nil && *iss.BedId == 5 {
			found = true
		}
	}
	if !found {
		t.Error("应检测到床5未配置机器类型")
	}
}

func TestRunPrecheck_InvalidZoneTag(t *testing.T) {
	board := makeBoardForTest()
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "INVALID_ZONE_TAG" && iss.PatientId != nil && *iss.PatientId == 102 {
			found = true
		}
	}
	if !found {
		t.Error("应检测到患者102分区标签无效")
	}
}

func TestRunPrecheck_HomeWardNotFound(t *testing.T) {
	board := makeBoardForTest()
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "HOME_WARD_NOT_FOUND" && iss.PatientId != nil && *iss.PatientId == 103 {
			found = true
		}
	}
	if !found {
		t.Error("应检测到患者103 HomeWardId不存在")
	}
}

func TestRunPrecheck_HomeWardNoExt(t *testing.T) {
	board := makeBoardForTest()
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "HOME_WARD_NO_EXT" && iss.PatientId != nil && *iss.PatientId == 107 {
			found = true
		}
	}
	if !found {
		t.Error("应检测到患者107 HomeWardId存在但无WardExt")
	}
}

func TestRunPrecheck_FixedBedNotFound(t *testing.T) {
	board := makeBoardForTest()
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	hdFound, hdfFound := false, false
	for _, iss := range issues {
		if iss.Type == "FIXED_HD_BED_NOT_FOUND" && iss.PatientId != nil && *iss.PatientId == 104 {
			hdFound = true
		}
		if iss.Type == "FIXED_HDF_BED_NOT_FOUND" && iss.PatientId != nil && *iss.PatientId == 105 {
			hdfFound = true
		}
	}
	if !hdFound {
		t.Error("应检测到患者104固定HD床位不存在")
	}
	if !hdfFound {
		t.Error("应检测到患者105固定HDF床位不存在")
	}
}

func TestRunPrecheck_FixedHDBedModeMismatch(t *testing.T) {
	board := makeBoardForTest()
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "FIXED_HD_BED_NOT_FOUND" && iss.PatientId != nil && *iss.PatientId == 106 {
			found = true
		}
	}
	if !found {
		t.Error("应检测到患者106固定HD床位不存在(床5无机器类型)")
	}
}

func TestRunPrecheck_FixedHDFBedModeMismatch(t *testing.T) {
	board := makeBoardForTest()
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "FIXED_HDF_BED_MODE_MISMATCH" && iss.PatientId != nil && *iss.PatientId == 108 {
			found = true
		}
	}
	if !found {
		t.Error("应检测到患者108固定HDF床位不支持HDF(床1仅HD)")
	}
}

func TestRunPrecheck_HDFPatientNoMachine(t *testing.T) {
	board := makeBoardForTest()
	delete(board.BedMachineExts, 2)
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "HDF_PATIENT_NO_MACHINE" && iss.PatientId != nil && *iss.PatientId == 100 {
			found = true
		}
	}
	if !found {
		t.Error("应检测到HDF患者100没有可用HDF机器")
	}
}

func TestRunPrecheck_NonDialysisDay(t *testing.T) {
	board := makeBoardForTest()
	board.CalendarEntries["2026-01-06"] = ScheduleBoardCalendar{
		Calendar: models.Calendar{
			Id:            1,
			TenantId:      3,
			CalDate:       time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
			IsDialysisDay: false,
		},
	}
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "NON_DIALYSIS_DAY" && iss.Date == "2026-01-06" {
			found = true
		}
	}
	if !found {
		t.Error("应检测到非透析日")
	}
}

func TestRunPrecheck_MachineOutage(t *testing.T) {
	board := makeBoardForTest()
	board.Outages = []models.MachineOutage{
		{Id: 1, TenantId: 3, BedId: 1, StartAt: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: ptrI64(1), OutageType: 10, Reason: "维护"},
	}
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "MACHINE_OUTAGE" && iss.BedId != nil && *iss.BedId == 1 {
			found = true
		}
	}
	if !found {
		t.Error("应检测到设备停机")
	}
}

func TestRunPrecheck_BedConflict(t *testing.T) {
	board := makeBoardForTest()
	board.Occupancies[1] = []ScheduleBoardOccupancy{
		{PatientShiftId: 1, PatientId: 100, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 1, WardId: 1},
		{PatientShiftId: 2, PatientId: 101, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 1, WardId: 1},
	}
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "BED_CONFLICT" {
			found = true
		}
	}
	if !found {
		t.Error("应检测到床位冲突")
	}
}

func TestRunPrecheck_PatientConflict(t *testing.T) {
	board := makeBoardForTest()
	board.Occupancies[1] = []ScheduleBoardOccupancy{
		{PatientShiftId: 1, PatientId: 100, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 1, WardId: 1},
		{PatientShiftId: 2, PatientId: 100, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 2, WardId: 1},
	}
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "PATIENT_CONFLICT" {
			found = true
		}
	}
	if !found {
		t.Error("应检测到患者冲突")
	}
}

func TestRunPrecheck_CRRTNotConfigured(t *testing.T) {
	board := makeBoardForTest()
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "CRRT_MACHINE_NOT_CONFIGURED" {
			found = true
		}
	}
	if !found {
		t.Error("应检测到无CRRT机器配置")
	}
}

// ===================== 容量测试 =====================

func TestComputeCapacity_SkipNonDialysisDay(t *testing.T) {
	board := makeBoardForTest()
	board.CalendarEntries["2026-01-05"] = ScheduleBoardCalendar{
		Calendar: models.Calendar{
			Id: 1, TenantId: 3,
			CalDate:       time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			IsDialysisDay: false,
		},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-05" {
			t.Error("非透析日不应出现在容量中")
		}
	}
}

func TestComputeCapacity_OpenWardFilter(t *testing.T) {
	board := makeBoardForTest()
	board.CalendarEntries["2026-01-05"] = ScheduleBoardCalendar{
		Calendar: models.Calendar{
			Id: 1, TenantId: 3,
			CalDate:       time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			IsDialysisDay: true,
		},
		OpenWards: []int64{1},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-05" && s.WardId == 2 {
			t.Error("病区2不在OpenWards中，不应出现在容量里")
		}
	}
}

func TestComputeCapacity_OutageDeduct(t *testing.T) {
	board := makeBoardForTest()
	// 停机覆盖床1，日期为 2026-01-05，ShiftId=1
	board.Outages = []models.MachineOutage{
		{Id: 1, TenantId: 3, BedId: 1, StartAt: time.Date(2026, 1, 5, 8, 0, 0, 0, time.UTC), ShiftId: ptrI64(1), OutageType: 10, Reason: "维护"},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-05" && s.ShiftId == 1 && s.WardId == 1 {
			// 病区1: 床1停机,床2 HDF=计入,床4 HD=计入,床5无机器=不计,床6扩展停用=不计 → 2
			if s.Capacity != 2 {
				t.Errorf("容量应扣除停机+停用+无机器, 期望2, 实际=%d", s.Capacity)
			}
		}
	}
}

func TestComputeCapacity_OutageEndBoundary(t *testing.T) {
	board := makeBoardForTest()
	// 停机 StartAt=01-05 08:00, EndAt=01-06 00:00 → 应只扣减 01-05，不扣 01-06
	board.Outages = []models.MachineOutage{
		{Id: 1, TenantId: 3, BedId: 1, StartAt: time.Date(2026, 1, 5, 8, 0, 0, 0, time.UTC), EndAt: timePtr(time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)), ShiftId: ptrI64(1), OutageType: 10, Reason: "维护"},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-06" && s.ShiftId == 1 && s.WardId == 1 {
			// 01-06 不应扣除床1
			if s.Capacity != 3 {
				t.Errorf("01-06容量不应扣除床1, 期望3, 实际=%d", s.Capacity)
			}
		}
	}
}

func TestComputeCapacity_OccupationPerWard(t *testing.T) {
	board := makeBoardForTest()
	// 病区1有2条占用, 病区2有1条占用
	board.Occupancies[1] = []ScheduleBoardOccupancy{
		{PatientShiftId: 1, PatientId: 100, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 1, WardId: 1},
		{PatientShiftId: 2, PatientId: 101, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 2, WardId: 1},
		{PatientShiftId: 3, PatientId: 102, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 3, WardId: 2},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-05" && s.ShiftId == 1 {
			if s.WardId == 1 && s.Occupied != 2 {
				t.Errorf("病区1 Occupied 应为2, 实际=%d", s.Occupied)
			}
			if s.WardId == 2 && s.Occupied != 1 {
				t.Errorf("病区2 Occupied 应为1, 实际=%d", s.Occupied)
			}
		}
	}
}

func TestComputeCapacity_DisabledBedAndExt(t *testing.T) {
	board := makeBoardForTest()
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-05" && s.ShiftId == 1 && s.WardId == 1 {
			// 病区1 beds: 1=HD(计入), 2=HDF(计入), 4=HD(计入), 5=无机器(不计), 6=扩展停用(不计)
			// 无 WardExt 的 wardId=2 不计 zone
			if s.Capacity != 3 {
				t.Errorf("病区1容量应为3(排除床5无机器+床6扩展停用), 实际=%d", s.Capacity)
			}
		}
	}
}

func TestComputeCapacity_DateRangeLimit(t *testing.T) {
	board := makeBoardForTest()
	board.StartDate = time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	board.EndDate = time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	dates := map[string]bool{}
	for _, s := range slots {
		dates[s.Date] = true
	}
	if !dates["2026-01-05"] || !dates["2026-01-06"] || !dates["2026-01-07"] {
		t.Error("应包含所有日期")
	}
	if dates["2026-01-08"] {
		t.Error("不应超出结束日期")
	}
}

func TestComputeCapacity_OccupiedDeduct(t *testing.T) {
	board := makeBoardForTest()
	board.Occupancies[1] = []ScheduleBoardOccupancy{
		{PatientShiftId: 1, PatientId: 100, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 1, WardId: 1},
		{PatientShiftId: 2, PatientId: 101, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 2, WardId: 1},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-05" && s.ShiftId == 1 && s.WardId == 1 {
			if s.Occupied != 2 {
				t.Errorf("占用应为2, 实际=%d", s.Occupied)
			}
		}
	}
}

// ===================== 其他测试 =====================

func TestComputeCapacity_OutageCrossMidnight(t *testing.T) {
	board := makeBoardForTest()
	// 跨午夜：01-05 20:00 ~ 01-06 01:00，应扣 01-05 和 01-06 两天
	board.Outages = []models.MachineOutage{
		{Id: 1, TenantId: 3, BedId: 1, StartAt: time.Date(2026, 1, 5, 20, 0, 0, 0, time.UTC), EndAt: timePtr(time.Date(2026, 1, 6, 1, 0, 0, 0, time.UTC)), ShiftId: ptrI64(1), OutageType: 10, Reason: "维护"},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	day5cap := -1
	day6cap := -1
	for _, s := range slots {
		if s.Date == "2026-01-05" && s.ShiftId == 1 && s.WardId == 1 {
			day5cap = s.Capacity
		}
		if s.Date == "2026-01-06" && s.ShiftId == 1 && s.WardId == 1 {
			day6cap = s.Capacity
		}
	}
	// 病区1: 3 beds 计入，床1停机 → 2（01-05和01-06都扣）
	if day5cap != 2 {
		t.Errorf("01-05容量应扣除跨午夜停机, 期望2, 实际=%d", day5cap)
	}
	if day6cap != 2 {
		t.Errorf("01-06容量应扣除跨午夜停机, 期望2, 实际=%d", day6cap)
	}
}

func TestComputeCapacity_WardIdZeroFallback(t *testing.T) {
	board := makeBoardForTest()
	// WardId=0, BedId=3 → 床3属于WardId=2，应将占用计入WardId=2
	board.Occupancies[1] = []ScheduleBoardOccupancy{
		{PatientShiftId: 1, PatientId: 100, Date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), ShiftId: 1, BedId: 3, WardId: 0},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-05" && s.ShiftId == 1 {
			if s.WardId == 2 && s.Occupied != 1 {
				t.Errorf("WardId=0/BedId=3 应向 WardId=2 统计, 期望Occupied=1, 实际=%d", s.Occupied)
			}
			if s.WardId == 0 {
				t.Error("WardId=0 的占用应通过 BedId 反查归属到实际病区")
			}
		}
	}
}

func TestRunPrecheck_CRRTWardNotConfigured(t *testing.T) {
	board := makeBoardForTest()
	board.WardExts = map[int64]models.WardExt{}
	issues := (&ScheduleBoardService{}).RunPrecheck(board)
	found := false
	for _, iss := range issues {
		if iss.Type == "CRRT_WARD_NOT_CONFIGURED" {
			found = true
		}
	}
	if !found {
		t.Error("WardExts为空时应返回CRRT_WARD_NOT_CONFIGURED")
	}
}

func TestComputeCapacity_OutageEndAtMidnight(t *testing.T) {
	board := makeBoardForTest()
	// EndAt=01-06 00:00:00，应只扣 01-05，不扣 01-06
	board.Outages = []models.MachineOutage{
		{Id: 1, TenantId: 3, BedId: 1, StartAt: time.Date(2026, 1, 5, 8, 0, 0, 0, time.UTC), EndAt: timePtr(time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)), ShiftId: ptrI64(1), OutageType: 10, Reason: "维护"},
	}
	slots := (&ScheduleBoardService{}).ComputeCapacity(board)
	for _, s := range slots {
		if s.Date == "2026-01-05" && s.ShiftId == 1 && s.WardId == 1 {
			// 病区1正常3 beds，床1停机 → 2
			if s.Capacity != 2 {
				t.Errorf("01-05容量应为2(扣除床1), 实际=%d", s.Capacity)
			}
		}
		if s.Date == "2026-01-06" && s.ShiftId == 1 && s.WardId == 1 {
			// 01-06 不应扣除床1
			if s.Capacity != 3 {
				t.Errorf("01-06容量应为3(不扣床1), 实际=%d", s.Capacity)
			}
		}
	}
}
func TestParseScheduleDate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2026-01-05", false},
		{"2026-01-05T08:00:00Z", false},
		{"invalid", true},
		{"", true},
	}
	for _, tt := range tests {
		_, err := ParseScheduleDate(tt.input)
		if tt.wantErr && err == nil {
			t.Errorf("ParseScheduleDate(%q) 期望错误", tt.input)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ParseScheduleDate(%q) 不期望错误: %v", tt.input, err)
		}
	}
}

func TestValidateMachineTypeAndModes_Regression(t *testing.T) {
	got, err := validateMachineTypeAndModes("HDF", "HD,CRRT")
	if err == nil {
		t.Error("HDF 机器不应接受 CRRT 模式")
	}
	if got != "" {
		t.Errorf("错误时返回值应为空, 实际=%q", got)
	}
}

func TestModeSupports(t *testing.T) {
	if !modeSupports("HD,HDF,HF", "HD") {
		t.Error("modeSupports应返回true")
	}
	if modeSupports("HD", "HDF") {
		t.Error("modeSupports(HD, HDF)应返回false")
	}
	if !modeSupports("hd , HDF", "HD") {
		t.Error("normalizeModes大小写应归一化, hd -> HD 应匹配")
	}
	if !modeSupports("hd , hdf", "HDF") {
		t.Error("normalizeModes应归一化并支持HDF")
	}
}

func TestLoadBoardInvalidParams(t *testing.T) {
	svc := &ScheduleBoardService{}
	_, err := svc.LoadBoard(0, time.Now(), time.Now())
	if err == nil {
		t.Error("tenantID=0 应报错")
	}

	_, err = svc.LoadBoard(1, time.Time{}, time.Now())
	if err == nil {
		t.Error("start 为零值应报错")
	}

	_, err = svc.LoadBoard(1, time.Now(), time.Now().AddDate(0, 0, -1))
	if err == nil {
		t.Error("end 早于 start 应报错")
	}
}

func TestLoadBoardDbNilLast(t *testing.T) {
	svc := &ScheduleBoardService{}
	_, err := svc.LoadBoard(1, time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Error("db=nil 应报错")
	}
	if err.Error() != "database not available" {
		t.Errorf("期望 database not available, 实际=%q", err.Error())
	}
}

func timePtr(t time.Time) *time.Time { return &t }
