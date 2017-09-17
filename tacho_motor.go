// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"math"
	"path/filepath"
	"strconv"
	"time"
)

var _ idSetter = (*TachoMotor)(nil)

// TachoMotor represents a handle to a tacho-motor.
type TachoMotor struct {
	id int

	err error
}

// Path returns the tacho-motor sysfs path.
func (*TachoMotor) Path() string { return filepath.Join(prefix, TachoMotorPath) }

// Type returns "motor".
func (*TachoMotor) Type() string { return motorPrefix }

// String satisfies the fmt.Stringer interface.
func (m *TachoMotor) String() string {
	if m == nil {
		return motorPrefix + "*"
	}
	return motorPrefix + strconv.Itoa(m.id)
}

// Err returns the error state of the TachoMotor and clears it.
func (m *TachoMotor) Err() error {
	err := m.err
	m.err = nil
	return err
}

// idInt and setID satisfy the idSetter interface.
func (m *TachoMotor) setID(id int) { *m = TachoMotor{id: id} }
func (m *TachoMotor) idInt() int {
	if m == nil {
		return -1
	}
	return m.id
}

// TachoMotorFor returns a TachoMotor for the given ev3 port name and driver. If the
// motor driver does not match the driver string, a TechoMotor for the port is
// returned with a DriverMismatch error.
// If port is empty, the first tacho-motor satisfying the driver name is returned.
func TachoMotorFor(port, driver string) (*TachoMotor, error) {
	id, err := deviceIDFor(port, driver, (*TachoMotor)(nil), -1)
	if id == -1 {
		return nil, err
	}
	return &TachoMotor{id: id}, err
}

// Next returns a TachoMotor for the next motor with the same device driver as
// the receiver.
func (m *TachoMotor) Next() (*TachoMotor, error) {
	driver, err := DriverFor(m)
	if err != nil {
		return nil, err
	}
	id, err := deviceIDFor("", driver, (*TachoMotor)(nil), m.id)
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
		m.err = newInvalidValueError(m, command, "", comm, avail)
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
	if sp < -100 || 100 < sp {
		m.err = newValueOutOfRangeError(m, dutyCycleSetpoint, sp, -100, 100)
		return m
	}
	m.err = setAttributeOf(m, dutyCycleSetpoint, strconv.Itoa(sp))
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
		m.err = newInvalidValueError(m, polarity, "", string(p), []string{string(Normal), string(Inversed)})
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
		m.err = newValueOutOfRangeError(m, position, pos, math.MinInt32, math.MaxInt32)
		return m
	}
	m.err = setAttributeOf(m, position, strconv.Itoa(pos))
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
	m.err = setAttributeOf(m, holdPIDkd, strconv.Itoa(k))
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
	m.err = setAttributeOf(m, holdPIDki, strconv.Itoa(k))
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
	m.err = setAttributeOf(m, holdPIDkp, strconv.Itoa(k))
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
		m.err = newValueOutOfRangeError(m, positionSetpoint, sp, math.MinInt32, math.MaxInt32)
		return m
	}
	m.err = setAttributeOf(m, positionSetpoint, strconv.Itoa(sp))
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
	m.err = setAttributeOf(m, speedSetpoint, strconv.Itoa(sp))
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
		m.err = newNegativeDurationError(m, rampUpSetpoint, sp)
		return m
	}
	m.err = setAttributeOf(m, rampUpSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
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
		m.err = newNegativeDurationError(m, rampDownSetpoint, sp)
		return m
	}
	m.err = setAttributeOf(m, rampDownSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
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
	m.err = setAttributeOf(m, speedPIDkd, strconv.Itoa(k))
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
	m.err = setAttributeOf(m, speedPIDki, strconv.Itoa(k))
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
	m.err = setAttributeOf(m, speedPIDkp, strconv.Itoa(k))
	return m
}

// State returns the current state of the TachoMotor.
func (m *TachoMotor) State() (MotorState, error) {
	if m.err != nil {
		return 0, m.Err()
	}
	return stateFrom(attributeOf(m, state))
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
		m.err = newInvalidValueError(m, stopAction, "", action, avail)
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
	if sp < 0 {
		m.err = newNegativeDurationError(m, timeSetpoint, sp)
		return m
	}
	m.err = setAttributeOf(m, timeSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
	return m
}

// Uevent returns the current uevent state for the TachoMotor.
func (m *TachoMotor) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(m, uevent))
}
