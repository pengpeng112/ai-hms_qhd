"""对拍验证(B 验收硬门禁)。

训练宽表前 N 行为基准:
  参考(A) = 行[expected] → preprocessor → model.predict_proba
  在线(B) = 反推 window/basic → build_feature_frame → preprocessor → model.predict_proba

断言:
  1. B 产出的 620 列向量与 A 仅在「已知不可在线复现的训练泄漏列」SKEW_COLS 上可不同,
     其余列逐列一致(<1e-6)。SKEW_COLS={'TreatmentId_std'}:训练拉平窗跨了治疗边界
     导致非 0;在线按治疗取窗恒 0(更干净)。
  2. 把 SKEW_COLS 用参考值覆盖回 B 后,predict_proba 与 A 逐行一致(<1e-6),
     证明特征流水线本身完全正确(差异仅源于该泄漏列)。

运行:.venv312/Scripts/python.exe study/parity_test.py
"""
import os, sys, pickle
BASE = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, BASE)
import numpy as np
import pandas as pd
from features import build_feature_frame, NUM_COLS, WINDOW

CSV = os.path.join(BASE, "study", "dixueya", "透析数据_拉平+基本信息.csv")
N = 20
SKEW_COLS = {"TreatmentId_std"}  # 训练泄漏列,在线不可复现(恒0),不计入逐列门禁
TOL = 1e-6

model = pickle.load(open(os.path.join(BASE, "models", "CatBoost_model.pkl"), "rb"))
pre = pickle.load(open(os.path.join(BASE, "models", "scaler.pkl"), "rb"))
expected = list(pre.feature_names_in_)

df = pd.read_csv(CSV, nrows=N)
ref = df.reindex(columns=expected)
ref_prob = model.predict_proba(pre.transform(ref))[:, 1]


def row_to_inputs(row):
    win = []
    for i in range(1, WINDOW + 1):
        pt = {"LogTime": 0.0}
        for c in NUM_COLS:
            col = f"{c}_{i}"
            pt[c] = float(row[col]) if col in row and pd.notna(row[col]) else np.nan
        win.append(pt)
    basic = {k: row.get(k) for k in
             ["Gender", "Age", "DialysisMethod", "DryWeight", "UFQuantity_y",
              "pre-Weight", "pre-SBP", "pre-DBP", "SBP", "DBP"]}
    return win, basic


col_ok = True
prob_ok = True
for idx in range(len(df)):
    win, basic = row_to_inputs(df.iloc[idx])
    frame = build_feature_frame(win, basic, expected)

    # (1) 逐列比对(排除 DialysisMethod 字符串列与 SKEW_COLS)。
    num_cols = [c for c in expected if c != "DialysisMethod"]
    a = frame[num_cols].iloc[0].astype(float)
    b = ref[num_cols].iloc[idx].astype(float)
    diff = (a - b).abs()
    mism = [c for c in num_cols if c not in SKEW_COLS and diff[c] > TOL]

    # (2) 覆盖 SKEW_COLS 后比 predict_proba。
    frame_ov = frame.copy()
    for c in SKEW_COLS:
        frame_ov[c] = float(ref[c].iloc[idx])
    p_ov = float(model.predict_proba(pre.transform(frame_ov))[0][1])
    dprob = abs(p_ov - float(ref_prob[idx]))

    if mism:
        col_ok = False
    if dprob > TOL:
        prob_ok = False
    print(f"row{idx}: prob_overlaid={p_ov:.6f} ref={ref_prob[idx]:.6f} dprob={dprob:.2e} | extra_mismatch={mism}")

print("PARITY", "PASS" if (col_ok and prob_ok) else "FAIL",
      f"(col_ok={col_ok}, prob_ok={prob_ok}, skew_cols={sorted(SKEW_COLS)})")
sys.exit(0 if (col_ok and prob_ok) else 1)
