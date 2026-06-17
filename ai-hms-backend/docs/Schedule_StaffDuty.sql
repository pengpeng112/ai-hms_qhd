-- 医护人力排班·月基线表（④ v1，契约04/05）
-- ⚠️ 本项目 AutoMigrate 永久禁用：此表须由 DBA 执行本脚本建表后，人力排班/当班解析方可运行。

CREATE TABLE IF NOT EXISTS "Schedule_StaffDuty" (
    "Id"             BIGSERIAL   PRIMARY KEY,
    "TenantId"       BIGINT      NOT NULL,
    "CreatorId"      BIGINT,
    "CreateTime"     TIMESTAMP   NOT NULL DEFAULT NOW(),
    "LastModifyTime" TIMESTAMP   NOT NULL DEFAULT NOW(),
    "StaffId"        BIGINT      NOT NULL,
    "StaffName"      VARCHAR(64),
    "DutyRole"       VARCHAR(32) NOT NULL,
    "WardId"         BIGINT      NOT NULL,
    "DutyDate"       DATE        NOT NULL,
    "Shift"          VARCHAR(16)
);

CREATE INDEX IF NOT EXISTS "idx_staffduty_lookup"
    ON "Schedule_StaffDuty" ("TenantId", "WardId", "DutyDate", "DutyRole");
