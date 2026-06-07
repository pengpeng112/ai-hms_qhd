# UI/UX 优化执行开发计划（Execution Plan）

> 状态：可执行 · 已对齐评审答复
> 来源：`docs/ui-ux-improvement-proposal.md`（评审稿）+ 用户答复（2026-05-30）
> 适用范围：`ai-hms-frontend/`
> 约束：遵循根 `CLAUDE.md`；前端最低闸门 `npm run lint` + `npm run build`；不做后端改动；不动业务字段映射

---

## 0. 评审答复落地（执行前必读）

| # | 答复 | 落地决定 |
|---|---|---|
| 字号下限 | 按使用与操作合理来 | 通用页面下限 12px；保留极少数业务密排例外（清单见 §1.3），用 `text-density-strict` 类显式标注，禁止其他地方再用 10/11px |
| 侧边栏分组 | 按使用频次重新分组 | 一级 5 组：`日常工作 / 患者中心 / 排班 / 资源管理 / 系统配置`，依据见 §2.2 |
| 任务栏 | 按建议来 | 默认收起 56px 窄条，仅显示图标 + 数字红点；点击/锁定展开 320px |
| PatientDetail Tab | 可以试试 | 主 Tab 4 个：`总览 / 治疗 / 病历 / 历史`，子 Tab 见 §3.3 |
| Dark 主题 | 按你判断 | **删除 dark 主题**（!important 块占 50+ 行，使用率不明，留下高维护成本）；保留 `light` + `high-contrast` 两套，主题切换 UI 同步去 dark 选项 |
| Dashboard 角色化 | 角色 mockup 后再开发 | U2 阶段以"出 mockup → 评审 → 实现"两步走，本计划只先列接口契约和占位 |
| AntD ConfigProvider | 可以 | 纳入 U0 同步落地，统一 `colorPrimary / borderRadius / 字号` |

---

## 1. 阶段总览与依赖

```
U0 设计系统 ──┬──> U1 框架布局 ──┬──> U2 Dashboard / PatientList
              │                   │
              │                   └──> U3 PatientDetail / Schedule
              │
              └──> U4 Monitoring 拆分 + 状态色统一
              │
              └──> U5 a11y / 打印 / 主题清理（最后）
```

| 阶段 | 工期 | 可独立交付 | 依赖 |
|---|---|---|---|
| U0 设计系统统一 | 3 天 | ✅ | 无 |
| U1 框架级布局（Header/Sidebar/Taskbar） | 5 天 | ✅ | U0 |
| U2 Dashboard + PatientList | 8 天（含 mockup 评审） | ✅ | U0、U1 |
| U3 PatientDetail + Schedule 视觉 | 8 天 | ✅ | U0、U1，与排班重构方案 P0/P1 并行 |
| U4 Monitoring 拆分 | 5 天 | ✅ | U0 |
| U5 主题清理 / 打印 / a11y 收尾 | 3 天 | ✅ | U0–U4 |

每阶段以 PR 为单位，PR 标题前缀 `feat(ui):` / `refactor(ui):` / `style(ui):`，影响面写 `frontend`。

---

## 2. 全局约束（Definition of Done）

每个阶段、每个 PR 必须满足：

1. `cd ai-hms-frontend && npm run lint` 通过
2. `cd ai-hms-frontend && npm run build` 通过（CI 最低门槛，见根 `CLAUDE.md`）
3. 不引入新的 `rounded-2xl` / `rounded-xl` 类（U0 上线后 ESLint 拦截）
4. 不引入 `text-[10px]` / `text-[11px]`（同上）
5. 不在业务页面新增 `!important`
6. 截图核对清单（§9）至少覆盖本 PR 所改页面
7. 中文注释；与现有代码注释语言保持一致
8. 不动后端接口；不动 i18n key（避免 i18n 漏翻）；新增样式 token 必须同步出现在 `index.css` 的 `@theme`

---

## 3. U0 — 设计系统统一（3 天）

### 3.1 任务 1：扩充 `index.css` token

**文件**：`ai-hms-frontend/src/index.css`

**操作**：在 `@theme { ... }` 内新增以下变量，保留现有 `--color-primary-*` 与 `--color-sidebar` 不变。

```css
@theme {
  /* ========== Radius（仅三档） ========== */
  --radius-sm: 4px;   /* 标签、徽章 */
  --radius-md: 8px;   /* 按钮、输入、卡片头 */
  --radius-lg: 12px;  /* 模态、面板、大卡片 */

  /* ========== 字号阶梯 ========== */
  --text-meta: 12px;
  --text-body: 14px;
  --text-h3:   16px;
  --text-h2:   20px;
  --text-h1:   24px;

  /* ========== 业务状态语义色 ========== */
  --state-treating:    #1677ff;
  --state-treating-bg: #e6f4ff;
  --state-waiting:     #faad14;
  --state-waiting-bg:  #fff7e6;
  --state-finished:    #52c41a;
  --state-finished-bg: #f6ffed;
  --state-alert:       #ff4d4f;
  --state-alert-bg:    #fff1f0;
  --state-offline:     #8c8c8c;
  --state-offline-bg:  #f5f5f5;

  /* ========== 表面（替代裸 bg-white / bg-gray-50） ========== */
  --color-surface:         #ffffff;
  --color-surface-sunken:  #f5f7fa;
  --color-surface-sidebar: #1f2a44;  /* 介于深蓝灰与浅蓝之间，过渡更自然 */
  --color-foreground:      #1f2937;
  --color-foreground-muted:#6b7280;
}
```

**对应 high-contrast 主题**（在 `.theme-high-contrast { }` 内同步覆盖 `--color-surface*` 与 `--color-foreground*`）。

**Tailwind v4 自动效果**：上述 token 自动暴露为 `bg-surface / text-foreground / rounded-md / text-meta` 等类，无需额外配置。

**验收**：

- [ ] 在 `index.css` 内可见全部 token
- [ ] 任意页面写 `<div class="bg-surface rounded-md text-meta">` 不报样式错
- [ ] `npm run build` 通过

### 3.2 任务 2：删除 dark 主题

**文件**：
- `ai-hms-frontend/src/index.css`：删除整个 `.theme-dark { ... }` 块，删除全部 `.theme-dark .xxx { !important }` 覆盖（约 50 行）
- `ai-hms-frontend/src/contexts/ThemeContextBase.tsx`：从 `ThemeType` 类型移除 `'dark'`
- `ai-hms-frontend/src/contexts/ThemeContext.tsx`：从 `root.classList.remove('theme-light', 'theme-dark', 'theme-high-contrast')` 移除 `'theme-dark'`
- `ai-hms-frontend/src/pages/Settings.tsx`：主题切换 UI 移除 dark 选项（保留 light / high-contrast）
- 全局搜索 `theme-dark`、`'dark'` 主题字符串，逐一处理；旧 localStorage 值迁移：`if (stored === 'dark') stored = 'light'`

**验收**：

- [ ] `grep -r "theme-dark" src/` 仅剩文档/注释
- [ ] 切换主题 UI 上不再出现 dark 选项
- [ ] 老用户 localStorage 残留 `'dark'` 自动回落到 `'light'`，不白屏

### 3.3 任务 3：定义"密排例外"白名单

**目的**：用户答复"按使用与操作合理来"，因此并非粗暴禁 10/11px，而是只允许在下列业务密排场景出现，并显式打标。

**白名单**（仅以下场景允许 10/11px，其他禁止）：

| 场景 | 示例位置 | 类名 |
|---|---|---|
| 表格右上角辅助计数（如"3/8"） | Schedule 床位计数 | `text-density-strict` + 注释 `// density:strict` |
| 角标计数（红点内数字） | Header 任务计数 | 无（保持现有 `[10px]` 但加注释） |
| 打印交班表 | `@media print` | `print:text-[10px]` |

实现：在 `index.css` 增加

```css
@layer utilities {
  .text-density-strict { font-size: 11px; line-height: 14px; }
}
```

**验收**：

- [ ] 全局以外的 `text-[10px]/[11px]` 被替换为 `text-meta` 或 `text-density-strict`
- [ ] 替换清单提交在 PR 描述中（51 文件，逐一列出）

### 3.4 任务 4：ESLint 规则拦截

**文件**：`ai-hms-frontend/eslint.config.js`

新增 `no-restricted-syntax` 规则（自定义 `Literal` 选择器或采用 `eslint-plugin-tailwindcss` + 自定义 message）。简化方案用 `no-restricted-syntax`：

```js
// eslint.config.js 追加 rules
rules: {
  'no-restricted-syntax': [
    'error',
    {
      selector: "Literal[value=/\\brounded-(xl|2xl|3xl)\\b/]",
      message: '禁止使用 rounded-xl/2xl/3xl，请改用 rounded-md / rounded-lg（见 docs/ui-ux-execution-plan.md §3.1）',
    },
    {
      selector: "Literal[value=/\\btext-\\[10px\\]|\\btext-\\[11px\\]/]",
      message: '禁止裸用 text-[10px]/[11px]，请改用 text-meta，或在白名单场景使用 text-density-strict',
    },
    {
      selector: "Literal[value=/\\b!important\\b/]",
      message: '禁止在业务页面新增 !important，主题适配请用 surface/foreground token',
    },
  ],
},
```

**注意**：`TemplateLiteral` 也需匹配。若上述选择器无法覆盖模板字符串场景，改用 `eslint-plugin-tailwindcss` 的 `classnames-order` + 自定义校验，二选一在 PR 中验证。

**验收**：

- [ ] 引入一个临时 `<div className="rounded-2xl">` 能被 ESLint 报错
- [ ] `npm run lint` 在主分支无新告警（即把现有违规一次性清掉）

### 3.5 任务 5：AntD ConfigProvider 主题对齐

**文件**：`ai-hms-frontend/src/App.tsx`

```tsx
<ConfigProvider
  locale={zhCN}
  theme={{
    token: {
      colorPrimary: '#1677ff',
      colorError:   '#ff4d4f',
      colorWarning: '#faad14',
      colorSuccess: '#52c41a',
      colorInfo:    '#1677ff',
      borderRadius: 8,        // 对齐 --radius-md
      borderRadiusLG: 12,     // 对齐 --radius-lg
      borderRadiusSM: 4,      // 对齐 --radius-sm
      fontSize: 14,           // 对齐 --text-body
      fontFamily: '-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif',
    },
    components: {
      Button: { borderRadius: 8 },
      Modal:  { borderRadiusLG: 12 },
      Table:  { borderRadius: 8, headerBg: '#f5f7fa' },
      Tag:    { borderRadiusSM: 4 },
    },
  }}
>
```

**验收**：

- [ ] AntD Modal/Button/Table/Tag 圆角与字号目测对齐 token
- [ ] PatientList、Schedule（含 antd Table/Modal）回归无样式破坏

### 3.6 U0 交付物

- PR-U0-1：token + dark 删除（任务 1+2）
- PR-U0-2：白名单替换 + ESLint 拦截（任务 3+4）
- PR-U0-3：AntD ConfigProvider 注入（任务 5）

可串可并；建议串行，便于 review。U0 完工后跑一次全量截图并归档至 `docs/ui-snapshots/U0/`。

---

## 4. U1 — 框架级布局（5 天）

### 4.1 任务 1：Header（`layouts/Header.tsx`）

**变更**：

- 高度 `h-12` → `h-14`
- 背景 `bg-[#f0f7ff]` → `bg-surface border-b border-gray-200`（去掉浅蓝染色）
- 左侧科室徽章：双行布局，主行科室名 `text-body font-medium`，次行病区名 `text-meta text-foreground-muted`
- 任务按钮：删除 `animate-bounce`；红点数量 ≥ 100 时显示 `99+`
- 右侧用户区合并为 Popover：
  - 触发：头像 + 姓名（可点）
  - 浮层内容：当前角色 / 切换角色 / 个人设置 / 退出登录
  - 用 antd `Popover`（与 ConfigProvider 一致）
- 面包屑：从 `<main>` 内 `PageBreadcrumb` 移到 Header 第二行
  - 仅当 `routeMeta.breadcrumb.length > 1` 时显示
  - 第二行高 `h-8`，背景与第一行同色，下边框 `border-b border-gray-100`
  - Header 总高 `h-14` → `h-[88px]`（深页面）/ `h-14`（浅页面）

**新增组件**：`layouts/HeaderUserMenu.tsx`（Popover 内容），保持单文件 < 120 行。

**验收**：

- [ ] 4 角色登录后头像 Popover 项与权限对齐
- [ ] 浅页面（Dashboard / PatientList）面包屑不显示，深页面（PatientDetail）显示
- [ ] Tab 键可聚焦头像、退出按钮，焦点环可见

### 4.2 任务 2：Sidebar 重组（`layouts/Sidebar.tsx`）

**重组依据**：使用频次（来自评审答复）。结合现有 `monitoring`、`dialysisProcessing` 是日常高频，`schedule` 单独成顶级（产品强调）。

新分组：

```
1. 日常工作  dailyWork    [dashboard, wardOverview, monitoring, dialysisProcessing]
2. 患者中心  patientCenter[patients, educationManagement(隐藏)]
3. 排班      schedule     [schedule]                       ← 顶级单项
4. 资源管理  resource     [inventory, deviceBinding, wardManagement, bedManagement]
5. 系统配置  systemConfig [masterData, treatmentConfig, dictConfig, userManagement, roleManagement, settings, statistics]
```

> 备注：`statistics` 使用频次低，与系统配置同放系统区；如产品异议可移回独立 `analytics` 顶级。

**新增交互**：

- **折叠态二级浮层**：sidebar `isOpen=false` 时，Hover 一级图标弹出 mini Popover 列出二级菜单
  - 用 antd `Popover trigger="hover"`，`placement="right"`，`mouseEnterDelay=0.2`
- **品牌区合并角色信息**：
  - 展开态：上行 `AI-HMS 智能透析`，下行 `当前角色：xxx` 灰字
  - 折叠态：仅显示 `AI` 字标
- **底部留白**：原"角色卡"区域改为版本号 + 帮助链接（先占位）

**配色**：背景从 `bg-slate-900` → `bg-[var(--color-surface-sidebar)]`（即 `#1f2a44`）。

**验收**：

- [ ] 折叠/展开切换流畅，Hover 浮层不抖动
- [ ] 5 组顶级菜单按角色权限过滤后仍正确显示
- [ ] 当前路由匹配时高亮无回归

### 4.3 任务 3：Taskbar 默认收起（`layouts/MainLayout.tsx`）

**变更**：

- `taskbarOpen` 默认值改 `false`（同时 `localStorage` 历史值兼容：`saved !== null ? saved === 'true' : false`）
- 收起态：右侧仍占 `w-14 (56px)` 窄条，纵排显示：
  - 顶部：图标按钮（点击展开/收起）
  - 下方：按 severity 分组的"红/橙/蓝"小圆点 + 数字徽章
  - 底部：锁定按钮（点击切到常驻 320px）
- 展开宽度从 `w-[340px]` 调整为 `w-[320px]`（节约空间）
- 任务卡片：
  - `rounded-2xl` → `rounded-md`
  - severity 由整片底色 → 左侧 4px 色条 + `bg-surface`（仅 hover 时浅色背景）
  - 顶部右上角增加 3 个图标按钮：标记已读 / 转交 / 跳转详情（已读、转交先占位 `disabled` + Tooltip "即将上线"，避免新增后端接口）
- 空态：`暂无任务` 不再大绿勾占满整列；改为顶部一行 `已清空 ✓`，下方仍可显示历史（先用占位）

**新增组件**：`layouts/TaskbarRail.tsx`（56px 窄条），`layouts/TaskCard.tsx`（卡片单元），便于复用与单测。

**验收**：

- [ ] 默认进入工作台，任务栏为窄条；点击图标展开 320px
- [ ] 锁定后下次刷新依然展开
- [ ] severity 色条与 §3.1 状态色 token 一致
- [ ] 卡片点击跳转行为不变（保留权限校验）

### 4.4 U1 交付物

- PR-U1-1：Header 重做（含面包屑迁移、Popover）
- PR-U1-2：Sidebar 重组（含折叠浮层）
- PR-U1-3：Taskbar 收起化（含 Rail/Card 拆分）

完工后归档截图：`docs/ui-snapshots/U1/`，与 U0 对比。

---
## 5. U2 — Dashboard + PatientList（8 天）

### 5.1 Dashboard 角色化首屏（5 天，含 mockup 评审 1.5 天）

**子阶段**：

- **Step 1（mockup，1 天）**：产出 3 张 mockup（医生 / 护士 / 管理员），输出到 `docs/ui-mockups/dashboard-{role}.md`（用文字框图或导出图片）。**评审通过前不实现**。
- **Step 2（评审，0.5 天）**：与产品/医护过 mockup，圈定每角色默认显示卡片清单。
- **Step 3（实现，3 天）**：

**默认卡片建议（mockup 出之前的初稿）**：

| 角色 | 默认卡片（按位置 1-4） |
|---|---|
| 医生 (DOCTOR_*) | 今日我的患者 / 待开方提醒 / 近 7 日治疗汇总 / 异常化验提醒 |
| 护士 (NURSE_*) | 今日班次床位矩阵 / 待执行医嘱 / 待透前评估 / 在线设备统计 |
| 管理员 (ADMIN) | 运营总览（4 KPI 卡） / 设备利用率 / 患者增长 / 治疗量趋势 |
| 调度护士 (NURSE_SCHEDULER) | 今日班次床位矩阵 / 待排班患者 / 床位利用率 / 排班合约异常 |

**实现规范**：

- 现有 `cardsConfig` + `localStorage` 自定义机制保留
- 新增 `getDefaultCards(role: AppRole): LocalCardConfig[]` 工具函数，替代当前 `DASHBOARD_CARDS` 单一来源
- localStorage 键由 `dashboard_layout_config_${role}` 改为 `dashboard_layout_config_v2_${role}`，旧键检测到自动迁移（找不到默认卡片时回落到 v2 默认）
- 每个角色卡片清单由 `constants/dashboardDefaults.ts` 集中维护
- 入口改进：将"自定义布局"按钮从下拉藏区移到右上角"⚙ 自定义"，目视一级
- 指标卡视觉：删除 `rounded-2xl + 渐变` 装饰；统一为 `bg-surface rounded-md border border-gray-100 p-4`，主指标 `text-h1` 大字，下方副指标 `text-meta`

**新增/修改文件**：

- 新增 `src/constants/dashboardDefaults.ts`
- 新增 `src/pages/dashboard/cards/`（按卡片类型拆 5–7 个组件，每个 < 200 行）
- 修改 `src/pages/Dashboard.tsx`（缩到 < 300 行，仅做编排）

**验收**：

- [ ] 4 种角色登录默认布局符合 mockup
- [ ] 老用户（已有 v1 localStorage）首次进入不空白
- [ ] 自定义布局功能不退化（增删卡片、保存、重置）

### 5.2 PatientList 优化（3 天）

**变更**：

- **顶部统计**：竖线分隔的 `text-sm` → 4 个 mini stat 卡（无边框，仅 `bg-surface-sunken rounded-md p-3`），数字 `text-h2` + 标签 `text-meta`
- **筛选 Tab 收束**：
  - 主筛 3 个：`全部 / 在科 / 转出`
  - 次筛收进右侧 `Filter` 弹层（antd `Popover`），含：今日治疗 / 透析中 / 我的患者 / 自定义条件（性别/年龄段/医保/默认模式）
  - `Filter` 按钮显示当前激活次筛数量徽章
- **表格**：
  - 行高从 `py-4`（约 56px）调到 `py-5`（约 64px）
  - hover 高亮：取消整行 `hover:bg-blue-50/30`，改为左侧 `hover:border-l-4 hover:border-state-treating`，`pl-2` 留位（防止抖动用 `border-l-4 border-transparent` 占位）
  - 删除按钮：从行内可见的 `<Trash2>` 按钮 → 行末"…"下拉（antd `Dropdown`），下拉内含"删除""转出""复制 ID"
- **空态**：`<EmptyState>` 增加"新增建档"快捷按钮

**新增/修改文件**：

- 修改 `src/pages/PatientList.tsx`
- 新增 `src/pages/patient-list/FilterPopover.tsx`
- 复用 `src/components/ui/EmptyState.tsx`

**验收**：

- [ ] 筛选切换、搜索、分页、删除、新增 5 项交互回归无破坏
- [ ] hover 不再有整行染色闪烁
- [ ] 误触删除概率显著降低（需 1 步下拉 + 确认 Dialog）

### 5.3 U2 交付物

- PR-U2-1：Dashboard mockup（仅文档）
- PR-U2-2：Dashboard 实现 + 拆分
- PR-U2-3：PatientList 优化

---

## 6. U3 — PatientDetail + Schedule 视觉（8 天）

### 6.1 PatientDetail Tab 收束（5 天）

**新主 Tab 结构**（4 个）：

```
总览 overview   ─ 单页：风险卡 + 最近一次透析摘要 + 关键化验 + 治疗趋势
治疗 treatment  ─ 子 Tab：[计划 plan / 方案医嘱 schemeOrder / 血管通路 vascular]
病历 records    ─ 子 Tab：[基本信息 basicInfo / 化验检查 labs / 月报 monthly]
历史 history    ─ 单页：原 HistoryTab
```

**实现规范**：

- 主 Tab 用 antd `Tabs`（与 ConfigProvider 风格一致）
- 子 Tab 用 antd `Segmented`（区分层级，避免双层 Tabs 视觉混淆）
- URL 同步：主 Tab 进 query `?tab=treatment&sub=schemeOrder`，刷新保持
- `OverviewTab` 改造为聚合页：删除原"总览"中冗余的快速跳转链；改为风险卡（红黄绿三档）+ "最近一次透析"+ "近 7 日趋势"
- **Focus Bar 改为右侧固定卡**：去掉 `setFocusBarOpen` 切换逻辑；右侧 320px 固定面板，内容 = 关键风险列表 + 最近一次透析摘要；窗口宽度 < 1280 时自动折叠为右上角"风险"按钮
- **顶部头像区**：5 个并列徽章 → 主信息 `姓名 + 年龄 + 床号` 大字；右上角风险等级胶囊（`高危/中危/低危` 对应 `--state-alert/--state-waiting/--state-finished`）；其他次信息（医保 / 主治医生 / 默认模式）压缩到第二行 `text-meta`

**新增/修改文件**：

- 修改 `src/pages/PatientDetail.tsx`（拆为 < 200 行的编排）
- 新增 `src/pages/patient-detail/PatientHeader.tsx`（头像区）
- 新增 `src/pages/patient-detail/FocusPanel.tsx`（右侧固定卡）
- 子 Tab 容器：`src/pages/patient-detail/tabs/TreatmentTabs.tsx`、`RecordsTabs.tsx`

**i18n 注意**：原 9 个 Tab 的 i18n key 保留供子 Tab 使用，不要删；新增主 Tab key：`patient.tab.overview/treatment/records/history`。

**验收**：

- [ ] 各角色进入患者详情，URL 直链 `?tab=...&sub=...` 都能正确高亮
- [ ] 现有所有原 Tab 内容均可达，无功能丢失
- [ ] 1280px 以下右侧风险卡不挤压主区

### 6.2 Schedule 视觉对齐（3 天，仅视觉，业务重构走排班方案 P0）

**变更**：

- 床位 × 日期网格格高从 `h-14` → `h-16`
- 已排卡片：模式色块（5 模式 5 种全色背景）→ 改为左侧 4px 色条 + `bg-surface`，姓名 `text-body`，模式 `text-meta`
- 历史日期格子保持灰底，但加 `cursor-not-allowed` 而非禁用全部点击（用于查看治疗记录跳转）
- 待排班队列：hover 闪烁（当前 `transition-colors` 偶发抖动）→ 静态 `hover:bg-surface-sunken`
- "本周完成度"指标：在顶部黄金位（替换当前"应用模板"按钮位置）显示，公式 `本周已排 / (合约应排 || 床位时间表应开)`，实现需后端给字段；P1 前先用前端兜底估算 + Tooltip 提示"估算值"
- 角标修复：依赖排班方案 P0 任务（后端 `PatientShiftWeekItem` 加 `sourceType` + `isManualAdjusted`）；本计划不提供后端实现，仅做前端联动准备
- "应用模板"按钮挪到右上角次级位置

**注意**：本节只改视觉与位置，不改交互流程；交互/数据层走 `docs/schedule-redesign-proposal.md`。

**验收**：

- [ ] 周视图、日视图、待排班队列、拖拽、右键操作、互换 全量回归通过
- [ ] 不同模式色条与 token 一致；色盲下仍能区分（模式缩写文字保留）

### 6.3 U3 交付物

- PR-U3-1：PatientDetail Tab 收束 + Header 重做
- PR-U3-2：Focus Bar 改固定卡
- PR-U3-3：Schedule 视觉对齐

---

## 7. U4 — Monitoring 拆分 + 状态色统一（5 天）

**目标**：把 `Monitoring.tsx` 1311 行拆到可维护粒度（每文件 < 350 行），并统一状态色。

### 7.1 拆分

**新结构**：

```
src/pages/monitoring/
  index.tsx                  编排（< 200 行）
  StatusGrid.tsx             设备/床位矩阵
  AlertList.tsx              报警流（severity 卡片）
  PatientPanel.tsx           当前选中患者实时数据
  hooks/
    useMonitoringData.ts     轮询/订阅逻辑抽取
    useAlertActions.ts       报警处理
  types.ts
```

**步骤**：

1. **先抽 hooks**（不动 UI），跑通现有页面 → 确认无回归
2. **再拆组件**（每次只搬一块，每搬一块跑一次 build）
3. 旧 `Monitoring.tsx` 保留为 `Monitoring.legacy.tsx` 一周后删除（避免 PR 过大）

### 7.2 状态色统一

- 全文件 grep `bg-(red|orange|yellow|green|blue)-(50|100|200)`，逐一映射到 `--state-*-bg`
- 边框/文字色映射到 `--state-*`
- 高优先级报警：`bg-state-alert-bg + border-l-4 border-state-alert`，禁止整片飘红
- 色盲补充：每条报警在文字前加 `Lucide icon`（`AlertTriangle / AlertCircle / Info`），与颜色双通道

### 7.3 U4 验收

- [ ] 拆分后每文件 < 350 行
- [ ] 报警弹出/消失/确认流程无回归
- [ ] 关闭浏览器颜色（模拟色盲）能从图标区分严重度

### 7.4 U4 交付物

- PR-U4-1：Hooks 抽取
- PR-U4-2：组件拆分
- PR-U4-3：状态色统一 + 色盲增强

---

## 8. U5 — 收尾（3 天）

### 8.1 主题 CSS 清理

- 全局搜 `!important`，逐项消除（仅保留 `@media print` 与第三方覆盖必要项）
- 删除 `index.css` 中过时的 `.theme-dark` 残余（U0 漏网部分）
- `.theme-high-contrast` 保留并补强：所有交互元素至少 7:1 对比度

### 8.2 打印样式

新增 `src/styles/print.css`（在 `index.css` 末尾 `@media print` 块）：

```css
@media print {
  aside, header, .no-print, [data-no-print] { display: none !important; }
  main { padding: 0; }
  table { page-break-inside: auto; }
  tr { page-break-inside: avoid; page-break-after: auto; }
}
```

测试目标：Schedule 周视图、PatientList、Monitoring 设备矩阵 三个页面打印效果可读。

### 8.3 a11y 收尾

- 表格行 `tr` 加 `tabIndex={0}`、`role="row"`，键盘 Enter 等价点击
- 所有可点击 `<div>` 改为 `<button>` 或加 `role="button" tabIndex={0}`
- 全局焦点环：`*:focus-visible { outline: 2px solid var(--color-primary-500); outline-offset: 2px; }`
- `text-gray-400` on `bg-surface` 全替换为 `text-foreground-muted`（满足 4.5:1）

### 8.4 U5 验收

- [ ] axe-core / Lighthouse a11y 评分 ≥ 90
- [ ] 打印 Schedule 周视图为 A4 横向，无侧栏遮挡
- [ ] high-contrast 主题切换后无白底白字 / 黑底黑字

---

## 9. 截图核对清单（每 PR 强制）

按 §9 模板提交对比截图（before / after）：

```
docs/ui-snapshots/
  U0/  tokens-and-antd.png
  U1/  header.png  sidebar.png  taskbar.png
  U2/  dashboard-doctor.png  dashboard-nurse.png  dashboard-admin.png  patient-list.png
  U3/  patient-detail-overview.png  patient-detail-treatment.png  schedule-week.png
  U4/  monitoring-grid.png  monitoring-alert.png
  U5/  print-schedule.png  high-contrast.png
```

PR 描述模板：

```markdown
影响面：frontend
关联：docs/ui-ux-execution-plan.md §3.1 任务 1（举例）
变更：
  - …
验证：
  - npm run lint  ✅
  - npm run build ✅
  - 截图：docs/ui-snapshots/U0/…
回归：
  - 角色 A/B/C 登录、… 路径
```

---

## 10. 风险预案

| 风险 | 影响 | 预案 |
|---|---|---|
| ESLint 一次性拦截后大量历史违规 | 主分支不绿 | U0 PR 内一次性把存量违规清理到 0；按目录分批，先 `pages/Schedule.tsx` 等大头 |
| AntD ConfigProvider token 变更触发样式回归 | 弹窗、表格视觉跳变 | U0 PR 必须含 PatientList、Schedule、Modal 关键路径回归截图 |
| Dashboard 角色化首屏 mockup 评审延期 | U2 阻塞 | 不影响 U3/U4，可先开 U3 分支；mockup 评审超 3 个工作日则降级为"统一中性首屏 + 用户自定义"作为兜底 |
| Sidebar 重组改变路由直链习惯 | 用户记忆混乱 | 路由路径不变，只改顶级分组；在 Header Popover "更新日志"提示一次 |
| Monitoring 拆分中数据轮询不稳定 | 实时性受影响 | hooks 抽取阶段保持原轮询周期，只迁移代码不改逻辑；上线后开关回滚 |

---

## 11. 验收与上线节奏

- 每阶段独立 PR，merge 到 `feat/ui-uplift` 长期分支
- U0 完成后即可上线（无业务风险）
- U1–U4 在 `feat/ui-uplift` 累积，每 5 天与 `main` rebase 一次
- U5 完成后再合 `main`，单次合并；合并前由 1 名医护现场试用 0.5 天

提交信息格式（见根 `CLAUDE.md`）：

```
refactor(ui): unify radius/text tokens (U0-1)
feat(ui): role-based dashboard skeleton (U2-2)
```

PR 标注影响面 `frontend`；新增 token / 主题键写入 PR 描述 "新增环境变量/Token" 段。

---

## 12. 参考与不在范围

**参考**：

- 评审稿：`docs/ui-ux-improvement-proposal.md`
- 排班重构：`docs/schedule-redesign-proposal.md`（U3 视觉与之并行）
- 字典工作流：项目根 `AGENTS.md`（任何字典字段改动前置检查）

**明确不在范围**（避免越权）：

- 不动后端任何接口与字段
- 不引入新依赖（除非 PR 里单独立项；目前规划无新依赖）
- 不改 i18n key（仅新增 key）
- 不改路由 path（仅 query 参数同步 URL）
- 不动业务字段映射与字典 typeCode

---

## 13. 给执行 AI 的硬性要求

1. **严格按阶段顺序** U0 → U1 → U2/U3/U4 并行 → U5；前一阶段未通过验收禁止启动后一阶段
2. 每个 PR ≤ 800 行 diff（不含截图 / 文档），超出拆 PR
3. 不允许跨阶段顺手改：例如做 U1 不顺手改 PatientList
4. 所有"不确定"必须先记录到 `docs/legacy-migration-pending-confirmation.md`（沿用项目既有机制），等待人工确认再实现
5. 触碰旧文件先 Read 全文，不允许凭关键字 sed 替换
6. 所有 PR 必须自带回归截图与上述 §9 模板填写
7. 文档/注释一律中文（与现有代码库一致，见根 `CLAUDE.md` Output Style）

