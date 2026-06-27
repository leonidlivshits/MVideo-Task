FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/price-service ./cmd/price-service \
    && go build -o /bin/migrate ./cmd/migrate

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /bin/price-service /bin/price-service
COPY --from=builder /bin/migrate /bin/migrate
COPY migrations ./migrations

EXPOSE 8080

CMD ["/bin/price-service"]
