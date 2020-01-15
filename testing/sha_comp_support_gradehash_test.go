// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	rand2 "crypto/rand"
	"fmt"
	"time"
)

// Routines for collecting stats on Hashing algorithms and comparing them to other
// implementations and libraries.

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
	fmt.Print("------------------------------------\n" +
		"Key For Data Printed while tests run:\n\n" +
		"| xxx,xxx :  of the number of hashes performed.  The test does the same number of sha hashes as lxr hashes\n\n" +
		"| bit-xxx :  This is the test, where the test (bit or add, or cnt, or dif) is followed by -xxx where xxx\n" +
		"               is either sha or lxr.  Like bit-sha or dif-lxr\n\n" +
		"| SB      :  How many bytes changed relative to expected number of the bytes that should change from one\n" +
		"               hash to the next.  You want zero, which means, over time, you have exactly the expected\n" +
		"               number of bytes changing\n\n" +
		"| xx - xx :  We count how many byte values we see. Possible values are 00 to FF.  All should be even, and\n" +
		"               no byte value should be favored.  We print which byte we saw the most, and which we saw the\n" +
		"               least. If the bytes change over time, that's good\n\n" +
		"| bits    :  Half the bits should change.  Averaged over all the hashes in the test, this is the difference\n" +
		"               between, say 128 for a 256 bit hash and how many bits have actually changed over the hashes\n\n" +
		"| score    :  On average, how many bits remain the same between hashes. Closer to 1/2 the bits in the hash is good\n" +
		"             Flip and Stay are picked to keep the difference positive, which is a better way to compare\n\n" +
		"| ffxxxxx :  The maximum unsigned high order eight bytes of the hash.  Like mining.  Both sha and lxr should\n" +
		"               should kinda take the same number of hashes to get kinda the same-ish max value\n\n" +
		"| cnt     :  Number of times we found a bigger hash in this run\n\n" +
		"| xxx hps :  We take the time executing sha and the time executing lxr and calculate a rough estimate of\n" +
		"               how many hashes per second we could be executing them.  Generally lxr is way slower\n" +
		"------------------------------------\n")
}

func (g *Gradehash) AddHash(src []byte, hash []byte) {

	for len(hash) > len(g.positionSums) {
		g.positionSums = append(g.positionSums, 0)
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
	if diff > g.difficulty {
		g.difficulty = diff
		g.diffHash = hash
		g.diffsrc = append(g.diffsrc[:0], src...)
		g.diffcnt++
		g.diffchanged = true
	}

}

func (g *Gradehash) Start() {
	g.start = time.Now().UnixNano()
}

func (g *Gradehash) Stop() {
	diff := time.Now().UnixNano() - g.start
	g.exctime += diff
}

// return the count of the number of hashes performed.
func (g *Gradehash) Report(name string) (hashcount string, report string) {

	if g.numhashes == 0 {
		report = fmt.Sprintln("no report data")
		return
	}

	freq := []float64{}
	for _, v := range g.bytefrequency {
		freq = append(freq, (float64(v) / float64(g.numhashes) / float64(len(g.positionSums))))
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
		delta := 1 - v*256
		score += delta * delta
	}

	spentv := float64(g.exctime) / 1000000000 // In seconds, divide by a billion
	hps := Comma(uint64(float64(g.numhashes) / spentv))
	spent := fmt.Sprintf("| %10s hps", hps)

	// Calculate how far off from half (128) we are.  Cause that is what matters.
	AvgBitsChanged := float64(g.bitsChanged) / float64(g.numhashes)

	bytesSame := float64(g.samebytes) / float64(g.numhashes)

	if score > 100 {
		score = 100
	}

	halfbits := float64(len(g.positionSums) * 8 / 2)
	avgChanged := ""
	avgChanged = fmt.Sprintf("bits: %11.8f", AvgBitsChanged-halfbits)

	hashcount = Comma(uint64(g.numhashes))

	report = fmt.Sprintf("%8s | SB %11.8f | %02x - %02x | score %12.10f | %s |",
		name,
		1.0/256*float64(len(g.diffHash))-bytesSame,
		maxb,
		minb,
		score,
		avgChanged)
	report += fmt.Sprintf(" %10x | cnt= %2d ",
		g.diffHash[:5],
		g.diffcnt)
	report += spent
	g.diffchanged = false
	return
}

// Consider a bigger number to be more difficult.  (This is a bit different
// than most PoW.  It is the same as viewing the return value as signed, and
// saying a smaller value is more difficult, due to the nature of signed
// values in binary.
func Difficulty(hash []byte) uint64 {
	// Calculate the Difficulty
	diff := uint64(0)
	for i := 0; i < 8; i++ {
		diff = diff<<8 + uint64(hash[i])
	}
	return diff
}

func Getbuf(length int) []byte {
	//buflen := minsample + rand.Intn(maxsample)
	nbuf := make([]byte, length)
	_, err := rand2.Reader.Read(nbuf)
	if err != nil {
		panic(err)
	}
	return nbuf
}

func Comma(n uint64) string {
	if n == 0 {
		return "0"
	}
	var s string
	for n > 999 {
		s = fmt.Sprintf(",%03d", n%1000) + s
		n /= 1000
	}
	return fmt.Sprintf("%d%s", n, s)
}
