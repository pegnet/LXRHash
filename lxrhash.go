package lxr

//todo go through and make all the types unsigned, to avoid conversions.

type LXRHash struct {
	ByteMap  []byte // Integer Offsets
	MapSize  int64  // Size of the translation table
	Passes   int    // Passes to generate the rand table
	Seed     int64  // An arbitrary number used to create the tables.
	HashSize uint32 // Number of bytes in the hash
}

// Hash()
// Takes a source of bytes, returns a 32 byte (256 bit) hash
func (w LXRHash) Hash(src []byte) []byte {

	// Keep the byte intermediate results as int64 values until reduced.
	hashes := make([]int64, w.HashSize)
	// The intital offset into the lookup table is the length of the input.
	// This prevents concatenation attacks, which adds to the protection from
	// the reduction pass.
	var lastStage = int64(len(src)) ^ w.Seed ^ int64(w.HashSize)
	// We keep a series of previous states, and roll them along through each
	// byte of source processed.
	var stages,stages2 [10]int64
	v := w.ByteMap[lastStage%w.MapSize]

	step := func(i int, v2 int64) int64 {
		lastStage = v2 ^ int64(i)<<16 ^ lastStage
		for i, stage := range stages {
			ui := uint64(i)
			stages[i] = stage<<(31^ui) ^ stage>>(7^ui) ^ lastStage
			lastStage = stage ^ lastStage<<5 ^ int64(v<<ui) ^
				int64(w.ByteMap[uint64(stage+v2)%uint64(w.MapSize)])
		}
		stages, stages2 = stages2, stages
		v = w.ByteMap[uint64(lastStage)%uint64(w.MapSize)] ^ v
		return hashes[uint32(i)%w.HashSize]
	}

	for i, v2 := range src {
		step(i, int64(v2))
		// Set one of the hashes[] using the last rolling value, the input byte v2,
		// the mapped byte v, and the previous hashes[] value
		hashes[uint32(i)%w.HashSize] = lastStage ^ int64(v^w.ByteMap[uint64(stages[0]+int64(v2))%uint64(w.MapSize)])
	}

	// Reduction pass
	// Done by Interating over hashes[] to produce the bytes[] hash
	//
	// At this point, we have HBits of state in hashes.  We need to reduce them down to a byte,
	// And we do so by doing a bit more bitwise math, and mapping the values through our byte map.

	bytes := make([]byte, w.HashSize)

	// Roll over all the hashes (32 int64 values)
	for i, h := range hashes {
		step(i, h)
		// Set a byte
		idx2 := int64(uint64(int64(v)^lastStage) % uint64(w.MapSize))
		bytes[i] = w.ByteMap[idx2]
	}

	// Return the resulting hash
	return bytes
}
