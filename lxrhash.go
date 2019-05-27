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
	var s, s2 [11]uint64
	// Since MapSize is specified in bits, the index mask is the size-1
	mk := lx.MapSize - 1

	// Define a function to move the state by one byte.
	step := func(i uint64, v2 uint64) {
		s[0] = s[0] ^ as ^ v2 ^ uint64(lx.ByteMap[(as^v2<<9)&mk])<<4
		for i := len(s) - 1; i >= 0; i-- {
			if i > 0 {
				s[i] = s[i-1]<<7 ^ s[i-1]>>1 ^ s[i]<<17 ^ s[i]>>3 ^ uint64(lx.ByteMap[(s[i]^v2<<9)&mk])<<16
			}
			as = s[i]<<32  ^ s[i]>>3 ^ as<<11 ^ as>>1
		}
		s, s2 = s2, s
	}

	for i, v2 := range src {
		idx := uint64(i)
		step(idx, uint64(v2))
		// Set one of the hs[] using the last rolling value, the input byte v2,
		// the mapped byte bytemap, and the previous hs[] value
		hash := hs[idx%lx.HashSize]
		hs[idx%lx.HashSize] = as ^ hash<<21 ^ hash>>1
	}

	// Reduction pass
	// Done by Interating over hs[] to produce the bytes[] hash
	//
	// At this point, we have HBits of state in hs.  We need to reduce them down to a byte,
	// And we do so by doing a bit more bitwise math, and mapping the values through our byte map.

	bytes := make([]byte, lx.HashSize)
	// Roll over all the hs (32 int64 values)
	for i, h := range hs {
		step(uint64(i), h)
		// Set a byte
		idx2 := (s[0] ^ as) & mk
		bytes[i] = lx.ByteMap[idx2]
	}

	// Return the resulting hash
	return bytes
}
