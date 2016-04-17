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

// String satisfies the fmt.Stringer interface.
func (m *DCMotor) String() string { return fmt.Sprint(motorPrefix, m.id) }

// DCMotorFor returns a DCMotor for the given ev3 port name and driver. If the
// motor driver does not match the driver string, a DCMotor for the port is
// returned with a DriverMismatch error.
// If port is empty, the first dc-motor satisfying the driver name is returned.
func DCMotorFor(port, driver string) (*DCMotor, error) {
	id, err := deviceIDFor(port, driver, DCMotorPath, motorPrefix)
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

// Address returns the ev3 port name for the DCMotor.
func (m *DCMotor) Address() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+address, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port address: %v", err)
	}
	return string(chomp(b)), err
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

// Driver returns the driver name for the DCMotor.
func (m *DCMotor) Driver() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+driverName, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port driver name: %v", err)
	}
	return string(chomp(b)), err
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

// DutyCycleSetPoint returns the current duty cycle set point value for the DCMotor.
func (m *DCMotor) DutyCycleSetPoint() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+dutyCycleSetPoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read duty cycle set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse duty cycle set point: %v", err)
	}
	return sp, nil
}

// SetDutyCycleSetPoint sets the duty cycle set point value for the DCMotor
func (m *DCMotor) SetDutyCycleSetPoint(sp int) error {
	if sp < -100 || sp > 100 {
		return fmt.Errorf("ev3dev: invalid duty cycle set point: %d (valid -100 - 100)", sp)
	}
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+dutyCycleSetPoint, m), fmt.Sprintln(sp))
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

// RampUpSetPoint returns the current ramp up set point value for the DCMotor.
func (m *DCMotor) RampUpSetPoint() (time.Duration, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+rampUpSetPoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read ramp up set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse ramp up set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetRampUpSetPoint sets the ramp up set point value for the DCMotor.
func (m *DCMotor) SetRampUpSetPoint(d time.Duration) error {
	if d < 0 || d > 10000 {
		return fmt.Errorf("ev3dev: invalid ramp up set point: %v (must be positive)", d)
	}
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+rampUpSetPoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set ramp up set point: %v", err)
	}
	return nil
}

// RampDownSetPoint returns the current ramp down set point value for the DCMotor.
func (m *DCMotor) RampDownSetPoint() (time.Duration, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+rampDownSetPoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read ramp down set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse ramp down set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetRampDownSetPoint sets the ramp down set point value for the DCMotor.
func (m *DCMotor) SetRampDownSetPoint(d time.Duration) error {
	if d < 0 || d > 10000 {
		return fmt.Errorf("ev3dev: invalid ramp down set point: %v (must be positive)", d)
	}
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+rampDownSetPoint, m), fmt.Sprintln(int(d/time.Millisecond)))
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

// StopCommand returns the stop action used when a stop command is issued
// to the DCMotor.
func (m *DCMotor) StopCommand() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+stopCommand, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read stop command: %v", err)
	}
	return string(chomp(b)), err
}

// SetStopCommand sets the stop action to be used when a stop command is
// issued to the DCMotor.
func (m *DCMotor) SetStopCommand(comm string) error {
	avail, err := m.StopCommands()
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
	err = m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+stopCommand, m), comm)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set dc-motor stop command: %v", err)
	}
	return nil
}

// StopCommands returns the available stop actions for the DCMotor.
func (m *DCMotor) StopCommands() ([]string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+stopCommands, m))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read dc-motor stop command: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}

// TimeSetPoint returns the current time set point value for the DCMotor.
func (m *DCMotor) TimeSetPoint() (time.Duration, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(DCMotorPath+"/%s/"+timeSetPoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read time set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse time set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetTimeSetPoint sets the time set point value for the DCMotor.
func (m *DCMotor) SetTimeSetPoint(d time.Duration) error {
	err := m.writeFile(fmt.Sprintf(DCMotorPath+"/%s/"+timeSetPoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set time set point: %v", err)
	}
	return nil
}
