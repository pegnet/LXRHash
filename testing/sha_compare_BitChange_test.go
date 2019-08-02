// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

func TestBitChange(t *testing.T) {
	LX.Init(Seed, MapSizeBits, HashSize, Passes)

	Gradehash{}.PrintHeader()

	numTests := 1
	for i := 0; i < numTests; i++ {
		go BitChangeTest()
	}

	time.Sleep(20 * time.Second)

}

func BitChangeTest() {
	var g1 Gradehash
	var g2 Gradehash

	cnt := int64(0)

	last := time.Now().Unix()
	for x := int64(0); x < 100000000000; x++ {
		// Get a new buffer of data.
		buf := Getbuf(1024)

		// pick one of 64 bytes
		for i := 0; i < len(buf); i++ {
			// pick one of 8 bits
			for j := 0; j < 8; j++ {
				cnt++
				// Calculate a bit to flip, and flip it.
				bit_to_flip := byte(1 << uint(j))
				buf[i] = buf[i] ^ bit_to_flip

				g1.Start()
				sv := sha256.Sum256(buf)
				g1.Stop()
				g1.AddHash(buf, sv[:])

				g2.Start()
				wv := LX.Hash(buf)
				g2.Stop()
				g2.AddHash(buf, wv)

				// flipping a bit again repairs it.
				buf[i] = buf[i] ^ bit_to_flip

				if true {
					if cnt > 1000 && time.Now().Unix()-last > 4 {
						last = time.Now().Unix()
						cnt = 0

						c, r1 := g1.Report("bit-sha")
						_, r2 := g2.Report("bit-lxr")
						// Print on one line, so if we run multiple tests at the same time, we don't
						// split the output, because go will ensure one print goes out uninterrupted.
						fmt.Printf("%10s %s\n%10s %s\n\n", c, r1, " ", r2)
					}
				} else {
					if cnt > 100000 {
						cnt = 0

						c, r1 := g1.Report("bit-sha")
						_, r2 := g2.Report("bit-lxr")
						// Print on one line, so if we run multiple tests at the same time, we don't
						// split the output, because go will ensure one print goes out uninterrupted.
						fmt.Printf("%10s %s\n%10s %s\n\n", c, r1, " ", r2)
					}

				}
			}

		}

	}
}
