import os
import pickle
import logging
from typing import List, Optional
import numpy as np

logger = logging.getLogger(__name__)

BASE_DIR      = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ARTIFACTS_DIR = os.path.join(BASE_DIR, "artifacts")
MODEL_PATH    = os.path.join(ARTIFACTS_DIR, "model.keras")
TOKENIZER_PATH = os.path.join(ARTIFACTS_DIR, "tokenizer.pkl")
SCALER_PATH   = os.path.join(ARTIFACTS_DIR, "feature_scaler.pkl")

TEMPERATURE = 2.0


class FakeReviewPredictor:

    def __init__(self):
        self.model     = None
        self.tokenizer = None
        self.scaler    = None
        self._fake_idx = 1  

    def load(self) -> None:
        for path, name in [(MODEL_PATH, "Модель"), (TOKENIZER_PATH, "Токенизатор"), (SCALER_PATH, "Scaler")]:
            if not os.path.exists(path):
                raise FileNotFoundError(f"{name} не найден: {path}\nЗапустите: python -m training.train")

        from keras.models import load_model
        self.model = load_model(MODEL_PATH)

        with open(TOKENIZER_PATH, "rb") as f:
            self.tokenizer = pickle.load(f)

        with open(SCALER_PATH, "rb") as f:
            self.scaler = pickle.load(f)

        logger.info(f"Модель загружена. Классов: {self.model.output_shape[-1]}")

    @property
    def is_loaded(self) -> bool:
        return self.model is not None and self.tokenizer is not None and self.scaler is not None

    def predict_batch(
        self,
        texts: List[str],
        ratings: Optional[List[Optional[str]]] = None,
    ) -> List[float]:

        if not self.is_loaded:
            raise RuntimeError("Модель не загружена")
        if not texts:
            return []

        from training.tokenizer import texts_to_padded
        from features.feature_pipeline import extract_features_batch

        text_seq = texts_to_padded(self.tokenizer, texts)

        features = extract_features_batch(texts, ratings)
        features_norm = self.scaler.transform(features).astype(np.float32)

        probs = self.model.predict([text_seq, features_norm], verbose=0)

        # Температурное масштабирование: поменять

        log_probs = np.log(np.clip(probs, 1e-7, 1.0))
        scaled = log_probs / TEMPERATURE
        
        scaled -= scaled.max(axis=1, keepdims=True)
        exp_scaled = np.exp(scaled)
        calibrated = exp_scaled / exp_scaled.sum(axis=1, keepdims=True)

        num_classes = calibrated.shape[1]
        if num_classes == 1:
            return [round(float(p[0]), 4) for p in calibrated]
        return [round(float(p[self._fake_idx]), 4) for p in calibrated]

    def predict_one(
        self,
        text: str,
        rating: Optional[str] = None,
    ) -> float:
        return self.predict_batch([text], [rating])[0]


predictor = FakeReviewPredictor()
