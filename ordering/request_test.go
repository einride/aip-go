package ordering

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseOrderBy(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		r := mockRequest{orderBy: "foo asc,bar desc"}
		expected := OrderBy{
			Fields: []Field{
				{Path: "foo"},
				{Path: "bar", Desc: true},
			},
		}
		actual, err := ParseOrderBy(r)
		assert.NilError(t, err)
		assert.DeepEqual(t, expected, actual)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		r := mockRequest{orderBy: "/foo"}
		actual, err := ParseOrderBy(r)
		assert.ErrorContains(t, err, "invalid character '/'")
		assert.DeepEqual(t, OrderBy{}, actual)
	})
}

type mockRequest struct {
	orderBy string
}

var _ Request = &mockRequest{}

func (m mockRequest) GetOrderBy() string {
	return m.orderBy
}
