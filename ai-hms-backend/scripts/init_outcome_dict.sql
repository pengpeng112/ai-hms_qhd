-- 转归字典初始化脚本
-- 功能：初始化 OUTCOME 字典数据（树形结构：一级类型 + 二级原因）
-- 创建日期：2026-02-11
-- 更新日期：2026-02-12 (合并 OUTCOME_TYPE/OUTCOME_REASON 为单一 OUTCOME)
-- 适用于：PostgreSQL 9.6+
-- 执行前请备份数据库

-- ============ 清理旧数据 ============

DELETE FROM dict_items WHERE type_code IN ('OUTCOME_TYPE', 'OUTCOME_REASON');
DELETE FROM dict_types WHERE code IN ('OUTCOME_TYPE', 'OUTCOME_REASON');

-- ============ 字典类型 (dict_types) ============

INSERT INTO dict_types (id, code, name, description, icon, sort_order, is_enabled, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'OUTCOME', '患者转归', '患者转归分类（一级：在科/转出，二级：具体原因）', '📋', 210, TRUE, NOW(), NOW())
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_enabled = EXCLUDED.is_enabled,
  updated_at = NOW();

-- ============ 字典项 (dict_items) ============

-- 一级分类（parent_code 为空）
INSERT INTO dict_items (id, type_code, code, name, description, parent_code, sort_order, is_enabled, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'OUTCOME', '10', '在科', '患者仍在科室治疗', NULL, 1, TRUE, NOW(), NOW()),
  (gen_random_uuid(), 'OUTCOME', '20', '转出', '患者转出科室', NULL, 2, TRUE, NOW(), NOW())
ON CONFLICT (type_code, code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  parent_code = EXCLUDED.parent_code,
  sort_order = EXCLUDED.sort_order,
  is_enabled = EXCLUDED.is_enabled,
  updated_at = NOW();

-- 二级分类 - "在科"(10)下的子项
INSERT INTO dict_items (id, type_code, code, name, description, parent_code, sort_order, is_enabled, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'OUTCOME', 'IN_DEPT', '在科', '患者仍在科室', '10', 1, TRUE, NOW(), NOW())
ON CONFLICT (type_code, code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  parent_code = EXCLUDED.parent_code,
  sort_order = EXCLUDED.sort_order,
  is_enabled = EXCLUDED.is_enabled,
  updated_at = NOW();

-- 二级分类 - "转出"(20)下的子项
INSERT INTO dict_items (id, type_code, code, name, description, parent_code, sort_order, is_enabled, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'OUTCOME', 'TRANSFER_OUT', '转外院', '转往外院治疗', '20', 1, TRUE, NOW(), NOW()),
  (gen_random_uuid(), 'OUTCOME', 'TRANSPLANT', '转肾移植', '转为肾移植治疗', '20', 2, TRUE, NOW(), NOW()),
  (gen_random_uuid(), 'OUTCOME', 'PD_TRANSFER', '转腹透', '转为腹膜透析', '20', 3, TRUE, NOW(), NOW()),
  (gen_random_uuid(), 'OUTCOME', 'CURED', '病愈', '患者病愈出院', '20', 4, TRUE, NOW(), NOW()),
  (gen_random_uuid(), 'OUTCOME', 'DEATH', '死亡', '患者死亡', '20', 5, TRUE, NOW(), NOW()),
  (gen_random_uuid(), 'OUTCOME', 'QUIT', '退出', '患者退出治疗', '20', 6, TRUE, NOW(), NOW())
ON CONFLICT (type_code, code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  parent_code = EXCLUDED.parent_code,
  sort_order = EXCLUDED.sort_order,
  is_enabled = EXCLUDED.is_enabled,
  updated_at = NOW();

-- ============ 验证查询 ============

SELECT
  di.type_code,
  di.code,
  di.name,
  di.parent_code,
  CASE WHEN di.parent_code IS NULL OR di.parent_code = '' THEN '一级' ELSE '二级' END AS level,
  di.sort_order
FROM dict_items di
WHERE di.type_code = 'OUTCOME'
ORDER BY
  CASE WHEN di.parent_code IS NULL OR di.parent_code = '' THEN 0 ELSE 1 END,
  di.parent_code,
  di.sort_order;
