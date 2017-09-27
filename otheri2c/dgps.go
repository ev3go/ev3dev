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
	dGPS_I2C_addr = 0x06 // dGPS Sensor I2C Address

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
	dGPS_UTC:              {sendSize: 3, recvSize: 4},
	dGPS_Status:           {sendSize: 3, recvSize: 1},
	dGPS_Latitude:         {sendSize: 3, recvSize: 4},
	dGPS_Longitude:        {sendSize: 3, recvSize: 4},
	dGPS_Velocity:         {sendSize: 3, recvSize: 3},
	dGPS_Heading:          {sendSize: 3, recvSize: 2},
	dGPS_DistanceToDest:   {sendSize: 3, recvSize: 4},
	dGPS_AngleToDest:      {sendSize: 3, recvSize: 2},
	dGPS_AngleSinceLast:   {sendSize: 3, recvSize: 2},
	dGPS_SetDestLatitude:  {sendSize: 7, recvSize: 0},
	dGPS_SetDestLongitude: {sendSize: 7, recvSize: 0},
	dGPS_ExtendedFirmware: {sendSize: 4, recvSize: 3},
	dGPS_Altitude:         {sendSize: 3, recvSize: 4},
	dGPS_HDOP:             {sendSize: 3, recvSize: 4},
	dGPS_SatellitesInView: {sendSize: 3, recvSize: 4},
}

type GPS struct {
	dev  *sysfs.I2C
	send [7]byte
	recv [4]byte
}

func OpenGPS(port string) (*GPS, error) {
	number, err := i2cDeviceNumberFor("i2c-" + port)
	if err != nil {
		return nil, err
	}
	d, err := sysfs.NewI2C(number)
	if err != nil {
		return nil, err
	}
	return &GPS{dev: d, send: [7]byte{1: 3}}, nil
}

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

func (d *GPS) Close() error {
	if d.dev == nil {
		return nil
	}
	err := d.dev.Close()
	d.dev = nil
	return err
}

func (d *GPS) tx(request byte) ([]byte, error) {
	time.Sleep(20 * time.Millisecond)
	c := dGPS_CommandLookup[request]
	d.send[0] = c.sendSize
	d.send[2] = request
	err := d.dev.Tx(dGPS_I2C_addr, d.send[:c.sendSize], d.recv[:c.recvSize])
	if err != nil {
		return nil, err
	}
	return d.recv[:c.recvSize], nil
}

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

func (d *GPS) Status() (ok bool, err error) {
	b, err := d.tx(dGPS_Status)
	if err != nil {
		return false, err
	}
	return b[0] == 1, nil
}

func (d *GPS) Latitude() (int, error) {
	b, err := d.tx(dGPS_Latitude)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

func (d *GPS) Longitude() (int, error) {
	b, err := d.tx(dGPS_Longitude)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

func (d *GPS) Altitude() (int, error) {
	b, err := d.tx(dGPS_Altitude)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

func (d *GPS) Velocity() (int, error) {
	b, err := d.tx(dGPS_Velocity)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b[:4]) >> 8), nil
}

func (d *GPS) Heading() (int, error) {
	b, err := d.tx(dGPS_Heading)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint16(b)), nil
}

func (d *GPS) DistanceToDest() (int, error) {
	b, err := d.tx(dGPS_DistanceToDest)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

func (d *GPS) AngleToDest() (int, error) {
	b, err := d.tx(dGPS_AngleToDest)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint16(b)), nil
}

func (d *GPS) AngleSinceLast() (int, error) {
	b, err := d.tx(dGPS_AngleSinceLast)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint16(b)), nil
}

func (d *GPS) HDOP() (int, error) {
	b, err := d.tx(dGPS_HDOP)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

func (d *GPS) SatellitesInView() (int, error) {
	b, err := d.tx(dGPS_SatellitesInView)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

func (d *GPS) ExtendedFirmware(use bool) ([]byte, error) {
	if use {
		d.send[3] = 1
	} else {
		d.send[3] = 0
	}
	return d.tx(dGPS_ExtendedFirmware)
}

func (d *GPS) SetDestLatitude(lat int) error {
	c := dGPS_CommandLookup[dGPS_SetDestLatitude]
	binary.BigEndian.PutUint32(d.send[3:c.sendSize], uint32(lat))
	_, err := d.tx(dGPS_SetDestLatitude)
	return err
}

func (d *GPS) SetDestLongitude(lon int) error {
	c := dGPS_CommandLookup[dGPS_SetDestLongitude]
	binary.BigEndian.PutUint32(d.send[3:c.sendSize], uint32(lon))
	_, err := d.tx(dGPS_SetDestLongitude)
	return err
}
