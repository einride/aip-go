package flop_test

import (
	"testing"

	"github.com/ysomad/flop"
	"github.com/ysomad/flop/internal/assert"
)

func TestParseOrderBy(t *testing.T) {
	t.Parallel()
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    flop.OrderBy
		wantErr assert.ErrorFunc
	}{
		{name: "empty", args: args{input: ""}, want: flop.OrderBy{}, wantErr: assert.NoError},
		{name: "whitespace only", args: args{input: " \n\t"}, want: flop.OrderBy{}, wantErr: assert.NoError},
		{
			name: "fields",
			args: args{input: "create_time desc, user.name asc, id"},
			want: flop.OrderBy{Fields: []flop.Field{
				{Path: "create_time", Desc: true},
				{Path: "user.name"},
				{Path: "id"},
			}},
			wantErr: assert.NoError,
		},
		{
			name: "explicit ascending and whitespace",
			args: args{input: "  create_time   asc ,\tuser_id desc  "},
			want: flop.OrderBy{Fields: []flop.Field{
				{Path: "create_time"},
				{Path: "user_id", Desc: true},
			}},
			wantErr: assert.NoError,
		},
		{name: "duplicate", args: args{input: "id, id desc"}, wantErr: assert.Error},
		{name: "empty field", args: args{input: "id,,name"}, wantErr: assert.Error},
		{name: "trailing comma", args: args{input: "id,"}, wantErr: assert.Error},
		{name: "invalid path", args: args{input: "user..name"}, wantErr: assert.Error},
		{name: "path starts with number", args: args{input: "1st_name"}, wantErr: assert.Error},
		{name: "non-ascii path", args: args{input: "имя"}, wantErr: assert.Error},
		{name: "invalid direction", args: args{input: "id DESC"}, wantErr: assert.Error},
		{name: "too many words", args: args{input: "id asc desc"}, wantErr: assert.Error},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := flop.ParseOrderBy(test.args.input)
			assert.Equal(t, test.want, got)
			test.wantErr(t, err)
		})
	}
}

func TestOrderBy_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", (flop.OrderBy{}).String())
	assert.Equal(t, "create_time desc, id", (flop.OrderBy{Fields: []flop.Field{
		{Path: "create_time", Desc: true},
		{Path: "id"},
	}}).String())
}

func TestField_SQL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		field flop.Field
		want  string
	}{
		{name: "ascending", field: flop.Asc("created_at"), want: "created_at ASC"},
		{name: "descending", field: flop.Desc("created_at"), want: "created_at DESC"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.field.SQL())
		})
	}
}
