package lxr

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"time"
)

var runStart int64

type Gradehash struct {
	bytefrequency [256]int
	numhashes     int
	positionSums  []int
	last          []byte
	exctime       int64
	start         int64
	samebytes     int
	bitsChanged   int
	bitsDelta     int
	diffsrc       []byte
	diffcnt       int
	diffchanged   bool
	difficulty    uint64
	diffHash      []byte
}

func (g Gradehash) PrintHeader() {
	fmt.Println ("Key For Data Printed while tests run:\n" +
		"Time of test\n" +
		"| bit-xxx -- Specifies the test (bit) and the hash function used (sha is sha246, and lxr is LXRHash\n" +
		"| sameBytes -- number of repeated bytes in the hash\n" +
		"| max,min : -- most frequent bytes in all hashes and by how much.  Lower values are better\n" +
		"| score -- looks at the square of the difference between the expected number of changed bytes from one hash to another.  Lower is better\n" +
		"| BitsFlipped -- average number of bits flipped from one hash to the next.  Should be close to 1/2 the bits in a hash.\n" +
		"| 16 bytes of source data :: 16 bytes of the hash of the data + nonce\n" +
		"| diff = high 64 bytes of the hash as a difficulty, if the hash leads with two bytes of zeros\n" +
		"| cnt= number of qualifying 'difficult' hashes found\n" +
		"| seconds number of seconds in test used to calculate this hash\n")
}

func (g *Gradehash) AddHash(src []byte, hash []byte) {

	for len(hash) > len(g.positionSums) {
		g.positionSums = append(g.positionSums,0)
	}

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
		if g.last[i] == hash[i] {
			g.samebytes++
		}
	}
	g.bitsDelta += changedhere
	g.last = hash

	diff := Difficulty(hash)
	if g.difficulty == 0 || (diff != 0 && diff < g.difficulty) {
		g.difficulty = diff
		g.diffHash = hash
		g.diffsrc = append(g.diffsrc[:0], src...)
		if g.difficulty > 0 {
			g.diffcnt++
			g.diffchanged = true
		}
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

	freq := []float64{}
	for _, v := range g.bytefrequency {
		freq = append(freq, (float64(v)/float64(g.numhashes)/float64(len(g.positionSums))))
	}
	maxn := float64(freq[0])
	maxb := 0
	minb := 0
	minn := float64(freq[0])
	score := float64(0)
	for i, v := range freq {
		if v > maxn {
			maxn = v
			maxb = i
		}
		if v < minn {
			minn = v
			minb = i
		}
		delta := 1-v*256
		score += delta*delta
	}

	spentSec := g.exctime / 1000000000
	millisec := (g.exctime - (spentSec * 1000000000)) / 1000000
	spent := fmt.Sprintf("| seconds %8d.%03d", spentSec, millisec)

	// Calculate how far off from half (128) we are.  Cause that is what matters.
	AvgBitsChanged := float64(g.bitsChanged)/float64(g.numhashes)

	bytesSame := float64(g.samebytes) / float64(g.numhashes)

	if score > 100 {
		score = 100
	}

	halfbits := float64(len(g.positionSums)*8/2)
	avgChanged := ""
	if AvgBitsChanged > halfbits {
		avgChanged = fmt.Sprintf("%16s %12.8f","BitsFlipped",AvgBitsChanged)
	}else{
		avgChanged = fmt.Sprintf("%16s %12.8f","BitsUnchanged",halfbits*2 - AvgBitsChanged)
	}

	fmt.Printf("\n%s | %8s %12s:: | sameBytes %10.6f | max,min : %3d% 10.6f : %3d %10.6f : | score %18.10f | %s |",
		runtime,
		name,
		humanize.Comma(int64(g.numhashes)),
		bytesSame,
		maxb, maxn,
		minb, minn,
		score,
		avgChanged)
	if len(g.diffsrc) > 16 && len(g.diffHash) > 16 {
		fmt.Printf(" \"%20x\"::%30x | diff=%16x | cnt=%5d ",
			g.diffsrc[:16],
			g.diffHash[:16],
			g.difficulty,
			g.diffcnt)
	}
	g.diffchanged = false
	fmt.Println("  ", spent)
}

func Difficulty(hash []byte) uint64 {
	// skip start leading bytes If they are not zero, the Difficulty is zero
	start := 2
	for _, v := range hash[:start] {
		if v != 0 {
			return 0
		}
	}
	// The next 8 bytes define the Difficulty.  A smaller number is more difficult

	// Shift v a byte left and add the new byte
	as := func(v uint64, b byte) uint64 {
		return (v << 8) + uint64(b)
	}

	// Calculate the Difficulty
	diff := uint64(0)
	for i := start; i < start+6; i++ {
		// Add each byte to an 8 byte Difficulty, shifting the previous values left a byte each round
		diff = as(diff, hash[i])
	}
	return diff
}
