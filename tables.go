// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"time"
)

// constants for building different sized lookup tables (ByteMap).  Right now, the lookup table is hard coded as
// a 1K table, but it can be far larger.
const (
	firstrand = uint64(2458719153079158768)
	firstb    = uint64(4631534797403582785)
	firstv    = uint64(3523455478921636871)
)

// Verbose enables or disables the output of progress indicators to the console
func (lx *LXRHash) Verbose(val bool) {
	lx.verbose = val
}

// Log is a wrapper function that only prints information when verbose is enabled
func (lx *LXRHash) Log(msg string) {
	if lx.verbose {
		fmt.Println(msg)
	}
}

// Init()
// We use our own algorithm for initializing the map struct.  This is an fairly large table of
// byte values we use to map bytes to other byte values to enhance the avalanche nature of the hash
// as well as increase the memory footprint of the hash.
//
// Seed is a 64 bit starting point
// MapSizeBits is the number of bits to use for the MapSize, i.e. 10 = mapsize of 1024
// HashSize is the number of bits in the hash; truncated to a byte bountry
// Passes is the number of shuffles of the ByteMap performed.  Each pass shuffles all byte values in the map
func (lx *LXRHash) Init(Seed, MapSizeBits, HashSize, Passes uint64) {
	if MapSizeBits < 8 {
		panic(fmt.Sprintf("Bad Map Size in Bits.  Must be between 8 and 34 bits, was %d", MapSizeBits))
	}

	MapSize := uint64(1) << MapSizeBits
	lx.ByteMap = make([]byte, int(MapSize))

	lx.HashSize = (HashSize + 7) / 8
	lx.MapSize = MapSize
	lx.MapSizeBits = MapSizeBits
	lx.Seed = Seed
	lx.Passes = Passes
	lx.ReadTable()

}

// ReadTable
// When a lookup table is on the disk, this will allow one to read it.
func (lx *LXRHash) ReadTable() {

	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	userPath := u.HomeDir
	lxrhashPath := userPath + "/.lxrhash"
	err = os.MkdirAll(lxrhashPath, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Could not create the directory %s", lxrhashPath))
	}

	filename := fmt.Sprintf(lxrhashPath+"/lrxhash-seed-%x-passes-%d-size-%d.dat", lx.Seed, lx.Passes, lx.MapSizeBits)
	// Try and load our byte map.
	lx.Log(fmt.Sprintf("Reading ByteMap Table %s", filename))

	start := time.Now()
	dat, err := ioutil.ReadFile(filename)
	// If loading fails, or it is the wrong size, generate it.  Otherwise just use it.
	if err != nil || len(dat) != int(lx.MapSize) {
		lx.Log("Table not found, Generating ByteMap Table")
		lx.GenerateTable()
		lx.Log("Writing ByteMap Table ")
		lx.WriteTable(filename)
	} else {
		lx.ByteMap = dat
	}
	lx.Log(fmt.Sprintf("Done. Total time taken: %s", time.Since(start)))
}

// WriteTable
// When playing around with the algorithm, it is nice to generate files and use them off the disk.  This
// allows the user to do that, and save the cost of regeneration between Lxrhash runs.
func (lx *LXRHash) WriteTable(filename string) {
	// Ah, the data file isn't good for us.  Delete it (if it exists)
	os.Remove(filename)

	// open output file
	fo, err := os.Create(filename)
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
	if _, err := fo.Write(lx.ByteMap[:]); err != nil {
		panic(err)
	}

}

// GenerateTable
// Build a table with a rather simplistic but with many passes, adequately randomly ordered bytes.
// We do some straight forward bitwise math to initialize and scramble our ByteMap.
func (lx *LXRHash) GenerateTable() {

	// Our own "random" generator that really is just used to shuffle values
	offset := lx.Seed ^ firstrand
	b := lx.Seed ^ firstb
	v := firstv
	MapMask := lx.MapSize - 1
	// The random index used to shuffle the ByteMap is itself computed through the ByteMap table
	// in a deterministic pattern.
	rand := func(i uint64) int64 {
		offset = offset<<9 ^ offset>>1 ^ offset>>7 ^ b
		v = uint64(lx.ByteMap[(offset^b)&MapMask]) ^ v<<8 ^ v>>1
		b = v<<7 ^ v<<13 ^ v<<33 ^ v<<52 ^ b<<9 ^ b>>1
		return int64(uint64(offset) & uint64(MapMask))
	}

	// Fill the ByteMap with bytes ranging from 0 to 255.  As long as Mapsize%256 == 0, this
	// looping and masking works just fine.
	lx.Log("Initializing the Table")
	for i := range lx.ByteMap {
		lx.ByteMap[i] = byte(i)
	}

	// Now what we want to do is just mix it all up.  Take every byte in the ByteMap list, and exchange it
	// for some other byte in the ByteMap list. Note that we do this over and over, mixing and more mixing
	// the ByteMap, but maintaining the ratio of each byte value in the ByteMap list.
	lx.Log("Shuffling the Table")
	period := time.Now().Unix()
	for loops := 0; loops < int(lx.Passes); loops++ {
		lx.Log(fmt.Sprintf("Pass %d", loops))
		for i := range lx.ByteMap {
			if (i+1)%1000 == 0 && time.Now().Unix()-period > 10 {
				lx.Log(fmt.Sprintf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete", i/1024000, len(lx.ByteMap)/1024000, 100*float64(i)/float64(len(lx.ByteMap))))
				period = time.Now().Unix()
			}

			j := rand(uint64(i))
			lx.ByteMap[i], lx.ByteMap[j] = lx.ByteMap[j], lx.ByteMap[i]
		}
		lx.Log(fmt.Sprintf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete", len(lx.ByteMap)/1024000, len(lx.ByteMap)/1024000, float64(100)))
	}
}
