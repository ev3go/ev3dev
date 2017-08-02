// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

var st = testStack(3)

func testStack(n int) stack {
	if n == 0 {
		return callers()
	}
	return testStack(n - 1)
}

var mockValueError = newInvalidValueError(mockDevice{}, "attr", "", "invalid", []string{"ok", "valid"})
