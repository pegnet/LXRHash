// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
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

// Init initializes the hash with the given values
//
// We use our own algorithm for initializing the map struct.  This is an fairly large table of
// byte values we use to map bytes to other byte values to enhance the avalanche nature of the hash
// as well as increase the memory footprint of the hash.
//
// Seed is a 64 bit starting point
// MapSizeBits is the number of bits to use for the MapSize, i.e. 10 = mapsize of 1024
// HashSize is the number of bits in the hash; truncated to a byte bountry
// Passes is the number of shuffles of the ByteMap performed.  Each pass shuffles all byte values in the map
//
// Panics when MapSizeBits is < 8 and on other error conditions
func (lx *LXRHash) Init(Seed, MapSizeBits, HashSize, Passes uint64) {
	tablePath, err := GetUserTablePath()
	if err != nil {
		panic(err)
	}
	if _, err = lx.initFromPath(Seed, MapSizeBits, HashSize, Passes, tablePath); err != nil {
		panic(err)
	}
}

// Initialize the hash with the given values, reading the hash table from the specified path
//
// Seed - a 64-bit starting value
// MapSizeBits - size of the map as an exponent of 2, i.e., 10 -> map size of 2 ^ 10 = 1024; between 8 and 34 bits inclusive.
// HashSize - number of bits in the hash, truncated to a byte boundary
// Passes - number of shuffles of the ByteMap
// TablePath - the file system path of the directory which holds hash table files
//
// Because hash table files can be large, sharing them between applications is desirable.  Suggested shared hash table paths:
//   Windows: %ProgramData%/LXRHash
//   macOS:   /Library/Application\ Support/org.pegnet.LXRHash
//   Linux:   /var/lib/LXRHash
//   BSD:     /var/db/LXRHash
//
func (lx *LXRHash) InitFromPath(Seed, MapSizeBits, HashSize, Passes uint64, TablePath string) (string, error) {
	if _, err := os.Stat(TablePath); err != nil {
		return "", err
	}
	return lx.initFromPath(Seed, MapSizeBits, HashSize, Passes, TablePath)
}

// ReadTable attempts to load the ByteMap from disk.
// If that doesn't exist, a new one will be generated and saved.
func (lx *LXRHash) ReadTable() {
	lxrHashPath, err := GetUserTablePath()
	if err != nil {
		panic(err)
	}
	if err = os.MkdirAll(lxrHashPath, os.ModePerm); err != nil {
		panic(fmt.Sprintf("Could not create the directory %s", lxrHashPath))
	}
	if _, err = lx.readTableFromPath(lxrHashPath); err != nil {
		panic(err)
	}
}

// WriteTable caches the bytemap to disk so it only has to be generated once
func (lx *LXRHash) WriteTable(filename string) {
	if err := lx.writeTable(filename); err != nil {
		panic(err)
	}
}

// GenerateTable generates the bytemap.
// Initializes the map with an incremental sequence of bytes,
// then does P passes, shuffling each element in a deterministic manner.
func (lx *LXRHash) GenerateTable() {
	lx.ByteMap = make([]byte, int(lx.MapSize))
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

func (lx *LXRHash) initFromPath(Seed, MapSizeBits, HashSize, Passes uint64, TablePath string) (string, error) {
	if MapSizeBits < 8 {
		return "", fmt.Errorf("Bad Map Size in Bits.  Must be between 8 and 34 bits, was %d", MapSizeBits)
	}

	MapSize := uint64(1) << MapSizeBits
	lx.HashSize = (HashSize + 7) / 8
	lx.MapSize = MapSize
	lx.MapSizeBits = MapSizeBits
	lx.Seed = Seed
	lx.Passes = Passes
	lxrhashtablepath, err := lx.readTableFromPath(TablePath)
	if err != nil {
		return "", err
	}
	return lxrhashtablepath, nil
}

func (lx *LXRHash) readTableFromPath(tablepath string) (string, error) {
	filename := fmt.Sprintf("lxrhash-seed-%x-passes-%d-size-%d.dat", lx.Seed, lx.Passes, lx.MapSizeBits)
	filepath := path.Join(tablepath, filename)

	// Try and load our byte map.
	lx.Log(fmt.Sprintf("Reading ByteMap Table %s", filepath))

	start := time.Now()
	dat, err := ioutil.ReadFile(filepath)
	// If loading fails, or it is the wrong size, generate it.  Otherwise just use it.
	if err != nil || len(dat) != int(lx.MapSize) {
		lx.Log("Table not found, Generating ByteMap Table")
		lx.GenerateTable()
		lx.Log("Writing ByteMap Table ")
		if err := lx.writeTable(filepath); err != nil {
			return "", err
		}
	} else {
		lx.ByteMap = dat
	}
	lx.Log(fmt.Sprintf("Finished Reading ByteMap Table. Total time taken: %s", time.Since(start)))
	return filepath, nil
}

func (lx *LXRHash) writeTable(filepath string) (result error) {
	tablePath := path.Dir(filepath)
	if err := os.MkdirAll(tablePath, os.ModePerm); err != nil {
		return fmt.Errorf("Could not create the directory %s", tablePath)
	}
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return err
	}
	// open output file
	fo, err := os.Create(filepath)
	if err != nil {
		return err
	}

	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			result = err
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
			return fmt.Errorf("error writing bytemap to disk: %d bytes written, %v", nn, err)
		}
	}
	err = w.Flush()
	return err
}

func GetUserTablePath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	userTablePath := filepath.Join(u.HomeDir, ".lxrhash")
	return userTablePath, nil
}
