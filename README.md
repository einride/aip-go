# flop

[![ci](https://github.com/ysomad/flop/actions/workflows/ci.yml/badge.svg)](https://github.com/ysomad/flop/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ysomad/flop.svg)](https://pkg.go.dev/github.com/ysomad/flop)

Declarative cursor and page-number pagination for Go 1.22+

Three packages:

| Package | Contents |
| --- | --- |
| `flop` | ordering, cursors, page numbers, `Paginator` |
| [`flop/expr`](./expr) | expression nodes, typed values, `Canonical` rendering |
| [`flop/filter`](./filter) | `Filter`, the AIP-160 parser, per-endpoint `Schema` |

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

`Sortable` and `Key` columns may both appear in client `order_by`. Keys are also
appended in declaration order when absent, making every resolved order unique.
A key already named by the client keeps its position and direction.
Comparison follows the natural ordering of each declared SQL column type; an
endpoint using different semantics must document them. Fields are consulted left
to right, so a later field only decides rows that tie on every earlier one: a
leading column without duplicates makes the rest of the order inert.

Declarations are validated once and the resulting paginator is immutable and
safe for concurrent package-level reuse. Column names are trusted declarations
and reach SQL verbatim; request strings can only select declared names.
Endpoints that do not expose filtering pass `filter.Disabled()` instead of a
schema; such endpoints accept only an empty filter.

## Filtering

The parser accepts the [AIP-160 grammar](https://google.aip.dev/assets/misc/ebnf-filtering.txt)

## Cursor pagination

```go
orderBy, err := flop.ParseOrderBy(req.OrderBy)
f, err := filter.Parse(req.Filter)
params, err := orders.Cursor(req.Cursor, req.PageSize, orderBy, f)
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
params, err := orders.Offset(req.Page, req.PageSize, orderBy, f)
rows, total, err := repo.ListPage(ctx, params)
page, err := flop.NewPage(rows, params, total)
```

Page numbers are a library extension, not an AIP-158 request shape. They enable
jump-to-page and totals but pay for `OFFSET` and a matching count query.
`Page.TotalItems` maps to AIP `total_size` and may be an estimate if the service
documents it as such; filtered totals should describe the filtered collection.

## Credits

[Einride AIP Go implementation](https://github.com/einride/aip-go)
- [Google AIP-132: Ordering](https://google.aip.dev/132)
- [Google AIP-158: Pagination](https://google.aip.dev/158)
- [Google AIP-160: Filtering](https://google.aip.dev/160)
- [AIP-160 filtering EBNF](https://google.aip.dev/assets/misc/ebnf-filtering.txt)
- [GitLab pagination guidelines](https://docs.gitlab.com/development/database/pagination_guidelines/)
- [Frontend pagination patterns](https://dev.to/zeeshanali0704/frontend-system-design-pagination-patterns-guide-11di)
