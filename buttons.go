// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"time"
)

// ButtonPoller allows polling of the ev3 buttons. The zero
// value is ready for use.
type ButtonPoller struct {
	buf []byte
}

// Poll returns a set of Button flags indicating which buttons
// were pressed when the call was made. Poll does not block.
func (b *ButtonPoller) Poll() (Button, error) {
	if b.buf == nil {
		b.buf = make([]byte, keyBufLen)
	}
	ev, err := os.Open(ButtonPath)
	if err != nil {
		return 0, fmt.Errorf("ev3dev: failed to open button event device: %v", err)
	}
	defer ev.Close()

	err = ioctl(ev.Fd(), eviocgkey(b.buf), reflect.ValueOf(b.buf).Index(0).Addr().Pointer())
	if err != nil {
		return 0, fmt.Errorf("ev3dev: failed to set ioctl command for button event device: %v", err)
	}
	return getButton(b.buf), nil
}

func getButton(buf []byte) Button {
	var pressed Button
	for i, bit := range &buttons {
		if isSet(bit, buf) {
			pressed |= 1 << uint(i)
		}
	}
	return pressed
}

// ButtonWaiter provides a mechanism to block waiting for button
// events.
type ButtonWaiter struct {
	Events <-chan ButtonEvent

	f    *os.File
	err  error
	done chan struct{}
	wg   sync.WaitGroup
}

// NewButtonWaiter returns a ButtonWaiter.
func NewButtonWaiter() (*ButtonWaiter, error) {
	ev, err := os.Open(ButtonPath)
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to open button event device: %v", err)
	}

	c := make(chan ButtonEvent)
	b := &ButtonWaiter{Events: c, f: ev, done: make(chan struct{})}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		var buf [16]byte
		for {
			select {
			case <-b.done:
				close(c)
				return
			default:
				_, err := io.ReadFull(ev, buf[:])
				if err != nil {
					c <- ButtonEvent{Err: err}
					continue
				}
				c <- getEvent(buf[:])
			}
		}
	}()
	return b, nil
}

func getEvent(buf []byte) ButtonEvent {
	sec := binary.LittleEndian.Uint32(buf[:4])
	usec := binary.LittleEndian.Uint32(buf[4:8])
	return ButtonEvent{
		Button:    keyTable[binary.LittleEndian.Uint16(buf[10:12])],
		TimeStamp: time.Duration(time.Duration(sec)*time.Second + time.Duration(usec)*time.Microsecond),
		Type:      uint(binary.LittleEndian.Uint16(buf[8:10])),
		Value:     uint(binary.LittleEndian.Uint32(buf[12:16])),
	}
}

// Close closes the backing events source file and the Events channel.
func (b *ButtonWaiter) Close() error {
	select {
	case <-b.done:
		return nil
	default:
		close(b.done)
		b.wg.Wait()
		return b.f.Close()
	}
}

// ButtonEvent is a button event, including the time of the event. The Err
// value reflects any error state arising from detecting the event.
type ButtonEvent struct {
	Button      Button
	TimeStamp   time.Duration
	Type, Value uint
	Err         error
}

// Button is a set of flags indicating which physical button was pressed.
type Button byte

const (
	Back Button = 1 << iota
	Left
	Middle
	Right
	Up
	Down
)

// buttons maps the button numbers used by the ev3 with the linux
// key codes. The order of elements in buttons must match the order
// of Button constants.
var buttons = [...]uint{
	key_backspace,
	key_left,
	key_enter,
	key_right,
	key_up,
	key_down,
}

const (
	key_backspace = 14
	key_enter     = 28
	key_up        = 103
	key_down      = 108
	key_left      = 105
	key_right     = 106

	key_max = 0x2ff

	keyBufLen = (key_max + 7) / 8
)

var keyTable = [key_max]Button{
	key_backspace: Back,
	key_left:      Left,
	key_enter:     Middle,
	key_right:     Right,
	key_up:        Up,
	key_down:      Down,
}
