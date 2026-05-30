import os
import sys
import pandas as pd

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, BASE_DIR)

from features.feature_pipeline import extract_features_batch, FEATURE_NAMES
from features.extractors import rating_sentiment_mismatch

INPUT_CSV  = os.path.join(BASE_DIR, "datasets", "raw_reviews.csv")
OUTPUT_CSV = os.path.join(BASE_DIR, "datasets", "labeled_reviews.csv")


RED_FLAGS = {
    "similarity_to_others": (">", 0.50),  #схожесть с другим 
    "superlative_ratio":    (">", 0.12),  #восторженный стиль
    "ad_cliche_ratio":      (">", 0.10),  #рекламные клише
    "generic_word_ratio":   (">", 0.15),  #много общих слов
    "rating_mismatch":      (">", 0.40),  #рейтинг не тональность
    "too_short":            ("<", 5),     #<5 слов
}

GREEN_FLAGS = {
    "similarity_to_others": ("<", 0.25),  #мало похоже на других
    "superlative_ratio":    ("<", 0.05),  #сдержанный стиль
    "ad_cliche_ratio":      ("<", 0.05),  #нет рекламных клише
    "generic_word_ratio":   ("<", 0.10),  #мало общих слов
    "has_date":             ("=", True),  #есть дата
    "long_text":            (">", 10),    #> 10 слов
}


def word_count(text: str) -> int:
    return len(text.split()) if isinstance(text, str) else 0


def _check(val, op, threshold) -> bool:
    if op == ">":  return val > threshold
    if op == "<":  return val < threshold
    if op == "=":  return val == threshold
    return False


def count_red_flags(row: dict, feats: dict) -> int:
    n = 0
    for fname, (op, thr) in RED_FLAGS.items():
        if fname == "rating_mismatch":
            n += _check(feats.get("_mismatch", 0), op, thr)
        elif fname == "too_short":
            n += _check(feats.get("_word_count", 0), op, thr)
        else:
            n += _check(feats.get(fname, 0), op, thr)
    return n


def count_green_flags(row: dict, feats: dict) -> int:
    n = 0
    for fname, (op, thr) in GREEN_FLAGS.items():
        if fname == "has_date":
            n += _check(bool(row.get("date", "")), op, thr)
        elif fname == "long_text":
            n += _check(feats.get("_word_count", 0), op, thr)
        else:
            n += _check(feats.get(fname, 0), op, thr)
    return n


def assign_label(red: int, green: int) -> int:
    if red >= 2:
        return 1
    if red == 1 and green == 0:
        return 1
    if green >= 2 and red <= 1:
        return 0
    return 2  


def build_full_text(row) -> str:

    parts = []
    pros = str(row.get("pros", "") or "").strip()
    cons = str(row.get("cons", "") or "").strip()
    text = str(row.get("text", "") or "").strip()

    if pros:
        parts.append(f"Достоинства: {pros}")
    if cons:
        parts.append(f"Недостатки: {cons}")
    if text:
        parts.append(text)
    return " ".join(parts)


def main():
    if not os.path.exists(INPUT_CSV):
        print(f"Не найден файл: {INPUT_CSV}")
        print("Запустите сначала: python -m data.collect_reviews")
        sys.exit(1)

    print(f"Читаем: {INPUT_CSV}")
    df = pd.read_csv(INPUT_CSV, dtype=str)  
    print(f"Загружено отзывов: {len(df)}")

    expected = {"product_id", "rating", "pros", "cons", "text", "date"}
    missing = expected - set(df.columns)
    if missing:
        print(f"В CSV отсутствуют колонки: {missing}")
        sys.exit(1)

    df = df.fillna("")

    df["full_text"] = df.apply(build_full_text, axis=1)

    df["text_only_rating"] = (df["full_text"].str.strip() == "").astype(int)
    no_text_count = df["text_only_rating"].sum()
    print(f"  Отзывов без текста (только рейтинг): {no_text_count}")


    df.loc[df["full_text"].str.strip() == "", "full_text"] = "[пусто]"

    print("\nВычисление признаков...")
    all_feats: list[dict] = []

    for pid, group in df.groupby("product_id"):
        texts   = group["full_text"].tolist()
        ratings = group["rating"].tolist()

        matrix = extract_features_batch(texts, ratings)

        for i, idx in enumerate(group.index):
            row = group.loc[idx]
            feat_dict = dict(zip(FEATURE_NAMES, matrix[i]))
            feat_dict["_mismatch"]   = rating_sentiment_mismatch(texts[i], ratings[i])
            feat_dict["_word_count"] = word_count(texts[i])
            feat_dict["__index__"]   = idx
            all_feats.append(feat_dict)

    feat_df = pd.DataFrame(all_feats).set_index("__index__")
    feat_df.index.name = None

    aux_cols = ["_mismatch", "_word_count"]
    feat_df_clean = feat_df.drop(columns=aux_cols)

    result = df.join(feat_df_clean)

    print("Разметка...")

    labels, reds, greens = [], [], []
    for idx in result.index:
        row  = result.loc[idx].to_dict()
        frow = feat_df.loc[idx].to_dict()
        red   = count_red_flags(row, frow)
        green = count_green_flags(row, frow)
        labels.append(assign_label(red, green))
        reds.append(red)
        greens.append(green)

    result["label"] = labels
    result["red_flags"] = reds
    result["green_flags"] = greens

    result = result.drop(columns=["full_text"], errors="ignore")

    label_names = {0: "genuine", 1: "fake", 2: "uncertain"}
    
    print("\nРаспределение меток:")
    for label_val, cnt in result["label"].value_counts().sort_index().items():
        pct = 100 * cnt / len(result)
        print(f"  {label_val} ({label_names[label_val]:9s}): {cnt:5d}  ({pct:.1f}%)")

    base_cols    = ["product_id", "rating", "pros", "cons", "text", "date", "text_only_rating"]
    feature_cols = FEATURE_NAMES
    meta_cols    = ["red_flags", "green_flags", "label"]
    final_cols   = [c for c in base_cols + feature_cols + meta_cols if c in result.columns]

    result[final_cols].to_csv(OUTPUT_CSV, index=False, encoding="utf-8")
    print(f"\nСохранено: {OUTPUT_CSV}")
    print(f"Колонки: {final_cols}")


if __name__ == "__main__":
    main()
