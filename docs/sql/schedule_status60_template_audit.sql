-- Status=60 模板审计查询
-- 用途：只读审计 Schedule_PatientShift.Status=60 的记录，供人工区分"转出人员"和"模板伪记录"。
-- 注意：本脚本只做 SELECT，不执行 UPDATE/DELETE/INSERT。
-- 执行后人工标记 IsTemplateCandidate，再决定哪些迁移到独立模板表。

SELECT
    "TenantId",
    "Id"              AS "PatientShiftId",
    "PatientId",
    "WardId",
    "BedId",
    "ShiftId",
    "PatientPlanId",
    "ShiftTiming",
    "TreatmentTime",
    "Status"          AS "LegacyStatus",
    "CreatorId",
    "CreateTime",
    "LastModifyTime"
FROM "Schedule_PatientShift"
WHERE "Status" = 60
ORDER BY "TenantId", "WardId", "PatientId", "TreatmentTime";

-- 人工审核提示：
-- IsTemplateCandidate=true 的判断依据（需人工逐条确认）：
--   1. 该记录有 PatientPlanId 且 ShiftTiming 为长期（20）→ 更像模板骨架。
--   2. 该记录没有 PatientPlanId 或 ShiftTiming 为空/临时（10）→ 可能是真正的转出人员。
--   3. 同一 Ward 中有多条同日期骨架 → 更像模板。
--   4. TreatmentTime 是非常久远或非常近期的日期 → 需看业务上下文。
-- 禁止直接按 Status=60 全量迁移。
