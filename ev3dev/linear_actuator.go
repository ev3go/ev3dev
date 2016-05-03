// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"strings"
	"time"
)

var _ idSetter = (*LinearActuator)(nil)

// LinearActuator represents a handle to a linear actuator tacho-motor.
type LinearActuator struct {
	id int

	err error
}

// Path returns the tacho-motor sysfs path.
func (*LinearActuator) Path() string { return TachoMotorPath }

// Type returns "linear".
func (*LinearActuator) Type() string { return linearPrefix }

// String satisfies the fmt.Stringer interface.
func (m *LinearActuator) String() string {
	if m == nil {
		return linearPrefix + "*"
	}
	return fmt.Sprint(linearPrefix, m.id)
}

// Err returns the error state of the LinearActuator and clears it.
func (m *LinearActuator) Err() error {
	err := m.err
	m.err = nil
	return err
}

// setID satisfies the idSetter interface.
func (m *LinearActuator) setID(id int) {
	*m = LinearActuator{id: id}
}

// LinearActuatorFor returns a LinearActuator for the given ev3 port name and driver.
// If the motor driver does not match the driver string, a LinearActuator for the port
// is returned with a DriverMismatch error.
// If port is empty, the first tacho-motor satisfying the driver name is returned.
func LinearActuatorFor(port, driver string) (*LinearActuator, error) {
	id, err := deviceIDFor(port, driver, (*LinearActuator)(nil))
	if id == -1 {
		return nil, err
	}
	return &LinearActuator{id: id}, err
}

// Commands returns the available commands for the LinearActuator.
func (m *LinearActuator) Commands() ([]string, error) {
	return stringSliceFrom(attributeOf(m, commands))
}

// Command issues a command to the LinearActuator.
func (m *LinearActuator) Command(comm string) *LinearActuator {
	if m.err != nil {
		return m
	}
	avail, err := m.Commands()
	if err != nil {
		m.err = err
		return m
	}
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
	m.err = setAttributeOf(m, command, comm)
	return m
}

// CountPerMeter returns the number of tacho counts in one meter of travel of the motor.
// Calls to CountPerMeter will return an error for non-linear motors.
func (m *LinearActuator) CountPerMeter() (int, error) {
	return intFrom(attributeOf(m, countPerMeter))
}

// FullTravelCount returns the the number of tacho counts in the full travel of the motor.
// Calls to FullTravelCount will return an error for non-linear motors.
func (m *LinearActuator) FullTravelCount() (int, error) {
	return intFrom(attributeOf(m, fullTravelCount))
}

// DutyCycle returns the current duty cycle value for the LinearActuator.
func (m *LinearActuator) DutyCycle() (int, error) {
	return intFrom(attributeOf(m, dutyCycle))
}

// DutyCycleSetpoint returns the current duty cycle setpoint value for the LinearActuator.
func (m *LinearActuator) DutyCycleSetpoint() (int, error) {
	return intFrom(attributeOf(m, dutyCycleSetpoint))
}

// SetDutyCycleSetpoint sets the duty cycle setpoint value for the LinearActuator
func (m *LinearActuator) SetDutyCycleSetpoint(sp int) *LinearActuator {
	if m.err != nil {
		return m
	}
	if sp < -100 || sp > 100 {
		m.err = fmt.Errorf("ev3dev: invalid duty cycle setpoint: %d (valid -100 - 100)", sp)
		return m
	}
	m.err = setAttributeOf(m, dutyCycleSetpoint, fmt.Sprint(sp))
	return m
}

// Polarity returns the current polarity of the LinearActuator.
func (m *LinearActuator) Polarity() (Polarity, error) {
	p, err := stringFrom(attributeOf(m, polarity))
	return Polarity(p), err
}

// SetPolarity sets the polarity of the LinearActuator
func (m *LinearActuator) SetPolarity(p Polarity) *LinearActuator {
	if m.err != nil {
		return m
	}
	if p != Normal && p != Inversed {
		m.err = fmt.Errorf("ev3dev: invalid polarity: %q (valid \"normal\" or \"inversed\")", p)
		return m
	}
	m.err = setAttributeOf(m, polarity, string(p))
	return m
}

// Position returns the current position value for the LinearActuator.
func (m *LinearActuator) Position() (int, error) {
	return intFrom(attributeOf(m, position))
}

// SetPosition sets the position value for the LinearActuator.
func (m *LinearActuator) SetPosition(pos int) *LinearActuator {
	if m.err != nil {
		return m
	}
	if pos != int(int32(pos)) {
		m.err = fmt.Errorf("ev3dev: invalid position: %d (valid in int32)", pos)
		return m
	}
	m.err = setAttributeOf(m, position, fmt.Sprint(pos))
	return m
}

// HoldPIDKd returns the derivative constant for the position PID for the LinearActuator.
func (m *LinearActuator) HoldPIDKd() (int, error) {
	return intFrom(attributeOf(m, holdPIDkd))
}

// SetHoldPIDKd sets the derivative constant for the position PID for the LinearActuator.
func (m *LinearActuator) SetHoldPIDKd(k int) *LinearActuator {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, holdPIDkd, fmt.Sprint(k))
	return m
}

// HoldPIDKi returns the integral constant for the position PID for the LinearActuator.
func (m *LinearActuator) HoldPIDKi() (int, error) {
	return intFrom(attributeOf(m, holdPIDki))
}

// SetHoldPIDKi sets the integral constant for the position PID for the LinearActuator.
func (m *LinearActuator) SetHoldPIDKi(k int) *LinearActuator {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, holdPIDki, fmt.Sprint(k))
	return m
}

// HoldPIDKp returns the proportional constant for the position PID for the LinearActuator.
func (m *LinearActuator) HoldPIDKp() (int, error) {
	return intFrom(attributeOf(m, holdPIDkp))
}

// SetHoldPIDKp sets the proportional constant for the position PID for the LinearActuator.
func (m *LinearActuator) SetHoldPIDKp(k int) *LinearActuator {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, holdPIDkp, fmt.Sprint(k))
	return m
}

// MaxSpeed returns the maximum value that is accepted by SpeedSetpoint.
func (m *LinearActuator) MaxSpeed() (int, error) {
	return intFrom(attributeOf(m, maxSpeed))
}

// PositionSetpoint returns the current position setpoint value for the LinearActuator.
func (m *LinearActuator) PositionSetpoint() (int, error) {
	return intFrom(attributeOf(m, positionSetpoint))
}

// SetPositionSetpoint sets the position setpoint value for the LinearActuator.
func (m *LinearActuator) SetPositionSetpoint(sp int) *LinearActuator {
	if m.err != nil {
		return m
	}
	if sp != int(int32(sp)) {
		m.err = fmt.Errorf("ev3dev: invalid position setpoint: %d (valid in int32)", sp)
		return m
	}
	m.err = setAttributeOf(m, positionSetpoint, fmt.Sprint(sp))
	return m
}

// Speed returns the current speed of the LinearActuator.
func (m *LinearActuator) Speed() (int, error) {
	return intFrom(attributeOf(m, speed))
}

// SpeedSetpoint returns the current speed setpoint value for the LinearActuator.
func (m *LinearActuator) SpeedSetpoint() (int, error) {
	return intFrom(attributeOf(m, speedSetpoint))
}

// SetSpeedSetpoint sets the speed setpoint value for the LinearActuator.
func (m *LinearActuator) SetSpeedSetpoint(sp int) *LinearActuator {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, speedSetpoint, fmt.Sprint(sp))
	return m
}

// RampUpSetpoint returns the current ramp up setpoint value for the LinearActuator.
func (m *LinearActuator) RampUpSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rampUpSetpoint))
}

// SetRampUpSetpoint sets the ramp up setpoint value for the LinearActuator.
func (m *LinearActuator) SetRampUpSetpoint(sp time.Duration) *LinearActuator {
	if m.err != nil {
		return m
	}
	if sp < 0 {
		m.err = fmt.Errorf("ev3dev: invalid ramp up setpoint: %v (must be positive)", sp)
		return m
	}
	m.err = setAttributeOf(m, rampUpSetpoint, fmt.Sprint(int(sp/time.Millisecond)))
	return m
}

// RampDownSetpoint returns the current ramp down setpoint value for the LinearActuator.
func (m *LinearActuator) RampDownSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rampDownSetpoint))
}

// SetRampDownSetpoint sets the ramp down setpoint value for the LinearActuator.
func (m *LinearActuator) SetRampDownSetpoint(sp time.Duration) *LinearActuator {
	if m.err != nil {
		return m
	}
	if sp < 0 {
		m.err = fmt.Errorf("ev3dev: invalid ramp down setpoint: %v (must be positive)", sp)
		return m
	}
	m.err = setAttributeOf(m, rampDownSetpoint, fmt.Sprint(int(sp/time.Millisecond)))
	return m
}

// SpeedPIDKd returns the derivative constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SpeedPIDKd() (int, error) {
	return intFrom(attributeOf(m, speedPIDkd))
}

// SetSpeedPIDKd sets the derivative constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SetSpeedPIDKd(sp int) *LinearActuator {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, speedPIDkd, fmt.Sprint(sp))
	return m
}

// SpeedPIDKi returns the integral constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SpeedPIDKi() (int, error) {
	return intFrom(attributeOf(m, speedPIDki))
}

// SetSpeedPIDKi sets the integral constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SetSpeedPIDKi(sp int) *LinearActuator {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, speedPIDki, fmt.Sprint(sp))
	return m
}

// SpeedPIDKp returns the proportional constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SpeedPIDKp() (int, error) {
	return intFrom(attributeOf(m, speedPIDkp))
}

// SetSpeedPIDKp sets the proportional constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SetSpeedPIDKp(sp int) *LinearActuator {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, speedPIDkp, fmt.Sprint(sp))
	return m
}

// State returns the current state of the LinearActuator.
func (m *LinearActuator) State() (MotorState, error) {
	if m.err != nil {
		return 0, m.Err()
	}
	data, _, err := attributeOf(m, state)
	if err != nil {
		return 0, err
	}
	var stat MotorState
	for _, s := range strings.Split(data, " ") {
		bit, ok := motorStateTable[s]
		if !ok {
			return 0, fmt.Errorf("ev3dev: unrecognized motor state value: %s in [%s]", s, data)
		}
		stat |= bit
	}
	return stat, nil
}

// StopAction returns the stop action used when a stop command is issued
// to the LinearActuator.
func (m *LinearActuator) StopAction() (string, error) {
	return stringFrom(attributeOf(m, stopAction))
}

// SetStopAction sets the stop action to be used when a stop command is
// issued to the LinearActuator.
func (m *LinearActuator) SetStopAction(action string) *LinearActuator {
	if m.err != nil {
		return m
	}
	avail, err := m.StopActions()
	if err != nil {
		m.err = err
		return m
	}
	ok := false
	for _, a := range avail {
		if a == action {
			ok = true
			break
		}
	}
	if !ok {
		m.err = fmt.Errorf("ev3dev: stop action %q not available for %s (available:%q)", action, m, avail)
		return m
	}
	m.err = setAttributeOf(m, stopAction, action)
	return m
}

// StopActions returns the available stop actions for the LinearActuator.
func (m *LinearActuator) StopActions() ([]string, error) {
	return stringSliceFrom(attributeOf(m, stopActions))
}

// TimeSetpoint returns the current time setpoint value for the LinearActuator.
func (m *LinearActuator) TimeSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, timeSetpoint))
}

// SetTimeSetpoint sets the time setpoint value for the LinearActuator.
func (m *LinearActuator) SetTimeSetpoint(sp time.Duration) *LinearActuator {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, timeSetpoint, fmt.Sprint(int(sp/time.Millisecond)))
	return m
}

// Uevent returns the current uevent state for the LinearActuator.
func (m *LinearActuator) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(m, uevent))
}
