// Copyright ©2016 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

// ServoMotor represents a handle to a servo-motor.
type ServoMotor struct {
	id int

	err error
}

// Path returns the servo-motor sysfs path.
func (*ServoMotor) Path() string { return ServoMotorPath }

// Type returns "motor".
func (*ServoMotor) Type() string { return motorPrefix }

// String satisfies the fmt.Stringer interface.
func (m *ServoMotor) String() string {
	if m == nil {
		return motorPrefix + "*"
	}
	return fmt.Sprint(motorPrefix, m.id)
}

// Err returns the error state of the ServoMotor and clears it.
func (m *ServoMotor) Err() error {
	err := m.err
	m.err = nil
	return err
}

// ServoMotorFor returns a ServoMotor for the given ev3 port name and driver.
// If the motor driver does not match the driver string, a ServoMotor for the port
// is returned with a DriverMismatch error.
// If port is empty, the first servo-motor satisfying the driver name is returned.
func ServoMotorFor(port, driver string) (*ServoMotor, error) {
	id, err := deviceIDFor(port, driver, (*ServoMotor)(nil))
	if id == -1 {
		return nil, err
	}
	return &ServoMotor{id: id}, err
}

func (m *ServoMotor) writeFile(path, data string) error {
	return ioutil.WriteFile(path, []byte(data), 0)
}

// Commands returns the available commands for the ServoMotor.
func (m *ServoMotor) Commands() []string {
	return []string{
		"run",
		"float",
	}
}

// Command issues a command to the ServoMotor.
func (m *ServoMotor) Command(comm string) *ServoMotor {
	if m.err != nil {
		return m
	}
	avail := m.Commands()
	ok := false
	for _, c := range avail {
		if c == comm {
			ok = true
			break
		}
	}
	if !ok {
		m.err = fmt.Errorf("ev3dev: command %q not available for %s (available:%q)", comm, m, avail)
		return m
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+command, m), comm)
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to issue servo-motor command: %v", err)
	}
	return m
}

// MaxPulseSetpoint returns the current max pulse set point value for the ServoMotor.
func (m *ServoMotor) MaxPulseSetpoint() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
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
func (m *ServoMotor) SetMaxPulseSetpoint(sp int) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 2300 || sp > 2700 {
		m.err = fmt.Errorf("ev3dev: invalid max pulse set point: %d (valid 2300-1700)", sp)
		return m
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+maxPulseSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set max pulse set point: %v", err)
	}
	return m
}

// MidPulseSetpoint returns the current mid pulse set point value for the ServoMotor.
func (m *ServoMotor) MidPulseSetpoint() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
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
func (m *ServoMotor) SetMidPulseSetpoint(sp int) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 1300 || sp > 1700 {
		m.err = fmt.Errorf("ev3dev: invalid mid pulse set point: %d (valid 1300-1700)", sp)
		return m
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+midPulseSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set mid pulse set point: %v", err)
	}
	return m
}

// MinPulseSetpoint returns the current min pulse set point value for the ServoMotor.
func (m *ServoMotor) MinPulseSetpoint() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
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
func (m *ServoMotor) SetMinPulseSetpoint(sp int) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 300 || sp > 700 {
		m.err = fmt.Errorf("ev3dev: invalid min pulse set point: %d (valid 300 - 700)", sp)
		return m
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+minPulseSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set min pulse set point: %v", err)
	}
	return m
}

// Polarity returns the current polarity of the ServoMotor.
func (m *ServoMotor) Polarity() (string, error) {
	if m.err != nil {
		return "", m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(ServoMotorPath+"/%s/"+polarity, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read polarity: %v", err)
	}
	return string(b), nil
}

// SetPolarity sets the polarity of the ServoMotor
func (m *ServoMotor) SetPolarity(p Polarity) *ServoMotor {
	if m.err != nil {
		return m
	}
	if p != Normal && p != Inversed {
		m.err = fmt.Errorf("ev3dev: invalid polarity: %q (valid \"normal\" or \"inversed\")", p)
		return m
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+polarity, m), string(p))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set polarity %v", err)
	}
	return m
}

// Position returns the current position value for the ServoMotor.
func (m *ServoMotor) Position() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
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
func (m *ServoMotor) SetPosition(pos int) *ServoMotor {
	if m.err != nil {
		return m
	}
	if pos != int(int32(pos)) {
		m.err = fmt.Errorf("ev3dev: invalid position: %d (valid in int32)", pos)
		return m
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+position, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set position: %v", err)
	}
	return m
}

// RateSetpoint returns the current rate set point value for the ServoMotor.
func (m *ServoMotor) RateSetpoint() (time.Duration, error) {
	if m.err != nil {
		return -1, m.Err()
	}
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
func (m *ServoMotor) SetRateSetpoint(d time.Duration) *ServoMotor {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(ServoMotorPath+"/%s/"+rateSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set rate set point: %v", err)
	}
	return m
}

// State returns the current state of the ServoMotor.
func (m *ServoMotor) State() (MotorState, error) {
	if m.err != nil {
		return 0, m.Err()
	}
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
