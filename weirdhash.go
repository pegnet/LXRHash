package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"crypto/sha256"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/dustin/go-humanize"
	"math/rand"
	"time"
)

type Gradehash struct {
	bytefrequency [256]int
	numhashes     int
	positionSums  [32]int
	last          []byte
	exctime       int64
	start         int64
}

func (g *Gradehash) AddHash(hash []byte) {
	g.last = hash

	for _, v := range hash {
		g.bytefrequency[v]++
	}

	g.numhashes++
	for i, v := range hash {
		g.positionSums[i] += int(v)
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
	minn := float64(0)
	score := float64(0)
	for _, v := range diffs {
		if v > maxn {
			maxn = v
		}
		if v < minn {
			minn = v
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

	fmt.Printf("\n%5s %12s:: avg %10.2f maxdiff %10.6f mindiff %10.6f score %20.2f:", name, humanize.Comma(int64(g.numhashes)), avg, maxn, minn, score)
	fmt.Printf(" %33x ", g.last)
	fmt.Print("  ", spent)
}

const Mapsiz = 0x800

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
		offset := 0
		rand := func(i int) int {
			offset += i + offset<<4 + offset>>3 ^ rands[offset&(Mapsiz-1)]
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
		for loops := 0; loops < 5000; loops++ {
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

func main() {
	var wh Whash
	var g0 Gradehash
	var g1 Gradehash
	var g2 Gradehash

	getbuf := func() []byte {
		nbuf := random.RandByteSliceOfLen(rand.Intn(1024000) + 10240)
		return nbuf
	}

	wh.Init()
	for i := 0; i < 1000000000; i++ {
		g0.Start()
		buf := getbuf()
		g0.Stop()
		g0.AddHash(buf[:32])

		g1.Start()
		sv := sha256.Sum256(buf)
		g1.Stop()
		g1.AddHash(sv[:])

		g2.Start()
		wv := wh.Hash(buf)
		g2.Stop()
		g2.AddHash(wv)

		if (i+1)%(100) == 0 {
			fmt.Println()
			g0.Report("rand")
			g1.Report("sha")
			g2.Report("wh")
		}
	}
}
