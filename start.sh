#!/usr/bin/env bash

# =============================================================================

# start.sh — Запуск проекта FakeCheck (Linux / macOS)

# Запуск: chmod +x start.sh && ./start.sh

# =============================================================================

set -e
ROOT="$(cd "$(dirname "$0")" && pwd)"
PIDS=()

step() { echo -e "\n\033[36m>>> $1\033[0m"; }
ok()   { echo -e "    \033[32m[OK]\033[0m $1"; }
warn() { echo -e "    \033[33m[WARN]\033[0m $1"; }
fail() { echo -e "    \033[31m[ERROR]\033[0m $1"; exit 1; }

# Завершаем дочерние процессы при выходе

cleanup() {
echo -e "\nОстановка серверов…"
for pid in "${PIDS[@]}"; do
kill "$pid" 2>/dev/null && echo "  Остановлен PID $pid"
done
}
trap cleanup EXIT INT TERM

# ── 1. Компиляция TypeScript ──────────────────────────────────────────────────

step "Компиляция TypeScript…"
cd "$ROOT"
if ! npx tsc –project tsconfig.json 2>&1; then
fail "Ошибка компиляции TypeScript"
fi
ok "TypeScript скомпилирован"

# ── 2. Go-сервер ──────────────────────────────────────────────────────────────

step "Запуск Go-сервера…"
cd "$ROOT/backend-go"
go run . &
GO_PID=$!
PIDS+=($GO_PID)
ok "Go-сервер запущен (PID $GO_PID) → http://localhost:8080"
sleep 2

# ── 3. Проверка ML-модели ─────────────────────────────────────────────────────

step "Проверка ML-модели…"
cd "$ROOT"
MODEL="backend-python/models/fake_review_model.h5"
TFIDF="backend-python/models/tfidf_vectorizer.pkl"
DATASET="backend-python/dataset/wb_reviews.csv"

if [ ! -f "$MODEL" ] || [ ! -f "$TFIDF" ]; then
warn "Модель не найдена — запускаем обучение"

```
if [ ! -f "$DATASET" ]; then
    warn "Датасет не найден — скачиваем"
    python3 backend-python/dataset/get_dataset.py || \
        warn "Не удалось скачать датасет. Запустите вручную: python3 backend-python/dataset/get_dataset.py"
else
    ok "Датасет уже есть"
fi

python3 backend-python/train.py || \
    warn "Ошибка обучения. Запустите вручную: python3 backend-python/train.py"
ok "Модель обучена"
```

else
ok "Модель уже обучена"
fi

# ── 4. Python-сервер ──────────────────────────────────────────────────────────

step "Запуск Python-сервера…"
cd "$ROOT/backend-python"
python3 -m uvicorn main:app –host 0.0.0.0 –port 8000 &
PY_PID=$!
PIDS+=($PY_PID)
ok "Python-сервер запущен (PID $PY_PID) → http://localhost:8000"

# ── Итог ──────────────────────────────────────────────────────────────────────

echo ""
echo "============================================="
echo -e "  \033[32mПроект запущен!\033[0m"
echo "  Сайт:   http://localhost:8080"
echo "  ML API: http://localhost:8000/health"
echo "  Для остановки нажмите Ctrl+C"
echo "============================================="

# Держим скрипт живым пока работает Go-сервер

wait $GO_PID