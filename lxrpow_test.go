// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

var Found chan solution

const LxrPoW_time int64 = 1200

type solution struct {
	time        time.Time // Time the solution was found
	solutionCnt uint64    // Solutions found
	totalCnt    int       // Total Hashes computed
	comment     string    // Label for what sort of hashing used for this solution
	hash        []byte    // The hash provided by some hashing function
	lhash       []byte    // The modified hash used to compute difficulty (the same for simple PoW)
	diff        uint64    // The reported difficulty
}

// TstHash()
// Test either simple Sha256 PoW, or the LxrPoW function()
// Runs a go routine forever, doing one minute blocks and hashing for a higher PoW.
// Comment is a label used in reporting, and lxrPoW is a bool that if true executes LxrPow otherwise simple PoW
func TstHash(comment string, lxrPoW bool) {
	powCnt := 0             // Count how many PoW solutions we find.  Each one should be greater than the last
	var pow, lastPow uint64 // pow is the current computed difficulty, lastPow is the last solution we found.

	// We just need to hash something.
	v := sha256.Sum256([]byte("xxxx")) // Usage is to do one sha256 of an entry, then adda nonce.  Start with a hash
	data := v[:]                       // We need this value as a slice
	var lhash []byte                   // lhash is the value we do the PoW calculation on. We get this by running the hash through the ByteMap
	start := time.Now()                // Track when we started, so we can figure outhow fast the hash is.

	for i := uint64(1); ; i++ { // This is our hashing loop, and we count how many we execute

		// Take the base hash (v) and add our nonce (the current count of hashes).
		n := i           // n is going to be our nonce.  We will shift away bytes until n is nil.
		data = data[:32] // We reuse the data slice (cuts down on garbage collection)
		for n > 0 {      // While the nonce isn't zero
			data = append(data, byte(n)) // Add the least significant byte remaining.  We don't care about reversing
			n = n >> 8                   // the order of bytes in the nonce.  We just need a nonce.
		}

		h1 := sha256.Sum256(data) // Take the hash of the data + nonce
		hash := h1[:]             // Make the result a slice

		if lxrPoW { // If using lxrPoW, the we do that, otherwise we just scrape the PoW from the hash as is
			lhash, pow = lx.LxrPoW(int(lastPow>>60), hash[:])
		} else {
			pow = lx.PoW(hash[:])
			lhash = hash
		}

		if pow > lastPow { // If we have a greater value for the PoW, then report it
			lastPow = pow
			powCnt++
			Found <- solution{time.Now(), i, powCnt, comment, hash, lhash, pow}
		}

		if i&0xFF == 0 { // Because checking the clock is expensive, just do it every 255 hashes
			if time.Now().Unix()-start.Unix() > 60 { // If 60 seconds have passed, then start a new block
				start = time.Now() // reset our timer

				// End the block, start another.
				pow = 0                 // Look for a new PoW
				lastPow = 0             // Have no lastPoW anymore, so reset that too
				v = sha256.Sum256(v[:]) // Compute a new base hash
			}
		}
	}
}

// This is our main function.  As written, it runs for the time specified by the runtime constant
func TestPoW(t *testing.T) {
	start := time.Now()               // Mark our start
	Found = make(chan solution, 1000) // Create a channel for the go routines to use to report results
	go TstHash("lxr", true)           // Start the hashing for lxrPoW
	go TstHash("hsh", false)          // Start the hashing for plain PoW
	lxrTotal := uint64(1)             // Total lxr PoW calculations
	hshTotal := uint64(1)             // Total normal PoW calculations
	lxrSeconds := uint64(1)           // Track the time the last lxr PoW was reported
	hshSeconds := uint64(1)           // Same for regular Hash

	// Now our feedback loop
	for {
		// If time is up, than return
		if time.Now().Unix()-start.Unix() > LxrPoW_time {
			return
		}
		// Get the next reported solution
		solution := <-Found
		// Take the seconds from the beginning
		seconds := uint64(solution.time.Sub(start).Seconds())
		if seconds == 0 {
			seconds = seconds + 1
		}

		// do the math for lxrPoW, or PoW (the comment tells us which we are)
		if solution.comment == "lxr" {
			lxrTotal = solution.solutionCnt
			lxrSeconds = seconds
		} else {
			hshTotal = solution.solutionCnt
			hshSeconds = seconds
		}
		// Calculate Ratio of plain PoW vs LxrPoW
		ratio := (hshTotal / hshSeconds) / (lxrTotal / lxrSeconds)
		// Report what we found
		fmt.Printf("Comment %5s Rate %10d/s Ratio %6d totalCnt %4d POW %016x Hash %x, LxrPoW %x\n",
			solution.comment, solution.solutionCnt/seconds, ratio,
			solution.totalCnt, solution.diff, solution.hash, solution.lhash)
	}
}
