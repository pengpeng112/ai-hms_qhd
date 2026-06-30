"""
IDH（透中低血压）预警 FastAPI 微服务 —— 与 ACTRS 平级，供 ai-hms 后端 HTTPScorer 调用。

加载 dixueya 训练保存的 CatBoost 模型 + ColumnTransformer 预处理器，暴露 POST /idh/score：
喂某治疗近 30 个 Device_DMLog 时点 + 基本信息 → 返回下次测量低血压概率。

模型来源：桌面/HMS开发/dixueya.rar（CatBoost_model.pkl + preprocessor/scaler.pkl）。
真模型表现：CatBoost+SMOTE，AUC≈0.91、特异度 0.997、敏感度 0.67（详见 IDH 评估文档）。

⚠️ 上线前务必做「对拍验证」：用一条训练样本跑通，比对概率与训练侧一致（见 README）。
"""
from __future__ import annotations
import os
import pickle
from pathlib import Path

from fastapi import FastAPI
from pydantic import BaseModel, Field

from features import build_feature_frame

MODEL_PATH = os.environ.get("IDH_MODEL_PATH", "models/CatBoost_model.pkl")
PREPROCESSOR_PATH = os.environ.get("IDH_PREPROCESSOR_PATH", "models/scaler.pkl")

app = FastAPI(title="IDH 预警服务", version="0.1.0")

_model = None
_preprocessor = None
_expected_cols: list[str] = []


def _load(path: str):
    p = Path(path)
    if not p.exists():
        return None
    with open(p, "rb") as f:
        return pickle.load(f)


@app.on_event("startup")
def _startup():
    global _model, _preprocessor, _expected_cols
    _model = _load(MODEL_PATH)
    _preprocessor = _load(PREPROCESSOR_PATH)
    if _preprocessor is not None and hasattr(_preprocessor, "feature_names_in_"):
        _expected_cols = list(_preprocessor.feature_names_in_)


class Sample(BaseModel):
    # 一个 DMLog 时点（键与训练 FLATTEN_COLS 对齐；缺失留空即可）。
    LogTime: float | None = None
    TMP: float | None = None
    UFVolume: float | None = None
    VenousPressure: float | None = None
    ArterialPressure: float | None = None
    BF: float | None = None
    Conductivity: float | None = None
    APumpSpeedDeviation: float | None = None
    BPumpSpeedDeviation: float | None = None
    HeparinPumpFlow: float | None = None
    AConductivity: float | None = None
    DialysateTemp: float | None = None
    TreatmentTime: float | None = None
    UFSetVolume: float | None = None
    UFQuantity: float | None = None
    BConductivity: float | None = None
    DeviceId: float | None = None
    SubstituateVolume: float | None = None
    HeparinVolume: float | None = None
    SubstituateSpeed: float | None = None


class BasicInfo(BaseModel):
    Gender: int | None = Field(default=None, description="男=1 / 女=0")
    Age: float | None = None
    DialysisMethod: str | None = None
    DryWeight: float | None = None
    UFQuantity_y: float | None = Field(default=None, alias="UFQuantity_y")
    pre_Weight: float | None = Field(default=None, alias="pre-Weight")
    pre_SBP: float | None = Field(default=None, alias="pre-SBP")
    pre_DBP: float | None = Field(default=None, alias="pre-DBP")
    SBP: float | None = None
    DBP: float | None = None

    class Config:
        populate_by_name = True


class ScoreRequest(BaseModel):
    treatmentId: int | None = None
    accessType: str | None = None
    window: list[Sample] = Field(default_factory=list)
    basic: BasicInfo = Field(default_factory=BasicInfo)


class ScoreResponse(BaseModel):
    available: bool
    probability: float = 0.0


@app.get("/health")
def health():
    return {"ok": True, "modelLoaded": _model is not None, "preprocessorLoaded": _preprocessor is not None}


@app.post("/idh/score", response_model=ScoreResponse)
def score(req: ScoreRequest):
    # 模型/预处理器未装载，或窗口为空 → 不可用（后端降级显示"待数据"）。
    if _model is None or _preprocessor is None or not req.window or not _expected_cols:
        return ScoreResponse(available=False)
    try:
        window = [s.model_dump() for s in req.window]
        basic = req.basic.model_dump(by_alias=True)
        feats = build_feature_frame(window, basic, _expected_cols)
        x = _preprocessor.transform(feats)
        prob = float(_model.predict_proba(x)[0][1])
        return ScoreResponse(available=True, probability=prob)
    except Exception:
        # 任何特征/推理异常都降级，绝不阻断实时监控。
        return ScoreResponse(available=False)
