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

// dcMotor is a dcMotor sysfs directory.
type dcMotor struct {
	address string
	driver  string

	// mu protects the underscore
	// prefix attributes below.
	mu sync.Mutex

	_lastCommand string
	_commands    []string

	_rampUpSet   time.Duration
	_rampDownSet time.Duration

	_timeSet time.Duration

	_dutyCycle    int
	_dutyCycleSet int

	_polarity Polarity

	_state MotorState

	_lastStopAction string
	_stopActions    []string

	_uevent map[string]string

	t *testing.T
}

func (m *dcMotor) commands() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._commands
}

func (m *dcMotor) lastCommand() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastCommand
}

func (m *dcMotor) dutyCycle() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._dutyCycle
}

func (m *dcMotor) setDutyCycle(sp int) {
	m.mu.Lock()
	m._dutyCycle = sp
	m.mu.Unlock()
}

func (m *dcMotor) state() MotorState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._state
}

func (m *dcMotor) setState(s MotorState) {
	m.mu.Lock()
	m._state = s
	m.mu.Unlock()
}

func (m *dcMotor) lastStopAction() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastStopAction
}

func (m *dcMotor) stopActions() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._stopActions
}

func (m *dcMotor) uevent() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._uevent
}

// dcMotorAddress is the address attribute.
type dcMotorAddress dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorAddress) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.address)
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorAddress) Size() (int64, error) {
	return size(m.address), nil
}

// dcMotorDriver is the driver_name attribute.
type dcMotorDriver dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorDriver) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.driver)
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorDriver) Size() (int64, error) {
	return size(m.driver), nil
}

// dcMotorCommands is the commands attribute.
type dcMotorCommands dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorCommands) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorCommands) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorCommands) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	sort.Strings(m._commands)
	return strings.Join(m._commands, " ")
}

// dcMotorCommand is the command attribute.
type dcMotorCommand dcMotor

// Truncate is a no-op.
func (m *dcMotorCommand) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *dcMotorCommand) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
func (m *dcMotorCommand) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorCommand) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastCommand
}

// dcMotorStopActions is the stop_actions attribute.
type dcMotorStopActions dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorStopActions) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorStopActions) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorStopActions) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	sort.Strings(m._stopActions)
	return strings.Join(m._stopActions, " ")
}

// dcMotorDutyCycle is the duty_cycle attribute.
type dcMotorDutyCycle dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorDutyCycle) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorDutyCycle) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorDutyCycle) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._dutyCycle)
}

// dcMotorDutyCycleSet is the duty_cycle_sp attribute.
type dcMotorDutyCycleSet dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorDutyCycleSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *dcMotorDutyCycleSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *dcMotorDutyCycleSet) WriteAt(b []byte, off int64) (int, error) {
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
func (m *dcMotorDutyCycleSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorDutyCycleSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(m._dutyCycleSet)
}

// dcMotorPolarity is the polarity attribute.
type dcMotorPolarity dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorPolarity) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *dcMotorPolarity) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *dcMotorPolarity) WriteAt(b []byte, off int64) (int, error) {
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
func (m *dcMotorPolarity) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorPolarity) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return string(m._polarity)
}

// dcMotorRampUpSet is the ramp_up_sp attribute.
type dcMotorRampUpSet dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorRampUpSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *dcMotorRampUpSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *dcMotorRampUpSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if time.Duration(i)*time.Millisecond > 10*time.Second {
		err = errors.New("ev3dev: duration out of range")
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._rampUpSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorRampUpSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorRampUpSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(int(m._rampUpSet / time.Millisecond))
}

// dcMotorRampDownSet is the ramp_down_sp attribute.
type dcMotorRampDownSet dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorRampDownSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *dcMotorRampDownSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *dcMotorRampDownSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if time.Duration(i)*time.Millisecond > 10*time.Second {
		err = errors.New("ev3dev: duration out of range")
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._rampDownSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorRampDownSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorRampDownSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(int(m._rampDownSet / time.Millisecond))
}

// dcMotorState is the state attribute.
type dcMotorState dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorState) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorState) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorState) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := strings.Replace(m._state.String(), "|", " ", -1)
	if s == MotorState(0).String() {
		return ""
	}
	return s
}

// dcMotorStopAction is the stop_actions attribute.
type dcMotorStopAction dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorStopAction) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *dcMotorStopAction) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *dcMotorStopAction) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
func (m *dcMotorStopAction) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorStopAction) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastStopAction
}

// dcMotorTimeSet is the time_sp attribute.
type dcMotorTimeSet dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorTimeSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *dcMotorTimeSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *dcMotorTimeSet) WriteAt(b []byte, off int64) (int, error) {
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
func (m *dcMotorTimeSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorTimeSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strconv.Itoa(int(m._timeSet / time.Millisecond))
}

// dcMotorUevent is the uevent attribute.
type dcMotorUevent dcMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *dcMotorUevent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *dcMotorUevent) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *dcMotorUevent) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := make([]string, 0, len(m._uevent))
	for k, v := range m._uevent {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(e)
	return strings.Join(e, "\n")
}

type dcMotorConn struct {
	id      int
	dcMotor *dcMotor
}

func connectedDCMotors(c ...dcMotorConn) []sisyphus.Node {
	n := make([]sisyphus.Node, len(c))
	for i, m := range c {
		n[i] = d(fmt.Sprintf("motor%d", m.id), 0775).With(
			ro(AddressName, 0444, (*dcMotorAddress)(m.dcMotor)),
			ro(DriverNameName, 0444, (*dcMotorDriver)(m.dcMotor)),
			ro(CommandsName, 0444, (*dcMotorCommands)(m.dcMotor)),
			wo(CommandName, 0222, (*dcMotorCommand)(m.dcMotor)),
			rw(PolarityName, 0666, (*dcMotorPolarity)(m.dcMotor)),
			ro(DutyCycleName, 0444, (*dcMotorDutyCycle)(m.dcMotor)),
			rw(DutyCycleSetpointName, 0666, (*dcMotorDutyCycleSet)(m.dcMotor)),
			rw(RampUpSetpointName, 0666, (*dcMotorRampUpSet)(m.dcMotor)),
			rw(RampDownSetpointName, 0666, (*dcMotorRampDownSet)(m.dcMotor)),
			ro(StateName, 0444, (*dcMotorState)(m.dcMotor)),
			ro(StopActionsName, 0444, (*dcMotorStopActions)(m.dcMotor)),
			rw(StopActionName, 0666, (*dcMotorStopAction)(m.dcMotor)),
			rw(TimeSetpointName, 0666, (*dcMotorTimeSet)(m.dcMotor)),
			ro(UeventName, 0444, (*dcMotorUevent)(m.dcMotor)),
		)
	}
	return n
}

func dcmotorsysfs(m ...dcMotorConn) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("dc-motor", 0775).With(
					connectedDCMotors(m...)...,
				),
			),
		),
	).Sync()
}

func TestDCMotor(t *testing.T) {
	const driver = "rcx-motor"
	conn := []dcMotorConn{
		{
			id: 5,
			dcMotor: &dcMotor{
				address: "outC",
				driver:  driver,

				_commands: []string{
					"run-forever",
					"run-timed",
					"run-direct",
					"stop",
				},

				_lastStopAction: "coast",
				_stopActions: []string{
					"coast",
					"brake",
				},

				_uevent: map[string]string{
					"LEGO_ADDRESS":     "outC",
					"LEGO_DRIVER_NAME": driver,
				},

				t: t,
			},
		},
		{
			id: 7,
			dcMotor: &dcMotor{
				address: "outD",
				driver:  driver,

				t: t,
			},
		},
	}

	fs := dcmotorsysfs(conn...)
	unmount := serve(fs, t)
	defer unmount()

	t.Run("new DCMotor", func(t *testing.T) {
		for _, r := range []struct{ port, driver string }{
			{port: "", driver: conn[0].dcMotor.driver},
			{port: conn[0].dcMotor.address, driver: conn[0].dcMotor.driver},
			{port: conn[0].dcMotor.address, driver: ""},
		} {
			got, err := DCMotorFor(r.port, r.driver)
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
					t.Errorf("unexpected value for have driver error: got:%q want:%q", merr.Have, conn[0].dcMotor.driver)
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
			wantAddr := conn[0].dcMotor.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := conn[0].dcMotor.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Next", func(t *testing.T) {
		m, err := DCMotorFor(conn[0].dcMotor.address, conn[0].dcMotor.driver)
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
		wantAddr := conn[1].dcMotor.address
		if gotAddr != wantAddr {
			t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
		}
		gotDriver, err := DriverFor(got)
		if err != nil {
			t.Errorf("unexpected error getting driver name:%v", err)
		}
		wantDriver := conn[1].dcMotor.driver
		if gotDriver != wantDriver {
			t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
		}
	})

	t.Run("FindAfter", func(t *testing.T) {
		var last *DCMotor
		for _, c := range conn {
			got := new(DCMotor)
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
			wantAddr := c.dcMotor.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := c.dcMotor.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Command", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			commands := m.Commands()
			want := c.dcMotor.commands()
			if !reflect.DeepEqual(commands, want) {
				t.Errorf("unexpected commands value: got:%q want:%q", commands, want)
			}
			for _, command := range commands {
				err := m.Command(command).Err()
				if err != nil {
					t.Errorf("unexpected error for command %q: %v", command, err)
				}

				got := c.dcMotor.lastCommand()
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

				got := c.dcMotor.lastCommand()
				dontwant := command
				if got == dontwant {
					t.Errorf("unexpected invalid command value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Duty cycle", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, sp := range []int{0, 64, 128, 192, 255} {
				c.dcMotor.setDutyCycle(sp)
				got, err := m.DutyCycle()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.dcMotor.dutyCycle()
				if got != want {
					t.Errorf("unexpected duty cycle value: got:%d want:%d", got, want)
				}
			}
		}
	})

	t.Run("Duty cycle setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
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
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
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

	t.Run("Ramp up setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{time.Millisecond, time.Second} {
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
			for _, v := range []time.Duration{-time.Millisecond, -time.Second, -time.Minute, time.Minute} {
				err := m.SetRampUpSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set position setpoint %d", v)
				}
			}
		}
	})

	t.Run("Ramp down setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{time.Millisecond, time.Second} {
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
			for _, v := range []time.Duration{-time.Millisecond, -time.Second, -time.Minute, time.Minute} {
				err := m.SetRampDownSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set position setpoint %d", v)
				}
			}
		}
	})

	t.Run("State", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
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
				c.dcMotor.setState(s)
				got, err := m.State()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.dcMotor.state()
				if got != want {
					t.Errorf("unexpected state value: got:%v want:%v", got, want)
				}
			}
		}
	})

	t.Run("Stop action", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			stopActions := m.StopActions()
			want := c.dcMotor.stopActions()
			if !reflect.DeepEqual(stopActions, want) {
				t.Errorf("unexpected stop actions value: got:%q want:%q", stopActions, want)
			}
			for _, stopAction := range stopActions {
				err := m.SetStopAction(stopAction).Err()
				if err != nil {
					t.Errorf("unexpected error for set stop action %q: %v", stopAction, err)
				}

				got := c.dcMotor.lastStopAction()
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

				got := c.dcMotor.lastStopAction()
				dontwant := stopAction
				if got == dontwant {
					t.Errorf("unexpected invalid stop action value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Time setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
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
					t.Errorf("expected error for set position setpoint %d", v)
				}
			}
		}
	})

	t.Run("Uevent", func(t *testing.T) {
		for _, c := range conn {
			m, err := DCMotorFor(c.dcMotor.address, c.dcMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := m.Uevent()
			if err != nil {
				t.Errorf("unexpected error getting uevent: %v", err)
			}
			want := c.dcMotor.uevent()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
			}
		}
	})
}
