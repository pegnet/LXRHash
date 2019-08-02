// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

func TestNonce(t *testing.T) {
	LX.Init(Seed, 25, HashSize, Passes)

	Gradehash{}.PrintHeader()

	numTests := 1
	for i := 0; i < numTests; i++ {
		go NonceTest()
	}

	time.Sleep(20 * time.Second)

}

const (
	elementsInLRU = 16 * 1024
	cacheLine     = uint(3) // 1 would be 2 bytes, 2 would be 4 bytes, 3 would be 8 bytes, etc.
)

func NonceTest() {
	var g1 Gradehash
	var g2 Gradehash

	cnt := int64(0)

	base := Getbuf(32)

	last := time.Now().Unix()

	hits := make(map[uint64]int)
	var list [elementsInLRU]uint64
	for i := range list {
		list[i] = 0xffffffffffffffff
	}
	hit := func(v uint64) bool {
		v = v >> cacheLine // Cache line of 64 bytes
		if hits[v] > 0 {
			for i := elementsInLRU - 1; i >= 0; i-- {
				if list[i] == v {
					if i > 0 {
						copy(list[1:i], list[0:i-1])
					}
					list[0] = v
					hits[v]++
					return true
				}
			}
		}
		copy(list[1:], list[:])
		list[0] = v
		hits[v] = 1

		return false
	}

	var numHits int
	for x := int64(0); x < 100000000000; x++ {
		// Get a new buffer of data.
		base = base[:32]
		for n := x + 1000; n > 0; n = n >> 8 {
			base = append(base, byte(n))
		}

		g1.Start()
		sv := sha256.Sum256(base)
		g1.Stop()
		g1.AddHash(base, sv[:])

		g2.Start()
		wv := LX.Hash(base)
		if hit(LX.FirstIdx) {
			numHits++
		}
		g2.Stop()
		g2.AddHash(base, wv)

		cnt++
		if cnt > 1000 && time.Now().Unix()-last > 4 {
			last = time.Now().Unix()
			cnt = 0

			c, r1 := g1.Report("non-sha")
			_, r2 := g2.Report("non-lxr")
			// Print on one line, so if we run multiple tests at the same time, we don't
			// split the output, because go will ensure one print goes out uninterrupted.
			fmt.Printf("%10s %s\n%10s %s Hits %d\n\n", c, r1, " ", r2, numHits)
		}
	}
}
