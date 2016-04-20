// Copyright ©2016 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
)

// PowerSupply represents a handle to a the ev3 power supply controller.
// The zero value is useable, reading from the legoev3-battery driver.
// Using another string value will
type PowerSupply string

// String satisfies the fmt.Stringer interface.
func (p PowerSupply) String() string {
	if p == "" {
		return "legoev3-battery"
	}
	return string(p)
}

// Voltage returns voltage measured from the power supply in volts.
func (p PowerSupply) Voltage() (float64, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(PowerSupplyPath+"/%s/"+voltageNow, p))
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to read voltage: %v", err)
	}
	v, err := strconv.ParseFloat(string(chomp(b)), 64)
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to parse voltage: %v", err)
	}
	return v * 1e-6, nil
}

// VoltageMin returns the minimum design voltage for the power supply in volts.
func (p PowerSupply) VoltageMin() (float64, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(PowerSupplyPath+"/%s/"+voltageMinDesign, p))
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to read voltage minimum: %v", err)
	}
	v, err := strconv.ParseFloat(string(chomp(b)), 64)
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to parse voltage minimum: %v", err)
	}
	return v * 1e-6, nil
}

// VoltageMax returns the maximum design voltage for the power supply in volts.
func (p PowerSupply) VoltageMax() (float64, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(PowerSupplyPath+"/%s/"+voltageMaxDesign, p))
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to read voltage maximum: %v", err)
	}
	v, err := strconv.ParseFloat(string(chomp(b)), 64)
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to parse voltage maximum: %v", err)
	}
	return v * 1e-6, nil
}

// Current returns the current drawn from the power supply in milliamps.
func (p PowerSupply) Current() (float64, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(PowerSupplyPath+"/%s/"+currentNow, p))
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to read current: %v", err)
	}
	v, err := strconv.ParseFloat(string(chomp(b)), 64)
	if err != nil {
		return math.NaN(), fmt.Errorf("ev3dev: failed to parse current: %v", err)
	}
	return v * 1e-3, nil
}

// Technology returns the battery technology of the power supply.
func (p PowerSupply) Technology() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%p/"+batteryTechnology, p))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read battery type: %v", err)
	}
	return string(chomp(b)), err
}

// Type returns the battery technology of the power supply.
func (p PowerSupply) Type() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(SensorPath+"/%p/"+batteryType, p))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read battery type: %v", err)
	}
	return string(chomp(b)), err
}
