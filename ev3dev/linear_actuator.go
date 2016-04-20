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

func (m *LinearActuator) writeFile(path, data string) error {
	return ioutil.WriteFile(path, []byte(data), 0)
}

// Commands returns the available commands for the LinearActuator.
func (m *LinearActuator) Commands() ([]string, error) {
	if m.err != nil {
		return nil, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+commands, m))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read tacho-motor commands: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
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
	err = m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+command, m), comm)
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to issue tacho-motor command: %v", err)
	}
	return m
}

// CountPerMeter returns the number of tacho counts in one meter of travel of the motor.
// Calls to CountPerMeter will return an error for non-linear motors.
func (m *LinearActuator) CountPerMeter() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+countPerMeter, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read count per meter: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse count per meter: %v", err)
	}
	return sp, nil
}

// FullTravelCount returns the the number of tacho counts in the full travel of the motor.
// Calls to FullTravelCount will return an error for non-linear motors.
func (m *LinearActuator) FullTravelCount() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+fullTravelCount, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read full travel count: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse full travel count: %v", err)
	}
	return sp, nil
}

// DutyCycle returns the current duty cycle value for the LinearActuator.
func (m *LinearActuator) DutyCycle() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+dutyCycle, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read duty cycle: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse duty cycle: %v", err)
	}
	return sp, nil
}

// DutyCycleSetpoint returns the current duty cycle set point value for the LinearActuator.
func (m *LinearActuator) DutyCycleSetpoint() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+dutyCycleSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read duty cycle set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse duty cycle set point: %v", err)
	}
	return sp, nil
}

// SetDutyCycleSetpoint sets the duty cycle set point value for the LinearActuator
func (m *LinearActuator) SetDutyCycleSetpoint(sp int) *LinearActuator {
	if m.err != nil {
		return m
	}
	if sp < -100 || sp > 100 {
		m.err = fmt.Errorf("ev3dev: invalid duty cycle set point: %d (valid -100 - 100)", sp)
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+dutyCycleSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set duty cycle set point: %v", err)
	}
	return m
}

// Polarity returns the current polarity of the LinearActuator.
func (m *LinearActuator) Polarity() (string, error) {
	if m.err != nil {
		return "", m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+polarity, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read polarity: %v", err)
	}
	return string(b), nil
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
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+polarity, m), string(p))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set polarity %v", err)
	}
	return m
}

// Position returns the current position value for the LinearActuator.
func (m *LinearActuator) Position() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+position, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read position: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse position: %v", err)
	}
	return pos, nil
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
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+position, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set position: %v", err)
	}
	return m
}

// HoldPIDKd returns the derivative constant for the position PID for the LinearActuator.
func (m *LinearActuator) HoldPIDKd() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDkd, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read hold PID Kd: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse hold PID Kd: %v", err)
	}
	return pos, nil
}

// SetHoldPIDKd sets the derivative constant for the position PID for the LinearActuator.
func (m *LinearActuator) SetHoldPIDKd(pos int) *LinearActuator {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDkd, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set hold PID Kd: %v", err)
	}
	return m
}

// HoldPIDKi returns the integral constant for the position PID for the LinearActuator.
func (m *LinearActuator) HoldPIDKi() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDki, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read hold PID Ki: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse hold PID Ki: %v", err)
	}
	return pos, nil
}

// SetHoldPIDKi sets the integral constant for the position PID for the LinearActuator.
func (m *LinearActuator) SetHoldPIDKi(pos int) *LinearActuator {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDki, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set hold PID Ki: %v", err)
	}
	return m
}

// HoldPIDKp returns the proportional constant for the position PID for the LinearActuator.
func (m *LinearActuator) HoldPIDKp() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDkp, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read hold PID Kp: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse hold PID Kp: %v", err)
	}
	return pos, nil
}

// SetHoldPIDKp sets the proportional constant for the position PID for the LinearActuator.
func (m *LinearActuator) SetHoldPIDKp(pos int) *LinearActuator {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDkp, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set hold PID Kp: %v", err)
	}
	return m
}

// MaxSpeed returns  the maximum value that is accepted by SpeedSetpoint.
func (m *LinearActuator) MaxSpeed() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+maxSpeed, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read max speed: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse max speed: %v", err)
	}
	return pos, nil
}

// PositionSetpoint returns the current position set point value for the LinearActuator.
func (m *LinearActuator) PositionSetpoint() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+positionSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read position set point: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse position set point: %v", err)
	}
	return pos, nil
}

// SetPositionSetpoint sets the position set point value for the LinearActuator.
func (m *LinearActuator) SetPositionSetpoint(pos int) *LinearActuator {
	if m.err != nil {
		return m
	}
	if pos != int(int32(pos)) {
		m.err = fmt.Errorf("ev3dev: invalid position set point: %d (valid in int32)", pos)
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+positionSetpoint, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set position set point: %v", err)
	}
	return m
}

// Speed returns the current speed set point value for the LinearActuator.
func (m *LinearActuator) Speed() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speed, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read speed: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse speed: %v", err)
	}
	return sp, nil
}

// SpeedSetpoint returns the current speed set point value for the LinearActuator.
func (m *LinearActuator) SpeedSetpoint() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read speed set point: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse speed set point: %v", err)
	}
	return sp, nil
}

// SetSpeedSetpoint sets the speed set point value for the LinearActuator.
func (m *LinearActuator) SetSpeedSetpoint(sp int) *LinearActuator {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set speed set point: %v", err)
	}
	return m
}

// RampUpSetpoint returns the current ramp up set point value for the LinearActuator.
func (m *LinearActuator) RampUpSetpoint() (time.Duration, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+rampUpSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read ramp up set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse ramp up set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetRampUpSetpoint sets the ramp up set point value for the LinearActuator.
func (m *LinearActuator) SetRampUpSetpoint(d time.Duration) *LinearActuator {
	if m.err != nil {
		return m
	}
	if d < 0 {
		m.err = fmt.Errorf("ev3dev: invalid ramp up set point: %v (must be positive)", d)
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+rampUpSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set ramp up set point: %v", err)
	}
	return m
}

// RampDownSetpoint returns the current ramp down set point value for the LinearActuator.
func (m *LinearActuator) RampDownSetpoint() (time.Duration, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+rampDownSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read ramp down set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse ramp down set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetRampDownSetpoint sets the ramp down set point value for the LinearActuator.
func (m *LinearActuator) SetRampDownSetpoint(d time.Duration) *LinearActuator {
	if m.err != nil {
		return m
	}
	if d < 0 {
		m.err = fmt.Errorf("ev3dev: invalid ramp down set point: %v (must be positive)", d)
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+rampDownSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set ramp down set point: %v", err)
	}
	return m
}

// SpeedPIDKd returns the derivative constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SpeedPIDKd() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDkd, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read speed PID Kd: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse speed PID Kd: %v", err)
	}
	return pos, nil
}

// SetSpeedPIDKd sets the derivative constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SetSpeedPIDKd(pos int) *LinearActuator {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDkd, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set speed PID Kd: %v", err)
	}
	return m
}

// SpeedPIDKi returns the integral constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SpeedPIDKi() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDki, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read speed PID Ki: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse speed PID Ki: %v", err)
	}
	return pos, nil
}

// SetSpeedPIDKi sets the integral constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SetSpeedPIDKi(pos int) *LinearActuator {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDki, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set speed PID Ki: %v", err)
	}
	return m
}

// SpeedPIDKp returns the proportional constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SpeedPIDKp() (int, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDkp, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read speed PID Kp: %v", err)
	}
	pos, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse speed PID Kp: %v", err)
	}
	return pos, nil
}

// SetSpeedPIDKp sets the proportional constant for the speed regulation PID for the LinearActuator.
func (m *LinearActuator) SetSpeedPIDKp(pos int) *LinearActuator {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDkp, m), fmt.Sprintln(pos))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set speed PID Kp: %v", err)
	}
	return m
}

// State returns the current state of the LinearActuator.
func (m *LinearActuator) State() (MotorState, error) {
	if m.err != nil {
		return 0, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+commands, m))
	if err != nil {
		return 0, fmt.Errorf("ev3dev: failed to read tacho-motor commands: %v", err)
	}
	var stat MotorState
	for _, s := range strings.Split(string(chomp(b)), " ") {
		stat |= motorStateTable[s]
	}
	return stat, nil
}

// StopAction returns the stop action used when a stop command is issued
// to the LinearActuator.
func (m *LinearActuator) StopAction() (string, error) {
	if m.err != nil {
		return "", m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+stopAction, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read stop command: %v", err)
	}
	return string(chomp(b)), err
}

// SetStopAction sets the stop action to be used when a stop command is
// issued to the LinearActuator.
func (m *LinearActuator) SetStopAction(comm string) *LinearActuator {
	if m.err != nil {
		return m
	}
	avail, err := m.StopActions()
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
		m.err = fmt.Errorf("ev3dev: stop command %q not available for %s (available:%q)", comm, m, avail)
		return m
	}
	err = m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+stopAction, m), comm)
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set tacho-motor stop command: %v", err)
	}
	return m
}

// StopActions returns the available stop actions for the LinearActuator.
func (m *LinearActuator) StopActions() ([]string, error) {
	if m.err != nil {
		return nil, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+stopActions, m))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read tacho-motor stop command: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}

// TimeSetpoint returns the current time set point value for the LinearActuator.
func (m *LinearActuator) TimeSetpoint() (time.Duration, error) {
	if m.err != nil {
		return -1, m.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+timeSetpoint, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read time set point: %v", err)
	}
	d, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse time set point: %v", err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

// SetTimeSetpoint sets the time set point value for the LinearActuator.
func (m *LinearActuator) SetTimeSetpoint(d time.Duration) *LinearActuator {
	if m.err != nil {
		return m
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+timeSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		m.err = fmt.Errorf("ev3dev: failed to set time set point: %v", err)
	}
	return m
}
