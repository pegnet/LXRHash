package testing_test

import (
	"crypto/sha256"
	"fmt"
	. "github.com/pegnet/LXRHash"
	"math/rand"
	"testing"
	"time"
)

func TestDifferentHashes(t *testing.T) {
	rand.Seed(123412341234)

	Lxrhash.Init(Seed, MapSizeBits, HashSize, Passes)

	Gradehash{}.PrintHeader()

	numTests := 8
	for i := 0; i < numTests; i++ {
		go DifferentHashes()
	}

	for {
		time.Sleep(1 * time.Second)
	}
}

func DifferentHashes() {
	var g1 Gradehash
	var g2 Gradehash

	buf := Getbuf()

	last := time.Now().Unix()
	cnt := 0
	for i := 1; i < 100000000000; i++ {

		// Get a new buffer of data.
		buf = Getbuf()

		g1.Start()
		sv := sha256.Sum256(buf)
		g1.Stop()
		g1.AddHash(buf, sv[:])

		g2.Start()
		wv := Lxrhash.Hash(buf)
		g2.Stop()
		g2.AddHash(buf, wv)
		cnt++
		if cnt > 1000 && time.Now().Unix()-last > 4 {
			last = time.Now().Unix()
			cnt = 0

			c, r1 := g1.Report("dif-sha")
			_, r2 := g2.Report("dif-lxr")
			// Print on one line, so if we run multiple tests at the same time, we don't
			// split the output, because go will ensure one print goes out uninterrupted.
			fmt.Printf("%10s %s\n%10s %s\n\n", c, r1, " ", r2)
		}
	}
}
