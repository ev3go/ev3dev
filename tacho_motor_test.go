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
	"sync"
	"syscall"
	"testing"
	"time"

	. "github.com/ev3go/ev3dev"

	"github.com/ev3go/sisyphus"
)

// tachoMotor is a tachoMotor sysfs directory.
type tachoMotor struct {
	address string
	driver  string

	// mu protects the underscore
	// prefix attributes below.
	mu sync.Mutex

	_lastCommand string
	_commands    []string

	_countPerRot int

	_maxSpeed int
	_speed    int
	_speedSet int

	_rampUpSet   time.Duration
	_rampDownSet time.Duration

	_timeSet time.Duration

	_dutyCycle    int
	_dutyCycleSet int

	_polarity Polarity

	_position    int
	_positionSet int

	_holdPIDkd int
	_holdPIDki int
	_holdPIDkp int

	_speedPIDkd int
	_speedPIDki int
	_speedPIDkp int

	_state MotorState

	_lastStopAction string
	_stopActions    []string

	_uevent map[string]string

	t *testing.T
}

func (m *tachoMotor) commands() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._commands
}

func (m *tachoMotor) lastCommand() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastCommand
}

func (m *tachoMotor) countsPerRot() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._countPerRot
}

func (m *tachoMotor) setCountsPerRot(n int) {
	m.mu.Lock()
	m._countPerRot = n
	m.mu.Unlock()
}

func (m *tachoMotor) maxSpeed() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._maxSpeed
}

func (m *tachoMotor) setMaxSpeed(s int) {
	m.mu.Lock()
	m._maxSpeed = s
	m.mu.Unlock()
}

func (m *tachoMotor) speed() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._speed
}

func (m *tachoMotor) setSpeed(s int) {
	m.mu.Lock()
	m._speed = s
	m.mu.Unlock()
}

func (m *tachoMotor) dutyCycle() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._dutyCycle
}

func (m *tachoMotor) setDutyCycle(n int) {
	m.mu.Lock()
	m._dutyCycle = n
	m.mu.Unlock()
}

func (m *tachoMotor) state() MotorState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._state
}

func (m *tachoMotor) setState(s MotorState) {
	m.mu.Lock()
	m._state = s
	m.mu.Unlock()
}

func (m *tachoMotor) lastStopAction() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastStopAction
}

func (m *tachoMotor) stopActions() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._stopActions
}

func (m *tachoMotor) uevent() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._uevent
}

// tachoMotorAddress is the address attribute.
type tachoMotorAddress tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorAddress) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.address)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorAddress) Size() (int64, error) {
	return size(m.address), nil
}

// tachoMotorDriver is the driver_name attribute.
type tachoMotorDriver tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorDriver) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.driver)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorDriver) Size() (int64, error) {
	return size(m.driver), nil
}

// tachoMotorCommands is the commands attribute.
type tachoMotorCommands tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorCommands) ReadAt(b []byte, offset int64) (int, error) {
	if len((*tachoMotor)(m).commands()) == 0 {
		return len(b), syscall.ENOTSUP
	}
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorCommands) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorCommands) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	sort.Strings(m._commands)
	return strings.Join(m._commands, " ")
}

// tachoMotorCommand is the command attribute.
type tachoMotorCommand tachoMotor

// Truncate is a no-op.
func (m *tachoMotorCommand) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorCommand) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m._commands) == 0 {
		return len(b), syscall.ENOTSUP
	}
	command := string(chomp(b))
	for _, c := range m._commands {
		if command == c {
			m._lastCommand = command
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorCommand) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorCommand) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastCommand
}

// tachoMotorStopActions is the stop_sactions attribute.
type tachoMotorStopActions tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorStopActions) ReadAt(b []byte, offset int64) (int, error) {
	if len((*tachoMotor)(m).stopActions()) == 0 {
		return len(b), syscall.ENOTSUP
	}
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorStopActions) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorStopActions) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	sort.Strings(m._stopActions)
	return strings.Join(m._stopActions, " ")
}

// tachoMotorCountsPerRot is the counts_per_rot attribute.
type tachoMotorCountsPerRot tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorCountsPerRot) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorCountsPerRot) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorCountsPerRot) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._countPerRot)
}

// tachoMotorDutyCycle is the duty_cycle attribute.
type tachoMotorDutyCycle tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorDutyCycle) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorDutyCycle) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorDutyCycle) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._dutyCycle)
}

// tachoMotorDutyCycleSet is the duty_cycle_sp attribute.
type tachoMotorDutyCycleSet tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorDutyCycleSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorDutyCycleSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorDutyCycleSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._dutyCycleSet = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorDutyCycleSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorDutyCycleSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._dutyCycleSet)
}

// tachoMotorPolarity is the polarity attribute.
type tachoMotorPolarity tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorPolarity) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorPolarity) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorPolarity) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := Polarity(b)
	switch p {
	case "normal", "inversed":
		m._polarity = p
	default:
		m.t.Errorf("unexpected error: %q", b)
		return len(b), syscall.EINVAL
	}
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorPolarity) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorPolarity) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return string(m._polarity)
}

// tachoMotorPosition is the position attribute.
type tachoMotorPosition tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorPosition) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorPosition) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorPosition) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._position = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorPosition) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorPosition) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._position)
}

// tachoMotorPositionSet is the position_sp attribute.
type tachoMotorPositionSet tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorPositionSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorPositionSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorPositionSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._positionSet = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorPositionSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorPositionSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._positionSet)
}

// tachoMotorHoldPIDkd is the hold_pid/Kd attribute.
type tachoMotorHoldPIDkd tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorHoldPIDkd) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorHoldPIDkd) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorHoldPIDkd) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._holdPIDkd = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorHoldPIDkd) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorHoldPIDkd) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._holdPIDkd)
}

// tachoMotorHoldPIDki is the hold_pid/Ki attribute.
type tachoMotorHoldPIDki tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorHoldPIDki) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorHoldPIDki) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorHoldPIDki) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._holdPIDki = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorHoldPIDki) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorHoldPIDki) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._holdPIDki)
}

// tachoMotorHoldPIDkp is the hold_pid/Kp attribute.
type tachoMotorHoldPIDkp tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorHoldPIDkp) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorHoldPIDkp) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorHoldPIDkp) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._holdPIDkp = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorHoldPIDkp) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorHoldPIDkp) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._holdPIDkp)
}

// tachoMotorMaxSpeed is the max_speed attribute.
type tachoMotorMaxSpeed tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorMaxSpeed) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorMaxSpeed) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorMaxSpeed) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._maxSpeed)
}

// tachoMotorSpeed is the speed attribute.
type tachoMotorSpeed tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorSpeed) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorSpeed) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorSpeed) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._speed)
}

// tachoMotorSpeedSet is the speed_sp attribute.
type tachoMotorSpeedSet tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorSpeedSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorSpeedSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorSpeedSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._speedSet = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorSpeedSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorSpeedSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._speedSet)
}

// tachoMotorRampUpSet is the ramp_up_sp attribute.
type tachoMotorRampUpSet tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorRampUpSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorRampUpSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorRampUpSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._rampUpSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorRampUpSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorRampUpSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(int(m._rampUpSet / time.Millisecond))
}

// tachoMotorRampDownSet is the ramp_down_sp attribute.
type tachoMotorRampDownSet tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorRampDownSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorRampDownSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorRampDownSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._rampDownSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorRampDownSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorRampDownSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(int(m._rampDownSet / time.Millisecond))
}

// tachoMotorSpeedPIDkd is the speed_pid/Kd attribute.
type tachoMotorSpeedPIDkd tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorSpeedPIDkd) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorSpeedPIDkd) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorSpeedPIDkd) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._speedPIDkd = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorSpeedPIDkd) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorSpeedPIDkd) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._speedPIDkd)
}

// tachoMotorSpeedPIDki is the speed_pid/Ki attribute.
type tachoMotorSpeedPIDki tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorSpeedPIDki) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorSpeedPIDki) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorSpeedPIDki) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._speedPIDki = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorSpeedPIDki) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorSpeedPIDki) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._speedPIDki)
}

// tachoMotorSpeedPIDkp is the speed_pid/Kp attribute.
type tachoMotorSpeedPIDkp tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorSpeedPIDkp) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorSpeedPIDkp) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorSpeedPIDkp) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._speedPIDkp = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorSpeedPIDkp) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorSpeedPIDkp) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._speedPIDkp)
}

// tachoMotorState is the state attribute.
type tachoMotorState tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorState) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorState) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorState) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := strings.Replace(m._state.String(), "|", " ", -1)
	if s == MotorState(0).String() {
		return ""
	}
	return s
}

// tachoMotorStopAction is the stop_actions attribute.
type tachoMotorStopAction tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorStopAction) ReadAt(b []byte, offset int64) (int, error) {
	if len((*tachoMotor)(m).stopActions()) == 0 {
		return len(b), syscall.ENOTSUP
	}
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorStopAction) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorStopAction) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m._stopActions) == 0 {
		return len(b), syscall.ENOTSUP
	}
	stopAction := string(chomp(b))
	for _, c := range m._stopActions {
		if stopAction == c {
			m._lastStopAction = stopAction
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorStopAction) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorStopAction) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastStopAction
}

// tachoMotorTimeSet is the time_sp attribute.
type tachoMotorTimeSet tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorTimeSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *tachoMotorTimeSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *tachoMotorTimeSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._timeSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorTimeSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorTimeSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(int(m._timeSet / time.Millisecond))
}

// tachoMotorUevent is the uevent attribute.
type tachoMotorUevent tachoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *tachoMotorUevent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *tachoMotorUevent) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *tachoMotorUevent) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := make([]string, 0, len(m._uevent))
	for k, v := range m._uevent {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(e)
	return strings.Join(e, "\n")
}

type tachoMotorConn struct {
	id         int
	tachoMotor *tachoMotor
}

func connectedTachoMotors(c ...tachoMotorConn) []sisyphus.Node {
	n := make([]sisyphus.Node, len(c))
	for i, m := range c {
		n[i] = d(fmt.Sprintf("motor%d", m.id), 0775).With(
			ro(AddressName, 0444, (*tachoMotorAddress)(m.tachoMotor)),
			ro(DriverNameName, 0444, (*tachoMotorDriver)(m.tachoMotor)),
			ro(CommandsName, 0444, (*tachoMotorCommands)(m.tachoMotor)),
			wo(CommandName, 0222, (*tachoMotorCommand)(m.tachoMotor)),
			ro(CountPerRotName, 0444, (*tachoMotorCountsPerRot)(m.tachoMotor)),
			rw(PolarityName, 0666, (*tachoMotorPolarity)(m.tachoMotor)),
			ro(DutyCycleName, 0444, (*tachoMotorDutyCycle)(m.tachoMotor)),
			rw(DutyCycleSetpointName, 0666, (*tachoMotorDutyCycleSet)(m.tachoMotor)),
			rw(PositionName, 0666, (*tachoMotorPosition)(m.tachoMotor)),
			rw(PositionSetpointName, 0666, (*tachoMotorPositionSet)(m.tachoMotor)),
			d(HoldPIDName, 777).With(
				rw(KdName, 0666, (*tachoMotorHoldPIDkd)(m.tachoMotor)),
				rw(KiName, 0666, (*tachoMotorHoldPIDki)(m.tachoMotor)),
				rw(KpName, 0666, (*tachoMotorHoldPIDkp)(m.tachoMotor)),
			),
			ro(MaxSpeedName, 0444, (*tachoMotorMaxSpeed)(m.tachoMotor)),
			ro(SpeedName, 0444, (*tachoMotorSpeed)(m.tachoMotor)),
			rw(SpeedSetpointName, 0666, (*tachoMotorSpeedSet)(m.tachoMotor)),
			rw(RampUpSetpointName, 0666, (*tachoMotorRampUpSet)(m.tachoMotor)),
			rw(RampDownSetpointName, 0666, (*tachoMotorRampDownSet)(m.tachoMotor)),
			d(SpeedPIDName, 777).With(
				rw(KdName, 0666, (*tachoMotorSpeedPIDkd)(m.tachoMotor)),
				rw(KiName, 0666, (*tachoMotorSpeedPIDki)(m.tachoMotor)),
				rw(KpName, 0666, (*tachoMotorSpeedPIDkp)(m.tachoMotor)),
			),
			ro(StateName, 0444, (*tachoMotorState)(m.tachoMotor)),
			ro(StopActionsName, 0444, (*tachoMotorStopActions)(m.tachoMotor)),
			rw(StopActionName, 0666, (*tachoMotorStopAction)(m.tachoMotor)),
			rw(TimeSetpointName, 0666, (*tachoMotorTimeSet)(m.tachoMotor)),
			ro(UeventName, 0444, (*tachoMotorUevent)(m.tachoMotor)),
		)
	}
	return n
}

func tachomotorsysfs(m ...tachoMotorConn) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("tacho-motor", 0775).With(
					connectedTachoMotors(m...)...,
				),
			),
		),
	).Sync()
}

func TestTachoMotor(t *testing.T) {
	const driver = "lego-ev3-l-motor"
	conn := []tachoMotorConn{
		{
			id: 5,
			tachoMotor: &tachoMotor{
				address: "outA",
				driver:  driver,

				_commands: []string{
					"run-forever",
					"run-to-abs-pos",
					"run-to-rel-pos",
					"run-timed",
					"run-direct",
					"stop",
					"reset",
				},

				_lastStopAction: "coast",
				_stopActions: []string{
					"coast",
					"brake",
					"hold",
				},

				_uevent: map[string]string{
					"LEGO_ADDRESS":     "outA",
					"LEGO_DRIVER_NAME": driver,
				},

				t: t,
			},
		},
		{
			id: 7,
			tachoMotor: &tachoMotor{
				address: "outB",
				driver:  driver,

				t: t,
			},
		},
	}

	fs := tachomotorsysfs(conn...)
	unmount := serve(fs, t)
	defer unmount()

	t.Run("new TachoMotor", func(t *testing.T) {
		for _, r := range []struct{ port, driver string }{
			{port: "", driver: conn[0].tachoMotor.driver},
			{port: conn[0].tachoMotor.address, driver: conn[0].tachoMotor.driver},
			{port: conn[0].tachoMotor.address, driver: ""},
		} {
			got, err := TachoMotorFor(r.port, r.driver)
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
					t.Errorf("unexpected value for have driver error: got:%q want:%q", merr.Have, conn[0].tachoMotor.driver)
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
			wantAddr := conn[0].tachoMotor.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := conn[0].tachoMotor.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Next", func(t *testing.T) {
		m, err := TachoMotorFor(conn[0].tachoMotor.address, conn[0].tachoMotor.driver)
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
		wantAddr := conn[1].tachoMotor.address
		if gotAddr != wantAddr {
			t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
		}
		gotDriver, err := DriverFor(got)
		if err != nil {
			t.Errorf("unexpected error getting driver name:%v", err)
		}
		wantDriver := conn[1].tachoMotor.driver
		if gotDriver != wantDriver {
			t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
		}
	})

	t.Run("FindAfter", func(t *testing.T) {
		var last *TachoMotor
		for _, c := range conn {
			got := new(TachoMotor)
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
			wantAddr := c.tachoMotor.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := c.tachoMotor.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Command", func(t *testing.T) {
		for _, c := range conn {
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			commands, err := m.Commands()
			want := c.tachoMotor.commands()
			if len(want) == 0 {
				if err == nil {
					t.Error("expected error getting commands from non-commandable tachoMotor")
				}
				continue
			}
			if err != nil {
				t.Fatalf("unexpected error getting commands: %v", err)
			}
			if !reflect.DeepEqual(commands, want) {
				t.Errorf("unexpected commands value: got:%q want:%q", commands, want)
			}
			for _, command := range commands {
				err := m.Command(command).Err()
				if err != nil {
					t.Errorf("unexpected error for command %q: %v", command, err)
				}

				got := c.tachoMotor.lastCommand()
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

				got := c.tachoMotor.lastCommand()
				dontwant := command
				if got == dontwant {
					t.Errorf("unexpected invalid command value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Count per rot", func(t *testing.T) {
		for _, c := range conn {
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, n := range []int{0, 64, 128, 192, 255} {
				c.tachoMotor.setCountsPerRot(n)
				got, err := m.CountPerRot()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.tachoMotor.countsPerRot()
				if got != want {
					t.Errorf("unexpected count per rot value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Duty cycle", func(t *testing.T) {
		for _, c := range conn {
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, p := range []int{0, 64, 128, 192, 255} {
				c.tachoMotor.setDutyCycle(p)
				got, err := m.DutyCycle()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.tachoMotor.dutyCycle()
				if got != want {
					t.Errorf("unexpected duty cycle value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Duty cycle setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, s := range []int{0, 64, 128, 192, 255} {
				c.tachoMotor.setMaxSpeed(s)
				got, err := m.MaxSpeed()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.tachoMotor.maxSpeed()
				if got != want {
					t.Errorf("unexpected max speed value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Speed", func(t *testing.T) {
		for _, c := range conn {
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, s := range []int{0, 64, 128, 192, 255} {
				c.tachoMotor.setSpeed(s)
				got, err := m.Speed()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.tachoMotor.speed()
				if got != want {
					t.Errorf("unexpected speed value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Speed setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, s := range []MotorState{
				0,
				Running,
				Running | Ramping,
				Running | Stalled,
				Running | Overloaded,
				Running | Stalled | Overloaded,
				Holding,
			} {
				c.tachoMotor.setState(s)
				got, err := m.State()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.tachoMotor.state()
				if got != want {
					t.Errorf("unexpected state value: got:%v want:%v", got, want)
				}
			}
		}
	})

	t.Run("Stop action", func(t *testing.T) {
		for _, c := range conn {
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			stopActions, err := m.StopActions()
			want := c.tachoMotor.stopActions()
			if len(want) == 0 {
				if err == nil {
					t.Error("expected error getting stop actions from non-stop action tachoMotor")
				}
				continue
			}
			if err != nil {
				t.Fatalf("unexpected error getting stop actions: %v", err)
			}
			if !reflect.DeepEqual(stopActions, want) {
				t.Errorf("unexpected stop actions value: got:%q want:%q", stopActions, want)
			}
			for _, stopAction := range stopActions {
				err := m.SetStopAction(stopAction).Err()
				if err != nil {
					t.Errorf("unexpected error for set stop action %q: %v", stopAction, err)
				}

				got := c.tachoMotor.lastStopAction()
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

				got := c.tachoMotor.lastStopAction()
				dontwant := stopAction
				if got == dontwant {
					t.Errorf("unexpected invalid stop action value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Time setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
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
			m, err := TachoMotorFor(c.tachoMotor.address, c.tachoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := m.Uevent()
			if err != nil {
				t.Errorf("unexpected error getting uevent: %v", err)
			}
			want := c.tachoMotor.uevent()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
			}
		}
	})
}
