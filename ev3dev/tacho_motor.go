// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"strings"
	"time"
)

var _ idSetter = (*TachoMotor)(nil)

// TachoMotor represents a handle to a tacho-motor.
type TachoMotor struct {
	id int

	err error
}

// Path returns the tacho-motor sysfs path.
func (*TachoMotor) Path() string { return TachoMotorPath }

// Type returns "motor".
func (*TachoMotor) Type() string { return motorPrefix }

// String satisfies the fmt.Stringer interface.
func (m *TachoMotor) String() string {
	if m == nil {
		return motorPrefix + "*"
	}
	return fmt.Sprint(motorPrefix, m.id)
}

// Err returns the error state of the TachoMotor and clears it.
func (m *TachoMotor) Err() error {
	err := m.err
	m.err = nil
	return err
}

// setID satisfies the idSetter interface.
func (m *TachoMotor) setID(id int) {
	*m = TachoMotor{id: id}
}

// TachoMotorFor returns a TachoMotor for the given ev3 port name and driver. If the
// motor driver does not match the driver string, a TechoMotor for the port is
// returned with a DriverMismatch error.
// If port is empty, the first tacho-motor satisfying the driver name is returned.
func TachoMotorFor(port, driver string) (*TachoMotor, error) {
	id, err := deviceIDFor(port, driver, (*TachoMotor)(nil))
	if id == -1 {
		return nil, err
	}
	return &TachoMotor{id: id}, err
}

// Commands returns the available commands for the TachoMotor.
func (m *TachoMotor) Commands() ([]string, error) {
	return stringSliceFrom(attributeOf(m, commands))
}

// Command issues a command to the TachoMotor.
func (m *TachoMotor) Command(comm string) *TachoMotor {
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

// CountPerRot returns the number of tacho counts in one rotation of the motor.
// Calls to CountPerRot will return an error for non-rotational motors.
func (m *TachoMotor) CountPerRot() (int, error) {
	return intFrom(attributeOf(m, countPerRot))
}

// DutyCycle returns the current duty cycle value for the TachoMotor.
func (m *TachoMotor) DutyCycle() (int, error) {
	return intFrom(attributeOf(m, dutyCycle))
}

// DutyCycleSetpoint returns the current duty cycle setpoint value for the TachoMotor.
func (m *TachoMotor) DutyCycleSetpoint() (int, error) {
	return intFrom(attributeOf(m, dutyCycleSetpoint))
}

// SetDutyCycleSetpoint sets the duty cycle setpoint value for the TachoMotor
func (m *TachoMotor) SetDutyCycleSetpoint(sp int) *TachoMotor {
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

// Polarity returns the current polarity of the TachoMotor.
func (m *TachoMotor) Polarity() (Polarity, error) {
	p, err := stringFrom(attributeOf(m, polarity))
	return Polarity(p), err
}

// SetPolarity sets the polarity of the TachoMotor
func (m *TachoMotor) SetPolarity(p Polarity) *TachoMotor {
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

// Position returns the current position value for the TachoMotor.
func (m *TachoMotor) Position() (int, error) {
	return intFrom(attributeOf(m, position))
}

// SetPosition sets the position value for the TachoMotor.
func (m *TachoMotor) SetPosition(pos int) *TachoMotor {
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

// HoldPIDKd returns the derivative constant for the position PID for the TachoMotor.
func (m *TachoMotor) HoldPIDKd() (int, error) {
	return intFrom(attributeOf(m, holdPIDkd))
}

// SetHoldPIDKd sets the derivative constant for the position PID for the TachoMotor.
func (m *TachoMotor) SetHoldPIDKd(k int) *TachoMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, holdPIDkd, fmt.Sprint(k))
	return m
}

// HoldPIDKi returns the integral constant for the position PID for the TachoMotor.
func (m *TachoMotor) HoldPIDKi() (int, error) {
	return intFrom(attributeOf(m, holdPIDki))
}

// SetHoldPIDKi sets the integral constant for the position PID for the TachoMotor.
func (m *TachoMotor) SetHoldPIDKi(k int) *TachoMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, holdPIDki, fmt.Sprint(k))
	return m
}

// HoldPIDKp returns the proportional constant for the position PID for the TachoMotor.
func (m *TachoMotor) HoldPIDKp() (int, error) {
	return intFrom(attributeOf(m, holdPIDkp))
}

// SetHoldPIDKp sets the proportional constant for the position PID for the TachoMotor.
func (m *TachoMotor) SetHoldPIDKp(k int) *TachoMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, holdPIDkp, fmt.Sprint(k))
	return m
}

// MaxSpeed returns the maximum value that is accepted by SpeedSetpoint.
func (m *TachoMotor) MaxSpeed() (int, error) {
	return intFrom(attributeOf(m, maxSpeed))
}

// PositionSetpoint returns the current position setpoint value for the TachoMotor.
func (m *TachoMotor) PositionSetpoint() (int, error) {
	return intFrom(attributeOf(m, positionSetpoint))
}

// SetPositionSetpoint sets the position setpoint value for the TachoMotor.
func (m *TachoMotor) SetPositionSetpoint(sp int) *TachoMotor {
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

// Speed returns the current speed of the TachoMotor.
func (m *TachoMotor) Speed() (int, error) {
	return intFrom(attributeOf(m, speed))
}

// SpeedSetpoint returns the current speed setpoint value for the TachoMotor.
func (m *TachoMotor) SpeedSetpoint() (int, error) {
	return intFrom(attributeOf(m, speedSetpoint))
}

// SetSpeedSetpoint sets the speed setpoint value for the TachoMotor.
func (m *TachoMotor) SetSpeedSetpoint(sp int) *TachoMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, speedSetpoint, fmt.Sprint(sp))
	return m
}

// RampUpSetpoint returns the current ramp up setpoint value for the TachoMotor.
func (m *TachoMotor) RampUpSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rampUpSetpoint))
}

// SetRampUpSetpoint sets the ramp up setpoint value for the TachoMotor.
func (m *TachoMotor) SetRampUpSetpoint(sp time.Duration) *TachoMotor {
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

// RampDownSetpoint returns the current ramp down setpoint value for the TachoMotor.
func (m *TachoMotor) RampDownSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rampDownSetpoint))
}

// SetRampDownSetpoint sets the ramp down setpoint value for the TachoMotor.
func (m *TachoMotor) SetRampDownSetpoint(sp time.Duration) *TachoMotor {
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

// SpeedPIDKd returns the derivative constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SpeedPIDKd() (int, error) {
	return intFrom(attributeOf(m, speedPIDkd))
}

// SetSpeedPIDKd sets the derivative constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SetSpeedPIDKd(k int) *TachoMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, speedPIDkd, fmt.Sprint(k))
	return m
}

// SpeedPIDKi returns the integral constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SpeedPIDKi() (int, error) {
	return intFrom(attributeOf(m, speedPIDki))
}

// SetSpeedPIDKi sets the integral constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SetSpeedPIDKi(k int) *TachoMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, speedPIDki, fmt.Sprint(k))
	return m
}

// SpeedPIDKp returns the proportional constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SpeedPIDKp() (int, error) {
	return intFrom(attributeOf(m, speedPIDkp))
}

// SetSpeedPIDKp sets the proportional constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SetSpeedPIDKp(k int) *TachoMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, speedPIDkp, fmt.Sprint(k))
	return m
}

// State returns the current state of the TachoMotor.
func (m *TachoMotor) State() (MotorState, error) {
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
// to the TachoMotor.
func (m *TachoMotor) StopAction() (string, error) {
	return stringFrom(attributeOf(m, stopAction))
}

// SetStopAction sets the stop action to be used when a stop command is
// issued to the TachoMotor.
func (m *TachoMotor) SetStopAction(action string) *TachoMotor {
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

// StopActions returns the available stop actions for the TachoMotor.
func (m *TachoMotor) StopActions() ([]string, error) {
	return stringSliceFrom(attributeOf(m, stopActions))
}

// TimeSetpoint returns the current time setpoint value for the TachoMotor.
func (m *TachoMotor) TimeSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, timeSetpoint))
}

// SetTimeSetpoint sets the time setpoint value for the TachoMotor.
func (m *TachoMotor) SetTimeSetpoint(sp time.Duration) *TachoMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, timeSetpoint, fmt.Sprint(int(sp/time.Millisecond)))
	return m
}

// Uevent returns the current uevent state for the TachoMotor.
func (m *TachoMotor) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(m, uevent))
}
