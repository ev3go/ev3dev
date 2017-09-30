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

var _ idSetter = (*LinearActuator)(nil)

// LinearActuator represents a handle to a linear actuator tacho-motor.
type LinearActuator struct {
	id int

	// Cached values:
	driver                                   string
	countPerMeter, fullTravelCount, maxSpeed int
	commands, stopActions                    []string

	err error
}

// Path returns the tacho-motor sysfs path.
func (*LinearActuator) Path() string { return filepath.Join(prefix, TachoMotorPath) }

// Type returns "linear".
func (*LinearActuator) Type() string { return linearPrefix }

// String satisfies the fmt.Stringer interface.
func (m *LinearActuator) String() string {
	if m == nil {
		return linearPrefix + "*"
	}
	return linearPrefix + strconv.Itoa(m.id)
}

// Err returns the error state of the LinearActuator and clears it.
func (m *LinearActuator) Err() error {
	err := m.err
	m.err = nil
	return err
}

// idInt and setID satisfy the idSetter interface.
func (m *LinearActuator) setID(id int) error {
	t := LinearActuator{id: id}
	var err error
	t.countPerMeter, err = intFrom(attributeOf(&t, countPerMeter))
	if err != nil {
		goto fail
	}
	t.fullTravelCount, err = intFrom(attributeOf(&t, fullTravelCount))
	if err != nil {
		goto fail
	}
	t.maxSpeed, err = intFrom(attributeOf(&t, maxSpeed))
	if err != nil {
		goto fail
	}
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
	*m = LinearActuator{id: -1}
	return err
}
func (m *LinearActuator) idInt() int {
	if m == nil {
		return -1
	}
	return m.id
}

// LinearActuatorFor returns a LinearActuator for the given ev3 port name and driver.
// If the motor driver does not match the driver string, a LinearActuator for the port
// is returned with a DriverMismatch error.
// If port is empty, the first tacho-motor satisfying the driver name is returned.
func LinearActuatorFor(port, driver string) (*LinearActuator, error) {
	id, err := deviceIDFor(port, driver, (*LinearActuator)(nil), -1)
	if id == -1 {
		return nil, err
	}
	var m LinearActuator
	_err := m.setID(id)
	if _err != nil {
		err = _err
	}
	return &m, err
}

// Next returns a LinearActuator for the next motor with the same device driver as
// the receiver.
func (m *LinearActuator) Next() (*LinearActuator, error) {
	driver, err := DriverFor(m)
	if err != nil {
		return nil, err
	}
	id, err := deviceIDFor("", driver, (*LinearActuator)(nil), m.id)
	if id == -1 {
		return nil, err
	}
	return &LinearActuator{id: id}, err
}

// Driver returns the driver used by the LinearActuator.
func (m *LinearActuator) Driver() string {
	return m.driver
}

// Commands returns the available commands for the LinearActuator.
func (m *LinearActuator) Commands() []string {
	if m.commands == nil {
		return nil
	}
	// Return a copy to prevent users
	// changing the values under our feet.
	avail := make([]string, len(m.commands))
	copy(avail, m.commands)
	return avail
}

// Command issues a command to the LinearActuator.
func (m *LinearActuator) Command(comm string) *LinearActuator {
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

// CountPerMeter returns the number of tacho counts in one meter of travel of the motor.
func (m *LinearActuator) CountPerMeter() int {
	return m.countPerMeter
}

// FullTravelCount returns the the number of tacho counts in the full travel of the motor.
func (m *LinearActuator) FullTravelCount() int {
	return m.fullTravelCount
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
	if sp < -100 || 100 < sp {
		m.err = newValueOutOfRangeError(m, dutyCycleSetpoint, sp, -100, 100)
		return m
	}
	m.err = setAttributeOf(m, dutyCycleSetpoint, strconv.Itoa(sp))
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
		m.err = newInvalidValueError(m, polarity, "", string(p), []string{string(Normal), string(Inversed)})
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
		m.err = newValueOutOfRangeError(m, position, pos, math.MinInt32, math.MaxInt32)
		return m
	}
	m.err = setAttributeOf(m, position, strconv.Itoa(pos))
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
	m.err = setAttributeOf(m, holdPIDkd, strconv.Itoa(k))
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
	m.err = setAttributeOf(m, holdPIDki, strconv.Itoa(k))
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
	m.err = setAttributeOf(m, holdPIDkp, strconv.Itoa(k))
	return m
}

// MaxSpeed returns the maximum value that is accepted by SpeedSetpoint.
func (m *LinearActuator) MaxSpeed() int {
	return m.maxSpeed
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
		m.err = newValueOutOfRangeError(m, positionSetpoint, sp, math.MinInt32, math.MaxInt32)
		return m
	}
	m.err = setAttributeOf(m, positionSetpoint, strconv.Itoa(sp))
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
	m.err = setAttributeOf(m, speedSetpoint, strconv.Itoa(sp))
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
		m.err = newNegativeDurationError(m, rampUpSetpoint, sp)
		return m
	}
	m.err = setAttributeOf(m, rampUpSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
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
		m.err = newNegativeDurationError(m, rampDownSetpoint, sp)
		return m
	}
	m.err = setAttributeOf(m, rampDownSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
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
	m.err = setAttributeOf(m, speedPIDkd, strconv.Itoa(sp))
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
	m.err = setAttributeOf(m, speedPIDki, strconv.Itoa(sp))
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
	m.err = setAttributeOf(m, speedPIDkp, strconv.Itoa(sp))
	return m
}

// State returns the current state of the LinearActuator.
func (m *LinearActuator) State() (MotorState, error) {
	if m.err != nil {
		return 0, m.Err()
	}
	return stateFrom(attributeOf(m, state))
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

// StopActions returns the available stop actions for the LinearActuator.
func (m *LinearActuator) StopActions() []string {
	if m.stopActions == nil {
		return nil
	}
	// Return a copy to prevent users
	// changing the values under our feet.
	avail := make([]string, len(m.stopActions))
	copy(avail, m.stopActions)
	return avail

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
	if sp < 0 {
		m.err = newNegativeDurationError(m, timeSetpoint, sp)
		return m
	}
	m.err = setAttributeOf(m, timeSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
	return m
}

// Uevent returns the current uevent state for the LinearActuator.
func (m *LinearActuator) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(m, uevent))
}
