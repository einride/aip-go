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
		wantErr bool
	}{
		{
			name: "zero page selects first page",
			args: args{pageSize: 10},
			want: want{pageNumber: 1, offset: 0, limit: 10},
		},
		{
			name: "requested page",
			args: args{pageNumber: 3, pageSize: 20},
			want: want{pageNumber: 3, offset: 40, limit: 20},
		},
		{
			name: "page size is not clamped",
			args: args{pageNumber: 2, pageSize: 101},
			want: want{pageNumber: 2, offset: 101, limit: 101},
		},
		{
			name: "maximum page and size do not overflow int64",
			args: args{pageNumber: math.MaxInt32, pageSize: math.MaxInt32},
			want: want{
				pageNumber: math.MaxInt32,
				offset:     int64(math.MaxInt32-1) * int64(math.MaxInt32),
				limit:      math.MaxInt32,
			},
		},
		{
			name:    "negative page number",
			args:    args{pageNumber: -1, pageSize: 10},
			wantErr: true,
		},
		{name: "zero page size", args: args{pageNumber: 1}, wantErr: true},
		{
			name:    "negative page size",
			args:    args{pageNumber: 1, pageSize: -1},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, gotErr := flop.NewOffsetParams(test.args.pageNumber, test.args.pageSize)
			assert.Equal(t, test.want.pageNumber, got.Page())
			assert.Equal(t, test.want.offset, got.Offset())
			assert.Equal(t, test.want.limit, got.Limit())
			if test.wantErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
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
		wantErr bool
	}{
		{
			name: "middle page",
			args: args{rows: []int32{3, 4}, params: secondPageParams, totalItems: 5},
			want: flop.OffsetPage[int32]{
				Items: []int32{3, 4}, Number: 2,
				TotalPages: 3, TotalItems: 5,
			},
		},
		{
			name: "last page",
			args: args{rows: []int32{5}, params: thirdPageParams, totalItems: 5},
			want: flop.OffsetPage[int32]{
				Items: []int32{5}, Number: 3,
				TotalPages: 3, TotalItems: 5,
			},
		},
		{
			name: "page beyond end",
			args: args{rows: []int32{}, params: thirdPageParams, totalItems: 0},
			want: flop.OffsetPage[int32]{
				Items: []int32{}, Number: 3,
				TotalPages: 0, TotalItems: 0,
			},
		},
		{
			name:    "too many rows",
			args:    args{rows: []int32{1, 2, 3}, params: secondPageParams, totalItems: 5},
			wantErr: true,
		},
		{
			name:    "negative total",
			args:    args{params: secondPageParams, totalItems: -1},
			wantErr: true,
		},
		{
			name:    "zero parameters",
			args:    args{params: flop.OffsetParams{}},
			wantErr: true,
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
			if test.wantErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
		})
	}
}
