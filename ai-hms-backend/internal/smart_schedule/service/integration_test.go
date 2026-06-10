package service

import (
	"os"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/repo"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/seed"
)

const testTenant int64 = 1

var (
	testDB   *gorm.DB
	testOnce sync.Once
)

var allTables = []string{
	"Schedule_PatientShift", "Schedule_CrrtSession", "Schedule_ConflictQueue",
	"Schedule_ScheduleTemplateItem", "Schedule_ScheduleTemplate", "Schedule_PlanChange",
	"Schedule_PatientProfile", "Schedule_Patient", "Schedule_Calendar",
	"Schedule_MachineOutage", "Schedule_Bed", "Schedule_Shift", "Schedule_Ward",
	"Schedule_TenantSetting",
}

func getDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("跳过集成测试:未设置 TEST_DATABASE_URL")
	}
	testOnce.Do(func() {
		g, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			t.Fatalf("连接测试库失败: %v", err)
		}
		if err := g.AutoMigrate(model.AllModels()...); err != nil {
			t.Fatalf("迁移失败: %v", err)
		}
		if err := repo.EnsureIndexes(g); err != nil {
			t.Fatalf("创建索引失败: %v", err)
		}
		testDB = g
	})
	return testDB
}

func setup(t *testing.T) (*gorm.DB, int64) {
	g := getDB(t)
	for _, tbl := range allTables {
		if err := g.Exec(`TRUNCATE TABLE "` + tbl + `" RESTART IDENTITY CASCADE`).Error; err != nil {
			t.Fatalf("清表 %s 失败: %v", tbl, err)
		}
	}
	if _, err := seed.Demo(g, testTenant); err != nil {
		t.Fatalf("seed 失败: %v", err)
	}
	return g, testTenant
}

func futureMonday() time.Time {
	return sched.MondayOf(time.Now().AddDate(0, 0, 7))
}

func TestGenerateConfirmCancelMakeup(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()

	res, err := GenerateSchedule(g, tenant, start, 2)
	if err != nil {
		t.Fatal(err)
	}
	if res.Drafts != 36 || res.Conflicts != 0 {
		t.Fatalf("生成应 36 草稿 0 冲突,得 %+v", res)
	}

	n, err := ConfirmPlan(g, tenant, 1, start, 2)
	if err != nil || n != 36 {
		t.Fatalf("整盘确认应 36,得 %d err=%v", n, err)
	}

	var s model.PatientShift
	if err := g.Where(`"TenantId"=? AND "PatientId"=? AND "TreatmentTime"=? AND "Status" NOT IN ?`,
		tenant, 1001, start, []int16{sched.StatusCancelled, sched.StatusAbsent}).First(&s).Error; err != nil {
		t.Fatalf("找不到 1001 周一排班: %v", err)
	}
	if err := CancelShift(g, tenant, s.Id, "测试请假"); err != nil {
		t.Fatalf("取消失败: %v", err)
	}

	diff, err := ComputeDiffs(g, tenant, start, 2)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, it := range diff.Items {
		if it.PatientId == 1001 {
			found = true
			if it.Diff != 1 {
				t.Errorf("1001 应少排 1,得 %d", it.Diff)
			}
		}
	}
	if !found {
		t.Error("差异检测未发现 1001 缺排")
	}

	mk, err := MakeupPatient(g, tenant, 1001, start, 2)
	if err != nil {
		t.Fatal(err)
	}
	if mk.Placed < 1 {
		t.Errorf("补透应至少补 1 次,得 %d", mk.Placed)
	}

	diff2, _ := ComputeDiffs(g, tenant, start, 2)
	for _, it := range diff2.Items {
		if it.PatientId == 1001 && it.Diff > 0 {
			t.Errorf("补透后 1001 仍缺 %d", it.Diff)
		}
	}
}

func TestGenerateIdempotent(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	if _, err := GenerateSchedule(g, tenant, start, 2); err != nil {
		t.Fatal(err)
	}
	res2, err := GenerateSchedule(g, tenant, start, 2)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Drafts != 0 {
		t.Errorf("重复生成应 0 草稿(幂等),得 %d", res2.Drafts)
	}
}

func TestHolidayDutyModePartialOpen(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	wed := start.AddDate(0, 0, 2)

	var cWard, aWard model.Ward
	g.Where(`"TenantId"=? AND "ZoneType"=?`, tenant, sched.ZoneC).First(&cWard)
	g.Where(`"TenantId"=? AND "ZoneType"=?`, tenant, sched.ZoneA).First(&aWard)

	if _, err := SetHoliday(g, tenant, wed, 20, itoa(cWard.Id)); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateSchedule(g, tenant, start, 2); err != nil {
		t.Fatal(err)
	}

	var aWed, aMon int64
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "WardId"=? AND "TreatmentTime"=? AND "Status" NOT IN ?`,
		tenant, aWard.Id, wed, []int16{sched.StatusCancelled}).Count(&aWed)
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "WardId"=? AND "TreatmentTime"=? AND "Status" NOT IN ?`,
		tenant, aWard.Id, start, []int16{sched.StatusCancelled}).Count(&aMon)
	if aWed != 0 {
		t.Errorf("值班模式只开C区,A区周三应 0 排班,得 %d", aWed)
	}
	if aMon == 0 {
		t.Error("A区周一应有排班(对照)")
	}
}

func TestPlanChangeCancelsFuture(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	if _, err := GenerateSchedule(g, tenant, start, 2); err != nil {
		t.Fatal(err)
	}
	res, err := ApplyPlanChange(g, tenant, 1003, "FREQ", itoa(int64(sched.FreqMonWedFri)), start)
	if err != nil {
		t.Fatal(err)
	}
	if res.Replanned == 0 {
		t.Errorf("方案变更应有未确认排班被重排,得 %d", res.Replanned)
	}
}

func TestAdminValidation(t *testing.T) {
	g, tenant := setup(t)

	if _, err := CreateWard(g, tenant, 1, &model.Ward{Name: "X", ZoneType: "X"}); err == nil {
		t.Error("非法 ZoneType 应被拒")
	}
	if _, err := CreateWard(g, tenant, 1, &model.Ward{Name: "D区", ZoneType: sched.ZoneA}); err != nil {
		t.Errorf("合法病区应成功: %v", err)
	}
	bad := &model.PatientProfile{PatientId: 8001, ZoneTag: sched.ZoneA, FreqPattern: sched.FreqMonWedFri,
		HdfEnabled: true, HdfWeekday: i16ptr(2)}
	if _, err := UpsertProfile(g, tenant, bad); err == nil {
		t.Error("HDF 日不在频率透析日内应被拒")
	}
}

func i16ptr(v int16) *int16 { return &v }

func TestBaseModeHFD(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	g.Model(&model.ScheduleTemplateItem{}).Where(`"TenantId"=? AND "PatientId"=?`, tenant, 1001).
		Update("DefaultMode", sched.ModeHFD)
	if _, err := GenerateSchedule(g, tenant, start, 2); err != nil {
		t.Fatal(err)
	}
	var s model.PatientShift
	if err := g.Where(`"TenantId"=? AND "PatientId"=? AND "TreatmentTime"=?`, tenant, 1001, start).First(&s).Error; err != nil {
		t.Fatalf("找不到 1001 排班: %v", err)
	}
	if s.DialysisMode != sched.ModeHFD {
		t.Errorf("1001 基础模式应为 HFD,得 %s", s.DialysisMode)
	}
	var m model.Machine
	g.Where(`"Id"=?`, s.MachineId).First(&m)
	if m.MachineType != sched.MachineHD {
		t.Errorf("HFD 应落在 HD 机,得 %s", m.MachineType)
	}
}

func TestDischargeStopsScheduling(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	if _, err := GenerateSchedule(g, tenant, start, 2); err != nil {
		t.Fatal(err)
	}
	if err := DischargePatient(g, tenant, 1, 1001, "出院"); err != nil {
		t.Fatal(err)
	}
	var active int64
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "PatientId"=? AND "Status" NOT IN ?`,
		tenant, 1001, []int16{sched.StatusCancelled, sched.StatusAbsent}).Count(&active)
	if active != 0 {
		t.Errorf("出组后 1001 不应有有效排班,得 %d", active)
	}
	if _, err := GenerateSchedule(g, tenant, start, 2); err != nil {
		t.Fatal(err)
	}
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "PatientId"=? AND "Status" NOT IN ?`,
		tenant, 1001, []int16{sched.StatusCancelled, sched.StatusAbsent}).Count(&active)
	if active != 0 {
		t.Errorf("出组病人重新生成仍不应被排,得 %d", active)
	}
}

func TestPlaceNewPatientMidCycle(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	var aWard model.Ward
	g.Where(`"TenantId"=? AND "ZoneType"=?`, tenant, sched.ZoneA).First(&aWard)
	var mor model.Shift
	g.Where(`"TenantId"=? AND "ShiftCode"=?`, tenant, "MORNING").First(&mor)

	if _, err := UpsertPatient(g, tenant, &model.Patient{Id: 3001, Name: "新病人", Gender: "男"}); err != nil {
		t.Fatal(err)
	}
	prof := &model.PatientProfile{PatientId: 3001, ZoneTag: sched.ZoneA, HomeWardId: &aWard.Id,
		FreqPattern: sched.FreqMonWedFri, WeeklyCount: 3, ShiftId: &mor.Id, DefaultMode: sched.ModeHD}
	if _, err := UpsertProfile(g, tenant, prof); err != nil {
		t.Fatal(err)
	}
	placed, _, err := PlaceNewPatientService(g, tenant, 3001, start, 2)
	if err != nil {
		t.Fatal(err)
	}
	if placed == 0 {
		t.Error("中途入组应排入至少 1 次")
	}
}

func TestWeeklyCountValidation(t *testing.T) {
	g, tenant := setup(t)
	bad := &model.PatientProfile{PatientId: 4001, ZoneTag: sched.ZoneA,
		WeeklyCount: 2, FreqPattern: sched.FreqMonWedFri}
	if _, err := UpsertProfile(g, tenant, bad); err == nil {
		t.Error("次数与星期组合不一致应被拒")
	}
}

func TestInfectionStatusAndWaive(t *testing.T) {
	g, tenant := setup(t)
	if err := SetInfectionStatus(g, tenant, 1001, sched.InfectionPositive); err != nil {
		t.Fatal(err)
	}
	var p model.Patient
	g.Where(`"TenantId"=? AND "Id"=?`, tenant, 1001).First(&p)
	if p.InfectionStatus != sched.InfectionPositive {
		t.Errorf("院感状态应为 positive,得 %s", p.InfectionStatus)
	}
	if err := WaiveInfection(g, tenant, 1, 1002); err != nil {
		t.Fatal(err)
	}
	var p2 model.Patient
	g.Where(`"TenantId"=? AND "Id"=?`, tenant, 1002).First(&p2)
	if p2.InfectionWaivedAt == nil {
		t.Error("豁免后应有 InfectionWaivedAt")
	}
}

func TestQualityMetrics(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	if _, err := GenerateSchedule(g, tenant, start, 2); err != nil {
		t.Fatal(err)
	}
	q, err := ComputeQuality(g, tenant, start, 2)
	if err != nil {
		t.Fatal(err)
	}
	if q.CapacitySlots <= 0 || q.UsedSlots <= 0 {
		t.Errorf("容量/已用应 > 0,得 cap=%d used=%d", q.CapacitySlots, q.UsedSlots)
	}
	if q.PatientsOnTarget == 0 {
		t.Error("应有达标病人")
	}
	if q.Score < 0 || q.Score > 100 {
		t.Errorf("综合分应 0-100,得 %d", q.Score)
	}
}
