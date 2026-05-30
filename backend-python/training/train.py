import os
import sys
import pickle
import numpy as np
import pandas as pd
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt

from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import (
    classification_report, confusion_matrix, ConfusionMatrixDisplay,
)
from keras.callbacks import EarlyStopping, ModelCheckpoint
from keras.utils import to_categorical

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, BASE_DIR)

from features.feature_pipeline import FEATURE_NAMES, extract_features_batch  # noqa: F401
from training.tokenizer import (
    fit_tokenizer, texts_to_padded, save_tokenizer,
    VOCAB_SIZE, MAX_LENGTH,
)
from training.model import build_model


DATASET_PATH = os.path.join(BASE_DIR, "datasets", "labeled_reviews.csv")
ARTIFACTS_DIR = os.path.join(BASE_DIR, "artifacts")
MODEL_PATH = os.path.join(ARTIFACTS_DIR, "model.keras")
TOKENIZER_PATH = os.path.join(ARTIFACTS_DIR, "tokenizer.pkl")
SCALER_PATH = os.path.join(ARTIFACTS_DIR, "feature_scaler.pkl")
HISTORY_PNG = os.path.join(ARTIFACTS_DIR, "training_history.png")
CONFUSION_PNG = os.path.join(ARTIFACTS_DIR, "confusion_matrix.png")
REPORT_TXT = os.path.join(ARTIFACTS_DIR, "classification_report.txt")

os.makedirs(ARTIFACTS_DIR, exist_ok=True)

BATCH_SIZE = 32
EPOCHS = 50    
PATIENCE = 5       
VALIDATION_SPLIT = 0.15
TEST_SIZE = 0.15
RANDOM_STATE = 42


def load_and_prepare_data():

    if not os.path.exists(DATASET_PATH):
        print(f"Не найден размеченный датасет: {DATASET_PATH}")
        print("Запустите сначала: python -m data.label_dataset")
        sys.exit(1)

    df = pd.read_csv(DATASET_PATH, dtype=str)
    df = df.fillna("")
    print(f"Загружено отзывов: {len(df)}")
    print(f"Распределение меток до фильтрации:")
    print(df["label"].value_counts().sort_index().to_string())

    df["label"] = df["label"].astype(int)
    df = df[df["label"].isin([0, 1])].reset_index(drop=True)
    print(f"\nПосле исключения uncertain: {len(df)} отзывов")
    print(df["label"].value_counts().sort_index().to_string())


    def build_text(row):
        parts = []
        pros = str(row.get("pros") or "").strip()
        cons = str(row.get("cons") or "").strip()
        text = str(row.get("text") or "").strip()
        if pros: parts.append(f"Достоинства: {pros}")
        if cons: parts.append(f"Недостатки: {cons}")
        if text: parts.append(text)
        return " ".join(parts) if parts else "[пусто]"

    df["full_text"] = df.apply(build_text, axis=1)
    texts  = df["full_text"].astype(str).tolist()
    labels = df["label"].astype(int).astype(np.int32).values

    if all(name in df.columns for name in FEATURE_NAMES):
        print("Признаки взяты из CSV")
        features = df[FEATURE_NAMES].values.astype(np.float32)
    else:
        print("Признаки в CSV отсутствуют — вычисляем заново")
        ratings  = df["rating"].astype(str).tolist() if "rating" in df.columns else None
        features = extract_features_batch(texts, ratings)

    num_classes = len(np.unique(labels))
    return texts, features, labels, num_classes


def plot_history(history, path: str):
    fig, axes = plt.subplots(1, 2, figsize=(12, 4))

    axes[0].plot(history.history["loss"], label="train")
    axes[0].plot(history.history["val_loss"], label="val")
    axes[0].set_title("Loss")
    axes[0].set_xlabel("Эпоха")
    axes[0].legend()

    axes[1].plot(history.history["accuracy"], label="train")
    axes[1].plot(history.history["val_accuracy"], label="val")
    axes[1].set_title("Accuracy")
    axes[1].set_xlabel("Эпоха")
    axes[1].legend()

    fig.tight_layout()
    fig.savefig(path, dpi=100)
    plt.close(fig)


def plot_confusion(y_true, y_pred, path: str):
    cm = confusion_matrix(y_true, y_pred)
    fig, ax = plt.subplots(figsize=(5, 4))
    ConfusionMatrixDisplay(cm, display_labels=["genuine", "fake"]).plot(
        ax=ax, cmap="Blues", values_format="d",
    )
    ax.set_title("Матрица ошибок")
    fig.tight_layout()
    fig.savefig(path, dpi=100)
    plt.close(fig)


def main():

    texts, features, labels, num_classes = load_and_prepare_data()
    print(f"\nКлассов для обучения: {num_classes}")

    X_train_texts, X_test_texts, X_train_feats, X_test_feats, y_train, y_test = train_test_split(
        texts, features, labels,
        test_size=TEST_SIZE,
        random_state=RANDOM_STATE,
        stratify=labels,
    )
    print(f"\nTrain: {len(X_train_texts)}, Test: {len(X_test_texts)}")

    y_train_cat = to_categorical(y_train, num_classes=num_classes)
    y_test_cat  = to_categorical(y_test,  num_classes=num_classes)

    print("\nОбучение токенизатора...")
    tokenizer = fit_tokenizer(X_train_texts)
    X_train_seq = texts_to_padded(tokenizer, X_train_texts)
    X_test_seq  = texts_to_padded(tokenizer, X_test_texts)
    save_tokenizer(tokenizer, TOKENIZER_PATH)
    print(f"Размер словаря: {min(VOCAB_SIZE, len(tokenizer.word_index) + 1)}")
    print(f"Сохранён: {TOKENIZER_PATH}")

    print("\nНормализация числовых признаков...")
    scaler = StandardScaler()
    X_train_feats_norm = scaler.fit_transform(X_train_feats).astype(np.float32)
    X_test_feats_norm  = scaler.transform(X_test_feats).astype(np.float32)
    with open(SCALER_PATH, "wb") as f:
        pickle.dump(scaler, f)
    print(f"Сохранён: {SCALER_PATH}")

    print(f"\nПостроение модели (классов: {num_classes}, признаков: {features.shape[1]})...")
    model = build_model(num_classes=num_classes, num_features=features.shape[1])
    model.summary()

    print("\nОбучение...")
    callbacks = [
        EarlyStopping(
            monitor="val_loss",
            patience=PATIENCE,
            restore_best_weights=True, 
            verbose=1,
        ),
        ModelCheckpoint(
            filepath=MODEL_PATH,
            monitor="val_loss",
            save_best_only=True,
            verbose=0,
        ),
    ]
    history = model.fit(
        x=[X_train_seq, X_train_feats_norm],
        y=y_train_cat,
        validation_split=VALIDATION_SPLIT,
        batch_size=BATCH_SIZE,
        epochs=EPOCHS,
        callbacks=callbacks,
        verbose=1,
    )

    print("\nОценка на тестовой выборке...")
    loss, accuracy = model.evaluate(
        [X_test_seq, X_test_feats_norm], y_test_cat, verbose=0,
    )
    print(f"Test loss: {loss:.4f}\nTest accuracy: {accuracy:.4f}")

    predictions = model.predict([X_test_seq, X_test_feats_norm], verbose=0)
    y_pred = np.argmax(predictions, axis=1)

    label_names = ["genuine", "fake"][:num_classes]
    report = classification_report(y_test, y_pred, target_names=label_names)
    print("\n--- Метрики ---")
    print(report)

    with open(REPORT_TXT, "w", encoding="utf-8") as f:
        f.write(f"Test loss: {loss:.4f}\nTest accuracy: {accuracy:.4f}\n\n")
        f.write(report)
    print(f"Отчёт сохранён: {REPORT_TXT}")

    plot_history(history, HISTORY_PNG)
    plot_confusion(y_test, y_pred, CONFUSION_PNG)
    print(f"Графики сохранены: {HISTORY_PNG}, {CONFUSION_PNG}")

    model.save(MODEL_PATH)
    print(f"\nМодель сохранена: {MODEL_PATH}")


if __name__ == "__main__":
    main()
