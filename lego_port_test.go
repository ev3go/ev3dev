// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"

	. "github.com/ev3go/ev3dev"

	"github.com/ev3go/sisyphus"
)

// legoPort is a legoPort sysfs directory.
type legoPort struct {
	address string
	driver  string

	// mu protects the underscore
	// prefix attributes below.
	mu sync.Mutex

	_mode  string
	_modes []string

	_device string

	_status string

	_uevent map[string]string

	t *testing.T
}

func (p *legoPort) modes() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p._modes
}

func (p *legoPort) device() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p._device
}

func (p *legoPort) status() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p._status
}

func (p *legoPort) setStatus(s string) {
	p.mu.Lock()
	p._status = s
	p.mu.Unlock()
}

func (p *legoPort) uevent() map[string]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p._uevent
}

// legoPortAddress is the address attribute.
type legoPortAddress legoPort

// ReadAt satisfies the io.ReaderAt interface.
func (p *legoPortAddress) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *legoPortAddress) Size() (int64, error) {
	return size(p), nil
}

// String returns a string representation of the attribute.
func (p *legoPortAddress) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.address
}

// legoPortDriver is the driver_name attribute.
type legoPortDriver legoPort

// ReadAt satisfies the io.ReaderAt interface.
func (p *legoPortDriver) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *legoPortDriver) Size() (int64, error) {
	return size(p), nil
}

// String returns a string representation of the attribute.
func (p *legoPortDriver) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.driver
}

// legoPortModes is the modes attribute.
type legoPortModes legoPort

// ReadAt satisfies the io.ReaderAt interface.
func (p *legoPortModes) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *legoPortModes) Size() (int64, error) {
	return size(p), nil
}

// String returns a string representation of the attribute.
func (p *legoPortModes) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	sort.Strings(p._modes)
	return strings.Join(p._modes, " ")
}

// legoPortMode is the mode attribute.
type legoPortMode legoPort

// ReadAt satisfies the io.ReaderAt interface.
func (p *legoPortMode) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Truncate is a no-op.
func (p *legoPortMode) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (p *legoPortMode) WriteAt(b []byte, off int64) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	mode := string(chomp(b))
	for _, c := range p._modes {
		if mode == c {
			p._mode = mode
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (p *legoPortMode) Size() (int64, error) {
	return size(p), nil
}

// String returns a string representation of the attribute.
func (p *legoPortMode) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p._mode
}

// legoPortStatus is the status attribute.
type legoPortStatus legoPort

// ReadAt satisfies the io.ReaderAt interface.
func (p *legoPortStatus) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *legoPortStatus) Size() (int64, error) {
	return size(p), nil
}

// String returns a string representation of the attribute.
func (p *legoPortStatus) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p._status
}

// legoPortSetDevice is the set_device attribute.
type legoPortSetDevice legoPort

// Truncate is a no-op.
func (p *legoPortSetDevice) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (p *legoPortSetDevice) WriteAt(b []byte, off int64) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p._device = string(chomp(b))
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (p *legoPortSetDevice) Size() (int64, error) {
	return size(p), nil
}

// String returns a string representation of the attribute.
func (p *legoPortSetDevice) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p._device
}

// legoPortUevent is the uevent attribute.
type legoPortUevent legoPort

// ReadAt satisfies the io.ReaderAt interface.
func (p *legoPortUevent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *legoPortUevent) Size() (int64, error) {
	return size(p), nil
}

// String returns a string representation of the attribute.
func (p *legoPortUevent) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	e := make([]string, 0, len(p._uevent))
	for k, v := range p._uevent {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(e)
	return strings.Join(e, "\n")
}

type legoPortConn struct {
	id       int
	legoPort *legoPort
}

func connectedLegoPorts(c ...legoPortConn) []sisyphus.Node {
	n := make([]sisyphus.Node, len(c))
	for i, p := range c {
		n[i] = d(fmt.Sprintf("port%d", p.id), 0775).With(
			ro(AddressName, 0444, (*legoPortAddress)(p.legoPort)),
			ro(DriverNameName, 0444, (*legoPortDriver)(p.legoPort)),
			ro(ModesName, 0444, (*legoPortModes)(p.legoPort)),
			rw(ModeName, 0666, (*legoPortMode)(p.legoPort)),
			wo(SetDeviceName, 0222, (*legoPortSetDevice)(p.legoPort)),
			ro(StatusName, 0444, (*legoPortStatus)(p.legoPort)),
			ro(UeventName, 0444, (*legoPortUevent)(p.legoPort)),
		)
	}
	return n
}

func legoportsysfs(p ...legoPortConn) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("lego-port", 0775).With(
					connectedLegoPorts(p...)...,
				),
			),
		),
	).Sync()
}

func TestLegoPort(t *testing.T) {
	const driver = "legoev3-input-port"
	conn := []legoPortConn{
		{
			id: 5,
			legoPort: &legoPort{
				address: "in1",
				driver:  driver,

				_modes: []string{"GYRO-ANG", "GYRO-RATE", "GYRO-FAS", "GYRO-G&A", "GYRO-CAL"},
				_mode:  "GYRO-ANG",

				_uevent: map[string]string{
					"DEVTYPE":          "legoev3-input-port",
					"LEGO_DRIVER_NAME": driver,
					"LEGO_ADDRESS":     "in1",
				},

				t: t,
			},
		},
		{
			id: 7,
			legoPort: &legoPort{
				address: "in4",
				driver:  driver,

				t: t,
			},
		},
	}

	fs := legoportsysfs(conn...)
	unmount := serve(fs, t)
	defer unmount()

	t.Run("new LegoPort", func(t *testing.T) {
		for _, r := range []struct{ port, driver string }{
			{port: "", driver: conn[0].legoPort.driver},
			{port: conn[0].legoPort.address, driver: conn[0].legoPort.driver},
			{port: conn[0].legoPort.address, driver: ""},
		} {
			got, err := LegoPortFor(r.port, r.driver)
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
					t.Errorf("unexpected value for have driver error: got:%q want:%q", merr.Have, conn[0].legoPort.driver)
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
			wantAddr := conn[0].legoPort.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := conn[0].legoPort.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
			methodDriver := got.Driver()
			if methodDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", methodDriver, wantDriver)
			}
		}
	})

	t.Run("Next", func(t *testing.T) {
		p, err := LegoPortFor(conn[0].legoPort.address, conn[0].legoPort.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := p.Next()
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
		wantAddr := conn[1].legoPort.address
		if gotAddr != wantAddr {
			t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
		}
		gotDriver, err := DriverFor(got)
		if err != nil {
			t.Errorf("unexpected error getting driver name:%v", err)
		}
		wantDriver := conn[1].legoPort.driver
		if gotDriver != wantDriver {
			t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
		}
	})

	t.Run("FindAfter", func(t *testing.T) {
		var last *LegoPort
		for _, c := range conn {
			got := new(LegoPort)
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
			wantAddr := c.legoPort.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := c.legoPort.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Mode", func(t *testing.T) {
		p, err := LegoPortFor(conn[0].legoPort.address, conn[0].legoPort.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		modes := p.Modes()
		want := conn[0].legoPort.modes()
		if !reflect.DeepEqual(modes, want) {
			t.Errorf("unexpected modes value: got:%q want:%q", modes, want)
		}
		for _, mode := range modes {
			err := p.SetMode(mode).Err()
			if err != nil {
				t.Errorf("unexpected error for mode %q: %v", mode, err)
			}

			got, err := p.Mode()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := mode
			if got != want {
				t.Errorf("unexpected mode value: got:%q want:%q", got, want)
			}
		}
		for _, mode := range []string{"invalid", "another"} {
			err := p.SetMode(mode).Err()
			if err == nil {
				t.Errorf("expected error for mode %q", mode)
			}

			got, err := p.Mode()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			dontwant := mode
			if got == dontwant {
				t.Errorf("unexpected invalid mode value: got:%q, don't want:%q", got, dontwant)
			}
		}
	})

	t.Run("SetDevice", func(t *testing.T) {
		p, err := LegoPortFor(conn[0].legoPort.address, conn[0].legoPort.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, device := range []string{"sensor", "motor", "driver"} {
			err := p.SetDevice(device).Err()
			if err != nil {
				t.Errorf("unexpected error for device %q: %v", device, err)
			}

			got := conn[0].legoPort.device()
			want := device
			if got != want {
				t.Errorf("unexpected device value: got:%q want:%q", got, want)
			}
		}
	})

	t.Run("Status", func(t *testing.T) {
		p, err := LegoPortFor(conn[0].legoPort.address, conn[0].legoPort.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, status := range []string{"device", "error", "no-device"} {
			conn[0].legoPort.setStatus(status)
			got, err := p.Status()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := conn[0].legoPort.status()
			if got != want {
				t.Errorf("unexpected status value: got:%q want:%q", got, want)
			}
		}
	})

	t.Run("Uevent", func(t *testing.T) {
		for _, c := range conn {
			p, err := LegoPortFor(c.legoPort.address, c.legoPort.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := p.Uevent()
			if err != nil {
				t.Errorf("unexpected error getting uevent: %v", err)
			}
			want := c.legoPort.uevent()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
			}
		}
	})

	t.Run("ConnectedTo", func(t *testing.T) {
		for i, c := range conn {
			p, err := LegoPortFor(c.legoPort.address, c.legoPort.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			name := []string{"in2:sensor", "outC:motor"}[i%2]

			path := filepath.Join(p.Path(), p.String())
			err = fs.Bind(path, d(name, 0775))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := ConnectedTo(p)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := name
			if got != want {
				t.Errorf("unexpected ConnectedTo value: got:%q want:%q", got, want)
			}
		}
	})
}
