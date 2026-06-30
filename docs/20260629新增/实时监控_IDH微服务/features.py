"""IDH 预警在线特征工程 —— 产出与训练 preprocessor 同名同序的原始特征行,自对齐到
preprocessor.feature_names_in_。差距(可行性闸实测):丢 LogTime_*;补当前血压
SBP/DBP/MeanArterialPressure;缺列(如 TreatmentId_std)由 reindex 填 0;DialysisMethod
未知类别夹到训练类别默认 HD(避免 OneHotEncoder(drop='first') 报错)。"""
from __future__ import annotations
import numpy as np
import pandas as pd

# 训练侧 2_0 拉平列顺序(LogTime 仅用于排序,不作展开特征)。
FLATTEN_COLS = [
    "LogTime", "TMP", "UFVolume", "VenousPressure", "ArterialPressure", "BF",
    "Conductivity", "APumpSpeedDeviation", "BPumpSpeedDeviation", "HeparinPumpFlow",
    "AConductivity", "DialysateTemp", "TreatmentTime", "UFSetVolume", "UFQuantity",
    "BConductivity", "DeviceId", "SubstituateVolume", "HeparinVolume", "SubstituateSpeed",
]
WINDOW = 30
BASIC_COLS = ["Gender", "Age", "DialysisMethod", "DryWeight",
              "UFQuantity_y", "pre-Weight", "pre-SBP", "pre-DBP"]
# 展开/统计用的数值设备列(排除 LogTime —— 训练 preprocessor 不含 LogTime_*)。
NUM_COLS = [c for c in FLATTEN_COLS if c != "LogTime"]
DIALYSIS_METHODS = {"HD", "HDF", "HF", "HFD", "HP", "HP+HD"}
DEFAULT_METHOD = "HD"


def _num(v) -> float:
    try:
        return float(v)
    except (TypeError, ValueError):
        return np.nan


def build_feature_frame(window: list[dict], basic: dict, expected_cols: list[str]) -> pd.DataFrame:
    """组装单行原始特征 DataFrame 并 reindex 到 expected_cols(同名同序)。
    window: DMLog 时点升序(末尾最新),键为 FLATTEN_COLS;不足 30 取已有。
    basic:  BASIC_COLS + 当前 SBP/DBP。
    expected_cols: preprocessor.feature_names_in_。
    """
    rows = window[-WINDOW:]
    flat: dict[str, float] = {}

    # 当前测量血压标量 + MAP(=(SBP+2*DBP)/3,与床卡同口径)。
    sbp = _num(basic.get("SBP"))
    dbp = _num(basic.get("DBP"))
    flat["SBP"] = sbp
    flat["DBP"] = dbp
    flat["MeanArterialPressure"] = (sbp + 2.0 * dbp) / 3.0 if not (np.isnan(sbp) or np.isnan(dbp)) else np.nan

    # 设备列拉平 {col}_{i+1}(不含 LogTime),外层时点、内层列(与训练一致)。
    cols = [f"{c}_{i + 1}" for i in range(WINDOW) for c in NUM_COLS]
    vals: list[float] = []
    for r in rows:
        for c in NUM_COLS:
            vals.append(_num(r.get(c)))
    for i in range(min(len(vals), len(cols))):
        flat[cols[i]] = vals[i]

    # 每设备列窗口均值/标准差。
    arr = pd.DataFrame([{c: _num(r.get(c)) for c in NUM_COLS} for r in rows])
    for c in NUM_COLS:
        flat[f"{c}_mean"] = float(arr[c].mean()) if len(arr) else np.nan
        flat[f"{c}_std"] = float(arr[c].std()) if len(arr) > 1 else 0.0

    # 基本信息(DialysisMethod 单独兜底)。
    for c in BASIC_COLS:
        if c == "DialysisMethod":
            continue
        if c == "Gender":
            flat[c] = int(basic["Gender"]) if basic.get("Gender") is not None else np.nan
        else:
            flat[c] = _num(basic.get(c))
    method = str(basic.get("DialysisMethod") or DEFAULT_METHOD)
    if method not in DIALYSIS_METHODS:
        method = DEFAULT_METHOD

    frame = pd.DataFrame([flat])
    # 自对齐:缺列补 0(数值),多列(LogTime_*)丢弃。DialysisMethod 单独置入(字符串)。
    frame = frame.reindex(columns=[c for c in expected_cols if c != "DialysisMethod"], fill_value=0.0)
    if "DialysisMethod" in expected_cols:
        frame["DialysisMethod"] = method
    # 恢复 expected 顺序。
    frame = frame.reindex(columns=expected_cols)
    return frame
