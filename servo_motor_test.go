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

// servoMotor is a servoMotor sysfs directory.
type servoMotor struct {
	address string
	driver  string

	// mu protects the underscore
	// prefix attributes below.
	mu sync.Mutex

	_lastCommand string

	_polarity Polarity

	_maxPulseSet time.Duration
	_midPulseSet time.Duration
	_minPulseSet time.Duration

	_positionSet int

	_rateSet time.Duration

	_state MotorState

	_uevent map[string]string

	t *testing.T
}

func (m *servoMotor) lastCommand() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastCommand
}

func (m *servoMotor) state() MotorState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._state
}

func (m *servoMotor) setState(s MotorState) {
	m.mu.Lock()
	m._state = s
	m.mu.Unlock()
}

func (m *servoMotor) uevent() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._uevent
}

// servoMotorAddress is the address attribute.
type servoMotorAddress servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorAddress) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.address)
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorAddress) Size() (int64, error) {
	return size(m.address), nil
}

// servoMotorDriver is the driver_name attribute.
type servoMotorDriver servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorDriver) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.driver)
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorDriver) Size() (int64, error) {
	return size(m.driver), nil
}

// servoMotorCommands is the commands attribute.
type servoMotorCommands servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorCommands) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorCommands) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorCommands) String() string {
	return "run float"
}

// servoMotorCommand is the command attribute.
type servoMotorCommand servoMotor

// Truncate is a no-op.
func (m *servoMotorCommand) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *servoMotorCommand) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	command := string(chomp(b))
	for _, c := range []string{"run", "float"} {
		if command == c {
			m._lastCommand = command
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorCommand) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorCommand) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m._lastCommand
}

// servoMotorMaxPulseSet is the max_pulse_sp attribute.
type servoMotorMaxPulseSet servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorMaxPulseSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *servoMotorMaxPulseSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *servoMotorMaxPulseSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	d := time.Duration(i) * time.Millisecond
	if err == nil && (d < 2300*time.Millisecond || 2700*time.Millisecond < d) {
		err = fmt.Errorf("ev3dev: invalid max pulse setpoint: %d (valid 2300ms-1700ms)", d)
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._maxPulseSet = d
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorMaxPulseSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorMaxPulseSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprint(int(m._maxPulseSet / time.Millisecond))
}

// servoMotorMidPulseSet is the mid_pulse_sp attribute.
type servoMotorMidPulseSet servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorMidPulseSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *servoMotorMidPulseSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *servoMotorMidPulseSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	d := time.Duration(i) * time.Millisecond
	if err == nil && (d < 1300*time.Millisecond || 1700*time.Millisecond < d) {
		err = fmt.Errorf("ev3dev: invalid mid pulse setpoint: %d (valid 1300ms-1700ms)", d)
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._midPulseSet = d
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorMidPulseSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorMidPulseSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprint(int(m._midPulseSet / time.Millisecond))
}

// servoMotorMinPulseSet is the min_pulse_sp attribute.
type servoMotorMinPulseSet servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorMinPulseSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *servoMotorMinPulseSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *servoMotorMinPulseSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	d := time.Duration(i) * time.Millisecond
	if err == nil && (d < 300*time.Millisecond || 700*time.Millisecond < d) {
		err = fmt.Errorf("ev3dev: invalid min pulse setpoint: %d (valid 300ms-700ms)", d)
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._minPulseSet = d
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorMinPulseSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorMinPulseSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprint(int(m._minPulseSet / time.Millisecond))
}

// servoMotorPolarity is the polarity attribute.
type servoMotorPolarity servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorPolarity) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *servoMotorPolarity) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *servoMotorPolarity) WriteAt(b []byte, off int64) (int, error) {
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
func (m *servoMotorPolarity) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorPolarity) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return string(m._polarity)
}

// servoMotorPositionSet is the position_sp attribute.
type servoMotorPositionSet servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorPositionSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *servoMotorPositionSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *servoMotorPositionSet) WriteAt(b []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if i < -100 || 100 < i {
		err = fmt.Errorf("ev3dev: set position out of range: %d not in -100 - 100", i)
	}
	if err != nil {
		m.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	m._positionSet = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorPositionSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorPositionSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprint(m._positionSet)
}

// servoMotorRateSet is the rate_sp attribute.
type servoMotorRateSet servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorRateSet) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Truncate is a no-op.
func (m *servoMotorRateSet) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (m *servoMotorRateSet) WriteAt(b []byte, off int64) (int, error) {
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
	m._rateSet = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorRateSet) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorRateSet) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprint(int(m._rateSet / time.Millisecond))
}

// servoMotorState is the state attribute.
type servoMotorState servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorState) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorState) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorState) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := strings.Replace(m._state.String(), "|", " ", -1)
	if s == MotorState(0).String() {
		return ""
	}
	return s
}

// servoMotorUevent is the uevent attribute.
type servoMotorUevent servoMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *servoMotorUevent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *servoMotorUevent) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *servoMotorUevent) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := make([]string, 0, len(m._uevent))
	for k, v := range m._uevent {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(e)
	return strings.Join(e, "\n")
}

type servoMotorConn struct {
	id         int
	servoMotor *servoMotor
}

func connectedServoMotors(c ...servoMotorConn) []sisyphus.Node {
	n := make([]sisyphus.Node, len(c))
	for i, m := range c {
		n[i] = d(fmt.Sprintf("motor%d", m.id), 0775).With(
			ro(AddressName, 0444, (*servoMotorAddress)(m.servoMotor)),
			ro(DriverNameName, 0444, (*servoMotorDriver)(m.servoMotor)),
			ro(CommandsName, 0444, (*servoMotorCommands)(m.servoMotor)),
			wo(CommandName, 0222, (*servoMotorCommand)(m.servoMotor)),
			rw(MaxPulseSetpointName, 0666, (*servoMotorMaxPulseSet)(m.servoMotor)),
			rw(MidPulseSetpointName, 0666, (*servoMotorMidPulseSet)(m.servoMotor)),
			rw(MinPulseSetpointName, 0666, (*servoMotorMinPulseSet)(m.servoMotor)),
			rw(PolarityName, 0666, (*servoMotorPolarity)(m.servoMotor)),
			rw(PositionSetpointName, 0666, (*servoMotorPositionSet)(m.servoMotor)),
			rw(RateSetpointName, 0666, (*servoMotorRateSet)(m.servoMotor)),
			ro(StateName, 0444, (*servoMotorState)(m.servoMotor)),
			ro(UeventName, 0444, (*servoMotorUevent)(m.servoMotor)),
		)
	}
	return n
}

func servomotorsysfs(m ...servoMotorConn) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("servo-motor", 0775).With(
					connectedServoMotors(m...)...,
				),
			),
		),
	).Sync()
}

func TestServoMotor(t *testing.T) {
	const driver = "lego-nxt-motor"
	conn := []servoMotorConn{
		{
			id: 5,
			servoMotor: &servoMotor{
				address: "outD",
				driver:  driver,

				_uevent: map[string]string{
					"LEGO_ADDRESS":     "outD",
					"LEGO_DRIVER_NAME": driver,
				},

				t: t,
			},
		},
		{
			id: 7,
			servoMotor: &servoMotor{
				address: "outA",
				driver:  driver,

				t: t,
			},
		},
	}

	fs := servomotorsysfs(conn...)
	unmount := serve(fs, t)
	defer unmount()

	t.Run("new ServoMotor", func(t *testing.T) {
		for _, r := range []struct{ port, driver string }{
			{port: "", driver: conn[0].servoMotor.driver},
			{port: conn[0].servoMotor.address, driver: conn[0].servoMotor.driver},
			{port: conn[0].servoMotor.address, driver: ""},
		} {
			got, err := ServoMotorFor(r.port, r.driver)
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
					t.Errorf("unexpected value for have driver error: got:%q want:%q", merr.Have, conn[0].servoMotor.driver)
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
			wantAddr := conn[0].servoMotor.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := conn[0].servoMotor.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Next", func(t *testing.T) {
		m, err := ServoMotorFor(conn[0].servoMotor.address, conn[0].servoMotor.driver)
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
		wantAddr := conn[1].servoMotor.address
		if gotAddr != wantAddr {
			t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
		}
		gotDriver, err := DriverFor(got)
		if err != nil {
			t.Errorf("unexpected error getting driver name:%v", err)
		}
		wantDriver := conn[1].servoMotor.driver
		if gotDriver != wantDriver {
			t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
		}
	})

	t.Run("FindAfter", func(t *testing.T) {
		var last *ServoMotor
		for _, c := range conn {
			got := new(ServoMotor)
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
			wantAddr := c.servoMotor.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := c.servoMotor.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Command", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, command := range m.Commands() {
				err := m.Command(command).Err()
				if err != nil {
					t.Errorf("unexpected error for command %q: %v", command, err)
				}

				got := c.servoMotor.lastCommand()
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

				got := c.servoMotor.lastCommand()
				dontwant := command
				if got == dontwant {
					t.Errorf("unexpected invalid command value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Max pulse setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{
				2300 * time.Millisecond,
				2400 * time.Millisecond,
				2500 * time.Millisecond,
				2600 * time.Millisecond,
				2700 * time.Millisecond,
			} {
				err := m.SetMaxPulseSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for set max pulse setpoint %d: %v", v, err)
				}

				got, err := m.MaxPulseSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected max pulse setpoint value: got:%d want:%d", got, want)
				}
			}
			for _, v := range []time.Duration{
				2299 * time.Millisecond,
				2701 * time.Millisecond,
			} {
				err := m.SetMaxPulseSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set max pulse setpoint %d", v)
				}
			}
		}
	})

	t.Run("Mid pulse setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{
				1300 * time.Millisecond,
				1400 * time.Millisecond,
				1500 * time.Millisecond,
				1600 * time.Millisecond,
				1700 * time.Millisecond,
			} {
				err := m.SetMidPulseSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for set rate setpoint %d: %v", v, err)
				}

				got, err := m.MidPulseSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected rate setpoint value: got:%d want:%d", got, want)
				}
			}
			for _, v := range []time.Duration{
				1299 * time.Millisecond,
				1701 * time.Millisecond,
			} {
				err := m.SetMidPulseSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set position setpoint %d", v)
				}
			}
		}
	})

	t.Run("Min pulse setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{
				300 * time.Millisecond,
				400 * time.Millisecond,
				500 * time.Millisecond,
				600 * time.Millisecond,
				700 * time.Millisecond,
			} {
				err := m.SetMinPulseSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for set rate setpoint %d: %v", v, err)
				}

				got, err := m.MinPulseSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected rate setpoint value: got:%d want:%d", got, want)
				}
			}
			for _, v := range []time.Duration{
				299 * time.Millisecond,
				701 * time.Millisecond,
			} {
				err := m.SetMinPulseSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set position setpoint %d", v)
				}
			}
		}
	})

	t.Run("Polarity", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
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

	t.Run("Position setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
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
			for _, v := range []int64{-101, 101} {
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

	t.Run("Rate setpoint", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, v := range []time.Duration{0, time.Millisecond, time.Second, time.Minute} {
				err := m.SetRateSetpoint(v).Err()
				if err != nil {
					t.Errorf("unexpected error for set rate setpoint %v: %v", v, err)
				}

				got, err := m.RateSetpoint()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := v
				if got != want {
					t.Errorf("unexpected rate setpoint value: got:%v want:%v", got, want)
				}
			}
			for _, v := range []time.Duration{-time.Millisecond, -time.Second, -time.Minute} {
				err := m.SetRateSetpoint(v).Err()
				if err == nil {
					t.Errorf("expected error for set position setpoint %v", v)
				}
			}
		}
	})

	t.Run("State", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
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
				c.servoMotor.setState(s)
				got, err := m.State()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				want := c.servoMotor.state()
				if got != want {
					t.Errorf("unexpected state value: got:%v want:%v", got, want)
				}
			}
		}
	})

	t.Run("Uevent", func(t *testing.T) {
		for _, c := range conn {
			m, err := ServoMotorFor(c.servoMotor.address, c.servoMotor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := m.Uevent()
			if err != nil {
				t.Errorf("unexpected error getting uevent: %v", err)
			}
			want := c.servoMotor.uevent()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
			}
		}
	})
}
