// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// Sensor represents a handle to a lego-sensor.
type Sensor struct {
	id int

	err error
}

// Path returns the lego-sensor sysfs path.
func (*Sensor) Path() string { return SensorPath }

// Type returns "sensor".
func (*Sensor) Type() string { return sensorPrefix }

// String satisfies the fmt.Stringer interface.
func (s *Sensor) String() string {
	if s == nil {
		return sensorPrefix + "*"
	}
	return fmt.Sprint(sensorPrefix, s.id)
}

// Err returns the error state of the Sensor and clears it.
func (s *Sensor) Err() error {
	err := s.err
	s.err = nil
	return err
}

// SensorFor returns a Sensor for the given ev3 port name and driver. If the
// sensor driver does not match the driver string, a Sensor for the port
// is returned with a DriverMismatch error.
// If port is empty, the first sensor satisfying the driver name is returned.
func SensorFor(port, driver string) (*Sensor, error) {
	id, err := deviceIDFor(port, driver, (*Sensor)(nil))
	if id == -1 {
		return nil, err
	}
	return &Sensor{id: id}, err
}

func (s *Sensor) writeFile(path, data string) error {
	return ioutil.WriteFile(path, []byte(data), 0)
}

// BinData returns the unscaled raw values from the Sensor.
func (s *Sensor) BinData() (string, error) {
	if s.err != nil {
		return "", s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+binData, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read bin data: %v", err)
	}
	return string(chomp(b)), err
}

// BinDataFormat returns the format of the values returned by BinData for the
// current mode.
//
// The returned values should be interpretted according to:
//
//  u8: Unsigned 8-bit integer (byte)
//  s8: Signed 8-bit integer (sbyte)
//  u16: Unsigned 16-bit integer (ushort)
//  s16: Signed 16-bit integer (short)
//  s16_be: Signed 16-bit integer, big endian
//  s32: Signed 32-bit integer (int)
//  s32_be: Signed 32-bit integer, big endian
//  float: IEEE 754 32-bit floating point (float)
func (s *Sensor) BinDataFormat() (string, error) {
	if s.err != nil {
		return "", s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+binDataFormat, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read bin data format: %v", err)
	}
	return string(chomp(b)), err
}

// Command issues a command to the Sensor.
func (s *Sensor) Command(comm string) *Sensor {
	if s.err != nil {
		return s
	}
	avail, err := s.Commands()
	if err != nil {
		s.err = err
		return s
	}
	ok := false
	for _, c := range avail {
		if c == comm {
			ok = true
			break
		}
	}
	if !ok {
		s.err = fmt.Errorf("ev3dev: command %q not available for %s (available:%q)", comm, s, avail)
		return s
	}
	err = s.writeFile(fmt.Sprintf(SensorPath+"/%s/"+command, s), comm)
	if err != nil {
		s.err = fmt.Errorf("ev3dev: failed to issue sensor command: %v", err)
	}
	return s
}

// Commands returns the available commands for the Sensor.
func (s *Sensor) Commands() ([]string, error) {
	if s.err != nil {
		return nil, s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+commands, s))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read sensor commands: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}

// Direct returns a file that can be used to directly communication
// with the sensor for using advanced features that are not otherwise
// available through the lego-sensor class. It is the responsibility
// of the user to provide the correct file operation flags, and to
// close the file after use.
func (s *Sensor) Direct(flag int) (*os.File, error) {
	if s.err != nil {
		return nil, s.Err()
	}
	return os.OpenFile(fmt.Sprintf(SensorPath+"/%s/"+direct, s), flag, 0)
}

// Decimals returns the number of decimal places for the values in the
// attributes of the current mode.
func (s *Sensor) Decimals() (int, error) {
	if s.err != nil {
		return -1, s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+decimals, s))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read number of decimals: %v", err)
	}
	places, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse number of decimals: %v", err)
	}
	return places, nil
}

// Modes returns the available modes for the Sensor.
func (s *Sensor) Modes() ([]string, error) {
	if s.err != nil {
		return nil, s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+modes, s))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read sensor modes: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), err
}

// Mode returns the currently selected mode of the Sensor.
func (s *Sensor) Mode() (string, error) {
	if s.err != nil {
		return "", s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+mode, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read sensor mode: %v", err)
	}
	return string(chomp(b)), err
}

// SetMode sets the mode of the Sensor.
func (s *Sensor) SetMode(mode string) *Sensor {
	if s.err != nil {
		return s
	}
	err := s.writeFile(fmt.Sprintf(SensorPath+"/%s/"+mode, s), mode)
	if err != nil {
		s.err = fmt.Errorf("ev3dev: failed to set sensor mode: %v", err)
	}
	return s
}

// NumValues returns number of values available from the Sensor.
func (s *Sensor) NumValues() (int, error) {
	if s.err != nil {
		return -1, s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+numValues, s))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read number of values available: %v", err)
	}
	places, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse number of values available: %v", err)
	}
	return places, nil
}

// PollRate returns the current polling rate value for the Sensor.
func (s *Sensor) PollRate() (time.Duration, error) {
	if s.err != nil {
		return -1, s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+pollRate, s))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read rate set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse rate set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetPollRate sets the polling rate value for the Sensor.
func (s *Sensor) SetPollRate(d time.Duration) *Sensor {
	if s.err != nil {
		return s
	}
	err := s.writeFile(fmt.Sprintf(SensorPath+"/%s/"+pollRate, s), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		s.err = fmt.Errorf("ev3dev: failed to set rate set point: %v", err)
	}
	return s
}

// Units returns the units of the measured value for the current mode for the Sensor.
func (s *Sensor) Units() (string, error) {
	if s.err != nil {
		return "", s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+units, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read units: %v", err)
	}
	return string(chomp(b)), err
}

// Value returns tthe value or values measured by the Sensor. Value will return
// and error if n is greater than or equal to the value returned by NumValues.
func (s *Sensor) Value(n int) (string, error) {
	if s.err != nil {
		return "", s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+value+"%d", s, n))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read value%d: %v", n, err)
	}
	return string(chomp(b)), err
}

// TextValues returns slice of strings string representing sensor-specific text values.
func (s *Sensor) TextValues() ([]string, error) {
	if s.err != nil {
		return nil, s.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+textValues, s))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read text values: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}
