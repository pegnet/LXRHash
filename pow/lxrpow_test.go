// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package pow

import (
	"crypto/sha256"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var lx = new(LxrPow).Init(30, 6)

const LxrPoW_time int64 = 120

func Test_LxrPoW(t *testing.T) {
	type result struct{ nonce, pow uint64 }
	results := make(chan result, 10)

	oprHash := sha256.Sum256([]byte("This is a test"))
	var best result
	var hashCnt int64
	start := time.Now()
	nonce := uint64(start.UnixNano())

	goCnt := 8
	for i := 0; i < goCnt; i++ {

		go func(instance int, a [32]byte) {
			var nPow, last uint64
			for {
				nonce = nonce<<17 ^ nonce>>9 ^ uint64(hashCnt) ^ uint64(instance) // diff nonce for each instance
				h := atomic.AddInt64(&hashCnt, 1)
				_, nPow = lx.LxrPoW(a[:], uint64(h))
				if nPow > last {
					last = nPow
					results <- result{nonce, nPow}
				}
			}
		}(i, oprHash)
	}

	// pull the results, and print the best hashes
	for i := 0; true; {
		r := <-results
		if r.pow > best.pow {
			i++
			best = r
			current := time.Since(start)
			rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
			fmt.Printf("  %3d time: %10s TH: %10d H/s: %12.5f Pow: %016x Hash: %64x Nonce: %016x\n",
				i, fmt.Sprintf("%3d:%02d:%02d", int(current.Hours()), int(current.Minutes())%60, int(current.Seconds())%60),
				hashCnt, rate, r.pow, oprHash, r.nonce)
		} else if r.pow > 0x02f0<<48 {
			current := time.Since(start)
			rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
			fmt.Printf("      time: %10s TH: %10d H/s: %12.5f Pow: %016x\n",
				fmt.Sprintf("%3d:%02d:%02d", int(current.Hours()), int(current.Minutes())%60, int(current.Seconds())%60),
				hashCnt, rate, r.pow)
		}
	}
}
