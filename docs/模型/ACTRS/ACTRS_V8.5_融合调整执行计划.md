# ACTRS V8.5 离线部署包融合调整执行计划

## 1. 背景

当前 ai-hms 已融合过 ACTRS 旧版能力，主要代码位于：

| 模块 | 文件 |
|---|---|
| ACTRS Go client | `ai-hms-backend/internal/integrations/actrs/client.go` |
| ACTRS 类型与映射 | `ai-hms-backend/internal/integrations/actrs/types.go`、`mapper.go` |
| 业务服务 | `ai-hms-backend/internal/services/actr_service.go` |
| API handler | `ai-hms-backend/internal/api/v1/actr_handler.go` |
| 前端面板 | `ai-hms-frontend/src/components/actr/ActrPanel.tsx`、`src/services/actrApi.ts` |
| 本地镜像表 | `patient_actr`，DDL 在 `docs/sql/deploy_new_tables.sql` |

旧融合资料在：

`docs/模型/ACTRS/ai-hms-ACTRS契约对拍+部署-交付-20260629(1)`

本次目标是把 ai-hms 与新的：

`docs/模型/ACTRS/ACTRS_V8.5_离线部署包/ACTRS_V8.5_离线部署包`

对齐。该 V8.5 包是容器化 FastAPI 服务，包含 JWT 权限、审计、DICOM、QC、mask 持久化、actr1/actr2/actr_norm 等增强。

### 1.1 版本方向确认（P0，执行前必须确认）

目录名显示本次是从 `ACTRS_V8.6.4_old` 切换到 `ACTRS_V8.5_离线部署包`。虽然业务描述称为“升级”，但从源码能力看，V8.6.4 比 V8.5 包含更多后续能力。因此执行前必须由团队确认：**本轮真实部署目标就是 V8.5 离线包，而不是 V8.6.4 线**。

| 能力 | V8.5 离线包 | V8.6.4_old | 影响 |
|---|---|---|---|
| `/api/xrays/{id}/correction` | 正常端点，query 参数 `correction` | 端点仍存在但测试标记 deprecated | 若目标是 V8.6.4，本计划的 correction 方案需改走新审核流 |
| confirm/reject/revert 审核流 | 无 | 有 | 切到 V8.5 会失去 V8.6.4 的医生审核状态流 |
| `correction_type` / training_corpus | 无 | 有 | 切到 V8.5 会失去训练语料和校正类型能力 |
| `migrate_v86()` | 无 | 有 | V8.6.4 侧有额外 DB 迁移逻辑 |

执行判定：

1. 若团队确认部署目标是 **V8.5 离线包**，继续执行本文。
2. 若团队实际目标是 **V8.6.4 或后续版本**，本文中 `ApplyCorrection` 的 query 参数方案不应直接执行，应重新按 V8.6.4 的 confirm/reject/revert/新 correction 契约制定计划。

## 2. 核心结论

当前 ai-hms 代码仍有多个与 V8.5 真服务不一致的点，不能只替换模型包或改版本号。

必须先做**契约修正**：

1. ACTRS V8.5 所有 API 都挂 `/api` 前缀，当前 Go client 未加 `/api`。
2. V8.5 `qc_pass` 是 `int(0/1)`，当前 Go 类型仍是 `bool`。
3. V8.5 医生校正接口是 query 参数 `correction`，当前 Go client 发送 JSON body `doctor_correction`。
4. V8.5 患者已存在时 `POST /api/patients` 返回 `409`，当前 `ensurePatientMapping` 未做 search fallback；这是映射丢失/环境切换场景的健壮性问题，优先级低于前三个契约阻断。
5. 当前 `.env.example` 写的是 `ACTRS_BASE_URL=http://localhost:8000/api`，与 V8.5 “BaseURL 不带 `/api`，client 内部补”约定冲突。
6. 当前 `upsertPatientACTR` 更新旧记录时未同步 `actr1/actr2/heart_width/lung_width/tilt_angle` 等 V8.5 字段。
7. 当前 ai-hms 入口只放行 `.dcm` 且上传上限 20MB，无法完整承接 V8.5 的 DICOM 能力（`.dicom/.ima`、大 DICOM）。
8. 当前 `GET /actr/status` 只返回 enabled/configured，不满足部署说明中的 reachable 自检要求。

已有 `docs/模型/ACTRS/ai-hms-ACTRS契约对拍+部署-交付-20260629(1)/纯改动-go.diff` 可作为参考，但**不能直接视为完整方案**：它覆盖 `/api` 前缀与 `qc_pass=*int`，但 `ApplyCorrection` 仍保留 JSON body，没有改成 V8.5 需要的 query 参数。因此本文取代该旧 diff 的执行计划，不能二者混用后认为完成。

## 3. 已核实的 V8.5 契约

### 3.1 路由前缀

V8.5 `main.py` 中所有 router 都挂载在 `/api`：

```py
app.include_router(patients.router, prefix="/api")
app.include_router(xrays.router, prefix="/api")
app.include_router(auth_router, prefix="/api")
```

因此 ai-hms `ACTRS_BASE_URL` 应只配置 `scheme://host:port`，client 内部拼 `/api/...`。

### 3.2 认证

`POST /api/auth/login`

请求：

```json
{"username":"...","password":"..."}
```

响应：

```json
{"access_token":"...","token_type":"bearer","role":"doctor","full_name":"...","user_id":1}
```

ai-hms 当前只需要 `access_token`，可保持兼容。

### 3.3 患者建档

`POST /api/patients` 创建患者；如果透析号已存在，返回 `409`：

```py
if existing:
    raise HTTPException(status_code=409, detail=f"透析号 {payload.dialysis_id} 已存在")
```

V8.5 也提供 `GET /api/patients?q=<dialysis_id>` 搜索。

### 3.4 胸片分析

`POST /api/patients/{patient_id}/xrays`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `file` | multipart file | JPG/PNG/DICOM |
| `analysis_date` | form string，可选 | 分析日期 |
| `notes` | form string，可选 | 备注 |

响应 `XrayOut` 包含：

| 字段 | 当前 ai-hms 是否接收 | 说明 |
|---|---:|---|
| `ctr` | 是 | CTR |
| `actr` | 是 | V8.5 主显示值，等于 actr2 |
| `actr1` / `actr2` / `actr_norm` | 是 | 类型已有 |
| `heart_width` / `lung_width` / `tilt_angle` | 是 | 类型已有 |
| `mask_path` | 是 | 类型已有 |
| `qc_pass` | **类型错误** | V8.5 是 `Optional[int]` |
| `qc_pa_ap` / `qc_warnings` | 是 | 类型已有 |
| `qc_rotation` / `qc_inspiration` | 否 | V8.5 新增，当前 ai-hms 不存 |
| `inference_time_ms` / `dicom_meta` | 否 | V8.5 新增，当前 ai-hms 不存 |
| `heart_area` / `lung_area` / `t1_dist` / `t2_dist` | 否 | V8.5 返回，当前 ai-hms 不存 |

### 3.5 医生校正

V8.5 `routers/xrays.py`：

```py
@router.patch("/xrays/{xray_id}/correction")
def apply_doctor_correction(xray_id: int, correction: float, ...)
```

即实际调用应是：

```http
PATCH /api/xrays/{id}/correction?correction=0.55
Authorization: Bearer <token>
```

当前 ai-hms 发送 JSON body：`{"doctor_correction": 0.55}`，与 V8.5 不一致。

## 4. 改造范围

### 4.1 本轮改造清单（按优先级）

| 优先级 | 文件 | 改造内容 |
|---|---|---|
| P0 | `internal/integrations/actrs/client.go` | 所有路径补 `/api`；`ACTRS_BASE_URL` 不带 `/api`；校正接口改 query 参数 |
| P0 | `internal/integrations/actrs/types.go` | `XrayOut.QCPass bool` 改为 `*int`；补充可选 V8.5 字段（至少用于解析，不一定入库） |
| P0 | `internal/integrations/actrs/mapper.go` | `QCPass *int` 转本地 int；修正更新字段映射 |
| P1 | `internal/api/v1/actr_handler.go` | DICOM 上传落地：上调或区分上传体积阈值；白名单补 `.dicom/.ima` |
| P1 | `internal/services/actr_service.go` | `upsertPatientACTR` 更新路径补 V8.5 字段；`Status` 增加 reachable |
| P1 | `.env.example` | `ACTRS_BASE_URL=http://localhost:8000`，注释明确不带 `/api`；timeout 注释匹配 DICOM/CPU 推理 |
| P1 | `ai-hms-frontend/src/components/actr/ActrPanel.tsx` | 上传 accept 同步补 `.dicom,.ima` |
| P1 | `internal/integrations/actrs/client_test.go` | 假服务改为 `/api` + `qc_pass: 0/1` + correction query |
| P1 | `internal/services/actr_service_test.go` | 假服务改为 V8.5 形状，覆盖校正 query、qc int、upsert 更新、reachable |
| P2 | `internal/services/actr_service.go` | `ensurePatientMapping` 增加搜索/409 fallback，必须精确匹配 `dialysis_id` |
| P2 | `internal/services/actr_service_test.go` | 覆盖搜索精确匹配和 409 fallback |
| P2 | `docs/模型/ACTRS/ACTRS接AI-HMS_落地总说明_20260629.md` | 更新接线说明（如仍作为总说明使用） |

### 4.2 可选：V8.5 全量字段镜像

当前 `patient_actr` 已包含 `actr1/actr2/actr_norm/heart_width/lung_width/tilt_angle/qc_pass/qc_pa_ap/qc_warnings/model_version/mask_path`。

但 V8.5 还提供以下字段，当前本地表没有：

| V8.5 字段 | 是否建议本轮入库 | 说明 |
|---|---:|---|
| `heart_area` / `lung_area` | 可选 | 面积值，科研/质控可用 |
| `t1_dist` / `t2_dist` | 可选 | 传统 CTR 拆分值 |
| `qc_rotation` / `qc_inspiration` | 建议后续 | 前端可展示更详细 QC |
| `inference_time_ms` | 可选 | 性能审计 |
| `dicom_meta` | 后续单独评估 | JSON 元数据，可能涉及敏感信息 |

本轮建议：**先不扩表**，保证 V8.5 可连通、可分析、可写入现有字段。扩表另起 DBA 审核脚本，因为数据库规则要求 app 运行时不得 DDL。

## 5. 详细实施步骤

### 5.1 修改 ACTRS client 路径与 BaseURL 规范

目标：`ACTRS_BASE_URL` 只填 `http://host:port`。

改造点：

```go
login:          POST c.baseURL + "/api/auth/login"
UpsertPatient:  POST "/api/patients"
ListXrays:      GET  "/api/patients/{id}/xrays"
AnalyzeXray:    POST "/api/patients/{id}/xrays"
ApplyCorrection: PATCH "/api/xrays/{id}/correction?correction=<value>"
```

同时在 `NewClient` 中做防呆：如果用户误把 `ACTRS_BASE_URL` 配成 `.../api`，可 trim 掉结尾 `/api` 并记录注释/测试，避免双 `/api/api`。

### 5.2 修正 `qc_pass` 类型

`types.go`：

```go
QCPass *int `json:"qc_pass"`
```

`mapper.go`：

```go
qc := 0
if x.QCPass != nil {
    qc = *x.QCPass
}
```

单测覆盖：`qc_pass=1`、`qc_pass=0`、`qc_pass=null/缺失`。

### 5.3 修正医生校正接口

当前：

```go
PATCH /xrays/{id}/correction
body: {"doctor_correction": 0.55}
```

V8.5 应改为：

```go
PATCH /api/xrays/{id}/correction?correction=0.55
```

`CorrectionRequest` 可以保留用于业务层，但 client 内部不再 JSON 编码为 body。`notes` 在 V8.5 correction 接口当前不接收，ai-hms 本地仍可保存医生备注到 `patient_actr.notes` 或暂不使用，需在计划中明确。

### 5.4 患者映射 409 fallback

当前 `ensurePatientMapping` 首次分析直接 `POST /patients`。若 ACTRS 中已有人但 ai-hms 本地 `external_patient_mappings` 丢失，会返回 `409`，分析失败。

优先级：P2。该项不是 V8.5 核心契约阻断，而是映射丢失、环境切换、人工清理映射表后的健壮性增强。若工期收紧，前三个契约阻断（`/api`、`qc_pass`、correction query）优先。

改造建议：

1. 先 `GET /api/patients?q=<dialysisNo>` 搜索。
2. 必须遍历搜索结果，只接受 `dialysis_id == dialysisNo` 的精确等值匹配；不能取第一条结果，因为 V8.5 搜索是 `name.contains(q) OR dialysis_id.contains(q)`，`D-001` 可能匹配到 `D-0010/D-0011`。
3. 若未找到，再 `POST /api/patients` 创建。
4. 若 `POST` 返回 `409`，再执行一次搜索 fallback。

需要新增 client 方法：

```go
SearchPatients(ctx, q string) ([]PatientOut, error)
```

`PatientOut` 保持 `ID/DialysisID/Name` 即可。

### 5.5 同步 V8.5 字段到本地更新路径

`MapXrayToPatientACTR` 新建路径已映射 `ACTR1/ACTR2/ACTRNorm/HeartWidth/LungWidth/TiltAngle/MaskPath`。

但 `ActrService.upsertPatientACTR` 更新已有记录时当前只更新：

```go
ctr, actr, actr_norm, qc_pass, qc_pa_ap, qc_warnings, model_version, analysis_date, image_path, overlay_path, mask_path
```

应补：

```go
actr1, actr2, heart_width, lung_width, tilt_angle
```

否则同一 ACTRS xray 二次同步时 V8.5 增强字段不会刷新。

`doctor_correction` / `notes` 需要避免反向覆盖：

1. `doctor_correction` 只在 ACTRS 新值非 nil 时更新，不用 nil 覆盖本地已有医生纠正值。
2. 若本地已有非空 `doctor_correction`，re-analyze 返回 nil 或旧值时不得清空本地值。
3. `notes` 只在新值非空或业务明确要求同步时更新，避免 ACTRS 侧空备注覆盖 ai-hms 本地医生备注。

### 5.6 DICOM 上传入口落地

V8.5 服务端支持 JPG/PNG/DICOM，DICOM 识别包括扩展名 `.dcm/.dicom/.ima`，并能处理部分无扩展但带 DICOM magic bytes 的文件。ai-hms 当前入口会先于 ACTRS 拦截请求，因此必须同步调整。

后端 `internal/api/v1/actr_handler.go`：

1. 当前 `maxUploadSize = 20 << 20` 对真实 PACS DICOM 偏小，常见 DICOM 可能 50-100MB+。本轮建议至少将 ACTR 上传阈值提升到 100MB，或按扩展名区分：普通图片 20MB，DICOM 100MB。
2. 白名单从 `.jpg/.jpeg/.png/.dcm` 扩为 `.jpg/.jpeg/.png/.dcm/.dicom/.ima`。
3. 无扩展 DICOM magic-byte 文件是否放行需另行评估；若要支持，应在读取少量文件头后判定，不能简单放开所有无扩展文件。

前端 `ActrPanel.tsx`：

```tsx
accept="image/*,.dcm,.dicom,.ima"
```

手工联调必须覆盖至少一个 `.dcm` 或 `.dicom/.ima` 文件，确认不在 ai-hms 入口被 413/400 拦截。

### 5.7 `.env.example`、timeout 与部署文档

改：

```ini
ACTRS_BASE_URL=http://localhost:8000
# 不要带 /api，Go client 内部自动补 /api
```

补充：

```ini
ACTRS_USERNAME=aihms_svc
ACTRS_PASSWORD=<ACTRS doctor/admin role account password>
ACTRS_TIMEOUT_SEC=60   # 分析路径含 DICOM 转换和 CPU 推理时建议 60s 起
```

注意：当前默认 10 秒可能不足以完成 CPU 胸片分析（V8.5 文档提到 CPU 推理约 8-30 秒，大 DICOM 还包含转换与落盘）。但当前 client timeout 是共享的，login/list/correction 也会使用它；把全局 timeout 调太大，会拖慢错误路径。

建议实现顺序：

1. 最小改造：`.env.example` 提示 `ACTRS_TIMEOUT_SEC=60`，生产按机器性能调整。
2. 更优改造：client 支持分析路径单独 timeout，例如 `ACTRS_ANALYZE_TIMEOUT_SEC=60-120`，login/list/correction 保持 10-15 秒。

### 5.8 `GET /actr/status` 增加 reachable

部署文档要求通过 `GET /actr/status` 检查 `enabled/reachable`。当前 `ActrService.Status()` 只返回 enabled/configured，无法完成连通性自检。本轮应把 reachable 列为必做，而不是可选。

建议返回：

```json
{
  "enabled": true,
  "configured": true,
  "reachable": true
}
```

实现方式：

1. 未启用或未配置时，`reachable=false`，不请求 ACTRS。
2. 已启用且已配置时，用短 timeout 调用轻量接口验证连通性，优先复用 login 或 health/version 端点；若 V8.5 无稳定 health 端点，可执行一次 login 验证。
3. reachable 检测不能使用分析长 timeout，避免状态接口卡住。

### 5.9 前端影响评估

当前 `ActrPanel.tsx` 已按 `qcPass: number` 展示：

```ts
qcPass === 1 ? QC 合格 : QC 不合格
```

因此后端 `PatientACTR.QCPass int` 与前端一致，不需要大改。

建议补充显示（可选）：

| 字段 | 前端增强 |
|---|---|
| `qcPaAp` | 显示 PA/AP/unknown 标签 |
| `qcWarnings` | 若为 JSON 字符串，解析成列表展示 |
| `modelVersion` | 小字显示模型版本 |
| `maskPath` | 暂不展示 |

本轮可不做前端增强，只保证 existing UI 不报错。

## 6. 测试计划

### 6.1 Go client 单测

文件：`ai-hms-backend/internal/integrations/actrs/client_test.go`

假服务必须模拟 V8.5：

| 路由 | 预期 |
|---|---|
| `/api/auth/login` | 返回 `access_token` |
| `/api/patients` | 需要 Bearer token；返回 patient |
| `/api/patients?q=<dialysisNo>` | 返回包含相近透析号的列表，用于验证精确匹配过滤 |
| `/api/patients/77/xrays` | `qc_pass` 返回 int |
| `/api/xrays/9/correction` | 校验 query `correction=0.5`，不得依赖 JSON body |

### 6.2 Mapper 单测

文件：`ai-hms-backend/internal/integrations/actrs/mapper_test.go`

覆盖：

1. `qc_pass=1` → 本地 `QCPass=1`。
2. `qc_pass=0` → 本地 `QCPass=0`。
3. `qc_pass` 缺失/nil → 本地 `QCPass=0`。
4. `actr1/actr2/actr_norm/mask_path/model_version` 被正确映射。

### 6.3 Service 单测

文件：`ai-hms-backend/internal/services/actr_service_test.go`

覆盖：

1. `ensurePatientMapping` 本地无映射、ACTRS 搜索已有透析号 → 遍历结果并按 `dialysis_id` 精确等值匹配，不 POST，直接写映射。
2. `POST /api/patients` 返回 409 后 fallback 搜索成功 → 写映射。
3. `Analyze` 返回 `qc_pass=1` 可持久化。
4. 二次 `Analyze` 同一 xray ID 更新 `actr1/actr2/heart_width/lung_width/tilt_angle`。
5. 二次 `Analyze` 不用 nil/空值覆盖本地非空 `doctor_correction/notes`。
6. `Correct` 调 ACTRS 时使用 query 参数 `correction`。
7. `Status` 在启用且配置完整时返回 `reachable`，并使用短超时路径。

### 6.4 Handler / 前端入口测试

覆盖：

1. `actr_handler.go` 放行 `.dcm/.dicom/.ima`。
2. DICOM 上传阈值不再被 20MB 硬拦截，或按扩展名区分普通图片和 DICOM 阈值。
3. `ActrPanel.tsx` 的 `accept` 包含 `image/*,.dcm,.dicom,.ima`。

### 6.5 手工联调

1. 启动 V8.5 离线包：`docker-compose up -d`。
2. 配置 ai-hms `.env`：`ACTRS_BASE_URL=http://host:port`，不要 `/api`。
3. `GET /api/v1/actr/status` 检查 `enabled/configured/reachable`。
4. 医生墙打开 ACTR 面板，上传 jpg/png。
5. 上传 `.dcm` 和至少一种 `.dicom/.ima` DICOM 文件，确认 ACTRS 侧可返回结果，ai-hms 不报 413/400。
6. 医生校正值，确认 ACTRS 无 422/404，ai-hms 本地 `doctor_correction/corrected_by` 更新。

## 7. 验证命令

后端：

```powershell
cd ai-hms-backend
go test ./internal/integrations/actrs ./internal/services -count=1 -timeout 120s
```

前端：

```powershell
cd ai-hms-frontend
npm run build
```

如果执行完整门禁：

```powershell
cd ai-hms-frontend
npm run lint
```

注意：若 lint 失败，需要区分是否为既有历史问题。

## 8. 不纳入本轮

1. 不提交 `pspnet_chestxray_best_model_4.pth`、`actrs.db`、uploads、backups、离线包运行产物。
2. 不把 V8.5 服务端源码合入 ai-hms 主应用；ACTRS 作为独立服务部署。
3. 不在 app runtime 执行 DDL。
4. 不扩展 `patient_actr` 表存储 `dicom_meta/qc_rotation/qc_inspiration/inference_time_ms`，如需要另起 DBA 审核脚本。
5. 不接 PACS 自动取片。
6. 不接 QC 评分页 CTR 槽位与 RNa 干体重外环二阶段逻辑，除非另立需求。

## 9. 文件清单建议

### 必改文件

```text
ai-hms-backend/internal/integrations/actrs/client.go
ai-hms-backend/internal/integrations/actrs/types.go
ai-hms-backend/internal/integrations/actrs/mapper.go
ai-hms-backend/internal/api/v1/actr_handler.go
ai-hms-backend/internal/integrations/actrs/client_test.go
ai-hms-backend/internal/integrations/actrs/mapper_test.go
ai-hms-backend/internal/services/actr_service.go
ai-hms-backend/internal/services/actr_service_test.go
ai-hms-backend/.env.example
ai-hms-frontend/src/components/actr/ActrPanel.tsx
docs/模型/ACTRS/ACTRS_V8.5_融合调整执行计划.md
```

### 视情况改

```text
ai-hms-frontend/src/services/actrApi.ts
docs/模型/ACTRS/ACTRS接AI-HMS_落地总说明_20260629.md
docs/sql/deploy_new_tables.sql   # 仅当决定扩 patient_actr 字段，且走部署阶段 SQL
```

## 10. 验收标准

1. 团队已书面确认本轮目标是 V8.5 离线包；若目标变为 V8.6.4，本计划需重写 correction 相关部分。
2. Go client 单测证明所有 ACTRS 路径均带 `/api`。
3. `qc_pass` int 解析不报错，`0/1/nil` 映射正确。
4. `ApplyCorrection` 使用 query 参数 `correction`，不再依赖 JSON body。
5. `ACTRS_BASE_URL` 配成 `http://host:port` 时能登录、建档/搜索、上传、校正。
6. 重复同步同一个 xray ID 时，V8.5 已有字段能更新到 `patient_actr`，且不以 nil/空值覆盖本地非空 `doctor_correction/notes`。
7. `GET /api/v1/actr/status` 返回 `enabled/configured/reachable`，部署自检可判断 ACTRS 连通性。
8. ai-hms 上传入口支持 `.dcm/.dicom/.ima`，DICOM 文件不会被 20MB 固定上限误拦截。
9. P2 若纳入本轮：本地映射表丢失但 ACTRS 已存在患者时，能通过 `dialysis_id` 精确等值搜索补回映射，不误关联相似透析号。
10. `go test ./internal/integrations/actrs ./internal/services` 通过。
11. `go build ./cmd/server` 通过。
12. `npm run build` 通过。

## 11. 提交信息建议

```text
fix(actrs): 对齐 V8.5 离线包 API 契约与 qc_pass 类型
```

如同时补患者 409 fallback、DICOM 入口和 `.env.example`：

```text
fix(actrs): 适配 V8.5 契约、DICOM 上传与患者映射回补
```
