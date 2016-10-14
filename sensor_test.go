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

	lastCommand string
	commands    []string

	direct []byte

	mode     string
	modes    []string
	units    map[string]string
	decimals map[string]int

	binData       []byte
	binDataFormat string

	values []string

	pollRate int

	uevent map[string]string

	t *testing.T
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
	if len(s.commands) == 0 {
		return len(b), syscall.ENOTSUP
	}
	return readAt(b, offset, s)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorCommands) Size() (int64, error) {
	return size(s), nil
}

func (s *sensorCommands) String() string {
	sort.Strings(s.commands)
	return strings.Join(s.commands, " ")
}

// sensorCommand is the command attribute.
type sensorCommand sensor

// Truncate is a no-op.
func (s *sensorCommand) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (s *sensorCommand) WriteAt(b []byte, off int64) (int, error) {
	if len(s.commands) == 0 {
		return len(b), syscall.ENOTSUP
	}
	command := string(chomp(b))
	for _, c := range s.commands {
		if command == c {
			s.lastCommand = command
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (s *sensorCommand) Size() (int64, error) {
	return size(s.lastCommand), nil
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

func (s *sensorModes) String() string {
	sort.Strings(s.modes)
	return strings.Join(s.modes, " ")
}

// sensorMode is the mode attribute.
type sensorMode sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorMode) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s.mode)
}

// Truncate is a no-op.
func (s *sensorMode) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (s *sensorMode) WriteAt(b []byte, off int64) (int, error) {
	mode := string(chomp(b))
	for _, c := range s.modes {
		if mode == c {
			s.mode = mode
			return len(b), nil
		}
	}
	return len(b), syscall.EINVAL
}

// Size returns the length of the backing data and a nil error.
func (s *sensorMode) Size() (int64, error) {
	return size(s.mode), nil
}

// sensorBinDataFormat is the bin_data_format attribute.
type sensorBinDataFormat sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorBinDataFormat) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s.binDataFormat)
}

// Size returns the length of the backing data and a nil error.
func (s *sensorBinDataFormat) Size() (int64, error) {
	return size(s.binDataFormat), nil
}

// sensorBinData is the bin_datas attribute.
type sensorBinData sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorBinData) ReadAt(b []byte, offset int64) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	if offset >= int64(len(s.binData)) {
		return 0, io.EOF
	}
	n := copy(b, s.binData[offset:])
	if n <= len(b) {
		return n, io.EOF
	}
	return n, nil
}

// Size returns the length of the backing data and a nil error.
func (s *sensorBinData) Size() (int64, error) {
	return int64(len(s.binData)), nil
}

// sensorDirect is the direct attribute.
type sensorDirect sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorDirect) ReadAt(b []byte, offset int64) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	if offset >= int64(len(s.direct)) {
		return 0, io.EOF
	}
	n := copy(b, s.direct[offset:])
	if n <= len(b) {
		return n, io.EOF
	}
	return n, nil
}

// Truncate truncates the Bytes at n bytes from the beginning of the slice.
func (s *sensorDirect) Truncate(n int64) error {
	if n < 0 || n > int64(len(s.direct)) {
		return syscall.EINVAL
	}
	tail := s.direct[n:cap(s.direct)]
	for i := range tail {
		tail[i] = 0
	}
	s.direct = s.direct[:n]
	return nil
}

// WriteAt satisfies the io.WriterAt interface.
func (s *sensorDirect) WriteAt(b []byte, off int64) (int, error) {
	if off >= int64(cap(s.direct)) {
		t := make([]byte, off+int64(len(b)))
		copy(t, s.direct)
		s.direct = t
	}
	s.direct = s.direct[:off]
	s.direct = append(s.direct, b...)
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (s *sensorDirect) Size() (int64, error) { return int64(len(s.direct)), nil }

// sensorUnits is the units attribute.
type sensorUnits sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorUnits) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s.units[s.mode])
}

// Size returns the length of the backing data and a nil error.
func (s *sensorUnits) Size() (int64, error) {
	return size(s.units[s.mode]), nil
}

// sensorPollRate is the poll_rate attribute.
type sensorPollRate sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorPollRate) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s.pollRate)
}

// Truncate is a no-op.
func (s *sensorPollRate) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (s *sensorPollRate) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		s.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	s.pollRate = i
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (s *sensorPollRate) Size() (int64, error) {
	return size(s.pollRate), nil
}

// sensorDecimals is the decimals attribute.
type sensorDecimals sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorDecimals) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, s.decimals[s.mode])
}

// Size returns the length of the backing data and a nil error.
func (s *sensorDecimals) Size() (int64, error) {
	return size(s.decimals[s.mode]), nil
}

// sensorNumValues is the num_values attribute.
type sensorNumValues sensor

// ReadAt satisfies the io.ReaderAt interface.
func (s *sensorNumValues) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, len(s.values))
}

// Size returns the length of the backing data and a nil error.
func (s *sensorNumValues) Size() (int64, error) {
	return size(len(s.values)), nil
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

func (s sensorValue) String() string {
	if len(s.values) <= s.n {
		return "\n"
	}
	return s.values[s.n]
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

func (s *sensorTextValues) String() string {
	return strings.Join(s.values, " ")
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

func (s *sensorUevent) String() string {
	e := make([]string, 0, len(s.uevent))
	for k, v := range s.uevent {
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

				modes:    []string{"GYRO-ANG", "GYRO-RATE", "GYRO-FAS", "GYRO-G&A", "GYRO-CAL"},
				mode:     "GYRO-ANG",
				units:    map[string]string{"GYRO-ANG": "deg", "GYRO-RATE": "d/s"},
				decimals: map[string]int{"GYRO-ANG": 0, "GYRO-RATE": 1},

				commands: []string{"operate", "twirl"},

				binDataFormat: "s16",
				binData:       []byte{0x01, 0x00, 0x02, 0x00},

				values: []string{"1", "2"},

				direct: []byte("initial state"),

				uevent: map[string]string{
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

				modes: []string{"GYRO-ANG", "GYRO-RATE", "GYRO-FAS", "GYRO-G&A", "GYRO-CAL"},
				mode:  "GYRO-ANG",
				units: map[string]string{"GYRO-ANG": "deg", "GYRO-RATE": "d/s"},

				binDataFormat: "s16",
				binData:       []byte{0x01, 0x00, 0x02, '\n'},

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
		if !reflect.DeepEqual(modes, conn[0].sensor.modes) {
			t.Errorf("unexpected modes value: got:%q want:%q", modes, conn[0].sensor.modes)
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
			if len(c.sensor.commands) == 0 {
				if err == nil {
					t.Error("expected error getting commands from non-commandable sensor")
				}
				continue
			}
			if err != nil {
				t.Fatalf("unexpected error getting commands: %v", err)
			}
			if !reflect.DeepEqual(commands, c.sensor.commands) {
				t.Errorf("unexpected commands value: got:%q want:%q", commands, c.sensor.commands)
			}
			for _, command := range commands {
				err := s.Command(command).Err()
				if err != nil {
					t.Errorf("unexpected error for command %q: %v", command, err)
				}

				got := c.sensor.lastCommand
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

				got := c.sensor.lastCommand
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
			if format != c.sensor.binDataFormat {
				t.Errorf("unexpected bin data format value: got:%q want:%q", format, c.sensor.binDataFormat)
			}
			data, err := s.BinData()
			if err != nil {
				t.Fatalf("unexpected error getting bin data: %v", err)
			}
			if !reflect.DeepEqual(data, c.sensor.binData) {
				t.Errorf("unexpected bin data value: got:%#x want:%#x", data, c.sensor.binData)
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
		if !reflect.DeepEqual(got, conn[0].sensor.direct) {
			t.Errorf("unexpected direct value: got:%q want:%q", got, conn[0].sensor.direct)
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
			want := conn[0].sensor.units[mode]
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
			want := conn[0].sensor.decimals[mode]
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
			want := len(c.sensor.values)
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
				want := c.sensor.values[i]
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
			want := c.sensor.values
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
			want := c.sensor.uevent
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
			}
		}
	})
}
