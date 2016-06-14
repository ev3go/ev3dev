// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fb

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
)

// NewMonochrome returns a new Monochrome image with the given bounds
// and stride. If stride is zero, a working stride is computed.
func NewMonochrome(r image.Rectangle, stride int) *Monochrome {
	w, h := r.Dx(), r.Dy()
	if stride == 0 {
		stride = (w + 7) / 8
	}
	pix := make([]uint8, stride*h)
	return &Monochrome{Pix: pix, Stride: stride, Rect: r}
}

// NewMonochromeWith returns a new Monochrome image with the given bounds
// and stride, backed by the []byte, pix. If stride is zero, a working
// stride is computed. If the length of pix is less than stride*h, an
// error is returned.
func NewMonochromeWith(pix []byte, r image.Rectangle, stride int) (draw.Image, error) {
	w, h := r.Dx(), r.Dy()
	if stride == 0 {
		stride = (w + 7) / 8
	}
	if len(pix) < stride*h {
		return nil, errors.New("ev3dev: bad pixel buffer length")
	}
	return &Monochrome{Pix: pix, Stride: stride, Rect: r}, nil
}

// Monochrome is an in-memory image whose At method returns Pixel values.
type Monochrome struct {
	// Pix holds the image's pixels, as bit values.
	// The pixel at (x, y) is the x%8^th bit in
	// Pix[(x-Rect.Min.X)/8 + (y-Rect.Min.Y)*Stride].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between
	// vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// ColorModel returns the monochrome color model.
func (p *Monochrome) ColorModel() color.Model { return MonochromeModel }

// Bounds returns the bounding rectangle for the image.
func (p *Monochrome) Bounds() image.Rectangle { return p.Rect }

// At returns the color of the pixel at (x, y).
func (p *Monochrome) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return Pixel(White)
	}
	i := p.pixOffset(x, y)
	return Pixel(p.Pix[i]&(1<<uint(x%8)) != 0)
}

// Set sets the color of the pixel at (x, y) to c.
func (p *Monochrome) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.pixOffset(x, y)
	if MonochromeModel.Convert(c).(Pixel) == Black {
		p.Pix[i] |= 1 << uint(x%8)
	} else {
		p.Pix[i] &^= 1 << uint(x%8)
	}
}

// pixOffset returns the index into p.Pix for the byte
// containing bit describing the pixel at (x, y).
func (p *Monochrome) pixOffset(x, y int) int {
	return (x-p.Rect.Min.X)/8 + (y-p.Rect.Min.Y)*p.Stride
}

// Pixel is a black and white monochrome pixel.
type Pixel bool

const (
	Black Pixel = true
	White Pixel = false
)

// RGBA returns the RGBA values for the receiver.
func (c Pixel) RGBA() (r, g, b, a uint32) {
	if c == Black {
		return 0, 0, 0, 0xffff
	}
	return 0xffff, 0xffff, 0xffff, 0xffff
}

// MonochromeModel is the color model for black and white images.
var MonochromeModel color.Model = color.ModelFunc(monoModel)

func monoModel(c color.Color) color.Color {
	if _, ok := c.(Pixel); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	y := (299*r + 587*g + 114*b + 500) / 1000
	return Pixel(uint16(y) < 0x8000)
}
