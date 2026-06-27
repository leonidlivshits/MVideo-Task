# MVideo Task

Сервис управления ценами товаров.

Может:

- устанавливать цены товаров;
- получать цену товара на конкретный момент времени;
- выгружать историю цен за период в CSV;

## Технологии

- Go 1.26
- PostgreSQL
- pgx / pgxpool
- golang-migrate
- Docker Compose
- OpenAPI / Swagger UI

## Секреты и конфигурация

Переменные находятся в .env.example

## Запуск через Docker Compose

Запуск:

```powershell
docker compose up --build
```

Swagger UI доступен по адресу:

```text
http://localhost:8081
```

## Миграции

Через Docker Compose:

```powershell
docker compose run --rm migrate
```

## HTTP API

OpenAPI-описание лежит в:

```text
docs/openapi.yaml
```

При запуске через Docker Compose документация открывается в Swagger UI:

```text
http://localhost:8081
```

## Формат CSV:

```csv
good_id,create_at,price
1,2026-06-27T10:00:00Z,100
```