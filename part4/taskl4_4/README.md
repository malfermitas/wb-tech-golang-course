# Go Memory & GC Metrics Exporter

Легковесный HTTP-сервер на Go, экспортирующий метрики памяти и сборщика мусора в формате Prometheus. Включает поддержку `pprof` и динамическое управление `GOGC`.

## Запуск

```bash
go run taskl4_4.go
```

Сервер запустится на `http://localhost:8080`.

## Эндпоинты

| Эндпоинт | Описание |
|----------|----------|
| `/metrics` | Метрики в формате Prometheus |
| `/debug/pprof/` | Интерфейс pprof |

## Примеры запросов

### Получить все метрики

```bash
curl http://localhost:8080/metrics
```

Пример ответа:
```
# HELP go_memstats_alloc_bytes Количество байт, выделенных и всё ещё используемых.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 123456
# HELP go_goroutines Текущее количество горутин.
# TYPE go_goroutines gauge
go_goroutines 8
...
```

### Получить конкретную метрику (текущий аллокации памяти)

```bash
curl http://localhost:8080/metrics | grep go_memstats_alloc_bytes
```

### Получить количество горутин

```bash
curl http://localhost:8080/metrics | grep go_goroutines
```

### Открыть pprof в браузере

```
http://localhost:8080/debug/pprof/
```

### Получить heap профиль

```bash
curl -o heap.prof http://localhost:8080/debug/pprof/heap
```

## Метрики

- `go_memstats_alloc_bytes` — текущая аллокация памяти
- `go_memstats_heap_alloc_bytes` — память в куче
- `go_memstats_heap_sys_bytes` — память полученая от ОС
- `go_memstats_gc_percentage` — GOGC процент
- `go_goroutines` — количество горутин
- `go_memstats_num_gc` — количество GC циклов
- `go_memstats_last_gc_time_seconds` — время последнего GC

## Настройка GOGC (динамически)

Изменить порог GC можно через переменную окружения:

```bash
GOGC=50 go run taskl4_4.go
```