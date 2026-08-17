# Shop pagination example

This example is a standard-library HTTP server backed by PostgreSQL through
`pgx`. Pagination and filter SQL are produced by the repository's `sqflop`
adapter from Squirrel builders; the local modules are wired in with `replace`
directives.

Start PostgreSQL, then run the server. On every start, the Go migration drops
and recreates the table, then generates 120,000 varied payments.

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

`page_size` defaults to 25 and is capped at 100; `page` defaults to 1. Every
payment column can be filtered; `created_at` is the one column that is filterable
but not sortable. Unqualified text searches cover `id`, `provider`, and
`user_id`; `AND`, `OR`, and `NOT` must be uppercase, and a miscased operator is a
filter syntax error rather than a text search. The default order is
`captured_at desc`, with the primary key `(id, captured_at)` appended by the
library to make cursor positions unique — `provider` is sortable and filterable
but is not part of the key.

## Ordering

`ORDER BY` fields are consulted left to right: a later field only decides rows
that tie on every earlier field. A leading column with no duplicates therefore
makes every field after it inert — the seeded `captured_at` is unique per row, so
`captured_at desc, amount asc` and `captured_at desc, amount desc` return the
same page. The response reports `order_by` as the server resolved it, key columns
included, so a client can tell an applied order from a tie-free window.

Over 1500 stripe payments share `amount = 0`, so there the secondary field
decides and the two pages differ — oldest captures versus newest:

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

Only the primary key `(id, captured_at)` is indexed, so ordering by any other
column sorts the whole filtered set — fine for a demo, but a real endpoint indexes
each sortable column with the appended key columns following it. A
mixed-direction order such as `amount desc, id asc` cannot use a
single-direction index either way.

Configuration:

- `ADDR` defaults to `:8080`.
- `DATABASE_URL` defaults to
  `postgres://shop@localhost:5432/shop?sslmode=disable`.

Every server start regenerates the sample rows, so restarting `go run .` is
enough to reset the data. Removing the volume also works if PostgreSQL itself
needs a clean state.

```sh
docker compose down -v
docker compose up -d --wait
go run .
```
