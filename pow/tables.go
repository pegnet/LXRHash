// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package pow

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/user"
	"time"
)

// Init initializes the hash with the given values
//
// We use our own algorithm for initializing the map struct.  This is an fairly large table of
// byte values we use to map bytes to other byte values to enhance the avalanche nature of the hash
// as well as increase the memory footprint of the hash.
//
// Bits is the number of bits used to address the ByteMap. If less than 8, set to 8.
// Passes is the number of shuffles of the ByteMap performed.  Each pass shuffles all byte values in the map
func (lx *LxrPow) Init(Bits, Passes uint64) *LxrPow {
	if Bits < 8 {
		Bits = 8
	}
	lx.MapSize = uint64(math.Pow(2, float64(Bits)))
	lx.Passes = Passes
	lx.ReadTable()
	return lx
}

// ReadTable attempts to load the ByteMap from disk.
// If that doesn't exist, a new one will be generated and saved.
func (lx *LxrPow) ReadTable() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	userPath := u.HomeDir
	lxrPowPath := userPath + "/.lxrpow"
	err = os.MkdirAll(lxrPowPath, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Could not create the directory %s", lxrPowPath))
	}
	bits := math.Log2(float64(lx.MapSize))
	filename := fmt.Sprintf(lxrPowPath+"/lxrpow-%x-passes-%d-bits.dat", lx.Passes, int64(bits))
	// Try and load our byte map.
	fmt.Printf("Reading ByteMap Table %s\n", filename)

	start := time.Now()
	dat, err := os.ReadFile(filename)
	// If loading fails, or it is the wrong size, generate it.  Otherwise just use it.
	if err != nil || len(dat) != int(lx.MapSize) {
		fmt.Println("Table not found, Generating ByteMap Table")
		lx.GenerateTable()
		fmt.Println("Writing ByteMap Table ")
		lx.WriteTable(filename)
	} else {
		lx.ByteMap = dat
	}
	fmt.Printf("Finished Reading ByteMap Table. Total time taken: %s\n", time.Since(start))
}

// WriteTable caches the byteMap to disk so it only has to be generated once
func (lx *LxrPow) WriteTable(filename string) {
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
	w := bufio.NewWriter(fo)
	bufSize := 4096 // 4KiB
	for i := 0; i < len(lx.ByteMap); i += bufSize {
		j := i + bufSize
		if j > len(lx.ByteMap) {
			j = len(lx.ByteMap)
		}
		if nn, err := w.Write(lx.ByteMap[i:j]); err != nil {
			panic(fmt.Sprintf("error writing ByteMap to disk: %d bytes written, %v", nn, err))
		}
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}
}

// GenerateTable generates the ByteMap.
// Initializes the map with an incremental sequence of bytes,
// then does P passes, shuffling each element in a deterministic manner.
func (lx *LxrPow) GenerateTable() {
	var offset uint64 = 824798434557
	lx.ByteMap = make([]byte, int(lx.MapSize))
	// Our own "random" generator that really is just used to shuffle values
	MapMask := lx.MapSize - 1
	// The random index used to shuffle the ByteMap is itself computed through the ByteMap table
	// in a deterministic pattern.
	rand := func(i uint64) int64 {
		offset = offset<<9 ^ offset>>1 ^ offset>>7 ^ uint64(lx.ByteMap[(offset)&MapMask])
		return int64(uint64(offset) & MapMask)
	}

	// Fill the ByteMap with bytes ranging from 0 to 255.  As long as MapSize%256 == 0, this
	// looping and masking works just fine.
	fmt.Println("Initializing the Table")
	for i := range lx.ByteMap {
		lx.ByteMap[i] = byte(i)
	}

	// Now what we want to do is just mix it all up.  Take every byte in the ByteMap list, and exchange it
	// for some other byte in the ByteMap list. Note that we do this over and over, mixing and more mixing
	// the ByteMap, but maintaining the ratio of each byte value in the ByteMap list.
	fmt.Println("Shuffling the Table")
	period := time.Now().Unix()
	for loops := 0; loops < int(lx.Passes); loops++ {
		fmt.Printf("Pass %d\n", loops)
		for i := range lx.ByteMap {
			if (i+1)%1000 == 0 && time.Now().Unix()-period > 10 {
				fmt.Printf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete\r",
					i/1024000,
					len(lx.ByteMap)/1024000,
					100*float64(i)/float64(len(lx.ByteMap)))
				period = time.Now().Unix()
			}
			j := rand(uint64(i))
			lx.ByteMap[i], lx.ByteMap[j] = lx.ByteMap[j], lx.ByteMap[i]
		}
		fmt.Printf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete",
			len(lx.ByteMap)/1024000, len(lx.ByteMap)/1024000, float64(100))
	}
}
