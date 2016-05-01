// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ev3dev provides low level access to the ev3dev control and sensor drivers.
// See documentation at http://www.ev3dev.org/docs/drivers/.
//
// The API provided in the ev3dev package allows fluent chaining of action calls.
// Methods for each of the device handle types are split into two classes: action
// and result. Action method calls return the receiver and result method calls
// return an error value generally with another result. Action methods result in
// a change of state in the robot while result methods return the requested attribute
// state of the robot.
//
// To allow fluent call chains, errors are sticky for action methods and are cleared
// and returned by result methods. In a chain of calls the first error that is caused
// by an action method prevents execution of all subsequent action method calls, and
// is returned by the first result method called on the device handle, clearing the
// error state. Any attribute value returned by a call chain returning a non-nil error
// is invalid.
package ev3dev

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	linearPrefix = "linear"
	motorPrefix  = "motor"
	portPrefix   = "port"
	sensorPrefix = "sensor"
)

const (
	// LEDPath is the path to the ev3 LED file system.
	LEDPath = "/sys/class/leds"

	// ButtonPath is the path to the ev3 button events.
	ButtonPath = "/dev/input/by-path/platform-gpio-keys.0-event"

	// LegoPortPath is the path to the ev3 lego-port file system.
	LegoPortPath = "/sys/class/lego-port"

	// SensorPath is the path to the ev3 lego-sensor file system.
	SensorPath = "/sys/class/lego-sensor"

	// TachoMotorPath is the path to the ev3 tacho-motor file system.
	TachoMotorPath = "/sys/class/tacho-motor"

	// ServoMotorPath is the path to the ev3 servo-motor file system.
	ServoMotorPath = "/sys/class/servo-motor"

	// DCMotorPath is the path to the ev3 dc-motor file system.
	DCMotorPath = "/sys/class/dc-motor"

	// PowerSupplyPath is the path to the ev3 power supply file system.
	PowerSupplyPath = "/sys/class/power_supply"
)

// These are the subsystem path definitions for all device classes.
const (
	address                   = "address"
	binData                   = "bin_data"
	batteryTechnology         = "technology"
	batteryType               = "type"
	binDataFormat             = "bin_data_format"
	brightness                = "brightness"
	command                   = "command"
	commands                  = "commands"
	countPerMeter             = "count_per_m"
	countPerRot               = "count_per_rot"
	currentNow                = "current_now"
	decimals                  = "decimals"
	delayOff                  = "delay_off"
	delayOn                   = "delay_on"
	direct                    = "direct"
	driverName                = "driver_name"
	dutyCycle                 = "duty_cycle"
	dutyCycleSetpoint         = "duty_cycle_sp"
	fullTravelCount           = "full_travel_count"
	holdPID                   = "hold_pid/"
	holdPIDkd                 = holdPID + kd
	holdPIDki                 = holdPID + ki
	holdPIDkp                 = holdPID + kp
	kd                        = "Kd"
	ki                        = "Ki"
	kp                        = "Kp"
	maxBrightness             = "max_brightness"
	maxPulseSetpoint          = "max_pulse_sp"
	midPulseSetpoint          = "mid_pulse_sp"
	minPulseSetpoint          = "min_pulse_sp"
	maxSpeed                  = "max_speed"
	mode                      = "mode"
	modes                     = "modes"
	numValues                 = "num_values"
	polarity                  = "polarity"
	pollRate                  = "poll_ms"
	position                  = "position"
	positionSetpoint          = "position_sp"
	power                     = "power/"
	powerAutosuspendDelay     = power + "autosuspend_delay_ms"
	powerControl              = power + "control"
	powerRuntimeActiveTime    = power + "runtime_active_time"
	powerRuntimeStatus        = power + "runtime_status"
	powerRuntimeSuspendedTime = power + "runtime_suspended_time"
	rampDownSetpoint          = "ramp_down_sp"
	rampUpSetpoint            = "ramp_up_sp"
	rateSetpoint              = "rate_sp"
	setDevice                 = "set_device"
	speed                     = "speed"
	speedPID                  = "speed_pid/"
	speedPIDkd                = speedPID + kd
	speedPIDki                = speedPID + ki
	speedPIDkp                = speedPID + kp
	speedSetpoint             = "speed_sp"
	state                     = "state"
	status                    = "status"
	stopAction                = "stop_action"
	stopActions               = "stop_actions"
	subsystem                 = "subsystem"
	textValues                = "text_values"
	timeSetpoint              = "time_sp"
	trigger                   = "trigger"
	uevent                    = "uevent"
	units                     = "units"
	value                     = "value"
	voltageMaxDesign          = "voltage_max_design"
	voltageMinDesign          = "voltage_min_design"
	voltageNow                = "voltage_now"
)

// Polarity represent motor polarity states.
type Polarity string

const (
	Normal   Polarity = "normal"
	Inversed Polarity = "inversed"
)

// MotorState is a flag set representing the state of a TachoMotor.
type MotorState uint

const (
	Running MotorState = 1 << iota
	Ramping
	Holding
	Overloaded
	Stalled
)

const (
	running    = "running"
	ramping    = "ramping"
	holding    = "holding"
	overloaded = "overloaded"
	stalled    = "stalled"
)

var motorStateTable = map[string]MotorState{
	running:    Running,
	ramping:    Ramping,
	holding:    Holding,
	overloaded: Overloaded,
	stalled:    Stalled,
}

var motorStates = []string{
	running,
	ramping,
	holding,
	overloaded,
	stalled,
}

// String satisfies the fmt.Stringer interface.
func (f MotorState) String() string {
	const stateMask = Running | Ramping | Holding | Overloaded | Stalled

	var b []byte
	for i, s := range motorStates {
		if f&(1<<uint(i)) != 0 {
			if i != 0 {
				b = append(b, '|')
			}
			b = append(b, s...)
		}
	}

	return string(b)
}

// DriverMismatch errors are returned when a device is found that
// does not match the requested driver.
type DriverMismatch struct {
	// Want is the string describing
	// the requested driver.
	Want string

	// Have is the string describing
	// the driver present on the device.
	Have string
}

func (e DriverMismatch) Error() string {
	return fmt.Sprintf("ev3dev: mismatched driver names: want %q but have %q", e.Want, e.Have)
}

// Device is an ev3dev API device.
type Device interface {
	// Path returns the sysfs path
	// for the device type.
	Path() string

	// Type returns the type of the
	// device, one of "linear", "motor",
	// "port" or "sensor".
	Type() string

	// Err returns and clears the
	// error state of the Device.
	Err() error

	fmt.Stringer
}

// AddressOf returns the port address of the Device.
func AddressOf(d Device) (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(d.Path()+"/%s/"+address, d))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read %s address: %v", d.Type(), err)
	}
	return string(chomp(b)), err
}

// DriverFor returns the driver name for the Device.
func DriverFor(d Device) (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(d.Path()+"/%s/"+driverName, d))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read %s driver name: %v", d.Type(), err)
	}
	return string(chomp(b)), err
}

// deviceIDFor returns the id for the given ev3 port name and driver of the Device.
// If the driver does not match the driver string, an id for the device is returned
// with a DriverMismatch error.
// If port is empty, the first device satisfying the driver name is returned.
func deviceIDFor(port, driver string, d Device) (int, error) {
	devices, err := devicesIn(d.Path())
	if err != nil {
		return -1, fmt.Errorf("ev3dev: could not get devices for %s: %v", d.Path(), err)
	}

	portBytes := []byte(port)
	driverBytes := []byte(driver)
	for _, device := range devices {
		if !strings.HasPrefix(device, d.Type()) {
			continue
		}
		id, err := strconv.Atoi(strings.TrimPrefix(device, d.Type()))
		if err != nil {
			return -1, fmt.Errorf("ev3dev: could not parse id from device name %q: %v", device, err)
		}

		if port == "" {
			path := filepath.Join(d.Path(), device, driverName)
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return -1, fmt.Errorf("ev3dev: could not read driver name %s: %v", path, err)
			}
			if !bytes.Equal(driverBytes, chomp(b)) {
				continue
			}
			return id, nil
		}

		path := filepath.Join(d.Path(), device, address)
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return -1, fmt.Errorf("ev3dev: could not read address %s: %v", path, err)
		}
		if !bytes.Equal(portBytes, chomp(b)) {
			continue
		}
		path = filepath.Join(d.Path(), device, driverName)
		b, err = ioutil.ReadFile(path)
		if err != nil {
			return -1, fmt.Errorf("ev3dev: could not read driver name %s: %v", path, err)
		}
		if !bytes.Equal(driverBytes, chomp(b)) {
			err = DriverMismatch{Want: driver, Have: string(b)}
		}
		return id, err
	}

	if port != "" {
		return -1, fmt.Errorf("ev3dev: could not find device for driver %q on port %s", driver, port)
	}
	return -1, fmt.Errorf("ev3dev: could not find device for driver %q", driver)
}

func devicesIn(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdirnames(0)
}

func attributeOf(d Device, attr string) (data string, _attr string, err error) {
	err = d.Err()
	if err != nil {
		return "", "", err
	}
	path := filepath.Join(d.Path(), d.String(), attr)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("ev3dev: failed to read attribute %s: %v", path, err)
	}
	return string(chomp(b)), attr, nil
}

func chomp(b []byte) []byte {
	if b[len(b)-1] == '\n' {
		return b[:len(b)-1]
	}
	return b
}

func intFrom(data, attr string, err error) (int, error) {
	if err != nil {
		return -1, err
	}
	i, err := strconv.Atoi(data)
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse %s: %v", attr, err)
	}
	return i, nil
}

func float64From(data, attr string, err error) (float64, error) {
	if err != nil {
		return math.NaN(), err
	}
	f, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to parse %s: %v", attr, err)
	}
	return f, nil
}

func durationFrom(data, attr string, err error) (time.Duration, error) {
	if err != nil {
		return -1, err
	}
	d, err := strconv.Atoi(data)
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse %s: %v", attr, err)
	}
	return time.Duration(d) * time.Millisecond, nil
}

func stringFrom(data, _ string, err error) (string, error) {
	return data, err
}

func stringSliceFrom(data, _ string, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	return strings.Split(data, " "), nil
}

func ueventFrom(data, attr string, err error) (map[string]string, error) {
	if err != nil {
		return nil, err
	}
	uevent := make(map[string]string)
	for _, l := range strings.Split(data, "\n") {
		parts := strings.Split(l, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("ev3dev: failed to parse %s: unexpected line %q", attr, l)
		}
		uevent[parts[0]] = parts[1]
	}
	return uevent, nil
}

func setAttributeOf(d Device, attr, data string) error {
	path := filepath.Join(d.Path(), d.String(), attr)
	err := ioutil.WriteFile(path, []byte(data), 0)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set attribute %s: %v", path, attr, err)
	}
	return nil
}
