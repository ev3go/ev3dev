// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// znap is a reimplementation of the control program for the Znap robot with
// the modification that the ultrasonic sensor is autodetected so that it works
// with the instructions as printed and with models built to work with the
// control program provided in the Mindstorms software.
//
// http://robotsquare.com/wp-content/uploads/2013/10/45544_45560_znap.pdf
package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ev3go/ev3dev"
)

func main() {
	b, err := ev3dev.TachoMotorFor("outB", "lego-ev3-m-motor")
	if err != nil {
		log.Fatalf("failed to find motor for jaw in outB: %v", err)
	}
	a, err := ev3dev.TachoMotorFor("outA", "lego-ev3-l-motor")
	if err != nil {
		log.Fatalf("failed to find left motor in outA: %v", err)
	}
	d, err := ev3dev.TachoMotorFor("outD", "lego-ev3-l-motor")
	if err != nil {
		log.Fatalf("failed to find left motor in outD: %v", err)
	}
	go func() {
		for {
			bstat, _ := b.State()
			bspeed, _ := b.Speed()
			log.Printf("outB: %s %d", bstat, bspeed)
			astat, _ := a.State()
			aspeed, _ := a.Speed()
			log.Printf("outA: %s %d", astat, aspeed)
			dstat, _ := d.State()
			dspeed, _ := d.Speed()
			log.Printf("outD: %s %d", dstat, dspeed)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	go znap()
	wander()
}

var check atomic.Value

func setChecking(c bool) { check.Store(c) }
func isChecking() bool   { c, _ := check.Load().(bool); return c }

var attacking atomic.Value

func setAttacking(p bool) { attacking.Store(p) }
func isAttacking() bool   { p, _ := attacking.Load().(bool); return p }

func znap() {
	jaw, err := ev3dev.TachoMotorFor("outB", "lego-ev3-m-motor")
	if err != nil {
		log.Fatalf("failed to find motor for jaw in outB: %v", err)
	}
	max, err := jaw.MaxSpeed()
	if err != nil {
		log.Fatalf("failed to read maximum jaw speed: %v", err)
	}
	max /= 2
	err = jaw.
		SetRampUpSetpoint(200 * time.Millisecond).
		SetRampDownSetpoint(200 * time.Millisecond).
		Err()
	if err != nil {
		log.Fatalf("failed to set jaw acceleration: %v", err)
	}

	us, err := ev3dev.SensorFor("", "lego-ev3-us")
	if err != nil {
		log.Fatalf("failed to find ultrasonic sensor: %v", err)
	}
	d, err := us.Decimals()
	if err != nil {
		log.Fatalf("failed to read decimal precision: %v", err)
	}
	s := 1 / math.Pow10(d)

	for {
		if !isChecking() {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		ds, err := us.Value(0)
		if err != nil {
			log.Fatalf("failed to read distance: %v", err)
		}
		dist, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			log.Fatalf("failed to parse distance: %v", err)
		}
		dist *= s
		if dist < 40 {
			setAttacking(true)
			if dist < 25 {
				err = jaw.
					SetSpeedSetpoint(-max).
					Command("run-forever").
					Err()
				if err != nil {
					log.Fatalf("failed to run jaw motor: %v", err)
				}

				// play t-rex roar

				time.Sleep(time.Second / 4)

				err = jaw.
					SetStopAction("coast").
					Command("stop").
					Err()
				if err != nil {
					log.Fatalf("failed to stop jaw motor: %v", err)
				}

				time.Sleep(time.Second)

				stat, ok, err := wait(jaw, ev3dev.Running, 0, 0, false, 10*time.Second)
				if err != nil {
					log.Fatalf("failed to wait for jaw motor to return: %v", err)
				}
				if !ok {
					log.Fatalf("failed to wait for jaw motor to return: %v", stat)
				}
				err = jaw.
					SetSpeedSetpoint(max).
					SetTimeSetpoint(time.Second).
					SetStopAction("hold").
					Command("run-timed").
					Err()
				if err != nil {
					log.Fatalf("failed to run jaw motor: %v", err)
				}
			} else {
				err = jaw.
					SetSpeedSetpoint(max).
					SetPositionSetpoint(-120).
					SetStopAction("hold").
					Command("run-to-rel-pos").
					Err()
				if err != nil {
					log.Fatalf("failed to run jaw motor: %v", err)
				}

				// play snake sound

				stat, ok, err := wait(jaw, ev3dev.Running, 0, 0, false, 10*time.Second)
				if err != nil {
					log.Fatalf("failed to wait for jaw motor to threaten: %v", err)
				}
				if !ok {
					log.Fatalf("failed to wait for jaw motor to threaten: %v", stat)
				}
				err = jaw.
					SetSpeedSetpoint(max).
					SetPositionSetpoint(120).
					SetStopAction("coast").
					Command("run-to-rel-pos").
					Err()
				if err != nil {
					log.Fatalf("failed to run jaw motor: %v", err)
				}
			}
			stat, ok, err := wait(jaw, ev3dev.Running, 0, 0, false, 10*time.Second)
			if err != nil {
				log.Fatalf("failed to wait for jaw motor to return: %v", err)
			}
			if !ok {
				log.Fatalf("failed to wait for jaw motor to return: %v", stat)
			}
			setAttacking(false)

			time.Sleep(time.Second / 2)
		}
	}
}

func wander() {
	left, err := ev3dev.TachoMotorFor("outA", "lego-ev3-l-motor")
	if err != nil {
		if err != nil {
			log.Fatalf("failed to find left motor in outA: %v", err)
		}
	}
	left.
		SetPolarity(ev3dev.Inversed).
		SetRampUpSetpoint(200 * time.Millisecond).
		SetRampDownSetpoint(200 * time.Millisecond).
		SetStopAction("brake").
		Err()
	if err != nil {
		log.Fatalf("failed to set initialize left track: %v", err)
	}
	right, err := ev3dev.TachoMotorFor("outD", "lego-ev3-l-motor")
	if err != nil {
		if err != nil {
			log.Fatalf("failed to find right motor in outD: %v", err)
		}
	}
	right.
		SetPolarity(ev3dev.Inversed).
		SetRampUpSetpoint(200 * time.Millisecond).
		SetRampDownSetpoint(200 * time.Millisecond).
		SetStopAction("brake").
		Err()
	if err != nil {
		log.Fatalf("failed to set initialize right track: %v", err)
	}
	max, err := left.MaxSpeed() // Assume left and right have same maximum.
	if err != nil {
		log.Fatalf("failed to read maximum track speed: %v", err)
	}
	max /= 2

	s := steering{left, right}
	for {
	rol:
		for {
			setChecking(true)
			if isAttacking() {
				break rol
			}
			for _, move := range []struct {
				speed int
				dir   float64
			}{
				{speed: max, dir: 1},
				{speed: max, dir: -1},
				{speed: max, dir: 0},
			} {
				err = s.steer(move.speed, (rand.Intn(3)+1)*360, move.dir)
				if err != nil {
					log.Fatalf("failed to steer %v/%v: %v", move.speed, move.dir, err)
				}
				if isAttacking() {
					break rol
				}
			}
		}
		setChecking(false)

		left.Command("stop")
		right.Command("stop")
		err = left.Err()
		if err != nil {
			log.Fatalf("failed to stop left track: %v", err)
		}
		err = right.Err()
		if err != nil {
			log.Fatalf("failed to stop right track: %v", err)
		}

		time.Sleep(2 * time.Second)
	}
}

type steering struct {
	left, right *ev3dev.TachoMotor
}

func (s steering) steer(speed, counts int, dir float64) error {
	if dir < -1 || 1 < dir || math.IsNaN(dir) {
		return fmt.Errorf("direction out of range: %v", dir)
	}

	var ls, rs, lcounts, rcounts int
	switch {
	case dir == 0:
		ls = speed
		rs = speed
		lcounts = counts
		rcounts = counts
	case dir < 0:
		rs = speed
		rcounts = counts
		dir = (dir + 0.5) * 2
		ls = int(math.Abs(dir * float64(speed)))
		lcounts = int(float64(rcounts) * dir)
	case dir > 0:
		ls = speed
		lcounts = counts
		dir = (0.5 - dir) * 2
		rs = int(math.Abs(dir * float64(speed)))
		rcounts = int(float64(lcounts) * dir)
	}

	var err error
	err = s.left.
		SetSpeedSetpoint(ls).
		SetPositionSetpoint(lcounts).
		Err()
	if err != nil {
		return err
	}
	err = s.right.
		SetSpeedSetpoint(rs).
		SetPositionSetpoint(rcounts).
		Err()
	if err != nil {
		return err
	}

	err = s.left.Command("run-to-rel-pos").Err()
	if err != nil {
		return err
	}
	err = s.right.Command("run-to-rel-pos").Err()
	if err != nil {
		s.left.Command("stop").Err()
		return err
	}
	var stat ev3dev.MotorState
	var ok bool
	stat, ok, err = wait(s.left, ev3dev.Running, 0, 0, false, 5*time.Second)
	if err != nil {
		log.Fatalf("failed to wait for left motor to stop: %v", err)
	}
	if !ok {
		log.Fatalf("failed to wait for left motor to stop: %v", stat)
	}
	stat, ok, err = wait(s.right, ev3dev.Running, 0, 0, false, 5*time.Second)
	if err != nil {
		log.Fatalf("failed to wait for right motor to stop: %v", err)
	}
	if !ok {
		log.Fatalf("failed to wait for right motor to stop: %v", stat)
	}
	return nil
}

type staterDevice interface {
	ev3dev.Device
	State() (ev3dev.MotorState, error)
}

// TODO(kortschak) Replace wait with the ev3dev wait function when the
// kernel supports polling on motors states.
//
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
func wait(d staterDevice, mask, want, not ev3dev.MotorState, any bool, timeout time.Duration) (stat ev3dev.MotorState, ok bool, err error) {
	path := filepath.Join(d.Path(), d.String(), "state")
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
