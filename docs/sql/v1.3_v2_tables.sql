-- 透析排班 v1.3 表结构 DDL (Schedule_v2_*)
-- 与主系统 Schedule_* 表完全隔离，零 DDL 影响
-- 所有语句幂等（IF NOT EXISTS），可重复执行

CREATE TABLE IF NOT EXISTS "Schedule_v2_Ward" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "Name"           VARCHAR(256) NOT NULL,
    "ZoneType"       VARCHAR(8) NOT NULL,
    "ParentWardId"   BIGINT,
    "IsSubZone"      BOOLEAN NOT NULL DEFAULT FALSE,
    "Sort"           INT NOT NULL DEFAULT 0,
    "IsDisabled"     BOOLEAN NOT NULL DEFAULT FALSE,
    "Note"           VARCHAR(512)
);
CREATE INDEX IF NOT EXISTS idx_sv2_ward_tenant ON "Schedule_v2_Ward"("TenantId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_Machine" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "WardId"         BIGINT NOT NULL,
    "Code"           VARCHAR(64) NOT NULL,
    "Name"           VARCHAR(256),
    "MachineType"    VARCHAR(8) NOT NULL,
    "PositionIndex"  INT NOT NULL,
    "IsDisabled"     BOOLEAN NOT NULL DEFAULT FALSE,
    "Sort"           INT NOT NULL DEFAULT 0,
    "LegacyBedId"    BIGINT,
    "Note"           VARCHAR(512)
);
CREATE INDEX IF NOT EXISTS idx_sv2_machine_tenant ON "Schedule_v2_Machine"("TenantId");
CREATE INDEX IF NOT EXISTS idx_sv2_machine_ward ON "Schedule_v2_Machine"("WardId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_MachineOutage" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "MachineId"      BIGINT NOT NULL,
    "StartAt"        TIMESTAMPTZ NOT NULL,
    "EndAt"          TIMESTAMPTZ,
    "OutageType"     SMALLINT NOT NULL DEFAULT 10,
    "Reason"         VARCHAR(512)
);
CREATE INDEX IF NOT EXISTS idx_sv2_outage_tenant ON "Schedule_v2_MachineOutage"("TenantId");
CREATE INDEX IF NOT EXISTS idx_sv2_outage_machine ON "Schedule_v2_MachineOutage"("MachineId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_Shift" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "Name"           VARCHAR(64) NOT NULL,
    "ShiftCode"      VARCHAR(16) NOT NULL,
    "StartTime"      VARCHAR(8),
    "EndTime"        VARCHAR(8),
    "Sort"           INT NOT NULL DEFAULT 0,
    "IsDisabled"     BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_sv2_shift_tenant ON "Schedule_v2_Shift"("TenantId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_Calendar" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "CalDate"        DATE NOT NULL,
    "IsDialysisDay"  BOOLEAN NOT NULL DEFAULT TRUE,
    "HolidayMode"    SMALLINT NOT NULL DEFAULT 0,
    "OpenWardIds"    TEXT,
    "OpenMachineIds" TEXT,
    "Note"           VARCHAR(256)
);
CREATE INDEX IF NOT EXISTS idx_sv2_cal_tenant ON "Schedule_v2_Calendar"("TenantId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_PatientProfile" (
    "Id"                 BIGSERIAL PRIMARY KEY,
    "TenantId"           BIGINT NOT NULL DEFAULT 1,
    "CreatorId"          BIGINT NOT NULL DEFAULT 0,
    "CreateTime"         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "PatientId"          BIGINT NOT NULL,
    "ZoneTag"            VARCHAR(8) NOT NULL,
    "HomeWardId"         BIGINT,
    "WeeklyCount"        SMALLINT NOT NULL DEFAULT 0,
    "FreqPattern"        SMALLINT NOT NULL,
    "ShiftId"            BIGINT,
    "DefaultMode"        VARCHAR(8) NOT NULL DEFAULT 'HD',
    "HdfEnabled"         BOOLEAN NOT NULL DEFAULT FALSE,
    "HdfWeekday"         SMALLINT,
    "HdfWeekParity"      SMALLINT,
    "FixedHdMachineId"   BIGINT,
    "FixedHdfMachineId"  BIGINT,
    "IsAdmissionRejected" BOOLEAN NOT NULL DEFAULT FALSE,
    "EffectiveFrom"      DATE,
    "PatientStatus"      SMALLINT NOT NULL DEFAULT 10,
    "DischargeReason"    VARCHAR(64),
    "DischargedAt"       TIMESTAMPTZ,
    "DischargedBy"       BIGINT
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sv2_profile_patient ON "Schedule_v2_PatientProfile"("PatientId");
CREATE INDEX IF NOT EXISTS idx_sv2_profile_tenant ON "Schedule_v2_PatientProfile"("TenantId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_PlanChange" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "PatientId"      BIGINT NOT NULL,
    "ChangeType"     VARCHAR(16) NOT NULL,
    "OldValue"       VARCHAR(64),
    "NewValue"       VARCHAR(64),
    "EffectiveDate"  DATE NOT NULL,
    "AffectedCount"  INT NOT NULL DEFAULT 0,
    "ProcessedAt"    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sv2_plan_tenant ON "Schedule_v2_PlanChange"("TenantId");
CREATE INDEX IF NOT EXISTS idx_sv2_plan_patient ON "Schedule_v2_PlanChange"("PatientId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_PatientShift" (
    "Id"                   BIGSERIAL PRIMARY KEY,
    "TenantId"             BIGINT NOT NULL DEFAULT 1,
    "CreatorId"            BIGINT NOT NULL DEFAULT 0,
    "CreateTime"           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "PatientId"            BIGINT NOT NULL,
    "ScheduleDate"         DATE NOT NULL,
    "ShiftId"              BIGINT,
    "WardId"               BIGINT NOT NULL,
    "MachineId"            BIGINT,
    "Status"               SMALLINT NOT NULL DEFAULT 0,
    "DialysisMode"         VARCHAR(8) NOT NULL DEFAULT 'HD',
    "SourceType"           SMALLINT NOT NULL DEFAULT 10,
    "RecordForm"           SMALLINT NOT NULL DEFAULT 10,
    "Confirm1At"           TIMESTAMPTZ,
    "Confirm2At"           TIMESTAMPTZ,
    "Confirm3At"           TIMESTAMPTZ,
    "Confirm1By"           BIGINT,
    "Confirm2By"           BIGINT,
    "Confirm3By"           BIGINT,
    "IsBorrowedSlot"       BOOLEAN NOT NULL DEFAULT FALSE,
    "CancelReason"         VARCHAR(256),
    "MakeupOfShiftId"      BIGINT,
    "SourceTemplateItemId" BIGINT,
    "IsLocked"             BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_sv2_ps_tenant ON "Schedule_v2_PatientShift"("TenantId");
CREATE INDEX IF NOT EXISTS idx_sv2_ps_patient ON "Schedule_v2_PatientShift"("PatientId");
CREATE INDEX IF NOT EXISTS idx_sv2_ps_date ON "Schedule_v2_PatientShift"("ScheduleDate");
CREATE INDEX IF NOT EXISTS idx_sv2_ps_ward ON "Schedule_v2_PatientShift"("WardId");
CREATE INDEX IF NOT EXISTS idx_sv2_ps_machine ON "Schedule_v2_PatientShift"("MachineId");
CREATE INDEX IF NOT EXISTS idx_sv2_ps_status ON "Schedule_v2_PatientShift"("Status");
CREATE INDEX IF NOT EXISTS idx_sv2_ps_shift ON "Schedule_v2_PatientShift"("ShiftId");

-- 唯一索引：防并发重复排班/双占机位
CREATE UNIQUE INDEX IF NOT EXISTS uq_v2_ps_patient_slot
    ON "Schedule_v2_PatientShift" ("TenantId","PatientId","ScheduleDate","ShiftId")
    WHERE "Status" NOT IN (70,80) AND "ShiftId" IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_v2_ps_machine_slot
    ON "Schedule_v2_PatientShift" ("TenantId","MachineId","ScheduleDate","ShiftId")
    WHERE "Status" NOT IN (70,80) AND "MachineId" IS NOT NULL AND "ShiftId" IS NOT NULL;

CREATE TABLE IF NOT EXISTS "Schedule_v2_CrrtSession" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "PatientShiftId" BIGINT NOT NULL,
    "MachineId"      BIGINT NOT NULL,
    "StartAt"        TIMESTAMPTZ NOT NULL,
    "EndAt"          TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sv2_crrt_shift ON "Schedule_v2_CrrtSession"("PatientShiftId");
CREATE INDEX IF NOT EXISTS idx_sv2_crrt_tenant ON "Schedule_v2_CrrtSession"("TenantId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_ScheduleTemplate" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "Name"           VARCHAR(128) NOT NULL,
    "Scope"          VARCHAR(8),
    "IsActive"       BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_sv2_tpl_tenant ON "Schedule_v2_ScheduleTemplate"("TenantId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_ScheduleTemplateItem" (
    "Id"               BIGSERIAL PRIMARY KEY,
    "TenantId"         BIGINT NOT NULL DEFAULT 1,
    "CreatorId"        BIGINT NOT NULL DEFAULT 0,
    "CreateTime"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "TemplateId"       BIGINT NOT NULL,
    "PatientId"        BIGINT NOT NULL,
    "ZoneTag"          VARCHAR(8) NOT NULL,
    "WardId"           BIGINT,
    "ShiftId"          BIGINT,
    "FreqPattern"      SMALLINT NOT NULL,
    "DefaultMode"      VARCHAR(8) NOT NULL DEFAULT 'HD',
    "FixedHdMachineId"  BIGINT,
    "FixedHdfMachineId" BIGINT,
    "HdfEnabled"       BOOLEAN NOT NULL DEFAULT FALSE,
    "HdfWeekday"       SMALLINT,
    "HdfWeekParity"    SMALLINT
);
CREATE INDEX IF NOT EXISTS idx_sv2_ti_tenant ON "Schedule_v2_ScheduleTemplateItem"("TenantId");
CREATE INDEX IF NOT EXISTS idx_sv2_ti_template ON "Schedule_v2_ScheduleTemplateItem"("TemplateId");
CREATE INDEX IF NOT EXISTS idx_sv2_ti_patient ON "Schedule_v2_ScheduleTemplateItem"("PatientId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_ConflictQueue" (
    "Id"               BIGSERIAL PRIMARY KEY,
    "TenantId"         BIGINT NOT NULL DEFAULT 1,
    "CreatorId"        BIGINT NOT NULL DEFAULT 0,
    "CreateTime"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "PatientId"        BIGINT,
    "ScheduleDate"     DATE,
    "ShiftId"          BIGINT,
    "WardId"           BIGINT,
    "ConflictType"     VARCHAR(24) NOT NULL,
    "Severity"         SMALLINT NOT NULL DEFAULT 10,
    "Detail"           TEXT,
    "SuggestedShiftId" BIGINT,
    "Status"           SMALLINT NOT NULL DEFAULT 0,
    "ResolvedBy"       BIGINT,
    "ResolvedAt"       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sv2_cq_tenant ON "Schedule_v2_ConflictQueue"("TenantId");
CREATE INDEX IF NOT EXISTS idx_sv2_cq_patient ON "Schedule_v2_ConflictQueue"("PatientId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_Patient" (
    "Id"                BIGINT PRIMARY KEY,
    "TenantId"          BIGINT NOT NULL DEFAULT 1,
    "Name"              VARCHAR(64) NOT NULL,
    "Gender"            VARCHAR(8),
    "InfectionStatus"   VARCHAR(16) NOT NULL DEFAULT 'unknown',
    "InfectionWaivedBy" BIGINT,
    "InfectionWaivedAt" TIMESTAMPTZ,
    "CreateTime"        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sv2_patient_tenant ON "Schedule_v2_Patient"("TenantId");

CREATE TABLE IF NOT EXISTS "Schedule_v2_TenantSetting" (
    "Id"             BIGSERIAL PRIMARY KEY,
    "TenantId"       BIGINT NOT NULL DEFAULT 1,
    "CreatorId"      BIGINT NOT NULL DEFAULT 0,
    "CreateTime"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "SettingKey"     VARCHAR(64) NOT NULL,
    "SettingValue"   VARCHAR(256) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sv2_ts_tenant ON "Schedule_v2_TenantSetting"("TenantId");
