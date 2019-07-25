// +build main

// go run render.go -imprint "Q0QQQ on 40m at $(date -u)" < 40m.wav > 40m.png
package main

import (
	"github.com/strickyak/kiwisdr-qrss-grabber/riff"
	"github.com/strickyak/kiwisdr-qrss-grabber/transform"

	"bufio"
	"flag"
	"fmt"
	"image/png"
	"io"
	"os"
)

var _ = fmt.Printf

var IMPRINT = flag.String("imprint", "", "ASCII to print on image")
var RAW = flag.Bool("raw", false, "raw signed 16bit little endian input")
var FFT = flag.Int("fft", 15, "number of bits for FFT, like 14 or 15")
var RATE = flag.Int("rate", 12000, "Sample rate per second")
var BASE = flag.Int("base", 7039000, "Base tuner frequency in Hz")
var CENTER = flag.Int("center", 800, "Hz from base to center of image")
var BW = flag.Int("bw", 300, "Hz bandwidth across image")
var OVERLAP = flag.Float64("overlap", 0.5, "amount of audio to overlap from FFT to next FFT, 0 to 0.99")

func main() {
	flag.Parse()

	cf := &transform.Config{
		Rate:    *RATE,
		Base:    *BASE,
		Center:  *CENTER,
		Bw:      *BW,
		FFTBits: *FFT,
		Overlap: *OVERLAP,
	}

	var sampleCh <-chan int16
	if *RAW {
		sampleCh = readRaw(os.Stdin)
	} else {
		sampleCh = riff.GoReadFile("/dev/stdin")
	}
	fftCh := transform.GoRun(sampleCh, cf)
	img := transform.BuildImage(fftCh, *IMPRINT)
	png.Encode(os.Stdout, img)
}

func readRaw(r io.Reader) <-chan int16 {
	ch := make(chan int16, 10000)
	go func() {
		defer close(ch)
		br := bufio.NewReader(r)
		for {
			lo, err := br.ReadByte()
			if err != nil {
				break
			}
			hi, err := br.ReadByte()
			if err != nil {
				break
			}
			ch <- int16((uint16(lo) & 0xFF) | ((uint16(hi) << 8) & 0xFF00))
		}
	}()
	return ch
}
