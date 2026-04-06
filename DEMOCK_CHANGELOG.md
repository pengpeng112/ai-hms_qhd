# DEMOCK Change Log

## [T-A1] 清理 constants.ts 死代码
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/constants.ts`
- 变更：删除全部 `MOCK_` 常量，仅保留 `DASHBOARD_CARDS`
- 编译验证：`npx tsc --noEmit` 通过

## [T-A2] 删除 mockHelpers.ts
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/utils/mockHelpers.ts`（删除）、`ai-hms-frontend/src/utils/index.ts`、`ai-hms-frontend/src/services/index.ts`
- 变更：删除 `mockHelpers` 文件及 re-export 链
- 编译验证：`npx tsc --noEmit` 通过

## [T-A3] DialysisProcessing 护士下拉动态化
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/pages/DialysisProcessing.tsx`
- 变更：新增护士列表状态，调用 `restApi.getUserList({ status: 'active' })`，所有护士下拉改为动态 `renderNurseOptions()`，无数据时显示占位项
- 编译验证：`npx tsc --noEmit` 通过

## [T-A4] DialysisProcessing 日期动态化
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/pages/DialysisProcessing.tsx`
- 变更：将全部 `2025-12-13` 相关默认值/展示改为当前时间函数
- 编译验证：`npx tsc --noEmit` 通过

## [T-A5] Monitoring 删除硬编码患者名
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/pages/Monitoring.tsx`
- 变更：`device.patientName` 空值改为 `--`；处方弹窗体重/超滤/血压参数改为设备数据或空串；耗材列表增加 `// TODO: 从 MaterialCatalog API 加载`
- 编译验证：`npx tsc --noEmit` 通过

## [T-B1] HistoryTab 接入真实治疗历史
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/pages/patient-detail/tabs/HistoryTab.tsx`
- 变更：移除 `mockHistoryData`，改为 `restApi.getTreatments`，新增加载态和“暂无治疗历史记录”空态
- 编译验证：`npx tsc --noEmit` 通过

## [T-B2] WardOverview processData 动态化
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/pages/WardOverview.tsx`
- 变更：`processData` 改为状态，按当天 `getTreatments` 统计状态分布生成饼图数据
- 编译验证：`npx tsc --noEmit` 通过

## [T-C1] 新建 ClinicalTask 数据模型
- 执行日期：2026-04-06
- 修改文件：`ai-hms-backend/internal/models/clinical_task.go`、`ai-hms-backend/internal/database/migrate.go`
- 变更：新增 `ClinicalTask` 模型并加入 AutoMigrate
- 编译验证：`go build ./...` 通过

## [T-C2] 新建 ClinicalTask Service + Handler
- 执行日期：2026-04-06
- 修改文件：`ai-hms-backend/internal/services/clinical_task_service.go`、`ai-hms-backend/internal/api/v1/clinical_task_handler.go`、`ai-hms-backend/cmd/server/main.go`
- 变更：新增 `GET /api/v1/clinical-tasks`、`PUT /api/v1/clinical-tasks/:id/status`
- 编译验证：`go build ./...` 通过

## [T-C3] MainLayout 替换 MOCK_TASKS
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/layouts/MainLayout.tsx`、`ai-hms-frontend/src/services/restClient.ts`
- 变更：删除 `MOCK_TASKS`，任务栏改为调用 `getClinicalTasks`；增加加载骨架与“暂无待处理任务”空态
- 编译验证：`npx tsc --noEmit` 通过

## [T-C4] Statistics 后端接口：quality
- 执行日期：2026-04-06
- 修改文件：`ai-hms-backend/internal/services/statistics_service.go`、`ai-hms-backend/internal/api/v1/statistics_handler.go`、`ai-hms-backend/cmd/server/main.go`
- 变更：新增 `GET /api/v1/statistics/quality`
- 编译验证：`go build ./...` 通过

## [T-C5] Statistics 后端接口：infection
- 执行日期：2026-04-06
- 修改文件：同 T-C4
- 变更：新增 `GET /api/v1/statistics/infection`
- 编译验证：`go build ./...` 通过

## [T-C6] Statistics 后端接口：vascular
- 执行日期：2026-04-06
- 修改文件：同 T-C4
- 变更：新增 `GET /api/v1/statistics/vascular`
- 编译验证：`go build ./...` 通过

## [T-C7] Statistics 后端接口：workload
- 执行日期：2026-04-06
- 修改文件：同 T-C4
- 变更：新增 `GET /api/v1/statistics/workload`
- 编译验证：`go build ./...` 通过

## [T-C8] Statistics.tsx 接入真实统计
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/pages/Statistics.tsx`、`ai-hms-frontend/src/services/restClient.ts`
- 变更：删除硬编码数组和固定卡片值，改为 4 组真实接口并行加载，增加 loading/error/empty 处理
- 编译验证：`npx tsc --noEmit` 通过

## [T-D1] 全量编译检查
- 执行日期：2026-04-06
- 命令：`npx tsc --noEmit`、`go build ./...`、`go vet ./...`
- 结果：全部通过

## [T-D2] Mock 残留扫描
- 执行日期：2026-04-06
- 命令：`rg -n "MOCK_" src/pages`、`rg -n "MOCK_" src/layouts`、`rg -n "武琼迪|李俊雅|高敬兰|..." src`、`rg -n "2025-12-13" src/pages`、`rg -n "mockHistoryData|mockHelpers|generateHistoryData" src`
- 结果：全部 0 命中

## [T-D3] 回归测试
- 执行日期：2026-04-06
- 结果：命令行环境未执行浏览器手工回归；已完成静态扫描与全量编译验证
