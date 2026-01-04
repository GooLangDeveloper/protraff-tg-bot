FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod tidy

COPY . .
RUN go build -ldflags="-s -w" -o bot

CMD ["./bot"]
