# IDH 预警微服务（实时监控 AI③ 真模型）

把 dixueya 研究的 **CatBoost+SMOTE** 透中低血压（IDH）预测模型，部署成与 ACTRS 平级的
FastAPI 微服务，供 ai-hms 后端 `internal/integrations/idh` 的 `HTTPScorer` 调用，
点亮实时监控床卡的"⚠ IDH 风险"。

| 项 | 内容 |
|---|---|
| 状态 | **已产品化 + 对拍通过**（子项目B，2026-06-28） |
| ai-hms 侧 | **Go 侧集成已由团队整合上推 `origin/master`（commit `c8a6089`）**——含 IDH A+B 契约/缓存/接线（团队复核整改：IDH 配置统一进 `config.IDHConfig`、缓存 key=`{tenantID,treatmentID}`）。本微服务按本包**独立部署**，与该 Go 侧对接。 |
| 真模型表现 | AUC≈0.91 · 特异度 0.997 · 敏感度 0.67 |
| 对拍 | `study/parity_test.py`：20 行训练样本，在线特征 → 概率与训练侧逐行一致 `dprob=0`（唯一已知 train/serve skew = `TreatmentId_std`，见下） |
| 接口 | `POST /idh/score` · `GET /health` |

---

## 一、环境红线（重要）

- **必须 Python 3.11 / 3.12**。**禁用 Python 3.14 / 3.15**：这些版本 sklearn 仅有 1.7.x wheel，
  而 1.7 删除了 `_RemainderColsList`，加载训练期保存的 `ColumnTransformer` pickle 会报
  `AttributeError: ... _RemainderColsList`。
- 锁定组合（已验证）：**Python 3.12 + scikit-learn==1.6.1 + catboost==1.2.10**。

## 二、文件

| 文件 | 说明 |
|---|---|
| `app.py` | FastAPI：加载模型+preprocessor，缓存 `feature_names_in_`，`POST /idh/score` |
| `features.py` | 在线特征工程：当前 BP(SBP/DBP/MAP) + 30 时点设备列拉平 + 均值/标准差 + 基本信息，`reindex` 自对齐到 preprocessor 的 620 列 |
| `requirements.txt` | 依赖（版本已钉死） |
| `models/` | `CatBoost_model.pkl`、`scaler.pkl`(ColumnTransformer) |
| `tests/` | `test_features.py`、`test_app.py`（pytest） |
| `study/parity_test.py` | 对拍脚本（验收硬门禁） |

## 三、准备模型文件

从 `桌面\HMS开发\dixueya.rar` 解出到 `models/`：
```bash
"C:/Program Files/7-Zip/7z.exe" e <path>/dixueya.rar -omodels "dixueya/CatBoost_model.pkl" "dixueya/scaler.pkl" -y
```
可用 `IDH_MODEL_PATH` / `IDH_PREPROCESSOR_PATH` 覆盖路径（默认 `models/CatBoost_model.pkl`、`models/scaler.pkl`）。

## 四、建环境 + 运行

```bash
# Windows，使用 Python 3.12
py -V:Astral/CPython3.12.12 -m venv .venv312      # 或任意 py3.11/3.12
.venv312\Scripts\python -m pip install -r requirements.txt
.venv312\Scripts\uvicorn app:app --host 0.0.0.0 --port 8910
curl http://127.0.0.1:8910/health   # {"ok":true,"modelLoaded":true,"preprocessorLoaded":true}
```

## 五、测试 + 对拍（验收硬门禁）

```bash
.venv312\Scripts\python -m pytest tests/ -q          # features + app 全过
.venv312\Scripts\python study\parity_test.py         # 须打印 PARITY PASS
```
对拍最近结果（2026-06-28，N=20）：每行 `dprob=0.00e+00`，`PARITY PASS (col_ok=True, prob_ok=True, skew_cols=['TreatmentId_std'])`。
> `TreatmentId_std` 是训练拉平时窗口跨治疗边界产生的泄漏列，**在线按治疗取窗恒 0**（更干净），
> 是唯一的 train/serve skew；对拍以"覆盖该列后概率逐行一致"证明特征流水线本身完全正确。

## 六、与 ai-hms 后端对接（子项目A 已接线）

- 后端设环境变量：`IDH_ENABLED=true`、`IDH_BASE_URL=http://<host>:8910`、可选 `IDH_TIMEOUT_SEC`。
- 后端 `internal/integrations/idh.HTTPScorer` POST 整个 `RiskInput`（契约已与本服务 `ScoreRequest` 逐字对齐，含 20 设备列 PascalCase + `basic` 的 `pre-Weight`/`pre-SBP`/`pre-DBP`/`UFQuantity_y` 及**当前 `SBP`/`DBP`**）。
- 当前 BP 由后端 during-signs 发来；缺失则服务端 `reindex` 兜底 0 + 降级。
- 全链路失败均降级为 `available:false` → 卡面"IDH 待数据"，绝不阻断实时监控（后端 stale-while-revalidate 缓存异步刷新）。
- 概率→分级（high/medium/low）切点在后端 `idh.LevelFromProbability`（0.5/0.2）。

## 七、本期边界 / 后续

- 仅 **IDH 低血压**一个目标；失衡/痉挛/心律失常留扩展。
- **阈值/SMOTE 重标定**需本中心带标签数据，本期未做（留后续）；当前用训练期默认切点。
