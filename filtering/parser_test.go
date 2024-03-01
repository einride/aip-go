package filtering

import (
	"testing"
	"time"

	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
)

func TestParser(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		expected      *expr.Expr
		errorContains string
	}{
		{
			filter:   "New York Giants",
			expected: Sequence(Text("New"), Text("York"), Text("Giants")),
		},

		{
			filter:   "New York Giants OR Yankees",
			expected: Sequence(Text("New"), Text("York"), Or(Text("Giants"), Text("Yankees"))),
		},

		{
			filter:   "New York (Giants OR Yankees)",
			expected: Sequence(Text("New"), Text("York"), Or(Text("Giants"), Text("Yankees"))),
		},

		{
			filter: "a b AND c AND d",
			expected: And(
				Sequence(Text("a"), Text("b")),
				Text("c"),
				Text("d"),
			),
		},

		{
			filter: "(a b) AND c AND d",
			expected: And(
				Sequence(Text("a"), Text("b")),
				Text("c"),
				Text("d"),
			),
		},

		{
			filter: "a < 10 OR a >= 100",
			expected: Or(
				LessThan(Text("a"), Int(10)),
				GreaterEquals(Text("a"), Int(100)),
			),
		},
		{
			filter: "a OR b OR c",
			expected: Or(
				Text("a"),
				Text("b"),
				Text("c"),
			),
		},

		{
			filter:   "NOT (a OR b)",
			expected: Not(Or(Text("a"), Text("b"))),
		},

		{
			filter:   `-file:".java"`,
			expected: Not(Has(Text("file"), String(".java"))),
		},

		{
			filter:   "-30",
			expected: Int(-30),
		},

		{
			filter:   "package=com.google",
			expected: Equals(Text("package"), Member(Text("com"), "google")),
		},

		{
			filter:   `msg != 'hello'`,
			expected: NotEquals(Text("msg"), String("hello")),
		},

		{
			filter:   `1 > 0`,
			expected: GreaterThan(Int(1), Int(0)),
		},

		{
			filter:   `2.5 >= 2.4`,
			expected: GreaterEquals(Float(2.5), Float(2.4)),
		},

		{
			filter:   `foo >= -2.4`,
			expected: GreaterEquals(Text("foo"), Float(-2.4)),
		},

		{
			filter:   `foo >= (-2.4)`,
			expected: GreaterEquals(Text("foo"), Float(-2.4)),
		},

		{
			filter:   `-2.5 >= -2.4`,
			expected: Not(GreaterEquals(Float(2.5), Float(-2.4))),
		},

		{
			filter:   `yesterday < request.time`,
			expected: LessThan(Text("yesterday"), Member(Text("request"), "time")),
		},

		{
			filter: `experiment.rollout <= cohort(request.user)`,
			expected: LessEquals(
				Member(Text("experiment"), "rollout"),
				Function("cohort", Member(Text("request"), "user")),
			),
		},

		{
			filter:   `prod`,
			expected: Text("prod"),
		},

		{
			filter:   `expr.type_map.1.type`,
			expected: Member(Member(Member(Text("expr"), "type_map"), "1"), "type"),
		},

		{
			filter:   `regex(m.key, '^.*prod.*$')`,
			expected: Function("regex", Member(Text("m"), "key"), String("^.*prod.*$")),
		},

		{
			filter:   `math.mem('30mb')`,
			expected: Function("math.mem", String("30mb")),
		},

		{
			filter: `(msg.endsWith('world') AND retries < 10)`,
			expected: And(
				Function("msg.endsWith", String("world")),
				LessThan(Text("retries"), Int(10)),
			),
		},

		{
			filter: `(endsWith(msg, 'world') AND retries < 10)`,
			expected: And(
				Function("endsWith", Text("msg"), String("world")),
				LessThan(Text("retries"), Int(10)),
			),
		},

		{
			filter:   "time.now()",
			expected: Function("time.now"),
		},

		{
			filter: `timestamp("2012-04-21T11:30:00-04:00")`,
			expected: Timestamp(
				time.Date(2012, time.April, 21, 11, 30, 0, 0, time.FixedZone("EST", -4*int(time.Hour.Seconds()))),
			),
		},

		{
			filter:   `duration("32s")`,
			expected: Duration(32 * time.Second),
		},

		{
			filter:   `duration("4h0m0s")`,
			expected: Duration(4 * time.Hour),
		},

		{
			filter: `
				start_time > timestamp("2006-01-02T15:04:05+07:00") AND 
				(driver = "driver1" OR start_driver = "driver1" OR end_driver = "driver1")
			`,
			expected: And(
				GreaterThan(
					Text("start_time"),
					Timestamp(
						time.Date(
							2006, time.January, 2, 15, 4, 5, 0,
							time.FixedZone("EST", 7*int(time.Hour.Seconds())),
						),
					),
				),
				Or(
					Equals(Text("driver"), String("driver1")),
					Equals(Text("start_driver"), String("driver1")),
					Equals(Text("end_driver"), String("driver1")),
				),
			),
		},

		{
			filter:   `annotations:schedule`,
			expected: Has(Text("annotations"), String("schedule")),
		},

		{
			filter:   `annotations.schedule = "test"`,
			expected: Equals(Member(Text("annotations"), "schedule"), String("test")),
		},

		{
			filter:        "<",
			errorContains: "unexpected token <",
		},

		{
			filter:        `(-2.5) >= -2.4`,
			errorContains: "unexpected token >=",
		},

		{
			filter:        `a = "foo`,
			errorContains: "unterminated string",
		},

		{
			filter:        "invalid = foo\xa0\x01bar",
			errorContains: "invalid UTF-8",
		},
	} {
		tt := tt
		t.Run(tt.filter, func(t *testing.T) {
			t.Parallel()
			var parser Parser
			parser.Init(tt.filter)
			actual, err := parser.Parse()
			if tt.errorContains != "" {
				if actual != nil {
					t.Log(actual.GetExpr().String())
				}
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(
					t,
					tt.expected,
					actual.GetExpr(),
					protocmp.Transform(),
					protocmp.IgnoreFields(&expr.Expr{}, "id"),
				)
				assertUniqueExprIDs(t, actual.GetExpr())
			}
		})
	}
}

func assertUniqueExprIDs(t *testing.T, exp *expr.Expr) {
	t.Helper()
	seenIDs := make(map[int64]struct{})
	Walk(func(currExpr, _ *expr.Expr) bool {
		if _, ok := seenIDs[currExpr.GetId()]; ok {
			t.Fatalf("duplicate expression ID '%d' for expr %v", currExpr.GetId(), currExpr)
		}
		seenIDs[currExpr.GetId()] = struct{}{}
		return true
	}, exp)
}
