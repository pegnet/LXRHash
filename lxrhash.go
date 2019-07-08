package lxr

const shift = 10

type LXRHash struct {
	ByteMap     []byte // Integer Offsets
	MapSize     uint64 // Size of the translation table
	MapSizeBits uint64 // Size of the ByteMap in Bits
	Passes      uint64 // Passes to generate the rand table
	Seed        uint64 // An arbitrary number used to create the tables.
	HashSize    uint64 // Number of bytes in the hash
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
	mk := lx.MapSize - 1

	B := func(v uint64) uint64 { return uint64(lx.ByteMap[v&mk]) }
	b := func(v uint64) byte { return byte(B(v)) }

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

	var idx uint64
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
	for i, h := range hs {
		step(h, uint64(i))          // Step the hash functions and then
		bytes[i] = b(as) ^ b(hs[i]) // Xor two resulting sequences
	}

	// Return the resulting hash
	return bytes
}
