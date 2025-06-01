FROM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o telegram-bot .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/telegram-bot .

CMD ["./telegram-bot"]