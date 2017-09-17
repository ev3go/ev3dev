// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev_test

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	. "github.com/ev3go/ev3dev"

	"github.com/ev3go/sisyphus"
)

// powerSupply is a power supply sysfs directory.
type powerSupply struct {
	voltage    float64 // V
	voltageMin float64 // V
	voltageMax float64 // V
	current    float64 // mA

	technology string
	typ        string

	uevent map[string]string
}

// powerSupplyVoltage is the voltage_now attribute.
type powerSupplyVoltage powerSupply

// ReadAt satisfies the io.ReaderAt interface.
func (p *powerSupplyVoltage) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *powerSupplyVoltage) Size() (int64, error) {
	return size(p), nil
}

func (p *powerSupplyVoltage) String() string {
	return strconv.Itoa(int(p.voltage * 1e6))
}

// powerSupplyVoltageMin is the voltage_min_design attribute.
type powerSupplyVoltageMin powerSupply

// ReadAt satisfies the io.ReaderAt interface.
func (p *powerSupplyVoltageMin) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *powerSupplyVoltageMin) Size() (int64, error) {
	return size(p), nil
}

func (p *powerSupplyVoltageMin) String() string {
	return strconv.Itoa(int(p.voltageMin * 1e6))
}

// powerSupplyVoltageMax is the voltage_max_design attribute.
type powerSupplyVoltageMax powerSupply

// ReadAt satisfies the io.ReaderAt interface.
func (p *powerSupplyVoltageMax) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *powerSupplyVoltageMax) Size() (int64, error) {
	return size(p), nil
}

func (p *powerSupplyVoltageMax) String() string {
	return strconv.Itoa(int(p.voltageMax * 1e6))
}

// powerSupplyCurrent is the current_now attribute.
type powerSupplyCurrent powerSupply

// ReadAt satisfies the io.ReaderAt interface.
func (p *powerSupplyCurrent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *powerSupplyCurrent) Size() (int64, error) {
	return size(p), nil
}

func (p *powerSupplyCurrent) String() string {
	return strconv.Itoa(int(p.current * 1e3))
}

// powerSupplyTechnology is the technology attribute.
type powerSupplyTechnology powerSupply

// ReadAt satisfies the io.ReaderAt interface.
func (p *powerSupplyTechnology) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p.technology)
}

// Size returns the length of the backing data and a nil error.
func (p *powerSupplyTechnology) Size() (int64, error) {
	return size(p.technology), nil
}

// powerSupplyType is the type attribute.
type powerSupplyType powerSupply

// ReadAt satisfies the io.ReaderAt interface.
func (p *powerSupplyType) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p.typ)
}

// Size returns the length of the backing data and a nil error.
func (p *powerSupplyType) Size() (int64, error) {
	return size(p.typ), nil
}

// powerSupplyUevent is the uevent attribute.
type powerSupplyUevent powerSupply

// ReadAt satisfies the io.ReaderAt interface.
func (p *powerSupplyUevent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, p)
}

// Size returns the length of the backing data and a nil error.
func (p *powerSupplyUevent) Size() (int64, error) {
	return size(p), nil
}

func (p *powerSupplyUevent) String() string {
	s := make([]string, 0, len(p.uevent))
	for k, v := range p.uevent {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(s)
	return strings.Join(s, "\n")
}

func powersupplysysfs(p *powerSupply) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("power_supply", 0775).With(
					d(PowerSupply("").String(), 0775).With(
						ro(VoltageNowName, 0444, (*powerSupplyVoltage)(p)),
						ro(VoltageMinDesignName, 0444, (*powerSupplyVoltageMin)(p)),
						ro(VoltageMaxDesignName, 0444, (*powerSupplyVoltageMax)(p)),
						ro(CurrentNowName, 0444, (*powerSupplyCurrent)(p)),
						ro(BatteryTechnologyName, 0444, (*powerSupplyTechnology)(p)),
						ro(BatteryTypeName, 0444, (*powerSupplyType)(p)),
						ro(UeventName, 0444, (*powerSupplyUevent)(p)),
					),
				),
			),
		),
	).Sync()
}

func TestPowerSupply(t *testing.T) {
	p := &powerSupply{
		voltage:    7.338464,
		voltageMin: 7.1,
		voltageMax: 7.5,
		current:    174.666,

		technology: "Li-ion",
		typ:        "Battery",

		uevent: map[string]string{
			"POWER_SUPPLY_NAME":               "legoev3-battery",
			"POWER_SUPPLY_TECHNOLOGY":         "Li-ion",
			"POWER_SUPPLY_VOLTAGE_MAX_DESIGN": "7500000",
			"POWER_SUPPLY_VOLTAGE_MIN_DESIGN": "7100000",
			"POWER_SUPPLY_VOLTAGE_NOW":        "7338464",
			"POWER_SUPPLY_CURRENT_NOW":        "174666",
			"POWER_SUPPLY_SCOPE":              "System",
		},
	}

	unmount := serve(powersupplysysfs(p), t)
	defer unmount()

	t.Run("Voltage", func(t *testing.T) {
		got, err := PowerSupply("").Voltage()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := p.voltage
		if got != want {
			t.Errorf("unexpected voltage value: got:%f want:%f", got, want)
		}
	})

	t.Run("VoltageMin", func(t *testing.T) {
		got, err := PowerSupply("").VoltageMin()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := p.voltageMin
		if got != want {
			t.Errorf("unexpected voltage min value: got:%f want:%f", got, want)
		}
	})

	t.Run("VoltageMax", func(t *testing.T) {
		got, err := PowerSupply("").VoltageMax()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := p.voltageMax
		if got != want {
			t.Errorf("unexpected voltage max value: got:%f want:%f", got, want)
		}
	})

	t.Run("Current", func(t *testing.T) {
		got, err := PowerSupply("").Current()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := p.current
		if got != want {
			t.Errorf("unexpected current value: got:%f want:%f", got, want)
		}
	})

	t.Run("Technology", func(t *testing.T) {
		got, err := PowerSupply("").Technology()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := p.technology
		if got != want {
			t.Errorf("unexpected technology value: got:%q want:%q", got, want)
		}
	})

	t.Run("Type", func(t *testing.T) {
		got, err := PowerSupply("").Type()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := p.typ
		if got != want {
			t.Errorf("unexpected type value: got:%q want:%q", got, want)
		}
	})

	t.Run("Uevent", func(t *testing.T) {
		got, err := PowerSupply("").Uevent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := p.uevent
		if !reflect.DeepEqual(got, want) {
			t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
		}
	})
}
