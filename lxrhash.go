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
	var s1, s2 uint64
	// Since MapSize is specified in bits, the index mask is the size-1
	mk := lx.MapSize - 1

	// Define a function to move the state by one byte.
	step := func(v2 uint64) {
		s1 = s1 ^ as ^ v2 ^ uint64(lx.ByteMap[(as^v2<<9)&mk])<<4
		s2 = s1<<23 ^ s1>>5 ^ s2<<17 ^ s2>>3 ^ uint64(lx.ByteMap[(s2^v2<<9)&mk])<<11
		as = s2<<29 ^ s2>>7 ^ as<<37 ^ as>>1 ^ uint64(lx.ByteMap[(s1^v2<<9)&mk])<<13
		s1, s2 = s2, s1
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
		bytes[i] = lx.ByteMap[as&mk] ^ bytes[i]
	}

	// Return the resulting hash
	return bytes
}
