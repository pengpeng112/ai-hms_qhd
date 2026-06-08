-- Schedule_ConflictQueue: 排班冲突/待处理队列
-- 规范 v1: 所有无法自动解决的约束进入此队，由人工裁决

CREATE TABLE IF NOT EXISTS "Schedule_ConflictQueue" (
  "Id" BIGINT PRIMARY KEY,
  "TenantId" BIGINT NOT NULL,
  "PatientId" BIGINT NULL,
  "ScheduleDate" TIMESTAMP NULL,
  "ShiftId" BIGINT NULL,
  "WardId" BIGINT NULL,
  "BedId" BIGINT NULL,
  "ConflictType" VARCHAR(64) NOT NULL,
  "Severity" SMALLINT NOT NULL DEFAULT 20,
  "Detail" TEXT NULL,
  "SuggestedShiftId" BIGINT NULL,
  "Status" SMALLINT NOT NULL DEFAULT 0,
  "HandledBy" BIGINT NULL,
  "HandledAt" TIMESTAMP NULL,
  "ResolutionNote" TEXT NULL,
  "CreatorId" BIGINT NULL,
  "CreateTime" TIMESTAMP NOT NULL DEFAULT NOW(),
  "LastModifyTime" TIMESTAMP NULL,
  "LastModifierId" BIGINT NULL
);
