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

// TachoMotor represents a handle to a tacho-motor.
type TachoMotor struct {
	mu sync.Mutex
	id int
}

// String satisfies the fmt.Stringer interface.
func (m *TachoMotor) String() string { return fmt.Sprint(motorPrefix, m.id) }

// TachoMotorFor returns a TachoMotor for the given ev3 port name and driver. If the
// motor driver does not match the driver string, a TechoMotor for the port is
// returned with a DriverMismatch error.
// If port is empty, the first tacho-motor satisfying the driver name is returned.
func TachoMotorFor(port, driver string) (*TachoMotor, error) {
	id, err := deviceIDFor(port, driver, TachoMotorPath, motorPrefix)
	if id == -1 {
		return nil, err
	}
	return &TachoMotor{id: id}, err
}

func (m *TachoMotor) writeFile(path, data string) error {
	defer m.mu.Unlock()
	m.mu.Lock()
	return ioutil.WriteFile(path, []byte(data), 0)
}

// Address returns the ev3 port name for the TachoMotor.
func (m *TachoMotor) Address() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+address, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port address: %v", err)
	}
	return string(chomp(b)), err
}

// Commands returns the available commands for the TachoMotor.
func (m *TachoMotor) Commands() ([]string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+commands, m))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read tacho-motor commands: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}

// Command issues a command to the TachoMotor.
func (m *TachoMotor) Command(comm string) error {
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
	err = m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+command, m), comm)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to issue tacho-motor command: %v", err)
	}
	return nil
}

// CountPerRot returns the number of tacho counts in one rotation of the motor.
// Calls to CountPerRot will return an error for non-rotational motors.
func (m *TachoMotor) CountPerRot() (int, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+countPerRot, m))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read count per rotation: %v", err)
	}
	sp, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse count per rotation: %v", err)
	}
	return sp, nil
}

// CountPerMeter returns the number of tacho counts in one meter of travel of the motor.
// Calls to CountPerMeter will return an error for non-linear motors.
func (m *TachoMotor) CountPerMeter() (int, error) {
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
func (m *TachoMotor) FullTravelCount() (int, error) {
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

// Driver returns the driver name for the TachoMotor.
func (m *TachoMotor) Driver() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+driverName, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port driver name: %v", err)
	}
	return string(chomp(b)), err
}

// DutyCycle returns the current duty cycle value for the TachoMotor.
func (m *TachoMotor) DutyCycle() (int, error) {
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

// DutyCycleSetpoint returns the current duty cycle set point value for the TachoMotor.
func (m *TachoMotor) DutyCycleSetpoint() (int, error) {
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

// SetDutyCycleSetpoint sets the duty cycle set point value for the TachoMotor
func (m *TachoMotor) SetDutyCycleSetpoint(sp int) error {
	if sp < -100 || sp > 100 {
		return fmt.Errorf("ev3dev: invalid duty cycle set point: %d (valid -100 - 100)", sp)
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+dutyCycleSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set duty cycle set point: %v", err)
	}
	return nil
}

// Polarity returns the current polarity of the TachoMotor.
func (m *TachoMotor) Polarity() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+polarity, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read polarity: %v", err)
	}
	return string(b), nil
}

// SetPolarity sets the polarity of the TachoMotor
func (m *TachoMotor) SetPolarity(p Polarity) error {
	if p != Normal && p != Inversed {
		return fmt.Errorf("ev3dev: invalid polarity: %q (valid \"normal\" or \"inversed\")", p)
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+polarity, m), string(p))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set polarity %v", err)
	}
	return nil
}

// Position returns the current position value for the TachoMotor.
func (m *TachoMotor) Position() (int, error) {
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

// SetPosition sets the position value for the TachoMotor.
func (m *TachoMotor) SetPosition(pos int) error {
	if pos != int(int32(pos)) {
		return fmt.Errorf("ev3dev: invalid position: %d (valid in int32)", pos)
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+position, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set position: %v", err)
	}
	return nil
}

// HoldPIDKd returns the derivative constant for the position PID for the TachoMotor.
func (m *TachoMotor) HoldPIDKd() (int, error) {
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

// SetHoldPIDKd sets the derivative constant for the position PID for the TachoMotor.
func (m *TachoMotor) SetHoldPIDKd(pos int) error {
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDkd, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set hold PID Kd: %v", err)
	}
	return nil
}

// HoldPIDKi returns the integral constant for the position PID for the TachoMotor.
func (m *TachoMotor) HoldPIDKi() (int, error) {
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

// SetHoldPIDKi sets the integral constant for the position PID for the TachoMotor.
func (m *TachoMotor) SetHoldPIDKi(pos int) error {
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDki, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set hold PID Ki: %v", err)
	}
	return nil
}

// HoldPIDKp returns the proportional constant for the position PID for the TachoMotor.
func (m *TachoMotor) HoldPIDKp() (int, error) {
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

// SetHoldPIDKp sets the proportional constant for the position PID for the TachoMotor.
func (m *TachoMotor) SetHoldPIDKp(pos int) error {
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+holdPIDkp, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set hold PID Kp: %v", err)
	}
	return nil
}

// PositionSetpoint returns the current position set point value for the TachoMotor.
func (m *TachoMotor) PositionSetpoint() (int, error) {
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

// SetPositionSetpoint sets the position set point value for the TachoMotor.
func (m *TachoMotor) SetPositionSetpoint(pos int) error {
	if pos != int(int32(pos)) {
		return fmt.Errorf("ev3dev: invalid position set point: %d (valid in int32)", pos)
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+positionSetpoint, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set position set point: %v", err)
	}
	return nil
}

// Speed returns the current speed set point value for the TachoMotor.
func (m *TachoMotor) Speed() (int, error) {
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

// SpeedSetpoint returns the current speed set point value for the TachoMotor.
func (m *TachoMotor) SpeedSetpoint() (int, error) {
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

// SetSpeedSetpoint sets the speed set point value for the TachoMotor.
func (m *TachoMotor) SetSpeedSetpoint(sp int) error {
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedSetpoint, m), fmt.Sprintln(sp))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set speed set point: %v", err)
	}
	return nil
}

// RampUpSetpoint returns the current ramp up set point value for the TachoMotor.
func (m *TachoMotor) RampUpSetpoint() (time.Duration, error) {
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

// SetRampUpSetpoint sets the ramp up set point value for the TachoMotor.
func (m *TachoMotor) SetRampUpSetpoint(d time.Duration) error {
	if d < 0 {
		return fmt.Errorf("ev3dev: invalid ramp up set point: %v (must be positive)", d)
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+rampUpSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set ramp up set point: %v", err)
	}
	return nil
}

// RampDownSetpoint returns the current ramp down set point value for the TachoMotor.
func (m *TachoMotor) RampDownSetpoint() (time.Duration, error) {
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

// SetRampDownSetpoint sets the ramp down set point value for the TachoMotor.
func (m *TachoMotor) SetRampDownSetpoint(d time.Duration) error {
	if d < 0 {
		return fmt.Errorf("ev3dev: invalid ramp down set point: %v (must be positive)", d)
	}
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+rampDownSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set ramp down set point: %v", err)
	}
	return nil
}

// SpeedPIDKd returns the derivative constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SpeedPIDKd() (int, error) {
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

// SetSpeedPIDKd sets the derivative constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SetSpeedPIDKd(pos int) error {
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDkd, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set speed PID Kd: %v", err)
	}
	return nil
}

// SpeedPIDKi returns the integral constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SpeedPIDKi() (int, error) {
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

// SetSpeedPIDKi sets the integral constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SetSpeedPIDKi(pos int) error {
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDki, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set speed PID Ki: %v", err)
	}
	return nil
}

// SpeedPIDKp returns the proportional constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SpeedPIDKp() (int, error) {
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

// SetSpeedPIDKp sets the proportional constant for the speed regulation PID for the TachoMotor.
func (m *TachoMotor) SetSpeedPIDKp(pos int) error {
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+speedPIDkp, m), fmt.Sprintln(pos))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set speed PID Kp: %v", err)
	}
	return nil
}

// State returns the current state of the TachoMotor.
func (m *TachoMotor) State() (MotorState, error) {
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

// StopCommand returns the stop action used when a stop command is issued
// to the TachoMotor.
func (m *TachoMotor) StopCommand() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+stopCommand, m))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read stop command: %v", err)
	}
	return string(chomp(b)), err
}

// SetStopCommand sets the stop action to be used when a stop command is
// issued to the TachoMotor.
func (m *TachoMotor) SetStopCommand(comm string) error {
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
	err = m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+stopCommand, m), comm)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set tacho-motor stop command: %v", err)
	}
	return nil
}

// StopCommands returns the available stop actions for the TachoMotor.
func (m *TachoMotor) StopCommands() ([]string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(TachoMotorPath+"/%s/"+stopCommands, m))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read tacho-motor stop command: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), nil
}

// TimeSetpoint returns the current time set point value for the TachoMotor.
func (m *TachoMotor) TimeSetpoint() (time.Duration, error) {
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

// SetTimeSetpoint sets the time set point value for the TachoMotor.
func (m *TachoMotor) SetTimeSetpoint(d time.Duration) error {
	err := m.writeFile(fmt.Sprintf(TachoMotorPath+"/%s/"+timeSetpoint, m), fmt.Sprintln(int(d/time.Millisecond)))
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set time set point: %v", err)
	}
	return nil
}
