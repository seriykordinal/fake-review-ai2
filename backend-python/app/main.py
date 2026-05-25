"""
FastAPI приложение FakeCheck ML.

Эндпоинты:
    GET  /health         — статус сервиса
    POST /predict        — анализ одного отзыва
    POST /predict_batch  — пакетный анализ ({"texts": [...]})

Запуск:
    uvicorn app.main:app --host 0.0.0.0 --port 8000
"""

import logging
import sys
import os
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from app.schemas import (
    TextRequest, TextBatchRequest,
    PredictionResponse, BatchPredictionResponse, HealthResponse,
)
from app.inference import predictor

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Загружаем модель...")
    predictor.load()
    logger.info("Модель загружена.")
    yield


app = FastAPI(
    title="FakeCheck ML Service",
    version="2.0.0",
    lifespan=lifespan,
)


@app.get("/health", response_model=HealthResponse)
async def health():
    return HealthResponse(
        status="ok" if predictor.is_loaded else "loading",
        model_loaded=predictor.is_loaded,
    )


@app.post("/predict", response_model=PredictionResponse)
async def predict(request: TextRequest):
    if not predictor.is_loaded:
        raise HTTPException(503, "Модель не загружена")
    prob = predictor.predict_one(text=request.text, rating=request.rating)
    return PredictionResponse(fake_probability=prob)


@app.post("/predict_batch", response_model=BatchPredictionResponse)
async def predict_batch(request: TextBatchRequest):
    """Go отправляет {"texts": [...]} — простой список строк."""
    if not predictor.is_loaded:
        raise HTTPException(503, "Модель не загружена")
    if not request.texts:
        return BatchPredictionResponse(probabilities=[])
    probs = predictor.predict_batch(texts=request.texts)
    return BatchPredictionResponse(probabilities=probs)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000)
