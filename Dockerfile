# Собираем приложение
FROM golang:1.24.3 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /telegram-bot ./cmd/bot

# Финальный образ
FROM alpine:latest

WORKDIR /app
COPY --from=builder /telegram-bot /app/telegram-bot

EXPOSE 2112  8080

CMD ["/app/telegram-bot"]