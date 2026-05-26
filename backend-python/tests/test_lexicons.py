"""
Тесты для features/lexicons.py — словари признаков.

Запуск:
    cd backend-python
    python -m pytest tests/test_lexicons.py -v
"""

import pytest
from features.lexicons import (
    AD_CLICHES,
    SUPERLATIVES,
    GENERIC_WORDS,
    SPECIFIC_DETAIL_MARKERS,
    POSITIVE_WORDS,
    NEGATIVE_WORDS,
)


class TestLexiconTypes:
    """Все словари должны быть frozenset (неизменяемые)."""

    def test_ad_cliches_type(self):
        assert isinstance(AD_CLICHES, frozenset)

    def test_superlatives_type(self):
        assert isinstance(SUPERLATIVES, frozenset)

    def test_generic_words_type(self):
        assert isinstance(GENERIC_WORDS, frozenset)

    def test_specific_detail_markers_type(self):
        assert isinstance(SPECIFIC_DETAIL_MARKERS, frozenset)

    def test_positive_words_type(self):
        assert isinstance(POSITIVE_WORDS, frozenset)

    def test_negative_words_type(self):
        assert isinstance(NEGATIVE_WORDS, frozenset)


class TestLexiconNotEmpty:
    def test_ad_cliches_not_empty(self):
        assert len(AD_CLICHES) > 0

    def test_superlatives_not_empty(self):
        assert len(SUPERLATIVES) > 0

    def test_generic_words_not_empty(self):
        assert len(GENERIC_WORDS) > 0

    def test_specific_detail_markers_not_empty(self):
        assert len(SPECIFIC_DETAIL_MARKERS) > 0

    def test_positive_words_not_empty(self):
        assert len(POSITIVE_WORDS) > 0

    def test_negative_words_not_empty(self):
        assert len(NEGATIVE_WORDS) > 0


class TestLexiconContent:
    """Проверяем наличие ключевых слов в каждом словаре."""

    def test_ad_cliches_contains_key_phrases(self):
        assert "рекомендую" in AD_CLICHES
        assert "всем рекомендую" in AD_CLICHES
        assert "берите не пожалеете" in AD_CLICHES

    def test_superlatives_contains_key_words(self):
        assert "лучший" in SUPERLATIVES
        assert "невероятный" in SUPERLATIVES
        assert "очень" in SUPERLATIVES

    def test_generic_words_contains_key_words(self):
        assert "хороший" in GENERIC_WORDS
        assert "отличный" in GENERIC_WORDS
        assert "всё хорошо" in GENERIC_WORDS

    def test_specific_markers_contains_key_words(self):
        assert "размер" in SPECIFIC_DETAIL_MARKERS
        assert "хлопок" in SPECIFIC_DETAIL_MARKERS
        assert "цвет" in SPECIFIC_DETAIL_MARKERS

    def test_positive_words_contains_key_words(self):
        assert "хороший" in POSITIVE_WORDS
        assert "рекомендую" in POSITIVE_WORDS
        assert "супер" in POSITIVE_WORDS

    def test_negative_words_contains_key_words(self):
        assert "плохой" in NEGATIVE_WORDS
        assert "брак" in NEGATIVE_WORDS
        assert "ужасный" in NEGATIVE_WORDS


class TestLexiconLowercase:
    """Все слова в словарях должны быть в нижнем регистре."""

    @pytest.mark.parametrize("lexicon,name", [
        (AD_CLICHES, "AD_CLICHES"),
        (SUPERLATIVES, "SUPERLATIVES"),
        (GENERIC_WORDS, "GENERIC_WORDS"),
        (SPECIFIC_DETAIL_MARKERS, "SPECIFIC_DETAIL_MARKERS"),
        (POSITIVE_WORDS, "POSITIVE_WORDS"),
        (NEGATIVE_WORDS, "NEGATIVE_WORDS"),
    ])
    def test_all_lowercase(self, lexicon, name):
        for word in lexicon:
            assert word == word.lower(), \
                f"{name}: '{word}' is not lowercase"


class TestLexiconNoOverlap:
    """Положительные и отрицательные слова не должны пересекаться."""

    def test_positive_negative_no_overlap(self):
        overlap = POSITIVE_WORDS & NEGATIVE_WORDS
        assert len(overlap) == 0, f"Overlap: {overlap}"
