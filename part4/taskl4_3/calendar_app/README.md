# Calendar App

HTTP-сервис для управления событиями календаря с напоминаниями.

## Возможности

- Создание, обновление, удаление событий
- Напоминания через heap (эффективный алгоритм)
- Автоматическая чистка старых событий (30 дней)
- Асинхронный логгер
- Graceful shutdown

## Запуск

### Из исходного кода

```bash
cd calendar_app
go build -o calendar_app ./cmd/server
./calendar_app
```

Сервер запустится на `http://localhost:8080`

### Docker

```bash
docker build -t calendar_app .
docker run -p 8080:8080 calendar_app
```

## API Endpoints

### Создать событие

```bash
POST /create_event
```

Параметры:
- `user_id` - ID пользователя
- `event` - название события
- `date` - дата события (RFC3339)
- `reminder_date` - дата напоминания (RFC3339)
- `description` - описание (опционально)

Пример:
```bash
curl -X POST "http://localhost:8080/create_event" \
  -d "user_id=user1" \
  -d "event=Встреча" \
  -d "date=2025-06-01T10:00:00Z" \
  -d "reminder_date=2025-05-31T10:00:00Z" \
  -d "description=Важная встреча"
```

### Обновить событие

```bash
POST /update_event
```

Параметры:
- `event_id` - ID события
- `user_id` - ID пользователя
- `event` - название события
- `date` - дата события
- `reminder_date` - дата напоминания
- `description` - описание

Пример:
```bash
curl -X POST "http://localhost:8080/update_event" \
  -d "event_id=<EVENT_ID>" \
  -d "event=Встреча (обновлено)"
```

### Удалить событие

```bash
POST /delete_event
```

Параметры:
- `event_id` - ID события

Пример:
```bash
curl -X POST "http://localhost:8080/delete_event" \
  -d "event_id=<EVENT_ID>"
```

### Получить события за день

```bash
GET /events_for_day?user_id=<USER_ID>&date=<DATE>
```

Пример:
```bash
curl "http://localhost:8080/events_for_day?user_id=user1&date=2025-06-01"
```

### Получить события за неделю

```bash
GET /events_for_week?user_id=<USER_ID>&date=<DATE>
```

### Получить события за месяц

```bash
GET /events_for_month?user_id=<USER_ID>&date=<DATE>
```

## Graceful Shutdown

При получении сигнала `SIGINT` или `SIGTERM` сервер:
1. Перестаёт принимать новые соединения
2. Ждёт завершения текущих запросов (до 10 сек)
3. Останавливает все фоновые воркеры

```bash
# Остановка
kill -SIGINT $(pidof calendar_app)
```

## Конфигурация

Параметры чистки задаются в `app.go`:

```go
cleanup.NewCleanupWorker(
    eventRepo,
    ctx,
    10*time.Minute,      // интервал проверки
    30*24*time.Hour,     // удалять события старше 30 дней
)
```

## Архитектура

```
┌─────────────────────────────────────────────┐
│                  main.go                     │
└─────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────┐
│              App (app.go)                    │
│  - создаёт зависимости                       │
│  - управляет lifecycle                       │
└─────────────────────────────────────────────┘
          │           │           │
          ▼           ▼           ▼
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │  HTTP   │ │Notifier │ │Cleanup  │
    │ Server  │ │ Worker  │ │ Worker  │
    └─────────┘ └─────────┘ └─────────┘
          │           │           │
          └───────────┴───────────┘
                      │
                      ▼
           ┌─────────────────┐
           │  In-Memory Repo │
           └─────────────────┘
```

## Требования

- Go 1.25+
- chi/v5
- google/uuid
