package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"time"
)

var runStart int64

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

	if runStart == 0 {
		runStart = time.Now().Unix()
	}

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
	g.bitsDelta += (changedhere - 128) * (changedhere - 128)
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
	now := time.Now().Unix()
	secs := now - runStart
	hrs := secs / 60 / 60
	secs = secs - hrs*60*60
	mins := secs / 60
	secs = secs - mins*60

	runtime := fmt.Sprintf("%4d:%02d:%02d", hrs, mins, secs)

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

	// Calculate how far off from half (128) we are.  Cause that is what matters.
	AvgBitsChanged := 128 - float64(g.bitsChanged)/float64(g.numhashes)
	if AvgBitsChanged < 0 {
		AvgBitsChanged *= -1
	}
	Deltascore := float64(g.bitsDelta) / float64(g.numhashes)
	fmt.Printf("\n%s | %5s %12s:: | max,min : %3d% 10.6f : %3d %10.6f : | score %14.2f | 128Delta:  %10.8f | Sqr(Delta) %10.6f |",
		runtime,
		name,
		humanize.Comma(int64(g.numhashes)),
		maxb, maxn,
		minb, minn,
		score,
		AvgBitsChanged,
		Deltascore)
	fmt.Printf(" \"%20x\"::%30x diff:=%16x", g.diffsrc[:16], g.diffHash[:16], g.difficulty)
	fmt.Print("  ", spent, "\n")
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
