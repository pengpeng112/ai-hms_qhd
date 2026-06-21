package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

type newTableDDL struct {
	Table      string
	Statements []string
}

var requiredNewTableDDL = map[string]newTableDDL{
	"exam_reports": {Table: "exam_reports", Statements: []string{
		`CREATE TABLE IF NOT EXISTS exam_reports (id varchar(36) NOT NULL, patient_id varchar(36) NOT NULL, exam_date timestamp, title varchar(200) NOT NULL, conclusion text, department varchar(100), external_report_id varchar(128), source_system varchar(32) NOT NULL DEFAULT 'HIS_ORACLE_EXAM', synced_at timestamp, created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, CONSTRAINT pk_exam_reports PRIMARY KEY (id))`,
		`CREATE INDEX IF NOT EXISTS idx_exam_reports_patient_date ON exam_reports (patient_id, exam_date DESC)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_reports_external_unique ON exam_reports (source_system, external_report_id, patient_id)`,
	}},
	"exam_report_items": {Table: "exam_report_items", Statements: []string{
		`CREATE TABLE IF NOT EXISTS exam_report_items (id varchar(36) NOT NULL, exam_report_id varchar(36) NOT NULL, item_name varchar(200) NOT NULL, item_code varchar(64), item_category varchar(100), item_result text, sort_order int NOT NULL DEFAULT 0, created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, CONSTRAINT pk_exam_report_items PRIMARY KEY (id))`,
		`CREATE INDEX IF NOT EXISTS idx_exam_report_items_report ON exam_report_items (exam_report_id, sort_order)`,
	}},
	"external_patient_mappings": {Table: "external_patient_mappings", Statements: []string{
		`CREATE TABLE IF NOT EXISTS external_patient_mappings (id varchar(36) NOT NULL, tenant_id bigint NOT NULL, legacy_patient_id bigint NOT NULL, external_system varchar(32) NOT NULL, external_patient_id varchar(64) NOT NULL, external_visit_id varchar(64), id_no varchar(64), dialysis_no varchar(64), hosp_no varchar(64), case_no varchar(64), outpatient_no varchar(64), medical_record_no varchar(64), patient_name varchar(128), match_status varchar(32) NOT NULL DEFAULT 'confirmed', last_synced_at timestamp, created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, CONSTRAINT pk_external_patient_mappings PRIMARY KEY (id))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_external_patient_mapping_unique ON external_patient_mappings (tenant_id, external_system, external_patient_id, COALESCE(external_visit_id, ''))`,
		`CREATE INDEX IF NOT EXISTS idx_external_patient_mapping_legacy ON external_patient_mappings (tenant_id, legacy_patient_id)`,
	}},
	"sync_job_configs": {Table: "sync_job_configs", Statements: []string{
		`CREATE TABLE IF NOT EXISTS sync_job_configs (id varchar(36) NOT NULL, job_code varchar(64) NOT NULL, source_system varchar(32) NOT NULL, sync_type varchar(64) NOT NULL, enabled boolean NOT NULL DEFAULT false, cron_expr varchar(64), interval_seconds int, batch_size int NOT NULL DEFAULT 500, timeout_seconds int NOT NULL DEFAULT 60, max_retry int NOT NULL DEFAULT 3, cursor_type varchar(32) NOT NULL DEFAULT 'time', cursor_value varchar(128), overwrite_policy varchar(32) NOT NULL DEFAULT 'fill_empty', last_run_at timestamp, next_run_at timestamp, created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, CONSTRAINT pk_sync_job_configs PRIMARY KEY (id))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sync_job_configs_code ON sync_job_configs (job_code)`,
	}},
	"sync_job_runs": {Table: "sync_job_runs", Statements: []string{
		`CREATE TABLE IF NOT EXISTS sync_job_runs (id varchar(36) NOT NULL, job_code varchar(64) NOT NULL, source_system varchar(32) NOT NULL, sync_type varchar(64) NOT NULL, status varchar(32) NOT NULL, started_at timestamp NOT NULL, finished_at timestamp, duration_ms bigint, fetched_count int NOT NULL DEFAULT 0, created_count int NOT NULL DEFAULT 0, updated_count int NOT NULL DEFAULT 0, skipped_count int NOT NULL DEFAULT 0, failed_count int NOT NULL DEFAULT 0, cursor_before varchar(128), cursor_after varchar(128), error_message text, created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, CONSTRAINT pk_sync_job_runs PRIMARY KEY (id))`,
		`CREATE INDEX IF NOT EXISTS idx_sync_job_runs_job_start ON sync_job_runs (job_code, started_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_job_runs_status ON sync_job_runs (status, started_at DESC)`,
	}},
	"sign_record": {Table: "sign_record", Statements: []string{
		`CREATE TABLE IF NOT EXISTS sign_record (id varchar(36) NOT NULL, tenant_id bigint NOT NULL, target_type varchar(16) NOT NULL, target_id varchar(64) NOT NULL, signer_id varchar(64) NOT NULL, signer_name varchar(64), sign_time timestamp NOT NULL, signature_blob text, created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, CONSTRAINT pk_sign_record PRIMARY KEY (id))`,
		`CREATE INDEX IF NOT EXISTS idx_sign_record_target ON sign_record (tenant_id, target_type, target_id)`,
	}},
	"Schedule_StaffDuty": {Table: "Schedule_StaffDuty", Statements: []string{
		`CREATE TABLE IF NOT EXISTS "Schedule_StaffDuty" ("Id" bigserial PRIMARY KEY, "TenantId" bigint NOT NULL, "CreatorId" bigint, "CreateTime" timestamp NOT NULL DEFAULT now(), "LastModifyTime" timestamp NOT NULL DEFAULT now(), "StaffId" bigint NOT NULL, "StaffName" varchar(64), "DutyRole" varchar(32) NOT NULL, "WardId" bigint NOT NULL, "DutyDate" date NOT NULL, "Shift" varchar(16))`,
		`CREATE INDEX IF NOT EXISTS "idx_staffduty_lookup" ON "Schedule_StaffDuty" ("TenantId", "WardId", "DutyDate", "DutyRole")`,
	}},
	"Schedule_StaffDutyOverride": {Table: "Schedule_StaffDutyOverride", Statements: []string{
		`CREATE TABLE IF NOT EXISTS "Schedule_StaffDutyOverride" ("Id" bigserial PRIMARY KEY, "TenantId" bigint NOT NULL, "CreatorId" bigint, "CreateTime" timestamp NOT NULL DEFAULT now(), "LastModifyTime" timestamp NOT NULL DEFAULT now(), "DutyDate" date NOT NULL, "WardId" bigint NOT NULL, "DutyRole" varchar(32) NOT NULL, "OriginalStaffId" bigint, "ActualStaffId" bigint NOT NULL, "ActualStaffName" varchar(64), "Reason" varchar(128), "ChangedBy" bigint)`,
		`CREATE INDEX IF NOT EXISTS "idx_staffdutyoverride_lookup" ON "Schedule_StaffDutyOverride" ("TenantId", "WardId", "DutyDate", "DutyRole")`,
	}},
	"Schedule_Patient": {Table: "Schedule_Patient", Statements: []string{
		`CREATE TABLE IF NOT EXISTS "Schedule_Patient" ("Id" bigint NOT NULL, "TenantId" bigint NOT NULL, "Name" varchar(64) NOT NULL, "Gender" varchar(8), "InfectionStatus" varchar(16) NOT NULL DEFAULT 'unknown', "InfectionWaivedBy" bigint, "InfectionWaivedAt" timestamptz, "CreateTime" timestamptz NOT NULL DEFAULT now(), "LastModifyTime" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("Id"))`,
		`CREATE INDEX IF NOT EXISTS "idx_schedule_patient_tenant" ON "Schedule_Patient" ("TenantId")`,
	}},
	"patient_infectious": {Table: "patient_infectious", Statements: []string{
		`CREATE TABLE IF NOT EXISTS patient_infectious (id varchar(36) PRIMARY KEY, tenant_id bigint NOT NULL, patient_id varchar(64) NOT NULL, screen_date date, items text, source varchar(8), result_overall varchar(8), positive_markers varchar(128), next_due_date date, disposition varchar(16), handled_doctor_id varchar(64), handled_headnurse_id varchar(64), handled_at timestamptz, zone_tag varchar(16), note varchar(256), created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now())`,
		`CREATE INDEX IF NOT EXISTS idx_inf_tenant_patient ON patient_infectious (tenant_id, patient_id)`,
	}},
	"patient_actr": {Table: "patient_actr", Statements: []string{
		`CREATE TABLE IF NOT EXISTS patient_actr (id varchar(36) PRIMARY KEY, tenant_id bigint NOT NULL, patient_id varchar(64) NOT NULL, dialysis_no varchar(64), actrs_xray_id bigint NOT NULL, analysis_date timestamptz, ctr numeric, actr numeric, actr1 numeric, actr2 numeric, actr_norm numeric, heart_width integer, lung_width integer, tilt_angle numeric, qc_pass integer NOT NULL DEFAULT 0, qc_pa_ap varchar(8), qc_warnings varchar(256), model_version varchar(32), source varchar(16), image_path varchar(256), overlay_path varchar(256), mask_path varchar(256), doctor_correction numeric, corrected_by varchar(64), corrected_at timestamptz, adopted_by varchar(64), adopted_at timestamptz, adopted_prescription_id varchar(32), adopted_dry_weight numeric, adopted_uf_quantity numeric, notes varchar(256), synced_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now())`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_actr_tenant_patient_xray ON patient_actr (tenant_id, patient_id, actrs_xray_id)`,
		`CREATE INDEX IF NOT EXISTS idx_actr_tenant_patient ON patient_actr (tenant_id, patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_actr_adopted_prescription ON patient_actr (tenant_id, adopted_prescription_id)`,
	}},
	"cnrds_report": {Table: "cnrds_report", Statements: []string{
		`CREATE TABLE IF NOT EXISTS cnrds_report (id varchar(36) PRIMARY KEY, tenant_id bigint NOT NULL, period varchar(16), report_type varchar(12), event_type varchar(16), patient_id varchar(64), content text, patient_count int NOT NULL DEFAULT 0, status varchar(12) NOT NULL DEFAULT 'draft', export_ref varchar(256), reviewed_by varchar(64), submitted_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now())`,
		`CREATE INDEX IF NOT EXISTS idx_cnrds_tenant_type_period ON cnrds_report (tenant_id, report_type, period)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_cnrds_monthly_unique ON cnrds_report (tenant_id, report_type, period) WHERE report_type = 'monthly'`,
	}},
	"water_quality": {Table: "water_quality", Statements: []string{
		`CREATE TABLE IF NOT EXISTS water_quality (id varchar(36) PRIMARY KEY, tenant_id bigint NOT NULL, test_date date, test_type varchar(24), sample_point varchar(16), device_id varchar(64), value numeric, unit varchar(16), standard_limit varchar(32), result varchar(8), source varchar(12), next_due_date date, handled_engineer_id varchar(64), handled_headnurse_id varchar(64), handled_at timestamptz, action varchar(256), created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now())`,
		`CREATE INDEX IF NOT EXISTS idx_wq_tenant_type_date ON water_quality (tenant_id, test_type, test_date)`,
	}},
	"disinfection_compliance": {Table: "disinfection_compliance", Statements: []string{
		`CREATE TABLE IF NOT EXISTS disinfection_compliance (id varchar(36) PRIMARY KEY, tenant_id bigint NOT NULL, disinfection_id bigint NOT NULL, device_id bigint, concentration varchar(32), residual_check varchar(8), result varchar(8), source varchar(12), doc_ref varchar(256), created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now())`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_disinfection ON disinfection_compliance (disinfection_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dc_tenant_device ON disinfection_compliance (tenant_id, device_id)`,
	}},
	"vascular_access_event": {Table: "vascular_access_event", Statements: []string{
		`CREATE TABLE IF NOT EXISTS vascular_access_event (id varchar(36) PRIMARY KEY, tenant_id bigint NOT NULL, access_id bigint NOT NULL, patient_id bigint, event_type varchar(16), event_date date, detail text, operator_id varchar(64), note varchar(256), created_at timestamptz NOT NULL DEFAULT now())`,
		`CREATE INDEX IF NOT EXISTS idx_vae_tenant_patient ON vascular_access_event (tenant_id, patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_vae_access ON vascular_access_event (access_id)`,
	}},
}

// EnsureRequiredNewTables creates only whitelisted independent new tables.
// Existing tables are never altered; legacy tables are not part of this list.
func EnsureRequiredNewTables(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	for _, t := range RequiredNewTables {
		if db.Migrator().HasTable(t.Table) {
			log.Printf("[SCHEMA] 新表 %q 已存在，跳过自动创建", t.Table)
			continue
		}
		ddl, ok := requiredNewTableDDL[t.Table]
		if !ok {
			return fmt.Errorf("缺少新表 %q 的白名单 DDL，拒绝自动创建", t.Table)
		}
		log.Printf("[SCHEMA] 新表 %q 不存在，开始按白名单 DDL 自动创建（功能：%s）", t.Table, t.Feature)
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, stmt := range ddl.Statements {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("自动创建新表 %q 失败: %w", t.Table, err)
		}
	}
	return nil
}
