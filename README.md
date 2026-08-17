# flop

[![ci](https://github.com/ysomad/flop/actions/workflows/ci.yml/badge.svg)](https://github.com/ysomad/flop/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ysomad/flop/graph/badge.svg)](https://codecov.io/gh/ysomad/flop)
[![Go Reference](https://pkg.go.dev/badge/github.com/ysomad/flop.svg)](https://pkg.go.dev/github.com/ysomad/flop)

Declarative cursor and page-number pagination for Go 1.22+

Four packages:

| Package | Contents |
| --- | --- |
| `flop` | ordering, cursors, page numbers, `Paginator` |
| [`flop/expr`](./expr) | expression nodes, typed values, `Canonical` rendering |
| [`flop/filter`](./filter) | `Filter`, the AIP-160 parser, per-endpoint `Schema` |
| [`flop/sqflop`](./sqflop) | Squirrel query builders and filter compiler |

## Declare an endpoint

Every endpoint explicitly declares its sortable fields, uniqueness keys,
filter schema, default sort, default page size, and maximum page size:

```go
var orderFilters, _ = filter.NewSchema(
	[]filter.Field{
		filter.NewField("state", filter.Enum("pending", "paid"), false),
		filter.NewField("total", filter.TypeInt, false),
		filter.NewField("customer_name", filter.TypeString, true),
		filter.NewField("tags", filter.TypeStringList, false),
	},
	filter.Limits{MaxBytes: 4096, MaxDepth: 16, MaxNodes: 128},
)

var orders, err = flop.NewPaginator(
	[]flop.Col[Order]{
		flop.Sortable("total", func(o Order) expr.Value {
			return expr.Int64Value(o.Total)
		}),
		flop.Sortable("created_at", func(o Order) expr.Value {
			return expr.TimeValue(o.CreatedAt)
		}),
		flop.Key("id", func(o Order) expr.Value {
			return expr.Int64Value(o.ID)
		}),
	},
	orderFilters,
	[]flop.Field{flop.Desc("created_at")},
	50,  // endpoint default
	250, // endpoint maximum
)
```

## Filtering

The parser accepts the [AIP-160 grammar](https://google.aip.dev/assets/misc/ebnf-filtering.txt)

## Cursor pagination

```go
orderBy, err := flop.ParseOrderBy(req.OrderBy)
f, err := filter.Parse(req.Filter)
params, err := orders.CursorParams(req.Cursor, req.PageSize, orderBy, f)
if err != nil {
	// Invalid order, filter, cursor, or non-positive page size: HTTP 400
}

rows, err := repo.List(ctx, params) // fetch params.Limit() rows
page, err := orders.CursorPage(rows, params)
```

## Squirrel

```go
// Every filter field selects the column of the same name.
filterCompiler, err := sqflop.NewFilterCompiler(orderFilters, nil)

// Or map only the columns that differ.
filterCompiler, err := sqflop.NewFilterCompiler(orderFilters, map[string]string{
	"customer_name": "c.name",
})

base := psql.Select("id", "total", "created_at").
	From("orders").
	Where(sq.Eq{"tenant_id": req.Tenant})
base = filterCompiler.Apply(base, params.Filter())

query := sqflop.Cursor(base, params)
sql, args, err := query.ToSql()
```


Page-number queries use the same declaration:

```go
params, err := orders.OffsetParams(req.Page, req.PageSize, orderBy, f)
rows, total, err := repo.ListPage(ctx, params)
page, err := flop.NewOffsetPage(rows, params, total)
```

## Credits

- [Einride AIP Go implementation](https://github.com/einride/aip-go)
- [Google AIP-132: Ordering](https://google.aip.dev/132)
- [Google AIP-158: Pagination](https://google.aip.dev/158)
- [Google AIP-160: Filtering](https://google.aip.dev/160)
- [AIP-160 filtering EBNF](https://google.aip.dev/assets/misc/ebnf-filtering.txt)
- [GitLab pagination guidelines](https://docs.gitlab.com/development/database/pagination_guidelines/)
- [Frontend pagination patterns](https://dev.to/zeeshanali0704/frontend-system-design-pagination-patterns-guide-11di)
