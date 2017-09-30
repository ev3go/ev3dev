// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package motorutil

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ev3go/ev3dev"
)

// Steering implements a paired-motor steering unit similar to an EV3-G steering block.
//
// Errors ocurring during steering operations are sticky. They are returned either by
// a call to Err or Wait.
type Steering struct {
	// Left and Right are the left and right motors to be
	// used by the steering module.
	Left, Right *ev3dev.TachoMotor

	// Timeout is the timeout for waiting for motors to
	// return to a non-driving state.
	//
	// See ev3dev.Wait documentation for timeout behaviour.
	Timeout time.Duration

	err error
}

// StopAction returns the stop action used when a stop command is issued
// to the TachoMotor devices held by the Steering. StopAction returns an
// error if the two motors do not agree on the stop action.
func (s *Steering) StopAction() (string, error) {
	err := s.Err()
	if err != nil {
		return "", err
	}

	lAction, err := s.Left.StopAction()
	if err != nil {
		return "", err
	}
	rAction, err := s.Right.StopAction()
	if err != nil {
		return "", err
	}
	if lAction != rAction {
		return "", actionMismatch{left: lAction, right: rAction}
	}
	return lAction, nil
}

type actionMismatch struct {
	left, right string
}

func (e actionMismatch) Error() string {
	return fmt.Sprintf("motorutil: stop action mismatch: %s != %s", e.left, e.right)
}

// StopActions returns the available stop actions for the TachoMotor.
// StopAction returns an error and the intersection of available actions
// if the two motors do not agree on the available stop actions.
func (s *Steering) StopActions() ([]string, error) {
	err := s.Err()
	if err != nil {
		return nil, err
	}

	lActions := s.Left.StopActions()
	rActions := s.Right.StopActions()
	sort.Strings(lActions)
	sort.Strings(rActions)
	if !equal(lActions, rActions) {
		err = actionsMismatch{left: lActions, right: rActions}
		lActions = intersect(lActions, rActions)
	}
	return lActions, err
}

// equal returns whether a and b are equal.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, va := range a {
		if va != b[i] {
			return false
		}
	}
	return true
}

// intersect returns a new []string containing the intersection of the sorted
// slices a and b.
func intersect(a, b []string) []string {
	var c []string
	var ia, ib int
	for ia < len(a) && ib < len(b) {
		av := a[ia]
		bv := b[ib]
		switch {
		case av == bv:
			c = append(c, av)
			ia++
			ib++
		case av < bv:
			ia++
		case av > bv:
			ib++
		}
	}
	return c
}

type actionsMismatch struct {
	left, right []string
}

func (e actionsMismatch) Error() string {
	return fmt.Sprintf("motorutil: available stop actions mismatch: %s != %s", e.left, e.right)
}

// SetStopAction sets the stop action to be used when a stop command is
// issued to the TachoMotor. SetStopAction returns on the first error
// encountered.
func (s *Steering) SetStopAction(action string) *Steering {
	if s.err != nil {
		return s
	}

	s.err = s.Left.SetStopAction(action).Err()
	if s.err != nil {
		return s
	}
	s.err = s.Right.SetStopAction(action).Err()
	return s
}

// SteerCounts steers in the given turn for the given tacho counts and at the
// specified speed. The valid range of turn is -100 (hard left) to +100 (hard right).
// If the product of counts and speed is negative, the turn will be made in reverse.
//
// See the ev3dev.SetSpeedSetPoint and ev3dev.SetPositionSetPoint documentation for
// speed and count behaviour.
func (s *Steering) SteerCounts(speed, turn, counts int) *Steering {
	if s.err != nil {
		return s
	}

	if turn < -100 || 100 < turn {
		s.err = directionError(turn)
		return s
	}

	// Make speed a velocity relative to the counts vector.
	if speed < 0 {
		counts = -counts
	}
	// leftSpeed and rightSpeed may be signed here,
	// but ev3dev ignores speed_sp for run-to-*-pos.
	leftSpeed, leftCounts, rightSpeed, rightCounts := motorRates(speed, turn, counts)

	s.err = s.Left.
		SetSpeedSetpoint(leftSpeed).
		SetPositionSetpoint(leftCounts).
		Err()
	if s.err != nil {
		return s
	}
	s.err = s.Right.
		SetSpeedSetpoint(rightSpeed).
		SetPositionSetpoint(rightCounts).
		Err()
	if s.err != nil {
		return s
	}

	// TODO(kortschak): Remove conditional stop when the
	// driver handles zero relative position change as a no-op.
	if leftCounts == 0 {
		s.err = s.Left.Command("stop").Err()
	} else {
		s.err = s.Left.Command("run-to-rel-pos").Err()
	}
	if s.err != nil {
		return s
	}
	// TODO(kortschak): Remove conditional stop when the
	// driver handles zero relative position change as a no-op.
	if rightCounts == 0 {
		s.err = s.Right.Command("stop").Err()
	} else {
		s.err = s.Right.Command("run-to-rel-pos").Err()
	}
	if s.err != nil {
		s.Left.Command("stop").Err()
	}
	return s
}

// SteerDuration steers in the given turn for the given duration, d,  and at the
// specified speed. The valid range of turn is -100 (hard left) to +100 (hard right).
// If speed is negative, the turn will be made in reverse.
//
// See the ev3dev.SetSpeedSetpoint and ev3dev.SetTimeSetpoint documentation for speed
// and duration behaviour.
func (s *Steering) SteerDuration(speed, turn int, d time.Duration) *Steering {
	if s.err != nil {
		return s
	}

	if turn < -100 || 100 < turn {
		s.err = directionError(turn)
		return s
	}
	if d < 0 {
		s.err = durationError(d)
		return s
	}

	leftSpeed, _, rightSpeed, _ := motorRates(speed, turn, 0)

	s.err = s.Left.
		SetSpeedSetpoint(leftSpeed).
		SetTimeSetpoint(d).
		Err()
	if s.err != nil {
		return s
	}
	s.err = s.Right.
		SetSpeedSetpoint(rightSpeed).
		SetTimeSetpoint(d).
		Err()
	if s.err != nil {
		return s
	}

	s.err = s.Left.Command("run-timed").Err()
	if s.err != nil {
		return s
	}
	s.err = s.Right.Command("run-timed").Err()
	if s.err != nil {
		s.Left.Command("stop").Err()
	}
	return s
}

func motorRates(speed, turn, counts int) (leftSpeed, leftCounts, rightSpeed, rightCounts int) {
	switch {
	case turn == 0:
		leftSpeed = speed
		rightSpeed = speed
		leftCounts = counts
		rightCounts = counts
	case turn < 0:
		rightSpeed = speed
		rightCounts = counts
		turn = (turn + 50) * 2
		leftSpeed = (speed * turn) / 100
		leftCounts = (rightCounts * turn) / 100
	case turn > 0:
		leftSpeed = speed
		leftCounts = counts
		turn = (50 - turn) * 2
		rightSpeed = (speed * turn) / 100
		rightCounts = (leftCounts * turn) / 100
	}
	return leftSpeed, leftCounts, rightSpeed, rightCounts
}

// Err returns the error state of the Steering and clears it.
func (s *Steering) Err() error {
	err := s.err
	s.err = nil
	return err
}

// Wait waits for the last steering operation to complete. A non-nil error will either
// implement the Cause method, which may be used to determine the underlying cause, or
// be an Errors holding errors that implement the Cause method.
func (s *Steering) Wait() error {
	if err := s.Err(); err != nil {
		return err
	}

	var errors [2]error

	var wg sync.WaitGroup
	for i, motor := range []struct {
		side   string
		device *ev3dev.TachoMotor
	}{
		{side: "left", device: s.Left},
		{side: "right", device: s.Right},
	} {
		i := i
		side := motor.side
		device := motor.device
		wg.Add(1)
		go func() {
			defer wg.Done()
			stat, ok, err := ev3dev.Wait(device, ev3dev.Running, 0, 0, false, s.Timeout)
			if err != nil {
				errors[i] = waitError{side: side, motor: device, cause: err}
			}
			if !ok {
				errors[i] = waitError{side: side, motor: device, cause: timeoutError(s.Timeout), stat: stat}
			}
		}()
	}
	wg.Wait()

	switch {
	case errors[0] != nil && errors[1] != nil:
		return Errors(errors[:])
	case errors[0] != nil:
		return errors[0]
	case errors[1] != nil:
		return errors[1]
	}
	return nil
}

// directionError is a ev3dev.ValidFloat64Ranger error.
type directionError int

var _ ev3dev.ValidRanger = directionError(0)

func (e directionError) Error() string {
	return fmt.Sprintf("motorutil: invalid turn: %d (must be in within -100 to 100)", e)
}

func (e directionError) Range() (value, min, max int) {
	return int(e), -100, 100
}

// durationError is a ev3dev.ValidDurationRanger error.
type durationError time.Duration

var _ ev3dev.ValidDurationRanger = durationError(0)

func (e durationError) Error() string {
	return fmt.Sprintf("motorutil: invalid duration: %v (must be positive)", time.Duration(e))
}

func (e durationError) DurationRange() (value, min, max time.Duration) {
	return time.Duration(e), 0, math.MaxInt64
}

// waitError is a Causer error.
type waitError struct {
	side  string
	motor *ev3dev.TachoMotor
	stat  ev3dev.MotorState
	cause error
}

func (e waitError) Error() string {
	if _, ok := e.cause.(timeoutError); ok {
		return fmt.Sprintf("motorutil: failed to wait for %s motor (%v) to stop (state=%v): %v", e.side, e.motor, e.stat, e.cause)
	}
	return fmt.Sprintf("motorutil: failed to wait for %s motor (%v) to stop: %v", e.side, e.motor, e.cause)
}

func (e waitError) Cause() error { return e.cause }

// timeoutError is a timeout failure.
type timeoutError time.Duration

func (e timeoutError) Error() string {
	return fmt.Sprintf("motorutil: wait timed out: longer than %v", time.Duration(e))
}

func (e timeoutError) Timeout() bool { return true }
