// +build main

// go run render.go -imprint "Q0QQQ on 40m at $(date -u)" < 40m.wav > 40m.png
package main

import (
	"github.com/strickyak/kiwisdr-qrss-grabber-client/riff"
	"github.com/strickyak/kiwisdr-qrss-grabber-client/transform"

	"flag"
	"fmt"
	"image/png"
	"os"
)

var _ = fmt.Printf

var IMPRINT = flag.String("imprint", "", "ASCII to print on image")

func main() {
	flag.Parse()

	sampleCh := riff.GoReadFile("/dev/stdin")
	fftCh := transform.GoRun(sampleCh, 0.5)
	img := transform.BuildImage(fftCh, *IMPRINT)
	png.Encode(os.Stdout, img)
}
