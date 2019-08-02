// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

func TestCount(t *testing.T) {
	LX.Init(Seed, MapSizeBits, HashSize, Passes)

	Gradehash{}.PrintHeader()

	numTests := 1
	for i := 0; i < numTests; i++ {
		go BitCountTest()
	}

	time.Sleep(20 * time.Second)

}

func BitCountTest() {
	var g1 Gradehash
	var g2 Gradehash

	cnt := int64(0)

	last := time.Now().Unix()
	for x := int64(0); x < 100000000000; x++ {
		// Get a new buffer of data.
		buf := Getbuf(1024)

		for i := 0; i < 10; i++ {
			buf[i] = 0
		}

		for i := 0; i < 1000000; i++ {
			// pick one of 64 bytes
			for i := 0; ; i++ {
				buf[i] += 1
				if buf[i] != 0 {
					break
				}
			}

			cnt++

			g1.Start()
			sv := sha256.Sum256(buf)
			g1.Stop()
			g1.AddHash(buf, sv[:])

			g2.Start()
			wv := LX.Hash(buf)
			g2.Stop()
			g2.AddHash(buf, wv)

			if cnt > 1000 && time.Now().Unix()-last > 4 {
				last = time.Now().Unix()
				cnt = 0

				c, r1 := g1.Report("cnt-sha")
				_, r2 := g2.Report("cnt-lxr")
				// Print on one line, so if we run multiple tests at the same time, we don't
				// split the output, because go will ensure one print goes out uninterrupted.
				fmt.Printf("%10s %s\n%10s %s\n\n", c, r1, " ", r2)
			}
		}

	}
}
