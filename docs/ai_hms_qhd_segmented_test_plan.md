# ai-hms_qhd 血液透析系统分段测试计划

> 适用项目：`pengpeng112/ai-hms_qhd`  
> 适用场景：本地 AI / Codex / Claude Code 基于测试库分段执行测试  
> 测试目标：验证当前修复后的程序是否具备进入科室试运行条件  
> 原则：只测测试库，不碰生产库；每段测试必须记录问题到本地 Markdown 文件。

---

## 0. 本地 AI 执行总原则

### 0.1 禁止事项

1. 禁止连接生产库。
2. 禁止对生产库执行 DDL/DML。
3. 禁止执行未审核的 AutoMigrate、DropTables。
4. 禁止把真实密码、token、连接串写入测试报告。
5. 禁止发现问题后直接大范围重构，必须先记录问题、定位证据、给出修复建议。

### 0.2 测试数据命名规则

所有测试数据统一使用前缀：

```text
TEST_AI_HMS_
```

例如：

```text
TEST_AI_HMS_患者A
TEST_AI_HMS_A区
TEST_AI_HMS_A01床
TEST_AI_HMS_低分子肝素
TEST_AI_HMS_透析器
```

### 0.3 本地测试记录文件

请在项目根目录创建测试记录目录：

```text
docs/local-test/
```

建议文件：

```text
docs/local-test/00-test-index.md
docs/local-test/01-env-startup.md
docs/local-test/02-auth-permission.md
docs/local-test/03-patient-plan-access.md
docs/local-test/04-order-test.md
docs/local-test/05-schedule-config.md
docs/local-test/06-schedule-template.md
docs/local-test/07-dialysis-execution.md
docs/local-test/08-inventory-device.md
docs/local-test/09-statistics-frontend.md
docs/local-test/10-security-concurrency.md
docs/local-test/issues.md
docs/local-test/final-report.md
```

所有发现的问题统一追加到：

```text
docs/local-test/issues.md
```

---

## 1. 问题记录格式

每发现一个问题，必须按下面格式追加：

```markdown
## ISSUE-编号：问题标题

- 级别：P0 / P1 / P2 / P3
- 模块：
- 测试阶段：
- 发现时间：
- 代码提交：
- 测试环境：
- 复现步骤：
  1.
  2.
  3.
- 预期结果：
- 实际结果：
- 接口请求：
- 接口响应：
- 涉及文件：
- 涉及数据库表：
- 数据库验证 SQL：
- 数据库验证结果：
- 日志摘要：
- 初步原因：
- 建议修复：
- 是否阻断试运行：是 / 否
- 当前状态：未修复 / 已修复待回归 / 已回归通过
```

问题级别定义：

```text
P0：会导致数据错乱、主从表不一致、越权、无法启动、事务不一致、核心流程无法完成。
P1：影响核心业务准确性或试运行质量，如排班冲突、审计缺失、状态错误。
P2：影响体验、提示、统计口径、部分边缘功能。
P3：文档、样式、轻微提示、优化建议。
```

---

## 2. 阶段一：环境、构建、启动测试

### 2.1 执行目标

确认当前代码、后端、前端、测试库、配置均可正常启动。

### 2.2 执行命令

```bash
git branch --show-current
git rev-parse HEAD
git status --short
```

后端：

```bash
cd ai-hms-backend
go version
go test ./...
go run ./cmd/server
```

前端：

```bash
cd ai-hms-frontend
node -v
npm -v
npm ci
npm run lint
npm run build
npm run dev
```

健康检查：

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
```

如果没有 `/healthz` 或 `/readyz`，记录为改进项，不一定阻断。

### 2.3 数据库检查

```sql
SELECT version();

SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

重点确认：

```text
Identity_Users
Identity_Roles
Identity_UserRoles
Authorization_Roles
Authorization_RoleUsers
Organ_Employee
Register_PatientInfomation
Plan_PatientPlan
Order_PatientOrder
Order_PatientDayOrder
Schedule_Ward
Schedule_Bed
Schedule_Shift
Schedule_PatientShift
Schedule_WardExt
Schedule_BedMachineExt
Schedule_PatientProfile
Schedule_PatientShiftExt
Schedule_ScheduleTemplate
Schedule_ScheduleTemplateItem
Schedule_ConflictQueue
Schedule_Calendar
Schedule_MachineOutage
```

### 2.4 输出文件

写入：

```text
docs/local-test/01-env-startup.md
```

必须记录：

```text
代码分支
提交号
Go 版本
Node/npm 版本
测试库名称
启动方式
go test 结果
npm build 结果
健康检查结果
缺失表清单
启动日志异常
```

---

## 3. 阶段二：认证、Token、权限测试

### 3.1 测试目标

验证登录、Token、角色权限、管理员接口保护。

### 3.2 测试用户

准备：

```text
TEST_AI_HMS_admin：管理员
TEST_AI_HMS_doctor：医生
TEST_AI_HMS_nurse：护士
TEST_AI_HMS_viewer：只读用户
```

### 3.3 登录接口

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "TEST_AI_HMS_admin",
  "password": "Test@123456"
}
```

验证：

```text
正确账号能登录
错误密码不能登录
禁用用户不能登录
返回 token
token 中包含 tenant_id
token 中 roles 正确
密码不出现在日志
```

### 3.4 权限接口测试

分别使用管理员、医生、护士、只读用户访问：

```text
GET /api/v1/me
用户管理接口
角色管理接口
权限管理接口
字典写接口
排班配置写接口
日志查看接口
HDIS 配置接口
医嘱新增接口
透析执行写接口
```

预期：

```text
未登录：401
无权限：403
有权限：200
普通用户不能直接 curl 管理员接口
```

### 3.5 输出文件

写入：

```text
docs/local-test/02-auth-permission.md
```

---

## 4. 阶段三：患者、治疗方案、血管通路测试

### 4.1 患者列表

测试：

```http
GET /api/v1/patients?page=1&pageSize=20
GET /api/v1/patients?name=TEST_AI_HMS
GET /api/v1/patients?onlyActive=true
GET /api/v1/patients?onlyTransferred=true
```

验证：

```text
分页正确
姓名模糊查询正确
在透/转出过滤正确
TenantId 不串数据
查询接口不产生写入
```

数据库核查：

```sql
SELECT *
FROM "Register_PatientInfomation"
WHERE "Name" LIKE 'TEST_AI_HMS_%'
ORDER BY "Id";
```

### 4.2 患者详情

测试：

```http
GET /api/v1/patients/{patientId}
GET /api/v1/patients/{patientId}/core
GET /api/v1/patients/{patientId}/basic-info
```

验证：

```text
基本信息正确
治疗方案正确
医嘱摘要正确
排班信息正确
不存在患者返回 404
非法 patientId 返回 400
```

### 4.3 治疗方案

测试：

```http
GET /api/v1/patients/{patientId}/treatment-plan
POST /api/v1/patients/{patientId}/treatment-plan
PUT /api/v1/patients/{patientId}/treatment-plan/{planId}
```

输入示例：

```json
{
  "dialysisMethod": "HDF",
  "oddWeekFrequency": 3,
  "evenWeekFrequency": 2,
  "duration": 4,
  "dryWeight": 65.5,
  "bloodFlow": 250,
  "anticoagulant": "TEST_AI_HMS_低分子肝素"
}
```

数据库核查：

```sql
SELECT *
FROM "Plan_PatientPlan"
WHERE "TenantId" = 3
  AND "PatientId" = :patientId
ORDER BY "LastModifyTime" DESC;
```

验证：

```text
新增方案落库
修改方案落库
停用方案后待排班不再出现
一个患者多个启用方案是否符合业务规则
```

### 4.4 血管通路

测试：

```http
GET /api/v1/patients/{patientId}/vascular-access
POST /api/v1/patients/{patientId}/vascular-access
PUT /api/v1/patients/{patientId}/vascular-access/{id}
```

输入示例：

```json
{
  "accessType": "AVF",
  "site": "左前臂",
  "buildDate": "2025-01-01",
  "status": "正常",
  "note": "TEST_AI_HMS_血管通路测试"
}
```

验证：

```text
新增可查
修改可查
不能跨患者修改
历史数据不被误删
```

### 4.5 输出文件

写入：

```text
docs/local-test/03-patient-plan-access.md
```

---

## 5. 阶段四：医嘱模块测试

### 5.1 目标

医嘱模块是 P0 重点。必须验证主医嘱、日医嘱、事务、状态、停用、并发。

### 5.2 医嘱列表

```http
GET /api/v1/patients/{patientId}/orders
GET /api/v1/patients/{patientId}/orders?type=longTerm
GET /api/v1/patients/{patientId}/orders?type=temporary
GET /api/v1/patients/{patientId}/orders?includeExpired=true
```

验证：

```text
长期/临时分类正确
停用默认不显示
includeExpired=true 可显示历史
状态映射正确
TenantId 不串
```

### 5.3 新增长期医嘱

```http
POST /api/v1/patients/{patientId}/orders
```

输入：

```json
{
  "type": "longTerm",
  "category": "药品",
  "name": "TEST_AI_HMS_低分子肝素",
  "content": "TEST_AI_HMS_低分子肝素 4000IU",
  "dose": "4000",
  "unit": "IU",
  "route": "静脉注射",
  "timing": "透析中",
  "execTiming": "during",
  "startTime": "2026-06-01",
  "notes": "TEST_AI_HMS_长期医嘱新增测试"
}
```

数据库核查：

```sql
SELECT *
FROM "Order_PatientOrder"
WHERE "Content" LIKE 'TEST_AI_HMS_%'
ORDER BY "CreateTime" DESC;

SELECT *
FROM "Order_PatientDayOrder"
WHERE "Content" LIKE 'TEST_AI_HMS_%'
ORDER BY "CreateTime" DESC;
```

验证：

```text
Order_PatientOrder 新增
Order_PatientDayOrder 新增或符合设计
TenantId/PatientId/PatientOrderId 一致
Classification/Content/Dosage/UseOpportunity/UseMethod/UseWay/Note 正确
CreatorId/OperatorId 正确
```

### 5.4 新增临时医嘱

输入：

```json
{
  "type": "temporary",
  "category": "处置",
  "name": "TEST_AI_HMS_临时补液",
  "content": "TEST_AI_HMS_临时补液 100ml",
  "dose": "100",
  "unit": "ml",
  "route": "静脉滴注",
  "timing": "透析后",
  "execTiming": "after",
  "startTime": "2026-06-01",
  "endTime": "2026-06-01",
  "notes": "TEST_AI_HMS_临时医嘱新增测试"
}
```

验证：

```text
类型映射正确
EndTime 正确
到期状态正确
不影响长期医嘱
```

### 5.5 修改医嘱

```http
PUT /api/v1/patients/{patientId}/orders/{orderId}
```

输入：

```json
{
  "content": "TEST_AI_HMS_低分子肝素 5000IU",
  "dose": "5000",
  "route": "静脉注射",
  "timing": "透析前",
  "notes": "TEST_AI_HMS_医嘱修改测试"
}
```

验证：

```text
更新老库真实字段
不是更新新系统小写字段
只更新目标患者目标医嘱
LastModifyTime 更新
```

### 5.6 停用医嘱

```http
POST /api/v1/patients/{patientId}/orders/{orderId}/stop
```

输入：

```json
{
  "stopReason": "TEST_AI_HMS_停用测试",
  "stopDate": "2026-06-02"
}
```

验证：

```text
主医嘱停用
EndTime 写入
日医嘱状态同步或符合设计
默认列表不显示
includeExpired=true 可显示
```

### 5.7 事务回滚

必须模拟：

```text
主医嘱写入成功后，日医嘱写入失败。
```

期望：

```text
主医嘱回滚
日医嘱无异常残留
接口返回错误
日志有失败原因
```

### 5.8 并发测试

模拟 10 个并发新增医嘱。

验证：

```text
无 ID 冲突
无主从不一致
无半条数据
符合是否允许重复医嘱的业务规则
```

### 5.9 输出文件

写入：

```text
docs/local-test/04-order-test.md
```

---

## 6. 阶段五：排班基础配置测试

### 6.1 病区

```http
GET /api/v1/wards
POST /api/v1/wards
PUT /api/v1/wards/{id}
DELETE /api/v1/wards/{id}
```

输入：

```json
{
  "name": "TEST_AI_HMS_A区",
  "patientType": "普通",
  "infectionType": "无",
  "sort": 1,
  "note": "TEST_AI_HMS_病区测试"
}
```

数据库：

```sql
SELECT *
FROM "Schedule_Ward"
WHERE "Name" LIKE 'TEST_AI_HMS_%';
```

### 6.2 床位

```http
GET /api/v1/beds
POST /api/v1/beds
PUT /api/v1/beds/{id}
DELETE /api/v1/beds/{id}
```

输入：

```json
{
  "name": "TEST_AI_HMS_A01",
  "wardId": 10001,
  "sort": 1,
  "note": "TEST_AI_HMS_床位测试"
}
```

数据库：

```sql
SELECT *
FROM "Schedule_Bed"
WHERE "Name" LIKE 'TEST_AI_HMS_%';
```

### 6.3 班次

```http
GET /api/v1/shifts
POST /api/v1/shifts
PUT /api/v1/shifts/{id}
DELETE /api/v1/shifts/{id}
```

输入：

```json
{
  "name": "TEST_AI_HMS_早班",
  "startTime": "08:00",
  "endTime": "12:00",
  "sort": 1
}
```

数据库：

```sql
SELECT *
FROM "Schedule_Shift"
WHERE "Name" LIKE 'TEST_AI_HMS_%';
```

### 6.4 验证点

```text
非管理员不能写
已被排班引用的数据不能物理删除
禁用后不能参与新排班
修改后前端列表刷新
```

### 6.5 输出文件

写入：

```text
docs/local-test/05-schedule-config.md
```

---

## 7. 阶段六：排班扩展、模板、周排班测试

### 7.1 病区扩展

```http
PUT /api/v1/schedule/config/wards
```

输入：

```json
{
  "wardId": 10001,
  "zoneType": "A",
  "isSubZone": false,
  "note": "TEST_AI_HMS_病区扩展"
}
```

数据库：

```sql
SELECT *
FROM "Schedule_WardExt"
WHERE "WardId" = 10001;
```

### 7.2 床位机器扩展

```http
PUT /api/v1/schedule/config/machines
```

输入：

```json
{
  "bedId": 10001,
  "machineCode": "TEST_AI_HMS_M001",
  "machineType": "HDF",
  "supportedModes": "HD,HDF,HF",
  "positionIndex": 1,
  "isDisabled": false,
  "note": "TEST_AI_HMS_机器扩展"
}
```

验证：

```text
HD 只支持 HD
HDF 支持 HD/HDF
CRRT 规则符合设计
禁用机器不参与排班
```

### 7.3 患者排班骨架

```http
PUT /api/v1/schedule/config/patient-profiles
```

输入：

```json
{
  "patientId": 10001,
  "zoneTag": "A",
  "homeWardId": 10001,
  "freqPattern": 30,
  "shiftId": 10001,
  "defaultMode": "HDF",
  "hdfEnabled": true,
  "hdfWeekday": 3,
  "hdfWeekParity": 1,
  "fixedHdBedId": 10001,
  "fixedHdfBedId": 10002,
  "isAdmissionRejected": false,
  "effectiveFrom": "2026-06-01"
}
```

数据库：

```sql
SELECT *
FROM "Schedule_PatientProfile"
WHERE "PatientId" = 10001;
```

### 7.4 奇偶周设置

```http
PUT /api/v1/schedule/config/settings
```

输入：

```json
{
  "settingKey": "OddEvenWeekAnchorMonday",
  "settingValue": "2026-06-01",
  "settingType": "date"
}
```

验证：

```text
必须是周一
周排班使用该配置
HDF 单双周使用该配置
不是直接使用 ISO 周
```

### 7.5 周排班查询

```http
GET /api/v1/schedule/week?startDate=2026-06-01&endDate=2026-06-07
```

验证：

```text
病区、床位、班次正确
已排班正确
待排班次数正确
取消/转出不计入
```

### 7.6 模板保存

```http
POST /api/v1/schedule/templates
```

输入：

```json
{
  "name": "TEST_AI_HMS_A区模板",
  "scope": "A",
  "wardId": 10001,
  "items": [
    {
      "patientId": 10001,
      "zoneTag": "A",
      "wardId": 10001,
      "shiftId": 10001,
      "freqPattern": 30,
      "fixedHdBedId": 10001,
      "fixedHdfBedId": 10002,
      "hdfEnabled": true,
      "hdfWeekday": 3,
      "hdfWeekParity": 1
    }
  ]
}
```

数据库：

```sql
SELECT *
FROM "Schedule_ScheduleTemplate"
WHERE "Name" LIKE 'TEST_AI_HMS_%';

SELECT *
FROM "Schedule_ScheduleTemplateItem"
WHERE "TenantId" = 3
ORDER BY "CreateTime" DESC;
```

### 7.7 模板应用

```http
POST /api/v1/schedule/templates/apply
```

输入：

```json
{
  "templateId": 10001,
  "targetDate": "2026-06-01",
  "wardId": 10001
}
```

验证：

```text
生成 Schedule_PatientShift
生成 Schedule_PatientShiftExt
不重复排班
非透析日不生成或明确提示
固定床冲突能识别
并发应用不重复
```

重复检查：

```sql
SELECT
  "TenantId",
  "PatientId",
  DATE("TreatmentTime") AS treatment_date,
  "ShiftId",
  COUNT(*) AS cnt
FROM "Schedule_PatientShift"
WHERE "Status" NOT IN (40, 50, 60)
GROUP BY "TenantId", "PatientId", DATE("TreatmentTime"), "ShiftId"
HAVING COUNT(*) > 1;

SELECT
  "TenantId",
  "BedId",
  DATE("TreatmentTime") AS treatment_date,
  "ShiftId",
  COUNT(*) AS cnt
FROM "Schedule_PatientShift"
WHERE "Status" NOT IN (40, 50, 60)
  AND "BedId" IS NOT NULL
GROUP BY "TenantId", "BedId", DATE("TreatmentTime"), "ShiftId"
HAVING COUNT(*) > 1;
```

### 7.8 输出文件

写入：

```text
docs/local-test/06-schedule-template.md
```

---

## 8. 阶段七：透析执行闭环测试

### 8.1 测试目标

从今日排班开始，完整模拟一次透析：

```text
今日排班
透前评估
首次核对
医嘱执行
透中监测
二次核对
透后评估
耗材扣减
设备消毒
健康宣教
透析汇总
完成透析
```

### 8.2 今日患者列表

验证：

```text
只显示今日有效排班
取消患者不显示或标记取消
感染患者有标识
床位、班次、模式正确
```

### 8.3 透前评估

输入示例：

```json
{
  "patientShiftId": 10001,
  "preWeight": 68.2,
  "dryWeight": 65.5,
  "temperature": 36.5,
  "pulse": 80,
  "systolicPressure": 140,
  "diastolicPressure": 85,
  "symptom": "TEST_AI_HMS_透前评估",
  "nurseId": 10001
}
```

验证：

```text
写入真实表
体重差计算正确
生命体征校验
可追溯 PatientShiftId
```

### 8.4 首次核对

输入：

```json
{
  "patientShiftId": 10001,
  "patientIdentityChecked": true,
  "dialysisModeChecked": true,
  "prescriptionChecked": true,
  "vascularAccessChecked": true,
  "anticoagulantChecked": true,
  "checkerId": 10001,
  "note": "TEST_AI_HMS_首次核对"
}
```

验证：

```text
核对记录写入
核对人正确
核对失败必须有原因
未核对是否允许进入下一步，按业务规则判断
```

### 8.5 医嘱执行

验证：

```text
当日医嘱加载
执行后 Order_PatientDayOrder 状态变化
执行人和执行时间正确
已停用医嘱不能执行
不能重复执行
```

### 8.6 透中监测

输入：

```json
{
  "patientShiftId": 10001,
  "recordTime": "2026-06-01 10:00:00",
  "bloodPressureHigh": 130,
  "bloodPressureLow": 80,
  "pulse": 76,
  "bloodFlow": 250,
  "ultrafiltrationVolume": 1200,
  "symptom": "TEST_AI_HMS_透中监测"
}
```

验证：

```text
允许多条记录
时间线正确
异常指标提示
修改/删除有权限
```

### 8.7 透后评估

输入：

```json
{
  "patientShiftId": 10001,
  "postWeight": 65.8,
  "temperature": 36.4,
  "pulse": 78,
  "systolicPressure": 125,
  "diastolicPressure": 78,
  "actualUltrafiltration": 2400,
  "complication": "无",
  "note": "TEST_AI_HMS_透后评估"
}
```

验证：

```text
透后评估写入
实际脱水量正确
未完成必要步骤是否允许透后评估
完成后核心记录是否锁定
```

### 8.8 耗材扣减

输入：

```json
{
  "patientShiftId": 10001,
  "items": [
    {
      "materialId": 10001,
      "quantity": 1,
      "unit": "支"
    }
  ],
  "note": "TEST_AI_HMS_耗材扣减"
}
```

验证：

```text
耗材记录新增
库存减少
库存流水新增
库存不足拒绝
部分失败整体回滚
并发扣减不出现负库存
```

### 8.9 设备消毒

输入：

```json
{
  "patientShiftId": 10001,
  "equipmentId": 10001,
  "disinfectUserId": 10001,
  "disinfectWay": "热化学消毒",
  "disinfectant": "TEST_AI_HMS_消毒液",
  "startTime": "2026-06-01 12:30:00",
  "endTime": "2026-06-01 13:00:00",
  "note": "TEST_AI_HMS_设备消毒"
}
```

验证：

```text
消毒记录新增
设备状态更新
结束时间不能早于开始时间
未消毒设备是否允许下一班使用，按规则判断
```

### 8.10 完成透析

验证：

```text
排班状态变为已完成
汇总可查
完成后不允许随意修改核心数据
今日列表状态更新
```

### 8.11 输出文件

写入：

```text
docs/local-test/07-dialysis-execution.md
```

---

## 9. 阶段八：库存、设备、字典、用户角色测试

### 9.1 库存耗材

测试：

```text
新增耗材
修改耗材
入库
出库
透析扣减
库存不足
库存预警
库存流水
并发扣减
```

重点验证：

```text
库存不能为负
流水完整
事务回滚
```

输出：

```text
docs/local-test/08-inventory-device.md
```

### 9.2 设备

测试：

```text
新增设备
绑定床位
状态改为正常/报警/离线/维护
设备停机
恢复使用
监控看板展示
停用设备不参与排班
```

### 9.3 字典

测试：

```text
新增字典
修改字典
禁用字典
已使用字典历史显示
非管理员不能写
```

### 9.4 用户角色

测试：

```text
新增用户
重置密码
禁用用户
新增角色
分配权限
删除角色前检查用户绑定
密码不能明文保存
```

---

## 10. 阶段九：统计、看板、前端 E2E

### 10.1 统计看板

测试：

```text
今日透析人数
在透患者数
设备使用率
排班完成率
医嘱执行率
HDF/HD/CRRT 统计
感染患者统计
```

验证：

```text
统计 SQL 与接口返回一致
取消排班不计入完成
转出患者不计入在透
TenantId 过滤正确
日期边界正确
```

输出：

```text
docs/local-test/09-statistics-frontend.md
```

### 10.2 前端 E2E 路径

手工或 Playwright 执行：

```text
登录
选择角色
患者列表
患者详情
新增医嘱
停用医嘱
排班页面
创建模板
应用模板
透析执行
透前评估
首次核对
医嘱执行
透中监测
透后评估
耗材扣减
设备消毒
完成透析
统计看板
```

失败时记录：

```text
截图
console error
network 请求
接口响应
后端日志
数据库记录
```

---

## 11. 阶段十：安全、并发、性能测试

### 11.1 安全测试

测试：

```text
SQL 注入
XSS
越权访问
无 token
错误 token
过期 token
普通用户访问管理员接口
日志是否泄露密码/token/连接串
CORS 限制
日志查看接口是否可读任意文件
```

输入示例：

```text
' OR 1=1 --
<script>alert(1)</script>
../../../etc/passwd
```

期望：

```text
非法输入 400
未认证 401
无权限 403
服务不 panic
日志不泄密
数据库不污染
```

### 11.2 并发测试

必须测试：

```text
10 个并发新增医嘱
10 个并发停用同一医嘱
5 个并发应用同一模板
5 个并发排同一床同一班
10 个并发扣同一耗材
2 个并发提交同一透后评估
```

记录：

```text
成功数
失败数
失败原因
最终数据库记录数
是否重复
是否负库存
是否主从不一致
```

### 11.3 性能测试

建议数据量：

```text
患者 1000
治疗方案 1000
排班 1 个月
医嘱每人 5 条
透中监测每次 5 条
库存流水 1 万条
```

目标：

```text
患者列表 < 1 秒
患者详情 < 1.5 秒
周排班 < 2 秒
今日透析列表 < 1 秒
医嘱列表 < 1 秒
统计看板 < 2 秒
```

重点检查：

```text
慢 SQL
N+1 查询
DATE(TreatmentTime) 影响索引
未按 TenantId 过滤
前端一次性加载全量
```

输出：

```text
docs/local-test/10-security-concurrency.md
```

---

## 12. 测试数据清理

测试完成后清理 `TEST_AI_HMS_%` 数据。

先统计：

```sql
SELECT COUNT(*) FROM "Register_PatientInfomation" WHERE "Name" LIKE 'TEST_AI_HMS_%';
SELECT COUNT(*) FROM "Order_PatientOrder" WHERE "Content" LIKE 'TEST_AI_HMS_%';
SELECT COUNT(*) FROM "Schedule_Ward" WHERE "Name" LIKE 'TEST_AI_HMS_%';
SELECT COUNT(*) FROM "Schedule_Bed" WHERE "Name" LIKE 'TEST_AI_HMS_%';
SELECT COUNT(*) FROM "Schedule_Shift" WHERE "Name" LIKE 'TEST_AI_HMS_%';
```

清理原则：

```text
优先测试库直接清理
涉及历史业务链路的数据，优先逻辑禁用
清理前记录数量
清理后再次统计
```

---

## 13. 最终报告

最终生成：

```text
docs/local-test/final-report.md
```

报告结构：

```markdown
# ai-hms_qhd 本地完整测试报告

## 1. 测试概况

- 测试时间：
- 执行人/AI：
- 代码分支：
- 提交号：
- 后端版本：
- 前端版本：
- 数据库版本：
- 测试库：
- 启动方式：

## 2. 测试结果总览

| 阶段 | 模块 | 结果 | P0 | P1 | P2 | P3 | 说明 |
|---|---|---|---:|---:|---:|---:|---|

## 3. 各模块测试摘要

## 4. 数据库增删改查验证摘要

## 5. 并发与事务测试摘要

## 6. 安全测试摘要

## 7. 性能测试摘要

## 8. 问题清单汇总

## 9. 阻断项

## 10. 是否建议进入科室试运行

结论：
- 建议 / 不建议
- 原因：
- 试运行前必须完成：
```

试运行准入标准：

```text
P0 = 0
核心流程通过
医嘱主从表一致
排班无重复冲突
耗材无负库存
权限无越权
日志无敏感信息
透析执行闭环可完成
```
