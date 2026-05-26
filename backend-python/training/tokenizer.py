"""
Токенизатор для текстовых отзывов.

Преобразует тексты в последовательности целочисленных идентификаторов слов
фиксированной длины. Готовая последовательность подаётся на Embedding-слой
нейронной сети.

Сохраняет и загружает токенизатор через pickle.
"""

import os
import pickle
from typing import List, Tuple

import numpy as np
from keras.preprocessing import Tokenizer
from keras.preprocessing.sequence import pad_sequences


# Гиперпараметры токенизации
VOCAB_SIZE  = 10000   # размер словаря (берём top-N самых частых слов)
MAX_LENGTH  = 100     # максимальная длина последовательности (слов)
OOV_TOKEN   = "<OOV>" # токен для неизвестных слов
PAD_TYPE    = "post"  # дополнение нулями в конце последовательности
TRUNC_TYPE  = "post"  # обрезание длинных последовательностей с конца


def fit_tokenizer(texts: List[str]) -> Tokenizer:
    """Обучает токенизатор на корпусе текстов."""
    tokenizer = Tokenizer(num_words=VOCAB_SIZE, oov_token=OOV_TOKEN)
    tokenizer.fit_on_texts(texts)
    return tokenizer


def texts_to_padded(tokenizer: Tokenizer, texts: List[str]) -> np.ndarray:
    """Преобразует список текстов в padded-последовательности."""
    sequences = tokenizer.texts_to_sequences(texts)
    padded = pad_sequences(
        sequences,
        maxlen=MAX_LENGTH,
        padding=PAD_TYPE,
        truncating=TRUNC_TYPE,
    )
    return padded


def save_tokenizer(tokenizer: Tokenizer, path: str) -> None:
    """Сохраняет токенизатор в pickle."""
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "wb") as f:
        pickle.dump(tokenizer, f, protocol=pickle.HIGHEST_PROTOCOL)


def load_tokenizer(path: str) -> Tokenizer:
    """Загружает токенизатор из pickle."""
    with open(path, "rb") as f:
        return pickle.load(f)


def prepare_training_data(
    texts: List[str],
) -> Tuple[Tokenizer, np.ndarray]:
    """Утилита: обучает токенизатор и сразу возвращает padded-данные."""
    tokenizer = fit_tokenizer(texts)
    padded = texts_to_padded(tokenizer, texts)
    return tokenizer, padded
