// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

var intFromTest = []struct {
	data    string
	attr    string
	err     error
	wantInt int
	wantErr error
}{
	{data: "", attr: "empty", err: nil, wantInt: -1, wantErr: errors.New(`ev3dev: failed to parse mock empty attribute path/mock/empty: strconv.Atoi: parsing "": invalid syntax at ev3dev_conv_test.go:`)},
	{data: "1", attr: "one", err: nil, wantInt: 1, wantErr: nil},
	{data: "0", attr: "zero", err: nil, wantInt: 0, wantErr: nil},
	{data: "-1", attr: "minus_one", err: nil, wantInt: -1, wantErr: nil},
	{data: "0", attr: "prior", err: errors.New("prior error"), wantInt: -1, wantErr: errors.New("prior error")},
}

func TestIntFrom(t *testing.T) {
	for _, test := range intFromTest {
		gotInt, gotErr := intFrom(mockDevice{}, test.data, test.attr, test.err)

		if !strings.HasPrefix(fmt.Sprint(gotErr), fmt.Sprint(test.wantErr)) {
			t.Errorf("unexpected error:\ngot:\n\t%v\nwant prefix:\n\t%v", gotErr, test.wantErr)
		}
		if gotInt != test.wantInt {
			t.Errorf("unexpected integer result: got:%d want:%d", gotInt, test.wantInt)
		}
	}
}

func isSame(a, b float64) bool {
	return a == b || (math.IsNaN(a) && math.IsNaN(b))
}

var float64FromTest = []struct {
	data        string
	attr        string
	err         error
	wantFloat64 float64
	wantErr     error
}{
	{data: "", attr: "empty", err: nil, wantFloat64: math.NaN(), wantErr: errors.New(`ev3dev: failed to parse mock empty attribute path/mock/empty: strconv.ParseFloat: parsing "": invalid syntax at ev3dev_conv_test.go:`)},
	{data: "1", attr: "one", err: nil, wantFloat64: 1, wantErr: nil},
	{data: "0", attr: "zero", err: nil, wantFloat64: 0, wantErr: nil},
	{data: "-1", attr: "minus_one", err: nil, wantFloat64: -1, wantErr: nil},
	{data: "0", attr: "prior", err: errors.New("prior error"), wantFloat64: math.NaN(), wantErr: errors.New("prior error")},
}

func TestFloat64From(t *testing.T) {
	for _, test := range float64FromTest {
		gotFloat64, gotErr := float64From(mockDevice{}, test.data, test.attr, test.err)

		if !strings.HasPrefix(fmt.Sprint(gotErr), fmt.Sprint(test.wantErr)) {
			t.Errorf("unexpected error:\ngot:\n\t%v\nwant prefix:\n\t%v", gotErr, test.wantErr)
		}
		if !isSame(gotFloat64, test.wantFloat64) {
			t.Errorf("unexpected float64 result: got:%f want:%f", gotFloat64, test.wantFloat64)
		}
	}
}

var durationFromTest = []struct {
	data         string
	attr         string
	err          error
	wantDuration time.Duration
	wantErr      error
}{
	{data: "", attr: "empty", err: nil, wantDuration: -1, wantErr: errors.New(`ev3dev: failed to parse mock empty attribute path/mock/empty: strconv.Atoi: parsing "": invalid syntax at ev3dev_conv_test.go:`)},
	{data: "1", attr: "one", err: nil, wantDuration: 1 * time.Millisecond, wantErr: nil},
	{data: "0", attr: "zero", err: nil, wantDuration: 0, wantErr: nil},
	{data: "-1", attr: "minus_one", err: nil, wantDuration: -1 * time.Millisecond, wantErr: nil},
	{data: "0", attr: "prior", err: errors.New("prior error"), wantDuration: -1, wantErr: errors.New("prior error")},
}

func TestDurationFrom(t *testing.T) {
	for _, test := range durationFromTest {
		gotDuration, gotErr := durationFrom(mockDevice{}, test.data, test.attr, test.err)

		if !strings.HasPrefix(fmt.Sprint(gotErr), fmt.Sprint(test.wantErr)) {
			t.Errorf("unexpected error:\ngot:\n\t%v\nwant prefix:\n\t%v", gotErr, test.wantErr)
		}
		if gotDuration != test.wantDuration {
			t.Errorf("unexpected duration result: got:%v want:%v", gotDuration, test.wantDuration)
		}
	}
}

var stringSliceFromTest = []struct {
	data        string
	attr        string
	err         error
	wantStrings []string
	wantErr     error
}{
	{data: "", attr: "empty", err: nil, wantStrings: nil, wantErr: nil},
	{data: "1", attr: "one", err: nil, wantStrings: []string{"1"}, wantErr: nil},
	{data: "0 1", attr: "two", err: nil, wantStrings: []string{"0", "1"}, wantErr: nil},
	{data: "0\t1", attr: "tab", err: nil, wantStrings: []string{"0\t1"}, wantErr: nil},
	{data: "0", attr: "prior", err: errors.New("prior error"), wantStrings: nil, wantErr: errors.New("prior error")},
}

func TestStringSliceFrom(t *testing.T) {
	for _, test := range stringSliceFromTest {
		gotStrings, gotErr := stringSliceFrom(mockDevice{}, test.data, test.attr, test.err)

		if fmt.Sprint(gotErr) != fmt.Sprint(test.wantErr) {
			t.Errorf("unexpected error:\ngot:\n\t%v\nwant prefix:\n\t%v", gotErr, test.wantErr)
		}
		if !reflect.DeepEqual(gotStrings, test.wantStrings) {
			t.Errorf("unexpected strings result: got:%v want:%v", gotStrings, test.wantStrings)
		}
	}
}

var stateFromTest = []struct {
	data      string
	attr      string
	err       error
	wantState MotorState
	wantErr   error
}{
	{data: "", attr: "empty", err: nil, wantState: 0, wantErr: nil},
	{data: running, attr: running, err: nil, wantState: Running, wantErr: nil},
	{data: ramping, attr: ramping, err: nil, wantState: Ramping, wantErr: nil},
	{data: holding, attr: holding, err: nil, wantState: Holding, wantErr: nil},
	{data: overloaded, attr: overloaded, err: nil, wantState: Overloaded, wantErr: nil},
	{data: stalled, attr: stalled, err: nil, wantState: Stalled, wantErr: nil},
	{data: running + " " + stalled, attr: running + " " + stalled, err: nil, wantState: Running | Stalled, wantErr: nil},
	{data: "invalid", attr: "invalid", err: nil, wantState: 0, wantErr: errors.New(`ev3dev: unrecognized motor state for mock state: "invalid" (valid:["holding" "overloaded" "ramping" "running" "stalled"]) at ev3dev.go:`)},
	{data: "0", attr: "prior", err: errors.New("prior error"), wantState: 0, wantErr: errors.New("prior error")},
}

func TestStateFrom(t *testing.T) {
	for _, test := range stateFromTest {
		gotState, gotErr := stateFrom(mockDevice{}, test.data, test.attr, test.err)

		if !strings.HasPrefix(fmt.Sprint(gotErr), fmt.Sprint(test.wantErr)) {
			t.Errorf("unexpected error:\ngot:\n\t%v\nwant prefix:\n\t%v", gotErr, test.wantErr)
		}
		if gotState != test.wantState {
			t.Errorf("unexpected state result: got:%v want:%v", gotState, test.wantState)
		}
	}
}

type ue map[string]string

var ueventFromTest = []struct {
	data        string
	attr        string
	err         error
	wantUevents map[string]string
	wantErr     error
}{
	{data: "", attr: "empty", err: nil, wantUevents: nil, wantErr: nil},
	{data: "one=1", attr: "one", err: nil, wantUevents: ue{"one": "1"}, wantErr: nil},
	{data: "zero=0\none=1", attr: "two", err: nil, wantUevents: ue{"zero": "0", "one": "1"}, wantErr: nil},
	{data: "0", attr: "zero", err: nil, wantUevents: nil, wantErr: errors.New(`ev3dev: failed to parse mock zero attribute path/mock/zero: unexpected line: "0" at ev3dev_conv_test.go:`)},
	{data: "0", attr: "prior", err: errors.New("prior error"), wantUevents: nil, wantErr: errors.New("prior error")},
}

func TestUeventFrom(t *testing.T) {
	for _, test := range ueventFromTest {
		gotUevents, gotErr := ueventFrom(mockDevice{}, test.data, test.attr, test.err)

		if !strings.HasPrefix(fmt.Sprint(gotErr), fmt.Sprint(test.wantErr)) {
			t.Errorf("unexpected error:\ngot:\n\t%v\nwant prefix:\n\t%v", gotErr, test.wantErr)
		}
		if !reflect.DeepEqual(gotUevents, test.wantUevents) {
			t.Errorf("unexpected uevent result: got:%v want:%v", gotUevents, test.wantUevents)
		}
	}
}
