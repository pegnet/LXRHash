package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

const (
	firstrand = 0x13ef13156da2756b
	Mapsiz    = 0x400
	MapMask   = Mapsiz - 1
	HBits     = 0x20
	HMask     = HBits - 1
	maxsample = 1
	minsample = 63
)

type PegHash struct {
	maps [Mapsiz]byte // Integer Offsets
	good bool
}

// generateAndWrite
// If we do not have a file with our already computed bytes, then what we want to do
// is do bitwise math to initialize and scramble our maps.  Once we have done this, we
// write out the file.  If we have the file already, then we don't need to do this.
func (w *PegHash) generateAndWrite() {
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
	for loops := 0; loops < 200000; loops++ {
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

}

// Init()
// We use our own algorithm for initializing the map struct.  This is an fairly large table of
// byte values we use to map bytes to other byte values to enhance the avalanche nature of the hash
// as well as increase the memory footprint of the hash.
func (w *PegHash) Init() {

	// Try and load our byte map.
	dat, err := ioutil.ReadFile("whashmaps.dat")

	// If loading fails, or it is the wrong size, generate it.  Otherwise just use it.
	if err != nil || len(dat) != Mapsiz {
		w.generateAndWrite()
	} else {
		copy(w.maps[:Mapsiz], dat)
	}
}

// Hash()
// Takes a source of bytes, returns a 32 byte (256 bit) hash
func (w PegHash) Hash(src []byte) []byte {

	// Keep the 32 byte intermediate result as int64 values until reduced.
	var hashes [HBits]int64
	// The intital offset into the lookup table is the length of the input.
	// This prevents concatenation attacks, which adds to the protection from
	// the reduction pass.
	var offset = int64(len(src))
	// We keep a series of previous states, and roll them along through each
	// byte of source processed.
	var last1, last2, last3 int64

	v := byte(offset)
	// Pass through the source bytes, building up lastX values, hashes[], and offset
	for i, v2 := range src {
		// Take the byte from source (v2) and map it through the lookup table
		// using the offset being maintained, and the rolling lastX values
		v = w.maps[(offset^int64(v2)^last1^last2^last3)&MapMask] ^ v

		// Roll the set of last values, leaving lingering influences from past
		// values.
		last3 = last2>>2 ^ last3 ^ last3>>5
		last2 = last1<<3 ^ last2 ^ last2>>7
		last1 = int64(v2) ^ last1<<1 ^ last1>>3

		// Set one of the hashes[] using the last rolling value, the input byte v2,
		// the mapped byte v, and the previous hashes[] value
		h := hashes[i&HMask]
		hashes[i&HMask] = last3 ^ int64(v^v2) ^ h

		// combine the l values, the previous offset, and the hashes[i]
		offset = last1<<7 ^ last2<<3 ^ last3<<9 ^ offset<<8 ^ offset>>1 ^ h
	}

	// Reduction pass
	// Done by Interating over hashes[] to produce the bytes[] hash
	//
	// At this point, we have HBits of state in hashes.  We need to reduce them down to a byte,
	// And we do so by doing a bit more bitwise math, and mapping the values through our byte map.

	var bytes [HBits]byte

	// Roll over all the hashes (32 int64 values)
	for i, h := range hashes {
		// Map each h using the offset and rolling values
		v := w.maps[(offset^h^last1^last2^last3)&MapMask]
		// Roll the last values
		last3 = last2>>2 ^ last3
		last2 = last1<<3 ^ last2
		last1 = h ^ last1<<1

		// Set a byte
		bytes[i] = w.maps[(int64(v)^offset)&MapMask]

		// combine the l values, the previous offset, and the hashes[i]
		offset = last1<<7 ^ last2<<3 ^ last3<<9 ^ offset<<8 ^ offset>>1 ^ h
	}

	// Return the resulting hash
	return bytes[:]
}
