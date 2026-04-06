# Используем официальный образ Go
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Этап финального образа
FROM alpine:latest

# Для работы приложения могут понадобиться сертификаты
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарный файл из builder-стадии
COPY --from=builder /app/main .

# Открываем порт
EXPOSE 8080

# Запуск приложения
CMD ["./main"]
