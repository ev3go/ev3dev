// Copyright ©2017 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package motorutil

import "testing"

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
