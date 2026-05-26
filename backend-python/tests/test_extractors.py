"""
Тесты для features/extractors.py — 5 признаков фейковых отзывов.

Запуск:
    cd backend-python
    python -m pytest tests/test_extractors.py -v
"""

import pytest
from features.extractors import (
    similarity_to_others,
    lack_of_specifics,
    superlative_ratio,
    ad_cliche_ratio,
    generic_word_ratio,
    rating_sentiment_mismatch,
    _tokenize,
    _bigrams,
)


# ── Хелперы: _tokenize, _bigrams ─────────────────────────────────────────────

class TestTokenize:
    def test_simple_russian(self):
        tokens = _tokenize("Отличный товар, всем рекомендую!")
        assert tokens == ["отличный", "товар", "всем", "рекомендую"]

    def test_mixed_languages(self):
        tokens = _tokenize("Размер XL подошёл")
        assert "размер" in tokens
        assert "xl" in tokens

    def test_empty_string(self):
        assert _tokenize("") == []

    def test_numbers_excluded(self):
        tokens = _tokenize("Купила 3 штуки за 500 рублей")
        assert "3" not in tokens
        assert "500" not in tokens
        assert "штуки" in tokens

    def test_punctuation_stripped(self):
        tokens = _tokenize("Супер!!! Класс...")
        assert tokens == ["супер", "класс"]


class TestBigrams:
    def test_normal(self):
        tokens = ["отличный", "товар", "рекомендую"]
        bg = _bigrams(tokens)
        assert bg == ["отличный товар", "товар рекомендую"]

    def test_single_token(self):
        assert _bigrams(["одно"]) == []

    def test_empty(self):
        assert _bigrams([]) == []


# ── 1. similarity_to_others ──────────────────────────────────────────────────

class TestSimilarityToOthers:
    def test_no_others(self):
        assert similarity_to_others("Отличный товар", []) == 0.0

    def test_identical_texts(self):
        text = "Отличный товар, всем рекомендую"
        others = [text]
        sim = similarity_to_others(text, others)
        assert sim == 1.0

    def test_completely_different(self):
        text = "Качество отличное материал хлопок"
        others = ["Ужасный брак вернула возврат"]
        sim = similarity_to_others(text, others)
        assert sim < 0.2

    def test_partially_similar(self):
        text = "Отличный товар рекомендую всем покупайте"
        others = ["Отличный товар советую всем берите"]
        sim = similarity_to_others(text, others)
        assert 0.1 < sim < 0.9

    def test_empty_text(self):
        assert similarity_to_others("", ["Текст"]) == 0.0

    def test_empty_other_text(self):
        assert similarity_to_others("Текст", [""]) == 0.0

    def test_returns_max_similarity(self):
        text = "Отличный товар рекомендую"
        others = [
            "Совершенно другой отзыв",
            "Отличный товар рекомендую",  # идентичный
        ]
        sim = similarity_to_others(text, others)
        assert sim == 1.0


# ── 2. lack_of_specifics ────────────────────────────────────────────────────

class TestLackOfSpecifics:
    def test_empty_text(self):
        assert lack_of_specifics("") == 1.0

    def test_all_specific(self):
        text = "размер хлопок цвет материал"
        result = lack_of_specifics(text)
        assert result < 0.1  # почти все слова конкретные

    def test_no_specifics(self):
        text = "отличный товар рекомендую советую"
        result = lack_of_specifics(text)
        assert result == 1.0

    def test_mixed(self):
        text = "Размер подошёл, хлопок натуральный, рекомендую всем"
        result = lack_of_specifics(text)
        assert 0.0 < result < 1.0

    def test_returns_float(self):
        result = lack_of_specifics("Обычный текст")
        assert isinstance(result, float)


# ── 3. superlative_ratio ─────────────────────────────────────────────────────

class TestSuperlativeRatio:
    def test_empty_text(self):
        assert superlative_ratio("") == 0.0

    def test_no_superlatives(self):
        text = "Купила платье, сидит нормально"
        result = superlative_ratio(text)
        assert result < 0.3

    def test_many_superlatives(self):
        text = "Лучший невероятный потрясающий великолепный!!!"
        result = superlative_ratio(text)
        assert result > 0.3

    def test_exclamation_marks_counted(self):
        text_calm = "Хороший товар"
        text_excited = "Хороший товар!!!!!!"
        r1 = superlative_ratio(text_calm)
        r2 = superlative_ratio(text_excited)
        assert r2 > r1

    def test_returns_bounded(self):
        result = superlative_ratio("Лучший" * 50 + "!" * 100)
        assert 0.0 <= result <= 1.0


# ── 4. ad_cliche_ratio ──────────────────────────────────────────────────────

class TestAdClicheRatio:
    def test_empty_text(self):
        assert ad_cliche_ratio("") == 0.0

    def test_no_cliches(self):
        text = "Платье из хлопка, размер подошёл, после стирки не село"
        result = ad_cliche_ratio(text)
        assert result < 0.2

    def test_many_cliches(self):
        text = "Рекомендую всем, берите не пожалеете, качество на высоте, товар огонь"
        result = ad_cliche_ratio(text)
        assert result > 0.3

    def test_single_cliche(self):
        text = "Хороший товар, рекомендую"
        result = ad_cliche_ratio(text)
        assert result > 0.0

    def test_bounded(self):
        text = " ".join(["рекомендую", "советую", "берите не пожалеете"] * 10)
        result = ad_cliche_ratio(text)
        assert 0.0 <= result <= 1.0


# ── 5. generic_word_ratio ───────────────────────────────────────────────────

class TestGenericWordRatio:
    def test_empty_text(self):
        assert generic_word_ratio("") == 1.0

    def test_all_generic(self):
        text = "хороший отличный классный понравился"
        result = generic_word_ratio(text)
        assert result == 1.0

    def test_no_generic(self):
        text = "хлопок размер стирка материал"
        result = generic_word_ratio(text)
        assert result == 0.0

    def test_mixed(self):
        text = "Хороший товар из хлопка, размер подошёл"
        result = generic_word_ratio(text)
        assert 0.0 < result < 1.0


# ── rating_sentiment_mismatch ────────────────────────────────────────────────

class TestRatingSentimentMismatch:
    def test_no_rating(self):
        assert rating_sentiment_mismatch("Отличный товар", None) == 0.0

    def test_invalid_rating(self):
        assert rating_sentiment_mismatch("Текст", "abc") == 0.0

    def test_empty_text(self):
        assert rating_sentiment_mismatch("", "5") == 0.0

    def test_positive_text_high_rating(self):
        """Позитивный текст + высокий рейтинг = мало несоответствия."""
        result = rating_sentiment_mismatch("Отличный, замечательный, рекомендую", "5")
        assert result < 0.3

    def test_negative_text_high_rating(self):
        """Негативный текст + высокий рейтинг = несоответствие."""
        result = rating_sentiment_mismatch("Ужасный брак, обман, вернула", "5")
        assert result > 0.3

    def test_positive_text_low_rating(self):
        """Позитивный текст + низкий рейтинг = несоответствие."""
        result = rating_sentiment_mismatch("Отличный, замечательный, рекомендую", "1")
        assert result > 0.3

    def test_bounded(self):
        result = rating_sentiment_mismatch("Ужасный кошмар брак", "5")
        assert 0.0 <= result <= 1.0

    def test_no_sentiment_words(self):
        result = rating_sentiment_mismatch("Размер подошёл хлопок", "3")
        assert result == 0.0
