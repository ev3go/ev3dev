// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

var errorTests = []struct {
	fn     func() error
	panics bool

	wantErrorPrefix string
}{
	{
		fn: func() error {
			return newInvalidValueError(nil, "", "", "", nil)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newInvalidValueError(mockDevice{}, "", "", "valid", []string{"ok", "valid"})
		},
		panics: true,
	},
	{
		fn: func() error {
			return newInvalidValueError(mockDevice{}, "attr", "", "invalid", []string{"ok", "valid"})
		},
		wantErrorPrefix: `ev3dev: invalid value for mock attr: "invalid" (valid:["ok" "valid"]) at errors_test.go:`,
	},
	{
		fn: func() error {
			return newInvalidValueError(mockDevice{}, "attr", "unexpected value", "surprise", []string{"ok", "valid"})
		},
		wantErrorPrefix: `ev3dev: unexpected value for mock attr: "surprise" (valid:["ok" "valid"]) at errors_test.go:`,
	},

	{
		fn: func() error {
			return newValueOutOfRangeError(nil, "", 0, -1, 1)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newValueOutOfRangeError(mockDevice{}, "", 0, 0, 0)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newValueOutOfRangeError(mockDevice{}, "attr", 0, 1, 2)
		},
		wantErrorPrefix: `ev3dev: invalid value for mock attr: 0 (must be in 1-2) at errors_test.go:`,
	},

	{
		fn: func() error {
			return newNegativeDurationError(nil, "", -1)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newNegativeDurationError(mockDevice{}, "", 0)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newNegativeDurationError(mockDevice{}, "attr", -1)
		},
		wantErrorPrefix: `ev3dev: invalid duration for mock attr: -1ns (must be positive) at errors_test.go:`,
	},

	{
		fn: func() error {
			return newDurationOutOfRangeError(nil, "", 0, -1, 1)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newDurationOutOfRangeError(mockDevice{}, "", 0, 0, 0)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newDurationOutOfRangeError(mockDevice{}, "attr", 0, 1, 2)
		},
		wantErrorPrefix: `ev3dev: invalid duration for mock attr: 0s (must be in 1ns-2ns) at errors_test.go:`,
	},
}

func panics(fn func() error) (err error, panicked bool) {
	defer func() {
		r := recover()
		panicked = r != nil
	}()
	err = fn()
	return err, panicked
}

func TestErrors(t *testing.T) {
	for i, test := range errorTests {
		got, panicked := panics(test.fn)
		if panicked != test.panics {
			t.Errorf("expected panic for test %d", i)
			continue
		}
		if panicked {
			continue
		}
		if errStr := fmt.Sprint(got); !strings.HasPrefix(errStr, test.wantErrorPrefix) {
			t.Errorf("unexpected error string:\ngot:\n\t%s\nwant prefix:\n\t%s", errStr, test.wantErrorPrefix)
		}
	}
}

const (
	wantCaller      = "stack_test.go:13 github.com/ev3go/ev3dev.testStack"
	wantTracePrefix = `github.com/ev3go/ev3dev.testStack
	stack_test.go:13
github.com/ev3go/ev3dev.testStack
	stack_test.go:13
github.com/ev3go/ev3dev.testStack
	stack_test.go:13
github.com/ev3go/ev3dev.init
	stack_test.go:7
main.init
`
)

func TestStack(t *testing.T) {
	gotCaller := st.caller(0)
	if gotCaller != wantCaller {
		t.Errorf("unexpected caller string: got:%q want:%q", gotCaller, wantCaller)
	}
	var buf bytes.Buffer
	_, err := st.writeTo(&buf)
	if err != nil {
		t.Errorf("unexpected error writing trace: %v", err)
	}
	gotTrace := buf.String()
	if !strings.HasPrefix(gotTrace, wantTracePrefix) {
		t.Errorf("unexpected trace string:\ngot:\n%s\nwant prefix:\n%s", gotTrace, wantTracePrefix)
	}
}
