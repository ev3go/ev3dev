// Copyright ©2017 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gps demonstrates use of the Dexter Industries dGPS device.
// It displays the time in UTC, the number of satellites in view
// and the HDOP. It also shows the current location, velocity and
// heading, and the distance and heading to a notable location.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ev3go/ev3dev/otheri2c"
)

func main() {
	var port = flag.String("port", "", "specify the sensor port the GPS is connected to")
	flag.Parse()
	if *port == "" {
		flag.Usage()
		os.Exit(2)
	}

	gps, err := otheri2c.OpenGPS(*port)
	if err != nil {
		log.Fatalf("failed to open GPS device: %v", err)
	}
	defer gps.Close()

	b, err := gps.ExtendedFirmware(true)
	if err != nil {
		log.Fatalf("error selecting extended firmware: %v", err)
	}
	fmt.Printf("GPS-X response: %v\n", b)

	stat, err := gps.Status()
	if err != nil {
		log.Fatalf("error getting satellite link status: %v", err)
	}
	fmt.Printf("satellite link ok=%t\n", stat)

	sats, err := gps.SatellitesInView()
	if err != nil {
		log.Fatalf("error getting satellites in view: %v", err)
	}
	fmt.Printf("satellites in view: %d\n", sats)

	hdop, err := gps.HDOP()
	if err != nil {
		log.Fatalf("error getting HDOP: %v", err)
	}
	fmt.Printf("satellite HDOP=%d\n", hdop)

	t, err := gps.UTC()
	if err != nil {
		log.Fatalf("error getting satellite time: %v", err)
	}
	fmt.Printf("satellite time=%v\n", t)

	lat, err := gps.Latitude()
	if err != nil {
		log.Fatalf("error getting latitude: %v", err)
	}
	lon, err := gps.Longitude()
	if err != nil {
		log.Fatalf("error getting longitude: %v", err)
	}
	alt, err := gps.Altitude()
	if err != nil {
		log.Fatalf("error getting altitude: %v", err)
	}
	vel, err := gps.Velocity()
	if err != nil {
		log.Fatalf("error getting velocity: %v", err)
	}
	head, err := gps.Heading()
	if err != nil {
		log.Fatalf("error getting heading: %v", err)
	}
	fmt.Printf("lat=%d.%d° lon=%d.%d° alt=%dm vel=%dcm/s head=%d°\n",
		lat/1e6, abs(lat%1e6), lon/1e6, abs(lon%1e6), alt, vel, head)

	err = gps.SetDestLatitude(55730966)
	if err != nil {
		log.Fatalf("error setting destination latitude: %v", err)
	}
	err = gps.SetDestLongitude(9010570)
	if err != nil {
		log.Fatalf("error setting longitude: %v", err)
	}

	dist, err := gps.DistanceToDest()
	if err != nil {
		log.Fatalf("error getting distance to destination: %v", err)
	}
	angle, err := gps.AngleToDest()
	if err != nil {
		log.Fatalf("error getting angle to destination: %v", err)
	}
	lastAngle, err := gps.AngleSinceLast()
	if err != nil {
		log.Fatalf("error getting angle since last: %v", err)
	}
	fmt.Printf("destination=%dm %d° traveled=%d°\n",
		dist, angle, lastAngle)
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
