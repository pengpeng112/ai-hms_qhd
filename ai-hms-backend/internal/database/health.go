package database

import (
	"log"

	"gorm.io/gorm"
)

// RequiredNewTable 应用依赖的独立新表。
type RequiredNewTable struct {
	Table   string // 表名
	Feature string // 依赖该表的功能
	DDL     string // 建表脚本路径
}

// RequiredNewTables 当前所有应用依赖的新表清单。新增依赖新表的功能时在此登记。
// 这些表是独立新表，不改老系统既有表；部署阶段可执行审核后的 CREATE TABLE IF NOT EXISTS 脚本。
// 接班记录复用老库 Schedule_CheckIn（已存在，不在此列）。
var RequiredNewTables = []RequiredNewTable{
	{Table: "sign_record", Feature: "待签 / 电子签", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "Schedule_StaffDuty", Feature: "人力排班·月基线", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "Schedule_StaffDutyOverride", Feature: "日覆盖 / 接班", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "Schedule_Patient", Feature: "智能排班·轻量患者档案", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "exam_reports", Feature: "HIS 检查报告主表", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "exam_report_items", Feature: "HIS 检查报告项目明细", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "external_patient_mappings", Feature: "外部患者映射", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "sync_job_configs", Feature: "同步任务配置", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "sync_job_runs", Feature: "同步任务运行历史", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "patient_infectious", Feature: "传染病筛查与阳性处置", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "patient_actr", Feature: "ACTRS CTR/ACTR 镜像", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "cnrds_report", Feature: "CNRDS 上报", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "water_quality", Feature: "透析用水/透析液质量监测", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "disinfection_compliance", Feature: "透析机消毒监管", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "vascular_access_event", Feature: "血管通路全生命周期", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "adverse_event", Feature: "不良事件登记", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "medication_admin", Feature: "长嘱给药执行", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "dry_weight_assessment", Feature: "干体重评估", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "patient_dry_weight", Feature: "确诊干体重锚定", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "nursing_doc", Feature: "护理文书", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "consent_record", Feature: "知情同意", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "charge_record", Feature: "收费归集清单", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "charge_line", Feature: "收费归集明细", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "his_price_item", Feature: "HIS价表镜像", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "monitoring_threshold", Feature: "实时监控固定阈值表", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "monitoring_vp_stratum", Feature: "实时监控VP分层阈值表", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "monitoring_setting", Feature: "实时监控标量配置", DDL: "docs/sql/deploy_new_tables.sql"},
}

// VerifyRequiredTables 只读检查新表是否已存在，返回缺失的表名列表（绝不执行任何 DDL）。
func VerifyRequiredTables(db *gorm.DB) []string {
	if db == nil {
		return nil
	}
	missing := make([]string, 0, len(RequiredNewTables))
	for _, t := range RequiredNewTables {
		if !db.Migrator().HasTable(t.Table) {
			missing = append(missing, t.Table)
		}
	}
	return missing
}

// LogRequiredTablesStatus 启动期把缺失的新表打成醒目告警（把"运行时静默 500"提前为"启动即可见"）。
func LogRequiredTablesStatus(db *gorm.DB) {
	missing := VerifyRequiredTables(db)
	if len(missing) == 0 {
		log.Println("[SCHEMA] 新表检查通过：应用依赖的新表均已存在")
		return
	}
	missingSet := make(map[string]struct{}, len(missing))
	for _, m := range missing {
		missingSet[m] = struct{}{}
	}
	for _, t := range RequiredNewTables {
		if _, ok := missingSet[t.Table]; ok {
			log.Printf("[SCHEMA] ⚠️ 缺少新表 %q（功能：%s）—— 请在部署阶段执行 %s 建表；缺失期间相关端点可能返回 500（AutoMigrate 永久禁用，应用运行时不会自动建表）",
				t.Table, t.Feature, t.DDL)
		}
	}
}
