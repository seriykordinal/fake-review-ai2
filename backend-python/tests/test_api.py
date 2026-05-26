"""
Тесты для app/main.py — FastAPI эндпоинты /health, /predict, /predict_batch.

Запуск:
    cd backend-python
    python -m pytest tests/test_api.py -v
"""

import pytest
from unittest.mock import patch, MagicMock
from fastapi.testclient import TestClient


# Мокаем predictor ДО импорта app, чтобы lifespan не пытался загрузить модель
@pytest.fixture()
def client():
    with patch("app.main.predictor") as mock_pred:
        mock_pred.is_loaded = True
        mock_pred.predict_one.return_value = 0.75
        mock_pred.predict_batch.return_value = [0.5, 0.6, 0.7]
        mock_pred.load.return_value = None

        from app.main import app
        with TestClient(app) as c:
            yield c, mock_pred


@pytest.fixture()
def client_model_not_loaded():
    with patch("app.main.predictor") as mock_pred:
        mock_pred.is_loaded = False
        mock_pred.load.return_value = None

        from app.main import app
        with TestClient(app) as c:
            yield c, mock_pred


# ── /health ──────────────────────────────────────────────────────────────────

class TestHealth:
    def test_health_loaded(self, client):
        c, _ = client
        resp = c.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "ok"
        assert data["model_loaded"] is True

    def test_health_not_loaded(self, client_model_not_loaded):
        c, _ = client_model_not_loaded
        resp = c.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "loading"
        assert data["model_loaded"] is False


# ── /predict ─────────────────────────────────────────────────────────────────

class TestPredict:
    def test_predict_success(self, client):
        c, mock_pred = client
        resp = c.post("/predict", json={"text": "Отличный товар"})
        assert resp.status_code == 200
        data = resp.json()
        assert data["fake_probability"] == 0.75
        mock_pred.predict_one.assert_called_once_with(text="Отличный товар", rating=None)

    def test_predict_with_rating(self, client):
        c, mock_pred = client
        resp = c.post("/predict", json={"text": "Текст", "rating": "5"})
        assert resp.status_code == 200
        mock_pred.predict_one.assert_called_once_with(text="Текст", rating="5")

    def test_predict_empty_text_rejected(self, client):
        c, _ = client
        resp = c.post("/predict", json={"text": ""})
        assert resp.status_code == 422  # Pydantic validation error

    def test_predict_missing_text(self, client):
        c, _ = client
        resp = c.post("/predict", json={})
        assert resp.status_code == 422

    def test_predict_invalid_json(self, client):
        c, _ = client
        resp = c.post("/predict", content="not json",
                       headers={"Content-Type": "application/json"})
        assert resp.status_code == 422

    def test_predict_model_not_loaded(self, client_model_not_loaded):
        c, _ = client_model_not_loaded
        resp = c.post("/predict", json={"text": "Текст"})
        assert resp.status_code == 503


# ── /predict_batch ───────────────────────────────────────────────────────────

class TestPredictBatch:
    def test_batch_success(self, client):
        c, mock_pred = client
        texts = ["Отзыв 1", "Отзыв 2", "Отзыв 3"]
        resp = c.post("/predict_batch", json={"texts": texts})
        assert resp.status_code == 200
        data = resp.json()
        assert data["probabilities"] == [0.5, 0.6, 0.7]
        mock_pred.predict_batch.assert_called_once_with(texts=texts)

    def test_batch_empty_list(self, client):
        c, _ = client
        resp = c.post("/predict_batch", json={"texts": []})
        assert resp.status_code == 200
        data = resp.json()
        assert data["probabilities"] == []

    def test_batch_missing_texts(self, client):
        c, _ = client
        resp = c.post("/predict_batch", json={})
        assert resp.status_code == 422

    def test_batch_model_not_loaded(self, client_model_not_loaded):
        c, _ = client_model_not_loaded
        resp = c.post("/predict_batch", json={"texts": ["Текст"]})
        assert resp.status_code == 503
