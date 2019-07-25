// +build main

/*

## Repeatedly grab QRSS images.

Example:

```
go run grabloop.go --seconds=600 --spool=spool --overlap=0.8 --mangain=70 \
  --bands='W6REK-20m:sybil.yak.net:14096000:850:300,W6REK-40m:sybil.yak.net:7039000:850:300'
```

When you run that command, nothing will happen until the current time in
Unix Seconds, modulo the number of seconds you specify, is 0.
For example, 600 seconds is 10 minutes, so that command will start listening on the
next 10-minute period, like 12:00 or 12:10 or 12:20.  Then it will record for
10 minutes, so you won't get images until after it finishes recording.
It immediately starts the next recording, but with a 2 second gap, for the
KiwiSDR to drop the old connection and be ready for the new one.

The --bands parameter is 5-tuples of ID:kiwihost:BaseFreq:CenterOffset:Bandwidth
separated by commas.  The radio is tuned to BaseFreq in Upper Sideband mode.
The center frequency to plot is BaseFreq+CenterOffset.
The bandwidth of the plot is Bandwidth.
So the frequencies plotted range from BaseFreq+CenterOffset-(Bandwidth/2)
to BaseFreq+CenterOffset+(Bandwidth/2).

Plots are dropped in subdirectories (named by ID) of the --spool directory.
The most recent image is also copied up to a fixed filename ID.png in the --spool directory.

*/
package main

import (
	"github.com/strickyak/go-kiwisdr-client/client"
	"github.com/strickyak/kiwisdr-qrss-grabber/transform"

	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var SECONDS = flag.Int64("seconds", 60, "How many seconds per loop. Should divide 86400.")

var BANDS = flag.String("bands", "", "Comma-separated groups of Label:ServerHost:Base:Center:BW")
var SPOOL = flag.String("spool", "/tmp/spool", "Spool directory for holding images.")

var AGC = flag.Bool("agc", false, "AGC in SDR ")
var MANGAIN = flag.Int("mangain", 50, "Manual Gain in SDR (if no AGC) (10 to 90?)")
var FFT = flag.Int("fft", 15, "number of bits for FFT, like 14 or 15")
var RATE = flag.Int("rate", 12000, "Sample rate per second")
var OVERLAP = flag.Float64("overlap", 0.5, "amount of audio to overlap from FFT to next FFT, 0 to 0.99")

var IDENTIFY = flag.String("identify", "QRSS/grabber", "how to identify to the Kiwi")
var TIMES = flag.Int("times", -1, "how many times to grab.  Forever if negative. (bug: stops 1 too soon)")

type Grabber struct {
	Index         int
	Times         int
	Label         string
	ClientConf    *client.Config
	Tuning        *client.Tuning
	TransformConf *transform.Config
}

var UNIT struct{}
var done = make(chan struct{}, 1)

func main() {
	flag.Parse()

	bands := strings.Split(*BANDS, ",")
	for i, band := range bands {
		words := strings.Split(band, ":")
		if len(words) != 5 {
			log.Fatal("Got %q but expected 5 parts", words)
		}
		base, err := strconv.ParseInt(words[2], 10, 64)
		if err != nil {
			log.Fatal("Cannot ParseInt %q", words[2])
		}
		center, err := strconv.ParseInt(words[3], 10, 64)
		if err != nil {
			log.Fatal("Cannot ParseInt %q", words[3])
		}
		bw, err := strconv.ParseInt(words[4], 10, 64)
		if err != nil {
			log.Fatal("Cannot ParseInt %q", words[4])
		}
		grabber := &Grabber{
			Index: i,
			Label: words[0],
			ClientConf: &client.Config{
				ServerHost:  words[1] + ":8073",
				Kind:        client.SND,
				Identify:    *IDENTIFY,
				AGC:         *AGC,
				ManGain:     *MANGAIN,
				NoWaterfall: true,
			},
			Tuning: &client.Tuning{
				Freq: base,
				Mode: client.USB_3500,
			},
			TransformConf: &transform.Config{
				Rate:    *RATE,
				Base:    int(base),
				Center:  int(center),
				Bw:      int(bw),
				FFTBits: *FFT,
				Overlap: *OVERLAP,
			},
			Times: *TIMES,
		}
		go grabber.Loop()
	}
	for _, _ = range bands {
		<-done
	}
}

const TimestampFormat = "2006-01-02--15-04-05"

func (g *Grabber) Loop() {
	var period int64 = 1000000000 * (*SECONDS) // nanoseconds
	for {
		if g.Times == 0 {
			break
		}
		g.Times--

		now := time.Now()
		nownano := now.UnixNano()
		targettime := (1 + (nownano / period)) * period
		waittime := targettime - nownano
		waittime += 1000000000 * int64(g.Index) // stagger by a second.
		log.Printf("period %d now %d targettime %d waittime %d",
			period/1000000000,
			nownano/1000000000,
			targettime/1000000000,
			waittime/1000000000)

		log.Printf("sleeping waittime...")
		time.Sleep(time.Duration(waittime) * time.Nanosecond)

		log.Printf("awake; g.RunOnce...")
		timestamp := time.Unix(0, targettime).UTC().Format(TimestampFormat)
		go g.RunOnce(timestamp)
		time.Sleep(time.Second)
	}
	done <- UNIT
}

func (g *Grabber) RunOnce(timestamp string) {
	const GAP = 2 // Seconds.
	c := client.Dial(g.ClientConf, g.Tuning)
	ac := client.NewAudioClient(c)
	packetChannel := ac.BackgroundPlayForDuration(
		time.Duration(*SECONDS-GAP) * time.Second)

	audioChannel := make(chan int16, 12000)
	log.Printf("go generateAudioFromPackets")
	go generateAudioFromPackets(packetChannel, audioChannel)
	fftCh := transform.GoRun(audioChannel, g.TransformConf)

	midFreq := g.TransformConf.Base + g.TransformConf.Center
	halfBw := g.TransformConf.Bw / 2
	imprint := fmt.Sprintf("%s %s %d+-%d", g.Label, timestamp, midFreq, halfBw)
	img := transform.BuildImage(fftCh, imprint)

	dirname := filepath.Join(*SPOOL, g.Label)
	filename := filepath.Join(dirname, g.Label+"--"+timestamp+".png")
	os.Mkdir(dirname, 0755)
	log.Printf("Create: %q", filename)
	w, err := os.Create(filename)
	if err != nil {
		log.Panicf("Cannot create: %q", filename)
	}
	defer func() {
		w.Close()
		shortname := filepath.Join(*SPOOL, g.Label+".png")
		cmd := exec.Command("/bin/cp", "-v", filename, shortname)
		cmd.Start()
		go cmd.Wait()
	}()
	png.Encode(w, img)
	log.Printf("Encoded image.")
}

func generateAudioFromPackets(packetChannel <-chan client.AudioPacket, audioChannel chan<- int16) {
	for packet := range packetChannel {
		for _, s := range packet.Samples {
			audioChannel <- s
		}
	}
	close(audioChannel)
}
