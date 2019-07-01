package transform

import (
	"github.com/strickyak/kiwisdr-qrss-grabber-client/font5x7"

	"image"
	"image/color"
	"log"
)

var _ = log.Printf

// const SampPerSec = 12000
// const TargetBW = 300
// const ImageHeight = 600
// // const FFTSize = (1 << 15)
// const FFTSize = (1 << 8)

func BuildImage(in <-chan []uint16, imprint string) image.Image {
	count := 0
	var sum uint64
	var min, max uint16
	var cols [][]uint16
	for arr := range in {
		for _, e := range arr {
			count++
			sum += uint64(e)
			if e < min || min == 0 {
				min = e
			}
			if e > max {
				max = e
			}
		}
		cols = append(cols, arr)
	}
	mean := float64(sum) / float64(count)
	log.Printf("BuildImage: min=%d mean=%.1f max=%d", min, mean, max)

	fakemin := uint16(int((float64(min) + mean) / 2)) // fake min is halfway between min & mean.

	span := float64(max - fakemin)
	scale := 0xFFFF / span

	wid := len(cols)
	hei := len(cols[0])
	log.Printf("BuildImage: wid=%d hei=%d", wid, hei)
	bounds := image.Rectangle{image.Point{0, 0}, image.Point{wid, hei}}
	img := image.NewRGBA64(bounds)
	for x, col := range cols {
		for y, e := range col {
			val := scale * float64(e-fakemin)
			if val < 0 {
				val = 0
			}
			if val > 0xFFFF {
				val = 0xFFFF
			}
			c := uint16(int(val))
			img.SetRGBA64(x, hei-1-y, color.RGBA64{c, c, c, 0xFFFF})
		}
	}
	green := color.RGBA64{0, 0xFFFF, 0, 0xFFFF}
	for i, r := range imprint {
		if r > 255 {
			r = 255
		}
		for x := 0; x < 5; x++ {
			for y := 0; y < 8; y++ {
				if font5x7.Pixel(byte(r), y, x) {
					img.SetRGBA64(10+i*10+x, 10+y, green)
				}
			}
		}
	}

	return img
}
