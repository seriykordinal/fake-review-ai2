from keras.models import Model
from keras.layers import (
    Input, Embedding, Conv1D, GlobalMaxPooling1D,
    Dense, Dropout, Concatenate, BatchNormalization,
)
from keras.optimizers import Adam
from keras.losses import CategoricalCrossentropy

from .tokenizer import VOCAB_SIZE, MAX_LENGTH

EMBEDDING_DIM = 128
CNN_FILTERS = 64
DROPOUT_RATE = 0.5
LEARNING_RATE = 0.001
LABEL_SMOOTHING = 0.2


def build_model(num_classes: int, num_features: int) -> Model:
    
    text_input = Input(shape=(MAX_LENGTH,), dtype="int32", name="text_input")

    x = Embedding(input_dim=VOCAB_SIZE, output_dim=EMBEDDING_DIM, input_length=MAX_LENGTH, name="embedding")(text_input)

    conv3 = Conv1D(CNN_FILTERS, kernel_size=3, activation="relu", padding="valid", name="conv_k3")(x)
    conv3 = GlobalMaxPooling1D(name="pool_k3")(conv3)

    conv5 = Conv1D(CNN_FILTERS, kernel_size=5, activation="relu", padding="valid", name="conv_k5")(x)
    conv5 = GlobalMaxPooling1D(name="pool_k5")(conv5)

    text_merged = Concatenate(name="conv_concat")([conv3, conv5])
    text_out = Dense(64, activation="relu", name="text_dense")(text_merged)



    features_input = Input(shape=(num_features,), dtype="float32", name="features_input")

    f = BatchNormalization(name="feat_bn")(features_input)
    f = Dense(16, activation="relu", name="feat_dense")(f)



    merged = Concatenate(name="merge")([text_out, f])

    h = Dense(64, activation="relu", name="hidden")(merged)
    h = Dropout(DROPOUT_RATE, name="dropout")(h)

    output = Dense(num_classes, activation="softmax", name="output")(h)

    model = Model(
        inputs=[text_input, features_input],
        outputs=output,
        name="FakeReviewDetector"
    )

    model.compile(
        optimizer=Adam(learning_rate=LEARNING_RATE),
        loss=CategoricalCrossentropy(label_smoothing=LABEL_SMOOTHING),
        metrics=["accuracy"]
    )

    return model
