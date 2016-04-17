// Copyright ©2016 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"os"
	"sync"
	"syscall"
)

const (
	// LCDWidth is the width of the LCD screen in pixel.
	LCDWidth = 178

	// LCDHeight is the height of the LCD screen in pixel.
	LCDHeight = 128

	// LCDStride is the width of the LCD screen memory in bytes.
	LCDStride = 24
)

// LCD is the draw image used draw directly to the ev3 LCD screen.
// Drawing operations are safe for concurrent use, but are not atomic
// beyond the pixel level.
var LCD draw.Image = frameBuffer("/dev/fb0")

func frameBuffer(path string) draw.Image {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	fbdev, err := syscall.Mmap(int(f.Fd()), 0, LCDHeight*LCDStride, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}
	for i := 0; i < LCDHeight*LCDStride; i++ {
		fbdev[i] = 0
	}
	fb0, err := newMonochromeWith(fbdev, image.Rect(0, 0, LCDWidth, LCDHeight), LCDStride)
	if err != nil {
		panic(err)
	}
	return &lcd{img: fb0}
}

// lcd is a reader/writer locked draw.Image.
type lcd struct {
	m   sync.RWMutex
	img draw.Image
}

func (p *lcd) ColorModel() color.Model { return p.img.ColorModel() }
func (p *lcd) Bounds() image.Rectangle { return p.img.Bounds() }
func (p *lcd) At(x, y int) color.Color {
	defer p.m.RUnlock()
	p.m.RLock()
	return p.img.At(x, y)
}
func (p *lcd) Set(x, y int, c color.Color) {
	p.m.Lock()
	p.img.Set(x, y, c)
	p.m.Unlock()
}

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

// newMonochromeWith returns a new Monochrome image with the given bounds
// and stride, backed by the []byte, pix. If stride is zero, a working
// stride is computed.
func newMonochromeWith(pix []byte, r image.Rectangle, stride int) (*Monochrome, error) {
	w, h := r.Dx(), r.Dy()
	if stride == 0 {
		stride = (w + 7) / 8
	}
	if len(pix) != stride*h {
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
