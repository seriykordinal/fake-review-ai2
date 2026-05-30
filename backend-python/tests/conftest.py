import pytest
from unittest.mock import MagicMock, patch
from fastapi.testclient import TestClient


@pytest.fixture
def client():
    """TestClient с замоканным предиктором (модель загружена)."""
    mock_pred = MagicMock()
    mock_pred.is_loaded = True
    mock_pred.predict_one.return_value = 0.75
    mock_pred.predict_batch.return_value = [0.5, 0.6, 0.7]

    with patch("app.main.predictor", mock_pred):
        from app.main import app
        with TestClient(app, raise_server_exceptions=False) as c:
            yield c, mock_pred


@pytest.fixture
def client_model_not_loaded():
    """TestClient с замоканным предиктором (модель НЕ загружена)."""
    mock_pred = MagicMock()
    mock_pred.is_loaded = False

    with patch("app.main.predictor", mock_pred):
        from app.main import app
        with TestClient(app, raise_server_exceptions=False) as c:
            yield c, mock_pred
