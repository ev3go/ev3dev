// Copyright ©2016 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// LED represents a handle to an ev3 LED.
//
// Interaction with shared physical resources is intrinsically
// subject to race conditions without a transactional model,
// which is not provided here. If concurrent access to LEDs is
// needed, the user is required to establish this model.
type LED struct {
	color string
	side  string

	err error
}

var (
	GreenLeft  *LED = &LED{color: "green", side: "left"}
	GreenRight *LED = &LED{color: "green", side: "right"}
	RedLeft    *LED = &LED{color: "red", side: "left"}
	RedRight   *LED = &LED{color: "red", side: "right"}
)

// String satisfies the fmt.Stringer interface.
func (l *LED) String() string { return fmt.Sprintf("ev3:%s:%s", l.side, l.color) }

// Err returns the error state of the LED and clears it.
func (l *LED) Err() error {
	err := l.err
	l.err = nil
	return err
}

func (l *LED) writeFile(path, data string) error {
	return ioutil.WriteFile(path, []byte(data), 0)
}

// MaxBrightness returns the maximum brightness value for the LED.
func (l *LED) MaxBrightness() (int, error) {
	if l.err != nil {
		return -1, l.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(LEDPath+"/%s/"+maxBrightness, l))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read maximum led brightness: %v", err)
	}
	bright, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse maximum led brightness: %v", err)
	}
	return bright, nil
}

// Brightness returns the current brightness value for the LED.
func (l *LED) Brightness() (int, error) {
	if l.err != nil {
		return -1, l.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(LEDPath+"/%s/"+brightness, l))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to read led brightness: %v", err)
	}
	bright, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		return -1, fmt.Errorf("ev3dev: failed to parse led brightness: %v", err)
	}
	return bright, nil
}

// SetBrightness sets the brightness of the LED.
func (l *LED) SetBrightness(bright int) *LED {
	if l.err != nil {
		return l
	}
	max, err := l.MaxBrightness()
	if err != nil {
		l.err = err
		return l
	}
	if bright < 0 || bright > max {
		l.err = fmt.Errorf("ev3dev: invalid led brightness: %d (valid 0-%d)", bright, max)
		return l
	}
	err = l.writeFile(fmt.Sprintf(LEDPath+"/%s/"+brightness, l), fmt.Sprintln(bright))
	if err != nil {
		l.err = fmt.Errorf("ev3dev: failed to set led brightness: %v", err)
	}
	return l
}

// Trigger returns the current and available triggers for the LED.
func (l *LED) Trigger() (current string, available []string, err error) {
	if l.err != nil {
		return "", nil, l.Err()
	}
	b, err := ioutil.ReadFile(fmt.Sprintf(LEDPath+"/%s/"+trigger, l))
	if err != nil {
		return "", nil, fmt.Errorf("ev3dev: failed to read led trigger: %v", err)
	}
	all := strings.Split(string(chomp(b)), " ")
	current = strings.Trim(all[0], "[]")
	return current, all[1:], err
}

// SetTrigger sets the trigger for the LED.
func (l *LED) SetTrigger(trig string) *LED {
	if l.err != nil {
		return l
	}
	_, avail, err := l.Trigger()
	if err != nil {
		l.err = err
		return l
	}
	ok := false
	for _, t := range avail {
		if t == trig {
			ok = true
			break
		}
	}
	if !ok {
		l.err = fmt.Errorf("ev3dev: led trigger %q not available for %s (available:%q)", mode, l, avail)
		return l
	}
	err = l.writeFile(fmt.Sprintf(LEDPath+"/%s/"+trigger, l), trig)
	if err != nil {
		l.err = fmt.Errorf("ev3dev: failed to set led trigger: %v", err)
	}
	return l
}
