// Copyright ©2017 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package otheri2c

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"periph.io/x/periph/host/sysfs"
)

const (
	dGPS_I2C_addr = 0x03 // dGPS Sensor I2C Address

	// From https://www.dexterindustries.com/manual/dgps-2/
	dGPS_UTC              = 0x00 // Fetch UTC
	dGPS_Status           = 0x01 // Status of satellite link: 0 no link, 1 link
	dGPS_Latitude         = 0x02 // Fetch latitude
	dGPS_Longitude        = 0x04 // Fetch longitude
	dGPS_Velocity         = 0x06 // Fetch velocity (cm/s)
	dGPS_Heading          = 0x07 // Fetch heading (degrees)
	dGPS_DistanceToDest   = 0x08 // Fetch distance to destination (m)
	dGPS_AngleToDest      = 0x09 // Fetch angle to destination (degrees)
	dGPS_AngleSinceLast   = 0x0a // Fetch angle travelled since last request
	dGPS_SetDestLatitude  = 0x0b // Set latitude of destination
	dGPS_SetDestLongitude = 0x0c // Set longitude of destination
	dGPS_ExtendedFirmware = 0x0d // Extended firmware
	dGPS_Altitude         = 0x0e // Altitude (m)
	dGPS_HDOP             = 0x0f // HDOP
	dGPS_SatellitesInView = 0x10 // Satellites in view
)

type dGPS_Command struct {
	sendSize, recvSize byte
}

// From https://www.dexterindustries.com/manual/dgps-2/
var dGPS_CommandLookup = [...]dGPS_Command{
	dGPS_UTC:              {sendSize: 1, recvSize: 4},
	dGPS_Status:           {sendSize: 1, recvSize: 1},
	dGPS_Latitude:         {sendSize: 1, recvSize: 4},
	dGPS_Longitude:        {sendSize: 1, recvSize: 4},
	dGPS_Velocity:         {sendSize: 1, recvSize: 3},
	dGPS_Heading:          {sendSize: 1, recvSize: 2},
	dGPS_DistanceToDest:   {sendSize: 1, recvSize: 4},
	dGPS_AngleToDest:      {sendSize: 1, recvSize: 2},
	dGPS_AngleSinceLast:   {sendSize: 1, recvSize: 2},
	dGPS_SetDestLatitude:  {sendSize: 5, recvSize: 0},
	dGPS_SetDestLongitude: {sendSize: 5, recvSize: 0},
	dGPS_ExtendedFirmware: {sendSize: 2, recvSize: 3},
	dGPS_Altitude:         {sendSize: 1, recvSize: 4},
	dGPS_HDOP:             {sendSize: 1, recvSize: 4},
	dGPS_SatellitesInView: {sendSize: 1, recvSize: 4},
}

// GPS is a handle to a dGPS device.
type GPS struct {
	dev  *sysfs.I2C
	send [5]byte
	recv [4]byte
}

// OpenGPS opens a GPS attached to the LEGO sensort port given. This
// can either be in the form inX for an EV3 input port where X is the
// physical port number, or N where N is the I²C bus number.
//
// The GPS should be closed when it is no longer needed.
func OpenGPS(port string) (*GPS, error) {
	number, err := i2cDeviceNumberFor("i2c-" + port)
	if err != nil {
		return nil, err
	}
	d, err := sysfs.NewI2C(number)
	if err != nil {
		return nil, err
	}
	return &GPS{dev: d}, nil
}

// i2cDeviceNumberFor returns the bus number for the path /dev/port after
// resolving any symlinks.
func i2cDeviceNumberFor(port string) (int, error) {
	dev, err := filepath.EvalSymlinks(filepath.Join("/dev", port))
	if err != nil {
		return -1, err
	}
	if filepath.Dir(dev) != "/dev" {
		return -1, fmt.Errorf("otheri2c: not a device: %q", dev)
	}
	base := filepath.Base(dev)
	if !strings.HasPrefix(base, "i2c-") {
		return -1, fmt.Errorf("otheri2c: not an I²C device: %q", dev)
	}
	return strconv.Atoi(base[len("i2c-"):])
}

// Close closes the device.
func (d *GPS) Close() error {
	if d.dev == nil {
		return nil
	}
	err := d.dev.Close()
	d.dev = nil
	return err
}

// tx performs an I²C message transaction.
func (d *GPS) tx(request byte) ([]byte, error) {
	time.Sleep(200 * time.Millisecond)
	c := dGPS_CommandLookup[request]
	d.send[0] = request
	for i := range &d.recv {
		d.recv[i] = 0
	}
	err := d.dev.Tx(dGPS_I2C_addr, d.send[:c.sendSize], d.recv[:c.recvSize])
	if err != nil {
		return nil, err
	}
	return d.recv[:c.recvSize], nil
}

// UTC returns the time in the UTC time zone obtained from the satellite.
//
// Note that the time returned by the satellite does not include the date,
// so UTC will fabricate a date based on the local time.
func (d *GPS) UTC() (time.Time, error) {
	b, err := d.tx(dGPS_UTC)
	if err != nil {
		return time.Time{}, err
	}
	utc := binary.BigEndian.Uint32(b)
	s := int(utc % 1e2)
	utc /= 1e2
	m := int(utc % 1e2)
	utc /= 1e2
	h := int(utc % 1e2)
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, h, m, s, 0, time.UTC), nil
}

// Status returns whether the GPS values are valid.
func (d *GPS) Status() (ok bool, err error) {
	b, err := d.tx(dGPS_Status)
	if err != nil {
		return false, err
	}
	return b[0] == 1, nil
}

// Latitude returns the current latitude in millionths of a degree.
// A negative latitude indicates southern latitudes.
func (d *GPS) Latitude() (int, error) {
	b, err := d.tx(dGPS_Latitude)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

// Longitude returns the current latitude in millionths of a degree.
// A negative longitude indicates western longitudes.
func (d *GPS) Longitude() (int, error) {
	b, err := d.tx(dGPS_Longitude)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

// Altitude returns the current altitude in meters. This is only valid if
// ExtendedFirmware(true) has been called prior to calling Altitude.
func (d *GPS) Altitude() (int, error) {
	b, err := d.tx(dGPS_Altitude)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

// Velocity returns the current velocity in centimeters.
func (d *GPS) Velocity() (int, error) {
	b, err := d.tx(dGPS_Velocity)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b[:4]) >> 8), nil
}

// Heading returns the current heading in degrees.
func (d *GPS) Heading() (int, error) {
	b, err := d.tx(dGPS_Heading)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint16(b)), nil
}

// DistanceToDest returns the distance to the current destination in meters.
func (d *GPS) DistanceToDest() (int, error) {
	b, err := d.tx(dGPS_DistanceToDest)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

// AngleToDest returns the heading to the current destination in degrees.
func (d *GPS) AngleToDest() (int, error) {
	b, err := d.tx(dGPS_AngleToDest)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint16(b)), nil
}

// AngleSinceLast returns the angle travelled since the last call to AngleSinceLast.
func (d *GPS) AngleSinceLast() (int, error) {
	b, err := d.tx(dGPS_AngleSinceLast)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint16(b)), nil
}

// HDOP returns the measure of the precision that can be expected. This is only
// valid if ExtendedFirmware(true) has been called prior to calling HDOP.
func (d *GPS) HDOP() (int, error) {
	b, err := d.tx(dGPS_HDOP)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

// SatellitesInView returns the number of satellites in view. This is only valid if
// ExtendedFirmware(true) has been called prior to calling SatellitesInView.
func (d *GPS) SatellitesInView() (int, error) {
	b, err := d.tx(dGPS_SatellitesInView)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

// ExtendedFirmware specifies whether to use the dGPS-X firmware extension.
// Turning on the extended firmware can slow down the dGPS sensor.
func (d *GPS) ExtendedFirmware(use bool) ([]byte, error) {
	if use {
		d.send[1] = 1
	} else {
		d.send[1] = 0
	}
	return d.tx(dGPS_ExtendedFirmware)
}

// SetDestLatitude sets the latitude of the destination in millionths of a degree.
// A negative latitude indicates southern latitudes.
func (d *GPS) SetDestLatitude(lat int) error {
	c := dGPS_CommandLookup[dGPS_SetDestLatitude]
	binary.BigEndian.PutUint32(d.send[1:c.sendSize], uint32(lat))
	_, err := d.tx(dGPS_SetDestLatitude)
	return err
}

// SetDestLongitude sets the longitude of the destination in millionths of a degree.
// A negative longitude indicates western longitudes.
func (d *GPS) SetDestLongitude(lon int) error {
	c := dGPS_CommandLookup[dGPS_SetDestLongitude]
	binary.BigEndian.PutUint32(d.send[1:c.sendSize], uint32(lon))
	_, err := d.tx(dGPS_SetDestLongitude)
	return err
}
