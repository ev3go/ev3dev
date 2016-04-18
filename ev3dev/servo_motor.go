// Copyright ©2016 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ServoMotor represents a handle to a servo-motor.
type ServoMotor struct {
	mu sync.Mutex
	id int
}

// String satisfies the fmt.Stringer interface.
func (m *ServoMotor) String() string { return fmt.Sprint(motorPrefix, m.id) }

// ServoMotorFor returns a ServoMotor for the given ev3 port name and driver.
// If the motor driver does not match the driver string, a ServoMotor for the port
// is returned with a DriverMismatch error.
// If port is empty, the first servo-motor satisfying the driver name is returned.
func ServoMotorFor(port, driver string) (*ServoMotor, error) {
	id, err := deviceIDFor(port, driver, ServoMotorPath, motorPrefix)
	if id == -1 {
		return nil, err
	}
	return &ServoMotor{id: id}, err
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

// MaxPulseSetpoint returns the current max pulse set point value for the ServoMotor.
func (m *ServoMotor) MaxPulseSetpoint() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+maxPulseSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read max pulse set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse max pulse set point: %v", err)
	}
	return sp, nil
}

// SetMaxPulseSetpoint sets the max pulse set point value for the ServoMotor
func (m *ServoMotor) SetMaxPulseSetpoint(sp int) error {
	if sp < 2300 || sp > 2700 {
		return fmt.Errorf("ev3dev: invalid max pulse set point: %d (valid 2300-1700)", sp)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+maxPulseSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set max pulse set point: %v", err)
	}
	return nil
}

// MidPulseSetpoint returns the current mid pulse set point value for the ServoMotor.
func (m *ServoMotor) MidPulseSetpoint() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+midPulseSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read mid pulse set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse mid pulse set point: %v", err)
	}
	return sp, nil
}

// SetMidPulseSetpoint sets the mid pulse set point value for the ServoMotor
func (m *ServoMotor) SetMidPulseSetpoint(sp int) error {
	if sp < 1300 || sp > 1700 {
		return fmt.Errorf("ev3dev: invalid mid pulse set point: %d (valid 1300-1700)", sp)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+midPulseSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set mid pulse set point: %v", err)
	}
	return nil
}

// MinPulseSetpoint returns the current min pulse set point value for the ServoMotor.
func (m *ServoMotor) MinPulseSetpoint() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+minPulseSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read min pulse set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse min pulse set point: %v", err)
	}
	return sp, nil
}

// SetMinPulseSetpoint sets the min pulse set point value for the ServoMotor
func (m *ServoMotor) SetMinPulseSetpoint(sp int) error {
	if sp < 300 || sp > 700 {
		return fmt.Errorf("ev3dev: invalid min pulse set point: %d (valid 300 - 700)", sp)
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+minPulseSetpoint, m), fmt.Sprintln(sp))
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

// RateSetpoint returns the current rate set point value for the ServoMotor.
func (m *ServoMotor) RateSetpoint() (time.Duration, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+rateSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read rate set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse rate set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetRateSetpoint sets the rate set point value for the ServoMotor.
func (m *ServoMotor) SetRateSetpoint(d time.Duration) error {
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+rateSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
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
