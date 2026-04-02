-- 转归字典结构迁移
-- 从两个独立的字典类型 (OUTCOME_TYPE, OUTCOME_REASON) 合并为一个树形结构 (OUTCOME)

-- 步骤 1: 创建新的 OUTCOME 字典类型
INSERT INTO dict_types (id, code, name, description, icon, sort_order, is_enabled, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'OUTCOME',
    '患者转归',
    '患者转归分类（一级：在科/转出，二级：具体原因）',
    '📋',
    210,
    true,
    NOW(),
    NOW()
)
ON CONFLICT (code) DO NOTHING;

-- 步骤 2: 迁移一级分类（从 OUTCOME_TYPE）
-- "在科"(10) - 一级，无父级
-- "转出"(20) - 一级，无父级
INSERT INTO dict_items (id, type_code, code, name, description, parent_code, sort_order, is_enabled, created_at, updated_at)
SELECT
    gen_random_uuid(),
    'OUTCOME' as type_code,
    code,
    name,
    description,
    NULL as parent_code,  -- 一级分类无父级
    sort_order,
    is_enabled,
    created_at,
    updated_at
FROM dict_items
WHERE type_code = 'OUTCOME_TYPE'
ON CONFLICT (type_code, code) DO NOTHING;

-- 步骤 3: 迁移二级分类（从 OUTCOME_REASON）
-- parent_code 保持原样 (10 或 20)，指向新的 OUTCOME 类型下的一级分类
INSERT INTO dict_items (id, type_code, code, name, description, parent_code, sort_order, is_enabled, created_at, updated_at)
SELECT
    gen_random_uuid(),
    'OUTCOME' as type_code,
    code,
    name,
    description,
    parent_code,  -- 保持原有的 parent_code (10 或 20)
    sort_order,
    is_enabled,
    created_at,
    updated_at
FROM dict_items
WHERE type_code = 'OUTCOME_REASON'
ON CONFLICT (type_code, code) DO NOTHING;

-- 步骤 4: 删除旧的字典项
DELETE FROM dict_items WHERE type_code IN ('OUTCOME_TYPE', 'OUTCOME_REASON');

-- 步骤 5: 删除旧的字典类型
DELETE FROM dict_types WHERE code IN ('OUTCOME_TYPE', 'OUTCOME_REASON');

-- 验证结果
SELECT
    type_code,
    code,
    name,
    parent_code,
    CASE WHEN parent_code IS NULL OR parent_code = '' THEN '一级' ELSE '二级' END as level
FROM dict_items
WHERE type_code = 'OUTCOME'
ORDER BY
    CASE WHEN parent_code IS NULL OR parent_code = '' THEN 0 ELSE 1 END,
    sort_order;
