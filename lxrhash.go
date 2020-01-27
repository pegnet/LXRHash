// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import "encoding/binary"

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

// AbortSettings indicated the proper settings to abort if a hash is found
// to be less than the target. Aborting early can save a few hash table
// accesses
func AbortSettings(target uint64) (abortByte int, abortVal uint8) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, target)
	for i := range data {
		if data[i] != 0xFF {
			return i, data[i]
		}
	}
	return -1, 0
}

type HashParallelItem struct {
	src                     []byte
	hs                      []uint64
	as, s1, s2, s3, idx, v2 uint64
}

// HashParallel takes the arbitrary input and returns the resulting hash of length HashSize.
// The batch must have at least one entry.
// The base is prefixed to all items in the batch.
func (lx LXRHash) HashParallel(base []byte, batch [][]byte) [][]byte {
	var work []*HashParallelItem
	for _, src := range batch {
		work = append(work, &HashParallelItem{
			src: src,
			as:  lx.Seed,
			hs:  make([]uint64, lx.HashSize),
		})
	}

	mk := lx.MapSize - 1

	B := func(v uint64) uint64 { return uint64(lx.ByteMap[v&mk]) }
	b := func(v uint64) byte { return byte(B(v)) }

	source := func(work *HashParallelItem, i int) byte {
		if i < len(base) {
			return base[i]
		}
		return work.src[i-len(base)]
	}

	faststep := func(work []*HashParallelItem, i int, idx uint64) {
		for _, h := range work {
			v2 := uint64(source(h, i))
			b := B(h.as ^ v2)
			h.as = h.as<<7 ^ h.as>>5 ^ v2<<20 ^ v2<<16 ^ v2 ^ b<<20 ^ b<<12 ^ b<<4
			h.s1 = h.s1<<9 ^ h.s1>>3 ^ h.hs[idx]
			h.hs[idx] = h.s1 ^ h.as
			h.s1, h.s2, h.s3 = h.s3, h.s1, h.s2
		}
	}

	step := func(work []*HashParallelItem, i int, idx uint64, reduce bool) {
		for _, h := range work {
			if reduce {
				h.v2 = uint64(h.hs[i])
			} else {
				h.v2 = uint64(source(h, i))
			}
			h.s1 = h.s1<<9 ^ h.s1>>1 ^ h.as ^ B(h.as>>5^h.v2)<<3
		}

		for _, h := range work {
			h.s1 = h.s1<<5 ^ h.s1>>3 ^ B(h.s1^h.v2)<<7
		}
		for _, h := range work {
			h.s1 = h.s1<<7 ^ h.s1>>7 ^ B(h.as^h.s1>>7)<<5
		}
		for _, h := range work {
			h.s1 = h.s1<<11 ^ h.s1>>5 ^ B(h.v2^h.as>>11^h.s1)<<27
			h.hs[idx] = h.s1 ^ h.as ^ h.hs[idx]<<7 ^ h.hs[idx]>>13
		}
		for _, h := range work {
			h.as = h.as<<17 ^ h.as>>5 ^ h.s1 ^ B(h.as^h.s1>>27^h.v2)<<3
		}
		for _, h := range work {
			h.as = h.as<<13 ^ h.as>>3 ^ B(h.as^h.s1)<<7
		}
		for _, h := range work {
			h.as = h.as<<15 ^ h.as>>7 ^ B(h.as>>7^h.s1)<<11
		}
		for _, h := range work {
			h.as = h.as<<9 ^ h.as>>11 ^ B(h.v2^h.as^h.s1)<<3
		}
		for _, h := range work {
			h.s1 = h.s1<<7 ^ h.s1>>27 ^ h.as ^ B(h.as>>3)<<13
		}
		for _, h := range work {
			h.s1 = h.s1<<3 ^ h.s1>>13 ^ B(h.s1^h.v2)<<11
		}
		for _, h := range work {
			h.s1 = h.s1<<8 ^ h.s1>>11 ^ B(h.as^h.s1>>11)<<9
		}
		for _, h := range work {
			h.s1 = h.s1<<6 ^ h.s1>>9 ^ B(h.v2^h.as^h.s1)<<3
		}
		for _, h := range work {
			h.as = h.as<<23 ^ h.as>>3 ^ h.s1 ^ B(h.as^h.v2^h.s1>>3)<<7
		}
		for _, h := range work {
			h.as = h.as<<17 ^ h.as>>7 ^ B(h.as^h.s1>>3)<<5
		}
		for _, h := range work {
			h.as = h.as<<13 ^ h.as>>5 ^ B(h.as>>5^h.s1)<<1
		}
		for _, h := range work {
			h.as = h.as<<11 ^ h.as>>1 ^ B(h.v2^h.as^h.s1)<<7
		}

		for _, h := range work {
			h.s1 = h.s1<<5 ^ h.s1>>3 ^ h.as ^ B(h.as>>7^h.s1>>3)<<6
		}
		for _, h := range work {
			h.s1 = h.s1<<8 ^ h.s1>>6 ^ B(h.s1^h.v2)<<11
		}
		for _, h := range work {
			h.s1 = h.s1<<11 ^ h.s1>>11 ^ B(h.as^h.s1>>11)<<5
		}
		for _, h := range work {
			h.s1 = h.s1<<7 ^ h.s1>>5 ^ B(h.v2^h.as>>7^h.as^h.s1)<<17
		}

		for _, h := range work {
			h.s2 = h.s2<<3 ^ h.s2>>17 ^ h.s1 ^ B(h.as^h.s2>>5^h.v2)<<13
		}
		for _, h := range work {
			h.s2 = h.s2<<6 ^ h.s2>>13 ^ B(h.s2)<<11
		}
		for _, h := range work {
			h.s2 = h.s2<<11 ^ h.s2>>11 ^ B(h.as^h.s1^h.s2>>11)<<23
		}
		for _, h := range work {
			h.s2 = h.s2<<4 ^ h.s2>>23 ^ B(h.v2^h.as>>8^h.as^h.s2>>10)<<1
			h.s1 = h.s2<<3 ^ h.s2>>1 ^ h.hs[idx] ^ h.v2
		}
		for _, h := range work {
			h.as = h.as<<9 ^ h.as>>7 ^ h.s1>>1 ^ B(h.s2>>1^h.hs[idx])<<5
			h.s1, h.s2, h.s3 = h.s3, h.s1, h.s2
		}

	}

	idx := uint64(0)
	// Fast spin to prevent caching state
	for i := 0; i < len(base)+len(work[0].src); i++ {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		faststep(work, i, idx)
		idx++
	}

	idx = 0
	// Actual work to compute the hash
	for i := 0; i < len(base)+len(work[0].src); i++ {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		step(work, i, idx, false)
		idx++
	}

	ret := make([][]byte, len(batch))
	for i := range ret {
		ret[i] = make([]byte, lx.HashSize)
	}

	for i := int64(lx.HashSize - 1); i >= 0; i-- {
		step(work, int(i), uint64(i), true) // Step the hash functions and then
		for j, h := range work {
			ret[j][i] = b(h.as) ^ b(h.hs[i]) // Xor two resulting sequences
		}
	}

	// Return the resulting hash
	return ret
}

func (lx LXRHash) fastStepf(v2, as, s1, s2, s3, idx uint64, hs []uint64) (uint64, uint64, uint64, uint64) {
	b := uint64(lx.ByteMap[(as^v2)&(lx.MapSize-1)])
	as = as<<7 ^ as>>5 ^ v2<<20 ^ v2<<16 ^ v2 ^ b<<20 ^ b<<12 ^ b<<4
	s1 = s1<<9 ^ s1>>3 ^ hs[idx]
	hs[idx] = s1 ^ as
	s1, s2, s3 = s3, s1, s2
	return as, s1, s2, s3
}

func (lx LXRHash) stepf(as, s1, s2, s3, v2 uint64, hs []uint64, idx uint64, mk uint64) (uint64, uint64, uint64, uint64) {
	s1 = s1<<9 ^ s1>>1 ^ as ^ uint64(lx.ByteMap[(as>>5^v2)&mk])<<3      // Shifts are not random.  They are selected to ensure that
	s1 = s1<<5 ^ s1>>3 ^ uint64(lx.ByteMap[(s1^v2)&mk])<<7              // Prior bytes pulled from the ByteMap contribute to the
	s1 = s1<<7 ^ s1>>7 ^ uint64(lx.ByteMap[(as^s1>>7)&mk])<<5           // next access of the ByteMap, either by contributing to
	s1 = s1<<11 ^ s1>>5 ^ uint64(lx.ByteMap[(v2^as>>11^s1)&mk])<<27     // the lower bits of the index, or in the upper bits that
	_ = 0                                                               // move the access further in the map.
	hs[idx] = s1 ^ as ^ hs[idx]<<7 ^ hs[idx]>>13                        //
	_ = 0                                                               // We also pay attention not only to where the ByteMap bits
	as = as<<17 ^ as>>5 ^ s1 ^ uint64(lx.ByteMap[(as^s1>>27^v2)&mk])<<3 // are applied, but what bits we use in the indexing of
	as = as<<13 ^ as>>3 ^ uint64(lx.ByteMap[(as^s1)&mk])<<7             // the ByteMap
	as = as<<15 ^ as>>7 ^ uint64(lx.ByteMap[(as>>7^s1)&mk])<<11         //
	as = as<<9 ^ as>>11 ^ uint64(lx.ByteMap[(v2^as^s1)&mk])<<3          // Tests run against this set of shifts show that the
	_ = 0                                                               // bytes pulled from the ByteMap are evenly distributed
	s1 = s1<<7 ^ s1>>27 ^ as ^ uint64(lx.ByteMap[(as>>3)&mk])<<13       // over possible byte values (0-255) and indexes into
	s1 = s1<<3 ^ s1>>13 ^ uint64(lx.ByteMap[(s1^v2)&mk])<<11            // the ByteMap are also evenly distributed, and the
	s1 = s1<<8 ^ s1>>11 ^ uint64(lx.ByteMap[(as^s1>>11)&mk])<<9         // deltas between bytes provided map to a curve expected
	s1 = s1<<6 ^ s1>>9 ^ uint64(lx.ByteMap[(v2^as^s1)&mk])<<3           // (fewer maximum and minimum deltas, and most deltas around
	_ = 0                                                               // zero.
	as = as<<23 ^ as>>3 ^ s1 ^ uint64(lx.ByteMap[(as^v2^s1>>3)&mk])<<7
	as = as<<17 ^ as>>7 ^ uint64(lx.ByteMap[(as^s1>>3)&mk])<<5
	as = as<<13 ^ as>>5 ^ uint64(lx.ByteMap[(as>>5^s1)&mk])<<1
	as = as<<11 ^ as>>1 ^ uint64(lx.ByteMap[(v2^as^s1)&mk])<<7
	s1 = s1<<5 ^ s1>>3 ^ as ^ uint64(lx.ByteMap[(as>>7^s1>>3)&mk])<<6
	s1 = s1<<8 ^ s1>>6 ^ uint64(lx.ByteMap[(s1^v2)&mk])<<11
	s1 = s1<<11 ^ s1>>11 ^ uint64(lx.ByteMap[(as^s1>>11)&mk])<<5
	s1 = s1<<7 ^ s1>>5 ^ uint64(lx.ByteMap[(v2^as>>7^as^s1)&mk])<<17
	s2 = s2<<3 ^ s2>>17 ^ s1 ^ uint64(lx.ByteMap[(as^s2>>5^v2)&mk])<<13
	s2 = s2<<6 ^ s2>>13 ^ uint64(lx.ByteMap[(s2)&mk])<<11
	s2 = s2<<11 ^ s2>>11 ^ uint64(lx.ByteMap[(as^s1^s2>>11)&mk])<<23
	s2 = s2<<4 ^ s2>>23 ^ uint64(lx.ByteMap[(v2^as>>8^as^s2>>10)&mk])<<1
	s1 = s2<<3 ^ s2>>1 ^ hs[idx] ^ v2
	as = as<<9 ^ as>>7 ^ s1>>1 ^ uint64(lx.ByteMap[(s2>>1^hs[idx])&mk])<<5

	s1, s2, s3 = s3, s1, s2

	return as, s1, s2, s3
}

// FlatHash takes the arbitrary input and returns the resulting hash of length HashSize
// Does not use anonymous functions
func (lx LXRHash) FlatHash(src []byte) []byte {
	// Keep the byte intermediate results as int64 values until reduced.
	hs := make([]uint64, lx.HashSize)
	// as accumulates the state as we walk through applying the source data through the lookup map
	// and combine it with the state we are building up.
	var as = lx.Seed
	// We keep a series of states, and roll them along through each byte of source processed.
	var s1, s2, s3 uint64
	// Since MapSize is specified in bits, the index mask is the size-1
	mk := lx.MapSize - 1

	idx := uint64(0)
	// Fast spin to prevent caching state
	for _, v2 := range src {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}

		as, s1, s2, s3 = lx.fastStepf(uint64(v2), as, s1, s2, s3, idx, hs)
		idx++
	}

	idx = 0
	// Actual work to compute the hash
	for _, v2 := range src {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}

		as, s1, s2, s3 = lx.stepf(as, s1, s2, s3, uint64(v2), hs, idx, mk)
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
		as, s1, s2, s3 = lx.stepf(as, s1, s2, s3, uint64(hs[i]), hs, uint64(i), mk)
		bytes[i] = lx.ByteMap[as&mk] ^ lx.ByteMap[hs[i]&mk] // Xor two resulting sequences
	}

	// Return the resulting hash
	return bytes
}

// Hash takes the arbitrary input and returns the resulting hash of length HashSize
func (lx LXRHash) Hash(src []byte) []byte {
	// Keep the byte intermediate results as int64 values until reduced.
	hs := make([]uint64, lx.HashSize)
	// as accumulates the state as we walk through applying the source data through the lookup map
	// and combine it with the state we are building up.
	var as = lx.Seed
	// We keep a series of states, and roll them along through each byte of source processed.
	var s1, s2, s3 uint64
	// Since MapSize is specified in bits, the index mask is the size-1
	mk := lx.MapSize - 1

	B := func(v uint64) uint64 { return uint64(lx.ByteMap[v&mk]) }
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
		if idx >= lx.HashSize { // Use an if to avoid modulo math
			idx = 0
		}
		faststep(uint64(v2), idx)
		idx++
	}

	idx = 0
	// Actual work to compute the hash
	for _, v2 := range src {
		if idx >= lx.HashSize { // Use an if to avoid modulo math
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

	bytes := make([]byte, lx.HashSize)
	// Roll over all the hs (one int64 value for every byte in the resulting hash) and reduce them to byte values
	for i := len(hs) - 1; i >= 0; i-- {
		step(hs[i], uint64(i))      // Step the hash functions and then
		bytes[i] = b(as) ^ b(hs[i]) // Xor two resulting sequences
	}

	// Return the resulting hash
	return bytes
}
