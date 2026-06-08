-- v1.2 智能排班 v2 独立表 DDL + 数据桥接
-- 执行前务必在测试环境验证，生产库需备份后执行
-- 12 张新表 + 3 个桥接 INSERT，不修改任何现有 Schedule_* 表

-- ========== 基础资源 v2 ==========

CREATE TABLE IF NOT EXISTS "Schedule_v2_Ward" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "Name" VARCHAR(256) NOT NULL,
  "ZoneType" VARCHAR(8) NOT NULL DEFAULT 'A',
  "ParentWardId" BIGINT,
  "IsSubZone" BOOLEAN DEFAULT FALSE,
  "Sort" INT DEFAULT 0,
  "IsDisabled" BOOLEAN DEFAULT FALSE,
  "Note" VARCHAR(512),
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_v2_Shift" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "Name" VARCHAR(64) NOT NULL,
  "ShiftCode" VARCHAR(16) NOT NULL,
  "StartTime" VARCHAR(8),
  "EndTime" VARCHAR(8),
  "Sort" INT DEFAULT 0,
  "IsDisabled" BOOLEAN DEFAULT FALSE,
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_Machine" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "WardId" BIGINT NOT NULL,
  "Code" VARCHAR(64) NOT NULL,
  "Name" VARCHAR(256),
  "MachineType" VARCHAR(8) NOT NULL DEFAULT 'HD',
  "PositionIndex" INT NOT NULL DEFAULT 0,
  "IsDisabled" BOOLEAN DEFAULT FALSE,
  "Sort" INT DEFAULT 0,
  "LegacyBedId" BIGINT,
  "Note" VARCHAR(512),
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_v2_Calendar" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "CalDate" DATE NOT NULL,
  "IsDialysisDay" BOOLEAN NOT NULL DEFAULT TRUE,
  "HolidayMode" SMALLINT NOT NULL DEFAULT 0,
  "OpenWardIds" TEXT,
  "OpenMachineIds" TEXT,
  "Note" VARCHAR(256),
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

-- ========== 排班核心 v2 ==========

CREATE TABLE IF NOT EXISTS "Schedule_v2_PatientShift" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "PatientId" BIGINT NOT NULL,
  "ScheduleDate" DATE NOT NULL,
  "ShiftId" BIGINT,
  "WardId" BIGINT NOT NULL,
  "MachineId" BIGINT,
  "Status" SMALLINT NOT NULL DEFAULT 0,
  "DialysisMode" VARCHAR(8) NOT NULL DEFAULT 'HD',
  "SourceType" SMALLINT NOT NULL DEFAULT 10,
  "RecordForm" SMALLINT NOT NULL DEFAULT 10,
  "Confirm1At" TIMESTAMP, "Confirm1By" BIGINT,
  "Confirm2At" TIMESTAMP, "Confirm2By" BIGINT,
  "Confirm3At" TIMESTAMP, "Confirm3By" BIGINT,
  "IsBorrowedSlot" BOOLEAN DEFAULT FALSE,
  "CancelReason" VARCHAR(256),
  "MakeupOfShiftId" BIGINT,
  "SourceTemplateItemId" BIGINT,
  "IsLocked" BOOLEAN DEFAULT FALSE,
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_v2_MachineOutage" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "MachineId" BIGINT NOT NULL,
  "StartAt" TIMESTAMP NOT NULL,
  "EndAt" TIMESTAMP,
  "ShiftId" BIGINT,
  "OutageType" SMALLINT NOT NULL DEFAULT 10,
  "Reason" VARCHAR(512),
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_v2_CrrtSession" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "PatientShiftId" BIGINT NOT NULL,
  "MachineId" BIGINT NOT NULL,
  "StartAt" TIMESTAMP NOT NULL,
  "EndAt" TIMESTAMP,
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

-- ========== 排班属性 v2 ==========

CREATE TABLE IF NOT EXISTS "Schedule_v2_PatientProfile" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "PatientId" BIGINT NOT NULL,
  "ZoneTag" VARCHAR(8) NOT NULL DEFAULT 'A',
  "HomeWardId" BIGINT,
  "FreqPattern" SMALLINT NOT NULL DEFAULT 10,
  "ShiftId" BIGINT,
  "DefaultMode" VARCHAR(8) NOT NULL DEFAULT 'HD',
  "HdfEnabled" BOOLEAN DEFAULT FALSE,
  "HdfWeekday" SMALLINT,
  "HdfWeekParity" SMALLINT,
  "FixedHdMachineId" BIGINT,
  "FixedHdfMachineId" BIGINT,
  "IsAdmissionRejected" BOOLEAN DEFAULT FALSE,
  "EffectiveFrom" DATE,
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_v2_ScheduleTemplate" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "Name" VARCHAR(128) NOT NULL,
  "Scope" VARCHAR(8),
  "IsActive" BOOLEAN DEFAULT TRUE,
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_v2_ScheduleTemplateItem" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "TemplateId" BIGINT NOT NULL,
  "PatientId" BIGINT NOT NULL,
  "ZoneTag" VARCHAR(8) NOT NULL DEFAULT 'A',
  "WardId" BIGINT,
  "ShiftId" BIGINT,
  "FreqPattern" SMALLINT NOT NULL DEFAULT 10,
  "FixedHdMachineId" BIGINT,
  "FixedHdfMachineId" BIGINT,
  "HdfEnabled" BOOLEAN DEFAULT FALSE,
  "HdfWeekday" SMALLINT,
  "HdfWeekParity" SMALLINT,
  "TemplateVersion" INT NOT NULL DEFAULT 1,
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_v2_ConflictQueue" (
  "Id" BIGSERIAL PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "PatientId" BIGINT,
  "ScheduleDate" DATE,
  "ShiftId" BIGINT,
  "WardId" BIGINT,
  "ConflictType" VARCHAR(24) NOT NULL,
  "Severity" SMALLINT NOT NULL DEFAULT 10,
  "Detail" TEXT,
  "SuggestedShiftId" BIGINT,
  "Status" SMALLINT NOT NULL DEFAULT 0,
  "ResolvedBy" BIGINT,
  "ResolvedAt" TIMESTAMP,
  "CreatorId" BIGINT,
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "Schedule_Patient" (
  "Id" BIGINT PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "Name" VARCHAR(64) NOT NULL,
  "Gender" VARCHAR(8),
  "CreateTime" TIMESTAMP DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP DEFAULT NOW()
);

-- ========== 数据桥接（一次性快照） ==========

-- 桥接1: 病区 Schedule_Ward + Schedule_WardExt → Schedule_v2_Ward
INSERT INTO "Schedule_v2_Ward" (
  "Id", "TenantId", "Name", "ZoneType", "ParentWardId", "IsSubZone",
  "Sort", "IsDisabled", "Note", "CreatorId", "CreateTime", "LastModifyTime"
)
SELECT
  w."Id", w."TenantId", w."Name",
  COALESCE(e."ZoneType", 'A'),
  e."ParentWardId",
  COALESCE(e."IsSubZone", FALSE),
  COALESCE(w."Sort", 0),
  COALESCE(w."IsDisabled", FALSE),
  w."Note",
  w."CreatorId",
  w."CreateTime",
  w."LastModifyTime"
FROM "Schedule_Ward" w
LEFT JOIN "Schedule_WardExt" e ON e."WardId" = w."Id" AND e."TenantId" = w."TenantId"
WHERE NOT EXISTS (
  SELECT 1 FROM "Schedule_v2_Ward" v2 WHERE v2."Id" = w."Id" AND v2."TenantId" = w."TenantId"
);

-- 桥接2: 班次 Schedule_Shift → Schedule_v2_Shift
INSERT INTO "Schedule_v2_Shift" (
  "Id", "TenantId", "Name", "ShiftCode", "StartTime", "EndTime",
  "Sort", "IsDisabled", "CreatorId", "CreateTime", "LastModifyTime"
)
SELECT
  "Id", "TenantId", "Name",
  CASE
    WHEN "Sort" = 1 THEN 'MORNING'
    WHEN "Sort" = 2 THEN 'AFTERNOON'
    WHEN "Sort" = 3 THEN 'NIGHT'
    ELSE 'OTHER'
  END,
  "StartTime", "EndTime",
  "Sort", "IsDisabled",
  "CreatorId", "CreateTime", "LastModifyTime"
FROM "Schedule_Shift"
WHERE NOT EXISTS (
  SELECT 1 FROM "Schedule_v2_Shift" v2 WHERE v2."Id" = "Schedule_Shift"."Id" AND v2."TenantId" = "Schedule_Shift"."TenantId"
);

-- 桥接3: 机器 Schedule_BedMachineExt + Schedule_Bed → Schedule_Machine
INSERT INTO "Schedule_Machine" (
  "TenantId", "WardId", "Code", "Name", "MachineType",
  "PositionIndex", "IsDisabled", "Sort", "LegacyBedId"
)
SELECT
  e."TenantId",
  COALESCE(b."WardId", 0),
  e."MachineCode",
  e."LegacyBedName",
  e."MachineType",
  e."PositionIndex",
  e."IsDisabled",
  e."PositionIndex",
  e."BedId"
FROM "Schedule_BedMachineExt" e
JOIN "Schedule_Bed" b ON b."Id" = e."BedId" AND b."TenantId" = e."TenantId"
WHERE NOT EXISTS (
  SELECT 1 FROM "Schedule_Machine" m
  WHERE m."LegacyBedId" = e."BedId" AND m."TenantId" = e."TenantId"
);
