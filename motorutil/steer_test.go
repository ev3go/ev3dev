// Copyright ©2017 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package motorutil

import (
	"reflect"
	"testing"
)

var motorRatesTests = []struct {
	speed, turn, counts int

	wantLeftSpeed, wantLeftCounts   int
	wantRightSpeed, wantRightCounts int
}{
	{
		speed: 100, turn: 0, counts: 10,
		wantLeftSpeed: 100, wantLeftCounts: 10,
		wantRightSpeed: 100, wantRightCounts: 10,
	},
	{
		speed: 100, turn: -25, counts: 10,
		wantLeftSpeed: 50, wantLeftCounts: 5,
		wantRightSpeed: 100, wantRightCounts: 10,
	},
	{
		speed: 100, turn: +25, counts: 10,
		wantLeftSpeed: 100, wantLeftCounts: 10,
		wantRightSpeed: 50, wantRightCounts: 5,
	},
	{
		speed: 100, turn: -50, counts: 10,
		wantLeftSpeed: 0, wantLeftCounts: 0,
		wantRightSpeed: 100, wantRightCounts: 10,
	},
	{
		speed: 100, turn: +50, counts: 10,
		wantLeftSpeed: 100, wantLeftCounts: 10,
		wantRightSpeed: 0, wantRightCounts: 0,
	},
	{
		speed: 100, turn: -75, counts: 10,
		wantLeftSpeed: -50, wantLeftCounts: -5,
		wantRightSpeed: 100, wantRightCounts: 10,
	},
	{
		speed: 100, turn: +75, counts: 10,
		wantLeftSpeed: 100, wantLeftCounts: 10,
		wantRightSpeed: -50, wantRightCounts: -5,
	},
	{
		speed: 100, turn: -100, counts: 10,
		wantLeftSpeed: -100, wantLeftCounts: -10,
		wantRightSpeed: 100, wantRightCounts: 10,
	},
	{
		speed: 100, turn: +100, counts: 10,
		wantLeftSpeed: 100, wantLeftCounts: 10,
		wantRightSpeed: -100, wantRightCounts: -10,
	},
}

func TestMotorRates(t *testing.T) {
	for _, speedDirection := range []int{1, -1} {
		for _, countDirection := range []int{1, 0, -1} {
			for _, test := range motorRatesTests {
				test.counts *= countDirection
				test.speed *= speedDirection

				leftSpeed, leftCounts, rightSpeed, rightCounts := motorRates(test.speed, test.turn, test.counts)

				if leftSpeed != test.wantLeftSpeed*speedDirection {
					t.Errorf("unexpected left motor speed for speed=%d turn=%d counts=%d: got:%d want:%d",
						test.speed, test.turn, test.counts,
						leftSpeed, test.wantLeftSpeed*speedDirection)
				}
				if leftCounts != test.wantLeftCounts*countDirection {
					t.Errorf("unexpected left motor counts for speed=%d turn=%d counts=%d: got:%d want:%d",
						test.speed, test.turn, test.counts,
						leftCounts, test.wantLeftCounts*countDirection)
				}
				if rightSpeed != test.wantRightSpeed*speedDirection {
					t.Errorf("unexpected right motor speed for speed=%d turn=%d counts=%d: got:%d want:%d",
						test.speed, test.turn, test.counts,
						rightSpeed, test.wantRightSpeed*speedDirection)
				}
				if rightCounts != test.wantRightCounts*countDirection {
					t.Errorf("unexpected right motor counts for speed=%d turn=%d counts=%d: got:%d want:%d",
						test.speed, test.turn, test.counts,
						rightCounts, test.wantRightCounts*countDirection)
				}
			}
		}
	}
}

var stringSetTests = []struct {
	a, b []string
	want []string
}{
	{
		a:    []string{"1", "2", "3", "4", "5"},
		b:    []string{"1", "2", "3", "4", "5"},
		want: []string{"1", "2", "3", "4", "5"},
	},
	{
		a:    []string{"1", "2", "4", "6", "8"},
		b:    []string{"2", "3", "4", "7", "8"},
		want: []string{"2", "4", "8"},
	},
	{
		a:    []string{"2", "3", "4", "7", "8"},
		b:    []string{"1", "2", "4", "6", "8"},
		want: []string{"2", "4", "8"},
	},
	{
		a:    []string{"4", "5", "6"},
		b:    []string{"1", "2", "3"},
		want: nil,
	},
	{
		a:    []string{"4", "5", "6"},
		b:    []string{"1", "2", "3", "5"},
		want: []string{"5"},
	},
}

func TestEqual(t *testing.T) {
	for _, test := range stringSetTests {
		got := equal(test.a, test.b)
		want := reflect.DeepEqual(test.a, test.b)
		if got != want {
			t.Errorf("unexpected equality between %v and %v: got: %t want:%t",
				test.a, test.b, got, want)
		}
	}
}

func TestIntersect(t *testing.T) {
	for _, test := range stringSetTests {
		got := intersect(test.a, test.b)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("unexpected intersection between %v and %v:\ngot: %v\nwant:%v",
				test.a, test.b, got, test.want)
		}
	}
}
