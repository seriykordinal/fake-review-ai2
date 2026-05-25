"""
Функции извлечения 5 признаков фейковых отзывов.

Признаки взяты из:
  - rskrf.ru — Роскачество: признаки фейковых отзывов
  - techscience.ru — n-граммы, TF-IDF, similarity как признаки фейков

Все функции принимают текст (str) и возвращают float.
Вспомогательные функции принимают дополнительные параметры.
"""

import re
from typing import List, Optional

from .lexicons import (
    AD_CLICHES, SUPERLATIVES, GENERIC_WORDS,
    SPECIFIC_DETAIL_MARKERS, POSITIVE_WORDS, NEGATIVE_WORDS,
)

_WORD_RE        = re.compile(r"\b[а-яёА-ЯЁa-zA-Z]+\b")
_REPEAT_PUNCT_RE = re.compile(r"([!?])\1+")


def _tokenize(text: str) -> List[str]:
    return [w.lower() for w in _WORD_RE.findall(text)]


def _bigrams(tokens: List[str]) -> List[str]:
    return [f"{tokens[i]} {tokens[i+1]}" for i in range(len(tokens) - 1)]


# ── 1. Шаблонность / повторяемость ────────────────────────────────────────────
# Источники: techscience, rskrf
# Метод: cosine similarity по TF-IDF между текущим отзывом и остальными
# отзывами того же товара. Высокое сходство = скоординированная накрутка.

def similarity_to_others(text: str, other_texts: List[str]) -> float:
    """Максимальное Жаккарда-сходство с другими отзывами товара.
    Значение от 0 до 1. Высокое (>0.6) = шаблонность."""
    if not other_texts:
        return 0.0
    bigrams_text = set(_bigrams(_tokenize(text)))
    if not bigrams_text:
        # Только уникаграммы если текст очень короткий
        bigrams_text = set(_tokenize(text))
    if not bigrams_text:
        return 0.0
    max_sim = 0.0
    for other in other_texts:
        other_bg = set(_bigrams(_tokenize(other)))
        if not other_bg:
            other_bg = set(_tokenize(other))
        if not other_bg:
            continue
        inter = bigrams_text & other_bg
        union = bigrams_text | other_bg
        sim = len(inter) / len(union)
        if sim > max_sim:
            max_sim = sim
    return round(max_sim, 4)


# ── 2. Отсутствие конкретики ───────────────────────────────────────────────────
# Источники: rskrf, techscience
# Метод: доля слов из словаря конкретных деталей в тексте.
# Чем выше — тем больше конкретики (признак подлинного отзыва).
# Инвертируем: 1 - ratio → высокое значение = мало конкретики = фейк.

def lack_of_specifics(text: str) -> float:
    """Отсутствие конкретики: 1 − доля конкретных слов.
    0 — много конкретики (подлинный), 1 — нет конкретики (подозрительно)."""
    tokens = _tokenize(text)
    if not tokens:
        return 1.0  # пустой текст — максимальная неконкретность
    specific_count = sum(1 for t in tokens if t in SPECIFIC_DETAIL_MARKERS)
    return round(1.0 - specific_count / len(tokens), 4)


# ── 3. Восторженный стиль ─────────────────────────────────────────────────────
# Источник: rskrf — "неадекватно восторженный взгляд",
# слова в превосходной степени как признак фейка

def superlative_ratio(text: str) -> float:
    """Доля слов из словаря превосходных степеней.
    Высокая доля = восторженный стиль = подозрительно."""
    tokens = _tokenize(text)
    if not tokens:
        return 0.0
    count = sum(1 for t in tokens if t in SUPERLATIVES)
    # Также считаем восклицания — часть эмоционального стиля
    excl = text.count("!")
    excl_ratio = min(excl / max(len(text), 1) * 5, 1.0)  # нормируем
    ratio = (count / len(tokens) + excl_ratio) / 2
    return round(ratio, 4)


# ── 4. Рекламная лексика ──────────────────────────────────────────────────────
# Источник: rskrf — рекламные формулировки и призывы как подозрительный признак

def ad_cliche_ratio(text: str) -> float:
    """Доля рекламных клише в тексте (по словарю + биграммам).
    Высокая = рекламный стиль = подозрительно."""
    tokens = _tokenize(text)
    if not tokens:
        return 0.0
    text_lower = text.lower()
    # Проверяем клише как подстроку (они могут быть фразами)
    cliche_hits = sum(1 for c in AD_CLICHES if c in text_lower)
    # Нормируем по длине текста в словах
    ratio = min(cliche_hits / max(len(tokens) / 5, 1), 1.0)
    return round(ratio, 4)


# ── 5. Много общих слов / мало предметных ────────────────────────────────────
# Источник: rskrf — абстрактные формулировки, применимые "к чему угодно"

def generic_word_ratio(text: str) -> float:
    """Доля обобщённых оценочных слов без конкретики.
    Высокая = "всё хорошо, ничего конкретного" = подозрительно."""
    tokens = _tokenize(text)
    if not tokens:
        return 1.0
    generic_count = sum(1 for t in tokens if t in GENERIC_WORDS)
    return round(generic_count / len(tokens), 4)


# ── Вспомогательный: несоответствие рейтинга и тональности ───────────────────
# Используется в разметке, не как обучающий признак

def rating_sentiment_mismatch(text: str, rating: Optional[str]) -> float:
    """Несоответствие между рейтингом и тональностью текста.
    0 — согласованы, 1 — сильное несоответствие."""
    if not rating:
        return 0.0
    try:
        r = float(str(rating).strip())
    except (ValueError, TypeError):
        return 0.0
    tokens = _tokenize(text)
    if not tokens:
        return 0.0
    pos = sum(1 for t in tokens if t in POSITIVE_WORDS)
    neg = sum(1 for t in tokens if t in NEGATIVE_WORDS)
    if pos + neg == 0:
        return 0.0
    sentiment = (pos - neg) / (pos + neg)          # -1..+1
    rating_norm = (r - 3.0) / 2.0                   # -1..+1
    return round(min(abs(sentiment - rating_norm) / 2.0, 1.0), 4)
