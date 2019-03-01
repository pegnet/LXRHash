package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/dustin/go-humanize"
	"math/rand"
	"time"
	"github.com/pegnet/LXR256"
)

const (
	maxsample = 1
	minsample = 63
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
	var lx lxr.LXRHash
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

		for time.Now().Unix()-start < 60/3 {
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

				d := lxr.Difficulty(wv)
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
		fmt.Printf("%5s Average Difficulty %25f  total hashes %28s #blocks %d \n",
			tag,
			sum/float64(i),
			humanize.Comma(int64(hashes)), i)

	}
}



func main() {

	go JustMine("SHA==")
	go JustMine("lxr--")

	// wait forever
	for {
		time.Sleep(1 * time.Second)
	}
}
