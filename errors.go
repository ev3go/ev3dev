// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"io"
	"math"
	"path/filepath"
	"runtime"
	"time"
)

// ValidValuer is an error caused by an invalid discrete value.
type ValidValuer interface {
	// Values returns the invalid value
	// and a slice of valid values.
	Values() (value string, valid []string)
}

// ValidRanger is an error caused by an invalid ranged integer value.
type ValidRanger interface {
	// Range returns the invalid value
	// and the range of valid values.
	Range() (value, min, max int)
}

// ValidDurationRanger is an error caused by an invalid ranged time.Duration value.
type ValidDurationRanger interface {
	// DurationRange returns the invalid value
	// and the range of valid values.
	DurationRange() (value, min, max time.Duration)
}

type invalidValueError struct {
	dev   Device
	attr  string
	mesg  string
	value string
	valid []string

	stack
}

func newInvalidValueError(dev Device, attr, message, value string, valid []string) invalidValueError {
	if dev == nil {
		panic("ev3dev: nil device")
	}
	for _, v := range valid {
		if v == value {
			panic(fmt.Sprintf("ev3dev: bad invalid value for %s %s: %q in %q",
				dev, attr, value, valid))
		}
	}
	return invalidValueError{
		dev:   dev,
		attr:  attr,
		mesg:  message,
		value: value,
		valid: valid,
		stack: callers(),
	}
}

func (e invalidValueError) Error() string {
	if e.mesg != "" {
		return fmt.Sprintf("ev3dev: %s for %s %s: %q (valid:%q) at %s",
			e.mesg, e.dev, e.attr, e.value, e.valid, e.caller(0))
	}
	return fmt.Sprintf("ev3dev: invalid value for %s %s: %q (valid:%q) at %s",
		e.dev, e.attr, e.value, e.valid, e.caller(0))
}

func (e invalidValueError) Format(fs fmt.State, c rune) {
	type naked invalidValueError
	switch c {
	case 'v':
		switch {
		case fs.Flag('+'):
			fmt.Fprintln(fs, e.Error())
			e.stack.writeTo(fs)
			return
		case fs.Flag('#'):
			n := fmt.Sprintf("%#v", naked(e))
			fmt.Fprintf(fs, "%T%s", e, n[len("ev3dev.naked"):])
			return
		}
		fallthrough
	case 's':
		io.WriteString(fs, e.Error())
	case 'q':
		fmt.Fprintf(fs, "%q", e.Error())
	default:
		fmt.Fprintf(fs, "%"+string(c), naked(e))
	}
}

func (e invalidValueError) Values() (value string, valid []string) {
	return e.value, e.valid
}

type valueOutOfRangeError struct {
	dev      Device
	attr     string
	value    int
	min, max int

	stack
}

func newValueOutOfRangeError(dev Device, attr string, v, min, max int) valueOutOfRangeError {
	if dev == nil {
		panic("ev3dev: nil device")
	}
	if min <= v && v <= max {
		panic(fmt.Sprintf("ev3dev: bad value out of range error for %s %s: %v in %v-%v",
			dev, attr, v, min, max))
	}
	return valueOutOfRangeError{
		dev:   dev,
		attr:  attr,
		value: v,
		min:   min,
		max:   max,
		stack: callers(),
	}
}

func (e valueOutOfRangeError) Error() string {
	return fmt.Sprintf("ev3dev: invalid value for %s %s: %d (must be in %d-%d) at %s",
		e.dev, e.attr, e.value, e.min, e.max, e.caller(0))
}

func (e valueOutOfRangeError) Format(fs fmt.State, c rune) {
	type naked valueOutOfRangeError
	switch c {
	case 'v':
		switch {
		case fs.Flag('+'):
			fmt.Fprintln(fs, e.Error())
			e.stack.writeTo(fs)
			return
		case fs.Flag('#'):
			n := fmt.Sprintf("%#v", naked(e))
			fmt.Fprintf(fs, "%T%s", e, n[len("ev3dev.naked"):])
			return
		}
		fallthrough
	case 's':
		io.WriteString(fs, e.Error())
	case 'q':
		fmt.Fprintf(fs, "%q", e.Error())
	default:
		fmt.Fprintf(fs, "%"+string(c), naked(e))
	}
}

func (e valueOutOfRangeError) Range() (value, min, max int) {
	return e.value, e.min, e.max
}

type idError struct {
	dev  Device
	attr string
	id   int

	stack
}

func newIDErrorFor(dev Device, id int) idError {
	if dev == nil {
		panic("ev3dev: nil device")
	}
	if id >= 0 {
		panic(fmt.Sprintf("ev3dev: bad id error for %s: %v not negative",
			dev, id))
	}
	return idError{
		dev:   dev,
		id:    id,
		stack: callers(),
	}
}

func (e idError) Error() string {
	return fmt.Sprintf("ev3dev: invalid id for %s: %v (must be positive) at %s",
		e.dev, e.id, e.caller(0))
}

func (e idError) Format(fs fmt.State, c rune) {
	type naked idError
	switch c {
	case 'v':
		switch {
		case fs.Flag('+'):
			fmt.Fprintln(fs, e.Error())
			e.stack.writeTo(fs)
			return
		case fs.Flag('#'):
			n := fmt.Sprintf("%#v", naked(e))
			fmt.Fprintf(fs, "%T%s", e, n[len("ev3dev.naked"):])
			return
		}
		fallthrough
	case 's':
		io.WriteString(fs, e.Error())
	case 'q':
		fmt.Fprintf(fs, "%q", e.Error())
	default:
		fmt.Fprintf(fs, "%"+string(c), naked(e))
	}
}

func (e idError) Range() (value, min, max int) {
	return e.id, 0, int(^uint(0) >> 1)
}

type negativeDurationError struct {
	dev      Device
	attr     string
	duration time.Duration

	stack
}

func newNegativeDurationError(dev Device, attr string, d time.Duration) negativeDurationError {
	if dev == nil {
		panic("ev3dev: nil device")
	}
	if d >= 0 {
		panic(fmt.Sprintf("ev3dev: bad negative duration for %s %s: %v not negative",
			dev, attr, d))
	}
	return negativeDurationError{
		dev:      dev,
		attr:     attr,
		duration: d,
		stack:    callers(),
	}
}

func (e negativeDurationError) Error() string {
	return fmt.Sprintf("ev3dev: invalid duration for %s %s: %v (must be positive) at %s",
		e.dev, e.attr, e.duration, e.caller(0))
}

func (e negativeDurationError) Format(fs fmt.State, c rune) {
	type naked negativeDurationError
	switch c {
	case 'v':
		switch {
		case fs.Flag('+'):
			fmt.Fprintln(fs, e.Error())
			e.stack.writeTo(fs)
			return
		case fs.Flag('#'):
			n := fmt.Sprintf("%#v", naked(e))
			fmt.Fprintf(fs, "%T%s", e, n[len("ev3dev.naked"):])
			return
		}
		fallthrough
	case 's':
		io.WriteString(fs, e.Error())
	case 'q':
		fmt.Fprintf(fs, "%q", e.Error())
	default:
		fmt.Fprintf(fs, "%"+string(c), naked(e))
	}
}

func (e negativeDurationError) DurationRange() (value, min, max time.Duration) {
	return e.duration, 0, math.MaxInt64
}

type durationOutOfRangeError struct {
	dev      Device
	attr     string
	duration time.Duration
	min, max time.Duration

	stack
}

func newDurationOutOfRangeError(dev Device, attr string, d, min, max time.Duration) durationOutOfRangeError {
	if dev == nil {
		panic("ev3dev: nil device")
	}
	if min <= d && d <= max {
		panic(fmt.Sprintf("ev3dev: bad duration out of range error for %s %s: %v in %v-%v",
			dev, attr, d, min, max))
	}
	return durationOutOfRangeError{
		dev:      dev,
		attr:     attr,
		duration: d,
		min:      min,
		max:      max,
		stack:    callers(),
	}
}

func (e durationOutOfRangeError) Error() string {
	return fmt.Sprintf("ev3dev: invalid duration for %s %s: %v (must be in %v-%v) at %s",
		e.dev, e.attr, e.duration, e.min, e.max, e.caller(0))
}

func (e durationOutOfRangeError) Format(fs fmt.State, c rune) {
	type naked durationOutOfRangeError
	switch c {
	case 'v':
		switch {
		case fs.Flag('+'):
			fmt.Fprintln(fs, e.Error())
			e.stack.writeTo(fs)
			return
		case fs.Flag('#'):
			n := fmt.Sprintf("%#v", naked(e))
			fmt.Fprintf(fs, "%T%s", e, n[len("ev3dev.naked"):])
			return
		}
		fallthrough
	case 's':
		io.WriteString(fs, e.Error())
	case 'q':
		fmt.Fprintf(fs, "%q", e.Error())
	default:
		fmt.Fprintf(fs, "%"+string(c), naked(e))
	}
}

func (e durationOutOfRangeError) DurationRange() (value, min, max time.Duration) {
	return e.duration, e.min, e.max
}

type attrOpError struct {
	dev  Device
	attr string
	data string
	op   string
	err  error

	stack
}

func newAttrOpError(dev Device, attr, data, op string, err error) attrOpError {
	return attrOpError{
		dev:   dev,
		attr:  attr,
		data:  data,
		op:    op,
		err:   err,
		stack: callers(),
	}
}

func (e attrOpError) Error() string {
	return fmt.Sprintf("ev3dev: failed to %s %s %s attribute %s: %v at %s",
		e.op, e.dev, e.attr, filepath.Join(e.dev.Path(), e.dev.String(), e.attr), e.err, e.caller(0))
}

func (e attrOpError) Format(fs fmt.State, c rune) {
	type naked attrOpError
	switch c {
	case 'v':
		switch {
		case fs.Flag('+'):
			fmt.Fprintln(fs, e.Error())
			e.stack.writeTo(fs)
			return
		case fs.Flag('#'):
			n := fmt.Sprintf("%#v", naked(e))
			fmt.Fprintf(fs, "%T%s", e, n[len("ev3dev.naked"):])
			return
		}
		fallthrough
	case 's':
		io.WriteString(fs, e.Error())
	case 'q':
		fmt.Fprintf(fs, "%q", e.Error())
	default:
		fmt.Fprintf(fs, "%"+string(c), naked(e))
	}
}

func (e attrOpError) Cause() error  { return e.err }
func (e attrOpError) Unwrap() error { return e.err }

type parseError struct {
	dev  Device
	attr string
	err  error

	stack
}

func newParseError(dev Device, attr string, err error) parseError {
	return parseError{
		dev:   dev,
		attr:  attr,
		err:   err,
		stack: callers(),
	}
}

func (e parseError) Error() string {
	return fmt.Sprintf("ev3dev: failed to parse %s %s attribute %s: %v at %s",
		e.dev, e.attr, filepath.Join(e.dev.Path(), e.dev.String(), e.attr), e.err, e.caller(1))
}

func (e parseError) Format(fs fmt.State, c rune) {
	type naked parseError
	switch c {
	case 'v':
		switch {
		case fs.Flag('+'):
			fmt.Fprintln(fs, e.Error())
			e.stack.writeTo(fs)
			return
		case fs.Flag('#'):
			n := fmt.Sprintf("%#v", naked(e))
			fmt.Fprintf(fs, "%T%s", e, n[len("ev3dev.naked"):])
			return
		}
		fallthrough
	case 's':
		io.WriteString(fs, e.Error())
	case 'q':
		fmt.Fprintf(fs, "%q", e.Error())
	default:
		fmt.Fprintf(fs, "%"+string(c), naked(e))
	}
}

func (e parseError) Cause() error  { return e.err }
func (e parseError) Unwrap() error { return e.err }

type syntaxError string

func (e syntaxError) Error() string { return fmt.Sprintf("unexpected line: %q", string(e)) }

type stack []uintptr

func callers() stack {
	var pc [64]uintptr
	n := runtime.Callers(3, pc[:])
	return pc[:n]
}

func (s stack) caller(depth int) string {
	if len(s) <= depth || s[depth] == 0 {
		return "<unknown caller>"
	}
	fn := runtime.FuncForPC(s[depth])
	file, line := fn.FileLine(s[depth])
	return fmt.Sprintf("%s:%d %s", filepath.Base(file), line, fn.Name())
}

func (s stack) writeTo(w io.Writer) (int, error) {
	var n int
	for _, pc := range s {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		_n, err := fmt.Fprintf(w, "%s\n\t%s:%d\n", fn.Name(), filepath.Base(file), line)
		n += _n
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

type causer interface {
	Cause() error
}

func cause(err error) error {
	c, ok := err.(causer)
	if ok {
		return c.Cause()
	}
	return err
}
