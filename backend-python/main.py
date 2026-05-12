
"""
FastAPI сервис для определения фейковых отзывов.

Датасет deepRost/wb-reviews имеет 3 класса:
  0 = негативный отзыв  -> используем как "фейк/плохой"
  1 = нейтральный
  2 = позитивный

fake_probability = P(класс == 0), т.е. вероятность негативного/подозрительного отзыва.

Запуск: uvicorn main:app --host 0.0.0.0 --port 8000
"""

import os
import pickle
import logging
from contextlib import asynccontextmanager
from typing import List

import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

logger = logging.getLogger(__name__)

BASE_DIR   = os.path.dirname(__file__)
MODEL_PATH = os.path.join(BASE_DIR, "models", "fake_review_model.h5")
TFIDF_PATH = os.path.join(BASE_DIR, "models", "tfidf_vectorizer.pkl")

_model = None
_tfidf = None


def load_artifacts():
    global _model, _tfidf
    for path in (MODEL_PATH, TFIDF_PATH):
        if not os.path.exists(path):
            raise FileNotFoundError(
                f"Файл не найден: {path}\n"
                "Запустите сначала: python train.py"
            )
    from keras.models import load_model
    _model = load_model(MODEL_PATH)
    with open(TFIDF_PATH, "rb") as f:
        _tfidf = pickle.load(f)

    num_classes = _model.output_shape[-1]
    logger.info(f"Модель загружена. Классов: {num_classes}")


@asynccontextmanager
async def lifespan(app: FastAPI):
    load_artifacts()
    yield


app = FastAPI(title="FakeReview Detector", lifespan=lifespan)


# ───────────── Схемы ─────────────
class TextRequest(BaseModel):
    text: str

class TextListRequest(BaseModel):
    texts: List[str]

class PredictionResponse(BaseModel):
    fake_probability: float

class BatchPredictionResponse(BaseModel):
    probabilities: List[float]


# ───────────── Предсказание ─────────────
def _predict_batch(texts: List[str]) -> List[float]:
    """
    Возвращает fake_probability для каждого текста.

    Логика:
    - Класс 0 = негативный отзыв (наиболее подозрительный / нечестный)
    - При 3 классах: fake = P(0), при 2 классах: fake = P(0)
    - Если модель выдаёт только 1 класс — используем его напрямую
    """
    X = _tfidf.transform(texts).toarray().astype(np.float32)
    probs = _model.predict(X, verbose=0)  # shape: (n, num_classes)

    num_classes = probs.shape[1]
    if num_classes == 1:
        # Бинарный выход без softmax — маловероятно, но на всякий случай
        return [round(float(p[0]), 4) for p in probs]
    else:
        # P(класс 0) = вероятность негативного/фейкового отзыва
        return [round(float(p[0]), 4) for p in probs]


# ───────────── Эндпоинты ─────────────
@app.post("/predict", response_model=PredictionResponse)
async def predict(request: TextRequest):
    if _model is None or _tfidf is None:
        raise HTTPException(503, "Модель не загружена")
    return PredictionResponse(fake_probability=_predict_batch([request.text])[0])


@app.post("/predict_batch", response_model=BatchPredictionResponse)
async def predict_batch(request: TextListRequest):
    if _model is None or _tfidf is None:
        raise HTTPException(503, "Модель не загружена")
    if not request.texts:
        return BatchPredictionResponse(probabilities=[])
    return BatchPredictionResponse(probabilities=_predict_batch(request.texts))


@app.get("/health")
async def health():
    return {"status": "ok", "model_loaded": _model is not None}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
