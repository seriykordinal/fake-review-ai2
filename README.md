# FakeCheck — Сервис выявления фейковых отзывов 

Веб-сервис для анализа отзывов на товары Wildberries с помощью машинного обучения.

## Стек

|Слой     |Технология                        |
|---------|----------------------------------|
|Фронтенд |TypeScript, HTML/CSS              |
|Go-сервер|Go, PostgreSQL                    |
|ML-сервис|Python + FastAPI, TensorFlow/Keras|
|Парсер   |chromedp (headless Chrome)        |

## Структура проекта

```


fake-review-ai1/
├── README.md
├── backend-go/
│   ├── config/
│   ├── database/
│   ├── go.mod
│   ├── go.sum
│   ├── handlers/
│   ├── main.go
│   ├── middleware/
│   ├── models/
│   └── services/
│       
├── backend-python/
│   ├── dataset/
│   ├── main.py
│   ├── models/
│   ├── requirements.txt
│   ├── train.py
│   └── venv/
|
├── frontend/
│   ├── admin-app.ts
│   ├── app.ts
│   ├── services/
│   ├── types/
│   └── utils/
│       
├── static/
|
├── node_modules/
├── package-lock.json
├── package.json
├── settings.example.json
├── settings.json
├── start.sh
│   
└── tsconfig.json

```

## Требования

- **Go** 1.21+
- **Python** 3.10+
- **Node.js** 18+ (для компиляции TypeScript)
- **PostgreSQL** 14+
- **Google Chrome** (для парсера WB)

Установка Python-зависимостей:

```bash
pip install requirements.txt
```

## Настройка

Все параметры хранятся в `settings.json` в корне проекта:

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
  "go_server":    { "port": 8080 },
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

**`debug: true`** — включает подробные логи во всех компонентах (Go-сервер, парсер).  
**`debug: false`** — выводятся только ошибки.

> ⚠️ Не коммитьте `settings.json` в репозиторий — он содержит пароли. Добавьте его в `.gitignore`.

## База данных

Создайте базу и запустите миграции:

```sql
CREATE DATABASE "fake-review-ai2";
```

Таблицы создаются автоматически при первом запуске Go-сервера.

## Запуск

### Быстрый старт (Windows)

```powershell
.\start.ps1
```

Скрипт автоматически:

1. Компилирует TypeScript
1. Запускает Go-сервер
1. Скачивает датасет (если отсутствует)
1. Обучает модель (если не обучена)
1. Запускает Python-сервер

### Ручной запуск

```bash
# 1. Компиляция TypeScript
npx tsc

# 2. Go-сервер
cd backend-go && go run .

# 3. Датасет (один раз)
python backend-python/dataset/get_dataset.py

# 4. Обучение модели (один раз)
python backend-python/train.py

# 5. Python-сервер
cd backend-python && python main.py
```

После запуска сайт доступен по адресу: **http://localhost:8080**

## Роли пользователей

|Роль         |Возможности                                                                      |
|-------------|---------------------------------------------------------------------------------|
|`user`       |Анализ отзывов и товаров, просмотр своей истории                                 |
|`admin`      |Всё выше + просмотр всех пользователей и анализов, удаление обычных пользователей|
|`super_admin`|Всё выше + смена ролей, удаление любых пользователей                             |

## API

### Публичные эндпоинты

|Метод |URL            |Описание                   |
|------|---------------|---------------------------|
|`POST`|`/api/register`|Регистрация                |
|`POST`|`/api/login`   |Вход                       |
|`POST`|`/api/verify`  |Подтверждение email        |
|`POST`|`/api/analyze` |Анализ одного текста отзыва|

### Защищённые (требуют JWT)

|Метод   |URL                   |Описание                         |
|--------|----------------------|---------------------------------|
|`GET`   |`/api/profile`        |Данные профиля                   |
|`POST`  |`/api/analyze_product`|Анализ всех отзывов товара по URL|
|`GET`   |`/api/history`        |История анализов пользователя    |
|`DELETE`|`/api/account`        |Удаление аккаунта                |

### Админ (требуют роль admin/super_admin)

|Метод   |URL                       |Описание             |
|--------|--------------------------|---------------------|
|`GET`   |`/api/admin/users`        |Список пользователей |
|`DELETE`|`/api/admin/users/{id}`   |Удаление пользователя|
|`POST`  |`/api/admin/users/role`   |Смена роли           |
|`GET`   |`/api/admin/analyses`     |Все анализы          |
|`DELETE`|`/api/admin/analyses/{id}`|Удаление анализа     |
|`GET`   |`/api/admin/stats`        |Глобальная статистика|