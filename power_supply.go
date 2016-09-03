// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import "path/filepath"

// PowerSupply represents a handle to a the ev3 power supply controller.
// The zero value is usable, reading from the first available device in
// the power supply file system, falling back to the legoev3-battery driver.
// Using another string value will read from the device of that name.
type PowerSupply string

// powerDevice is used to fake a Device. The Type and Err methods
// do not have meaningful semantics.
type powerDevice struct {
	PowerSupply
}

// Path returns the power-supply sysfs path.
func (p PowerSupply) Path() string { return filepath.Join(prefix, PowerSupplyPath) }

func (powerDevice) Type() string { panic("ev3dev: unexpected call of powerDevice Type") }

// String satisfies the fmt.Stringer interface.
//
// String scans the PowerSupplyPath directory if p is the zero value.
// To avoid this the user should set p to the returned value on the first
// use.
func (p PowerSupply) String() string {
	if p == "" {
		cand, err := devicesIn(p.Path())
		if err != nil || len(cand) == 0 {
			return "legoev3-battery"
		}
		return cand[0]
	}
	return string(p)
}

// Err always returns nil since the power device does not support call chains.
func (powerDevice) Err() error { return nil }

// Voltage returns voltage measured from the power supply in volts.
func (p PowerSupply) Voltage() (float64, error) {
	v, err := float64From(attributeOf(powerDevice{p}, voltageNow))
	return v * 1e-6, err
}

// VoltageMin returns the minimum design voltage for the power supply in volts.
func (p PowerSupply) VoltageMin() (float64, error) {
	v, err := float64From(attributeOf(powerDevice{p}, voltageMinDesign))
	return v * 1e-6, err
}

// VoltageMax returns the maximum design voltage for the power supply in volts.
func (p PowerSupply) VoltageMax() (float64, error) {
	v, err := float64From(attributeOf(powerDevice{p}, voltageMaxDesign))
	return v * 1e-6, err
}

// Current returns the current drawn from the power supply in milliamps.
func (p PowerSupply) Current() (float64, error) {
	v, err := float64From(attributeOf(powerDevice{p}, currentNow))
	return v * 1e-3, err
}

// Technology returns the battery technology of the power supply.
func (p PowerSupply) Technology() (string, error) {
	return stringFrom(attributeOf(powerDevice{p}, batteryTechnology))
}

// Type returns the battery type of the power supply.
func (p PowerSupply) Type() (string, error) {
	return stringFrom(attributeOf(powerDevice{p}, batteryType))
}

// Uevent returns the current uevent state for the power supply.
func (p PowerSupply) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(powerDevice{p}, uevent))
}
