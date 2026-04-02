-- =============================================================================
-- seed_phase1.sql — 首轮环境初始化种子数据
-- 目标 DB: host=10.10.8.83, dbname=Postgre, user=amdin
-- 依赖: AutoMigrate 已执行（表结构已就绪）
-- 执行方式: 通过 seed_phase1.sh 脚本调用，或手动执行
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 可登录用户
--    username: test_admin / test_doctor
--    password: Test@123456  (bcrypt 哈希, cost=10)
--    哈希值: $2a$10$U6Hl3Fc/sjbAtHpjN.QE4uvRFkSFNTFRcCMYGL1tWTvtyaNsdZOLS
--    生成方式: go run scripts/gen_bcrypt_temp.go
-- -----------------------------------------------------------------------------
INSERT INTO users (
    id,
    username,
    password,
    real_name,
    phone,
    email,
    role,
    status,
    department_id,
    created_at,
    updated_at
) VALUES (
    'seed-user-001',
    'test_admin',
    '$2a$10$U6Hl3Fc/sjbAtHpjN.QE4uvRFkSFNTFRcCMYGL1tWTvtyaNsdZOLS',
    '测试管理员',
    '13800138000',
    'test_admin@aihms.local',
    'ADMIN',
    'active',
    NULL,
    NOW(),
    NOW()
)
ON CONFLICT (username) DO UPDATE
    SET password   = EXCLUDED.password,
        real_name  = EXCLUDED.real_name,
        role       = EXCLUDED.role,
        status     = EXCLUDED.status,
        updated_at = NOW();

-- 额外测试用户：普通医生账号（同密码）
INSERT INTO users (
    id,
    username,
    password,
    real_name,
    phone,
    email,
    role,
    status,
    department_id,
    created_at,
    updated_at
) VALUES (
    'seed-user-002',
    'test_doctor',
    '$2a$10$U6Hl3Fc/sjbAtHpjN.QE4uvRFkSFNTFRcCMYGL1tWTvtyaNsdZOLS',
    '测试医生',
    '13900139000',
    'test_doctor@aihms.local',
    'DOCTOR_CHIEF',
    'active',
    NULL,
    NOW(),
    NOW()
)
ON CONFLICT (username) DO UPDATE
    SET password   = EXCLUDED.password,
        real_name  = EXCLUDED.real_name,
        role       = EXCLUDED.role,
        status     = EXCLUDED.status,
        updated_at = NOW();

-- -----------------------------------------------------------------------------
-- 2. 患者基础数据
--    Patient.ID 类型: varchar(36) — 使用带前缀的字符串 ID
--    注意: Treatment.PatientId 是 int64 (bigint)，与此 varchar ID 不兼容 [TYPE-MISMATCH]
--    影响范围: Treatment_Treatment, patient_shifts, hospitalizations 三张表
-- -----------------------------------------------------------------------------
INSERT INTO patients (
    id,
    name,
    age,
    gender,
    bed_number,
    diagnosis,
    risk_level,
    status,
    patient_type,
    insurance_type,
    dry_weight,
    default_mode,
    doctor_id,
    doctor_name,
    admission_date,
    created_at,
    updated_at
) VALUES (
    'seed-patient-001',
    '张三',
    52,
    '男',
    'A01',
    '慢性肾功能衰竭（尿毒症期）',
    '高危',
    'active',
    '门诊',
    '职工医保',
    65.50,
    'HDF',
    'seed-user-002',
    '测试医生',
    '2023-06-01',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE
    SET name       = EXCLUDED.name,
        updated_at = NOW();

INSERT INTO patients (
    id,
    name,
    age,
    gender,
    bed_number,
    diagnosis,
    risk_level,
    status,
    patient_type,
    insurance_type,
    dry_weight,
    default_mode,
    doctor_id,
    doctor_name,
    admission_date,
    created_at,
    updated_at
) VALUES (
    'seed-patient-002',
    '李四',
    47,
    '女',
    'A02',
    '糖尿病肾病',
    '中危',
    'active',
    '住院',
    '居民医保',
    58.00,
    'HD',
    'seed-user-002',
    '测试医生',
    '2024-01-15',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE
    SET name       = EXCLUDED.name,
        updated_at = NOW();

-- -----------------------------------------------------------------------------
-- 3. 患者扩展基础信息（patient_basic_infos）
-- -----------------------------------------------------------------------------
INSERT INTO patient_basic_infos (
    id,
    patient_id,
    pinyin,
    birthday,
    ethnicity,
    id_type,
    id_number,
    visit_category,
    dialysis_no,
    nurse_name,
    first_dialysis_date,
    first_hospital_date,
    phone,
    address,
    contact_name,
    contact_phone,
    created_at,
    updated_at
) VALUES (
    'seed-pbi-001',
    'seed-patient-001',
    'Zhang San',
    '1972-03-15',
    '汉',
    '身份证',
    '110101197203150012',
    '门诊',
    'HD-2023-001',
    '王护士',
    '2023-06-01',
    '2023-06-01',
    '13700137001',
    '北京市朝阳区XX街道XX号',
    '张大明',
    '13700137002',
    NOW(),
    NOW()
)
ON CONFLICT (patient_id) DO UPDATE
    SET updated_at = NOW();

INSERT INTO patient_basic_infos (
    id,
    patient_id,
    pinyin,
    birthday,
    ethnicity,
    id_type,
    id_number,
    visit_category,
    dialysis_no,
    nurse_name,
    first_dialysis_date,
    first_hospital_date,
    phone,
    address,
    contact_name,
    contact_phone,
    created_at,
    updated_at
) VALUES (
    'seed-pbi-002',
    'seed-patient-002',
    'Li Si',
    '1977-08-20',
    '汉',
    '身份证',
    '110101197708200023',
    '住院',
    'HD-2024-002',
    '赵护士',
    '2024-01-15',
    '2024-01-15',
    '13600136001',
    '北京市海淀区YY街道YY号',
    '李大强',
    '13600136002',
    NOW(),
    NOW()
)
ON CONFLICT (patient_id) DO UPDATE
    SET updated_at = NOW();

-- -----------------------------------------------------------------------------
-- 4. 血管通路（vascular_accesses）
--    PatientID: varchar(36) — 与 Patient.ID 兼容
-- -----------------------------------------------------------------------------
INSERT INTO vascular_accesses (
    id,
    patient_id,
    access_type,
    site,
    side,
    access_number,
    intervention_count,
    is_default,
    is_disabled,
    notes,
    created_at,
    updated_at
) VALUES (
    'seed-va-001',
    'seed-patient-001',
    '自体动静脉内瘘AVF',
    '左前臂',
    'L',
    1,
    0,
    true,
    false,
    '初期流量不稳，现已扩张良好',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO vascular_accesses (
    id,
    patient_id,
    access_type,
    site,
    side,
    access_number,
    intervention_count,
    is_default,
    is_disabled,
    notes,
    created_at,
    updated_at
) VALUES (
    'seed-va-002',
    'seed-patient-002',
    '带隧道和涤纶套的透析导管TCC',
    '右颈内',
    'R',
    1,
    0,
    true,
    false,
    '临时导管，待内瘘成熟后撤管',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 5. 感染信息（infection_infos）
--    PatientID: varchar(36) — 与 Patient.ID 兼容
-- -----------------------------------------------------------------------------
INSERT INTO infection_infos (
    id,
    patient_id,
    hbs_ag,
    hcv_ab,
    hiv_ab,
    tpa_b,
    update_date,
    created_at,
    updated_at
) VALUES (
    'seed-inf-001',
    'seed-patient-001',
    '阴性',
    '阴性',
    '阴性',
    '阴性',
    NOW(),
    NOW(),
    NOW()
)
ON CONFLICT (patient_id) DO UPDATE
    SET updated_at = NOW();

INSERT INTO infection_infos (
    id,
    patient_id,
    hbs_ag,
    hcv_ab,
    hiv_ab,
    tpa_b,
    update_date,
    created_at,
    updated_at
) VALUES (
    'seed-inf-002',
    'seed-patient-002',
    '阴性',
    '阳性',
    '阴性',
    '阴性',
    NOW(),
    NOW(),
    NOW()
)
ON CONFLICT (patient_id) DO UPDATE
    SET updated_at = NOW();

-- -----------------------------------------------------------------------------
-- 6. 治疗方案（treatment_plans）
--    PatientID: varchar(36) — 与 Patient.ID 兼容
-- -----------------------------------------------------------------------------
INSERT INTO treatment_plans (
    id,
    patient_id,
    weekly_frequency,
    biweekly_frequency,
    duration,
    dry_weight,
    extra_weight,
    status,
    doctor_id,
    start_date,
    notes,
    created_at,
    updated_at
) VALUES (
    'seed-tp-001',
    'seed-patient-001',
    3,
    3,
    4,
    65.50,
    0.60,
    '启用',
    'seed-user-002',
    '2023-06-01',
    'AVF-左前臂，每周三次HDF',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO treatment_plans (
    id,
    patient_id,
    weekly_frequency,
    biweekly_frequency,
    duration,
    dry_weight,
    extra_weight,
    status,
    doctor_id,
    start_date,
    notes,
    created_at,
    updated_at
) VALUES (
    'seed-tp-002',
    'seed-patient-002',
    2,
    2,
    4,
    58.00,
    0.50,
    '启用',
    'seed-user-002',
    '2024-01-15',
    'TCC-右颈内，每周两次HD',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 7. 病史记录（medical_histories）
--    PatientID: varchar(36) — 与 Patient.ID 兼容
-- -----------------------------------------------------------------------------
INSERT INTO medical_histories (
    id,
    patient_id,
    current_illness,
    past_history,
    transfusion_history,
    marital_history,
    family_history,
    disease_diagnosis,
    primary_disease_name,
    primary_disease_content,
    complication_name,
    complication_content,
    created_at,
    updated_at
) VALUES (
    'seed-mh-001',
    'seed-patient-001',
    '慢性肾功能衰竭，规律透析中',
    '高血压病史15年，口服降压药控制',
    '无输血史',
    '已婚，育有一子',
    '父亲有高血压病史',
    '慢性肾功能衰竭（尿毒症期）',
    '慢性肾小球肾炎',
    'IgA肾病 IV期',
    '肾性高血压',
    '血压控制尚可',
    NOW(),
    NOW()
)
ON CONFLICT (patient_id) DO UPDATE
    SET updated_at = NOW();

INSERT INTO medical_histories (
    id,
    patient_id,
    current_illness,
    past_history,
    transfusion_history,
    marital_history,
    family_history,
    disease_diagnosis,
    primary_disease_name,
    primary_disease_content,
    complication_name,
    complication_content,
    created_at,
    updated_at
) VALUES (
    'seed-mh-002',
    'seed-patient-002',
    '糖尿病肾病，维持性血液透析',
    '发现血糖高20年，胰岛素注射治疗',
    '无输血史',
    '已婚，育有两子',
    '母亲有糖尿病',
    '糖尿病肾病',
    '2型糖尿病肾病',
    '未活检',
    '糖尿病视网膜病变',
    '已行激光治疗',
    NOW(),
    NOW()
)
ON CONFLICT (patient_id) DO UPDATE
    SET updated_at = NOW();

-- -----------------------------------------------------------------------------
-- 8. 字典数据说明
-- 字典初始化通过 API 完成，不在此 SQL 中写入：
--   POST /api/v1/dict/items/init
--   该接口触发 DictService.InitDefaultDictData()
--   会自动创建 25 类型 + 约120 条字典项（DIALYSIS_MODE, ANTICOAGULANT 等）
--   首次部署时，请先执行本 SQL，启动后端服务，然后调用上述 API 初始化字典。
--
-- 如需直接 SQL 初始化字典，参考：scripts/init_outcome_dict.sql
-- -----------------------------------------------------------------------------

-- -----------------------------------------------------------------------------
-- 验证插入结果
-- -----------------------------------------------------------------------------
SELECT
    u.id,
    u.username,
    u.real_name,
    u.role,
    u.status
FROM users u
WHERE u.id IN ('seed-user-001', 'seed-user-002')
ORDER BY u.id;

SELECT
    p.id,
    p.name,
    p.gender,
    p.age,
    p.status,
    p.bed_number,
    p.risk_level,
    va.access_type AS vascular_access_type,
    mh.primary_disease_name,
    ii.hbs_ag
FROM patients p
LEFT JOIN vascular_accesses va ON va.patient_id = p.id AND va.is_default = true
LEFT JOIN medical_histories mh ON mh.patient_id = p.id
LEFT JOIN infection_infos ii ON ii.patient_id = p.id
WHERE p.id IN ('seed-patient-001', 'seed-patient-002')
ORDER BY p.id;
