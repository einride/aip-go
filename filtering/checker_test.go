package filtering

import (
	"testing"

	syntaxv1 "go.einride.tech/aip/proto/gen/einride/example/syntax/v1"
	"gotest.tools/v3/assert"
)

func TestChecker(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		filter        string
		declarations  []DeclarationOption
		errorContains string
	}{
		{
			filter: "New York Giants",
			declarations: []DeclarationOption{
				DeclareIdent("New", TypeBool),
				DeclareIdent("York", TypeBool),
				DeclareIdent("Giants", TypeBool),
			},
			errorContains: "undeclared function 'FUZZY'",
		},

		{
			filter: "New York Giants OR Yankees",
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("New", TypeBool),
				DeclareIdent("York", TypeBool),
				DeclareIdent("Giants", TypeBool),
				DeclareIdent("Yankees", TypeBool),
			},
			errorContains: "undeclared function 'FUZZY'",
		},

		{
			filter: "New York (Giants OR Yankees)",
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("New", TypeBool),
				DeclareIdent("York", TypeBool),
				DeclareIdent("Giants", TypeBool),
				DeclareIdent("Yankees", TypeBool),
			},
			errorContains: "undeclared function 'FUZZY'",
		},

		{
			filter: "a b AND c AND d",
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("a", TypeBool),
				DeclareIdent("b", TypeBool),
				DeclareIdent("c", TypeBool),
				DeclareIdent("d", TypeBool),
			},
			errorContains: "undeclared function 'FUZZY'",
		},

		{
			filter: "a",
			declarations: []DeclarationOption{
				DeclareIdent("a", TypeBool),
			},
		},

		{
			filter: "(a b) AND c AND d",
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("a", TypeBool),
				DeclareIdent("b", TypeBool),
				DeclareIdent("c", TypeBool),
				DeclareIdent("d", TypeBool),
			},
			errorContains: "undeclared function 'FUZZY'",
		},

		{
			filter: `author = "Karin Boye" AND NOT read`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("author", TypeString),
				DeclareIdent("read", TypeBool),
			},
		},

		{
			filter: "a < 10 OR a >= 100",
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("a", TypeInt),
			},
		},

		{
			filter: "NOT (a OR b)",
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("a", TypeBool),
				DeclareIdent("b", TypeBool),
			},
		},

		{
			filter: `-file:".java"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("file", TypeString),
			},
		},

		{
			filter:        "-30",
			errorContains: "non-bool result type",
		},

		{
			filter: "package=com.google",
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("package", TypeString),
				DeclareIdent("com", TypeMap(TypeString, TypeString)),
			},
		},

		{
			filter: `msg != 'hello'`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("msg", TypeString),
			},
		},

		{
			filter: `1 > 0`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
			},
		},

		{
			filter: `2.5 >= 2.4`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
			},
		},

		{
			filter: `foo >= -2.4`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("foo", TypeFloat),
			},
		},

		{
			filter: `foo >= (-2.4)`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("foo", TypeFloat),
			},
		},

		{
			filter: `-2.5 >= -2.4`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
			},
		},

		{
			filter: `yesterday < request.time`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("yesterday", TypeTimestamp),
			},
			// TODO: Add support for structs.
			errorContains: "undeclared identifier 'request'",
		},

		{
			filter: `experiment.rollout <= cohort(request.user)`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareFunction("cohort", NewFunctionOverload("cohort_string", TypeFloat, TypeString)),
			},
			// TODO: Add support for structs.
			errorContains: "undeclared identifier 'experiment'",
		},

		{
			filter: `prod`,
			declarations: []DeclarationOption{
				DeclareIdent("prod", TypeBool),
			},
		},

		{
			filter: `expr.type_map.1.type`,
			// TODO: Add support for structs.
			errorContains: "undeclared identifier 'expr'",
		},

		{
			filter: `regex(m.key, '^.*prod.*$')`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("m", TypeMap(TypeString, TypeString)),
				DeclareFunction("regex", NewFunctionOverload("regex_string", TypeBool, TypeString, TypeString)),
			},
		},

		{
			filter: `math.mem('30mb')`,
			declarations: []DeclarationOption{
				DeclareFunction("math.mem", NewFunctionOverload("math.mem_string", TypeBool, TypeString)),
			},
		},

		{
			filter: `(msg.endsWith('world') AND retries < 10)`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("retries", TypeInt),
			},
			errorContains: "undeclared function 'msg.endsWith'",
		},

		{
			filter: `(endsWith(msg, 'world') AND retries < 10)`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareFunction("endsWith", NewFunctionOverload("endsWith_string", TypeBool, TypeString, TypeString)),
				DeclareIdent("retries", TypeInt),
				DeclareIdent("msg", TypeString),
			},
		},

		{
			filter: "expire_time > time.now()",
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareFunction("time.now", NewFunctionOverload("time.now", TypeTimestamp)),
				DeclareIdent("expire_time", TypeTimestamp),
			},
		},

		{
			filter: `time.now() > timestamp("2012-04-21T11:30:00-04:00")`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareFunction("time.now", NewFunctionOverload("time.now", TypeTimestamp)),
			},
		},

		{
			filter: `time.now() > timestamp("INVALID")`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareFunction("time.now", NewFunctionOverload("time.now", TypeTimestamp)),
			},
			errorContains: "invalid timestamp",
		},

		{
			filter: `ttl > duration("30s")`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("ttl", TypeDuration),
			},
		},

		{
			filter: `ttl > duration("INVALID")`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("ttl", TypeDuration),
			},
			errorContains: "invalid duration",
		},

		{
			filter: `ttl > duration(input_field)`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("ttl", TypeDuration),
				DeclareIdent("input_field", TypeString),
			},
		},

		{
			filter: `create_time > timestamp("2006-01-02T15:04:05+07:00")`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
		},

		{
			filter: `create_time > timestamp("2006-01-02T15:04:05+07:00")`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
		},

		{
			filter: `
				start_time > timestamp("2006-01-02T15:04:05+07:00") AND
				(driver = "driver1" OR start_driver = "driver1" OR end_driver = "driver1")
			`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("start_time", TypeTimestamp),
				DeclareIdent("driver", TypeString),
				DeclareIdent("start_driver", TypeString),
				DeclareIdent("end_driver", TypeString),
			},
		},

		{
			filter: `annotations:schedule`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("annotations", TypeMap(TypeString, TypeString)),
			},
		},

		{
			filter: `annotations.schedule = "test"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("annotations", TypeMap(TypeString, TypeString)),
			},
		},

		{
			filter: `enum = ENUM_ONE`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareEnumIdent("enum", syntaxv1.Enum(0).Type()),
			},
		},

		{
			filter: `enum = ENUM_ONE AND NOT enum2 = ENUM_TWO`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareEnumIdent("enum", syntaxv1.Enum(0).Type()),
				DeclareEnumIdent("enum2", syntaxv1.Enum(0).Type()),
			},
		},

		{
			filter: `create_time = "2022-08-12 22:22:22"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
			errorContains: "invalid timestamp. Should be in RFC3339 format",
		},

		{
			filter: `create_time = "2022-08-12T22:22:22+01:00"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
		},

		{
			filter: `create_time != "2022-08-12T22:22:22+01:00"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
		},

		{
			filter: `create_time < "2022-08-12T22:22:22+01:00"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
		},

		{
			filter: `create_time > "2022-08-12T22:22:22+01:00"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
		},

		{
			filter: `create_time <= "2022-08-12T22:22:22+01:00"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
		},

		{
			filter: `create_time >= "2022-08-12T22:22:22+01:00"`,
			declarations: []DeclarationOption{
				DeclareStandardFunctions(),
				DeclareIdent("create_time", TypeTimestamp),
			},
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
			parsedExpr, err := parser.Parse()
			if err != nil && tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)
			declarations, err := NewDeclarations(tt.declarations...)
			if err != nil && tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)
			var checker Checker
			checker.Init(parsedExpr.GetExpr(), parsedExpr.GetSourceInfo(), declarations)
			checkedExpr, err := checker.Check()
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
				return
			}
			assert.NilError(t, err)
			assert.Assert(t, checkedExpr != nil)
		})
	}
}
