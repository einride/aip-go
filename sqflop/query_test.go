package sqflop

import (
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/filter"
	"github.com/ysomad/flop/internal/assert"
)

type queryRow struct {
	ID        int64
	CreatedAt time.Time
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func queryPaginator(t *testing.T) *flop.Paginator[queryRow] {
	t.Helper()
	paginator, err := flop.NewPaginator(
		[]flop.Col[queryRow]{
			flop.Sortable("created_at", func(row queryRow) expr.Value { return expr.TimeValue(row.CreatedAt) }),
			flop.Key("id", func(row queryRow) expr.Value { return expr.Int64Value(row.ID) }),
		},
		filter.Disabled(),
		[]flop.Field{flop.Desc("created_at")},
		2,
		100,
	)
	assert.NoError(t, err)
	return paginator
}

func TestCursor(t *testing.T) {
	t.Parallel()
	paginator := queryPaginator(t)
	first, err := paginator.Cursor("", 2, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	now := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	page, err := paginator.CursorPage([]queryRow{
		{ID: 1, CreatedAt: now},
		{ID: 2, CreatedAt: now.Add(-time.Second)},
		{ID: 3, CreatedAt: now.Add(-2 * time.Second)},
	}, first)
	assert.NoError(t, err)
	next, err := paginator.Cursor(page.NextCursor, 2, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	badPosition := next
	badPosition.After = append(flop.CursorPosition(nil), next.After...)
	badPosition.After[0].Path = "id"
	shortPosition := next
	shortPosition.After = append(flop.CursorPosition(nil), next.After[:1]...)
	nullPosition := next
	nullPosition.After = append(flop.CursorPosition(nil), next.After...)
	nullPosition.After[0].Value = expr.NullValue()
	singleOrder, err := flop.NewCursorParams(
		"",
		1,
		flop.OrderBy{Fields: []flop.Field{flop.Asc("id")}},
		filter.Filter{},
	)
	assert.NoError(t, err)
	singleOrder.After = flop.CursorPosition{{Path: "id", Value: expr.Int64Value(1)}}
	withoutOrder, err := flop.NewCursorParams("", 1, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	invalidOrder, err := flop.NewCursorParams(
		"",
		1,
		flop.OrderBy{Fields: []flop.Field{{Path: "bad-name"}}},
		filter.Filter{},
	)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		base     sq.SelectBuilder
		params   flop.CursorParams
		wantSQL  string
		wantArgs []any
		wantErr  bool
	}{
		{
			name: "first page", base: psql.Select("id").From("orders").Where(sq.Eq{"tenant_id": 7}), params: first,
			wantSQL: "SELECT id FROM orders WHERE tenant_id = $1 ORDER BY created_at DESC, id ASC LIMIT 3", wantArgs: []any{7},
		},
		{
			name: "seek with mixed directions", base: psql.Select("id").From("orders").Where(sq.Eq{"tenant_id": 7}), params: next,
			wantSQL:  "SELECT id FROM orders WHERE tenant_id = $1 AND (created_at < $2 OR (created_at = $3 AND id > $4)) ORDER BY created_at DESC, id ASC LIMIT 3",
			wantArgs: []any{7, now.Add(-time.Second), now.Add(-time.Second), int64(2)},
		},
		{name: "base order", base: psql.Select("id").From("orders").OrderBy("id"), params: first, wantErr: true},
		{name: "missing order", base: psql.Select("id").From("orders"), params: withoutOrder, wantErr: true},
		{name: "invalid order column", base: psql.Select("id").From("orders"), params: invalidOrder, wantErr: true},
		{name: "position mismatch", base: psql.Select("id").From("orders"), params: badPosition, wantErr: true},
		{
			name:    "position length mismatch",
			base:    psql.Select("id").From("orders"),
			params:  shortPosition,
			wantErr: true,
		},
		{name: "null position", base: psql.Select("id").From("orders"), params: nullPosition, wantErr: true},
		{
			name:     "single order field",
			base:     psql.Select("id").From("orders"),
			params:   singleOrder,
			wantSQL:  "SELECT id FROM orders WHERE id > $1 ORDER BY id ASC LIMIT 2",
			wantArgs: []any{int64(1)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			gotSQL, gotArgs, err := Cursor(test.base, test.params).ToSql()
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.wantSQL, gotSQL)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}

func TestPage(t *testing.T) {
	t.Parallel()
	paginator := queryPaginator(t)
	params, err := paginator.Offset(3, 20, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	firstPage, err := paginator.Offset(1, 20, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	withoutOrder, err := flop.NewParams(1, 10)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		base    sq.SelectBuilder
		params  flop.Params
		wantSQL string
		wantErr bool
	}{
		{
			name:    "offset page",
			base:    psql.Select("id").From("orders"),
			params:  params,
			wantSQL: "SELECT id FROM orders ORDER BY created_at DESC, id ASC LIMIT 20 OFFSET 40",
		},
		{
			name:    "first page",
			base:    psql.Select("id").From("orders"),
			params:  firstPage,
			wantSQL: "SELECT id FROM orders ORDER BY created_at DESC, id ASC LIMIT 20",
		},
		{name: "base order", base: psql.Select("id").From("orders").OrderBy("id"), params: params, wantErr: true},
		{name: "invalid parameters", base: psql.Select("id").From("orders"), params: flop.Params{}, wantErr: true},
		{name: "missing order", base: psql.Select("id").From("orders"), params: withoutOrder, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			gotSQL, _, err := Page(test.base, test.params).ToSql()
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.wantSQL, gotSQL)
		})
	}
}

func TestCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		base     sq.SelectBuilder
		wantSQL  string
		wantArgs []any
		wantErr  bool
	}{
		{
			name: "count", base: psql.Select("id").From("orders").Where(sq.Eq{"tenant_id": 7}).Limit(20).Offset(40),
			wantSQL: "SELECT count(*) FROM orders WHERE tenant_id = $1", wantArgs: []any{7},
		},
		{name: "base order", base: psql.Select("id").From("orders").OrderBy("id"), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			gotSQL, gotArgs, err := Count(test.base).ToSql()
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.wantSQL, gotSQL)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}
