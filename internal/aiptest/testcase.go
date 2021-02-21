package aiptest

import (
	"go.einride.tech/aip/internal/codegen"
)

type testCase struct {
	enabled bool
	name    string
	fn      func(file *codegen.File)
}

func disabledTestCase() testCase {
	return testCase{}
}

func newTestCase(name string, fn func(f *codegen.File)) testCase {
	return testCase{
		enabled: true,
		name:    name,
		fn:      fn,
	}
}

func (t testCase) Name() string {
	return t.name
}

func (t testCase) FuncName() string {
	return "test" + t.name
}
