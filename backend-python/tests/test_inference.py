"""
Тесты для app/inference.py — FakeReviewPredictor.

Тестируем логику предиктора без реальной модели (моки).

Запуск:
    cd backend-python
    python -m pytest tests/test_inference.py -v
"""

import pytest
import numpy as np
from unittest.mock import patch, MagicMock
from app.inference import FakeReviewPredictor, TEMPERATURE


class TestFakeReviewPredictor:
    def test_init(self):
        pred = FakeReviewPredictor()
        assert pred.model is None
        assert pred.tokenizer is None
        assert pred.scaler is None

    def test_is_loaded_false_by_default(self):
        pred = FakeReviewPredictor()
        assert pred.is_loaded is False

    def test_is_loaded_partial(self):
        pred = FakeReviewPredictor()
        pred.model = MagicMock()
        assert pred.is_loaded is False

        pred.tokenizer = MagicMock()
        assert pred.is_loaded is False

        pred.scaler = MagicMock()
        assert pred.is_loaded is True

    def test_load_missing_artifacts(self):
        pred = FakeReviewPredictor()
        with patch("os.path.exists", return_value=False):
            with pytest.raises(FileNotFoundError):
                pred.load()

    def test_predict_batch_not_loaded(self):
        pred = FakeReviewPredictor()
        with pytest.raises(RuntimeError, match="не загружена"):
            pred.predict_batch(["текст"])

    def test_predict_batch_empty(self):
        pred = FakeReviewPredictor()
        pred.model = MagicMock()
        pred.tokenizer = MagicMock()
        pred.scaler = MagicMock()
        result = pred.predict_batch([])
        assert result == []

    def test_temperature_value(self):
        assert TEMPERATURE == 2.0

    def test_fake_idx(self):
        pred = FakeReviewPredictor()
        assert pred._fake_idx == 1


class TestTemperatureScaling:
    """Тестируем логику temperature scaling изолированно."""

    def test_temperature_makes_probs_softer(self):
        """При T>1 вероятности должны стать ближе к 0.5."""
        # Имитируем: исходный softmax [0.1, 0.9]
        probs = np.array([[0.1, 0.9]])

        log_probs = np.log(np.clip(probs, 1e-7, 1.0))
        scaled = log_probs / TEMPERATURE
        scaled -= scaled.max(axis=1, keepdims=True)
        exp_scaled = np.exp(scaled)
        calibrated = exp_scaled / exp_scaled.sum(axis=1, keepdims=True)

        # После T=2.0, [0.1, 0.9] → ближе к [0.3, 0.7]
        assert calibrated[0, 1] < 0.9
        assert calibrated[0, 1] > 0.5
        assert calibrated[0, 0] > 0.1

    def test_temperature_1_preserves(self):
        """При T=1 вероятности не меняются."""
        probs = np.array([[0.2, 0.8]])
        T = 1.0

        log_probs = np.log(np.clip(probs, 1e-7, 1.0))
        scaled = log_probs / T
        scaled -= scaled.max(axis=1, keepdims=True)
        exp_scaled = np.exp(scaled)
        calibrated = exp_scaled / exp_scaled.sum(axis=1, keepdims=True)

        np.testing.assert_allclose(calibrated, probs, atol=1e-6)

    def test_calibrated_sums_to_one(self):
        """Калиброванные вероятности суммируются в 1."""
        probs = np.array([[0.05, 0.95], [0.5, 0.5], [0.9, 0.1]])

        log_probs = np.log(np.clip(probs, 1e-7, 1.0))
        scaled = log_probs / TEMPERATURE
        scaled -= scaled.max(axis=1, keepdims=True)
        exp_scaled = np.exp(scaled)
        calibrated = exp_scaled / exp_scaled.sum(axis=1, keepdims=True)

        for i in range(len(probs)):
            assert abs(calibrated[i].sum() - 1.0) < 1e-6
