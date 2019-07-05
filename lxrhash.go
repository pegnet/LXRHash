package lxr

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
	step := func(v2 uint64, idx uint64) {
		s1 = s1<<9 ^ s1>>7 ^ as ^ B(as)<<7
		s1 = s1 ^ B(s1^v2)<<3
		s1 = s1 ^ B(as^s1)<<17
		s1 = s1 ^ B(v2^as>>11^s1)<<23

		as = as<<17 ^ as>>1 ^ s1 ^ B(as^v2)<<9
		as = as ^ B(s1)<<8
		as = as ^ B(as^s1)<<23
		as = as ^ B(v2^as>>17^as^s1)<<17

		s1 = s1<<7 ^ s1>>27 ^ as ^ B(as)<<11
		s1 = s1 ^ B(s1^v2)<<9
		s1 = s1 ^ B(as^s1)<<33
		s1 = s1 ^ B(v2^as>>13^as^s1)<<23

		as = as<<23 ^ as>>3 ^ s1 ^ B(as^v2)<<13
		as = as ^ B(as^s1)<<11
		as = as ^ B(as^s1)<<40
		as = as ^ B(v2^as>>9^as^s1)<<17

		s1 = s1<<5 ^ s1>>3 ^ as ^ B(as)<<15
		s1 = s1 ^ B(s1^v2)<<5
		s1 = s1 ^ B(as^s1)<<43
		s1 = s1 ^ B(v2^as>>23^as^s1)<<23

		s2 = s2<<3 ^ s2>>17 ^ s1 ^ B(as^s1^v2)<<17
		s2 = s2 ^ B(s2)<<2
		s2 = s2 ^ B(as^s1^s2)<<51
		s2 = s2 ^ B(v2^as>>15^as^s2)<<17

		s1 = s1 ^ hs[idx]

		s1, s2, s3 = s3, s1, s2
	}

	for i, v2 := range src {
		idx := uint64(i) % lx.HashSize
		step(uint64(v2), idx)
		// Set one of the hs[] using the last rolling value, the input byte v2,
		// the mapped byte bytemap, and the previous hs[] value
		hs[idx] = as ^ hs[idx]
	}

	// Reduction pass
	// Done by Interating over hs[] to produce the bytes[] hash
	//
	// At this point, we have HBits of state in hs.  We need to reduce them down to a byte,
	// And we do so by doing a bit more bitwise math, and mapping the values through our byte map.

	bytes := make([]byte, lx.HashSize)
	// Roll over all the hs (32 int64 values)
	for i, h := range hs {
		step(h, uint64(i))
		// Set a byte
		bytes[i] = b(as) ^ b(hs[i])
	}

	// Return the resulting hash
	return bytes
}
