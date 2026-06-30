# IDH 阈值重标定 · 一期(本中心评估 + 偏敏感度调阈值)

数据已确认充足(合格治疗 23815 / 标注点 320645 / IDH 率 9.3% / 窗口≥10 占 92.8%),现在就能做。
**分工:团队跑提取(生产库),把小 CSV 回传;我跑分析、出阈值、改 Go。**

## 一致性原则(重要)
提取打分**复用与线上同一套** `features.build_feature_frame` + `models/`(CatBoost+scaler),且**只用线上可得的 9 列 DMLog**(其余训练列线上恒 0,这里同样为 0)。
→ 重标定出的阈值对**线上实际特征**有效(不是用"完整 620 特征"标定一个线上达不到的乐观阈值)。

## 团队步骤(生产库侧,只读)

1. 用子项目B 的环境:`实时监控_IDH微服务\.venv312`。
2. **仅提取脚本**需要补装数据库驱动（微服务运行不需要 `psycopg2-binary`）：
   `.venv312\Scripts\python -m pip install -r study\requirements-recal.txt`
3. 配只读连接(环境变量):
   `set PG_DSN=host=<ip> port=5432 dbname=<db> user=<只读用户> password=<***>`
4. **先小批 dry-run**(确认表名/字段/连通):
   `.venv312\Scripts\python study\recal_extract_score.py --tenant 3 --limit-treatments 20 --out dryrun.csv`
   - 看 dryrun.csv 有几百行、prob∈[0,1]、label∈{0,1} 即正常。
   - 若报字段/表名错(各院 schema 可能略不同),按报错改脚本顶部 SQL(基本信息那几张表:Treatment_BeforeSigns / Register_PatientInfomation / Plan_PatientPlan / Plan_PatientPrescription)。
5. **全量跑**(约 32 万行,可断点续——按 `(treatment_id, operate_time)` 逐行去重，中断后重跑不会丢行):
   `.venv312\Scripts\python study\recal_extract_score.py --tenant 3 --out recal_scored.csv`
6. 把 **`recal_scored.csv`** 回传(只含 treatment_id, operate_time, prob, label,无患者敏感信息,体量小)。

## 我方步骤(拿到 CSV 后)

`.venv312\Scripts\python study\recal_analysis.py recal_scored.csv --target-sens 0.85 --out ./recal_out`
→ 自动创建 `--out` 目录；产出 `recalibration_report.md`(本中心 AUC、**偏敏感度工作阈值**、风险三级切点、概率校准结论)+ `threshold_sweep.csv`。
→ 据此把 ai-hms `idh.LevelFromProbability` 切点改成推荐值(做成可配 `IDH_LEVEL_HIGH`/`IDH_LEVEL_MEDIUM`);校准偏差大时上 isotonic/Platt(可选)。

## 标签口径(与训练/盘点一致)
某血压测点 t 的"下一次测量" t+1 低血压 = `下次SBP<100 或 SBP跌>20 或 MAP跌>10`(MAP=(SBP+2·DBP)/3)。

## 边界
- 一期**只调阈值/校准、不重训权重**;若本中心 AUC 明显低于 0.91 再议二期重训。
- 自测脚本已在合成数据上验证分析链路(AUC/扫描/选点/校准)跑通。
