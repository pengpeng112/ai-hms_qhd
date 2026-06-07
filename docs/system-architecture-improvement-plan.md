# 系统功能架构改进计划

生成日期：2026-06-01

## 1. 背景与目标

当前系统由 `ai-hms-backend/` 和 `ai-hms-frontend/` 两个活动应用组成。后端直接连接老血透 PostgreSQL，前端为 React + Vite + TypeScript。系统正处于“新 UI / 新接口逐步迁移到老血透库”的阶段，已经完成部分老库对接，但仍存在权限、租户、新表残留、患者写操作、前端 API 分层等架构风险。

本计划用于复核后再交给其他 AI 或开发人员实施。计划只描述改造，不包含本次直接代码实现。

### 总体目标

- 后端高危接口具备明确权限控制。
- 后端只读写已确认的老血透权威表，不依赖未确认的新表。
- 租户过滤策略统一，不再散落硬编码 `legacyTenantID = 3`。
- 服务启动不会默认执行可能改写老库数据的后台任务。
- 患者创建、更新、删除不再使用新库字段或硬删除方式破坏老库。
- 前端路由权限、菜单权限、后端权限策略保持一致。
- 前端 API 调用方式收敛，不新增直接 `fetch('/api/...')` 或扩张历史 `restClient.ts` facade。

### 非目标

- 不做全局 UI 重构。
- 不执行 DDL，不恢复 `AutoMigrate`。
- 不做微服务拆分。
- 不全量重写业务模型。
- 不新增新数据库表，除非后续单独确认。

## 2. 当前关键问题

| 编号 | 优先级 | 问题 | 主要位置 | 风险 |
|---|---|---|---|---|
| A-01 | P0 | 后端多数业务路由只有登录认证，无角色/权限控制 | `cmd/server/main.go`、`internal/middleware/auth.go` | 登录用户可能越权管理用户、角色、权限、患者、治疗、库存 |
| A-02 | P0 | 多处硬编码 `legacyTenantID = 3` | `auth_service.go`、`treatment_service.go`、`dashboard_service.go`、`inventory_service.go` | 多租户隔离失效或部署到错误租户 |
| A-03 | P0 | 服务启动自动执行医嘱停用任务，且无租户过滤 | `cmd/server/main.go`、`order_cron.go` | 启动即改写老库，可能跨租户停用医嘱 |
| A-04 | P0 | 患者写操作混用新表/新字段，并硬删除患者 | `patient_service.go`、`patient_basic_info.go` | 写坏老库患者数据或破坏历史数据 |
| A-05 | P0 | 新表模型/读写路径残留 | `models/*.go`、`services/*.go` | 访问不存在或非权威表，形成数据割裂 |
| A-06 | P1 | 前端只隐藏菜单，不做路由级权限守卫 | `AuthGuard.tsx`、`Sidebar.tsx`、`role.ts` | 可通过 URL 直达隐藏页面 |
| A-07 | P1 | 前端 API 层职责混乱 | `restClient.ts`、各 `*Api.ts` | 合约漂移、类型重复、维护困难 |
| A-08 | P1 | 存在直接 `fetch` 和错误 token key | `ScheduleTemplateList.tsx`、`ApplyTemplateModal.tsx` | 页面不可用，绕过统一错误处理 |
| A-09 | P1 | REST 登录、OAuth、本地预览登录并存 | `auth.ts`、`restAuth.ts`、`AuthGuard.tsx` | 认证入口混乱，存在误配置风险 |
| A-10 | P2 | 路由、菜单、面包屑、权限码分散维护 | `router.tsx`、`Sidebar.tsx`、`routeMeta.ts`、`role.ts` | 孤儿页面、权限和导航不同步 |

## 3. 改造原则

- 先保护老库，再完善功能。
- 不确定字段先禁用写入或标注待确认，不猜测落库。
- 老库 SQL 表名和列名统一使用双引号。
- 高危写接口先加权限，再谈业务完善。
- 每次只改一个模块或一类问题，改完必须验证。
- 不修改用户未授权的无关代码和数据。

## 4. Phase 0：安全保护与写操作收敛

目标：先降低越权和老库误写风险。

### 4.1 后端高危接口权限分组

涉及文件：

- `ai-hms-backend/cmd/server/main.go`
- `ai-hms-backend/internal/middleware/auth.go`
- `ai-hms-backend/internal/api/v1/*_handler.go`

改造内容：

- 在 `main.go` 中拆分路由组：`protected`、`admin`、`clinicalWrite`。
- `admin` 使用 `RequireRoles("ADMIN")` 或后续权限码中间件。
- 管理类接口先全部纳入 `admin`：用户、角色、权限、HDIS 设置、系统日志、字典写操作。
- 临床写接口纳入 `clinicalWrite`：治疗编辑、患者删除、医嘱停用/复制/分组、库存调整。
- 只读接口仍可保留在 `protected`。

验收标准：

- 普通角色访问用户/角色/权限管理接口返回 403。
- ADMIN 访问同类接口正常。
- 路由注册处能清晰看出每类接口的权限边界。

### 4.2 医嘱定时任务默认关闭

涉及文件：

- `ai-hms-backend/config/*`
- `ai-hms-backend/cmd/server/main.go`
- `ai-hms-backend/internal/services/order_cron.go`

改造内容：

- 新增配置 `ORDER_CRON_ENABLED=false`，默认不启动 `StartOrderCron()`。
- `disableExpiredLegacyOrders()` 增加 `TenantId` 条件。
- 如果后续要开启，增加 PostgreSQL advisory lock 或单实例开关。
- 日志明确输出当前定时任务是否启用。

验收标准：

- 默认启动服务不会修改 `Order_PatientOrder`。
- 开启配置后，SQL 必须包含 `TenantId` 过滤。

### 4.3 临时限制患者高危写操作

涉及文件：

- `ai-hms-backend/internal/services/patient_service.go`
- `ai-hms-backend/internal/api/v1/patient_handler.go`

改造内容：

- 患者删除先改为禁用或返回“暂不支持直接删除”，禁止硬删除。
- 患者更新只允许已确认的老库字段，未确认字段不写入。
- 创建患者如字段映射未完整确认，先要求管理员权限并收敛为最小必填字段。
- 移除或绕开 `patient_basic_infos` 新表写入。

验收标准：

- 代码中不再出现患者硬删除 `Delete(&models.Patient{})`。
- 更新 SQL 不再使用 `bed_number`、`risk_level`、`status`、`patient_type` 等新库字段名。
- 患者创建/更新只写老库真实字段。

## 5. Phase 1：租户策略统一

目标：统一所有服务的租户来源，避免跨租户读写。

### 5.1 明确租户模式

需要用户确认：

- 当前系统是否永久只服务 `TenantId=3`？
- 是否存在未来多租户部署需求？

方案 A：单租户模式

- 新增必填配置 `LEGACY_TENANT_ID`。
- 登录 JWT 中写入该配置值。
- 所有服务通过统一 helper 获取租户，不再散落硬编码。

方案 B：多租户模式

- 登录时从老库用户/员工/机构关系解析真实 `TenantId`。
- handler 从 JWT context 获取 tenantID 并传入 service。
- 所有 service 查询必须使用传入的 tenantID。

建议：如果短期只服务当前院区，先采用方案 A，但必须配置化，不继续硬编码。

### 5.2 清理 `legacyTenantID` 散落使用

涉及文件：

- `auth_service.go`
- `patient_service.go`
- `treatment_service.go`
- `dashboard_service.go`
- `inventory_service.go`
- `dict_service.go`
- 其他 grep 到 `legacyTenantID` 的 service

改造内容：

- 建立统一函数，例如 `LegacyTenantID()` 或从 config 注入。
- 新服务方法优先接收 `tenantID int64`。
- 对暂不改签名的大服务，先用统一 helper 替换常量。

验收标准：

- 全仓搜索 `legacyTenantID = 3` 仅允许保留在统一配置/兼容层。
- 核心业务 SQL 均有明确 `TenantId` 来源。

## 6. Phase 2：新表残留清理

目标：确保系统只依赖老库权威表或明确允许的兼容配置。

### 6.1 建立表使用白名单

新增文档建议：

- `docs/legacy-table-usage-whitelist.md`

内容：

- 模块名。
- 允许使用的老库表。
- 是否只读。
- 是否允许写入。
- 字段映射文档链接。

### 6.2 清理或隔离新表模型

重点文件：

- `models/patient_basic_info.go`
- `models/treatment.go`
- `models/dict.go`
- `models/clinical_task.go`
- `models/integration_hdis_setting.go`

处理策略：

- `patient_basic_infos`：迁移到 `Register_PatientInfomation` 及相关老库表。
- `orders`：已迁移部分必须统一到 `Order_PatientOrder` 等老库表。
- `dict_types/dict_items`：如仍需兼容，要明确是否允许保留；否则统一到 `CodeDictionary_CodeDictionarys`。
- `clinical_tasks`：如老库无对应表，接口应返回暂不支持或改为从治疗/医嘱/排班派生。
- `integration_hdis_settings`：如必须保存配置，需要用户明确批准新配置表或改为环境变量。

验收标准：

- `TableName()` 返回的新表名均被白名单解释。
- 不在白名单中的新表读写路径已删除、禁用或改造。

## 7. Phase 3：患者写操作专项重构

目标：将患者写操作改为安全、明确、可审计。

### 7.1 患者创建

改造内容：

- 明确创建患者最小字段集：`TenantId`、`Name`、`Gender`、`IDNo`、`PhoneNo`、`PatientType` 等需按老库字段确认。
- 禁止写 `gorm:"-"` 字段。
- 不写 `patient_basic_infos`。
- 创建后必须能通过患者列表和患者详情读取。

验收标准：

- 创建 SQL 只涉及老库真实表。
- 创建字段在老库结构文档中有对应。

### 7.2 患者更新

改造内容：

- 建立字段映射表：前端字段 -> 老库列 -> 是否允许写。
- 只允许写已确认字段。
- 干体重、诊断等派生字段继续写入各自权威表，不写患者主表不存在列。

验收标准：

- 不再出现 snake_case 新库字段名更新老库患者表。
- 修改基本信息后详情页能正确回显。

### 7.3 患者删除

改造内容：

- 禁止硬删除。
- 如老库有状态字段，采用状态停用/转归方式。
- 如果无法确认，则接口返回 409 或 400，提示“不支持直接删除患者”。

验收标准：

- 无硬删除患者主表代码。
- 删除按钮前端根据接口能力禁用或提示。

## 8. Phase 4：前端权限与 API 层收敛

目标：让前端权限、路由、API 调用与后端一致。

### 8.1 路由级权限守卫

涉及文件：

- `src/components/AuthGuard.tsx`
- `src/services/role.ts`
- `src/router.tsx`

改造内容：

- 新增 `PermissionGuard` 或增强 `AuthGuard`。
- 根据当前 path 获取 menuKey/permissionCode。
- 未授权时显示 403 页面或跳转默认首页。

验收标准：

- 非管理员直接访问 `/user-management`、`/role-management` 被拦截。
- 菜单隐藏和路由访问规则一致。

### 8.2 修复直接 fetch

涉及文件：

- `src/pages/ScheduleTemplateList.tsx`
- `src/components/schedule/ApplyTemplateModal.tsx`

改造内容：

- 禁止直接读取 `auth_token`。
- 使用统一 `apiClient`。
- 改为后端真实路径，如 `/api/v1/patient-shifts/templates`。

验收标准：

- 全仓搜索 `fetch('/api`、`auth_token` 无业务调用残留。

### 8.3 API 模块化收敛

涉及文件：

- `src/services/restClient.ts`
- `src/services/index.ts`
- 各 `*Api.ts`

改造内容：

- `restClient.ts` 只保留 `apiClient`、通用响应类型、错误处理、少量历史 facade。
- 新业务统一写独立 `*Api.ts`。
- 页面逐步从 `restApi.xxx` 迁移到独立模块。

验收标准：

- 新增页面或新功能不得使用 `restApi.xxx`。
- 每个业务 API 模块有清晰 DTO 类型。

## 9. Phase 5：认证体系整理

目标：明确生产认证路径，降低误配置风险。

需要用户确认：

- 生产是否继续使用 OAuth？
- 是否保留本地预览登录？

建议方案：

- 如果生产只用 REST 登录：隔离或移除 `auth.ts` OAuth 入口。
- OAuth 配置如保留，改为环境变量，不硬编码。
- 本地预览登录仅在 `DEV` 或显式 `VITE_ENABLE_LOCAL_PREVIEW_LOGIN=true` 时显示入口。
- token shape 统一到 `utils/token.ts`。

验收标准：

- 生产构建没有硬编码 OAuth 域名和 clientId。
- 登录、登出、刷新用户态只有一套主流程。

## 10. Phase 6：路由配置单一来源

目标：减少路由、菜单、面包屑、权限码不同步。

涉及文件：

- `src/router.tsx`
- `src/layouts/Sidebar.tsx`
- `src/layouts/routeMeta.ts`
- `src/services/role.ts`

改造内容：

- 新增统一配置，例如 `src/router/routeConfig.tsx`。
- 每项包含：`path`、`element`、`title`、`menuKey`、`permissionCode`、`icon`、`hiddenInMenu`。
- 由配置生成路由、菜单、面包屑和权限判断。

验收标准：

- 排班模板等页面不再是孤儿路由。
- 新增页面只需改一处配置。

## 11. 推荐实施顺序

1. Phase 0.1 后端高危接口权限分组。
2. Phase 0.2 医嘱定时任务默认关闭。
3. Phase 0.3 患者删除/更新高危路径先收敛。
4. Phase 1 租户配置化。
5. Phase 2 新表残留白名单和清理。
6. Phase 3 患者写操作专项重构。
7. Phase 4 前端路由权限和 API 调用修复。
8. Phase 5 认证体系整理。
9. Phase 6 路由配置单一来源。

## 12. 每阶段验证命令

后端：

```powershell
go build -o "$env:TEMP\ai-hms-backend-check.exe" ./cmd/server
go test ./internal/services ./internal/api/v1 -count=1 -short
go vet ./...
```

前端：

```powershell
npm run lint
npm run build
```

静态检查建议：

```powershell
rg "AutoMigrate|DropTable|CreateTable|ALTER TABLE|DROP TABLE" ai-hms-backend
rg "legacyTenantID|patient_basic_infos|dict_types|dict_items|clinical_tasks|integration_hdis_settings" ai-hms-backend/internal
rg "fetch\('/api|fetch\(\"/api|auth_token|TODO|Mock|占位" ai-hms-frontend/src
```

## 13. 需要用户决策的问题

1. 当前系统是否确定只服务 `TenantId=3`？
2. 患者创建、更新、删除是否允许直接写老库？还是先禁用写操作？
3. 角色权限最终以 `Identity_Roles` 还是 `Authorization_Roles` 为业务权限体系？
4. `integration_hdis_settings` 这类配置是否允许继续存数据库？如果允许，是否允许新表？
5. 生产是否保留 OAuth 登录？
6. 排班模板是否继续使用 `Schedule_PatientShift.Status=60` 作为模板实现？

## 14. 交给开发 AI 的执行要求

- 每次只执行一个 Phase 或一个子任务。
- 开发前先读取本计划、`AGENTS.md`、老库结构文档。
- 不做计划外重构。
- 不执行 DDL。
- 不改 `.env`、不提交密钥。
- 修改后必须运行对应验证命令。
- 遇到字段含义不确定，记录到待确认文档，不猜测写库。

---

# 附录 A：复核意见

复核人：opencode (代码交叉验证)
复核日期：2026-06-01

## A.1 计划中问题断言的验证结论

| 编号 | 问题断言 | 验证结论 | 补充事实 |
|---|---|---|---|
| A-01 | 后端多数路由只有登录认证，无角色控制 | **属实** | `main.go` 中仅 `hdis_settings` 和 `log` 两组路由使用了 `RequireRoles("ADMIN")`，其余 20+ 组路由全部仅挂 `AuthMiddleware`，无任何角色/权限码校验 |
| A-02 | 多处硬编码 `legacyTenantID = 3` | **属实，但规模被低估** | 计划中列出 5 个文件，实际跨 **16 个 service**、**198 处引用**。涉及：`treatment_service.go`(50)、`treatment_config_service.go`(40)、`patient_service.go`(22)、`vascular_access_service.go`(18)、`device_service.go`(14)、`medical_history_service.go`(14)、`prescription_service.go`(10)、`dashboard_service.go`(8)、`patient_core_service.go`(6)、`inventory_service.go`(3)、`key_indicator_service.go`(3)、`order_service.go`(3)、`auth_service.go`(4)、`lab_report_service.go`(1)、`order_cron.go`(隐含)、测试文件(2) |
| A-03 | 服务启动自动执行医嘱停用任务，无租户过滤 | **属实** | `main.go:48` 无条件调用 `StartOrderCron()`；`order_cron.go` 的 SQL 完全没有 `TenantId` 条件 |
| A-04 | 患者写操作混用新表/新字段，硬删除 | **属实，但严重度更高** | (1) `PatientBasicInfo` 新表有 **37 处引用**跨 8 个文件；(2) `Delete()` 函数注释明确写着"硬删除"，同时会 `Delete(&models.PatientBasicInfo{})`；(3) 更新使用 `bed_number`/`risk_level`/`status`/`patient_type` 等小写下划线字段名，与老库大写驼峰命名不匹配 |
| A-05 | 新表模型/读写路径残留 | **属实，但缺失关键细节** | 计划遗漏了 `orders` 新表问题：`models.Order.TableName()` 返回 `"orders"`（新表），`order_service.go` 中有 **8 处** `Model(&models.Order{})` 操作新表，同时同文件有 `legacyPatientOrder.TableName()` 指向老表 `"Order_PatientOrder"`。这是**新旧表并存混用**而非仅仅残留 |
| A-06 | 前端只隐藏菜单不做路由权限守卫 | **属实** | `AuthGuard.tsx` 仅检查 `isLoggedIn()` + `hasSelectedRole()`，无具体角色/权限码校验 |
| A-07 | 前端 API 层职责混乱 | **属实** | `restClient.ts` 2651 行，混合 HTTP 配置 + 50+ 类型定义 + 全部业务 API |
| A-08 | 直接 fetch 和错误 token key | **属实，但不完整** | 计划只提到 2 处 `auth_token`，实际还有 3 处绕过 `getToken()` 直接读 `localStorage.getItem('hdis_access_token')`：`MaterialTab.tsx:179`、`DrugTab.tsx:312`、`BatchImportModal.tsx:220` |
| A-09 | 三路登录并存 | **属实** | OAuth 硬编码 `authServer: 'https://auth.aihhd.com'` + `clientId: 'yanshi7779'`；本地预览登录用伪造 admin token `'local-preview-token'` |
| A-10 | 路由/菜单/权限码分散 | **属实** | 无进一步细节差异 |

## A.2 计划整改方案的补充和修正

### A.2.1 Phase 0.1 权限分组 — 修正

计划中建议 `RequireRoles("ADMIN")`，但当前代码 `RequireRoles` 仅接受角色名称字符串并做精确匹配。需注意：

1. **角色名称不一致**：老库有两套角色表 `Identity_Roles`（仅 ADMIN）和 `Authorization_Roles`（7 条业务角色：管理员/医生/护士长/护士/主任/安全管理员/运维管理员）。`RequireRoles("ADMIN")` 只匹配 `Identity_Roles` 的 ADMIN 行，不匹配 `Authorization_Roles` 的"管理员"。
2. **建议补充**：权限中间件应同时支持角色码和权限码两种粒度。当前只有角色维度的 `RequireRoles`，缺少更细的权限码（如 `patient:delete`、`order:edit`）控制能力。
3. **计划遗漏**：`permission_service.go` 中角色权限关联使用 `Authorization_RolePermissions.RoleId` JOIN `Authorization_Roles`，而角色列表 CRUD 使用 `Identity_Roles`。这两种表混用本身就是 bug，计划中需要明确统一方案。

### A.2.2 Phase 0.2 定时任务 — 补充

计划遗漏了一点：当前 `order_cron.go` 中 `disableExpiredLegacyOrders` **没有事务包裹**，也没有错误重试。除了加 `TenantId` 过滤外，还建议：
- 加事务保护。
- 加 `LIMIT` / 批量处理，避免一次性更新大量记录导致锁表。
- 加幂等保护（已停用的不再更新）。

### A.2.3 Phase 0.3 患者写操作 — 关键修正

计划中只提到"移除或绕开 `patient_basic_infos` 新表写入"，但实际影响远超预期：

1. **`patient_basic_service.go` 整个文件都基于新表**，不是简单绕开就能解决。需要评估是否有前端页面依赖该服务返回的数据。
2. **`patient_service.go:899` 硬删除关联基本信息** `s.db.Where("patient_id = ?", id).Delete(&models.PatientBasicInfo{})`——如果绕开新表，这条删除也要同步移除。
3. **`lis_sync_service.go`、`key_indicator_service.go`、`exam_report_sync_service.go`** 也引用 `PatientBasicInfo`，改动需要联动处理。
4. **`models/patient_basic_info.go` 已标注 DEPRECATED**，计划应明确是整个文件和对应 service 一起移除/重写，还是逐步迁移。

### A.2.4 Phase 2 新表残留 — 关键补充

计划中遗漏了以下关键细节：

1. **`dict_service.go` 存在新旧双路径**：`listNewSchemaDictTypes()` 查 `dict_types` 新表，`listLegacyCodeDictionaryTypes()` 查老库 `CodeDictionary_CodeDictionarys`。合并逻辑在 `ListTypes()` 中做 fallback。计划只说"统一到 CodeDictionary"，但没提 fallback 逻辑的处理。
2. **`orders` 和 `Order_PatientOrder` 并存**：`order_service.go` 中写操作（停用/复制/分组）使用 `Model(&models.Order{})` 操作 `orders` 新表，读操作（列表/详情）使用 `Table("Order_PatientOrder")` 读老库。这是**读写分离到不同表**的 bug，不是简单残留。
3. **`clinical_task_service.go` 完全基于新表**，且有 `ensureTables()` 空占位函数（曾是 AutoMigrate）。如果移除新表，该服务的写接口需要改为从现有业务数据派生或返回"暂不支持"。

### A.2.5 Phase 4 前端修复 — 补充

1. **`auth_token` → `hdis_access_token`** 问题不仅限于排班模板页面。还有 3 处直接读 `localStorage.getItem('hdis_access_token')` 绕过 `getToken()` 工具函数（`MaterialTab.tsx`、`DrugTab.tsx`、`BatchImportModal.tsx`），这些也应一并修复。
2. **OAuth 配置硬编码**：`authServer: 'https://auth.aihhd.com'` 和 `clientId: 'yanshi7779'` 应改为环境变量 `VITE_OAUTH_AUTH_SERVER` 和 `VITE_OAUTH_CLIENT_ID`。

### A.2.6 其他遗漏问题

以下问题在原计划中未列出但实际存在：

1. **SQL 日志安全隐患**：`database.go` 设置 `ParameterizedQueries: false`，会导致 GORM 将参数值直接拼入 SQL 日志，可能泄露患者敏感信息（身份证号、诊断等）。计划中未提及此问题。
2. **GORM 双引号一致性**：计划原则中提到"老库 SQL 表名和列名统一使用双引号"，但未明确要求对新建的 GORM 模型也统一使用双引号 tag（如 `gorm:"column:\"ColumnName\""`）。当前代码中部分新模型（如 `PatientBasicInfo`）使用小写下划线字段名，与老库大写驼峰的列名不一致。
3. **角色/权限体系结构性冲突**：`permission_service.go` 中列表/创建/更新/删除角色操作 `Identity_Roles`，但权限关联 `Authorization_RolePermissions.RoleId` JOIN `Authorization_Roles.Id`。两套角色表的 ID 可能碰撞（`Identity_Roles` 的 ADMIN Id 与 `Authorization_Roles` 中某行 Id 相同），导致权限分配串角色。这是 **P0 级 bug**，但原计划仅在"需要用户决策"中提到了角色表选择，未将其作为独立 bug 处理。

## A.3 新增建议条目

基于复核，建议在计划中新增以下条目：

### A-11（P0）角色权限关联的两表混用 bug

- 问题：`permission_service.go:65` 权限列表 JOIN `Authorization_Roles`，但 `permission_service.go:146-224` 角色增删改查操作 `Identity_Roles`。两表 Id 空间不同，角色 ID 可能碰撞。
- 修复：统一业务权限管理使用的角色来源。**不要直接把 `Authorization_RolePermissions.RoleId` 改去关联 `Identity_Roles.Id`**，因为当前老库权限关系实际指向 `Authorization_Roles.Id`。保守方案是：权限分配、权限树、应用角色管理统一使用 `Authorization_Roles`；`Identity_Roles` 暂保留为 ASP.NET Identity 登录角色来源之一。若后续要彻底合并两套角色，必须先做只读映射审计和用户确认，不能直接改关系表。
- 优先级：**P0**，因为当前存在创建角色后权限关联指向错误角色的风险。

### A-12（P1）医嘱服务新旧表并存读写

- 问题：`order_service.go` 读操作走 `Order_PatientOrder` 老库，写操作（停用/复制/分组）走 `orders` 新表。更新操作 `Model(&models.Order{})` 会写 `orders` 表，可能对患者医嘱数据产生幽灵记录。
- 修复：所有医嘱写操作统一到 `Order_PatientOrder`。
- 优先级：**P1**，因为新表 `orders` 在老库中不存在，写操作实际上会报错或创建不应存在的数据。

### A-13（P1）SQL 日志参数泄露

- 问题：`database.go:45` 设置 `ParameterizedQueries: false`，GORM 会将 SQL 参数值（包含患者身份证号、诊断等）直接写入日志。
- 修复：生产环境应改为 `ParameterizedQueries: true`，或加配置开关。
- 优先级：**P1**。

### A-14（P2）前端 3 处直接读取 localStorage 的 token 绕过

- 问题：`MaterialTab.tsx:179`、`DrugTab.tsx:312`、`BatchImportModal.tsx:220` 直接读取 `localStorage.getItem('hdis_access_token')`，绕过 `getToken()` 工具函数（未处理 token 过期/刷新逻辑）。
- 修复：统一使用 `getToken()` 函数。
- 优先级：**P2**，功能不受影响但维护风险高。

## A.4 实施顺序建议修正

原计划实施顺序为 Phase 0→1→2→3→4→5→6。基于复核，建议修正：

1. **Phase 0.1** 后端高危接口权限分组 — 保持
2. **Phase 0.2** 医嘱定时任务默认关闭 — 保持
3. **Phase 0.3** 患者删除/更新高危路径先收敛 — **需扩大范围**：同时处理 `order_service.go` 新旧表并存问题（A-12），因为医嘱写新表同样有数据破坏风险
4. **Phase 0.4（新增）** 角色权限两表混用修复（A-11）— 新增，因为这是当前运行的活跃 bug；执行时以 `Authorization_Roles` 作为业务权限角色源，不要把既有关联强行改到 `Identity_Roles`
5. **Phase 1** 租户配置化 — 保持
6. **Phase 2** 新表残留白名单和清理 — **补充**：明确解决 `dict_service.go` 双路径 fallback、`orders`/`Order_PatientOrder` 并存、`clinical_task_service.go` 功能替代
7. **Phase 3** 患者写操作专项重构 — 保持，但**必须联动** `patient_basic_service.go` 整体重写和 3 个联动服务的调整
8. **Phase 4** 前端路由权限和 API 调用修复 — 保持，**补充** 3 处 token 绕过修复
9. **Phase 5** 认证体系整理 — 保持，**补充** OAuth 硬编码改为环境变量
10. **Phase 6** 路由配置单一来源 — 保持

---

# 附录 B：二次复核与可执行开发计划

复核日期：2026-06-01

## B.1 二次复核结论

文档主体和附录 A 的大部分判断准确，可以作为后续改造依据。但必须按以下纠偏执行：

- `A-11` 的修复方向已更正：不要把 `Authorization_RolePermissions.RoleId` 无迁移地改去关联 `Identity_Roles.Id`。
- 当前代码事实是：权限关系查询和写入使用 `Authorization_Roles` + `Authorization_RolePermissions`，但应用角色 CRUD 使用 `Identity_Roles`。
- 已知数据库事实是：`Authorization_RolePermissions.RoleId` 与 `Authorization_Roles.Id` 对齐。
- 保守收敛方向是：业务权限管理统一到 `Authorization_Roles`；`Identity_Roles` 暂时保留为 ASP.NET Identity 登录角色来源之一。
- 后续开发 AI 必须先修 P0 安全项，再处理结构性重构，不允许一开始大范围重写。

## B.2 执行总原则

- 每次只执行一个任务包，完成后运行对应验证命令。
- 不执行 DDL，不恢复 `AutoMigrate`，不创建、删除、修改老库表结构。
- 不提交 `.env`、日志、构建产物、真实密钥。
- 不批量重写 `ai-hms-frontend/src/services/restClient.ts`，该文件存在非 UTF-8/GB2312 编码风险；如需改动，优先新增独立 API 模块。
- 字段含义不确定时，记录到 `docs/legacy-migration-uncertain-field-checklist.md`，不要猜字段写库。
- 对患者、医嘱、治疗、库存这类老库核心数据，宁可暂时禁用危险写操作，也不要写入未确认字段。

## B.3 Task P0-01：后端高危接口权限保护

目标：先阻断普通登录用户越权管理系统数据。

涉及文件：

- `ai-hms-backend/cmd/server/main.go`
- `ai-hms-backend/internal/middleware/auth.go`
- `ai-hms-backend/internal/services/auth_service.go`

执行步骤：

1. 在 `auth_service.go` 中把用户角色加载从“单个主角色”调整为“角色集合”。角色集合应合并 `Identity_UserRoles` + `Identity_Roles` 和 `Authorization_RoleUsers` + `Authorization_Roles`。
2. 角色集合需要去重、去空格，并保留 `ADMIN` 标准化逻辑。
3. 在 `middleware/auth.go` 中保留 `RequireRoles`，可新增 `RequireAnyRole` 或复用现有实现。
4. 在 `main.go` 中拆分路由组：`protected` 只要求登录，`admin` 要求 `ADMIN`、`管理员`、`安全管理员`、`运维管理员` 任一角色，`clinicalWrite` 要求 `ADMIN`、`管理员`、`主任`、`医生`、`护士长`、`护士` 任一角色。
5. 至少将 `/users/**`、`/app-roles/**`、`/app-permissions/tree`、`/permissions/**`、`/role-permissions/**`、`/hdis/settings/**`、`/logs/**` 迁入 `admin`。
6. 字典写操作、患者删除、医嘱写操作、库存写操作可以先按保守策略迁入 `admin` 或 `clinicalWrite`，不要保持裸 `protected`。

验收标准：

- 非管理员角色访问用户、角色、权限接口返回 403。
- 管理员角色可访问用户、角色、权限接口。
- `/me` 返回的 `roles` 包含用户所有角色，而不是仅一个主角色。

验证命令：

```powershell
cd ai-hms-backend
go test ./internal/services ./internal/api/v1 -count=1 -short
go build -o "$env:TEMP\ai-hms-backend-check.exe" ./cmd/server
```

## B.4 Task P0-02：医嘱定时任务默认关闭

目标：服务启动默认不改写老库医嘱。

涉及文件：

- `ai-hms-backend/config/*`
- `ai-hms-backend/cmd/server/main.go`
- `ai-hms-backend/internal/services/order_cron.go`

执行步骤：

1. 新增配置 `ORDER_CRON_ENABLED=false`。
2. `main.go` 中仅在配置为 true 时调用 `services.StartOrderCron()`。
3. `disableExpiredLegacyOrders()` 增加 `"TenantId" = ?` 过滤，租户值来自统一配置或传入参数。
4. 日志明确输出“order cron disabled”或“order cron enabled”。
5. 不做任何 DDL，不新增任务状态表。

验收标准：

- 默认配置启动服务不会调用 `StartOrderCron()`。
- 开启配置后，更新 `Order_PatientOrder` 的 SQL 带 `TenantId` 条件。

## B.5 Task P0-03：角色权限两表混用修复

目标：避免角色列表、权限分配操作不同角色表导致权限串角色。

涉及文件：

- `ai-hms-backend/internal/services/permission_service.go`
- `ai-hms-backend/internal/api/v1/role_management_handler.go`
- `ai-hms-frontend/src/services/role.ts`
- 角色管理页面相关文件

执行步骤：

1. 将 `PermissionService.ListRoles/CreateRole/UpdateRole/DeleteRole` 的业务角色数据源统一为 `Authorization_Roles`，不要继续在应用角色管理里操作 `Identity_Roles`。
2. `CreateRole` 写 `Authorization_Roles` 时必须补齐老库必填字段；如果字段含义不确定，先禁用新增/删除角色，仅允许读取和权限分配。
3. `SetRolePermissionCodes` 继续使用 `Authorization_RolePermissions.RoleId = Authorization_Roles.Id`。
4. 角色权限 API 入参建议从 `role name` 逐步改为 `role id` 或稳定 code；若短期保持 role name，必须按 `Authorization_Roles.Name` 查找。
5. `Identity_Roles` 暂时只用于登录身份角色读取和用户管理，不参与 `Authorization_RolePermissions`。

验收标准：

- `/app-roles` 返回 `Authorization_Roles` 中的业务角色，而不是只返回 `Identity_Roles` 的 ADMIN。
- 给某个业务角色保存权限后，`Authorization_RolePermissions.RoleId` 指向同名 `Authorization_Roles.Id`。
- 不出现 `Authorization_RolePermissions.RoleId = Identity_Roles.Id` 的写入路径。

## B.6 Task P0-04：患者高危写操作先收敛

目标：阻止患者硬删除和未确认字段写入老库。

涉及文件：

- `ai-hms-backend/internal/services/patient_service.go`
- `ai-hms-backend/internal/api/v1/patient_handler.go`
- 患者管理前端页面

执行步骤：

1. `PatientService.Delete` 先改为返回业务错误，例如 409：“老库患者暂不支持直接删除”。
2. 删除 `Delete(&models.PatientBasicInfo{})` 关联删除路径，避免继续操作 `patient_basic_infos` 新表。
3. `Update` 中移除 `bed_number`、`risk_level`、`patient_type` 等小写下划线新库字段更新。
4. `status` 若要保留，必须映射到老库真实列 `TreatmentStatus`；如语义不明确，先不写。
5. 前端删除按钮根据接口能力禁用或提示“不支持直接删除患者”。

验收标准：

- 全仓不再存在患者主表硬删除路径。
- 患者更新不再写 `bed_number`、`risk_level`、`patient_type` 小写字段。
- 患者删除接口不会改写老库数据。

## B.7 Task P1-01：租户配置化

目标：消除散落的 `legacyTenantID = 3`，但不改变业务行为。

涉及文件：

- `ai-hms-backend/config/*`
- `ai-hms-backend/internal/services/auth_service.go`
- 所有 grep 到 `legacyTenantID` 的 service

执行步骤：

1. 新增必填或显式配置 `LEGACY_TENANT_ID`。短期默认值可按当前部署继续使用 3，但代码中不能再散落硬编码常量。
2. 在统一位置提供 `LegacyTenantID()` 或配置注入方式。
3. `auth_service.go` 登录生成 JWT 时使用统一租户配置。
4. 逐个 service 替换 `legacyTenantID` 常量引用，不改变 SQL 条件语义。
5. 替换完成后删除 `auth_service.go` 中全局 `legacyTenantID` 常量，或仅保留在统一配置兼容层。

验收标准：

- `rg "legacyTenantID\s*int64\s*=\s*3|legacyTenantID\s*=\s*3" ai-hms-backend/internal` 无结果。
- 核心老库查询仍保留 `TenantId` 过滤。
- 登录 JWT 中 `tenant_id` 与配置一致。

## B.8 Task P1-02：医嘱服务新旧表并存修复

目标：所有医嘱读写统一到老库 `Order_PatientOrder`，不再写新表 `orders`。

涉及文件：

- `ai-hms-backend/internal/services/order_service.go`
- `ai-hms-backend/internal/models/treatment.go`
- `ai-hms-backend/internal/api/v1/order_handler.go`

执行步骤：

1. 审计 `order_service.go` 中所有 `models.Order{}`、`s.db.Model(&models.Order{})`、`s.db.Create(&copied)`、`s.db.First(&refreshed)` 路径。
2. `Update`、`Group`、`Ungroup`、`Stop`、`Copy`、`Revise`、`CreateFromTemplate` 必须全部改为操作 `Order_PatientOrder` 或明确返回“不支持”。
3. 不要继续使用 `models.Order.TableName() == "orders"` 承载老库写操作。
4. 字段映射必须使用老库字段：例如 `EndTime`、`IsDisabled`、`OrderGroup`、`Content`、`Dosage`、`UseMethod`、`UseWay`、`Note` 等，具体以老库结构文档和现有 `legacyPatientOrder` 为准。
5. 若某些前端语义如 `group_id`、`stop_reason` 老库无明确字段，先禁用该操作或记录待确认，不要写新表字段。

验收标准：

- `rg "Model\(&models\.Order\{\}\)|Create\(&copied\)|First\(&refreshed" ai-hms-backend/internal/services/order_service.go` 无新表写路径。
- 医嘱列表、创建仍正常。
- 不再通过 `orders` 表更新或创建医嘱。

## B.9 Task P1-03：新表残留白名单与隔离

目标：明确哪些新表路径必须删除、禁用或暂时保留。

涉及文件：

- `ai-hms-backend/internal/models/patient_basic_info.go`
- `ai-hms-backend/internal/services/patient_basic_service.go`
- `ai-hms-backend/internal/services/dict_service.go`
- `ai-hms-backend/internal/services/clinical_task_service.go`
- `ai-hms-backend/internal/services/hdis_settings_service.go`
- `ai-hms-backend/internal/models/treatment.go`

执行步骤：

1. 新建 `docs/legacy-table-usage-whitelist.md`，列出每个模块允许访问的老库表、读写属性、迁移状态。
2. `dict_service.go` 保留老库 `CodeDictionary_CodeDictionarys` 为主路径，移除或显式隔离 `dict_types/dict_items` fallback。
3. `clinical_task_service.go` 如没有老库权威表，写接口先返回“不支持”，读接口可从排班、治疗、医嘱派生。
4. `integration_hdis_settings` 如需继续使用，必须在文档中标注为“待用户批准的新配置存储”；否则改为环境变量或只读配置。
5. `patient_basic_infos` 不再作为患者基础信息权威来源，后续由 Phase 3 专项重写。

验收标准：

- 每个 `TableName()` 返回的新表名都能在白名单中找到处理结论。
- 不在白名单中的新表读写路径已删除、禁用或改为老库表。

## B.10 Task P1-04：SQL 日志参数保护

目标：降低日志泄露患者敏感信息的风险。

涉及文件：

- `ai-hms-backend/internal/database/database.go`
- `ai-hms-backend/config/*`

执行步骤：

1. 增加配置项控制 GORM `ParameterizedQueries`。
2. 生产环境默认 `ParameterizedQueries=true`。
3. 开发环境如需排查 SQL，可显式配置为 false。
4. 不改变现有日志级别配置语义。

验收标准：

- 默认配置不会把 SQL 参数值直接写入日志。
- `go vet ./...` 通过。

## B.11 Task P2-01：前端直接 fetch 和 token key 修复

目标：所有 API 请求走统一客户端和统一 token 工具。

涉及文件：

- `ai-hms-frontend/src/pages/ScheduleTemplateList.tsx`
- `ai-hms-frontend/src/components/schedule/ApplyTemplateModal.tsx`
- `ai-hms-frontend/src/pages/TreatmentConfig/tabs/MaterialTab.tsx`
- `ai-hms-frontend/src/pages/TreatmentConfig/tabs/DrugTab.tsx`
- `ai-hms-frontend/src/pages/TreatmentConfig/components/BatchImportModal.tsx`
- `ai-hms-frontend/src/services/*Api.ts`

执行步骤：

1. 新增或复用独立排班模板 API 模块，不要直接改大体量 `restClient.ts`。
2. `ScheduleTemplateList.tsx` 和 `ApplyTemplateModal.tsx` 移除 `fetch('/api/...')` 和 `auth_token`。
3. 修正模板接口路径为后端真实接口；当前候选为 `/api/v1/patient-shifts/templates`，执行前需核对 handler。
4. `MaterialTab.tsx`、`DrugTab.tsx`、`BatchImportModal.tsx` 改为通过统一 API 客户端或 `getToken()` 获取 token。

验收标准：

- `rg "fetch\('/api|fetch\(\"/api|auth_token" ai-hms-frontend/src` 无结果。
- 除统一 token 工具和 HTTP 客户端外，业务页面不直接读取 `localStorage.getItem('hdis_access_token')`。
- `npm run lint` 和 `npm run build` 通过。

## B.12 Task P2-02：前端路由权限守卫

目标：菜单隐藏和 URL 直达权限一致。

涉及文件：

- `ai-hms-frontend/src/components/AuthGuard.tsx`
- `ai-hms-frontend/src/router.tsx`
- `ai-hms-frontend/src/services/role.ts`
- `ai-hms-frontend/src/layouts/Sidebar.tsx`

执行步骤：

1. 新增 `PermissionGuard` 或增强 `AuthGuard`。
2. 根据当前 path 获取所需角色或权限码。
3. 未授权时显示 403 页面或跳转默认首页。
4. 先覆盖用户管理、角色管理、权限管理、系统设置等管理页面。

验收标准：

- 非管理员直接访问用户管理、角色管理、权限管理页面被拦截。
- 管理员访问正常。
- 菜单隐藏和 URL 直达规则一致。

## B.13 Task P2-03：认证入口整理

目标：明确生产环境登录入口，避免 OAuth、REST、本地预览三套逻辑互相干扰。

涉及文件：

- `ai-hms-frontend/src/services/auth.ts`
- `ai-hms-frontend/src/services/restAuth.ts`
- `ai-hms-frontend/src/components/AuthGuard.tsx`
- `ai-hms-frontend/src/pages/Login.tsx`

执行步骤：

1. 如果继续保留 OAuth，将 `authServer` 和 `clientId` 改为环境变量，不允许硬编码生产域名和 clientId。
2. 如果生产只用 REST 登录，将 OAuth 回调处理隔离到显式开关下。
3. 本地预览登录只允许在 `import.meta.env.DEV` 或 `VITE_ENABLE_LOCAL_PREVIEW_LOGIN=true` 时出现。
4. token 保存、读取、清理统一走 `utils/token.ts`。

验收标准：

- 生产构建中没有硬编码 OAuth 域名和 clientId。
- 登录、登出、token 检查只有一套主流程。

## B.14 Task P3-01：患者基础信息专项迁移

目标：用老库权威表替代 `patient_basic_infos`。

涉及文件：

- `ai-hms-backend/internal/services/patient_basic_service.go`
- `ai-hms-backend/internal/services/patient_service.go`
- `ai-hms-backend/internal/services/lis_sync_service.go`
- `ai-hms-backend/internal/services/key_indicator_service.go`
- `ai-hms-backend/internal/services/exam_report_sync_service.go`
- `ai-hms-backend/internal/models/patient_basic_info.go`

执行步骤：

1. 梳理前端患者基础信息页面字段，建立“前端字段 -> 老库表列 -> 是否可写”映射。
2. 读取优先使用 `Register_PatientInfomation`，必要时联查诊断、血管通路、计划、感染、病历等老库表。
3. 写入只允许已确认字段，未确认字段记录到待确认文档。
4. `patient_basic_service.go` 不再直接读写 `patient_basic_infos`。
5. 依赖 `PatientBasicInfo` 的同步服务需改为老库患者主键和真实字段。

验收标准：

- `rg "patient_basic_infos|PatientBasicInfo" ai-hms-backend/internal` 只剩已标注废弃、无运行路径的兼容代码，或完全无结果。
- 患者详情页基础信息可正常读取。
- 未确认字段不会写库。

## B.15 Task P3-02：路由、菜单、权限配置单一来源

目标：减少孤儿页面和权限配置漂移。

涉及文件：

- `ai-hms-frontend/src/router.tsx`
- `ai-hms-frontend/src/layouts/Sidebar.tsx`
- `ai-hms-frontend/src/layouts/routeMeta.ts`
- `ai-hms-frontend/src/services/role.ts`

执行步骤：

1. 新增统一路由配置，例如 `src/router/routeConfig.tsx`。
2. 每个路由配置包含 `path`、`element`、`title`、`menuKey`、`requiredRoles` 或 `permissionCode`、`hiddenInMenu`。
3. 由配置生成路由、菜单、面包屑和权限判断。
4. 先覆盖管理类页面和排班模板页面，不必一次性迁移所有页面。

验收标准：

- 新增页面只需维护一处路由元数据。
- 排班模板页面不再是孤儿路由。

## B.16 总验证清单

每完成一个后端任务包，运行：

```powershell
cd ai-hms-backend
go test ./internal/services ./internal/api/v1 -count=1 -short
go vet ./...
go build -o "$env:TEMP\ai-hms-backend-check.exe" ./cmd/server
```

每完成一个前端任务包，运行：

```powershell
cd ai-hms-frontend
npm run lint
npm run build
```

最终静态检查：

```powershell
rg "AutoMigrate|DropTable|CreateTable|ALTER TABLE|DROP TABLE" ai-hms-backend
rg "legacyTenantID\s*=\s*3|legacyTenantID\s+int64\s*=\s*3" ai-hms-backend/internal
rg "Model\(&models\.Order\{\}\)|patient_basic_infos|dict_types|dict_items|clinical_tasks|integration_hdis_settings" ai-hms-backend/internal
rg "fetch\('/api|fetch\(\"/api|auth_token" ai-hms-frontend/src
```

## B.17 给开发 AI 的建议执行顺序

1. `Task P0-01` 后端高危接口权限保护。
2. `Task P0-02` 医嘱定时任务默认关闭。
3. `Task P0-03` 角色权限两表混用修复。
4. `Task P0-04` 患者高危写操作先收敛。
5. `Task P1-01` 租户配置化。
6. `Task P1-02` 医嘱服务新旧表并存修复。
7. `Task P1-03` 新表残留白名单与隔离。
8. `Task P1-04` SQL 日志参数保护。
9. `Task P2-01` 前端直接 fetch 和 token key 修复。
10. `Task P2-02` 前端路由权限守卫。
11. `Task P2-03` 认证入口整理。
12. `Task P3-01` 患者基础信息专项迁移。
13. `Task P3-02` 路由、菜单、权限配置单一来源。

## B.18 交付格式要求

后续开发 AI 每完成一个任务包，应返回：

- 修改文件清单。
- 关键行为变化。
- 未处理或需要用户确认的字段/接口。
- 验证命令和结果。
- 是否触及老库写操作。
