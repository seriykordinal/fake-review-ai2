import re
from typing import List, Optional

from .lexicons import (
    AD_CLICHES, SUPERLATIVES, GENERIC_WORDS,
    SPECIFIC_DETAIL_MARKERS, POSITIVE_WORDS, NEGATIVE_WORDS,
)

_WORD_RE = re.compile(r"\b[а-яёА-ЯЁa-zA-Z]+\b")

def _tokenize(text: str) -> List[str]:
    return [w.lower() for w in _WORD_RE.findall(text)]


def _bigrams(tokens: List[str]) -> List[str]:
    return [f"{tokens[i]} {tokens[i+1]}" for i in range(len(tokens) - 1)]


#Шаблонность 

def similarity_to_others(text: str, other_texts: List[str]) -> float:

    if not other_texts:
        return 0.0
    bigrams_text = set(_bigrams(_tokenize(text)))
    if not bigrams_text:
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


#Отсутствие конкретики

def lack_of_specifics(text: str) -> float:

    tokens = _tokenize(text)
    if not tokens:
        return 1.0 
    specific_count = sum(1 for t in tokens if t in SPECIFIC_DETAIL_MARKERS)
    return round(1.0 - specific_count / len(tokens), 4)


#Восторженный стиль

def superlative_ratio(text: str) -> float:
    tokens = _tokenize(text)
    if not tokens:
        return 0.0
    count = sum(1 for t in tokens if t in SUPERLATIVES)
    excl = text.count("!")
    excl_ratio = min(excl / max(len(text), 1) * 5, 1.0) 
    ratio = (count / len(tokens) + excl_ratio) / 2
    return round(ratio, 4)


#Рекламная лексика

def ad_cliche_ratio(text: str) -> float:

    tokens = _tokenize(text)
    if not tokens:
        return 0.0
    text_lower = text.lower()
    cliche_hits = sum(1 for c in AD_CLICHES if c in text_lower)
    ratio = min(cliche_hits / max(len(tokens) / 5, 1), 1.0)
    return round(ratio, 4)


#Много общих слов, мало предметных

def generic_word_ratio(text: str) -> float:
    tokens = _tokenize(text)
    if not tokens:
        return 1.0
    generic_count = sum(1 for t in tokens if t in GENERIC_WORDS)
    return round(generic_count / len(tokens), 4)


#несоответствие рейтинга и тональности

def rating_sentiment_mismatch(text: str, rating: Optional[str]) -> float:

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
    sentiment = (pos - neg) / (pos + neg)
    rating_norm = (r - 3.0) / 2.0
    return round(min(abs(sentiment - rating_norm) / 2.0, 1.0), 4)
