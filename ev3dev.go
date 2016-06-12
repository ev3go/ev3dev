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
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
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

type staterDevice interface {
	Device
	State() (MotorState, error)
}

// wait blocks until the wanted motor state under the motor state mask is
// reached, or the timeout is reached.
// The last unmasked motor state is returned unless the timeout was reached
// before the motor state was read.
// When the any parameter is false, wait will return ok as true if
//  (stat^not) & mask == want
// and when any is true wait return false if
//  (stat^not) & mask != 0.
// Otherwise ok will return false indicating that the returned state did
// not match the request.
func wait(d staterDevice, mask, want, not MotorState, any bool, timeout time.Duration) (stat MotorState, ok bool, err error) {
	path := filepath.Join(d.Path(), d.String(), state)
	f, err := os.Open(path)
	if err != nil {
		return 0, false, err
	}
	defer f.Close()

	// If any state in the mask is wanted, we just
	// need to check that (stat^not)&mask is not zero.
	if any {
		want = 0
	}

	end := time.Now().Add(timeout)
	for timeout < 0 || time.Since(end) < 0 {
		fds := []unix.PollFd{{Fd: int32(f.Fd()), Events: unix.POLLIN}}
		_timeout := timeout
		if timeout >= 0 {
			if remain := end.Sub(time.Now()); remain < timeout {
				_timeout = remain
			}
		}
		n, err := unix.Poll(fds, int(_timeout/time.Millisecond))
		if n == 0 {
			return 0, false, err
		}

		stat, err = d.State()
		if err != nil {
			return stat, false, err
		}

		// Check that we have the wanted state.
		if ((stat^not)&mask == want) != any {
			return stat, true, nil
		}

		relax := 50 * time.Millisecond
		if remain := end.Sub(time.Now()); remain < relax {
			relax = remain / 2
		}
		time.Sleep(relax)
	}

	return stat, false, nil
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

// idSetter is a Device that can set its
// device id and clear its error field.
type idSetter interface {
	Device

	// idInt returns the id integer value of
	// the device.
	// A nil idSetter must return -1.
	idInt() int

	// setID sets the device id to the given id
	// and clears the error field to allow an
	// already used Device to be reused.
	setID(id int)
}

// FindAfter finds the first device after d matching the class of the
// dst Device with the given driver name, or returns an error. The
// concrete types of d and dst must match. On return with a nil
// error, dst is usable as a handle for the device.
// If d is nil, FindAfter finds the first matching device.
//
// Only ev3dev.Device implementations are supported.
func FindAfter(d, dst Device, driver string) error {
	_, ok := dst.(idSetter)
	if !ok {
		return fmt.Errorf("ev3dev: device type %T not supported", dst)
	}

	after := -1
	if d != nil {
		if reflect.TypeOf(d) != reflect.TypeOf(dst) {
			return fmt.Errorf("ev3dev: device types do not match %T != %T", d, dst)
		}
		after = d.(idSetter).idInt()
	}

	id, err := deviceIDFor("", driver, dst, after)
	if err != nil {
		return err
	}
	dst.(idSetter).setID(id)
	return nil
}

// IsConnected returns whether the Device is connected.
func IsConnected(d Device) (ok bool, err error) {
	_, err = os.Stat(fmt.Sprintf(d.Path()+"/%s", d))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		err = nil
	}
	return false, err
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
// If port is empty, the first device satisfying the driver name with an id after the
// specified after parameter is returned.
func deviceIDFor(port, driver string, d Device, after int) (int, error) {
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
			if id <= after {
				continue
			}
			path := filepath.Join(d.Path(), device, driverName)
			b, err := ioutil.ReadFile(path)
			if os.IsNotExist(err) {
				// If the device disappeared
				// try the next one.
				continue
			}
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
			err = DriverMismatch{Want: driver, Have: string(chomp(b))}
		}
		return id, err
	}

	if port != "" {
		return -1, fmt.Errorf("ev3dev: could not find device for driver %q on port %s", driver, port)
	}
	if after < 0 {
		return -1, fmt.Errorf("ev3dev: could not find device for driver %q", driver)
	}
	return -1, fmt.Errorf("ev3dev: could find device with driver name %q after %s%d", driver, d.Type(), after)
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
		return fmt.Errorf("ev3dev: failed to set attribute %s: %v", path, err)
	}
	return nil
}
