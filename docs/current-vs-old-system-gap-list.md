# 当前系统与 `old_system` 截图差异核查清单

核查日期：2026-05-31
修复日期：2026-05-31

## 修复状态总览

| 编号 | 问题 | 状态 |
|------|------|------|
| P0-1 | 透析小结可编辑保存 | **已修复** |
| P0-2 | 机器消毒登记接通 | **已修复** |
| P0-3 | 床位/病区后端路由 | **已修复** |
| P1-4 | 床位FEP/采集连接字段 | **已修复** |
| P1-5 | 病区负责医护多选 | **已修复** |
| P1-6 | 宣教管理CRUD | **已修复** |
| P1-7 | 医嘱模板组隐藏 | **已修复** |
| P1-8 | 处方动态调整记录 | 待确认 → `docs/uncertain-items-2026-05-31.md` |
| P2-9 | 用户多角色 | **已修复** |
| P2-10 | 角色停用 | 待确认 → `docs/uncertain-items-2026-05-31.md` |
| P2-11 | 字典老库来源控制 | **已修复** |
| P2-12 | 同日多治疗 | 待确认 → `docs/uncertain-items-2026-05-31.md` |

验证结果：
- 后端 `go build ./cmd/server` ✅ 通过
- 前端 `npm run lint` ✅ 0 errors, 2 warnings
- 前端 `npm run build` ✅ 通过

核查范围：
- 参考截图：`old_system/` 下透析执行 2.1-2.9、用户管理、角色管理、床位管理、病区管理、字典配置截图。
- 当前实现：`ai-hms-frontend/src/pages/**`、`ai-hms-frontend/src/services/**`、部分后端路由注册。
- 验证：已运行 `ai-hms-frontend npm run lint`，0 errors，2 warnings。

说明：本次是基于截图与代码的核查，没有启动浏览器逐像素截图比对。若某处属于你们有意改版，应以业务确认结果为准。

## P0 必须优先修复

### 1. 透析小结页面截图要求可编辑保存，当前代码仍是只读/禁用

参考截图：`old_system/2.9透析执行—透析小结/主界面.png`

当前问题：
- 截图中“医生小结”“治疗/护理小结”为可输入文本域，“保存小结”按钮可点击。
- 当前 `DialysisSummary.tsx` 中两个 textarea 都是 `readOnly`，保存按钮 `disabled` 且提示“功能待后端接口就绪”。
- 前端 `restClient.ts` 已有 `updateTreatmentSummary(id, data)`，后端也已有 `/api/v1/treatments/:id/summary`，但页面没有接入。

涉及文件：
- `ai-hms-frontend/src/pages/dialysis-processing/execution/DialysisSummary.tsx`
- `ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx`
- `ai-hms-frontend/src/services/restClient.ts`

解决方案：
- 在 `DialysisSummary` 中增加本地表单状态：`doctorSummary`、`treatmentSummary`。
- 当 `treatment` 变化时同步初值，避免切换患者残留旧数据。
- 点击“保存小结”调用 `restApi.updateTreatmentSummary(treatment.id, { doctorSummary, treatmentSummary })`。
- 保存成功后刷新当前治疗记录，或由父组件传入 `onTreatmentUpdated(updatedTreatment)` 更新 `currentTreatment`。
- 移除 textarea 的 `readOnly` 和按钮的禁用状态；仅在 `treatmentLoading`、无 `treatment`、保存中时禁用。

验收标准：
- 进入“透析小结”后可编辑两个小结字段。
- 点击保存后页面提示成功，刷新后内容仍存在。
- 切换患者时不会显示上一位患者的小结内容。

### 2. 双人核对页机器消毒登记按钮仍未接通

参考截图：`old_system/2.3透析执行—双人核对/主界面.png`

当前问题：
- 截图中机器消毒登记卡片应可“确认并保存消毒登记”。
- 当前 `Verification.tsx` 的按钮仍然禁用，文案为“功能待后端接口就绪”。
- 前端 `restClient.ts` 已有 `saveTreatmentDisinfection(treatmentId, data)`，后端也已新增 `/api/v1/treatments/:id/disinfection`，但页面未调用。
- 当前登记人 fallback 使用 `patient.name`，这不符合截图里的“登记人 test_admin”语义，消毒登记人应来自当前登录用户或选中员工。

涉及文件：
- `ai-hms-frontend/src/pages/dialysis-processing/execution/Verification.tsx`
- `ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx`
- `ai-hms-frontend/src/services/restClient.ts`

解决方案：
- 给 `Verification` 增加 `onSaveDisinfection` prop，或在组件内直接调用 `restApi.saveTreatmentDisinfection`。
- 构造 payload：`type: '机器消毒'`、`disinfectant: '500mg/L含氯消毒液'`、`startTime: disinfectionTime`、`disinfectUserId: currentUserId`、必要时补 `description/note`。
- 将登记人展示改为 `currentUser?.name || staffNameById.get(currentUserId) || '--'`，不要 fallback 到患者姓名。
- 保存成功后提示“消毒登记已保存”，并刷新治疗记录或更新局部状态。

验收标准：
- 机器消毒登记按钮可点击。
- 成功保存到后端后无“功能待后端接口就绪”文案。
- 登记人显示当前登录人或当前护士，不显示患者姓名。

### 3. 床位管理、病区管理前端已完成，但后端路由缺失

参考截图：
- `old_system/床位管理/主界面.png`
- `old_system/床位管理/编辑床位.png`
- `old_system/病区管理/主界面.png`
- `old_system/病区管理/床位编辑.png`

当前问题：
- 前端调用 `/api/v1/beds` 和 `/api/v1/wards`。
- 后端代码中未找到 `RegisterBedRoutes`、`RegisterWardRoutes`、`/beds`、`/wards` 路由注册。
- 页面 UI 会加载失败或只能显示空数据，无法达到截图中的真实列表效果。

涉及文件：
- `ai-hms-frontend/src/pages/BedManagement.tsx`
- `ai-hms-frontend/src/pages/WardManagement.tsx`
- `ai-hms-frontend/src/services/managementApi.ts`
- `ai-hms-backend/internal/api/v1/*`
- `ai-hms-backend/internal/services/*`
- `ai-hms-backend/cmd/server/main.go`

解决方案：
- 后端新增 Ward service/handler：列表、新增、编辑、删除或禁用。
- 后端新增 Bed service/handler：列表、新增、编辑、删除或禁用。
- 使用老库对应表，字段不确定时先查 `老血透数据库表结构-合并版.md` 和数据库实际 `information_schema.columns`。
- 路由注册到 `cmd/server/main.go` 的 protected API group。
- 前端不要改路径，保持 `/api/v1/wards`、`/api/v1/beds`。

验收标准：
- 病区管理能展示截图中的列：名称、患者类型、传染病、负责医护、床位数、状态、操作。
- 床位管理能展示截图中的列：名称、所属病区、默认设备、设备数、状态、操作。
- 新增/编辑保存后列表刷新。

## P1 高优先级修复

### 4. 床位编辑弹窗字段少于截图要求

参考截图：`old_system/床位管理/编辑床位.png`

当前问题：
- 当前 `BedManagement.tsx` 编辑弹窗仅包含：床位名称、所属病区、排序、备注、是否禁用。
- `managementApi.ts` 类型已预留 `fepId`、`acquisiteConnectId`、`equipments`，说明目标应包含采集连接/FEP/设备绑定等信息。
- 截图主界面有“默认设备”“设备数”，但编辑弹窗无法配置设备，因此主界面数据不可维护。

解决方案：
- 在床位编辑弹窗补充设备相关字段：默认设备、设备绑定列表、采集连接/FEP。
- 若老系统是单床多设备，则设备列表应支持添加/删除并标记默认设备。
- 后端 Bed payload 接收并保存 `equipments`、`fepId`、`acquisiteConnectId`。

需要确认：
- 老库设备绑定关系表名和字段是否以 `BedId/EquipmentId/IsDefault` 存储，需查库确认。

### 5. 病区编辑弹窗的“负责医护”不应是自由文本

参考截图：`old_system/病区管理/主界面.png`

当前问题：
- 主界面展示“负责医护”为多人姓名列表。
- 当前 `WardManagement.tsx` 中 `responsibleUsers` 是普通输入框，容易写入姓名文本，无法稳定关联用户 ID。

解决方案：
- 将“责任人员/负责医护”改为多选用户下拉，数据源使用 `userApi.getList({ status: 'active' })`。
- 前端 payload 使用用户 ID 列表或逗号字符串，后端按老库字段真实格式保存。
- 列表展示 `responsibleUserNames`，避免直接展示 ID。

需要确认：
- 老库病区责任医护字段是单字段字符串、关联表，还是 JSON/逗号 ID。

### 6. 宣教管理页面与截图级别差异大，且没有增删改

参考资料：`old_system/history/3.1宣教管理/1.宣教管理.har`，以及透析执行健康宣教截图。

当前问题：
- `EducationManagement.tsx` 只有 44 行，仅列表 + 刷新。
- 页面用原生 `fetch('/api/v1/health-educations')`，未复用 `educationManagementApi`。
- 没有新增、编辑、禁用/删除、内容描述、类型/分类、排序、附件等能力。
- `managementApi.ts` 已定义 `educationManagementApi.create/update/remove`，但后端目前只注册了 `GET /health-educations`，没有管理端 POST/PUT/DELETE。

解决方案：
- 前端改用 `educationManagementApi`。
- 重构宣教管理为与其他主数据页面一致的列表 + 新增/编辑弹窗。
- 后端补齐 `POST /health-educations`、`PUT /health-educations/:id`、`DELETE /health-educations/:id` 或状态禁用接口。
- 字段建议：标题/名称、类型、分类、内容描述、排序、附件、状态、备注。

验收标准：
- 宣教管理可以维护 `Auxiliary_HealthEducation` 内容。
- 透析执行健康宣教页的题目下拉能立即使用管理端新增的内容。

### 7. 医嘱弹窗“从模板组调取”仍是占位功能

参考截图：`old_system/2.5透析执行—透析医嘱/弹窗界面新增医嘱.png`

当前问题：
- 弹窗有“直接新增录入 / 从模板组调取”两个 tab。
- 当前 `MedicalOrders.tsx` 中选择模板 tab 后显示“功能待后端接口就绪”，保存也直接拦截。

解决方案：
- 若老系统 HAR 中已有模板组接口，应补 `orderTemplateApi`，在 tab 中展示模板组、模板医嘱明细，支持勾选导入。
- 若模板组暂不迁移，应隐藏“从模板组调取”tab，避免展示不可用入口。

需要确认：
- 这次修复目标是否包含医嘱模板组调取。如果不包含，建议直接隐藏入口而不是保留占位。

### 8. 当日处方动态调整记录固定暂无数据

参考截图：`old_system/2.2透析执行—当日处方/当日处方 主界面.png`

当前问题：
- 当前 `TodayPrescription.tsx` 中“当日处方动态调整记录”表格固定渲染“暂无数据”。
- 如果老系统要求记录处方修改历史，则当前页面无法反映真实调整记录。

解决方案：
- 后端提供处方变更历史接口，例如 `GET /patients/:id/prescriptions/:prescriptionId/change-logs`。
- 编辑保存处方时写入调整记录，记录调整内容、调整人、调整时间。
- 前端用接口数据替换固定空行。

需要确认：
- 老库是否有处方调整日志表；如果没有，需要决定是否只展示当前无数据。

## P2 中优先级修复

### 9. 用户管理界面细节与截图不完全一致

参考截图：
- `old_system/用户管理/主界面.png`
- `old_system/用户管理/用户编辑.png`

当前问题：
- 截图标题为“人员管理”，当前页面标题也为“人员管理”，但代码路径仍是 `UserManagement.tsx`，可接受。
- 截图状态列显示带“启用”文字的开关/标签；当前 Ant `Switch` 只显示开关，不显示文字。
- 截图角色分配为多选标签；当前 form 字段是单选 `role`。
- 后端 `UserService.SetUserRoles` 已支持角色数组，但当前页面编辑弹窗只提交单角色。

解决方案：
- 将角色字段改为 `Select mode="multiple"`，提交时调用 `userApi.setRoles(userId, roleIds)` 或 create/update 后同步角色。
- 状态列 Switch 加 `checkedChildren="启用"`、`unCheckedChildren="停用"`，或按截图用标签化展示。
- 编辑回显使用 `record.roles` 多角色数组。

### 10. 角色管理状态/停用语义与截图不一致

参考截图：
- `old_system/角色管理/主界面.png`
- `old_system/角色管理/编辑界面.png`
- `old_system/角色管理/权限分配界面.png`

当前问题：
- 截图操作列是“编辑 / 权限 / 停用”，当前 `RoleManagement.tsx` 是“编辑 / 权限 / 删除”。
- 截图状态列显示 `active`，当前代码显示“已启用/已禁用”Tag。
- 后端 `DeleteRole` 可能直接删除角色，风险高；老系统更像停用。

解决方案：
- 将删除操作改为停用/启用，前端按钮文案改“停用”。
- 后端新增或调整角色状态字段映射，优先软停用，不做物理删除。
- 若老库角色表没有状态字段，需要确认是否继续显示 `active` 常量，或用删除替代停用。

需要确认：
- `Authorization_Roles` 是否存在状态/禁用字段；此前查库只有 5 列，可能无状态列。

### 11. 字典配置使用了可维护按钮，但老库来源实际只读语义不清

参考截图：
- `old_system/字典配置/透析治疗.png`
- `old_system/字典配置/其他字典.png`

当前问题：
- 截图中老库来源行的编辑、查看、删除图标是灰色弱化状态。
- 当前 `DictConfig.tsx` 有“老库来源，当前仅支持只读查看”标签，但仍渲染新增按钮和操作入口，需要确认是否实际 disabled。

解决方案：
- 老库来源类型：禁用新增、编辑、删除，仅允许查看。
- 本地维护类型：允许完整增删改启停。
- 顶部“新增字典值”按钮应根据 `type.source === 'legacy'` disabled，并给出 tooltip。

### 12. 透析执行左侧患者列表出现重复患者时缺少排班维度区分

参考截图：多张透析执行截图中同一患者可能重复出现，例如“李大军 300408”多次。

当前问题：
- 当前 `DialysisExecution.tsx` 以 `patient.id` 作为选中 ID。
- 如果列表中同一患者对应多条排班/治疗记录，单用患者 ID 无法区分不同床位/班次/治疗记录。

解决方案：
- 患者列表数据应以排班记录或治疗记录为粒度，row key 使用 `scheduleId` 或 `treatmentId`。
- 当前 `getPatientTreatmentByDate(patientId, date)` 只能取当天一条治疗记录；如果同日多次透析，需要补班次/床位参数。

需要确认：
- 当前业务是否允许同一患者同日多条排班/治疗记录。

## 已确认通过或基本一致

- 透前评估整体布局、字段组、底部提交区与截图接近。
- 当日处方主卡片、基础参数、抗凝方案、材料清单结构与截图接近。
- 透后评估主布局与截图接近。
- 健康宣教执行页布局、宣教题目、详情、历史、保存按钮与截图接近。
- 角色权限树弹窗整体与截图接近。
- `npm run lint` 当前 0 errors，仅 `PostAssessment.tsx` 有 2 个 hooks 依赖 warning。

## 建议执行顺序

1. 先接通透析小结保存和机器消毒登记，这是已有前后端 API 但页面未调用，风险最低、收益最高。
2. 补齐 `/wards` 和 `/beds` 后端路由，否则病区/床位页面无法真实使用。
3. 完善宣教管理 CRUD，保证透析执行健康宣教的数据源可维护。
4. 再处理用户多角色、角色停用、医嘱模板组、处方调整日志等业务语义问题。

## 需要你确认的问题

1. 医嘱“从模板组调取”这次是否必须实现？如果不是，建议隐藏 tab。
2. 角色管理“停用”是否要真实落库？如果老库无状态字段，需要确定替代方案。
3. 病区负责医护、床位设备绑定的老库关系表字段是否已有明确映射？如果没有，需要先查库确认。
4. 同一患者同日是否允许多条排班/治疗记录？这会影响透析执行左侧列表的 row key 和治疗记录加载接口。
