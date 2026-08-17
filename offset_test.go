package flop_test

import (
	"math"
	"testing"

	"github.com/ysomad/flop"
	"github.com/ysomad/flop/internal/assert"
)

func TestNewOffsetParams(t *testing.T) {
	t.Parallel()
	type args struct {
		pageNumber int32
		pageSize   int32
	}
	type want struct {
		pageNumber int32
		offset     int64
		limit      int32
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr assert.ErrorFunc
	}{
		{
			name:    "zero page selects first page",
			args:    args{pageSize: 10},
			want:    want{pageNumber: 1, offset: 0, limit: 10},
			wantErr: assert.NoError,
		},
		{
			name:    "requested page",
			args:    args{pageNumber: 3, pageSize: 20},
			want:    want{pageNumber: 3, offset: 40, limit: 20},
			wantErr: assert.NoError,
		},
		{
			name:    "page size is not clamped",
			args:    args{pageNumber: 2, pageSize: 101},
			want:    want{pageNumber: 2, offset: 101, limit: 101},
			wantErr: assert.NoError,
		},
		{
			name: "maximum page and size do not overflow int64",
			args: args{pageNumber: math.MaxInt32, pageSize: math.MaxInt32},
			want: want{
				pageNumber: math.MaxInt32,
				offset:     int64(math.MaxInt32-1) * int64(math.MaxInt32),
				limit:      math.MaxInt32,
			},
			wantErr: assert.NoError,
		},
		{
			name:    "negative page number",
			args:    args{pageNumber: -1, pageSize: 10},
			wantErr: assert.Error,
		},
		{name: "zero page size", args: args{pageNumber: 1}, wantErr: assert.Error},
		{
			name:    "negative page size",
			args:    args{pageNumber: 1, pageSize: -1},
			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, gotErr := flop.NewOffsetParams(test.args.pageNumber, test.args.pageSize)
			assert.Equal(t, test.want.pageNumber, got.Page())
			assert.Equal(t, test.want.offset, got.Offset())
			assert.Equal(t, test.want.limit, got.Limit())
			test.wantErr(t, gotErr)
		})
	}
}

func TestNewOffsetPage(t *testing.T) {
	t.Parallel()
	secondPageParams, err := flop.NewOffsetParams(2, 2)
	assert.NoError(t, err)
	thirdPageParams, err := flop.NewOffsetParams(3, 2)
	assert.NoError(t, err)
	type args struct {
		rows       []int32
		params     flop.OffsetParams
		totalItems int64
	}
	tests := []struct {
		name    string
		args    args
		want    flop.OffsetPage[int32]
		wantErr assert.ErrorFunc
	}{
		{
			name: "middle page",
			args: args{rows: []int32{3, 4}, params: secondPageParams, totalItems: 5},
			want: flop.OffsetPage[int32]{
				Items: []int32{3, 4}, Number: 2,
				TotalPages: 3, TotalItems: 5,
			},
			wantErr: assert.NoError,
		},
		{
			name: "last page",
			args: args{rows: []int32{5}, params: thirdPageParams, totalItems: 5},
			want: flop.OffsetPage[int32]{
				Items: []int32{5}, Number: 3,
				TotalPages: 3, TotalItems: 5,
			},
			wantErr: assert.NoError,
		},
		{
			name: "page beyond end",
			args: args{rows: []int32{}, params: thirdPageParams, totalItems: 0},
			want: flop.OffsetPage[int32]{
				Items: []int32{}, Number: 3,
				TotalPages: 0, TotalItems: 0,
			},
			wantErr: assert.NoError,
		},
		{
			name:    "too many rows",
			args:    args{rows: []int32{1, 2, 3}, params: secondPageParams, totalItems: 5},
			wantErr: assert.Error,
		},
		{
			name:    "negative total",
			args:    args{params: secondPageParams, totalItems: -1},
			wantErr: assert.Error,
		},
		{
			name:    "zero parameters",
			args:    args{params: flop.OffsetParams{}},
			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, gotErr := flop.NewOffsetPage(
				test.args.rows,
				test.args.params,
				test.args.totalItems,
			)
			assert.Equal(t, test.want, got)
			test.wantErr(t, gotErr)
		})
	}
}
