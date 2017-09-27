// Copyright ©2017 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

	b, err := gps.ExtendedFirmware(true)
	if err != nil {
		log.Fatalf("error selecting extended firmware: %v", err)
	}
	fmt.Printf("GPS-X response: %v\n", b)

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
	fmt.Printf("%d° %d° %dm %dcm/s %d°\n", lat, lon, alt, vel, head)

	err = gps.SetDestLatitude(0)
	if err != nil {
		log.Fatalf("error setting destination latitude: %v", err)
	}
	err = gps.SetDestLongitude(0)
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
	fmt.Printf("%dm %d° %d°\n", dist, angle, lastAngle)

}
