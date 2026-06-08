# AI-HMS 血透系统排班模块专项复核与整改计划

适用仓库：pengpeng112/ai-hms_qhd  
适用分支：fix/legacy-ui-restore  
目标：对照排班规则、老系统逻辑和当前代码，复核排班管理一致性、卡顿原因、未完成功能，并按阶段整改。

---

## 1. 总体判断

当前版本已具备基础排班能力：病区、床位、班次 CRUD，排班扩展表，患者排班骨架，模板保存，模板应用，重复排班检测，以及与透析执行闭环打通。

但当前版本仍更接近“老库兼容 + 模板应用 + 基础排班管理”，尚未完全达到新版规则要求的“智能排班算法子程序”。后续整改应先解决卡顿和数据一致性，再补齐三级确认、冲突队列、请假缺席、设备停机、假日、CRRT，最后实现 HDF 两轮优先、固定机位、连片相邻等智能分配算法。

---

## 2. 整改总原则

1. 禁止连接生产库直接测试。
2. 禁止对生产库执行未经审核的 DDL/DML。
3. 禁止一次性重写全部排班逻辑。
4. 禁止为了解决冲突而自动强排患者。
5. 所有冲突必须遵循“主方案 → 备选方案 → 报警人工裁决”。
6. 当前阶段优先保持老库兼容，不强行切换到全新 Schedule_Machine 模型。
7. 所有新增功能必须有接口测试、数据库验证、前端验证和回归报告。

---

## 3. 阶段一：规则实现矩阵复核

### 目标

确认当前代码哪些规则已实现、哪些只是建表、哪些完全未实现。

### 检查文件

后端重点检查：

```text
ai-hms-backend/internal/services/schedule_week_service.go
ai-hms-backend/internal/services/schedule_template_service.go
ai-hms-backend/internal/services/schedule_board_service.go
ai-hms-backend/internal/services/schedule_config_service.go
ai-hms-backend/internal/api/v1/schedule_*.go
ai-hms-backend/internal/models/schedule.go
ai-hms-backend/internal/models/schedule_ext.go
docs/sql/schedule_extension_tables.sql
```

前端重点检查：

```text
ai-hms-frontend/src/pages/Schedule.tsx
ai-hms-frontend/src/pages/ScheduleTemplateList.tsx
ai-hms-frontend/src/pages/ScheduleTemplateEditor.tsx
ai-hms-frontend/src/services/schedule*.ts
ai-hms-frontend/src/services/restClient.ts
```

### 输出矩阵

| 规则 | 当前实现文件/函数 | 状态 | 说明 |
|---|---|---|---|
| 模板独立存表 |  | 已实现/部分/未实现 |  |
| 模板应用生成排班 |  | 已实现/部分/未实现 |  |
| 重复排班检测 |  | 已实现/部分/未实现 |  |
| HDF 两轮优先排位 |  | 已实现/部分/未实现 |  |
| HD 优先 HD 机，溢出 HDF |  | 已实现/部分/未实现 |  |
| 固定机位 |  | 已实现/部分/未实现 |  |
| 新患者多日同机 |  | 已实现/部分/未实现 |  |
| 奇偶周锚点 |  | 已实现/部分/未实现 |  |
| 三级确认 |  | 已实现/部分/未实现 |  |
| 冲突队列 |  | 已实现/部分/未实现 |  |
| 临时透析 |  | 已实现/部分/未实现 |  |
| 请假/缺席 |  | 已实现/部分/未实现 |  |
| 设备停机 |  | 已实现/部分/未实现 |  |
| 假日非透析日 |  | 已实现/部分/未实现 |  |
| CRRT |  | 已实现/部分/未实现 |  |
| 历史排班保护 |  | 已实现/部分/未实现 |  |

### 阶段产出

```text
docs/local-test/schedule-01-rule-implementation-matrix.md
```

---

## 4. 阶段二：排班卡顿性能定位

### 目标

定位排班管理不流畅原因，区分是 SQL 慢、接口返回过大、前端渲染重，还是拖拽/保存逻辑问题。

### 后端接口耗时记录

记录以下接口的耗时、返回大小、重复请求次数：

```text
GET /api/v1/schedule/week
GET /api/v1/schedule/rules/board
GET /api/v1/schedule/config/wards
GET /api/v1/schedule/config/machines
GET /api/v1/schedule/templates
POST /api/v1/schedule/templates/apply
PUT /api/v1/schedule/patient-shifts/{id}
```

### 慢 SQL 定位

重点检查：

```text
Schedule_PatientShift
Schedule_PatientShiftExt
Schedule_Ward
Schedule_Bed
Schedule_Shift
Schedule_WardExt
Schedule_BedMachineExt
Schedule_PatientProfile
Schedule_ScheduleTemplate
Schedule_ScheduleTemplateItem
Schedule_TenantSetting
```

必须检查是否存在：

1. `DATE("TreatmentTime")` 导致索引失效。
2. 缺少 `TenantId` 条件。
3. 未限制当前周起止时间。
4. 一次扫描历史全量排班。
5. N+1 查询。
6. 缺少组合索引。

对慢 SQL 执行：

```sql
EXPLAIN ANALYZE
-- 粘贴实际慢 SQL
;
```

### 前端定位

使用 DevTools 和 React Profiler 检查：

1. 切换周是否整页重渲染。
2. 拖拽时是否频繁 setState。
3. 保存一个排班是否刷新整周。
4. 是否一次性渲染全部病区、床位、日期、班次。
5. Tooltip/Popover 是否大量挂载。
6. 搜索是否缺少 debounce。
7. 单元格组件是否缺少 memo。

### 阶段产出

```text
docs/local-test/schedule-02-performance-diagnosis.md
```

---

## 5. 阶段三：后端性能整改

### 5.1 日期查询改造

将：

```sql
DATE("TreatmentTime") = DATE(?)
```

改为：

```sql
"TreatmentTime" >= ?
AND "TreatmentTime" < ?
```

周查询统一改为：

```sql
"TreatmentTime" >= :weekStart
AND "TreatmentTime" <  :weekEndPlusOne
```

### 5.2 索引优化

根据实际 SQL 验证后，在测试库考虑添加：

```sql
CREATE INDEX IF NOT EXISTS idx_schedule_patientshift_tenant_time_shift
ON "Schedule_PatientShift" ("TenantId", "TreatmentTime", "ShiftId");

CREATE INDEX IF NOT EXISTS idx_schedule_patientshift_tenant_patient_time
ON "Schedule_PatientShift" ("TenantId", "PatientId", "TreatmentTime");

CREATE INDEX IF NOT EXISTS idx_schedule_patientshift_tenant_bed_time_shift
ON "Schedule_PatientShift" ("TenantId", "BedId", "TreatmentTime", "ShiftId");

CREATE INDEX IF NOT EXISTS idx_schedule_patientshiftext_tenant_shift
ON "Schedule_PatientShiftExt" ("TenantId", "PatientShiftId");

CREATE INDEX IF NOT EXISTS idx_schedule_patientprofile_tenant_patient
ON "Schedule_PatientProfile" ("TenantId", "PatientId");

CREATE INDEX IF NOT EXISTS idx_schedule_bedmachine_tenant_bed
ON "Schedule_BedMachineExt" ("TenantId", "BedId");

CREATE INDEX IF NOT EXISTS idx_schedule_templateitem_tenant_template
ON "Schedule_ScheduleTemplateItem" ("TenantId", "TemplateId");
```

### 5.3 接口返回瘦身

周排班接口只返回：

```text
当前周
当前病区
有效排班
必要患者字段
必要床位字段
必要扩展字段
```

不要默认返回全部患者、全部历史排班、全部医嘱、全部治疗记录、全部模板历史。

### 5.4 配置缓存

可缓存：

```text
病区
床位
班次
床位机器能力
租户排班参数
模板头
```

缓存写操作后主动失效。

### 阶段产出

```text
docs/local-test/schedule-03-backend-performance-fix.md
```

---

## 6. 阶段四：前端流畅度整改

### 6.1 默认按病区加载

默认只加载当前病区或当前用户负责病区，不默认加载全部病区。

### 6.2 矩阵虚拟滚动

排班矩阵必须考虑虚拟化，推荐：

```text
@tanstack/react-virtual
react-window
react-virtualized
```

目标是只渲染可视区域床位行。

### 6.3 组件 memo 化

重点组件：

```text
ScheduleGrid
WardSection
BedRow
ScheduleCell
PatientShiftCard
```

要求使用：

```text
React.memo
useMemo
useCallback
```

避免一个单元格变化导致全表重渲染。

### 6.4 拖拽优化

拖动中只维护轻量 `dragState`，不要实时更新整周排班大对象。

Drop 后只更新源单元格和目标单元格，保存成功后局部替换。

### 6.5 保存后局部刷新

禁止修改一个排班后重新请求整周并重渲染全表。

改为后端返回 updated shift，前端只替换对应 cell，必要时后台静默刷新差异统计。

### 6.6 搜索 debounce

患者搜索使用 300ms debounce，输入过程中不要每个 keypress 重算全表。

### 阶段产出

```text
docs/local-test/schedule-04-frontend-smoothness-fix.md
```

---

## 7. 阶段五：并发和一致性补强

### 7.1 并发测试

必须测试：

1. 5 并发应用同一模板。
2. 5 并发排同一床位同一班次。
3. 10 并发调整同一患者同一天同班次。
4. 10 并发保存不同单元格。
5. 2 并发取消同一排班。
6. 2 并发确认同一排班日。

### 7.2 唯一约束

确认数据库是否存在有效保护：

```text
同一患者 + 同日 + 同班次，不可重复有效排班
同一床位/机器 + 同日 + 同班次，不可重复有效占用
取消/缺席不占有效位
```

候选唯一索引：

```sql
CREATE UNIQUE INDEX IF NOT EXISTS uq_patient_shift_active_patient
ON "Schedule_PatientShift" ("TenantId", "PatientId", "TreatmentTime", "ShiftId")
WHERE "Status" NOT IN (70, 80);

CREATE UNIQUE INDEX IF NOT EXISTS uq_patient_shift_active_bed
ON "Schedule_PatientShift" ("TenantId", "BedId", "TreatmentTime", "ShiftId")
WHERE "Status" NOT IN (70, 80)
  AND "BedId" IS NOT NULL;
```

注意：如果当前老库状态码不是 70/80，需要按真实状态码调整。

### 7.3 模板应用事务

`ApplyTemplate` 必须：

1. 在单事务内完成。
2. 任一患者排班失败，不产生半套数据。
3. 重复冲突返回明确错误。
4. 同一模板 + 同一目标周可考虑应用级锁。

### 阶段产出

```text
docs/local-test/schedule-05-concurrency-consistency.md
```

---

## 8. 阶段六：补齐排班业务闭环

### 8.1 三级确认

实现：

```text
第一次确认：护士长确认 2/4 周计划
第二次确认：主班/护士长确认次日全天，周一提前到周六
第三次确认：当日早上确认
```

字段：

```text
Confirm1At / Confirm1By
Confirm2At / Confirm2By
Confirm3At / Confirm3By
```

接口建议：

```text
POST /api/v1/schedule/confirm/level1
POST /api/v1/schedule/confirm/level2
POST /api/v1/schedule/confirm/level3
```

### 8.2 冲突队列

补齐 `Schedule_ConflictQueue` 闭环：

```text
GET /api/v1/schedule/conflicts
POST /api/v1/schedule/conflicts/{id}/resolve
POST /api/v1/schedule/conflicts/{id}/ignore
```

冲突类型：

```text
NO_MACHINE
HDF_NO_MACHINE
WARD_FULL
SLOT_SPILLED
NEW_PATIENT_UNPLACED
MACHINE_OUTAGE
HOLIDAY_REPLAN
PLAN_CHANGE
DUPLICATE_PATIENT
DUPLICATE_BED
```

### 8.3 请假/缺席

实现：

```text
提前请假 → 取消留痕，释放机位
当天缺席 → 缺席留痕，机位可借
系统不自动补透，只提示差异
```

### 8.4 临时透析

实现：

```text
当前班次当前区找空位
下一班次兜底
可借请假/缺席机位
不挤走规律患者
不进入模板
需医生医嘱 + 护士长/主班确认
```

### 8.5 设备停机

实现：

```text
临时停机：生成调整草稿，不改固定机位
长期停机：人工永久迁移
受影响患者进入冲突队列
```

### 8.6 假日非透析日

实现：

```text
周日默认非透析日
法定假日可设非透析日
假日值班模式只开放部分区/床/机
受影响排班进入冲突队列
```

### 8.7 CRRT

实现：

```text
C 区特殊排班
不走三班模板
记录起止时间
CRRT 机只做 CRRT
避免机器时间区间重叠
```

### 阶段产出

```text
docs/local-test/schedule-06-business-workflow-fix.md
```

---

## 9. 阶段七：智能排班算法增强

该阶段放在基础排班稳定后执行。

### 9.1 HDF 两轮算法

在每个 `区 × 日期 × 班次` 内：

1. 第一轮排 HDF 次。
2. 第二轮排 HD 次。
3. HD 优先 HD 机。
4. HD 机满后才使用空闲 HDF 机。
5. 无 HDF 机写冲突队列，不强排。

### 9.2 固定机位

实现：

```text
老患者优先使用上一次机器/床位
固定 HD 位和固定 HDF 位分开记忆
人工特批优先于自动记忆
```

### 9.3 新患者多日同机

实现：

```text
首选所有透析日都空闲的同一台机
找不到则逐次分配
仍失败则冲突队列
```

### 9.4 连片/相邻偏好

实现：

```text
同班次同方案患者尽量相邻
优先保护整台全空机
不做负载均衡式打散
```

### 9.5 HDF 奇偶周负载均衡

实现：

```text
使用 OddEvenWeekAnchorMonday
HdfWeekParity 只服务 HDF 错峰
新入组患者分配到 HDF 机负载较轻的奇/偶侧
```

### 阶段产出

```text
docs/local-test/schedule-07-smart-algorithm.md
```

---

## 10. 验收标准

### 10.1 功能验收

1. 模板保存、模板应用稳定。
2. 排班不重复。
3. 历史/已治疗排班受保护。
4. HDF 机型匹配正确。
5. 非透析日不生成常规排班。
6. 设备停机不再安排。
7. 冲突进入冲突队列。
8. 三级确认可追溯。
9. 透析执行仍能读取排班。

### 10.2 性能验收

建议目标：

```text
GET /api/v1/schedule/week < 1.5s
GET /api/v1/schedule/rules/board < 2s
切换周前端交互 < 1s
拖拽卡片无明显卡顿
保存单个排班 < 800ms
页面首屏渲染 < 2s
```

### 10.3 并发验收

1. 并发应用模板不重复。
2. 并发排同一床位只有一个成功。
3. 并发排同一患者同日同班只有一个成功。
4. 不出现半套模板应用结果。
5. 不出现主从表不一致。

---

## 11. 最终报告

整改完成后生成：

```text
docs/local-test/schedule-final-rectification-report.md
```

必须包含：

1. 规则实现矩阵。
2. 已实现功能。
3. 未实现功能。
4. 性能优化前后对比。
5. SQL 优化前后 EXPLAIN ANALYZE。
6. 前端渲染优化前后对比。
7. 并发测试结果。
8. 数据库 DDL 变更。
9. 是否建议扩大试运行。
10. 仍需人工确认的规则。

---

## 12. 推荐执行提示词

请从阶段一开始执行，不要跳阶段。每完成一个阶段，先生成阶段报告和问题清单，不要直接进入下一阶段。

若发现 P0 问题，例如重复排班、跨租户数据、核心接口 500、模板应用产生半套数据，请立即停止后续开发，先记录问题并修复。
