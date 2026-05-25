"""
Единая точка вычисления вектора признаков для отзыва.

5 признаков, основанных на источниках rskrf и techscience:
  1. similarity_to_others   — шаблонность (сходство с другими отзывами товара)
  2. lack_of_specifics      — отсутствие конкретики
  3. superlative_ratio      — восторженный стиль
  4. ad_cliche_ratio        — рекламная лексика
  5. generic_word_ratio     — много общих слов вместо предметных

Использование:
    from features.feature_pipeline import extract_features, FEATURE_NAMES
    vec = extract_features("Отличный товар!", rating="5", other_texts=[...])
    # → numpy array shape (5,)
"""

from typing import List, Optional
import numpy as np

from . import extractors as ex


FEATURE_NAMES = [
    "similarity_to_others",   # шаблонность / повторяемость
    "lack_of_specifics",      # отсутствие конкретики
    "superlative_ratio",      # восторженный стиль
    "ad_cliche_ratio",        # рекламная лексика
    "generic_word_ratio",     # много общих слов
]

NUM_FEATURES = len(FEATURE_NAMES)  # 5


def extract_features(
    text: str,
    rating: Optional[str] = None,
    other_texts: Optional[List[str]] = None,
) -> np.ndarray:
    """
    Возвращает вектор из 5 признаков для одного отзыва.

    Параметры:
        text        — объединённый текст (pros + cons + comment)
        rating      — оценка "1".."5", может быть None
        other_texts — другие отзывы того же товара (для similarity)
    """
    if other_texts is None:
        other_texts = []

    return np.array([
        ex.similarity_to_others(text, other_texts),
        ex.lack_of_specifics(text),
        ex.superlative_ratio(text),
        ex.ad_cliche_ratio(text),
        ex.generic_word_ratio(text),
    ], dtype=np.float32)


def extract_features_batch(
    texts: List[str],
    ratings: Optional[List[Optional[str]]] = None,
) -> np.ndarray:
    """
    Извлекает признаки для группы отзывов одного товара.
    similarity_to_others вычисляется внутри группы автоматически.

    Возвращает numpy array формы (N, 5).
    """
    n = len(texts)
    ratings = ratings or [None] * n
    matrix = np.zeros((n, NUM_FEATURES), dtype=np.float32)
    for i, text in enumerate(texts):
        others = texts[:i] + texts[i + 1:]
        matrix[i] = extract_features(text, ratings[i], others)
    return matrix
