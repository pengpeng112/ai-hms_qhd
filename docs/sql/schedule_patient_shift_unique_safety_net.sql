-- Schedule_PatientShift 并发唯一安全网（仅供人工审查/执行）
-- 用途：在应用层冲突检查之外，为老主表补充数据库级并发保护。
-- 注意：本脚本不会由程序自动执行；执行前必须先在目标库人工审查重复数据和业务状态语义。
-- 老库 Schedule_PatientShift.Status 语义：10 草稿 / 20 已确认 / 30 用户确认 / 40 用户取消 / 50 排班取消 / 60 转出人员。
-- 本安全网只约束活跃状态，排除 40/50/60。

-- ============================================================
-- 1. 执行前预检：同患者同日同班重复
-- ============================================================
SELECT
    "TenantId",
    "PatientId",
    DATE("TreatmentTime") AS "ScheduleDate",
    "ShiftId",
    COUNT(*) AS "DupCount",
    ARRAY_AGG("Id" ORDER BY "Id") AS "Ids"
FROM "Schedule_PatientShift"
WHERE "Status" NOT IN (40, 50, 60)
GROUP BY "TenantId", "PatientId", DATE("TreatmentTime"), "ShiftId"
HAVING COUNT(*) > 1
ORDER BY "DupCount" DESC, "TenantId", "ScheduleDate", "ShiftId";

-- ============================================================
-- 2. 执行前预检：同床同日同班重复
-- ============================================================
SELECT
    "TenantId",
    "BedId",
    DATE("TreatmentTime") AS "ScheduleDate",
    "ShiftId",
    COUNT(*) AS "DupCount",
    ARRAY_AGG("Id" ORDER BY "Id") AS "Ids"
FROM "Schedule_PatientShift"
WHERE "BedId" IS NOT NULL
  AND "Status" NOT IN (40, 50, 60)
GROUP BY "TenantId", "BedId", DATE("TreatmentTime"), "ShiftId"
HAVING COUNT(*) > 1
ORDER BY "DupCount" DESC, "TenantId", "ScheduleDate", "ShiftId";

-- ============================================================
-- 3. 唯一索引：同患者同日同班只能有一条活跃排班
-- ============================================================
-- 生产执行建议使用 CONCURRENTLY，避免长时间阻塞写入。
-- PostgreSQL 要求 CREATE INDEX CONCURRENTLY 不能放在显式事务中执行。
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS "uidx_PatientShift_ActivePatientDayShift"
ON "Schedule_PatientShift" ("TenantId", "PatientId", DATE("TreatmentTime"), "ShiftId")
WHERE "Status" NOT IN (40, 50, 60);

-- ============================================================
-- 4. 唯一索引：同床同日同班只能有一条活跃排班
-- ============================================================
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS "uidx_PatientShift_ActiveBedDayShift"
ON "Schedule_PatientShift" ("TenantId", "BedId", DATE("TreatmentTime"), "ShiftId")
WHERE "BedId" IS NOT NULL
  AND "Status" NOT IN (40, 50, 60);

-- ============================================================
-- 5. 人工回滚语句（如需）
-- ============================================================
-- DROP INDEX CONCURRENTLY IF EXISTS "uidx_PatientShift_ActivePatientDayShift";
-- DROP INDEX CONCURRENTLY IF EXISTS "uidx_PatientShift_ActiveBedDayShift";
