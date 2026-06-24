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
    target_type    VARCHAR(32)  NOT NULL,
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

-- 10. patient_infectious — 传染病筛查与阳性处置（规则A1 / 契约05批次2）
CREATE TABLE IF NOT EXISTS patient_infectious (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id varchar(64) NOT NULL,
    screen_date date,
    items text,
    source varchar(8),
    result_overall varchar(8),
    positive_markers varchar(128),
    next_due_date date,
    disposition varchar(16),
    handled_doctor_id varchar(64),
    handled_headnurse_id varchar(64),
    handled_at timestamptz,
    zone_tag varchar(16),
    note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_inf_tenant_patient ON patient_infectious (tenant_id, patient_id);

-- 11. patient_actr — ACTRS CTR/ACTR 镜像（契约05/07）
CREATE TABLE IF NOT EXISTS patient_actr (
    id               varchar(36) PRIMARY KEY,
    tenant_id        bigint NOT NULL,
    patient_id       varchar(64) NOT NULL,
    dialysis_no      varchar(64),
    actrs_xray_id    bigint NOT NULL,
    analysis_date    timestamptz,
    ctr              numeric,
    actr             numeric,
    actr1            numeric,
    actr2            numeric,
    actr_norm        numeric,
    heart_width      integer,
    lung_width       integer,
    tilt_angle       numeric,
    qc_pass          integer NOT NULL DEFAULT 0,
    qc_pa_ap         varchar(8),
    qc_warnings      varchar(256),
    model_version    varchar(32),
    source           varchar(16),
    image_path       varchar(256),
    overlay_path     varchar(256),
    mask_path        varchar(256),
    doctor_correction numeric,
    corrected_by     varchar(64),
    corrected_at     timestamptz,
    adopted_by       varchar(64),
    adopted_at       timestamptz,
    adopted_prescription_id varchar(32),
    adopted_dry_weight       numeric,
    adopted_uf_quantity      numeric,
    notes            varchar(256),
    synced_at        timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_actr_tenant_patient_xray
    ON patient_actr (tenant_id, patient_id, actrs_xray_id);
CREATE INDEX IF NOT EXISTS idx_actr_tenant_patient
    ON patient_actr (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_actr_adopted_prescription
    ON patient_actr (tenant_id, adopted_prescription_id);

-- 12. cnrds_report — CNRDS 上报包（规则 A4）
CREATE TABLE IF NOT EXISTS cnrds_report (
    id            varchar(36) PRIMARY KEY,
    tenant_id     bigint NOT NULL,
    period        varchar(16),
    report_type   varchar(12),
    event_type    varchar(16),
    patient_id    varchar(64),
    content       text,
    patient_count int NOT NULL DEFAULT 0,
    status        varchar(12) NOT NULL DEFAULT 'draft',
    export_ref    varchar(256),
    reviewed_by   varchar(64),
    submitted_at  timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cnrds_tenant_type_period
    ON cnrds_report (tenant_id, report_type, period);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cnrds_monthly_unique
    ON cnrds_report (tenant_id, report_type, period) WHERE report_type = 'monthly';

-- 13. water_quality — 透析用水/透析液质量监测（规则A2 / 契约05批次2）
CREATE TABLE IF NOT EXISTS water_quality (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    test_date date,
    test_type varchar(24),
    sample_point varchar(16),
    device_id varchar(64),
    value numeric,
    unit varchar(16),
    standard_limit varchar(32),
    result varchar(8),
    source varchar(12),
    next_due_date date,
    handled_engineer_id varchar(64),
    handled_headnurse_id varchar(64),
    handled_at timestamptz,
    action varchar(256),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_wq_tenant_type_date ON water_quality (tenant_id, test_type, test_date);

-- 14. disinfection_compliance — 透析机消毒监管伴生表（规则A3 / 契约05批次2）
CREATE TABLE IF NOT EXISTS disinfection_compliance (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    disinfection_id bigint NOT NULL,
    device_id bigint,
    concentration varchar(32),
    residual_check varchar(8),
    result varchar(8),
    source varchar(12),
    doc_ref varchar(256),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_disinfection ON disinfection_compliance (disinfection_id);
CREATE INDEX IF NOT EXISTS idx_dc_tenant_device ON disinfection_compliance (tenant_id, device_id);

-- 15. vascular_access_event — 血管通路全生命周期节点（规则B1 / 契约05批次2）
CREATE TABLE IF NOT EXISTS vascular_access_event (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    access_id bigint NOT NULL,
    patient_id bigint,
    event_type varchar(16),
    event_date date,
    detail text,
    operator_id varchar(64),
    note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_vae_tenant_patient ON vascular_access_event (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_vae_access ON vascular_access_event (access_id);

COMMENT ON TABLE vascular_access_event IS '血管通路全生命周期事件表';

COMMENT ON COLUMN vascular_access_event.id IS '主键ID';
COMMENT ON COLUMN vascular_access_event.tenant_id IS '租户ID';
COMMENT ON COLUMN vascular_access_event.access_id IS '血管通路ID，对应老库 Register_VascularAccess.Id';
COMMENT ON COLUMN vascular_access_event.patient_id IS '患者ID，冗余保存用于按患者查询';
COMMENT ON COLUMN vascular_access_event.event_type IS '事件类型：establish/maturation/first_use/physical_check/complication/intervention/failure/replacement';
COMMENT ON COLUMN vascular_access_event.event_date IS '事件发生日期';
COMMENT ON COLUMN vascular_access_event.detail IS '事件明细JSON，按事件类型约定结构';
COMMENT ON COLUMN vascular_access_event.operator_id IS '操作人ID';
COMMENT ON COLUMN vascular_access_event.note IS '备注';
COMMENT ON COLUMN vascular_access_event.created_at IS '创建时间';

-- 16. adverse_event 不良事件登记（规则B2）
CREATE TABLE IF NOT EXISTS adverse_event (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id bigint,
    treatment_id bigint,
    event_type varchar(64) NOT NULL,
    severity varchar(16) NOT NULL DEFAULT 'mild',
    occurred_at timestamptz NOT NULL,
    description text,
    handling text,
    outcome text,
    reporter_id varchar(64),
    reported_to text,
    reported_at timestamptz,
    within_6h boolean,
    status varchar(16) NOT NULL DEFAULT 'registered',
    cqi_linked boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ae_tenant_patient ON adverse_event (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_ae_status_severity ON adverse_event (tenant_id, status, severity);
CREATE INDEX IF NOT EXISTS idx_ae_within6h ON adverse_event (tenant_id, within_6h);

COMMENT ON TABLE adverse_event IS '不良事件登记表（规则B2）';
COMMENT ON COLUMN adverse_event.id IS '主键ID';
COMMENT ON COLUMN adverse_event.tenant_id IS '租户ID';
COMMENT ON COLUMN adverse_event.patient_id IS '患者ID';
COMMENT ON COLUMN adverse_event.treatment_id IS '关联治疗ID';
COMMENT ON COLUMN adverse_event.event_type IS '事件分类（COMPLICATION字典项）';
COMMENT ON COLUMN adverse_event.severity IS '严重程度：mild轻/moderate中/severe重';
COMMENT ON COLUMN adverse_event.occurred_at IS '发生时间';
COMMENT ON COLUMN adverse_event.description IS '发生经过描述';
COMMENT ON COLUMN adverse_event.handling IS '处理措施';
COMMENT ON COLUMN adverse_event.outcome IS '转归结果';
COMMENT ON COLUMN adverse_event.reporter_id IS '上报人ID';
COMMENT ON COLUMN adverse_event.reported_to IS '上报对象JSON数组';
COMMENT ON COLUMN adverse_event.reported_at IS '上报时间';
COMMENT ON COLUMN adverse_event.within_6h IS '严重事件是否6小时内上报';
COMMENT ON COLUMN adverse_event.status IS '状态：registered已登记/reported已上报/acknowledged已受理/processing处理中/closed已结案';
COMMENT ON COLUMN adverse_event.cqi_linked IS '是否纳入CQI质控';
COMMENT ON COLUMN adverse_event.created_at IS '创建时间';
COMMENT ON COLUMN adverse_event.updated_at IS '更新时间';

-- 17. medication_admin 长嘱给药执行（规则B3）
CREATE TABLE IF NOT EXISTS medication_admin (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id bigint,
    order_id bigint NOT NULL,
    treatment_id bigint,
    drug_name varchar(128) NOT NULL,
    category varchar(32),
    dose varchar(64),
    route varchar(32),
    timing varchar(32),
    administered_by varchar(64) NOT NULL,
    administered_name varchar(64),
    administered_at timestamptz NOT NULL DEFAULT now(),
    second_check_by varchar(64),
    second_check_name varchar(64),
    second_check_at timestamptz,
    status varchar(16) NOT NULL DEFAULT 'recorded',
    note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ma_tenant_patient ON medication_admin (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_ma_treatment ON medication_admin (tenant_id, treatment_id);
CREATE INDEX IF NOT EXISTS idx_ma_order ON medication_admin (tenant_id, order_id);

COMMENT ON TABLE medication_admin IS '长嘱给药执行记录表（规则B3）';
COMMENT ON COLUMN medication_admin.id IS '主键ID';
COMMENT ON COLUMN medication_admin.tenant_id IS '租户ID';
COMMENT ON COLUMN medication_admin.patient_id IS '患者ID';
COMMENT ON COLUMN medication_admin.order_id IS '长嘱ID，对应老库 Order_PatientOrder.Id';
COMMENT ON COLUMN medication_admin.treatment_id IS '关联治疗ID';
COMMENT ON COLUMN medication_admin.drug_name IS '药品名称';
COMMENT ON COLUMN medication_admin.category IS '药品种类：EPO/iron/phos_binder/vitamin_d/antihypertensive/other';
COMMENT ON COLUMN medication_admin.dose IS '给药剂量';
COMMENT ON COLUMN medication_admin.route IS '给药途径：iv/po/im/sc/ivgtt';
COMMENT ON COLUMN medication_admin.timing IS '给药时机：pre_dialysis/start/1h/2h/end/post_dialysis';
COMMENT ON COLUMN medication_admin.administered_by IS '执行护士ID';
COMMENT ON COLUMN medication_admin.administered_name IS '执行护士姓名';
COMMENT ON COLUMN medication_admin.administered_at IS '给药执行时间';
COMMENT ON COLUMN medication_admin.second_check_by IS '核对人ID';
COMMENT ON COLUMN medication_admin.second_check_name IS '核对人姓名';
COMMENT ON COLUMN medication_admin.second_check_at IS '核对时间';
COMMENT ON COLUMN medication_admin.status IS '双核状态：recorded已记录/verified已验证';
COMMENT ON COLUMN medication_admin.note IS '备注';
COMMENT ON COLUMN medication_admin.created_at IS '创建时间';
COMMENT ON COLUMN medication_admin.updated_at IS '更新时间';

-- 18. dry_weight_assessment 干体重评估记录（规则B4）
CREATE TABLE IF NOT EXISTS dry_weight_assessment (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id bigint,
    assess_type varchar(16) NOT NULL DEFAULT 'daily',
    phase varchar(16) NOT NULL DEFAULT 'induction',
    sbp int,
    dbp int,
    heart_rate int,
    edema boolean NOT NULL DEFAULT false,
    palpitation boolean NOT NULL DEFAULT false,
    heart_failure boolean NOT NULL DEFAULT false,
    cramp boolean NOT NULL DEFAULT false,
    ctr double precision,
    actr double precision,
    bia_oh double precision,
    bia_tbw double precision,
    bia_ecw double precision,
    post_weight double precision,
    target_weight double precision,
    decision varchar(16),
    adjust_kg double precision,
    rna_setting double precision,
    main_met boolean NOT NULL DEFAULT false,
    failed_reasons text,
    assessor_id varchar(64),
    assessor_name varchar(64),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dwa_tenant_patient ON dry_weight_assessment (tenant_id, patient_id);

COMMENT ON TABLE dry_weight_assessment IS '干体重评估记录表（规则B4）';
COMMENT ON COLUMN dry_weight_assessment.id IS '主键ID';
COMMENT ON COLUMN dry_weight_assessment.tenant_id IS '租户ID';
COMMENT ON COLUMN dry_weight_assessment.patient_id IS '患者ID';
COMMENT ON COLUMN dry_weight_assessment.assess_type IS '评估类型：daily日常/cycle周期';
COMMENT ON COLUMN dry_weight_assessment.phase IS '阶段：induction诱导期/maintenance维持期';
COMMENT ON COLUMN dry_weight_assessment.sbp IS '收缩压mmHg';
COMMENT ON COLUMN dry_weight_assessment.dbp IS '舒张压mmHg';
COMMENT ON COLUMN dry_weight_assessment.heart_rate IS '心率bpm';
COMMENT ON COLUMN dry_weight_assessment.edema IS '是否显性水肿';
COMMENT ON COLUMN dry_weight_assessment.palpitation IS '是否心慌气短';
COMMENT ON COLUMN dry_weight_assessment.heart_failure IS '是否心衰';
COMMENT ON COLUMN dry_weight_assessment.cramp IS '是否肌肉痉挛';
COMMENT ON COLUMN dry_weight_assessment.ctr IS '心胸比CTR';
COMMENT ON COLUMN dry_weight_assessment.actr IS 'ACTR值（优先医生修正）';
COMMENT ON COLUMN dry_weight_assessment.bia_oh IS 'BIA OH值';
COMMENT ON COLUMN dry_weight_assessment.bia_tbw IS 'BIA TBW值';
COMMENT ON COLUMN dry_weight_assessment.bia_ecw IS 'BIA ECW值';
COMMENT ON COLUMN dry_weight_assessment.post_weight IS '透后体重kg';
COMMENT ON COLUMN dry_weight_assessment.target_weight IS '目标干体重kg';
COMMENT ON COLUMN dry_weight_assessment.decision IS '决策：hold维持/lower下调/raise上调';
COMMENT ON COLUMN dry_weight_assessment.adjust_kg IS '调整幅度kg';
COMMENT ON COLUMN dry_weight_assessment.rna_setting IS 'RNa设置值';
COMMENT ON COLUMN dry_weight_assessment.main_met IS '主判据是否全满足';
COMMENT ON COLUMN dry_weight_assessment.failed_reasons IS '未达标项JSON数组';
COMMENT ON COLUMN dry_weight_assessment.assessor_id IS '评估人ID';
COMMENT ON COLUMN dry_weight_assessment.assessor_name IS '评估人姓名';
COMMENT ON COLUMN dry_weight_assessment.created_at IS '创建时间';

-- 19. patient_dry_weight 患者确定干体重（规则B4）
CREATE TABLE IF NOT EXISTS patient_dry_weight (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id bigint NOT NULL,
    dry_weight double precision NOT NULL,
    standard_actr double precision,
    standard_ctr double precision,
    phase varchar(16) NOT NULL DEFAULT 'maintenance',
    confirmed_by varchar(64),
    confirmed_name varchar(64),
    confirmed_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pdw_tenant_patient ON patient_dry_weight (tenant_id, patient_id);

COMMENT ON TABLE patient_dry_weight IS '患者确定干体重表（规则B4，一患者一条）';
COMMENT ON COLUMN patient_dry_weight.id IS '主键ID';
COMMENT ON COLUMN patient_dry_weight.tenant_id IS '租户ID';
COMMENT ON COLUMN patient_dry_weight.patient_id IS '患者ID，唯一';
COMMENT ON COLUMN patient_dry_weight.dry_weight IS '确定干体重kg';
COMMENT ON COLUMN patient_dry_weight.standard_actr IS '锚定ACTR基线';
COMMENT ON COLUMN patient_dry_weight.standard_ctr IS '锚定CTR基线';
COMMENT ON COLUMN patient_dry_weight.phase IS '阶段：induction诱导期/maintenance维持期';
COMMENT ON COLUMN patient_dry_weight.confirmed_by IS '确定人ID';
COMMENT ON COLUMN patient_dry_weight.confirmed_name IS '确定人姓名';
COMMENT ON COLUMN patient_dry_weight.confirmed_at IS '确定时间';
COMMENT ON COLUMN patient_dry_weight.created_at IS '创建时间';
COMMENT ON COLUMN patient_dry_weight.updated_at IS '更新时间';

-- 20. nursing_doc — 护理文书：量表评估/护理记录/护理计划（规则C1 / 契约05批次2）
CREATE TABLE IF NOT EXISTS nursing_doc (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id varchar(64),
    treatment_id varchar(64),
    doc_type varchar(12),
    scale_type varchar(16),
    score int,
    risk_level varchar(12),
    content text,
    nurse_id varchar(64),
    nurse_name varchar(64),
    recorded_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_nd_tenant_patient ON nursing_doc (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_nd_treatment ON nursing_doc (treatment_id);
CREATE INDEX IF NOT EXISTS idx_nd_type ON nursing_doc (doc_type);

-- 21. consent_record — 知情同意（规则C2 / 契约05批次2；复用 sign_record 待签线）
CREATE TABLE IF NOT EXISTS consent_record (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id varchar(64),
    consent_type varchar(16),
    template_version varchar(32),
    signed_by varchar(64),
    sign_record_id varchar(64),
    issued_by varchar(64),
    signed_at timestamptz,
    expires_at timestamptz,
    status varchar(12),
    doc_ref varchar(256),
    note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cr_tenant_patient ON consent_record (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_cr_type ON consent_record (consent_type);
CREATE INDEX IF NOT EXISTS idx_cr_status ON consent_record (status);

COMMIT;

-- ================================================================
-- 字典种子数据：COMPLICATION
-- ================================================================
BEGIN;
INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT v.*
FROM (VALUES
    ('HYPOTENSION', 'COMPLICATION', '低血压', 0, false, 10, true),
    ('ARRHYTHMIA', 'COMPLICATION', '心律失常', 0, false, 20, true),
    ('ALLERGY', 'COMPLICATION', '过敏反应', 0, false, 30, true),
    ('AIR_EMBOLISM', 'COMPLICATION', '空气栓塞', 0, false, 40, true),
    ('COAGULATION', 'COMPLICATION', '凝血/透析器凝血', 0, false, 50, true),
    ('FEVER', 'COMPLICATION', '发热/致热原反应', 0, false, 60, true),
    ('CRAMP', 'COMPLICATION', '肌肉痉挛', 0, false, 70, true),
    ('PUNCTURE_BLEED', 'COMPLICATION', '穿刺点渗血/血肿', 0, false, 80, true),
    ('HEMOLYSIS', 'COMPLICATION', '溶血', 0, false, 90, true),
    ('CHEST_PAIN', 'COMPLICATION', '胸痛/背痛', 0, false, 100, true),
    ('NAUSEA_VOMIT', 'COMPLICATION', '恶心呕吐', 0, false, 110, true),
    ('HEADACHE', 'COMPLICATION', '头痛', 0, false, 120, true),
    ('HEART_FAILURE', 'COMPLICATION', '急性心衰/肺水肿', 0, false, 130, true),
    ('ACCESS_RELATED', 'COMPLICATION', '通路相关并发症', 0, false, 140, true)
) AS v("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
WHERE NOT EXISTS (
    SELECT 1 FROM "CodeDictionary_CodeDictionarys" d
    WHERE d."Code" = v."Code" AND d."Type" = v."Type"
);
COMMIT;

-- ================================================================
-- 22 - 24：C4 收费归集 + HIS 价表同步
-- ================================================================

-- 22. charge_record 收费归集清单头（规则C4）
CREATE TABLE IF NOT EXISTS charge_record (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id bigint,
    treatment_id bigint NOT NULL,
    prescription_id bigint,
    charge_date timestamptz,
    shift varchar(16),
    dialysis_mode varchar(16),
    access_type varchar(16),
    crrt_hours decimal(5,2),
    total_amount decimal(10,2),
    status varchar(16) NOT NULL DEFAULT 'draft',
    recorded_by varchar(64),
    recorded_name varchar(64),
    checked_by varchar(64),
    checked_name varchar(64),
    checked_at timestamptz,
    exported_at timestamptz,
    pushed_at timestamptz,
    note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chr_tenant_patient ON charge_record (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_chr_treatment ON charge_record (treatment_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_chr_tenant_treatment_active
ON charge_record (tenant_id, treatment_id)
WHERE status <> 'cancelled';
CREATE INDEX IF NOT EXISTS idx_chr_date ON charge_record (tenant_id, charge_date);
CREATE INDEX IF NOT EXISTS idx_chr_status ON charge_record (tenant_id, status);

COMMENT ON TABLE charge_record IS '收费归集清单头（规则C4）';

-- 23. charge_line 收费归集清单明细（规则C4）
CREATE TABLE IF NOT EXISTS charge_line (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    charge_record_id varchar(36) NOT NULL,
    category varchar(16) NOT NULL,
    item_code varchar(64),
    item_name varchar(128) NOT NULL,
    spec varchar(64),
    unit varchar(16),
    quantity decimal(10,2),
    unit_price decimal(10,2),
    amount decimal(10,2),
    billable boolean NOT NULL DEFAULT true,
    source varchar(8) NOT NULL DEFAULT 'auto',
    charge_item_id bigint,
    his_price_item_id varchar(36),
    his_item_code varchar(20),
    his_item_class varchar(1),
    his_item_name varchar(120),
    price_source varchar(32),
    matched_status varchar(16),
    note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cl_record ON charge_line (charge_record_id);
CREATE INDEX IF NOT EXISTS idx_cl_his_item_code ON charge_line (his_item_code);
CREATE INDEX IF NOT EXISTS idx_cl_match_status ON charge_line (matched_status);
CREATE INDEX IF NOT EXISTS idx_cl_category ON charge_line (category);

COMMENT ON TABLE charge_line IS '收费归集清单明细（规则C4）';

-- 24. his_price_item HIS price_list 本地镜像
CREATE TABLE IF NOT EXISTS his_price_item (
    id varchar(36) PRIMARY KEY,
    source_system varchar(32) NOT NULL DEFAULT 'HIS_ORACLE',
    item_class varchar(1),
    item_code varchar(20) NOT NULL,
    item_name varchar(120),
    item_spec varchar(50),
    units varchar(30),
    price decimal(9,3),
    prefer_price decimal(9,3),
    foreigner_price decimal(9,3),
    performed_by varchar(8),
    fee_type_mask integer,
    class_on_inp_rcpt varchar(1),
    class_on_outp_rcpt varchar(1),
    class_on_reckoning varchar(10),
    subj_code varchar(10),
    class_on_mr varchar(4),
    memo varchar(100),
    start_date timestamp,
    stop_date timestamp,
    operator_code varchar(8),
    enter_date timestamp,
    high_price decimal(10,4),
    material_code varchar(20),
    score_1 decimal(10,2),
    score_2 decimal(10,2),
    price_name_code varchar(20),
    control_flag varchar(1),
    input_code varchar(100),
    input_code_wb varchar(100),
    std_code_1 varchar(20),
    changed_memo varchar(40),
    class_on_insur_mr varchar(24),
    package_spec varchar(20),
    firm_id varchar(10),
    charge_according varchar(23),
    license_id varchar(20),
    update_flag decimal,
    dept_name varchar(100),
    update_flag_syb decimal,
    mr_bill_class varchar(4),
    class_on_mr_add varchar(4),
    cwtj_code varchar(20),
    high_value decimal(9,3),
    drg_code varchar(8),
    insur_update integer,
    stop_operator varchar(8),
    limit_quantity decimal(10,0),
    is_active boolean NOT NULL DEFAULT true,
    synced_at timestamp NOT NULL DEFAULT now(),
    sync_run_id varchar(36),
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_his_price_item_code
ON his_price_item (source_system, item_code);

CREATE INDEX IF NOT EXISTS idx_his_price_item_name
ON his_price_item (item_name);

CREATE INDEX IF NOT EXISTS idx_his_price_item_input_code
ON his_price_item (input_code);

CREATE INDEX IF NOT EXISTS idx_his_price_item_class
ON his_price_item (item_class);

CREATE INDEX IF NOT EXISTS idx_his_price_item_active
ON his_price_item (is_active, stop_date);

COMMENT ON TABLE his_price_item IS 'HIS price_list 本地镜像，用于收费归集查价和项目匹配';
