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
	wantGoSyntax    string
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
		wantGoSyntax:    `ev3dev.invalidValueError{dev:ev3dev.mockDevice{}, attr:"attr", mesg:"", value:"invalid", valid:[]string{"ok", "valid"}, stack:ev3dev.stack{0x0, 0x0, 0x0, 0x0, 0x0}}`,
	},
	{
		fn: func() error {
			return newInvalidValueError(mockDevice{}, "attr", "unexpected value", "surprise", []string{"ok", "valid"})
		},
		wantErrorPrefix: `ev3dev: unexpected value for mock attr: "surprise" (valid:["ok" "valid"]) at errors_test.go:`,
		wantGoSyntax:    `ev3dev.invalidValueError{dev:ev3dev.mockDevice{}, attr:"attr", mesg:"unexpected value", value:"surprise", valid:[]string{"ok", "valid"}, stack:ev3dev.stack{0x0, 0x0, 0x0, 0x0, 0x0}}`,
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
		wantGoSyntax:    `ev3dev.valueOutOfRangeError{dev:ev3dev.mockDevice{}, attr:"attr", value:0, min:1, max:2, stack:ev3dev.stack{0x0, 0x0, 0x0, 0x0, 0x0}}`,
	},

	{
		fn: func() error {
			return newIDErrorFor(nil, -1)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newIDErrorFor(mockDevice{}, 0)
		},
		panics: true,
	},
	{
		fn: func() error {
			return newIDErrorFor(mockDevice{}, -1)
		},
		wantErrorPrefix: `ev3dev: invalid id for mock: -1 (must be positive) at errors_test.go:`,
		wantGoSyntax:    `ev3dev.idError{dev:ev3dev.mockDevice{}, attr:"", id:-1, stack:ev3dev.stack{0x0, 0x0, 0x0, 0x0, 0x0}}`,
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
		wantGoSyntax:    `ev3dev.negativeDurationError{dev:ev3dev.mockDevice{}, attr:"attr", duration:-1, stack:ev3dev.stack{0x0, 0x0, 0x0, 0x0, 0x0}}`,
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
		wantGoSyntax:    `ev3dev.durationOutOfRangeError{dev:ev3dev.mockDevice{}, attr:"attr", duration:0, min:1, max:2, stack:ev3dev.stack{0x0, 0x0, 0x0, 0x0, 0x0}}`,
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

		var s stack
		switch got := got.(type) {
		case invalidValueError:
			s = got.stack
		case valueOutOfRangeError:
			s = got.stack
		case idError:
			s = got.stack
		case negativeDurationError:
			s = got.stack
		case durationOutOfRangeError:
			s = got.stack
		default:
			panic(fmt.Sprintf("unexpected error type: %T", got))
		}
		// Zero out the frames to ensure consistent results.
		for i := range s {
			s[i] = 0
		}
		if errStr := fmt.Sprintf("%#v", got); errStr != test.wantGoSyntax {
			t.Errorf("unexpected error Go syntax string: got:%s want:%s", errStr, test.wantGoSyntax)
		}
	}
}

const (
	wantCaller = "stack_test.go:13 github.com/ev3go/ev3dev.testStack"

	// Expected output for go1.13 runtime.
	wantTracePrefix113 = `github.com/ev3go/ev3dev.testStack
	stack_test.go:13
github.com/ev3go/ev3dev.testStack
	stack_test.go:13
github.com/ev3go/ev3dev.testStack
	stack_test.go:13
github.com/ev3go/ev3dev.init
	stack_test.go:7
runtime.doInit
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
	if !hasAnyPrefix(gotTrace, wantTracePrefix113) {
		t.Errorf("unexpected trace string:\ngot:\n%s\nwant prefix:\n%s", gotTrace, wantTracePrefix113)
	}
}

const (
	// Expected output for go1.13 runtime.
	wantGoSyntax113         = `ev3dev.invalidValueError{dev:ev3dev.mockDevice{}, attr:"attr", mesg:"", value:"invalid", valid:[]string{"ok", "valid"}, stack:ev3dev.stack{0x0, 0x0, 0x0, 0x0, 0x0}}`
	wantErrorTracePrefix113 = `ev3dev: invalid value for mock attr: "invalid" (valid:["ok" "valid"]) at stack_test.go:16 github.com/ev3go/ev3dev.init
github.com/ev3go/ev3dev.init
	stack_test.go:16
runtime.doInit`
)

func TestPrintTrace(t *testing.T) {
	gotTrace := fmt.Sprintf("%+v", mockValueError)
	if !hasAnyPrefix(gotTrace, wantErrorTracePrefix113) {
		t.Errorf("unexpected trace string:\ngot:\n%s\nwant:\n%s", gotTrace, wantErrorTracePrefix113)
	}
	for i := range mockValueError.stack {
		mockValueError.stack[i] = 0
	}
	gotGoSyntax := fmt.Sprintf("%#v", mockValueError)
	if !matchesAny(gotGoSyntax, wantGoSyntax113) {
		t.Errorf("unexpected Go syntax string: got:%s want:%s", gotGoSyntax, wantGoSyntax113)
	}
}

func hasAnyPrefix(q string, prefixes ...string) bool {
	for _, pre := range prefixes {
		if strings.HasPrefix(q, pre) {
			return true
		}
	}
	return false
}

func matchesAny(q string, targets ...string) bool {
	for _, t := range targets {
		if q == t {
			return true
		}
	}
	return false
}
