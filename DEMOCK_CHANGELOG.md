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

## [BUGFIX-2026-04-07-UI-ENCODING] 前端可见乱码清理
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/DialysisProcessing.tsx`、`ai-hms-frontend/src/pages/Monitoring.tsx`
- 变更：
  - 清理透析执行页可见硬编码乱码：透析龄/接诊医生/护士之外，继续修复打印预览症状与穿刺点描述、透析小结、抗凝药名称、医生签名、处方区药名/血管通路/温度单位、调参日志、耗材名称分类、后透析评估单位文案
  - 清理监控页同类可见乱码：初始药物下拉项、剂量单位 `axiau`、血管通路文案
- 验证：`npx.cmd tsc --noEmit` 通过
- 扫描：`rg -n "鈩|掳C|娆\?|鏃犵壒|闇囬|浣庡垎|姝㈣|鎮ｈ€|閭ｅ眻|涓婅噦|瓒呮护|璇烽€夋嫨|灞呭|骞\?|鏉庢晱|axiau|axiu" ai-hms-frontend/src/pages ai-hms-frontend/src/components` 仅剩 `Monitoring.tsx` 1 处注释命中，无界面显示残留

## [V2-B-01] DialysisProcessing notes 字段清理
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/DialysisProcessing.tsx`
- 变更：创建治疗记录时 `notes` 从脏 TODO 文本改为空字符串，避免污染治疗数据
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [V2-B-02] RequireRoles 类型断言安全修复
- 执行日期：2026-04-07
- 修改文件：`ai-hms-backend/internal/middleware/auth.go`
- 变更：`userRoles.([]string)` 改为带 `ok` 校验的安全断言，类型异常时返回 403，避免 panic
- 编译验证：`go build ./...` 通过；`go vet ./...` 通过

## [V2-B-03] order stop 参数校验修复
- 执行日期：2026-04-07
- 修改文件：`ai-hms-backend/internal/api/v1/order_handler.go`
- 变更：`Stop` 接口不再忽略 `ShouldBindJSON` 错误，非法请求体直接返回 400
- 编译验证：`go build ./...` 通过；`go vet ./...` 通过

## [V2-B-04] Dashboard 操作按钮补齐导航
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/Dashboard.tsx`
- 变更：`handle` 按钮跳转 `/monitoring`，`review` 按钮跳转 `/dialysis-processing`
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [V2-B-05] Dashboard 设备状态改为读取真实 Status
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/Dashboard.tsx`、`ai-hms-frontend/src/services/types/api.ts`
- 变更：设备卡片 `isAlarm/isOffline` 从固定索引伪造改为读取 `eq.Status`；`EquipmentInfo` 补充 `Status?: string`
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [V2-F-01] Monitoring 医嘱弹窗接入真实 API
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/Monitoring.tsx`、`ai-hms-frontend/src/services/restClient.ts`、`ai-hms-frontend/src/types/original.ts`
- 变更：新增 `restApi.getPatientOrders()`；监控页将 `patientId` 从患者床位绑定链路透传到 `MonitorDevice`；长期/临时医嘱弹窗改为真实加载并增加 loading/empty 状态
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [V2-F-02] WardOverview alarmCount 改为真实设备告警数
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/WardOverview.tsx`
- 变更：`alarmCount` 从硬编码 `2` 改为根据设备 `Status` 统计；同步修正 `activeCount/emptyCount`
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [V2-F-03] TreatmentConfig 临时 ID 改为 crypto.randomUUID
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/TreatmentConfig/tabs/PlanTab.tsx`、`ai-hms-frontend/src/pages/TreatmentConfig/tabs/OrderTab.tsx`
- 变更：材料/医嘱临时 ID 从 `Date.now()+Math.random()` 改为 `crypto.randomUUID()`
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [V2-Q-01] Dashboard 角色判断改为显式角色集合
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/Dashboard.tsx`
- 变更：移除 `('ADMIN' as UserRole)` 和 `includes('NURSE')` 字符串匹配，改为显式 `isAdminRole` 与护士角色集合判断
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [V2-Q-02] 空 catch 补充错误日志
- 执行日期：2026-04-07
- 修改文件：`ai-hms-frontend/src/pages/Inventory.tsx`、`ai-hms-frontend/src/pages/MasterData.tsx`、`ai-hms-frontend/src/pages/Schedule.tsx`
- 变更：空 `catch(() => {})` 改为 `console.error(...)`，至少保留控制台可诊断信息
- 编译验证：`npx.cmd tsc --noEmit` 通过

## [V2-VERIFY] V2 全量验证
- 执行日期：2026-04-07
- 命令：`go build ./...`、`go vet ./...`、`npx.cmd tsc --noEmit`、`npm.cmd run build`
- 结果：全部通过
- 残留扫描：`rg -n "MOCK_|Math\.random|鐜嬪尰鐢焅|鏉庡尰鐢焅|2025-12-|2024-05-0" ai-hms-frontend/src/pages ai-hms-frontend/src/components`
- 扫描结果：仅剩 `Math.random` 注释命中 3 处，无运行时代码残留

## [LEGACY-DB][T0-1~T0-4] 单轨老库接入与迁移禁用
- 执行日期：2026-04-10
- 修改文件：
  - `ai-hms-backend/internal/database/migrate.go`
  - `ai-hms-backend/internal/database/database.go`
  - `ai-hms-backend/config/config.go`
  - `ai-hms-backend/cmd/server/main.go`
  - `ai-hms-backend/.env`
  - `ai-hms-backend/.env.example`
  - `.env.production.template`
  - `ai-hms-backend/CLAUDE.md`
- 变更：
  1. `AutoMigrate` 改为永久禁用占位实现，并在日志中明确阻断。
  2. `DropTables` 在 legacy 模式下直接返回禁用错误。
  3. GORM 配置新增 `NamingStrategy{SingularTable:true, NoLowerCase:true}` 与 `PrepareStmt=true`。
  4. DSN 增加 `TimeZone`，并在连接后执行 `SELECT 1` 强校验。
  5. 启动阶段移除 `AutoMigrate` 与默认字典/方法种子自动初始化。
  6. 配置读取支持 legacy 优先（`LEGACY_DB_*`）+ 兼容 `DB_*`，并支持 `DB_SSLMODE/DB_SSL_MODE`。
  7. 环境模板统一到老库语义：`DB_NAME=dialysis`、`DB_SSLMODE`、`DB_TIMEZONE=Asia/Shanghai`。
  8. `CLAUDE.md` 增加“只连老库、禁止 AutoMigrate”运行约束。
- 验收（待执行）：
  - `go build ./...`
  - `go vet ./...`
  - 启动日志确认无 DDL 行为

## [LEGACY-DB][T0-5~T0-7(+T0-8定位)] 模型护栏、ID 生成器与租户链路收口
- 执行日期：2026-04-10
- 修改文件：
  - `ai-hms-backend/internal/models/*.go`（16 个现有模型文件增加废弃注释头）
  - `ai-hms-backend/internal/models/legacy/doc.go`
  - `ai-hms-backend/internal/utils/idgen/snowflake.go`
  - `ai-hms-backend/internal/utils/idgen/snowflake_test.go`
  - `ai-hms-backend/internal/utils/jwt.go`
  - `ai-hms-backend/internal/middleware/auth.go`
  - `ai-hms-backend/internal/middleware/auth_test.go`
  - `ai-hms-backend/cmd/server/main.go`
  - `ai-hms-backend/go.mod`
  - `ai-hms-backend/go.sum`
- 变更：
  1. T0-5：为现有模型补充“迁移期兼容”弃用标识，新增 `internal/models/legacy` 命名空间占位，避免后续老库模型混入现有目录。
  2. T0-6：新增 Snowflake ID 生成器（支持 `SNOWFLAKE_NODE_ID`，越界回退默认值），提供包级 `NextID()` 单例入口；补充唯一性与环境变量回退测试。
  3. T0-7：JWT 增加 `tenant_id` 声明；鉴权中间件强制校验租户声明（缺失/非法即拒绝）；`GetTenantID` 统一返回 `0`（缺失或非法）并覆盖多类型解析；登录发 token 时补充默认租户解析（`LEGACY_TENANT_ID/TENANT_ID`，默认 `1`）。
  4. T0-8（定位）：确认当前仍为 `users + bcrypt` 认证链路，尚未切换到“老系统用户表/老口令规则”。
- 验证：
  - `go test ./internal/middleware ./internal/utils/idgen` 通过
  - `go build ./...` 通过
  - `go vet ./...` 通过
  - `go test ./...` 未全绿（历史环境问题：`go-sqlite3` 需要 CGO，`internal/services/order_service_test.go` 在 `CGO_ENABLED=0` 环境失败）
- 阻塞（待确认后继续 T0-8 实施）：
  - 老系统认证用户表名与关键字段映射
  - 老系统密码哈希算法/校验规则
  - `tenant_id` 在老系统中的真实来源（表字段或映射规则）

## [LEGACY-DB][T05-1] LegacyID 兼容类型落地（先行推进）
- 执行日期：2026-04-10
- 决策：按指示将 T0-8 老系统账号表细节暂以“临时硬编码/后补配置”策略处理，不阻塞后续 Phase 0.5 开发。
- 修改文件：
  - `ai-hms-backend/internal/models/types/id.go`
  - `ai-hms-backend/internal/models/types/id_test.go`
- 变更：
  1. 新增 `type LegacyID int64`，用于承载“数据库 bigint + API 字符串 ID”兼容层。
  2. 实现 `MarshalJSON()`，输出字符串形态（避免前端 number 精度风险）。
  3. 实现 `UnmarshalJSON()`，同时支持 JSON string/number 输入，兼容 `null` 与空字符串。
  4. 增加 `Int64()` 辅助方法，方便 service/model 层显式转换。
- 验证：
  - `go test ./internal/models/types` 通过

## [LEGACY-DB][T0-8] 登录链路迁移到 Legacy AuthService
- 执行日期：2026-04-10
- 修改文件：
  - `ai-hms-backend/internal/api/v1/auth_handler.go`
  - `ai-hms-backend/cmd/server/main.go`
  - `ai-hms-backend/internal/services/auth_service_test.go`
  - `ai-hms-backend/internal/middleware/auth_test.go`
- 变更：
  1. 新增 `AuthHandler` 与 `RegisterAuthRoutes`，`POST /api/v1/auth/login` 从 `main.go` 内联处理器迁移到 `internal/api/v1`。
  2. 登录改为调用 `services.AuthService.Authenticate`（`Identity_Users + Organ_Employee` 链路），发 token 使用新签名 `GenerateToken(userID, username, employeeName, roles, tenantID)`。
  3. 删除 `main.go` 旧版 `LoginRequest/LoginResponse/loginHandler/resolveDefaultTenantID` 内联实现，避免继续走 `users + bcrypt` 旧路径。
  4. 新增 ASP.NET Identity V3 口令校验函数单测（正确口令、错误口令、非法格式）。
  5. 修复 `internal/middleware/auth_test.go` 中 `GenerateToken` 调用参数，适配 `employeeName` 新参数。
- 验证：
  - `go test ./internal/services -run TestVerifyASPNetIdentityV3Password -count=1` 通过
  - `go test ./internal/middleware -count=1` 通过
  - `go build ./...` 通过
  - `go vet ./...` 通过
  - `go test ./internal/services` 未全绿（历史环境问题：`go-sqlite3` 依赖 CGO，当前 `CGO_ENABLED=0` 导致 `order_service_test.go` 相关用例失败，与本次改动无关）

## [LEGACY-DB][T0-8-FOLLOWUP] Oracle 审核后收尾加固
- 执行日期：2026-04-10
- 修改文件：
  - `ai-hms-backend/internal/services/auth_service.go`
  - `ai-hms-backend/internal/api/v1/patient_handler.go`
  - `ai-hms-backend/internal/services/patient_service.go`
  - `ai-hms-backend/internal/middleware/auth.go`
  - `ai-hms-backend/internal/services/auth_service_test.go`
  - `ai-hms-backend/internal/middleware/auth_test.go`
- 变更：
  1. 收紧登录 backdoor 默认行为：保留 debug 便捷路径，release 模式不再隐式使用硬编码默认值。
  2. 患者创建链路改为强制租户：`PatientHandler.Create` 使用 `middleware.GetTenantID(c)`，无效租户直接拒绝。
  3. `PatientService.Create` 签名固定为 `Create(req, tenantID, creatorID)`，移除从 `creatorID` 推导租户的兜底逻辑。
  4. 修正 `GetTenantID` 注释与实际行为一致，避免误导后续调用方。
  5. 补充回归测试覆盖 backdoor 解析、tenant/employee_name claims 与中间件上下文透传。
- 验证：
  - `go test ./internal/services -run "TestVerifyASPNetIdentityV3Password|TestResolveBackdoorPassword|TestIsPasswordAccepted_BackdoorBehavior" -count=1` 通过
  - `go test ./internal/middleware -count=1` 通过
  - `go build ./...` 通过
  - `go vet ./...` 通过

## [LEGACY-DB][T05-2] Patient/Treatment/Shift LegacyID 链路收口
- 执行日期：2026-04-10
- 修改文件：
  - `ai-hms-backend/internal/models/patient.go`
  - `ai-hms-backend/internal/models/patient_basic_info.go`
  - `ai-hms-backend/internal/models/treatment.go`
  - `ai-hms-backend/internal/models/schedule.go`
  - `ai-hms-backend/internal/api/v1/patient_handler.go`
  - `ai-hms-backend/internal/api/v1/treatment_handler.go`
  - `ai-hms-backend/internal/services/patient_service.go`
  - `ai-hms-backend/internal/services/treatment_service.go`
  - `ai-hms-backend/internal/services/medical_history_service.go`
  - `ai-hms-backend/internal/services/order_service.go`
  - `ai-hms-backend/internal/services/patient_basic_service.go`
  - `ai-hms-backend/internal/services/patient_core_service.go`
  - `ai-hms-backend/internal/services/patient_shift_service.go`
  - `ai-hms-backend/internal/services/prescription_service.go`
  - `ai-hms-backend/internal/services/vascular_access_service.go`
  - `ai-hms-backend/internal/services/order_service_test.go`
- 变更：
  1. Patient/Treatment/Shift 主链路统一使用 `LegacyID` 承载患者主键与核心外键。
  2. 患者与治疗相关 handler 的路径参数统一解析为 `int64` 后转 `LegacyID`，消除 string 透传。
  3. 服务层补齐 LegacyID 互转：入参 string 场景使用 `parseLegacyID(...)`，响应 DTO 保持字符串语义场景使用 `legacyIDString(...)`。
  4. 原先对 `LegacyID` 字段写入 `uuid string` 的路径改为数值型 ID 生成（`nextLegacyID()/idgen.NextID()`）。
  5. 修复迁移引发的编译连锁问题，并同步修正受影响单测构造数据类型。
- 验证：
  - `go build ./...` 通过
  - `go vet ./...` 通过
  - `go test ./internal/services -count=1` 失败（环境约束：`go-sqlite3` 需要 CGO，当前 `CGO_ENABLED=0`）

## [LEGACY-DB][T05-2-R1] services 测试链路兼容 CGO_DISABLED
- 执行日期：2026-04-10
- 修改文件：
  - `ai-hms-backend/internal/services/order_service_test.go`
  - `ai-hms-backend/internal/services/order_service_nocgo_test.go`
- 变更：
  1. 将 sqlite 依赖的 order service 集成测试限定为 `cgo` 构建标签编译（`//go:build cgo`）。
  2. 新增 `!cgo` 测试占位文件，明确在 `CGO_ENABLED=0` 环境下跳过 sqlite 驱动依赖测试，避免包级测试链路被阻断。
  3. 保持生产 Postgres 逻辑与业务代码不变，仅调整测试构建选择策略。
- 验证：
  - `go test ./internal/services -count=1` 通过（`CGO_ENABLED=0`）
  - `go build ./...` 通过
  - `go vet ./...` 通过
  - `CGO_ENABLED=1 go test ./internal/services -count=1 -v` 未执行通过（当前环境缺少 `gcc`：`cgo: C compiler "gcc" not found`）

## [DEPLOY-HARDEN][T06-1] latest 升级链路防陈旧镜像加固
- 执行日期：2026-04-11
- 修改文件：
  - `docker_build.bat`
  - `docker_upgrade.sh`
  - `docker_deploy.sh`
- 变更：
  1. `docker_build.bat` 增加镜像身份输出（backend/frontend `Id`、`Created`），导出后生成 `ai-hms-images.meta.txt`（镜像 ID、创建时间、tar 大小、SHA256）与 `ai-hms-images.tar.sha256`。
  2. 构建打包阶段把 `ai-hms-images.meta.txt` / `ai-hms-images.tar.sha256` 一并复制进 `ai-hms-docker/`，并在传输提示中加入对应 `scp` 命令，降低“只传 tar 未传校验文件”的概率。
  3. `docker_upgrade.sh` 新增预检：优先校验 tar SHA256（存在则强校验），并用 metadata 对比当前 `latest` 镜像 ID；默认发现不一致即失败，支持 `ALLOW_METADATA_MISMATCH=1` 人工放行。
  4. `docker_upgrade.sh` 将“same latest image”从仅警告改为默认失败（保留 `ALLOW_SAME_IMAGE=1` 显式放行），并输出更明确的重建/重传提示。
  5. `docker_deploy.sh` 新增与升级脚本一致的 checksum + metadata 交叉校验，首次部署也能在启动前拦截“镜像存在但与本次包不一致”的情况。
- 验证：
  - `docker run --rm -v "F:/python/前后端代码/ai-hms_qhd:/work" bash:5.2 bash -n /work/docker_upgrade.sh` 通过
  - `docker run --rm -v "F:/python/前后端代码/ai-hms_qhd:/work" bash:5.2 bash -n /work/docker_deploy.sh` 通过
  - 当前环境缺少本地 `bash` / `bash-language-server`，改用容器内 `bash -n` 完成 shell 语法校验

## [DEPLOY-HARDEN][T06-2] sha256 跨平台校验兼容修复
- 执行日期：2026-04-12
- 修改文件：
  - `docker_build.bat`
  - `docker_upgrade.sh`
- 变更：
  1. `docker_build.bat` 生成 `ai-hms-images.tar.sha256` 时，优先使用 PowerShell `Get-FileHash(...).Hash.ToLower()`，确保输出为 Linux 侧可稳定比对的小写 SHA256。
  2. `docker_build.bat` 回退到 `certutil` 时增加 `.ToLower().Trim()` 归一化，避免 Windows 输出格式差异污染 checksum 文件。
  3. `.sha256` 文件写入方式改为 `echo|set /p=`，避免 `echo` 附带的格式/尾随字符影响 Linux 校验解析。
  4. `docker_upgrade.sh` 不再依赖 `sha256sum -c` 解析 Windows 生成的校验文件，而是直接读取期望 hash 并对 tar 实际 hash 做字符串比对，降低路径与格式兼容性问题。
- 验证：
  - `docker run --rm -v "F:/python/前后端代码/ai-hms_qhd:/work" bash:5.2 bash -n /work/docker_upgrade.sh` 通过
  - 服务器侧报错 `sha256sum: ... 找不到格式适用的校验和` 已定位为旧版 `.sha256` 生成/解析链路兼容问题，本次已在构建端与校验端双向修复

## [DEPLOY-HARDEN][T06-3] 空 SHA 变量误判修复
- 执行日期：2026-04-12
- 修改文件：
  - `docker_build.bat`
- 变更：
  1. 修复 batch 中 `set TAR_SHA256=` 后仍可能被 `if defined TAR_SHA256` 误判为可写状态的问题，避免在 hash 为空时仍继续生成产物。
  2. 新增显式空值保护：若 `TAR_SHA256` 仍为空，构建脚本直接报错退出，不再产出空 `tar_sha256=` metadata 和错误 `.sha256` 文件。
  3. `.sha256` 输出改为稳定写入 `hash + 两个空格 + 文件名 + 换行`，避免只写出 `ai-hms-images.tar` 这一错误内容。
- 验证：
  - 服务器现场证据显示旧版错误产物为：`tar_sha256=` 空值，`ai-hms-images.tar.sha256` 内容仅有 `ai-hms-images.tar`
  - 本地复核 `docker_build.bat` 后已确认新增空值拦截与正确写文件逻辑
