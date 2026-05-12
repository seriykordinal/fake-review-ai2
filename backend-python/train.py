import os
import pickle
import numpy as np
import pandas as pd
import matplotlib
matplotlib.use("Agg")  
import matplotlib.pyplot as plt

from sklearn.model_selection import train_test_split
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics import confusion_matrix, classification_report
from sklearn.preprocessing import LabelEncoder

from keras.models import Sequential
from keras.layers import Dense
from keras.optimizers import Adam

BASE_DIR     = os.path.dirname(__file__)
DATASET_PATH = os.path.join(BASE_DIR, "dataset", "wb_reviews.csv")
MODELS_DIR   = os.path.join(BASE_DIR, "models")
MODEL_PATH   = os.path.join(MODELS_DIR, "fake_review_model.h5")
TFIDF_PATH   = os.path.join(MODELS_DIR, "tfidf_vectorizer.pkl")
CM_PATH      = os.path.join(MODELS_DIR, "confusion_matrix.png")

os.makedirs(MODELS_DIR, exist_ok=True)

print("Загружаем датасет...")
if not os.path.exists(DATASET_PATH):
    raise FileNotFoundError(
        f"Датасет не найден: {DATASET_PATH}\n"
        "Запустите сначала: python dataset/get_dataset.py"
    )

df = pd.read_csv(DATASET_PATH)
print(f"Загружено строк: {len(df)}")
print(f"Распределение меток:\n{df['sentiment'].value_counts().sort_index().to_string()}\n")

X_text = df["review_text"].astype(str).values
y_raw  = df["sentiment"].values

le = LabelEncoder()
y = le.fit_transform(y_raw).astype(np.int32)
classes     = le.classes_
num_classes = len(classes)
print(f"Классы: {classes}  (всего: {num_classes})")

print("Векторизация TF-IDF...")
tfidf = TfidfVectorizer(
    max_features=5000,
    ngram_range=(1, 2),
    sublinear_tf=True,
    strip_accents="unicode",
    analyzer="word",
    min_df=2,
)
X = tfidf.fit_transform(X_text).toarray().astype(np.float32)
print(f"Размерность признаков: {X.shape}")

with open(TFIDF_PATH, "wb") as f:
    pickle.dump(tfidf, f)
print(f"TF-IDF сохранён: {TFIDF_PATH}")

X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42, stratify=y
)
print(f"Train: {len(X_train)}, Test: {len(X_test)}\n")


model = Sequential([
    Dense(units=32, activation="relu", input_shape=(X_train.shape[1],)),
    Dense(units=16, activation="relu"),
    Dense(units=num_classes, activation="softmax"),
])

model.summary()

model.compile(
    optimizer=Adam(learning_rate=0.0001),
    loss="sparse_categorical_crossentropy",
    metrics=["accuracy"],
)

history = model.fit(
    x=X_train,
    y=y_train,
    validation_split=0.1,
    batch_size=64,
    epochs=10,
    verbose=1,
)

loss, accuracy = model.evaluate(X_test, y_test, verbose=0)
print(f"\nТочность на тесте: {accuracy:.4f}  |  Loss: {loss:.4f}")

predictions = model.predict(X_test, verbose=0)
y_pred = np.argmax(predictions, axis=1)

print("\n--- Метрики ---")
label_names = [str(c) for c in classes]
print(classification_report(y_test, y_pred, target_names=label_names))

cm = confusion_matrix(y_test, y_pred)

fig, ax = plt.subplots(figsize=(5, 4))
im = ax.imshow(cm, interpolation="nearest", cmap=plt.cm.Blues)
plt.colorbar(im, ax=ax)

ax.set(
    xticks=range(len(label_names)), yticks=range(len(label_names)),
    xticklabels=label_names, yticklabels=label_names,
    xlabel="Предсказанная метка", ylabel="Истинная метка",
    title="Матрица ошибок",
)

thresh = cm.max() / 2.0
for i in range(cm.shape[0]):
    for j in range(cm.shape[1]):
        ax.text(j, i, cm[i, j], ha="center", va="center",
                color="white" if cm[i, j] > thresh else "black")

fig.tight_layout()
fig.savefig(CM_PATH, dpi=100)
print(f"\nМатрица ошибок сохранена: {CM_PATH}")
plt.close(fig)

model.save(MODEL_PATH)
print(f"Модель сохранена: {MODEL_PATH}")





