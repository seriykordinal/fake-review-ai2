# FakeCheck — Сервис выявления фейковых отзывов

Веб-сервис для автоматического анализа отзывов на товары Wildberries. Собирает отзывы через headless-браузер, оценивает вероятность их фиктивного происхождения с помощью нейронной сети и показывает результаты в удобном интерфейсе.

## Стек

| Слой | Технология |
|---|---|
| Фронтенд | TypeScript, HTML/CSS |
| Go-сервер | Go 1.26, PostgreSQL, Gorilla Mux |
| ML-сервис | Python 3.10+, FastAPI, TensorFlow/Keras |
| Парсер | chromedp (headless Chrome) |

## Как это работает

1. Пользователь вставляет ссылку на товар Wildberries
2. Go-сервер запускает headless Chrome через chromedp и собирает все отзывы
3. Тексты отзывов пакетом отправляются на Python ML-сервис
4. Нейронная сеть (CNN + числовые признаки) возвращает вероятность фейка для каждого отзыва
5. Результаты сохраняются в PostgreSQL и отображаются в интерфейсе

Повторный запрос на тот же товар возвращается из кэша мгновенно — без повторного парсинга.

## Структура проекта

```
fake-review-ai2/
├── backend-go/            # Go API-сервер
│   ├── config/            # Загрузка settings.json
│   ├── database/          # Репозитории (users, analyses, stats)
│   ├── handlers/          # HTTP-обработчики
│   ├── middleware/        # JWT-аутентификация, проверка роли
│   ├── models/            # Структуры данных
│   ├── services/          # Бизнес-логика, парсер WB, JWT, email
│   ├── utils/             # WriteJSON, WriteJSONError
│   └── main.go
│
├── backend-python/        # Python ML-сервис
│   ├── app/               # FastAPI-приложение (main, predictor, models)
│   ├── data/              # Сбор и разметка датасета
│   ├── features/          # Инженерия числовых признаков (5 штук)
│   ├── training/          # Архитектура модели, токенизатор, обучение
│   └── tests/             # pytest-тесты API, предиктора, схем
│
├── frontend/              # TypeScript-исходники
│   ├── services/          # Запросы к API (auth, product, admin)
│   ├── types/             # Типы данных
│   └── utils/             # Вспомогательные функции, работа с DOM
│
├── static/                # Скомпилированный JS + HTML + CSS
├── settings.example.json  # Пример конфигурации
├── package.json           # TypeScript-компилятор
└── tsconfig.json
```

## Требования

- **Go** 1.21+
- **Python** 3.10+
- **Node.js** 18+ (для компиляции TypeScript)
- **PostgreSQL** 14+
- **Google Chrome** (для парсера Wildberries)

## Установка и запуск

### 1. Клонировать репозиторий

```bash
git clone https://github.com/seriykordinal/fake-review-ai2.git
cd fake-review-ai2
```

### 2. Создать конфигурацию

Скопировать пример и заполнить своими значениями:

```bash
cp settings.example.json settings.json
```

```json
{
  "database": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "your_password",
    "dbname": "fake-review-ai2"
  },
  "jwt": {
    "secret": "замените_на_случайную_строку"
  },
  "go_server":     { "port": 8080 },
  "python_server": { "host": "localhost", "port": 8000 },
  "email": {
    "from": "your@gmail.com",
    "password": "app_password",
    "smtp_host": "smtp.gmail.com",
    "smtp_port": "587"
  },
  "email_verification_enabled": false,
  "debug": false
}
```

> ⚠️ Не коммитьте `settings.json` — он уже добавлен в `.gitignore`.

`debug: true` включает подробные логи парсера и Go-сервера.

### 3. Создать базу данных

```sql
CREATE DATABASE "fake-review-ai2";
```

Таблицы создаются автоматически при первом запуске Go-сервера.

### 4. Установить Python-зависимости

```bash
pip install -r backend-python/requirements.txt
```

### 5. Скомпилировать TypeScript

```bash
npx tsc
```

### 6. Обучить модель (один раз)

```bash
# Собрать отзывы для датасета (нужен запущенный Go-сервер и аккаунт)
python -m backend-python.data.collect_reviews

# Разметить датасет
python -m backend-python.data.label_dataset

# Обучить модель
python -m backend-python.training.train
```

Артефакты сохраняются в `backend-python/artifacts/`: `model.keras`, `tokenizer.pkl`, `feature_scaler.pkl`.

### 7. Запустить серверы

```bash
# Терминал 1 — Go-сервер
cd backend-go && go run .

# Терминал 2 — Python ML-сервис
cd backend-python && uvicorn app.main:app --host 0.0.0.0 --port 8000
```

Сайт доступен по адресу: **http://localhost:8080**

## Роли пользователей

| Роль | Возможности |
|---|---|
| `user` | Анализ одного отзыва, анализ товара по URL, просмотр своей истории, управление аккаунтом |
| `admin` | Всё выше + список всех пользователей и анализов, удаление записей, просмотр статистики |
| `super_admin` | Всё выше + смена ролей любому пользователю |

## API

### Публичные эндпоинты

| Метод | URL | Описание |
|---|---|---|
| `POST` | `/api/register` | Регистрация |
| `POST` | `/api/login` | Вход |
| `POST` | `/api/verify` | Подтверждение email |
| `POST` | `/api/analyze` | Анализ одного текста отзыва |

### Защищённые (требуют JWT в заголовке `Authorization: Bearer <token>`)

| Метод | URL | Описание |
|---|---|---|
| `GET` | `/api/profile` | Данные профиля |
| `POST` | `/api/analyze_product` | Анализ всех отзывов товара по URL |
| `GET` | `/api/history` | История анализов пользователя |
| `DELETE` | `/api/account` | Удаление аккаунта |

### Админ (требуют роль `admin` или `super_admin`)

| Метод | URL | Описание |
|---|---|---|
| `GET` | `/api/admin/users` | Список пользователей |
| `DELETE` | `/api/admin/users/{id}` | Удаление пользователя |
| `POST` | `/api/admin/users/role` | Смена роли |
| `GET` | `/api/admin/analyses` | Все анализы товаров |
| `DELETE` | `/api/admin/analyses/{id}` | Удаление анализа |
| `GET` | `/api/admin/stats` | Глобальная статистика |

### ML-сервис (внутренний, порт 8000)

| Метод | URL | Описание |
|---|---|---|
| `GET` | `/health` | Статус сервиса и состояние модели |
| `POST` | `/predict` | Анализ одного отзыва `{"text": "..."}` |
| `POST` | `/predict_batch` | Пакетный анализ `{"texts": [...]}` |

## Тесты

```bash
# Go
cd backend-go && go test ./...

# Python
cd backend-python && pytest tests/ -v
```

73 теста покрывают конфигурацию, JWT-middleware, обработчики, парсер URL, Pydantic-схемы, предиктор и HTTP API ML-сервиса.
