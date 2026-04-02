-- 测试数据 SQL
-- 用于验证 API 格式，不从前端 Mock 导入
-- 手动创建 2-3 个具有代表性的患者即可

-- 1. 测试患者：张三 (治疗中状态)
INSERT INTO patients (id, name, age, gender, bed_number, diagnosis, risk_level, status, patient_type, insurance_type, dry_weight, default_mode, doctor_name, created_at, updated_at)
VALUES (
  'test-patient-001',
  '张三',
  45,
  '男',
  'A01',
  '慢性肾功能衰竭（尿毒症期）',
  '高危',
  'active',
  '门诊',
  '职工医保',
  65.5,
  'HDF',
  '李医生',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

-- 2. 测试患者：李四 (待诊状态)
INSERT INTO patients (id, name, age, gender, bed_number, diagnosis, risk_level, status, patient_type, insurance_type, dry_weight, default_mode, doctor_name, created_at, updated_at)
VALUES (
  'test-patient-002',
  '李四',
  52,
  '女',
  'A02',
  '糖尿病肾病',
  '中危',
  'active',
  '住院',
  '居民医保',
  58.0,
  'HD',
  '王医生',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

-- 3. 血管通路数据
INSERT INTO vascular_accesses (id, patient_id, type, site, status, first_use_date, notes, created_at, updated_at)
VALUES
  ('test-va-001', 'test-patient-001', '动静脉内瘘', '左前臂', '正常', '2023-01-20', '初期流量不稳，现已扩张良好', NOW(), NOW()),
  ('test-va-002', 'test-patient-002', '中心静脉导管', '右颈内', '正常', '2023-06-15', '临时导管', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 4. 病史记录
INSERT INTO medical_histories (id, patient_id, primary_disease, pathology, allergies, medical_history, complications, created_at, updated_at)
VALUES
  ('test-mh-001', 'test-patient-001', '慢性肾小球肾炎', 'IgA肾病 IV期', '青霉素', '高血压病史15年', '肾性高血压', NOW(), NOW()),
  ('test-mh-002', 'test-patient-002', '糖尿病肾病', '未活检', '无', '发现血糖高20年', '糖尿病视网膜病变', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 5. 感染信息
INSERT INTO infection_infos (id, patient_id, hbs_ag, hcv_ab, hiv_ab, tpa_b, update_date, created_at, updated_at)
VALUES
  ('test-inf-001', 'test-patient-001', '阴性', '阴性', '阴性', '阴性', '2026-01-20', NOW(), NOW()),
  ('test-inf-002', 'test-patient-002', '阴性', '阴性', '阴性', '阴性', '2026-01-20', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 6. 治疗方案（简化版，用于测试）
INSERT INTO treatment_plans (id, patient_id, weekly_frequency, biweekly_frequency, duration, dry_weight, extra_weight, notes, created_at, updated_at)
VALUES
  ('test-tp-001', 'test-patient-001', 3, 3, 4, 65.5, 0.6, 'AVF-左前臂', NOW(), NOW()),
  ('test-tp-002', 'test-patient-002', 2, 2, 4, 58.0, 0.5, 'TCC-右颈内', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 7. 住院信息
INSERT INTO hospitalizations (patient_id, case_no, hosp_no, admission_date, hosp_ward, hosp_bed, attend_dr, status, create_time, last_modify_time)
VALUES
  ('test-patient-002', 'CASE2026001', 'HOSP2026001', '2026-01-20', '肾内科', 'A02', '王医生', 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 验证插入结果
SELECT
  p.id,
  p.name,
  p.status,
  p.bed_number,
  va.type as vascular_access_type,
  mh.primary_disease,
  ii.hbs_ag
FROM patients p
LEFT JOIN vascular_accesses va ON va.patient_id = p.id
LEFT JOIN medical_histories mh ON mh.patient_id = p.id
LEFT JOIN infection_infos ii ON ii.patient_id = p.id
WHERE p.id LIKE 'test-%'
ORDER BY p.id;
