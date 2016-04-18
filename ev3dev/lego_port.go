// Copyright ©2016 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Path returns the lego-port sysfs path.
func (*LegoPort) Path() string { return LegoPortPath }

// Path returns "port".
func (*LegoPort) Type() string { return portPrefix }

// LegoPort represents a handle to a lego-port.
type LegoPort struct {
	id int
}

// String satisfies the fmt.Stringer interface.
func (p *LegoPort) String() string { return fmt.Sprint(portPrefix, p.id) }

// LegoPortFor returns a LegoPort for the given ev3 port name and driver. If the
// lego-port driver does not match the driver string, a LegoPort for the port
// is returned with a DriverMismatch error.
// If port is empty, the first port satisfying the driver name is returned.
func LegoPortFor(port, driver string) (*LegoPort, error) {
	id, err := deviceIDFor(port, driver, (*LegoPort)(nil))
	if id == -1 {
		return nil, err
	}
	return &LegoPort{id: id}, err
}

func (p *LegoPort) writeFile(path, data string) error {
	return ioutil.WriteFile(path, []byte(data), 0)
}

// Modes returns the available modes for the LegoPort.
func (p *LegoPort) Modes() ([]string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(LegoPortPath+"/%s/"+modes, p))
	if err != nil {
		return nil, fmt.Errorf("ev3dev: failed to read port modes: %v", err)
	}
	return strings.Split(string(chomp(b)), " "), err
}

// Mode returns the currently selected mode of the LegoPort.
func (p *LegoPort) Mode() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(LegoPortPath+"/%s/"+mode, p))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port mode: %v", err)
	}
	return string(chomp(b)), err
}

// SetMode sets the mode of the LegoPort.
func (p *LegoPort) SetMode(mode string) error {
	err := p.writeFile(fmt.Sprintf(LegoPortPath+"/%s/"+mode, p), mode)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set port mode: %v", err)
	}
	return nil
}

// SetDevice sets the device of the LegoPort.
func (p *LegoPort) SetDevice(dev string) error {
	err := p.writeFile(fmt.Sprintf(LegoPortPath+"/%s/"+setDevice, p), dev)
	if err != nil {
		return fmt.Errorf("ev3dev: failed to set port device: %v", err)
	}
	return nil
}

// Status returns the current status of the LegoPort.
func (p *LegoPort) Status() (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf(LegoPortPath+"/%s/"+status, p))
	if err != nil {
		return "", fmt.Errorf("ev3dev: failed to read port status: %v", err)
	}
	return string(chomp(b)), err
}

// ConnectedTo returns a description of the device attached to p in the form
// {inX,outY}:DEVICE-NAME, where X is in {1-4} and Y is in {A-D}.
func ConnectedTo(p *LegoPort) (string, error) {
	if p.id < 0 {
		return "", fmt.Errorf("ev3dev: invalid lego port number: %d", p.id)
	}
	f, err := os.Open(fmt.Sprintf(LegoPortPath+"/%s", p))
	if err != nil {
		return "", err
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return "", err
	}
	for _, n := range names {
		switch {
		case strings.HasPrefix(n, "in"):
			if len(n) >= 4 && n[3] == ':' && '1' <= n[2] && n[2] <= '4' {
				return n, nil
			}
		case strings.HasPrefix(n, "out"):
			if len(n) >= 5 && n[4] == ':' && 'A' <= n[3] && n[3] <= 'D' {
				return n, nil
			}
		}
	}
	return "", nil
}
