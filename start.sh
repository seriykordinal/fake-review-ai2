#!/bin/bash
echo "НЕ РАБОТАЕТ!!!!!"
exit 0




echo "=== Запуск двух процессов (Go + Python) ==="
echo "Для завершения нажмите Ctrl+C"

# Переход в директорию бэкенда на Go
echo "[1/2] Переход в ./backend-go, компиляция TypeScript..."
cd ./backend-go || { echo "Ошибка: не удалось перейти в ./backend-go"; exit 1; }
npx tsc
echo "Запуск Go-процесса..."
go run . &
PID_GO=$!
echo "   Go процесс запущен, PID: $PID_GO"



# Переход в директорию бэкенда на Python
echo "[2/2] Переход в ../backend-python..."
cd ../backend-python || { echo "Ошибка: не удалось перейти в ../backend-python"; exit 1; }
echo "Запуск Python-процесса..."


if [ ! -f "requirements.txt" ]; then
    echo "   ВНИМАНИЕ: файл requirements.txt не найден, зависимости не будут установлены."
else
    # Создание виртуального окружения, если его нет
    if [ ! -d "venv" ]; then
        echo "   Создание виртуального окружения (venv)..."
        python3 -m venv venv
    fi
    
    # Активация виртуального окружения и установка зависимостей
    echo "   Активация venv и установка зависимостей из requirements.txt..."
    source venv/bin/activate
    pip install --upgrade pip
    pip install -r requirements.txt
    echo "   Зависимости установлены."
fi


python3 main.py &
PID_PYTHON=$!
echo "   Python процесс запущен, PID: $PID_PYTHON"




echo "----------------------------------------"
echo "Оба процесса работают. Нажмите Ctrl+C для останова."
echo ""



# Функция для корректного завершения
cleanup() {
    echo ""
    echo "=== Получен сигнал завершения (Ctrl+C) ==="
    echo "Останавливаем Go процесс (PID $PID_GO)..."
    kill $PID_GO 2>/dev/null
    echo "Останавливаем Python процесс (PID $PID_PYTHON)..."
    kill $PID_PYTHON 2>/dev/null
    echo "Ожидание завершения процессов..."
    wait $PID_GO $PID_PYTHON 2>/dev/null
    echo "Все процессы остановлены. Скрипт завершён."
    exit 0
}

# Устанавливаем обработчик на Ctrl+C (SIGINT) и SIGTERM
trap cleanup SIGINT SIGTERM

# Ожидаем завершения любого из процессов (например, если один упадёт сам)
wait