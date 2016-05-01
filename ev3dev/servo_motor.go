// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
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
	m.err = setAttributeOf(m, command, comm)
	return m
}

// MaxPulseSetpoint returns the current max pulse setpoint value for the ServoMotor.
func (m *ServoMotor) MaxPulseSetpoint() (int, error) {
	return intFrom(attributeOf(m, maxPulseSetpoint))
}

// SetMaxPulseSetpoint sets the max pulse setpoint value for the ServoMotor
func (m *ServoMotor) SetMaxPulseSetpoint(sp int) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 2300 || sp > 2700 {
		m.err = fmt.Errorf("ev3dev: invalid max pulse setpoint: %d (valid 2300-1700)", sp)
		return m
	}
	m.err = setAttributeOf(m, maxPulseSetpoint, string(sp))
	return m
}

// MidPulseSetpoint returns the current mid pulse setpoint value for the ServoMotor.
func (m *ServoMotor) MidPulseSetpoint() (int, error) {
	return intFrom(attributeOf(m, midPulseSetpoint))
}

// SetMidPulseSetpoint sets the mid pulse setpoint value for the ServoMotor
func (m *ServoMotor) SetMidPulseSetpoint(sp int) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 1300 || sp > 1700 {
		m.err = fmt.Errorf("ev3dev: invalid mid pulse setpoint: %d (valid 1300-1700)", sp)
		return m
	}
	m.err = setAttributeOf(m, midPulseSetpoint, string(sp))
	return m
}

// MinPulseSetpoint returns the current min pulse setpoint value for the ServoMotor.
func (m *ServoMotor) MinPulseSetpoint() (int, error) {
	return intFrom(attributeOf(m, minPulseSetpoint))
}

// SetMinPulseSetpoint sets the min pulse setpoint value for the ServoMotor
func (m *ServoMotor) SetMinPulseSetpoint(sp int) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 300 || sp > 700 {
		m.err = fmt.Errorf("ev3dev: invalid min pulse setpoint: %d (valid 300 - 700)", sp)
		return m
	}
	m.err = setAttributeOf(m, minPulseSetpoint, string(sp))
	return m
}

// Polarity returns the current polarity of the ServoMotor.
func (m *ServoMotor) Polarity() (Polarity, error) {
	p, err := stringFrom(attributeOf(m, polarity))
	return Polarity(p), err
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
	m.err = setAttributeOf(m, polarity, string(p))
	return m
}

// PositionSetpoint returns the current position setpoint value for
// the ServoMotor.
func (m *ServoMotor) PositionSetpoint() (int, error) {
	return intFrom(attributeOf(m, positionSetpoint))
}

// SetPositionSetpoint sets the position value for the ServoMotor.
func (m *ServoMotor) SetPositionSetpoint(sp int) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp != int(int32(sp)) {
		m.err = fmt.Errorf("ev3dev: invalid position: %d (valid in int32)", sp)
		return m
	}
	m.err = setAttributeOf(m, positionSetpoint, fmt.Sprint(sp))
	return m
}

// RateSetpoint returns the current rate setpoint value for the ServoMotor.
func (m *ServoMotor) RateSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rateSetpoint))
}

// SetRateSetpoint sets the rate setpoint value for the ServoMotor.
func (m *ServoMotor) SetRateSetpoint(sp time.Duration) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 0 {
		m.err = fmt.Errorf("ev3dev: invalid ramp up setpoint: %v (must be positive)", sp)
		return m
	}
	m.err = setAttributeOf(m, rateSetpoint, fmt.Sprint(int(sp/time.Millisecond)))
	return m
}

// State returns the current state of the ServoMotor.
func (m *ServoMotor) State() (MotorState, error) {
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
