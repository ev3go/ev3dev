// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fb

import (
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestXRGB(t *testing.T) {
	for _, test := range testImages {
		golden := filepath.FromSlash("testdata/" + test + "-xrgb.png")

		src, err := decodeImage(filepath.FromSlash("testdata/" + test + ".png"))
		if err != nil {
			t.Fatalf("failed to read src image file %v.png: %v", test, err)
		}

		got := NewXRGB(src.Bounds())
		draw.Draw(got, got.Bounds(), src, src.Bounds().Min, draw.Src)

		if *genGolden {
			f, err := os.Create(golden)
			if err != nil {
				t.Fatalf("failed to create golden image file %v-xrgb.png: %v", test, err)
			}
			defer f.Close()
			err = png.Encode(f, got)
			if err != nil {
				t.Fatalf("failed to encode golden image %v-xrgb.png: %v", test, err)
			}
			continue
		}

		gol, err := decodeImage(golden)
		if err != nil {
			t.Fatalf("failed to read golden image file %v-xrgb.png: %v", test, err)
		}
		want := NewXRGB(gol.Bounds())
		draw.Draw(want, want.Bounds(), gol, gol.Bounds().Min, draw.Src)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("XRGB from source does not match expected image for %v test", test)
		}
	}
}
