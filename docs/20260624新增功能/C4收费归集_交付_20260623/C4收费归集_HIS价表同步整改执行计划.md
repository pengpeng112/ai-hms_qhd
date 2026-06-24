# C4 收费归集 + HIS 价表同步整改执行计划

> 日期：2026-06-24  
> 适用仓库：`ai-hms_qhd` 当前 `master`  
> 交付来源：`docs/20260624新增功能/C4收费归集_交付_20260623`  
> 目标：合入 C4 收费归集功能，同时接入 HIS Oracle `price_list` 价表同步；HIS 真实推送本期只预留、不启用。

## 1. 背景与结论

C4 原交付包基于 `5d66b8d`，当前仓库已经继续合入 A1/A2/A3、B2/B3/B4、C1/C2/C3 等功能，因此不能直接执行 `git am`、`git merge` 或整包覆盖。

本次整改采用手工融合：新增文件可提取，公共文件必须逐段合并。

C4 原交付包与本计划的关键差异：

| 项目 | C4 原交付包 | 本次整改 |
|---|---|---|
| 价表来源 | `billing_catalog.json` 为主要参考价来源 | HIS `price_list` 同步到 `his_price_item` 后优先查价 |
| HIS 推送 | `NoopPusher` 可把清单标记为 `pushed` | 本期禁用 push，不进入 `pushed` 状态 |
| 导出 | 前端 Excel 导出 | 保留，并记录 `exported_at` |
| 公共文件 | 基于 `5d66b8d` patch | 当前仓库手工融合 |
| 后续对账 | 文档预留 | 明确等 HIS 费用清单表结构后开发 |

回滚边界：

- 代码回滚按 Git 回退本次 C4/HIS 价表相关提交。
- 数据库回滚不自动执行 `DROP TABLE`，如需删除 `charge_record`、`charge_line`、`his_price_item`，必须 DBA 人工评估后执行。
- 如果只需临时关闭功能，后端保留路由但前端隐藏入口，或通过配置禁用收费归集入口；不要删除历史清单数据。
- HIS 价表同步失败不影响血透治疗主流程，只影响收费归集查价和价表搜索。

核心结论：

- C4 继续定位为“治疗费用项归集引擎”，不是收费系统。
- HIS `price_list` 是收费项目字典和参考价格的权威来源。
- `billing_catalog.json` 只保留为治疗费、护理费、注射费等规则兜底，不再作为药品/耗材长期权威价格来源。
- HIS 真实推送接口本期不开发、不启用；前端不显示“推送 HIS”。
- 后续 HIS 实际费用清单同步和对账需要等 HIS 表结构提供后再开发，本期只预留边界。

## 2. 已确认信息

### 2.1 Oracle 访问方式

`price_list` 在 Oracle 中的实际访问方式与现有检查报告同步一致，复用当前 HIS Oracle 配置和客户端：

- 配置来源：`HIS_ORACLE_HOST`、`HIS_ORACLE_PORT`、`HIS_ORACLE_SERVICE`、`HIS_ORACLE_USER`、`HIS_ORACLE_PASSWORD`
- 客户端：`ai-hms-backend/internal/integrations/his_oracle/client.go`
- 同步任务表：`sync_job_configs`、`sync_job_runs`

SQL 中表名先按现有检查报告同步风格使用 schema 前缀：

```sql
FROM his.price_list
```

如果现场账号同义词已映射，也可后续调整为：

```sql
FROM price_list
```

### 2.2 `item_class` 值域

| CLASS_CODE | CLASS_NAME |
|---|---|
| A | 西药 |
| B | 中药 |
| C | 化验 |
| D | 检查 |
| E | 治疗 |
| F | 手术 |
| G | 麻醉 |
| H | 血费 |
| I | 材料 |
| J | 床位 |
| K | 护理 |
| L | 膳食 |
| Z | 杂费 |
| Q | 其他 |

C4 归集中的主要映射：

| C4 类别 | HIS `item_class` | 说明 |
|---|---|---|
| A 治疗费 | E | 血液透析、血液透析滤过、血液灌流等 |
| B 耗材费 | I | 透析器、血液回路、穿刺针、灌流器等 |
| C 护理费 | K | 造瘘护理、置管护理、中换药等 |
| D 注射费 | E 或 K | 需按 HIS `price_list` 实际项目确认，本期允许人工匹配 |
| E 药品项 | A/B | 西药/中药，血透常用药一般为 A |

### 2.3 唯一键

`price_list.item_code` 已确认全院唯一。

本地镜像表使用唯一索引：

```sql
UNIQUE (source_system, item_code)
```

### 2.4 HIS 费用清单

HIS 实际费用清单表名和字段后续提供。

本期不开发 HIS 实际费用清单同步和对账，只在设计中预留。

### 2.5 新表许可

本期允许新增 `his_price_item` 表。

## 3. 本期范围

### 3.1 本期开发

- 合入 C4 收费归集后端模型、服务、API。
- 合入 C4 患者详情“收费归集”前端 Tab。
- 新增 HIS `price_list` 同步到本地 `his_price_item`。
- 收费归集查价优先使用 `his_price_item`。
- 保留 Excel 导出，供护士人工录入 HIS。
- 支持护士确认、双人核对、手工增删改明细。
- 支持 HIS 价表搜索，便于手工匹配收费项目。

### 3.2 本期不开发

- 不开发真实 HIS 推送。
- 不启用 `/charges/:id/push`。
- 不将清单状态自动标为 `pushed`。
- 不开发 HIS 退费回传。
- 不开发 HIS 实际费用清单同步。
- 不开发医保拆分、结算、收款、发票。
- 不在服务启动或请求路径执行 DDL。

## 4. 数据库设计

所有 DDL 只追加到：

```text
docs/sql/deploy_new_tables.sql
```

由 DBA 或部署阶段执行，应用运行时不得执行 DDL。

### 4.1 `his_price_item`

用途：保存 HIS Oracle `price_list` 的本地镜像，作为收费项目字典和参考价格来源。

建议 DDL：

```sql
-- 24. his_price_item HIS price_list 本地镜像
CREATE TABLE IF NOT EXISTS his_price_item (
    id varchar(36) PRIMARY KEY,
    source_system varchar(32) NOT NULL DEFAULT 'HIS_ORACLE',
    item_class varchar(1),
    item_code varchar(20) NOT NULL,
    item_name varchar(120),
    item_spec varchar(50),
    units varchar(30),
    price decimal(9,3),
    prefer_price decimal(9,3),
    foreigner_price decimal(9,3),
    performed_by varchar(8),
    fee_type_mask integer,
    class_on_inp_rcpt varchar(1),
    class_on_outp_rcpt varchar(1),
    class_on_reckoning varchar(10),
    subj_code varchar(10),
    class_on_mr varchar(4),
    memo varchar(100),
    start_date timestamp,
    stop_date timestamp,
    operator_code varchar(8),
    enter_date timestamp,
    high_price decimal(10,4),
    material_code varchar(20),
    score_1 decimal(10,2),
    score_2 decimal(10,2),
    price_name_code varchar(20),
    control_flag varchar(1),
    input_code varchar(100),
    input_code_wb varchar(100),
    std_code_1 varchar(20),
    changed_memo varchar(40),
    class_on_insur_mr varchar(24),
    package_spec varchar(20),
    firm_id varchar(10),
    charge_according varchar(23),
    license_id varchar(20),
    update_flag decimal,
    dept_name varchar(100),
    update_flag_syb decimal,
    mr_bill_class varchar(4),
    class_on_mr_add varchar(4),
    cwtj_code varchar(20),
    high_value decimal(9,3),
    drg_code varchar(8),
    insur_update integer,
    stop_operator varchar(8),
    limit_quantity decimal(10,0),
    is_active boolean NOT NULL DEFAULT true,
    synced_at timestamp NOT NULL DEFAULT now(),
    sync_run_id varchar(36),
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_his_price_item_code
ON his_price_item (source_system, item_code);

CREATE INDEX IF NOT EXISTS idx_his_price_item_name
ON his_price_item (item_name);

CREATE INDEX IF NOT EXISTS idx_his_price_item_input_code
ON his_price_item (input_code);

CREATE INDEX IF NOT EXISTS idx_his_price_item_class
ON his_price_item (item_class);

CREATE INDEX IF NOT EXISTS idx_his_price_item_active
ON his_price_item (is_active, stop_date);

COMMENT ON TABLE his_price_item IS 'HIS price_list 本地镜像，用于收费归集查价和项目匹配';
```

字段映射要求：

| HIS `price_list` 字段 | 本地字段 | 说明 |
|---|---|---|
| `stop_date` | `stop_date` | 原样保存 HIS 停用时间 |
| `operator` | `operator_code` | 避免使用 SQL 保留/易混名称 |
| `item_code` | `item_code` | 全院唯一，作为 upsert 业务键 |
| `price` | `price` | 默认参考价来源 |
| `prefer_price` | `prefer_price` | 仅展示或后续策略使用，本期不优先 |

`is_active` 必须每次同步重新计算，保证幂等：

```text
is_active = stop_date IS NULL OR stop_date > now()
```

如果 HIS 后续对同一 `item_code` 下发新的 `stop_date` 或恢复启用，同步任务必须覆盖本地 `stop_date` 和 `is_active`，不能只在新增时计算。

### 4.2 `charge_record`

用途：血透系统生成的收费归集清单头。

C4 原 DDL 可用，但建议新增导出字段，避免本期误用 `pushed`：

```sql
-- 22. charge_record 收费归集清单头（规则C4）
CREATE TABLE IF NOT EXISTS charge_record (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id bigint,
    treatment_id bigint NOT NULL,
    prescription_id bigint,
    charge_date timestamptz,
    shift varchar(16),
    dialysis_mode varchar(16),
    access_type varchar(16),
    crrt_hours decimal(5,2),
    total_amount decimal(10,2),
    status varchar(16) NOT NULL DEFAULT 'draft',
    recorded_by varchar(64),
    recorded_name varchar(64),
    checked_by varchar(64),
    checked_name varchar(64),
    checked_at timestamptz,
    exported_at timestamptz,
    pushed_at timestamptz,
    note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cr_tenant_patient ON charge_record (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_cr_treatment ON charge_record (treatment_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cr_tenant_treatment_active
ON charge_record (tenant_id, treatment_id)
WHERE status <> 'cancelled';
CREATE INDEX IF NOT EXISTS idx_cr_date ON charge_record (tenant_id, charge_date);
CREATE INDEX IF NOT EXISTS idx_cr_status ON charge_record (tenant_id, status);

COMMENT ON TABLE charge_record IS '收费归集清单头（规则C4）';
```

一致性要求：

- `tenant_id + treatment_id` 在非取消状态下只允许一张有效清单，防止重复生成草稿。
- `treatment_id` 必填：血透收费归集必然绑定一次治疗，不允许为 NULL。已设为 `NOT NULL`，部分唯一索引才不会因 NULL 不冲突而失效。
- `total_amount` 只统计 `billable=true` 且 `amount IS NOT NULL` 的明细。
- `status` 本期只允许 `draft`、`confirmed`、`checked`、`cancelled`；`pushed`、`settled` 保留但不进入正常流程。
- GORM 模型、API DTO、前端 TypeScript 类型必须与 DDL 同步，尤其 `exported_at`、`pushed_at`、`total_amount`、`status`。
- 时间类型统一用 `timestamptz`，与 `his_price_item` 的 `timestamp` 区分清楚：归集清单需要带时区（跨班次跨日期），HIS 镜像保持 Oracle `DATE` 语义用 `timestamp`。

### 4.3 `charge_line`

用途：血透系统生成的收费归集清单明细。

在 C4 原 DDL 基础上增加 HIS 价表关联和匹配状态：

```sql
-- 23. charge_line 收费归集清单明细（规则C4）
CREATE TABLE IF NOT EXISTS charge_line (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    charge_record_id varchar(36) NOT NULL,
    category varchar(16) NOT NULL,
    item_code varchar(64),
    item_name varchar(128) NOT NULL,
    spec varchar(64),
    unit varchar(16),
    quantity decimal(10,2),
    unit_price decimal(10,2),
    amount decimal(10,2),
    billable boolean NOT NULL DEFAULT true,
    source varchar(8) NOT NULL DEFAULT 'auto',
    charge_item_id bigint,
    his_price_item_id varchar(36),
    his_item_code varchar(20),
    his_item_class varchar(1),
    his_item_name varchar(120),
    price_source varchar(32),
    matched_status varchar(16),
    note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cl_record ON charge_line (charge_record_id);
CREATE INDEX IF NOT EXISTS idx_cl_his_item_code ON charge_line (his_item_code);
CREATE INDEX IF NOT EXISTS idx_cl_match_status ON charge_line (matched_status);
CREATE INDEX IF NOT EXISTS idx_cl_category ON charge_line (category);

COMMENT ON TABLE charge_line IS '收费归集清单明细（规则C4）';
```

字段枚举建议：

| 字段 | 允许值 |
|---|---|
| `category` | `treatment`、`material`、`nursing`、`injection`、`drug` |
| `source` | `auto`、`manual` |
| `price_source` | `his_price_list`、`billing_catalog`、`manual`、`unknown` |
| `matched_status` | `matched`、`multiple`、`unmatched`、`manual` |

一致性要求：

- `quantity` 默认按 1 处理，但落库前必须明确写入，避免前端和后端计算差异。
- `amount = quantity * unit_price` 由后端统一计算，前端只展示。
- `unit_price=NULL` 表示待核价，不应按 0 元自动计入合计。
- 删除明细只允许在 `draft` 或 `confirmed` 状态下执行，`checked` 后锁定。
- `charge_record_id` 不在 DDL 加外键约束（便于部署/迁移），但服务层必须在删除清单时级联删除所属明细，或在 DDL 后补 `ON DELETE CASCADE` 约束由 DBA 评估。
- `unit` 与 `his_price_item.units` 命名不一致（复数 vs 单数），这是为了让清单明细面向"实际使用单位"，HIS 镜像保持 HIS 原字段。映射时由服务层负责拷贝。

### 4.4 后续预留表

以下表本期不建议创建，等 HIS 费用清单表结构确认后再落地：

- `his_charge_record`
- `his_charge_line`
- `charge_reconcile_result`

## 5. HIS `price_list` 同步设计

### 5.1 Oracle 查询 SQL

初版采用全量同步，按 `item_code` upsert。

```sql
SELECT
    item_class,
    item_code,
    item_name,
    item_spec,
    units,
    price,
    prefer_price,
    foreigner_price,
    performed_by,
    fee_type_mask,
    class_on_inp_rcpt,
    class_on_outp_rcpt,
    class_on_reckoning,
    subj_code,
    class_on_mr,
    memo,
    start_date,
    stop_date,
    operator,
    enter_date,
    high_price,
    material_code,
    score_1,
    score_2,
    price_name_code,
    control_flag,
    input_code,
    input_code_wb,
    std_code_1,
    changed_memo,
    class_on_insur_mr,
    package_spec,
    firm_id,
    charge_according,
    license_id,
    update_flag,
    dept_name,
    update_flag_syb,
    mr_bill_class,
    class_on_mr_add,
    cwtj_code,
    high_value,
    drg_code,
    insur_update,
    stop_operator,
    limit_quantity
FROM his.price_list
```

分页示例（Oracle 风格，与现有 `his_oracle/client.go` 一致）：

```sql
SELECT * FROM (
    SELECT inner_q.*, ROWNUM rn FROM (
        SELECT /* 同上字段列表 */
        FROM his.price_list
        ORDER BY item_code
    ) inner_q WHERE ROWNUM <= :end_row
) WHERE rn > :start_row
```

分页参数由后端按 `batch_size` 推进，不依赖 `cursor_value`。

### 5.2 启用状态判断

本地 `is_active` 建议按以下规则计算：

```text
stop_date IS NULL OR stop_date > 当前时间
```

如果后续确认 `control_flag` 或 `update_flag` 表示停用/删除，再补充规则。

### 5.2.1 大数据量与超时保护

`price_list` 可能是全院级价表，数据量大于血透实际使用项目。本期同步必须做保护：

- Oracle 查询必须分页，默认每批 `1000` 条，可配置。
- 单次同步必须设置超时，默认 `60` 秒，可配置到 `sync_job_configs.timeout_seconds`。
- 每批写入本地库后更新运行计数，避免一次事务过大。
- 如果中途失败，已写入批次保留，`sync_job_runs.status` 记为 `partial` 或 `failed`，错误写入 `error_message`。
- 手动同步按钮应防重复点击；后端也要检测同一 `job_code` 是否已有 running 任务。
- 搜索接口默认 `activeOnly=true`，分页返回，禁止无条件返回全表。
- `enabled=false` 只控制自动定时调度，不影响手动点击同步按钮触发；手动入口调用同步服务，不受 `enabled` 限制。

### 5.3 同步任务编码

在 `internal/models/sync_job.go` 增加：

```go
SyncJobCodeHisPriceList = "his_price_list"
SyncTypeHisPriceList = "his_price_list"
```

建议默认配置：

| 字段 | 值 |
|---|---|
| `job_code` | `his_price_list` |
| `source_system` | `HIS_ORACLE` |
| `sync_type` | `his_price_list` |
| `enabled` | `false` |
| `batch_size` | `1000` |
| `cursor_type` | `full` |
| `enabled` | `false`（控制自动调度）|

说明：当前模型常量只有 `time`、`mixed`，但 HIS 价表初版是全量同步，不建议用 `mixed` 表达无游标任务，避免误导后续维护。建议新增：

```go
CursorTypeFull = "full"
```

如果为了最小改动暂不新增常量，也必须在文档和代码注释中明确：该任务忽略 `cursor_value`，每次按 `item_code` 全量 upsert。

### 5.4 本地 upsert 规则

唯一键：`source_system + item_code`。

同步行为：

- HIS 有、本地无：新增。
- HIS 有、本地有：覆盖更新 HIS 字段。
- HIS 无、本地有：本期不删除，可保留；后续如需可标记 `is_active=false`。
- `synced_at` 每次同步更新。

失败补偿：

- 同步任务失败后可直接重跑，upsert 保证幂等。
- 若某批次写入失败，不回滚已成功批次；运行记录标记 `partial` 或 `failed`。
- 前端应展示最近一次同步时间、状态、成功/失败数量。
- 收费归集时如果 HIS 价表为空，应降级使用 `billing_catalog.json` 和待核价，不阻断生成草稿。

## 6. 收费归集查价规则

查价优先级：

1. `his_price_item` 精确匹配。
2. `billing_catalog.json` 固定规则兜底。
3. 人工补录价格。
4. 待核价。

### 6.1 药品

来源：`medication_admin` 给药记录。

匹配顺序：

1. 如果给药记录已有 HIS 编码，按 `his_price_item.item_code` 精确匹配。
2. 否则按药名匹配 `item_name`，限定 `item_class IN ('A', 'B')`。
3. 多条匹配时前端提示人工选择。
4. 未匹配时仍进入清单，`matched_status='unmatched'`，`unit_price=NULL`。

### 6.2 耗材

来源：`Treatment_MaterialTrace`。

匹配顺序：

1. 如果老库耗材能拿到 HIS 编码，按 `item_code` 精确匹配。
2. 否则按耗材名称/规格匹配 `his_price_item`，限定 `item_class='I'`。
3. C4 catalog 中不可计费项继续按不可计费展示，不计入合计。
4. 未匹配时显示“待核价”。

与现有 `Stock_ChargeItem` 的关系：

- `Stock_ChargeItem` 继续作为老血透系统库存/耗材业务来源，不改老表结构。
- `Treatment_MaterialTrace.ChargeItemId` 仍优先关联 `Stock_ChargeItem.Id` 获取血透侧耗材名称、单位、数量。
- `his_price_item` 只提供 HIS 收费编码、HIS 名称、价格和停用状态。
- 两者之间本期不新增强制映射表；先通过编码、名称、规格匹配，无法唯一匹配时人工选择并写入 `charge_line.his_item_code`。
- 不允许把 HIS `price_list` 反写到 `Stock_ChargeItem`。

### 6.3 治疗费

来源：治疗模式。

优先匹配 `item_class='E'` 的 HIS 项目。

如果无法匹配，使用 `billing_catalog.json` 中固定治疗规则。

### 6.4 护理费

来源：血管通路类型。

优先匹配 `item_class='K'` 的 HIS 项目。

如果无法匹配，使用 `billing_catalog.json` 中固定护理规则。

### 6.5 注射费

来源：本次治疗是否存在静脉给药。

由于注射费可能在 HIS 中归为治疗或护理，本期允许：

- 优先按项目名称“静脉注射”匹配。
- `item_class` 可放宽为 `E` 或 `K`。
- 多条匹配时人工确认。

## 7. HIS 推送禁用策略

C4 原交付中 `NoopPusher` 会把清单标记为 `pushed`，本期需要调整。

本期策略：

- 可以保留 `HISChargePusher` interface。
- 不注册 `/charges/:id/push` 路由，或注册后固定返回 `501`。
- 前端不显示“推送 HIS”按钮。
- `DefaultPusher` 不参与业务流程。
- 清单导出 Excel 后只记录 `exported_at`，不写 `pushed_at`。
- 状态机不自动进入 `pushed`。

建议状态流：

```text
draft -> confirmed -> checked -> cancelled
```

导出 Excel 是动作，不是状态：

```text
checked --export--> checked + exported_at
```

后续真实 HIS 推送启用后再开放：

```text
checked -> pushed -> settled / cancelled
```

## 7.1 权限与审计要求

收费归集涉及费用金额，必须限制关键操作权限并记录操作者。

权限建议：

| 操作 | 建议角色 |
|---|---|
| 查看清单 | 医生、护士、护士长、管理员 |
| 生成草稿 | 护士、护士长、管理员 |
| 编辑明细 | 护士、护士长、管理员 |
| 确认清单 | 护士、护士长、管理员 |
| 双人核对 | 护士、护士长、管理员 |
| 取消清单 | 护士长、管理员，或确认前允许原记录人取消 |
| 同步 HIS 价表 | 管理员或系统同步管理员 |

审计要求：

- `charge_record.recorded_by/recorded_name` 记录确认人或记账人。
- `charge_record.checked_by/checked_name/checked_at` 记录核对人。
- 手工新增、修改、删除 `charge_line` 应在服务日志中记录清单 ID、行 ID、操作者、修改前后关键值。
- HIS 价表同步必须写 `sync_job_runs`，包含开始时间、结束时间、成功/失败数量和错误信息。
- 后续如需要更强审计，可新增 `charge_audit_log`，本期先不建表。

双人核对规则：

- C4 原规则“不强制核对人不等于记账人”，本期保持该规则。
- 前端仍应展示确认人与核对人，方便现场管理追踪。

## 8. 后端文件执行计划

### 8.1 新增文件

从 C4 patch 手工提取并按当前代码调整：

- `ai-hms-backend/internal/config/billing_catalog.json`
- `ai-hms-backend/internal/config/billing_catalog.go`
- `ai-hms-backend/internal/config/billing_catalog_test.go`
- `ai-hms-backend/internal/models/charge_record.go`
- `ai-hms-backend/internal/models/charge_line.go`
- `ai-hms-backend/internal/services/billing_service.go`
- `ai-hms-backend/internal/services/billing_service_test.go`
- `ai-hms-backend/internal/services/billing_pusher.go`
- `ai-hms-backend/internal/api/v1/billing_handler.go`

新增 HIS 价表同步文件：

- `ai-hms-backend/internal/models/his_price_item.go`
- `ai-hms-backend/internal/integrations/his_oracle/price_list.go`
- `ai-hms-backend/internal/services/his_price_sync_service.go`
- `ai-hms-backend/internal/services/his_price_service.go`
- `ai-hms-backend/internal/api/v1/his_price_handler.go`

### 8.2 修改文件

公共文件必须手工融合：

- `ai-hms-backend/cmd/server/main.go`
- `ai-hms-backend/internal/database/health.go`
- `ai-hms-backend/internal/models/sync_job.go`
- `docs/sql/deploy_new_tables.sql`

### 8.3 `main.go` 路由原则

注册：

```go
v1api.RegisterBillingRoutes(protected)
v1api.RegisterHisPriceRoutes(protected, cfg.HisOracle, cfg.LegacyTenantID)
```

不注册或禁用：

```go
// RegisterChargePushRoutes 本期不启用
```

如果保留 `/charges/:id/push`，handler 必须返回：

```text
501 HIS 收费推送接口暂未启用，请导出 Excel 后人工录入 HIS
```

## 9. 前端文件执行计划

### 9.1 新增文件

- `ai-hms-frontend/src/services/billingApi.ts`
- `ai-hms-frontend/src/services/hisPriceApi.ts`
- `ai-hms-frontend/src/lib/billingExcel.ts`
- `ai-hms-frontend/src/pages/patient-detail/tabs/BillingTab.tsx`

### 9.2 修改文件

- `ai-hms-frontend/src/services/index.ts`
- `ai-hms-frontend/src/pages/patient-detail/tabs/index.ts`
- `ai-hms-frontend/src/pages/PatientDetail.tsx`

### 9.3 UI 要求

患者详情侧栏新增：

```text
收费归集
```

Tab 内按钮：

- 生成归集草稿
- 确认
- 双人核对
- 导出 Excel
- 取消
- 添加/编辑/删除明细
- HIS 价表搜索/匹配

不显示：

- 推送 HIS
- HIS 结算
- HIS 退费
- HIS 对账

### 9.4 患者详情入口上下文风险

C4 原 README 已说明：患者详情页 `Patient` 类型不一定包含 `currentTreatmentId` / `currentPrescriptionId`。

本期处理原则：

- `BillingTab` 可以展示该患者历史收费清单。
- 如果缺少当前治疗上下文，不允许直接生成草稿，前端提示“请在治疗执行页发起归集，或先选择本次治疗”。
- 更推荐后续把“生成归集草稿”接入透析执行流，因为规则来源是“治疗启动后归集”。
- 如果本期必须在患者详情页生成，后端需要提供“查询患者最近未收费治疗记录”的接口，不能由前端猜测治疗 ID。
- `PatientDetail.tsx` 只能手工追加侧栏项和 switch 分支，不得整体替换当前文件。

## 10. API 设计

### 10.1 收费归集 API

| 方法 | 路径 | 本期状态 | 说明 |
|---|---|---|---|
| POST | `/api/v1/charges/build` | 启用 | 生成归集草稿 |
| GET | `/api/v1/charges` | 启用 | 清单列表 |
| GET | `/api/v1/charges/:id` | 启用 | 清单详情 |
| POST | `/api/v1/charges/:id/lines` | 启用 | 新增手工明细 |
| PATCH | `/api/v1/charges/lines/:lineId` | 启用 | 修改明细 |
| DELETE | `/api/v1/charges/lines/:lineId` | 启用 | 删除明细 |
| POST | `/api/v1/charges/:id/confirm` | 启用 | 护士确认 |
| POST | `/api/v1/charges/:id/check` | 启用 | 双人核对 |
| POST | `/api/v1/charges/:id/exported` | 启用 | 记录 Excel 已导出 |
| POST | `/api/v1/charges/:id/cancel` | 启用 | 取消清单 |
| POST | `/api/v1/charges/:id/push` | 禁用 | 本期不启用 |

API 行为约束：

- `POST /charges/build` 对同一 `tenant_id + treatment_id` 必须幂等；已有非取消清单时返回现有清单，不重复创建。
- `confirm` 只允许从 `draft` 进入 `confirmed`。
- `check` 只允许从 `confirmed` 进入 `checked`。
- `exported` 只记录导出时间，不改变状态。
- `cancel` 必须记录原因到 `note` 或扩展字段，不能静默取消。
- 所有修改类接口必须从登录上下文取操作者，不能信任前端传入的操作者姓名。

### 10.2 HIS 价表 API

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/his-price-items` | 搜索本地 HIS 价表 |
| POST | `/api/v1/his-price-items/sync` | 手动同步 HIS `price_list` |
| GET | `/api/v1/his-price-items/:itemCode` | 按编码查询 |

搜索参数建议：

| 参数 | 说明 |
|---|---|
| `keyword` | 名称、拼音码、五笔码、编码 |
| `itemClass` | A/B/E/I/K 等 |
| `activeOnly` | 是否仅启用项目 |
| `page` | 页码 |
| `pageSize` | 每页数量 |

## 11. 同步中心接入

现有同步中心已有检查报告同步和任务运行记录能力。

本期建议：

- 后端先接入 `sync_job_configs` / `sync_job_runs`。
- 前端同步中心是否展示 HIS 价表任务可作为小步增强。
- 手动同步按钮可以先放在收费归集 Tab 的“价表管理/同步”入口，或同步中心。

## 12. 实施步骤

### 阶段 1：准备与静态融合

- [ ] 读取 C4 patch，确认新增文件内容。
- [ ] 不执行 `git am`，手工提取新增文件。
- [ ] 调整 Go package import，适配当前仓库路径。
- [ ] 手工合并 `deploy_new_tables.sql`，加入表 22/23/24（核对当前脚本已到 21 `consent_record`，且其后的 `COMPLICATION` seed 事务仍在末尾）。
- [ ] 手工合并 `health.go`，加入 `charge_record`、`charge_line`、`his_price_item`。

### 阶段 2：HIS 价表同步

- [ ] 新增 `HisPriceItem` 模型。
- [ ] 新增 Oracle `QueryPriceList` 查询。
- [ ] 新增 `HisPriceSyncService`。
- [ ] 新增 `HisPriceService` 搜索能力。
- [ ] 新增 `HisPriceHandler`。
- [ ] 新增 sync job code。
- [ ] 新增 `CursorTypeFull` 或在代码注释中明确 full sync 行为。
- [ ] 增加 running 任务互斥，防止重复同步。
- [ ] 编写同步服务测试，Oracle 查询用接口/桩隔离。

### 阶段 3：C4 服务整改

- [ ] 新增/调整 `ChargeRecord`、`ChargeLine` 模型。
- [ ] 合入 `BillingService`。
- [ ] 修改查价逻辑：优先查 `his_price_item`。
- [ ] 移除或禁用 `MarkPushed` 正常业务入口。
- [ ] 增加 `MarkExported`。
- [ ] 保留 `HISChargePusher` 接口但不启用。
- [ ] 补状态机测试。
- [ ] 补同一治疗重复 build 幂等测试。
- [ ] 补并发 build 冲突测试。
- [ ] 补 checked 后明细不可编辑测试。

### 阶段 4：API 装配

- [ ] 注册收费归集 API。
- [ ] 注册 HIS 价表 API。
- [ ] 不注册 push 路由，或 push 返回 501。
- [ ] 确认所有接口保留鉴权。

### 阶段 5：前端融合

- [ ] 新增 `billingApi.ts`。
- [ ] 新增 `hisPriceApi.ts`。
- [ ] 新增 `billingExcel.ts`。
- [ ] 新增 `BillingTab.tsx`。
- [ ] 手工修改 `PatientDetail.tsx`，增加“收费归集”。
- [ ] 不显示“推送 HIS”按钮。
- [ ] 导出 Excel 后调用 `markExported`。

### 阶段 6：验证

- [ ] 后端服务测试。
- [ ] 后端构建。
- [ ] 后端 vet。
- [ ] 前端 build。
- [ ] 有 HIS Oracle 环境时执行 `price_list` 手动同步冒烟。

## 13. 测试计划

### 13.1 后端命令

```powershell
go test ./internal/services -count=1
go test ./internal/config ./internal/api/v1 -count=1
go build -o "$env:TEMP\check.exe" ./cmd/server
go vet ./internal/services ./internal/api/v1 ./internal/database ./internal/config ./cmd/server
```

说明：Windows 环境避免默认执行 `go build ./...`，以免 Oracle/外部驱动环境造成无关失败。

### 13.2 前端命令

```powershell
npm run build
```

### 13.3 HIS 价表冒烟

在具备 HIS Oracle 连接的环境执行：

- [ ] 测试 HIS Oracle 连接成功。
- [ ] 同步 `price_list` 前一批数据。
- [ ] 检查 `his_price_item` 总数。
- [ ] 按 `item_code` 精确查询。
- [ ] 按药品名称模糊查询，限定 `item_class=A`。
- [ ] 按材料名称模糊查询，限定 `item_class=I`。
- [ ] 检查 `stop_date` 已过期项目是否 `is_active=false`。

### 13.4 C4 业务冒烟

- [ ] 对一个有治疗记录的患者生成收费草稿。
- [ ] 确认治疗费行存在。
- [ ] 确认耗材行来自 `Treatment_MaterialTrace`。
- [ ] 确认药品行来自 `medication_admin`。
- [ ] 已匹配 HIS 价表的行显示 `his_item_code` 和参考价。
- [ ] 未匹配行显示待核价。
- [ ] 确认、核对流程成功。
- [ ] 导出 Excel 成功，并记录 `exported_at`。
- [ ] 前端无“推送 HIS”入口。

### 13.5 并发与失败场景

- [ ] 同一 `treatment_id` 连续两次生成草稿，只返回同一张非取消清单。
- [ ] 同一 `treatment_id` 并发生成草稿，不产生两张非取消清单。
- [ ] HIS 价表为空时，仍可生成清单，未匹配项目显示待核价。
- [ ] HIS 价表同步中断后可重跑，已同步项目不重复。
- [ ] `checked` 状态后编辑、删除明细接口返回失败。
- [ ] 未授权角色调用确认、核对、同步价表接口返回失败。

## 14. 风险与处理

| 风险 | 处理 |
|---|---|
| C4 patch 基线过旧 | 不直接套 patch，手工融合公共文件 |
| `price_list` 数据量大 | 分批查询，默认 batch size 1000 |
| 无更新时间字段 | 初版全量同步 + upsert |
| 全量同步耗时长 | 任务超时、分批提交、running 互斥、失败可重跑 |
| 名称匹配不唯一 | 前端提供人工选择，不自动猜测 |
| 患者详情缺少当前治疗上下文 | 只展示历史清单；生成草稿需治疗执行页或后端查询当前治疗 |
| 重复生成收费草稿 | `tenant_id + treatment_id` 非取消唯一索引 + build 幂等 |
| 金额操作权限过宽 | 按角色限制确认、核对、取消、同步价表 |
| 注射费分类不确定 | `E/K` 均允许匹配，需现场确认后收敛 |
| HIS 推送误启用 | 默认不注册路由，前端不显示按钮 |
| HIS 实际费用表未知 | 本期不做对账，只预留设计 |
| 老库 DDL 风险 | 新表只写部署 SQL，不运行时建表 |
| `treatment_id=NULL` 导致唯一索引失效 | `treatment_id` 设 `NOT NULL`，血透场景必绑治疗 |
| `charge_line` 孤儿行 | 服务层删除清单时级联删除明细，或建议 DBA 加 `ON DELETE CASCADE` |
| 时间类型跨表不一致 | 清单用 `timestamptz`，HIS 镜像用 `timestamp`，服务层负责转换 |

## 15. 待后续确认

- HIS 实际费用清单表名。
- HIS 实际费用清单字段：患者号、住院号/门诊号、费用日期、项目编码、项目名称、规格、单位、数量、单价、金额、收费状态、退费标识。
- 血透患者与 HIS 患者/住院号的关联优先级。
- 注射费在 HIS 中的准确 `item_code` 和 `item_class`。
- 治疗费、护理费固定项目是否能提供标准 `item_code` 清单。
- 是否需要同步停用历史价，或仅保留当前启用价格。
- 收费归集相关角色权限是否按“护士/护士长/管理员”执行，还是需要新增收费员角色。
- 是否需要新增独立 `charge_audit_log` 表满足审计要求。

## 16. 最终交付标准

- `charge_record`、`charge_line`、`his_price_item` DDL 已追加到 `docs/sql/deploy_new_tables.sql`。
- `health.go` 能检查三张表。
- HIS `price_list` 可手动同步到 `his_price_item`。
- 收费归集优先使用 HIS 价表查价。
- 收费清单可确认、核对、导出 Excel。
- HIS 推送按钮不显示，后端不启用真实推送。
- 同一治疗不会产生多张非取消收费清单。
- HIS 价表同步失败可重跑，不影响治疗主流程。
- 后端测试、构建、vet 通过。
- 前端 build 通过。
