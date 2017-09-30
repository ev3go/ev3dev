// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"path/filepath"
	"strconv"
	"time"
)

var _ idSetter = (*ServoMotor)(nil)

// ServoMotor represents a handle to a servo-motor.
type ServoMotor struct {
	id int

	// Cached value:
	driver string

	err error
}

// Path returns the servo-motor sysfs path.
func (*ServoMotor) Path() string { return filepath.Join(prefix, ServoMotorPath) }

// Type returns "motor".
func (*ServoMotor) Type() string { return motorPrefix }

// String satisfies the fmt.Stringer interface.
func (m *ServoMotor) String() string {
	if m == nil {
		return motorPrefix + "*"
	}
	return motorPrefix + strconv.Itoa(m.id)
}

// Err returns the error state of the ServoMotor and clears it.
func (m *ServoMotor) Err() error {
	err := m.err
	m.err = nil
	return err
}

// idInt and setID satisfy the idSetter interface.
func (m *ServoMotor) setID(id int) error {
	t := ServoMotor{id: id}
	var err error
	t.driver, err = DriverFor(&t)
	if err != nil {
		*m = ServoMotor{id: -1}
		return err
	}
	*m = t
	return nil
}
func (m *ServoMotor) idInt() int {
	if m == nil {
		return -1
	}
	return m.id
}

// ServoMotorFor returns a ServoMotor for the given ev3 port name and driver.
// If the motor driver does not match the driver string, a ServoMotor for the port
// is returned with a DriverMismatch error.
// If port is empty, the first servo-motor satisfying the driver name is returned.
func ServoMotorFor(port, driver string) (*ServoMotor, error) {
	id, err := deviceIDFor(port, driver, (*ServoMotor)(nil), -1)
	if id == -1 {
		return nil, err
	}
	var m ServoMotor
	_err := m.setID(id)
	if _err != nil {
		err = _err
	}
	return &m, err
}

// Next returns a ServoMotor for the next motor with the same device driver as
// the receiver.
func (m *ServoMotor) Next() (*ServoMotor, error) {
	driver, err := DriverFor(m)
	if err != nil {
		return nil, err
	}
	id, err := deviceIDFor("", driver, (*ServoMotor)(nil), m.id)
	if id == -1 {
		return nil, err
	}
	return &ServoMotor{id: id}, err
}

// Driver returns the driver used by the ServoMotor.
func (p *ServoMotor) Driver() string {
	return p.driver
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
		m.err = newInvalidValueError(m, command, "", comm, avail)
		return m
	}
	m.err = setAttributeOf(m, command, comm)
	return m
}

// MaxPulseSetpoint returns the current max pulse setpoint value for the ServoMotor.
func (m *ServoMotor) MaxPulseSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, maxPulseSetpoint))
}

// SetMaxPulseSetpoint sets the max pulse setpoint value for the ServoMotor
func (m *ServoMotor) SetMaxPulseSetpoint(sp time.Duration) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 2300*time.Millisecond || 2700*time.Millisecond < sp {
		m.err = newDurationOutOfRangeError(m, maxPulseSetpoint, sp, 2300*time.Millisecond, 2700*time.Millisecond)
		return m
	}
	m.err = setAttributeOf(m, maxPulseSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
	return m
}

// MidPulseSetpoint returns the current mid pulse setpoint value for the ServoMotor.
func (m *ServoMotor) MidPulseSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, midPulseSetpoint))
}

// SetMidPulseSetpoint sets the mid pulse setpoint value for the ServoMotor
func (m *ServoMotor) SetMidPulseSetpoint(sp time.Duration) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 1300*time.Millisecond || 1700*time.Millisecond < sp {
		m.err = newDurationOutOfRangeError(m, midPulseSetpoint, sp, 1300*time.Millisecond, 1700*time.Millisecond)
		return m
	}
	m.err = setAttributeOf(m, midPulseSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
	return m
}

// MinPulseSetpoint returns the current min pulse setpoint value for the ServoMotor.
func (m *ServoMotor) MinPulseSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, minPulseSetpoint))
}

// SetMinPulseSetpoint sets the min pulse setpoint value for the ServoMotor
func (m *ServoMotor) SetMinPulseSetpoint(sp time.Duration) *ServoMotor {
	if m.err != nil {
		return m
	}
	if sp < 300*time.Millisecond || 700*time.Millisecond < sp {
		m.err = newDurationOutOfRangeError(m, minPulseSetpoint, sp, 300*time.Millisecond, 700*time.Millisecond)
		return m
	}
	m.err = setAttributeOf(m, minPulseSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
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
		m.err = newInvalidValueError(m, polarity, "", string(p), []string{string(Normal), string(Inversed)})
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
	if sp < -100 || 100 < sp {
		m.err = newValueOutOfRangeError(m, positionSetpoint, sp, -100, 100)
		return m
	}
	m.err = setAttributeOf(m, positionSetpoint, strconv.Itoa(sp))
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
		m.err = newNegativeDurationError(m, rateSetpoint, sp)
		return m
	}
	m.err = setAttributeOf(m, rateSetpoint, strconv.Itoa(int(sp/time.Millisecond)))
	return m
}

// State returns the current state of the ServoMotor.
func (m *ServoMotor) State() (MotorState, error) {
	if m.err != nil {
		return 0, m.Err()
	}
	return stateFrom(attributeOf(m, state))
}

// Uevent returns the current uevent state for the ServoMotor.
func (m *ServoMotor) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(m, uevent))
}
