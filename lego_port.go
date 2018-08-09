// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var _ idSetter = (*LegoPort)(nil)

// Path returns the lego-port sysfs path.
func (*LegoPort) Path() string { return filepath.Join(prefix, LegoPortPath) }

// Type returns "port".
func (*LegoPort) Type() string { return portPrefix }

// LegoPort represents a handle to a lego-port.
type LegoPort struct {
	id int

	// Cached values:
	modes []string

	// Mode cached value:
	mode string

	// Device cached value:
	driver string

	err error
}

// String satisfies the fmt.Stringer interface.
func (p *LegoPort) String() string {
	if p == nil {
		return portPrefix + "*"
	}
	return portPrefix + strconv.Itoa(p.id)
}

// Err returns the error state of the LegoPort and clears it.
func (p *LegoPort) Err() error {
	err := p.err
	p.err = nil
	return err
}

// idInt and setID satisfy the idSetter interface.
func (p *LegoPort) setID(id int) error {
	t := LegoPort{id: id}
	var err error
	t.modes, err = stringSliceFrom(attributeOf(&t, modes))
	if err != nil {
		goto fail
	}
	t.mode, err = stringFrom(attributeOf(&t, mode))
	if err != nil {
		goto fail
	}
	t.driver, err = DriverFor(&t)
	if err != nil {
		goto fail
	}
	*p = t
	return nil

fail:
	*p = LegoPort{id: -1}
	return err
}
func (p *LegoPort) idInt() int {
	if p == nil {
		return -1
	}
	return p.id
}

// LegoPortFor returns a LegoPort for the given ev3 port name and driver. If the
// lego-port driver does not match the driver string, a LegoPort for the port
// is returned with a DriverMismatch error.
// If port is empty, the first port satisfying the driver name is returned.
func LegoPortFor(port, driver string) (*LegoPort, error) {
	id, err := deviceIDFor(port, driver, (*LegoPort)(nil), -1)
	if id == -1 {
		return nil, err
	}
	var p LegoPort
	_err := p.setID(id)
	if _err != nil {
		err = _err
	}
	return &p, err
}

// Next returns a LegoPort for the next port with the same device driver as
// the receiver.
func (p *LegoPort) Next() (*LegoPort, error) {
	driver, err := DriverFor(p)
	if err != nil {
		return nil, err
	}
	id, err := deviceIDFor("", driver, (*LegoPort)(nil), p.id)
	if id == -1 {
		return nil, err
	}
	return &LegoPort{id: id}, err
}

// Driver returns the driver used by the LegoPort.
func (p *LegoPort) Driver() string {
	return p.driver
}

// Modes returns the available modes for the LegoPort.
func (p *LegoPort) Modes() []string {
	if p.modes == nil {
		return nil
	}
	// Return a copy to prevent users
	// changing the values under our feet.
	avail := make([]string, len(p.modes))
	copy(avail, p.modes)
	return avail
}

// Mode returns the currently selected mode of the LegoPort.
func (p *LegoPort) Mode() string {
	return p.mode
}

// SetMode sets the mode of the LegoPort.
func (p *LegoPort) SetMode(m string) *LegoPort {
	if p.err != nil {
		return p
	}
	ok := false
	for _, a := range p.modes {
		if a == m {
			ok = true
			break
		}
	}
	if !ok {
		p.err = newInvalidValueError(p, mode, "", m, p.Modes())
		return p
	}
	p.err = setAttributeOf(p, mode, m)
	if p.err == nil {
		p.mode, p.err = stringFrom(attributeOf(p, mode))
	}
	return p
}

// SetDevice sets the device of the LegoPort.
func (p *LegoPort) SetDevice(d string) *LegoPort {
	if p.err != nil {
		return p
	}
	p.err = setAttributeOf(p, setDevice, d)
	if p.err == nil {
		p.driver, p.err = DriverFor(p)
	}
	return p
}

// Status returns the current status of the LegoPort.
func (p *LegoPort) Status() (string, error) {
	return stringFrom(attributeOf(p, status))
}

// Uevent returns the current uevent state for the LegoPort.
func (p *LegoPort) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(p, uevent))
}

// ConnectedTo returns a description of the device attached to p in the form
// CONNECTION:PORT:DEVICE where the connection is the underlying transport
// used by the port and is in {"spi0.1", "serial0-0", "ev3-ports", "evb-ports",
// "pistorms"} and the port is the name physically printed on the device — with
// the prefix "in" and "out" for EV3.
func ConnectedTo(p *LegoPort) (string, error) {
	if p.id < 0 {
		return "", newIDErrorFor(p, p.id)
	}
	f, err := os.Open(filepath.Join(p.Path(), p.String()))
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
		case strings.Contains(n, "-ports:"):
			suff := n[strings.Index(n, ":")+1:]
			switch {
			case strings.HasPrefix(suff, "in"):
				if len(suff) >= 4 && suff[3] == ':' && '1' <= suff[2] && suff[2] <= '4' {
					return n, nil
				}
			case strings.HasPrefix(suff, "out"):
				if len(suff) >= 5 && suff[4] == ':' && 'A' <= suff[3] && suff[3] <= 'D' {
					return n, nil
				}
			}
		case strings.Contains(n, "spi0.1:"):
			suff := n[strings.Index(n, ":")+1:]
			switch {
			case strings.HasPrefix(suff, "S"):
				if len(suff) >= 3 && suff[2] == ':' && '1' <= suff[1] && suff[1] <= '9' {
					return n, nil
				}
			case strings.HasPrefix(suff, "M"):
				if len(suff) >= 3 && suff[2] == ':' && 'A' <= suff[1] && suff[1] <= 'P' {
					return n, nil
				}
			}
		}
	}
	return "", nil
}
