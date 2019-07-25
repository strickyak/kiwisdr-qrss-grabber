package transform

import (
	"github.com/strickyak/kiwisdr-qrss-grabber/font5x7"

	"image"
	"image/color"
	"log"
	"math"
)

var _ = log.Printf

func BuildImage(in <-chan []float64, imprint string) image.Image {
	var count float64
	var sum float64
	var sumsq float64
	var min, max float64
	var cols [][]float64
	for arr := range in {
		for _, e := range arr {
			count++
			sum += float64(e)
			sumsq += float64(e * e)
			if e < min || min == 0 {
				min = e
			}
			if e > max {
				max = e
			}
		}
		cols = append(cols, arr)
	}
	mean := sum / count
	log.Printf("BuildImage: min=%.1f mean=%.1f max=%.1f", min, mean, max)

	naïve_variance := ((sumsq) - ((sum) * (sum) / (count))) / (count - 1)
	std_dev := math.Sqrt(naïve_variance)
	leap := 1.0 * std_dev

	min1 := (sum)/(count) + leap
	max1 := min1 + leap
	min2 := max1
	max2 := min2 + leap
	min3 := max2
	max3 := min3 + leap
	_ = max3

	span := (max1 - min1)
	scale := 0xFFFF / span

	wid := len(cols)
	hei := len(cols[0])
	log.Printf("BuildImage: wid=%d hei=%d", wid, hei)
	bounds := image.Rectangle{image.Point{0, 0}, image.Point{wid, hei}}
	img := image.NewRGBA64(bounds)
	for x, col := range cols {
		for y, e := range col {
			// First fill green.
			green := scale * ((e) - min1)
			if green < 0 {
				green = 0
			}
			if green > 0xFFFF {
				green = 0xFFFF
			}
			// Then fill red, which with green makes yellow.
			red := scale * ((e) - min2)
			if red < 0 {
				red = 0
			}
			if red > 0xFFFF {
				red = 0xFFFF
			}
			// Then fill blue, which with green and red makes white.
			blue := scale / 3 * ((e) - min3)
			if blue < 0 {
				blue = 0
			}
			if blue > 0xFFFF {
				blue = 0xFFFF
			}
			img.SetRGBA64(x, hei-1-y, color.RGBA64{uint16(int(red)), uint16(int(green)), uint16(int(blue)), 0xFFFF})
		}
	}
	// ink := color.RGBA64{0x2222, 0xFFFF, 0x2222, 0xFFFF} // Green
	// ink := color.RGBA64{0xE600, 0x4C00, 0xE600, 0xFFFF}  // purple
	ink := color.RGBA64{0xFFFF, 0x0000, 0xFFFF, 0xFFFF} // magenta
	// ink := color.RGBA64{0xe666, 0x6666, 0x3333, 0xFFFF}  // Burnt Orange
	// ink := color.RGBA64{0x3333, 0x3333, 0xffff, 0xFFFF}  // Blue
	const N = 2 // screen pixels per chargen pixel
	for i, r := range imprint {
		if r > 255 {
			r = 255
		}
		for x := 0; x < 5; x++ {
			for y := 0; y < 8; y++ {
				if font5x7.Pixel(byte(r), y, x) {
					for a := 0; a < N; a++ {
						for b := 0; b < N; b++ {
							img.SetRGBA64(a+N*(10-y), b+4+i*7*N+x*N, ink)
						}
					}
				}
			}
		}
	}

	return img
}
