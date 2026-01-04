FROM golang:1.21-alpine

WORKDIR /app

# Сначала копируем go.mod и исходник
COPY go.mod ./
COPY main.go ./

# Генерируем go.sum и скачиваем зависимости
RUN go mod tidy

# Собираем бинарник
RUN go build -ldflags="-s -w" -o bot

CMD ["./bot"]
