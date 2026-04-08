# Примеры запросов

## Запуск сервера

```bash
cd calendar_app
go run ./cmd/server
```

## Создание события

### Формат формы (application/x-www-form-urlencoded)

```bash
curl -X POST "http://localhost:8080/create_event" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "user_id=user1" \
  -d "event=Встреча с клиентом" \
  -d "date=2025-06-01T14:00:00Z" \
  -d "reminder_date=2025-06-01T13:00:00Z" \
  -d "description=Обсуждение проекта"
```

### JSON

```bash
curl -X POST "http://localhost:8080/create_event" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "event": "Встреча с клиентом",
    "date": "2025-06-01T14:00:00Z",
    "reminder_date": "2025-06-01T13:00:00Z",
    "description": "Обсуждение проекта"
  }'
```

**Ответ:**
```json
{
  "result": "event created",
  "event_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Обновление события

```bash
curl -X POST "http://localhost:8080/update_event" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "event_id=550e8400-e29b-41d4-a716-446655440000" \
  -d "event=Встреча перенесена" \
  -d "date=2025-06-02T15:00:00Z"
```

**Ответ:**
```json
{
  "result": "event updated",
  "event_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Удаление события

```bash
curl -X POST "http://localhost:8080/delete_event" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "event_id=550e8400-e29b-41d4-a716-446655440000"
```

**Ответ:**
```json
{
  "result": "event deleted"
}
```

## Получение событий за день

```bash
curl "http://localhost:8080/events_for_day?user_id=user1&date=2025-06-01"
```

**Ответ:**
```json
{
  "result": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "user1",
      "date": "2025-06-01T14:00:00Z",
      "reminder_time": "2025-06-01T13:00:00Z",
      "title": "Встреча с клиентом",
      "description": "Обсуждение проекта",
      "status": "EventStatusCreated",
      "created_at": "2025-05-15T10:00:00Z",
      "updated_at": "2025-05-15T10:00:00Z"
    }
  ]
}
```

## Получение событий за неделю

```bash
curl "http://localhost:8080/events_for_week?user_id=user1&date=2025-06-01"
```

## Получение событий за месяц

```bash
curl "http://localhost:8080/events_for_month?user_id=user1&date=2025-06-01"
```

## Тестирование напоминаний

Создайте событие с напоминанием в прошлом:

```bash
curl -X POST "http://localhost:8080/create_event" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "event": "Тестовое напоминание",
    "date": "2025-01-01T10:00:00Z",
    "reminder_date": "2025-01-01T09:00:00Z"
  }'
```

Через 5 секунд в консоли появится:
```
🔔 Напоминание: Тестовое напоминание
```

## Полный пример сценария

```bash
# 1. Создаём событие
EVENT_ID=$(curl -s -X POST "http://localhost:8080/create_event" \
  -d "user_id=user1" \
  -d "event=День рождения" \
  -d "date=2025-12-31T00:00:00Z" \
  -d "reminder_date=2025-12-30T00:00:00Z" | jq -r '.event_id')

echo "Created event: $EVENT_ID"

# 2. Проверяем за день
curl "http://localhost:8080/events_for_day?user_id=user1&date=2025-12-31"

# 3. Обновляем
curl -X POST "http://localhost:8080/update_event" \
  -d "event_id=$EVENT_ID" \
  -d "event=День рождения (обновлено)"

# 4. Удаляем
curl -X POST "http://localhost:8080/delete_event" \
  -d "event_id=$EVENT_ID"
```
