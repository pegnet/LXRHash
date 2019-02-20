package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/FactomProject/factomd/common/primitives/random"
	"math/rand"
	"time"
)

func BitChangeTest() {
	var wh PegHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	getbuf := func() []byte {
		nbuf := random.RandByteSliceOfLen(rand.Intn(maxsample) + minsample)
		return nbuf
	}
	getbuf2 := func() []byte {
		bytes := make([]byte, rand.Intn(maxsample)+minsample)
		for i := 0; i < len(bytes); i++ {
			if rand.Intn(10) == 0 {
				bytes[i] = 32
			} else {
				if rand.Intn(2) == 0 {
					bytes[i] = byte(65 + rand.Intn(26)) //A=65 and Z = 65+25
				} else {
					bytes[i] = byte(97 + rand.Intn(26))
				}
			}
		}
		return bytes
	}
	_ = getbuf
	_ = getbuf2

	start := time.Now()

	wh.Init()
	buf := getbuf2()
	for i := 0; i < 100000000000; i++ {

		// Get a new buffer of data.
		buf = getbuf2()

		// pick one of 64 bytes
		for i := 0; i < len(buf); i++ {
			// pick one of 8 bits
			for j := 0; j < 8; j++ {

				// Calculate a bit to flip, and flip it.
				bit_to_flip := byte(1 << uint(j))
				buf[i] = buf[i] ^ bit_to_flip

				g1.Start()
				sv := sha256.Sum256(buf)
				g1.Stop()
				g1.AddHash(buf, sv[:])

				g2.Start()
				wv := wh.Hash(buf)
				g2.Stop()
				g2.AddHash(buf, wv)

				// flipping a bit again repairs it.
				buf[i] = buf[i] ^ bit_to_flip

			}
		}

		t := time.Now()
		if t.Unix()-start.Unix() > 5 {
			fmt.Println("\n", string(buf))
			g1.Report("sha")
			g2.Report("wh")
			start = t
		}

	}
}

func main() {
	var wh PegHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	rand.Seed(13243442344225879)

	getbuf := func() []byte {
		nbuf := random.RandByteSliceOfLen(rand.Intn(maxsample) + minsample)
		return nbuf
	}
	getbuf2 := func() []byte {
		bytes := make([]byte, rand.Intn(maxsample)+minsample)
		for i := 0; i < len(bytes); i++ {
			if rand.Intn(10) == 0 {
				bytes[i] = 32
			} else {
				if rand.Intn(2) == 0 {
					bytes[i] = byte(65 + rand.Intn(26)) //A=65 and Z = 65+25
				} else {
					bytes[i] = byte(97 + rand.Intn(26))
				}
			}
		}
		return bytes
	}
	_ = getbuf
	_ = getbuf2

	start := time.Now()

	wh.Init()
	buf := getbuf2()
	for i := 1; i < 100000000000; i++ {

		// Get a new buffer of data.
		buf = getbuf2()

		g1.Start()
		sv := sha256.Sum256(buf)
		g1.Stop()
		g1.AddHash(buf, sv[:])

		g2.Start()
		wv := wh.Hash(buf)
		g2.Stop()
		g2.AddHash(buf, wv)

		if i%1000000 == 0 {
			t := time.Now()
			if t.Unix()-start.Unix() > 5 {
				fmt.Println()
				g1.Report("sha")
				g2.Report("wh")
				start = t
			}
		}
	}
}
