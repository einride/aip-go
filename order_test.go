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
		wantErr bool
	}{
		{name: "empty", args: args{input: ""}, want: flop.OrderBy{}},
		{name: "whitespace only", args: args{input: " \n\t"}, want: flop.OrderBy{}},
		{
			name: "fields",
			args: args{input: "create_time desc, user.name asc, id"},
			want: flop.OrderBy{Fields: []flop.Field{
				{Path: "create_time", Desc: true},
				{Path: "user.name"},
				{Path: "id"},
			}},
		},
		{
			name: "explicit ascending and whitespace",
			args: args{input: "  create_time   asc ,\tuser_id desc  "},
			want: flop.OrderBy{Fields: []flop.Field{
				{Path: "create_time"},
				{Path: "user_id", Desc: true},
			}},
		},
		{name: "duplicate", args: args{input: "id, id desc"}, wantErr: true},
		{name: "empty field", args: args{input: "id,,name"}, wantErr: true},
		{name: "trailing comma", args: args{input: "id,"}, wantErr: true},
		{name: "invalid path", args: args{input: "user..name"}, wantErr: true},
		{name: "path starts with number", args: args{input: "1st_name"}, wantErr: true},
		{name: "non-ascii path", args: args{input: "имя"}, wantErr: true},
		{name: "invalid direction", args: args{input: "id DESC"}, wantErr: true},
		{name: "too many words", args: args{input: "id asc desc"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := flop.ParseOrderBy(test.args.input)
			assert.Equal(t, test.want, got)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
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
