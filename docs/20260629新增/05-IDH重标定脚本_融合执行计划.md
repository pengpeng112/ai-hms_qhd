# IDH 重标定脚本融合执行计划

## 1. 通俗功能说明

当前 IDH 模型已经能给出低血压概率，但它的概率切点来自原训练数据，不一定最适合本中心。这个项目用本中心历史数据重新评估模型概率和真实低血压标签，找出更适合本中心的“高风险/中风险”阈值。

简单说：不是重新训练模型，而是用本院数据重新决定“概率多少算高危”，让提醒更符合本院实际。

## 2. 当前系统状态

- 数据可用性已经确认充足：合格治疗 23,815，可标注点 320,645，IDH 率 9.3%，窗口充足度良好。
- Go 侧 `LevelFromProbability` 目前仍是固定切点：0.5 high、0.2 medium。
- 交付脚本包括：
  - `study/recal_extract_score.py`：生产库只读提取并打分，输出小 CSV。
  - `study/recal_analysis.py`：分析 CSV，输出阈值扫描和报告。
- 本项目一期只调阈值/校准，不重训模型权重。
- 用户已确认：目标敏感度固定使用 `0.85`，不改为更激进的 `0.90`。
- 切点当前不存数据库；按当前代码形态，推荐代码默认值先保持现状 `high=0.5`、`medium=0.2`，重标定报告确认后的本中心切点优先通过环境变量 `IDH_LEVEL_HIGH` / `IDH_LEVEL_MEDIUM` 配置。

## 3. 融合目标

- 把重标定脚本纳入 IDH 微服务部署资料或单独研究目录。
- 生产库侧只读导出 `recal_scored.csv`。
- 本地分析得到推荐 high/medium 切点。
- Go 侧将 IDH 风险切点做成可配置，支持 `IDH_LEVEL_HIGH`、`IDH_LEVEL_MEDIUM`。

## 4. 不做事项

- 不在 AI-HMS 后端请求路径跑重标定。
- 不在生产库执行 DDL。
- 不导出患者姓名、身份证、电话等敏感信息。
- 不直接把重标定结果写死，除非已经拿到完整 `recalibration_report.md` 并经用户确认。

## 5. 执行阶段

### 5.1 阶段一：准备脚本运行目录

建议把脚本放在 IDH 微服务目录的 `study/` 下，或单独 `tools/idh-recalibration/`。如果放入主仓，应只放脚本和 README，不放模型 pickle 和生产 CSV。

需要文件：

- `study/recal_extract_score.py`
- `study/recal_analysis.py`
- `study/README_IDH重标定.md`

### 5.2 阶段二：生产库侧只读提取

环境要求：

- 使用 IDH 微服务同一 Python 3.12 环境。
- 安装 `psycopg2-binary`。
- 使用只读数据库账号。

先小批 dry-run：

```powershell
set PG_DSN=host=<ip> port=5432 dbname=<db> user=<readonly> password=<***>
.venv312\Scripts\python study\recal_extract_score.py --tenant 3 --limit-treatments 20 --out dryrun.csv
```

检查：

- `dryrun.csv` 有数据。
- `prob` 在 0 到 1 之间。
- `label` 只有 0/1。
- 行内不包含患者敏感信息。

全量提取：

```powershell
.venv312\Scripts\python study\recal_extract_score.py --tenant 3 --out recal_scored.csv
```

输出：`recal_scored.csv`，列为 `treatment_id, operate_time, prob, label`。

### 5.3 阶段三：本地分析

```powershell
.venv312\Scripts\python study\recal_analysis.py recal_scored.csv --target-sens 0.85 --out recal_out
```

目标敏感度固定为 `0.85`，除非用户再次明确修改，否则不要改成 `0.90` 或其它值。

输出：

- `recalibration_report.md`
- `threshold_sweep.csv`

重点读取：

- AUC
- PRC-AUC
- 推荐 high 阈值
- 推荐 medium 阈值
- 敏感度、特异度、PPV、报警率
- 校准偏差是否需要 isotonic/Platt

### 5.4 阶段四：Go 侧切点配置化

改动文件建议：

- `ai-hms-backend/internal/integrations/idh/types.go`
- `ai-hms-backend/internal/integrations/idh/scorer.go`
- `ai-hms-backend/config/config.go`
- `ai-hms-backend/cmd/server/main.go`
- `ai-hms-backend/internal/integrations/idh/scorer_test.go`
- `ai-hms-backend/.env.example`

建议实现：

1. `idh.Config` 增加：
   - `LevelHigh float64`
   - `LevelMedium float64`
2. `HTTPScorer` 持有切点配置。
3. 将 `LevelFromProbability(p)` 保留默认行为，新增可配置版本，例如：
   - `LevelFromProbabilityWithCuts(p, high, medium float64)`
4. `HTTPScorer.Score` 用配置切点映射 level。
5. `config.IDHConfig` 增加：
   - `LevelHigh`
   - `LevelMedium`
6. 环境变量：
   - `IDH_LEVEL_HIGH`
   - `IDH_LEVEL_MEDIUM`
7. 默认值在拿到报告前保持当前值：high=0.5，medium=0.2。
8. 拿到报告并确认后，推荐优先只在部署环境中设置 `IDH_LEVEL_HIGH` / `IDH_LEVEL_MEDIUM`；不要急于把本中心切点写死进代码默认值，避免多院区/多环境共用代码时误用。

## 6. 验证命令

Python 脚本：

```powershell
cd <IDH微服务目录>
.venv312\Scripts\python study\recal_extract_score.py --tenant 3 --limit-treatments 20 --out dryrun.csv
.venv312\Scripts\python study\recal_analysis.py dryrun.csv --target-sens 0.85 --out dryrun_out
```

Go 侧：

```powershell
cd ai-hms-backend
gofmt -w internal/integrations/idh/*.go config/config.go cmd/server/main.go
go vet ./internal/integrations/idh ./internal/config ./cmd/server
go test ./internal/integrations/idh ./internal/config -count=1 -timeout 120s
go build -o "$env:TEMP\check.exe" ./cmd/server
```

## 7. 风险与边界

- 一期只是切点重标定，不改变模型权重；如果本中心 AUC 明显低，需要另做二期重训。
- 提取脚本会读取大量历史数据，必须用只读账号，并建议离峰运行。
- `recal_scored.csv` 虽不含姓名，但仍含治疗 ID 和时间，应按内部数据管理，不应提交 git。
- 如果概率校准偏差大，是否上线 calibrator 需要单独决策；不要在没有评估报告时直接改线上概率。

## 8. 尚需确认的问题

- 若未来只有单中心部署且重标定稳定，是否再把环境变量中的推荐切点固化为代码默认值；当前推荐先走环境变量。
- 重标定分析报告出来后，是否允许我再改 Go 侧默认切点？
