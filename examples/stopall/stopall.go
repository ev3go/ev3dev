// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// stopall stops all motors.
package main

import (
	"log"

	"github.com/ev3go/ev3dev/motorutil"
)

func main() {
	err := motorutil.ResetAll()
	if err != nil {
		log.Fatal(err)
	}
}
