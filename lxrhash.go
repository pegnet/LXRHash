package lxr

type LXRHash struct {
	ByteMap     []byte // Integer Offsets
	MapSize     uint64 // Size of the translation table
	MapSizeBits uint64 // Size of the ByteMap in Bits
	Passes      uint64 // Passes to generate the rand table
	Seed        uint64 // An arbitrary number used to create the tables.
	HashSize    uint64 // Number of bytes in the hash
}

func (w LXRHash) Hash(src []byte) []byte {

	// Keep the byte intermediate results as int64 values until reduced.
	hashes := make([]uint64, w.HashSize)
	// The initial offset into the lookup table is the length of the input.
	// This prevents concatenation attacks, which adds to the protection from
	// the reduction pass.
	var lastStage = w.Seed
	// We keep a series of previous states, and roll them along through each
	// byte of source processed.
	var stages, stages2 [11]uint64
	MapMask := w.MapSize - 1

	// Define a function to move the state by one byte.
	step := func(i uint64, v2 uint64) {
		stages[0] = stages[0] ^ lastStage ^ v2 ^ uint64(w.ByteMap[(lastStage^v2<<9)%w.MapSize])<<4
		for i := len(stages) - 1; i >= 0; i-- {
			stage := stages[i]
			if i > 0 {
				stages[i] = stages[i-1]<<7 ^ stages[i-1]>>1 ^ stage ^ uint64(w.ByteMap[(stage^v2<<9)&MapMask])<<16
			}
			lastStage = stage<<32 ^ lastStage<<11 ^ lastStage>>1
		}
		stages, stages2 = stages2, stages
	}

	for i, v2 := range src {
		idx := uint64(i)
		step(idx, uint64(v2))
		// Set one of the hashes[] using the last rolling value, the input byte v2,
		// the mapped byte bytemap, and the previous hashes[] value
		hash := hashes[idx%w.HashSize]
		hashes[idx%w.HashSize] = lastStage ^ hash<<21 ^ hash>>1
	}

	// Reduction pass
	// Done by Interating over hashes[] to produce the bytes[] hash
	//
	// At this point, we have HBits of state in hashes.  We need to reduce them down to a byte,
	// And we do so by doing a bit more bitwise math, and mapping the values through our byte map.

	bytes := make([]byte, w.HashSize)
	// Roll over all the hashes (32 int64 values)
	for i, h := range hashes {
		step(uint64(i), h)
		// Set a byte
		idx2 := (stages[0] ^ lastStage) & MapMask
		bytes[i] = w.ByteMap[idx2]
	}

	// Return the resulting hash
	return bytes
}
