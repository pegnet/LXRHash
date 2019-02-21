package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/FactomProject/factomd/common/primitives/random"
	"math/rand"
)

func Getbuf() []byte {
	nbuf := random.RandByteSliceOfLen(rand.Intn(maxsample) + minsample)
	return nbuf
}

func Getbuf2() []byte {
	bytes := make([]byte, rand.Intn(maxsample)+minsample)
	for i := 0; i < len(bytes); i++ {
		values := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()-=_+[]\\{}|;'\",./<>?`~          "
		bytes[i] = values[rand.Intn(len(values))]
	}
	return bytes
}

func BitChangeTest() {
	var wh PegHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	wh.Init()
	buf := Getbuf2()
	cnt := 0

	for x := 0; x < 100000000000; x++ {
		// Get a new buffer of data.
		buf = Getbuf2()

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
				wv := wh.Hash(buf)
				g2.Stop()
				g2.AddHash(buf, wv)

				// flipping a bit again repairs it.
				buf[i] = buf[i] ^ bit_to_flip

			}
		}

		if cnt > 1000000 {
			cnt = 0

			fmt.Println()
			g1.Report("1 sha")
			g2.Report("1 wh")

		}

	}
}

func DifferentHashes() {
	var wh PegHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	rand.Seed(13243442344225879)

	wh.Init()
	buf := Getbuf2()
	for i := 1; i < 100000000000; i++ {

		// Get a new buffer of data.
		buf = Getbuf2()

		g1.Start()
		sv := sha256.Sum256(buf)
		g1.Stop()
		g1.AddHash(buf, sv[:])

		g2.Start()
		wv := wh.Hash(buf)
		g2.Stop()
		g2.AddHash(buf, wv)

		if i%1000000 == 0 {
			fmt.Println()
			g1.Report("2 sha")
			g2.Report("2 wh")

		}
	}
}

func main() {
	go BitChangeTest()
	DifferentHashes()
}
