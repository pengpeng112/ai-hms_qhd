-- seed_infectious_screen_item.sql
-- 传染病筛查项字典初始化（规则A1 / 契约05批次2）
-- 生成日期：2026-06-20
-- 说明：对应 models.DictTypeInfectiousScreenItem = "INFECTIOUS_SCREEN_ITEM"
--       前端 InfectiousPanel.tsx 的 SCREEN_ITEMS 下拉值须与本处 Code 完全一致。
-- 幂等：每项 WHERE NOT EXISTS，可重复执行。
-- 执行者：DBA（apply 代码后、功能启用前）

INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT 'HBsAg', 'INFECTIOUS_SCREEN_ITEM', '乙肝表面抗原(HBsAg)', 0, false, 10, true
WHERE NOT EXISTS (SELECT 1 FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' AND "Code"='HBsAg');

INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT '抗-HBs', 'INFECTIOUS_SCREEN_ITEM', '乙肝表面抗体(抗-HBs)', 0, false, 20, true
WHERE NOT EXISTS (SELECT 1 FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' AND "Code"='抗-HBs');

INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT 'HBeAg', 'INFECTIOUS_SCREEN_ITEM', '乙肝e抗原(HBeAg)', 0, false, 30, true
WHERE NOT EXISTS (SELECT 1 FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' AND "Code"='HBeAg');

INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT 'HBcAb', 'INFECTIOUS_SCREEN_ITEM', '乙肝核心抗体(HBcAb)', 0, false, 40, true
WHERE NOT EXISTS (SELECT 1 FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' AND "Code"='HBcAb');

INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT '抗-HCV', 'INFECTIOUS_SCREEN_ITEM', '丙肝抗体(抗-HCV)', 0, false, 50, true
WHERE NOT EXISTS (SELECT 1 FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' AND "Code"='抗-HCV');

INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT 'HIV 抗体', 'INFECTIOUS_SCREEN_ITEM', 'HIV抗体', 0, false, 60, true
WHERE NOT EXISTS (SELECT 1 FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' AND "Code"='HIV 抗体');

INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT 'TPPA', 'INFECTIOUS_SCREEN_ITEM', '梅毒螺旋体明胶凝集试验(TPPA)', 0, false, 70, true
WHERE NOT EXISTS (SELECT 1 FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' AND "Code"='TPPA');

INSERT INTO "CodeDictionary_CodeDictionarys" ("Code", "Type", "Name", "OrganId", "IsDisabled", "Sort", "Builtin")
SELECT 'RPR', 'INFECTIOUS_SCREEN_ITEM', '快速血浆反应素环状卡片试验(RPR)', 0, false, 80, true
WHERE NOT EXISTS (SELECT 1 FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' AND "Code"='RPR');

-- 校验
SELECT "Code", "Name", "Sort" FROM "CodeDictionary_CodeDictionarys" WHERE "Type"='INFECTIOUS_SCREEN_ITEM' ORDER BY "Sort";
