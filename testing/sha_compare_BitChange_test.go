package testing_test

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/pegnet/LXRHash"
)

func TestBitChange(t *testing.T) {
	rand.Seed(123412341234)

	Lxrhash.Init(Seed, MapSizeBits, HashSize, Passes)

	rand.Seed(123412341234)

	Lxrhash.Init(Seed, MapSizeBits, HashSize, Passes)
	numTests := 1

	Gradehash{}.PrintHeader()

	ctl := make(chan int, 1)
	res := make(chan int, 1)
	for i := 0; i < numTests; i++ {
		go BitChange(&ctl, &res)
	}

	time.Sleep(2 * time.Minute)

	fmt.Println("*************************** kill *********************************************")
	for i := 0; i < numTests; i++ {
		ctl <- 0
		<-res
	}

	fmt.Println("Diff Count")
	g := Lxrhash.DiffCnt
	all := int64(0)
	for i := 0; i < len(g); i += 1 {
		for j := 0; j < 1 && i+j < len(g); j++ {
			fmt.Printf("%16x %8d", (i+j)<<10, g[i+j])
			all += g[i+j]
		}
		fmt.Println()
	}
	fmt.Println("BCnt ", Lxrhash.BCnt, " counted ", all)

	fmt.Println("Access Count")
	g = Lxrhash.AccessCnt
	all = int64(0)
	for i := 0; i < len(g); i += 1 {
		for j := 0; j < 1 && i+j < len(g); j++ {
			fmt.Printf("%16x %8d", (i+j)<<10, g[i+j])
			all += g[i+j]
		}
		fmt.Println()
	}
	fmt.Println("BCnt ", Lxrhash.BCnt, " counted ", all)

	fmt.Println("Byte Returned Count")
	for i := 0; i < 256; i++ {
		fmt.Printf("%3d %10d\n", i, Lxrhash.BytesRtn[i])
	}

	fmt.Println("Byte Difference Count")
	for i := 0; i < 512; i++ {
		fmt.Printf("%3d %10d\n", i, Lxrhash.DiffBytesRtn[i])
	}

}

func BitChange(ctl, res *chan int) {
	var g1 Gradehash
	var g2 Gradehash

	buf := Getbuf()
	cnt := 0

	last := time.Now().Unix()
	for x := 0; x < 100000000000; x++ {

		select {
		case <-*ctl:
			*res <- 0
			return
		default:
		}

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
				wv := Lxrhash.Hash(buf)
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
