package main

import (
	rand2 "crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/pegnet/LXR256"
	"math/rand"
	"time"
)

const (
	maxsample = 1
	minsample = 63
	line      = "--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------"
)

func Getbuf() []byte {
	//buflen := minsample + rand.Intn(maxsample)
	nbuf := make([]byte, 128, 128)
	_, err := rand2.Reader.Read(nbuf)
	if err != nil {
		panic(err)
	}
	return nbuf
}

func Getbuf32() []byte {
	nbuf := make([]byte, 32, 32)
	_, err := rand2.Reader.Read(nbuf)
	if err != nil {
		panic(err)
	}
	return nbuf
}

func BitCountTest(Seed, MaxSize int64, HashSize, Passes, rate int) {
	var wh lxr.LXRHash
	var g1 lxr.Gradehash
	var g2 lxr.Gradehash

	wh.Init(Seed, MaxSize, int(HashSize), Passes)
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

			if cnt >= rate {
				cnt = 0

				g1.Report("cnt-sha")
				g2.Report("cnt-lxr")
				fmt.Print(line)

			}
		}

	}
}

func AddByteTest(Seed, MaxSize int64, HashSize, Passes, rate int) {
	var wh lxr.LXRHash
	var g1 lxr.Gradehash
	var g2 lxr.Gradehash

	wh.Init(Seed, MaxSize, int(HashSize), Passes)
	buf := Getbuf()
	cnt := 0

	for x := 0; x < 100000000000; x++ {
		// Get a new buffer of data.
		buf = []byte{byte(x)}

		for i := 0; i < 1000; i++ {
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

			if cnt > rate {
				cnt = 0

				g1.Report("add-sha")
				g2.Report("add-lxr")
				fmt.Print(line)

			}
		}

	}
}

func BitChangeTest(Seed, MaxSize int64, HashSize, Passes, rate int) {
	var wh lxr.LXRHash
	var g1 lxr.Gradehash
	var g2 lxr.Gradehash

	wh.Init(Seed, MaxSize, int(HashSize), Passes)
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

				if cnt > rate {
					cnt = 0

					g1.Report("bit-sha")
					g2.Report("bit-lxr")
					fmt.Print(line)

				}
			}

		}

	}
}

func DifferentHashes(Seed, MaxSize int64, HashSize, Passes, rate int) {
	var wh lxr.LXRHash
	var g1 lxr.Gradehash
	var g2 lxr.Gradehash

	wh.Init(Seed, MaxSize, int(HashSize), Passes)
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
			g2.Report("diff-lxr")
			fmt.Print(line)

		}

	}
}

func main() {
	rand.Seed(123412341234)

	Seed := int64(123412341234)
	MaxSize := int64(1024)
	Passes := 5
	rate := 100000
	HashSize := 256
	_ = rate

	lxrHash := lxr.LXRHash{}
	lxrHash.Init(Seed, MaxSize, int(HashSize), Passes)
	lxr.Gradehash{}.PrintHeader()

	//go BitCountTest(rate)
	go BitChangeTest(Seed, MaxSize, HashSize, Passes, rate)
	//go DifferentHashes(rate)
	//go AddByteTest(rate)

	for {
		time.Sleep(1 * time.Second)
	}
}
