package sqflop

import (
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
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
		[]flop.Column[queryRow]{
			flop.Sortable("created_at", func(row queryRow) expr.Value { return expr.TimeValue(row.CreatedAt) }),
			flop.Key("id", func(row queryRow) expr.Value { return expr.Int64Value(row.ID) }),
		},
		flop.DisabledFilter(),
		[]flop.OrderField{flop.Desc("created_at")},
		2,
		100,
	)
	assert.NoError(t, err)
	return paginator
}

func TestCursorQuery(t *testing.T) {
	t.Parallel()
	paginator := queryPaginator(t)
	first, err := paginator.CursorParams("", 2, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	now := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	page, err := paginator.CursorPage([]queryRow{
		{ID: 1, CreatedAt: now},
		{ID: 2, CreatedAt: now.Add(-time.Second)},
		{ID: 3, CreatedAt: now.Add(-2 * time.Second)},
	}, first)
	assert.NoError(t, err)
	next, err := paginator.CursorParams(page.NextCursor, 2, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	singleOrder := singleOrderParams(t)
	withoutOrder, err := flop.NewCursorParams("", 1, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	invalidOrder, err := flop.NewCursorParams(
		"",
		1,
		flop.OrderBy{Fields: []flop.OrderField{{Path: "bad-name"}}},
		flop.Filter{},
	)
	assert.NoError(t, err)
	mismatchedPosition := mismatchedPositionParams(t, now)

	type args struct {
		base   sq.SelectBuilder
		params flop.CursorParams
	}
	tests := []struct {
		name     string
		args     args
		wantSQL  string
		wantArgs []any
		wantErr  assert.ErrorFunc
	}{
		{
			name: "first page",
			args: args{base: psql.Select("id").From("orders").Where(sq.Eq{"tenant_id": 7}), params: first},
			wantSQL: "SELECT id FROM orders WHERE tenant_id = $1 " +
				"ORDER BY created_at DESC, id ASC LIMIT 3",
			wantArgs: []any{7},
			wantErr:  assert.NoError,
		},
		{
			name: "seek with mixed directions",
			args: args{base: psql.Select("id").From("orders").Where(sq.Eq{"tenant_id": 7}), params: next},
			wantSQL: "SELECT id FROM orders WHERE tenant_id = $1 " +
				"AND (created_at < $2 OR (created_at = $3 AND id > $4)) " +
				"ORDER BY created_at DESC, id ASC LIMIT 3",
			wantArgs: []any{7, now.Add(-time.Second), now.Add(-time.Second), int64(2)},
			wantErr:  assert.NoError,
		},
		{
			name:    "base order",
			args:    args{base: psql.Select("id").From("orders").OrderBy("id"), params: first},
			wantErr: assert.Error,
		},
		{
			name:    "missing order",
			args:    args{base: psql.Select("id").From("orders"), params: withoutOrder},
			wantErr: assert.Error,
		},
		{
			name:    "invalid order column",
			args:    args{base: psql.Select("id").From("orders"), params: invalidOrder},
			wantErr: assert.Error,
		},
		{
			name:    "cursor position mismatch",
			args:    args{base: psql.Select("id").From("orders"), params: mismatchedPosition},
			wantErr: assert.Error,
		},
		{
			name:     "single order field",
			args:     args{base: psql.Select("id").From("orders"), params: singleOrder},
			wantSQL:  "SELECT id FROM orders WHERE id > $1 ORDER BY id ASC LIMIT 2",
			wantArgs: []any{int64(1)},
			wantErr:  assert.NoError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			gotSQL, gotArgs, err := CursorQuery(test.args.base, test.args.params).ToSql()
			test.wantErr(t, err)
			if err != nil {
				return
			}
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
		[]flop.Column[queryRow]{
			flop.Key("id", func(row queryRow) expr.Value { return expr.Int64Value(row.ID) }),
		},
		flop.DisabledFilter(),
		[]flop.OrderField{flop.Asc("id")},
		1,
		100,
	)
	assert.NoError(t, err)
	first, err := paginator.CursorParams("", 1, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	page, err := paginator.CursorPage([]queryRow{{ID: 1}, {ID: 2}}, first)
	assert.NoError(t, err)
	params, err := paginator.CursorParams(page.NextCursor, 1, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	return params
}

// mismatchedPositionParams returns parameters whose position does not match
// their order. Paginator.CursorParams validates the two against each other, but
// NewCursorParams only checks the binding, which covers the order and filter and
// not the position, so the mismatch survives a cursor round trip.
func mismatchedPositionParams(t *testing.T, now time.Time) flop.CursorParams {
	t.Helper()
	orderBy := flop.OrderBy{Fields: []flop.OrderField{flop.Asc("id")}}
	params, err := flop.NewCursorParams("", 1, orderBy, flop.Filter{})
	assert.NoError(t, err)
	page, err := flop.NewCursorPage(
		[]queryRow{{ID: 1}, {ID: 2}},
		params,
		func(queryRow) (flop.CursorPosition, error) {
			return flop.CursorPosition{
				{Path: "created_at", Value: expr.TimeValue(now)},
				{Path: "id", Value: expr.Int64Value(1)},
			}, nil
		},
	)
	assert.NoError(t, err)
	mismatched, err := flop.NewCursorParams(page.NextCursor, 1, orderBy, flop.Filter{})
	assert.NoError(t, err)
	return mismatched
}

func Test_cursorPredicate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	orderBy := flop.OrderBy{Fields: []flop.OrderField{flop.Desc("created_at"), flop.Asc("id")}}

	type args struct {
		orderBy flop.OrderBy
		pos     flop.CursorPosition
	}
	tests := []struct {
		name     string
		args     args
		wantSQL  string
		wantArgs []any
		wantErr  assert.ErrorFunc
	}{
		{
			name: "single ascending field",
			args: args{
				orderBy: flop.OrderBy{Fields: []flop.OrderField{flop.Asc("id")}},
				pos:     flop.CursorPosition{{Path: "id", Value: expr.Int64Value(2)}},
			},
			wantSQL:  "id > ?",
			wantArgs: []any{int64(2)},
			wantErr:  assert.NoError,
		},
		{
			name: "single descending field",
			args: args{
				orderBy: flop.OrderBy{Fields: []flop.OrderField{flop.Desc("created_at")}},
				pos:     flop.CursorPosition{{Path: "created_at", Value: expr.TimeValue(now)}},
			},
			wantSQL:  "created_at < ?",
			wantArgs: []any{now},
			wantErr:  assert.NoError,
		},
		{
			name: "mixed directions",
			args: args{
				orderBy: orderBy,
				pos: flop.CursorPosition{
					{Path: "created_at", Value: expr.TimeValue(now)},
					{Path: "id", Value: expr.Int64Value(2)},
				},
			},
			wantSQL:  "(created_at < ? OR (created_at = ? AND id > ?))",
			wantArgs: []any{now, now, int64(2)},
			wantErr:  assert.NoError,
		},
		{
			name: "three fields",
			args: args{
				orderBy: flop.OrderBy{Fields: []flop.OrderField{
					flop.Desc("created_at"), flop.Asc("name"), flop.Asc("id"),
				}},
				pos: flop.CursorPosition{
					{Path: "created_at", Value: expr.TimeValue(now)},
					{Path: "name", Value: expr.StringValue("a")},
					{Path: "id", Value: expr.Int64Value(2)},
				},
			},
			wantSQL: "(created_at < ? OR (created_at = ? AND name > ?) " +
				"OR (created_at = ? AND name = ? AND id > ?))",
			wantArgs: []any{now, now, "a", now, "a", int64(2)},
			wantErr:  assert.NoError,
		},
		{
			name: "position mismatch",
			args: args{
				orderBy: orderBy,
				pos: flop.CursorPosition{
					{Path: "id", Value: expr.TimeValue(now)},
					{Path: "id", Value: expr.Int64Value(2)},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "position length mismatch",
			args: args{
				orderBy: orderBy,
				pos: flop.CursorPosition{
					{Path: "created_at", Value: expr.TimeValue(now)},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "null position",
			args: args{
				orderBy: orderBy,
				pos: flop.CursorPosition{
					{Path: "created_at", Value: expr.NullValue()},
					{Path: "id", Value: expr.Int64Value(2)},
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := cursorPredicate(test.args.orderBy, test.args.pos)
			test.wantErr(t, err)
			if err != nil {
				return
			}
			gotSQL, gotArgs, err := got.ToSql()
			assert.NoError(t, err)
			assert.Equal(t, test.wantSQL, gotSQL)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}

// OffsetParams carries an order only through Paginator.OffsetParams, which
// resolves it first, so applyOrder cannot fail here. That branch stays as a
// defense and is the one uncovered block in the package.
func TestOffsetQuery(t *testing.T) {
	t.Parallel()
	paginator := queryPaginator(t)
	params, err := paginator.OffsetParams(3, 20, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	firstPage, err := paginator.OffsetParams(1, 20, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)
	withoutOrder, err := flop.NewOffsetParams(1, 10, flop.OrderBy{}, flop.Filter{})
	assert.NoError(t, err)

	type args struct {
		base   sq.SelectBuilder
		params flop.OffsetParams
	}
	tests := []struct {
		name     string
		args     args
		wantSQL  string
		wantArgs []any
		wantErr  assert.ErrorFunc
	}{
		{
			name: "base arguments are preserved",
			args: args{
				base:   psql.Select("id").From("orders").Where(sq.Eq{"tenant_id": 7}),
				params: params,
			},
			wantSQL: "SELECT id FROM orders WHERE tenant_id = $1 " +
				"ORDER BY created_at DESC, id ASC LIMIT 20 OFFSET 40",
			wantArgs: []any{7},
			wantErr:  assert.NoError,
		},
		{
			name:    "offset page",
			args:    args{base: psql.Select("id").From("orders"), params: params},
			wantSQL: "SELECT id FROM orders ORDER BY created_at DESC, id ASC LIMIT 20 OFFSET 40",
			wantErr: assert.NoError,
		},
		{
			name:    "first page",
			args:    args{base: psql.Select("id").From("orders"), params: firstPage},
			wantSQL: "SELECT id FROM orders ORDER BY created_at DESC, id ASC LIMIT 20",
			wantErr: assert.NoError,
		},
		{
			name:    "base order",
			args:    args{base: psql.Select("id").From("orders").OrderBy("id"), params: params},
			wantErr: assert.Error,
		},
		{
			name:    "empty parameters",
			args:    args{base: psql.Select("id").From("orders"), params: flop.OffsetParams{}},
			wantErr: assert.Error,
		},
		{
			name:    "missing order",
			args:    args{base: psql.Select("id").From("orders"), params: withoutOrder},
			wantErr: assert.Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			gotSQL, gotArgs, err := OffsetQuery(test.args.base, test.args.params).ToSql()
			test.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.wantSQL, gotSQL)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}

func TestCountQuery(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		base     sq.SelectBuilder
		wantSQL  string
		wantArgs []any
		wantErr  assert.ErrorFunc
	}{
		{
			name: "count", base: psql.Select("id").From("orders").Where(sq.Eq{"tenant_id": 7}).Limit(20).Offset(40),
			wantSQL: "SELECT count(*) FROM orders WHERE tenant_id = $1", wantArgs: []any{7},
			wantErr: assert.NoError,
		},
		{name: "base order", base: psql.Select("id").From("orders").OrderBy("id"), wantErr: assert.Error},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			gotSQL, gotArgs, err := CountQuery(test.base).ToSql()
			test.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.wantSQL, gotSQL)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}
