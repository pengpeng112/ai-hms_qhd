-- 排班模块索引优化 + 唯一约束
-- 在执行 CREATE INDEX 前数据库应有实际数据，确保索引有效
-- 生产环境禁止直接执行，先评估执行计划

-- ===== 组合索引（提升排班查询性能） =====

-- 周视图常用: TenantId + TreatmentTime + ShiftId
CREATE INDEX IF NOT EXISTS idx_ps_tenant_treatment_shift
  ON "Schedule_PatientShift" ("TenantId", "TreatmentTime", "ShiftId");

-- 患者排班历史: TenantId + PatientId + TreatmentTime
CREATE INDEX IF NOT EXISTS idx_ps_tenant_patient_treatment
  ON "Schedule_PatientShift" ("TenantId", "PatientId", "TreatmentTime");

-- 床位占用检测: TenantId + BedId + TreatmentTime + ShiftId
CREATE INDEX IF NOT EXISTS idx_ps_tenant_bed_treatment_shift
  ON "Schedule_PatientShift" ("TenantId", "BedId", "TreatmentTime", "ShiftId");

-- 排班扩展通过外键关联
CREATE INDEX IF NOT EXISTS idx_pse_tenant_shiftid
  ON "Schedule_PatientShiftExt" ("TenantId", "PatientShiftId");

-- 患者骨架查询
CREATE INDEX IF NOT EXISTS idx_spp_tenant_patientid
  ON "Schedule_PatientProfile" ("TenantId", "PatientId");

-- 床位机器扩展查询
CREATE INDEX IF NOT EXISTS idx_bme_tenant_bedid
  ON "Schedule_BedMachineExt" ("TenantId", "BedId");

-- 模板项查询
CREATE INDEX IF NOT EXISTS idx_sti_tenant_template
  ON "Schedule_ScheduleTemplateItem" ("TenantId", "TemplateId");

-- ===== 唯一约束（防并发重复排班） =====
-- 注意: 状态码 70=已取消, 80=缺席 是老库遗留码
-- 如实际代码使用新状态值，请替换 NOT IN 子句

-- 同一患者+同日+同班次，不可重复有效排班
-- (取消/缺席不占有效位)
CREATE UNIQUE INDEX IF NOT EXISTS uq_ps_active_patient_date_shift
  ON "Schedule_PatientShift" ("TenantId", "PatientId", "TreatmentTime", "ShiftId")
  WHERE "Status" NOT IN (40, 50, 60);  -- 排除已取消/用户取消/转出

-- 同一床位+同日+同班次，不可重复有效占用
-- (取消/缺席不占有效位，且BedId非空)
CREATE UNIQUE INDEX IF NOT EXISTS uq_ps_active_bed_date_shift
  ON "Schedule_PatientShift" ("TenantId", "BedId", "TreatmentTime", "ShiftId")
  WHERE "Status" NOT IN (40, 50, 60) AND "BedId" IS NOT NULL;
