import csv
import os
import re
import sys
import time
import requests
from typing import Optional, List

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
PRODUCTS_FILE = os.path.join(BASE_DIR, "data", "products_list.txt")
OUTPUT_CSV    = os.path.join(BASE_DIR, "datasets", "raw_reviews.csv")

os.makedirs(os.path.dirname(OUTPUT_CSV), exist_ok=True)

GO_SERVER_URL = "http://localhost:8080"

FIELDNAMES = ["product_id", "rating", "pros", "cons", "text", "date"]


def extract_product_id(line: str) -> Optional[int]:
    line = line.strip()
    if not line or line.startswith("#"):
        return None
    m = re.search(r"/catalog/(\d+)/", line)
    if m:
        return int(m.group(1))
    m = re.search(r"(\d{5,})", line)
    if m:
        return int(m.group(1))
    return None


def get_token() -> str:
    print("Авторизация в Go-сервере...")
    email = input("Email: ").strip()
    password = input("Password: ").strip()
    try:
        r = requests.post(
            f"{GO_SERVER_URL}/api/login",
            json={"email": email, "password": password},
            timeout=10,
        )
        if r.status_code != 200:
            print(f"Не удалось войти: {r.text}")
            sys.exit(1)
        token = r.json().get("token", "")
        print("Авторизация успешна.\n")
        return token
    
    except Exception as e:
        print(f"Go-сервер недоступен ({GO_SERVER_URL}): {e}")
        sys.exit(1)


def fetch_reviews(product_id: int, token: str) -> List[dict]:
    url = f"{GO_SERVER_URL}/api/analyze_product?parse_only=true"
    product_url = f"https://www.wildberries.ru/catalog/{product_id}/detail.aspx"
    
    try:
        r = requests.post(
            url,
            json={"product_url": product_url},
            headers={"Authorization": f"Bearer {token}"},
            timeout=600,  
        )
        if r.status_code != 200:
            print(f"Go-сервер вернул {r.status_code}: {r.text[:100]}")
            return []
        data = r.json()
        return data.get("results", [])
    except Exception as e:
        print(f"Ошибка вызова Go-сервера: {e}")
        return []


def parse_review(review: dict, product_id: int) -> dict:
    return {
        "product_id":   product_id,
        "rating":       review.get("rating", ""),
        "pros":         review.get("pros", "").strip(),
        "cons":         review.get("cons", "").strip(),
        "text":         review.get("text", "").strip(),
        "date":         review.get("date", ""),
    }


def collect_all():
    if not os.path.exists(PRODUCTS_FILE):
        print(f"Файл не найден: {PRODUCTS_FILE}")
        print("Создайте его и добавьте URL или ID товаров (по одному на строку).")
        sys.exit(1)

    with open(PRODUCTS_FILE, "r", encoding="utf-8") as f:
        product_ids = [pid for pid in (extract_product_id(line) for line in f) if pid]

    if not product_ids:
        print("Список товаров пуст.")
        sys.exit(1)

    print(f"Товаров для сбора: {len(product_ids)}")

    token = get_token()

    total_reviews = 0
    skipped = 0

    with open(OUTPUT_CSV, "w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=FIELDNAMES)
        writer.writeheader()

        for i, pid in enumerate(product_ids, 1):
            print(f"[{i}/{len(product_ids)}] Товар {pid}", end=" ")

            reviews = fetch_reviews(pid, token)

            if not reviews:
                print("— отзывов не найдено, пропуск")
                skipped += 1
                continue

            count = 0
            for rev in reviews:
                row = parse_review(rev, pid)
                if not (row["text"] or row["pros"] or row["cons"]):
                    continue
                writer.writerow(row)
                count += 1

            total_reviews += count
            print(f"— {count} отзывов")

            time.sleep(2.0)

    print(f"\n{'='*50}")
    print(f"Готово!")
    print(f"Товаров обработано: {len(product_ids) - skipped} из {len(product_ids)}")
    print(f"Всего отзывов: {total_reviews}")
    print(f"Сохранено: {OUTPUT_CSV}")
    print(f"{'='*50}")


if __name__ == "__main__":
    collect_all()
