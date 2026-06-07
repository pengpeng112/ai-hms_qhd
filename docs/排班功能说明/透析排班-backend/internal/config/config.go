// Package config 读取租户级排班配置(设计 §0.1),带默认值兜底。
package config

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
)

// getSetting 读一条租户配置;不存在返回空串。
func getSetting(g *gorm.DB, tenant int64, key string) string {
	var s model.TenantSetting
	err := g.Where(`"TenantId" = ? AND "SettingKey" = ?`, tenant, key).First(&s).Error
	if err != nil {
		return ""
	}
	return s.SettingValue
}

// AnchorMonday 奇偶周基准周一(决策 21 / D-6),默认 2025-01-06。
// 不变量:必须是周一;非法或缺失时回退默认值。
func AnchorMonday(g *gorm.DB, tenant int64) time.Time {
	v := getSetting(g, tenant, sched.CfgAnchorMonday)
	if v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil && t.Weekday() == time.Monday {
			return t
		}
	}
	return sched.DefaultAnchorMonday
}

// DraftWeeks 一次生成草稿的周数(2 或 4),默认 2。
func DraftWeeks(g *gorm.DB, tenant int64) int {
	if v := getSetting(g, tenant, sched.CfgDraftWeeks); v != "" {
		if n, err := strconv.Atoi(v); err == nil && (n == 2 || n == 4) {
			return n
		}
	}
	return 2
}

// LowSlotWarnThreshold 余位预警阈值(开放项 D-1,暂默认 2)。
func LowSlotWarnThreshold(g *gorm.DB, tenant int64) int {
	if v := getSetting(g, tenant, sched.CfgLowSlotWarn); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 2
}
