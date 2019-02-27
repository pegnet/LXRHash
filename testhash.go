package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/dustin/go-humanize"
	"math/rand"
	"time"
)

func Getbuf() []byte {
	nbuf := random.RandByteSliceOfLen(rand.Intn(maxsample) + minsample)
	return nbuf
}

func Getbuf32() []byte {
	nbuf := random.RandByteSliceOfLen(32)
	return nbuf
}

func JustMine(tag string) {
	var lx LXRHash
	lx.Init()

	var sum float64

	for i := 1; i < 1000000; i++ {
		buf := Getbuf32()   // Mine this
		nonce := Getbuf32() // Nonce for mining
		nonce = append(nonce, buf...)
		for i := 0; i < 8; i++ {
			nonce[i] = 0
		}

		bestNonce := []byte{}
		bestNonce = append(bestNonce, nonce...)
		diff := uint64(0)
		cnt := 0
		hashes := 0
		start := time.Now().Unix()

		for time.Now().Unix()-start < 8*60 {
			for i := 0; i < 100000; i++ {
				// increment the nonce
				for i := 0; ; i++ {
					nonce[i] += 1
					if nonce[i] != 0 {
						break
					}
				}

				cnt++

				var wv []byte
				if tag == "SHA==" {
					wv2 := sha256.Sum256(nonce)
					wv = wv2[:]
				} else {
					wv = lx.Hash(nonce)
				}

				d := difficulty(wv)
				if d == 0 {
					continue
				}
				if diff == 0 || diff > d {
					secs := time.Now().Unix() - start
					min := secs / 60
					secs = secs % 60
					diff = d
					bestNonce = append(bestNonce[:0], nonce...)
					fmt.Printf("%5s,%15s %2d:%02d Difficulty %28s [nonce %32x src %32x] hash %32x \n",
						tag,
						humanize.Comma(int64(cnt)),
						min, secs,
						humanize.Comma(int64(diff)),
						nonce[:32], nonce[32:],
						wv)
				}
			}
		}
		hashes += cnt
		sum += float64(diff) / float64(cnt)
		fmt.Printf("%5s Average Difficulty %25f  total hashes %20d #blocks %d \n", tag, sum/float64(i), hashes, i)

	}
}

func BitCountTest(rate int) {
	var wh LXRHash
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

			if cnt >= rate {
				cnt = 0

				g1.Report("cnt-sha")
				g2.Report("cnt- wh")

			}
		}

	}
}

func AddByteTest(rate int) {
	var wh LXRHash
	wh.Init()
	var g1 Gradehash
	var g2 Gradehash

	wh.Init()
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
				g2.Report("add- wh")

			}
		}

	}
}

func BitChangeTest(rate int) {
	var wh LXRHash
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
	var wh LXRHash
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

	}
}

func main() {
	rate := 1000000
	_ = rate

	//go BitCountTest(rate)
	//go BitChangeTest(rate)
	//go DifferentHashes(rate)
	//go AddByteTest(rate)

	go JustMine("SHA==")
	go JustMine("lxr--")
	// wait forever
	for {
		time.Sleep(1 * time.Second)
	}
}
