"""Pydantic-схемы запросов и ответов FastAPI."""

from typing import List, Optional
from pydantic import BaseModel, Field


class TextRequest(BaseModel):
    """Запрос на анализ одного отзыва."""
    text: str = Field(..., min_length=1)
    rating: Optional[str] = Field(None)


class TextBatchRequest(BaseModel):
    """Пакетный запрос — Go отправляет {"texts": [...]}"""
    texts: List[str]


class PredictionResponse(BaseModel):
    fake_probability: float = Field(..., ge=0.0, le=1.0)


class BatchPredictionResponse(BaseModel):
    probabilities: List[float]


class HealthResponse(BaseModel):
    status: str
    model_loaded: bool
