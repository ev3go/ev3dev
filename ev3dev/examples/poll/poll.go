// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// keys demonstrates key polling. It should be run from the command line.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ev3go/ev3/ev3dev"
)

func main() {
	var b ev3dev.ButtonPoller

	for i := 0; i < 30; i++ {
		b, err := b.Poll()
		if err != nil {
			log.Fatalf("failed to poll keys: %v")
		}
		fmt.Printf("%6b\n", b)
		time.Sleep(5 * time.Second)
	}
}
