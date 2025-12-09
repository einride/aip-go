package filtering

import (
	"testing"

	"github.com/google/cel-go/cel"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
)

func TestProtoDeclarations(t *testing.T) {
	t.Parallel()

	protoMsg := fullProtobufMessage(t)

	testCases := []struct {
		name            string
		opts            []FilterOption
		filter          string
		expectedExpr    *expr.Expr
		expectedExprStr string
		expectError     bool
	}{
		// String fields
		{
			name:            "ok - string field",
			opts:            []FilterOption{WithFilterableFields("string_field")},
			filter:          `string_field = "test"`,
			expectedExpr:    Equals(Text("string_field"), String("test")),
			expectedExprStr: "=(string_field, \"test\")",
			expectError:     false,
		},
		// Boolean fields
		{
			name:            "ok - bool field true",
			opts:            []FilterOption{WithFilterableFields("bool_field")},
			filter:          `bool_field`,
			expectedExpr:    Text("bool_field"),
			expectedExprStr: "bool_field",
			expectError:     false,
		},
		{
			name:            "ok - bool field false",
			opts:            []FilterOption{WithFilterableFields("bool_field")},
			filter:          `NOT bool_field`,
			expectedExpr:    Not(Text("bool_field")),
			expectedExprStr: "NOT(bool_field)",
			expectError:     false,
		},
		// Integer fields (various sizes)
		{
			name:            "ok - int32 field",
			opts:            []FilterOption{WithFilterableFields("int32_field")},
			filter:          `int32_field = 42`,
			expectedExpr:    Equals(Text("int32_field"), Int(42)),
			expectedExprStr: "=(int32_field, 42)",
			expectError:     false,
		},
		{
			name:            "ok - int64 field",
			opts:            []FilterOption{WithFilterableFields("int64_field")},
			expectedExpr:    GreaterThan(Text("int64_field"), Int(100)),
			filter:          `int64_field > 100`,
			expectedExprStr: ">(int64_field, 100)",
			expectError:     false,
		},
		{
			name:            "ok - sint32 field",
			opts:            []FilterOption{WithFilterableFields("sint32_field")},
			filter:          `sint32_field < -10`,
			expectedExpr:    LessThan(Text("sint32_field"), Int(-10)),
			expectedExprStr: "<(sint32_field, -10)",
			expectError:     false,
		},
		{
			name:            "ok - sint64 field",
			opts:            []FilterOption{WithFilterableFields("sint64_field")},
			filter:          `sint64_field >= -1000`,
			expectedExpr:    GreaterEquals(Text("sint64_field"), Int(-1000)),
			expectedExprStr: ">=(sint64_field, -1000)",
			expectError:     false,
		},
		{
			name:            "ok - sfixed32 field",
			opts:            []FilterOption{WithFilterableFields("sfixed32_field")},
			filter:          `sfixed32_field = 123`,
			expectedExpr:    Equals(Text("sfixed32_field"), Int(123)),
			expectedExprStr: "=(sfixed32_field, 123)",
			expectError:     false,
		},
		{
			name:            "ok - sfixed64 field",
			opts:            []FilterOption{WithFilterableFields("sfixed64_field")},
			filter:          `sfixed64_field <= 456`,
			expectedExpr:    LessEquals(Text("sfixed64_field"), Int(456)),
			expectedExprStr: "<=(sfixed64_field, 456)",
			expectError:     false,
		},
		// Unsigned integer fields
		{
			name:            "ok - uint32 field",
			opts:            []FilterOption{WithFilterableFields("uint32_field")},
			filter:          `uint32_field = 789`,
			expectedExpr:    Equals(Text("uint32_field"), Int(789)),
			expectedExprStr: "=(uint32_field, 789)",
			expectError:     false,
		},
		{
			name:            "ok - uint64 field",
			opts:            []FilterOption{WithFilterableFields("uint64_field")},
			filter:          `uint64_field > 1000`,
			expectedExpr:    GreaterThan(Text("uint64_field"), Int(1000)),
			expectedExprStr: ">(uint64_field, 1000)",
			expectError:     false,
		},
		{
			name:            "ok - fixed32 field",
			opts:            []FilterOption{WithFilterableFields("fixed32_field")},
			filter:          `fixed32_field < 2000`,
			expectedExpr:    LessThan(Text("fixed32_field"), Int(2000)),
			expectedExprStr: "<(fixed32_field, 2000)",
			expectError:     false,
		},
		{
			name:            "ok - fixed64 field",
			opts:            []FilterOption{WithFilterableFields("fixed64_field")},
			filter:          `fixed64_field >= 3000`,
			expectedExpr:    GreaterEquals(Text("fixed64_field"), Int(3000)),
			expectedExprStr: ">=(fixed64_field, 3000)",
			expectError:     false,
		},
		// Float fields
		{
			name:            "ok - float field",
			opts:            []FilterOption{WithFilterableFields("float_field")},
			filter:          `float_field > 3.14`,
			expectedExpr:    GreaterThan(Text("float_field"), Float(3.14)),
			expectedExprStr: ">(float_field, 3.14)",
			expectError:     false,
		},
		{
			name:            "ok - double field",
			opts:            []FilterOption{WithFilterableFields("double_field")},
			filter:          `double_field <= 2.71`,
			expectedExpr:    LessEquals(Text("double_field"), Float(2.71)),
			expectedExprStr: "<=(double_field, 2.71)",
			expectError:     false,
		},
		// Enum field
		{
			name:            "ok - enum field",
			opts:            []FilterOption{WithFilterableFields("enum_field")},
			filter:          `enum_field = ENUM_VALUE_ONE`,
			expectedExpr:    Equals(Text("enum_field"), Text("ENUM_VALUE_ONE")),
			expectedExprStr: "=(enum_field, ENUM_VALUE_ONE)",
			expectError:     false,
		},
		// Timestamp field (well-known type)
		{
			name:            "ok - timestamp field",
			opts:            []FilterOption{WithFilterableFields("timestamp_field")},
			filter:          `timestamp_field > "2023-01-01T00:00:00Z"`,
			expectedExpr:    GreaterThan(Text("timestamp_field"), String("2023-01-01T00:00:00Z")),
			expectedExprStr: ">(timestamp_field, \"2023-01-01T00:00:00Z\")",
			expectError:     false,
		},
		// Nested message field
		{
			name:            "ok - nested message field",
			opts:            []FilterOption{WithFilterableFields("nested_message.nested_string")},
			filter:          `nested_message.nested_string = "nested_value"`,
			expectedExpr:    Equals(Member(Text("nested_message"), "nested_string"), String("nested_value")),
			expectedExprStr: "=(nested_message.nested_string, \"nested_value\")",
			expectError:     false,
		},
		// Deeply nested field
		{
			name:            "ok - deeply nested field",
			opts:            []FilterOption{WithFilterableFields("nested_message.deep_nested.deep_string")},
			filter:          `nested_message.deep_nested.deep_string = "deep_value"`,
			expectedExpr:    Equals(Member(Member(Text("nested_message"), "deep_nested"), "deep_string"), String("deep_value")),
			expectedExprStr: "=(nested_message.deep_nested.deep_string, \"deep_value\")",
			expectError:     false,
		},
		{
			name:            "ok - deeply nested field, filter by parent field",
			opts:            []FilterOption{WithFilterableFields("nested_message.deep_nested")},
			filter:          `nested_message.deep_nested.deep_string = "deep_value"`,
			expectedExpr:    Equals(Member(Member(Text("nested_message"), "deep_nested"), "deep_string"), String("deep_value")),
			expectedExprStr: "=(nested_message.deep_nested.deep_string, \"deep_value\")",
			expectError:     false,
		},
		{
			name:            "ok - deeply nested field, filter by parent parent",
			opts:            []FilterOption{WithFilterableFields("nested_message")},
			filter:          `nested_message.deep_nested.deep_string = "deep_value"`,
			expectedExpr:    Equals(Member(Member(Text("nested_message"), "deep_nested"), "deep_string"), String("deep_value")),
			expectedExprStr: "=(nested_message.deep_nested.deep_string, \"deep_value\")",
			expectError:     false,
		},
		// Complex expressions
		{
			name:   "ok - multiple field types",
			opts:   []FilterOption{WithFilterableFields("string_field", "int32_field", "bool_field")},
			filter: `string_field = "test" AND int32_field > 10 AND bool_field`,
			expectedExpr: And(
				Equals(Text("string_field"), String("test")),
				GreaterThan(Text("int32_field"), Int(10)),
				Text("bool_field"),
			),
			expectedExprStr: "AND(AND(=(string_field, \"test\"), >(int32_field, 10)), bool_field)",
			expectError:     false,
		},
		// Unsupported fields (these should be skipped during declaration)
		{
			name:        "error - list field should not be declared",
			opts:        []FilterOption{WithFilterableFields("string_list")},
			filter:      `string_list = "test"`,
			expectError: true,
		},
		{
			name:        "error -map field should not be declared",
			opts:        []FilterOption{WithFilterableFields("string_map")},
			filter:      `string_map = "test"`,
			expectError: true,
		},
		{
			name:            "error - no filterable fields option provided",
			opts:            []FilterOption{},
			filter:          `string_field = "test"`,
			expectedExprStr: "=(string_field, \"test\")",
			expectError:     true,
		},
		{
			name:            "error - empty filterable fields explicitly specified (no fields available)",
			opts:            []FilterOption{WithFilterableFields()},
			filter:          `string_field = "test"`,
			expectedExprStr: "=(string_field, \"test\")",
			expectError:     true,
		},
		{
			name:            "error - field not in filterable fields list",
			opts:            []FilterOption{WithFilterableFields("bool_field")},
			filter:          `string_field = "test"`,
			expectedExprStr: "=(string_field, \"test\")",
			expectError:     true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Given
			// When
			declarations, err := ProtoDeclarations(protoMsg, tt.opts...)

			// Then
			assert.NilError(t, err)
			assert.Assert(t, declarations != nil)

			// Create a mock request with the test filter
			req := &mockRequest{
				filter: tt.filter,
			}

			// Parse the filter using our declarations
			f, err := ParseFilter(req, declarations)

			if tt.expectError {
				assert.Assert(t, err != nil, "expected error for filter: %s", tt.filter)
				return
			}

			assert.NilError(t, err, "unexpected error for filter: %s", tt.filter)

			if tt.expectedExpr != nil {
				assert.DeepEqual(
					t,
					tt.expectedExpr,
					f.CheckedExpr.GetExpr(),
					protocmp.Transform(),
					protocmp.IgnoreFields(&expr.Expr{}, "id"),
				)
			}

			// Convert to string for comparison
			outExpr, err := cel.AstToString(cel.CheckedExprToAst(f.CheckedExpr))
			assert.NilError(t, err)
			assert.Equal(t, tt.expectedExprStr, outExpr)
		})
	}
}

type mockRequest struct {
	filter string
}

func (m *mockRequest) GetFilter() string {
	return m.filter
}
