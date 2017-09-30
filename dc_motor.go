// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"path/filepath"
	"strconv"
	"time"
)

var _ idSetter = (*DCMotor)(nil)

// DCMotor represents a handle to a dc-motor.
type DCMotor struct {
	id int

	// Cached values:
	driver                string
	commands, stopActions []string

	err error
}

// Path returns the dc-motor sysfs path.
func (*DCMotor) Path() string { return filepath.Join(prefix, DCMotorPath) }

// Type returns "motor".
func (*DCMotor) Type() string { return motorPrefix }

// String satisfies the fmt.Stringer interface.
func (m *DCMotor) String() string {
	if m == nil {
		return motorPrefix + "*"
	}
	return motorPrefix + strconv.Itoa(m.id)
}

// Err returns the error state of the DCMotor and clears it.
func (m *DCMotor) Err() error {
	err := m.err
	m.err = nil
	return err
}

// idInt and setID satisfy the idSetter interface.
func (m *DCMotor) setID(id int) error {
	t := DCMotor{id: id}
	var err error
	t.commands, err = stringSliceFrom(attributeOf(&t, commands))
	if err != nil {
		goto fail
	}
	t.stopActions, err = stringSliceFrom(attributeOf(&t, stopActions))
	if err != nil {
		goto fail
	}
	t.driver, err = DriverFor(&t)
	if err != nil {
		goto fail
	}
	*m = t
	return nil

fail:
	*m = DCMotor{id: -1}
	return err
}
func (m *DCMotor) idInt() int {
	if m == nil {
		return -1
	}
	return m.id
}

// DCMotorFor returns a DCMotor for the given ev3 port name and driver. If the
// motor driver does not match the driver string, a DCMotor for the port is
// returned with a DriverMismatch error.
// If port is empty, the first dc-motor satisfying the driver name is returned.
func DCMotorFor(port, driver string) (*DCMotor, error) {
	id, err := deviceIDFor(port, driver, (*DCMotor)(nil), -1)
	if id == -1 {
		return nil, err
	}
	var m DCMotor
	_err := m.setID(id)
	if _err != nil {
		err = _err
	}
	return &m, err
}

// Next returns a DCMotor for the next motor with the same device driver as
// the receiver.
func (m *DCMotor) Next() (*DCMotor, error) {
	driver, err := DriverFor(m)
	if err != nil {
		return nil, err
	}
	id, err := deviceIDFor("", driver, (*DCMotor)(nil), m.id)
	if id == -1 {
		return nil, err
	}
	return &DCMotor{id: id}, err
}

// Driver returns the driver used by the DCMotor.
func (m *DCMotor) Driver() string {
	return m.driver
}

// Commands returns the available commands for the DCMotor.
func (m *DCMotor) Commands() []string {
	if m.commands == nil {
		return nil
	}
	// Return a copy to prevent users
	// changing the values under our feet.
	avail := make([]string, len(m.commands))
	copy(avail, m.commands)
	return avail
}

// Command issues a command to the DCMotor.
func (m *DCMotor) Command(comm string) *DCMotor {
	if m.err != nil {
		return m
	}
	ok := false
	for _, c := range m.commands {
		if c == comm {
			ok = true
			break
		}
	}
	if !ok {
		m.err = newInvalidValueError(m, command, "", comm, m.Commands())
		return m
	}
	m.err = setAttributeOf(m, command, comm)
	return m
}

// DutyCycle returns the current duty cycle value for the DCMotor.
func (m *DCMotor) DutyCycle() (int, error) {
	return intFrom(attributeOf(m, dutyCycle))
}

// DutyCycleSetpoint returns the current duty cycle setpoint value for the DCMotor.
func (m *DCMotor) DutyCycleSetpoint() (int, error) {
	return intFrom(attributeOf(m, dutyCycleSetpoint))
}

// SetDutyCycleSetpoint sets the duty cycle setpoint value for the DCMotor
func (m *DCMotor) SetDutyCycleSetpoint(sp int) *DCMotor {
	if m.err != nil {
		return m
	}
	if sp < -100 || 100 < sp {
		m.err = newValueOutOfRangeError(m, dutyCycleSetpoint, sp, -100, 100)
		return m
	}
	m.err = setAttributeOf(m, dutyCycleSetpoint, strconv.Itoa(sp))
	return m
}

// Polarity returns the current polarity of the DCMotor.
func (m *DCMotor) Polarity() (Polarity, error) {
	p, err := stringFrom(attributeOf(m, polarity))
	return Polarity(p), err
}

// SetPolarity sets the polarity of the DCMotor
func (m *DCMotor) SetPolarity(p Polarity) *DCMotor {
	if m.err != nil {
		return m
	}
	if p != Normal && p != Inversed {
		m.err = newInvalidValueError(m, polarity, "", string(p), []string{string(Normal), string(Inversed)})
		return m
	}
	m.err = setAttributeOf(m, polarity, string(p))
	return m
}

// RampUpSetpoint returns the current ramp up setpoint value for the DCMotor.
func (m *DCMotor) RampUpSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rampUpSetpoint))
}

// SetRampUpSetpoint sets the ramp up setpoint value for the DCMotor.
func (m *DCMotor) SetRampUpSetpoint(sp time.Duration) *DCMotor {
	if m.err != nil {
		return m
	}
	if sp < 0 || 10*time.Second < sp {
		m.err = newDurationOutOfRangeError(m, rampUpSetpoint, sp, 0, 10*time.Second)
		return m
	}
	m.err = setAttributeOf(m, rampUpSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
	return m
}

// RampDownSetpoint returns the current ramp down setpoint value for the DCMotor.
func (m *DCMotor) RampDownSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rampDownSetpoint))
}

// SetRampDownSetpoint sets the ramp down setpoint value for the DCMotor.
func (m *DCMotor) SetRampDownSetpoint(sp time.Duration) *DCMotor {
	if m.err != nil {
		return m
	}
	if sp < 0 || 10*time.Second < sp {
		m.err = newDurationOutOfRangeError(m, rampDownSetpoint, sp, 0, 10*time.Second)
		return m
	}
	m.err = setAttributeOf(m, rampDownSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
	return m
}

// State returns the current state of the DCMotor.
func (m *DCMotor) State() (MotorState, error) {
	if m.err != nil {
		return 0, m.Err()
	}
	return stateFrom(attributeOf(m, state))
}

// StopAction returns the stop action used when a stop command is issued
// to the DCMotor.
func (m *DCMotor) StopAction() (string, error) {
	return stringFrom(attributeOf(m, stopAction))
}

// SetStopAction sets the stop action to be used when a stop command is
// issued to the DCMotor.
func (m *DCMotor) SetStopAction(action string) *DCMotor {
	if m.err != nil {
		return m
	}
	ok := false
	for _, a := range m.stopActions {
		if a == action {
			ok = true
			break
		}
	}
	if !ok {
		m.err = newInvalidValueError(m, stopAction, "", action, m.StopActions())
		return m
	}
	m.err = setAttributeOf(m, stopAction, action)
	return m
}

// StopActions returns the available stop actions for the DCMotor.
func (m *DCMotor) StopActions() []string {
	if m.stopActions == nil {
		return nil
	}
	// Return a copy to prevent users
	// changing the values under our feet.
	avail := make([]string, len(m.stopActions))
	copy(avail, m.stopActions)
	return avail
}

// TimeSetpoint returns the current time setpoint value for the DCMotor.
func (m *DCMotor) TimeSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, timeSetpoint))
}

// SetTimeSetpoint sets the time setpoint value for the DCMotor.
func (m *DCMotor) SetTimeSetpoint(sp time.Duration) *DCMotor {
	if m.err != nil {
		return m
	}
	if sp < 0 {
		m.err = newNegativeDurationError(m, timeSetpoint, sp)
		return m
	}
	m.err = setAttributeOf(m, timeSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
	return m
}

// Uevent returns the current uevent state for the DCMotor.
func (m *DCMotor) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(m, uevent))
}
