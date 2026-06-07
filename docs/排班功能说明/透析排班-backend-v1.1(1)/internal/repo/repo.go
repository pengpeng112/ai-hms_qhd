// Package repo 持久化层:在 GORM 与排班引擎之间搬运数据。
// 把"从库加载快照(Board)"与"草稿/冲突回写"集中于此,使引擎保持纯算法。
package repo

import (
	"time"

	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
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
		`"TenantId" = ? AND "ScheduleDate" BETWEEN ? AND ? AND "ShiftId" IS NOT NULL`,
		tenant, start, end,
	).Find(&existing).Error; err != nil {
		return nil, err
	}
	var outages []*model.MachineOutage
	if err := g.Where(`"TenantId" = ?`, tenant).Find(&outages).Error; err != nil {
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
	return tx.CreateInBatches(drafts, 200).Error
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
