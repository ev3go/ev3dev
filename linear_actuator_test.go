// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev_test

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	. "github.com/ev3go/ev3dev"

	"github.com/ev3go/sisyphus"
)

// linearActuator is a linearActuator sysfs directory.
type linearActuator struct {
	address string
	driver  string

	lastCommand string
	commands    []string

	countPerM int

	maxSpeed int
	speed    int
	speedSet int

	rampUpSet   time.Duration
	rampDownSet time.Duration

	timeSet time.Duration

	dutyCycle    int
	dutyCycleSet int

	polarity Polarity

	position    int
	positionSet int

	holdPIDkd int
	holdPIDki int
	holdPIDkp int

	speedPIDkd int
	speedPIDki int
	speedPIDkp int

	state MotorState

	lastStopAction string
	stopActions    []string

	uevent map[string]string

	t *testing.T
}

// linearActuatorAddress is the address attribute.
type linearActuatorAddress linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorAddress) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.address)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorAddress) Size() (int64, error) {
	return size(m.address), nil
}

// linearActuatorDriver is the driver_name attribute.
type linearActuatorDriver linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorDriver) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.driver)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorDriver) Size() (int64, error) {
	return size(m.driver), nil
}

// linearActuatorCommands is the commands attribute.
type linearActuatorCommands linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorCommands) ReadAt(b []byte, offset int64) (int, error) {
	if len(m.commands) == 0 {
		return len(b), syscall.ENOTSUP
	}
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorCommands) Size() (int64, error) {
	return size(m), nil
}

func (m *linearActuatorCommands) String() string {
	sort.Strings(m.commands)
	return strings.Join(m.commands, " ")
}

// linearActuatorCommand is the command attribute.
type linearActuatorCommand linearActuator

// Truncate is a no-op.
func (m *linearActuatorCommand) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorCommand) WriteAt(b []byte, off int64) (int, error) {
	if len(m.commands) == 0 {
		return len(b), syscall.ENOTSUP
	}
	command := string(chomp(b))
	for _, c := range m.commands {
		if command == c {
			m.lastCommand = command
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorCommand) Size() (int64, error) {
	return size(m.lastCommand), nil
}

// linearActuatorStopActions is the stop_actions attribute.
type linearActuatorStopActions linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorStopActions) ReadAt(b []byte, offset int64) (int, error) {
	if len(m.stopActions) == 0 {
		return len(b), syscall.ENOTSUP
	}
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorStopActions) Size() (int64, error) {
	return size(m), nil
}

func (m *linearActuatorStopActions) String() string {
	sort.Strings(m.stopActions)
	return strings.Join(m.stopActions, " ")
}

// linearActuatorCountsPerMeter is the counts_per_m attribute.
type linearActuatorCountsPerMeter linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorCountsPerMeter) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.countPerM)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorCountsPerMeter) Size() (int64, error) {
	return size(m.countPerM), nil
}

// linearActuatorDutyCycle is the duty_cycle attribute.
type linearActuatorDutyCycle linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorDutyCycle) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.dutyCycle)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorDutyCycle) Size() (int64, error) {
	return size(m.dutyCycle), nil
}

// linearActuatorDutyCycleSet is the duty_cycle_sp attribute.
type linearActuatorDutyCycleSet linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorDutyCycleSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.dutyCycleSet)
}

// Truncate is a no-op.
func (m *linearActuatorDutyCycleSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorDutyCycleSet) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.dutyCycleSet = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorDutyCycleSet) Size() (int64, error) {
	return size(m.dutyCycleSet), nil
}

// linearActuatorPolarity is the polarity attribute.
type linearActuatorPolarity linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorPolarity) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.polarity)
}

// Truncate is a no-op.
func (m *linearActuatorPolarity) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorPolarity) WriteAt(b []byte, off int64) (int, error) {
	p := Polarity(b)
	switch p {
	case "normal", "inversed":
		m.polarity = p
	default:
		m.t.Errorf("unexpected error: %q", b)
		return len(b), syscall.EINVAL
	}
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorPolarity) Size() (int64, error) {
	return size(m.polarity), nil
}

// linearActuatorPosition is the position attribute.
type linearActuatorPosition linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorPosition) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.position)
}

// Truncate is a no-op.
func (m *linearActuatorPosition) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorPosition) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.position = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorPosition) Size() (int64, error) {
	return size(m.position), nil
}

// linearActuatorPositionSet is the position_sp attribute.
type linearActuatorPositionSet linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorPositionSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.positionSet)
}

// Truncate is a no-op.
func (m *linearActuatorPositionSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorPositionSet) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.positionSet = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorPositionSet) Size() (int64, error) {
	return size(m.positionSet), nil
}

// linearActuatorHoldPIDkd is the hold_pid/Kd attribute.
type linearActuatorHoldPIDkd linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorHoldPIDkd) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.holdPIDkd)
}

// Truncate is a no-op.
func (m *linearActuatorHoldPIDkd) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorHoldPIDkd) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.holdPIDkd = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorHoldPIDkd) Size() (int64, error) {
	return size(m.holdPIDkd), nil
}

// linearActuatorHoldPIDki is the hold_pid/Ki attribute.
type linearActuatorHoldPIDki linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorHoldPIDki) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.holdPIDki)
}

// Truncate is a no-op.
func (m *linearActuatorHoldPIDki) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorHoldPIDki) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.holdPIDki = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorHoldPIDki) Size() (int64, error) {
	return size(m.holdPIDki), nil
}

// linearActuatorHoldPIDkp is the hold_pid/Kp attribute.
type linearActuatorHoldPIDkp linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorHoldPIDkp) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.holdPIDkp)
}

// Truncate is a no-op.
func (m *linearActuatorHoldPIDkp) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorHoldPIDkp) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.holdPIDkp = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorHoldPIDkp) Size() (int64, error) {
	return size(m.holdPIDkp), nil
}

// linearActuatorMaxSpeed is the max_speed attribute.
type linearActuatorMaxSpeed linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorMaxSpeed) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.maxSpeed)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorMaxSpeed) Size() (int64, error) {
	return size(m.maxSpeed), nil
}

// linearActuatorSpeed is the speed attribute.
type linearActuatorSpeed linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorSpeed) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.speed)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorSpeed) Size() (int64, error) {
	return size(m.speed), nil
}

// linearActuatorSpeedSet is the speed_sp attribute.
type linearActuatorSpeedSet linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorSpeedSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.speedSet)
}

// Truncate is a no-op.
func (m *linearActuatorSpeedSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorSpeedSet) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.speedSet = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorSpeedSet) Size() (int64, error) {
	return size(m.speedSet), nil
}

// linearActuatorRampUpSet is the ramp_up_sp attribute.
type linearActuatorRampUpSet linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorRampUpSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *linearActuatorRampUpSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorRampUpSet) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.rampUpSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorRampUpSet) Size() (int64, error) {
	return size(m), nil
}

func (m *linearActuatorRampUpSet) String() string {
	return fmt.Sprint(int(m.rampUpSet / time.Millisecond))
}

// linearActuatorRampDownSet is the ramp_down_sp attribute.
type linearActuatorRampDownSet linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorRampDownSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *linearActuatorRampDownSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorRampDownSet) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.rampDownSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorRampDownSet) Size() (int64, error) {
	return size(m), nil
}

func (m *linearActuatorRampDownSet) String() string {
	return fmt.Sprint(int(m.rampDownSet / time.Millisecond))
}

// linearActuatorSpeedPIDkd is the speed_pid/Kd attribute.
type linearActuatorSpeedPIDkd linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorSpeedPIDkd) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.speedPIDkd)
}

// Truncate is a no-op.
func (m *linearActuatorSpeedPIDkd) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorSpeedPIDkd) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.speedPIDkd = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorSpeedPIDkd) Size() (int64, error) {
	return size(m.speedPIDkd), nil
}

// linearActuatorSpeedPIDki is the speed_pid/Ki attribute.
type linearActuatorSpeedPIDki linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorSpeedPIDki) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.speedPIDki)
}

// Truncate is a no-op.
func (m *linearActuatorSpeedPIDki) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorSpeedPIDki) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.speedPIDki = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorSpeedPIDki) Size() (int64, error) {
	return size(m.speedPIDki), nil
}

// linearActuatorSpeedPIDkp is the speed_pid/Kp attribute.
type linearActuatorSpeedPIDkp linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorSpeedPIDkp) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.speedPIDkp)
}

// Truncate is a no-op.
func (m *linearActuatorSpeedPIDkp) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorSpeedPIDkp) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.speedPIDkp = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorSpeedPIDkp) Size() (int64, error) {
	return size(m.speedPIDkp), nil
}

// linearActuatorState is the state attribute.
type linearActuatorState linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorState) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorState) Size() (int64, error) {
	return size(m), nil
}

func (m *linearActuatorState) String() string {
	s := strings.Replace(m.state.String(), "|", " ", -1)
	if s == MotorState(0).String() {
		return ""
	}
	return s
}

// linearActuatorStopAction is the stop_actions attribute.
type linearActuatorStopAction linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorStopAction) ReadAt(b []byte, offset int64) (int, error) {
	if len(m.stopActions) == 0 {
		return len(b), syscall.ENOTSUP
	}
	return readAt(b, offset, m.lastStopAction)
}

// Truncate is a no-op.
func (m *linearActuatorStopAction) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorStopAction) WriteAt(b []byte, off int64) (int, error) {
	if len(m.stopActions) == 0 {
		return len(b), syscall.ENOTSUP
	}
	stopAction := string(chomp(b))
	for _, c := range m.stopActions {
		if stopAction == c {
			m.lastStopAction = stopAction
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorStopAction) Size() (int64, error) {
	return size(m.lastStopAction), nil
}

// linearActuatorTimeSet is the time_sp attribute.
type linearActuatorTimeSet linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorTimeSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *linearActuatorTimeSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *linearActuatorTimeSet) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m.timeSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorTimeSet) Size() (int64, error) {
	return size(m), nil
}

func (m *linearActuatorTimeSet) String() string {
	return fmt.Sprint(int(m.timeSet / time.Millisecond))
}

// linearActuatorUevent is the uevent attribute.
type linearActuatorUevent linearActuator

// ReadAt satisfies the io.ReaderAt interface.
func (m *linearActuatorUevent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *linearActuatorUevent) Size() (int64, error) {
	return size(m), nil
}

func (m *linearActuatorUevent) String() string {
	e := make([]string, 0, len(m.uevent))
	for k, v := range m.uevent {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(e)
	return strings.Join(e, "\n")
}

type linearActuatorConn struct {
	id             int
	linearActuator *linearActuator
}

func connectedLinearActuators(c ...linearActuatorConn) []sisyphus.Node {
	n := make([]sisyphus.Node, len(c))
	for i, m := range c {
		n[i] = d(fmt.Sprintf("linear%d", m.id), 0775).With(
			ro(AddressName, 0444, (*linearActuatorAddress)(m.linearActuator)),
			ro(DriverNameName, 0444, (*linearActuatorDriver)(m.linearActuator)),
			ro(CommandsName, 0444, (*linearActuatorCommands)(m.linearActuator)),
			wo(CommandName, 0222, (*linearActuatorCommand)(m.linearActuator)),
			ro(CountPerMeterName, 0444, (*linearActuatorCountsPerMeter)(m.linearActuator)),
			rw(PolarityName, 0666, (*linearActuatorPolarity)(m.linearActuator)),
			ro(DutyCycleName, 0444, (*linearActuatorDutyCycle)(m.linearActuator)),
			rw(DutyCycleSetpointName, 0666, (*linearActuatorDutyCycleSet)(m.linearActuator)),
			rw(PositionName, 0666, (*linearActuatorPosition)(m.linearActuator)),
			rw(PositionSetpointName, 0666, (*linearActuatorPositionSet)(m.linearActuator)),
			d(HoldPIDName, 777).With(
				rw(KdName, 0666, (*linearActuatorHoldPIDkd)(m.linearActuator)),
				rw(KiName, 0666, (*linearActuatorHoldPIDki)(m.linearActuator)),
				rw(KpName, 0666, (*linearActuatorHoldPIDkp)(m.linearActuator)),
			),
			ro(MaxSpeedName, 0444, (*linearActuatorMaxSpeed)(m.linearActuator)),
			ro(SpeedName, 0444, (*linearActuatorSpeed)(m.linearActuator)),
			rw(SpeedSetpointName, 0666, (*linearActuatorSpeedSet)(m.linearActuator)),
			rw(RampUpSetpointName, 0666, (*linearActuatorRampUpSet)(m.linearActuator)),
			rw(RampDownSetpointName, 0666, (*linearActuatorRampDownSet)(m.linearActuator)),
			d(SpeedPIDName, 777).With(
				rw(KdName, 0666, (*linearActuatorSpeedPIDkd)(m.linearActuator)),
				rw(KiName, 0666, (*linearActuatorSpeedPIDki)(m.linearActuator)),
				rw(KpName, 0666, (*linearActuatorSpeedPIDkp)(m.linearActuator)),
			),
			ro(StateName, 0444, (*linearActuatorState)(m.linearActuator)),
			ro(StopActionsName, 0444, (*linearActuatorStopActions)(m.linearActuator)),
			rw(StopActionName, 0666, (*linearActuatorStopAction)(m.linearActuator)),
			rw(TimeSetpointName, 0666, (*linearActuatorTimeSet)(m.linearActuator)),
			ro(UeventName, 0444, (*linearActuatorUevent)(m.linearActuator)),
		)
	}
	return n
}

func linearactuatorsysfs(m ...linearActuatorConn) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("tacho-motor", 0775).With(
					connectedLinearActuators(m...)...,
				),
			),
		),
	).Sync()
}

func TestLinearActuator(t *testing.T) {
	const driver = "lego-ev3-gyro"
	conn := []linearActuatorConn{
		{
			id: 5,
			linearActuator: &linearActuator{
				address: "outA",
				driver:  driver,

				commands: []string{
					"run-forever",
					"run-to-abs-pos",
					"run-to-rel-pos",
					"run-timed",
					"run-direct",
					"stop",
					"reset",
				},

				lastStopAction: "coast",
				stopActions: []string{
					"coast",
					"brake",
					"hold",
				},

				uevent: map[string]string{
					"LEGO_ADDRESS":     "outA",
					"LEGO_DRIVER_NAME": "lact-l12-ev3-100",
				},

				t: t,
			},
		},
		{
			id: 7,
			linearActuator: &linearActuator{
				address: "outB",
				driver:  driver,

				t: t,
			},
		},
	}

	fs := linearactuatorsysfs(conn...)
	unmount := serve(fs, t)
	defer unmount()

	t.Run("new LinearActuator", func(t *testing.T) {
		for _, r := range []struct{ port, driver string }{
			{port: "", driver: conn[0].linearActuator.driver},
			{port: conn[0].linearActuator.address, driver: conn[0].linearActuator.driver},
			{port: conn[0].linearActuator.address, driver: ""},
		} {
			got, err := LinearActuatorFor(r.port, r.driver)
			if r.driver == driver {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				merr, ok := err.(DriverMismatch)
				if !ok {
					t.Errorf("unexpected error type for driver mismatch: got:%T want:%T", err, merr)
				}
				if merr.Have != driver {
					t.Errorf("unexpected value for have driver error: got:%q want:%q", merr.Have, conn[0].linearActuator.driver)
				}
			}
			ok, err := IsConnected(got)
			if err != nil {
				t.Errorf("unexpected error getting connection status:%v", err)
			}
			if !ok {
				t.Error("expected device to be connected")
			}
			gotAddr, err := AddressOf(got)
			if err != nil {
				t.Errorf("unexpected error getting address: %v", err)
			}
			wantAddr := conn[0].linearActuator.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := conn[0].linearActuator.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Next", func(t *testing.T) {
		m, err := LinearActuatorFor(conn[0].linearActuator.address, conn[0].linearActuator.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := m.Next()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ok, err := IsConnected(got)
		if err != nil {
			t.Errorf("unexpected error getting connection status:%v", err)
		}
		if !ok {
			t.Error("expected device to be connected")
		}
		gotAddr, err := AddressOf(got)
		if err != nil {
			t.Errorf("unexpected error getting address: %v", err)
		}
		wantAddr := conn[1].linearActuator.address
		if gotAddr != wantAddr {
			t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
		}
		gotDriver, err := DriverFor(got)
		if err != nil {
			t.Errorf("unexpected error getting driver name:%v", err)
		}
		wantDriver := conn[1].linearActuator.driver
		if gotDriver != wantDriver {
			t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
		}
	})

	t.Run("FindAfter", func(t *testing.T) {
		var last *LinearActuator
		for _, c := range conn {
			got := new(LinearActuator)
			err := FindAfter(last, got, driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			last = got
			ok, err := IsConnected(got)
			if err != nil {
				t.Errorf("unexpected error getting connection status:%v", err)
			}
			if !ok {
				t.Error("expected device to be connected")
			}
			gotAddr, err := AddressOf(got)
			if err != nil {
				t.Errorf("unexpected error getting address: %v", err)
			}
			wantAddr := c.linearActuator.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := c.linearActuator.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Command", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			commands, err := m.Commands()
			if len(c.linearActuator.commands) == 0 {
				if err == nil {
					t.Error("expected error getting commands from non-commandable linearActuator")
				}
				continue
			}
			if err != nil {
				t.Fatalf("unexpected error getting commands: %v", err)
			}
			if !reflect.DeepEqual(commands, c.linearActuator.commands) {
				t.Errorf("unexpected commands value: got:%q want:%q", commands, c.linearActuator.commands)
			}
			for _, command := range commands {
				err := m.Command(command).Err()
				if err != nil {
					t.Errorf("unexpected error for command %q: %v", command, err)
				}

				got := c.linearActuator.lastCommand
				want := command
				if got != want {
					t.Errorf("unexpected command value: got:%q want:%q", got, want)
				}
			}
			for _, command := range []string{"invalid", "another"} {
				err := m.Command(command).Err()
				if err == nil {
					t.Errorf("expected error for command %q", command)
				}

				got := c.linearActuator.lastCommand
				dontwant := command
				if got == dontwant {
					t.Errorf("unexpected invalid command value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Count per meter", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, c.linearActuator.countPerM = range []int{0, 64, 128, 192, 255} {
				got, err := m.CountPerMeter()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.linearActuator.countPerM
				if got != want {
					t.Errorf("unexpected count per meter value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Duty cycle", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, c.linearActuator.dutyCycle = range []int{0, 64, 128, 192, 255} {
				got, err := m.DutyCycle()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.linearActuator.dutyCycle
				if got != want {
					t.Errorf("unexpected duty cycle value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Duty cycle setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetDutyCycleSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for duty cycle setpoint %d: %v", v, err)
				}

				got, err := m.DutyCycleSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected duty cycle setpoint value: got:%d want:%d", got, want)
				}
			}
			for _, v := range []int{-101, 101} {
				err := m.SetDutyCycleSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for duty cycle setpoint %d", v)
				}
			}
		}
	})

	t.Run("Polarity", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, polarity := range []Polarity{"normal", "inversed"} {
				err := m.SetPolarity(polarity).Err()
				if err != nil {
					t.Errorf("unexpected error for set polarity %q: %v", polarity, err)
				}

				got, err := m.Polarity()
				if err != nil {
					t.Errorf("unexpected error for polarity %q: %v", polarity, err)
				}
				want := polarity
				if got != want {
					t.Errorf("unexpected polarity value: got:%q want:%q", got, want)
				}
			}
			for _, polarity := range []Polarity{"invalid", "another"} {
				err := m.SetPolarity(polarity).Err()
				if err == nil {
					t.Errorf("expected error for set polarity %q", polarity)
				}

				got, err := m.Polarity()
				if err != nil {
					t.Errorf("unexpected error for polarity %q: %v", polarity, err)
				}
				dontwant := polarity
				if got == dontwant {
					t.Errorf("unexpected invalid polarity value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Position", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetPosition(v).Err()
				if err != nil {
					t.Errorf("unexpected error for set position %d: %v", v, err)
				}

				got, err := m.Position()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected position value: got:%d want:%d", got, want)
				}
			}
			for _, v := range []int64{-2147483649, 2147483648} {
				if int64(int(v)) != v {
					continue
				}
				err := m.SetPosition(int(v)).Err()
				if err == nil {
					t.Errorf("expected error for set position %d", v)
				}
			}
		}
	})

	t.Run("Position setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetPositionSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for set position setpoint %d: %v", v, err)
				}

				got, err := m.PositionSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected position setpoint value: got:%d want:%d", got, want)
				}
			}
			for _, v := range []int64{-2147483649, 2147483648} {
				if int64(int(v)) != v {
					continue
				}
				err := m.SetPositionSetpoint(int(v)).Err()
				if err == nil {
					t.Errorf("expected error for set position setpoint %d", v)
				}
			}
		}
	})

	t.Run("Hold PID Kd", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetHoldPIDKd(v).Err()
				if err != nil {
					t.Errorf("unexpected error for hold PID Kd %d: %v", v, err)
				}

				got, err := m.HoldPIDKd()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected hold PID Kd value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Hold PID Ki", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetHoldPIDKi(v).Err()
				if err != nil {
					t.Errorf("unexpected error for hold PID Ki %d: %v", v, err)
				}

				got, err := m.HoldPIDKi()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected hold PID Ki value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Hold PID Kp", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetHoldPIDKp(v).Err()
				if err != nil {
					t.Errorf("unexpected error for hold PID Kp %d: %v", v, err)
				}

				got, err := m.HoldPIDKp()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected hold PID Kp value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Max speed", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, c.linearActuator.maxSpeed = range []int{0, 64, 128, 192, 255} {
				got, err := m.MaxSpeed()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.linearActuator.maxSpeed
				if got != want {
					t.Errorf("unexpected max speed value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Speed", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, c.linearActuator.speed = range []int{0, 64, 128, 192, 255} {
				got, err := m.Speed()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.linearActuator.speed
				if got != want {
					t.Errorf("unexpected speed value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Speed setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetSpeedSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for speed setpoint %d: %v", v, err)
				}

				got, err := m.SpeedSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected speed setpoint value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Ramp up setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{time.Millisecond, time.Second, time.Minute} {
				err := m.SetRampUpSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for ramp up setpoint %d: %v", v, err)
				}

				got, err := m.RampUpSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected ramp up setpoint value: got:%v want:%v", got, want)
				}
			}
			for _, v := range []time.Duration{-time.Millisecond, -time.Second, -time.Minute} {
				err := m.SetRampUpSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set position setpoint %d", v)
				}
			}
		}
	})

	t.Run("Ramp down setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{time.Millisecond, time.Second, time.Minute} {
				err := m.SetRampDownSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for ramp down setpoint %d: %v", v, err)
				}

				got, err := m.RampDownSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected ramp down setpoint value: got:%v want:%v", got, want)
				}
			}
			for _, v := range []time.Duration{-time.Millisecond, -time.Second, -time.Minute} {
				err := m.SetRampDownSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set position setpoint %d", v)
				}
			}
		}
	})

	t.Run("Speed PID Kd", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetSpeedPIDKd(v).Err()
				if err != nil {
					t.Errorf("unexpected error for speed PID Kd %d: %v", v, err)
				}

				got, err := m.SpeedPIDKd()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected speed PID Kd value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Speed PID Ki", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetSpeedPIDKi(v).Err()
				if err != nil {
					t.Errorf("unexpected error for speed PID Ki %d: %v", v, err)
				}

				got, err := m.SpeedPIDKi()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected speed PID Ki value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Speed PID Kp", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []int{-100, -50, 0, 50, 100} {
				err := m.SetSpeedPIDKp(v).Err()
				if err != nil {
					t.Errorf("unexpected error for speed PID Kp %d: %v", v, err)
				}

				got, err := m.SpeedPIDKp()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected speed PID Kp value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("State", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, c.linearActuator.state = range []MotorState{
				0,
				Running,
				Running | Ramping,
				Running | Stalled,
				Running | Overloaded,
				Running | Stalled | Overloaded,
				Holding,
			} {
				got, err := m.State()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.linearActuator.state
				if got != want {
					t.Errorf("unexpected state value: got:%v want:%v", got, want)
				}
			}
		}
	})

	t.Run("Stop action", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			stopActions, err := m.StopActions()
			if len(c.linearActuator.stopActions) == 0 {
				if err == nil {
					t.Error("expected error getting stop actions from non-stop action linearActuator")
				}
				continue
			}
			if err != nil {
				t.Fatalf("unexpected error getting stop actions: %v", err)
			}
			if !reflect.DeepEqual(stopActions, c.linearActuator.stopActions) {
				t.Errorf("unexpected stop actions value: got:%q want:%q", stopActions, c.linearActuator.stopActions)
			}
			for _, stopAction := range stopActions {
				err := m.SetStopAction(stopAction).Err()
				if err != nil {
					t.Errorf("unexpected error for set stop action %q: %v", stopAction, err)
				}

				got := c.linearActuator.lastStopAction
				want := stopAction
				if got != want {
					t.Errorf("unexpected stop action value: got:%q want:%q", got, want)
				}

				got, err = m.StopAction()
				if err != nil {
					t.Errorf("unexpected error for stop action %q: %v", stopAction, err)
				}
				if got != want {
					t.Errorf("unexpected stop action value: got:%q want:%q", got, want)
				}
			}
			for _, stopAction := range []string{"invalid", "another"} {
				err := m.SetStopAction(stopAction).Err()
				if err == nil {
					t.Errorf("expected error for set stop action %q", stopAction)
				}

				got := c.linearActuator.lastStopAction
				dontwant := stopAction
				if got == dontwant {
					t.Errorf("unexpected invalid stop action value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Time setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{time.Millisecond, time.Second, time.Minute} {
				err := m.SetTimeSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for time setpoint %d: %v", v, err)
				}

				got, err := m.TimeSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected time setpoint value: got:%v want:%v", got, want)
				}
			}
			for _, v := range []time.Duration{-time.Millisecond, -time.Second, -time.Minute} {
				err := m.SetTimeSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set time setpoint %d", v)
				}
			}
		}
	})

	t.Run("Uevent", func(t *testing.T) {
		for _, c := range conn {
			m, err := LinearActuatorFor(c.linearActuator.address, c.linearActuator.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := m.Uevent()
			if err != nil {
				t.Errorf("unexpected error getting uevent: %v", err)
			}
			want := c.linearActuator.uevent
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
			}
		}
	})
}
