# 排班业务闭环补齐报告

> 日期: 2026-06-08

## 已实现

| 功能 | 后端 | 前端 | 状态 |
|------|------|------|------|
| 三级确认 | `schedule_confirm_service.go` + 2 API | 待开发 UI | **已实现** |
| 冲突队列 | model + service + 3 API | API已通，UI待开发 | **已实现** |
| 请假/取消 | `CancelPatientShift` + API | `restClient.ts` | **已实现** |
| 缺席 | `MarkAbsent` + API | `restClient.ts` | **已实现** |
| 临时透析 | 常量+`SourceTemporary` | 显示"临"角标 | **部分**(缺专用API) |
| 设备停机 | `MachineOutage`模型 | 无CRUD API | **部分**(缺管理API) |
| 假日非透析日 | `Calendar`+引擎跳过 | 无假日管理UI | **已实现** |
| CRRT | `CrrtSession`模型+预检 | 无排入/列表API | **部分**(缺核心API) |
| 排班生成 | 引擎+`POST /generate` | `GenerateModal` | **已实现** |
| 模板管理 | service+API | 列表+编辑器+应用 | **已实现** |

## 待补齐（优先级顺序）

| 优先级 | 功能 | 工作量 | 说明 |
|--------|------|--------|------|
| P1 | 临时透析专用API | 2h | POST /schedule/temporary |
| P1 | 设备停机CRUD | 2h | GET/POST /schedule/outages |
| P1 | CRRT排入+列表 | 3h | POST + GET /schedule/crrt |
| P2 | 冲突队列前端UI | 3h | ConflictDrawer |
| P2 | 确认按钮前端 | 2h | 排班页确认按钮 |
| P2 | 假日管理UI | 2h | CalendarModal |
