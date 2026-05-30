from typing import List, Optional
import numpy as np

from . import extractors as ex


FEATURE_NAMES = [
    "similarity_to_others",   # шаблонность/повторяемость
    "lack_of_specifics",      #отсутствие конкретики
    "superlative_ratio",      #восторженный стиль
    "ad_cliche_ratio",        #рекламная лексика
    "generic_word_ratio",     #много общих слов
]

NUM_FEATURES = len(FEATURE_NAMES) 


def extract_features(
    text: str,
    rating: Optional[str] = None,
    other_texts: Optional[List[str]] = None,
) -> np.ndarray:
 
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

    n = len(texts)
    ratings = ratings or [None] * n
    matrix = np.zeros((n, NUM_FEATURES), dtype=np.float32)
    for i, text in enumerate(texts):
        others = texts[:i] + texts[i + 1:]
        matrix[i] = extract_features(text, ratings[i], others)
    return matrix
