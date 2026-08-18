# Shop pagination example

```sh
docker compose up -d --wait
go run .
```

List payments with a cursor:

```sh
curl -sS http://localhost:8080/cursor-payments \
  -H 'content-type: application/json' \
  -d '{
    "page_size": 3,
    "order_by": "amount desc",
    "filter": "provider = stripe AND amount >= 1000"
  }'
```

Send the returned `next_cursor` with the same order and filter for the next
page. Page size may change between requests.

```json
{
  "page_size": 10,
  "order_by": "amount desc",
  "filter": "provider = stripe AND amount >= 1000",
  "cursor": "..."
}
```

List offset page 20, including filtered totals:

```sh
curl -sS http://localhost:8080/offset-payments \
  -H 'content-type: application/json' \
  -d '{
    "page": 20,
    "page_size": 10,
    "order_by": "provider, amount desc",
    "filter": "created_at >= 2024-01-01T00:00:00Z"
  }'
```

## Ordering

```sh
curl -sS http://localhost:8080/offset-payments \
  -H 'content-type: application/json' \
  -d '{
    "page_size": 3,
    "order_by": "amount asc, captured_at desc",
    "filter": "provider = \"stripe\" AND amount = 0"
  }'
```

```sh
curl -sS http://localhost:8080/offset-payments \
  -H 'content-type: application/json' \
  -d '{
    "page_size": 3,
    "order_by": "amount asc, captured_at asc",
    "filter": "provider = \"stripe\" AND amount = 0"
  }'
```

Configuration:

- `ADDR` defaults to `:8080`.
- `DATABASE_URL` defaults to
  `postgres://shop@localhost:5432/shop?sslmode=disable`.
