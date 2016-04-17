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

// ServoMotorPath is the path to the ev3 servo-motor file system.
const ServoMotorPath = "/sys/class/servo-motor"

// ServoMotor represents a handle to a servo-motor.
type ServoMotor struct {
	mu sync.Mutex
	id int
}

// String satisfies the fmt.Stringer interface.
func (m *ServoMotor) String() string { return fmt.Sprint(motorPrefix, m.id) }

// ServoMotorFor returns a ServoMotor for the given ev3 port name and driver.
// If port is empty, the first servo-motor satisfying the driver name is returned.
func ServoMotorFor(port, driver string) (*ServoMotor, error) {
	p, err := LegoPortFor(port)
	if err != nil {
		return nil, err
	}
	if p == nil {
		for id := 0; id < 8; id++ {
			m, err := ServoMotorFor(fmt.Sprint(portPrefix, id), driver)
			if err == nil {
				return m, err
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
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	if len(names) != 1 {
		return nil, fmt.Errorf("ev3dev: more than one device in path %s: %q", path, names)
	}
	name := names[0]
	if !strings.HasPrefix(name, motorPrefix) {
		return nil, fmt.Errorf("ev3dev: device in path %s not a motor: %q", path, name)
	}
	id, err := strconv.Atoi(strings.TrimPrefix(name, motorPrefix))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: could not parse id from device name %q: %v", name, err)
	}
	m := &ServoMotor{id: id}
	d, err := m.Driver()
	if err != nil {
		return nil, fmt.Errorf("ev3dev: could not get driver name: %v", err)
	}
	if d != driver {
		err = fmt.Errorf("ev3dev: mismatched driver names: want %q but have %q", driver, d)
	}
	return m, err
}

func (m *ServoMotor) writeFile(path, data string) error {
	defer m.mu.Unlock()
	m.mu.Lock()
	return ioutil.WriteFile(path, []byte(data), 0)
}

// Address returns the ev3 port name for the ServoMotor.
func (m *ServoMotor) Address() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+address, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port address: %v", err)
	}
	return string(chomp(b)), err
}

// Commands returns the available commands for the ServoMotor.
func (m *ServoMotor) Commands() []string {
	return []string{
		"run",
		"float",
	}
}

// Command issues a command to the ServoMotor.
func (m *ServoMotor) Command(comm string) error {
	avail := m.Commands()
	ok := false
	for _, c := range avail {
		if c == comm {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("ev3dev: command %q not available for %s (available:%q)", comm, m, avail)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+command, m), comm)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to issue servo-motor command: %v", err)
	}
	return nil
}

// Driver returns the driver name for the ServoMotor.
func (m *ServoMotor) Driver() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+driverName, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port driver name: %v", err)
	}
	return string(chomp(b)), err
}

// MaxPulseSetPoint returns the current max pulse set point value for the ServoMotor.
func (m *ServoMotor) MaxPulseSetPoint() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+maxPulseSetPoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read max pulse set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse max pulse set point: %v", err)
	}
	return sp, nil
}

// SetMaxPulseSetPoint sets the max pulse set point value for the ServoMotor
func (m *ServoMotor) SetMaxPulseSetPoint(sp int) error {
	if sp < 2300 || sp > 2700 {
		return fmt.Errorf("ev3dev: invalid max pulse set point: %d (valid 2300-1700)", sp)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+maxPulseSetPoint, m), fmt.Sprintln(sp))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set max pulse set point: %v", err)
	}
	return nil
}

// MidPulseSetPoint returns the current mid pulse set point value for the ServoMotor.
func (m *ServoMotor) MidPulseSetPoint() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+midPulseSetPoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read mid pulse set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse mid pulse set point: %v", err)
	}
	return sp, nil
}

// SetMidPulseSetPoint sets the mid pulse set point value for the ServoMotor
func (m *ServoMotor) SetMidPulseSetPoint(sp int) error {
	if sp < 1300 || sp > 1700 {
		return fmt.Errorf("ev3dev: invalid mid pulse set point: %d (valid 1300-1700)", sp)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+midPulseSetPoint, m), fmt.Sprintln(sp))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set mid pulse set point: %v", err)
	}
	return nil
}

// MinPulseSetPoint returns the current min pulse set point value for the ServoMotor.
func (m *ServoMotor) MinPulseSetPoint() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+minPulseSetPoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read min pulse set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse min pulse set point: %v", err)
	}
	return sp, nil
}

// SetMinPulseSetPoint sets the min pulse set point value for the ServoMotor
func (m *ServoMotor) SetMinPulseSetPoint(sp int) error {
	if sp < 300 || sp > 700 {
		return fmt.Errorf("ev3dev: invalid min pulse set point: %d (valid 300 - 700)", sp)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+minPulseSetPoint, m), fmt.Sprintln(sp))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set min pulse set point: %v", err)
	}
	return nil
}

// Polarity returns the current polarity of the ServoMotor.
func (m *ServoMotor) Polarity() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+polarity, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read polarity: %v", err)
	}
	return string(b), nil
}

// SetPolarity sets the polarity of the ServoMotor
func (m *ServoMotor) SetPolarity(p Polarity) error {
	if p != Normal && p != Inversed {
		return fmt.Errorf("ev3dev: invalid polarity: %q (valid \"normal\" or \"inversed\")", p)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+polarity, m), string(p))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set polarity %v", err)
	}
	return nil
}

// Position returns the current position value for the ServoMotor.
func (m *ServoMotor) Position() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+position, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read position: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse position: %v", err)
	}
	return pos, nil
}

// SetPosition sets the position value for the ServoMotor.
func (m *ServoMotor) SetPosition(pos int) error {
	if pos != int(int32(pos)) {
		return fmt.Errorf("ev3dev: invalid position: %d (valid in int32)", pos)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+position, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set position: %v", err)
	}
	return nil
}

// RateSetPoint returns the current rate set point value for the ServoMotor.
func (m *ServoMotor) RateSetPoint() (time.Duration, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+rateSetPoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read rate set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse rate set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetRateSetPoint sets the rate set point value for the ServoMotor.
func (m *ServoMotor) SetRateSetPoint(d time.Duration) error {
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+rateSetPoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set rate set point: %v", err)
	}
	return nil
}

// State returns the current state of the ServoMotor.
func (m *ServoMotor) State() (MotorState, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+commands, m))
	if err != nil {
		return 0, fmt.Errorf("ev3dev: failed to read servo-motor commands: %v", err)
	}
	var stat MotorState
	for _, s := range strings.Split(string(chomp(b)), " ") {
		stat |= motorStateTable[s]
	}
	return stat, nil
}
