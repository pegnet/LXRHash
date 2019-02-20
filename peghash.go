package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

const (
	firstrand = 0x13ef13156da2756b
	Mapsiz    = 0x400
	HBits     = 0x20
	HMask     = HBits - 1
	maxsample = 1
	minsample = 63
)

type PegHash struct {
	maps [Mapsiz]byte // Integer Offsets
}

// Init()
// We use our own algorithm for initializing the map struct.  This is an fairly large table of
// byte values we use to map bytes to other byte values to enhance the avalanche nature of the hash
// as well as increase the memory footprint of the hash.
func (w *PegHash) Init() {
	dat, err := ioutil.ReadFile("whashmaps.dat")
	if err != nil || len(dat) != Mapsiz {

		// Ah, the data file isn't good for us.  Delete it (if it exists)
		os.Remove("whashmaps.dat")

		// Our own "random" generator that really is just used to shuffle values
		rands := [Mapsiz]int{}
		offset := firstrand
		rand := func(i int) int {
			offset = offset ^ (i << 30) ^ offset<<7 ^ offset>>1&offset>>9 ^ rands[offset&(Mapsiz-1)]
			rands[i] = offset ^ rands[i]
			return rands[i] & (Mapsiz - 1)
		}

		// Fill the maps with bytes ranging from 0 to 255.  As long as Mapsize%256 == 0, this
		// looping and masking works just fine.
		for i := range w.maps {
			w.maps[i] = byte(i)
		}

		// Now what we want to do is just mix it all up.  Take every byte in the maps list, and exchange it
		// for some other byte in the maps list. Note that we do this over and over, mixing and more mixing
		// the maps, but maintaining the ratio of each byte value in the maps list.
		for loops := 0; loops < 15000; loops++ {
			fmt.Println("Pass ", loops)
			for i := range w.maps {
				j := rand(i)
				w.maps[i], w.maps[j] = w.maps[j], w.maps[i]
			}
		}

		// open output file
		fo, err := os.Create("whashmaps.dat")
		if err != nil {
			panic(err)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		// write a chunk
		if _, err := fo.Write(w.maps[:]); err != nil {
			panic(err)
		}

	} else {
		copy(w.maps[:Mapsiz], dat)
	}
}

// Hash()
// Takes a source of bytes, returns a 32 byte (256 bit) hash
func (w PegHash) Hash(src []byte) []byte {
	hashes := [HBits]int64{}
	i := int32(1)
	off1 := int64(len(src)) << 30

	for _, v := range src {
		i0 := (i + 0) & HMask
		i1 := (i + 5) & HMask

		h0 := hashes[i0]
		h1 := hashes[i1]

		// Shift up a byte what is in offset, combined with offset shifted down a bit, combined with a byte and index
		bi := int64(w.maps[(off1^int64(v))&(Mapsiz-1)]) ^ int64(i)
		off1 = (off1 << 7) ^ (off1 >> 1) ^ (^(off1 & h0) >> 9) ^ (bi << 28) ^ h1

		hashes[i0] = (h0 << 7) ^ (h0 >> 1) ^ (h0 >> 9) ^ off1
		hashes[i1] = h1 ^ h0 ^ int64(w.maps[(off1^bi)&(Mapsiz-1)])<<30
		i += 1
	}

	var b byte
	var bytes [HBits]byte

	off2 := off1
	for i, v := range hashes {
		b = byte(v^off1^off2) ^ b
		off1 = off1>>9 ^ off1>>1 ^ off1<<7 ^ v ^ int64(i)
		off2 = off2>>7 ^ off2<<1 ^ off2<<9 ^ v ^ int64(i)
		bytes[i] = w.maps[(int64(w.maps[b])+off1)&(Mapsiz-1)]
	}
	return bytes[:]
}
