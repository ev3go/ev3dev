// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// findafter demonstrates finding the two first available large motors.
//
// The program should be run from the command line after attaching motors
// to the ev3. Invoke the command with the driver name to see the ports
// the motors are connected to.
package main

import (
	"fmt"
	"log"

	"github.com/ev3go/ev3/ev3dev"
)

func main() {
	motors := [2]ev3dev.TachoMotor{}

	var last *ev3dev.TachoMotor
	for i := range &motors {
		err := ev3dev.FindAfter(last, &motors[i], "lego-ev3-l-motor")
		if err != nil {
			log.Fatalf("failed to find motor: %v", err)
		}
		last = &motors[i]
	}

	for i := range &motors {
		addr, err := ev3dev.AddressOf(&motors[i])
		if err != nil {
			log.Fatalf("failed to get address of lego-ev3-l-motor %d device: %v", i, err)
		}
		fmt.Printf("lego-ev3-l-motor %d is in %s\n", i, addr)
	}
}
