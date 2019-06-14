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
	line = "--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------"
)

var lxrhash lxr.LXRHash

func Getbuf() []byte {
	//buflen := minsample + rand.Intn(maxsample)
	nbuf := make([]byte, 1024, 1024)
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

func BitCountTest() {
	var g1 lxr.Gradehash
	var g2 lxr.Gradehash

	buf := Getbuf()
	cnt := 0

	last := time.Now().Unix()
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
			wv := lxrhash.Hash(buf)
			g2.Stop()
			g2.AddHash(buf, wv)

			if cnt > 1000 && time.Now().Unix()-last > 4  {
				last = time.Now().Unix()
				cnt = 0

				g1.Report("bit-sha")
				g2.Report("bit-lxr")
				fmt.Print(line)

			}
		}

	}
}

func AddByteTest() {
	var g1 lxr.Gradehash
	var g2 lxr.Gradehash

	buf := Getbuf()
	cnt := 0
	last := time.Now().Unix()
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
			wv := lxrhash.Hash(buf)
			g2.Stop()
			g2.AddHash(buf, wv)

			buf = append(buf, byte(rand.Intn(255)))

			if cnt > 1000 && time.Now().Unix()-last > 4  {
				last = time.Now().Unix()
				cnt = 0

				g1.Report("bit-sha")
				g2.Report("bit-lxr")
				fmt.Print(line)

			}
		}

	}
}

func BitChangeTest() {
	var g1 lxr.Gradehash
	var g2 lxr.Gradehash

	buf := Getbuf()
	cnt := 0

	last := time.Now().Unix()
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
				wv := lxrhash.Hash(buf)
				g2.Stop()
				g2.AddHash(buf, wv)

				// flipping a bit again repairs it.
				buf[i] = buf[i] ^ bit_to_flip

				if cnt > 1000 && time.Now().Unix()-last > 4  {
					last = time.Now().Unix()
					cnt = 0

					g1.Report("bit-sha")
					g2.Report("bit-lxr")
					fmt.Print(line)

				}
			}

		}

	}
}

func DifferentHashes() {
	var g1 lxr.Gradehash
	var g2 lxr.Gradehash

	buf := Getbuf()

	last := time.Now().Unix()
	cnt :=0
	for i := 1; i < 100000000000; i++ {

		// Get a new buffer of data.
		buf = Getbuf()

		g1.Start()
		sv := sha256.Sum256(buf)
		g1.Stop()
		g1.AddHash(buf, sv[:])

		g2.Start()
		wv := lxrhash.Hash(buf)
		g2.Stop()
		g2.AddHash(buf, wv)
		cnt++
		if cnt > 1000 && time.Now().Unix()-last > 4  {
			last = time.Now().Unix()
			cnt = 0

			g1.Report("bit-sha")
			g2.Report("bit-lxr")
			fmt.Print(line)

		}
	}
}

// GenAll()
// Generate all the map files for a particular seed, and number of passes
func GenAll(Seed, Passes uint64) {

	for i:= uint64(8); i < 33; i++ {
		meg := float64(uint64(1)<<i)/1000000
		fmt.Printf("Processing Map of %12.4f MB, %d bits\n",meg, i)
		lxrHash := lxr.LXRHash{}
		lxrHash.Init(Seed, i, 255, Passes)
	}
}



func main() {
	rand.Seed(123412341234)

	Seed := uint64(46898902133)
	MaxSizeBits := uint64(33)
	Passes := uint64(5)
	HashSize := uint64(256)

	lxrhash.Init(Seed, MaxSizeBits, HashSize, Passes)

	//GenAll(Seed,Passes)

	lxr.Gradehash{}.PrintHeader()

	//go BitCountTest(rate)
	go BitChangeTest()


	//go DifferentHashes(rate)
	//go AddByteTest(rate)

	for {
		time.Sleep(1 * time.Second)
	}
}
