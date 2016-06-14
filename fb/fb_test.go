// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fb

import (
	"flag"
	"image"
	"os"
)

var genGolden = flag.Bool("gen.golden", false, "generate golden image files")

func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

var testImages = []string{
	"gopherbrick",
	"corner",
	"black",
}
