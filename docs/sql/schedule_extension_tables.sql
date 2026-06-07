-- 透析排班新规则扩展表 DDL
-- 用途：在老血透库基础上新增排班规则扩展表，不修改任何老表结构。
-- 注意：本脚本只定义表结构和索引，不含 DROP TABLE 或 ALTER 老表。
-- 执行前请确认当前 schema 和目标数据库。
-- 所有表和列名沿用老库 PascalCase 风格，双引号引用。

-- ============================================================
-- 4.1 Schedule_WardExt  病区扩展（A/B/C 分区 + 子区树）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_WardExt" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "WardId"          BIGINT NOT NULL,
    "ZoneType"        VARCHAR(8) NOT NULL DEFAULT 'A',
    "ParentWardId"    BIGINT,
    "IsSubZone"       BOOLEAN NOT NULL DEFAULT FALSE,
    "Note"            VARCHAR(512),
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_WardExt_Tenant_Ward"    ON "Schedule_WardExt" ("TenantId", "WardId");
CREATE        INDEX IF NOT EXISTS "idx_WardExt_Tenant_Zone"    ON "Schedule_WardExt" ("TenantId", "ZoneType");


-- ============================================================
-- 4.2 Schedule_BedMachineExt  床位/机器扩展（HD/HDF/CRRT 能力）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_BedMachineExt" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "BedId"           BIGINT NOT NULL,
    "MachineCode"     VARCHAR(64),
    "MachineType"     VARCHAR(8) NOT NULL DEFAULT 'HD',
    "SupportedModes"  VARCHAR(64) NOT NULL DEFAULT 'HD',
    "PositionIndex"   INT NOT NULL DEFAULT 0,
    "IsDisabled"      BOOLEAN NOT NULL DEFAULT FALSE,
    "LegacyBedName"   VARCHAR(256),
    "Note"            VARCHAR(512),
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_BedMachineExt_Tenant_Bed"   ON "Schedule_BedMachineExt" ("TenantId", "BedId");
CREATE        INDEX IF NOT EXISTS "idx_BedMachineExt_Tenant_Type"  ON "Schedule_BedMachineExt" ("TenantId", "MachineType");


-- ============================================================
-- 4.3 Schedule_PatientProfile  患者排班骨架（人工权限属性，算法只读）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_PatientProfile" (
    "Id"                   BIGSERIAL PRIMARY KEY,
    "TenantId"             BIGINT NOT NULL,
    "PatientId"            BIGINT NOT NULL,
    "ZoneTag"              VARCHAR(8) NOT NULL DEFAULT 'A',
    "HomeWardId"           BIGINT,
    "FreqPattern"          SMALLINT NOT NULL DEFAULT 10,
    "ShiftId"              BIGINT,
    "DefaultMode"          VARCHAR(8) NOT NULL DEFAULT 'HD',
    "HdfEnabled"           BOOLEAN NOT NULL DEFAULT FALSE,
    "HdfWeekday"           SMALLINT,
    "HdfWeekParity"        SMALLINT,
    "FixedHdBedId"         BIGINT,
    "FixedHdfBedId"        BIGINT,
    "IsAdmissionRejected"  BOOLEAN NOT NULL DEFAULT FALSE,
    "EffectiveFrom"        DATE,
    "CreatorId"            BIGINT,
    "CreateTime"           TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_PatientProfile_Tenant_Patient"     ON "Schedule_PatientProfile" ("TenantId", "PatientId");
CREATE        INDEX IF NOT EXISTS "idx_PatientProfile_Tenant_Zone_Shift"  ON "Schedule_PatientProfile" ("TenantId", "ZoneTag", "ShiftId");


-- ============================================================
-- 4.4 Schedule_PatientShiftExt  排班记录扩展（新状态机正交字段 + 三级确认）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_PatientShiftExt" (
    "Id"                    BIGSERIAL PRIMARY KEY,
    "TenantId"              BIGINT NOT NULL,
    "PatientShiftId"        BIGINT NOT NULL,
    "DialysisMode"          VARCHAR(8) NOT NULL DEFAULT 'HD',
    "SourceType"            SMALLINT NOT NULL DEFAULT 10,
    "RecordForm"            SMALLINT NOT NULL DEFAULT 10,
    "Confirm1At"            TIMESTAMP,
    "Confirm2At"            TIMESTAMP,
    "Confirm3At"            TIMESTAMP,
    "Confirm1By"            BIGINT,
    "Confirm2By"            BIGINT,
    "Confirm3By"            BIGINT,
    "IsBorrowedSlot"        BOOLEAN NOT NULL DEFAULT FALSE,
    "BorrowedFromShiftId"   BIGINT,
    "IsLocked"              BOOLEAN NOT NULL DEFAULT FALSE,
    "CancelReason"          VARCHAR(256),
    "SourceTemplateItemId"  BIGINT,
    "SourceTemplateVersion" INT,
    "RuleStatus"            SMALLINT NOT NULL DEFAULT 10,
    "ApprovedBy"            BIGINT,
    "CreatorId"             BIGINT,
    "CreateTime"            TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"        TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_PatientShiftExt_Tenant_Shift"   ON "Schedule_PatientShiftExt" ("TenantId", "PatientShiftId");
CREATE        INDEX IF NOT EXISTS "idx_PatientShiftExt_Tenant_Status"  ON "Schedule_PatientShiftExt" ("TenantId", "RuleStatus");


-- ============================================================
-- 4.5 Schedule_ScheduleTemplate  模板头（替代 Status=60）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_ScheduleTemplate" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "Name"            VARCHAR(128) NOT NULL,
    "Scope"           VARCHAR(8),
    "WardId"          BIGINT,
    "IsActive"        BOOLEAN NOT NULL DEFAULT TRUE,
    "Version"         INT NOT NULL DEFAULT 1,
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_ScheduleTemplate_Tenant_Active" ON "Schedule_ScheduleTemplate" ("TenantId", "IsActive");


-- ============================================================
-- 4.6 Schedule_ScheduleTemplateItem  模板项（1 病人 1 项的稳定骨架）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_ScheduleTemplateItem" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "TemplateId"      BIGINT NOT NULL,
    "PatientId"       BIGINT NOT NULL,
    "ZoneTag"         VARCHAR(8) NOT NULL DEFAULT 'A',
    "WardId"          BIGINT,
    "ShiftId"         BIGINT,
    "FreqPattern"     SMALLINT NOT NULL DEFAULT 10,
    "FixedHdBedId"    BIGINT,
    "FixedHdfBedId"   BIGINT,
    "HdfEnabled"      BOOLEAN NOT NULL DEFAULT FALSE,
    "HdfWeekday"      SMALLINT,
    "HdfWeekParity"   SMALLINT,
    "TemplateVersion" INT NOT NULL DEFAULT 1,
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_TemplateItem_Tenant_Tmpl_Patient" ON "Schedule_ScheduleTemplateItem" ("TenantId", "TemplateId", "PatientId");


-- ============================================================
-- 4.7 Schedule_ConflictQueue  冲突/待处理队列
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_ConflictQueue" (
    "Id"                       BIGSERIAL PRIMARY KEY,
    "TenantId"                 BIGINT NOT NULL,
    "PatientId"                BIGINT,
    "ScheduleDate"             DATE,
    "ShiftId"                  BIGINT,
    "WardId"                   BIGINT,
    "ConflictType"             VARCHAR(24) NOT NULL,
    "Severity"                 SMALLINT NOT NULL DEFAULT 10,
    "Detail"                   TEXT,
    "SuggestedDate"            DATE,
    "SuggestedShiftId"         BIGINT,
    "SuggestedBedId"           BIGINT,
    "SuggestedPatientShiftId"  BIGINT,
    "Status"                   SMALLINT NOT NULL DEFAULT 0,
    "ResolvedBy"               BIGINT,
    "ResolvedAt"               TIMESTAMP,
    "CreatorId"                BIGINT,
    "CreateTime"               TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"           TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_ConflictQueue_Tenant_Status_Date" ON "Schedule_ConflictQueue" ("TenantId", "Status", "ScheduleDate");


-- ============================================================
-- 4.8 Schedule_MachineOutage  设备停机时段
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_MachineOutage" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "BedId"           BIGINT NOT NULL,
    "StartAt"         TIMESTAMP NOT NULL,
    "EndAt"           TIMESTAMP,
    "ShiftId"         BIGINT,
    "OutageType"      SMALLINT NOT NULL DEFAULT 10,
    "Reason"          VARCHAR(512),
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_MachineOutage_Tenant_Bed_Start_End" ON "Schedule_MachineOutage" ("TenantId", "BedId", "StartAt", "EndAt");


-- ============================================================
-- 4.9 Schedule_Calendar  机构日历（非透析日、假日值班）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_Calendar" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "CalDate"         DATE NOT NULL,
    "IsDialysisDay"   BOOLEAN NOT NULL DEFAULT TRUE,
    "HolidayMode"     SMALLINT NOT NULL DEFAULT 0,
    "Note"            VARCHAR(256),
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_Calendar_Tenant_Date" ON "Schedule_Calendar" ("TenantId", "CalDate");


-- ============================================================
-- 4.9a Schedule_CalendarOpenWard  假日值班开放病区（关联表）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_CalendarOpenWard" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "CalendarId"      BIGINT NOT NULL,
    "WardId"          BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_CalendarOpenWard_Tenant_Cal_Ward" ON "Schedule_CalendarOpenWard" ("TenantId", "CalendarId", "WardId");


-- ============================================================
-- 4.9b Schedule_CalendarOpenBed  假日值班开放机器（关联表）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_CalendarOpenBed" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "CalendarId"      BIGINT NOT NULL,
    "BedId"           BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_CalendarOpenBed_Tenant_Cal_Bed" ON "Schedule_CalendarOpenBed" ("TenantId", "CalendarId", "BedId");


-- ============================================================
-- 4.10 Schedule_CrrtSession  CRRT 占用（跨班跨天，不同于规律透析）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_CrrtSession" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "PatientShiftId"  BIGINT NOT NULL,
    "BedId"           BIGINT NOT NULL,
    "StartAt"         TIMESTAMP NOT NULL,
    "EndAt"           TIMESTAMP,
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_CrrtSession_Tenant_Shift"     ON "Schedule_CrrtSession" ("TenantId", "PatientShiftId");
CREATE        INDEX IF NOT EXISTS "idx_CrrtSession_Tenant_Bed_Start_End" ON "Schedule_CrrtSession" ("TenantId", "BedId", "StartAt", "EndAt");


-- ============================================================
-- 4.11 Schedule_PlanChange  方案变更记录
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_PlanChange" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "PatientId"       BIGINT NOT NULL,
    "ChangeType"      VARCHAR(16) NOT NULL,
    "OldValue"        VARCHAR(64),
    "NewValue"        VARCHAR(64),
    "EffectiveDate"   DATE NOT NULL,
    "AffectedCount"   INT NOT NULL DEFAULT 0,
    "ProcessedAt"     TIMESTAMP,
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_PlanChange_Tenant_Patient_Date" ON "Schedule_PlanChange" ("TenantId", "PatientId", "EffectiveDate");


-- ============================================================
-- 4.12 Schedule_TenantSetting  排班配置（奇偶周锚点、阈值等）
-- ============================================================
CREATE TABLE IF NOT EXISTS "Schedule_TenantSetting" (
    "Id"              BIGSERIAL PRIMARY KEY,
    "TenantId"        BIGINT NOT NULL,
    "SettingKey"      VARCHAR(64) NOT NULL,
    "SettingValue"    VARCHAR(256) NOT NULL,
    "SettingType"     VARCHAR(16) NOT NULL DEFAULT 'string',
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_TenantSetting_Tenant_Key" ON "Schedule_TenantSetting" ("TenantId", "SettingKey");
