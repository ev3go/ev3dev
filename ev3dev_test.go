// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev_test

import (
	"fmt"
	"io"
	"testing"
	"time"

	"bazil.org/fuse"

	"github.com/ev3go/ev3dev"
	"github.com/ev3go/sisyphus"
)

var (
	epoch time.Time
	clock func() time.Time
)

func init() {
	loc, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		panic(err)
	}
	epoch = time.Date(2013, time.September, 1, 0, 0, 0, 0, loc)
	clock = func() time.Time { return epoch }
}

var (
	d  = sisyphus.MustNewDir
	ro = sisyphus.MustNewRO
	rw = sisyphus.MustNewRW
	wo = sisyphus.MustNewWO
)

func readAt(b []byte, offset int64, val interface{}) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	s := fmt.Sprintln(val)
	if offset >= int64(len(s)) {
		return 0, io.EOF
	}
	n := copy(b, s[offset:])
	if n <= len(b) {
		return n, io.EOF
	}
	return n, nil
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func size(val interface{}) int64 {
	return int64(len(fmt.Sprintln(val)))
}

func chomp(b []byte) []byte {
	if b[len(b)-1] == '\n' {
		return b[:len(b)-1]
	}
	return b
}

func serve(fs *sisyphus.FileSystem, t *testing.T) (unmount func()) {
	c, err := sisyphus.Serve(ev3dev.Prefix, fs, nil, fuse.AllowNonEmptyMount())
	if err != nil {
		t.Fatalf("failed to open server: %v", err)
	}
	return func() {
		// Allow some time for the
		// server to be ready to close.
		time.Sleep(time.Second)

		err = c.Close()
		if err != nil {
			t.Errorf("failed to close server: %v", err)
		}
	}
}
