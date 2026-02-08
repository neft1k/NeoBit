# NeoBIT: прототип работы с векторами в Postgres

## 1. Цель
Прототип сервиса на Go для работы с векторными данными в PostgreSQL:
- хранение эмбеддингов в `pgvector`;
- фоновый импорт датасета (не блокирует API);
- фоновая кластеризация документов с `cluster_id IS NULL`;
- REST API по требованиям ТЗ;
- запуск через Docker Compose.

## 2. Выбранное расширение Postgres
Выбрано расширение: **pgvector**.

Причины:
- нативный тип `vector(N)`;
- SQL-операции расстояний/сходства;
- поддержка индексов HNSW / IVFFlat;
- простая интеграция с Go (`pgx` + `pgvector-go`).

В проекте используется:
- `VECTOR(384)` для эмбеддингов;
- индекс `HNSW` с `vector_cosine_ops`.

## 3. Что реализовано по ТЗ
- схема БД + миграции (`goose`);
- таблицы `documents` и `clusters`;
- фоновой импорт данных (батчами);
- фоновая кластеризация новых документов;
- обязательный REST API:
  - `GET /clusters?limit=&offset=`
  - `GET /clusters/{id}/documents?limit=&offset=`
  - `GET /documents/{id}`
  - `POST /documents/`
- Docker Compose: Postgres (pgvector) + app + Prometheus + подготовка среза датасета;
- graceful shutdown сервера и воркеров.

## 4. Архитектура
- `internal/repository/*` — доступ к БД;
- `internal/service/document` — бизнес-логика документов;
- `internal/service/importer` — импорт датасета;
- `internal/service/cluster` — кластеризация + worker;
- `internal/transport/http/handler/*` — HTTP-слой;
- `internal/server` — wiring, роутер, запуск/остановка воркеров.

## 5. Схема БД
Миграция: `internal/db/migrations/000001_init.sql`

### Таблица `clusters`
- `id BIGSERIAL PRIMARY KEY`
- `algorithm TEXT`
- `k INT`
- `centroid VECTOR(384)`
- `created_at`, `updated_at`

### Таблица `documents`
- `id BIGSERIAL PRIMARY KEY`
- `hn_id BIGINT`
- `title`, `url`, `by`, `text`
- `score INT`
- `time TIMESTAMPTZ`
- `embedding VECTOR(384) NOT NULL`
- `cluster_id BIGINT NULL REFERENCES clusters(id)`
- `created_at`, `updated_at`

### Индексы
- `idx_documents_cluster_id` на `documents(cluster_id)`
- `idx_documents_embedding_hnsw` на `documents USING hnsw (embedding vector_cosine_ops)`

## 6. Запуск
### Требования
- Docker + Docker Compose
- `goose` (для миграций)

### Шаги
1. Запуск контейнеров:
```bash
docker compose up --build
```

2. Применение миграций (отдельно, вручную):
```bash
make migrate-up
```

3. Просмотр логов:
```bash
docker compose logs -f app
```

Примечание:
- сервис `dataset` в compose формирует срез датасета `200k` в `hn_200k.csv`;
- `app` импортирует этот файл в фоне батчами.

## 7. Примеры API
Базовый URL: `http://localhost:8080`

### Создать документ
`embedding` должен содержать **384 float-значения**.

```bash
EMB=$(python3 - <<'PY'
print("[" + ",".join(["0.01"]*384) + "]")
PY
)

curl -X POST http://localhost:8080/documents/ \
  -H "Content-Type: application/json" \
  -d "{
    \"hn_id\": 123456789,
    \"title\": \"Example\",
    \"url\": \"https://example.com\",
    \"by\": \"user1\",
    \"score\": 10,
    \"time\": \"2024-01-01T12:00:00Z\",
    \"text\": \"sample text\",
    \"embedding\": $EMB
  }"
```

### Получить документ по id
```bash
curl "http://localhost:8080/documents/1"
```

### Получить список кластеров
```bash
curl "http://localhost:8080/clusters?limit=20&offset=0"
```

### Получить документы кластера
```bash
curl "http://localhost:8080/clusters/1/documents?limit=20&offset=0"
```

## 8. Метрики и оценка кластеризации
Endpoint метрик:
- `http://localhost:8080/metrics`
- Prometheus UI: `http://localhost:9090`

Собираемые метрики:
- `http_requests_total`
- `http_request_duration_seconds`
- `http_active_requests`
- `rate_limit_exceeded_total`
- `cluster_size_min`
- `cluster_size_max`
- `cluster_size_avg`
- `pct_clustered`

SQL-проверки из ТЗ:
```sql
-- Размеры кластеров
SELECT cluster_id, COUNT(*) AS size
FROM documents
GROUP BY cluster_id
ORDER BY size DESC;

-- Доля кластеризованных документов
SELECT
  100.0 * COUNT(*) FILTER (WHERE cluster_id IS NOT NULL) / COUNT(*) AS pct_clustered
FROM documents;
```

## 9. Команды
```bash
# Postgres
make db-up
make db-logs
make db-down

# Миграции
make migrate-up
make migrate-down
make migrate-status

# Тесты
go test ./...

# Линтер (если установлен)
golangci-lint run ./...
```

## 10. Ограничения текущего прототипа
- импорт-воркер запускается один раз при старте приложения;
- кластеризация в текущем коде упрощенная (прототипный вариант);
- миграции запускаются вручную, не автоматически.
