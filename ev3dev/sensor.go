// Copyright ©2016 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const sensor = "lego-sensor"

// Sensor represents a handle to a lego-sensor.
type Sensor struct {
	mu sync.Mutex
	id int
}

// String satisfies the fmt.Stringer interface.
func (s *Sensor) String() string { return fmt.Sprint(sensorPrefix, s.id) }

// SensorFor returns a Sensor for the given ev3 port name and driver. If the
// sensor driver does not match the driver string, a Sensor for the port
// is returned with a DriverMismatch error.
// If port is empty, the first sensor satisfying the driver name is returned.
func SensorFor(port, driver string) (*Sensor, error) {
	p, err := LegoPortFor(port)
	if err != nil {
		return nil, err
	}
	if p == nil {
		for id := 0; id < 8; id++ {
			s, err := SensorFor(fmt.Sprint(portPrefix, id), driver)
			if err == nil {
				return s, err
			}
		}
		return nil, fmt.Errorf("ev3dev: could not find device for driver %q", driver)
	}

	dev, err := ConnectedTo(p)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(LegoPortPath, p.String(), dev)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	files, err := f.Readdirnames(0)
	f.Close()
	if err != nil {
		return nil, err
	}
	var mapping string
	for _, n := range files {
		parts := strings.SplitN(n, ":", 2)
		if parts[0] == port {
			mapping = n
			break
		}
	}
	path = filepath.Join(path, mapping, sensor)
	f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	files, err = f.Readdirnames(0)
	f.Close()
	if len(files) != 1 {
		return nil, fmt.Errorf("ev3dev: more than one device in path %s: %q", path, files)
	}
	device := files[0]
	if !strings.HasPrefix(device, sensorPrefix) {
		return nil, fmt.Errorf("ev3dev: device in path %s not a sensor: %q", path, device)
	}
	id, err := strconv.Atoi(strings.TrimPrefix(device, sensorPrefix))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: could not parse id from device name %q: %v", device, err)
	}
	s := &Sensor{id: id}
	d, err := s.Driver()
	if err != nil {
		return nil, fmt.Errorf("ev3dev: could not get driver name: %v", err)
	}
	if d != driver {
		err = DriverMismatch{Want: driver, Have: d}
	}
	return s, err
}

func (s *Sensor) writeFile(path, data string) error {
	defer s.mu.Unlock()
	s.mu.Lock()
	return ioutil.WriteFile(path, []byte(data), 0)
}

// Address returns the ev3 port name for the Sensor.
func (s *Sensor) Address() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+address, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port address: %v", err)
	}
	return string(chomp(b)), err
}

// BinData returns the unscaled raw values from the Sensor.
func (s *Sensor) BinData() (string, error) {
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
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+binDataFormat, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read bin data format: %v", err)
	}
	return string(chomp(b)), err
}

// Command issues a command to the Sensor.
func (s *Sensor) Command(comm string) error {
	avail, err := s.Commands()
	if err != nil {
		return err
	}
	ok := false
	for _, c := range avail {
		if c == comm {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("ev3dev: command %q not available for %s (available:%q)", comm, s, avail)
	}
	err = s.writeFile(fmt.Sprintf(SensorPath+"/%s/"+command, s), comm)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to issue sensor command: %v", err)
	}
	return nil
}

// Commands returns the available commands for the Sensor.
func (s *Sensor) Commands() ([]string, error) {
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
	return os.OpenFile(fmt.Sprintf(SensorPath+"/%s/"+direct, s), flag, 0)
}

// Decimals returns the number of decimal places for the values in the
// attributes of the current mode.
func (s *Sensor) Decimals() (int, error) {
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

// Driver returns the driver name for the Sensor.
func (s *Sensor) Driver() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+driverName, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read sensor driver name: %v", err)
	}
	return string(chomp(b)), err
}

// Modes returns the available modes for the Sensor.
func (s *Sensor) Modes() ([]string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+modes, s))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read sensor modes: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), err
}

// Mode returns the currently selected mode of the Sensor.
func (s *Sensor) Mode() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+mode, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read sensor mode: %v", err)
	}
	return string(chomp(b)), err
}

// SetMode sets the mode of the Sensor.
func (s *Sensor) SetMode(mode string) error {
	err := s.writeFile(fmt.Sprintf(SensorPath+"/%s/"+mode, s), mode)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set sensor mode: %v", err)
	}
	return nil
}

// NumValues returns number of values available from the Sensor.
func (s *Sensor) NumValues() (int, error) {
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
func (s *Sensor) SetPollRate(d time.Duration) error {
	err := s.writeFile(fmt.Sprintf(SensorPath+"/%s/"+pollRate, s), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set rate set point: %v", err)
	}
	return nil
}

// Units returns the units of the measured value for the current mode for the Sensor.
func (s *Sensor) Units() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+units, s))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read units: %v", err)
	}
	return string(chomp(b)), err
}

// Value returns tthe value or values measured by the Sensor. Value will return
// and error if n is greater than or equal to the value returned by NumValues.
func (s *Sensor) Value(n int) (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+value+"%d", s, n))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read value%d: %v", n, err)
	}
	return string(chomp(b)), err
}

// TextValues returns slice of strings string representing sensor-specific text values.
func (s *Sensor) TextValues() ([]string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%s/"+textValues, s))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read text values: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}
