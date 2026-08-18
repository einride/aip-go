package flop_test

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/internal/assert"
)

type paginatorRow struct {
	ID        int64
	CreatedAt time.Time
	NullTime  bool
}

func testPaginator(t *testing.T) *flop.Paginator[paginatorRow] {
	t.Helper()
	schema, err := flop.NewFilterSchema([]flop.FilterField{
		flop.NewFilterField("age", flop.TypeInt, false),
		flop.NewFilterField("active", flop.TypeBool, false),
		flop.NewFilterField("name", flop.TypeString, true),
		flop.NewFilterField("state", flop.Enum("active", "disabled"), false),
		flop.NewFilterField("created_at", flop.TypeTime, false),
		flop.NewFilterField("wait", flop.TypeDuration, false),
	}, flop.FilterLimits{MaxBytes: 1024, MaxDepth: 16, MaxNodes: 64})
	assert.NoError(t, err)
	paginator, err := flop.NewPaginator(
		[]flop.Column[paginatorRow]{
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
		[]flop.OrderField{flop.Desc("created_at")},
		2,
		5,
	)
	assert.NoError(t, err)
	return paginator
}

func TestNewPaginator(t *testing.T) {
	t.Parallel()

	key := flop.Key("id", func(row paginatorRow) expr.Value { return expr.Int64Value(row.ID) })
	type args struct {
		columns     []flop.Column[paginatorRow]
		schema      *flop.FilterSchema
		order       []flop.OrderField
		defaultSize int32
		maxSize     int32
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "no columns",
			args: args{
				schema:      flop.DisabledFilter(),
				order:       []flop.OrderField{flop.Asc("id")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "nil schema",
			args: args{
				columns:     []flop.Column[paginatorRow]{key},
				order:       []flop.OrderField{flop.Asc("id")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "zero-value schema",
			args: args{
				columns:     []flop.Column[paginatorRow]{key},
				schema:      new(flop.FilterSchema),
				order:       []flop.OrderField{flop.Asc("id")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "no default order",
			args: args{
				columns:     []flop.Column[paginatorRow]{key},
				schema:      flop.DisabledFilter(),
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "invalid sizes",
			args: args{
				columns:     []flop.Column[paginatorRow]{key},
				schema:      flop.DisabledFilter(),
				order:       []flop.OrderField{flop.Asc("id")},
				defaultSize: 2,
				maxSize:     1,
			},
		},
		{
			name: "invalid column",
			args: args{
				columns: []flop.Column[paginatorRow]{
					flop.Key("bad-name", func(row paginatorRow) expr.Value { return expr.Int64Value(row.ID) }),
				},
				schema:      flop.DisabledFilter(),
				order:       []flop.OrderField{flop.Asc("id")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "nil value function",
			args: args{
				columns:     []flop.Column[paginatorRow]{flop.Key[paginatorRow]("id", nil)},
				schema:      flop.DisabledFilter(),
				order:       []flop.OrderField{flop.Asc("id")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "duplicate column",
			args: args{
				columns:     []flop.Column[paginatorRow]{key, key},
				schema:      flop.DisabledFilter(),
				order:       []flop.OrderField{flop.Asc("id")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "no key column",
			args: args{
				columns: []flop.Column[paginatorRow]{
					flop.Sortable("id", func(row paginatorRow) expr.Value { return expr.Int64Value(row.ID) }),
				},
				schema:      flop.DisabledFilter(),
				order:       []flop.OrderField{flop.Asc("id")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "unknown default order",
			args: args{
				columns:     []flop.Column[paginatorRow]{key},
				schema:      flop.DisabledFilter(),
				order:       []flop.OrderField{flop.Asc("missing")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
		{
			name: "duplicate default order",
			args: args{
				columns:     []flop.Column[paginatorRow]{key},
				schema:      flop.DisabledFilter(),
				order:       []flop.OrderField{flop.Asc("id"), flop.Desc("id")},
				defaultSize: 1,
				maxSize:     1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := flop.NewPaginator(
				test.args.columns,
				test.args.schema,
				test.args.order,
				test.args.defaultSize,
				test.args.maxSize,
			)
			assert.Error(t, err)
		})
	}
}

func TestPaginator_CursorParams(t *testing.T) {
	t.Parallel()
	paginator := testPaginator(t)
	seed, err := paginator.CursorParams("", 1, flop.OrderBy{}, flop.Filter{})
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
	// Decodes cleanly and carries a matching binding, but puts an integer where
	// created_at holds a time.
	wrongKind := forge(cursorField("created_at", intValue), cursorField("id", intValue))
	type args struct {
		cursor  string
		size    int32
		order   string
		orderBy flop.OrderBy
		filter  string
	}
	type cursorParams struct {
		limit   int32
		after   flop.CursorPosition
		orderBy flop.OrderBy
		filter  string
	}
	tests := []struct {
		name             string
		args             args
		wantCursorParams cursorParams
		wantErr          assert.ErrorFunc
	}{
		{
			name: "defaults and canonical filter",
			args: args{filter: `age = '42' active = true`},
			wantCursorParams: cursorParams{
				limit:   3,
				orderBy: flop.OrderBy{Fields: []flop.OrderField{flop.Desc("created_at"), flop.Asc("id")}},
				filter:  `(age = 42 AND active = true)`,
			},
			wantErr: assert.NoError,
		},
		{
			name: "custom order retains key direction",
			args: args{size: 5, order: "id desc"},
			wantCursorParams: cursorParams{
				limit:   6,
				orderBy: flop.OrderBy{Fields: []flop.OrderField{flop.Desc("id")}},
			},
			wantErr: assert.NoError,
		},
		{
			name: "page size clamps",
			args: args{size: 99},
			wantCursorParams: cursorParams{
				limit:   6,
				orderBy: flop.OrderBy{Fields: []flop.OrderField{flop.Desc("created_at"), flop.Asc("id")}},
			},
			wantErr: assert.NoError,
		},
		{name: "negative page size", args: args{size: -1}, wantErr: assert.Error},
		{name: "malformed cursor", args: args{cursor: "bad", size: 1}, wantErr: assert.Error},
		{name: "short cursor position", args: args{cursor: shortPosition, size: 1}, wantErr: assert.Error},
		{name: "mismatched cursor path", args: args{cursor: wrongPath, size: 1}, wantErr: assert.Error},
		{name: "null cursor value", args: args{cursor: nullPosition, size: 1}, wantErr: assert.Error},
		{name: "cursor value of the wrong kind", args: args{cursor: wrongKind, size: 1}, wantErr: assert.Error},
		{name: "unknown order", args: args{size: 1, order: "unknown"}, wantErr: assert.Error},
		{
			name: "invalid order path",
			args: args{
				size:    1,
				orderBy: flop.OrderBy{Fields: []flop.OrderField{{Path: "bad-name"}}},
			},
			wantErr: assert.Error,
		},
		{
			name: "duplicate order",
			args: args{
				size:    1,
				orderBy: flop.OrderBy{Fields: []flop.OrderField{flop.Asc("id"), flop.Desc("id")}},
			},
			wantErr: assert.Error,
		},
		{name: "unknown filter", args: args{size: 1, filter: "unknown = 1"}, wantErr: assert.Error},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			orderBy, err := flop.ParseOrderBy(test.args.order)
			assert.NoError(t, err)
			if len(test.args.orderBy.Fields) != 0 {
				orderBy = test.args.orderBy
			}
			filter, err := flop.ParseFilter(test.args.filter)
			assert.NoError(t, err)
			got, err := paginator.CursorParams(test.args.cursor, test.args.size, orderBy, filter)
			test.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.wantCursorParams, cursorParams{
				limit:   got.Limit(),
				after:   got.After(),
				orderBy: got.OrderBy(),
				filter:  got.Filter().String(),
			})
		})
	}
}

func TestPaginator_OffsetParams(t *testing.T) {
	t.Parallel()
	paginator := testPaginator(t)
	type args struct {
		page   int32
		size   int32
		order  string
		filter string
	}
	type offsetParams struct {
		page    int32
		offset  int64
		limit   int32
		orderBy flop.OrderBy
		filter  string
	}
	defaultOrder := flop.OrderBy{Fields: []flop.OrderField{flop.Desc("created_at"), flop.Asc("id")}}
	tests := []struct {
		name             string
		args             args
		wantOffsetParams offsetParams
		wantErr          assert.ErrorFunc
	}{
		{
			name:             "defaults",
			wantOffsetParams: offsetParams{page: 1, limit: 2, orderBy: defaultOrder},
			wantErr:          assert.NoError,
		},
		{
			name:             "page and size",
			args:             args{page: 3, size: 2},
			wantOffsetParams: offsetParams{page: 3, offset: 4, limit: 2, orderBy: defaultOrder},
			wantErr:          assert.NoError,
		},
		{
			name:             "size clamps",
			args:             args{page: 2, size: 99},
			wantOffsetParams: offsetParams{page: 2, offset: 5, limit: 5, orderBy: defaultOrder},
			wantErr:          assert.NoError,
		},
		{
			name: "custom order and filter",
			args: args{page: 1, size: 1, order: "id desc", filter: "age = 42"},
			wantOffsetParams: offsetParams{
				page:    1,
				limit:   1,
				orderBy: flop.OrderBy{Fields: []flop.OrderField{flop.Desc("id")}},
				filter:  "age = 42",
			},
			wantErr: assert.NoError,
		},
		{name: "negative page", args: args{page: -1}, wantErr: assert.Error},
		{name: "negative size", args: args{size: -1}, wantErr: assert.Error},
		{name: "unknown order", args: args{order: "missing"}, wantErr: assert.Error},
		{name: "unknown filter", args: args{filter: "missing = 1"}, wantErr: assert.Error},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			order, err := flop.ParseOrderBy(test.args.order)
			assert.NoError(t, err)
			filter, err := flop.ParseFilter(test.args.filter)
			assert.NoError(t, err)
			got, err := paginator.OffsetParams(test.args.page, test.args.size, order, filter)
			test.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.wantOffsetParams, offsetParams{
				page:    got.Page(),
				offset:  got.Offset(),
				limit:   got.Limit(),
				orderBy: got.OrderBy(),
				filter:  got.Filter().String(),
			})
		})
	}
}

func TestPaginator_CursorPage(t *testing.T) {
	t.Parallel()
	paginator := testPaginator(t)
	params, err := paginator.CursorParams("", 2, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	unknownParams, err := flop.NewCursorParams(
		"",
		2,
		flop.OrderBy{Fields: []flop.OrderField{flop.Asc("missing")}},
		flop.Filter{},
	)
	assert.NoError(t, err)
	emptyParams := flop.CursorParams{}
	now := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	// A nil params selects the paginator's own first-page parameters.
	type args struct {
		rows   []paginatorRow
		params *flop.CursorParams
	}
	// A page carries an opaque cursor, so the mirror holds the position it
	// decodes back to.
	type cursorPage struct {
		items int
		after flop.CursorPosition
	}
	tests := []struct {
		name           string
		args           args
		wantCursorPage cursorPage
		wantErr        assert.ErrorFunc
	}{
		{
			name:           "last page",
			args:           args{rows: []paginatorRow{{ID: 1, CreatedAt: now}}},
			wantCursorPage: cursorPage{items: 1},
			wantErr:        assert.NoError,
		},
		{
			name: "next page",
			args: args{rows: []paginatorRow{
				{ID: 1, CreatedAt: now},
				{ID: 2, CreatedAt: now.Add(-time.Second)},
				{ID: 3, CreatedAt: now.Add(-2 * time.Second)},
			}},
			wantCursorPage: cursorPage{
				items: 2,
				after: flop.CursorPosition{
					{Path: "created_at", Value: expr.TimeValue(now.Add(-time.Second))},
					{Path: "id", Value: expr.Int64Value(2)},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "null cursor column",
			args: args{
				rows: []paginatorRow{{ID: 1, CreatedAt: now}, {ID: 2, NullTime: true}, {ID: 3, CreatedAt: now}},
			},
			wantErr: assert.Error,
		},
		{name: "missing order", args: args{params: &emptyParams}, wantErr: assert.Error},
		{
			name: "undeclared cursor column",
			args: args{
				rows:   []paginatorRow{{ID: 1}, {ID: 2}, {ID: 3}},
				params: &unknownParams,
			},
			wantErr: assert.Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			caseParams := params
			if test.args.params != nil {
				caseParams = *test.args.params
			}
			page, err := paginator.CursorPage(test.args.rows, caseParams)
			test.wantErr(t, err)
			if err != nil {
				return
			}
			got := cursorPage{items: len(page.Items)}
			if page.NextCursor != "" {
				next, err := paginator.CursorParams(page.NextCursor, 1, flop.OrderBy{}, flop.Filter{})
				assert.NoError(t, err)
				got.after = next.After()
			}
			assert.Equal(t, test.wantCursorPage, got)
		})
	}
}
