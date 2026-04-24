# AI-HMS 交接说明文档

更新时间：2026-04-24  
用途：给后续 Codex 接手开发、联调和老库迁移用的快速说明。

## 1. 当前整体状态

本轮工作主要分成三条线：

1. 老血透数据库迁移持续推进，后端多个患者管理、治疗执行、诊疗配置接口已切到老库或老库优先。
2. 透析执行前端已从单文件页面迁移为模块化结构，并完成了主流程真实联调。
3. 字典配置页面已从卡片式平铺改成“左侧分类 + 右侧维护”的结构，并按业务口径区分“普通业务字典”和“其他字典”。

当前仓库处于“功能持续重构 + 老库迁移并行”的状态，交接时不要把已有的未提交修改直接覆盖掉。

## 2. 已完成的主要工作

### 2.1 老血透数据库迁移

已完成/已验证的迁移方向主要包括：

- 患者管理相关接口逐步切换到老库表
- 治疗详情、单条治疗详情、治疗记录单弹窗、透前/透后评估、透中监测等字段逐步对照老库
- 诊疗配置中的方案模板、医嘱模板、材料目录、药品目录已改为老库读取优先
- 设备管理已按老库表补齐主要字段
- 字典服务已优先读取 `CodeDictionary_CodeDictionarys`

已形成的关键文档：

- [legacy-migration-session-summary-2026-04-21.md](/F:/python/前后端代码/ai-hms_qhd/docs/legacy-migration-session-summary-2026-04-21.md)
- [treatment-execution-legacy-dev-record-2026-04-21.md](/F:/python/前后端代码/ai-hms_qhd/docs/treatment-execution-legacy-dev-record-2026-04-21.md)
- [dictionary-type-mapping-dev.md](/F:/python/前后端代码/ai-hms_qhd/docs/dictionary-type-mapping-dev.md)
- [legacy-migration-uncertain-field-checklist.md](/F:/python/前后端代码/ai-hms_qhd/docs/legacy-migration-uncertain-field-checklist.md)
- [patient-management-dictionary-uncertain-2026-04-23.md](/F:/python/前后端代码/ai-hms_qhd/docs/patient-management-dictionary-uncertain-2026-04-23.md)

### 2.2 透析执行前端重构

透析执行已不再是原来的单文件页面，主入口改为模块化结构，路径在：

- [ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx](/F:/python/前后端代码/ai-hms_qhd/ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx)

相关子页面已拆分到：

- `ai-hms-frontend/src/pages/dialysis-processing/execution/`

已完成真实联调的页面：

- 透前评估
- 当日处方
- 双人核对
- 透析医嘱
- 透中监测
- 透后评估
- 透析小结
- 健康宣教

当前前端透析执行的关键状态：

- `/dialysis-processing` 路由保持不变
- 页面结构已经切到新版模块化目录
- 真实患者、治疗、处方、核对、医嘱、透中监测、透后评估接口已经接入
- 构建已通过，当前可继续在此基础上做回显校验或字段补齐

### 2.3 字典配置页面改造

字典配置页面已经改造成：

- 左侧字典分类
- 右侧字典值维护表格
- “普通业务字典”点击一级分类后，右侧直接展示该分类下所有相关字典类型的维护表
- “其他字典”需要先选具体二级字典类型，再维护值

已修改页面：

- [ai-hms-frontend/src/pages/DictConfig.tsx](/F:/python/前后端代码/ai-hms_qhd/ai-hms-frontend/src/pages/DictConfig.tsx)

当前字典分类口径：

- `透析治疗`
- `血管通路`
- `医嘱处方`
- `人员信息`
- `临床诊疗`
- `转归记录`
- `其他字典`

当前字典接口仍沿用：

- `GET /api/v1/dict/types`
- `GET /api/v1/dict/items/{typeCode}`
- `GET /api/v1/dict/items/{typeCode}/tree`
- `POST /api/v1/dict/items`
- `PUT /api/v1/dict/items/{id}`
- `DELETE /api/v1/dict/items/{id}`
- `POST /api/v1/dict/items/{id}/toggle`

## 3. 当前已确认的关键口径

### 3.1 字典配置口径

- 普通业务字典不再做二级点击，左侧点一级分类后，右侧直接维护该分类下的字典值。
- 其他字典保留二级目录，先选具体字典类型再维护。
- 后端字典服务优先读取老库 `CodeDictionary_CodeDictionarys`。
- 未命中的 `Type` 会保持为“其他字典”归类。

### 3.2 透析执行口径

- 透前评估、当日处方、双人核对、透中监测、透后评估都已经接入真实接口。
- 当天无治疗记录时，前端会先创建治疗记录再保存对应子表数据。
- 透后提交当前会同步结束治疗状态，这一口径已经做了实现，但仍建议后续根据现场流程再确认一次。

### 3.3 老库迁移口径

- 能明确映射的字段，优先按老库表对照落地。
- 不能明确映射的字段，不做主观猜测，统一记录到交接/待确认文档。
- 前端行为优先保持不变，尽量把兼容适配放在后端。

## 4. 仍需继续确认的内容

以下内容目前仍建议保留在待确认清单，不要直接删掉：

- 透析执行中部分扩展字段的老库口径是否完全一致
- 透中监测里 `dialysateFlow` 是否有明确老库承接字段
- 透后评估里部分 symptom code 和提交即完成治疗的口径是否需要现场再次确认
- 健康宣教模块目前没有明确后端数据源，当前仅保留占位说明
- 月度评估小结的接口和字段口径是否需要新增
- 设备性能检测登记表的字段定义目前仍不完整

这些不确定项已分散记录在现有的 `.md` 文档里，后续接手时优先查看：

- [legacy-migration-session-summary-2026-04-21.md](/F:/python/前后端代码/ai-hms_qhd/docs/legacy-migration-session-summary-2026-04-21.md)
- [legacy-migration-uncertain-field-checklist.md](/F:/python/前后端代码/ai-hms_qhd/docs/legacy-migration-uncertain-field-checklist.md)
- [patient-management-dictionary-uncertain-2026-04-23.md](/F:/python/前后端代码/ai-hms_qhd/docs/patient-management-dictionary-uncertain-2026-04-23.md)

## 5. 建议后续开发顺序

1. 继续做透析执行页面的逐页回显核对，重点看处方、核对、透后评估的字段完整性。
2. 继续按老库口径清理患者管理模块里剩余的不确定字段。
3. 补齐字典配置页面里未覆盖的字典类型归类，如果现场确认有新的业务分类，再更新左侧分类映射。
4. 如果需要继续深挖老库，请优先查阅 `老血透数据库表结构-合并版.md` 和现有迁移文档，不要重新猜表结构。

## 6. 交接提醒

- 当前仓库里有大量未提交修改，接手前先看 `git status`，不要随意回滚非自己改动。
- 先读文档，再改代码，尤其是老库迁移相关功能。
- 任何无法明确映射到老库字段的内容，都要继续写回待确认文档。
