// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

// LXRHash holds one instance of a hash function with a specific seed and map size
type LXRHash struct {
	ByteMap     []byte // Integer Offsets
	MapSize     uint64 // Size of the translation table
	MapMask     uint64 // Bit Mask of valid bytemap indices
	MapSizeBits uint64 // Size of the ByteMap in Bits
	Passes      uint64 // Passes to generate the rand table
	Seed        uint64 // An arbitrary number used to create the tables.
	HashSize    uint64 // Number of bytes in the hash
	FirstIdx    uint64 // First Index used by LXRHash. (variance measures distribution of ByteMap access)
	verbose     bool
}

// Byte retrieves a single byte from the HashMap for the specified index
func (lx *LXRHash) Byte(v uint64) byte {
	return lx.ByteMap[v&lx.MapMask]
}

// UByte retrieves a single byte as uint64 from the HashMap for the specified index
func (lx *LXRHash) UByte(v uint64) uint64 {
	return uint64(lx.ByteMap[v&lx.MapMask])
}

// faststep is a simpler step that aims to intialize different start values for different length inputs.
// prevents someone from pre-calculating a common prefix
func (lx *LXRHash) faststep(hs []uint64, as, s1, s2, s3, v2, idx uint64) (uint64, uint64, uint64, uint64) {
	b := lx.UByte(as ^ v2)
	as = as<<7 ^ as>>5 ^ v2<<20 ^ v2<<16 ^ v2 ^ b<<20 ^ b<<12 ^ b<<4
	s1 = s1<<9 ^ s1>>3 ^ hs[idx]
	hs[idx] = s1 ^ as
	return as, s3, s1, s2
}

// step is a set of shifts intended to give the hash its pseudorandom attribute.
// Shifts are not random, they are selected to ensure that prior bytes pulled
// from the ByteMap contribute to the next access of the ByteMap, either by
// contributing to the lower bits of the index, or in the upper bits that
// move the access further in the map.
//
// We also pay attention not only to where the ByteMap bits are applied,
// but what bits we use in the indexing of the ByteMap
//
// Tests run against this set of shifts show that the bytes pulled from the
// ByteMap are evenly distributed over possible byte values (0-255) and indexes
// into the ByteMap are also evenly distributed, and the deltas between bytes
// provided map to a curve expected (fewer maximum and minimum deltas, and
// most deltas around zero.)
func (lx *LXRHash) step(hs []uint64, as, s1, s2, s3, v2, idx uint64) (uint64, uint64, uint64, uint64) {
	s1 = s1<<9 ^ s1>>1 ^ as ^ lx.UByte(as>>5^v2)<<3
	s1 = s1<<5 ^ s1>>3 ^ lx.UByte(s1^v2)<<7
	s1 = s1<<7 ^ s1>>7 ^ lx.UByte(as^s1>>7)<<5
	s1 = s1<<11 ^ s1>>5 ^ lx.UByte(v2^as>>11^s1)<<27

	hs[idx] = s1 ^ as ^ hs[idx]<<7 ^ hs[idx]>>13

	as = as<<17 ^ as>>5 ^ s1 ^ lx.UByte(as^s1>>27^v2)<<3
	as = as<<13 ^ as>>3 ^ lx.UByte(as^s1)<<7
	as = as<<15 ^ as>>7 ^ lx.UByte(as>>7^s1)<<11
	as = as<<9 ^ as>>11 ^ lx.UByte(v2^as^s1)<<3

	s1 = s1<<7 ^ s1>>27 ^ as ^ lx.UByte(as>>3)<<13
	s1 = s1<<3 ^ s1>>13 ^ lx.UByte(s1^v2)<<11
	s1 = s1<<8 ^ s1>>11 ^ lx.UByte(as^s1>>11)<<9
	s1 = s1<<6 ^ s1>>9 ^ lx.UByte(v2^as^s1)<<3

	as = as<<23 ^ as>>3 ^ s1 ^ lx.UByte(as^v2^s1>>3)<<7
	as = as<<17 ^ as>>7 ^ lx.UByte(as^s1>>3)<<5
	as = as<<13 ^ as>>5 ^ lx.UByte(as>>5^s1)<<1
	as = as<<11 ^ as>>1 ^ lx.UByte(v2^as^s1)<<7

	s1 = s1<<5 ^ s1>>3 ^ as ^ lx.UByte(as>>7^s1>>3)<<6
	s1 = s1<<8 ^ s1>>6 ^ lx.UByte(s1^v2)<<11
	s1 = s1<<11 ^ s1>>11 ^ lx.UByte(as^s1>>11)<<5
	s1 = s1<<7 ^ s1>>5 ^ lx.UByte(v2^as>>7^as^s1)<<17

	s2 = s2<<3 ^ s2>>17 ^ s1 ^ lx.UByte(as^s2>>5^v2)<<13
	s2 = s2<<6 ^ s2>>13 ^ lx.UByte(s2)<<11
	s2 = s2<<11 ^ s2>>11 ^ lx.UByte(as^s1^s2>>11)<<23
	s2 = s2<<4 ^ s2>>23 ^ lx.UByte(v2^as>>8^as^s2>>10)<<1

	s1 = s2<<3 ^ s2>>1 ^ hs[idx] ^ v2
	as = as<<9 ^ as>>7 ^ s1>>1 ^ lx.UByte(s2>>1^hs[idx])<<5

	return as, s3, s1, s2
}

// Hash takes the arbitrary input and returns the resulting hash of length HashSize
func (lx *LXRHash) Hash(src []byte) []byte {

	// intermediate results
	hs := make([]uint64, lx.HashSize)

	// state variables
	var as, s1, s2, s3, idx uint64
	as = lx.Seed

	// fast steps
	for _, v := range src {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		as, s1, s2, s3 = lx.faststep(hs, as, s1, s2, s3, uint64(v), idx)
		idx++
	}

	idx = 0
	// Actual work to compute the hash
	for i := range src {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		if i == 0 {
			lx.FirstIdx = (as>>5 ^ hs[i]) & lx.MapMask
		}
		as, s1, s2, s3 = lx.step(hs, as, s1, s2, s3, uint64(src[i]), idx)
		idx++
	}

	// Reduction pass
	// Done by Interating over hs[] to produce the bytes[] hash
	//
	// At this point, we have HBits of state in hs.  We need to reduce them down to a byte,
	// And we do so by doing a bit more bitwise math, and mapping the values through our byte map.

	bytes := make([]byte, lx.HashSize)
	// Roll over all the hs (one int64 value for every byte in the resulting hash) and reduce them to byte values
	for i := len(hs) - 1; i >= 0; i-- {
		as, s1, s2, s3 = lx.step(hs, as, s1, s2, s3, hs[i], uint64(i)) // Step the hash functions and then
		bytes[i] = lx.Byte(as) ^ lx.Byte(hs[i])                        // Xor two resulting sequences
	}

	// Return the resulting hash
	return bytes
}
