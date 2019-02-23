package main

import (
	"crypto/sha256"
	"github.com/FactomProject/factomd/common/primitives/random"
	"math/rand"
	"time"
)

func Getbuf() []byte {
	nbuf := random.RandByteSliceOfLen(rand.Intn(maxsample) + minsample)
	return nbuf
}

func BitCountTest(rate int) {
	var wh PegHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	wh.Init()
	buf := Getbuf()
	cnt := 0

	for x := 0; x < 100000000000; x++ {
		// Get a new buffer of data.
		buf = Getbuf()

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
			wv := wh.Hash(buf)
			g2.Stop()
			g2.AddHash(buf, wv)
			time.Sleep(time.Millisecond)
			if cnt >= rate {
				cnt = 0

				g1.Report("cnt-sha")
				g2.Report("cnt- wh")

			}
		}

	}
}

func AddByteTest(rate int) {
	var wh PegHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	wh.Init()
	buf := Getbuf()
	cnt := 0

	for x := 0; x < 100000000000; x++ {
		// Get a new buffer of data.
		buf = Getbuf()

		for i := 0; i < 10000; i++ {
			cnt++

			g1.Start()
			sv := sha256.Sum256(buf)
			g1.Stop()
			g1.AddHash(buf, sv[:])

			g2.Start()
			wv := wh.Hash(buf)
			g2.Stop()
			g2.AddHash(buf, wv)

			buf = append(buf, byte(rand.Intn(255)))
			time.Sleep(time.Millisecond)
			if cnt > rate {
				cnt = 0

				g1.Report("add-sha")
				g2.Report("add- wh")

			}
		}

	}
}

func BitChangeTest(rate int) {
	var wh PegHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	wh.Init()
	buf := Getbuf()
	cnt := 0

	for x := 0; x < 100000000000; x++ {
		// Get a new buffer of data.
		buf = Getbuf()

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
				time.Sleep(time.Millisecond)

				if cnt > rate {
					cnt = 0

					g1.Report("bit-sha")
					g2.Report("bit- wh")

				}
			}

		}

	}
}

func DifferentHashes(rate int) {
	var wh PegHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	rand.Seed(13243442344225879)

	wh.Init()
	buf := Getbuf()
	for i := 1; i < 100000000000; i++ {

		// Get a new buffer of data.
		buf = Getbuf()

		g1.Start()
		sv := sha256.Sum256(buf)
		g1.Stop()
		g1.AddHash(buf, sv[:])

		g2.Start()
		wv := wh.Hash(buf)
		g2.Stop()
		g2.AddHash(buf, wv)

		if i%rate == 0 {

			g1.Report("diff-sha")
			g2.Report("diff- wh")

		}
		time.Sleep(time.Millisecond)
	}
}

func main() {
	rate := 1000

	go BitCountTest(rate)
	go BitChangeTest(rate)
	go DifferentHashes(rate)
	go AddByteTest(rate)

	// wait forever
	for {
		time.Sleep(1 * time.Second)
	}
}
