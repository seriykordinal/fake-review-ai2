import io
import os
import sys
import requests
import pandas as pd

DATASET_URL = "https://huggingface.co/datasets/deepRost/wb-reviews/resolve/main/wb-review-dataset.csv"
OUTPUT_PATH = os.path.join(os.path.dirname(__file__), "wb_reviews.csv")

ENCODINGS = ["utf-8-sig", "utf-8", "windows-1251", "latin1"]


def download_dataset(url: str) -> bytes:
    print(f"Скачиваем датасет: {url}")
    response = requests.get(url, timeout=60)
    response.raise_for_status()
    print(f"Скачано {len(response.content) / 1024:.1f} KB")
    return response.content


def parse_csv(content: bytes) -> pd.DataFrame:
    for enc in ENCODINGS:
        try:
            text = content.decode(enc)
            df = pd.read_csv(io.StringIO(text), sep=None, engine="python")
            sample = str(df.iloc[0, 0]) if len(df) > 0 else ""
            if "?" not in sample and any(ord(c) > 127 for c in sample[:100]):
                print(f"Кодировка: {enc}")
                return df
        except Exception:
            continue
    raise ValueError("Не удалось определить кодировку датасета")


def normalize_columns(df: pd.DataFrame) -> pd.DataFrame:
    col_map = {
        "text": "review_text",
        "review": "review_text",
        "comment": "review_text",
        "label": "sentiment",
        "target": "sentiment",
        "class": "sentiment",
    }
    df = df.rename(columns={c: col_map[c] for c in df.columns if c in col_map})

    if "review_text" not in df.columns and len(df.columns) == 2:
        df.columns = ["review_text", "sentiment"]

    required = {"review_text", "sentiment"}
    missing = required - set(df.columns)
    if missing:
        raise ValueError(f"Отсутствуют колонки: {missing}. Есть: {df.columns.tolist()}")

    return df[["review_text", "sentiment"]]


def main():
    content = download_dataset(DATASET_URL)
    df = parse_csv(content)
    df = normalize_columns(df)

    # Базовая очистка
    before = len(df)
    df = df.dropna(subset=["review_text", "sentiment"])
    df["review_text"] = df["review_text"].astype(str).str.strip()
    df = df[df["review_text"].str.len() > 3]
    print(f"Строк после очистки: {len(df)} (удалено {before - len(df)})")

    print(f"\nРаспределение меток:\n{df['sentiment'].value_counts().to_string()}")

    df.to_csv(OUTPUT_PATH, index=False, encoding="utf-8")
    print(f"\nСохранено: {OUTPUT_PATH}")
    print(f"Размер: {len(df)} строк, {len(df.columns)} колонок")


if __name__ == "__main__":
    main()
