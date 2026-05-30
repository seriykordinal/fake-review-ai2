import os
import pickle
from typing import List, Tuple

import numpy as np
from tensorflow.keras.preprocessing.text import Tokenizer
from keras.preprocessing.sequence import pad_sequences


VOCAB_SIZE = 10000   
MAX_LENGTH = 100     
OOV_TOKEN = "<OOV>" 
PAD_TYPE = "post"  
TRUNC_TYPE = "post"  


def fit_tokenizer(texts: List[str]) -> Tokenizer:
    tokenizer = Tokenizer(num_words=VOCAB_SIZE, oov_token=OOV_TOKEN)
    tokenizer.fit_on_texts(texts)
    return tokenizer


def texts_to_padded(tokenizer: Tokenizer, texts: List[str]) -> np.ndarray:
    sequences = tokenizer.texts_to_sequences(texts)
    padded = pad_sequences(
        sequences,
        maxlen=MAX_LENGTH,
        padding=PAD_TYPE,
        truncating=TRUNC_TYPE,
    )
    return padded


def save_tokenizer(tokenizer: Tokenizer, path: str) -> None:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "wb") as f:
        pickle.dump(tokenizer, f, protocol=pickle.HIGHEST_PROTOCOL)


def load_tokenizer(path: str) -> Tokenizer:
    with open(path, "rb") as f:
        return pickle.load(f)


def prepare_training_data(
    texts: List[str],
) -> Tuple[Tokenizer, np.ndarray]:
    tokenizer = fit_tokenizer(texts)
    padded = texts_to_padded(tokenizer, texts)
    return tokenizer, padded
