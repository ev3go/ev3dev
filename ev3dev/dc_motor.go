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

// DCMotor represents a handle to a dc-motor.
type DCMotor struct {
	mu sync.Mutex
	id int
}

// Path returns the dc-motor sysfs path.
func (*DCMotor) Path() string { return DCMotorPath }

// Path returns "motor".
func (*DCMotor) Type() string { return motorPrefix }

// String satisfies the fmt.Stringer interface.
func (m *DCMotor) String() string {
	if m == nil {
		return motorPrefix + "*"
	}
	return fmt.Sprint(motorPrefix, m.id)
}

// DCMotorFor returns a DCMotor for the given ev3 port name and driver. If the
// motor driver does not match the driver string, a DCMotor for the port is
// returned with a DriverMismatch error.
// If port is empty, the first dc-motor satisfying the driver name is returned.
func DCMotorFor(port, driver string) (*DCMotor, error) {
	id, err := deviceIDFor(port, driver, (*DCMotor)(nil))
	if id == -1 {
		return nil, err
	}
	return &DCMotor{id: id}, err
}

func (m *DCMotor) writeFile(path, data string) error {
	defer m.mu.Unlock()
	m.mu.Lock()
	return ioutil.WriteFile(path, []byte(data), 0)
}

// Commands returns the available commands for the DCMotor.
func (m *DCMotor) Commands() ([]string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+commands, m))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read dc-motor commands: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}

// Command issues a command to the DCMotor.
func (m *DCMotor) Command(comm string) error {
	avail, err := m.Commands()
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
		return fmt.Errorf("ev3dev: command %q not available for %s (available:%q)", comm, m, avail)
	}
	err = m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+command, m), comm)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to issue dc-motor command: %v", err)
	}
	return nil
}

// DutyCycle returns the current duty cycle value for the DCMotor.
func (m *DCMotor) DutyCycle() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+dutyCycle, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read duty cycle: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse duty cycle: %v", err)
	}
	return sp, nil
}

// DutyCycleSetpoint returns the current duty cycle set point value for the DCMotor.
func (m *DCMotor) DutyCycleSetpoint() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+dutyCycleSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read duty cycle set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse duty cycle set point: %v", err)
	}
	return sp, nil
}

// SetDutyCycleSetpoint sets the duty cycle set point value for the DCMotor
func (m *DCMotor) SetDutyCycleSetpoint(sp int) error {
	if sp < -100 || sp > 100 {
		return fmt.Errorf("ev3dev: invalid duty cycle set point: %d (valid -100 - 100)", sp)
	}
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+dutyCycleSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set duty cycle set point: %v", err)
	}
	return nil
}

// Polarity returns the current polarity of the DCMotor.
func (m *DCMotor) Polarity() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+polarity, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read polarity: %v", err)
	}
	return string(b), nil
}

// SetPolarity sets the polarity of the DCMotor
func (m *DCMotor) SetPolarity(p Polarity) error {
	if p != Normal && p != Inversed {
		return fmt.Errorf("ev3dev: invalid polarity: %q (valid \"normal\" or \"inversed\")", p)
	}
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+polarity, m), string(p))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set polarity %v", err)
	}
	return nil
}

// RampUpSetpoint returns the current ramp up set point value for the DCMotor.
func (m *DCMotor) RampUpSetpoint() (time.Duration, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+rampUpSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read ramp up set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse ramp up set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetRampUpSetpoint sets the ramp up set point value for the DCMotor.
func (m *DCMotor) SetRampUpSetpoint(d time.Duration) error {
	if d < 0 || d > 10000 {
		return fmt.Errorf("ev3dev: invalid ramp up set point: %v (must be positive)", d)
	}
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+rampUpSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set ramp up set point: %v", err)
	}
	return nil
}

// RampDownSetpoint returns the current ramp down set point value for the DCMotor.
func (m *DCMotor) RampDownSetpoint() (time.Duration, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+rampDownSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read ramp down set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse ramp down set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetRampDownSetpoint sets the ramp down set point value for the DCMotor.
func (m *DCMotor) SetRampDownSetpoint(d time.Duration) error {
	if d < 0 || d > 10000 {
		return fmt.Errorf("ev3dev: invalid ramp down set point: %v (must be positive)", d)
	}
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+rampDownSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set ramp down set point: %v", err)
	}
	return nil
}

// State returns the current state of the DCMotor.
func (m *DCMotor) State() (MotorState, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+commands, m))
	if err != nil {
		return 0, fmt.Errorf("ev3dev: failed to read dc-motor commands: %v", err)
	}
	var stat MotorState
	for _, s := range strings.Split(string(chomp(b)), " ") {
		stat |= motorStateTable[s]
	}
	return stat, nil
}

// StopAction returns the stop action used when a stop command is issued
// to the DCMotor.
func (m *DCMotor) StopAction() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+stopAction, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read stop command: %v", err)
	}
	return string(chomp(b)), err
}

// SetStopAction sets the stop action to be used when a stop command is
// issued to the DCMotor.
func (m *DCMotor) SetStopAction(comm string) error {
	avail, err := m.StopActions()
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
		return fmt.Errorf("ev3dev: stop command %q not available for %s (available:%q)", comm, m, avail)
	}
	err = m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+stopAction, m), comm)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set dc-motor stop command: %v", err)
	}
	return nil
}

// StopActions returns the available stop actions for the DCMotor.
func (m *DCMotor) StopActions() ([]string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+stopActions, m))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read dc-motor stop command: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}

// TimeSetpoint returns the current time set point value for the DCMotor.
func (m *DCMotor) TimeSetpoint() (time.Duration, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+timeSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read time set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse time set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetTimeSetpoint sets the time set point value for the DCMotor.
func (m *DCMotor) SetTimeSetpoint(d time.Duration) error {
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+timeSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set time set point: %v", err)
	}
	return nil
}
