-- ============================================================
-- IDH 重标定 · 数据可用性盘点 (PostgreSQL / ai-hms 老库)
-- 目的:判断本中心历史数据是否够"现在就重标定",还是要继续攒。
-- 用法:把每段单独在生产库(只读)跑;先改 params 里的 TenantId(ai-hms 老库默认 3)。
-- 标签定义(同训练):某测点的"下一次测量"低血压 = 下次 SBP<100 或 SBP 跌>20 或 MAP 跌>10。
-- MAP = (SBP + 2*DBP)/3。
-- 注意:Q5 是逐血压点关联数 DMLog,大库可能较慢,可先只跑 Q1–Q4。
-- ============================================================


-- Q1. 总体盘点:DMLog 与血压记录的量级与时间跨度 ------------------
WITH params AS (SELECT 3::bigint AS tenant)
SELECT 'Device_DMLog' AS tbl,
       count(*)                              AS rows_total,
       count(DISTINCT "TreatmentId")         AS treatments,
       min("LogTime")::date                  AS first_day,
       max("LogTime")::date                  AS last_day
FROM "Device_DMLog", params WHERE "TenantId" = params.tenant
UNION ALL
SELECT 'Treatment_DuringSigns' AS tbl,
       count(*) FILTER (WHERE "SBP" IS NOT NULL AND "DBP" IS NOT NULL),
       count(DISTINCT "TreatmentId"),
       min("OperateTime")::date,
       max("OperateTime")::date
FROM "Treatment_DuringSigns", params WHERE "TenantId" = params.tenant;


-- Q2. 合格治疗(同时有 DMLog 且 有血压)= 重标定可用队列规模 --------
WITH params AS (SELECT 3::bigint AS tenant),
elig AS (
  SELECT t."Id" AS tid
  FROM "Treatment_Treatment" t, params p
  WHERE t."TenantId" = p.tenant
    AND EXISTS (SELECT 1 FROM "Device_DMLog" d
                WHERE d."TreatmentId" = t."Id" AND d."TenantId" = p.tenant)
    AND EXISTS (SELECT 1 FROM "Treatment_DuringSigns" s
                WHERE s."TreatmentId" = t."Id" AND s."TenantId" = p.tenant
                  AND s."SBP" IS NOT NULL AND s."DBP" IS NOT NULL)
)
SELECT count(*) AS eligible_treatments FROM elig;


-- Q3. 合格治疗的"密度"(每次治疗 DMLog 行数 / 血压测点数 的中位数与四分位) --
WITH params AS (SELECT 3::bigint AS tenant),
elig AS (
  SELECT t."Id" AS tid
  FROM "Treatment_Treatment" t, params p
  WHERE t."TenantId" = p.tenant
    AND EXISTS (SELECT 1 FROM "Device_DMLog" d WHERE d."TreatmentId"=t."Id" AND d."TenantId"=p.tenant)
    AND EXISTS (SELECT 1 FROM "Treatment_DuringSigns" s WHERE s."TreatmentId"=t."Id" AND s."TenantId"=p.tenant AND s."SBP" IS NOT NULL AND s."DBP" IS NOT NULL)
),
dlog AS (
  SELECT d."TreatmentId" tid, count(*) c
  FROM "Device_DMLog" d, params p
  WHERE d."TenantId"=p.tenant AND d."TreatmentId" IN (SELECT tid FROM elig)
  GROUP BY d."TreatmentId"
),
bpc AS (
  SELECT s."TreatmentId" tid, count(*) c
  FROM "Treatment_DuringSigns" s, params p
  WHERE s."TenantId"=p.tenant AND s."SBP" IS NOT NULL AND s."DBP" IS NOT NULL
    AND s."TreatmentId" IN (SELECT tid FROM elig)
  GROUP BY s."TreatmentId"
)
SELECT
  (SELECT percentile_cont(0.25) WITHIN GROUP (ORDER BY c) FROM dlog) AS dmlog_p25,
  (SELECT percentile_cont(0.5)  WITHIN GROUP (ORDER BY c) FROM dlog) AS dmlog_median,
  (SELECT percentile_cont(0.75) WITHIN GROUP (ORDER BY c) FROM dlog) AS dmlog_p75,
  (SELECT percentile_cont(0.25) WITHIN GROUP (ORDER BY c) FROM bpc)  AS bp_p25,
  (SELECT percentile_cont(0.5)  WITHIN GROUP (ORDER BY c) FROM bpc)  AS bp_median,
  (SELECT percentile_cont(0.75) WITHIN GROUP (ORDER BY c) FROM bpc)  AS bp_p75;


-- Q4. 可标注点数 + 本中心 IDH 事件率(核心指标) -------------------
-- label_points = 有"下一次测量"的血压点(能形成 t→t+1 对);idh_events = 其中触发低血压标签数。
WITH params AS (SELECT 3::bigint AS tenant),
elig AS (
  SELECT t."Id" AS tid
  FROM "Treatment_Treatment" t, params p
  WHERE t."TenantId" = p.tenant
    AND EXISTS (SELECT 1 FROM "Device_DMLog" d WHERE d."TreatmentId"=t."Id" AND d."TenantId"=p.tenant)
    AND EXISTS (SELECT 1 FROM "Treatment_DuringSigns" s WHERE s."TreatmentId"=t."Id" AND s."TenantId"=p.tenant AND s."SBP" IS NOT NULL AND s."DBP" IS NOT NULL)
),
bp AS (
  SELECT s."TreatmentId" tid, s."OperateTime" t, s."SBP" sbp, s."DBP" dbp
  FROM "Treatment_DuringSigns" s, params p
  WHERE s."TenantId"=p.tenant AND s."SBP" IS NOT NULL AND s."DBP" IS NOT NULL
    AND s."TreatmentId" IN (SELECT tid FROM elig)
),
paired AS (
  SELECT tid, t, sbp, dbp,
         LEAD(sbp) OVER w AS nsbp,
         LEAD(dbp) OVER w AS ndbp
  FROM bp
  WINDOW w AS (PARTITION BY tid ORDER BY t)
)
SELECT
  count(*) FILTER (WHERE nsbp IS NOT NULL) AS label_points,
  count(*) FILTER (WHERE nsbp IS NOT NULL AND
      (nsbp < 100 OR (sbp - nsbp) > 20 OR ((sbp + 2*dbp)/3 - (nsbp + 2*ndbp)/3) > 10)
  ) AS idh_events,
  round(100.0 * count(*) FILTER (WHERE nsbp IS NOT NULL AND
      (nsbp < 100 OR (sbp - nsbp) > 20 OR ((sbp + 2*dbp)/3 - (nsbp + 2*ndbp)/3) > 10)
  ) / nullif(count(*) FILTER (WHERE nsbp IS NOT NULL), 0), 1) AS idh_rate_pct
FROM paired;


-- Q5.（可选,可能慢）窗口充足度:每个血压点前 120 分钟内有多少 DMLog 行 ----
-- 训练用前 30 个 DMLog 时点;这里看有多少血压点能凑到 >=10 / >=30 个前置 DMLog。
WITH params AS (SELECT 3::bigint AS tenant),
elig AS (
  SELECT t."Id" AS tid
  FROM "Treatment_Treatment" t, params p
  WHERE t."TenantId" = p.tenant
    AND EXISTS (SELECT 1 FROM "Device_DMLog" d WHERE d."TreatmentId"=t."Id" AND d."TenantId"=p.tenant)
    AND EXISTS (SELECT 1 FROM "Treatment_DuringSigns" s WHERE s."TreatmentId"=t."Id" AND s."TenantId"=p.tenant AND s."SBP" IS NOT NULL AND s."DBP" IS NOT NULL)
),
bp AS (
  SELECT s."TreatmentId" tid, s."OperateTime" t
  FROM "Treatment_DuringSigns" s, params p
  WHERE s."TenantId"=p.tenant AND s."SBP" IS NOT NULL AND s."DBP" IS NOT NULL
    AND s."TreatmentId" IN (SELECT tid FROM elig)
)
SELECT
  count(*) AS bp_points,
  count(*) FILTER (WHERE w.cnt >= 10) AS with_ge10_dmlog,
  count(*) FILTER (WHERE w.cnt >= 30) AS with_ge30_dmlog,
  round(100.0 * count(*) FILTER (WHERE w.cnt >= 10) / nullif(count(*),0), 1) AS pct_ge10,
  round(100.0 * count(*) FILTER (WHERE w.cnt >= 30) / nullif(count(*),0), 1) AS pct_ge30
FROM bp
CROSS JOIN LATERAL (
  SELECT count(*) AS cnt
  FROM "Device_DMLog" d, params p
  WHERE d."TenantId" = p.tenant AND d."TreatmentId" = bp.tid
    AND d."LogTime" <= bp.t
    AND d."LogTime" >  bp.t - interval '120 minutes'
) w;


-- ============================================================
-- 怎么看结果(粗略经验门槛,越多越好):
--  · Q2 eligible_treatments  >= 200~300        → 队列规模够
--  · Q4 label_points         >= 2000           → 标注点够画稳 ROC/调阈值
--  · Q4 idh_rate_pct         约 5%~40%         → 类不平衡可控(太低需 SMOTE/重采样)
--  · Q5 pct_ge10             >= 60%~70%         → 多数测点能凑出特征窗(>=30 更好)
-- 都达标 → "现在就能做" IDH 重标定(离线评估+调阈值,开一轮 spec→执行)。
-- 明显不足 → 继续上线攒数据,过段时间再盘点。
-- ============================================================
