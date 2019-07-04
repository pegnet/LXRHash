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

	// Define a function to move the state by one byte.
	step := func(v2 uint64) {
		s1 = s1<<9 ^ s1>>7 ^ as ^ B(as)<<7^B(s1^v2)<<3 ^ B(as^s1)<<17^ B(v2^as>>11^s1)<<23
		as = as<<17 ^ as>>1 ^ s1 ^ B(as^v2)<<9^B(s1)<<8 ^ B(as^s1)<<23^ B(v2^as>>17^as^s1)<<17
		s1 = s1<<7 ^ s1>>27 ^ as ^ B(as)<<11^B(s1^v2)<<9 ^ B(as^s1)<<33^ B(v2^as>>13^as^s1)<<23
		as = as<<23 ^ as>>3 ^ s1 ^ B(as^v2)<<13^B(s1)<<11 ^ B(as^s1)<<40^ B(v2^as>>9^as^s1)<<17
		s1 = s1<<5 ^ s1>>3 ^ as ^ B(as)<<15^B(s1^v2)<<5 ^ B(as^s1)<<43^ B(v2^as>>23^as^s1)<<23
		s2 = s2<<3 ^ s2>>17 ^ s1 ^ B(as^v2)<<17^B(s1)<<2 ^ B(as^s1)<<51^ B(v2^as>>15^as^s1)<<17
		s1, s2, s3 = s3, s1, s2
	}

	for i, v2 := range src {
		idx := uint64(i) % lx.HashSize
		step(uint64(v2))
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
		step(h)
		// Set a byte
		bytes[i] = b(as) ^ bytes[i]
	}

	// Return the resulting hash
	return bytes
}
