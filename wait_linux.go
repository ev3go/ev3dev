// Copyright ©2017 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package ev3dev

import (
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
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
func Wait(d StaterDevice, mask, want, not MotorState, any bool, timeout time.Duration) (stat MotorState, ok bool, err error) {
	// We use a direct implementation of the State method here
	// to ensure we are polling on the same file as we are reading
	// from. Also, since we are potentially probing the state
	// repeatedly, we save file opens.
	//
	// This also allows us to test the code, which would not
	// otherwise be possible since sysiphus cannot do POLLPRI
	// polling, due to limitations in FUSE.

	// Check if we can proceed.
	err = d.Err()
	if err != nil {
		return 0, false, err
	}

	path := filepath.Join(d.Path(), d.String(), state)
	f, err := os.Open(path)
	if err != nil {
		return 0, false, err
	}
	defer f.Close()

	// See if we can exit early.
	stat, err = motorState(d, f)
	if err != nil {
		return stat, false, err
	}
	if stateIsOK(stat, mask, want, not, any) {
		return stat, true, nil
	}

	var fds []unix.PollFd
	if canPoll {
		fds = []unix.PollFd{{Fd: int32(f.Fd()), Events: unix.POLLIN}}

		// Read a single byte to mark f as unchanged.
		f.ReadAt([]byte{0}, 0)
	}

	end := time.Now().Add(timeout)
	for timeout < 0 || time.Since(end) < 0 {
		if canPoll {
			_timeout := timeout
			if timeout >= 0 {
				if remain := end.Sub(time.Now()); remain < timeout {
					_timeout = remain
				}
			}
			n, err := unix.Poll(fds, int(_timeout/time.Millisecond))
			if n == 0 {
				return 0, false, err
			}
		}
		stat, err = motorState(d, f)
		if err != nil {
			return stat, false, err
		}
		if stateIsOK(stat, mask, want, not, any) {
			return stat, true, nil
		}

		relax := 50 * time.Millisecond
		if remain := end.Sub(time.Now()); remain < relax {
			relax = remain / 2
		}
		time.Sleep(relax)
	}

	return stat, false, nil
}
