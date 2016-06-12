// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"sync"
	"syscall"
)

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

// NewFrameBuffer returns an uninitialized FrameBuffer using the
// device at path, a frame buffer that is w by h and with the given
// stride.
// The new function is a callback that constructs an appropriate
// draw.Image for the frame buffer bytes.
func NewFrameBuffer(path string, new func(buf []byte, rect image.Rectangle, stride int) (draw.Image, error), w, h, stride int) FrameBuffer {
	return &lcd{path: path, new: new, w: w, h: h, stride: stride}
}

// lcd is a reader/writer locked draw.Image.
type lcd struct {
	path   string
	w, h   int
	stride int
	new    func([]byte, image.Rectangle, int) (draw.Image, error)

	mu    sync.RWMutex
	img   draw.Image
	f     *os.File
	fbdev []byte
}

func (p *lcd) Init(zero bool) error {
	p.mu.RLock()
	if p.f == nil {
		defer p.mu.Unlock()
		p.mu.RUnlock()
		p.mu.Lock()
		return p.frameBuffer(zero)
	}
	p.mu.RUnlock()
	if zero {
		p.mu.Lock()
		for i := range p.fbdev {
			p.fbdev[i] = 0
		}
		p.mu.Unlock()
	}
	return nil
}

func (p *lcd) frameBuffer(zero bool) error {
	var err error
	p.f, err = os.OpenFile(p.path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	p.fbdev, err = syscall.Mmap(int(p.f.Fd()), 0, p.h*p.stride, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		p.f.Close()
		return err
	}
	if zero {
		for i := range p.fbdev {
			p.fbdev[i] = 0
		}
	}
	p.img, err = p.new(p.fbdev, image.Rect(0, 0, p.w, p.h), p.stride)
	return err
}

func (p *lcd) Close() (err error) {
	defer func() {
		p.mu.Unlock()
		_err := p.f.Close()
		p.f = nil
		if err == nil {
			err = _err
		}
	}()
	p.mu.Lock()
	return syscall.Munmap(p.fbdev)
}

func (p *lcd) ColorModel() color.Model { return p.img.ColorModel() }
func (p *lcd) Bounds() image.Rectangle { return p.img.Bounds() }
func (p *lcd) At(x, y int) color.Color {
	defer p.mu.RUnlock()
	p.mu.RLock()
	if p.f == nil {
		return nil
	}
	return p.img.At(x, y)
}
func (p *lcd) Set(x, y int, c color.Color) {
	p.mu.RLock()
	if p.f == nil {
		p.mu.RUnlock()
		return
	}
	p.mu.RUnlock()
	p.mu.Lock()
	p.img.Set(x, y, c)
	p.mu.Unlock()
}
