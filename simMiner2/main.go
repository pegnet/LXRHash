package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	lxr "github.com/pegnet/LXRHash"
)

var total uint64
var lx *lxr.LXRHash

func minerRefactor(base []byte, id int) {
	ninc := NewNonceIncrementer(id)
	var data []byte
	for {
		ninc.NextNonce()
		data = append(base, ninc.Nonce...)
		lx.Hash(data)
		atomic.AddUint64(&total, 1)
	}
}

func minerClosures(base []byte, id int) {
	ninc := NewNonceIncrementer(id)
	var data []byte
	for {
		ninc.NextNonce()
		data = append(base, ninc.Nonce...)
		lx.HashClosures(data)
		atomic.AddUint64(&total, 1)
	}
}

func main() {
	lx = new(lxr.LXRHash)
	lx.Init(lxr.Seed, lxr.MapSizeBits, lxr.HashSize, lxr.Passes)

	base := sha256.Sum256([]byte("foo"))

	closures := false
	if len(os.Args) > 1 && os.Args[1] == "true" {
		closures = true
	}

	fmt.Println("Using closures?", closures)
	fmt.Println(runtime.Version())

	start := time.Now()
	for i := 0; i < runtime.NumCPU(); i++ {
		if closures {
			go minerClosures(base[:], i)
		} else {
			go minerRefactor(base[:], i)
		}
	}

	for i := 0; i < 60; i++ {
		time.Sleep(10 * time.Second)
		since := time.Since(start)
		fmt.Printf("%s\t%d h\t%.0f hps\n", since, total, float64(total)/since.Seconds())
	}
}
