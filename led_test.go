// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev_test

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	. "github.com/ev3go/ev3dev"

	"github.com/ev3go/ev3"
	"github.com/ev3go/sisyphus"
)

// led is a led sysfs directory.
type led struct {
	maxBrightness int
	brightness    int

	delayOn  time.Duration
	delayOff time.Duration

	trigger map[string]bool

	uevent map[string]string

	t *testing.T
}

// ledMaxBrightness is the max_brightness attribute.
type ledMaxBrightness led

// ReadAt satisfies the io.ReaderAt interface.
func (l *ledMaxBrightness) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, l.maxBrightness)
}

// Size returns the length of the backing data and a nil error.
func (l *ledMaxBrightness) Size() (int64, error) {
	return size(l.maxBrightness), nil
}

// ledBrightness is the brightness attribute.
type ledBrightness led

// ReadAt satisfies the io.ReaderAt interface.
func (l *ledBrightness) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, l.brightness)
}

// Truncate is a no-op.
func (l *ledBrightness) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (l *ledBrightness) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if err != nil {
		l.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	if 0 <= i && i <= l.maxBrightness {
		l.brightness = i
	}
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (l *ledBrightness) Size() (int64, error) {
	return size(l.brightness), nil
}

// ledDelayOn is the delay_on attribute.
type ledDelayOn led

// ReadAt satisfies the io.ReaderAt interface.
func (l *ledDelayOn) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, l)
}

// Truncate is a no-op.
func (l *ledDelayOn) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (l *ledDelayOn) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if err != nil {
		l.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	l.delayOn = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (l *ledDelayOn) Size() (int64, error) {
	return size(l), nil
}

func (l *ledDelayOn) String() string {
	return fmt.Sprint(int(l.delayOn / time.Millisecond))
}

// ledDelayOff is the delay_off attribute.
type ledDelayOff led

// ReadAt satisfies the io.ReaderAt interface.
func (l *ledDelayOff) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, l)
}

// Truncate is a no-op.
func (l *ledDelayOff) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (l *ledDelayOff) WriteAt(b []byte, off int64) (int, error) {
	i, err := strconv.Atoi(string(chomp(b)))
	if i < 0 {
		err = errors.New("ev3dev: negative duration")
	}
	if err != nil {
		l.t.Errorf("unexpected error: %v", err)
		return len(b), syscall.EINVAL
	}
	l.delayOff = time.Duration(i) * time.Millisecond
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (l *ledDelayOff) Size() (int64, error) {
	return size(l), nil
}

func (l *ledDelayOff) String() string {
	return fmt.Sprint(int(l.delayOff / time.Millisecond))
}

// ledTrigger is the trigger attribute.
type ledTrigger led

// ReadAt satisfies the io.ReaderAt interface.
func (l *ledTrigger) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, l)
}

// Truncate is a no-op.
func (l *ledTrigger) Truncate(_ int64) error { return nil }

// WriteAt satisfies the io.WriterAt interface.
func (l *ledTrigger) WriteAt(b []byte, off int64) (int, error) {
	set := string(chomp(b))
	for k := range l.trigger {
		l.trigger[k] = false
	}
	l.trigger[set] = true
	return len(b), nil
}

// Size returns the length of the backing data and a nil error.
func (l *ledTrigger) Size() (int64, error) {
	return size(l), nil
}

func (l *ledTrigger) String() string {
	s := make([]string, 0, len(l.trigger))
	for k, set := range l.trigger {
		if set {
			s = append(s, fmt.Sprintf("[%s]", k))
			continue
		}
		s = append(s, k)
	}
	sort.Strings(s)
	return strings.Join(s, " ")
}

// ledUevent is the uevent attribute.
type ledUevent led

// ReadAt satisfies the io.ReaderAt interface.
func (l *ledUevent) ReadAt(b []byte, offset int64) (int, error) {
	return readAt(b, offset, l)
}

// Size returns the length of the backing data and a nil error.
func (l *ledUevent) Size() (int64, error) {
	return size(l), nil
}

func (l *ledUevent) String() string {
	s := make([]string, 0, len(l.uevent))
	for k, v := range l.uevent {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(s)
	return strings.Join(s, "\n")
}

func ledsysfs(l *led) *sisyphus.FileSystem {
	return sisyphus.NewFileSystem(0775, clock).With(
		d("sys", 0775).With(
			d("class", 0775).With(
				d("leds", 0775).With(
					d(ev3.GreenLeft.String(), 0775).With(
						ro(MaxBrightnessName, 0444, (*ledMaxBrightness)(l)),
						rw(BrightnessName, 0666, (*ledBrightness)(l)),
						rw(TriggerName, 0666, (*ledTrigger)(l)),
						rw(DelayOnName, 0666, (*ledDelayOn)(l)),
						rw(DelayOffName, 0666, (*ledDelayOff)(l)),
						ro(UeventName, 0444, (*ledUevent)(l)),
					),
				),
			),
		),
	).Sync()
}

func TestLED(t *testing.T) {
	l := &led{
		brightness:    255,
		maxBrightness: 255,

		trigger: map[string]bool{
			"none":                                      true,
			"mmc0":                                      false,
			"timer":                                     false,
			"heartbeat":                                 false,
			"default-on":                                false,
			"transient":                                 false,
			"legoev3-battery-charging-or-full":          false,
			"legoev3-battery-charging":                  false,
			"legoev3-battery-full":                      false,
			"legoev3-battery-charging-blink-full-solid": false,
			"rfkill0": false,
		},

		delayOn:  0,
		delayOff: 0,

		uevent: map[string]string{
			"LED_NAME": "left:green",
			"ACTIVE":   "1",
		},

		t: t,
	}

	unmount := serve(ledsysfs(l), t)
	defer unmount()

	t.Run("MaxBrightness", func(t *testing.T) {
		got, err := ev3.GreenLeft.MaxBrightness()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := l.maxBrightness
		if got != want {
			t.Errorf("unexpected maximum brightness value: got:%d want:%d", got, want)
		}
	})

	t.Run("Brightness", func(t *testing.T) {
		for _, b := range []int{0, 64, 128, 192, 255} {
			err := ev3.GreenLeft.SetBrightness(b).Err()
			if err != nil {
				t.Errorf("unexpected error for brightness %d: %v", b, err)
			}

			got, err := ev3.GreenLeft.Brightness()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := b
			if got != want {
				t.Errorf("unexpected brightness value: got:%d want:%d", got, want)
			}
		}
		for _, b := range []int{-1, 256} {
			err := ev3.GreenLeft.SetBrightness(b).Err()
			if err == nil {
				t.Errorf("expected error for brightness %d", b)
			}
		}
	})

	t.Run("Trigger", func(t *testing.T) {
		_, avail, err := ev3.GreenLeft.Trigger()
		if err != nil {
			t.Errorf("unexpected error getting available triggers: %v", err)
		}
		for _, trig := range avail {
			err := ev3.GreenLeft.SetTrigger(trig).Err()
			if err != nil {
				t.Errorf("unexpected error for trigger %q: %v", trig, err)
			}

			got, _, err := ev3.GreenLeft.Trigger()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := trig
			if got != want {
				t.Errorf("unexpected trigger value: got:%q want:%q", got, want)
			}
		}
		for _, trig := range []string{"invalid", "another"} {
			err := ev3.GreenLeft.SetTrigger(trig).Err()
			if err == nil {
				t.Errorf("expected error for trigger %q", trig)
			}

			got, _, err := ev3.GreenLeft.Trigger()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			dontwant := trig
			if got == dontwant {
				t.Errorf("unexpected invalid trigger value: got:%q don't want:%q", got, dontwant)
			}
		}
	})

	t.Run("Delay on", func(t *testing.T) {
		for _, d := range []time.Duration{0, time.Millisecond, time.Second} {
			err := ev3.GreenLeft.SetDelayOn(d).Err()
			if err != nil {
				t.Errorf("unexpected error for delay on %v: %v", d, err)
			}

			got, err := ev3.GreenLeft.DelayOn()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := d
			if got != want {
				t.Errorf("unexpected delay on value: got:%v want:%v", got, want)
			}
		}
		for _, d := range []time.Duration{-time.Millisecond, -time.Second, -time.Minute} {
			err := ev3.GreenLeft.SetDelayOn(d).Err()
			if err == nil {
				t.Errorf("expected error for set delay on %d", d)
			}
		}
		for _, d := range []time.Duration{0, time.Nanosecond, time.Microsecond} {
			err := ev3.GreenLeft.SetDelayOn(d).Err()
			if err != nil {
				t.Errorf("unexpected error for delay on %v: %v", d, err)
			}

			got, err := ev3.GreenLeft.DelayOn()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := time.Duration(0)
			if got != want {
				t.Errorf("unexpected delay on value: got:%v want:%v", got, want)
			}
		}
	})

	t.Run("Delay off", func(t *testing.T) {
		for _, d := range []time.Duration{0, time.Millisecond, time.Second} {
			err := ev3.GreenLeft.SetDelayOff(d).Err()
			if err != nil {
				t.Errorf("unexpected error for delay off %v: %v", d, err)
			}

			got, err := ev3.GreenLeft.DelayOff()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := d
			if got != want {
				t.Errorf("unexpected delay off value: got:%v want:%v", got, want)
			}
		}
		for _, d := range []time.Duration{-time.Millisecond, -time.Second, -time.Minute} {
			err := ev3.GreenLeft.SetDelayOff(d).Err()
			if err == nil {
				t.Errorf("expected error for set delay off %d", d)
			}
		}
		for _, d := range []time.Duration{0, time.Nanosecond, time.Microsecond} {
			err := ev3.GreenLeft.SetDelayOff(d).Err()
			if err != nil {
				t.Errorf("unexpected error for delay off %v: %v", d, err)
			}

			got, err := ev3.GreenLeft.DelayOff()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			want := time.Duration(0)
			if got != want {
				t.Errorf("unexpected delay off value: got:%v want:%v", got, want)
			}
		}
	})

	t.Run("Uevent", func(t *testing.T) {
		got, err := ev3.GreenLeft.Uevent()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		want := l.uevent
		if !reflect.DeepEqual(got, want) {
			t.Errorf("unexpected uevent value: got:%v want:%v", got, want)
		}
	})
}
