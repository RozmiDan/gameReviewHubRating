FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o rating_service ./cmd/app/main.go

RUN ls -lah /app

FROM debian:latest

WORKDIR /app

COPY --from=builder /app/rating_service /app/rating_service

COPY config/config.prod.yaml /app/config.prod.yaml

ENV CONFIG_PATH="/app/config.prod.yaml"

EXPOSE 8080

CMD ["/app/rating_service"]
