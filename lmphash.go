package lxr

import (
	"fmt"
	"io"
	"os"
	"os/user"
)

type LMPHash struct {
	fp          *os.File // File Pointer to ByteMap
	MapSize     uint64   // Size of the translation table
	MapSizeBits uint64   // Size of the ByteMap in Bits
	Passes      uint64   // Passes to generate the rand table
	Seed        uint64   // An arbitrary number used to create the tables.
	HashSize    uint64   // Number of bytes in the hash
	verbose     bool
}

// Verbose enables or disables the output of progress indicators to the console
func (lmp *LMPHash) Verbose(val bool) {
	lmp.verbose = val
}

// Log is a wrapper function that only prints information when verbose is enabled
func (lmp *LMPHash) Log(msg string) {
	if lmp.verbose {
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
func (lmp *LMPHash) Init(Seed, MapSizeBits, HashSize, Passes uint64) {
	if MapSizeBits < 8 {
		panic(fmt.Sprintf("Bad Map Size in Bits.  Must be between 8 and 34 bits, was %d", MapSizeBits))
	}

	MapSize := uint64(1) << MapSizeBits
	lmp.HashSize = (HashSize + 7) / 8
	lmp.MapSize = MapSize
	lmp.MapSizeBits = MapSizeBits
	lmp.Seed = Seed
	lmp.Passes = Passes
	lmp.openFile()
}

func (lmp *LMPHash) openFile() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	userPath := u.HomeDir
	lxrhashPath := userPath + "/.lxrhash"

	filename := fmt.Sprintf(lxrhashPath+"/lxrhash-seed-%x-passes-%d-size-%d.dat", lmp.Seed, lmp.Passes, lmp.MapSizeBits)
	fp, err := os.Open(filename)
	if err != nil {
		panic("file doesn't exist -- file generation not supported in low mem profile hash")
	}
	lmp.fp = fp
	lmp.Log(fmt.Sprintf("ByteMap file loaded into memory"))
}

func (lmp LMPHash) Hash(src []byte) []byte {
	// Keep the byte intermediate results as int64 values until reduced.
	hs := make([]uint64, lmp.HashSize)
	// as accumulates the state as we walk through applying the source data through the lookup map
	// and combine it with the state we are building up.
	var as = lmp.Seed
	// We keep a series of states, and roll them along through each byte of source processed.
	var s1, s2, s3 uint64
	// Since MapSize is specified in bits, the index mask is the size-1
	mk := lmp.MapSize - 1

	buf := make([]byte, 1)
	B := func(v uint64) uint64 {
		lmp.fp.Seek(int64(v&mk), io.SeekStart)
		lmp.fp.Read(buf)
		return uint64(buf[0])
	}
	b := func(v uint64) byte { return byte(B(v)) }

	faststep := func(v2 uint64, idx uint64) {
		b := B(as ^ v2)
		as = as<<7 ^ as>>5 ^ v2<<20 ^ v2<<16 ^ v2 ^ b<<20 ^ b<<12 ^ b<<4
		s1 = s1<<9 ^ s1>>3 ^ hs[idx]
		hs[idx] = s1 ^ as
		s1, s2, s3 = s3, s1, s2
	}

	// Define a function to move the state by one byte.  This is not intended to be fast
	// Requires the previous byte read to process the next byte read.  Forces serial evaluation
	// and removes the possibility of scheduling byte access.
	//
	// (Note that use of _ = 0 in lines below are to keep go fmt from messing with comments on the right of the page)
	step := func(v2 uint64, idx uint64) {
		s1 = s1<<9 ^ s1>>1 ^ as ^ B(as>>5^v2)<<3      // Shifts are not random.  They are selected to ensure that
		s1 = s1<<5 ^ s1>>3 ^ B(s1^v2)<<7              // Prior bytes pulled from the ByteMap contribute to the
		s1 = s1<<7 ^ s1>>7 ^ B(as^s1>>7)<<5           // next access of the ByteMap, either by contributing to
		s1 = s1<<11 ^ s1>>5 ^ B(v2^as>>11^s1)<<27     // the lower bits of the index, or in the upper bits that
		_ = 0                                         // move the access further in the map.
		hs[idx] = s1 ^ as ^ hs[idx]<<7 ^ hs[idx]>>13  //
		_ = 0                                         // We also pay attention not only to where the ByteMap bits
		as = as<<17 ^ as>>5 ^ s1 ^ B(as^s1>>27^v2)<<3 // are applied, but what bits we use in the indexing of
		as = as<<13 ^ as>>3 ^ B(as^s1)<<7             // the ByteMap
		as = as<<15 ^ as>>7 ^ B(as>>7^s1)<<11         //
		as = as<<9 ^ as>>11 ^ B(v2^as^s1)<<3          // Tests run against this set of shifts show that the
		_ = 0                                         // bytes pulled from the ByteMap are evenly distributed
		s1 = s1<<7 ^ s1>>27 ^ as ^ B(as>>3)<<13       // over possible byte values (0-255) and indexes into
		s1 = s1<<3 ^ s1>>13 ^ B(s1^v2)<<11            // the ByteMap are also evenly distributed, and the
		s1 = s1<<8 ^ s1>>11 ^ B(as^s1>>11)<<9         // deltas between bytes provided map to a curve expected
		s1 = s1<<6 ^ s1>>9 ^ B(v2^as^s1)<<3           // (fewer maximum and minimum deltas, and most deltas around
		_ = 0                                         // zero.
		as = as<<23 ^ as>>3 ^ s1 ^ B(as^v2^s1>>3)<<7
		as = as<<17 ^ as>>7 ^ B(as^s1>>3)<<5
		as = as<<13 ^ as>>5 ^ B(as>>5^s1)<<1
		as = as<<11 ^ as>>1 ^ B(v2^as^s1)<<7

		s1 = s1<<5 ^ s1>>3 ^ as ^ B(as>>7^s1>>3)<<6
		s1 = s1<<8 ^ s1>>6 ^ B(s1^v2)<<11
		s1 = s1<<11 ^ s1>>11 ^ B(as^s1>>11)<<5
		s1 = s1<<7 ^ s1>>5 ^ B(v2^as>>7^as^s1)<<17

		s2 = s2<<3 ^ s2>>17 ^ s1 ^ B(as^s2>>5^v2)<<13
		s2 = s2<<6 ^ s2>>13 ^ B(s2)<<11
		s2 = s2<<11 ^ s2>>11 ^ B(as^s1^s2>>11)<<23
		s2 = s2<<4 ^ s2>>23 ^ B(v2^as>>8^as^s2>>10)<<1

		s1 = s2<<3 ^ s2>>1 ^ hs[idx] ^ v2
		as = as<<9 ^ as>>7 ^ s1>>1 ^ B(s2>>1^hs[idx])<<5

		s1, s2, s3 = s3, s1, s2
	}

	idx := uint64(0)
	// Fast spin to prevent caching state
	for _, v2 := range src {
		if idx >= lmp.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		faststep(uint64(v2), idx)
		idx++
	}

	idx = 0
	// Actual work to compute the hash
	for _, v2 := range src {
		if idx >= lmp.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		step(uint64(v2), idx)
		idx++
	}

	// Reduction pass
	// Done by Interating over hs[] to produce the bytes[] hash
	//
	// At this point, we have HBits of state in hs.  We need to reduce them down to a byte,
	// And we do so by doing a bit more bitwise math, and mapping the values through our byte map.

	bytes := make([]byte, lmp.HashSize)
	// Roll over all the hs (one int64 value for every byte in the resulting hash) and reduce them to byte values
	for i := len(hs) - 1; i >= 0; i-- {
		step(hs[i], uint64(i))      // Step the hash functions and then
		bytes[i] = b(as) ^ b(hs[i]) // Xor two resulting sequences
	}

	// Return the resulting hash
	return bytes
}
