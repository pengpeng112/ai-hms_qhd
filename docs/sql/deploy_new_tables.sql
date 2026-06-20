-- ================================================================
-- 部署阶段自动建表脚本：独立新表
-- ================================================================
-- 适用范围：只创建不影响老系统既有表的独立新表。
-- 执行方式：部署阶段幂等执行；应用运行时仍禁止 AutoMigrate/DDL。
-- 注意：本脚本不包含任何老表 ALTER TABLE。
-- ================================================================

BEGIN;

-- 1. HIS 检查报告主表
CREATE TABLE IF NOT EXISTS exam_reports (
    id                  VARCHAR(36)  NOT NULL,
    patient_id          VARCHAR(36)  NOT NULL,
    exam_date           TIMESTAMP,
    title               VARCHAR(200) NOT NULL,
    conclusion          TEXT,
    department          VARCHAR(100),
    external_report_id  VARCHAR(128),
    source_system       VARCHAR(32)  NOT NULL DEFAULT 'HIS_ORACLE_EXAM',
    synced_at           TIMESTAMP,
    created_at          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_exam_reports PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_exam_reports_patient_date
    ON exam_reports (patient_id, exam_date DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_reports_external_unique
    ON exam_reports (source_system, external_report_id, patient_id);

-- 2. HIS 检查报告项目明细表
CREATE TABLE IF NOT EXISTS exam_report_items (
    id              VARCHAR(36)  NOT NULL,
    exam_report_id  VARCHAR(36)  NOT NULL,
    item_name       VARCHAR(200) NOT NULL,
    item_code       VARCHAR(64),
    item_category   VARCHAR(100),
    item_result     TEXT,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_exam_report_items PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_exam_report_items_report
    ON exam_report_items (exam_report_id, sort_order);

-- 3. 外部患者映射表
CREATE TABLE IF NOT EXISTS external_patient_mappings (
    id                  VARCHAR(36)  NOT NULL,
    tenant_id           BIGINT       NOT NULL,
    legacy_patient_id   BIGINT       NOT NULL,
    external_system     VARCHAR(32)  NOT NULL,
    external_patient_id VARCHAR(64)  NOT NULL,
    external_visit_id   VARCHAR(64),
    id_no               VARCHAR(64),
    dialysis_no         VARCHAR(64),
    hosp_no             VARCHAR(64),
    case_no             VARCHAR(64),
    outpatient_no       VARCHAR(64),
    medical_record_no   VARCHAR(64),
    patient_name        VARCHAR(128),
    match_status        VARCHAR(32)  NOT NULL DEFAULT 'confirmed',
    last_synced_at      TIMESTAMP,
    created_at          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_external_patient_mappings PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_external_patient_mapping_unique
    ON external_patient_mappings (tenant_id, external_system, external_patient_id, COALESCE(external_visit_id, ''));

CREATE INDEX IF NOT EXISTS idx_external_patient_mapping_legacy
    ON external_patient_mappings (tenant_id, legacy_patient_id);

-- 4. 同步任务配置表
CREATE TABLE IF NOT EXISTS sync_job_configs (
    id                VARCHAR(36)  NOT NULL,
    job_code          VARCHAR(64)  NOT NULL,
    source_system     VARCHAR(32)  NOT NULL,
    sync_type         VARCHAR(64)  NOT NULL,
    enabled           BOOLEAN      NOT NULL DEFAULT false,
    cron_expr         VARCHAR(64),
    interval_seconds  INT,
    batch_size        INT          NOT NULL DEFAULT 500,
    timeout_seconds   INT          NOT NULL DEFAULT 60,
    max_retry         INT          NOT NULL DEFAULT 3,
    cursor_type       VARCHAR(32)  NOT NULL DEFAULT 'time',
    cursor_value      VARCHAR(128),
    overwrite_policy  VARCHAR(32)  NOT NULL DEFAULT 'fill_empty',
    last_run_at       TIMESTAMP,
    next_run_at       TIMESTAMP,
    created_at        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_sync_job_configs PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sync_job_configs_code
    ON sync_job_configs (job_code);

-- 5. 同步任务运行历史表
CREATE TABLE IF NOT EXISTS sync_job_runs (
    id             VARCHAR(36)  NOT NULL,
    job_code       VARCHAR(64)  NOT NULL,
    source_system  VARCHAR(32)  NOT NULL,
    sync_type      VARCHAR(64)  NOT NULL,
    status         VARCHAR(32)  NOT NULL,
    started_at     TIMESTAMP    NOT NULL,
    finished_at    TIMESTAMP,
    duration_ms    BIGINT,
    fetched_count  INT          NOT NULL DEFAULT 0,
    created_count  INT          NOT NULL DEFAULT 0,
    updated_count  INT          NOT NULL DEFAULT 0,
    skipped_count  INT          NOT NULL DEFAULT 0,
    failed_count   INT          NOT NULL DEFAULT 0,
    cursor_before  VARCHAR(128),
    cursor_after   VARCHAR(128),
    error_message  TEXT,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_sync_job_runs PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_sync_job_runs_job_start
    ON sync_job_runs (job_code, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_sync_job_runs_status
    ON sync_job_runs (status, started_at DESC);

-- 6. 统一电子签留痕表
CREATE TABLE IF NOT EXISTS sign_record (
    id             VARCHAR(36)  NOT NULL,
    tenant_id      BIGINT       NOT NULL,
    target_type    VARCHAR(16)  NOT NULL,
    target_id      VARCHAR(64)  NOT NULL,
    signer_id      VARCHAR(64)  NOT NULL,
    signer_name    VARCHAR(64),
    sign_time      TIMESTAMP    NOT NULL,
    signature_blob TEXT,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_sign_record PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_sign_record_target
    ON sign_record (tenant_id, target_type, target_id);

-- 7. 医护人力排班月基线表
CREATE TABLE IF NOT EXISTS "Schedule_StaffDuty" (
    "Id"             BIGSERIAL   PRIMARY KEY,
    "TenantId"       BIGINT      NOT NULL,
    "CreatorId"      BIGINT,
    "CreateTime"     TIMESTAMP   NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMP   NOT NULL DEFAULT NOW(),
    "StaffId"        BIGINT      NOT NULL,
    "StaffName"      VARCHAR(64),
    "DutyRole"       VARCHAR(32) NOT NULL,
    "WardId"         BIGINT      NOT NULL,
    "DutyDate"       DATE        NOT NULL,
    "Shift"          VARCHAR(16)
);

CREATE INDEX IF NOT EXISTS "idx_staffduty_lookup"
    ON "Schedule_StaffDuty" ("TenantId", "WardId", "DutyDate", "DutyRole");

-- 8. 当日覆盖/顶班/换班表
CREATE TABLE IF NOT EXISTS "Schedule_StaffDutyOverride" (
    "Id"              BIGSERIAL   PRIMARY KEY,
    "TenantId"        BIGINT      NOT NULL,
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP   NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP   NOT NULL DEFAULT NOW(),
    "DutyDate"        DATE        NOT NULL,
    "WardId"          BIGINT      NOT NULL,
    "DutyRole"        VARCHAR(32) NOT NULL,
    "OriginalStaffId" BIGINT,
    "ActualStaffId"   BIGINT      NOT NULL,
    "ActualStaffName" VARCHAR(64),
    "Reason"          VARCHAR(128),
    "ChangedBy"       BIGINT
);

CREATE INDEX IF NOT EXISTS "idx_staffdutyoverride_lookup"
    ON "Schedule_StaffDutyOverride" ("TenantId", "WardId", "DutyDate", "DutyRole");

-- 9. 智能排班轻量患者档案表
CREATE TABLE IF NOT EXISTS "Schedule_Patient" (
    "Id"                BIGINT      NOT NULL,
    "TenantId"          BIGINT      NOT NULL,
    "Name"              VARCHAR(64) NOT NULL,
    "Gender"            VARCHAR(8),
    "InfectionStatus"   VARCHAR(16) NOT NULL DEFAULT 'unknown',
    "InfectionWaivedBy" BIGINT,
    "InfectionWaivedAt" TIMESTAMPTZ,
    "CreateTime"        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "idx_schedule_patient_tenant"
    ON "Schedule_Patient" ("TenantId");

COMMIT;
