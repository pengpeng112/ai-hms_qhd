# 实时监控 IDH 微服务融合执行计划

## 1. 通俗功能说明

这个目录是一个独立部署的 Python FastAPI 服务。它加载 CatBoost 低血压预测模型，接收 AI-HMS 后端传来的某次治疗近 30 个设备时点、当前血压和患者基本信息，返回下一次测量发生透中低血压的概率。

简单说：AI-HMS 负责把实时监控数据整理好发过去，Python 微服务负责算“马上低血压的风险有多高”。

## 2. 当前系统状态

- Go 侧接口已经具备：`internal/integrations/idh.HTTPScorer` 会 POST `/idh/score`。
- Go 侧配置已经具备：`IDH_ENABLED`、`IDH_BASE_URL`、`IDH_TIMEOUT_SEC` 已进入 `config.IDHConfig`。
- Go 侧默认禁用：未设置 `IDH_ENABLED=true` 时，床卡仍显示 IDH 不可用，不影响实时监控。
- 当前仓库不应直接合入 `models/*.pkl`，模型文件应随 Python 微服务独立部署。

## 3. 融合目标

- 不把 Python 微服务和模型文件并入 AI-HMS 主服务运行进程。
- 只做“部署与对接”融合：明确部署步骤、健康检查、环境变量、验收门禁。
- 保持 Go 侧失败降级：微服务不可用时返回 `available:false`，不阻塞床卡。

## 4. 不做事项

- 不把 `models/CatBoost_model.pkl`、`models/scaler.pkl` 提交进主仓代码。
- 不在 AI-HMS 后端启动 Python 服务。
- 不在应用启动或请求路径执行 DDL。
- 不宣称真实模型已生产验证，除非对拍和联调通过。

## 5. 执行步骤

### 5.1 部署 Python 微服务

1. 在单独目录部署 `实时监控_IDH微服务/`，不要复制到 `ai-hms-backend/internal`。
2. 使用 Python 3.11 或 3.12，禁止 Python 3.14/3.15。
3. 安装依赖：
   ```powershell
   py -3.12 -m venv .venv312
   .venv312\Scripts\python -m pip install -r requirements.txt
   ```
4. 确认模型文件存在：`models/CatBoost_model.pkl`、`models/scaler.pkl`。
5. 启动服务：
   ```powershell
   .venv312\Scripts\uvicorn app:app --host 0.0.0.0 --port 8910
   ```
6. 健康检查：
   ```powershell
   curl http://127.0.0.1:8910/health
   ```
7. 期望：`ok=true`、`modelLoaded=true`、`preprocessorLoaded=true`。

### 5.2 微服务验收

1. 运行单元测试：`.venv312\Scripts\python -m pytest tests/ -q`。
2. 运行对拍：`.venv312\Scripts\python study\parity_test.py`。
3. 必须看到：`PARITY PASS`。
4. 对拍不通过时，不允许开启生产 `IDH_ENABLED=true`。

### 5.3 AI-HMS 后端配置

在 `ai-hms-backend/.env` 或部署环境中配置：

```env
IDH_ENABLED=true
IDH_BASE_URL=http://<idh-service-host>:8910
IDH_TIMEOUT_SEC=5
```

如果 `IDH_ENABLED=true` 但 `IDH_BASE_URL` 为空，后端应保持 StubScorer 并记录日志。

### 5.4 联调检查

1. 启动 AI-HMS 后端。
2. 打开实时监控床卡。
3. 观察后端日志是否出现：`[IDH] HTTPScorer enabled`。
4. 对有 DMLog 窗口的治疗，确认 `idHRisk.available` 可变为 `true`。
5. 停止 Python 微服务后，确认实时监控接口仍正常返回，只是 IDH 不可用。

## 6. 可交给其他 AI 的任务

- 新增 `docs/idh-microservice-deployment.md`，记录部署、对拍、回滚。
- 新增示例 `.env` 片段，禁止写真实密码。
- 可选新增 Dockerfile，但不要把模型文件打入 git。

## 7. 验证命令

Python 微服务：

```powershell
cd <实时监控_IDH微服务目录>
.venv312\Scripts\python -m pytest tests/ -q
.venv312\Scripts\python study\parity_test.py
```

AI-HMS 后端：

```powershell
cd ai-hms-backend
go test ./internal/integrations/idh ./internal/services -count=1 -timeout 120s
go build -o "$env:TEMP\check.exe" ./cmd/server
```

## 8. 风险与不确定项

- 模型 pickle 与 scikit-learn 版本强绑定，必须锁定 Python 3.11/3.12 和 sklearn 1.6.1。
- 当前 Go 侧只传生产库已确认的 9 个 DMLog 核心列，其余模型列为 null/0，这是已知边界。
- 若后续启用本中心重标定或概率校准，需要另走 `IDH重标定脚本` 项目计划。
- **依赖差异**：`requirements.txt` 只包含微服务运行所需依赖。重标定提取脚本（`recal_extract_score.py`）需要连生产库，必须额外 `pip install psycopg2-binary`。建议在 `study/` 目录下放一份单独的 `requirements-recal.txt`（含 `psycopg2-binary`），与微服务运行环境隔离，避免微服务部署人员遗漏。
