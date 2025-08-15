# --------- Builder stage ---------
FROM golang:1.24-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk update && \
    apk upgrade && \
    apk add --no-cache bash git build-base && \
    rm -rf /var/cache/apk/*

WORKDIR /app

# Копируем go.mod и go.sum и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Сборка бинарника
RUN go build -o order-service ./cmd/main.go

# --------- Final stage ---------
FROM alpine:latest

# Устанавливаем bash (для ENTRYPOINT) и curl (если нужен wait-for-it)
RUN apk add --no-cache bash curl

WORKDIR /app

# Копируем бинарник и конфиги из builder stage
COPY --from=builder /app/order-service .
COPY --from=builder /app/config ./config

# Копируем скрипты
COPY start.sh wait-for-it.sh ./
RUN chmod +x start.sh wait-for-it.sh

# Открываем порт
EXPOSE 8081

# Запуск через bash
ENTRYPOINT ["bash", "start.sh"]
