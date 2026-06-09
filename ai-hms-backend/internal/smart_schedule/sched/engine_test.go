package sched

import (
	"testing"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
)

func p64(v int64) *int64 { return &v }

func buildBoard() *Board {
	wards := []*model.Ward{
		{BaseModel: model.BaseModel{Id: 1, TenantId: 1}, Name: "A区", ZoneType: ZoneA},
	}
	shifts := []*model.Shift{
		{BaseModel: model.BaseModel{Id: 10, TenantId: 1}, Name: "上午班", ShiftCode: "MORNING", Sort: 1},
		{BaseModel: model.BaseModel{Id: 11, TenantId: 1}, Name: "下午班", ShiftCode: "AFTERNOON", Sort: 2},
		{BaseModel: model.BaseModel{Id: 12, TenantId: 1}, Name: "晚班", ShiftCode: "NIGHT", Sort: 3},
	}
	machines := []*model.Machine{
		{BaseModel: model.BaseModel{Id: 101, TenantId: 1}, WardId: 1, Code: "A-01", MachineType: MachineHD, PositionIndex: 1},
		{BaseModel: model.BaseModel{Id: 102, TenantId: 1}, WardId: 1, Code: "A-02", MachineType: MachineHD, PositionIndex: 2},
		{BaseModel: model.BaseModel{Id: 103, TenantId: 1}, WardId: 1, Code: "A-03", MachineType: MachineHDF, PositionIndex: 3},
	}
	return NewBoard(anchor, wards, machines, shifts, nil, nil, nil)
}

func TestGenerateTwoRoundAndHDF(t *testing.T) {
	b := buildBoard()
	e := NewEngine(b)
	items := []*model.ScheduleTemplateItem{
		{BaseModel: model.BaseModel{Id: 1, TenantId: 1}, PatientId: 1001, ZoneTag: ZoneA,
			WardId: p64(1), ShiftId: p64(10), FreqPattern: FreqMonWedFri},
		{BaseModel: model.BaseModel{Id: 2, TenantId: 1}, PatientId: 1002, ZoneTag: ZoneA,
			WardId: p64(1), ShiftId: p64(10), FreqPattern: FreqMonWedFri,
			HdfEnabled: true, HdfWeekday: i16(3)},
	}
	dates := e.ExpandDialysisDates(day(2025, 1, 6), 1)
	e.Generate(items, dates)

	if len(b.Conflicts) != 0 {
		t.Fatalf("不应有冲突,得 %d 条:%+v", len(b.Conflicts), b.Conflicts)
	}
	if len(b.Drafts) != 6 {
		t.Fatalf("应生成 6 条草稿,得 %d", len(b.Drafts))
	}

	wed := find(b.Drafts, 1002, day(2025, 1, 8))
	if wed == nil {
		t.Fatal("缺 1002 周三排班")
	}
	if wed.DialysisMode != ModeHDF {
		t.Errorf("1002 周三应 HDF,得 %s", wed.DialysisMode)
	}
	if wed.MachineId == nil || *wed.MachineId != 103 {
		t.Errorf("1002 周三应在 HDF 机 103,得 %v", wed.MachineId)
	}

	mon := find(b.Drafts, 1002, day(2025, 1, 6))
	if mon == nil || mon.DialysisMode != ModeHD {
		t.Fatalf("1002 周一应 HD,得 %+v", mon)
	}
	if mon.MachineId == nil || (*mon.MachineId != 101 && *mon.MachineId != 102) {
		t.Errorf("1002 周一应在 HD 机,得 %v", mon.MachineId)
	}
}

func TestOverflowToHDFWhenHDFull(t *testing.T) {
	b := buildBoard()
	e := NewEngine(b)
	var items []*model.ScheduleTemplateItem
	for k := 0; k < 3; k++ {
		items = append(items, &model.ScheduleTemplateItem{
			BaseModel:   model.BaseModel{Id: int64(k + 1), TenantId: 1},
			PatientId:   int64(2001 + k), ZoneTag: ZoneA,
			WardId:      p64(1), ShiftId: p64(10), FreqPattern: FreqOnePerWk,
		})
	}
	dates := e.ExpandDialysisDates(day(2025, 1, 6), 1)
	e.Generate(items, dates)

	if len(b.Conflicts) != 0 {
		t.Fatalf("3 人 3 台机不应冲突,得 %+v", b.Conflicts)
	}
	thu := day(2025, 1, 9)
	used := map[int64]bool{}
	for _, d := range b.Drafts {
		if dkey(d.ScheduleDate) == dkey(thu) && d.MachineId != nil {
			used[*d.MachineId] = true
		}
	}
	if !used[103] {
		t.Errorf("HD 机满后第 3 人应溢出到 HDF 机 103,实际占用 %v", used)
	}
}

func TestAssignHdfWeekParityBalances(t *testing.T) {
	items := []*model.ScheduleTemplateItem{
		{BaseModel: model.BaseModel{Id: 1}, PatientId: 1, HdfEnabled: true, HdfWeekday: i16(3)},
		{BaseModel: model.BaseModel{Id: 2}, PatientId: 2, HdfEnabled: true, HdfWeekday: i16(3)},
	}
	e := NewEngine(buildBoard())
	e.AssignHdfWeekParity(items)
	if items[0].HdfWeekParity == nil || items[1].HdfWeekParity == nil {
		t.Fatal("应为两人都分配奇偶周")
	}
	if *items[0].HdfWeekParity == *items[1].HdfWeekParity {
		t.Errorf("同 HDF 日两人应错峰到不同奇偶周,得 %d/%d", *items[0].HdfWeekParity, *items[1].HdfWeekParity)
	}
}

func find(drafts []*model.PatientShift, patientID int64, d time.Time) *model.PatientShift {
	for _, s := range drafts {
		if s.PatientId == patientID && dkey(s.ScheduleDate) == dkey(d) {
			return s
		}
	}
	return nil
}
