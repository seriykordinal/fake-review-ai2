"""
Тесты для features/feature_pipeline.py — единая точка извлечения признаков.

Запуск:
    cd backend-python
    python -m pytest tests/test_feature_pipeline.py -v
"""

import numpy as np
import pytest
from features.feature_pipeline import (
    extract_features,
    extract_features_batch,
    FEATURE_NAMES,
    NUM_FEATURES,
)


class TestConstants:
    def test_num_features(self):
        assert NUM_FEATURES == 5

    def test_feature_names_count(self):
        assert len(FEATURE_NAMES) == 5

    def test_feature_names_content(self):
        expected = [
            "similarity_to_others",
            "lack_of_specifics",
            "superlative_ratio",
            "ad_cliche_ratio",
            "generic_word_ratio",
        ]
        assert FEATURE_NAMES == expected


class TestExtractFeatures:
    def test_returns_numpy_array(self):
        result = extract_features("Отличный товар")
        assert isinstance(result, np.ndarray)

    def test_shape(self):
        result = extract_features("Отличный товар")
        assert result.shape == (5,)

    def test_dtype(self):
        result = extract_features("Текст")
        assert result.dtype == np.float32

    def test_all_values_bounded(self):
        result = extract_features("Рекомендую всем, берите не пожалеете!")
        for i, val in enumerate(result):
            assert 0.0 <= val <= 1.0, f"feature {FEATURE_NAMES[i]} = {val}"

    def test_empty_text(self):
        result = extract_features("")
        assert result.shape == (5,)

    def test_with_rating(self):
        result = extract_features("Хороший товар", rating="5")
        assert result.shape == (5,)

    def test_with_other_texts(self):
        result = extract_features(
            "Отличный товар",
            other_texts=["Такой же отличный товар", "Другой отзыв"],
        )
        assert result.shape == (5,)
        # similarity_to_others (индекс 0) должен быть > 0
        assert result[0] > 0.0

    def test_no_other_texts(self):
        result = extract_features("Текст", other_texts=None)
        assert result[0] == 0.0  # similarity = 0 когда нет других отзывов

    def test_suspicious_review_has_high_features(self):
        """Типичный фейковый отзыв должен давать высокие значения признаков."""
        fake = "Лучший товар! Рекомендую всем! Качество на высоте! Берите не пожалеете!!!"
        result = extract_features(fake)
        # Должны быть повышены: superlative, ad_cliche, generic
        assert result[2] > 0.1  # superlative_ratio
        assert result[3] > 0.1  # ad_cliche_ratio

    def test_genuine_review_has_specifics(self):
        """Подлинный отзыв с конкретикой — низкий lack_of_specifics."""
        genuine = "Размер подошёл, хлопок натуральный, после стирки не село, ношу каждый день"
        result = extract_features(genuine)
        assert result[1] < 0.8  # lack_of_specifics ниже


class TestExtractFeaturesBatch:
    def test_returns_numpy_array(self):
        texts = ["Текст 1", "Текст 2", "Текст 3"]
        result = extract_features_batch(texts)
        assert isinstance(result, np.ndarray)

    def test_shape(self):
        texts = ["Текст 1", "Текст 2"]
        result = extract_features_batch(texts)
        assert result.shape == (2, 5)

    def test_dtype(self):
        result = extract_features_batch(["Текст"])
        assert result.dtype == np.float32

    def test_empty_list(self):
        result = extract_features_batch([])
        assert result.shape == (0, 5)

    def test_single_text(self):
        result = extract_features_batch(["Один отзыв"])
        assert result.shape == (1, 5)

    def test_similarity_computed_within_batch(self):
        """Одинаковые тексты в батче дают высокий similarity."""
        texts = [
            "Отличный товар рекомендую всем покупайте",
            "Отличный товар рекомендую всем покупайте",
            "Совсем другой уникальный отзыв про хлопок",
        ]
        result = extract_features_batch(texts)
        # Первые два одинаковые — similarity высокий
        assert result[0, 0] > 0.5
        assert result[1, 0] > 0.5
        # Третий отличается — similarity ниже
        assert result[2, 0] < result[0, 0]

    def test_with_ratings(self):
        texts = ["Отзыв 1", "Отзыв 2"]
        ratings = ["5", "3"]
        result = extract_features_batch(texts, ratings)
        assert result.shape == (2, 5)

    def test_with_none_ratings(self):
        result = extract_features_batch(["Текст"], ratings=None)
        assert result.shape == (1, 5)

    def test_values_bounded(self):
        texts = [
            "Рекомендую!!!",
            "Ужасный товар, брак",
            "Нормально, размер подошёл",
        ]
        result = extract_features_batch(texts)
        assert np.all(result >= 0.0)
        assert np.all(result <= 1.0)
