# 旧系统前端界面还原执行计划

## 1. 用户原始要求

系统升级后导致当前前端界面出现错乱，需要尽量还原到升级前的旧系统界面和功能状态。

旧系统目前仍可访问，信息如下：

- 旧系统地址：`http://10.10.8.84:3000`
- 用户名：`test_admin`
- 密码：`Test@123456`
- 可通过本机浏览器访问旧系统进行核对。

旧系统的每个界面截图和 HAR 信息已经保存在本地路径：

- `F:\python\前后端代码\ai-hms_qhd\old_system`

处理要求：

- `old_system/history` 文件夹不用处理。
- `old_system` 下其他目录均视为旧系统备份资料。
- 需要综合旧系统在线页面、旧系统截图、HAR 信息和当前系统代码，对当前系统界面进行分析并还原。
- 目标是让当前系统与之前的界面和功能尽量一致。
- 如果发现明显错误且可以修复，可以一并修改。
- 本文档用于交给其他 AI 执行，因此步骤必须尽量详细，避免执行方误判。

## 2. 当前仓库与前端结构

仓库根目录：

- `F:\python\前后端代码\ai-hms_qhd`

当前主应用只关注：

- 后端：`ai-hms-backend/`
- 前端：`ai-hms-frontend/`

以下目录仅作为参考或素材，不是当前主应用入口：

- `ai-hms-v1.3-透析执行/`
- `gorm/`
- `old_system/`

当前前端技术栈：

- React 19
- Vite
- TypeScript
- Tailwind CSS v4
- Ant Design 6
- React Router

当前前端关键文件：

- 前端入口：`ai-hms-frontend/src/App.tsx`
- 路由：`ai-hms-frontend/src/router.tsx`
- 全局样式：`ai-hms-frontend/src/index.css`
- 主布局：`ai-hms-frontend/src/layouts/MainLayout.tsx`
- 侧边栏：`ai-hms-frontend/src/layouts/Sidebar.tsx`
- 顶栏：`ai-hms-frontend/src/layouts/Header.tsx`
- 页面目录：`ai-hms-frontend/src/pages`
- API facade：`ai-hms-frontend/src/services/restClient.ts`
- 服务统一导出：`ai-hms-frontend/src/services/index.ts`

前端验证命令：

```bash
cd ai-hms-frontend
npm run lint
npm run build
```

注意：`npm run build` 已包含 TypeScript 编译检查。

## 3. 旧系统资料范围

需要处理的旧系统目录如下：

- `old_system/2.1透析执行—透前评估`
- `old_system/2.2透析执行—当日处方`
- `old_system/2.3透析执行—双人核对`
- `old_system/2.5透析执行—透析医嘱`
- `old_system/2.6透析执行—透中监测`
- `old_system/2.7透析执行—透后评估`
- `old_system/2.8透析执行—健康宣教`
- `old_system/2.9透析执行—透析小结`
- `old_system/病区管理`
- `old_system/床位管理`
- `old_system/用户管理`
- `old_system/角色管理`
- `old_system/字典配置`

明确不处理：

- `old_system/history`

每个有效目录通常包含：

- 一个或多个 `.png` 截图。
- 一个 `.har` 文件。

执行时应优先使用截图确认视觉和布局，使用 HAR 确认接口、字段、请求方式、响应结构和交互行为。

### 3.1 明确缺口

旧系统备份中存在从 `2.3透析执行—双人核对` 直接跳到 `2.5透析执行—透析医嘱` 的情况，当前本地备份没有 `2.4` 目录。

处理规则：

- 本次默认不处理 `2.4`，因为本地无截图和 HAR 依据。
- 如果在线旧系统中确认存在对应 `2.4` 页面，以在线旧系统为准，先记录页面名称、截图、接口信息，再决定是否纳入本轮。
- 不要凭猜测新增 `2.4` 页面或 tab。

### 3.2 不在本次还原范围内的页面

以下当前前端页面没有对应的 `old_system` 备份，默认保持现状，不在本次还原范围内。如果执行过程中发现这些页面也有明显错乱，只记录问题，不要顺手修改，除非用户明确追加范围。

| 路径 | 当前页面 | 状态 |
| --- | --- | --- |
| `/dashboard` | `Dashboard` | 无旧备份，默认不处理 |
| `/patients`、`/patients/:id` | `PatientList`、`PatientDetail` | 无旧备份，默认不处理 |
| `/monitoring` | `Monitoring` | 无旧备份，默认不处理 |
| `/schedule` | `Schedule` | 无旧备份，默认不处理 |
| `/statistics` | `Statistics` | 无旧备份，默认不处理 |
| `/settings` | `Settings` | 无旧备份，默认不处理 |
| `/ward-overview` | `WardOverview` | 无旧备份，默认不处理 |
| `/login`、`/role-select` | `Login`、`RoleSelect` | 无旧备份，默认不处理；但若无法登录当前系统，应先记录阻塞 |
| `/inventory` | `Inventory` | 无旧备份，默认不处理 |
| `/device-binding` | `DeviceManagement` | 无旧备份，默认不处理 |
| `/master-data` | `MasterData` | 无旧备份，默认不处理 |
| `/treatment-config` | `TreatmentConfig` | 无旧备份，默认不处理 |
| `/education-management` | `EducationManagement` | 无旧备份，默认不处理 |
| `/schedule-templates`、`/schedule-templates/edit` | `ScheduleTemplateList`、`ScheduleTemplateEditor` | 无旧备份，默认不处理 |

透析执行目录中存在 `Disinfection.tsx`、`Consumables.tsx` 等当前代码文件，但旧系统备份中无对应截图和 HAR。本次不要主动把它们加入旧系统 tab，也不要删除它们；仅在当前路由实际引用且造成错误时再单独处理。

## 4. 还原优先级

建议按以下顺序执行，避免先改具体页面后被公共布局覆盖：

1. 还原公共布局：侧边栏、顶栏、主内容区、全局样式。
2. 还原透析执行主框架：患者列表、标签页、患者摘要、内容滚动结构。
3. 逐个还原透析执行 2.1 到 2.9 子页面。
4. 还原用户管理。
5. 还原角色管理和权限分配弹窗。
6. 还原字典配置。
7. 还原病区管理。
8. 还原床位管理。
9. 全量验证和问题收敛。

## 5. 总体执行原则

- 每次只迁移或还原一个页面、一个模块或一个公共组件。
- 不要一次性大范围重写整个前端。
- 不要改动与本任务无关的文件。
- 不要删除当前已有 API 调用和权限逻辑，除非确认其导致页面错误。
- 默认保持当前前端 API 路径、响应结构、字段名和交互不变。
- 如果 HAR 中旧字段与当前字段不一致，优先在前端做兼容映射。
- 如果确认是后端接口缺失或明显错误，再考虑改后端。
- 不要恢复后端自动迁移、seed、DDL 或默认账号兜底。
- 不要提交 `.env*`、密钥、HAR、日志、二进制、`dist/` 等产物。
- 不要把旧系统截图复制进前端静态资源作为 UI 替代。
- 不要用假数据掩盖真实接口问题，除非页面原本就是静态配置页。
- 如字段含义不确定，记录到 `docs/legacy-migration-uncertain-field-checklist.md`。
- 前端代码必须通过 `npm run lint` 和 `npm run build`。
- 必须先阅读仓库根目录 `AGENTS.md`，其中包含后端禁止 DDL、字典字段、前端 lint、构建和 Git 约束。
- 修改菜单、页面标题、分组名等文案时，如果当前组件使用 `i18next` 的 `t('xxx')`，优先修改 `src/i18n/locales/zh-CN/*.json` 中对应文案，不要在组件中重复硬编码。
- 涉及新增 API facade 时，优先检查 `src/services/restClient.ts`、`src/services/index.ts` 和相关拆分 service，保持现有导出风格。
- 涉及字典/下拉字段时，除 `src/services/dictApi.ts` 外，还要检查 `src/hooks/useDictOptions.ts` 和 `src/hooks/useDictName.ts`，优先复用已有 hook。
- 如不清楚 Tailwind 或样式类限制，先阅读 `ai-hms-frontend/eslint.config.js`。

## 6. 旧系统整体视觉特征

根据旧截图，旧系统整体视觉如下：

- 左侧是深色固定侧边栏。
- 顶部是浅蓝灰色顶栏。
- 主内容区是浅灰白背景。
- 品牌文案为 `AI-HMS 智能透析`。
- 左侧菜单激活态为亮蓝色块，白色文字。
- 菜单分组包含：`日常工作`、`患者中心`、`资源管理`、`配置中心`、`系统管理`。
- 顶栏左侧有折叠菜单按钮。
- 顶栏左侧有科室胶囊，文案类似 `当前科室：肾内透析中心 · 第一病区`。
- 顶栏右侧有任务或日历图标、用户名、角色、头像、退出图标。
- 页面标题使用黑色粗体。
- 表格页面整体简洁，不要过度卡片化。
- 搜索框和新增按钮通常位于页面右上角。
- 主按钮为亮蓝色。
- 删除操作为红色。
- 状态标签通常为浅蓝、浅绿或浅红底色。
- 弹窗为白底、圆角、灰色遮罩。

建议统一的视觉变量：

- 主蓝色：先用浏览器取色器从在线旧系统确认，确认前临时参考 `#1d63ff`。
- 侧栏背景：先用浏览器取色器从在线旧系统确认，确认前临时参考 `#0f172a`。
- 页面背景：先用浏览器取色器从在线旧系统确认，确认前临时参考 `#f5f7fb`。
- 主文字：`#0f172a`。
- 次级文字：`#64748b`。
- 边框：`#e5eaf2`。

颜色落地规则：

- 不要在每个页面零散覆盖颜色。
- 优先修改 `src/index.css` 中的 CSS 变量和 `src/App.tsx` 中的 Ant Design `theme.token`。
- 在线旧系统可访问时，用浏览器 DevTools 取色后记录固定值，再替换临时参考值。

注意：项目 ESLint 限制新增裸 `rounded-xl/2xl/3xl`、`text-[10px]/text-[11px]`、`!important`。需要新增圆角或密排样式时，优先使用已有设计 token 或在全局 CSS 中封装语义类，并确认 lint 允许。

## 7. 页面映射表

| 旧系统资料 | 当前前端文件 | 目标 |
| --- | --- | --- |
| `2.1透析执行—透前评估` | `src/pages/dialysis-processing/execution/PreAssessment.tsx` | 还原透前评估表单、底部信息条和提交区 |
| `2.2透析执行—当日处方` | `src/pages/dialysis-processing/execution/TodayPrescription.tsx` | 还原当日处方展示 |
| `2.3透析执行—双人核对` | `src/pages/dialysis-processing/execution/Verification.tsx` | 还原首次核对和二次核对 |
| `2.5透析执行—透析医嘱` | `src/pages/dialysis-processing/execution/MedicalOrders.tsx` | 还原医嘱列表和新增医嘱弹窗 |
| `2.6透析执行—透中监测` | `src/pages/dialysis-processing/execution/MidMonitoring.tsx` | 还原 KPI、监测流水表格、新增监测点弹窗 |
| `2.7透析执行—透后评估` | `src/pages/dialysis-processing/execution/PostAssessment.tsx` | 还原透后评估表单 |
| `2.8透析执行—健康宣教` | `src/pages/dialysis-processing/execution/HealthEducation.tsx` | 还原健康宣教记录 |
| `2.9透析执行—透析小结` | `src/pages/dialysis-processing/execution/DialysisSummary.tsx` | 还原透析小结 |
| `用户管理` | `src/pages/UserManagement.tsx` | 还原人员管理列表和编辑弹窗 |
| `角色管理` | `src/pages/RoleManagement.tsx` | 还原角色列表、编辑、权限分配 |
| `字典配置` | `src/pages/DictConfig.tsx` | 还原字典分类和字典项维护 |
| `病区管理` | `src/pages/WardManagement.tsx` | 还原病区列表和床位编辑 |
| `床位管理` | `src/pages/BedManagement.tsx` | 还原床位列表和编辑弹窗 |

透析执行总框架文件：

- `src/pages/dialysis-processing/DialysisExecution.tsx`
- `src/pages/dialysis-processing/components/PatientListSidebar.tsx`
- `src/pages/dialysis-processing/components/PatientSummaryHeader.tsx`
- `src/pages/dialysis-processing/types.ts`

## 8. 第一阶段：准备和对照

### 8.1 必读和分支准备

执行前必须先阅读：

- `AGENTS.md`
- `ai-hms-frontend/AGENTS.md`，如果存在。
- `ai-hms-frontend/eslint.config.js`

执行前先运行：

```bash
git status
```

目的：

- 确认当前工作区是否已有别人修改。
- 不要回滚、覆盖或删除他人修改。

建议新建工作分支，方便隔离和回滚：

```bash
git checkout -b fix/legacy-ui-restore
```

如果分支已存在，改用其他清晰名称，例如：

```bash
git checkout -b fix/legacy-ui-restore-2
```

不要主动 commit 或 push，除非用户明确要求。

确认前端依赖可用：

```bash
cd ai-hms-frontend
npm install
```

如果依赖已经完整安装，`npm install` 不应改变业务代码；若 `package-lock.json` 发生非预期变化，记录后再决定是否保留。

### 8.2 第 0 阶段：诊断当前错乱

在修改任何文件前，先启动当前前端，记录当前错乱表现：

```bash
cd ai-hms-frontend
npm run dev
```

诊断要求：

1. 打开当前系统本地地址。
2. 尽量登录当前系统；如果无法登录，先记录阻塞，不要绕过认证逻辑。
3. 按 `old_system` 备份范围逐个访问对应页面。
4. 对照旧系统截图和在线旧系统，记录差异。
5. 差异必须写具体，不要只写“错乱”。

差异记录示例：

- 页面：透析执行 - 透前评估。
- 旧系统表现：左侧患者列表宽约 240px，选中项浅蓝背景，主区域顶部固定 tab。
- 当前表现：患者列表宽度过大，主区域出现双滚动条，底部提交按钮被遮挡。
- 初步判断：布局高度和滚动容器冲突。
- 计划处理文件：`DialysisExecution.tsx`、`PatientListSidebar.tsx`。

没有诊断记录不要直接进入代码修改。

### 8.3 建立对照清单

然后建立对照清单：

1. 遍历 `old_system`，忽略 `history`。
2. 记录每个目录下的截图和 HAR。
3. 记录当前前端对应页面文件。
4. 每个页面开始修改前，先看截图，再看 HAR，再看当前代码。
5. 修改完成后记录已还原、未还原、原因。

建议在执行过程中维护状态：

- 未开始
- 分析中
- 修改中
- 已验证
- 有阻塞

### 8.4 阶段性验证要求

不要等全部页面改完才验证。每完成以下任一阶段后，都应执行前端验证：

- 公共布局还原完成。
- 透析执行总框架还原完成。
- 每完成 1 个透析执行 tab。
- 用户管理完成。
- 角色管理完成。
- 字典配置完成。
- 病区管理完成。
- 床位管理完成。

验证命令：

```bash
cd ai-hms-frontend
npm run lint
npm run build
```

如果阶段性验证失败，先修复失败，再继续下一个页面。

## 9. 第二阶段：公共布局还原

### 9.1 `MainLayout.tsx`

目标：还原旧系统的整体框架。

要求：

- 页面整体为左侧菜单、顶部顶栏、右侧主内容布局。
- 侧栏固定，不随主内容滚动。
- 顶栏固定在主内容上方。
- 主内容区内部滚动。
- 旧系统无右侧任务栏。优先确认当前 `TaskbarRail`、`TaskCard` 是否造成宽度挤压。
- 如果任务栏造成布局错乱，应默认隐藏右侧任务栏渲染，或用 feature flag 控制；不要让它挤压主页面。
- 如果暂时保留任务栏，必须默认收起，且收起态不能影响主内容宽度。
- 避免页面出现双滚动条或底部操作栏被遮挡。

重点检查：

- `h-screen` 与内部 `h-full` 是否造成滚动错乱。
- `main` 的 `p-4` 是否导致旧系统页面边距不一致。
- 透析执行页面是否需要自己管理 padding。

### 9.2 `Sidebar.tsx`

目标：还原旧系统侧栏。

要求：

- 展开宽度接近 `240px`。
- 折叠宽度接近 `64px`。
- 背景为深色。
- 顶部品牌为 `AI-HMS 智能透析`。
- 菜单分组与旧系统一致。
- 激活菜单为亮蓝底白字。
- 非激活菜单为浅灰蓝文字。
- 底部显示当前登录身份、用户名、角色。
- 保留当前权限过滤逻辑，不要让无权限菜单全部显示。

菜单顺序建议参考旧图：

- 日常工作：工作台、病区概览、实时监控、透析执行、排班管理。
- 患者中心：患者管理。
- 资源管理：耗材管理、设备管理、病区管理、床位管理。
- 配置中心：主数据管理、诊疗配置、字典配置。
- 系统管理：用户管理、角色管理。

### 9.3 `Header.tsx`

目标：还原旧系统顶栏。

要求：

- 高度约 `48px`。
- 背景浅蓝灰。
- 左侧折叠按钮是白色圆角方块。
- 科室信息是白色胶囊。
- 科室文案为 `当前科室：肾内透析中心 · 第一病区` 或当前数据源中的科室。
- 右侧显示任务或日历图标、用户名、角色、头像、退出按钮。
- 用户信息排布紧凑，不要变成大卡片。

### 9.4 `index.css` 和 `App.tsx`

目标：统一全局视觉。

要求：

- 保持当前字体。
- 调整 Ant Design token，使按钮、表格、弹窗更接近旧系统。
- 不要大范围使用 `!important`。
- 可以新增少量语义类，如 `legacy-page`、`legacy-card`、`legacy-table`。
- 滚动条样式不要完全隐藏，透析执行页面需要能看到滚动状态。

## 10. 第三阶段：透析执行总框架还原

涉及文件：

- `src/pages/dialysis-processing/DialysisExecution.tsx`
- `src/pages/dialysis-processing/components/PatientListSidebar.tsx`
- `src/pages/dialysis-processing/components/PatientSummaryHeader.tsx`
- `src/pages/dialysis-processing/types.ts`

目标：先把 2.1 到 2.9 共用框架还原，再还原具体 tab。

要求：

- 左侧患者列表宽度约 `240px`。
- 左侧患者列表独立滚动。
- 主内容独立滚动。
- 中间有患者列表折叠按钮。
- 顶部标签页固定在主内容顶部。
- 标签顺序：`透前评估`、`当日处方`、`双人核对`、`透析医嘱`、`透中监测`、`透后评估`、`健康宣教`、`透析小结`。
- 标签选中态为蓝底白字。
- 标签未选中态为浅灰底深色字。
- 患者摘要展示姓名、排床状态、ID、性别年龄、费用类型、透析龄。
- 右侧展示干体重和治疗方案。
- 先检查 `src/pages/dialysis-processing/types.ts` 中 `ExecutionTab` 常量，确认标签顺序和文案与旧截图一致。
- 如需调整 tab 文案或顺序，优先修改 `ExecutionTab`，不要在多个组件中分别硬编码。

患者列表要求：

- 顶部有筛选：漏斗图标、`在科`、`全部`、总人数。
- 搜索框 placeholder：`搜索姓名 / 床位 / 患者ID`。
- 患者行显示姓名、ID、性别、年龄、就诊状态、排床状态。
- 选中行浅蓝背景，左侧蓝色竖线。
- 不要使用大卡片式患者列表。

## 11. 第四阶段：透析执行各 tab 还原

本阶段是计划内必做项，不是可选项。`ai-hms-frontend/eslint.config.js` 中对 `src/pages/dialysis-processing/**` 的临时豁免，只表示这些文件中的旧样式类暂时不会阻塞 `npm run lint`，不代表这些页面已经完成旧系统视觉和功能还原。

执行者必须按以下顺序逐个处理透析执行 tab，并在每个 tab 修改前查看对应截图、解析对应 HAR、核对当前代码。每完成一个 tab 后运行 `npm run lint` 和 `npm run build`。如果某个功能缺少当前接口，不要伪造成功，应保留旧系统布局并使用禁用态或明确提示，同时记录缺口。

透析执行 tab 还原顺序：

1. `PreAssessment.tsx`：透前评估。
2. `TodayPrescription.tsx`：当日处方。
3. `Verification.tsx`：双人核对。
4. `MedicalOrders.tsx`：透析医嘱。
5. `MidMonitoring.tsx`：透中监测。
6. `PostAssessment.tsx`：透后评估。
7. `HealthEducation.tsx`：健康宣教。
8. `DialysisSummary.tsx`：透析小结。

当前已完成范围仅包括透析执行外层框架，例如 `DialysisExecution.tsx`、`PatientListSidebar.tsx`、`PatientSummaryHeader.tsx` 的初步还原。各 tab 内部表单、表格、弹窗、底部操作区仍需继续按旧系统逐页还原。

### 11.1 透前评估

参考资料：

- `old_system/2.1透析执行—透前评估/主界面.png`
- `old_system/2.1透析执行—透前评估/10.10.8.84.har`

当前文件：

- `src/pages/dialysis-processing/execution/PreAssessment.tsx`

还原要求：

- 分组卡片：`体重与容量评估`、`生命体征监测`、`血管通路与神志状态`。
- 字段一行多列紧凑排布。
- 输入框右侧显示单位，如 `KG`、`L`、`mmHg`、`次/分`、`℃`。
- 必填字段显示红色星号。
- 透前体重下方有 `患者拒测`、`卧床` 等复选项。
- 血压支持收缩压和舒张压格式。
- 测压部位、穿刺点、神志状态等使用下拉或选择控件。
- 底部保留深色信息条，显示透析机开始时间、本日值班医生、当前登录人。
- 底部操作区保留 `称重照片历史`、提示信息、`暂存草稿`、`提交透前评估`。
- 如果暂存草稿无接口，不要假保存成功，可以禁用或提示“接口暂未提供”。

### 11.2 当日处方

参考资料：

- `old_system/2.2透析执行—当日处方/当日处方 主界面.png`
- `old_system/2.2透析执行—当日处方/10.10.8.84.har`

当前文件：

- `src/pages/dialysis-processing/execution/TodayPrescription.tsx`

还原要求：

- 对照 HAR 确认旧页面字段。
- 展示治疗模式、治疗方案、干体重、透析器、透析液、血流量、透析时长、抗凝、置换量等核心处方信息。
- 保持旧系统分组和字段顺序。
- 只读字段不要伪装成可编辑字段。

### 11.3 双人核对

参考资料：

- `old_system/2.3透析执行—双人核对/主界面.png`
- `old_system/2.3透析执行—双人核对/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/dialysis-processing/execution/Verification.tsx`

还原要求：

- 区分首次核对和二次核对。
- 显示核对项目、核对结果、核对人、核对时间。
- 保留当前接口 `saveTreatmentFirstCheck` 和 `saveTreatmentSecondCheck`。
- 按旧系统布局还原按钮和状态。

### 11.4 透析医嘱

参考资料：

- `old_system/2.5透析执行—透析医嘱/主界面.png`
- `old_system/2.5透析执行—透析医嘱/弹窗界面新增医嘱.png`
- `old_system/2.5透析执行—透析医嘱/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/dialysis-processing/execution/MedicalOrders.tsx`

还原要求：

- 主界面显示医嘱列表。
- 右上角或页面明显位置提供新增医嘱按钮。
- 新增医嘱弹窗字段、按钮、布局按旧截图还原。
- 删除、编辑、停止等操作以当前接口能力为准。
- 如果接口缺失，保留视觉但给出明确不可用提示，不要假成功。

### 11.5 透中监测

参考资料：

- `old_system/2.6透析执行—透中监测/主界面.png`
- `old_system/2.6透析执行—透中监测/弹窗界面.png`
- `old_system/2.6透析执行—透中监测/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/dialysis-processing/execution/MidMonitoring.tsx`

还原要求：

- 顶部 KPI 卡片显示：平均动脉压、实时跨膜压、当前血流量、超滤速率、异常预警。
- 主表格标题：`实时监测记录流水`。
- 标题旁显示 `REAL-TIME FEED` 标签。
- 右上角按钮：`+ 录入新监测点`。
- 表格需要支持横向滚动。
- 表格分组表头尽量接近旧系统。
- 表格底部显示机位状态和同步间隔。
- 页面底部显示黄色提示条。
- 新增监测点弹窗按旧截图字段顺序和单位还原。

### 11.6 透后评估

参考资料：

- `old_system/2.7透析执行—透后评估/主界面.png`
- `old_system/2.7透析执行—透后评估/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/dialysis-processing/execution/PostAssessment.tsx`

还原要求：

- 还原透后体重、生命体征、通路情况、并发症、护理记录等分组。
- 保存和提交逻辑使用当前接口。
- 不要改变当前治疗记录状态流转，除非确认旧系统一致。

### 11.7 健康宣教

参考资料：

- `old_system/2.8透析执行—健康宣教/主界面.png`
- `old_system/2.8透析执行—健康宣教/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/dialysis-processing/execution/HealthEducation.tsx`

还原要求：

- 展示宣教内容或宣教记录。
- 显示宣教状态、宣教人、时间等信息。
- 保持旧系统操作按钮。
- 接口不足时优先保证页面不报错，并记录缺口。

### 11.8 透析小结

参考资料：

- `old_system/2.9透析执行—透析小结/主界面.png`
- `old_system/2.9透析执行—透析小结/透析小结.har`

当前文件：

- `src/pages/dialysis-processing/execution/DialysisSummary.tsx`

还原要求：

- 还原小结展示字段。
- 还原保存、提交、打印等按钮，具体以旧系统和当前接口为准。
- 不要把无法实现的功能伪装成已实现。

## 12. 第五阶段：系统管理类页面还原

### 12.1 用户管理

参考资料：

- `old_system/用户管理/主界面.png`
- `old_system/用户管理/用户编辑.png`
- `old_system/用户管理/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/UserManagement.tsx`

还原要求：

- 面包屑：`系统设置 > 用户管理`。
- 页面标题旧图为 `人员管理`。
- 右上角搜索框 placeholder：`搜索姓名/用户名/拼音`。
- 右上角按钮：`刷新`、`新增人员`。
- 表格列：`用户名`、`真实姓名`、`性别`、`年龄`、`人员类型`、`角色`、`手机号`、`状态`、`操作`。
- 状态使用蓝色启用开关或标签。
- 操作包括：`编辑`、`重置密码`、`删除`。
- 删除操作必须二次确认。
- 编辑弹窗按旧截图还原字段顺序。

### 12.2 角色管理

参考资料：

- `old_system/角色管理/主界面.png`
- `old_system/角色管理/编辑界面.png`
- `old_system/角色管理/权限分配界面.png`
- `old_system/角色管理/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/RoleManagement.tsx`

还原要求：

- 角色列表按旧系统还原。
- 编辑弹窗按旧截图还原。
- 权限分配弹窗标题类似：`分配权限 - 系统管理员`。
- 权限分配弹窗宽度约 `760px`。
- 权限树按菜单分组展示。
- 权限项展示权限编码，如 `menu.dashboard`、`task.alert.view`。
- 说明条文案：`权限按左侧菜单分组展示：勾选菜单控制入口可见，勾选下级操作用于后续按钮级控制。`
- 底部按钮：`取消`、`确定`。
- 保留当前权限保存逻辑，不要写死权限。

### 12.3 字典配置

参考资料：

- `old_system/字典配置/透析治疗.png`
- `old_system/字典配置/其他字典.png`
- `old_system/字典配置/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/DictConfig.tsx`

还原要求：

- 改字典前必须先查看：`src/services/dictApi.ts`。
- 同时检查：`src/hooks/useDictOptions.ts` 和 `src/hooks/useDictName.ts`。
- 如涉及后端，查看：`ai-hms-backend/internal/services/dict_service.go`。
- 不清楚的字典项不要硬编码。
- 优先复用现有字典 hook 和 service，不要在页面内重复写死字典请求。
- 页面应支持字典分类、字典项列表、新增、编辑、启用状态。
- 透析治疗类字典和其他字典的切换方式按旧截图还原。

### 12.4 病区管理

参考资料：

- `old_system/病区管理/主界面.png`
- `old_system/病区管理/床位编辑.png`
- `old_system/病区管理/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/WardManagement.tsx`

还原要求：

- 还原病区列表。
- 还原搜索、刷新、新增、编辑、删除等操作。
- 还原床位编辑弹窗。
- 字段以 HAR 和当前接口为准，不要凭字段名猜测。

### 12.5 床位管理

参考资料：

- `old_system/床位管理/主界面.png`
- `old_system/床位管理/编辑床位.png`
- `old_system/床位管理/hdis.sdent.com.cn.har`

当前文件：

- `src/pages/BedManagement.tsx`

还原要求：

- 还原床位列表。
- 还原编辑床位弹窗。
- 字段包括床位编号、病区、状态等，具体以旧系统和当前接口为准。

## 13. HAR 分析要求

每个页面修改前必须分析对应 HAR。

### 13.1 HAR 打开方式

推荐方法一：使用 Chrome DevTools。

1. 打开 Chrome。
2. 按 `F12` 打开 DevTools。
3. 切换到 `Network` 标签。
4. 将对应 `.har` 文件拖入 Network 面板。
5. 查看请求 URL、方法、参数、响应。

推荐方法二：使用 PowerShell 初步列出请求。

```powershell
$har = Get-Content -LiteralPath "old_system\用户管理\hdis.sdent.com.cn.har" -Raw | ConvertFrom-Json
$har.log.entries | ForEach-Object { "$($_.request.method) $($_.request.url)" }
```

查看请求体：

```powershell
$har.log.entries | Where-Object { $_.request.postData } | ForEach-Object { $_.request.postData.text }
```

查看响应文本：

```powershell
$har.log.entries | ForEach-Object { $_.response.content.text }
```

常用搜索关键词：

- `request`
- `url`
- `method`
- `postData`
- `response`
- `content`
- `text`
- `items`
- `data`

分析内容：

- 请求 URL。
- 请求方法。
- 查询参数。
- 请求体。
- 响应字段。
- 分页参数。
- 枚举值。
- 新增、编辑、删除接口。
- 弹窗数据来源。

执行注意：

- HAR 可能包含 token 或敏感信息，不要提交。
- 不要把 HAR 原文复制进代码注释。
- 只记录必要字段和行为。
- 如果 HAR 内容为空或响应不可读，用在线旧系统重新打开页面核对。

### 13.2 HAR 到当前代码的映射

分析 HAR 后，必须回到当前代码确认是否已有对应 service：

- 优先查 `src/services/restClient.ts`。
- 再查 `src/services/*.ts`。
- 如新增拆分服务，必须从 `src/services/index.ts` 导出。
- 不要在页面组件中直接写裸 `axios` 调用，除非项目已有同类模式。

## 14. 在线旧系统核对要求

访问旧系统：

- `http://10.10.8.84:3000`
- `test_admin`
- `Test@123456`

核对步骤：

1. 登录旧系统。
2. 打开待还原页面。
3. 对照本地截图确认布局是否一致。
4. 如果在线旧系统和截图不一致，以在线旧系统为准。
5. 点开所有关键弹窗。
6. 记录按钮、字段、默认值、交互行为。

重点核对弹窗：

- 用户编辑。
- 角色编辑。
- 权限分配。
- 床位编辑。
- 病区床位编辑。
- 透析医嘱新增。
- 透中监测新增。

## 15. 功能验收清单

每个页面完成后至少检查：

- 页面能打开，不白屏。
- 页面布局不横向溢出，除非表格区域设计为横向滚动。
- 搜索框输入不报错。
- 刷新按钮可用。
- 新增按钮能打开弹窗。
- 编辑弹窗能回显数据。
- 保存按钮能调用接口。
- 删除按钮有二次确认。
- 状态开关能正确显示。
- 空数据有空状态。
- 加载中有加载状态。
- 接口失败有错误提示。
- 不显示 `undefined`、`null`、`NaN`。
- 中文文案与旧系统尽量一致。
- 1366x768 下可操作。
- 1920x1080 下不显得错乱。
- 弹窗底部按钮可见。
- 表格操作列不被遮挡。
- 透析执行底部操作区不被页面滚动遮挡。

## 16. 可能的高风险点

- 当前布局可能存在 `h-screen`、`h-full`、内部滚动叠加导致的错乱。
- 当前透析执行页面 padding 和卡片圆角可能与旧系统差异较大。
- 当前右侧任务栏可能挤压主页面宽度。
- 当前权限菜单过滤可能导致 `test_admin` 看不到旧系统完整菜单。
- 当前 Tailwind v4 和 Ant Design 混用，可能导致按钮高度、表格行高不统一。
- 当前 ESLint 严格，新增未使用变量会导致 build 失败。
- `useEffect` 内同步 `setState` 可能触发 React Hooks lint 规则。
- 字典项不要硬编码，尤其是患者、治疗、医保、证件类型相关字段。
- 当前前端使用 i18next，直接修改组件文案可能不生效，必须同步检查 `src/i18n/locales/zh-CN/*.json`。
- 旧系统无截图依据的页面不要顺手改，避免引入新的视觉回归。

## 16.1 阻塞处理规则

遇到以下情况时，按规则处理，不要凭猜测继续扩大修改范围。

旧系统在线不可达：

- 仅依赖本地截图和 HAR 还原布局。
- 功能行为标记为“待在线确认”。
- 不要猜测截图以外的弹窗或流程。

HAR 无法解析或响应内容缺失：

- 使用截图和在线旧系统核对视觉。
- 当前接口字段以现有代码和当前接口响应为准。
- 记录 HAR 缺失问题。

当前接口缺失：

- 保留旧系统布局。
- 对应按钮或功能使用禁用态，并提示 `功能待后端接口就绪`。
- 不要伪造保存成功、删除成功或提交成功。

字段无法映射：

- 记录到 `docs/legacy-migration-uncertain-field-checklist.md`。
- 暂停该字段的业务逻辑收敛。
- 可先显示安全兜底值，如 `--`，但不得把不确定字段写入提交 payload。

登录当前系统失败：

- 先记录错误信息和接口响应。
- 不要绕过 `AuthGuard`、`LoginGuard` 或本地 token 逻辑。
- 只在用户明确要求时再处理登录链路。

## 17. 最终验证

前端最终必须执行：

```bash
cd ai-hms-frontend
npm run lint
npm run build
```

如果修改后端接口，执行：

```bash
cd ai-hms-backend
go test ./internal/services ./internal/api/v1
```

Windows 上如需后端构建，避免在仓库内生成 exe：

```powershell
cd ai-hms-backend
go build -o "$env:TEMP\ai-hms-backend-check.exe" ./cmd/server
```

## 18. 最终交付说明格式

执行完成后，应向用户说明：

- 已还原哪些页面。
- 每个页面参考了哪些截图和 HAR。
- 修改了哪些主要文件。
- 哪些功能已验证。
- 哪些功能因接口缺失或字段不确定未完全还原。
- `npm run lint` 是否通过。
- `npm run build` 是否通过。
- 是否有后端改动及对应测试结果。

不要主动 commit 或 push，除非用户明确要求。
