package service

import (
	"os"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/db"
	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
	"github.com/sdsph/dialysis-scheduling/internal/seed"
)

// 业务层集成测试。需真实 PostgreSQL:设置环境变量 TEST_DATABASE_URL 后运行,例如:
//   set TEST_DATABASE_URL=host=127.0.0.1 user=postgres password=postgres dbname=aihms_test port=5432 sslmode=disable
//   go test ./internal/service/ -v
// 未设置则自动跳过(不影响无库环境的 go test ./...)。
// 各用例使用独立 tenant 隔离,无需清库。

const testTenant int64 = 1

var (
	testDB   *gorm.DB
	testOnce sync.Once
)

// allTables 所有 Schedule_* 表(供清库)。
var allTables = []string{
	"Schedule_PatientShift", "Schedule_CrrtSession", "Schedule_ConflictQueue",
	"Schedule_ScheduleTemplateItem", "Schedule_ScheduleTemplate", "Schedule_PlanChange",
	"Schedule_PatientProfile", "Schedule_Patient", "Schedule_Calendar",
	"Schedule_MachineOutage", "Schedule_Machine", "Schedule_Shift", "Schedule_Ward",
	"Schedule_TenantSetting",
}

func getDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("跳过集成测试:未设置 TEST_DATABASE_URL")
	}
	testOnce.Do(func() {
		g, err := db.Open(dsn)
		if err != nil {
			t.Fatalf("连接测试库失败: %v", err)
		}
		if err := db.Migrate(g); err != nil {
			t.Fatalf("迁移失败: %v", err)
		}
		testDB = g
	})
	return testDB
}

// setup 取库 + 清空全部表 + 写入演示数据(用例顺序执行,清库保证互不干扰)。
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
	return sched.MondayOf(time.Now().AddDate(0, 0, 7)) // 下周一,保证可编辑(非历史)
}

// 生成 → 整盘确认 → 取消 → 差异 → 补透 全链路。
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

	// 取消 1001 在 start(周一,一三五)的排班
	var s model.PatientShift
	if err := g.Where(`"TenantId"=? AND "PatientId"=? AND "ScheduleDate"=? AND "Status" NOT IN ?`,
		tenant, 1001, start, []int16{sched.StatusCancelled, sched.StatusAbsent}).First(&s).Error; err != nil {
		t.Fatalf("找不到 1001 周一排班: %v", err)
	}
	if err := CancelShift(g, tenant, s.Id, "测试请假"); err != nil {
		t.Fatalf("取消失败: %v", err)
	}

	// 差异检测:1001 应少 1
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

	// 补透:应补回至少 1 次
	mk, err := MakeupPatient(g, tenant, 1001, start, 2)
	if err != nil {
		t.Fatal(err)
	}
	if mk.Placed < 1 {
		t.Errorf("补透应至少补 1 次,得 %d", mk.Placed)
	}

	// 再查差异:1001 不应再缺
	diff2, _ := ComputeDiffs(g, tenant, start, 2)
	for _, it := range diff2.Items {
		if it.PatientId == 1001 && it.Diff > 0 {
			t.Errorf("补透后 1001 仍缺 %d", it.Diff)
		}
	}
}

// 生成幂等:同范围重复生成不应新增草稿。
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

// 假日值班部分开放:仅开 C 区时,A 区当天无排班。
func TestHolidayDutyModePartialOpen(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	wed := start.AddDate(0, 0, 2) // 周三(一三五病人当天有排)

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
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "WardId"=? AND "ScheduleDate"=? AND "Status" NOT IN ?`,
		tenant, aWard.Id, wed, []int16{sched.StatusCancelled}).Count(&aWed)
	g.Model(&model.PatientShift{}).Where(`"TenantId"=? AND "WardId"=? AND "ScheduleDate"=? AND "Status" NOT IN ?`,
		tenant, aWard.Id, start, []int16{sched.StatusCancelled}).Count(&aMon)
	if aWed != 0 {
		t.Errorf("值班模式只开C区,A区周三应 0 排班,得 %d", aWed)
	}
	if aMon == 0 {
		t.Error("A区周一应有排班(对照)")
	}
}

// 方案变更:生效日后未确认排班被取消待重排。
func TestPlanChangeCancelsFuture(t *testing.T) {
	g, tenant := setup(t)
	start := futureMonday()
	if _, err := GenerateSchedule(g, tenant, start, 2); err != nil {
		t.Fatal(err)
	}
	// 1003 二四六 → 改为一三五,生效日=start
	res, err := ApplyPlanChange(g, tenant, 1003, "FREQ", itoa(int64(sched.FreqMonWedFri)), start)
	if err != nil {
		t.Fatal(err)
	}
	if res.Replanned == 0 {
		t.Errorf("方案变更应有未确认排班被重排,得 %d", res.Replanned)
	}
}

// 管理录入校验:非法分区/非法 HDF 日被拒,合法通过。
func TestAdminValidation(t *testing.T) {
	g, tenant := setup(t)

	if _, err := CreateWard(g, tenant, 1, &model.Ward{Name: "X", ZoneType: "X"}); err == nil {
		t.Error("非法 ZoneType 应被拒")
	}
	if _, err := CreateWard(g, tenant, 1, &model.Ward{Name: "D区", ZoneType: sched.ZoneA}); err != nil {
		t.Errorf("合法病区应成功: %v", err)
	}
	// 一三五病人把 HDF 日设周二(非透析日)→ 应拒
	bad := &model.PatientProfile{PatientId: 8001, ZoneTag: sched.ZoneA, FreqPattern: sched.FreqMonWedFri,
		HdfEnabled: true, HdfWeekday: i16ptr(2)}
	if _, err := UpsertProfile(g, tenant, bad); err == nil {
		t.Error("HDF 日不在频率透析日内应被拒")
	}
}

func i16ptr(v int16) *int16 { return &v }

// 质量评分:生成后各指标合理。
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
