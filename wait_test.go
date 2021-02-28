// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ev3go/sisyphus"

	. "github.com/ev3go/ev3dev"
)

var stateIsOKTests = []struct {
	state MotorState

	mask MotorState
	want MotorState
	not  MotorState
	any  bool

	wantOK bool
}{
	{
		state:  0,
		wantOK: true,
	},
	{
		state:  Running,
		mask:   ^Running,
		wantOK: true,
	},
	{
		state:  Running,
		mask:   Running,
		want:   0,
		wantOK: false,
	},
	{
		state:  Running,
		mask:   Running,
		want:   Running,
		wantOK: true,
	},
	{
		state:  Running,
		mask:   Running | Stalled,
		want:   Running,
		not:    Stalled,
		wantOK: true,
	},
	{
		state:  Running | Stalled,
		mask:   Running | Stalled,
		want:   Running,
		not:    Stalled,
		wantOK: false,
	},
	{
		state:  Ramping,
		mask:   Running | Stalled | Overloaded | Ramping,
		any:    true,
		wantOK: true,
	},
	{
		state:  Ramping,
		mask:   Running | Stalled | Overloaded | Ramping,
		any:    true,
		not:    Ramping,
		wantOK: false,
	},
	{
		state:  Running | Stalled,
		mask:   Running | Stalled | Overloaded | Ramping,
		any:    true,
		not:    Ramping,
		wantOK: true,
	},
	{
		state:  Running | Stalled | Overloaded,
		mask:   Running | Stalled | Overloaded,
		any:    true,
		not:    Stalled | Overloaded,
		wantOK: false,
	},
}

func TestStateIsOK(t *testing.T) {
	for _, test := range stateIsOKTests {
		gotOK := StateIsOK(test.state, test.mask, test.want, test.not, test.any)
		if gotOK != test.wantOK {
			t.Errorf("unexpected result for state=%s mask=%s want=%s not=%s any=%t: got:%t want:%t",
				test.state, test.mask, test.want, test.not, test.any, gotOK, test.wantOK)
		}
	}
}

// waitMotor is a tacho motor sysfs directory.
type waitMotor struct {
	address string
	driver  string

	mu    sync.Mutex
	state MotorState

	t *testing.T
}

func (m *waitMotor) setState(s MotorState) {
	m.mu.Lock()
	m.state = s
	m.mu.Unlock()
}

// waitMotorAddress is the address attribute.
type waitMotorAddress waitMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *waitMotorAddress) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.address)
}

// Size returns the length of the backing data and a nil error.
func (m *waitMotorAddress) Size() (int64, error) {
	return size(m.address), nil
}

// waitMotorDriver is the driver_name attribute.
type waitMotorDriver waitMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *waitMotorDriver) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m.driver)
}

// Size returns the length of the backing data and a nil error.
func (m *waitMotorDriver) Size() (int64, error) {
	return size(m.driver), nil
}

// waitMotorCommands is the commands attribute.
type waitMotorCommands waitMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *waitMotorCommands) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, "none")
}

// Size returns the length of the backing data and a nil error.
func (m *waitMotorCommands) Size() (int64, error) {
	return size("none"), nil
}

// waitMotorStopActions is the stop_actions attribute.
type waitMotorStopActions waitMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *waitMotorStopActions) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, "none")
}

// Size returns the length of the backing data and a nil error.
func (m *waitMotorStopActions) Size() (int64, error) {
	return size("none"), nil
}

// waitMotorMaxSpeed is the max_speed attribute.
type waitMotorMaxSpeed waitMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *waitMotorMaxSpeed) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, "1200")
}

// Size returns the length of the backing data and a nil error.
func (m *waitMotorMaxSpeed) Size() (int64, error) {
	return size("1200"), nil
}

// waitMotorCountsPerRot is the count_per_rot attribute.
type waitMotorCountPerRot waitMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *waitMotorCountPerRot) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, "360")
}

// Size returns the length of the backing data and a nil error.
func (m *waitMotorCountPerRot) Size() (int64, error) {
	return size("360"), nil
}

// waitMotorState is the state attribute.
type waitMotorState waitMotor

// ReadAt satisfies the io.ReaderAt interface.
func (m *waitMotorState) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, m)
}

// Size returns the length of the backing data and a nil error.
func (m *waitMotorState) Size() (int64, error) {
	return size(m), nil
}

// String returns a string representation of the attribute.
func (m *waitMotorState) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := strings.Replace(m.state.String(), "|", " ", -1)
	if s == MotorState(0).String() {
		return ""
	}
	return s
}

type waitMotorConn struct {
	id        int
	waitMotor *waitMotor
}

func connectedWaitMotors(c ...waitMotorConn) []sisyphus.Node {
	n := make([]sisyphus.Node, len(c))
	for i, m := range c {
		n[i] = d(fmt.Sprintf("motor%d", m.id), 0775).With(
			ro(AddressName, 0444, (*waitMotorAddress)(m.waitMotor)),
			ro(DriverNameName, 0444, (*waitMotorDriver)(m.waitMotor)),
			ro(CommandsName, 0444, (*waitMotorCommands)(m.waitMotor)),
			ro(CountPerRotName, 0444, (*waitMotorCountPerRot)(m.waitMotor)),
			ro(MaxSpeedName, 0444, (*waitMotorMaxSpeed)(m.waitMotor)),
			ro(StopActionsName, 0444, (*waitMotorStopActions)(m.waitMotor)),
			ro(StateName, 0444, (*waitMotorState)(m.waitMotor)),
		)
	}
	return n
}

func waitmotorsysfs(m ...waitMotorConn) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("tacho-motor", 0775).With(
					connectedWaitMotors(m...)...,
				),
			),
		),
	).Sync()
}

type period struct {
	s MotorState
	d time.Duration
}

func sequenceDuration(s []period) time.Duration {
	var d time.Duration
	for _, p := range s {
		d += p.d
	}
	return d
}

var waitTests = []struct {
	states []period

	mask    MotorState
	query   MotorState
	not     MotorState
	any     bool
	timeout time.Duration

	want   MotorState
	wantOK bool
}{
	{
		states: []period{
			{s: 0, d: 200 * time.Millisecond},
			{s: Running, d: 200 * time.Millisecond},
		},

		mask:    Running,
		query:   Running,
		timeout: 100 * time.Millisecond,

		want:   0,
		wantOK: false,
	},
	{
		states: []period{
			{s: 0, d: 200 * time.Millisecond},
			{s: Running, d: 200 * time.Millisecond},
		},

		mask:    Running,
		query:   Running,
		timeout: time.Second,

		want:   Running,
		wantOK: true,
	},
	{
		states: []period{
			{s: 0, d: 200 * time.Millisecond},
			{s: Running | Stalled, d: 200 * time.Millisecond},
			{s: Running, d: 200 * time.Millisecond},
		},

		mask:    Running | Stalled,
		query:   Running,
		not:     Stalled,
		timeout: time.Second,

		want:   Running,
		wantOK: true,
	},
	{
		states: []period{
			{s: Running | Stalled | Overloaded, d: 200 * time.Millisecond},
			{s: Holding, d: 200 * time.Millisecond},
		},

		mask:    Running | Stalled | Overloaded,
		any:     true,
		not:     Stalled | Overloaded,
		timeout: time.Second,

		want:   Holding,
		wantOK: true,
	},
	{
		states: []period{
			{s: 0, d: 200 * time.Millisecond},
			{s: Running | Stalled, d: 200 * time.Millisecond},
			{s: Running, d: 200 * time.Millisecond},
		},

		mask:    Running | Stalled,
		any:     true,
		timeout: time.Second,

		want:   Running | Stalled,
		wantOK: true,
	},
	{
		states: []period{
			{s: 0, d: 200 * time.Millisecond},
			{s: Running, d: 200 * time.Millisecond},
		},

		mask:    Running,
		query:   Running,
		timeout: -1,

		want:   Running,
		wantOK: true,
	},
}

func TestWaitMotor(t *testing.T) {
	const driver = "lego-ev3-l-motor"
	conn := waitMotorConn{
		id: 5,
		waitMotor: &waitMotor{
			address: "outA",
			driver:  driver,

			t: t,
		},
	}

	fs := waitmotorsysfs(conn)
	unmount := serve(fs, t)
	defer unmount()

	t.Run("State", func(t *testing.T) {
		m, err := TachoMotorFor(conn.waitMotor.address, conn.waitMotor.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, test := range waitTests {
			timeout := 10 * sequenceDuration(test.states)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, p := range test.states {
					conn.waitMotor.setState(p.s)
					err = fs.InvalidatePath(filepath.Join(m.Path(), m.String(), StateName))
					if err != nil {
						t.Fatalf("unexpected error invalidating state: %v", err)
					}
					time.Sleep(p.d)
				}
			}()

			ok := make(chan bool, 1)
			go func() {
				select {
				case <-time.After(timeout):
					t.Fatalf("failed to timeout after %v", timeout)
				case <-ok:
				}
			}()

			got, gotOK, err := Wait(m, test.mask, test.query, test.not, test.any, test.timeout)
			ok <- true
			if err != nil {
				// FIXME(kortschak): Occasionally FUSE appears to be in a state
				// where the reported size of the attribute is 1 ("\n") but the value
				// of the attribute is "running\n". The first cached byte is read
				// and the rest of the attribute bytes are appended giving "\nunning".
				// This causes an error state to be returned by Wait.
				// I suspect that this would be fixed with FUSE direct_io.
				goto next
			}
			if got != test.want {
				t.Errorf("unexpected motor state: got:%v want:%v", got, test.want)
			}
			if gotOK != test.wantOK {
				t.Errorf("unexpected success: got:%t want:%t", gotOK, test.wantOK)
			}

		next:
			wg.Wait()
		}
	})
}
