# C4 收费归集 — 移交说明

**模块**：C4 收费归集（治疗费用项归集引擎，非收费系统；真正收费在 HIS）
**分支**：`feat/billing-c4`　**基线**：`5d66b8d`（origin/master 最新：A1-A4/B1-B4/C1-C3 全部已合入）
**交付日期**：2026-06-23
**规则依据**：`ai-hms_规则C4_收费归集.md` v1.0（2026-06-23 angela 逐项审定）

## 怎么合入（二选一）

### 方式 A：bundle（推荐，含完整历史）
```bash
git fetch ./billing-c4.bundle feat/billing-c4:feat/billing-c4
git checkout <你的集成分支>   # 基于 5d66b8d
git merge feat/billing-c4
```

### 方式 B：patches（按序 am）
```bash
git checkout <你的集成分支>   # 需基于 5d66b8d
git am patches/0001-*.patch patches/0002-*.patch patches/0003-*.patch patches/0004-*.patch patches/0005-*.patch
```

## 5 个提交

| # | 提交 | 内容 |
|---|------|------|
| 1 | 后端脚手架 | 参考价目 catalog（go:embed billing_catalog.json）+ charge_record/charge_line 模型 + DDL 表22/23 + health 自检 |
| 2 | 归集引擎 | BuildDraft 五类费用项（A治疗费/B耗材费/C护理费/D注射费/E药品）+ SaveDraft 落库 + 合计 |
| 3 | 生命周期+报表 | 清单读取/增删改行 + 状态流 confirm/check/push/cancel + List + DaySummary 报表 |
| 4 | API 层 | HISPusher 预留接口（NoopPusher 桩）+ billing_handler 11 端点 + main 装配 |
| 5 | 前端 | billingApi 客户端 + 护士友好 Excel 导出 + 患者「收费归集」Tab + 患者详情侧栏装配 |

## 部署前置

**建两张新表**（执行 `docs/sql/deploy_new_tables.sql` 表 22/23，或整脚本幂等重跑）：
- `charge_record`（清单头）
- `charge_line`（清单明细）

启动时 health.go 自检会列出缺表告警。AutoMigrate 永久禁用，必须手工建表。

## 落地说明

### 系统定位（关键）
本模块**只做费用项归集 + 护士核对 + 导出/推送给 HIS**，**不做真实收费/医保拆分/结算/退费/发票**。参考价仅供护士核对，实际以 HIS 为准。

### 五类费用项
- **A 治疗费**：按治疗模式自动挂 1 项固定价（HD399/HDF599/HF399/HP550/HD+HP850/PE1680/DFPP336）；CRRT 按时长 115×小时 + 加收 35（默认不勾，billable=false 占位待护士手动启用）。
- **B 耗材费**：读 `Treatment_MaterialTrace` 实际用量，按 ChargeItemId 套 catalog 参考价。三档：可计费（进清单计金额）/不可计费（内瘘包·上下机包·护理包·注射器·敷贴敷料·透析液A/B，进清单灰显不计金额）/待核价（catalog 无此项，进清单 unitPrice=null 提示护士补价）。
- **C 护理费**：按患者血管通路类型自动挂——AVF/AVG→造瘘护理17；TCC/NCC→置管护理12+中换药×2(每次21)=54。
- **D 注射费**：本次有静脉给药则挂 1 项 3 元（每次治疗最多 1 项）。
- **E 药品项**：读 `medication_admin` 给药记录，列药名+规格，**不带价**（供 HIS 查价）。

### 参考价权威来源
`billing_catalog.json`（go:embed，院方改价重构建即生效）。数据来自用户①表（最新标准）join 老库 `charge_item` 的 ChargeItemId。**老库 Stock_ChargeItem.Price 全是 0，不可用，已弃。** 当前 30 项有价、约 60 项待核价（多为①表未列的药品/监测项，本就归 HIS 或不在耗材范围）、11 项不可计费。院方可随时补价。

### 收费流程 / 状态机
`draft → confirmed → checked → pushed → settled / cancelled`
- 治疗启动 → 系统归集生成草稿 → 当班护士点确认 → 另一护士核对（系统记录核对人，**不强制核对人≠记账人**）→ 推送/导出 → HIS 回传结算/退费。
- draft/confirmed 可编辑（增删改行重算合计）；checked 起锁定不可编辑。

### HIS 推送（预留，本期不实现）
`HISPusher` 接口 + `NoopPusher` 桩（照搬 CNRDS Exporter 模式）。当前 push 只把清单标 `pushed` 并返回引导语；真正推送（HTTP/HL7/文件）待 HIS 厂商出接口后加实现即可，调用方不变。
**临时过渡**：护士在患者「收费归集」Tab 点「导出 Excel」，得到护士友好清单（固定列序、参考总价行、不可计费灰显标注、可整列复制），人工录入 HIS。

### 前端入口
- **患者详情「收费归集」Tab**（已挂 PatientDetail 侧栏，知情同意后）：单患者清单明细、增删改、确认/核对/推送/取消、导出 Excel。
- ⚠️ **归集生成**按钮需"本次治疗/处方"上下文（`currentTreatmentId`/`currentPrescriptionId`）。Patient 类型暂无这两字段，v1 拿不到时给引导提示「请在治疗执行页发起归集」。**建议团队**：要么在 PatientDetail 载入时补这两字段，要么把"归集生成"动作接到透析执行流页（更贴合规则"治疗启动后归集"）。展示+编辑+状态+导出已全可用。

## 验证（已全过）
- 后端：`go build ./...` exit 0；`go vet ./...` clean；`go test ./internal/services/ -run TestBilling` 8 用例 PASS；`go test ./internal/config/ -run TestLoadBillingCatalog` PASS；全包 `go test ./internal/...` 无回归。
- 前端：`npx tsc --noEmit` exit 0。

## 本期未做（后续迭代）
- 独立「收费管理」一级页（当班全患者列表 + 批量导出 + 班次/日/月报表）——后端 DaySummary 已就绪、前端 exportChargeBatch 已就绪，差页面+路由+菜单装配（方案 Task 16）。
- 月报/按项目排行/按医保类型分布聚合（后端补 MonthByPatient/ByItem）。
- 检查检验费纳入（规则留接口，暂不启用）。
- HIS 真实推送实现 + 退费回传取消。
- 待核价项院方补全。

无 push 权限，补丁移交团队。同款交付模式见既有 A/B/C 各模块交付包。
