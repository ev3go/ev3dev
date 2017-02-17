// Copyright ©2017 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
)

const (
	tone = 0x02

	ev_snd = 0x12
	ev_max = 0x1f

	evBufLen = (ev_max + 7) / 8
)

// Speaker is an evdev sound device.
type Speaker struct {
	path string
	f    *os.File
	buf  [16]byte
}

// NewSpeaker returns a new Speaker based on the given evdev snd device path.
func NewSpeaker(path string) *Speaker {
	s := Speaker{path: path}
	binary.LittleEndian.PutUint16(s.buf[8:10], ev_snd)
	binary.LittleEndian.PutUint16(s.buf[10:12], tone)
	return &s
}

func hasSound(path string) (bool, error) {
	ev, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("ev3dev: failed to open sound event device: %v", err)
	}
	defer ev.Close()

	var buf [evBufLen]byte
	err = ioctl(ev.Fd(), eviocgbit(0, ev_max), reflect.ValueOf(buf[:]).Index(0).Addr().Pointer())
	if err != nil {
		return false, fmt.Errorf("ev3dev: failed to set ioctl command for sound event device: %v", err)
	}
	return isSet(ev_snd, buf[:]), nil
}

// Init prepares a Speaker for use.
func (s *Speaker) Init() error {
	ok, err := hasSound(s.path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("ev3dev: sound events not available for %q", s.path)
	}

	s.f, err = os.OpenFile(s.path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to open sound event device: %v", err)
	}
	return nil
}

// Tone plays a tone at the specified frequency from the ev3 speaker.
// If freq is zero, playing is stopped.
func (s *Speaker) Tone(freq uint32) error {
	binary.LittleEndian.PutUint32(s.buf[12:16], freq)
	_, err := s.f.Write(s.buf[:])
	return err
}

// Close closes the Speaker. After return, the Speaker may not be used unless
// Init is called again.
func (s *Speaker) Close() error {
	err := s.f.Close()
	s.f = nil
	return err
}
