-- 当日覆盖（顶班/换班/请假）表（④ v2，契约04/05）
-- ⚠️ AutoMigrate 永久禁用：DBA 须执行本脚本建表。

CREATE TABLE IF NOT EXISTS "Schedule_StaffDutyOverride" (
    "Id"              BIGSERIAL   PRIMARY KEY,
    "TenantId"        BIGINT      NOT NULL,
    "CreatorId"       BIGINT,
    "CreateTime"      TIMESTAMP   NOT NULL DEFAULT NOW(),
    "LastModifyTime"  TIMESTAMP   NOT NULL DEFAULT NOW(),
    "DutyDate"        DATE        NOT NULL,
    "WardId"          BIGINT      NOT NULL,
    "DutyRole"        VARCHAR(32) NOT NULL,
    "OriginalStaffId" BIGINT,
    "ActualStaffId"   BIGINT      NOT NULL,
    "ActualStaffName" VARCHAR(64),
    "Reason"          VARCHAR(128),
    "ChangedBy"       BIGINT
);

CREATE INDEX IF NOT EXISTS "idx_staffdutyoverride_lookup"
    ON "Schedule_StaffDutyOverride" ("TenantId", "WardId", "DutyDate", "DutyRole");
