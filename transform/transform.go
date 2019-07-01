package transform

import (
	"log"
	"math"
	"math/cmplx"
	"mjibson/go-dsp/fft" // vendored.
)

var _ = log.Printf

const SampPerSec = 12000
const Nyquist = SampPerSec / 2
const ImageHeight = 600
const FFTSize = 1 << 14 // 14 or 15

// const TargetBW = 300
// const LoF = 1200
// const HiF = LoF + TargetBW

// const TargetBW = 1000
// const LoF = 800
// const HiF = LoF + TargetBW

// Large band capture for now.
const TargetBW = 700
const LoF = 1000
const HiF = LoF + TargetBW

type Transform struct {
	Samps   [FFTSize]float64
	Shaped  [FFTSize]float64
	Overlap float64
}

func GoRun(in <-chan int16, overlap float64) <-chan []uint16 {
	t := &Transform{
		Overlap: overlap,
	}
	out := make(chan []uint16)
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

var once1 bool

func (t *Transform) NormalizeFFT(fft []complex128) []uint16 {
	hi := int(FFTSize / 2 * HiF / Nyquist)
	lo := int(FFTSize / 2 * LoF / Nyquist)
	tall := hi - lo
	if once1 {
		log.Printf("FFTSize=%d Nyquist=%d", FFTSize, Nyquist)
		log.Printf("hi=%d lo=%d tall=%d", hi, lo, tall)
	}
	norm := make([]uint16, tall)
	if false {
		var min, max float64
		for i := 0; i < FFTSize; i++ {
			abs := 1000 * math.Log10(cmplx.Abs(fft[i]))
			if abs > 65000 {
				abs = 65000
			}
			if abs < min || min == 0 {
				min = abs
			}
			if abs > max {
				max = abs
			}
		}
		// log.Printf("min: %f    max: %f", min, max)
		for i := 0; i < FFTSize; i++ {
			abs := 1000 * math.Log10(cmplx.Abs(fft[i]))
			x := 65000 * (abs - min) / (max - min)
			//log.Printf("     abs: %f        x: %f", abs, x)
			if x > 65000 {
				x = 65000
			}
			norm[i] = uint16(int(x))
		}
	} else {
		for i := 0; i < tall; i++ {
			abs := 1000 * math.Log10(cmplx.Abs(fft[i+lo]))
			if abs > 65000 {
				abs = 65000
			}
			norm[i] = uint16(int(abs))
		}
	}
	return norm
}

const RampLen = SampPerSec / 100 // 10ms
var RaisedCosine [RampLen]float64

func init() {
	for i := 0; i < RampLen; i++ {
		theta := float64(i) * math.Pi
		// Rtarts at 0.0, rises to 1.0.
		RaisedCosine[i] = 1.0 - math.Cos(theta)/2.0
	}
}

func (t *Transform) CopyAndShapeSampsToShaped() {
	copy(t.Shaped[:], t.Samps[:])

	// Shape front and back.
	for i := 0; i < RampLen; i++ {
		t.Shaped[i] *= RaisedCosine[i]
		t.Shaped[FFTSize-1-i] *= RaisedCosine[i]
	}

	// Eliminate DC bias.
	sum := 0.0
	for _, e := range t.Shaped {
		sum += e
	}
	avg := sum / float64(len(t.Shaped))
	for i, _ := range t.Shaped {
		t.Shaped[i] -= avg
	}
}
func (t *Transform) Shift() {
	shiftSize := int(t.Overlap * FFTSize)
	shiftFrom := FFTSize - shiftSize
	copy(t.Samps[:shiftSize], t.Samps[shiftFrom:])
}

// returns true if eof was hit.
func (t *Transform) Load(in <-chan int16) bool {
	eof := false
	shiftSize := int(t.Overlap * FFTSize)
	shiftFrom := FFTSize - shiftSize
	for i := 0; i < shiftSize; i++ {
		if eof {
			t.Samps[shiftFrom+i] = 0.0
		} else {
			x, ok := <-in
			t.Samps[shiftFrom+i] = float64(x)
			eof = (!ok)
		}
	}
	return eof
}
func (t *Transform) FFT() []complex128 {
	return fft.FFTReal(t.Shaped[:])
}
