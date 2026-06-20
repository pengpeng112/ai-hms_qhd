-- ================================================================
-- DBA 人工脚本：老表扩展字段与必要回填
-- ================================================================
-- 适用范围：修改 legacy 既有 Schedule_* 表。
-- 执行要求：必须 DBA/研发确认维护窗口、备份、锁影响和字段语义后执行。
-- 禁止：应用启动、部署普通步骤、请求路径中自动执行本脚本。
-- ================================================================

BEGIN;

-- 1. Schedule_Ward：病区分区能力
-- ZoneType：A/B/C 分区，用于普通/隔离/HDF-CRRT 区域识别。
-- ParentWardId：父病区 ID，用于子分区归属。
-- IsSubZone：是否子分区。
ALTER TABLE "Schedule_Ward" ADD COLUMN IF NOT EXISTS "ZoneType" VARCHAR(8) NOT NULL DEFAULT 'A';
ALTER TABLE "Schedule_Ward" ADD COLUMN IF NOT EXISTS "ParentWardId" BIGINT;
ALTER TABLE "Schedule_Ward" ADD COLUMN IF NOT EXISTS "IsSubZone" BOOLEAN DEFAULT false;

-- 2. Schedule_Shift：统一班次编码
-- ShiftCode：统一早/中/晚班编码，避免仅依赖老 Type 字段。
ALTER TABLE "Schedule_Shift" ADD COLUMN IF NOT EXISTS "ShiftCode" VARCHAR(16) NOT NULL DEFAULT 'MORNING';

-- 3. Schedule_Bed：床位向设备/机位语义扩展
-- MachineType：设备类型，如 HD/HDF/CRRT。
-- SupportedModes：支持治疗模式集合。
-- PositionIndex：同病区内展示/排班排序。
-- LegacyBedName：保留老床位名，便于追溯。
-- Code：设备/机位编码。
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "MachineType" VARCHAR(8) NOT NULL DEFAULT 'HD';
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "SupportedModes" VARCHAR(64) NOT NULL DEFAULT 'HD';
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "PositionIndex" INT NOT NULL DEFAULT 0;
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "LegacyBedName" VARCHAR(256);
ALTER TABLE "Schedule_Bed" ADD COLUMN IF NOT EXISTS "Code" VARCHAR(64);

-- 4. Schedule_PatientShift：排班核心表 V2 状态与确认链路
-- DialysisMode：本次排班治疗模式。
-- SourceType：排班来源类型。
-- RecordForm：记录形态，普通/临时/补排等。
-- Confirm1At/2At/3At：三级确认时间。
-- Confirm1By/2By/3By：三级确认人员。
-- IsBorrowedSlot：是否借用机位。
-- CancelReason：取消/撤销原因。
-- SourceTemplateItemId：来源模板项追踪。
-- IsLocked：是否锁定排班。
-- MachineId：机位 ID，替代单纯 BedId 语义。
-- MakeupOfShiftId：补排来源排班 ID。
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "DialysisMode" VARCHAR(8) NOT NULL DEFAULT 'HD';
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "SourceType" SMALLINT NOT NULL DEFAULT 10;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "RecordForm" SMALLINT NOT NULL DEFAULT 10;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm1At" TIMESTAMPTZ;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm2At" TIMESTAMPTZ;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm3At" TIMESTAMPTZ;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm1By" BIGINT;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm2By" BIGINT;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "Confirm3By" BIGINT;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "IsBorrowedSlot" BOOLEAN DEFAULT false;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "CancelReason" VARCHAR(256);
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "SourceTemplateItemId" BIGINT;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "IsLocked" BOOLEAN DEFAULT false;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "MachineId" BIGINT NOT NULL DEFAULT 0;
ALTER TABLE "Schedule_PatientShift" ADD COLUMN IF NOT EXISTS "MakeupOfShiftId" BIGINT;

-- 老列默认值修正：避免新写入时因老列无默认值导致 GORM INSERT 失败。
ALTER TABLE "Schedule_PatientShift" ALTER COLUMN "PatientPlanId" SET DEFAULT 0;
ALTER TABLE "Schedule_PatientShift" ALTER COLUMN "ShiftTiming" SET DEFAULT 0;
ALTER TABLE "Schedule_PatientShift" ALTER COLUMN "BedId" SET DEFAULT 0;
ALTER TABLE "Schedule_PatientShift" ALTER COLUMN "ShiftId" SET DEFAULT 0;

-- 5. Schedule_PatientProfile：排班患者扩展状态
-- WeeklyCount：每周透析次数。
-- PatientStatus：排班患者状态。
-- DischargeReason/At/By：转出/停排记录。
-- FixedHdMachineId/FixedHdfMachineId：固定 HD/HDF 机位。
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "WeeklyCount" SMALLINT;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "PatientStatus" SMALLINT NOT NULL DEFAULT 10;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "DischargeReason" VARCHAR(64);
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "DischargedAt" TIMESTAMPTZ;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "DischargedBy" BIGINT;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "FixedHdMachineId" BIGINT;
ALTER TABLE "Schedule_PatientProfile" ADD COLUMN IF NOT EXISTS "FixedHdfMachineId" BIGINT;

-- 6. Schedule_ScheduleTemplateItem：模板默认模式和固定机位
-- DefaultMode：模板项默认治疗模式。
-- FixedHdMachineId/FixedHdfMachineId：模板固定机位。
ALTER TABLE "Schedule_ScheduleTemplateItem" ADD COLUMN IF NOT EXISTS "DefaultMode" VARCHAR(8) NOT NULL DEFAULT 'HD';
ALTER TABLE "Schedule_ScheduleTemplateItem" ADD COLUMN IF NOT EXISTS "FixedHdMachineId" BIGINT;
ALTER TABLE "Schedule_ScheduleTemplateItem" ADD COLUMN IF NOT EXISTS "FixedHdfMachineId" BIGINT;

-- 7. Schedule_Calendar：开放范围配置
-- OpenWardIds/OpenMachineIds：按日限制开放病区/机位。
ALTER TABLE "Schedule_Calendar" ADD COLUMN IF NOT EXISTS "OpenWardIds" TEXT;
ALTER TABLE "Schedule_Calendar" ADD COLUMN IF NOT EXISTS "OpenMachineIds" TEXT;

-- 8. Schedule_MachineOutage：设备停机增加机位语义
ALTER TABLE "Schedule_MachineOutage" ADD COLUMN IF NOT EXISTS "MachineId" BIGINT NOT NULL DEFAULT 0;
ALTER TABLE "Schedule_MachineOutage" ALTER COLUMN "BedId" SET DEFAULT 0;

-- 9. Schedule_CrrtSession：CRRT 占用增加机位语义
ALTER TABLE "Schedule_CrrtSession" ADD COLUMN IF NOT EXISTS "MachineId" BIGINT NOT NULL DEFAULT 0;
ALTER TABLE "Schedule_CrrtSession" ALTER COLUMN "BedId" SET DEFAULT 0;

-- 10. Schedule_ConflictQueue：冲突处理建议与处理人字段
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "SuggestedDate" DATE;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "SuggestedShiftId" BIGINT;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "SuggestedBedId" BIGINT;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "SuggestedPatientShiftId" BIGINT;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "ResolvedBy" BIGINT;
ALTER TABLE "Schedule_ConflictQueue" ADD COLUMN IF NOT EXISTS "ResolvedAt" TIMESTAMPTZ;

COMMIT;

-- ================================================================
-- DBA 人工回填 SQL：执行前必须确认老扩展表是否存在和业务规则是否正确
-- ================================================================

-- A. 若存在 Schedule_BedMachineExt，可将设备扩展信息合并回 Schedule_Bed。
-- UPDATE "Schedule_Bed" b
-- SET "MachineType"    = COALESCE(e."MachineType", 'HD'),
--     "SupportedModes" = COALESCE(e."SupportedModes", 'HD'),
--     "PositionIndex"  = COALESCE(e."PositionIndex", 0),
--     "LegacyBedName"  = e."LegacyBedName",
--     "Code"           = COALESCE(e."MachineCode", b."Name")
-- FROM "Schedule_BedMachineExt" e
-- WHERE b."Id" = e."BedId" AND e."IsDisabled" = false;

-- B. 若存在 Schedule_PatientShiftExt，可将排班扩展信息合并回 Schedule_PatientShift。
-- UPDATE "Schedule_PatientShift" ps
-- SET "DialysisMode"         = COALESCE(ext."DialysisMode", 'HD'),
--     "SourceType"           = COALESCE(ext."SourceType", 10),
--     "RecordForm"           = COALESCE(ext."RecordForm", 10),
--     "Confirm1At"           = ext."Confirm1At",
--     "Confirm2At"           = ext."Confirm2At",
--     "Confirm3At"           = ext."Confirm3At",
--     "Confirm1By"           = ext."Confirm1By",
--     "Confirm2By"           = ext."Confirm2By",
--     "Confirm3By"           = ext."Confirm3By",
--     "IsBorrowedSlot"       = COALESCE(ext."IsBorrowedSlot", false),
--     "CancelReason"         = ext."CancelReason",
--     "SourceTemplateItemId" = ext."SourceTemplateItemId",
--     "IsLocked"             = COALESCE(ext."IsLocked", false),
--     "MachineId"            = ps."BedId"
-- FROM "Schedule_PatientShiftExt" ext
-- WHERE ps."Id" = ext."PatientShiftId";

-- C. ZoneType/ShiftCode 自动推断必须由 DBA 与业务确认后再执行。
-- UPDATE "Schedule_Ward" SET "ZoneType" = 'A';
-- UPDATE "Schedule_Ward" SET "ZoneType" = 'B'
--   WHERE "InfectionType" IN ('乙肝', '丙肝', 'HCV', 'HBV', 'positive');
-- UPDATE "Schedule_Ward" SET "ZoneType" = 'C'
--   WHERE "PatientType" IN ('CRRT', 'HDF')
--     AND "InfectionType" NOT IN ('乙肝', '丙肝', 'HCV', 'HBV', 'positive');
-- UPDATE "Schedule_Shift" SET "ShiftCode" = 'MORNING'   WHERE "Type" = 1;
-- UPDATE "Schedule_Shift" SET "ShiftCode" = 'AFTERNOON' WHERE "Type" = 2;
-- UPDATE "Schedule_Shift" SET "ShiftCode" = 'NIGHT'     WHERE "Type" = 3;
-- UPDATE "Schedule_Shift" SET "ShiftCode" = 'MORNING'   WHERE "Type" IS NULL OR "Type" NOT IN (1,2,3);

-- D. MachineId 从 BedId 初始化。
-- UPDATE "Schedule_PatientShift" SET "MachineId" = "BedId" WHERE "MachineId" = 0 OR "MachineId" IS NULL;
-- UPDATE "Schedule_MachineOutage" SET "MachineId" = "BedId" WHERE "MachineId" = 0;
-- UPDATE "Schedule_CrrtSession" SET "MachineId" = "BedId" WHERE "MachineId" = 0;
-- UPDATE "Schedule_PatientProfile" SET "FixedHdMachineId" = "FixedHdBedId" WHERE "FixedHdMachineId" IS NULL;
-- UPDATE "Schedule_PatientProfile" SET "FixedHdfMachineId" = "FixedHdfBedId" WHERE "FixedHdfMachineId" IS NULL;
