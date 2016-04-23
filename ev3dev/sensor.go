// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"os"
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

// BinData returns the unscaled raw values from the Sensor.
func (s *Sensor) BinData() (string, error) {
	return stringFrom(attributeOf(s, binData))
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
	return stringFrom(attributeOf(s, binDataFormat))
}

// Commands returns the available commands for the Sensor.
func (s *Sensor) Commands() ([]string, error) {
	return stringSliceFrom(attributeOf(s, commands))
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
	s.err = setAttributeOf(s, command, comm)
	return s
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
	return intFrom(attributeOf(s, decimals))
}

// Modes returns the available modes for the Sensor.
func (s *Sensor) Modes() ([]string, error) {
	return stringSliceFrom(attributeOf(s, modes))
}

// Mode returns the currently selected mode of the Sensor.
func (s *Sensor) Mode() (string, error) {
	return stringFrom(attributeOf(s, mode))
}

// SetMode sets the mode of the Sensor.
func (s *Sensor) SetMode(m string) *Sensor {
	if s.err != nil {
		return s
	}
	s.err = setAttributeOf(s, mode, m)
	return s
}

// NumValues returns number of values available from the Sensor.
func (s *Sensor) NumValues() (int, error) {
	return intFrom(attributeOf(s, numValues))
}

// PollRate returns the current polling rate value for the Sensor.
func (s *Sensor) PollRate() (time.Duration, error) {
	return durationFrom(attributeOf(s, pollRate))
}

// SetPollRate sets the polling rate value for the Sensor.
func (s *Sensor) SetPollRate(d time.Duration) *Sensor {
	if s.err != nil {
		return s
	}
	s.err = setAttributeOf(s, pollRate, fmt.Sprintln(int(d/time.Millisecond)))
	return s
}

// Units returns the units of the measured value for the current mode for the Sensor.
func (s *Sensor) Units() (string, error) {
	return stringFrom(attributeOf(s, units))
}

// Value returns tthe value or values measured by the Sensor. Value will return
// and error if n is greater than or equal to the value returned by NumValues.
func (s *Sensor) Value(n int) (string, error) {
	return stringFrom(attributeOf(s, value))
}

// TextValues returns slice of strings string representing sensor-specific text values.
func (s *Sensor) TextValues() ([]string, error) {
	return stringSliceFrom(attributeOf(s, textValues))
}
