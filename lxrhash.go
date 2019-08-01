// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

// LXRHash holds one instance of a hash function with a specific seed and map size
type LXRHash struct {
	ByteMap     []byte // Integer Offsets
	MapSize     uint64 // Size of the translation table
	MapSizeBits uint64 // Size of the ByteMap in Bits
	Passes      uint64 // Passes to generate the rand table
	Seed        uint64 // An arbitrary number used to create the tables.
	HashSize    uint64 // Number of bytes in the hash
	verbose     bool
}

func (lx LXRHash) B(v uint64) uint64 {
	return uint64(lx.ByteMap[v&(lx.MapSize-1)])
}

func (lx LXRHash) b(v uint64) byte {
	return byte(lx.B(v))
}

func (lx LXRHash) step(as, s1, s2, s3, v2, idx uint64) (uint64, uint64, uint64, uint64, uint64) {
	s1 = s1<<9 ^ s1>>1 ^ as ^ lx.B(as>>5^v2)<<3
	s1 = s1<<5 ^ s1>>3 ^ lx.B(s1^v2)<<7
	s1 = s1<<7 ^ s1>>7 ^ lx.B(as^s1>>7)<<5
	s1 = s1<<11 ^ s1>>5 ^ lx.B(v2^as>>11^s1)<<27

	idx = s1 ^ as ^ idx<<7 ^ idx>>13

	as = as<<17 ^ as>>5 ^ s1 ^ lx.B(as^s1>>27^v2)<<3
	as = as<<13 ^ as>>3 ^ lx.B(as^s1)<<7
	as = as<<15 ^ as>>7 ^ lx.B(as>>7^s1)<<11
	as = as<<9 ^ as>>11 ^ lx.B(v2^as^s1)<<3

	s1 = s1<<7 ^ s1>>27 ^ as ^ lx.B(as>>3)<<13
	s1 = s1<<3 ^ s1>>13 ^ lx.B(s1^v2)<<11
	s1 = s1<<8 ^ s1>>11 ^ lx.B(as^s1>>11)<<9
	s1 = s1<<6 ^ s1>>9 ^ lx.B(v2^as^s1)<<3

	as = as<<23 ^ as>>3 ^ s1 ^ lx.B(as^v2^s1>>3)<<7
	as = as<<17 ^ as>>7 ^ lx.B(as^s1>>3)<<5
	as = as<<13 ^ as>>5 ^ lx.B(as>>5^s1)<<1
	as = as<<11 ^ as>>1 ^ lx.B(v2^as^s1)<<7

	s1 = s1<<5 ^ s1>>3 ^ as ^ lx.B(as>>7^s1>>3)<<6
	s1 = s1<<8 ^ s1>>6 ^ lx.B(s1^v2)<<11
	s1 = s1<<11 ^ s1>>11 ^ lx.B(as^s1>>11)<<5
	s1 = s1<<7 ^ s1>>5 ^ lx.B(v2^as>>7^as^s1)<<17

	s2 = s2<<3 ^ s2>>17 ^ s1 ^ lx.B(as^s2>>5^v2)<<13
	s2 = s2<<6 ^ s2>>13 ^ lx.B(s2)<<11
	s2 = s2<<11 ^ s2>>11 ^ lx.B(as^s1^s2>>11)<<23
	s2 = s2<<4 ^ s2>>23 ^ lx.B(v2^as>>8^as^s2>>10)<<1

	s1 = s2<<3 ^ s2>>1 ^ idx ^ v2
	as = as<<9 ^ as>>7 ^ s1>>1 ^ lx.B(s2>>1^idx)<<5

	s1, s2, s3 = s3, s1, s2
	return as, s1, s2, s3, idx
}

func (lx LXRHash) faststep(as, s1, s2, s3, v2 uint64, idx uint64, hs []uint64) (uint64, uint64, uint64, uint64) {
	as = idx<<1 ^ idx>>3 ^ as<<7 ^ as>>5
	s1 = s1<<9 ^ s1>>3 ^ as
	hs[idx] = s1 ^ as
	as, s1, s2, s3 = s3, as, s1, s2
	return as, s1, s2, s3
}

func (lx LXRHash) Hash(src []byte) []byte {
	// Keep the byte intermediate results as int64 values until reduced.
	hs := make([]uint64, lx.HashSize)
	// as accumulates the state as we walk through applying the source data through the lookup map
	// and combine it with the state we are building up.
	var as = lx.Seed
	// We keep a series of states, and roll them along through each byte of source processed.
	var s1, s2, s3 uint64
	// Since MapSize is specified in bits, the index mask is the size-1

	idx := uint64(0)
	// Fast spin to prevent caching state
	for _, v2 := range src {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		as, s1, s2, s3 = lx.faststep(as, s1, s2, s3, uint64(v2), idx, hs)
		idx++
	}

	idx = 0
	// Actual work to compute the hash
	for _, v2 := range src {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		as, s1, s2, s3, hs[idx] = lx.step(as, s1, s2, s3, uint64(v2), hs[idx])
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
		as, s1, s2, s3, hs[i] = lx.step(as, s1, s2, s3, hs[i], hs[i]) // Step the hash functions and then
		bytes[i] = lx.b(as) ^ lx.b(hs[i])                             // Xor two resulting sequences
	}

	// Return the resulting hash
	return bytes
}
