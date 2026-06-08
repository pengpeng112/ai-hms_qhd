package service

import (
	"math"
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/config"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/repo"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

// 排班质量评分(P2 / 工程师反馈 #22):给排班一个可量化的体检报告。

// QualityResult 质量评分结果。
type QualityResult struct {
	WeekStart        string  `json:"weekStart"`
	Weeks            int     `json:"weeks"`
	PatientsTotal    int     `json:"patientsTotal"`    // 有应排的病人数
	PatientsOnTarget int     `json:"patientsOnTarget"` // 应排=已排 的病人数
	OnTargetRate     float64 `json:"onTargetRate"`     // 达标率 0~1
	CapacitySlots    int     `json:"capacitySlots"`    // 可用透析位 = (HD+HDF机) × 班次 × 透析日
	UsedSlots        int     `json:"usedSlots"`        // 已占用透析位
	Utilization      float64 `json:"utilization"`      // 机器利用率 0~1
	PatientsScheduled int    `json:"patientsScheduled"`// 有排班的病人数
	SingleMachine    int     `json:"singleMachine"`    // 全部排班都在同一台机的病人数
	StabilityRate    float64 `json:"stabilityRate"`    // 位置稳定率 0~1
	OpenConflicts    int     `json:"openConflicts"`    // 待处理冲突数
	Score            int     `json:"score"`            // 综合分 0~100(达标率60% + 稳定率40%)
}

func round2(f float64) float64 { return math.Round(f*100) / 100 }

// ComputeQuality 计算 [weekStart,+weeks] 的排班质量指标。
func ComputeQuality(g *gorm.DB, tenant int64, weekStart time.Time, weeks int) (*QualityResult, error) {
	anchor := config.AnchorMonday(g, tenant)
	start := dayStart(weekStart)
	end := start.AddDate(0, 0, weeks*7)

	board, err := repo.LoadBoard(g, tenant, anchor, start, end)
	if err != nil {
		return nil, err
	}
	dialysisDays := 0
	var dates []time.Time
	for i := 0; i < weeks*7; i++ {
		d := start.AddDate(0, 0, i)
		if board.IsDialysisDay(d) {
			dialysisDays++
			dates = append(dates, d)
		}
	}

	res := &QualityResult{WeekStart: start.Format("2006-01-02"), Weeks: weeks}

	// 达标率:逐病人比应排 vs 已排。
	var profiles []model.PatientProfile
	if err := g.Where(`"TenantId" = ? AND "IsAdmissionRejected" = false`, tenant).Find(&profiles).Error; err != nil {
		return nil, err
	}
	activeStatuses := []int16{sched.StatusDraft, sched.StatusConfirmed, sched.StatusInDialysis, sched.StatusCompleted, sched.StatusAbsent}
	type cnt struct {
		PatientId int64
		N         int
		DM        int
	}
	var rows []cnt
	g.Model(&model.PatientShift{}).
		Select(`"PatientId" AS patient_id, count(*) AS n, count(distinct "MachineId") AS dm`).
		Where(`"TenantId" = ? AND "ScheduleDate" >= ? AND "ScheduleDate" < ? AND "RecordForm" = ? AND "Status" IN ? AND "ShiftId" IS NOT NULL`,
			tenant, start, end, sched.RecordFormRegular, activeStatuses).
		Group(`"PatientId"`).Scan(&rows)
	schedCount := map[int64]int{}
	distinctMachine := map[int64]int{}
	for _, r := range rows {
		schedCount[r.PatientId] = r.N
		distinctMachine[r.PatientId] = r.DM
	}

	for _, p := range profiles {
		if p.FreqPattern == sched.FreqTemporary {
			continue
		}
		expected := 0
		for _, d := range dates {
			if sched.IsDue(p.FreqPattern, d) {
				expected++
			}
		}
		if expected == 0 {
			continue
		}
		res.PatientsTotal++
		if schedCount[p.PatientId] == expected {
			res.PatientsOnTarget++
		}
	}

	// 位置稳定率:有排班的病人里,全部排在同一台机的占比。
	for pid, dm := range distinctMachine {
		if schedCount[pid] > 0 {
			res.PatientsScheduled++
			if dm == 1 {
				res.SingleMachine++
			}
		}
	}

	// 机器利用率:容量 =(HD+HDF 机)× 班次 × 透析日。
	var machineCnt, shiftCnt int64
	g.Model(&model.Machine{}).Where(`"TenantId" = ? AND "IsDisabled" = false AND "MachineType" IN ?`,
		tenant, []string{sched.MachineHD, sched.MachineHDF}).Count(&machineCnt)
	g.Model(&model.Shift{}).Where(`"TenantId" = ? AND "IsDisabled" = false`, tenant).Count(&shiftCnt)
	res.CapacitySlots = int(machineCnt) * int(shiftCnt) * dialysisDays
	var used int64
	g.Model(&model.PatientShift{}).
		Where(`"TenantId" = ? AND "ScheduleDate" >= ? AND "ScheduleDate" < ? AND "RecordForm" = ? AND "Status" IN ? AND "MachineId" IS NOT NULL AND "ShiftId" IS NOT NULL`,
			tenant, start, end, sched.RecordFormRegular, activeStatuses).Count(&used)
	res.UsedSlots = int(used)

	// 待处理冲突。
	var conf int64
	g.Model(&model.ConflictQueue{}).Where(`"TenantId" = ? AND "Status" = 0`, tenant).Count(&conf)
	res.OpenConflicts = int(conf)

	// 比率与综合分。
	if res.PatientsTotal > 0 {
		res.OnTargetRate = round2(float64(res.PatientsOnTarget) / float64(res.PatientsTotal))
	}
	if res.CapacitySlots > 0 {
		res.Utilization = round2(float64(res.UsedSlots) / float64(res.CapacitySlots))
	}
	if res.PatientsScheduled > 0 {
		res.StabilityRate = round2(float64(res.SingleMachine) / float64(res.PatientsScheduled))
	}
	res.Score = int(math.Round(res.OnTargetRate*60 + res.StabilityRate*40))
	return res, nil
}
