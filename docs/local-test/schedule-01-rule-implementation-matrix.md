# 排班规则实现矩阵复核

> 复核日期: 2026-06-08 | 分支: fix/legacy-ui-restore

## 复核结果

| # | 规则 | 状态 | 后端关键文件 | 前端 |
|---|------|------|-------------|------|
| 1 | 模板独立存表 | **已实现** | `schedule_ext.go:103-140` + `schedule_template_service.go` | `ScheduleTemplateList.tsx` |
| 2 | 模板应用生成排班 | **已实现** | `schedule_template_service.go:259` | `ApplyTemplateModal.tsx` |
| 3 | 重复排班检测 | **已实现** | `patient_shift_service.go:456-510` | 通过后端 400 响应 |
| 4 | HDF两轮优先排位 | **已实现** | `schedule_engine/engine.go:107-217` | — |
| 5 | HD优先HD机，溢出HDF | **已实现** | `schedule_engine/placement.go:25-60` | — |
| 6 | 固定机位 | **已实现** | `models/schedule_ext.go:60-61` + `placement.go` | `ScheduleTemplateEditor.tsx` |
| 7 | 新患者多日同机 | **未实现** | 无代码 | 无代码 |
| 8 | 奇偶周锚点 | **已实现** | `schedule_generate_service.go:550` + `hdf.go` | — |
| 9 | 三级确认 | **已实现** | `schedule_confirm_service.go` + API | 无独立UI |
| 10 | 冲突队列 | **已实现** | `schedule_ext.go:144` + `conflict_service.go` | `restClient.ts:2158` |
| 11 | 临时透析 | **部分** | 常量有 `SourceTemporary=20`, 缺专用端点 | 显示"临"角标 |
| 12 | 请假/缺席 | **已实现** | `conflict_service.go:94-119` + API | `restClient.ts:2176` |
| 13 | 设备停机 | **部分** | 模型 `MachineOutage` 完整, 缺 CRUD API | — |
| 14 | 假日非透析日 | **已实现** | `Calendar` + 引擎跳过非透析日 | — |
| 15 | CRRT | **部分** | 模型 `CrrtSession` 完整, 缺插入/列出 API | — |
| 16 | 历史排班保护 | **已实现** | `patient_shift_service.go:514` + `ValidateShiftEdit` | `isDateLocked` |

**统计: 已实现 11/16, 部分 4/16, 未实现 1/16**

## P0 问题

**无 P0 问题发现**。没有重复排班、跨租户数据、致命 500 或半套数据问题。
