// Read microsoft RIFF format as written by `kiwiclient/kiwirecorder.py`.
// Can assume little-endian signed 16-bit samples, 1 channel, 12000 samples per second.
// (This will not work for IQ mode, which has 2 channels.)
package riff

import (
	"bufio"
	"log"
	"os"
)

func GoReadFile(filename string) <-chan int16 {
	r, err := os.Open(filename)
	if err != nil {
		log.Panicf("riff: cannot open %q: %v", filename, err)
	}
	br := bufio.NewReader(r)
	ch := make(chan int16)
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Fatalf("fatal: err in bg riff reader: %v", r)
			}
		}()
		ReadAll(br, ch)
	}()
	return ch
}

func ReadAll(r *bufio.Reader, ch chan int16) {
CHUNKS:
	for {
		// Read Chunk Descriptor
		id := read4s(r)
		log.Printf("Chunk id %q", id)
		if id == "" {
			break CHUNKS
		}
		if id != "RIFF" {
			log.Panic("riff: Chunk id not 'RIFF': %q", id)
		}
		_ = read4u(r) // size
		format := read4s(r)
		if format != "WAVE" {
			log.Panic("riff: Chunk format not 'WAVE': %q", format)
		}

	SUBCHUNKS:
		for {
			id = read4s(r)
			log.Printf("subchunk id %q", id)
			switch id {
			case "":
				break SUBCHUNKS
			case "fmt ":
				readFmtSubchunk(r)
			case "data":
				readDataSubchunk(r, ch)
			default:
				log.Panic("Unknown subchunk id: %q", id)
			}
		}
	}
	close(ch)
}

func readFmtSubchunk(r *bufio.Reader) {
	size := read4u(r)
	if size != 16 {
		log.Panicf("riff: fmt size not 16: %u", size)
	}
	format := read2u(r)
	numch := read2u(r)
	samprate := read4u(r)
	byterate := read4u(r)
	align := read2u(r)
	bitsPer := read2u(r)
	// if format != 1 || numch != 1 || (samprate != /*0x2ee1/ 1002 {
	if format != 1 || numch != 1 || samprate < 12000 || samprate > 12001 {
		log.Panicf("riff: bad format: %d %d %d %d %d %d", format, numch, samprate, byterate, align, bitsPer)
	}
}

func readDataSubchunk(r *bufio.Reader, ch chan int16) {
	size := read4u(r)
	log.Printf("data size %d", size)
	for i := 0; i < int(size); i += 2 {
		sample := read2u(r)
		x := int(sample)
		if x > 0x7FFF {
			// Adjust x into range of int16:
			// if it was too big, it should be negative.
			x -= 0x10000
		}
		ch <- int16(x)
	}
}

// read4s returns "" if short.
func read4s(r *bufio.Reader) string {
	var z []byte
	for i := 0; i < 4; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return ""
		}
		z = append(z, b)
	}
	return string(z)
}

// read4u panics if short.
func read4u(r *bufio.Reader) uint {
	var z uint
	var i uint
	for i = 0; i < 4; i++ {
		b, err := r.ReadByte()
		if err != nil {
			log.Panicf("riff: short file")
		}
		z |= uint(b) << (i * 8)
	}
	return z
}

// read2u panics if short.
func read2u(r *bufio.Reader) uint {
	var z uint
	var i uint
	for i = 0; i < 2; i++ {
		b, err := r.ReadByte()
		if err != nil {
			log.Panicf("riff: short file")
		}
		z |= uint(b) << (i * 8)
	}
	return z
}
