package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/dustin/go-humanize"
	"math/rand"
	"time"
)

const (
	firstrand = 0x13ef13156da2756b
	Mapsiz    = 0x400
	HBits     = 0x20
	HMask     = HBits - 1
	maxsample = 1
	minsample = 63
)

type Gradehash struct {
	bytefrequency [256]int
	numhashes     int
	positionSums  [32]int
	last          []byte
	exctime       int64
	start         int64
	bitsChanged   int
	bitsDelta     int
	diffsrc       []byte
	difficulty    uint64
	diffHash      []byte
}

func (g *Gradehash) AddHash(src []byte, hash []byte) {

	for _, v := range hash {
		g.bytefrequency[v]++
	}

	g.numhashes++
	for i, v := range hash {
		g.positionSums[i] += int(v)
	}
	changedhere := 0
	// pick one of 64 bytes
	for i := 0; i < len(g.last); i++ {
		// pick one of 8 bits
		for j := 0; j < 8; j++ {

			// Calculate a bit to check
			bit_to_flip := byte(1 << uint(j))
			if (g.last[i] & bit_to_flip) != (hash[i] & bit_to_flip) {
				g.bitsChanged++
				changedhere++
			}

		}
	}
	g.bitsDelta += (changedhere - 128) * (changedhere - 128) * 100000
	g.last = hash

	diff := difficulty(hash)
	if g.difficulty == 0 || (diff != 0 && diff < g.difficulty) {
		g.difficulty = diff
		g.diffHash = hash
		g.diffsrc = src
	}

}

func (g *Gradehash) Start() {
	g.start = time.Now().UnixNano()
}

func (g *Gradehash) Stop() {
	diff := time.Now().UnixNano() - g.start
	g.exctime += diff
}

func (g *Gradehash) Report(name string) {
	if g.numhashes == 0 {
		fmt.Println("no report data")
		return
	}

	sum := float64(0)
	for _, v := range g.bytefrequency {
		sum += float64(v)
	}

	avg := sum / 256 // Sum of all byte values generated over all hashes divided by the bytes possible
	diffs := []float64{}
	for _, v := range g.bytefrequency {
		diffs = append(diffs, (float64(v) - avg))
	}
	maxn := float64(0)
	maxb := 0
	minb := 0
	minn := float64(0)
	score := float64(0)
	for i, v := range diffs {
		if v > maxn {
			maxn = v
			maxb = i
		}
		if v < minn {
			minn = v
			minb = i
		}
		diff := v / float64(g.numhashes) * 10000000000 // Normalize the diff for the samples we have taken, and scale up.
		score += diff * diff                           // base the score on the square of the difference
		// The square of the difference doesn't have validity if the diff > -1 and < 1.
	}
	maxn = maxn / float64(g.numhashes)
	minn = minn / float64(g.numhashes)
	score = score / float64(g.numhashes)

	spentSec := g.exctime / 1000000000
	millisec := (g.exctime - (spentSec * 1000000000)) / 1000000
	spent := fmt.Sprintf("seconds %8d.%03d", spentSec, millisec)

	AvgBitsChanged := float64(g.bitsChanged) / float64(g.numhashes)
	Deltascore := g.bitsDelta / g.numhashes
	fmt.Printf("\n%5s %12s:: avg %10.2f maxdiff %3d=%10.6f mindiff %3d=%10.6f score %20.2f bitschanged %6.2f  DeltaScore: %20d",
		name, humanize.Comma(int64(g.numhashes)), avg, maxb, maxn, minb, minn, score, AvgBitsChanged, Deltascore)
	fmt.Printf(" \"%30s\"::%30x diff:=%16x", g.diffsrc, g.diffHash[:16], g.difficulty)
	fmt.Print("  ", spent)
}

func difficulty(hash []byte) uint64 {
	// skip start leading bytes.  If they are not zero, the difficulty is zero
	start := 2
	for _, v := range hash[:start] {
		if v != 0 {
			return 0
		}
	}
	// The next 8 bytes define the difficulty.  A smaller number is more difficult

	// Shift v a byte left and add the new byte
	as := func(v uint64, b byte) uint64 {
		return (v << 8) + uint64(b)
	}

	// Calculate the difficulty
	diff := uint64(0)
	for i := start; i < start+8; i++ {
		// Add each byte to an 8 byte difficulty, shifting the previous values left a byte each round
		diff = as(diff, hash[i])
	}
	return diff
}

type Whash struct {
	maps [Mapsiz]byte // Integer Offsets
}

func (w *Whash) Init() {
	dat, err := ioutil.ReadFile("whashmaps.dat")
	if err != nil || len(dat) != Mapsiz {

		// Ah, the data file isn't good for us.  Delete it (if it exists)
		os.Remove("whashmaps.dat")

		// Our own "random" generator that really is just used to shuffle values
		rands := [Mapsiz]int{}
		offset := firstrand
		rand := func(i int) int {
			offset = offset ^ (i << 30) + offset<<5 + offset>>5 ^ rands[offset&(Mapsiz-1)]
			rands[i] += offset
			return rands[i] & (Mapsiz - 1)
		}

		// Fill the maps with bytes ranging from 0 to 255.  As long as Mapsize%256 == 0, this
		// looping and masking works just fine.
		for i := range w.maps {
			w.maps[i] = byte(i)
		}

		// Now what we want to do is just mix it all up.  Take every byte in the maps list, and exchange it
		// for some other byte in the maps list. Note that we do this over and over, mixing and more mixing
		// the maps, but maintaining the ratio of each byte value in the maps list.
		for loops := 0; loops < 15000; loops++ {
			fmt.Println("Pass ", loops)
			for i := range w.maps {
				j := rand(i)
				w.maps[i], w.maps[j] = w.maps[j], w.maps[i]
			}
		}

		// open output file
		fo, err := os.Create("whashmaps.dat")
		if err != nil {
			panic(err)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		// write a chunk
		if _, err := fo.Write(w.maps[:]); err != nil {
			panic(err)
		}

	} else {
		copy(w.maps[:Mapsiz], dat)
	}
}

func (w Whash) Convert(offset int, ints [32]int) (bytes [32]byte) {
	for i, v := range ints {
		bytes[i] = w.maps[(offset+v)&(Mapsiz-1)]
		offset += v + int(bytes[i])
	}
	return
}

func (w Whash) Int(i int) int {
	return int(w.maps[i&(Mapsiz-1)])
}

// Takes a source of bytes, returns a 32 byte hash
func (w Whash) Hash(src []byte) []byte {
	offset := len(src)
	hashes := [32]int{}
	b := byte(1)
	for len(src) < 32 {
		src = append(src, b)
		b++
	}

	offset += int(src[0])<<8 + int(src[31])
	offset += (offset << 5) + (offset >> 7)
	for i := range hashes {
		offset = w.Int(offset+int(src[i])) + offset + (offset << 5) + (offset >> 7)
		hashes[i] = offset
	}

	//step := func(part []byte) {
	for i, v := range src {
		hi := i & 0x1F
		offset += (offset << 1) ^ (offset >> 1) + int(w.maps[(offset^int(v))&(Mapsiz-1)]) + i
		hv := hashes[hi]
		hashes[hi] += offset + hv>>3 + hv<<5 + hv ^ offset
	}
	//}

	//step(src)
	c := w.Convert(offset, hashes)
	return c[:]
}

func (w Whash) Convert2(off1, off2 int64, ints [32]int64) (bytes [32]byte) {
	var b byte
	for i, v := range ints {
		b = byte(v^off1^off2) ^ b
		off1 = off1>>9 ^ off1>>1 ^ off1<<7 ^ v ^ int64(i)
		off2 = off2>>7 ^ off2<<1 ^ off2<<9 ^ v ^ int64(i)
		bytes[i] = w.maps[(int64(w.maps[b])+off1)&(Mapsiz-1)]
	}
	return
}

// Takes a source of bytes, returns a 32 byte hash
func (w Whash) Hash2(src []byte) []byte {
	hashes := [HBits]int64{}
	i := int32(1)
	off1 := int64(len(src)) << 30
	step := func(v byte) {
		i0 := (i + 0) & HMask
		i1 := (i + 5) & HMask

		h0 := hashes[i0]
		h1 := hashes[i1]

		// Shift up a byte what is in offset, combined with offset shifted down a bit, combined with a byte and index
		bi := int64(w.maps[(off1^int64(v))&(Mapsiz-1)]) ^ int64(i)
		off1 = (off1 << 7) ^ (off1 >> 1) ^ (^(off1 & h0) >> 9) ^ (bi << 28) ^ h1

		hashes[i0] = (h0 << 7) ^ (h0 >> 1) ^ (h0 >> 9) ^ off1
		hashes[i1] = h1 ^ h0 ^ int64(w.maps[(off1^bi)&(Mapsiz-1)])<<30
		i += 1
	}
	for _, v := range src {
		step(v)
	}

	c := w.Convert2(off1, off1, hashes)
	return c[:]
}

func BitChangeTest() {
	var wh Whash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	getbuf := func() []byte {
		nbuf := random.RandByteSliceOfLen(rand.Intn(maxsample) + minsample)
		return nbuf
	}
	getbuf2 := func() []byte {
		bytes := make([]byte, rand.Intn(maxsample)+minsample)
		for i := 0; i < len(bytes); i++ {
			if rand.Intn(10) == 0 {
				bytes[i] = 32
			} else {
				if rand.Intn(2) == 0 {
					bytes[i] = byte(65 + rand.Intn(26)) //A=65 and Z = 65+25
				} else {
					bytes[i] = byte(97 + rand.Intn(26))
				}
			}
		}
		return bytes
	}
	_ = getbuf
	_ = getbuf2

	start := time.Now()

	wh.Init()
	buf := getbuf2()
	for i := 0; i < 100000000000; i++ {

		// Get a new buffer of data.
		buf = getbuf2()

		// pick one of 64 bytes
		for i := 0; i < len(buf); i++ {
			// pick one of 8 bits
			for j := 0; j < 8; j++ {

				// Calculate a bit to flip, and flip it.
				bit_to_flip := byte(1 << uint(j))
				buf[i] = buf[i] ^ bit_to_flip

				g1.Start()
				sv := sha256.Sum256(buf)
				g1.Stop()
				g1.AddHash(buf, sv[:])

				g2.Start()
				wv := wh.Hash2(buf)
				g2.Stop()
				g2.AddHash(buf, wv)

				// flipping a bit again repairs it.
				buf[i] = buf[i] ^ bit_to_flip

			}
		}

		t := time.Now()
		if t.Unix()-start.Unix() > 5 {
			fmt.Println("\n", string(buf))
			g1.Report("sha")
			g2.Report("wh")
			start = t
		}

	}
}

func main() {
	var wh Whash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	getbuf := func() []byte {
		nbuf := random.RandByteSliceOfLen(rand.Intn(maxsample) + minsample)
		return nbuf
	}
	getbuf2 := func() []byte {
		bytes := make([]byte, rand.Intn(maxsample)+minsample)
		for i := 0; i < len(bytes); i++ {
			if rand.Intn(10) == 0 {
				bytes[i] = 32
			} else {
				if rand.Intn(2) == 0 {
					bytes[i] = byte(65 + rand.Intn(26)) //A=65 and Z = 65+25
				} else {
					bytes[i] = byte(97 + rand.Intn(26))
				}
			}
		}
		return bytes
	}
	_ = getbuf
	_ = getbuf2

	start := time.Now()

	wh.Init()
	buf := getbuf2()
	for i := 0; i < 100000000000; i++ {

		// Get a new buffer of data.
		buf = getbuf2()

		g1.Start()
		sv := sha256.Sum256(buf)
		g1.Stop()
		g1.AddHash(buf, sv[:])

		g2.Start()
		wv := wh.Hash2(buf)
		g2.Stop()
		g2.AddHash(buf, wv)

		if i%10000 == 0 {
			t := time.Now()
			if t.Unix()-start.Unix() > 5 {
				fmt.Println()
				g1.Report("sha")
				g2.Report("wh")
				start = t
			}
		}
	}
}
