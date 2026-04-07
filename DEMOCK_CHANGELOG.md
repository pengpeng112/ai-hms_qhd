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

## [T-B3] DialysisProcessing 治疗记录提交
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/pages/DialysisProcessing.tsx`、`ai-hms-frontend/src/services/restClient.ts`
- 变更：补齐 `createTreatment/updateTreatment/updateTreatmentStatus`；流程中“进入监测”触发创建/更新治疗记录并置为进行中；“透后提交下一步”触发状态更新为已完成
- 备注：`notes` 字段临时写入 `// TODO: 补充治疗子表 API`，未使用任何 mock fallback
- 编译验证：`npx tsc --noEmit` 通过

## [P5-1] 阶段五：后端权限模型与接口基础
- 执行日期：2026-04-06
- 修改文件：`ai-hms-backend/internal/models/permission.go`、`ai-hms-backend/internal/services/permission_service.go`、`ai-hms-backend/internal/api/v1/permission_handler.go`、`ai-hms-backend/internal/database/migrate.go`、`ai-hms-backend/cmd/server/main.go`
- 变更：新增 `permissions` / `role_permissions` 模型与迁移；新增权限列表/保存接口和角色权限查询/覆盖接口；主路由注册 `RegisterPermissionRoutes`
- 编译验证：`go build ./...` 通过

## [P5-2] 阶段五：前端菜单权限改为后端驱动
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/services/role.ts`、`ai-hms-frontend/src/services/restClient.ts`、`ai-hms-frontend/src/layouts/Sidebar.tsx`
- 变更：删除 `FALLBACK_ROLE_USERS` mock 兜底；新增 `getRolePermissions/getPermissions/setRolePermissions` API；Sidebar 改为异步加载角色权限并仅展示后端返回授权菜单，无本地静态角色菜单 fallback
- 编译验证：`npx.cmd tsc --noEmit` 通过；`go build ./...` 通过

## [P5-3] 阶段五：任务栏权限过滤改为后端驱动
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/layouts/MainLayout.tsx`、`ai-hms-frontend/src/services/role.ts`
- 变更：删除 `MainLayout` 中按 `UserRole` 的硬编码任务可见性逻辑；改为加载角色权限码并按权限集合过滤任务类型（无权限即不展示）
- 编译验证：`npx.cmd tsc --noEmit` 通过；`go build ./...` 通过

## [P5-4] 阶段五：默认权限与角色授权初始化
- 执行日期：2026-04-06
- 修改文件：`ai-hms-backend/internal/services/permission_service.go`
- 变更：新增 `InitDefaultPermissions`（菜单级+任务级权限定义与角色基线授权，幂等）；新增 `ensureDefaultsInitialized`，在权限读取接口首次调用时自动初始化，避免前端权限化后出现空菜单
- 编译验证：`go build ./...` 通过；`npx.cmd tsc --noEmit` 通过

## [T-D1-R2] 全量编译复验
- 执行日期：2026-04-06
- 命令：`npx.cmd tsc --noEmit`、`go build ./...`、`go vet ./...`
- 结果：全部通过

## [T-D2-R2] Mock 残留复扫与补充清理
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/pages/Monitoring.tsx`
- 变更：移除 `generateMiniGraphData` 等随机模拟逻辑，监控曲线与生命体征改为无数据空态，不再使用 `Math.random` 生成假数据
- 扫描命令：`rg -n "MOCK_" ...`、`rg -n "武琼迪|李俊雅|高敬兰|..." ...`、`rg -n "2025-12-13|2024-05-04|2024-05-10" ...`、`rg -n "generateHistoryData|generateMiniGraphData|mockHistoryData|mockHelpers|return\s+MOCK_" ...`
- 结果：全部 0 命中；`src/utils/mockHelpers.ts` 为 `missing`

## [T-D3-R2] 回归验证复核（命令行可验证项）
- 执行日期：2026-04-06
- 校验点：护士下拉动态加载、治疗历史真实接口、病区 processData 动态化、统计页面真实接口、任务栏真实任务与空态、监控页 patientName 空态显示
- 结果：代码路径与关键调用复核通过；受当前环境限制，浏览器手工点击回归仍需在 UI 环境执行

## [T-D1-R3] 编译与构建链修复完成
- 执行日期：2026-04-06
- 修改文件：`ai-hms-frontend/src/services/restClient.ts`、`ai-hms-frontend/src/pages/DialysisProcessing.tsx`、`ai-hms-frontend/src/pages/Monitoring.tsx`、`ai-hms-frontend/src/pages/Statistics.tsx`、`ai-hms-frontend/src/pages/patient-detail/tabs/HistoryTab.tsx`
- 变更：修复历史遗留的注释粘连/字符串引号损坏/单位字段损坏等语法问题，恢复前端构建链
- 验证：`npx.cmd tsc -b` 通过；`npm.cmd run build` 通过；`go build ./...` 通过；`go vet ./...` 通过

## [T-D2-R3] Mock 残留再扫描
- 执行日期：2026-04-06
- 命令：`rg -n "MOCK_" ...`、`rg -n "武琼迪|李俊雅|高敬兰|..." ...`、`rg -n "2025-12-13|2024-05-04|2024-05-10" ...`、`rg -n "generateHistoryData|generateMiniGraphData|mockHistoryData|mockHelpers|return\s+MOCK_" ...`
- 结果：全部 0 命中

## [P5-5] 阶段五：任务处理操作级权限控制
- 执行日期：2026-04-07
- 修改文件：`ai-hms-backend/internal/services/permission_service.go`、`ai-hms-frontend/src/layouts/MainLayout.tsx`
- 变更：
  - 后端默认权限新增 `task.*.handle`（alert/prescription/order/assessment）并写入角色基线授权
  - 前端任务栏新增操作级权限校验：仅有 `task.*.handle` 才可点击处理；无权限显示“无处理权限”并禁用点击
- 编译验证：`npx.cmd tsc --noEmit` 通过；`go build ./...` 通过；`go vet ./...` 通过
- 残留扫描：`rg -n "MOCK_" ai-hms-frontend/src/pages ai-hms-frontend/src/layouts`、`rg -n "武琪迪|李俊雅|高敬兰" ai-hms-frontend/src`、`rg -n "2025-12-13" ai-hms-frontend/src/pages`、`rg -n "mockHistoryData|mockHelpers|generateHistoryData|generateMiniGraphData|return\\s+MOCK_" ai-hms-frontend/src` 均 0 命中

## [BUGFIX-2026-04-07] 历史改造后缺陷修复
- 执行日期：2026-04-07
- 修改文件：
  - `ai-hms-frontend/src/layouts/MainLayout.tsx`
  - `ai-hms-frontend/src/pages/patient-detail/tabs/HistoryTab.tsx`
  - `ai-hms-frontend/src/pages/DialysisProcessing.tsx`
  - `ai-hms-backend/internal/services/permission_service.go`
- 变更：
  - 修复任务栏路由：`ORDER/ASSESSMENT` 从 `/dialysis` 改为 `/dialysis-processing`
  - 增加患者 ID 数值安全校验，避免 `Number(...)` 为 `NaN` 时继续请求治疗相关 API
  - 后端权限默认初始化改为幂等增量补齐，支持已有环境自动补入新增权限（含 `task.*.handle`）
- 编译验证：`npx.cmd tsc --noEmit` 通过；`go build ./...` 通过；`go vet ./...` 通过

## [FIX-1+2] statistics 接口 tenant_id 隔离修复
- 执行日期：2026-04-07
- 修改文件：`ai-hms-backend/internal/services/statistics_service.go`、`ai-hms-backend/internal/api/v1/statistics_handler.go`
- 变更：4 个统计方法（`QualityByYear/InfectionByYear/VascularByYear/WorkloadByYearMonth`）增加 `tenantId` 参数；所有 DB 查询加入 `tenant_id = ?` 过滤；Handler 层从 `middleware.GetTenantID(c)` 提取并传入
- 编译验证：`go build ./...` 通过；`go vet ./...` 通过

## [FIX-3] Monitoring.tsx OrderListModal 硬编码医嘱清除
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/Monitoring.tsx`
- 变更：删除 `longOrders/tempOrders` 硬编码初始值（含 `2025-12-01` 日期、`王医生` 等），改为空数组并增加后续 API 对接 TODO 注释
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [FIX-4] permission_service.go 初始化并发安全修复
- 执行日期：2026-04-07
- 修改文件：`ai-hms-backend/internal/services/permission_service.go`
- 变更：引入包级 `sync.Once`，`ensureDefaultsInitialized()` 在进程生命周期内只执行一次
- 编译验证：`go build ./...` 通过；`go vet ./...` 通过

## [FIX-5] NURSE_SCHEDULER 角色常量化
- 执行日期：2026-04-07
- 修改文件：`ai-hms-backend/internal/models/user.go`、`ai-hms-backend/internal/services/permission_service.go`
- 变更：`user.go` 新增 `RoleNurseScheduler = "NURSE_SCHEDULER"`；`permission_service.go` 中角色键从硬编码字符串改为 `models.RoleNurseScheduler`
- 编译验证：`go build ./...` 通过；`go vet ./...` 通过

## [FIX-6] parseYear 年份上界校验
- 执行日期：2026-04-07
- 修改文件：`ai-hms-backend/internal/api/v1/statistics_handler.go`
- 变更：`parseYear()` 合法范围限定为 `[2000, currentYear+1]`，超出范围回退当前年份
- 编译验证：`go build ./...` 通过；`go vet ./...` 通过
