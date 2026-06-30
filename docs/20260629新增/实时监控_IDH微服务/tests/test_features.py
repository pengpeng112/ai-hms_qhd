import os, sys, pickle
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
import numpy as np
from features import build_feature_frame, FLATTEN_COLS

BASE = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def _expected_cols():
    with open(os.path.join(BASE, "models", "scaler.pkl"), "rb") as f:
        pre = pickle.load(f)
    return list(pre.feature_names_in_)


def test_frame_matches_expected_columns_exactly():
    expected = _expected_cols()
    win = [{c: float(i + 1) for c in FLATTEN_COLS} for i in range(30)]
    basic = {"Gender": 1, "Age": 65.0, "DialysisMethod": "HD", "DryWeight": 60.0,
             "UFQuantity_y": 2000.0, "pre-Weight": 62.0, "pre-SBP": 140.0, "pre-DBP": 80.0,
             "SBP": 150.0, "DBP": 90.0}
    frame = build_feature_frame(win, basic, expected)
    assert list(frame.columns) == expected  # 同名同序


def test_current_bp_and_map_filled():
    expected = _expected_cols()
    win = [{c: 1.0 for c in FLATTEN_COLS} for _ in range(30)]
    basic = {"DialysisMethod": "HD", "SBP": 150.0, "DBP": 90.0}
    frame = build_feature_frame(win, basic, expected)
    assert float(frame.iloc[0]["SBP"]) == 150.0
    assert float(frame.iloc[0]["DBP"]) == 90.0
    # MAP = (SBP + 2*DBP)/3 = (150+180)/3 = 110
    assert abs(float(frame.iloc[0]["MeanArterialPressure"]) - 110.0) < 1e-9


def test_unknown_dialysis_method_clamped_to_hd():
    expected = _expected_cols()
    win = [{c: 1.0 for c in FLATTEN_COLS} for _ in range(30)]
    basic = {"DialysisMethod": "ZZZ", "SBP": 120.0, "DBP": 70.0}
    frame = build_feature_frame(win, basic, expected)
    assert frame.iloc[0]["DialysisMethod"] == "HD"


def test_logtime_flatten_dropped():
    expected = _expected_cols()
    win = [{c: 1.0 for c in FLATTEN_COLS} for _ in range(30)]
    frame = build_feature_frame(win, {"DialysisMethod": "HD"}, expected)
    assert not any(c.startswith("LogTime_") for c in frame.columns)
