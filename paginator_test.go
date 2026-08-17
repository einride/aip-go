package flop_test

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/filter"
	"github.com/ysomad/flop/internal/assert"
)

type paginatorRow struct {
	ID        int64
	CreatedAt time.Time
	NullTime  bool
}

func testPaginator(t *testing.T) *flop.Paginator[paginatorRow] {
	t.Helper()
	schema, err := filter.NewSchema([]filter.Field{
		filter.NewField("age", filter.TypeInt, false),
		filter.NewField("active", filter.TypeBool, false),
		filter.NewField("name", filter.TypeString, true),
		filter.NewField("state", filter.Enum("active", "disabled"), false),
		filter.NewField("created_at", filter.TypeTime, false),
		filter.NewField("wait", filter.TypeDuration, false),
	}, filter.Limits{MaxBytes: 1024, MaxDepth: 16, MaxNodes: 64})
	assert.NoError(t, err)
	paginator, err := flop.NewPaginator(
		[]flop.Col[paginatorRow]{
			flop.Sortable("created_at", func(row paginatorRow) expr.Value {
				if row.NullTime {
					return expr.NullValue()
				}
				return expr.TimeValue(row.CreatedAt)
			}),
			flop.Key("id", func(row paginatorRow) expr.Value {
				return expr.Int64Value(row.ID)
			}),
		},
		schema,
		[]flop.Field{flop.Desc("created_at")},
		2,
		5,
	)
	assert.NoError(t, err)
	return paginator
}

func TestNewPaginator(t *testing.T) {
	t.Parallel()

	key := flop.Key("id", func(row paginatorRow) expr.Value { return expr.Int64Value(row.ID) })
	tests := []struct {
		name        string
		columns     []flop.Col[paginatorRow]
		schema      *filter.Schema
		order       []flop.Field
		defaultSize int32
		maxSize     int32
	}{
		{
			name:        "no columns",
			schema:      filter.Disabled(),
			order:       []flop.Field{flop.Asc("id")},
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name:        "nil schema",
			columns:     []flop.Col[paginatorRow]{key},
			order:       []flop.Field{flop.Asc("id")},
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name:        "zero-value schema",
			columns:     []flop.Col[paginatorRow]{key},
			schema:      new(filter.Schema),
			order:       []flop.Field{flop.Asc("id")},
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name:        "no default order",
			columns:     []flop.Col[paginatorRow]{key},
			schema:      filter.Disabled(),
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name:        "invalid sizes",
			columns:     []flop.Col[paginatorRow]{key},
			schema:      filter.Disabled(),
			order:       []flop.Field{flop.Asc("id")},
			defaultSize: 2,
			maxSize:     1,
		},
		{
			name: "invalid column",
			columns: []flop.Col[paginatorRow]{
				flop.Key("bad-name", func(row paginatorRow) expr.Value { return expr.Int64Value(row.ID) }),
			},
			schema:      filter.Disabled(),
			order:       []flop.Field{flop.Asc("id")},
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name:        "nil value function",
			columns:     []flop.Col[paginatorRow]{flop.Key[paginatorRow]("id", nil)},
			schema:      filter.Disabled(),
			order:       []flop.Field{flop.Asc("id")},
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name:        "duplicate column",
			columns:     []flop.Col[paginatorRow]{key, key},
			schema:      filter.Disabled(),
			order:       []flop.Field{flop.Asc("id")},
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name: "no key column",
			columns: []flop.Col[paginatorRow]{
				flop.Sortable("id", func(row paginatorRow) expr.Value { return expr.Int64Value(row.ID) }),
			},
			schema:      filter.Disabled(),
			order:       []flop.Field{flop.Asc("id")},
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name:        "unknown default order",
			columns:     []flop.Col[paginatorRow]{key},
			schema:      filter.Disabled(),
			order:       []flop.Field{flop.Asc("missing")},
			defaultSize: 1,
			maxSize:     1,
		},
		{
			name:        "duplicate default order",
			columns:     []flop.Col[paginatorRow]{key},
			schema:      filter.Disabled(),
			order:       []flop.Field{flop.Asc("id"), flop.Desc("id")},
			defaultSize: 1,
			maxSize:     1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := flop.NewPaginator(test.columns, test.schema, test.order, test.defaultSize, test.maxSize)
			assert.Error(t, err)
		})
	}
}

func TestPaginator_CursorParams(t *testing.T) {
	t.Parallel()
	paginator := testPaginator(t)
	seed, err := paginator.CursorParams("", 1, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	seedTime := time.Unix(1, 0)
	seedPage, err := paginator.CursorPage(
		[]paginatorRow{{ID: 1, CreatedAt: seedTime}, {ID: 2, CreatedAt: seedTime}},
		seed,
	)
	assert.NoError(t, err)
	payload, err := base64.RawURLEncoding.DecodeString(seedPage.NextCursor)
	assert.NoError(t, err)
	prefix := append([]byte(nil), payload[:1+sha256.Size]...)
	intValue := append([]byte{byte(expr.Int64Kind)}, make([]byte, 8)...)
	forge := func(fields ...[]byte) string {
		plaintext := appendUvarint(append([]byte(nil), prefix...), uint64(len(fields)))
		for _, field := range fields {
			plaintext = append(plaintext, field...)
		}
		return base64.RawURLEncoding.EncodeToString(plaintext)
	}
	shortPosition := forge(cursorField("id", intValue))
	wrongPath := forge(cursorField("wrong_path", intValue), cursorField("id", intValue))
	nullPosition := forge(cursorField("created_at", []byte{byte(expr.NullKind)}), cursorField("id", intValue))
	tests := []struct {
		name       string
		cursor     string
		size       int32
		order      string
		orderBy    flop.OrderBy
		filter     string
		wantLimit  int32
		wantOrder  flop.OrderBy
		wantFilter string
		wantErr    assert.ErrorFunc
	}{
		{
			name: "defaults and canonical filter", filter: `age = '42' active = true`,
			wantLimit:  3,
			wantOrder:  flop.OrderBy{Fields: []flop.Field{flop.Desc("created_at"), flop.Asc("id")}},
			wantFilter: `(age = 42 AND active = true)`,
			wantErr:    assert.NoError,
		},
		{
			name: "custom order retains key direction", size: 5, order: "id desc",
			wantLimit: 6, wantOrder: flop.OrderBy{Fields: []flop.Field{flop.Desc("id")}},
			wantErr: assert.NoError,
		},
		{
			name: "page size clamps", size: 99,
			wantLimit: 6,
			wantOrder: flop.OrderBy{Fields: []flop.Field{flop.Desc("created_at"), flop.Asc("id")}},
			wantErr:   assert.NoError,
		},
		{name: "negative page size", size: -1, wantErr: assert.Error},
		{name: "malformed cursor", cursor: "bad", size: 1, wantErr: assert.Error},
		{name: "short cursor position", cursor: shortPosition, size: 1, wantErr: assert.Error},
		{name: "mismatched cursor path", cursor: wrongPath, size: 1, wantErr: assert.Error},
		{name: "null cursor value", cursor: nullPosition, size: 1, wantErr: assert.Error},
		{name: "unknown order", size: 1, order: "unknown", wantErr: assert.Error},
		{
			name:    "invalid order path",
			size:    1,
			orderBy: flop.OrderBy{Fields: []flop.Field{{Path: "bad-name"}}},
			wantErr: assert.Error,
		},
		{
			name:    "duplicate order",
			size:    1,
			orderBy: flop.OrderBy{Fields: []flop.Field{flop.Asc("id"), flop.Desc("id")}},
			wantErr: assert.Error,
		},
		{name: "unknown filter", size: 1, filter: "unknown = 1", wantErr: assert.Error},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			orderBy, err := flop.ParseOrderBy(test.order)
			assert.NoError(t, err)
			if len(test.orderBy.Fields) != 0 {
				orderBy = test.orderBy
			}
			filter, err := filter.Parse(test.filter)
			assert.NoError(t, err)
			got, err := paginator.CursorParams(test.cursor, test.size, orderBy, filter)
			test.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.wantLimit, got.Limit())
			assert.Equal(t, test.wantOrder, got.OrderBy())
			assert.Equal(t, test.wantFilter, got.Filter().String())
		})
	}
}

func TestPaginator_OffsetParams(t *testing.T) {
	t.Parallel()
	paginator := testPaginator(t)
	tests := []struct {
		name       string
		page       int32
		size       int32
		order      string
		filter     string
		wantPage   int32
		wantLimit  int32
		wantOffset int64
		wantErr    assert.ErrorFunc
	}{
		{name: "defaults", wantPage: 1, wantLimit: 2, wantErr: assert.NoError},
		{
			name: "page and size", page: 3, size: 2,
			wantPage: 3, wantLimit: 2, wantOffset: 4, wantErr: assert.NoError,
		},
		{
			name: "size clamps", page: 2, size: 99,
			wantPage: 2, wantLimit: 5, wantOffset: 5, wantErr: assert.NoError,
		},
		{
			name:      "custom order and filter",
			page:      1,
			size:      1,
			order:     "id desc",
			filter:    "age = 42",
			wantPage:  1,
			wantLimit: 1,
			wantErr:   assert.NoError,
		},
		{name: "negative page", page: -1, wantErr: assert.Error},
		{name: "negative size", size: -1, wantErr: assert.Error},
		{name: "unknown order", order: "missing", wantErr: assert.Error},
		{name: "unknown filter", filter: "missing = 1", wantErr: assert.Error},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			order, err := flop.ParseOrderBy(test.order)
			assert.NoError(t, err)
			filter, err := filter.Parse(test.filter)
			assert.NoError(t, err)
			got, err := paginator.OffsetParams(test.page, test.size, order, filter)
			test.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.wantPage, got.Page())
			assert.Equal(t, test.wantLimit, got.Limit())
			assert.Equal(t, test.wantOffset, got.Offset())
			assert.True(t, len(got.OrderBy().Fields) > 0)
			assert.Equal(t, test.filter, got.Filter().String())
		})
	}
}

func TestPaginator_CursorPage(t *testing.T) {
	t.Parallel()
	paginator := testPaginator(t)
	params, err := paginator.CursorParams("", 2, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	unknownParams, err := flop.NewCursorParams(
		"",
		2,
		flop.OrderBy{Fields: []flop.Field{flop.Asc("missing")}},
		filter.Filter{},
	)
	assert.NoError(t, err)
	emptyParams := flop.CursorParams{}
	now := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		rows      []paginatorRow
		wantItems int
		wantAfter flop.CursorPosition
		params    *flop.CursorParams
		wantErr   assert.ErrorFunc
	}{
		{
			name: "last page", rows: []paginatorRow{{ID: 1, CreatedAt: now}},
			wantItems: 1, wantErr: assert.NoError,
		},
		{
			name: "next page",
			rows: []paginatorRow{
				{ID: 1, CreatedAt: now},
				{ID: 2, CreatedAt: now.Add(-time.Second)},
				{ID: 3, CreatedAt: now.Add(-2 * time.Second)},
			},
			wantItems: 2,
			wantAfter: flop.CursorPosition{
				{Path: "created_at", Value: expr.TimeValue(now.Add(-time.Second))},
				{Path: "id", Value: expr.Int64Value(2)},
			},
			wantErr: assert.NoError,
		},
		{
			name:    "null cursor column",
			rows:    []paginatorRow{{ID: 1, CreatedAt: now}, {ID: 2, NullTime: true}, {ID: 3, CreatedAt: now}},
			wantErr: assert.Error,
		},
		{name: "missing order", params: &emptyParams, wantErr: assert.Error},
		{
			name: "undeclared cursor column", params: &unknownParams,
			rows: []paginatorRow{{ID: 1}, {ID: 2}, {ID: 3}}, wantErr: assert.Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			caseParams := params
			if test.params != nil {
				caseParams = *test.params
			}
			page, err := paginator.CursorPage(test.rows, caseParams)
			test.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.wantItems, len(page.Items))
			if len(test.wantAfter) == 0 {
				assert.Equal(t, "", page.NextCursor)
				return
			}
			next, err := paginator.CursorParams(page.NextCursor, 1, flop.OrderBy{}, filter.Filter{})
			assert.NoError(t, err)
			assert.Equal(t, test.wantAfter, next.After())
		})
	}
}
