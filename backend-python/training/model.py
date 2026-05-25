"""
Архитектура нейронной сети для детектирования фейковых отзывов.

Текстовая часть: Embedding → Conv1D × 2 → GlobalMaxPooling → Dense(64)
Признаки:        5 числовых → BatchNorm → Dense(16)
Объединение:     Concatenate → Dense(64) + Dropout → softmax

Используем Functional API для двух входов, но архитектура каждой
ветки описана в Sequential-стиле для читаемости.
"""

from tensorflow.keras.models import Model
from tensorflow.keras.layers import (
    Input, Embedding, Conv1D, GlobalMaxPooling1D,
    Dense, Dropout, Concatenate, BatchNormalization,
)
from tensorflow.keras.optimizers import Adam
from tensorflow.keras.losses import CategoricalCrossentropy

from .tokenizer import VOCAB_SIZE, MAX_LENGTH

EMBEDDING_DIM   = 128
CNN_FILTERS     = 64
DROPOUT_RATE    = 0.5       # увеличен для снижения переобучения
LEARNING_RATE   = 1e-3
LABEL_SMOOTHING = 0.2       # смягчает метки: 0→0.1, 1→0.9


def build_model(num_classes: int, num_features: int) -> Model:
    """
    Строит модель с двумя входами: текст (последовательность токенов)
    и числовые признаки (5 значений от 0 до 1).

    Параметры:
        num_classes  — 2 (genuine / fake) или 3 (+ uncertain)
        num_features — 5 (из feature_pipeline.NUM_FEATURES)
    """

    # ── Текстовая ветка ───────────────────────────────────────────────────
    text_input = Input(shape=(MAX_LENGTH,), dtype="int32", name="text_input")

    x = Embedding(
        input_dim=VOCAB_SIZE,
        output_dim=EMBEDDING_DIM,
        input_length=MAX_LENGTH,
        name="embedding",
    )(text_input)

    # Два параллельных свёрточных слоя — триграммы и пятиграммы
    conv3 = Conv1D(CNN_FILTERS, kernel_size=3, activation="relu",
                   padding="valid", name="conv_k3")(x)
    conv3 = GlobalMaxPooling1D(name="pool_k3")(conv3)

    conv5 = Conv1D(CNN_FILTERS, kernel_size=5, activation="relu",
                   padding="valid", name="conv_k5")(x)
    conv5 = GlobalMaxPooling1D(name="pool_k5")(conv5)

    text_merged = Concatenate(name="conv_concat")([conv3, conv5])
    text_out = Dense(64, activation="relu", name="text_dense")(text_merged)

    # ── Ветка числовых признаков ──────────────────────────────────────────
    features_input = Input(shape=(num_features,), dtype="float32",
                           name="features_input")

    f = BatchNormalization(name="feat_bn")(features_input)
    f = Dense(16, activation="relu", name="feat_dense")(f)

    # ── Объединение и классификатор ───────────────────────────────────────
    merged = Concatenate(name="merge")([text_out, f])

    h = Dense(64, activation="relu", name="hidden")(merged)
    h = Dropout(DROPOUT_RATE, name="dropout")(h)

    output = Dense(num_classes, activation="softmax", name="output")(h)

    # ── Сборка ────────────────────────────────────────────────────────────
    model = Model(
        inputs=[text_input, features_input],
        outputs=output,
        name="FakeReviewDetector",
    )

    # Label smoothing: модель не учится выдавать 0.0 и 1.0,
    # а стремится к 0.05 и 0.95 — это даёт реальные вероятности
    model.compile(
        optimizer=Adam(learning_rate=LEARNING_RATE),
        loss=CategoricalCrossentropy(label_smoothing=LABEL_SMOOTHING),
        metrics=["accuracy"],
    )

    return model
