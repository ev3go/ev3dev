// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

// sensor is a sensor sysfs directory.
type sensor struct {
	address string
	driver  string

	// mu protects the underscore
	// prefix attributes below.
	mu sync.Mutex

	_lastCommand string
	_commands    []string

	_direct []byte

	_mode     string
	_modes    []string
	_units    map[string]string
	_decimals map[string]int

	_binData       []byte
	_binDataFormat string

	_values []string

	_pollRate int

	_uevent map[string]string

	t *testing.T
}

func (s *sensor) commands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._commands
}

func (s *sensor) lastCommand() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._lastCommand
}

func (s *sensor) direct() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._direct
}

func (s *sensor) modes() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._modes
}

func (s *sensor) units() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._units[s._mode]
}

func (s *sensor) decimals() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._decimals[s._mode]
}

func (s *sensor) binData() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._binData
}

func (s *sensor) binDataFormat() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._binDataFormat
}

func (s *sensor) values() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._values
}

func (s *sensor) uevent() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._uevent
}

// sensorAddress is the address attribute.
type sensorAddress sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorAddress) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s.address)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorAddress) Size() (int64, error) {
	return size(s.address), nil
}

// sensorDriver is the driver_name attribute.
type sensorDriver sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorDriver) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s.driver)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorDriver) Size() (int64, error) {
	return size(s.driver), nil
}

// sensorCommands is the commands attribute.
type sensorCommands sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorCommands) ReadAt(b []byte, offset int64) (int, error) {
	if len(s._commands) == 0 {
		return len(b), syscall.ENOTSUP
	}
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorCommands) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorCommands) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	sort.Strings(s._commands)
	return strings.Join(s._commands, " ")
}

// sensorCommand is the command attribute.
type sensorCommand sensor

// Truncate is a no-op.
func (s *sensorCommand) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (s *sensorCommand) WriteAt(b []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s._commands) == 0 {
		return len(b), syscall.ENOTSUP
	}
	command := string(chomp(b))
	for _, c := range s._commands {
		if command == c {
			s._lastCommand = command
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (s *sensorCommand) Size() (int64, error) {
	return size(s._lastCommand), nil
}

// sensorModes is the modes attribute.
type sensorModes sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorModes) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorModes) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorModes) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	sort.Strings(s._modes)
	return strings.Join(s._modes, " ")
}

// sensorMode is the mode attribute.
type sensorMode sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorMode) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Truncate is a no-op.
func (s *sensorMode) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (s *sensorMode) WriteAt(b []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mode := string(chomp(b))
	for _, c := range s._modes {
		if mode == c {
			s._mode = mode
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (s *sensorMode) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorMode) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._mode
}

// sensorBinDataFormat is the bin_data_format attribute.
type sensorBinDataFormat sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorBinDataFormat) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorBinDataFormat) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorBinDataFormat) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._binDataFormat
}

// sensorBinData is the bin_datas attribute.
type sensorBinData sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorBinData) ReadAt(b []byte, offset int64) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	d := s.String()
	if offset >= int64(len(d)) {
		return 0, io.EOF
	}
	n := copy(b, d[offset:])
	if n <= len(b) {
		return n, io.EOF
	}
	return n, nil
}

// Size returns the length of the backing data and a nil error.
func (s *sensorBinData) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorBinData) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return string(s._binData)
}

// sensorDirect is the direct attribute.
type sensorDirect sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorDirect) ReadAt(b []byte, offset int64) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	d := s.String()
	if offset >= int64(len(d)) {
		return 0, io.EOF
	}
	n := copy(b, d[offset:])
	if n <= len(b) {
		return n, io.EOF
	}
	return n, nil
}

// Truncate truncates the Bytes at n bytes from the beginning of the slice.
func (s *sensorDirect) Truncate(n int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n < 0 || n > int64(len(s._direct)) {
		return syscall.EINVAL
	}
	tail := s._direct[n:cap(s._direct)]
	for i := range tail {
		tail[i] = 0
	}
	s._direct = s._direct[:n]
	return nil
}

// WriteAt satisfies the io.WriterAt interface.
func (s *sensorDirect) WriteAt(b []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if off >= int64(cap(s._direct)) {
		t := make([]byte, off+int64(len(b)))
		copy(t, s._direct)
		s._direct = t
	}
	s._direct = s._direct[:off]
	s._direct = append(s._direct, b...)
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (s *sensorDirect) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorDirect) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return string(s._direct)
}

// sensorUnits is the units attribute.
type sensorUnits sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorUnits) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorUnits) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorUnits) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._units[s._mode]
}

// sensorPollRate is the poll_rate attribute.
type sensorPollRate sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorPollRate) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Truncate is a no-op.
func (s *sensorPollRate) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (s *sensorPollRate) WriteAt(b []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		s.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	s._pollRate = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (s *sensorPollRate) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorPollRate) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprint(s._pollRate)
}

// sensorDecimals is the decimals attribute.
type sensorDecimals sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorDecimals) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorDecimals) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorDecimals) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprint(s._decimals[s._mode])
}

// sensorNumValues is the num_values attribute.
type sensorNumValues sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorNumValues) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorNumValues) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorNumValues) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprint(len(s._values))
}

// sensorValue is the valueN attribute.
type sensorValue struct {
	n int
	*sensor
}

// ReadAt satisfies the io.ReaderAt interface.
func (s sensorValue) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s sensorValue) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s sensorValue) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s._values) <= s.n {
		return "\n"
	}
	return s._values[s.n]
}

// sensorTextValues is the text_values attribute.
type sensorTextValues sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorTextValues) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorTextValues) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorTextValues) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.Join(s._values, " ")
}

// sensorUevent is the uevent attribute.
type sensorUevent sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorUevent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorUevent) Size() (int64, error) {
	return size(s), nil
}

// String returns a string representation of the attribute.
func (s *sensorUevent) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := make([]string, 0, len(s._uevent))
	for k, v := range s._uevent {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(e)
	return strings.Join(e, "\n")
}

type sensorConn struct {
	id     int
	sensor *sensor
}

func connectedSensors(c ...sensorConn) []sisyphus.Node {
	n := make([]sisyphus.Node, len(c))
	for i, s := range c {
		n[i] = d(fmt.Sprintf("sensor%d", s.id), 0775).With(
			ro(AddressName, 0444, (*sensorAddress)(s.sensor)),
			ro(DriverNameName, 0444, (*sensorDriver)(s.sensor)),
			ro(ModesName, 0444, (*sensorModes)(s.sensor)),
			rw(ModeName, 0666, (*sensorMode)(s.sensor)),
			ro(CommandsName, 0444, (*sensorCommands)(s.sensor)),
			wo(CommandName, 0222, (*sensorCommand)(s.sensor)),
			ro(BinDataFormatName, 0444, (*sensorBinDataFormat)(s.sensor)),
			ro(BinDataName, 0444, (*sensorBinData)(s.sensor)),
			rw(DirectName, 0666, (*sensorDirect)(s.sensor)),
			rw(PollRateName, 0666, (*sensorPollRate)(s.sensor)),
			ro(UnitsName, 0444, (*sensorUnits)(s.sensor)),
			ro(DecimalsName, 0444, (*sensorDecimals)(s.sensor)),
			ro(NumValuesName, 0444, (*sensorNumValues)(s.sensor)),
			ro(ValueName+"0", 0444, sensorValue{0, s.sensor}),
			ro(ValueName+"1", 0444, sensorValue{1, s.sensor}),
			ro(TextValuesName, 0444, (*sensorTextValues)(s.sensor)),
			ro(UeventName, 0444, (*sensorUevent)(s.sensor)),
		)
	}
	return n
}

func sensorsysfs(s ...sensorConn) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("lego-sensor", 0775).With(
					connectedSensors(s...)...,
				),
			),
		),
	).Sync()
}

func TestSensor(t *testing.T) {
	const driver = "lego-ev3-gyro"
	conn := []sensorConn{
		{
			id: 5,
			sensor: &sensor{
				address: "in2",
				driver:  driver,

				_modes:    []string{"GYRO-ANG", "GYRO-RATE", "GYRO-FAS", "GYRO-G&A", "GYRO-CAL"},
				_mode:     "GYRO-ANG",
				_units:    map[string]string{"GYRO-ANG": "deg", "GYRO-RATE": "d/s"},
				_decimals: map[string]int{"GYRO-ANG": 0, "GYRO-RATE": 1},

				_commands: []string{"operate", "twirl"},

				_binDataFormat: "s16",
				_binData:       []byte{0x01, 0x00, 0x02, 0x00},

				_values: []string{"1", "2"},

				_direct: []byte("initial state"),

				_uevent: map[string]string{
					"LEGO_ADDRESS":     "in2",
					"LEGO_DRIVER_NAME": driver,
				},

				t: t,
			},
		},
		{
			id: 7,
			sensor: &sensor{
				address: "in4",
				driver:  driver,

				_modes: []string{"GYRO-ANG", "GYRO-RATE", "GYRO-FAS", "GYRO-G&A", "GYRO-CAL"},
				_mode:  "GYRO-ANG",
				_units: map[string]string{"GYRO-ANG": "deg", "GYRO-RATE": "d/s"},

				_binDataFormat: "s16",
				_binData:       []byte{0x01, 0x00, 0x02, '\n'},

				t: t,
			},
		},
	}

	fs := sensorsysfs(conn...)
	unmount := serve(fs, t)
	defer unmount()

	t.Run("new Sensor", func(t *testing.T) {
		for _, r := range []struct{ port, driver string }{
			{port: "", driver: conn[0].sensor.driver},
			{port: conn[0].sensor.address, driver: conn[0].sensor.driver},
			{port: conn[0].sensor.address, driver: ""},
		} {
			got, err := SensorFor(r.port, r.driver)
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
					t.Errorf("unexpected value for have driver error: got:%q want:%q", merr.Have, conn[0].sensor.driver)
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
			wantAddr := conn[0].sensor.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := conn[0].sensor.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Next", func(t *testing.T) {
		s, err := SensorFor(conn[0].sensor.address, conn[0].sensor.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := s.Next()
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
		wantAddr := conn[1].sensor.address
		if gotAddr != wantAddr {
			t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
		}
		gotDriver, err := DriverFor(got)
		if err != nil {
			t.Errorf("unexpected error getting driver name:%v", err)
		}
		wantDriver := conn[1].sensor.driver
		if gotDriver != wantDriver {
			t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
		}
	})

	t.Run("FindAfter", func(t *testing.T) {
		var last *Sensor
		for _, c := range conn {
			got := new(Sensor)
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
			wantAddr := c.sensor.address
			if gotAddr != wantAddr {
				t.Errorf("unexpected value for address: got:%q want:%q", gotAddr, wantAddr)
			}
			gotDriver, err := DriverFor(got)
			if err != nil {
				t.Errorf("unexpected error getting driver name:%v", err)
			}
			wantDriver := c.sensor.driver
			if gotDriver != wantDriver {
				t.Errorf("unexpected value for driver name: got:%q want:%q", gotDriver, wantDriver)
			}
		}
	})

	t.Run("Mode", func(t *testing.T) {
		s, err := SensorFor(conn[0].sensor.address, conn[0].sensor.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		modes, err := s.Modes()
		if err != nil {
			t.Fatalf("unexpected error getting modes: %v", err)
		}
		want := conn[0].sensor.modes()
		if !reflect.DeepEqual(modes, want) {
			t.Errorf("unexpected modes value: got:%q want:%q", modes, want)
		}
		for _, mode := range modes {
			err := s.SetMode(mode).Err()
			if err != nil {
				t.Errorf("unexpected error for mode %q: %v", mode, err)
			}

			got, err := s.Mode()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := mode
			if got != want {
				t.Errorf("unexpected mode value: got:%q want:%q", got, want)
			}
		}
		for _, mode := range []string{"invalid", "another"} {
			err := s.SetMode(mode).Err()
			if err == nil {
				t.Errorf("expected error for mode %q", mode)
			}

			got, err := s.Mode()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			dontwant := mode
			if got == dontwant {
				t.Errorf("unexpected invalid mode value: got:%q, don't want:%q", got, dontwant)
			}
		}
	})

	t.Run("Command", func(t *testing.T) {
		for _, c := range conn {
			s, err := SensorFor(c.sensor.address, c.sensor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			commands, err := s.Commands()
			want := c.sensor.commands()
			if len(want) == 0 {
				if err == nil {
					t.Error("expected error getting commands from non-commandable sensor")
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
				err := s.Command(command).Err()
				if err != nil {
					t.Errorf("unexpected error for command %q: %v", command, err)
				}

				got := c.sensor.lastCommand()
				want := command
				if got != want {
					t.Errorf("unexpected command value: got:%q want:%q", got, want)
				}
			}
			for _, command := range []string{"invalid", "another"} {
				err := s.Command(command).Err()
				if err == nil {
					t.Errorf("expected error for command %q", command)
				}

				got := c.sensor.lastCommand()
				dontwant := command
				if got == dontwant {
					t.Errorf("unexpected invalid command value: got:%q don't want:%q", got, dontwant)
				}
			}
		}
	})

	t.Run("Binary data", func(t *testing.T) {
		for _, c := range conn {
			s, err := SensorFor(c.sensor.address, c.sensor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			format, err := s.BinDataFormat()
			if err != nil {
				t.Fatalf("unexpected error getting bin data format: %v", err)
			}
			wantFmt := c.sensor.binDataFormat()
			if format != wantFmt {
				t.Errorf("unexpected bin data format value: got:%q want:%q", format, wantFmt)
			}
			data, err := s.BinData()
			if err != nil {
				t.Fatalf("unexpected error getting bin data: %v", err)
			}
			wantData := c.sensor.binData()
			if !reflect.DeepEqual(data, wantData) {
				t.Errorf("unexpected bin data value: got:%#x want:%#x", data, wantData)
			}
		}
	})

	t.Run("Direct", func(t *testing.T) {
		s, err := SensorFor(conn[0].sensor.address, conn[0].sensor.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		f, err := s.Direct(os.O_RDWR)
		if err != nil {
			t.Fatalf("unexpected error getting direct connection: %v", err)
		}
		defer f.Close()
		got, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatalf("unexpected error reading direct connection: %v", err)
		}
		want := conn[0].sensor.direct()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("unexpected direct value: got:%q want:%q", got, want)
		}
		err = f.Truncate(0)
		if err != nil {
			t.Fatalf("unexpected error truncating: %v", err)
		}
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatalf("unexpected error seeking to start: %v", err)
		}
		new := []byte("new data")
		_, err = f.Write(new)
		if err != nil {
			t.Fatalf("unexpected error writing new data: %v", err)
		}
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatalf("unexpected error seeking to start: %v", err)
		}
		got, err = ioutil.ReadAll(f)
		if err != nil {
			t.Fatalf("unexpected error reading direct connection: %v", err)
		}
		if !reflect.DeepEqual(got, new) {
			t.Errorf("unexpected direct value: got:%q want:%q", got, new)
		}
	})

	t.Run("Poll rate", func(t *testing.T) {
		s, err := SensorFor(conn[0].sensor.address, conn[0].sensor.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, r := range []time.Duration{1 * time.Millisecond, 100 * time.Millisecond, 1 * time.Second} {
			err := s.SetPollRate(r).Err()
			if err != nil {
				t.Errorf("unexpected error for set poll rate %d: %v", r, err)
			}

			got, err := s.PollRate()
			if err != nil {
				t.Errorf("unexpected error getting poll rate: %v", err)
			}
			want := r
			if got != want {
				t.Errorf("unexpected poll rate value: got:%d want:%d", got, want)
			}
		}
	})

	t.Run("Units", func(t *testing.T) {
		s, err := SensorFor(conn[0].sensor.address, conn[0].sensor.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		modes, err := s.Modes()
		if err != nil {
			t.Fatalf("unexpected error getting modes: %v", err)
		}
		for _, mode := range modes {
			err := s.SetMode(mode).Err()
			if err != nil {
				t.Errorf("unexpected error for mode %q: %v", mode, err)
			}

			err = fs.InvalidatePath(filepath.Join(s.Path(), s.String(), UnitsName))
			if err != nil {
				t.Fatalf("unexpected error invalidating units: %v", err)
			}

			got, err := s.Units()
			if err != nil {
				t.Errorf("unexpected error getting units: %v", err)
			}
			want := conn[0].sensor.units()
			if got != want {
				t.Errorf("unexpected mode value: got:%q want:%q", got, want)
			}
		}
	})

	t.Run("Decimals", func(t *testing.T) {
		s, err := SensorFor(conn[0].sensor.address, conn[0].sensor.driver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		modes, err := s.Modes()
		if err != nil {
			t.Fatalf("unexpected error getting modes: %v", err)
		}
		for _, mode := range modes {
			err := s.SetMode(mode).Err()
			if err != nil {
				t.Errorf("unexpected error for mode %q: %v", mode, err)
			}

			err = fs.InvalidatePath(filepath.Join(s.Path(), s.String(), DecimalsName))
			if err != nil {
				t.Fatalf("unexpected error invalidating decimals: %v", err)
			}

			got, err := s.Decimals()
			if err != nil {
				t.Errorf("unexpected error getting decimals: %v", err)
			}
			want := conn[0].sensor.decimals()
			if got != want {
				t.Errorf("unexpected decimals value: got:%d want:%d", got, want)
			}
		}
	})

	t.Run("Number of values", func(t *testing.T) {
		for _, c := range conn {
			s, err := SensorFor(c.sensor.address, c.sensor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := s.NumValues()
			if err != nil {
				t.Errorf("unexpected error getting num values: %v", err)
			}
			want := len(c.sensor.values())
			if got != want {
				t.Errorf("unexpected num values: got:%d want:%d", got, want)
			}
		}
	})

	t.Run("Value", func(t *testing.T) {
		for _, c := range conn {
			s, err := SensorFor(c.sensor.address, c.sensor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			n, err := s.NumValues()
			if err != nil {
				t.Errorf("unexpected error getting num values: %v", err)
			}
			for i := 0; i < n; i++ {
				got, err := s.Value(i)
				if err != nil {
					t.Errorf("unexpected error getting value %d: %v", i, err)
				}
				want := c.sensor.values()[i]
				if got != want {
					t.Errorf("unexpected value: got:%q want:%q", got, want)
				}
			}
		}
	})

	t.Run("Text values", func(t *testing.T) {
		for _, c := range conn {
			s, err := SensorFor(c.sensor.address, c.sensor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := s.TextValues()
			if err != nil {
				t.Errorf("unexpected error getting text values: %v", err)
			}
			want := c.sensor.values()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected text values: got:%q want:%q", got, want)
			}
		}
	})

	t.Run("Uevent", func(t *testing.T) {
		for _, c := range conn {
			s, err := SensorFor(c.sensor.address, c.sensor.driver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := s.Uevent()
			if err != nil {
				t.Errorf("unexpected error getting uevent: %v", err)
			}
			want := c.sensor.uevent()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
			}
		}
	})
}
