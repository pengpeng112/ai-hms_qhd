package service

import (
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/config"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/repo"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

// 差异检测(决策 13 / 规范 §8):比较每位病人在周期内的"应排次数"与"已排次数"。
// 应排 = 按频率模式落在透析日的次数;已排 = 已分配机位且未取消的记录数(含缺席,排除取消/待排)。

// DiffItem 单个病人的应排/已排差异。
type DiffItem struct {
	PatientId   int64  `json:"patientId"`
	PatientName string `json:"patientName"`
	FreqPattern int16  `json:"freqPattern"`
	Expected    int    `json:"expected"`
	Scheduled   int    `json:"scheduled"`
	Diff        int    `json:"diff"` // 应排 - 已排;>0=少排(需补),<0=多排
}

// DiffResult 差异检测结果。
type DiffResult struct {
	WeekStart string     `json:"weekStart"`
	Weeks     int        `json:"weeks"`
	Items     []DiffItem `json:"items"` // 仅含 Diff != 0 者
}

// ComputeDiffs 计算 [weekStart, +weeks] 内各病人的应排/已排差异(仅返回有差异者)。
func ComputeDiffs(g *gorm.DB, tenant int64, weekStart time.Time, weeks int) (*DiffResult, error) {
	anchor := config.AnchorMonday(g, tenant)
	start := dayStart(weekStart)
	end := start.AddDate(0, 0, weeks*7)

	// 透析日集合(跳过非透析日)。
	board, err := repo.LoadBoard(g, tenant, anchor, start, end)
	if err != nil {
		return nil, err
	}
	var dates []time.Time
	for i := 0; i < weeks*7; i++ {
		d := start.AddDate(0, 0, i)
		if board.IsDialysisDay(d) {
			dates = append(dates, d)
		}
	}

	// 病人骨架(排除拒收、临时频率)。
	var profiles []*model.PatientProfile
	if err := g.Where(`"TenantId" = ? AND "IsAdmissionRejected" = false`, tenant).Find(&profiles).Error; err != nil {
		return nil, err
	}

	// 已排次数:已分配机位、未取消(含缺席留痕)的记录,按病人聚合。
	type cnt struct {
		PatientId int64
		N         int
	}
	var rows []cnt
	if err := g.Model(&model.PatientShift{}).
		Select(`"PatientId" AS patient_id, count(*) AS n`).
		Where(`"TenantId" = ? AND "TreatmentTime" >= ? AND "TreatmentTime" < ? AND "Status" IN ?`,
			tenant, start, end, []int16{sched.StatusDraft, sched.StatusConfirmed, sched.StatusInDialysis, sched.StatusCompleted, sched.StatusAbsent}).
		Group(`"PatientId"`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	scheduled := map[int64]int{}
	for _, r := range rows {
		scheduled[r.PatientId] = r.N
	}

	names := patientNames(g, tenant)
	res := &DiffResult{WeekStart: start.Format("2006-01-02"), Weeks: weeks}
	for _, p := range profiles {
		if p.FreqPattern == sched.FreqTemporary {
			continue // 临时病人无固定应排次数
		}
		expected := 0
		for _, d := range dates {
			if sched.IsDue(p.FreqPattern, d) {
				expected++
			}
		}
		sch := scheduled[p.PatientId]
		diff := expected - sch
		if diff != 0 {
			res.Items = append(res.Items, DiffItem{
				PatientId: p.PatientId, PatientName: names[p.PatientId], FreqPattern: p.FreqPattern,
				Expected: expected, Scheduled: sch, Diff: diff,
			})
		}
	}
	return res, nil
}
