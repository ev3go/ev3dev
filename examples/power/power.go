// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// power demonstrates using the PowerSupply type.
//
// The program should be run from the command line after attaching a device
// to the ev3. Invoke the command with an optional driver name to see the
// power supply stats.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ev3go/ev3dev"
)

func main() {
	driver := flag.String("driver", "", "specify the sensor driver name (required)")
	flag.Parse()

	p := ev3dev.PowerSupply(*driver)
	p = ev3dev.PowerSupply(p.String()) // Cache the driver name if not given.

	v, err := p.Voltage()
	if err != nil {
		log.Fatalf("could not read voltage: %v", err)
	}

	i, err := p.Current()
	if err != nil {
		log.Fatalf("could not read current: %v", err)
	}

	vMax, err := p.VoltageMax()
	if err != nil {
		log.Fatalf("could not read max design voltage: %v", err)
	}

	vMin, err := p.VoltageMin()
	if err != nil {
		log.Fatalf("could not read min design voltage: %v", err)
	}

	fmt.Printf("current power stats: V=%.2fV I=%.0fmA P=%.3fW (designed voltage range:%.2fV-%.2fV)\n", v, i, i*v/1000, vMin, vMax)
}
