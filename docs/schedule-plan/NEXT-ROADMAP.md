# 排班扩展表融合路线图

> 状态: 阶段 3 已验证, 后续阶段待确认执行  
> 日期: 2026-06-07  
> 适用范围: `ai-hms-backend/`, `ai-hms-frontend/`, `docs/sql/`  
> 重要说明: 本文基于“老表主记录 + Schedule_* 扩展表”的融合路线。它不同于早期 `docs/schedule-plan/README.md` 中“只改老表字段、不新建表”的保守路线。后续执行前必须由用户确认采用哪条路线, 不允许混用。

## 1. 当前结论

当前仓库已完成排班融合阶段 0 到阶段 3:

| 阶段 | 内容 | 状态 |
| --- | --- | --- |
| 阶段 0 | 扩展表 DDL、Status=60 审计 SQL、字段映射、扩展模型 | 已完成 |
| 阶段 0.5 | 连接池、健康检查、5xx 脱敏、请求日志、权限错误修正 | 已完成 |
| 阶段 1 | WardExt、BedMachineExt、PatientProfile、TenantSetting、Calendar 配置 CRUD | 已完成 |
| 阶段 2 | Board 快照、只读预检、冲突查询入口 | 已完成 |
| 阶段 3 | 独立模板替代 `Schedule_PatientShift.Status=60` | 已完成 |

阶段 3 的核心状态:

- 模板读写已切到 `Schedule_ScheduleTemplate` + `Schedule_ScheduleTemplateItem`。
- 旧 `Status=60` 仅保留审计, 不再作为模板保存或应用来源。
- `ApplyTemplate` 只按显式模板项生成草稿, 写 `Schedule_PatientShift.Status=10` 和 `Schedule_PatientShiftExt.RuleStatus=10`。
- 前端已完成模板列表、编辑头信息、应用模板弹窗的基础适配。
- 当前没有编译或测试阻塞。

## 2. 已验证命令

最近一次核查结果:

| 范围 | 命令 | 结果 |
| --- | --- | --- |
| 后端排班聚焦测试 | `go test ./internal/services -count=1 -short -v -run "Test(Save|List|Apply|RunPrecheck|LoadBoard|Upsert|Create|Update|Delete|Validate|Calendar|Bed|Ward|PatientProfile|TenantSetting)"` | PASS |
| 后端服务/API 覆盖率 | `go test ./internal/services ./internal/api/v1 -count=1 -short -cover` | PASS, services 10.1%, api/v1 0.0% |
| 后端全量测试 | `go test ./... -count=1 -short` | PASS |
| 后端构建与 vet | `go build -o "$env:TEMP\\ai-hms-backend-check.exe" ./cmd/server`; `go vet ./...` | PASS |
| 前端 lint | `npm run lint` | PASS, 0 errors, 3 warnings |
| 前端构建 | `npm run build` | PASS |
| v1.2 参考实现全量测试 | `go test ./... -count=1 -short` | PASS |
| v1.2 算法核心 | `go test ./internal/sched/ -count=1 -v` | PASS, 8/8 |

注意: 前端 lint 的 3 个 warning 属于既有非排班模板文件, 不作为本轮阻塞。

## 3. 当前能力矩阵

| 能力 | 当前仓库 | 说明 |
| --- | --- | --- |
| 老表真实排班主记录 | 已有 | 继续写 `Schedule_PatientShift` |
| 扩展表模型 | 已有 | `internal/models/schedule_ext.go` |
| 扩展 DDL | 已有 | `docs/sql/schedule_extension_tables.sql`, 只生成不执行 |
| 配置维护 | 已有 | 病区扩展、机器能力、患者骨架、日历、租户配置 |
| Board 快照 | 已有 | 面向单日只读预检 |
| 预检 | 已有 | 容量、机器能力、日历、停机、占用、CRRT 基础检查 |
| 独立模板 | 已有 | 替代 `Status=60` |
| 应用模板 | 已有 | 生成草稿, 不做自动算法 |
| 自动排班引擎 | 未实现 | 参考 v1.2 `internal/sched` |
| 草稿批量生成幂等 | 未实现 | 尚未接 `ON CONFLICT DO NOTHING` 或 DB 唯一约束 |
| 冲突队列写入 | 未实现 | 当前只读预检, 阶段 3 不写 `Schedule_ConflictQueue` |
| 冲突分页/resolve | 未实现 | v1.2 有参考实现 |
| 假日值班部分开放 | 部分基础 | Calendar/OpenWard/OpenBed 表与配置存在, 生成逻辑未接入 |
| 质量评分 | 未实现 | v1.2 有 `quality_service.go` 参考 |
| 停机扰动/补透/方案变更/CRRT 完整流程 | 未实现 | 后续阶段处理 |
| 模板项增删改 UI | 未实现 | 当前只能编辑模板头并保存已有 items |
| API handler 测试 | 缺口大 | `api/v1` 覆盖率 0.0% |

## 4. v1.2 可迁移能力

`docs/排班功能说明/透析排班-backend-v1.2/backend` 可作为后续实现参考, 但不能直接覆盖当前仓库。

可借鉴模块:

| v1.2 文件 | 可借鉴点 | 当前适配要求 |
| --- | --- | --- |
| `internal/sched/engine.go` | 两轮分配、HDF 奇偶周、排不开生成冲突 | 改为读取当前 `Schedule_*Ext` 和老表 Bed/Ward/Shift |
| `internal/sched/board.go` | 内存 Board、占用、机器能力、日历/停机判断 | 与当前 `ScheduleBoardService` 合并或抽象为纯算法输入 |
| `internal/repo/repo.go` | LoadBoard、SaveDrafts、SaveConflicts、幂等写入 | SQL 必须适配老库大小写和扩展表结构 |
| `internal/service/schedule_service.go` | `GenerateSchedule` 编排 | 不直接引入 AutoMigrate 或演示模型 |
| `internal/service/template_service.go` | 从 PatientProfile 重建模板 | 可迁到当前 `ScheduleTemplateService` |
| `internal/service/quality_service.go` | 达标率、利用率、稳定率、综合分 | 需定义当前仓库 DTO 和前端卡片 |
| `internal/service/integration_test.go` | `TEST_DATABASE_URL` 守卫的 PostgreSQL 集成测试 | 可复制测试模式, 不复制 schema 初始化策略 |

不能照搬的部分:

- v1.2 自动迁移建表逻辑;当前仓库永久禁止 AutoMigrate。
- v1.2 演示鉴权和本地模型;当前仓库必须使用已有 JWT/中间件与老库主档。
- v1.2 简化表结构;当前仓库需要保留 `Schedule_PatientShift` 主记录和大小写敏感列名。

## 5. 后续优先级

### P0 · 安全网与测试

目标: 在不进入自动生成算法前, 先补并发安全、幂等口径和测试防线。

任务:

1. 新增人工审核 SQL 草案, 不执行 DDL。
2. 设计 PostgreSQL 唯一索引:
   - 同租户、同患者、同日期、同班次不重复。
   - 同租户、同床位、同日期、同班次只允许一人。
   - 排除取消、缺席、转出等无效状态。
   - 排除 CRRT 或跨班记录, 避免与规律三班网格混用。
3. 服务层捕获唯一约束冲突, 转为友好业务错误。
4. 补 `schedule_handler`、`schedule_config_handler`、`schedule_rules_handler` 的 API handler 测试。
5. 补模板应用幂等/重复应用边界测试。

验收:

- SQL 只落文件, 文件顶部明确“DBA 审核后手工执行, 应用不会自动执行”。
- `go test ./internal/services ./internal/api/v1 -count=1 -short` 通过。
- `api/v1` 覆盖率不再为 0.0%。

### P1 · 模板维护闭环

目标: 让独立模板从“可应用”变为“可维护”。

任务:

1. 增加模板详情接口, 支持按模板 ID 获取头 + items。
2. 增加模板项增删改后端接口或扩展 SaveTemplate 的 items 全量保存语义。
3. 前端 `ScheduleTemplateEditor` 支持新增、删除、编辑模板项。
4. 从 `Schedule_PatientProfile` 重建模板, 参考 v1.2 `RebuildTemplateFromProfiles`。
5. 明确 `ShiftTiming` 和 `PatientPlanId` 来源, 不再硬编码或置空。

验收:

- 无需 `Status=60` 即可新建一套可应用模板。
- 模板项校验覆盖 `Scope`、`WardId`、机器能力、HDF 星期与频率。
- 前端可完成“创建模板 -> 添加模板项 -> 保存 -> 应用到日期”的闭环。

### P2 · 自动生成引擎沙箱

目标: 先接入纯算法, 只生成内存结果和冲突预览, 不写库。

任务:

1. 新增独立算法包, 迁入 v1.2 的周序号、奇偶周、HDF 替换、机型能力测试。
2. 将当前 `ScheduleBoardService` 输出转换为算法 Board 输入。
3. 实现 `PreviewGenerateSchedule`, 返回 drafts + conflicts。
4. 前端增加“生成预览”入口, 展示拟排结果和冲突, 不保存。

验收:

- v1.2 `internal/sched` 的 8 个核心测试在当前仓库有等价测试。
- 预览接口不会写 `Schedule_PatientShift`、`Schedule_PatientShiftExt`、`Schedule_ConflictQueue`。
- 排不开只返回冲突, 不自动挪位。

### P3 · 写入草稿与冲突队列

目标: 在 P0 唯一约束和 P2 预览稳定后, 才允许批量写草稿。

任务:

1. 实现 `GenerateSchedule` 写入事务。
2. 写 `Schedule_PatientShift.Status=10` 和 `Schedule_PatientShiftExt.RuleStatus=10`。
3. 写 `Schedule_ConflictQueue`。
4. 使用 DB 唯一约束或 `ON CONFLICT DO NOTHING` 保证重复生成幂等。
5. 增加 `GET /api/v1/schedule/conflicts?limit=&offset=`。
6. 增加 `POST /api/v1/schedule/conflicts/:id/resolve`。

验收:

- 重复生成同一周期不会重复插入有效排班。
- 并发写同一患者/床位由数据库兜底。
- 冲突处理仍是人工裁决, 系统不自动解决冲突。

### P4 · 扰动与评分

目标: 补齐临床扰动场景和运营评价。

任务:

1. 假日值班部分开放接入生成逻辑。
2. 设备停机迁移建议。
3. 补透建议。
4. 方案变更影响未确认排班。
5. CRRT 独立占机流程。
6. 排班质量评分接口和前端卡片。

验收:

- 停机、假日、补透、方案变更均只生成建议或待确认草稿。
- 不自动改已确认/治疗中/历史排班。
- 质量评分指标有单元测试覆盖。

## 6. 红线

- 不恢复 `AutoMigrate`, 不在应用启动时执行 DDL。
- 不执行 `docs/sql/schedule_extension_tables.sql` 或新增 SQL; SQL 只能人工审核后手工执行。
- 不迁移、删除、继续写入 `Status=60` 模板。
- 不把 v1.2 参考实现整包复制进当前仓库。
- 不绕过当前 JWT/权限中间件引入默认超级用户。
- 不在状态口径未确认前开发依赖“完成/取消/转出”的统计报表。
- 不自动解决冲突; 系统只给建议, 人工确认后才写入。

## 7. 建议下一步

如果继续编码, 建议从 P0 开始:

1. 新增 `docs/sql/schedule_patient_shift_unique_constraints.sql` 草案, 只供 DBA 审核。
2. 给模板应用和配置/规则 handler 补 API 测试。
3. 为唯一约束冲突设计统一业务错误和测试。

不要直接进入自动排班引擎; 当前更大的风险是并发安全网与 API 覆盖率不足。
