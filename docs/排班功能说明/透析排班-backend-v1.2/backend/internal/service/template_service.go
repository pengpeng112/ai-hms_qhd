package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
)

// 模板管理(P1 #5)。模板 = 稳定病人骨架的快照,供"模板复制生成"使用(决策 11/16)。
// 这里提供"由当前病人骨架重建生效模板"(最常用),以及模板/模板项查询。

func ListTemplates(g *gorm.DB, tenant int64) ([]model.ScheduleTemplate, error) {
	var ts []model.ScheduleTemplate
	err := g.Where(`"TenantId" = ?`, tenant).Order(`"Id"`).Find(&ts).Error
	return ts, err
}

func ListTemplateItems(g *gorm.DB, tenant, templateID int64) ([]model.ScheduleTemplateItem, error) {
	var its []model.ScheduleTemplateItem
	err := g.Where(`"TenantId" = ? AND "TemplateId" = ?`, tenant, templateID).Order(`"PatientId"`).Find(&its).Error
	return its, err
}

// RebuildTemplateResult 重建结果。
type RebuildTemplateResult struct {
	TemplateId int64  `json:"templateId"`
	Items      int    `json:"items"`
	Name       string `json:"name"`
}

// RebuildTemplateFromProfiles 把当前所有病人排班骨架(Profile)快照成一份新的生效模板。
// 旧的生效模板置为失效;新模板含每位"非拒收、非临时、骨架完整"病人一项。
func RebuildTemplateFromProfiles(g *gorm.DB, tenant int64, name string) (*RebuildTemplateResult, error) {
	if name == "" {
		name = "标准周模板 " + time.Now().Format("2006-01-02 15:04")
	}
	var profiles []model.PatientProfile
	if err := g.Where(`"TenantId" = ? AND "IsAdmissionRejected" = false`, tenant).Find(&profiles).Error; err != nil {
		return nil, err
	}

	res := &RebuildTemplateResult{Name: name}
	err := g.Transaction(func(tx *gorm.DB) error {
		// 旧生效模板失效
		if err := tx.Model(&model.ScheduleTemplate{}).Where(`"TenantId" = ? AND "IsActive" = true`, tenant).
			Update("IsActive", false).Error; err != nil {
			return err
		}
		tpl := &model.ScheduleTemplate{BaseModel: model.BaseModel{TenantId: tenant}, Name: name, Scope: "ALL", IsActive: true}
		if err := tx.Create(tpl).Error; err != nil {
			return err
		}
		res.TemplateId = tpl.Id

		var items []*model.ScheduleTemplateItem
		for _, p := range profiles {
			if p.FreqPattern == sched.FreqTemporary || p.HomeWardId == nil || p.ShiftId == nil {
				continue // 临时/骨架不全者不进模板
			}
			items = append(items, &model.ScheduleTemplateItem{
				BaseModel: model.BaseModel{TenantId: tenant}, TemplateId: tpl.Id, PatientId: p.PatientId,
				ZoneTag: p.ZoneTag, WardId: p.HomeWardId, ShiftId: p.ShiftId, FreqPattern: p.FreqPattern,
				FixedHdMachineId: p.FixedHdMachineId, FixedHdfMachineId: p.FixedHdfMachineId,
				HdfEnabled: p.HdfEnabled, HdfWeekday: p.HdfWeekday, HdfWeekParity: p.HdfWeekParity,
			})
		}
		res.Items = len(items)
		if len(items) > 0 {
			return tx.CreateInBatches(items, 200).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ResolveConflict 处理一条冲突/建议(P1 #6):accept=已处理(采纳建议)、ignore=已忽略。
// 对 SLOT_SPILLED 等"已生成建议草稿"的冲突,采纳即关闭(草稿已在,等正常确认流程);
// 系统不在此自动改排——仍是人工裁决,只是把队列项收尾。
func ResolveConflict(g *gorm.DB, tenant, by, id int64, accept bool) error {
	var c model.ConflictQueue
	if err := g.Where(`"TenantId" = ? AND "Id" = ?`, tenant, id).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	status := int16(20) // 已忽略
	if accept {
		status = 10 // 已处理
	}
	now := time.Now()
	return g.Model(&model.ConflictQueue{}).Where(`"TenantId" = ? AND "Id" = ?`, tenant, id).
		Updates(map[string]interface{}{"Status": status, "ResolvedBy": by, "ResolvedAt": now}).Error
}
