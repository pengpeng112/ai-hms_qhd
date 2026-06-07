# 排班管理模块 — 业务逻辑说明文档

> 供代码审查用。配合 `schedule_source.txt`（完整源码）阅读。

---

## 1. 模块概览

排班管理是透析系统的核心模块，负责管理每周的患者透析排班计划。

**核心功能**：周视图查看、患者排班创建/修改、拖拽移床/互换、排班模板存取、班次设置。

**涉及老库表**（全部在 `Schedule_` 命名空间）：

| 表名 | 用途 | 读写 |
|------|------|------|
| `Schedule_Ward` | 病区 | R/W |
| `Schedule_Bed` | 床位 | R/W |
| `Schedule_Shift` | 班次（早班/中班/晚班） | R/W |
| `Schedule_PatientShift` | 患者排班记录（**核心表**） | R/W |
| `Schedule_PatientShift_Template` | 排班模板 | R |
| `Schedule_Timeslot` | 时间段 | R |
| `Schedule_BedEquipmentRel` | 床位-设备绑定 | R |

---

## 2. 数据模型（核心字段）

### 2.1 Schedule_PatientShift（患者排班）

```
Id              BIGINT PK    -- 自增
TenantId        BIGINT       -- 租户
PatientId       BIGINT       -- 患者ID → Register_PatientInfomation
WardId          BIGINT       -- 病区 → Schedule_Ward
BedId           BIGINT       -- 床位 → Schedule_Bed
ShiftId         BIGINT       -- 班次 → Schedule_Shift
ShiftDate       DATE         -- 排班日期
Status          INT          -- 状态: 0=待执行 10=草稿 20=已确认 30=已透析 40=已完成 50=已取消 60=转出人员
Content         TEXT         -- 备注/内容
IsDelete        BOOL         -- 逻辑删除
```

**状态枚举映射**（后端 `legacy_enum_maps.go`）：
```
Old Status 0  = New "pending"     (待执行)
Old Status 10 = New "draft"       (草稿)
Old Status 20 = New "confirmed"   (已确认)
Old Status 30 = New "dialyzed"    (已透析)
Old Status 40 = New "completed"   (已完成)
Old Status 50 = New "cancelled"   (已取消)
Old Status 60 = New "transferred" (转出人员)
```

**特殊约定**：`Status = 60` 的排班记录被视为**排班模板**（不作为实际排班显示，仅作模板来源）。

### 2.2 Schedule_Shift（班次）

```
Id              BIGINT PK
TenantId        BIGINT
Name            VARCHAR     -- 班次名称（早班/中班/晚班）
StartTime       TIME        -- 开始时间
EndTime         TIME        -- 结束时间
Sort            INT         -- 排序
IsDisabled      BOOL        -- 是否禁用
```

---

## 3. API 路由结构

所有排班路由在 `/api/v1/` 下，均需登录认证：

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/shifts` | 获取班次列表 |
| POST | `/shifts` | 创建班次 |
| PUT | `/shifts/:id` | 更新班次 |
| DELETE | `/shifts/:id` | 删除班次 |
| GET | `/patient-shifts` | 分页查询患者排班 |
| GET | `/patient-shifts/:id` | 获取单条排班 |
| POST | `/patient-shifts` | 创建患者排班 |
| PUT | `/patient-shifts/:id` | 更新患者排班 |
| DELETE | `/patient-shifts/:id` | 取消排班 |
| POST | `/patient-shifts/:id/move` | 移动排班（换床） |
| GET | `/schedule/week` | **周视图聚合数据**（核心接口） |
| POST | `/patient-shifts/swap` | 互换两条排班 |
| POST | `/patient-shifts/batch` | 批量保存排班 |
| GET | `/schedule/template` | 获取排班模板列表 |
| POST | `/schedule/template/apply` | 应用模板到当前周 |
| POST | `/schedule/template/save` | 保存当前周排班为模板 |

---

## 4. 核心业务逻辑

### 4.1 周视图聚合（`schedule_week_service.go`）

**接口**: `GET /schedule/week?date=2026-06-04&wardId=1`

**流程**:
1. 查询所有启用的 `Schedule_Ward`，按 `Sort` 排序
2. 对每个病区，联查 `Schedule_Bed`（过滤已禁用的）
3. 查询所有启用的 `Schedule_Shift`
4. 查询当周（Mon-Sun）的患者排班 `Schedule_PatientShift`
5. JOIN `Register_PatientInfomation` 获取患者姓名
6. JOIN `Plan_PatientPlan` 获取干体重和治疗模式
7. 组装成 `WeekWardItem → WeekBedItem → WeekShiftItem[]` 的三层嵌套结构
8. 返回待排班患者列表（已有排班的患者不在此列）

**数据结构**:
```
{
  wards: [{
    id, name, beds: [{
      id, name, shifts: [{
        shiftId, shiftName, patientShift: { ... } | null
      }]
    }]
  }],
  shifts: [{ id, name, startTime, endTime }],
  pendingPatients: [{ id, name, gender, age, ... }]
}
```

### 4.2 患者排班 CRUD（`patient_shift_service.go`）

**创建排班**:
1. 校验必填字段（PatientId、WardId、BedId、ShiftId、ShiftDate）
2. **冲突检测**：查询同一 `ShiftDate + ShiftId` 是否有同一患者的排班
3. 写入 `Schedule_PatientShift`，Status 默认 10（草稿）

**更新排班**:
1. 通过 Id 查找记录
2. 更新指定字段（病区/床位/患者/内容等）
3. `LastModifyTime` 设置当前时间

**取消排班**:
1. 设置 `IsDelete = true`（逻辑删除）
2. 同时设置 `LastModifyTime`

**换床（Move）**:
1. 更新 `BedId` 字段
2. 更新 `LastModifyTime`

**互换（Swap）**:
1. 开启数据库事务
2. 读取两条 `Schedule_PatientShift` 完整记录
3. 交换所有排班字段（WardId、BedId、ShiftId、PatientId、ShiftDate、Content、Status）
4. 更新 `LastModifyTime`
5. 提交事务

### 4.3 排班模板（`patient_shift_service.go`）

模板复用 `Schedule_PatientShift` 表，以 `Status = 60` 区分普通排班和模板。

**保存模板**:
1. 将指定日期的所有排班记录 `Status` 改为 60
2. 或由前端传入模板数据，批量写入 Status=60 的记录

**应用模板**:
1. 查找 `Status = 60` 的模板排班
2. 复制到指定日期的周（生成新的普通排班，Status=20）
3. 跳过已存在排班的床位/班次

**模板列表**:
1. 查询 `Status = 60` 的记录
2. 按日期分组，返回模板名称/日期/周期等信息

### 4.4 前端拖拽逻辑（`useScheduleDragDrop.ts`）

**拖放流程**:
1. 用户从待排班列表拖拽患者到周视图空格
2. 调用 `movePatientShift`:
   - 如果该空格已有排班（crossCell），调用 `PUT /patient-shifts/:id/move`
   - 如果是空格，创建新排班 `POST /patient-shifts`
3. 用户将患者A拖放到患者B的格子上：
   - 弹出确认框
   - 调用 `POST /patient-shifts/swap`

**日期锁定**：历史日期的排班不可拖拽修改。

### 4.5 前端弹窗逻辑（`useScheduleModals.ts`）

四种弹窗状态管理：
- 创建/编辑排班：选患者 → 确认提交
- 右键操作菜单：修改患者、移动床位、换床、取消排班
- 换床弹窗：选目标床位 → 确认
- 移动弹窗：选目标日期/班次 → 确认

---

## 5. 前端路由与权限

| 路由 | 页面 | 菜单键 | 权限要求 |
|------|------|--------|----------|
| `/schedule` | 排班管理主页 | `schedule` | `NURSE_SCHEDULER` 角色或有 `schedule` 权限 |
| `/schedule-templates` | 排班模板列表 | `schedule` | 同上 |
| `/schedule-templates/edit` | 编辑模板 | - | 同上 |
| `/shift-config` | 班次设置 | `schedule` | 同上 |

权限守卫在前端 `role.ts:getMenuKeyByPath()` 和 `AuthGuard.tsx` 中实现。

---

## 6. 已知风险点

1. **模板与普通排班共用同一表**：`Schedule_PatientShift` 同时存储排班和模板，仅靠 `Status=60` 区分。删除排班可能误删模板。
2. **互换事务**：Swap 在事务内读取两条记录再写入，存在幻读风险。
3. **待排班患者列表**：通过排除已有排班的患者来计算，在大数据量下性能依赖索引。
4. **租户硬编码**：排班查询全部使用 `LegacyTenantID`（环境变量配置的租户），非多租户灵活。
5. **前端直接调用 `restApi` 方法**：`ScheduleTemplateList.tsx` 和 `ApplyTemplateModal.tsx` 通过 `restApi` facade 调用，类型转换使用了 `as unknown as` 双重断言。
6. **状态字段类型不一致**：`Schedule_PatientShift.Status` 在数据库是 INT，前端使用 string 枚举值。

---

## 7. 文件清单

### 后端

| 文件 | 说明 |
|------|------|
| `internal/api/v1/schedule_handler.go` | 所有排班路由处理器（732行） |
| `internal/services/patient_shift_service.go` | 患者排班CRUD/冲突检测/互换/模板（720行） |
| `internal/services/shift_service.go` | 班次CRUD（187行） |
| `internal/services/schedule_week_service.go` | 周视图聚合查询（356行） |
| `internal/models/schedule.go` | GORM数据模型 |

### 前端

| 文件 | 说明 |
|------|------|
| `src/pages/Schedule.tsx` | 排班管理主页（821行） |
| `src/pages/ScheduleTemplateList.tsx` | 模板列表 |
| `src/pages/ScheduleTemplateEditor.tsx` | 模板编辑器 |
| `src/pages/ShiftConfig.tsx` | 班次设置 |
| `src/hooks/useScheduleModals.ts` | 弹窗状态管理 |
| `src/hooks/useScheduleDragDrop.ts` | 拖拽逻辑 |
| `src/components/schedule/ApplyTemplateModal.tsx` | 应用模板弹窗 |
| `src/services/restClient.ts` | API类型定义和方法 |
| `src/services/scheduleTemplate.ts` | 模板类型定义 |
