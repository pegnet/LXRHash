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
	hashes := make([]int64,w.HashSize)
	// The intital offset into the lookup table is the length of the input.
	// This prevents concatenation attacks, which adds to the protection from
	// the reduction pass.
	var offset = int64(len(src))
	// We keep a series of previous states, and roll them along through each
	// byte of source processed.
	var last1, last2, last3 int64

	v := byte(offset)
	var idx1, idx2 int64

	step := func(i int, v2 int64) {
		// combine the l values, the previous offset, and the hashes[i]
		offset = last1<<7 ^ last2<<3 ^ last3<<9 ^ offset<<8 ^ offset>>1 ^ idx2 ^ int64(v)

		// Take the byte from source (v2) and map it through the lookup table
		// using the offset being maintained, and the rolling lastX values
		idx1 = int64(uint64(idx1^offset^v2) % uint64(w.MapSize))
		v = w.ByteMap[idx1] ^ v

		// Roll the set of last values, leaving lingering influences from past
		// values.
		last3 = last2>>2 ^ last3
		last2 = last1<<3 ^ last2
		last1 = int64(v) ^ v2 ^ last1<<1

		idx2 = idx2 ^ hashes[uint32(i)%w.HashSize]
	}

	// Pass through the source bytes, building up lastX values, hashes[], and offset
	for i, v2 := range src {
		step(i, int64(v2))
		// Set one of the hashes[] using the last rolling value, the input byte v2,
		// the mapped byte v, and the previous hashes[] value
		hashes[uint32(i)%w.HashSize] = last3 ^ int64(v^v2) ^ idx2

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
		idx2 := int64(uint64(int64(v)^offset) % uint64(w.MapSize))
		bytes[i] = w.ByteMap[idx2]
	}

	// Return the resulting hash
	return bytes
}
