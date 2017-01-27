// Copyright ©2017 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !linux

package ev3dev

import (
	"time"
)

// Wait blocks until the wanted motor state under the motor state mask is
// reached, or the timeout is reached. If timeout is negative Wait will wait
// indefinitely for the wanted motor state has been reached.
// The last unmasked motor state is returned unless the timeout was reached
// before the motor state was read.
// When the any parameter is false, Wait will return ok as true if
//  (state&mask)^not == want|not
// and when any is true Wait return false if
//  (state&mask)^not != 0 && state&mask&not == 0 .
// Otherwise ok will return false indicating that the returned state did
// not match the request.
// Wait will not set the error state of the StaterDevice, but will clear and
// return it if it is not nil.
//
// Wait is not implemented without a linux OS (needs unix.Poll).
func Wait(d StaterDevice, mask, want, not MotorState, any bool, timeout time.Duration) (stat MotorState, ok bool, err error) {
	panic("ev3dev: needs GOOS=linux")
}
