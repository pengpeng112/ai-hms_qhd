"""可行性闸探针：在 py3.12 venv 下加载 CatBoost + preprocessor，构造样本窗口，
端到端 transform+predict_proba，并比对 build_feature_frame 产列 vs preprocessor 期望列。"""
import pickle, sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
import numpy as np
from features import build_feature_frame, FLATTEN_COLS

base = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
with open(os.path.join(base, "models/CatBoost_model.pkl"), "rb") as f:
    model = pickle.load(f)
with open(os.path.join(base, "models/scaler.pkl"), "rb") as f:
    pre = pickle.load(f)

print("model:", type(model).__name__, "n_features_in_=", getattr(model, "n_features_in_", "?"))
fni = list(getattr(pre, "feature_names_in_", []))
print("preprocessor expects", len(fni), "input cols")
print("  first8:", fni[:8])
print("  last10:", fni[-10:])

win = [{c: 1.0 for c in FLATTEN_COLS} for _ in range(30)]
basic = {"Gender": 1, "Age": 65.0, "DialysisMethod": "HD", "DryWeight": 60.0,
         "UFQuantity_y": 2000.0, "pre-Weight": 62.0, "pre-SBP": 140.0, "pre-DBP": 80.0}
feat = build_feature_frame(win, basic)
print("built frame cols:", feat.shape[1])
built, exp = set(feat.columns), set(fni)
print("MISSING(expected-not-built) n=%d:" % len(exp - built), sorted(exp - built)[:25])
print("EXTRA(built-not-expected) n=%d:" % len(built - exp), sorted(built - exp)[:25])
try:
    X = pre.transform(feat)
    p = float(model.predict_proba(X)[0][1])
    print("END-TO-END SCORE OK  prob=", round(p, 4))
except Exception as e:
    print("TRANSFORM/PREDICT FAILED:", repr(e)[:400])
