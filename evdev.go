// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import "syscall"

// Constants from uapi/asm-generic/ioctl.h and uapi/linux/input.h.
const (
	_ioc_read = 2

	_ioc_nrbits   = 8
	_ioc_typebits = 8
	_ioc_sizebits = 14
	_ioc_dirbits  = 2

	_ioc_nrmask   = 1<<_ioc_nrbits - 1
	_ioc_typemask = 1<<_ioc_typebits - 1
	_ioc_sizemask = 1<<_ioc_sizebits - 1
	_ioc_dirmask  = 1<<_ioc_dirbits - 1

	// Calculate shifts for _ioc fields.
	_ioc_nrshift   = 0                              // bits  0- 7
	_ioc_typeshift = _ioc_nrshift + _ioc_nrbits     // bits  8-15
	_ioc_sizeshift = _ioc_typeshift + _ioc_typebits // bits 16-29
	_ioc_dirshift  = _ioc_sizeshift + _ioc_sizebits // bits 30-31
)

func eviocgbit(ev byte, bits uint16) uintptr {
	return _ioc_read<<_ioc_dirshift | uintptr(bits)<<_ioc_sizeshift | 'E'<<_ioc_typeshift | (0x20+uintptr(ev))<<_ioc_nrshift
}

func eviocgkey(buf []byte) uintptr {
	return _ioc_read<<_ioc_dirshift | uintptr(len(buf))<<_ioc_sizeshift | 'E'<<_ioc_typeshift | 0x18<<_ioc_nrshift
}

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if errno != 0 {
		return errno
	}
	return nil
}

func isSet(bit uint, buf []byte) bool {
	return buf[bit>>3]&(1<<(bit&7)) != 0
}
