# 🚀 URL Link Shortener

> Высоконагруженный REST API сервис для сокращения URL с кэшированием в Redis и хранением в PostgreSQL.

[![Go Version](https://img.shields.io/badge/Go-1.25.5-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-✅-2496ED?logo=docker)](https://docker.com)

## 📖 О проекте

URL Link Shortener — это backend-сервис для создания коротких ссылок с поддержкой:

- ⚡ **Мгновенного редиректа** через кэш Redis
- 📊 **Статистики переходов** с асинхронным подсчётом
- 🔒 **Graceful shutdown** и обработкой ошибок
- 🧪 **Unit-тестов** с моками для всех слоёв
- 🐳 **Полной контейнеризации** через Docker Compose

> ⚠️ Проект является учебным и демонстрирует применение Clean Architecture, Dependency Injection и best practices в Go.

## ✨ Возможности

| Функция                   | Описание                                                                       |
| ------------------------- | ------------------------------------------------------------------------------ |
| 🔗 **Создание ссылок**    | POST `/api/v1/shorten` — валидация, генерация кода, сохранение в БД + кэш      |
| 🔄 **Редирект**           | GET `/:code` — мгновенный ответ из кэша или БД, асинхронный подсчёт просмотров |
| 📈 **Статистика**         | GET `/api/v1/:code/stats` — общее количество переходов + метаданные ссылки     |
| ❤️ **Health Check**       | GET `/health` — проверка подключения к PostgreSQL и Redis                      |
| 🗑️ **Авто-очистка**       | Ссылки с истёкшим сроком (`expires_at`) или `is_active=false` возвращают 404   |
| 🛡️ **CORS**               | Гибкая настройка разрешённых доменов через `CORS_ORIGINS`                      |
| 🪵 **Structured Logging** | JSON-логи с `request_id`, методом, статусом, latency и client_ip               |

## 🛠 Технологический стек

```yaml
Language: Go 1.25.5
Framework: Gin v1.9+
Database: PostgreSQL 15 (pgx/v5)
Cache: Redis 7 (go-redis/v9)
Logging: Zap (structured, JSON/console)
Testing: testify + ручные моки
Container: Docker + Docker Compose
Architecture: Clean Architecture + Dependency Injection
```

## 📸 Скриншоты

### Архитектура базы данных

![Схема базы данных](./docs/screenshots/shortener_db_schema.png)

## 🚀 Быстрый старт

### Требования

- Docker & Docker Compose **или** Go 1.21+

### 1. Клонирование

```bash
git clone https://github.com/Dimidroll06/url-link-shortener.git
cd url-link-shortener
```

### 2. Настройка окружения

```bash
# Скопируйте шаблон
cp .env.example .env

# При необходимости отредактируйте пароли и порты
nano .env
```

### 3. Запуск через Docker (рекомендуется)

```bash
# Сборка и запуск всех сервисов
docker-compose up --build

# Запуск в фоновом режиме
docker-compose up -d

# Просмотр логов
docker-compose logs -f app
```

### 4. Запуск локально (без Docker)

```bash
# Убедитесь, что PostgreSQL и Redis запущены
# Примените миграции
migrate -path migrations -database "postgres://user:pass@localhost:5432/db?sslmode=disable" up

# Запустите приложение
go run ./cmd/server
```

### 5. Проверка работы

```bash
# Health check
curl http://localhost:8080/health

# Создать короткую ссылку
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# Перейти по короткой ссылке (редирект)
curl -I http://localhost:8080/abc12345

# Получить статистику
curl http://localhost:8080/api/v1/abc12345/stats
```

## 📡 API Reference

[📃 OpenAPI документация](./docs/openapi.yaml)

### 🔗 Создать короткую ссылку

```http
POST /api/v1/shorten
Content-Type: application/json
```

**Request:**

```json
{
  "url": "https://example.com/very/long/path"
}
```

**Success Response (201 Created):**

```json
{
  "short_code": "abc12345",
  "original_url": "https://example.com/very/long/path",
  "short_url": "http://localhost:8080/abc12345"
}
```

**Error Responses:**
| Status | Body | Описание |
|-|-|-|
| `400` | `{"error":"invalid request body"}` | Неверный JSON или пустой URL |
| `400` | `{"error":"invalid URL scheme: ..."}` | URL не начинается с `http://` или `https://` |
| `400` | `{"error":"url too long: ..."}` | Длина URL > 2048 символов |
| `409` | `{"error":"short code already exists"}` | Сгенерированный код уже занят (крайне редко) |
| `500` | `{"error":"internal server error"}` | Ошибка БД или внутренняя ошибка |

### 🔄 Редирект по короткому коду

```http
GET /:code
```

**Success Response (302 Found):**

```
Location: https://example.com/very/long/path
```

**Error Responses:**
| Status | Body | Описание |
|-|-|-|
| `404` | `{"error":"link not found"}` | Ссылка не найдена, истекла или деактивирована |
| `500` | `{"error":"internal server error"}` | Ошибка сервиса |

> 💡 Используется код `302` (временный редирект), чтобы статистика переходов собиралась корректно. Браузеры не кэшируют 302, в отличие от 301.

### 📊 Получить статистику

```http
GET /api/v1/:code/stats
```

**Success Response (200 OK):**

```json
{
  "short_code": "abc12345",
  "original_url": "https://example.com/very/long/path",
  "total_accesses": 42,
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-02-15T10:30:00Z"
}
```

**Error Responses:**
| Status | Body | Описание |
|-|-|-|
| `404` | `{"error":"link not found"}` | Ссылка не найдена или невалидна |
| `500` | `{"error":"failed to get statistics"}` | Ошибка при получении статистики |

### ❤️ Health Check

```http
GET /health
```

**Success (200 OK):**

```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error (503 Service Unavailable):**

```json
{
  "status": "unhealthy",
  "error": "database" // или "redis"
}
```

## ⚙️ Конфигурация

Все настройки задаются через переменные окружения или файл `.env`.

### 🔐 Обязательные переменные (для production)

| Переменная       | Описание                                   | Пример                |
| ---------------- | ------------------------------------------ | --------------------- |
| `DB_PASSWORD`    | Пароль к PostgreSQL                        | `secure_password_123` |
| `REDIS_PASSWORD` | Пароль к Redis (если включён)              | `secure_password_123` |
| `JWT_SECRET`     | Секрет для JWT (если добавите авторизацию) | `change_me_in_prod`   |

### 🌐 Сервер

| Переменная     | Default                 | Описание                                    |
| -------------- | ----------------------- | ------------------------------------------- |
| `SERVER_PORT`  | `8080`                  | Порт HTTP-сервера                           |
| `GIN_MODE`     | `release`               | Режим Gin: `debug`, `release`, `test`       |
| `APP_ENV`      | `development`           | Окружение: `development`, `production`      |
| `BASE_URL`     | `http://localhost:8080` | Базовый URL для генерации коротких ссылок   |
| `CORS_ORIGINS` | `*`                     | Разрешённые домены для CORS (через запятую) |

### 🗄️ PostgreSQL

| Переменная    | Default     | Описание            |
| ------------- | ----------- | ------------------- |
| `DB_HOST`     | `localhost` | Хост базы данных    |
| `DB_PORT`     | `5432`      | Порт PostgreSQL     |
| `DB_USER`     | `postgres`  | Пользователь БД     |
| `DB_PASSWORD` | —           | Пароль пользователя |
| `DB_NAME`     | `postgres`  | Имя базы данных     |

### 🗄️ Redis

| Переменная       | Default     | Описание                |
| ---------------- | ----------- | ----------------------- |
| `REDIS_HOST`     | `localhost` | Хост Redis              |
| `REDIS_PORT`     | `6379`      | Порт Redis              |
| `REDIS_PASSWORD` | —           | Пароль Redis            |
| `REDIS_DB`       | `0`         | Номер базы данных Redis |

### ⚙️ Прочие настройки

| Переменная            | Default | Описание                                              |
| --------------------- | ------- | ----------------------------------------------------- |
| `SHUTDOWN_TIMEOUT`    | `30s`   | Таймаут graceful shutdown                             |
| `URL_EXPIRATION_DAYS` | `30`    | Срок жизни ссылки по умолчанию (0 = бессрочно)        |
| `LOG_LEVEL`           | `info`  | Уровень логирования: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT`          | `json`  | Формат логов: `json`, `console`                       |

> 💡 **Совет**: В production установите `CORS_ORIGINS=https://yourdomain.com` вместо `*` для безопасности.

## 📂 Структура проекта

```
url-link-shortener/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа: DI, инициализация, запуск сервера
│
├── internal/                    # Приватный код приложения (Clean Architecture)
│   ├── adapters/                # Инфраструктурный слой (внешние зависимости)
│   │   ├── cache/               # Реализация кэша (Redis)
│   │   ├── handlers/            # HTTP-хендлеры (Gin)
│   │   ├── repository/          # Реализация репозиториев (PostgreSQL)
│   │   └── server/              # Настройка HTTP-сервера, middleware, graceful shutdown
│   │
│   ├── config/                  # Загрузка и валидация конфигурации (.env)
│   │
│   ├── core/                    # Domain layer (не зависит от внешнего мира)
│   │   ├── domain/              # Сущности: URL, URLStats
│   │   ├── errors/              # Централизованные ошибки приложения
│   │   ├── ports/               # Интерфейсы: Repository, Cache, Service
│   │   └── services/            # Бизнес-логика: URLService, StatsService
│   │
│   └── tests/                   # Тестовая инфраструктура
│       ├── mock/                # Mock-реализации интерфейсов
│       │   ├── cache/
│       │   ├── repository/
│       │   └── services/
│       └── unit/                # Unit-тесты
│           ├── handlers/
│           └── services/
│
├── migrations/                  # SQL-миграции (golang-migrate формат)
│   ├── 001_create_urls_table.up.sql
│   ├── 001_create_urls_table.down.sql
│   ├── 002_create_statistics_table.up.sql
│   └── 002_create_statistics_table.down.sql
│
├── .env.example                 # Шаблон переменных окружения
├── docker-compose.yml           # Оркестрация: app + postgres + redis + migrate
├── Dockerfile                   # Multi-stage сборка Go-приложения
├── go.mod                       # Зависимости Go
└── README.md                    # Этот файл
```

## 🧪 Тестирование

### Запуск всех тестов

```bash
# Unit-тесты с race detector
go test -v -race ./...

# Только тесты сервисов
go test -v -race ./internal/tests/unit/services/...

# Только тесты хендлеров
go test -v -race ./internal/tests/unit/handlers/...
```

### Покрытие кода

```bash
# Генерация отчёта
go test ./... -coverprofile=coverage.out

# HTML-отчёт в браузере
go tool cover -html=coverage.out -o coverage.html
```

### Очистка кэша тестов

```bash
# Если тесты ведут себя странно — очистите кэш
go clean -testcache
```

## 🔍 Observability

### Structured Logging

Приложение логирует все запросы в JSON-формате (если `LOG_FORMAT=json`):

```json
{
  "level": "info",
  "msg": "request completed",
  "method": "POST",
  "path": "/api/v1/shorten",
  "status": 201,
  "latency": "2.345ms",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_ip": "127.0.0.1",
  "user_agent": "curl/7.68.0"
}
```

### Request ID

Каждый запрос получает уникальный `X-Request-ID`:

- Генерируется автоматически, если не передан в заголовке
- Добавляется в логи для трассировки
- Возвращается в ответе для отладки

```bash
curl -i http://localhost:8080/health
# Ответ:
# X-Request-Id: 550e8400-e29b-41d4-a716-446655440000
```

### Health Checks

Эндпоинт `/health` используется Docker Compose для:

- Проверки готовности приложения (`start_period`)
- Автоматического перезапуска при падении БД/Redis
- Graceful shutdown при деплое

## 🐳 Docker

### Сборка образа

```bash
docker build -t url-shortener:latest .
```

### Запуск через Compose

```bash
# Полная сборка и запуск
docker-compose up --build

# Фоновый режим + просмотр логов
docker-compose up -d && docker-compose logs -f app

# Остановка с graceful shutdown (35 секунд)
docker-compose stop app
```

### Переменные в Docker

- Приложение внутри контейнера подключается к БД по хосту `db`, а не `localhost`
- Для локальной разработки (без Docker) используйте `DB_HOST=localhost` в `.env`
- Один файл `.env` работает для обоих сценариев благодаря переопределению в `docker-compose.yml`

## 🤝 Contributing

1.  Создайте feature-ветку: `git checkout -b feature/amazing-feature`
2.  Внесите изменения и добавьте тесты
3.  Убедитесь, что все тесты проходят: `go test ./...`
4.  Закоммитьте: `git commit -m 'Add amazing feature'`
5.  Откройте Pull Request

См. [CONTRIBUTING.md](CONTRIBUTING.md) для деталей.

## 📄 Лицензия

Этот проект распространяется под лицензией MIT. Подробнее см. в файле [LICENSE](LICENSE).

<p align="center">
  Made with ❤️ using Go
</p>
