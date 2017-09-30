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
	"log"
	"math"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ev3go/ev3dev"
	"github.com/ev3go/ev3dev/motorutil"
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
	max := jaw.MaxSpeed() / 2
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

				stat, ok, err := ev3dev.Wait(jaw, ev3dev.Running, 0, 0, false, 10*time.Second)
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

				stat, ok, err := ev3dev.Wait(jaw, ev3dev.Running, 0, 0, false, 10*time.Second)
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
			stat, ok, err := ev3dev.Wait(jaw, ev3dev.Running, 0, 0, false, 10*time.Second)
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
		SetStopAction("hold").
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
		SetStopAction("hold").
		Err()
	if err != nil {
		log.Fatalf("failed to set initialize right track: %v", err)
	}
	max := left.MaxSpeed() / 2 // Assume left and right have same maximum.

	s := motorutil.Steering{Left: left, Right: right, Timeout: 5 * time.Second}
	for {
	rol:
		for {
			setChecking(true)
			if isAttacking() {
				break rol
			}
			for _, move := range []struct {
				speed int
				dir   int
			}{
				{speed: max, dir: 100},
				{speed: max, dir: -100},
				{speed: max, dir: 0},
			} {
				err = s.SteerCounts(move.speed, move.dir, (rand.Intn(3)+1)*360).Wait()
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
