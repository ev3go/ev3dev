// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package motorutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ev3go/ev3dev"
)

// Errors is a collection of errors.
type Errors []error

func (e Errors) Error() string {
	if e == nil {
		return "<nil>"
	}
	if len(e) == 0 {
		return "<empty>"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("motorutil: multiple errors: %q", []error(e))
}

// ResetAll resets all the connected motors in the classes tacho-motor,
// servo-motor and dc-motor. Each motor class uses a different reset or
// stop command. ResetAll sends "reset" to tacho-motors, "float" to
// servo-motors and "stop" to dc-motors.
func ResetAll() error {
	paths, err := devicesIn(ev3dev.LegoPortPath)
	if err != nil {
		return err
	}
	var errors Errors
	for _, path := range paths {
		port, err := portFor(ev3dev.LegoPortPath, path)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		p, err := ev3dev.LegoPortFor(port, "")
		if _, ok := err.(ev3dev.DriverMismatch); err != nil && !ok {
			errors = append(errors, err)
			continue
		}

		// Find motors.
		status, err := p.Status()
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if !strings.HasSuffix(status, "motor") {
			continue
		}

		// Get the address of the motor.
		uevent, err := p.Uevent()
		if err != nil {
			errors = append(errors, err)
			continue
		}
		const legoAddress = "LEGO_ADDRESS"
		addr, ok := uevent[legoAddress]
		if !ok {
			errors = append(errors, fmt.Errorf("motorutil: cannot determine "+legoAddress+" for port %q", p))
			continue
		}

		// Send the correct stop command to the motor.
		switch status {
		case "tacho-motor":
			// For our purposes here, this
			// includes linear-actuators.
			t, err := ev3dev.TachoMotorFor(addr, "")
			if _, ok := err.(ev3dev.DriverMismatch); err != nil && !ok {
				errors = append(errors, err)
				continue
			}
			err = t.Command("reset").Err()
			if err != nil {
				errors = append(errors, err)
			}
		case "servo-motor":
			s, err := ev3dev.ServoMotorFor(addr, "")
			if _, ok := err.(ev3dev.DriverMismatch); err != nil && !ok {
				errors = append(errors, err)
				continue
			}
			err = s.Command("float").Err()
			if err != nil {
				errors = append(errors, err)
			}
		case "dc-motor":
			d, err := ev3dev.DCMotorFor(addr, "")
			if _, ok := err.(ev3dev.DriverMismatch); err != nil && !ok {
				errors = append(errors, err)
				continue
			}
			err = d.Command("stop").Err()
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	switch len(errors) {
	case 0:
		return nil
	case 1:
		return errors[0]
	default:
		return errors
	}
}

func devicesIn(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdirnames(0)
}

func portFor(path, base string) (string, error) {
	path = filepath.Join(path, base, "address")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("motorutil: failed to read port: %v", err)
	}
	return string(chomp(b)), nil
}

func chomp(b []byte) []byte {
	if b[len(b)-1] == '\n' {
		return b[:len(b)-1]
	}
	return b
}
