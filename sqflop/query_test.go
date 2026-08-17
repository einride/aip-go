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
	first, err := paginator.CursorParams("", 2, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	now := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	page, err := paginator.CursorPage([]queryRow{
		{ID: 1, CreatedAt: now},
		{ID: 2, CreatedAt: now.Add(-time.Second)},
		{ID: 3, CreatedAt: now.Add(-2 * time.Second)},
	}, first)
	assert.NoError(t, err)
	next, err := paginator.CursorParams(page.NextCursor, 2, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	singleOrder := singleOrderParams(t)
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

// singleOrderParams round-trips a cursor through a paginator ordered by a
// single key column, so params carry a one-field position.
func singleOrderParams(t *testing.T) flop.CursorParams {
	t.Helper()
	paginator, err := flop.NewPaginator(
		[]flop.Col[queryRow]{
			flop.Key("id", func(row queryRow) expr.Value { return expr.Int64Value(row.ID) }),
		},
		filter.Disabled(),
		[]flop.Field{flop.Asc("id")},
		1,
		100,
	)
	assert.NoError(t, err)
	first, err := paginator.CursorParams("", 1, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	page, err := paginator.CursorPage([]queryRow{{ID: 1}, {ID: 2}}, first)
	assert.NoError(t, err)
	params, err := paginator.CursorParams(page.NextCursor, 1, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	return params
}

// Paginator never produces a position inconsistent with its order, so these
// defenses are exercised against cursorPredicate directly.
func TestCursorPredicate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	orderBy := flop.OrderBy{Fields: []flop.Field{flop.Desc("created_at"), flop.Asc("id")}}

	tests := []struct {
		name     string
		position flop.CursorPosition
	}{
		{
			name: "position mismatch",
			position: flop.CursorPosition{
				{Path: "id", Value: expr.TimeValue(now)},
				{Path: "id", Value: expr.Int64Value(2)},
			},
		},
		{
			name: "position length mismatch",
			position: flop.CursorPosition{
				{Path: "created_at", Value: expr.TimeValue(now)},
			},
		},
		{
			name: "null position",
			position: flop.CursorPosition{
				{Path: "created_at", Value: expr.NullValue()},
				{Path: "id", Value: expr.Int64Value(2)},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := cursorPredicate(orderBy, test.position)
			assert.Error(t, err)
		})
	}
}

func TestPage(t *testing.T) {
	t.Parallel()
	paginator := queryPaginator(t)
	params, err := paginator.OffsetParams(3, 20, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	firstPage, err := paginator.OffsetParams(1, 20, flop.OrderBy{}, filter.Filter{})
	assert.NoError(t, err)
	withoutOrder, err := flop.NewOffsetParams(1, 10)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		base    sq.SelectBuilder
		params  flop.OffsetParams
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
		{name: "invalid parameters", base: psql.Select("id").From("orders"), params: flop.OffsetParams{}, wantErr: true},
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
