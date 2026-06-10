-- =============================================================================
-- 排班模块唯一索引 —— DBA 执行脚本
-- =============================================================================
-- 用途：为智能排班 v2 创建保证并发安全的唯一索引。
--      应用本身遵守"老库运行时严禁 DDL"红线，不再自动建索引（原 repo.EnsureIndexes
--      运行时调用已移除），改由本脚本经 DBA 审核后在老血透生产库执行；应用启动仅做
--      存在性校验（repo.VerifyIndexes），缺失时在 /schedule/health 告警。
--
-- 真值源：与 internal/smart_schedule/repo/repo.go 的 RequiredUniqueIndexes 保持一致。
--         修改索引定义时两处需同步。
--
-- 执行前提（务必按顺序）：
--   第 1 步：跑"重复数据探测"，确认两个目标索引均无冲突行（返回 0 行）。
--   第 2 步：若有冲突行，先由业务/DBA 清理或修正（取消/合并），不可强建。
--   第 3 步：再执行"创建索引"。CONCURRENTLY 不可置于事务块中，请逐条单独执行。
--
-- 环境：PostgreSQL（schema = public）
-- =============================================================================


-- -----------------------------------------------------------------------------
-- 第 1 步：重复数据探测（建索引前必须返回 0 行）
-- -----------------------------------------------------------------------------

-- 1a. uq_ps_patient_slot 冲突探测：
--     同一租户 + 病人 + 治疗时间 + 班次，存在多条有效（未取消70/未缺席80、ShiftId>0）排班。
SELECT "TenantId", "PatientId", "TreatmentTime", "ShiftId", COUNT(*) AS dup_count
FROM "Schedule_PatientShift"
WHERE "Status" NOT IN (70, 80)
  AND "ShiftId" > 0
GROUP BY "TenantId", "PatientId", "TreatmentTime", "ShiftId"
HAVING COUNT(*) > 1
ORDER BY dup_count DESC;

-- 1b. uq_ps_machine_slot 冲突探测：
--     同一租户 + 机位 + 治疗时间 + 班次，存在多条有效排班。
--     ⚠️ 注意：谓词沿用历史定义 "MachineId IS NOT NULL"。现行模型 MachineId 为
--        int64 NOT NULL DEFAULT 0，故未分配机位（MachineId=0）的多条排班也会被算作冲突。
--        若探测结果大量来自 MachineId=0，多半是"尚未分配机位"的正常草稿，
--        此时应优先考虑把索引谓词改为 "MachineId > 0"（见文末"可选语义优化"），
--        而非删除真实数据。是否调整由 DBA 与研发共同决定。
SELECT "TenantId", "MachineId", "TreatmentTime", "ShiftId", COUNT(*) AS dup_count
FROM "Schedule_PatientShift"
WHERE "Status" NOT IN (70, 80)
  AND "MachineId" IS NOT NULL
  AND "ShiftId" > 0
GROUP BY "TenantId", "MachineId", "TreatmentTime", "ShiftId"
HAVING COUNT(*) > 1
ORDER BY dup_count DESC;


-- -----------------------------------------------------------------------------
-- 第 2 步：创建唯一索引（确认上述探测均为 0 行后执行）
-- -----------------------------------------------------------------------------
-- 生产库推荐用 CONCURRENTLY 避免长写锁；CONCURRENTLY 不能在事务中执行，请逐条单独运行。
-- 若工具不支持 CONCURRENTLY，可去掉该关键字（会短暂持有写锁）。

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_ps_patient_slot
    ON "Schedule_PatientShift" ("TenantId", "PatientId", "TreatmentTime", "ShiftId")
    WHERE "Status" NOT IN (70, 80) AND "ShiftId" > 0;

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_ps_machine_slot
    ON "Schedule_PatientShift" ("TenantId", "MachineId", "TreatmentTime", "ShiftId")
    WHERE "Status" NOT IN (70, 80) AND "MachineId" IS NOT NULL AND "ShiftId" > 0;


-- -----------------------------------------------------------------------------
-- 第 3 步：验证（应各返回 1 行，valid = t）
-- -----------------------------------------------------------------------------
SELECT i.indexname, idx.indisvalid AS valid
FROM pg_indexes i
JOIN pg_class c ON c.relname = i.indexname
JOIN pg_index idx ON idx.indexrelid = c.oid
WHERE i.schemaname = 'public'
  AND i.indexname IN ('uq_ps_patient_slot', 'uq_ps_machine_slot');


-- =============================================================================
-- 回滚（仅在需要撤销时执行）
-- =============================================================================
-- DROP INDEX CONCURRENTLY IF EXISTS uq_ps_patient_slot;
-- DROP INDEX CONCURRENTLY IF EXISTS uq_ps_machine_slot;


-- =============================================================================
-- 可选语义优化（需研发确认后再用，会改变索引语义，务必同步 repo.go 的 RequiredUniqueIndexes）
-- =============================================================================
-- 将机位唯一索引谓词由 "MachineId IS NOT NULL" 改为 "MachineId > 0"，
-- 使未分配机位（MachineId=0）的草稿排班不参与机位唯一性约束：
--
-- DROP INDEX CONCURRENTLY IF EXISTS uq_ps_machine_slot;
-- CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_ps_machine_slot
--     ON "Schedule_PatientShift" ("TenantId", "MachineId", "TreatmentTime", "ShiftId")
--     WHERE "Status" NOT IN (70, 80) AND "MachineId" > 0 AND "ShiftId" > 0;
