// Copyright ©2016 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// demo is a reimplementation of the Demo program loaded on new ev3 bricks,
// without sound or graphics. It demonstrates the use of the ev3dev Go API.
// The control does not make full use of the ev3dev API where it could.
package main

import (
	"log"
	"time"

	"github.com/kortschak/ev3/ev3dev"
)

// TODO(kortschak): Add the gopher as a replacement for the eyes in the original demo.

func main() {
	// Get the handle for the medium motor on outA.
	outA, err := ev3dev.TachoMotorFor("outA", "lego-ev3-m-motor")
	if err != nil {
		log.Fatalf("failed to find medium motor on outA: %v", err)
	}
	err = outA.SetStopCommand("brake")
	if err != nil {
		log.Fatalf("failed to set brake stop for medium motor on outA: %v", err)
	}

	// Get the handle for the left large motor on outB.
	outB, err := ev3dev.TachoMotorFor("outB", "lego-ev3-l-motor")
	if err != nil {
		log.Fatalf("failed to find left large motor on outB: %v", err)
	}
	err = outB.SetStopCommand("brake")
	if err != nil {
		log.Fatalf("failed to set brake stop for left large motor on outB: %v", err)
	}

	// Get the handle for the right large motor on outC.
	outC, err := ev3dev.TachoMotorFor("outC", "lego-ev3-l-motor")
	if err != nil {
		log.Fatalf("failed to find right large motor on outC: %v", err)
	}
	err = outC.SetStopCommand("brake")
	if err != nil {
		log.Fatalf("failed to set brake stop for right large motor on outB: %v", err)
	}

	for i := 0; i < 2; i++ {
		// Run medium motor on outA at power 50, wait for 0.5 second and then brake.
		outA.SetDutyCycleSetPoint(50)
		outA.Command("run-forever")
		time.Sleep(time.Second / 2)
		outA.Command("stop")

		// Run large motors on B+C at power 70, wait for 2 second and then brake.
		outB.SetDutyCycleSetPoint(70)
		outC.SetDutyCycleSetPoint(70)
		outB.Command("run-forever")
		outC.Command("run-forever")
		time.Sleep(2 * time.Second)
		outB.Command("stop")
		outC.Command("stop")

		// Run medium motor on outA at power -75, wait for 0.5 second and then brake.
		outA.SetDutyCycleSetPoint(-75)
		outA.Command("run-forever")
		time.Sleep(time.Second / 2)
		outA.Command("stop")

		// Run large motors on B at power -50 and C at power 50, wait for 1 second and then brake.
		outB.SetDutyCycleSetPoint(-50)
		outC.SetDutyCycleSetPoint(50)
		outB.Command("run-forever")
		outC.Command("run-forever")
		time.Sleep(time.Second)
		outB.Command("stop")
		outC.Command("stop")
	}
}
