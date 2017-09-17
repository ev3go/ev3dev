// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"
)

// LED represents a handle to an ev3 LED.
//
// Interaction with shared physical resources is intrinsically
// subject to race conditions without a transactional model,
// which is not provided here. If concurrent access to LEDs is
// needed, the user is required to establish this model.
type LED struct {
	Name fmt.Stringer

	err error
}

// ledDevice is used to fake a Device. The Type method do not
// have meaningful semantics.
type ledDevice struct {
	*LED
}

// Path returns the LED sysfs path.
func (l *LED) Path() string { return filepath.Join(prefix, LEDPath) }

func (ledDevice) Type() string { panic("ev3dev: unexpected call of ledDevice Type") }

// String satisfies the fmt.Stringer interface.
func (l *LED) String() string { return l.Name.String() }

// Err returns the error state of the LED and clears it.
func (l *LED) Err() error {
	err := l.err
	l.err = nil
	return err
}

// MaxBrightness returns the maximum brightness value for the LED.
func (l *LED) MaxBrightness() (int, error) {
	return intFrom(attributeOf(ledDevice{l}, maxBrightness))
}

// Brightness returns the current brightness value for the LED.
func (l *LED) Brightness() (int, error) {
	return intFrom(attributeOf(ledDevice{l}, brightness))
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
		l.err = newValueOutOfRangeError(ledDevice{l}, brightness, bright, 0, max)
		return l
	}
	l.err = setAttributeOf(ledDevice{l}, brightness, strconv.Itoa(bright))
	return l
}

// Trigger returns the current and available triggers for the LED.
func (l *LED) Trigger() (current string, available []string, err error) {
	all, err := stringSliceFrom(attributeOf(ledDevice{l}, trigger))
	if err != nil {
		return "", nil, err
	}
	for i, t := range all {
		if t[0] == '[' && t[len(t)-1] == ']' {
			all[i] = t[1 : len(t)-1]
			current = all[i]
		}
	}
	if current == "" {
		return "", available, errors.New("ev3dev: could not find current trigger")
	}
	return current, all, err
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
		l.err = newInvalidValueError(ledDevice{l}, trigger, "", trig, avail)
		return l
	}
	l.err = setAttributeOf(ledDevice{l}, trigger, trig)
	return l
}

// DelayOff returns the duration for which the LED is off when using the timer trigger.
func (l *LED) DelayOff() (time.Duration, error) {
	return durationFrom(attributeOf(ledDevice{l}, delayOff))
}

// SetDelayOff sets the duration for which the LED is off when using the timer trigger.
func (l *LED) SetDelayOff(d time.Duration) *LED {
	if l.err != nil {
		return l
	}
	if d < 0 {
		l.err = newNegativeDurationError(ledDevice{l}, delayOff, d)
		return l
	}
	l.err = setAttributeOf(ledDevice{l}, delayOff, strconv.Itoa(int(d/time.Millisecond)))
	return l
}

// DelayOn returns the duration for which the LED is on when using the timer trigger.
func (l *LED) DelayOn() (time.Duration, error) {
	return durationFrom(attributeOf(ledDevice{l}, delayOn))
}

// SetDelayOn sets the duration for which the LED is on when using the timer trigger.
func (l *LED) SetDelayOn(d time.Duration) *LED {
	if l.err != nil {
		return l
	}
	if d < 0 {
		l.err = newNegativeDurationError(ledDevice{l}, delayOn, d)
		return l
	}
	l.err = setAttributeOf(ledDevice{l}, delayOn, strconv.Itoa(int(d/time.Millisecond)))
	return l
}

// Uevent returns the current uevent state for the LED.
func (l *LED) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(ledDevice{l}, uevent))
}
