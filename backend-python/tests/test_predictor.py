import pytest
from unittest.mock import patch, MagicMock
from app.predictor import FakeReviewPredictor


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


