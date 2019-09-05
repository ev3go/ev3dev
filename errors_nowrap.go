// Copyright ©2019 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !go1.13

// TODO(kortschak): Remove this when go1.12 no longer supported.

package ev3dev

import "strings"

func wrapped(format string) string { return strings.Replace(format, "%w", "%v", 1) }
