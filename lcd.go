// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import "image/draw"

// FrameBuffer is the linux frame buffer image interface.
type FrameBuffer interface {
	draw.Image

	// Init initializes the frame buffer. If zero
	// is true the frame buffer is zeroed. It is
	// safe to call Init on an already initialized
	// FrameBuffer.
	Init(zero bool) error

	// Close closes the backing file. The FrameBuffer
	// is not usable after a call to Close without a
	// following call to Init.
	Close() error
}
