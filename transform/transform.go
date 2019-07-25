package transform

import (
	"log"
	"math"
	"math/cmplx"

	"mjibson/go-dsp/fft" // vendored.
)

var _ = log.Printf

type Config struct {
	Rate    int
	Base    int
	Center  int
	Bw      int
	FFTBits int // Log2 of FFT size, maximum 15.
	Overlap float64
}

type Transform struct {
	*Config
	Size   int
	Samps  []float64
	Shaped []float64
}

func GoRun(in <-chan int16, cf *Config) <-chan []float64 {
	size := 1 << uint(cf.FFTBits)
	t := &Transform{
		Config: cf,
		Size:   size,
		Samps:  make([]float64, size),
		Shaped: make([]float64, size),
	}
	out := make(chan []float64)
	go func() {
		for {
			t.Shift()
			eof := t.Load(in)
			t.CopyAndShapeSampsToShaped()
			fft := t.FFT()
			out <- t.NormalizeFFT(fft) // From complex values to int16s
			if eof {
				break
			}
		}
		close(out)
	}()
	return out
}

var logOnce bool

// Extract region of interest and convert to unsigned 16bit magnitudes.
func (t *Transform) NormalizeFFT(fft []complex128) []float64 {
	// Offsets above base frequeny for lo & hi edges of image.
	hiOff := t.Center + t.Bw/2
	loOff := t.Center - t.Bw/2
	// Indices of region of interest in the FFT output.
	Nyquist := t.Rate / 2
	hi := int(t.Size / 2 * hiOff / Nyquist)
	lo := int(t.Size / 2 * loOff / Nyquist)
	tall := hi - lo + 1
	if logOnce {
		log.Printf("t.Size=%d Nyquist=%d", t.Size, Nyquist)
		log.Printf("hi=%d lo=%d tall=%d", hi, lo, tall)
	}
	// FFT results for region of interest.
	results := make([]float64, tall)
	for i := 0; i < tall; i++ {
		// Trial and error lead to using math.Sqrt.
		abs1 := math.Sqrt(cmplx.Abs(fft[i+lo]))
		abs2 := math.Sqrt(cmplx.Abs(fft[t.Size-1-(i+lo)]))
		// Average of the positive & negative freq magnitudes.
		results[i] = (abs1 + abs2) / 2
	}
	return results
}

const RampLen = 12000 / 100       // 10ms if 12k samples per second.
var RaisedCosine [RampLen]float64 // Lookup table.

func init() { // Precompute raised-cosine lookup table.
	for i := 0; i < RampLen; i++ {
		theta := float64(i) * math.Pi
		// Starts at 0.0, rises to 1.0.
		RaisedCosine[i] = 1.0 - math.Cos(theta)/2.0
	}
}

func (t *Transform) CopyAndShapeSampsToShaped() {
	copy(t.Shaped[:], t.Samps[:])

	// Shape front and back.
	for i := 0; i < RampLen; i++ {
		t.Shaped[i] *= RaisedCosine[i]
		t.Shaped[t.Size-1-i] *= RaisedCosine[i]
	}
}

func (t *Transform) Shift() {
	shiftSize := int(t.Overlap * float64(t.Size))
	shiftFrom := t.Size - shiftSize
	copy(t.Samps[:shiftSize], t.Samps[shiftFrom:])
}

// returns true if eof was hit.
func (t *Transform) Load(in <-chan int16) bool {
	eof := false
	shiftSize := int((1.0 - t.Overlap) * float64(t.Size))
	shiftFrom := t.Size - shiftSize
	for i := 0; i < shiftSize; i++ {
		if eof {
			t.Samps[shiftFrom+i] = 0.0
		} else {
			samp, ok := <-in
			t.Samps[shiftFrom+i] = float64(samp)
			eof = (!ok)
		}
	}
	return eof
}

func (t *Transform) FFT() []complex128 {
	// Use FFT implementation from mjibson/go-dsp/fft.
	return fft.FFTReal(t.Shaped[:t.Size])
}
