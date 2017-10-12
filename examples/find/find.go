// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// find demonstrates finding the first available sensor for a driver name.
//
// The program should be run from the command line after attaching a device
// to the ev3. Invoke the command with the driver name to see the sensor
// values.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/ev3go/ev3dev"
)

func main() {
	driver := flag.String("driver", "", "specify the sensor driver name (required)")
	flag.Parse()
	if *driver == "" {
		flag.Usage()
		os.Exit(1)
	}

	s := &ev3dev.Sensor{}

	err := ev3dev.FindAfter(nil, s, *driver)
	if err != nil {
		log.Fatalf("failed to find sensor: %v", err)
	}

	n := s.NumValues()
	u := s.Units()
	d := s.Decimals()

	addr, err := ev3dev.AddressOf(s)
	if err != nil {
		log.Fatalf("failed to get address of %s device: %v", *driver, err)
	}

	fmt.Printf("%s sensor device in %s port\n", *driver, addr)

	for i := 0; i < n; i++ {
		v, err := s.Value(i)
		if err != nil {
			log.Fatalf("failed to get of value %d: %v", i, err)
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Fatalf("failed to parse value: %v", err)
		}
		fmt.Printf("value%d = %v %s\n", i, f/math.Pow10(d), u)
	}
}
