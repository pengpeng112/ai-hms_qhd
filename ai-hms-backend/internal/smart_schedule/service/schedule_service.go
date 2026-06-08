// Package service 编排排班用例:加载快照 → 跑引擎 → 事务回写。
package service

import (
	"time"

	"gorm.io/gorm"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/config"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/repo"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

// filterDischarged 过滤掉已出组病人的模板项(决策 27)。
func filterDischarged(g *gorm.DB, tenant int64, items []*model.ScheduleTemplateItem) []*model.ScheduleTemplateItem {
	var ids []int64
	g.Model(&model.PatientProfile{}).Where(`"TenantId" = ? AND "PatientStatus" = ?`, tenant, sched.PatientDischarged).
		Pluck("PatientId", &ids)
	if len(ids) == 0 {
		return items
	}
	discharged := map[int64]bool{}
	for _, id := range ids {
		discharged[id] = true
	}
	out := items[:0]
	for _, it := range items {
		if !discharged[it.PatientId] {
			out = append(out, it)
		}
	}
	return out
}

// GenerateResult 生成结果摘要。
type GenerateResult struct {
	StartDate      string `json:"startDate"`
	Weeks          int    `json:"weeks"`
	DialysisDays   int    `json:"dialysisDays"`
	Drafts         int    `json:"drafts"`
	Conflicts      int    `json:"conflicts"`
	ParityAssigned int    `json:"parityAssigned"`
}

// GenerateSchedule 生成未来 weeks 周(2 或 4)的草稿排班(规范 §6,决策 11)。
// 流程:读锚点 → 取生效模板项 → 加载 Board → 引擎两轮分配 → 事务回写草稿/冲突/奇偶周。
func GenerateSchedule(g *gorm.DB, tenant int64, start time.Time, weeks int) (*GenerateResult, error) {
	anchor := config.AnchorMonday(g, tenant)

	items, err := repo.GetActiveTemplateItems(g, tenant)
	if err != nil {
		return nil, err
	}
	items = filterDischarged(g, tenant, items) // 出组病人不再排(决策 27)

	end := start.AddDate(0, 0, weeks*7)
	board, err := repo.LoadBoard(g, tenant, anchor, start, end)
	if err != nil {
		return nil, err
	}

	eng := sched.NewEngine(board)
	eng.SpillHorizonDays = config.SpillHorizonDays(g, tenant) // 顺延窗口配置化(决策 22)
	assignments := eng.AssignHdfWeekParity(items)              // 先定奇偶周并捕获,供持久化
	dates := eng.ExpandDialysisDates(start, weeks)
	eng.Generate(items, dates) // 内部再调一次 AssignHdfWeekParity,已分配者幂等跳过

	err = g.Transaction(func(tx *gorm.DB) error {
		if err := repo.PersistParity(tx, tenant, assignments); err != nil {
			return err
		}
		if err := repo.SaveDrafts(tx, tenant, board.Drafts); err != nil {
			return err
		}
		return repo.SaveConflicts(tx, tenant, board.Conflicts)
	})
	if err != nil {
		return nil, err
	}

	return &GenerateResult{
		StartDate:      start.Format("2006-01-02"),
		Weeks:          weeks,
		DialysisDays:   len(dates),
		Drafts:         len(board.Drafts),
		Conflicts:      len(board.Conflicts),
		ParityAssigned: len(assignments),
	}, nil
}
