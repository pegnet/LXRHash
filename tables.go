package lxr

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// constants for building different sized lookup tables (ByteMap).  Right now, the lookup table is hard coded as
// a 1K table, but it can be far larger.
const (
	firstrand = uint64(2458719153079158768)
	firstb    = uint64(4631534797403582785)
	firstv    = uint64(3523455478921636871)
)

// Init()
// We use our own algorithm for initializing the map struct.  This is an fairly large table of
// byte values we use to map bytes to other byte values to enhance the avalanche nature of the hash
// as well as increase the memory footprint of the hash.
//
// Seed is a 64 bit starting point
// MapSizeBits is the number of bits to use for the MapSize, i.e. 10 = mapsize of 1024
// HashSize is the number of bits in the hash; truncated to a byte bountry
// Passes is the number of shuffles of the ByteMap performed.  Each pass shuffles all byte values in the map

func (w *LXRHash) Init(Seed, MapSizeBits, HashSize, Passes uint64) {

	if MapSizeBits < 8 {
		panic(fmt.Sprintf("Bad Map Size in Bits.  Must be between 8 and 34 bits, was %d", MapSizeBits))
	}

	MapSize := uint64(1) << MapSizeBits
	w.ByteMap = make([]byte, int(MapSize))

	w.HashSize = (HashSize + 7) / 8
	w.MapSize = MapSize
	w.MapSizeBits = MapSizeBits
	w.Seed = Seed
	w.Passes = Passes
	w.ReadTable()

}

// ReadTable
// When a lookup table is on the disk, this will allow one to read it.
func (w *LXRHash) ReadTable() {
	filename := fmt.Sprintf("lrxhash.seed-%x.passes-%d.size-%d.dat", w.Seed, w.Passes, w.MapSizeBits)
	// Try and load our byte map.
	println("Reading ByteMap Table ", filename)
	dat, err := ioutil.ReadFile(filename)

	// If loading fails, or it is the wrong size, generate it.  Otherwise just use it.
	if err != nil || len(dat) != int(w.MapSize) {
		println("Table not found, Generating ByteMap Table ")
		w.GenerateTable()
		fmt.Println("writeing ByteMap Table ")
		w.WriteTable(filename)
		fmt.Println("Done")
	} else {
		w.ByteMap = dat
	}
}

// WriteTable
// When playing around with the algorithm, it is nice to generate files and use them off the disk.  This
// allows the user to do that, and save the cost of regeneration between test runs.
func (w *LXRHash) WriteTable(filename string) {
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
	if _, err := fo.Write(w.ByteMap[:]); err != nil {
		panic(err)
	}

}

// GenerateTable
// Build a table with a rather simplistic but with many passes, adequately randomly ordered bytes.
// We do some straight forward bitwise math to initialize and scramble our ByteMap.
func (w *LXRHash) GenerateTable() {

	// Our own "random" generator that really is just used to shuffle values
	offset := w.Seed ^ firstrand
	b := w.Seed ^ firstb
	v := firstv
	MapMask := w.MapSize - 1
	// The random index used to shuffle the ByteMap is itself computed through the ByteMap table
	// in a deterministic pattern.
	rand := func(i uint64) int64 {
		offset = offset<<9 ^ offset>>1 ^ offset>>7 ^ b
		v = uint64(w.ByteMap[(offset^b)&MapMask]) ^ v<<8 ^ v>>1
		b = v<<7 ^ v<<13 ^ v<<33 ^ v<<52 ^ b<<9 ^ b>>1
		return int64(uint64(offset) % uint64(w.MapSize))
	}

	start := time.Now().Unix()
	period := start
	// Fill the ByteMap with bytes ranging from 0 to 255.  As long as Mapsize%256 == 0, this
	// looping and masking works just fine.
	println("Initalize the Table")
	for i := range w.ByteMap {
		if (i+1)%1000 == 0 && time.Now().Unix()-period > 10 {
			println(" Index ", i+1, " of ", len(w.ByteMap))
			period = time.Now().Unix()
		}
		w.ByteMap[i] = byte(i)
	}

	// Now what we want to do is just mix it all up.  Take every byte in the ByteMap list, and exchange it
	// for some other byte in the ByteMap list. Note that we do this over and over, mixing and more mixing
	// the ByteMap, but maintaining the ratio of each byte value in the ByteMap list.
	println("Shuffling the Table")
	for loops := 0; loops < int(w.Passes); loops++ {
		fmt.Println("Pass ", loops)
		for i := range w.ByteMap {
			if (i+1)%1000 == 0 && time.Now().Unix()-period > 10 {
				fmt.Printf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete\n", i/1024000, len(w.ByteMap)/1024000, 100*float64(i)/float64(len(w.ByteMap)))
				period = time.Now().Unix()
			}

			j := rand(uint64(i))
			w.ByteMap[i], w.ByteMap[j] = w.ByteMap[j], w.ByteMap[i]
		}
		fmt.Printf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete\n", len(w.ByteMap)/1024000, len(w.ByteMap)/1024000, float64(100))
	}
}
