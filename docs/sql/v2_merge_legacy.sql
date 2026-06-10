-- ================================================================
-- 智能排班 V2 融合老表 — 数据库变更脚本
-- 
-- 目标：在现有 Schedule_* 表上添加 V2 需要的列，实现零表迁移融合。
-- 执行前提：已备份数据库。
-- 执行方式：psql -h 10.20.1.153 -U postgres -d dialysis -f v2_merge_legacy.sql
-- ================================================================

BEGIN;

-- ================================================================
-- 1. Schedule_Ward（加3列）
-- ================================================================
ALTER TABLE "Schedule_Ward" ADD COLUMN IF NOT EXISTS "ZoneType" varchar(8) NOT NULL DEFAULT 'A';
ALTER TABLE "Schedule_Ward" ADD COLUMN IF NOT EXISTS "ParentWardId" bigint;
ALTER TABLE "Schedule_Ward" ADD COLUMN IF NOT EXISTS "IsSubZone" boolean DEFAULT false;

-- ================================================================
-- 2. Schedule_Shift（加1列）
-- ================================================================
ALTER TABLE "Schedule_Shift" ADD COLUMN IF NOT EXISTS "ShiftCode" varchar(16) NOT NULL DEFAULT 'MORNING';

-- ================================================================
-- 3. Schedule_Bed（加5列）
-- ================================================================
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "MachineType" varchar(8) NOT NULL DEFAULT 'HD';
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "SupportedModes" varchar(64) NOT NULL DEFAULT 'HD';
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "PositionIndex" int NOT NULL DEFAULT 0;
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "LegacyBedName" varchar(256);
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "Code" varchar(64);

-- ================================================================
-- 4. Schedule_PatientShift（加15列）
-- ================================================================
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "DialysisMode" varchar(8) NOT NULL DEFAULT 'HD';
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "SourceType" smallint NOT NULL DEFAULT 10;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "RecordForm" smallint NOT NULL DEFAULT 10;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm1At" timestamptz;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm2At" timestamptz;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm3At" timestamptz;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm1By" bigint;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm2By" bigint;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm3By" bigint;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "IsBorrowedSlot" boolean DEFAULT false;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "CancelReason" varchar(256);
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "SourceTemplateItemId" bigint;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "IsLocked" boolean DEFAULT false;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "MachineId" bigint NOT NULL DEFAULT 0;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "MakeupOfShiftId" bigint;

-- 修复老列缺失默认值问题(避免GORM INSERT报错)
ALTER TABLE "Schedule_PatientShift" ALTER COLUMN "PatientPlanId" SET DEFAULT 0;
ALTER TABLE "Schedule_PatientShift" ALTER COLUMN "ShiftTiming" SET DEFAULT 0;
ALTER TABLE "Schedule_PatientShift" ALTER COLUMN "BedId" SET DEFAULT 0;
ALTER TABLE "Schedule_PatientShift" ALTER COLUMN "ShiftId" SET DEFAULT 0;

-- ================================================================
-- 5. Schedule_PatientProfile（加7列）
-- ================================================================
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "WeeklyCount" smallint;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "PatientStatus" smallint NOT NULL DEFAULT 10;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "DischargeReason" varchar(64);
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "DischargedAt" timestamptz;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "DischargedBy" bigint;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "FixedHdMachineId" bigint;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "FixedHdfMachineId" bigint;

-- ================================================================
-- 6. Schedule_ScheduleTemplateItem（加3列）
-- ================================================================
ALTER TABLE "Schedule_ScheduleTemplateItem" ADD COLUMN IF NOT EXISTS "DefaultMode" varchar(8) NOT NULL DEFAULT 'HD';
ALTER TABLE "Schedule_ScheduleTemplateItem" ADD COLUMN IF NOT EXISTS "FixedHdMachineId" bigint;
ALTER TABLE "Schedule_ScheduleTemplateItem" ADD COLUMN IF NOT EXISTS "FixedHdfMachineId" bigint;

-- ================================================================
-- 7. Schedule_Calendar（加2列）
-- ================================================================
ALTER TABLE "Schedule_Calendar" ADD COLUMN IF NOT EXISTS "OpenWardIds" text;
ALTER TABLE "Schedule_Calendar" ADD COLUMN IF NOT EXISTS "OpenMachineIds" text;

-- ================================================================
-- 8. Schedule_MachineOutage（加1列）
-- ================================================================
ALTER TABLE "Schedule_MachineOutage" ADD COLUMN IF NOT EXISTS "MachineId" bigint NOT NULL DEFAULT 0;
ALTER TABLE "Schedule_MachineOutage" ALTER COLUMN "BedId" SET DEFAULT 0;

-- ================================================================
-- 9. Schedule_CrrtSession（加1列）
-- ================================================================
ALTER TABLE "Schedule_CrrtSession" ADD COLUMN IF NOT EXISTS "MachineId" bigint NOT NULL DEFAULT 0;
ALTER TABLE "Schedule_CrrtSession" ALTER COLUMN "BedId" SET DEFAULT 0;

-- ================================================================
-- 10. Schedule_ConflictQueue（加5列,统一V2模型的列名）
-- ================================================================
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "SuggestedDate" date;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "SuggestedShiftId" bigint;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "SuggestedBedId" bigint;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "SuggestedPatientShiftId" bigint;
-- ⚠️ 若老库列名为 HandledBy/HandledAt, 需要重命名为 ResolvedBy/ResolvedAt:
-- ALTER TABLE "Schedule_ConflictQueue" RENAME COLUMN "HandledBy" TO "ResolvedBy";
-- ALTER TABLE "Schedule_ConflictQueue" RENAME COLUMN "HandledAt" TO "ResolvedAt";
-- 若老库本身就用 ResolvedBy/ResolvedAt, 则以下 ADD COLUMN 会跳过:
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "ResolvedBy" bigint;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "ResolvedAt" timestamptz;

COMMIT;

-- ================================================================
-- 新建表（需一次性创建）
-- ================================================================

-- 11. Schedule_Patient（V2轻量病人档）
CREATE TABLE IF NOT EXISTS "Schedule_Patient" (
  "Id" bigint NOT NULL,
  "TenantId" bigint NOT NULL,
  "Name" varchar(64) NOT NULL,
  "Gender" varchar(8),
  "InfectionStatus" varchar(16) NOT NULL DEFAULT 'unknown',
  "InfectionWaivedBy" bigint,
  "InfectionWaivedAt" timestamptz,
  "CreateTime" timestamptz NOT NULL DEFAULT now(),
  "LastModifyTime" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "idx_schedule_patient_tenant" ON "Schedule_Patient" ("TenantId");

-- ================================================================
-- 数据同步（执行顺序不可变）
-- ================================================================

-- 12. BedMachineExt → Bed 数据合并
UPDATE "Schedule_Bed" b
SET "MachineType"    = COALESCE(e."MachineType", 'HD'),
    "SupportedModes" = COALESCE(e."SupportedModes", 'HD'),
    "PositionIndex"  = COALESCE(e."PositionIndex", 0),
    "LegacyBedName"  = e."LegacyBedName",
    "Code"           = COALESCE(e."MachineCode", b."Name")
FROM "Schedule_BedMachineExt" e
WHERE b."Id" = e."BedId" AND e."IsDisabled" = false;

-- 13. PatientShiftExt → PatientShift 数据合并
UPDATE "Schedule_PatientShift" ps
SET "DialysisMode"         = COALESCE(ext."DialysisMode", 'HD'),
    "SourceType"           = COALESCE(ext."SourceType", 10),
    "RecordForm"           = COALESCE(ext."RecordForm", 10),
    "Confirm1At"           = ext."Confirm1At",
    "Confirm2At"           = ext."Confirm2At",
    "Confirm3At"           = ext."Confirm3At",
    "Confirm1By"           = ext."Confirm1By",
    "Confirm2By"           = ext."Confirm2By",
    "Confirm3By"           = ext."Confirm3By",
    "IsBorrowedSlot"       = COALESCE(ext."IsBorrowedSlot", false),
    "CancelReason"         = ext."CancelReason",
    "SourceTemplateItemId" = ext."SourceTemplateItemId",
    "IsLocked"             = COALESCE(ext."IsLocked", false),
    "MachineId"            = ps."BedId"
FROM "Schedule_PatientShiftExt" ext
WHERE ps."Id" = ext."PatientShiftId";

-- 14. Ward ZoneType 推断（⚠️ 需人工确认规则）
--    A区=普通透析区, B区=感染隔离区(乙肝/丙肝), C区=CRRT/HDF区
UPDATE "Schedule_Ward" SET "ZoneType" = 'A';
UPDATE "Schedule_Ward" SET "ZoneType" = 'B'
  WHERE "InfectionType" IN ('乙肝', '丙肝', 'HCV', 'HBV', 'positive');
UPDATE "Schedule_Ward" SET "ZoneType" = 'C'
  WHERE "PatientType" IN ('CRRT', 'HDF')
    AND "InfectionType" NOT IN ('乙肝', '丙肝', 'HCV', 'HBV', 'positive');

-- 15. Shift ShiftCode 推断（⚠️ 需人工确认规则）
--    Type=1→早班MORNING, Type=2→中班AFTERNOON, Type=3→晚班NIGHT
UPDATE "Schedule_Shift" SET "ShiftCode" = 'MORNING'   WHERE "Type" = 1;
UPDATE "Schedule_Shift" SET "ShiftCode" = 'AFTERNOON' WHERE "Type" = 2;
UPDATE "Schedule_Shift" SET "ShiftCode" = 'NIGHT'     WHERE "Type" = 3;
UPDATE "Schedule_Shift" SET "ShiftCode" = 'MORNING'   WHERE "Type" IS NULL OR "Type" NOT IN (1,2,3);

-- 16. MachineId 列同步（等于 BedId）
UPDATE "Schedule_PatientShift" SET "MachineId" = "BedId"
  WHERE "MachineId" = 0 OR "MachineId" IS NULL;
UPDATE "Schedule_MachineOutage" SET "MachineId" = "BedId"
  WHERE "MachineId" = 0;
UPDATE "Schedule_CrrtSession" SET "MachineId" = "BedId"
  WHERE "MachineId" = 0;
UPDATE "Schedule_PatientProfile" SET "FixedHdMachineId" = "FixedHdBedId"
  WHERE "FixedHdMachineId" IS NULL;
UPDATE "Schedule_PatientProfile" SET "FixedHdfMachineId" = "FixedHdfBedId"
  WHERE "FixedHdfMachineId" IS NULL;

-- ================================================================
-- ⚠️ 执行前必须手动确认以下项：
-- ================================================================
-- 1. Schedule_ConflictQueue 若由 schedule_conflict_queue.sql 创建,
--    其列名为 HandledBy/HandledAt 而非 ResolvedBy/ResolvedAt,
--    需取消上方 RENAME COLUMN 注释并执行。
-- 2. ZoneType 推断规则是否匹配实际病区分类。
-- 3. ShiftCode 推断规则是否匹配实际班次类型。
-- 4. BedMachineExt / PatientShiftExt 表是否存在(V2新建的扩展表),
--    若不存在则跳过步骤12-13。
-- ================================================================
-- 验证查询（执行后手动检查）：
-- SELECT "ZoneType", "Name", "PatientType", "InfectionType" FROM "Schedule_Ward";
-- SELECT "ShiftCode", "Name", "Type" FROM "Schedule_Shift";
-- SELECT COUNT(*) AS has_machine FROM "Schedule_PatientShift" WHERE "MachineId" > 0;
-- SELECT COUNT(*) AS total FROM "Schedule_PatientShift";