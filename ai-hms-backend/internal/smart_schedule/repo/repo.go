// Package repo 持久化层:在 GORM 与排班引擎之间搬运数据。
// 把"从库加载快照(Board)"与"草稿/冲突回写"集中于此,使引擎保持纯算法。
package repo

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model"
	"github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched"
)

// GetActiveTemplate 取当前生效模板(IsActive=true 的第一条)。
func GetActiveTemplate(g *gorm.DB, tenant int64) (*model.ScheduleTemplate, error) {
	var t model.ScheduleTemplate
	err := g.Where(`"TenantId" = ? AND "IsActive" = true`, tenant).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetActiveTemplateItems 取生效模板下的全部模板项(稳定病人骨架)。
func GetActiveTemplateItems(g *gorm.DB, tenant int64) ([]*model.ScheduleTemplateItem, error) {
	t, err := GetActiveTemplate(g, tenant)
	if err != nil {
		return nil, err
	}
	var items []*model.ScheduleTemplateItem
	err = g.Where(`"TenantId" = ? AND "TemplateId" = ?`, tenant, t.Id).Find(&items).Error
	return items, err
}

// LoadBoard 把 [start, end] 区间内排班所需资源与现有占用装载成内存快照 Board。
func LoadBoard(g *gorm.DB, tenant int64, anchor, start, end time.Time) (*sched.Board, error) {
	var wards []*model.Ward
	if err := g.Where(`"TenantId" = ?`, tenant).Find(&wards).Error; err != nil {
		return nil, err
	}
	var machines []*model.Machine
	if err := g.Where(`"TenantId" = ?`, tenant).Find(&machines).Error; err != nil {
		return nil, err
	}
	var shifts []*model.Shift
	if err := g.Where(`"TenantId" = ?`, tenant).Find(&shifts).Error; err != nil {
		return nil, err
	}
	// 加载范围内所有(规律)排班记录(含取消/缺席),供 Board 同时构建"机位占用"与"病人占位集合";
	// NewBoard 内部只用有效记录算机位占用,取消/缺席仅进病人占位集合(生成时跳过,不复活)。
	var existing []*model.PatientShift
	if err := g.Where(
		`"TenantId" = ? AND "TreatmentTime" BETWEEN ? AND ? AND "ShiftId" > 0`,
		tenant, start, end,
	).Find(&existing).Error; err != nil {
		return nil, err
	}
	var outages []*model.MachineOutage
	if err := g.Where(`"TenantId" = ? AND "StartAt" < ? AND ("EndAt" IS NULL OR "EndAt" > ?)`, tenant, end, start).
		Find(&outages).Error; err != nil {
		return nil, err
	}
	var calendar []*model.Calendar
	if err := g.Where(`"TenantId" = ? AND "CalDate" BETWEEN ? AND ?`, tenant, start, end).
		Find(&calendar).Error; err != nil {
		return nil, err
	}
	return sched.NewBoard(anchor, wards, machines, shifts, existing, outages, calendar), nil
}

// SaveDrafts 批量写入草稿排班(补 TenantId)。
func SaveDrafts(tx *gorm.DB, tenant int64, drafts []*model.PatientShift) error {
	if len(drafts) == 0 {
		return nil
	}
	for _, d := range drafts {
		d.TenantId = tenant
	}
	// 并发安全:若与已有(唯一索引)冲突则跳过该行,不让整批失败——
	// 配合"生成幂等",并发重复生成不会产生重复排班,也不会报错。
	return tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(drafts, 200).Error
}

// SaveConflicts 批量写入冲突/待处理队列(补 TenantId)。
func SaveConflicts(tx *gorm.DB, tenant int64, conflicts []*model.ConflictQueue) error {
	if len(conflicts) == 0 {
		return nil
	}
	for _, c := range conflicts {
		c.TenantId = tenant
	}
	return tx.CreateInBatches(conflicts, 200).Error
}

// PersistParity 写回 HDF 奇偶周分配到病人 Profile 与对应模板项。
func PersistParity(tx *gorm.DB, tenant int64, assignments []sched.ParityAssignment) error {
	for _, a := range assignments {
		parity := a.Parity
		if err := tx.Model(&model.PatientProfile{}).
			Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, a.PatientId).
			Update("HdfWeekParity", parity).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ScheduleTemplateItem{}).
			Where(`"TenantId" = ? AND "PatientId" = ?`, tenant, a.PatientId).
			Update("HdfWeekParity", parity).Error; err != nil {
			return err
		}
	}
	return nil
}

// IndexSpec 描述排班并发安全所依赖的一个唯一索引：名称 + 创建 DDL。
type IndexSpec struct {
	Name string
	DDL  string
}

// RequiredUniqueIndexes 排班模块依赖的唯一索引清单（单一真值源）。
//
// 这些索引保证"同一病人/机位 + 时间 + 班次"不会被并发生成重复排班行。
// 老库守则禁止运行时 DDL，因此它们必须由 DBA 通过
// scripts/schedule_unique_indexes.sql 预先建好；应用启动只用 VerifyIndexes
// 做存在性校验，绝不自动创建。本清单需与该 SQL 脚本保持一致。
//
// 注意：uq_ps_machine_slot 的谓词沿用历史定义 "MachineId IS NOT NULL"。
// 由于现行模型 MachineId 为 int64 NOT NULL DEFAULT 0，该谓词恒真，会把
// 未分配机位（MachineId=0）的多条排班视为冲突。是否应改为 "MachineId > 0"
// 属独立的索引语义优化，不在本次运行时 DDL 整改范围内，详见 SQL 脚本中的说明。
var RequiredUniqueIndexes = []IndexSpec{
	{
		Name: "uq_ps_patient_slot",
		DDL: `CREATE UNIQUE INDEX IF NOT EXISTS uq_ps_patient_slot
		 ON "Schedule_PatientShift" ("TenantId","PatientId","TreatmentTime","ShiftId")
		 WHERE "Status" NOT IN (70,80) AND "ShiftId" > 0`,
	},
	{
		Name: "uq_ps_machine_slot",
		DDL: `CREATE UNIQUE INDEX IF NOT EXISTS uq_ps_machine_slot
		 ON "Schedule_PatientShift" ("TenantId","MachineId","TreatmentTime","ShiftId")
		 WHERE "Status" NOT IN (70,80) AND "MachineId" IS NOT NULL AND "ShiftId" > 0`,
	},
}

// VerifyIndexes 校验排班唯一索引是否已存在（只读，不执行任何 DDL）。
//
// 返回缺失的索引名列表，供调用方在启动日志与 /schedule/health 中告警。
// 这是老库"运行时严禁 DDL"红线下对原 EnsureIndexes 的安全替代：
// 索引创建交由 DBA 执行 scripts/schedule_unique_indexes.sql，应用只读校验。
func VerifyIndexes(g *gorm.DB) ([]string, error) {
	names := make([]string, 0, len(RequiredUniqueIndexes))
	for _, idx := range RequiredUniqueIndexes {
		names = append(names, idx.Name)
	}

	var present []string
	if err := g.Table("pg_indexes").
		Where("schemaname = ? AND indexname IN ?", "public", names).
		Pluck("indexname", &present).Error; err != nil {
		return nil, err
	}

	have := make(map[string]bool, len(present))
	for _, n := range present {
		have[n] = true
	}
	var missing []string
	for _, n := range names {
		if !have[n] {
			missing = append(missing, n)
		}
	}
	return missing, nil
}

// EnsureIndexes 创建排班唯一索引（执行 DDL）。
//
// ⚠️ 仅供单元/集成测试夹具与 DBA 离线场景使用，严禁在运行时对老生产库调用——
// 老库守则禁止任何运行时 DDL。生产环境请改用 scripts/schedule_unique_indexes.sql
// 由 DBA 审核执行；应用启动只通过 VerifyIndexes 做存在性校验。
func EnsureIndexes(g *gorm.DB) error {
	for _, idx := range RequiredUniqueIndexes {
		if err := g.Exec(idx.DDL).Error; err != nil {
			return err
		}
	}
	return nil
}


