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

// NewXRGB returns a new XRGB image with the given bounds.
func NewXRGB(r image.Rectangle) *XRGB {
	w, h := r.Dx(), r.Dy()
	stride := 4 * w
	pix := make([]uint8, stride*h)
	return &XRGB{Pix: pix, Stride: stride, Rect: r}
}

// NewXRGBWith returns a new XRGB image with the given bounds,
// backed by the []byte, pix. If stride is zero, a working stride
// is computed. If the length of pix is less than stride*h, an
// error is returned.
func NewXRGBWith(pix []byte, r image.Rectangle, stride int) (draw.Image, error) {
	w, h := r.Dx(), r.Dy()
	if stride == 0 {
		stride = 4 * w
	}
	if len(pix) < stride*h {
		return nil, errors.New("ev3dev: bad pixel buffer length")
	}
	return &XRGB{Pix: pix, Stride: stride, Rect: r}, nil
}

// XRGB is an in-memory image whose At method returns PixelXRGB values.
type XRGB struct {
	// Pix holds the image's pixels, as 32 bit XRGB
	// values stored in little-endian order.
	// The pixel at (x, y) is the four bytes at
	// Pix[4*(x-Rect.Min.X) + (y-Rect.Min.Y)*Stride].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between
	// vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// ColorModel returns the XRGB color model.
func (p *XRGB) ColorModel() color.Model { return RGB565Model }

// Bounds returns the bounding rectangle for the image.
func (p *XRGB) Bounds() image.Rectangle { return p.Rect }

// At returns the color of the pixel at (x, y).
func (p *XRGB) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.RGBA{}
	}
	i := p.pixOffset(x, y)
	return PixelXRGB{R: p.Pix[i+2], G: p.Pix[i+1], B: p.Pix[i]}
}

// Set sets the color of the pixel at (x, y) to c.
func (p *XRGB) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.pixOffset(x, y)
	xrgb := XRGBModel.Convert(c).(PixelXRGB)
	p.Pix[i+2] = xrgb.R
	p.Pix[i+1] = xrgb.G
	p.Pix[i] = xrgb.B
}

// pixOffset returns the index into p.Pix for the first byte
// containing the pixel at (x, y).
func (p *XRGB) pixOffset(x, y int) int {
	return 4*(x-p.Rect.Min.X) + (y-p.Rect.Min.Y)*p.Stride
}

// PixelXRGB is an XRGB pixel.
type PixelXRGB struct {
	R, G, B byte
}

// RGBA returns the RGBA values for the receiver.
func (c PixelXRGB) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	return r, g, b, 0xffff
}

// XRGBModel is the color model for XRGB images.
var XRGBModel color.Model = color.ModelFunc(xrgbModel)

func xrgbModel(c color.Color) color.Color {
	if _, ok := c.(PixelXRGB); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return PixelXRGB{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8)}
}
