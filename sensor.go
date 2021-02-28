// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var _ idSetter = (*Sensor)(nil)

// Sensor represents a handle to a lego-sensor.
type Sensor struct {
	id int

	// Cached values:
	driver, firmwareVersion string
	commands, modes         []string

	// Mode cached values:
	decimals, numValues        int
	mode, units, binDataFormat string

	err error
}

// Path returns the lego-sensor sysfs path.
func (*Sensor) Path() string { return filepath.Join(prefix, SensorPath) }

// Type returns "sensor".
func (*Sensor) Type() string { return sensorPrefix }

// String satisfies the fmt.Stringer interface.
func (s *Sensor) String() string {
	if s == nil {
		return sensorPrefix + "*"
	}
	return sensorPrefix + strconv.Itoa(s.id)
}

// Err returns the error state of the Sensor and clears it.
func (s *Sensor) Err() error {
	err := s.err
	s.err = nil
	return err
}

// idInt and setID satisfy the idSetter interface.
func (s *Sensor) setID(id int) error {
	t := Sensor{id: id}
	var err error
	t.firmwareVersion, err = stringFrom(attributeOf(&t, firmwareVersion))
	if err != nil {
		goto fail
	}
	t.commands, err = stringSliceFrom(attributeOf(&t, commands))
	if err != nil {
		goto fail
	}
	t.modes, err = stringSliceFrom(attributeOf(&t, modes))
	if err != nil {
		goto fail
	}
	t.driver, err = DriverFor(&t)
	if err != nil {
		goto fail
	}
	err = t.cacheModeAttrs()
	if err != nil {
		goto fail
	}
	*s = t
	return nil

fail:
	*s = Sensor{id: -1}
	return err
}
func (s *Sensor) idInt() int {
	if s == nil {
		return -1
	}
	return s.id
}

// SensorFor returns a Sensor for the given ev3 port name and driver. If the
// sensor driver does not match the driver string, a Sensor for the port
// is returned with a DriverMismatch error.
// If port is empty, the first sensor satisfying the driver name is returned.
func SensorFor(port, driver string) (*Sensor, error) {
	id, err := deviceIDFor(port, driver, (*Sensor)(nil), -1)
	if id == -1 {
		return nil, err
	}
	var s Sensor
	_err := s.setID(id)
	if _err != nil {
		err = _err
	}
	return &s, err
}

// Next returns a Sensor for the next sensor with the same device driver as
// the receiver.
func (s *Sensor) Next() (*Sensor, error) {
	driver, err := DriverFor(s)
	if err != nil {
		return nil, err
	}
	id, err := deviceIDFor("", driver, (*Sensor)(nil), s.id)
	if id == -1 {
		return nil, err
	}
	return &Sensor{id: id}, err
}

// BinData returns the unscaled raw values from the Sensor.
func (s *Sensor) BinData() ([]byte, error) {
	err := s.Err()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(s.Path(), s.String(), binData)
	b, err := readFile(path)
	if err != nil {
		return nil, newAttrOpError(s, binData, string(b), "read", err)
	}
	return b, nil
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
func (s *Sensor) BinDataFormat() string {
	return s.binDataFormat
}

// Driver returns the driver used by the Sensor.
func (s *Sensor) Driver() string {
	return s.driver
}

// Commands returns the available commands for the Sensor.
func (s *Sensor) Commands() []string {
	if s.commands == nil {
		return nil
	}
	// Return a copy to prevent users
	// changing the values under our feet.
	avail := make([]string, len(s.commands))
	copy(avail, s.commands)
	return avail
}

// Command issues a command to the Sensor.
func (s *Sensor) Command(comm string) *Sensor {
	if s.err != nil {
		return s
	}
	ok := false
	for _, c := range s.commands {
		if c == comm {
			ok = true
			break
		}
	}
	if !ok {
		s.err = newInvalidValueError(s, command, "", comm, s.Commands())
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
	return os.OpenFile(filepath.Join(s.Path(), s.String(), direct), flag, 0)
}

// Decimals returns the number of decimal places for the values in the
// attributes of the current mode.
func (s *Sensor) Decimals() int {
	return s.decimals
}

// FirmwareVersion returns the firmware version of the Sensor.
func (s *Sensor) FirmwareVersion() string {
	return s.firmwareVersion
}

// Modes returns the available modes for the Sensor.
func (s *Sensor) Modes() []string {
	if s.modes == nil {
		return nil
	}
	// Return a copy to prevent users
	// changing the values under our feet.
	avail := make([]string, len(s.modes))
	copy(avail, s.modes)
	return avail
}

// Mode returns the currently selected mode of the Sensor.
func (s *Sensor) Mode() (string, error) {
	return stringFrom(attributeOf(s, mode))
}

// SetMode sets the mode of the Sensor. Calling SetMode invalidates and refreshes
// cached values for BinDataFormat, Decimals, Mode, NumValues and Units.
func (s *Sensor) SetMode(m string) *Sensor {
	if s.err != nil {
		return s
	}
	ok := false
	for _, a := range s.modes {
		if a == m {
			ok = true
			break
		}
	}
	if !ok {
		s.err = newInvalidValueError(s, mode, "", m, s.Modes())
		return s
	}
	s.err = setAttributeOf(s, mode, m)
	if s.err == nil {
		s.err = s.cacheModeAttrs()
	}
	return s
}

func (s *Sensor) cacheModeAttrs() error {
	var err error
	s.decimals, err = intFrom(attributeOf(s, decimals))
	if err != nil {
		return err
	}
	s.numValues, err = intFrom(attributeOf(s, numValues))
	if err != nil {
		return err
	}
	s.mode, err = stringFrom(attributeOf(s, mode))
	if err != nil {
		return err
	}
	s.units, err = stringFrom(attributeOf(s, units))
	if err != nil {
		return err
	}
	s.binDataFormat, err = stringFrom(attributeOf(s, binDataFormat))
	if err != nil {
		return err
	}
	return nil
}

// NumValues returns number of values available from the Sensor.
func (s *Sensor) NumValues() int {
	return s.numValues
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
	s.err = setAttributeOf(s, pollRate, strconv.Itoa(int(d/time.Millisecond)))
	return s
}

// Units returns the units of the measured value for the current mode for the Sensor.
func (s *Sensor) Units() string {
	return s.units
}

// Value returns tthe value or values measured by the Sensor. Value will return
// and error if n is greater than or equal to the value returned by NumValues.
func (s *Sensor) Value(n int) (string, error) {
	return stringFrom(attributeOf(s, value+strconv.Itoa(n)))
}

// TextValues returns slice of strings string representing sensor-specific text values.
func (s *Sensor) TextValues() ([]string, error) {
	return stringSliceFrom(attributeOf(s, textValues))
}

// Uevent returns the current uevent state for the Sensor.
func (s *Sensor) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(s, uevent))
}
