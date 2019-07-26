package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"

	lxr "github.com/pegnet/LXRHash"
)

var total uint64
var now = time.Now()

var prt chan string

func mine(useLXR bool, data []byte) uint64 {

	LX := new(lxr.LXRHash)
	LX.Init(lxr.Seed, lxr.MapSizeBits, lxr.HashSize, lxr.Passes)

	cd := uint64(0)
	dlen := len(data)
	for i := 0; i < 100000000; i++ {
		data = data[:dlen]
		for b := i; b > 0; b = b >> 8 {
			data = append(data, byte(b))
		}
		var hash []byte
		if useLXR {
			hash = LX.Hash(data)
		} else {
			h := sha256.Sum256(data)
			hash = h[:]
		}
		d := uint64(0)
		for i := 0; i < 8; i++ {
			d = d<<8 + uint64(hash[i])
		}
		if cd < d {
			cd = d
			running := time.Since(now)
			hps := float64(total) / running.Seconds()
			prt <- fmt.Sprintf("%10d %16x %8x %10.1f hps\n", total, cd, i, hps)

		}
	}
	return cd
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage:\n\nJustHash <hash>\n\n<hash> is equal to LXRHash to sim mine LXRHash\n<hash> is equal to Sha256 to sim mine Sha256")
		os.Exit(0)
	}

	h := strings.ToLower(os.Args[1])
	hash := h == "lxrhash"
	if !hash && h != "sha256" {
		fmt.Println("Usage:\n\nJustHash <hash>\n\n<hash> is equal to LXRHash to sim mine LXRHash\n<hash> is equal to Sha256 to sim mine Sha256")
	}

	prt = make(chan string, 500)
	go mine(hash, []byte("000000000200000000020000000002000"))
	go mine(hash, []byte("000000000200000000020000000002001"))
	go mine(hash, []byte("000000000200000000020000000002002"))
	go mine(hash, []byte("000000000200000000020000000002003"))

	go mine(hash, []byte("000000000200000000020000000002004"))
	go mine(hash, []byte("000000000200000000020000000002005"))
	go mine(hash, []byte("000000000200000000020000000002006"))
	go mine(hash, []byte("000000000200000000020000000002006"))

	for {
		select {
		case s := <-prt:
			fmt.Print(s)
			continue
		default:
		}
		time.Sleep(10 * time.Second)
		running := time.Since(now)
		hps := float64(total) / running.Seconds()
		prt <- fmt.Sprintf("%10d %16s %8s %10.1f hps\n", total, "", "", hps)

	}
}
