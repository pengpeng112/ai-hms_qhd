import os, sys
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from fastapi.testclient import TestClient
import app as appmod
from features import FLATTEN_COLS


def _win():
    return [{k: 1.0 for k in FLATTEN_COLS} for _ in range(30)]


def test_health_model_loaded():
    with TestClient(appmod.app) as c:
        r = c.get("/health")
        assert r.status_code == 200
        body = r.json()
        assert body["modelLoaded"] is True
        assert body["preprocessorLoaded"] is True


def test_score_returns_probability():
    payload = {
        "treatmentId": 1, "accessType": "AVF",
        "window": _win(),
        "basic": {"Gender": 1, "Age": 65, "DialysisMethod": "HD", "DryWeight": 60,
                  "UFQuantity_y": 2000, "pre-Weight": 62, "pre-SBP": 140, "pre-DBP": 80,
                  "SBP": 150, "DBP": 90},
    }
    with TestClient(appmod.app) as c:
        r = c.post("/idh/score", json=payload)
        assert r.status_code == 200
        body = r.json()
        assert body["available"] is True
        assert 0.0 <= body["probability"] <= 1.0


def test_score_empty_window_unavailable():
    with TestClient(appmod.app) as c:
        r = c.post("/idh/score", json={"window": [], "basic": {}})
        assert r.json()["available"] is False
