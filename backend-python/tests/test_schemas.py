"""
Тесты для app/schemas.py — Pydantic-модели запросов и ответов.

Запуск:
    cd backend-python
    python -m pytest tests/test_schemas.py -v
"""

import pytest
from pydantic import ValidationError
from app.schemas import (
    TextRequest,
    TextBatchRequest,
    PredictionResponse,
    BatchPredictionResponse,
    HealthResponse,
)


# ── TextRequest ──────────────────────────────────────────────────────────────

class TestTextRequest:
    def test_valid(self):
        req = TextRequest(text="Отличный товар")
        assert req.text == "Отличный товар"
        assert req.rating is None

    def test_with_rating(self):
        req = TextRequest(text="Текст", rating="5")
        assert req.rating == "5"

    def test_empty_text_rejected(self):
        with pytest.raises(ValidationError):
            TextRequest(text="")

    def test_missing_text_rejected(self):
        with pytest.raises(ValidationError):
            TextRequest()

    def test_rating_optional(self):
        req = TextRequest(text="Текст")
        assert req.rating is None


# ── TextBatchRequest ─────────────────────────────────────────────────────────

class TestTextBatchRequest:
    def test_valid(self):
        req = TextBatchRequest(texts=["Текст 1", "Текст 2"])
        assert len(req.texts) == 2

    def test_empty_list(self):
        req = TextBatchRequest(texts=[])
        assert req.texts == []

    def test_missing_texts_rejected(self):
        with pytest.raises(ValidationError):
            TextBatchRequest()


# ── PredictionResponse ───────────────────────────────────────────────────────

class TestPredictionResponse:
    def test_valid(self):
        resp = PredictionResponse(fake_probability=0.75)
        assert resp.fake_probability == 0.75

    def test_zero(self):
        resp = PredictionResponse(fake_probability=0.0)
        assert resp.fake_probability == 0.0

    def test_one(self):
        resp = PredictionResponse(fake_probability=1.0)
        assert resp.fake_probability == 1.0

    def test_negative_rejected(self):
        with pytest.raises(ValidationError):
            PredictionResponse(fake_probability=-0.1)

    def test_above_one_rejected(self):
        with pytest.raises(ValidationError):
            PredictionResponse(fake_probability=1.1)


# ── BatchPredictionResponse ──────────────────────────────────────────────────

class TestBatchPredictionResponse:
    def test_valid(self):
        resp = BatchPredictionResponse(probabilities=[0.1, 0.5, 0.9])
        assert len(resp.probabilities) == 3

    def test_empty(self):
        resp = BatchPredictionResponse(probabilities=[])
        assert resp.probabilities == []


# ── HealthResponse ───────────────────────────────────────────────────────────

class TestHealthResponse:
    def test_ok(self):
        resp = HealthResponse(status="ok", model_loaded=True)
        assert resp.status == "ok"
        assert resp.model_loaded is True

    def test_loading(self):
        resp = HealthResponse(status="loading", model_loaded=False)
        assert resp.model_loaded is False
