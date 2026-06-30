"""IDH 阈值重标定 · 分析(本地跑团队回传的 recal_scored.csv)。

输入 CSV 列:treatment_id, operate_time, prob, label  (label∈{0,1})
产出:阈值扫描表 threshold_sweep.csv、ROC/PRC 数据、推荐工作阈值(偏敏感度)、
      风险三级切点、概率校准结论 → recalibration_report.md。

用法:python recal_analysis.py recal_scored.csv [--target-sens 0.85]
依赖:numpy pandas scikit-learn(子项目B 的 .venv312 已具备)。
"""
from __future__ import annotations
import argparse
import os
import sys
import numpy as np
import pandas as pd
from sklearn.metrics import roc_auc_score, average_precision_score, roc_curve

try:  # Windows GBK 控制台兜底:报告用 UTF-8 输出,不被 ‑/中文呛到。
    sys.stdout.reconfigure(encoding="utf-8")
except Exception:
    pass


def confusion_at(y, p, thr):
    pred = (p >= thr).astype(int)
    tp = int(((pred == 1) & (y == 1)).sum())
    fp = int(((pred == 1) & (y == 0)).sum())
    tn = int(((pred == 0) & (y == 0)).sum())
    fn = int(((pred == 0) & (y == 1)).sum())
    sens = tp / (tp + fn) if (tp + fn) else 0.0
    spec = tn / (tn + fp) if (tn + fp) else 0.0
    ppv = tp / (tp + fp) if (tp + fp) else 0.0
    npv = tn / (tn + fn) if (tn + fn) else 0.0
    f1 = 2 * ppv * sens / (ppv + sens) if (ppv + sens) else 0.0
    alert = (tp + fp) / len(y) if len(y) else 0.0
    return dict(threshold=round(float(thr), 3), tp=tp, fp=fp, tn=tn, fn=fn,
                sensitivity=round(sens, 4), specificity=round(spec, 4),
                ppv=round(ppv, 4), npv=round(npv, 4), f1=round(f1, 4),
                alert_rate=round(alert, 4))


def sweep(y, p, grid):
    return pd.DataFrame([confusion_at(y, p, t) for t in grid])


def pick_sensitivity(df, target):
    """满足 sensitivity>=target 的最小阈值行(报警最少/最不打扰的偏敏感工作点)。"""
    ok = df[df["sensitivity"] >= target]
    if len(ok) == 0:
        return df.iloc[df["sensitivity"].idxmax()]
    return ok.sort_values("threshold", ascending=False).iloc[0]


def youden(df):
    j = df["sensitivity"] + df["specificity"] - 1
    return df.iloc[int(j.idxmax())]


def calibration_bins(y, p, nbins=10):
    edges = np.linspace(0, 1, nbins + 1)
    rows = []
    for i in range(nbins):
        m = (p >= edges[i]) & (p < edges[i + 1] if i < nbins - 1 else p <= edges[i + 1])
        if m.sum() == 0:
            continue
        rows.append(dict(bin=f"[{edges[i]:.1f},{edges[i+1]:.1f})",
                         n=int(m.sum()),
                         pred_mean=round(float(p[m].mean()), 4),
                         obs_rate=round(float(y[m].mean()), 4)))
    return pd.DataFrame(rows)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("csv")
    ap.add_argument("--target-sens", type=float, default=0.85)
    ap.add_argument("--out", default=".")
    args = ap.parse_args()

    if not os.path.isfile(args.csv):
        sys.exit(f"CSV 文件不存在: {args.csv}")

    df = pd.read_csv(args.csv)
    if len(df) == 0:
        sys.exit("CSV 为空，无法分析。请检查 dry-run 提取是否成功。")

    required = {"treatment_id", "operate_time", "prob", "label"}
    missing = required - set(df.columns)
    if missing:
        sys.exit(f"CSV 缺少必要列: {missing}。期望列: {sorted(required)}")

    try:
        p = df["prob"].astype(float).to_numpy()
    except (ValueError, TypeError) as e:
        sys.exit(f"prob 列无法转为数值: {e}")

    if (p < 0).any() or (p > 1).any():
        sys.exit(f"prob 必须在 [0,1] 区间，当前范围 [{p.min():.4f}, {p.max():.4f}]")

    try:
        y = df["label"].astype(int).to_numpy()
    except (ValueError, TypeError) as e:
        sys.exit(f"label 列无法转为整数: {e}")

    invalid = set(y) - {0, 1}
    if invalid:
        sys.exit(f"label 只能是 0/1，发现意外值: {sorted(invalid)}")

    n, pos = len(y), int(y.sum())
    if pos == 0:
        sys.exit("正例(label=1)数量为 0，无法计算 AUC/切点。请检查数据质量。")
    if pos == n:
        sys.exit("负例(label=0)数量为 0，无法计算 AUC/切点。请检查数据质量。")

    auc = roc_auc_score(y, p) if pos < n else float("nan")
    prc = average_precision_score(y, p) if pos < n else float("nan")

    os.makedirs(args.out, exist_ok=True)

    grid = np.round(np.arange(0.01, 1.00, 0.01), 2)
    sw = sweep(y, p, grid)
    sw.to_csv(os.path.join(args.out, "threshold_sweep.csv"), index=False)

    sens_row = pick_sensitivity(sw, args.target_sens)
    j_row = youden(sw)
    cal = calibration_bins(y, p)

    # 风险三级切点:high=偏敏感工作阈值;medium=youden 与 high 的较小者再下探;low 以下。
    high = float(sens_row["threshold"])
    med = round(min(float(j_row["threshold"]), high) * 0.6, 3)

    lines = []
    lines.append("# IDH 阈值重标定 · 报告\n")
    lines.append(f"- 样本 n={n}, 正例={pos}({pos/n*100:.1f}%)")
    lines.append(f"- **本中心 AUC={auc:.4f}**, PRC-AUC={prc:.4f}")
    lines.append(f"- 目标敏感度 ≥ {args.target_sens}\n")
    lines.append("## 偏敏感度工作点(推荐)")
    lines.append(f"- threshold=**{sens_row['threshold']}**  sens={sens_row['sensitivity']} "
                 f"spec={sens_row['specificity']} ppv={sens_row['ppv']} 报警率={sens_row['alert_rate']}")
    lines.append(f"- 混淆 TP={sens_row['tp']} FP={sens_row['fp']} TN={sens_row['tn']} FN={sens_row['fn']}\n")
    lines.append("## Youden's J 平衡点(参照)")
    lines.append(f"- threshold={j_row['threshold']} sens={j_row['sensitivity']} spec={j_row['specificity']}\n")
    lines.append("## 风险三级切点建议")
    lines.append(f"- high ≥ {high}, medium ≥ {med}, 其余 low")
    lines.append(f"- 接入 Go:`IDH_LEVEL_HIGH={high}` `IDH_LEVEL_MEDIUM={med}`\n")
    lines.append("## 概率校准(可靠性)")
    lines.append(cal.to_string(index=False))
    drift = float((cal["pred_mean"] - cal["obs_rate"]).abs().mean()) if len(cal) else 0.0
    lines.append(f"\n- 平均 |pred-obs| = {drift:.4f} "
                 f"({'偏差大,建议上 isotonic/Platt 校准' if drift > 0.05 else '校准尚可,可不校准'})")

    report = "\n".join(lines)
    with open(os.path.join(args.out, "recalibration_report.md"), "w", encoding="utf-8") as f:
        f.write(report)
    print(report)


if __name__ == "__main__":
    main()
